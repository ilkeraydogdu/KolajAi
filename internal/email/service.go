package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Service handles email sending operations
type Service struct {
	Config      *Config
	Templates   map[string]*template.Template
	TemplateDir string
}

// NewService creates a new email service
func NewService(config *Config, templateDir string) (*Service, error) {
	service := &Service{
		Config:      config,
		Templates:   make(map[string]*template.Template),
		TemplateDir: templateDir,
	}

	// Load email templates
	err := service.LoadTemplates()
	if err != nil {
		return nil, err
	}

	return service, nil
}

// LoadTemplates loads all email templates
func (s *Service) LoadTemplates() error {
	log.Printf("Loading internal email templates from: %s", s.TemplateDir)

	// Load base template
	baseTemplatePath := filepath.Join(s.TemplateDir, "email_template.gohtml")
	log.Printf("Loading base template from: %s", baseTemplatePath)
	baseTemplate, err := template.ParseFiles(baseTemplatePath)
	if err != nil {
		log.Printf("Failed to parse base email template: %v", err)
		return fmt.Errorf("failed to parse base email template: %w", err)
	}
	log.Printf("Base template loaded successfully")

	// Load all template files
	templateFiles, err := filepath.Glob(filepath.Join(s.TemplateDir, "*.gohtml"))
	if err != nil {
		log.Printf("Failed to find email templates: %v", err)
		return fmt.Errorf("failed to find email templates: %w", err)
	}
	log.Printf("Found %d template files", len(templateFiles))

	for _, file := range templateFiles {
		if file == baseTemplatePath {
			log.Printf("Skipping base template: %s", file)
			continue // Skip base template
		}

		log.Printf("Loading template file: %s", file)
		name := filepath.Base(file)
		name = name[:len(name)-7] // Remove .gohtml
		log.Printf("Template name: %s", name)

		// Template içeriğini okuyorum
		content, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Failed to read template file %s: %v", file, err)
			return fmt.Errorf("failed to read email template %s: %w", name, err)
		}
		log.Printf("Template content: %s", string(content))

		// Yeni bir template oluşturup, base template'i kopyalıyorum
		newTemplate, err := template.Must(baseTemplate.Clone()).ParseFiles(file)
		if err != nil {
			log.Printf("Failed to parse email template %s: %v", name, err)
			return fmt.Errorf("failed to parse email template %s: %w", name, err)
		}

		// Template'leri listeleyerek tanımları kontrol ediyorum
		templateNames := []string{}
		for _, t := range newTemplate.Templates() {
			templateNames = append(templateNames, t.Name())
		}
		log.Printf("Templates in %s: %v", name, templateNames)

		// Verify template has necessary templates defined
		if t := newTemplate.Lookup("email_template"); t == nil {
			log.Printf("Warning: Template %s does not define 'email_template'", name)
		}

		if t := newTemplate.Lookup("email_content"); t == nil {
			log.Printf("Warning: Template %s does not define 'email_content'", name)
		}

		s.Templates[name] = newTemplate
		log.Printf("Successfully loaded template: %s", name)
	}

	// Log all available templates
	log.Printf("Loaded templates: %v", s.listTemplateNames())

	return nil
}

// listTemplateNames returns all loaded template names
func (s *Service) listTemplateNames() []string {
	names := make([]string, 0, len(s.Templates))
	for name := range s.Templates {
		names = append(names, name)
	}
	return names
}

// SendEmail sends an email
func (s *Service) SendEmail(to, subject, body string) error {
	log.Printf("Sending email to: %s, subject: %s", to, subject)

	// Debug mode for local development
	if s.Config.Debug {
		log.Printf("DEBUG EMAIL:\nTo: %s\nSubject: %s\nBody: %s", to, subject, body)
		return nil
	}

	// Create email data
	emailData := &EmailData{
		To:      []string{to},
		Subject: subject,
	}

	// Add retry logic
	maxRetries := 3
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		err := s.sendMail(emailData, body)
		if err == nil {
			return nil
		}
		lastErr = err
		log.Printf("Attempt %d failed: %v", i+1, err)
		time.Sleep(time.Second * time.Duration(i+1)) // Exponential backoff
	}

	return fmt.Errorf("failed to send email after %d attempts: %w", maxRetries, lastErr)
}

