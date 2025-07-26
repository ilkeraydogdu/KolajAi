package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Config represents the application configuration
type Config struct {
	AppName      string             `json:"app_name"`
	Environment  string             `json:"environment"`
	Server       ServerConfig       `json:"server"`
	Database     DatabaseConfig     `json:"database"`
	Template     TemplateConfig     `json:"template"`
	Email        EmailConfig        `json:"email"`
	Notification NotificationConfig `json:"notification"`
	Routes       RoutesConfig       `json:"routes"`
	Auth         AuthConfig         `json:"auth"`
	Logger       LoggerConfig       `json:"logger"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
	IdleTimeout  int    `json:"idle_timeout"`
	Debug        bool   `json:"debug"`
	Prefork      bool   `json:"prefork"`
	BaseURL      string `json:"base_url"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret            string        `json:"jwt_secret"`
	JWTExpiration        time.Duration `json:"jwt_expiration"`
	RefreshTokenDuration time.Duration `json:"refresh_token_duration"`
	PasswordMinLength    int           `json:"password_min_length"`
	RequireUppercase     bool          `json:"require_uppercase"`
	RequireSpecialChar   bool          `json:"require_special_char"`
	RequireNumber        bool          `json:"require_number"`
	MaxLoginAttempts     int           `json:"max_login_attempts"`
	LockoutDuration      time.Duration `json:"lockout_duration"`
}

// EmailConfig holds email configuration
type EmailConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	FromName string `json:"from_name"`
	TLS      bool   `json:"tls"`
}

// TemplateConfig holds template configuration
type TemplateConfig struct {
	Dir           string `json:"dir"`
	PartialsDir   string `json:"partials_dir"`
	ComponentsDir string `json:"components_dir"`
	Extension     string `json:"extension"`
	BaseLayout    string `json:"base_layout"`
	Cache         bool   `json:"cache"`
}

