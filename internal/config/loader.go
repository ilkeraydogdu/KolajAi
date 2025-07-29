package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// LoadConfig loads configuration from YAML file with environment variable support
func LoadConfig(filename string) (*Config, error) {
	// Load .env file if exists
	loadEnvFile(".env")

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Replace environment variables
	data = []byte(expandEnvVars(string(data)))

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
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

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	// Validate server configuration
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// Validate security configuration
	if len(config.Security.EncryptionKey) != 32 && len(config.Security.EncryptionKey) != 44 {
		// 32 bytes raw or 44 bytes base64 encoded
		return fmt.Errorf("encryption key must be 32 bytes")
	}

	if config.Security.JWTSecret == "" {
		return fmt.Errorf("JWT secret cannot be empty")
	}

	if config.Security.CSRFSecret == "" {
		return fmt.Errorf("CSRF secret cannot be empty")
	}

	if config.Security.SessionSecret == "" {
		return fmt.Errorf("session secret cannot be empty")
	}

	// Validate database configuration
	if config.Database.Driver == "" {
		return fmt.Errorf("database driver cannot be empty")
	}

	// Validate cache configuration
	if config.Cache.Driver == "" {
		config.Cache.Driver = "memory"
	}

	// Parse durations
	if config.Cache.DefaultTTL != "" {
		if _, err := time.ParseDuration(config.Cache.DefaultTTL); err != nil {
			return fmt.Errorf("invalid cache default TTL: %w", err)
		}
	}

	// Validate email configuration if SMTP is configured
	if config.Email.SMTPHost != "" {
		if config.Email.SMTPPort <= 0 {
			return fmt.Errorf("invalid SMTP port")
		}
		if config.Email.FromEmail == "" {
			return fmt.Errorf("from email cannot be empty when SMTP is configured")
		}
	}

	// Validate marketplace configurations
	for name, marketplace := range config.Marketplace {
		if marketplace.Enabled {
			if marketplace.APIKey == "" {
				return fmt.Errorf("marketplace %s: API key cannot be empty when enabled", name)
			}
			if marketplace.Environment == "" {
				marketplace.Environment = "sandbox"
			}
		}
	}

	// Validate payment configurations
	for name, payment := range config.Payment {
		if payment.Enabled {
			if payment.APIKey == "" {
				return fmt.Errorf("payment %s: API key cannot be empty when enabled", name)
			}
			if payment.Environment == "" {
				payment.Environment = "sandbox"
			}
		}
	}

	return nil
}

// GetDefaultConfig returns default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Environment: "development",
		Server: ServerConfig{
			Port:           8081,
			Host:           "0.0.0.0",
			Domain:         "localhost",
			ReadTimeout:    10,
			WriteTimeout:   10,
			IdleTimeout:    120,
			MaxHeaderBytes: 1048576,
		},
		Database: DatabaseConfig{
			Driver:          "sqlite3",
			Name:            "kolajAi.db",
			MaxOpenConns:    25,
			MaxIdleConns:    25,
			ConnMaxLifetime: 5,
		},
		Security: SecurityConfig{
			EncryptionKey:    "defaultencryptionkey32byteslong!",
			JWTSecret:        "defaultjwtsecret",
			CSRFSecret:       "defaultcsrfsecret",
			SessionSecret:    "defaultsessionsecret",
			TwoFactorEnabled: false,
			RateLimitEnabled: true,
			CORSEnabled:      true,
			AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8081"},
			Vault: VaultConfig{
				Enabled: false,
			},
			CredentialRotation: CredentialRotationConfig{
				Enabled:               true,
				CheckInterval:         "1h",
				DefaultRotationPeriod: "720h",
				NotifyBefore:          "72h",
			},
		},
		Cache: CacheConfig{
			Driver:         "memory",
			DefaultTTL:     "30m",
			MaxMemoryUsage: 1073741824,
		},
		Email: EmailConfig{
			SMTPHost:  "smtp.gmail.com",
			SMTPPort:  587,
			FromEmail: "noreply@kolajAi.com",
			FromName:  "KolajAI",
			UseSSL:    false,
			UseTLS:    true,
		},
		SEO: SEOConfig{
			SiteName:        "KolajAI Enterprise Marketplace",
			SiteDescription: "Advanced AI-powered e-commerce platform",
			SiteKeywords:    []string{"e-commerce", "AI", "marketplace", "online shopping"},
			DefaultLanguage: "tr",
			Languages:       []string{"tr", "en", "ar"},
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		},
		AI: AIConfig{
			DefaultCredits: 100,
			CreditCosts: map[string]int{
				"image_generation":   40,
				"content_generation": 10,
				"template_creation":  50,
				"chat_message":       5,
				"image_analysis":     15,
			},
		},
		Marketplace: make(map[string]MarketplaceConfig),
		Payment:     make(map[string]PaymentConfig),
	}
}