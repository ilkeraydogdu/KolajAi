package notifications

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// NotificationManager handles comprehensive notification management
type NotificationManager struct {
	db        *sql.DB
	channels  map[string]NotificationChannel
	templates map[string]NotificationTemplate
	config    NotificationConfig
	queue     NotificationQueue
}

// NotificationConfig holds notification configuration
type NotificationConfig struct {
	DefaultChannel     string                          `json:"default_channel"`
	RetryAttempts      int                             `json:"retry_attempts"`
	RetryDelay         time.Duration                   `json:"retry_delay"`
	BatchSize          int                             `json:"batch_size"`
	QueueSize          int                             `json:"queue_size"`
	Workers            int                             `json:"workers"`
	EnableRateLimiting bool                            `json:"enable_rate_limiting"`
	RateLimits         map[string]RateLimit            `json:"rate_limits"`
	Templates          map[string]NotificationTemplate `json:"templates"`
	Channels           map[string]ChannelConfig        `json:"channels"`
	UserPreferences    UserPreferenceConfig            `json:"user_preferences"`
}

// RateLimit defines rate limiting configuration
type RateLimit struct {
	MaxPerMinute int           `json:"max_per_minute"`
	MaxPerHour   int           `json:"max_per_hour"`
	MaxPerDay    int           `json:"max_per_day"`
	BurstSize    int           `json:"burst_size"`
	Window       time.Duration `json:"window"`
}

