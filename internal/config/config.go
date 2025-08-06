package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Global config instance
var cfg *Config

// Config holds the application configuration
type Config struct {
	Environment string         `yaml:"environment"`
	Server      ServerConfig   `yaml:"server"`
	Database    DatabaseConfig `yaml:"database"`
	Security    SecurityConfig `yaml:"security"`
	Cache       CacheConfig    `yaml:"cache"`
	Email       EmailConfig    `yaml:"email"`
	SEO         SEOConfig      `yaml:"seo"`
	Logging     LoggingConfig  `yaml:"logging"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port           int    `yaml:"port"`
	Host           string `yaml:"host"`
	Domain         string `yaml:"domain"`
	ReadTimeout    int    `yaml:"read_timeout"`
	WriteTimeout   int    `yaml:"write_timeout"`
	IdleTimeout    int    `yaml:"idle_timeout"`
	MaxHeaderBytes int    `yaml:"max_header_bytes"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver          string `yaml:"driver"`
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Name            string `yaml:"name"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	EncryptionKey    string   `yaml:"encryption_key"`
	JWTSecret        string   `yaml:"jwt_secret"`
	CSRFSecret       string   `yaml:"csrf_secret"`
	SessionSecret    string   `yaml:"session_secret"`
	TwoFactorEnabled bool     `yaml:"two_factor_enabled"`
	RateLimitEnabled bool     `yaml:"rate_limit_enabled"`
	CORSEnabled      bool     `yaml:"cors_enabled"`
	AllowedOrigins   []string `yaml:"allowed_origins"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Driver         string        `yaml:"driver"`
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	Password       string        `yaml:"password"`
	DB             int           `yaml:"db"`
	DefaultTTL     time.Duration `yaml:"default_ttl"`
	MaxMemoryUsage int64         `yaml:"max_memory_usage"`
}

// EmailConfig holds email configuration
type EmailConfig struct {
	SMTPHost     string `yaml:"smtp_host"`
	SMTPPort     int    `yaml:"smtp_port"`
	SMTPUser     string `yaml:"smtp_user"`
	SMTPPassword string `yaml:"smtp_password"`
	FromEmail    string `yaml:"from_email"`
	FromName     string `yaml:"from_name"`
	UseSSL       bool   `yaml:"use_ssl"`
	UseTLS       bool   `yaml:"use_tls"`
}

