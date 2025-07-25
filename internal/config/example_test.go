package config_test

import (
	"fmt"
	"log"
	"os"

	"go-fiber-template/internal/config"
)

func ExampleLoad() {
	// Set some environment variables for the example
	os.Setenv("APP_ENV", "development")
	os.Setenv("PORT", "8080")
	os.Setenv("JWT_SECRET", "this-is-a-very-long-secret-key-for-testing-purposes")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Use configuration
	fmt.Printf("Environment: %v\n", cfg.Env)
	fmt.Printf("Server Address: %s\n", cfg.Server.GetServerAddress())
	fmt.Printf("Database DSN: %s\n", cfg.Database.GetDSN())
	fmt.Printf("Is Development: %v\n", cfg.IsDevelopment())

	// Clean up
	os.Unsetenv("APP_ENV")
	os.Unsetenv("PORT")
	os.Unsetenv("JWT_SECRET")

	// Output:
	// Environment: development
	// Server Address: localhost:8080
	// Database DSN: host=localhost port=5432 user=postgres password= dbname=go_fiber_template sslmode=disable
	// Is Development: true
}
