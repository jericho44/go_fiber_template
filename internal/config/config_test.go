package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Save original environment variables
	originalEnvs := make(map[string]string)
	envVars := []string{
		"APP_ENV", "PORT", "HOST", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD",
		"DB_NAME", "DB_SSLMODE", "JWT_SECRET", "JWT_ACCESS_EXPIRY", "JWT_REFRESH_EXPIRY",
		"CORS_ORIGINS", "RATE_LIMIT_MAX", "RATE_LIMIT_WINDOW", "SWAGGER_ENABLED",
		"SWAGGER_HOST", "SWAGGER_BASE_PATH",
	}

	for _, env := range envVars {
		originalEnvs[env] = os.Getenv(env)
	}

	// Clean up function
	cleanup := func() {
		for _, env := range envVars {
			if originalValue, exists := originalEnvs[env]; exists && originalValue != "" {
				os.Setenv(env, originalValue)
			} else {
				os.Unsetenv(env)
			}
		}
	}
	defer cleanup()

	t.Run("Load with default values", func(t *testing.T) {
		// Clear all environment variables
		for _, env := range envVars {
			os.Unsetenv(env)
		}

		// Set minimum required values
		os.Setenv("JWT_SECRET", "this-is-a-very-long-secret-key-for-testing-purposes")

		config, err := Load()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Test default values
		if config.Env != Development {
			t.Errorf("Expected environment to be %v, got %v", Development, config.Env)
		}
		if config.Server.Port != 3000 {
			t.Errorf("Expected port to be 3000, got %d", config.Server.Port)
		}
		if config.Server.Host != "localhost" {
			t.Errorf("Expected host to be localhost, got %s", config.Server.Host)
		}
		if config.Database.Port != 5432 {
			t.Errorf("Expected DB port to be 5432, got %d", config.Database.Port)
		}
	})

	t.Run("Load with custom values", func(t *testing.T) {
		// Set custom environment variables
		os.Setenv("APP_ENV", "production")
		os.Setenv("PORT", "8080")
		os.Setenv("HOST", "0.0.0.0")
		os.Setenv("DB_HOST", "db.example.com")
		os.Setenv("DB_PORT", "5433")
		os.Setenv("DB_USER", "testuser")
		os.Setenv("DB_PASSWORD", "testpass")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("DB_SSLMODE", "require")
		os.Setenv("JWT_SECRET", "this-is-a-very-long-secret-key-for-testing-purposes")
		os.Setenv("JWT_ACCESS_EXPIRY", "30m")
		os.Setenv("JWT_REFRESH_EXPIRY", "720h")
		os.Setenv("CORS_ORIGINS", "https://example.com,https://api.example.com")
		os.Setenv("RATE_LIMIT_MAX", "200")
		os.Setenv("RATE_LIMIT_WINDOW", "5m")
		os.Setenv("SWAGGER_ENABLED", "false")
		os.Setenv("SWAGGER_HOST", "api.example.com")
		os.Setenv("SWAGGER_BASE_PATH", "/v2")

		config, err := Load()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Test custom values
		if config.Env != Production {
			t.Errorf("Expected environment to be %v, got %v", Production, config.Env)
		}
		if config.Server.Port != 8080 {
			t.Errorf("Expected port to be 8080, got %d", config.Server.Port)
		}
		if config.Server.Host != "0.0.0.0" {
			t.Errorf("Expected host to be 0.0.0.0, got %s", config.Server.Host)
		}
		if config.Database.Host != "db.example.com" {
			t.Errorf("Expected DB host to be db.example.com, got %s", config.Database.Host)
		}
		if config.Database.Port != 5433 {
			t.Errorf("Expected DB port to be 5433, got %d", config.Database.Port)
		}
		if config.JWT.AccessExpiry != 30*time.Minute {
			t.Errorf("Expected JWT access expiry to be 30m, got %v", config.JWT.AccessExpiry)
		}
		if len(config.CORS.Origins) != 2 {
			t.Errorf("Expected 2 CORS origins, got %d", len(config.CORS.Origins))
		}
		if config.Rate.Max != 200 {
			t.Errorf("Expected rate limit max to be 200, got %d", config.Rate.Max)
		}
		if !config.IsProduction() {
			t.Error("Expected IsProduction() to return true")
		}
	})

	t.Run("Load with invalid environment", func(t *testing.T) {
		os.Setenv("APP_ENV", "invalid")
		os.Setenv("JWT_SECRET", "this-is-a-very-long-secret-key-for-testing-purposes")

		_, err := Load()
		if err == nil {
			t.Error("Expected error for invalid environment")
		}
	})

	t.Run("Load with invalid port", func(t *testing.T) {
		os.Setenv("APP_ENV", "development")
		os.Setenv("PORT", "invalid")
		os.Setenv("JWT_SECRET", "this-is-a-very-long-secret-key-for-testing-purposes")

		_, err := Load()
		if err == nil {
			t.Error("Expected error for invalid port")
		}
	})

	t.Run("Load with invalid JWT expiry", func(t *testing.T) {
		os.Setenv("APP_ENV", "development")
		os.Setenv("JWT_SECRET", "this-is-a-very-long-secret-key-for-testing-purposes")
		os.Setenv("JWT_ACCESS_EXPIRY", "invalid")

		_, err := Load()
		if err == nil {
			t.Error("Expected error for invalid JWT access expiry")
		}
	})
}

