package services

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"time"

	"kolajAi/internal/models"
	"kolajAi/internal/repository"
)

// EmailService handles email operations
type EmailService struct {
	repo     *repository.BaseRepository
	db       *sql.DB
	config   EmailConfig
	provider EmailProvider
}

// EmailConfig holds email configuration
type EmailConfig struct {
	Provider     string            `json:"provider"`
	SMTPHost     string            `json:"smtp_host"`
	SMTPPort     int               `json:"smtp_port"`
	Username     string            `json:"username"`
	Password     string            `json:"password"`
	FromEmail    string            `json:"from_email"`
	FromName     string            `json:"from_name"`
	APIKey       string            `json:"api_key"`
	APISecret    string            `json:"api_secret"`
	CustomConfig map[string]string `json:"custom_config"`
}

// EmailProvider interface for different email providers
type EmailProvider interface {
	SendEmail(req *EmailRequest) error
	SendBulkEmail(req *BulkEmailRequest) error
	GetDeliveryStatus(messageID string) (*EmailStatus, error)
}

// EmailRequest represents an email sending request
type EmailRequest struct {
	To          []string          `json:"to"`
	CC          []string          `json:"cc,omitempty"`
	BCC         []string          `json:"bcc,omitempty"`
	Subject     string            `json:"subject"`
	HTMLBody    string            `json:"html_body,omitempty"`
	TextBody    string            `json:"text_body,omitempty"`
	FromEmail   string            `json:"from_email,omitempty"`
	FromName    string            `json:"from_name,omitempty"`
	ReplyTo     string            `json:"reply_to,omitempty"`
	Attachments []EmailAttachment `json:"attachments,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	TemplateID  string            `json:"template_id,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Priority    EmailPriority     `json:"priority"`
	TrackOpens  bool              `json:"track_opens"`
	TrackClicks bool              `json:"track_clicks"`
}

// BulkEmailRequest represents bulk email sending request
type BulkEmailRequest struct {
	Template    EmailTemplate     `json:"template"`
	Recipients  []EmailRecipient  `json:"recipients"`
	FromEmail   string            `json:"from_email,omitempty"`
	FromName    string            `json:"from_name,omitempty"`
	ReplyTo     string            `json:"reply_to,omitempty"`
	Priority    EmailPriority     `json:"priority"`
	TrackOpens  bool              `json:"track_opens"`
	TrackClicks bool              `json:"track_clicks"`
}