// ChannelConfig holds channel-specific configuration
type ChannelConfig struct {
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Timeout     time.Duration          `json:"timeout"`
	RetryPolicy RetryPolicy            `json:"retry_policy"`
	Settings    map[string]interface{} `json:"settings"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxAttempts   int           `json:"max_attempts"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// UserPreferenceConfig holds user preference settings
type UserPreferenceConfig struct {
	AllowOptOut     bool     `json:"allow_opt_out"`
	DefaultChannels []string `json:"default_channels"`
	RequiredTypes   []string `json:"required_types"`
	OptOutTypes     []string `json:"opt_out_types"`
}

// Notification represents a notification
type Notification struct {
	ID          string                 `json:"id"`
	Type        NotificationType       `json:"type"`
	Category    string                 `json:"category"`
	Priority    NotificationPriority   `json:"priority"`
	Recipients  []Recipient            `json:"recipients"`
	Subject     string                 `json:"subject"`
	Content     string                 `json:"content"`
	Data        map[string]interface{} `json:"data"`
	Channels    []string               `json:"channels"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Template    string                 `json:"template,omitempty"`
	Language    string                 `json:"language"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	Status      NotificationStatus     `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	SentAt      *time.Time             `json:"sent_at,omitempty"`
	DeliveredAt *time.Time             `json:"delivered_at,omitempty"`
	ReadAt      *time.Time             `json:"read_at,omitempty"`
	ClickedAt   *time.Time             `json:"clicked_at,omitempty"`
	Attempts    int                    `json:"attempts"`
	LastError   string                 `json:"last_error,omitempty"`
	TrackingID  string                 `json:"tracking_id"`
	ParentID    string                 `json:"parent_id,omitempty"`
	ThreadID    string                 `json:"thread_id,omitempty"`
}

// NotificationType represents different notification types
type NotificationType string

const (
	NotificationTypeInfo          NotificationType = "info"
	NotificationTypeWarning       NotificationType = "warning"
	NotificationTypeError         NotificationType = "error"
	NotificationTypeSuccess       NotificationType = "success"
	NotificationTypeMarketing     NotificationType = "marketing"
	NotificationTypeTransactional NotificationType = "transactional"
	NotificationTypeSystem        NotificationType = "system"
	NotificationTypeReminder      NotificationType = "reminder"
	NotificationTypeUpdate        NotificationType = "update"
	NotificationTypePromotion     NotificationType = "promotion"
)

// NotificationPriority represents notification priority levels
type NotificationPriority string

const (
	PriorityLow      NotificationPriority = "low"
	PriorityNormal   NotificationPriority = "normal"
	PriorityHigh     NotificationPriority = "high"
	PriorityCritical NotificationPriority = "critical"
	PriorityUrgent   NotificationPriority = "urgent"
)

// NotificationStatus represents notification status
type NotificationStatus string

const (
	StatusPending   NotificationStatus = "pending"
	StatusQueued    NotificationStatus = "queued"
	StatusSending   NotificationStatus = "sending"
	StatusSent      NotificationStatus = "sent"
	StatusDelivered NotificationStatus = "delivered"
	StatusRead      NotificationStatus = "read"
	StatusClicked   NotificationStatus = "clicked"
	StatusFailed    NotificationStatus = "failed"
	StatusExpired   NotificationStatus = "expired"
	StatusCancelled NotificationStatus = "cancelled"
)

// Recipient represents a notification recipient
type Recipient struct {
	ID          string                 `json:"id"`
	Type        RecipientType          `json:"type"`
	Address     string                 `json:"address"`
	Name        string                 `json:"name"`
	Language    string                 `json:"language"`
	Timezone    string                 `json:"timezone"`
	Preferences map[string]bool        `json:"preferences"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RecipientType represents recipient types
type RecipientType string

const (
	RecipientTypeUser  RecipientType = "user"
	RecipientTypeEmail RecipientType = "email"
	RecipientTypePhone RecipientType = "phone"
	RecipientTypeGroup RecipientType = "group"
	RecipientTypeRole  RecipientType = "role"
)

// NotificationChannel interface for different delivery channels
type NotificationChannel interface {
	Send(ctx context.Context, notification *Notification, recipient *Recipient) error
	GetName() string
	GetPriority() int
	IsEnabled() bool
	ValidateRecipient(recipient *Recipient) error
	GetDeliveryStatus(trackingID string) (*DeliveryStatus, error)
}

// DeliveryStatus represents delivery status for a channel
type DeliveryStatus struct {
	Status      string                 `json:"status"`
	DeliveredAt *time.Time             `json:"delivered_at,omitempty"`
	ReadAt      *time.Time             `json:"read_at,omitempty"`
	ClickedAt   *time.Time             `json:"clicked_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NotificationTemplate represents a notification template
type NotificationTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        NotificationType       `json:"type"`
	Category    string                 `json:"category"`
	Language    string                 `json:"language"`
	Subject     string                 `json:"subject"`
	Content     string                 `json:"content"`
	HTMLContent string                 `json:"html_content"`
	Variables   []TemplateVariable     `json:"variables"`
	Channels    []string               `json:"channels"`
	Priority    NotificationPriority   `json:"priority"`
	TTL         time.Duration          `json:"ttl"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	IsActive    bool                   `json:"is_active"`
}

// TemplateVariable represents a template variable
type TemplateVariable struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value"`
	Description  string      `json:"description"`
	Validation   string      `json:"validation"`
}

// NotificationQueue interface for queuing notifications
type NotificationQueue interface {
	Enqueue(notification *Notification) error
	Dequeue() (*Notification, error)
	Size() int
	Clear() error
	GetPending() ([]*Notification, error)
}

// NotificationStats represents notification statistics
type NotificationStats struct {
	TotalSent       int                          `json:"total_sent"`
	TotalDelivered  int                          `json:"total_delivered"`
	TotalRead       int                          `json:"total_read"`
	TotalClicked    int                          `json:"total_clicked"`
	TotalFailed     int                          `json:"total_failed"`
	ByType          map[NotificationType]int     `json:"by_type"`
	ByChannel       map[string]int               `json:"by_channel"`
	ByPriority      map[NotificationPriority]int `json:"by_priority"`
	DeliveryRate    float64                      `json:"delivery_rate"`
	ReadRate        float64                      `json:"read_rate"`
	ClickRate       float64                      `json:"click_rate"`
	AvgDeliveryTime time.Duration                `json:"avg_delivery_time"`
	TrendData       []NotificationTrendPoint     `json:"trend_data"`
}

// NotificationTrendPoint represents a point in trend data
type NotificationTrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Sent      int       `json:"sent"`
	Delivered int       `json:"delivered"`
	Read      int       `json:"read"`
	Clicked   int       `json:"clicked"`
	Failed    int       `json:"failed"`
}

