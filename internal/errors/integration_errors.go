package errors

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// Authentication errors
	ErrorCodeAuthenticationFailed ErrorCode = "AUTHENTICATION_FAILED"
	ErrorCodeCredentialsExpired   ErrorCode = "CREDENTIALS_EXPIRED"
	ErrorCodeInvalidCredentials   ErrorCode = "INVALID_CREDENTIALS"
	ErrorCodeAccessDenied         ErrorCode = "ACCESS_DENIED"

	// Network errors
	ErrorCodeNetworkTimeout       ErrorCode = "NETWORK_TIMEOUT"
	ErrorCodeConnectionFailed     ErrorCode = "CONNECTION_FAILED"
	ErrorCodeServiceUnavailable   ErrorCode = "SERVICE_UNAVAILABLE"
	ErrorCodeTooManyRequests      ErrorCode = "TOO_MANY_REQUESTS"

	// Validation errors
	ErrorCodeValidationFailed     ErrorCode = "VALIDATION_FAILED"
	ErrorCodeInvalidInput         ErrorCode = "INVALID_INPUT"
	ErrorCodeMissingRequiredField ErrorCode = "MISSING_REQUIRED_FIELD"
	ErrorCodeInvalidFormat        ErrorCode = "INVALID_FORMAT"

	// Business logic errors
	ErrorCodeProductNotFound      ErrorCode = "PRODUCT_NOT_FOUND"
	ErrorCodeOrderNotFound        ErrorCode = "ORDER_NOT_FOUND"
	ErrorCodeInsufficientStock    ErrorCode = "INSUFFICIENT_STOCK"
	ErrorCodePriceChanged         ErrorCode = "PRICE_CHANGED"
	ErrorCodeCategoryNotFound     ErrorCode = "CATEGORY_NOT_FOUND"

	// API errors
	ErrorCodeAPIError             ErrorCode = "API_ERROR"
	ErrorCodeInvalidAPIVersion    ErrorCode = "INVALID_API_VERSION"
	ErrorCodeQuotaExceeded        ErrorCode = "QUOTA_EXCEEDED"
	ErrorCodeRateLimitExceeded    ErrorCode = "RATE_LIMIT_EXCEEDED"

	// Internal errors
	ErrorCodeInternalError        ErrorCode = "INTERNAL_ERROR"
	ErrorCodeConfigurationError   ErrorCode = "CONFIGURATION_ERROR"
	ErrorCodeDatabaseError        ErrorCode = "DATABASE_ERROR"
	ErrorCodeCacheError           ErrorCode = "CACHE_ERROR"

	// Integration specific errors
	ErrorCodeProviderError        ErrorCode = "PROVIDER_ERROR"
	ErrorCodeMappingError         ErrorCode = "MAPPING_ERROR"
	ErrorCodeSyncError            ErrorCode = "SYNC_ERROR"
	ErrorCodeWebhookError         ErrorCode = "WEBHOOK_ERROR"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

const (
	SeverityCritical ErrorSeverity = "critical"
	SeverityHigh     ErrorSeverity = "high"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityLow      ErrorSeverity = "low"
	SeverityInfo     ErrorSeverity = "info"
)

