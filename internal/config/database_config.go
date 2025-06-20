package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Note: DatabaseConfig is now defined in config.go to avoid duplication

// GetDatabaseConfig returns database configuration from environment
func GetDatabaseConfig() *DatabaseConfig {
	config := &DatabaseConfig{
		Host:                  getEnv("DB_HOST", "mysql"),
		Port:                  getEnvAsInt("DB_PORT", 3306),
		Database:              getEnv("DB_DATABASE", "TS-Lager"),
		Username:              getEnv("DB_USERNAME", "tsweb"),
		Password:              getEnv("DB_PASSWORD", "web"),
		MaxOpenConns:          getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:          getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime:       getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		ConnMaxIdleTime:       getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		SlowQueryThreshold:    getEnvAsDuration("DB_SLOW_QUERY_THRESHOLD", 500*time.Millisecond),
		EnableQueryLogging:    getEnvAsBool("DB_ENABLE_QUERY_LOGGING", false),
		PrepareStmt:           getEnvAsBool("DB_PREPARE_STMT", true),
		DisableForeignKeyConstraintWhenMigrating: getEnvAsBool("DB_DISABLE_FK_WHEN_MIGRATING", true),
	}

	// Set log level based on environment
	if getEnvAsBool("DB_DEBUG", false) {
		config.LogLevel = logger.Info
	} else {
		config.LogLevel = logger.Warn
	}

	return config
}

// ConnectDatabase connects to the database with optimized settings
func ConnectDatabase(config *DatabaseConfig) (*gorm.DB, error) {
	// Build DSN with performance optimizations
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s&interpolateParams=true",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	// Configure GORM with performance settings
	gormConfig := &gorm.Config{
		PrepareStmt:                              config.PrepareStmt,
		DisableForeignKeyConstraintWhenMigrating: config.DisableForeignKeyConstraintWhenMigrating,
		Logger:                                   createLogger(config),
	}

	// Connect to database
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool parameters for optimal performance
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Database connected successfully with %d max connections", config.MaxOpenConns)
	
	return db, nil
}

// createLogger creates a configured logger for GORM
func createLogger(config *DatabaseConfig) logger.Interface {
	logConfig := logger.Config{
		SlowThreshold:             config.SlowQueryThreshold,
		LogLevel:                  config.LogLevel,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	}

	if config.EnableQueryLogging {
		return logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logConfig,
		)
	}

	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             config.SlowQueryThreshold,
			LogLevel:                  logger.Error, // Only log errors in production
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}

