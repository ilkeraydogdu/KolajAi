package services

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"kolajAi/internal/models"
	"kolajAi/internal/repository"
)

// NotificationService handles notification operations
type NotificationService struct {
	repo        *repository.BaseRepository
	db          *sql.DB
	emailSvc    *EmailService
	channels    map[string]NotificationChannel
}

// NotificationChannel interface for different notification channels
type NotificationChannel interface {
	Send(notification *models.Notification) error
	GetDeliveryStatus(notificationID uint) (*DeliveryStatus, error)
	SupportsScheduling() bool
}

// DeliveryStatus represents notification delivery status
type DeliveryStatus struct {
	Status      string    `json:"status"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// NotificationRequest represents a notification sending request
type NotificationRequest struct {
	UserID      uint                   `json:"user_id"`
	Type        models.NotificationType `json:"type"`
	Channel     models.NotificationChannel `json:"channel"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	TemplateID  string                 `json:"template_id,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Priority    models.NotificationPriority `json:"priority"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// BulkNotificationRequest represents bulk notification request
type BulkNotificationRequest struct {
	UserIDs     []uint                 `json:"user_ids"`
	Type        models.NotificationType `json:"type"`
	Channel     models.NotificationChannel `json:"channel"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	TemplateID  string                 `json:"template_id,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Priority    models.NotificationPriority `json:"priority"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
}

// NotificationStats represents notification statistics
type NotificationStats struct {
	TotalSent      int     `json:"total_sent"`
	TotalDelivered int     `json:"total_delivered"`
	TotalRead      int     `json:"total_read"`
	TotalFailed    int     `json:"total_failed"`
	DeliveryRate   float64 `json:"delivery_rate"`
	ReadRate       float64 `json:"read_rate"`
}

// NewNotificationService creates a new notification service
func NewNotificationService(repo *repository.BaseRepository, db *sql.DB, emailSvc *EmailService) *NotificationService {
	service := &NotificationService{
		repo:     repo,
		db:       db,
		emailSvc: emailSvc,
		channels: make(map[string]NotificationChannel),
	}

	// Initialize notification channels
	service.channels["email"] = NewEmailChannel(emailSvc)
	service.channels["sms"] = NewSMSChannel()
	service.channels["push"] = NewPushChannel()
	service.channels["whatsapp"] = NewWhatsAppChannel()
	service.channels["in_app"] = NewInAppChannel(db)

	return service
}

// SendNotification sends a single notification
func (s *NotificationService) SendNotification(req *NotificationRequest) (*models.Notification, error) {
	// Validate request
	if err := s.validateNotificationRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check user preferences
	if !s.canSendToUser(req.UserID, req.Type, req.Channel) {
		return nil, errors.New("user has disabled this notification type/channel")
	}

	// Process template if specified
	if req.TemplateID != "" {
		if err := s.processTemplate(req); err != nil {
			return nil, fmt.Errorf("template processing failed: %w", err)
		}
	}

	// Create notification record
	notification := &models.Notification{
		RecipientID:   req.UserID,
		RecipientType: models.RecipientTypeUser,
		Type:          req.Type,
		Title:         req.Title,
		Message:       req.Message,
		Priority:      req.Priority,
		Status:        models.NotificationStatusPending,
		Channels:      []models.NotificationChannel{req.Channel},
		ScheduledAt:   req.ScheduledAt,
		ExpiresAt:     req.ExpiresAt,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Save to database
	id, err := s.repo.Create("notifications", notification)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}
	notification.ID = uint(id)

	// Send immediately or schedule
	if req.ScheduledAt == nil || req.ScheduledAt.Before(time.Now()) {
		if err := s.sendNotificationNow(notification); err != nil {
			s.updateNotificationStatus(notification.ID, models.NotificationStatusFailed, err.Error())
			return notification, fmt.Errorf("failed to send notification: %w", err)
		}
	}

	return notification, nil
}

// SendBulkNotification sends bulk notifications
func (s *NotificationService) SendBulkNotification(req *BulkNotificationRequest) ([]*models.Notification, error) {
	if len(req.UserIDs) == 0 {
		return nil, errors.New("no recipients specified")
	}

	var notifications []*models.Notification
	var errors []error

	// Send to each user
	for _, userID := range req.UserIDs {
		notificationReq := &NotificationRequest{
			UserID:      userID,
			Type:        req.Type,
			Channel:     req.Channel,
			Title:       req.Title,
			Message:     req.Message,
			Data:        req.Data,
			TemplateID:  req.TemplateID,
			Variables:   req.Variables,
			Priority:    req.Priority,
			ScheduledAt: req.ScheduledAt,
		}

		notification, err := s.SendNotification(notificationReq)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to send to user %d: %w", userID, err))
		}
		notifications = append(notifications, notification)
	}

	if len(errors) > 0 {
		return notifications, fmt.Errorf("bulk notification partially failed: %d errors occurred", len(errors))
	}

	return notifications, nil
}