// SEOConfig holds SEO configuration
type SEOConfig struct {
	SiteName         string   `yaml:"site_name"`
	SiteDescription  string   `yaml:"site_description"`
	SiteKeywords     []string `yaml:"site_keywords"`
	DefaultLanguage  string   `yaml:"default_language"`
	Languages        []string `yaml:"languages"`
	GoogleAnalytics  string   `yaml:"google_analytics"`
	GoogleTagManager string   `yaml:"google_tag_manager"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// NotificationConfig represents notification configuration
type NotificationConfig struct {
	EnableEmail bool                        `yaml:"enable_email"`
	EnableSMS   bool                        `yaml:"enable_sms"`
	EnablePush  bool                        `yaml:"enable_push"`
	DefaultTTL  int                         `yaml:"default_ttl"`
	Templates   map[string]string           `yaml:"templates"`
	Types       map[string]NotificationType `yaml:"types"`
}

// NotificationType represents a notification type configuration
type NotificationType struct {
	Name     string `yaml:"name"`
	Template string `yaml:"template"`
	Subject  string `yaml:"subject"`
	Enabled  bool   `yaml:"enabled"`
	Priority int    `yaml:"priority"`
	Icon     string `yaml:"icon"`
	Color    string `yaml:"color"`
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	// Load .env file if exists
	loadEnvFile(".env")

	config := &Config{}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return GetDefaultConfig(), nil
	}

	// Read config file
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Replace environment variables
	data = []byte(expandEnvVars(string(data)))

	// Parse YAML
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	// Override with environment variables
	overrideWithEnv(config)

	// Set global config
	cfg = config

	return config, nil
}

// GetDefaultConfig returns default configuration
func GetDefaultConfig() *Config {
	config := &Config{
		Environment: getEnv("APP_ENV", "development"),
		Server: ServerConfig{
			Port:           getEnvAsInt("SERVER_PORT", 8081),
			Host:           getEnv("SERVER_HOST", "0.0.0.0"),
			Domain:         getEnv("SERVER_DOMAIN", "localhost"),
			ReadTimeout:    getEnvAsInt("SERVER_READ_TIMEOUT", 10),
			WriteTimeout:   getEnvAsInt("SERVER_WRITE_TIMEOUT", 10),
			IdleTimeout:    getEnvAsInt("SERVER_IDLE_TIMEOUT", 120),
			MaxHeaderBytes: getEnvAsInt("SERVER_MAX_HEADER_BYTES", 1048576),
		},
		Database: DatabaseConfig{
			Driver:          getEnv("DB_DRIVER", "sqlite3"),
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 3306),
			Name:            getEnv("DB_NAME", "kolajAi.db"),
			User:            getEnv("DB_USER", ""),
			Password:        getEnv("DB_PASSWORD", ""),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 5),
		},
		Security: SecurityConfig{
			EncryptionKey:    getEnv("ENCRYPTION_KEY", "supersecretkey32byteslongforencryption"),
			JWTSecret:        getEnv("JWT_SECRET", "jwtsecretkey"),
			CSRFSecret:       getEnv("CSRF_SECRET", "csrfsecretkey"),
			SessionSecret:    getEnv("SESSION_SECRET", "sessionsecretkey"),
			TwoFactorEnabled: getEnvAsBool("TWO_FACTOR_ENABLED", false),
			RateLimitEnabled: getEnvAsBool("RATE_LIMIT_ENABLED", true),
			CORSEnabled:      getEnvAsBool("CORS_ENABLED", true),
			AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8081"},
		},
		Cache: CacheConfig{
			Driver:         getEnv("CACHE_DRIVER", "memory"),
			Host:           getEnv("CACHE_HOST", "localhost"),
			Port:           getEnvAsInt("CACHE_PORT", 6379),
			Password:       getEnv("CACHE_PASSWORD", ""),
			DB:             getEnvAsInt("CACHE_DB", 0),
			DefaultTTL:     time.Duration(getEnvAsInt("CACHE_DEFAULT_TTL", 1800)) * time.Second,
			MaxMemoryUsage: int64(getEnvAsInt("CACHE_MAX_MEMORY", 1073741824)), // 1GB
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
			SMTPUser:     getEnv("SMTP_USER", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromEmail:    getEnv("FROM_EMAIL", "noreply@kolajAi.com"),
			FromName:     getEnv("FROM_NAME", "KolajAI"),
			UseSSL:       getEnvAsBool("SMTP_USE_SSL", false),
			UseTLS:       getEnvAsBool("SMTP_USE_TLS", true),
		},
		SEO: SEOConfig{
			SiteName:         getEnv("SITE_NAME", "KolajAI Enterprise Marketplace"),
			SiteDescription:  getEnv("SITE_DESCRIPTION", "Advanced AI-powered e-commerce platform"),
			SiteKeywords:     []string{"e-commerce", "AI", "marketplace", "online shopping"},
			DefaultLanguage:  getEnv("DEFAULT_LANGUAGE", "tr"),
			Languages:        []string{"tr", "en", "ar"},
			GoogleAnalytics:  getEnv("GOOGLE_ANALYTICS", ""),
			GoogleTagManager: getEnv("GOOGLE_TAG_MANAGER", ""),
		},
		Logging: LoggingConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			Output:     getEnv("LOG_OUTPUT", "stdout"),
			MaxSize:    getEnvAsInt("LOG_MAX_SIZE", 100),
			MaxBackups: getEnvAsInt("LOG_MAX_BACKUPS", 3),
			MaxAge:     getEnvAsInt("LOG_MAX_AGE", 28),
			Compress:   getEnvAsBool("LOG_COMPRESS", true),
		},
	}

	// Set global config
	cfg = config

	return config
}

// overrideWithEnv overrides config values with environment variables
func overrideWithEnv(config *Config) {
	if env := getEnv("APP_ENV", ""); env != "" {
		config.Environment = env
	}
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

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(configPath, data, 0644)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Security.EncryptionKey == "" {
		return fmt.Errorf("encryption key is required")
	}

	if len(c.Security.EncryptionKey) < 32 {
		return fmt.Errorf("encryption key must be at least 32 characters long")
	}

	return nil
}

// GetNotificationConfig returns the notification configuration
func GetNotificationConfig() (NotificationConfig, bool) {
	if cfg == nil {
		return NotificationConfig{}, false
	}

	// Return default notification config
	return NotificationConfig{
		EnableEmail: true,
		EnableSMS:   false,
		EnablePush:  true,
		DefaultTTL:  3600,
		Templates: map[string]string{
			"welcome":  "welcome.html",
			"order":    "order.html",
			"payment":  "payment.html",
			"shipping": "shipping.html",
		},
		Types: map[string]NotificationType{
			"welcome": {
				Name:     "Welcome",
				Template: "welcome",
				Subject:  "Welcome to KolajAI",
				Enabled:  true,
				Priority: 1,
				Icon:     "welcome",
				Color:    "blue",
			},
			"order": {
				Name:     "Order",
				Template: "order",
				Subject:  "Order Confirmation",
				Enabled:  true,
				Priority: 2,
				Icon:     "shopping-cart",
				Color:    "green",
			},
		},
	}, true
}

// expandEnvVars expands ${VAR} or ${VAR:-default} patterns in the string
func expandEnvVars(s string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		// Remove ${ and }
		varExpr := match[2 : len(match)-1]

		// Check for default value syntax
		parts := strings.SplitN(varExpr, ":-", 2)
		varName := parts[0]
		defaultValue := ""
		if len(parts) > 1 {
			defaultValue = parts[1]
		}

		// Get environment variable or use default
		if value := os.Getenv(varName); value != "" {
			return value
		}
		return defaultValue
	})
}

// loadEnvFile loads environment variables from a .env file
func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		// .env file is optional
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		// Only set if not already set
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}
