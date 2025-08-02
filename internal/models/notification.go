package models

import (
	"time"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Notification represents a notification
type Notification struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	
	// Recipient Information
	RecipientID   uint      `json:"recipient_id" gorm:"index;not null"`
	RecipientType RecipientType `json:"recipient_type" gorm:"not null"`
	
	// Notification Content
	Title       string    `json:"title" gorm:"size:200;not null" validate:"required,max=200"`
	Message     string    `json:"message" gorm:"type:text;not null" validate:"required"`
	Type        NotificationType `json:"type" gorm:"not null"`
	Category    NotificationCategory `json:"category" gorm:"not null"`
	Priority    NotificationPriority `json:"priority" gorm:"default:'normal'"`
	
	// Channels
	Channels    []NotificationChannel `json:"channels" gorm:"type:json"`
	
	// Status and Tracking
	Status      NotificationStatus `json:"status" gorm:"default:'pending'"`
	ReadAt      *time.Time `json:"read_at"`
	ClickedAt   *time.Time `json:"clicked_at"`
	DismissedAt *time.Time `json:"dismissed_at"`
	
	// Delivery Information
	SentAt      *time.Time `json:"sent_at"`
	DeliveredAt *time.Time `json:"delivered_at"`
	FailedAt    *time.Time `json:"failed_at"`
	RetryCount  int        `json:"retry_count" gorm:"default:0"`
	MaxRetries  int        `json:"max_retries" gorm:"default:3"`
	
	// Scheduling
	ScheduledAt *time.Time `json:"scheduled_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	
	// Action and Navigation
	ActionURL   string     `json:"action_url" gorm:"size:500"`
	ActionText  string     `json:"action_text" gorm:"size:100"`
	DeepLink    string     `json:"deep_link" gorm:"size:500"`
	
	// Rich Content
	ImageURL    string     `json:"image_url" gorm:"size:500"`
	IconURL     string     `json:"icon_url" gorm:"size:500"`
	BadgeCount  *int       `json:"badge_count"`
	
	// Grouping and Threading
	GroupKey    string     `json:"group_key" gorm:"size:100;index"`
	ThreadID    string     `json:"thread_id" gorm:"size:100;index"`
	ParentID    *uint      `json:"parent_id" gorm:"index"`
	Parent      *Notification `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	
	// Context and Metadata
	EntityType  string     `json:"entity_type" gorm:"size:50"`  // order, product, user, etc.
	EntityID    *uint      `json:"entity_id" gorm:"index"`
	Metadata    NotificationMetadata `json:"metadata" gorm:"type:json"`
	
	// Personalization
	Language    string     `json:"language" gorm:"size:5;default:'tr'"`
	Timezone    string     `json:"timezone" gorm:"size:50;default:'Europe/Istanbul'"`
	
	// Analytics
	ViewCount   int        `json:"view_count" gorm:"default:0"`
	ClickCount  int        `json:"click_count" gorm:"default:0"`
	
	// Error Information
	ErrorCode   string     `json:"error_code" gorm:"size:100"`
	ErrorMessage string    `json:"error_message" gorm:"size:500"`
	
	// Delivery Results per Channel
	DeliveryResults []NotificationDeliveryResult `json:"delivery_results,omitempty" gorm:"foreignKey:NotificationID"`
	
	// Timestamps
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// RecipientType represents the type of notification recipient
type RecipientType string

const (
	RecipientTypeUser     RecipientType = "user"
	RecipientTypeCustomer RecipientType = "customer"
	RecipientTypeVendor   RecipientType = "vendor"
	RecipientTypeAdmin    RecipientType = "admin"
	RecipientTypeSystem   RecipientType = "system"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeInfo      NotificationType = "info"
	NotificationTypeSuccess   NotificationType = "success"
	NotificationTypeWarning   NotificationType = "warning"
	NotificationTypeError     NotificationType = "error"
	NotificationTypeMarketing NotificationType = "marketing"
	NotificationTypeSystem    NotificationType = "system"
	NotificationTypeTransactional NotificationType = "transactional"
)

// NotificationCategory represents notification categories
type NotificationCategory string

const (
	NotificationCategoryOrder       NotificationCategory = "order"
	NotificationCategoryPayment     NotificationCategory = "payment"
	NotificationCategoryShipping    NotificationCategory = "shipping"
	NotificationCategoryProduct     NotificationCategory = "product"
	NotificationCategoryAccount     NotificationCategory = "account"
	NotificationCategoryPromotion   NotificationCategory = "promotion"
	NotificationCategoryReview      NotificationCategory = "review"
	NotificationCategorySupport     NotificationCategory = "support"
	NotificationCategorySystem      NotificationCategory = "system"
	NotificationCategoryAI          NotificationCategory = "ai"
	NotificationCategoryMarketplace NotificationCategory = "marketplace"
)

// NotificationPriority represents notification priority levels
type NotificationPriority string

const (
	NotificationPriorityLow      NotificationPriority = "low"
	NotificationPriorityNormal   NotificationPriority = "normal"
	NotificationPriorityHigh     NotificationPriority = "high"
	NotificationPriorityCritical NotificationPriority = "critical"
	NotificationPriorityUrgent   NotificationPriority = "urgent"
)

// NotificationChannel represents delivery channels
type NotificationChannel string

const (
	NotificationChannelInApp    NotificationChannel = "in_app"
	NotificationChannelEmail    NotificationChannel = "email"
	NotificationChannelSMS      NotificationChannel = "sms"
	NotificationChannelPush     NotificationChannel = "push"
	NotificationChannelWebPush  NotificationChannel = "web_push"
	NotificationChannelWhatsApp NotificationChannel = "whatsapp"
	NotificationChannelSlack    NotificationChannel = "slack"
	NotificationChannelDiscord  NotificationChannel = "discord"
	NotificationChannelTelegram NotificationChannel = "telegram"
)

// NotificationStatus represents notification status
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusScheduled NotificationStatus = "scheduled"
	NotificationStatusSending   NotificationStatus = "sending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusRead      NotificationStatus = "read"
	NotificationStatusClicked   NotificationStatus = "clicked"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusExpired   NotificationStatus = "expired"
	NotificationStatusCancelled NotificationStatus = "cancelled"
)