// SendTransactionalNotification sends predefined transactional notifications
func (s *NotificationService) SendTransactionalNotification(notificationType string, userID uint, variables map[string]interface{}) error {
	templates := s.getTransactionalTemplates()
	
	template, exists := templates[notificationType]
	if !exists {
		return fmt.Errorf("unknown notification type: %s", notificationType)
	}

	// Send to all enabled channels for this notification type
	enabledChannels := s.getEnabledChannelsForUser(userID, template.Type)
	
	for _, channel := range enabledChannels {
		req := &NotificationRequest{
			UserID:    userID,
			Type:      template.Type,
			Channel:   channel,
			Title:     template.Title,
			Message:   template.Message,
			Variables: variables,
			Priority:  template.Priority,
		}

		_, err := s.SendNotification(req)
		if err != nil {
			fmt.Printf("Warning: Failed to send %s notification via %s to user %d: %v\n", 
				notificationType, channel, userID, err)
		}
	}

	return nil
}

// SendOrderStatusNotification sends order status update notification
func (s *NotificationService) SendOrderStatusNotification(orderID uint, customerID uint, status string) error {
	variables := map[string]interface{}{
		"OrderID": orderID,
		"Status":  status,
		"OrderURL": fmt.Sprintf("https://kolaj.ai/orders/%d", orderID),
	}

	return s.SendTransactionalNotification("order_status_update", customerID, variables)
}

// SendPaymentNotification sends payment notification
func (s *NotificationService) SendPaymentNotification(paymentID uint, customerID uint, amount float64, status string) error {
	variables := map[string]interface{}{
		"PaymentID": paymentID,
		"Amount":    fmt.Sprintf("%.2f TL", amount),
		"Status":    status,
	}

	return s.SendTransactionalNotification("payment_update", customerID, variables)
}

// SendPromotionalNotification sends promotional notification
func (s *NotificationService) SendPromotionalNotification(userIDs []uint, title, message string, data map[string]interface{}) error {
	req := &BulkNotificationRequest{
		UserIDs:  userIDs,
		Type:     models.NotificationTypeMarketing,
		Channel:  models.NotificationChannelPush, // Default to push for promotions
		Title:    title,
		Message:  message,
		Data:     data,
		Priority: models.NotificationPriorityLow,
	}

	_, err := s.SendBulkNotification(req)
	return err
}

