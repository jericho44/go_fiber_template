package middleware

import (
	"context"
	"log"
	"os"
	"time"

	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// DBSecurityMiddleware provides database security monitoring
type DBSecurityMiddleware struct {
	queryLogger      *utils.DBQueryLogger
	queryBuilder     *utils.QueryBuilder
	securityMonitor  *utils.QuerySecurityMonitor
	metricsCollector *utils.QueryMetricsCollector
	securityTester   *utils.DBSecurityTester
	config           *DBSecurityConfig
}

// DBSecurityConfig holds configuration for database security middleware
type DBSecurityConfig struct {
	EnableQueryLogging    bool
	EnableSecurityMonitor bool
	EnableMetrics         bool
	LogLevel              string
	SlowQueryThreshold    time.Duration
	MaxQueryLength        int
	EnableRealTimeAlerts  bool
}

// DefaultDBSecurityConfig returns default configuration
func DefaultDBSecurityConfig() *DBSecurityConfig {
	return &DBSecurityConfig{
		EnableQueryLogging:    true,
		EnableSecurityMonitor: true,
		EnableMetrics:         true,
		LogLevel:              "INFO",
		SlowQueryThreshold:    1 * time.Second,
		MaxQueryLength:        10000,
		EnableRealTimeAlerts:  true,
	}
}

// NewDBSecurityMiddleware creates a new database security middleware
func NewDBSecurityMiddleware(db *gorm.DB, config *DBSecurityConfig) *DBSecurityMiddleware {
	if config == nil {
		config = DefaultDBSecurityConfig()
	}

	// Create logger
	logger := log.New(os.Stdout, "[DB-SECURITY] ", log.LstdFlags)

	// Create query logger
	queryLoggerConfig := &utils.DBQueryLoggerConfig{
		EnableQueryLogging:    config.EnableQueryLogging,
		EnableSlowQueryLog:    true,
		EnableSuspiciousLog:   true,
		SlowQueryThreshold:    config.SlowQueryThreshold,
		LogLevel:              config.LogLevel,
		IncludeStackTrace:     false,
		MaxQueryLength:        config.MaxQueryLength,
		SensitiveDataPatterns: []string{"password", "token", "secret", "key", "hash"},
	}
	queryLogger := utils.NewDBQueryLogger(logger, queryLoggerConfig)

	// Create query builder
	queryBuilder := utils.NewQueryBuilder(db, queryLogger)

	// Create metrics collector
	metricsCollector := utils.NewQueryMetricsCollector()

	// Create security monitor
	securityMonitor := utils.NewQuerySecurityMonitor(queryLogger, metricsCollector)

	// Create security tester
	securityTester := utils.NewDBSecurityTester(db, queryBuilder)

	return &DBSecurityMiddleware{
		queryLogger:      queryLogger,
		queryBuilder:     queryBuilder,
		securityMonitor:  securityMonitor,
		metricsCollector: metricsCollector,
		securityTester:   securityTester,
		config:           config,
	}
}

// Handler returns the middleware handler function
func (m *DBSecurityMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Add database security context
		ctx := m.enrichContext(c)
		c.SetUserContext(ctx)

		// Continue with the request
		return c.Next()
	}
}

// enrichContext enriches the context with security information
func (m *DBSecurityMiddleware) enrichContext(c *fiber.Ctx) context.Context {
	ctx := c.UserContext()

	// Add correlation ID if not present
	correlationID := c.Get("X-Correlation-ID")
	if correlationID == "" {
		correlationID = generateCorrelationID()
		c.Set("X-Correlation-ID", correlationID)
	}
	ctx = context.WithValue(ctx, "correlation_id", correlationID)

	// Add user information if available
	if userID := c.Locals("user_id"); userID != nil {
		ctx = context.WithValue(ctx, "user_id", userID)
	}

	// Add request information
	ctx = context.WithValue(ctx, "ip_address", c.IP())
	ctx = context.WithValue(ctx, "user_agent", c.Get("User-Agent"))
	ctx = context.WithValue(ctx, "request_path", c.Path())
	ctx = context.WithValue(ctx, "request_method", c.Method())

	return ctx
}

// GetQueryBuilder returns the secure query builder
func (m *DBSecurityMiddleware) GetQueryBuilder() *utils.QueryBuilder {
	return m.queryBuilder
}

// GetSecurityMonitor returns the security monitor
func (m *DBSecurityMiddleware) GetSecurityMonitor() *utils.QuerySecurityMonitor {
	return m.securityMonitor
}

// GetMetricsCollector returns the metrics collector
func (m *DBSecurityMiddleware) GetMetricsCollector() *utils.QueryMetricsCollector {
	return m.metricsCollector
}

// GetSecurityTester returns the security tester
func (m *DBSecurityMiddleware) GetSecurityTester() *utils.DBSecurityTester {
	return m.securityTester
}