// NotificationMetadata holds additional notification data
type NotificationMetadata struct {
	Campaign        string                 `json:"campaign,omitempty"`
	Source          string                 `json:"source,omitempty"`
	Medium          string                 `json:"medium,omitempty"`
	UtmParams       map[string]string      `json:"utm_params,omitempty"`
	CustomFields    map[string]interface{} `json:"custom_fields,omitempty"`
	TemplateID      string                 `json:"template_id,omitempty"`
	TemplateVersion string                 `json:"template_version,omitempty"`
	Variables       map[string]interface{} `json:"variables,omitempty"`
	ABTestGroup     string                 `json:"ab_test_group,omitempty"`
}

// NotificationDeliveryResult represents delivery result for each channel
type NotificationDeliveryResult struct {
	ID             uint                `json:"id" gorm:"primaryKey"`
	NotificationID uint                `json:"notification_id" gorm:"index;not null"`
	Notification   Notification        `json:"notification" gorm:"foreignKey:NotificationID"`
	
	Channel        NotificationChannel `json:"channel" gorm:"not null"`
	Status         DeliveryStatus      `json:"status" gorm:"default:'pending'"`
	
	// Provider Information
	Provider       string              `json:"provider" gorm:"size:100"`
	ProviderID     string              `json:"provider_id" gorm:"size:255"`
	ProviderResponse map[string]interface{} `json:"provider_response" gorm:"type:json"`
	
	// Delivery Details
	SentAt         *time.Time          `json:"sent_at"`
	DeliveredAt    *time.Time          `json:"delivered_at"`
	FailedAt       *time.Time          `json:"failed_at"`
	RetryCount     int                 `json:"retry_count" gorm:"default:0"`
	
	// Error Information
	ErrorCode      string              `json:"error_code" gorm:"size:100"`
	ErrorMessage   string              `json:"error_message" gorm:"size:500"`
	
	// Channel-specific data
	ChannelData    ChannelData         `json:"channel_data" gorm:"type:json"`
	
	// Timestamps
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

// DeliveryStatus represents delivery status for each channel
type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "pending"
	DeliveryStatusSending   DeliveryStatus = "sending"
	DeliveryStatusSent      DeliveryStatus = "sent"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusFailed    DeliveryStatus = "failed"
	DeliveryStatusBounced   DeliveryStatus = "bounced"
	DeliveryStatusRejected  DeliveryStatus = "rejected"
)

