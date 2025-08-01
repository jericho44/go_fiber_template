package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Environment represents the application environment
type Environment string

const (
	Development Environment = "development"
	Staging     Environment = "staging"
	Production  Environment = "production"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Rate     RateConfig
	Redis    RedisConfig
	Swagger  SwaggerConfig
	Env      Environment
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	User            string        `json:"user"`
	Password        string        `json:"-"` // Don't expose password in JSON
	Name            string        `json:"name"`
	SSLMode         string        `json:"ssl_mode"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	MaxOpenConns    int           `json:"max_open_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	Secret        string        `json:"-"` // Don't expose secret in JSON
	AccessExpiry  time.Duration `json:"access_expiry"`
	RefreshExpiry time.Duration `json:"refresh_expiry"`
}

// CORSConfig holds CORS-related configuration
type CORSConfig struct {
	Origins []string `json:"origins"`
}

// RateConfig holds rate limiting configuration
type RateConfig struct {
	Max    int           `json:"max"`
	Window time.Duration `json:"window"`
}

// RedisConfig holds Redis-related configuration
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"-"` // Don't expose password in JSON
	DB       int    `json:"db"`
	Enabled  bool   `json:"enabled"`
}

// SwaggerConfig holds Swagger documentation configuration
type SwaggerConfig struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	BasePath string `json:"base_path"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	config := &Config{}

	// Load environment
	env := getEnv("APP_ENV", "development")
	switch env {
	case "development":
		config.Env = Development
	case "staging":
		config.Env = Staging
	case "production":
		config.Env = Production
	default:
		return nil, fmt.Errorf("invalid environment: %s", env)
	}

	// Load server configuration
	port, err := strconv.Atoi(getEnv("PORT", "3000"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT value: %v", err)
	}
	config.Server = ServerConfig{
		Port: port,
		Host: getEnv("HOST", "localhost"),
	}

	// Load database configuration
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT value: %v", err)
	}

	maxIdleConns, err := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "10"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_IDLE_CONNS value: %v", err)
	}

	maxOpenConns, err := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "100"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_OPEN_CONNS value: %v", err)
	}

	connMaxLifetime, err := time.ParseDuration(getEnv("DB_CONN_MAX_LIFETIME", "1h"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME value: %v", err)
	}

	connMaxIdleTime, err := time.ParseDuration(getEnv("DB_CONN_MAX_IDLE_TIME", "30m"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_CONN_MAX_IDLE_TIME value: %v", err)
	}

	config.Database = DatabaseConfig{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            dbPort,
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", ""),
		Name:            getEnv("DB_NAME", "go_fiber_template"),
		SSLMode:         getEnv("DB_SSLMODE", "disable"),
		MaxIdleConns:    maxIdleConns,
		MaxOpenConns:    maxOpenConns,
		ConnMaxLifetime: connMaxLifetime,
		ConnMaxIdleTime: connMaxIdleTime,
	}

	// Load JWT configuration
	accessExpiry, err := time.ParseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_EXPIRY value: %v", err)
	}
	refreshExpiry, err := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXPIRY value: %v", err)
	}
	config.JWT = JWTConfig{
		Secret:        getEnv("JWT_SECRET", ""),
		AccessExpiry:  accessExpiry,
		RefreshExpiry: refreshExpiry,
	}

	// Load CORS configuration
	originsStr := getEnv("CORS_ORIGINS", "http://localhost:3000")
	origins := strings.Split(originsStr, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	config.CORS = CORSConfig{
		Origins: origins,
	}

	// Load rate limiting configuration
	rateMax, err := strconv.Atoi(getEnv("RATE_LIMIT_MAX", "100"))
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_MAX value: %v", err)
	}
	rateWindow, err := time.ParseDuration(getEnv("RATE_LIMIT_WINDOW", "1m"))
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_WINDOW value: %v", err)
	}
	config.Rate = RateConfig{
		Max:    rateMax,
		Window: rateWindow,
	}

	// Load Redis configuration
	redisPort, err := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_PORT value: %v", err)
	}
	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB value: %v", err)
	}
	redisEnabled, err := strconv.ParseBool(getEnv("REDIS_ENABLED", "false"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_ENABLED value: %v", err)
	}
	config.Redis = RedisConfig{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     redisPort,
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       redisDB,
		Enabled:  redisEnabled,
	}

	// Load Swagger configuration
	swaggerEnabled, err := strconv.ParseBool(getEnv("SWAGGER_ENABLED", "true"))
	if err != nil {
		return nil, fmt.Errorf("invalid SWAGGER_ENABLED value: %v", err)
	}
	config.Swagger = SwaggerConfig{
		Enabled:  swaggerEnabled,
		Host:     getEnv("SWAGGER_HOST", "localhost:3000"),
		BasePath: getEnv("SWAGGER_BASE_PATH", "/api/v1"),
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %v", err)
	}

	return config, nil
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server config: %v", err)
	}
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database config: %v", err)
	}
	if err := c.JWT.Validate(); err != nil {
		return fmt.Errorf("jwt config: %v", err)
	}
	if err := c.CORS.Validate(); err != nil {
		return fmt.Errorf("cors config: %v", err)
	}
	if err := c.Rate.Validate(); err != nil {
		return fmt.Errorf("rate config: %v", err)
	}
	if err := c.Redis.Validate(); err != nil {
		return fmt.Errorf("redis config: %v", err)
	}
	return nil
}