func TestValidation(t *testing.T) {
	t.Run("Valid configuration", func(t *testing.T) {
		config := &Config{
			Env: Development,
			Server: ServerConfig{
				Port: 3000,
				Host: "localhost",
			},
			Database: DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				Name:     "testdb",
				SSLMode:  "disable",
			},
			JWT: JWTConfig{
				Secret:        "this-is-a-very-long-secret-key-for-testing-purposes",
				AccessExpiry:  15 * time.Minute,
				RefreshExpiry: 168 * time.Hour,
			},
			CORS: CORSConfig{
				Origins: []string{"http://localhost:3000"},
			},
			Rate: RateConfig{
				Max:    100,
				Window: time.Minute,
			},
		}

		if err := config.Validate(); err != nil {
			t.Errorf("Expected no validation error, got %v", err)
		}
	})

	t.Run("Invalid server port", func(t *testing.T) {
		config := &ServerConfig{
			Port: 0,
			Host: "localhost",
		}

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for invalid port")
		}
	})

	t.Run("Invalid server host", func(t *testing.T) {
		config := &ServerConfig{
			Port: 3000,
			Host: "",
		}

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for empty host")
		}
	})

	t.Run("Invalid database config", func(t *testing.T) {
		config := &DatabaseConfig{
			Host:     "",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Name:     "testdb",
			SSLMode:  "disable",
		}

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for empty host")
		}
	})

	t.Run("Invalid SSL mode", func(t *testing.T) {
		config := &DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Name:     "testdb",
			SSLMode:  "invalid",
		}

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for invalid SSL mode")
		}
	})

	t.Run("Invalid JWT secret", func(t *testing.T) {
		config := &JWTConfig{
			Secret:        "short",
			AccessExpiry:  15 * time.Minute,
			RefreshExpiry: 168 * time.Hour,
		}

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for short JWT secret")
		}
	})

	t.Run("Invalid JWT expiry order", func(t *testing.T) {
		config := &JWTConfig{
			Secret:        "this-is-a-very-long-secret-key-for-testing-purposes",
			AccessExpiry:  168 * time.Hour,
			RefreshExpiry: 15 * time.Minute,
		}

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for access expiry >= refresh expiry")
		}
	})

	t.Run("Empty CORS origins", func(t *testing.T) {
		config := &CORSConfig{
			Origins: []string{},
		}

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for empty CORS origins")
		}
	})

	t.Run("Invalid rate limit", func(t *testing.T) {
		config := &RateConfig{
			Max:    0,
			Window: time.Minute,
		}

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for zero rate limit max")
		}
	})
}

func TestEnvironmentMethods(t *testing.T) {
	tests := []struct {
		env          Environment
		isDev        bool
		isStaging    bool
		isProduction bool
	}{
		{Development, true, false, false},
		{Staging, false, true, false},
		{Production, false, false, true},
	}

	for _, test := range tests {
		config := &Config{Env: test.env}

		if config.IsDevelopment() != test.isDev {
			t.Errorf("For env %v, expected IsDevelopment() to be %v", test.env, test.isDev)
		}
		if config.IsStaging() != test.isStaging {
			t.Errorf("For env %v, expected IsStaging() to be %v", test.env, test.isStaging)
		}
		if config.IsProduction() != test.isProduction {
			t.Errorf("For env %v, expected IsProduction() to be %v", test.env, test.isProduction)
		}
	}
}

func TestDatabaseDSN(t *testing.T) {
	config := &DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		Name:     "testdb",
		SSLMode:  "disable",
	}

	expected := "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable"
	actual := config.GetDSN()

	if actual != expected {
		t.Errorf("Expected DSN %s, got %s", expected, actual)
	}
}

func TestServerAddress(t *testing.T) {
	config := &ServerConfig{
		Host: "localhost",
		Port: 3000,
	}

	expected := "localhost:3000"
	actual := config.GetServerAddress()

	if actual != expected {
		t.Errorf("Expected server address %s, got %s", expected, actual)
	}
}

func TestGetEnv(t *testing.T) {
	// Test with existing environment variable
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result := getEnv("TEST_VAR", "default")
	if result != "test_value" {
		t.Errorf("Expected test_value, got %s", result)
	}

	// Test with non-existing environment variable
	result = getEnv("NON_EXISTING_VAR", "default")
	if result != "default" {
		t.Errorf("Expected default, got %s", result)
	}
}