// ChannelData holds channel-specific delivery data
type ChannelData struct {
	// Email specific
	Subject         string `json:"subject,omitempty"`
	FromEmail       string `json:"from_email,omitempty"`
	ToEmail         string `json:"to_email,omitempty"`
	MessageID       string `json:"message_id,omitempty"`
	
	// SMS specific
	FromPhone       string `json:"from_phone,omitempty"`
	ToPhone         string `json:"to_phone,omitempty"`
	SMSProvider     string `json:"sms_provider,omitempty"`
	
	// Push specific
	DeviceToken     string `json:"device_token,omitempty"`
	Platform        string `json:"platform,omitempty"`
	AppID           string `json:"app_id,omitempty"`
	
	// WhatsApp specific
	WhatsAppNumber  string `json:"whatsapp_number,omitempty"`
	TemplateID      string `json:"template_id,omitempty"`
	
	// Custom fields
	CustomData      map[string]interface{} `json:"custom_data,omitempty"`
}

// NotificationTemplate represents notification templates
type NotificationTemplate struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	
	// Template Information
	Name        string    `json:"name" gorm:"size:200;not null" validate:"required,max=200"`
	Code        string    `json:"code" gorm:"size:100;unique;not null" validate:"required"`
	Description string    `json:"description" gorm:"type:text"`
	Version     string    `json:"version" gorm:"size:20;default:'1.0'"`
	
	// Template Content
	Type        NotificationType     `json:"type" gorm:"not null"`
	Category    NotificationCategory `json:"category" gorm:"not null"`
	Priority    NotificationPriority `json:"priority" gorm:"default:'normal'"`
	
	// Multi-channel Templates
	Templates   ChannelTemplates     `json:"templates" gorm:"type:json"`
	
	// Configuration
	IsActive    bool                 `json:"is_active" gorm:"default:true"`
	IsDefault   bool                 `json:"is_default" gorm:"default:false"`
	
	// Targeting
	TargetAudience AudienceFilter    `json:"target_audience" gorm:"type:json"`
	
	// Scheduling Rules
	SchedulingRules SchedulingRules  `json:"scheduling_rules" gorm:"type:json"`
	
	// A/B Testing
	ABTestConfig ABTestConfig        `json:"ab_test_config" gorm:"type:json"`
	
	// Analytics
	SentCount     int               `json:"sent_count" gorm:"default:0"`
	DeliveredCount int              `json:"delivered_count" gorm:"default:0"`
	OpenedCount   int               `json:"opened_count" gorm:"default:0"`
	ClickedCount  int               `json:"clicked_count" gorm:"default:0"`
	
	// Vendor/Admin
	VendorID    *uint               `json:"vendor_id" gorm:"index"`
	Vendor      *Vendor             `json:"vendor,omitempty" gorm:"foreignKey:VendorID"`
	CreatedBy   uint                `json:"created_by" gorm:"index;not null"`
	
	// Timestamps
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	DeletedAt   *time.Time          `json:"deleted_at,omitempty" gorm:"index"`
}