// NotificationConfig holds notification configuration
type NotificationConfig struct {
	DefaultType string                      `json:"default_type"`
	DefaultTTL  time.Duration               `json:"default_ttl"`
	Timeout     int                         `json:"timeout"`
	Position    string                      `json:"position"`
	ShowClose   bool                        `json:"show_close"`
	AutoClose   bool                        `json:"auto_close"`
	Types       map[string]NotificationType `json:"types"`
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level      string `json:"level"`
	File       string `json:"file"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
	Compress   bool   `json:"compress"`
}

// RoutesConfig holds route configuration
type RoutesConfig struct {
	API struct {
		Prefix     string   `json:"prefix"`
		Version    string   `json:"version"`
		Middleware []string `json:"middleware"`
	} `json:"api"`
	Web struct {
		Middleware []string `json:"middleware"`
	} `json:"web"`
	Assets struct {
		Path   string `json:"path"`
		Prefix string `json:"prefix"`
	} `json:"assets"`
}

// TableConfig represents a database table configuration
type TableConfig struct {
	Name    string `json:"name"`
	Columns []struct {
		Name       string `json:"name"`
		Type       string `json:"type"`
		PrimaryKey bool   `json:"primary_key,omitempty"`
		Nullable   bool   `json:"nullable,omitempty"`
		Default    string `json:"default,omitempty"`
	} `json:"columns"`
	Indexes []struct {
		Name    string   `json:"name"`
		Columns []string `json:"columns"`
		Unique  bool     `json:"unique,omitempty"`
	} `json:"indexes,omitempty"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Driver          string        `json:"driver"`
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	Username        string        `json:"username"`
	Password        string        `json:"password"`
	Database        string        `json:"database"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	Charset         string        `json:"charset"`
	ParseTime       bool          `json:"parse_time"`
}

// GetConfig returns the application configuration
func GetConfig() (*Config, error) {
	// First, try to load from environment variable
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		// If not set, use default path
		configPath = "config.json"
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config if it doesn't exist
		config := createDefaultConfig()
		return config, nil
	}

	// Read config file
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse config
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// createDefaultConfig creates a default configuration
func createDefaultConfig() *Config {
	return &Config{
		AppName:     "KolajAI",
		Environment: "development",
		Server: ServerConfig{
			Host:         "localhost",
			Port:         8080,
			ReadTimeout:  60,
			WriteTimeout: 60,
			IdleTimeout:  120,
			Debug:        true,
			Prefork:      false,
			BaseURL:      "http://localhost:8080",
		},
		Database: DatabaseConfig{
			Driver:          "mysql",
			Host:            "localhost",
			Port:            3306,
			Username:        "root",
			Password:        "",
			Database:        "kolajai",
			MaxOpenConns:    25,
			MaxIdleConns:    25,
			ConnMaxLifetime: 5 * time.Minute,
			Charset:         "utf8mb4",
			ParseTime:       true,
		},
		Template: TemplateConfig{
			Dir:           "web/templates",
			PartialsDir:   "web/templates/partials",
			ComponentsDir: "web/templates/components",
			Extension:     ".gohtml",
			BaseLayout:    "layout",
			Cache:         true,
		},
		Email: EmailConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user@example.com",
			Password: "password",
			From:     "noreply@example.com",
			FromName: "KolajAI",
			TLS:      true,
		},
		Notification: NotificationConfig{
			DefaultType: "info",
			DefaultTTL:  5 * time.Second,
			Timeout:     5000,
			Position:    "top-right",
			ShowClose:   true,
			AutoClose:   true,
			Types:       GetDefaultNotificationTypes(),
		},
		Routes: RoutesConfig{
			API: struct {
				Prefix     string   `json:"prefix"`
				Version    string   `json:"version"`
				Middleware []string `json:"middleware"`
			}{
				Prefix:  "/api",
				Version: "v1",
				Middleware: []string{
					"cors",
					"auth",
					"logger",
				},
			},
			Web: struct {
				Middleware []string `json:"middleware"`
			}{
				Middleware: []string{
					"session",
					"csrf",
					"logger",
				},
			},
			Assets: struct {
				Path   string `json:"path"`
				Prefix string `json:"prefix"`
			}{
				Path:   "web/static",
				Prefix: "/static",
			},
		},
		Auth: AuthConfig{
			JWTSecret:            "secret",
			JWTExpiration:        24 * time.Hour,
			RefreshTokenDuration: 7 * 24 * time.Hour,
			PasswordMinLength:    8,
			RequireUppercase:     true,
			RequireSpecialChar:   true,
			RequireNumber:        true,
			MaxLoginAttempts:     5,
			LockoutDuration:      15 * time.Minute,
		},
		Logger: LoggerConfig{
			Level:      "debug",
			File:       "logs/app.log",
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
		},
	}
}

// LoadConfig loads the application configuration from a file
func LoadConfig(path string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	// Read config file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse config
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(config *Config, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetServerConfig returns the server configuration
func GetServerConfig() (ServerConfig, bool) {
	config, err := GetConfig()
	if err != nil {
		log.Printf("Failed to get config: %v", err)
		return ServerConfig{}, false
	}
	return config.Server, true
}

// GetDatabaseConfig returns the database configuration
func GetDatabaseConfig() (DatabaseConfig, bool) {
	config, err := GetConfig()
	if err != nil {
		log.Printf("Failed to get config: %v", err)
		return DatabaseConfig{}, false
	}
	return config.Database, true
}

// GetTemplateConfig returns the template configuration
func GetTemplateConfig() (TemplateConfig, bool) {
	config, err := GetConfig()
	if err != nil {
		log.Printf("Failed to get config: %v", err)
		return TemplateConfig{}, false
	}
	return config.Template, true
}

// GetEmailConfig returns the email configuration
func GetEmailConfig() (EmailConfig, bool) {
	config, err := GetConfig()
	if err != nil {
		log.Printf("Failed to get config: %v", err)
		return EmailConfig{}, false
	}
	return config.Email, true
}

// GetNotificationConfig returns the notification configuration
func GetNotificationConfig() (NotificationConfig, bool) {
	config, err := GetConfig()
	if err != nil {
		log.Printf("Failed to get config: %v", err)
		return NotificationConfig{}, false
	}
	return config.Notification, true
}

// GetRoutesConfig returns the routes configuration
func GetRoutesConfig() (RoutesConfig, bool) {
	config, err := GetConfig()
	if err != nil {
		log.Printf("Failed to get config: %v", err)
		return RoutesConfig{}, false
	}
	return config.Routes, true
}

// GetAuthConfig returns the authentication configuration
func GetAuthConfig() (AuthConfig, bool) {
	config, err := GetConfig()
	if err != nil {
		log.Printf("Failed to get config: %v", err)
		return AuthConfig{}, false
	}
	return config.Auth, true
}

// GetLoggerConfig returns the logger configuration
func GetLoggerConfig() (LoggerConfig, bool) {
	config, err := GetConfig()
	if err != nil {
		log.Printf("Failed to get config: %v", err)
		return LoggerConfig{}, false
	}
	return config.Logger, true
}