// EmailRecipient represents a bulk email recipient
type EmailRecipient struct {
	Email     string                 `json:"email"`
	Name      string                 `json:"name,omitempty"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	ID       string `json:"id"`
	Subject  string `json:"subject"`
	HTMLBody string `json:"html_body"`
	TextBody string `json:"text_body"`
}

// EmailAttachment represents an email attachment
type EmailAttachment struct {
	Filename    string `json:"filename"`
	Content     []byte `json:"content"`
	ContentType string `json:"content_type"`
	Disposition string `json:"disposition"` // attachment or inline
}

// EmailPriority represents email priority levels
type EmailPriority string

const (
	EmailPriorityLow    EmailPriority = "low"
	EmailPriorityNormal EmailPriority = "normal"
	EmailPriorityHigh   EmailPriority = "high"
	EmailPriorityUrgent EmailPriority = "urgent"
)

// EmailStatus represents email delivery status
type EmailStatus struct {
	MessageID    string            `json:"message_id"`
	Status       string            `json:"status"`
	DeliveredAt  *time.Time        `json:"delivered_at,omitempty"`
	OpenedAt     *time.Time        `json:"opened_at,omitempty"`
	ClickedAt    *time.Time        `json:"clicked_at,omitempty"`
	BouncedAt    *time.Time        `json:"bounced_at,omitempty"`
	Error        string            `json:"error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// EmailLog represents email sending log
type EmailLog struct {
	ID          int                    `json:"id"`
	MessageID   string                 `json:"message_id"`
	ToEmail     string                 `json:"to_email"`
	FromEmail   string                 `json:"from_email"`
	Subject     string                 `json:"subject"`
	Status      string                 `json:"status"`
	Provider    string                 `json:"provider"`
	TemplateID  string                 `json:"template_id,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Error       string                 `json:"error,omitempty"`
	SentAt      time.Time              `json:"sent_at"`
	DeliveredAt *time.Time             `json:"delivered_at,omitempty"`
	OpenedAt    *time.Time             `json:"opened_at,omitempty"`
	ClickedAt   *time.Time             `json:"clicked_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// NewEmailService creates a new email service
func NewEmailService(repo *repository.BaseRepository, db *sql.DB, config EmailConfig) *EmailService {
	service := &EmailService{
		repo:   repo,
		db:     db,
		config: config,
	}

	// Initialize email provider based on configuration
	switch strings.ToLower(config.Provider) {
	case "smtp":
		service.provider = NewSMTPProvider(config)
	case "sendgrid":
		service.provider = NewSendGridProvider(config)
	case "mailgun":
		service.provider = NewMailgunProvider(config)
	case "ses":
		service.provider = NewSESProvider(config)
	default:
		service.provider = NewSMTPProvider(config) // Default to SMTP
	}

	return service
}

// SendEmail sends a single email
func (s *EmailService) SendEmail(req *EmailRequest) (*EmailLog, error) {
	// Validate request
	if err := s.validateEmailRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Set defaults
	if req.FromEmail == "" {
		req.FromEmail = s.config.FromEmail
	}
	if req.FromName == "" {
		req.FromName = s.config.FromName
	}
	if req.Priority == "" {
		req.Priority = EmailPriorityNormal
	}

	// Process template if specified
	if req.TemplateID != "" {
		if err := s.processTemplate(req); err != nil {
			return nil, fmt.Errorf("template processing failed: %w", err)
		}
	}

	// Generate message ID
	messageID := s.generateMessageID()

	// Create email log
	emailLog := &EmailLog{
		MessageID:  messageID,
		ToEmail:    strings.Join(req.To, ","),
		FromEmail:  req.FromEmail,
		Subject:    req.Subject,
		Status:     "pending",
		Provider:   s.config.Provider,
		TemplateID: req.TemplateID,
		Variables:  req.Variables,
		SentAt:     time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Send email
	err := s.provider.SendEmail(req)
	if err != nil {
		emailLog.Status = "failed"
		emailLog.Error = err.Error()
	} else {
		emailLog.Status = "sent"
	}

	// Save email log
	id, logErr := s.saveEmailLog(emailLog)
	if logErr != nil {
		fmt.Printf("Warning: Failed to save email log: %v\n", logErr)
	} else {
		emailLog.ID = int(id)
	}

	if err != nil {
		return emailLog, fmt.Errorf("failed to send email: %w", err)
	}

	return emailLog, nil
}

// SendBulkEmail sends bulk emails
func (s *EmailService) SendBulkEmail(req *BulkEmailRequest) ([]*EmailLog, error) {
	if len(req.Recipients) == 0 {
		return nil, errors.New("no recipients specified")
	}

	var emailLogs []*EmailLog
	var errors []error

	// Send to each recipient
	for _, recipient := range req.Recipients {
		emailReq := &EmailRequest{
			To:          []string{recipient.Email},
			Subject:     req.Template.Subject,
			HTMLBody:    req.Template.HTMLBody,
			TextBody:    req.Template.TextBody,
			FromEmail:   req.FromEmail,
			FromName:    req.FromName,
			ReplyTo:     req.ReplyTo,
			Variables:   recipient.Variables,
			Priority:    req.Priority,
			TrackOpens:  req.TrackOpens,
			TrackClicks: req.TrackClicks,
		}

		// Process template variables for this recipient
		if err := s.processTemplateVariables(emailReq); err != nil {
			errors = append(errors, fmt.Errorf("template processing failed for %s: %w", recipient.Email, err))
			continue
		}

		emailLog, err := s.SendEmail(emailReq)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to send to %s: %w", recipient.Email, err))
		}
		emailLogs = append(emailLogs, emailLog)
	}

	if len(errors) > 0 {
		return emailLogs, fmt.Errorf("bulk email partially failed: %d errors occurred", len(errors))
	}

	return emailLogs, nil
}

// SendTransactionalEmail sends predefined transactional emails
func (s *EmailService) SendTransactionalEmail(emailType string, recipient string, variables map[string]interface{}) error {
	templates := s.getTransactionalTemplates()
	
	template, exists := templates[emailType]
	if !exists {
		return fmt.Errorf("unknown email type: %s", emailType)
	}

	req := &EmailRequest{
		To:        []string{recipient},
		Subject:   template.Subject,
		HTMLBody:  template.HTMLBody,
		TextBody:  template.TextBody,
		Variables: variables,
		Priority:  EmailPriorityHigh,
	}

	_, err := s.SendEmail(req)
	return err
}

// SendWelcomeEmail sends welcome email to new users
func (s *EmailService) SendWelcomeEmail(user *models.User, verificationToken string) error {
	variables := map[string]interface{}{
		"UserName":          user.Name,
		"VerificationToken": verificationToken,
		"VerificationURL":   fmt.Sprintf("https://kolaj.ai/verify-email?token=%s", verificationToken),
	}

	return s.SendTransactionalEmail("welcome", user.Email, variables)
}

// SendPasswordResetEmail sends password reset email
func (s *EmailService) SendPasswordResetEmail(user *models.User, resetToken string) error {
	variables := map[string]interface{}{
		"UserName":   user.Name,
		"ResetToken": resetToken,
		"ResetURL":   fmt.Sprintf("https://kolaj.ai/reset-password?token=%s", resetToken),
		"ExpiresIn":  "1 hour",
	}

	return s.SendTransactionalEmail("password_reset", user.Email, variables)
}

// SendOrderConfirmationEmail sends order confirmation email
func (s *EmailService) SendOrderConfirmationEmail(order *models.Order, customer *models.Customer) error {
	variables := map[string]interface{}{
		"CustomerName": customer.GetFullName(),
		"OrderID":      order.ID,
		"OrderTotal":   fmt.Sprintf("%.2f %s", order.TotalAmount, order.Currency),
		"OrderDate":    order.CreatedAt.Format("2006-01-02 15:04"),
		"OrderURL":     fmt.Sprintf("https://kolaj.ai/orders/%d", order.ID),
	}

	return s.SendTransactionalEmail("order_confirmation", customer.User.Email, variables)
}

// SendShippingNotificationEmail sends shipping notification email
func (s *EmailService) SendShippingNotificationEmail(shipment *models.Shipment, customer *models.Customer) error {
	variables := map[string]interface{}{
		"CustomerName":    customer.GetFullName(),
		"OrderID":         shipment.OrderID,
		"TrackingNumber":  shipment.TrackingNumber,
		"TrackingURL":     fmt.Sprintf("https://kolaj.ai/track/%s", shipment.TrackingNumber),
		"EstimatedDelivery": "",
	}

	if shipment.EstimatedDeliveryDate != nil {
		variables["EstimatedDelivery"] = shipment.EstimatedDeliveryDate.Format("2006-01-02")
	}

	return s.SendTransactionalEmail("shipping_notification", customer.User.Email, variables)
}

// GetEmailStatus retrieves email delivery status
func (s *EmailService) GetEmailStatus(messageID string) (*EmailStatus, error) {
	if messageID == "" {
		return nil, errors.New("message ID is required")
	}

	return s.provider.GetDeliveryStatus(messageID)
}

// GetEmailLogs retrieves email logs with pagination
func (s *EmailService) GetEmailLogs(limit, offset int) ([]*EmailLog, error) {
	query := `SELECT id, message_id, to_email, from_email, subject, status, provider, 
			  template_id, variables, error, sent_at, delivered_at, opened_at, 
			  clicked_at, created_at, updated_at 
			  FROM email_logs ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query email logs: %w", err)
	}
	defer rows.Close()

	var logs []*EmailLog
	for rows.Next() {
		var log EmailLog
		var variablesJSON sql.NullString

		err := rows.Scan(
			&log.ID, &log.MessageID, &log.ToEmail, &log.FromEmail,
			&log.Subject, &log.Status, &log.Provider, &log.TemplateID,
			&variablesJSON, &log.Error, &log.SentAt, &log.DeliveredAt,
			&log.OpenedAt, &log.ClickedAt, &log.CreatedAt, &log.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email log: %w", err)
		}

		// Parse variables JSON
		if variablesJSON.Valid {
			json.Unmarshal([]byte(variablesJSON.String), &log.Variables)
		}

		logs = append(logs, &log)
	}

	return logs, nil
}

// Helper methods

func (s *EmailService) validateEmailRequest(req *EmailRequest) error {
	if len(req.To) == 0 {
		return errors.New("at least one recipient is required")
	}

	for _, email := range req.To {
		if !s.isValidEmail(email) {
			return fmt.Errorf("invalid email address: %s", email)
		}
	}

	if req.Subject == "" {
		return errors.New("subject is required")
	}

	if req.HTMLBody == "" && req.TextBody == "" {
		return errors.New("email body is required")
	}

	return nil
}

func (s *EmailService) isValidEmail(email string) bool {
	// Basic email validation
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func (s *EmailService) processTemplate(req *EmailRequest) error {
	// Get template from database or predefined templates
	templates := s.getTransactionalTemplates()
	template, exists := templates[req.TemplateID]
	if !exists {
		return fmt.Errorf("template not found: %s", req.TemplateID)
	}

	req.Subject = template.Subject
	req.HTMLBody = template.HTMLBody
	req.TextBody = template.TextBody

	return s.processTemplateVariables(req)
}

func (s *EmailService) processTemplateVariables(req *EmailRequest) error {
	if req.Variables == nil {
		return nil
	}

	// Process subject
	if tmpl, err := template.New("subject").Parse(req.Subject); err == nil {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, req.Variables); err == nil {
			req.Subject = buf.String()
		}
	}

	// Process HTML body
	if req.HTMLBody != "" {
		if tmpl, err := template.New("html").Parse(req.HTMLBody); err == nil {
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, req.Variables); err == nil {
				req.HTMLBody = buf.String()
			}
		}
	}

	// Process text body
	if req.TextBody != "" {
		if tmpl, err := template.New("text").Parse(req.TextBody); err == nil {
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, req.Variables); err == nil {
				req.TextBody = buf.String()
			}
		}
	}

	return nil
}

func (s *EmailService) generateMessageID() string {
	return fmt.Sprintf("%d@kolaj.ai", time.Now().UnixNano())
}

func (s *EmailService) saveEmailLog(log *EmailLog) (int64, error) {
	variablesJSON, _ := json.Marshal(log.Variables)

	query := `INSERT INTO email_logs (message_id, to_email, from_email, subject, status, 
			  provider, template_id, variables, error, sent_at, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := s.db.Exec(query, log.MessageID, log.ToEmail, log.FromEmail,
		log.Subject, log.Status, log.Provider, log.TemplateID, string(variablesJSON),
		log.Error, log.SentAt, log.CreatedAt, log.UpdatedAt)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (s *EmailService) getTransactionalTemplates() map[string]EmailTemplate {
	return map[string]EmailTemplate{
		"welcome": {
			ID:      "welcome",
			Subject: "Welcome to KolajAI - Verify Your Email",
			HTMLBody: `
				<h1>Welcome to KolajAI, {{.UserName}}!</h1>
				<p>Thank you for joining our marketplace. Please verify your email address by clicking the link below:</p>
				<a href="{{.VerificationURL}}" style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Verify Email</a>
				<p>If you didn't create this account, please ignore this email.</p>
			`,
			TextBody: `Welcome to KolajAI, {{.UserName}}!
			
Thank you for joining our marketplace. Please verify your email address by visiting: {{.VerificationURL}}

If you didn't create this account, please ignore this email.`,
		},
		"password_reset": {
			ID:      "password_reset",
			Subject: "Reset Your Password - KolajAI",
			HTMLBody: `
				<h1>Password Reset Request</h1>
				<p>Hello {{.UserName}},</p>
				<p>We received a request to reset your password. Click the link below to set a new password:</p>
				<a href="{{.ResetURL}}" style="background-color: #dc3545; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Reset Password</a>
				<p>This link will expire in {{.ExpiresIn}}.</p>
				<p>If you didn't request this, please ignore this email.</p>
			`,
			TextBody: `Password Reset Request

Hello {{.UserName}},

We received a request to reset your password. Visit this link to set a new password: {{.ResetURL}}

This link will expire in {{.ExpiresIn}}.

If you didn't request this, please ignore this email.`,
		},
		"order_confirmation": {
			ID:      "order_confirmation",
			Subject: "Order Confirmation #{{.OrderID}} - KolajAI",
			HTMLBody: `
				<h1>Order Confirmation</h1>
				<p>Hello {{.CustomerName}},</p>
				<p>Thank you for your order! Here are the details:</p>
				<ul>
					<li>Order ID: {{.OrderID}}</li>
					<li>Order Total: {{.OrderTotal}}</li>
					<li>Order Date: {{.OrderDate}}</li>
				</ul>
				<a href="{{.OrderURL}}" style="background-color: #28a745; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">View Order</a>
			`,
			TextBody: `Order Confirmation

Hello {{.CustomerName}},

Thank you for your order! Here are the details:
- Order ID: {{.OrderID}}
- Order Total: {{.OrderTotal}}
- Order Date: {{.OrderDate}}

View your order: {{.OrderURL}}`,
		},
		"shipping_notification": {
			ID:      "shipping_notification",
			Subject: "Your Order Has Shipped - KolajAI",
			HTMLBody: `
				<h1>Your Order Has Shipped!</h1>
				<p>Hello {{.CustomerName}},</p>
				<p>Great news! Your order #{{.OrderID}} has been shipped.</p>
				<p>Tracking Number: {{.TrackingNumber}}</p>
				{{if .EstimatedDelivery}}<p>Estimated Delivery: {{.EstimatedDelivery}}</p>{{end}}
				<a href="{{.TrackingURL}}" style="background-color: #17a2b8; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Track Package</a>
			`,
			TextBody: `Your Order Has Shipped!

Hello {{.CustomerName}},

Great news! Your order #{{.OrderID}} has been shipped.
Tracking Number: {{.TrackingNumber}}
{{if .EstimatedDelivery}}Estimated Delivery: {{.EstimatedDelivery}}{{end}}

Track your package: {{.TrackingURL}}`,
		},
	}
}

// SMTP Provider Implementation
type SMTPProvider struct {
	config EmailConfig
}

func NewSMTPProvider(config EmailConfig) *SMTPProvider {
	return &SMTPProvider{config: config}
}

func (p *SMTPProvider) SendEmail(req *EmailRequest) error {
	// Basic SMTP implementation
	auth := smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.SMTPHost)
	
	to := req.To
	msg := p.buildMessage(req)
	
	addr := fmt.Sprintf("%s:%d", p.config.SMTPHost, p.config.SMTPPort)
	return smtp.SendMail(addr, auth, req.FromEmail, to, []byte(msg))
}

func (p *SMTPProvider) SendBulkEmail(req *BulkEmailRequest) error {
	// For SMTP, send individual emails
	for _, recipient := range req.Recipients {
		emailReq := &EmailRequest{
			To:       []string{recipient.Email},
			Subject:  req.Template.Subject,
			HTMLBody: req.Template.HTMLBody,
			TextBody: req.Template.TextBody,
			FromEmail: req.FromEmail,
			FromName: req.FromName,
			Variables: recipient.Variables,
		}
		
		if err := p.SendEmail(emailReq); err != nil {
			return err
		}
	}
	return nil
}

func (p *SMTPProvider) GetDeliveryStatus(messageID string) (*EmailStatus, error) {
	// SMTP doesn't provide delivery status by default
	return &EmailStatus{
		MessageID: messageID,
		Status:    "sent",
	}, nil
}

func (p *SMTPProvider) buildMessage(req *EmailRequest) string {
	var msg bytes.Buffer
	
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", req.FromName, req.FromEmail))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(req.To, ",")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", req.Subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	
	if req.HTMLBody != "" {
		msg.WriteString(req.HTMLBody)
	} else {
		msg.WriteString(req.TextBody)
	}
	
	return msg.String()
}

// Placeholder implementations for other providers
type SendGridProvider struct{ config EmailConfig }
type MailgunProvider struct{ config EmailConfig }
type SESProvider struct{ config EmailConfig }

func NewSendGridProvider(config EmailConfig) *SendGridProvider { return &SendGridProvider{config} }
func NewMailgunProvider(config EmailConfig) *MailgunProvider   { return &MailgunProvider{config} }
func NewSESProvider(config EmailConfig) *SESProvider           { return &SESProvider{config} }

func (p *SendGridProvider) SendEmail(req *EmailRequest) error { 
	return errors.New("SendGrid provider disabled - use SMTP instead") 
}
func (p *SendGridProvider) SendBulkEmail(req *BulkEmailRequest) error { 
	return errors.New("SendGrid provider disabled - use SMTP instead") 
}
func (p *SendGridProvider) GetDeliveryStatus(messageID string) (*EmailStatus, error) { 
	return nil, errors.New("SendGrid provider disabled - use SMTP instead") 
}

func (p *MailgunProvider) SendEmail(req *EmailRequest) error { 
	return errors.New("Mailgun provider disabled - use SMTP instead") 
}
func (p *MailgunProvider) SendBulkEmail(req *BulkEmailRequest) error { 
	return errors.New("Mailgun provider disabled - use SMTP instead") 
}
func (p *MailgunProvider) GetDeliveryStatus(messageID string) (*EmailStatus, error) { 
	return nil, errors.New("Mailgun provider disabled - use SMTP instead") 
}

func (p *SESProvider) SendEmail(req *EmailRequest) error { 
	return errors.New("SES provider disabled - use SMTP instead") 
}
func (p *SESProvider) SendBulkEmail(req *BulkEmailRequest) error { 
	return errors.New("SES provider disabled - use SMTP instead") 
}
func (p *SESProvider) GetDeliveryStatus(messageID string) (*EmailStatus, error) { 
	return nil, errors.New("SES provider disabled - use SMTP instead") 
}