// ChannelTemplates holds templates for different channels
type ChannelTemplates struct {
	InApp    *InAppTemplate    `json:"in_app,omitempty"`
	Email    *EmailTemplate    `json:"email,omitempty"`
	SMS      *SMSTemplate      `json:"sms,omitempty"`
	Push     *PushTemplate     `json:"push,omitempty"`
	WhatsApp *WhatsAppTemplate `json:"whatsapp,omitempty"`
}

// InAppTemplate represents in-app notification template
type InAppTemplate struct {
	Title     string `json:"title"`
	Message   string `json:"message"`
	ImageURL  string `json:"image_url,omitempty"`
	ActionURL string `json:"action_url,omitempty"`
	ActionText string `json:"action_text,omitempty"`
}

// EmailTemplate represents email notification template
type EmailTemplate struct {
	Subject     string `json:"subject"`
	HTMLBody    string `json:"html_body"`
	TextBody    string `json:"text_body"`
	FromName    string `json:"from_name"`
	FromEmail   string `json:"from_email"`
	ReplyTo     string `json:"reply_to,omitempty"`
	Attachments []EmailAttachment `json:"attachments,omitempty"`
}

// SMSTemplate represents SMS notification template
type SMSTemplate struct {
	Message   string `json:"message"`
	FromPhone string `json:"from_phone,omitempty"`
}

// PushTemplate represents push notification template
type PushTemplate struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	ImageURL string `json:"image_url,omitempty"`
	IconURL  string `json:"icon_url,omitempty"`
	Sound    string `json:"sound,omitempty"`
	Badge    *int   `json:"badge,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// WhatsAppTemplate represents WhatsApp notification template
type WhatsAppTemplate struct {
	TemplateID string                 `json:"template_id"`
	Language   string                 `json:"language"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// EmailAttachment represents email attachment
type EmailAttachment struct {
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Content     string `json:"content"` // base64 encoded
	Size        int64  `json:"size"`
}

// AudienceFilter represents audience targeting rules
type AudienceFilter struct {
	CustomerTiers    []CustomerTier `json:"customer_tiers,omitempty"`
	Countries        []string       `json:"countries,omitempty"`
	Languages        []string       `json:"languages,omitempty"`
	MinOrderCount    *int           `json:"min_order_count,omitempty"`
	MaxOrderCount    *int           `json:"max_order_count,omitempty"`
	MinTotalSpent    *float64       `json:"min_total_spent,omitempty"`
	MaxTotalSpent    *float64       `json:"max_total_spent,omitempty"`
	RegistrationDays *int           `json:"registration_days,omitempty"`
	CustomFilters    map[string]interface{} `json:"custom_filters,omitempty"`
}

// SchedulingRules represents notification scheduling rules
type SchedulingRules struct {
	TimeZone        string    `json:"timezone,omitempty"`
	SendHours       []int     `json:"send_hours,omitempty"`        // 0-23
	SendDays        []int     `json:"send_days,omitempty"`         // 0-6 (Sunday-Saturday)
	NoSendBefore    string    `json:"no_send_before,omitempty"`    // HH:MM
	NoSendAfter     string    `json:"no_send_after,omitempty"`     // HH:MM
	RespectDND      bool      `json:"respect_dnd"`                 // Do Not Disturb
	MaxPerDay       *int      `json:"max_per_day,omitempty"`
	MaxPerWeek      *int      `json:"max_per_week,omitempty"`
	MinInterval     *int      `json:"min_interval,omitempty"`      // minutes between notifications
}

// ABTestConfig represents A/B testing configuration
type ABTestConfig struct {
	IsEnabled       bool                   `json:"is_enabled"`
	TestName        string                 `json:"test_name,omitempty"`
	VariantA        map[string]interface{} `json:"variant_a,omitempty"`
	VariantB        map[string]interface{} `json:"variant_b,omitempty"`
	TrafficSplit    float64                `json:"traffic_split,omitempty"` // 0.0-1.0
	TestDuration    *int                   `json:"test_duration,omitempty"` // days
	WinningVariant  string                 `json:"winning_variant,omitempty"`
}

