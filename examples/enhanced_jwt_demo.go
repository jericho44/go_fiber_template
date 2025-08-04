package main

import (
	"fmt"
	"log"
	"time"

	"go-fiber-template/internal/services"
	"go-fiber-template/internal/utils"
)

// Enhanced JWT Token Rotation Demo
// This demonstrates the enhanced JWT token rotation and security features
func main() {
	fmt.Println("=== Enhanced JWT Token Rotation Demo ===")

	// Create JWT manager
	jwtManager := utils.NewJWTManager("demo-secret-key", 15*time.Minute, 7*24*time.Hour)

	// Demo 1: Basic token generation with context
	fmt.Println("\n1. Generating token pair with enhanced context...")

	tokenCtx := utils.TokenContext{
		UserID:            1,
		Email:             "demo@example.com",
		SessionID:         "demo-session-123",
		DeviceFingerprint: "demo-device-fingerprint",
		IPAddress:         "192.168.1.100",
	}

	tokenPair, err := jwtManager.GenerateTokenPairWithContext(tokenCtx)
	if err != nil {
		log.Fatalf("Failed to generate token pair: %v", err)
	}

	fmt.Printf("✓ Access Token generated (length: %d)\n", len(tokenPair.AccessToken))
	fmt.Printf("✓ Refresh Token generated (length: %d)\n", len(tokenPair.RefreshToken))
	fmt.Printf("✓ Session ID: %s\n", tokenPair.SessionID)
	fmt.Printf("✓ Expires at: %s\n", time.Unix(tokenPair.ExpiresAt, 0).Format(time.RFC3339))

	// Demo 2: Token validation
	fmt.Println("\n2. Validating access token...")

	claims, err := jwtManager.ValidateToken(tokenPair.AccessToken)
	if err != nil {
		log.Fatalf("Failed to validate token: %v", err)
	}

	fmt.Printf("✓ Token validated successfully\n")
	fmt.Printf("✓ User ID: %d\n", claims.UserID)
	fmt.Printf("✓ Email: %s\n", claims.Email)
	fmt.Printf("✓ Token Type: %s\n", claims.TokenType)
	fmt.Printf("✓ Session ID: %s\n", claims.SessionID)
	fmt.Printf("✓ Device Fingerprint: %s\n", claims.DeviceFingerprint)
	fmt.Printf("✓ IP Address: %s\n", claims.IPAddress)

	// Demo 3: Token fingerprinting
	fmt.Println("\n3. Demonstrating device fingerprinting...")

	fingerprintGen := utils.NewDeviceFingerprintGenerator()

	// Simulate different browser/device combinations
	testCases := []struct {
		name       string
		userAgent  string
		acceptLang string
		acceptEnc  string
		ipAddress  string
	}{
		{
			name:       "Chrome on Windows",
			userAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			acceptLang: "en-US,en;q=0.9",
			acceptEnc:  "gzip, deflate, br",
			ipAddress:  "192.168.1.100",
		},
		{
			name:       "Firefox on Linux",
			userAgent:  "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0",
			acceptLang: "en-US,en;q=0.5",
			acceptEnc:  "gzip, deflate",
			ipAddress:  "192.168.1.101",
		},
		{
			name:       "Safari on macOS",
			userAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
			acceptLang: "en-us",
			acceptEnc:  "gzip, deflate, br",
			ipAddress:  "192.168.1.102",
		},
	}

	for _, tc := range testCases {
		fingerprint := fingerprintGen.GenerateFingerprint(
			tc.userAgent,
			tc.acceptLang,
			tc.acceptEnc,
			"",
			tc.ipAddress,
		)
		fmt.Printf("✓ %s fingerprint: %s\n", tc.name, fingerprint[:16]+"...")
	}

	// Demo 4: Token hash generation for blacklisting
	fmt.Println("\n4. Demonstrating token hashing for security...")

	accessTokenHash := jwtManager.GetTokenHash(tokenPair.AccessToken)
	refreshTokenHash := jwtManager.GetTokenHash(tokenPair.RefreshToken)

	fmt.Printf("✓ Access token hash: %s\n", accessTokenHash[:16]+"...")
	fmt.Printf("✓ Refresh token hash: %s\n", refreshTokenHash[:16]+"...")

	// Demo 5: Token expiry extraction
	fmt.Println("\n5. Extracting token expiry information...")

	accessExpiry, err := jwtManager.GetTokenExpiry(tokenPair.AccessToken)
	if err != nil {
		log.Fatalf("Failed to get access token expiry: %v", err)
	}

	refreshExpiry, err := jwtManager.GetTokenExpiry(tokenPair.RefreshToken)
	if err != nil {
		log.Fatalf("Failed to get refresh token expiry: %v", err)
	}

	fmt.Printf("✓ Access token expires: %s (in %v)\n",
		accessExpiry.Format(time.RFC3339),
		time.Until(accessExpiry).Round(time.Second))
	fmt.Printf("✓ Refresh token expires: %s (in %v)\n",
		refreshExpiry.Format(time.RFC3339),
		time.Until(refreshExpiry).Round(time.Second))

	// Demo 6: Enhanced request context simulation
	fmt.Println("\n6. Simulating enhanced request context...")

	requestCtx := services.RequestContext{
		IPAddress: "192.168.1.100",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		Endpoint:  "GET /api/protected",
		Location:  "US",
		Timestamp: time.Now(),
	}

	fmt.Printf("✓ Request IP: %s\n", requestCtx.IPAddress)
	fmt.Printf("✓ User Agent: %s\n", requestCtx.UserAgent[:50]+"...")
	fmt.Printf("✓ Endpoint: %s\n", requestCtx.Endpoint)
	fmt.Printf("✓ Location: %s\n", requestCtx.Location)
	fmt.Printf("✓ Timestamp: %s\n", requestCtx.Timestamp.Format(time.RFC3339))

	// Demo 7: Anomaly detection configuration
	fmt.Println("\n7. Demonstrating anomaly detection configuration...")

	anomalyConfig := services.DefaultAnomalyDetectorConfig()

	fmt.Printf("✓ Max IP changes per hour: %d\n", anomalyConfig.MaxIPChangesPerHour)
	fmt.Printf("✓ Max user agent changes per day: %d\n", anomalyConfig.MaxUserAgentChangesPerDay)
	fmt.Printf("✓ Max requests per minute: %d\n", anomalyConfig.MaxRequestsPerMinute)
	fmt.Printf("✓ Max devices per user: %d\n", anomalyConfig.MaxDevicesPerUser)
	fmt.Printf("✓ Max location changes per hour: %d\n", anomalyConfig.MaxLocationChangesPerHour)

	// Demo 8: Security levels demonstration
	fmt.Println("\n8. Security levels for different scenarios...")

	securityScenarios := []struct {
		name        string
		description string
		level       string
	}{
		{
			name:        "Standard Login",
			description: "Same device, same location, normal usage",
			level:       "standard",
		},
		{
			name:        "Enhanced Security",
			description: "IP change or user agent change detected",
			level:       "enhanced",
		},
		{
			name:        "High Security",
			description: "Multiple anomalies or suspicious patterns",
			level:       "high_security",
		},
		{
			name:        "Token Rotation",
			description: "Token refreshed with security context changes",
			level:       "rotated",
		},
	}

	for _, scenario := range securityScenarios {
		fmt.Printf("✓ %s (%s): %s\n", scenario.name, scenario.level, scenario.description)
	}

	fmt.Println("\n=== Demo completed successfully! ===")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("• Enhanced token generation with session context")
	fmt.Println("• Device fingerprinting for security")
	fmt.Println("• Token hashing for blacklisting")
	fmt.Println("• Request context tracking")
	fmt.Println("• Anomaly detection configuration")
	fmt.Println("• Security level determination")
	fmt.Println("• Token expiry management")

	fmt.Println("\nEnhanced JWT Security Features:")
	fmt.Println("• Token rotation with new session IDs")
	fmt.Println("• Redis-based token blacklisting")
	fmt.Println("• Comprehensive anomaly detection")
	fmt.Println("• Session context validation")
	fmt.Println("• Security event logging")
	fmt.Println("• Device fingerprint validation")
	fmt.Println("• IP subnet change detection")
}
