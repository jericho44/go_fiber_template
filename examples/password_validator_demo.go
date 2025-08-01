package main

import (
	"context"
	"fmt"

	"go-fiber-template/internal/config"
	"go-fiber-template/internal/services"
)

func main() {
	fmt.Println("Password Validator Demo")
	fmt.Println("======================")

	// Create a password validator without database (for demo purposes)
	cfg := &config.Config{}
	passwordValidator := services.NewPasswordValidator(cfg, nil)

	// Test passwords
	testPasswords := []string{
		"weak",
		"password123",
		"Password123",
		"Password123!",
		"VeryStr0ng!P@ssw0rd2024",
		"ComplexP@ssw0rd123!",
	}

	fmt.Println("\n1. Password Complexity Validation:")
	fmt.Println("----------------------------------")
	for _, password := range testPasswords {
		result := passwordValidator.ValidateComplexity(password)
		fmt.Printf("Password: %-25s | Valid: %-5t | Score: %3d | Level: %s\n",
			password, result.Valid, result.Score, result.Strength.Level)

		if !result.Valid && len(result.Errors) > 0 {
			fmt.Printf("  Errors: %v\n", result.Errors[0])
		}
		if len(result.Suggestions) > 0 {
			fmt.Printf("  Suggestion: %s\n", result.Suggestions[0])
		}
		fmt.Println()
	}

	fmt.Println("\n2. Common Password Detection:")
	fmt.Println("-----------------------------")
	commonPasswords := []string{"password", "123456", "admin", "UniqueP@ssw0rd2024!"}
	for _, password := range commonPasswords {
		err := passwordValidator.CheckCommonPasswords(password)
		if err != nil {
			fmt.Printf("Password: %-20s | Status: COMMON (rejected)\n", password)
		} else {
			fmt.Printf("Password: %-20s | Status: UNIQUE (accepted)\n", password)
		}
	}

	fmt.Println("\n3. Password Strength Analysis:")
	fmt.Println("------------------------------")
	strengthPasswords := []string{"a", "abc123", "Password123", "VeryStr0ng!P@ssw0rd2024"}
	for _, password := range strengthPasswords {
		result := passwordValidator.GetPasswordStrength(password)
		fmt.Printf("Password: %-25s | Score: %3d | Level: %-12s | Time to crack: %s\n",
			password, result.Score, result.Level, result.TimeToCrack)
	}

	fmt.Println("\n4. Secure Password Generation:")
	fmt.Println("------------------------------")
	for i := 0; i < 3; i++ {
		generated := passwordValidator.GenerateSecurePassword()
		result := passwordValidator.GetPasswordStrength(generated)
		fmt.Printf("Generated: %-20s | Score: %3d | Level: %s\n",
			generated, result.Score, result.Level)
	}

	fmt.Println("\n5. Comprehensive Policy Validation:")
	fmt.Println("-----------------------------------")
	ctx := context.Background()
	testPassword := "TestP@ssw0rd123!"

	result := passwordValidator.ValidatePasswordPolicy(ctx, 0, testPassword)
	fmt.Printf("Password: %s\n", testPassword)
	fmt.Printf("Valid: %t\n", result.Valid)
	fmt.Printf("Score: %d\n", result.Score)
	fmt.Printf("Strength Level: %s\n", result.Strength.Level)

	if len(result.Errors) > 0 {
		fmt.Printf("Errors: %v\n", result.Errors)
	}
	if len(result.Warnings) > 0 {
		fmt.Printf("Warnings: %v\n", result.Warnings)
	}
	if len(result.Suggestions) > 0 {
		fmt.Printf("Suggestions: %v\n", result.Suggestions)
	}

	fmt.Printf("\nRequirements met:\n")
	for req, met := range result.Requirements {
		status := "✗"
		if met {
			status = "✓"
		}
		fmt.Printf("  %s %s\n", status, req)
	}

	fmt.Println("\nDemo completed successfully!")
}
