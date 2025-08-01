package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-fiber-template/internal/config"
	"go-fiber-template/internal/database"
	"go-fiber-template/internal/models"
	"go-fiber-template/internal/repositories"
	"go-fiber-template/internal/services"
)

func main() {
	fmt.Println("🔒 Account Lockout Demo")
	fmt.Println("=======================")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	if err := database.Initialize(&cfg.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	db := database.GetDB()

	// Clean up any existing test data
	db.Where("email = ?", "demo@example.com").Delete(&models.User{})

	// Create services
	txManager := repositories.NewTransactionManager(db)
	passwordHistoryRepo := repositories.NewPasswordHistoryRepository(db)
	passwordValidator := services.NewPasswordValidator(cfg, passwordHistoryRepo)

	// Create account locker with demo configuration
	lockoutConfig := &services.LockoutConfig{
		MaxFailedAttempts: 3,
		BaseLockDuration:  5 * time.Second, // Short duration for demo
		MaxLockDuration:   30 * time.Second,
		ProgressiveMode:   true,
	}
	accountLocker := services.NewAccountLocker(db, txManager, lockoutConfig)

	// Create auth service
	authService := services.NewAuthService(db, txManager, passwordValidator, accountLocker)

	ctx := context.Background()

	// Step 1: Register a demo user
	fmt.Println("\n📝 Step 1: Registering demo user...")
	registerReq := services.RegisterRequest{
		Email:     "demo@example.com",
		Password:  "DemoPassword123!@#",
		FirstName: "Demo",
		LastName:  "User",
	}

	_, err = authService.Register(ctx, registerReq)
	if err != nil {
		log.Fatalf("Failed to register user: %v", err)
	}
	fmt.Println("✅ User registered successfully")

	// Step 2: Demonstrate failed login attempts
	fmt.Println("\n🚫 Step 2: Attempting login with wrong password...")
	loginReq := services.LoginRequest{
		Email:    "demo@example.com",
		Password: "WrongPassword123!",
	}

	for i := 1; i <= 4; i++ {
		fmt.Printf("\n🔑 Attempt %d: ", i)

		// Check if account is locked before attempting
		isLocked, err := accountLocker.IsAccountLocked(ctx, loginReq.Email)
		if err != nil {
			log.Printf("Error checking lock status: %v", err)
			continue
		}

		if isLocked {
			fmt.Printf("❌ Account is locked - cannot attempt login")

			// Get lockout info
			info, err := accountLocker.GetLockoutInfo(ctx, loginReq.Email)
			if err == nil && info.LockedUntil != nil {
				fmt.Printf(" (locked until %s)", info.LockedUntil.Format("15:04:05"))
			}
			continue
		}

		// Attempt login
		_, err = authService.Login(ctx, loginReq)
		if err != nil {
			fmt.Printf("❌ Login failed: %s", err.Error())

			// Get updated lockout info
			info, err := accountLocker.GetLockoutInfo(ctx, loginReq.Email)
			if err == nil {
				fmt.Printf(" (Failed attempts: %d/%d)", info.FailedAttempts, lockoutConfig.MaxFailedAttempts)
				if info.IsLocked {
					fmt.Printf(" - ACCOUNT LOCKED!")
				}
			}
		} else {
			fmt.Printf("✅ Login successful")
		}
	}

	// Step 3: Wait for lock to expire
	fmt.Println("\n⏳ Step 3: Waiting for account lock to expire...")
	for {
		isLocked, err := accountLocker.IsAccountLocked(ctx, loginReq.Email)
		if err != nil {
			log.Printf("Error checking lock status: %v", err)
			break
		}

		if !isLocked {
			fmt.Println("✅ Account is now unlocked")
			break
		}

		fmt.Print(".")
		time.Sleep(1 * time.Second)
	}

	// Step 4: Successful login
	fmt.Println("\n🎉 Step 4: Attempting login with correct password...")
	correctLoginReq := services.LoginRequest{
		Email:    "demo@example.com",
		Password: "DemoPassword123!@#",
	}

	_, err = authService.Login(ctx, correctLoginReq)
	if err != nil {
		fmt.Printf("❌ Login failed: %s\n", err.Error())
	} else {
		fmt.Println("✅ Login successful - failed attempts reset")
	}

	// Step 5: Verify failed attempts are reset
	info, err := accountLocker.GetLockoutInfo(ctx, loginReq.Email)
	if err == nil {
		fmt.Printf("📊 Current status: Failed attempts: %d, Locked: %t\n", info.FailedAttempts, info.IsLocked)
	}

	// Step 6: Demonstrate admin unlock
	fmt.Println("\n👨‍💼 Step 6: Demonstrating admin unlock functionality...")

	// Lock the account again
	for i := 0; i < 3; i++ {
		accountLocker.RecordFailedAttempt(ctx, loginReq.Email)
	}

	isLocked, _ := accountLocker.IsAccountLocked(ctx, loginReq.Email)
	fmt.Printf("Account locked: %t\n", isLocked)

	// Admin unlock
	err = accountLocker.UnlockAccount(ctx, loginReq.Email)
	if err != nil {
		fmt.Printf("❌ Admin unlock failed: %s\n", err.Error())
	} else {
		fmt.Println("✅ Admin unlock successful")
	}

	// Verify unlock
	isLocked, _ = accountLocker.IsAccountLocked(ctx, loginReq.Email)
	info, _ = accountLocker.GetLockoutInfo(ctx, loginReq.Email)
	fmt.Printf("📊 After admin unlock: Failed attempts: %d, Locked: %t\n", info.FailedAttempts, isLocked)

	// Clean up
	db.Where("email = ?", "demo@example.com").Delete(&models.User{})
	fmt.Println("\n🧹 Demo completed - test data cleaned up")
}
