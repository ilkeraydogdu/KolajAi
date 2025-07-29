package config

import "time"

// Config represents the main configuration structure
type Config struct {
	Environment string                       `yaml:"environment"`
	Server      ServerConfig                 `yaml:"server"`
	Database    DatabaseConfig               `yaml:"database"`
	Security    SecurityConfig               `yaml:"security"`
	Cache       CacheConfig                  `yaml:"cache"`
	Email       EmailConfig                  `yaml:"email"`
	SEO         SEOConfig                    `yaml:"seo"`
	Logging     LoggingConfig                `yaml:"logging"`
	AI          AIConfig                     `yaml:"ai"`
	Marketplace map[string]MarketplaceConfig `yaml:"marketplace"`
	Payment     map[string]PaymentConfig     `yaml:"payment"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port           int    `yaml:"port"`
	Host           string `yaml:"host"`
	Domain         string `yaml:"domain"`
	ReadTimeout    int    `yaml:"read_timeout"`
	WriteTimeout   int    `yaml:"write_timeout"`
	IdleTimeout    int    `yaml:"idle_timeout"`
	MaxHeaderBytes int    `yaml:"max_header_bytes"`
}

// DatabaseConfig represents database configuration
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

// SecurityConfig represents security configuration
type SecurityConfig struct {
	EncryptionKey      string                    `yaml:"encryption_key"`
	JWTSecret          string                    `yaml:"jwt_secret"`
	CSRFSecret         string                    `yaml:"csrf_secret"`
	SessionSecret      string                    `yaml:"session_secret"`
	TwoFactorEnabled   bool                      `yaml:"two_factor_enabled"`
	RateLimitEnabled   bool                      `yaml:"rate_limit_enabled"`
	CORSEnabled        bool                      `yaml:"cors_enabled"`
	AllowedOrigins     []string                  `yaml:"allowed_origins"`
	Vault              VaultConfig               `yaml:"vault"`
	CredentialRotation CredentialRotationConfig  `yaml:"credential_rotation"`
}

// VaultConfig represents HashiCorp Vault configuration
type VaultConfig struct {
	Enabled bool   `yaml:"enabled"`
	Address string `yaml:"address"`
	Token   string `yaml:"token"`
	Path    string `yaml:"path"`
}

// CredentialRotationConfig represents credential rotation configuration
type CredentialRotationConfig struct {
	Enabled               bool   `yaml:"enabled"`
	CheckInterval         string `yaml:"check_interval"`
	DefaultRotationPeriod string `yaml:"default_rotation_period"`
	NotifyBefore          string `yaml:"notify_before"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	Driver         string `yaml:"driver"`
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Password       string `yaml:"password"`
	DB             int    `yaml:"db"`
	DefaultTTL     string `yaml:"default_ttl"`
	MaxMemoryUsage int64  `yaml:"max_memory_usage"`
}

// EmailConfig represents email configuration
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

// SEOConfig represents SEO configuration
type SEOConfig struct {
	SiteName         string   `yaml:"site_name"`
	SiteDescription  string   `yaml:"site_description"`
	SiteKeywords     []string `yaml:"site_keywords"`
	DefaultLanguage  string   `yaml:"default_language"`
	Languages        []string `yaml:"languages"`
	GoogleAnalytics  string   `yaml:"google_analytics"`
	GoogleTagManager string   `yaml:"google_tag_manager"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// AIConfig represents AI service configuration
type AIConfig struct {
	OpenAIKey      string         `yaml:"openai_key"`
	AnthropicKey   string         `yaml:"anthropic_key"`
	StabilityAIKey string         `yaml:"stability_ai_key"`
	ReplicateKey   string         `yaml:"replicate_key"`
	HuggingFaceKey string         `yaml:"huggingface_key"`
	DefaultCredits int            `yaml:"default_credits"`
	CreditCosts    map[string]int `yaml:"credit_costs"`
}

// MarketplaceConfig represents marketplace integration configuration
type MarketplaceConfig struct {
	Enabled     bool   `yaml:"enabled"`
	APIKey      string `yaml:"api_key"`
	APISecret   string `yaml:"api_secret"`
	MerchantID  string `yaml:"merchant_id"`
	SupplierID  string `yaml:"supplier_id"`
	AccessKey   string `yaml:"access_key"`
	SecretKey   string `yaml:"secret_key"`
	BranchCode  string `yaml:"branch_code"`
	Environment string `yaml:"environment"`
}

// PaymentConfig represents payment integration configuration
type PaymentConfig struct {
	Enabled     bool   `yaml:"enabled"`
	APIKey      string `yaml:"api_key"`
	APISecret   string `yaml:"api_secret"`
	Environment string `yaml:"environment"`
}