package config

import (
	"encoding/json"
	"os"
	"time"
	
	"gorm.io/gorm/logger"
)

type Config struct {
	Database DatabaseConfig `json:"database"`
	Server   ServerConfig   `json:"server"`
	UI       UIConfig       `json:"ui"`
	Email    EmailConfig    `json:"email"`
	Invoice  InvoiceConfig  `json:"invoice"`
	PDF      PDFConfig      `json:"pdf"`
	Security SecurityConfig `json:"security"`
	Logging  LoggingConfig  `json:"logging"`
	Backup   BackupConfig   `json:"backup"`
}

type DatabaseConfig struct {
	Host                  string        `json:"host"`
	Port                  int           `json:"port"`
	Database              string        `json:"database"`
	Username              string        `json:"username"`
	Password              string        `json:"password"`
	PoolSize              int           `json:"pool_size"` // Kept for backwards compatibility
	MaxOpenConns          int           `json:"max_open_conns"`
	MaxIdleConns          int           `json:"max_idle_conns"`
	ConnMaxLifetime       time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime       time.Duration `json:"conn_max_idle_time"`
	SlowQueryThreshold    time.Duration `json:"slow_query_threshold"`
	EnableQueryLogging    bool          `json:"enable_query_logging"`
	LogLevel              logger.LogLevel `json:"-"` // Not serializable
	PrepareStmt           bool          `json:"prepare_stmt"`
	DisableForeignKeyConstraintWhenMigrating bool `json:"disable_fk_when_migrating"`
}

type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

type UIConfig struct {
	ThemeDark      string            `json:"theme_dark"`
	ThemeLight     string            `json:"theme_light"`
	CurrentTheme   string            `json:"current_theme"`
	Colors         map[string]string `json:"colors"`
	AutoSave       bool              `json:"auto_save"`
	AutoSaveInterval int             `json:"auto_save_interval"`
	CacheTimeout   int               `json:"cache_timeout"`
	WindowWidth    int               `json:"window_width"`
	WindowHeight   int               `json:"window_height"`
}

type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	FromEmail    string `json:"from_email"`
	FromName     string `json:"from_name"`
	UseTLS       bool   `json:"use_tls"`
}

type InvoiceConfig struct {
	DefaultTaxRate          float64 `json:"default_tax_rate"`
	DefaultPaymentTerms     int     `json:"default_payment_terms"`
	AutoCalculateRentalDays bool    `json:"auto_calculate_rental_days"`
	ShowLogoOnInvoice       bool    `json:"show_logo_on_invoice"`
	InvoiceNumberPrefix     string  `json:"invoice_number_prefix"`
	InvoiceNumberFormat     string  `json:"invoice_number_format"`
	CurrencySymbol          string  `json:"currency_symbol"`
	CurrencyCode            string  `json:"currency_code"`
	DateFormat              string  `json:"date_format"`
}

type PDFConfig struct {
	Generator string            `json:"generator"`
	PaperSize string            `json:"paper_size"`
	Margins   map[string]string `json:"margins"`
}

type SecurityConfig struct {
	SessionTimeout    int    `json:"session_timeout"`
	PasswordMinLength int    `json:"password_min_length"`
	MaxLoginAttempts  int    `json:"max_login_attempts"`
	LockoutDuration   int    `json:"lockout_duration"`
	EncryptionKey     string `json:"encryption_key"`
}

type LoggingConfig struct {
	Level      string `json:"level"`
	File       string `json:"file"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
}

type BackupConfig struct {
	Enabled       bool   `json:"enabled"`
	Interval      int    `json:"interval"`
	RetentionDays int    `json:"retention_days"`
	Path          string `json:"path"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return getDefaultConfig(), nil // Return default config if file doesn't exist
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(c)
}

func getDefaultConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:                  "localhost",
			Port:                  3306,
			Database:              "jobscanner",
			Username:              "root",
			Password:              "",
			PoolSize:              5,
			MaxOpenConns:          25,
			MaxIdleConns:          5,
			ConnMaxLifetime:       5 * time.Minute,
			ConnMaxIdleTime:       5 * time.Minute,
			SlowQueryThreshold:    500 * time.Millisecond,
			EnableQueryLogging:    false,
			LogLevel:              logger.Warn,
			PrepareStmt:           true,
			DisableForeignKeyConstraintWhenMigrating: true,
		},
		Server: ServerConfig{
			Port: 8080,
			Host: "localhost",
		},
		UI: UIConfig{
			ThemeDark:        "darkly",
			ThemeLight:       "flatly",
			CurrentTheme:     "dark",
			AutoSave:         true,
			AutoSaveInterval: 300,
			CacheTimeout:     300,
			WindowWidth:      1400,
			WindowHeight:     800,
			Colors: map[string]string{
				"primary":     "#007bff",
				"background":  "#ffffff",
				"text":        "#000000",
				"selection":   "#e9ecef",
				"success":     "#28a745",
				"error":       "#dc3545",
				"warning":     "#ffc107",
				"dark_bg":     "#2b2b2b",
				"dark_text":   "#ffffff",
			},
		},
		Email: EmailConfig{
			SMTPHost:     "localhost",
			SMTPPort:     587,
			SMTPUsername: "",
			SMTPPassword: "",
			FromEmail:    "noreply@rentalcore.com",
			FromName:     "RentalCore",
			UseTLS:       true,
		},
		Invoice: InvoiceConfig{
			DefaultTaxRate:          19.0,
			DefaultPaymentTerms:     30,
			AutoCalculateRentalDays: true,
			ShowLogoOnInvoice:       true,
			InvoiceNumberPrefix:     "INV-",
			InvoiceNumberFormat:     "{prefix}{year}{month}{sequence:4}",
			CurrencySymbol:          "€",
			CurrencyCode:            "EUR",
			DateFormat:              "DD.MM.YYYY",
		},
		PDF: PDFConfig{
			Generator: "auto",
			PaperSize: "A4",
			Margins: map[string]string{
				"top":    "1cm",
				"bottom": "1cm",
				"left":   "1cm",
				"right":  "1cm",
			},
		},
		Security: SecurityConfig{
			SessionTimeout:    3600,
			PasswordMinLength: 8,
			MaxLoginAttempts:  5,
			LockoutDuration:   900,
			EncryptionKey:     "TS-Lager-Default-Encryption-Key-Change-In-Production",
		},
		Logging: LoggingConfig{
			Level:      "info",
			File:       "logs/app.log",
			MaxSize:    100,
			MaxBackups: 5,
			MaxAge:     30,
		},
		Backup: BackupConfig{
			Enabled:       true,
			Interval:      86400,
			RetentionDays: 30,
			Path:          "backups/",
		},
	}
}