// GetUserNotifications retrieves user notifications with pagination
func (s *NotificationService) GetUserNotifications(userID uint, limit, offset int) ([]*models.Notification, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	query := `SELECT id, recipient_id, type, title, message, priority, status, 
			  scheduled_at, sent_at, delivered_at, read_at, expires_at, created_at, updated_at 
			  FROM notifications WHERE recipient_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := s.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*models.Notification
	for rows.Next() {
		notification, err := s.scanNotification(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// GetUnreadNotifications retrieves unread notifications for user
func (s *NotificationService) GetUnreadNotifications(userID uint) ([]*models.Notification, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	query := `SELECT id, recipient_id, type, title, message, priority, status, 
			  scheduled_at, sent_at, delivered_at, read_at, expires_at, created_at, updated_at 
			  FROM notifications WHERE recipient_id = ? AND read_at IS NULL 
			  ORDER BY created_at DESC`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query unread notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*models.Notification
	for rows.Next() {
		notification, err := s.scanNotification(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// MarkAsRead marks notification as read
func (s *NotificationService) MarkAsRead(notificationID uint, userID uint) error {
	if notificationID == 0 || userID == 0 {
		return errors.New("notification ID and user ID are required")
	}

	query := `UPDATE notifications SET read_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = ? AND recipient_id = ? AND read_at IS NULL`

	result, err := s.db.Exec(query, notificationID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("notification not found or already read")
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for user
func (s *NotificationService) MarkAllAsRead(userID uint) error {
	if userID == 0 {
		return errors.New("user ID is required")
	}

	query := `UPDATE notifications SET read_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
			  WHERE recipient_id = ? AND read_at IS NULL`

	_, err := s.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// GetNotificationStats retrieves notification statistics
func (s *NotificationService) GetNotificationStats(userID uint, days int) (*NotificationStats, error) {
	since := time.Now().AddDate(0, 0, -days)

	query := `SELECT 
				COUNT(*) as total_sent,
				COUNT(CASE WHEN delivered_at IS NOT NULL THEN 1 END) as total_delivered,
				COUNT(CASE WHEN read_at IS NOT NULL THEN 1 END) as total_read,
				COUNT(CASE WHEN status = 'failed' THEN 1 END) as total_failed
			  FROM notifications 
			  WHERE recipient_id = ? AND created_at >= ?`

	var stats NotificationStats
	err := s.db.QueryRow(query, userID, since).Scan(
		&stats.TotalSent, &stats.TotalDelivered, &stats.TotalRead, &stats.TotalFailed,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification stats: %w", err)
	}

	// Calculate rates
	if stats.TotalSent > 0 {
		stats.DeliveryRate = float64(stats.TotalDelivered) / float64(stats.TotalSent) * 100
		stats.ReadRate = float64(stats.TotalRead) / float64(stats.TotalSent) * 100
	}

	return &stats, nil
}

// ProcessScheduledNotifications processes scheduled notifications
func (s *NotificationService) ProcessScheduledNotifications() error {
	query := `SELECT id, recipient_id, type, title, message, priority, status, 
			  scheduled_at, sent_at, delivered_at, read_at, expires_at, created_at, updated_at 
			  FROM notifications 
			  WHERE status = 'pending' AND scheduled_at IS NOT NULL AND scheduled_at <= CURRENT_TIMESTAMP`

	rows, err := s.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query scheduled notifications: %w", err)
	}
	defer rows.Close()

	var processed int
	for rows.Next() {
		notification, err := s.scanNotification(rows)
		if err != nil {
			fmt.Printf("Warning: Failed to scan scheduled notification: %v\n", err)
			continue
		}

		if err := s.sendNotificationNow(notification); err != nil {
			s.updateNotificationStatus(notification.ID, models.NotificationStatusFailed, err.Error())
			fmt.Printf("Warning: Failed to send scheduled notification %d: %v\n", notification.ID, err)
		} else {
			processed++
		}
	}

	fmt.Printf("Processed %d scheduled notifications\n", processed)
	return nil
}

// Helper methods

func (s *NotificationService) validateNotificationRequest(req *NotificationRequest) error {
	if req.UserID == 0 {
		return errors.New("user ID is required")
	}

	if req.Type == "" {
		return errors.New("notification type is required")
	}

	if req.Channel == "" {
		return errors.New("notification channel is required")
	}

	if req.Title == "" && req.Message == "" {
		return errors.New("title or message is required")
	}

	return nil
}

func (s *NotificationService) canSendToUser(userID uint, notificationType models.NotificationType, channel models.NotificationChannel) bool {
	// Check user notification preferences
	query := `SELECT COUNT(*) FROM user_notification_preferences 
			  WHERE user_id = ? AND notification_type = ? AND channel = ? AND enabled = 1`

	var count int
	err := s.db.QueryRow(query, userID, notificationType, channel).Scan(&count)
	if err != nil {
		// If no preferences found, allow by default
		return true
	}

	return count > 0
}

func (s *NotificationService) processTemplate(req *NotificationRequest) error {
	templates := s.getTransactionalTemplates()
	template, exists := templates[req.TemplateID]
	if !exists {
		return fmt.Errorf("template not found: %s", req.TemplateID)
	}

	req.Type = template.Type
	req.Title = template.Title
	req.Message = template.Message
	req.Priority = template.Priority

	// Process template variables
	if req.Variables != nil {
		req.Title = s.processTemplateVariables(req.Title, req.Variables)
		req.Message = s.processTemplateVariables(req.Message, req.Variables)
	}

	return nil
}

func (s *NotificationService) processTemplateVariables(text string, variables map[string]interface{}) string {
	result := text
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

func (s *NotificationService) sendNotificationNow(notification *models.Notification) error {
	// Use first channel from channels array
	if len(notification.Channels) == 0 {
		return errors.New("no channels specified for notification")
	}
	
	channelName := string(notification.Channels[0])
	channel, exists := s.channels[channelName]
	if !exists {
		return fmt.Errorf("unsupported channel: %s", channelName)
	}

	// Update status to sending
	s.updateNotificationStatus(notification.ID, models.NotificationStatusSending, "")

	// Send notification
	err := channel.Send(notification)
	if err != nil {
		return err
	}

	// Update status to sent
	s.updateNotificationStatus(notification.ID, models.NotificationStatusSent, "")
	s.updateSentAt(notification.ID)

	return nil
}

func (s *NotificationService) updateNotificationStatus(id uint, status models.NotificationStatus, errorMsg string) {
	query := `UPDATE notifications SET status = ?, error = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	s.db.Exec(query, status, errorMsg, id)
}