// Implement driver.Valuer interfaces
func (nm NotificationMetadata) Value() (driver.Value, error) {
	return json.Marshal(nm)
}

func (nm *NotificationMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, nm)
}

func (cd ChannelData) Value() (driver.Value, error) {
	return json.Marshal(cd)
}

func (cd *ChannelData) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, cd)
}

func (ct ChannelTemplates) Value() (driver.Value, error) {
	return json.Marshal(ct)
}

func (ct *ChannelTemplates) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, ct)
}

func (af AudienceFilter) Value() (driver.Value, error) {
	return json.Marshal(af)
}

func (af *AudienceFilter) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, af)
}

func (sr SchedulingRules) Value() (driver.Value, error) {
	return json.Marshal(sr)
}

func (sr *SchedulingRules) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, sr)
}

func (ab ABTestConfig) Value() (driver.Value, error) {
	return json.Marshal(ab)
}

func (ab *ABTestConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, ab)
}

// TableName methods
func (Notification) TableName() string {
	return "notifications"
}

func (NotificationDeliveryResult) TableName() string {
	return "notification_delivery_results"
}

func (NotificationTemplate) TableName() string {
	return "notification_templates"
}

// Notification methods
func (n *Notification) IsRead() bool {
	return n.ReadAt != nil
}

func (n *Notification) IsClicked() bool {
	return n.ClickedAt != nil
}

func (n *Notification) IsExpired() bool {
	return n.ExpiresAt != nil && time.Now().After(*n.ExpiresAt)
}

func (n *Notification) ShouldRetry() bool {
	return n.Status == NotificationStatusFailed && n.RetryCount < n.MaxRetries
}

func (n *Notification) CanBeSent() bool {
	now := time.Now()
	return n.Status == NotificationStatusPending &&
		   (n.ScheduledAt == nil || n.ScheduledAt.Before(now)) &&
		   !n.IsExpired()
}

func (n *Notification) MarkAsRead() {
	if n.ReadAt == nil {
		now := time.Now()
		n.ReadAt = &now
		n.Status = NotificationStatusRead
	}
}

func (n *Notification) MarkAsClicked() {
	if n.ClickedAt == nil {
		now := time.Now()
		n.ClickedAt = &now
		n.Status = NotificationStatusClicked
		n.ClickCount++
	}
}

func (n *Notification) GetDeliveryResult(channel NotificationChannel) *NotificationDeliveryResult {
	for _, result := range n.DeliveryResults {
		if result.Channel == channel {
			return &result
		}
	}
	return nil
}

// NotificationTemplate methods
func (nt *NotificationTemplate) HasChannel(channel NotificationChannel) bool {
	switch channel {
	case NotificationChannelInApp:
		return nt.Templates.InApp != nil
	case NotificationChannelEmail:
		return nt.Templates.Email != nil
	case NotificationChannelSMS:
		return nt.Templates.SMS != nil
	case NotificationChannelPush:
		return nt.Templates.Push != nil
	case NotificationChannelWhatsApp:
		return nt.Templates.WhatsApp != nil
	default:
		return false
	}
}

func (nt *NotificationTemplate) GetDeliveryRate() float64 {
	if nt.SentCount == 0 {
		return 0
	}
	return float64(nt.DeliveredCount) / float64(nt.SentCount) * 100
}

func (nt *NotificationTemplate) GetOpenRate() float64 {
	if nt.DeliveredCount == 0 {
		return 0
	}
	return float64(nt.OpenedCount) / float64(nt.DeliveredCount) * 100
}

func (nt *NotificationTemplate) GetClickRate() float64 {
	if nt.OpenedCount == 0 {
		return 0
	}
	return float64(nt.ClickedCount) / float64(nt.OpenedCount) * 100
}