// ApplyPerformanceIndexes applies database indexes for performance
func ApplyPerformanceIndexes(db *gorm.DB) error {
	log.Println("Applying performance indexes...")

	indexes := []string{
		// Device indexes
		"CREATE INDEX IF NOT EXISTS idx_devices_productid ON devices(productID)",
		"CREATE INDEX IF NOT EXISTS idx_devices_status ON devices(status)",
		"CREATE INDEX IF NOT EXISTS idx_devices_search ON devices(deviceID, serialnumber)",
		
		// Job-Device relationship indexes
		"CREATE INDEX IF NOT EXISTS idx_jobdevices_deviceid ON jobdevices(deviceID)",
		"CREATE INDEX IF NOT EXISTS idx_jobdevices_jobid ON jobdevices(jobID)",
		"CREATE INDEX IF NOT EXISTS idx_jobdevices_composite ON jobdevices(deviceID, jobID)",
		
		// Jobs indexes
		"CREATE INDEX IF NOT EXISTS idx_jobs_customerid ON jobs(customerID)",
		"CREATE INDEX IF NOT EXISTS idx_jobs_statusid ON jobs(statusID)",
		"CREATE INDEX IF NOT EXISTS idx_jobs_dates ON jobs(startDate, endDate)",
		"CREATE INDEX IF NOT EXISTS idx_jobs_customer_status ON jobs(customerID, statusID)",
		
		// Invoice indexes
		"CREATE INDEX IF NOT EXISTS idx_invoices_customerid ON invoices(customer_id)",
		"CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status)",
		"CREATE INDEX IF NOT EXISTS idx_invoices_dates ON invoices(issue_date, due_date)",
		"CREATE INDEX IF NOT EXISTS idx_invoices_number ON invoices(invoice_number)",
		
		// Customer indexes
		"CREATE INDEX IF NOT EXISTS idx_customers_search_company ON customers(companyname)",
		"CREATE INDEX IF NOT EXISTS idx_customers_search_name ON customers(firstname, lastname)",
		"CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email)",
		
		// Product indexes
		"CREATE INDEX IF NOT EXISTS idx_products_categoryid ON products(categoryID)",
		"CREATE INDEX IF NOT EXISTS idx_products_status ON products(status)",
		
		// Composite indexes for complex queries
		"CREATE INDEX IF NOT EXISTS idx_devices_product_status ON devices(productID, status)",
		"CREATE INDEX IF NOT EXISTS idx_jobs_status_dates ON jobs(statusID, startDate, endDate)",
		
		// Invoice line items
		"CREATE INDEX IF NOT EXISTS idx_invoice_line_items_invoice ON invoice_line_items(invoice_id)",
		"CREATE INDEX IF NOT EXISTS idx_invoice_line_items_device ON invoice_line_items(device_id)",
		
		// Session management
		"CREATE INDEX IF NOT EXISTS idx_sessions_userid ON sessions(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at)",
		
		// Email and invoice templates
		"CREATE INDEX IF NOT EXISTS idx_email_templates_type ON email_templates(template_type)",
		"CREATE INDEX IF NOT EXISTS idx_email_templates_default ON email_templates(template_type, is_default)",
		"CREATE INDEX IF NOT EXISTS idx_invoice_templates_default ON invoice_templates(is_default)",
		"CREATE INDEX IF NOT EXISTS idx_invoice_templates_active ON invoice_templates(is_active)",
	}

	successCount := 0
	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
		} else {
			successCount++
		}
	}

	log.Printf("Successfully applied %d/%d performance indexes", successCount, len(indexes))
	return nil
}

// GetDatabaseStats returns database connection statistics
func GetDatabaseStats(db *gorm.DB) (map[string]interface{}, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	
	return map[string]interface{}{
		"max_open_connections":     stats.MaxOpenConnections,
		"open_connections":         stats.OpenConnections,
		"in_use":                   stats.InUse,
		"idle":                     stats.Idle,
		"wait_count":               stats.WaitCount,
		"wait_duration":            stats.WaitDuration.String(),
		"max_idle_closed":          stats.MaxIdleClosed,
		"max_idle_time_closed":     stats.MaxIdleTimeClosed,
		"max_lifetime_closed":      stats.MaxLifetimeClosed,
	}, nil
}

// OptimizeDatabaseSettings applies MySQL-specific optimizations
func OptimizeDatabaseSettings(db *gorm.DB) error {
	log.Println("Applying MySQL optimization settings...")

	optimizations := []string{
		// Query cache optimization
		"SET SESSION query_cache_type = ON",
		"SET SESSION query_cache_size = 67108864", // 64MB
		
		// InnoDB optimizations
		"SET SESSION innodb_buffer_pool_size = 134217728", // 128MB
		"SET SESSION sort_buffer_size = 2097152",           // 2MB
		"SET SESSION read_buffer_size = 131072",            // 128KB
		"SET SESSION join_buffer_size = 262144",            // 256KB
		
		// Timeout settings
		"SET SESSION wait_timeout = 28800",                 // 8 hours
		"SET SESSION interactive_timeout = 28800",          // 8 hours
		
		// Character set
		"SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci",
	}

	for _, sql := range optimizations {
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("Warning: Failed to apply optimization: %s - %v", sql, err)
		}
	}

	log.Println("Database optimization settings applied")
	return nil
}

// Helper functions for environment variables

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}