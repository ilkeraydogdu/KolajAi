package email

import (
	"crypto/tls"
	"fmt"
)

// Config holds the email configuration
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	FromAddr string
	FromName string
	TLS      bool
	Debug    bool // Debug mode for local development
}

// NewConfig creates a new email configuration
func NewConfig(host string, port int, username, password, fromName, fromAddr string, useTLS bool) *Config {
	return &Config{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		FromName: fromName,
		FromAddr: fromAddr,
		TLS:      useTLS,
	}
}

// DefaultConfig returns a default email configuration
func DefaultConfig() *Config {
	return &Config{
		Host:     "ilkeraydogdu.com.tr",
		Port:     465,
		Username: "mail@ilkeraydogdu.com.tr",
		Password: "ilkN.2801",
		FromAddr: "mail@ilkeraydogdu.com.tr",
		FromName: "Pofuduk DİJİTAL",
		TLS:      true,
		Debug:    true,
	}
}

// SMTPAddr returns the SMTP address with port
func (c *Config) SMTPAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// TLSConfig returns the TLS configuration
func (c *Config) TLSConfig() *tls.Config {
	return &tls.Config{
		ServerName:         c.Host,
		InsecureSkipVerify: false, // Güvenlik için false yapıyoruz
		MinVersion:         tls.VersionTLS12,
	}
}