// UserPreference represents user notification preferences
type UserPreference struct {
	UserID     string                    `json:"user_id"`
	Channels   map[string]bool           `json:"channels"`
	Types      map[NotificationType]bool `json:"types"`
	Categories map[string]bool           `json:"categories"`
	Frequency  string                    `json:"frequency"`
	QuietHours QuietHours                `json:"quiet_hours"`
	Language   string                    `json:"language"`
	Timezone   string                    `json:"timezone"`
	OptedOut   bool                      `json:"opted_out"`
	UpdatedAt  time.Time                 `json:"updated_at"`
}

// QuietHours represents quiet hours configuration
type QuietHours struct {
	Enabled   bool   `json:"enabled"`
	StartTime string `json:"start_time"` // "22:00"
	EndTime   string `json:"end_time"`   // "08:00"
	Timezone  string `json:"timezone"`
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(db *sql.DB, config NotificationConfig) *NotificationManager {
	nm := &NotificationManager{
		db:        db,
		channels:  make(map[string]NotificationChannel),
		templates: make(map[string]NotificationTemplate),
		config:    config,
	}

	nm.createNotificationTables()
	nm.loadTemplates()
	nm.initializeChannels()
	nm.startWorkers()

	return nm
}

// createNotificationTables creates necessary tables for notifications
func (nm *NotificationManager) createNotificationTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS notifications (
			id VARCHAR(128) PRIMARY KEY,
			type VARCHAR(50) NOT NULL,
			category VARCHAR(100),
			priority VARCHAR(20) NOT NULL,
			recipients TEXT NOT NULL,
			subject VARCHAR(500),
			content TEXT,
			data TEXT,
			channels TEXT,
			scheduled_at DATETIME,
			expires_at DATETIME,
			template_id VARCHAR(128),
			language VARCHAR(10),
			tags TEXT,
			metadata TEXT,
			status VARCHAR(20) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			sent_at DATETIME,
			delivered_at DATETIME,
			read_at DATETIME,
			clicked_at DATETIME,
			attempts INT DEFAULT 0,
			last_error TEXT,
			tracking_id VARCHAR(128),
			parent_id VARCHAR(128),
			thread_id VARCHAR(128),
			INDEX idx_type (type),
			INDEX idx_status (status),
			INDEX idx_created_at (created_at),
			INDEX idx_scheduled_at (scheduled_at),
			INDEX idx_tracking_id (tracking_id)
		)`,
		`CREATE TABLE IF NOT EXISTS notification_deliveries (
			id VARCHAR(128) PRIMARY KEY,
			notification_id VARCHAR(128) NOT NULL,
			recipient_id VARCHAR(128) NOT NULL,
			channel VARCHAR(50) NOT NULL,
			status VARCHAR(20) NOT NULL,
			delivered_at DATETIME,
			read_at DATETIME,
			clicked_at DATETIME,
			error_message TEXT,
			attempts INT DEFAULT 0,
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_notification_id (notification_id),
			INDEX idx_recipient_id (recipient_id),
			INDEX idx_channel (channel),
			INDEX idx_status (status),
			FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS notification_templates (
			id VARCHAR(128) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			category VARCHAR(100),
			language VARCHAR(10) NOT NULL,
			subject VARCHAR(500),
			content TEXT,
			html_content TEXT,
			variables TEXT,
			channels TEXT,
			priority VARCHAR(20),
			ttl_seconds INT,
			metadata TEXT,
			is_active BOOLEAN DEFAULT TRUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_type (type),
			INDEX idx_category (category),
			INDEX idx_language (language),
			INDEX idx_is_active (is_active)
		)`,
		`CREATE TABLE IF NOT EXISTS user_notification_preferences (
			user_id VARCHAR(128) PRIMARY KEY,
			channels TEXT,
			types TEXT,
			categories TEXT,
			frequency VARCHAR(50) DEFAULT 'immediate',
			quiet_hours TEXT,
			language VARCHAR(10),
			timezone VARCHAR(50),
			opted_out BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_opted_out (opted_out)
		)`,
		`CREATE TABLE IF NOT EXISTS notification_analytics (
			id VARCHAR(128) PRIMARY KEY,
			notification_id VARCHAR(128) NOT NULL,
			event_type VARCHAR(50) NOT NULL,
			channel VARCHAR(50),
			recipient_id VARCHAR(128),
			user_agent TEXT,
			ip_address VARCHAR(45),
			location TEXT,
			device_info TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			metadata TEXT,
			INDEX idx_notification_id (notification_id),
			INDEX idx_event_type (event_type),
			INDEX idx_timestamp (timestamp),
			FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE
		)`,
	}

	for _, query := range queries {
		if _, err := nm.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create notification table: %w", err)
		}
	}

	return nil
}

// RegisterChannel registers a notification channel
func (nm *NotificationManager) RegisterChannel(channel NotificationChannel) {
	nm.channels[channel.GetName()] = channel
}

// SendNotification sends a notification
func (nm *NotificationManager) SendNotification(ctx context.Context, notification *Notification) error {
	// Set defaults
	if notification.ID == "" {
		notification.ID = nm.generateNotificationID()
	}
	if notification.TrackingID == "" {
		notification.TrackingID = nm.generateTrackingID()
	}
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}
	notification.UpdatedAt = time.Now()
	notification.Status = StatusPending

	// Apply template if specified
	if notification.Template != "" {
		if err := nm.applyTemplate(notification); err != nil {
			return fmt.Errorf("failed to apply template: %w", err)
		}
	}

	// Validate notification
	if err := nm.validateNotification(notification); err != nil {
		return fmt.Errorf("notification validation failed: %w", err)
	}

	// Process recipients
	if err := nm.processRecipients(notification); err != nil {
		return fmt.Errorf("failed to process recipients: %w", err)
	}

	// Store notification
	if err := nm.storeNotification(notification); err != nil {
		return fmt.Errorf("failed to store notification: %w", err)
	}

	// Handle scheduling
	if notification.ScheduledAt != nil && notification.ScheduledAt.After(time.Now()) {
		return nm.scheduleNotification(notification)
	}

	// Send immediately
	return nm.sendNotificationNow(ctx, notification)
}

// SendBulkNotifications sends multiple notifications
func (nm *NotificationManager) SendBulkNotifications(ctx context.Context, notifications []*Notification) error {
	// Process in batches
	batchSize := nm.config.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	for i := 0; i < len(notifications); i += batchSize {
		end := i + batchSize
		if end > len(notifications) {
			end = len(notifications)
		}

		batch := notifications[i:end]
		for _, notification := range batch {
			if err := nm.SendNotification(ctx, notification); err != nil {
				// Log error but continue with other notifications
				nm.logError(fmt.Sprintf("Failed to send notification %s: %v", notification.ID, err))
			}
		}
	}

	return nil
}

// GetNotification retrieves a notification by ID
func (nm *NotificationManager) GetNotification(id string) (*Notification, error) {
	query := `
		SELECT id, type, category, priority, recipients, subject, content, data,
		       channels, scheduled_at, expires_at, template_id, language, tags,
		       metadata, status, created_at, updated_at, sent_at, delivered_at,
		       read_at, clicked_at, attempts, last_error, tracking_id, parent_id, thread_id
		FROM notifications WHERE id = ?
	`

	row := nm.db.QueryRow(query, id)
	return nm.scanNotification(row)
}

// GetUserNotifications retrieves notifications for a user
func (nm *NotificationManager) GetUserNotifications(userID string, limit, offset int) ([]*Notification, error) {
	query := `
		SELECT n.id, n.type, n.category, n.priority, n.recipients, n.subject, n.content, n.data,
		       n.channels, n.scheduled_at, n.expires_at, n.template_id, n.language, n.tags,
		       n.metadata, n.status, n.created_at, n.updated_at, n.sent_at, n.delivered_at,
		       n.read_at, n.clicked_at, n.attempts, n.last_error, n.tracking_id, n.parent_id, n.thread_id
		FROM notifications n
		WHERE JSON_EXTRACT(n.recipients, '$[*].id') LIKE ?
		ORDER BY n.created_at DESC
		LIMIT ? OFFSET ?
	`

	// Using parameterized query to prevent SQL injection
	userIDPattern := "%" + userID + "%"
	rows, err := nm.db.Query(query, userIDPattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]*Notification, 0)
	for rows.Next() {
		notification, err := nm.scanNotification(rows)
		if err != nil {
			continue
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// MarkAsRead marks a notification as read
func (nm *NotificationManager) MarkAsRead(notificationID, userID string) error {
	now := time.Now()

	// Update notification
	query := `UPDATE notifications SET read_at = ?, status = 'read', updated_at = ? WHERE id = ?`
	_, err := nm.db.Exec(query, now, now, notificationID)
	if err != nil {
		return err
	}

	// Track analytics
	nm.trackEvent(notificationID, "read", userID, nil)

	return nil
}

// MarkAsClicked marks a notification as clicked
func (nm *NotificationManager) MarkAsClicked(notificationID, userID string, metadata map[string]interface{}) error {
	now := time.Now()

	// Update notification
	query := `UPDATE notifications SET clicked_at = ?, status = 'clicked', updated_at = ? WHERE id = ?`
	_, err := nm.db.Exec(query, now, now, notificationID)
	if err != nil {
		return err
	}

	// Track analytics
	nm.trackEvent(notificationID, "clicked", userID, metadata)

	return nil
}

// GetUserPreferences retrieves user notification preferences
func (nm *NotificationManager) GetUserPreferences(userID string) (*UserPreference, error) {
	query := `
		SELECT user_id, channels, types, categories, frequency, quiet_hours,
		       language, timezone, opted_out, updated_at
		FROM user_notification_preferences WHERE user_id = ?
	`

	row := nm.db.QueryRow(query, userID)

	var pref UserPreference
	var channelsJSON, typesJSON, categoriesJSON, quietHoursJSON string

	err := row.Scan(
		&pref.UserID, &channelsJSON, &typesJSON, &categoriesJSON,
		&pref.Frequency, &quietHoursJSON, &pref.Language, &pref.Timezone,
		&pref.OptedOut, &pref.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return default preferences
			return nm.getDefaultUserPreferences(userID), nil
		}
		return nil, err
	}

	// Parse JSON fields
	json.Unmarshal([]byte(channelsJSON), &pref.Channels)
	json.Unmarshal([]byte(typesJSON), &pref.Types)
	json.Unmarshal([]byte(categoriesJSON), &pref.Categories)
	json.Unmarshal([]byte(quietHoursJSON), &pref.QuietHours)

	return &pref, nil
}

// UpdateUserPreferences updates user notification preferences
func (nm *NotificationManager) UpdateUserPreferences(userID string, preferences *UserPreference) error {
	channelsJSON, _ := json.Marshal(preferences.Channels)
	typesJSON, _ := json.Marshal(preferences.Types)
	categoriesJSON, _ := json.Marshal(preferences.Categories)
	quietHoursJSON, _ := json.Marshal(preferences.QuietHours)

	query := `
		INSERT INTO user_notification_preferences 
		(user_id, channels, types, categories, frequency, quiet_hours, language, timezone, opted_out)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		channels = VALUES(channels), types = VALUES(types), categories = VALUES(categories),
		frequency = VALUES(frequency), quiet_hours = VALUES(quiet_hours),
		language = VALUES(language), timezone = VALUES(timezone), opted_out = VALUES(opted_out),
		updated_at = NOW()
	`

	_, err := nm.db.Exec(query, userID, string(channelsJSON), string(typesJSON),
		string(categoriesJSON), preferences.Frequency, string(quietHoursJSON),
		preferences.Language, preferences.Timezone, preferences.OptedOut)

	return err
}

// GetNotificationStats retrieves notification statistics
func (nm *NotificationManager) GetNotificationStats(startDate, endDate time.Time) (*NotificationStats, error) {
	stats := &NotificationStats{
		ByType:     make(map[NotificationType]int),
		ByChannel:  make(map[string]int),
		ByPriority: make(map[NotificationPriority]int),
	}

	// Get basic counts
	query := `
		SELECT 
		    COUNT(*) as total_sent,
		    SUM(CASE WHEN status IN ('delivered', 'read', 'clicked') THEN 1 ELSE 0 END) as total_delivered,
		    SUM(CASE WHEN status IN ('read', 'clicked') THEN 1 ELSE 0 END) as total_read,
		    SUM(CASE WHEN status = 'clicked' THEN 1 ELSE 0 END) as total_clicked,
		    SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as total_failed
		FROM notifications 
		WHERE created_at BETWEEN ? AND ?
	`

	err := nm.db.QueryRow(query, startDate, endDate).Scan(
		&stats.TotalSent, &stats.TotalDelivered, &stats.TotalRead,
		&stats.TotalClicked, &stats.TotalFailed,
	)
	if err != nil {
		return nil, err
	}

	// Calculate rates
	if stats.TotalSent > 0 {
		stats.DeliveryRate = float64(stats.TotalDelivered) / float64(stats.TotalSent)
		stats.ReadRate = float64(stats.TotalRead) / float64(stats.TotalSent)
		stats.ClickRate = float64(stats.TotalClicked) / float64(stats.TotalSent)
	}

	// Get breakdown by type
	typeQuery := `
		SELECT type, COUNT(*) FROM notifications 
		WHERE created_at BETWEEN ? AND ?
		GROUP BY type
	`
	rows, err := nm.db.Query(typeQuery, startDate, endDate)
	if err == nil {
		for rows.Next() {
			var notType string
			var count int
			if err := rows.Scan(&notType, &count); err == nil {
				stats.ByType[NotificationType(notType)] = count
			}
		}
		rows.Close()
	}

	return stats, nil
}

// Helper methods

func (nm *NotificationManager) generateNotificationID() string {
	return fmt.Sprintf("notif_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond())
}

func (nm *NotificationManager) generateTrackingID() string {
	return fmt.Sprintf("track_%d", time.Now().UnixNano())
}

func (nm *NotificationManager) applyTemplate(notification *Notification) error {
	template, exists := nm.templates[notification.Template]
	if !exists {
		return fmt.Errorf("template not found: %s", notification.Template)
	}

	// Apply template values
	if notification.Subject == "" {
		notification.Subject = template.Subject
	}
	if notification.Content == "" {
		notification.Content = template.Content
	}
	if notification.Type == "" {
		notification.Type = template.Type
	}
	if notification.Priority == "" {
		notification.Priority = template.Priority
	}

	// Replace variables in subject and content
	notification.Subject = nm.replaceTemplateVariables(notification.Subject, notification.Data)
	notification.Content = nm.replaceTemplateVariables(notification.Content, notification.Data)

	return nil
}

func (nm *NotificationManager) replaceTemplateVariables(text string, data map[string]interface{}) string {
	result := text
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

func (nm *NotificationManager) validateNotification(notification *Notification) error {
	if notification.Type == "" {
		return fmt.Errorf("notification type is required")
	}
	if len(notification.Recipients) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	if notification.Subject == "" && notification.Content == "" {
		return fmt.Errorf("either subject or content is required")
	}
	return nil
}

func (nm *NotificationManager) processRecipients(notification *Notification) error {
	// Expand group recipients, apply user preferences, etc.
	// This is a simplified version
	return nil
}

func (nm *NotificationManager) storeNotification(notification *Notification) error {
	recipientsJSON, _ := json.Marshal(notification.Recipients)
	dataJSON, _ := json.Marshal(notification.Data)
	channelsJSON, _ := json.Marshal(notification.Channels)
	tagsJSON, _ := json.Marshal(notification.Tags)
	metadataJSON, _ := json.Marshal(notification.Metadata)

	query := `
		INSERT INTO notifications (
			id, type, category, priority, recipients, subject, content, data,
			channels, scheduled_at, expires_at, template_id, language, tags,
			metadata, status, created_at, updated_at, tracking_id, parent_id, thread_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := nm.db.Exec(query,
		notification.ID, notification.Type, notification.Category, notification.Priority,
		string(recipientsJSON), notification.Subject, notification.Content, string(dataJSON),
		string(channelsJSON), notification.ScheduledAt, notification.ExpiresAt,
		notification.Template, notification.Language, string(tagsJSON),
		string(metadataJSON), notification.Status, notification.CreatedAt,
		notification.UpdatedAt, notification.TrackingID, notification.ParentID, notification.ThreadID,
	)

	return err
}

func (nm *NotificationManager) scheduleNotification(notification *Notification) error {
	// Implementation would add to scheduler
	// For now, just update status
	notification.Status = StatusQueued
	return nm.updateNotificationStatus(notification.ID, StatusQueued)
}

func (nm *NotificationManager) sendNotificationNow(ctx context.Context, notification *Notification) error {
	notification.Status = StatusSending
	nm.updateNotificationStatus(notification.ID, StatusSending)

	// Send through configured channels
	for _, channelName := range notification.Channels {
		channel, exists := nm.channels[channelName]
		if !exists || !channel.IsEnabled() {
			continue
		}

		for _, recipient := range notification.Recipients {
			if err := channel.Send(ctx, notification, &recipient); err != nil {
				nm.logError(fmt.Sprintf("Failed to send via %s to %s: %v", channelName, recipient.Address, err))
				continue
			}
		}
	}

	// Update status
	notification.Status = StatusSent
	notification.SentAt = &time.Time{}
	*notification.SentAt = time.Now()

	return nm.updateNotificationStatus(notification.ID, StatusSent)
}

func (nm *NotificationManager) updateNotificationStatus(id string, status NotificationStatus) error {
	query := `UPDATE notifications SET status = ?, updated_at = NOW() WHERE id = ?`
	_, err := nm.db.Exec(query, status, id)
	return err
}

func (nm *NotificationManager) scanNotification(scanner interface{}) (*Notification, error) {
	// This would implement scanning from SQL row to Notification struct
	// Simplified implementation
	return &Notification{}, nil
}

func (nm *NotificationManager) trackEvent(notificationID, eventType, userID string, metadata map[string]interface{}) {
	metadataJSON, _ := json.Marshal(metadata)

	query := `
		INSERT INTO notification_analytics (id, notification_id, event_type, recipient_id, metadata)
		VALUES (?, ?, ?, ?, ?)
	`

	eventID := fmt.Sprintf("event_%d", time.Now().UnixNano())
	nm.db.Exec(query, eventID, notificationID, eventType, userID, string(metadataJSON))
}

func (nm *NotificationManager) getDefaultUserPreferences(userID string) *UserPreference {
	return &UserPreference{
		UserID:     userID,
		Channels:   map[string]bool{"email": true, "push": true},
		Types:      make(map[NotificationType]bool),
		Categories: make(map[string]bool),
		Frequency:  "immediate",
		QuietHours: QuietHours{Enabled: false},
		Language:   "tr",
		Timezone:   "Europe/Istanbul",
		OptedOut:   false,
		UpdatedAt:  time.Now(),
	}
}

func (nm *NotificationManager) loadTemplates() {
	// Implementation would load templates from database
}

func (nm *NotificationManager) initializeChannels() {
	// Implementation would initialize configured channels
}

func (nm *NotificationManager) startWorkers() {
	// Implementation would start background workers for processing queue
}

func (nm *NotificationManager) logError(message string) {
	// Implementation would log errors
	fmt.Printf("Notification Error: %s\n", message)
}
