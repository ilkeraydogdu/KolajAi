package integrations

import (
	"context"
	"time"
)

// IntegrationType represents the type of integration
type IntegrationType string

const (
	IntegrationTypePayment      IntegrationType = "payment"
	IntegrationTypeMarketplace  IntegrationType = "marketplace"
	IntegrationTypeShipping     IntegrationType = "shipping"
	IntegrationTypeAccounting   IntegrationType = "accounting"
	IntegrationTypeCommunication IntegrationType = "communication"
	IntegrationTypeAnalytics    IntegrationType = "analytics"
	IntegrationTypeAI           IntegrationType = "ai"
	IntegrationTypeStorage      IntegrationType = "storage"
	IntegrationTypeAuth         IntegrationType = "auth"
)

// IntegrationStatus represents the status of an integration
type IntegrationStatus string

const (
	IntegrationStatusActive   IntegrationStatus = "active"
	IntegrationStatusInactive IntegrationStatus = "inactive"
	IntegrationStatusError    IntegrationStatus = "error"
	IntegrationStatusPending  IntegrationStatus = "pending"
)

// Integration represents a base integration
type Integration struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        IntegrationType        `json:"type"`
	Provider    string                 `json:"provider"`
	Version     string                 `json:"version"`
	Status      IntegrationStatus      `json:"status"`
	Config      map[string]interface{} `json:"config"`
	Credentials Credentials            `json:"-"` // Never expose in JSON
	Metadata    IntegrationMetadata    `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Credentials represents encrypted credentials for an integration
type Credentials struct {
	APIKey      string            `json:"-"`
	APISecret   string            `json:"-"`
	AccessToken string            `json:"-"`
	RefreshToken string           `json:"-"`
	Extra       map[string]string `json:"-"`
}

// IntegrationMetadata contains metadata about the integration
type IntegrationMetadata struct {
	LastHealthCheck   time.Time         `json:"last_health_check"`
	LastSync          time.Time         `json:"last_sync"`
	ErrorCount        int               `json:"error_count"`
	SuccessCount      int               `json:"success_count"`
	AverageResponseTime time.Duration   `json:"average_response_time"`
	RateLimit         RateLimitInfo     `json:"rate_limit"`
	Capabilities      []string          `json:"capabilities"`
}

// RateLimitInfo contains rate limiting information
type RateLimitInfo struct {
	RequestsPerMinute int       `json:"requests_per_minute"`
	RequestsRemaining int       `json:"requests_remaining"`
	ResetsAt          time.Time `json:"resets_at"`
}

// IntegrationProvider is the base interface for all integration providers
type IntegrationProvider interface {
	// Initialize sets up the integration with credentials and config
	Initialize(ctx context.Context, credentials Credentials, config map[string]interface{}) error
	
	// HealthCheck verifies the integration is working
	HealthCheck(ctx context.Context) error
	
	// GetCapabilities returns the capabilities of this integration
	GetCapabilities() []string
	
	// GetRateLimit returns current rate limit information
	GetRateLimit() RateLimitInfo
	
	// Close cleans up any resources
	Close() error
}

// WebhookHandler handles incoming webhooks from integrations
type WebhookHandler interface {
	// ValidateWebhook validates the webhook signature/authenticity
	ValidateWebhook(headers map[string]string, body []byte) error
	
	// ProcessWebhook processes the webhook payload
	ProcessWebhook(ctx context.Context, event WebhookEvent) error
}

// WebhookEvent represents an incoming webhook event
type WebhookEvent struct {
	ID            string                 `json:"id"`
	IntegrationID string                 `json:"integration_id"`
	Type          string                 `json:"type"`
	Timestamp     time.Time              `json:"timestamp"`
	Headers       map[string]string      `json:"headers"`
	Payload       map[string]interface{} `json:"payload"`
	Signature     string                 `json:"signature"`
}

// RetryPolicy defines retry behavior for failed operations
type RetryPolicy struct {
	MaxAttempts     int           `json:"max_attempts"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []string      `json:"retryable_errors"`
}

// CircuitBreakerConfig defines circuit breaker configuration
type CircuitBreakerConfig struct {
	FailureThreshold   int           `json:"failure_threshold"`
	SuccessThreshold   int           `json:"success_threshold"`
	Timeout            time.Duration `json:"timeout"`
	HalfOpenMaxCalls   int           `json:"half_open_max_calls"`
}

// IntegrationError represents an error from an integration
type IntegrationError struct {
	Code       string    `json:"code"`
	Message    string    `json:"message"`
	Provider   string    `json:"provider"`
	Retryable  bool      `json:"retryable"`
	Timestamp  time.Time `json:"timestamp"`
	RequestID  string    `json:"request_id"`
	StatusCode int       `json:"status_code"`
}

func (e *IntegrationError) Error() string {
	return e.Message
}

// IntegrationRequest represents a request to an integration
type IntegrationRequest struct {
	ID        string                 `json:"id"`
	Method    string                 `json:"method"`
	Endpoint  string                 `json:"endpoint"`
	Headers   map[string]string      `json:"headers"`
	Body      interface{}            `json:"body"`
	Timeout   time.Duration          `json:"timeout"`
	Retries   int                    `json:"retries"`
}

// IntegrationResponse represents a response from an integration
type IntegrationResponse struct {
	ID         string                 `json:"id"`
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       interface{}            `json:"body"`
	Duration   time.Duration          `json:"duration"`
	Error      *IntegrationError      `json:"error,omitempty"`
}

// IntegrationLogger defines logging interface for integrations
type IntegrationLogger interface {
	LogRequest(integration string, request IntegrationRequest)
	LogResponse(integration string, response IntegrationResponse)
	LogError(integration string, err error)
	LogWebhook(integration string, event WebhookEvent)
}

// IntegrationMetrics defines metrics collection interface
type IntegrationMetrics interface {
	RecordRequest(integration string, method string, duration time.Duration, success bool)
	RecordWebhook(integration string, eventType string, success bool)
	RecordError(integration string, errorCode string)
	GetMetrics(integration string) map[string]interface{}
}

// IntegrationCache defines caching interface for integrations
type IntegrationCache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
	Clear(pattern string) error
}

// IntegrationEventBus defines event bus for integration events
type IntegrationEventBus interface {
	Publish(event IntegrationEvent) error
	Subscribe(eventType string, handler func(IntegrationEvent)) error
	Unsubscribe(eventType string) error
}

// IntegrationEvent represents an event in the integration system
type IntegrationEvent struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	IntegrationID string                 `json:"integration_id"`
	Timestamp     time.Time              `json:"timestamp"`
	Data          map[string]interface{} `json:"data"`
}

// Default retry policy
var DefaultRetryPolicy = RetryPolicy{
	MaxAttempts:   3,
	InitialDelay:  1 * time.Second,
	MaxDelay:      30 * time.Second,
	BackoffFactor: 2.0,
	RetryableErrors: []string{
		"TIMEOUT",
		"RATE_LIMIT",
		"TEMPORARY_FAILURE",
		"SERVICE_UNAVAILABLE",
	},
}

// Default circuit breaker config
var DefaultCircuitBreakerConfig = CircuitBreakerConfig{
	FailureThreshold: 5,
	SuccessThreshold: 2,
	Timeout:          60 * time.Second,
	HalfOpenMaxCalls: 3,
}