// Validate validates server configuration
func (s *ServerConfig) Validate() error {
	if s.Port <= 0 || s.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", s.Port)
	}
	if s.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	return nil
}

// Validate validates database configuration
func (d *DatabaseConfig) Validate() error {
	if d.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	if d.Port <= 0 || d.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", d.Port)
	}
	if d.User == "" {
		return fmt.Errorf("user cannot be empty")
	}
	if d.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	validSSLModes := []string{"disable", "require", "verify-ca", "verify-full"}
	isValid := false
	for _, mode := range validSSLModes {
		if d.SSLMode == mode {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("ssl_mode must be one of: %v, got %s", validSSLModes, d.SSLMode)
	}
	if d.MaxIdleConns < 0 {
		return fmt.Errorf("max_idle_conns must be non-negative, got %d", d.MaxIdleConns)
	}
	if d.MaxOpenConns <= 0 {
		return fmt.Errorf("max_open_conns must be positive, got %d", d.MaxOpenConns)
	}
	if d.MaxIdleConns > d.MaxOpenConns {
		return fmt.Errorf("max_idle_conns (%d) cannot be greater than max_open_conns (%d)", d.MaxIdleConns, d.MaxOpenConns)
	}
	if d.ConnMaxLifetime <= 0 {
		return fmt.Errorf("conn_max_lifetime must be positive")
	}
	if d.ConnMaxIdleTime <= 0 {
		return fmt.Errorf("conn_max_idle_time must be positive")
	}
	return nil
}

// Validate validates JWT configuration
func (j *JWTConfig) Validate() error {
	if j.Secret == "" {
		return fmt.Errorf("secret cannot be empty")
	}
	if len(j.Secret) < 32 {
		return fmt.Errorf("secret must be at least 32 characters long")
	}
	if j.AccessExpiry <= 0 {
		return fmt.Errorf("access expiry must be positive")
	}
	if j.RefreshExpiry <= 0 {
		return fmt.Errorf("refresh expiry must be positive")
	}
	if j.AccessExpiry >= j.RefreshExpiry {
		return fmt.Errorf("access expiry must be less than refresh expiry")
	}
	return nil
}

// Validate validates CORS configuration
func (c *CORSConfig) Validate() error {
	if len(c.Origins) == 0 {
		return fmt.Errorf("at least one origin must be specified")
	}
	for _, origin := range c.Origins {
		if origin == "" {
			return fmt.Errorf("origin cannot be empty")
		}
	}
	return nil
}

// Validate validates rate limiting configuration
func (r *RateConfig) Validate() error {
	if r.Max <= 0 {
		return fmt.Errorf("max must be positive, got %d", r.Max)
	}
	if r.Window <= 0 {
		return fmt.Errorf("window must be positive")
	}
	return nil
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Env == Development
}

// IsStaging returns true if the environment is staging
func (c *Config) IsStaging() bool {
	return c.Env == Staging
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Env == Production
}

// GetDSN returns the database connection string
func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

// GetAdminDSN returns the database connection string for admin operations (connects to postgres database)
func (d *DatabaseConfig) GetAdminDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.SSLMode)
}

// Validate validates Redis configuration
func (r *RedisConfig) Validate() error {
	if r.Enabled {
		if r.Host == "" {
			return fmt.Errorf("host cannot be empty when Redis is enabled")
		}
		if r.Port <= 0 || r.Port > 65535 {
			return fmt.Errorf("port must be between 1 and 65535, got %d", r.Port)
		}
		if r.DB < 0 {
			return fmt.Errorf("db must be non-negative, got %d", r.DB)
		}
	}
	return nil
}

// GetServerAddress returns the server address in host:port format
func (s *ServerConfig) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// GetRedisAddress returns the Redis address in host:port format
func (r *RedisConfig) GetRedisAddress() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}