// sendMail sends an email using SMTP
func (s *Service) sendMail(data *EmailData, body string) error {
	log.Printf("Attempting to connect to SMTP server: %s:%d", s.Config.Host, s.Config.Port)

	// Prepare email headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.Config.FromName, s.Config.FromAddr)
	headers["To"] = strings.Join(data.To, ", ")
	headers["Subject"] = data.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	if len(data.CC) > 0 {
		headers["Cc"] = strings.Join(data.CC, ", ")
	}

	if len(data.BCC) > 0 {
		headers["Bcc"] = strings.Join(data.BCC, ", ")
	}

	// Prepare email message
	var message bytes.Buffer
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	// Create SMTP client with SSL/TLS
	tlsConfig := &tls.Config{
		ServerName:         s.Config.Host,
		InsecureSkipVerify: true, // Kendinden imzalı sertifika için
	}

	// Create connection
	conn, err := tls.Dial("tcp", s.Config.SMTPAddr(), tlsConfig)
	if err != nil {
		log.Printf("TLS Connection Error: %v", err)
		return fmt.Errorf("TLS bağlantı hatası: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.Config.Host)
	if err != nil {
		log.Printf("SMTP Client Error: %v", err)
		return fmt.Errorf("SMTP istemci hatası: %w", err)
	}
	defer client.Close()

	// Authenticate
	auth := smtp.PlainAuth("", s.Config.Username, s.Config.Password, s.Config.Host)
	if err = client.Auth(auth); err != nil {
		log.Printf("SMTP Authentication Error: %v", err)
		return fmt.Errorf("SMTP kimlik doğrulama hatası: %w", err)
	}

	// Set sender
	if err = client.Mail(s.Config.FromAddr); err != nil {
		log.Printf("SMTP Sender Error: %v", err)
		return fmt.Errorf("SMTP gönderici hatası: %w", err)
	}

	// Set recipients
	recipients := append(data.To, append(data.CC, data.BCC...)...)
	for _, recipient := range recipients {
		if err = client.Rcpt(recipient); err != nil {
			log.Printf("SMTP Recipient Error: %v", err)
			return fmt.Errorf("SMTP alıcı hatası: %w", err)
		}
	}

	// Send email body
	w, err := client.Data()
	if err != nil {
		log.Printf("SMTP Data Error: %v", err)
		return fmt.Errorf("SMTP veri hatası: %w", err)
	}

	_, err = w.Write(message.Bytes())
	if err != nil {
		log.Printf("SMTP Write Error: %v", err)
		return fmt.Errorf("SMTP yazma hatası: %w", err)
	}

	err = w.Close()
	if err != nil {
		log.Printf("SMTP Close Error: %v", err)
		return fmt.Errorf("SMTP kapatma hatası: %w", err)
	}

	log.Printf("Email sent successfully to: %s", strings.Join(data.To, ", "))
	return nil
}

// SendTemplateEmail sends an email using a template
func (s *Service) SendTemplateEmail(to, subject, templateName string, data map[string]interface{}) error {
	// Find template
	tmpl, exists := s.Templates[templateName]
	if !exists {
		return fmt.Errorf("email template '%s' not found", templateName)
	}

	// Add standard data
	if data == nil {
		data = make(map[string]interface{})
	}
	data["Subject"] = subject
	data["Year"] = time.Now().Year()
	data["SiteName"] = "Pofuduk DİJİTAL"
	data["SiteURL"] = "https://pofudukdijital.com"

	// Render template
	var buffer bytes.Buffer

	// Email_template şablonunu doğrudan çalıştırıyoruz
	if err := tmpl.ExecuteTemplate(&buffer, "email_template", data); err != nil {
		// Hata durumunda farklı bir yöntem dene - base template'i çağır
		log.Printf("Error executing template '%s': %v, trying Execute method", templateName, err)

		if err := tmpl.Execute(&buffer, data); err != nil {
			log.Printf("Error with Execute method too: %v", err)
			return fmt.Errorf("failed to render email template: %w", err)
		}
	}

	// Email içeriğini alıyoruz
	renderedContent := buffer.String()

	// Debug modda template render sorununu tespit etmek için içeriği yazdır
	if s.Config.Debug {
		log.Printf("Template %s rendered to: %s", templateName, renderedContent)
	}

	// Send email
	return s.SendEmail(to, subject, renderedContent)
}

// SendStructuredEmail sends an email using EmailData struct
func (s *Service) SendStructuredEmailFromData(data *EmailData) error {
	// Debug mode for local development
	if s.Config.Debug {
		log.Printf("DEBUG EMAIL:\nTo: %s\nSubject: %s\nType: %s",
			strings.Join(data.To, ", "), data.Subject, data.Type)
		return nil
	}

	// Find template based on email type
	templateName := string(data.Type)
	tmpl, exists := s.Templates[templateName]
	if !exists {
		return fmt.Errorf("email template for type '%s' not found", data.Type)
	}

	// Render template
	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Send email
	return s.sendMail(data, buffer.String())
}

// LogEmailSend logs an email send attempt to the database
func (s *Service) LogEmailSend(userID interface{}, to, subject, emailType, errorMsg string) error {
	// TODO: Implement database logging
	return nil
}

// SendWelcomeEmail sends a welcome email
func (s *Service) SendWelcomeEmail(to, name string) error {
	data := map[string]interface{}{
		"Name": name,
		"Year": time.Now().Year(),
	}
	return s.SendTemplateEmail(to, "Hoş Geldiniz - Pofuduk Dijital", "welcome", data)
}

// SendPasswordResetEmail sends a password reset email
func (s *Service) SendPasswordResetEmail(to, name, resetLink string) error {
	data := map[string]interface{}{
		"Name":      name,
		"ResetLink": resetLink,
		"Year":      time.Now().Year(),
	}
	return s.SendTemplateEmail(to, "Şifre Sıfırlama - Pofuduk Dijital", "password_reset", data)
}

// SendVerificationEmail sends an account verification email
func (s *Service) SendVerificationEmail(to, name, verificationLink string) error {
	data := map[string]interface{}{
		"Name":             name,
		"VerificationLink": verificationLink,
		"Year":             time.Now().Year(),
	}
	return s.SendTemplateEmail(to, "Hesap Doğrulama - Pofuduk Dijital", "verification", data)
}

// SendPasswordChangedEmail sends a password changed notification
func (s *Service) SendPasswordChangedEmail(to, name string) error {
	data := map[string]interface{}{
		"Name": name,
		"Year": time.Now().Year(),
	}
	return s.SendTemplateEmail(to, "Şifreniz Değiştirildi - Pofuduk Dijital", "password_changed", data)
}

// SendCustomEmail sends a custom email using EmailData
func (s *Service) SendCustomEmail(data *EmailData) error {
	return s.SendStructuredEmailFromData(data)
}