func (s *NotificationService) updateSentAt(id uint) {
	query := `UPDATE notifications SET sent_at = CURRENT_TIMESTAMP WHERE id = ?`
	s.db.Exec(query, id)
}

func (s *NotificationService) scanNotification(rows *sql.Rows) (*models.Notification, error) {
	var notification models.Notification

	err := rows.Scan(
		&notification.ID, &notification.RecipientID, &notification.Type,
		&notification.Title, &notification.Message, &notification.Priority,
		&notification.Status, &notification.ScheduledAt, &notification.SentAt,
		&notification.DeliveredAt, &notification.ReadAt, &notification.ExpiresAt,
		&notification.CreatedAt, &notification.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &notification, nil
}

func (s *NotificationService) getEnabledChannelsForUser(userID uint, notificationType models.NotificationType) []models.NotificationChannel {
	// Default channels if no preferences found
	defaultChannels := []models.NotificationChannel{
		models.NotificationChannelInApp,
		models.NotificationChannelEmail,
	}

	query := `SELECT channel FROM user_notification_preferences 
			  WHERE user_id = ? AND notification_type = ? AND enabled = 1`

	rows, err := s.db.Query(query, userID, notificationType)
	if err != nil {
		return defaultChannels
	}
	defer rows.Close()

	var channels []models.NotificationChannel
	for rows.Next() {
		var channel models.NotificationChannel
		if err := rows.Scan(&channel); err == nil {
			channels = append(channels, channel)
		}
	}

	if len(channels) == 0 {
		return defaultChannels
	}

	return channels
}

// SimpleTemplate represents a simplified template for transactional notifications
type SimpleTemplate struct {
	ID       string
	Type     models.NotificationType
	Title    string
	Message  string
	Priority models.NotificationPriority
}

func (s *NotificationService) getTransactionalTemplates() map[string]*SimpleTemplate {
	return map[string]*SimpleTemplate{
		"order_status_update": {
			ID:       "order_status_update",
			Type:     models.NotificationTypeTransactional,
			Title:    "Order Status Update",
			Message:  "Your order #{{.OrderID}} status has been updated to {{.Status}}",
			Priority: models.NotificationPriorityHigh,
		},
		"payment_update": {
			ID:       "payment_update",
			Type:     models.NotificationTypeTransactional,
			Title:    "Payment Update",
			Message:  "Your payment of {{.Amount}} has been {{.Status}}",
			Priority: models.NotificationPriorityHigh,
		},
		"new_message": {
			ID:       "new_message",
			Type:     models.NotificationTypeInfo,
			Title:    "New Message",
			Message:  "You have received a new message from {{.SenderName}}",
			Priority: models.NotificationPriorityNormal,
		},
		"system_maintenance": {
			ID:       "system_maintenance",
			Type:     models.NotificationTypeSystem,
			Title:    "System Maintenance",
			Message:  "Scheduled maintenance will begin at {{.MaintenanceTime}}",
			Priority: models.NotificationPriorityHigh,
		},
	}
}

// Channel implementations

// EmailChannel implementation
type EmailChannel struct {
	emailService *EmailService
}

func NewEmailChannel(emailService *EmailService) *EmailChannel {
	return &EmailChannel{emailService: emailService}
}

func (c *EmailChannel) Send(notification *models.Notification) error {
	if c.emailService == nil {
		return errors.New("email service not configured")
	}

	// Get user email
	var userEmail string
	query := "SELECT email FROM users WHERE id = ?"
	err := c.emailService.db.QueryRow(query, notification.RecipientID).Scan(&userEmail)
	if err != nil {
		return fmt.Errorf("failed to get user email: %w", err)
	}

	// Send email notification
	req := &EmailRequest{
		To:       []string{userEmail},
		Subject:  notification.Title,
		HTMLBody: fmt.Sprintf("<h3>%s</h3><p>%s</p>", notification.Title, notification.Message),
		TextBody: fmt.Sprintf("%s\n\n%s", notification.Title, notification.Message),
		Priority: EmailPriorityNormal,
	}

	_, err = c.emailService.SendEmail(req)
	return err
}

func (c *EmailChannel) GetDeliveryStatus(notificationID uint) (*DeliveryStatus, error) {
	return &DeliveryStatus{Status: "sent"}, nil
}

func (c *EmailChannel) SupportsScheduling() bool {
	return true
}

// Placeholder channel implementations
type SMSChannel struct{}
type PushChannel struct{}
type WhatsAppChannel struct{}
type InAppChannel struct{ db *sql.DB }

func NewSMSChannel() *SMSChannel           { return &SMSChannel{} }
func NewPushChannel() *PushChannel         { return &PushChannel{} }
func NewWhatsAppChannel() *WhatsAppChannel { return &WhatsAppChannel{} }
func NewInAppChannel(db *sql.DB) *InAppChannel { return &InAppChannel{db: db} }

func (c *SMSChannel) Send(notification *models.Notification) error {
	fmt.Printf("SMS: %s - %s\n", notification.Title, notification.Message)
	return nil
}
func (c *SMSChannel) GetDeliveryStatus(notificationID uint) (*DeliveryStatus, error) {
	return &DeliveryStatus{Status: "sent"}, nil
}
func (c *SMSChannel) SupportsScheduling() bool { return true }

func (c *PushChannel) Send(notification *models.Notification) error {
	fmt.Printf("Push: %s - %s\n", notification.Title, notification.Message)
	return nil
}
func (c *PushChannel) GetDeliveryStatus(notificationID uint) (*DeliveryStatus, error) {
	return &DeliveryStatus{Status: "sent"}, nil
}
func (c *PushChannel) SupportsScheduling() bool { return true }

func (c *WhatsAppChannel) Send(notification *models.Notification) error {
	fmt.Printf("WhatsApp: %s - %s\n", notification.Title, notification.Message)
	return nil
}
func (c *WhatsAppChannel) GetDeliveryStatus(notificationID uint) (*DeliveryStatus, error) {
	return &DeliveryStatus{Status: "sent"}, nil
}
func (c *WhatsAppChannel) SupportsScheduling() bool { return true }

func (c *InAppChannel) Send(notification *models.Notification) error {
	// In-app notifications are just stored in database, no external sending required
	return nil
}
func (c *InAppChannel) GetDeliveryStatus(notificationID uint) (*DeliveryStatus, error) {
	return &DeliveryStatus{Status: "delivered"}, nil
}
func (c *InAppChannel) SupportsScheduling() bool { return true }