// RunSecurityTests runs database security tests
func (m *DBSecurityMiddleware) RunSecurityTests(ctx context.Context) *utils.SecurityTestSuite {
	return m.securityTester.RunAllSecurityTests(ctx)
}

// GetSecurityReport generates a security report
func (m *DBSecurityMiddleware) GetSecurityReport() map[string]interface{} {
	return m.securityMonitor.GetSecurityReport()
}

// generateCorrelationID generates a unique correlation ID
func generateCorrelationID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// DBSecurityHandler provides HTTP endpoints for database security
type DBSecurityHandler struct {
	middleware *DBSecurityMiddleware
}

// NewDBSecurityHandler creates a new database security handler
func NewDBSecurityHandler(middleware *DBSecurityMiddleware) *DBSecurityHandler {
	return &DBSecurityHandler{
		middleware: middleware,
	}
}

// GetSecurityReport handles GET /api/admin/db-security/report
func (h *DBSecurityHandler) GetSecurityReport(c *fiber.Ctx) error {
	report := h.middleware.GetSecurityReport()

	return c.JSON(fiber.Map{
		"success": true,
		"data":    report,
	})
}

// RunSecurityTests handles POST /api/admin/db-security/test
func (h *DBSecurityHandler) RunSecurityTests(c *fiber.Ctx) error {
	ctx := c.UserContext()
	testSuite := h.middleware.RunSecurityTests(ctx)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    testSuite,
	})
}

// GetMetrics handles GET /api/admin/db-security/metrics
func (h *DBSecurityHandler) GetMetrics(c *fiber.Ctx) error {
	metrics := h.middleware.GetMetricsCollector().GetMetrics()

	return c.JSON(fiber.Map{
		"success": true,
		"data":    metrics,
	})
}

// ResetMetrics handles POST /api/admin/db-security/metrics/reset
func (h *DBSecurityHandler) ResetMetrics(c *fiber.Ctx) error {
	h.middleware.GetMetricsCollector().Reset()

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Metrics reset successfully",
	})
}

// RegisterRoutes registers database security routes
func (h *DBSecurityHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/admin/db-security")

	api.Get("/report", h.GetSecurityReport)
	api.Post("/test", h.RunSecurityTests)
	api.Get("/metrics", h.GetMetrics)
	api.Post("/metrics/reset", h.ResetMetrics)
}

// DBSecurityGORMPlugin provides GORM plugin for database security
type DBSecurityGORMPlugin struct {
	monitor *utils.QuerySecurityMonitor
}

// NewDBSecurityGORMPlugin creates a new GORM plugin for database security
func NewDBSecurityGORMPlugin(monitor *utils.QuerySecurityMonitor) *DBSecurityGORMPlugin {
	return &DBSecurityGORMPlugin{
		monitor: monitor,
	}
}

// Name returns the plugin name
func (p *DBSecurityGORMPlugin) Name() string {
	return "db_security"
}

// Initialize initializes the plugin
func (p *DBSecurityGORMPlugin) Initialize(db *gorm.DB) error {
	// Register callbacks for query monitoring
	db.Callback().Query().Before("gorm:query").Register("db_security:before_query", p.beforeQuery)
	db.Callback().Query().After("gorm:query").Register("db_security:after_query", p.afterQuery)

	db.Callback().Create().Before("gorm:create").Register("db_security:before_create", p.beforeQuery)
	db.Callback().Create().After("gorm:create").Register("db_security:after_create", p.afterQuery)

	db.Callback().Update().Before("gorm:update").Register("db_security:before_update", p.beforeQuery)
	db.Callback().Update().After("gorm:update").Register("db_security:after_update", p.afterQuery)

	db.Callback().Delete().Before("gorm:delete").Register("db_security:before_delete", p.beforeQuery)
	db.Callback().Delete().After("gorm:delete").Register("db_security:after_delete", p.afterQuery)

	return nil
}

// beforeQuery is called before query execution
func (p *DBSecurityGORMPlugin) beforeQuery(db *gorm.DB) {
	db.Set("db_security:start_time", time.Now())
}

// afterQuery is called after query execution
func (p *DBSecurityGORMPlugin) afterQuery(db *gorm.DB) {
	startTime, exists := db.Get("db_security:start_time")
	if !exists {
		return
	}

	start, ok := startTime.(time.Time)
	if !ok {
		return
	}

	duration := time.Since(start)

	// Monitor the query
	if p.monitor != nil {
		ctx := db.Statement.Context
		if ctx == nil {
			ctx = context.Background()
		}

		query := db.Statement.SQL.String()
		args := db.Statement.Vars
		err := db.Error

		p.monitor.MonitorQuery(ctx, query, args, duration, err)
	}
}