// IntegrationError represents a standardized integration error
type IntegrationError struct {
	Code         ErrorCode              `json:"code"`
	Message      string                 `json:"message"`
	Provider     string                 `json:"provider"`
	Operation    string                 `json:"operation"`
	Severity     ErrorSeverity          `json:"severity"`
	Retryable    bool                   `json:"retryable"`
	RetryAfter   *time.Duration         `json:"retry_after,omitempty"`
	StatusCode   int                    `json:"status_code,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	TraceID      string                 `json:"trace_id,omitempty"`
	RequestID    string                 `json:"request_id,omitempty"`
	StackTrace   string                 `json:"stack_trace,omitempty"`
	Cause        error                  `json:"-"` // Original error, not serialized
	Metadata     map[string]string      `json:"metadata,omitempty"`
}

// Error implements the error interface
func (e *IntegrationError) Error() string {
	if e.Provider != "" && e.Operation != "" {
		return fmt.Sprintf("[%s:%s] %s: %s", e.Provider, e.Operation, e.Code, e.Message)
	} else if e.Provider != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Provider, e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// String returns a string representation of the error
func (e *IntegrationError) String() string {
	return e.Error()
}

// JSON returns the error as JSON
func (e *IntegrationError) JSON() ([]byte, error) {
	return json.Marshal(e)
}

// Is checks if the error matches the given error code
func (e *IntegrationError) Is(code ErrorCode) bool {
	return e.Code == code
}

// IsRetryable returns whether the error is retryable
func (e *IntegrationError) IsRetryable() bool {
	return e.Retryable
}

// GetSeverity returns the error severity
func (e *IntegrationError) GetSeverity() ErrorSeverity {
	return e.Severity
}

// WithContext adds context to the error
func (e *IntegrationError) WithContext(key string, value interface{}) *IntegrationError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithMetadata adds metadata to the error
func (e *IntegrationError) WithMetadata(key, value string) *IntegrationError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
	return e
}

// WithCause sets the underlying cause of the error
func (e *IntegrationError) WithCause(cause error) *IntegrationError {
	e.Cause = cause
	return e
}

// WithStackTrace adds stack trace to the error
func (e *IntegrationError) WithStackTrace() *IntegrationError {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	e.StackTrace = string(buf[:n])
	return e
}

// Unwrap returns the underlying cause of the error
func (e *IntegrationError) Unwrap() error {
	return e.Cause
}

// NewIntegrationError creates a new integration error
func NewIntegrationError(code ErrorCode, message, provider string) *IntegrationError {
	return &IntegrationError{
		Code:      code,
		Message:   message,
		Provider:  provider,
		Severity:  SeverityMedium,
		Retryable: false,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
		Metadata:  make(map[string]string),
	}
}

// NewRetryableError creates a new retryable integration error
func NewRetryableError(code ErrorCode, message, provider string, retryAfter time.Duration) *IntegrationError {
	err := NewIntegrationError(code, message, provider)
	err.Retryable = true
	err.RetryAfter = &retryAfter
	return err
}

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(provider, message string) *IntegrationError {
	return &IntegrationError{
		Code:      ErrorCodeAuthenticationFailed,
		Message:   message,
		Provider:  provider,
		Severity:  SeverityHigh,
		Retryable: false,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
		Metadata:  make(map[string]string),
	}
}

// NewValidationError creates a new validation error
func NewValidationError(provider, message string, context map[string]interface{}) *IntegrationError {
	return &IntegrationError{
		Code:      ErrorCodeValidationFailed,
		Message:   message,
		Provider:  provider,
		Severity:  SeverityMedium,
		Retryable: false,
		Context:   context,
		Timestamp: time.Now(),
		Metadata:  make(map[string]string),
	}
}

// NewNetworkError creates a new network error
func NewNetworkError(provider, message string, retryable bool) *IntegrationError {
	err := &IntegrationError{
		Code:      ErrorCodeConnectionFailed,
		Message:   message,
		Provider:  provider,
		Severity:  SeverityHigh,
		Retryable: retryable,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
		Metadata:  make(map[string]string),
	}

	if retryable {
		retryAfter := 30 * time.Second
		err.RetryAfter = &retryAfter
	}

	return err
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(provider, message string, retryAfter time.Duration) *IntegrationError {
	return &IntegrationError{
		Code:       ErrorCodeRateLimitExceeded,
		Message:    message,
		Provider:   provider,
		Severity:   SeverityMedium,
		Retryable:  true,
		RetryAfter: &retryAfter,
		Timestamp:  time.Now(),
		Context:    make(map[string]interface{}),
		Metadata:   make(map[string]string),
	}
}

// NewAPIError creates a new API error
func NewAPIError(provider, message string, statusCode int) *IntegrationError {
	severity := SeverityMedium
	retryable := false

	// Determine severity and retryability based on status code
	switch {
	case statusCode >= 500:
		severity = SeverityHigh
		retryable = true
	case statusCode == 429:
		severity = SeverityMedium
		retryable = true
	case statusCode >= 400 && statusCode < 500:
		severity = SeverityMedium
		retryable = false
	}

	err := &IntegrationError{
		Code:       ErrorCodeAPIError,
		Message:    message,
		Provider:   provider,
		Severity:   severity,
		Retryable:  retryable,
		StatusCode: statusCode,
		Timestamp:  time.Now(),
		Context:    make(map[string]interface{}),
		Metadata:   make(map[string]string),
	}

	if retryable {
		retryAfter := 60 * time.Second
		err.RetryAfter = &retryAfter
	}

	return err
}

// ErrorCategory represents error categories for grouping
type ErrorCategory string

const (
	CategoryAuthentication ErrorCategory = "authentication"
	CategoryNetwork        ErrorCategory = "network"
	CategoryValidation     ErrorCategory = "validation"
	CategoryBusiness       ErrorCategory = "business"
	CategoryAPI            ErrorCategory = "api"
	CategoryInternal       ErrorCategory = "internal"
	CategoryIntegration    ErrorCategory = "integration"
)

// GetErrorCategory returns the category of an error code
func GetErrorCategory(code ErrorCode) ErrorCategory {
	switch code {
	case ErrorCodeAuthenticationFailed, ErrorCodeCredentialsExpired, 
		 ErrorCodeInvalidCredentials, ErrorCodeAccessDenied:
		return CategoryAuthentication
	case ErrorCodeNetworkTimeout, ErrorCodeConnectionFailed, 
		 ErrorCodeServiceUnavailable:
		return CategoryNetwork
	case ErrorCodeValidationFailed, ErrorCodeInvalidInput, 
		 ErrorCodeMissingRequiredField, ErrorCodeInvalidFormat:
		return CategoryValidation
	case ErrorCodeProductNotFound, ErrorCodeOrderNotFound, 
		 ErrorCodeInsufficientStock, ErrorCodePriceChanged, 
		 ErrorCodeCategoryNotFound:
		return CategoryBusiness
	case ErrorCodeAPIError, ErrorCodeInvalidAPIVersion, 
		 ErrorCodeQuotaExceeded, ErrorCodeRateLimitExceeded, 
		 ErrorCodeTooManyRequests:
		return CategoryAPI
	case ErrorCodeInternalError, ErrorCodeConfigurationError, 
		 ErrorCodeDatabaseError, ErrorCodeCacheError:
		return CategoryInternal
	case ErrorCodeProviderError, ErrorCodeMappingError, 
		 ErrorCodeSyncError, ErrorCodeWebhookError:
		return CategoryIntegration
	default:
		return CategoryInternal
	}
}

// IsTemporaryError checks if an error is temporary and should be retried
func IsTemporaryError(err error) bool {
	if integrationErr, ok := err.(*IntegrationError); ok {
		return integrationErr.IsRetryable()
	}
	return false
}

// GetRetryDelay returns the recommended retry delay for an error
func GetRetryDelay(err error) time.Duration {
	if integrationErr, ok := err.(*IntegrationError); ok {
		if integrationErr.RetryAfter != nil {
			return *integrationErr.RetryAfter
		}
	}
	return 30 * time.Second // Default retry delay
}

// ErrorStats represents error statistics
type ErrorStats struct {
	TotalErrors    int                    `json:"total_errors"`
	ErrorsByCode   map[ErrorCode]int      `json:"errors_by_code"`
	ErrorsByProvider map[string]int       `json:"errors_by_provider"`
	ErrorsByCategory map[ErrorCategory]int `json:"errors_by_category"`
	ErrorsBySeverity map[ErrorSeverity]int `json:"errors_by_severity"`
	LastError      *IntegrationError      `json:"last_error"`
	LastUpdated    time.Time              `json:"last_updated"`
}

// NewErrorStats creates a new error statistics tracker
func NewErrorStats() *ErrorStats {
	return &ErrorStats{
		ErrorsByCode:     make(map[ErrorCode]int),
		ErrorsByProvider: make(map[string]int),
		ErrorsByCategory: make(map[ErrorCategory]int),
		ErrorsBySeverity: make(map[ErrorSeverity]int),
		LastUpdated:      time.Now(),
	}
}

// RecordError records an error in the statistics
func (es *ErrorStats) RecordError(err *IntegrationError) {
	es.TotalErrors++
	es.ErrorsByCode[err.Code]++
	es.ErrorsByProvider[err.Provider]++
	es.ErrorsByCategory[GetErrorCategory(err.Code)]++
	es.ErrorsBySeverity[err.Severity]++
	es.LastError = err
	es.LastUpdated = time.Now()
}

// GetErrorRate returns the error rate for a specific provider
func (es *ErrorStats) GetErrorRate(provider string, totalRequests int) float64 {
	if totalRequests == 0 {
		return 0.0
	}
	errors := es.ErrorsByProvider[provider]
	return float64(errors) / float64(totalRequests) * 100.0
}

// Reset resets the error statistics
func (es *ErrorStats) Reset() {
	es.TotalErrors = 0
	es.ErrorsByCode = make(map[ErrorCode]int)
	es.ErrorsByProvider = make(map[string]int)
	es.ErrorsByCategory = make(map[ErrorCategory]int)
	es.ErrorsBySeverity = make(map[ErrorSeverity]int)
	es.LastError = nil
	es.LastUpdated = time.Now()
}

// ErrorHandler handles integration errors with proper logging and metrics
type ErrorHandler struct {
	stats  *ErrorStats
	logger ErrorLogger
}

// ErrorLogger interface for error logging
type ErrorLogger interface {
	LogError(err *IntegrationError)
	LogWarning(message string, context map[string]interface{})
	LogInfo(message string, context map[string]interface{})
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger ErrorLogger) *ErrorHandler {
	return &ErrorHandler{
		stats:  NewErrorStats(),
		logger: logger,
	}
}

// HandleError handles an integration error
func (eh *ErrorHandler) HandleError(err *IntegrationError) {
	// Record error statistics
	eh.stats.RecordError(err)

	// Log the error
	if eh.logger != nil {
		eh.logger.LogError(err)
	}

	// Add stack trace for critical errors
	if err.Severity == SeverityCritical && err.StackTrace == "" {
		err.WithStackTrace()
	}
}

// GetStats returns the current error statistics
func (eh *ErrorHandler) GetStats() *ErrorStats {
	return eh.stats
}

// ResetStats resets the error statistics
func (eh *ErrorHandler) ResetStats() {
	eh.stats.Reset()
}