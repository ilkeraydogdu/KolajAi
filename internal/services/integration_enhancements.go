package services

import (
	"fmt"
	"sync"
	"time"
)

// RateLimitManager manages rate limiting for marketplace integrations
type RateLimitManager struct {
	limits map[string]*RateLimit
	mu     sync.RWMutex
}

// RateLimit represents rate limiting configuration for an integration
type RateLimit struct {
	IntegrationID     string
	RequestsPerMinute int
	RequestsPerHour   int
	RequestsPerDay    int
	CurrentUsage      int
	LastReset         time.Time
	IsBlocked         bool
	BlockUntil        time.Time
	WindowSize        time.Duration
	BurstSize         int
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	State           CircuitState
	FailureCount    int
	LastFailureTime time.Time
	Threshold       int
	Timeout         time.Duration
	mu              sync.RWMutex
}

// CircuitState represents circuit breaker states
type CircuitState string

const (
	StateClosed   CircuitState = "closed"
	StateOpen     CircuitState = "open"
	StateHalfOpen CircuitState = "half_open"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts       int
	InitialDelay      time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
	RetryableErrors   []string
	Jitter            bool
}

// IntegrationMetrics tracks integration performance metrics
type IntegrationMetrics struct {
	IntegrationID   string
	TotalRequests   int64
	SuccessfulRequests int64
	FailedRequests  int64
	AverageResponseTime time.Duration
	LastRequestTime time.Time
	ErrorRate       float64
	SuccessRate     float64
	mu              sync.RWMutex
}

// NewRateLimitManager creates a new rate limit manager
func NewRateLimitManager() *RateLimitManager {
	return &RateLimitManager{
		limits: make(map[string]*RateLimit),
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		State:     StateClosed,
		Threshold: threshold,
		Timeout:   timeout,
	}
}

// NewRetryConfig creates a new retry configuration
func NewRetryConfig(maxAttempts int, initialDelay time.Duration) *RetryConfig {
	return &RetryConfig{
		MaxAttempts:       maxAttempts,
		InitialDelay:      initialDelay,
		MaxDelay:          initialDelay * 10,
		BackoffMultiplier: 2.0,
		RetryableErrors:   []string{"timeout", "connection_error", "rate_limit"},
		Jitter:            true,
	}
}

// NewIntegrationMetrics creates new integration metrics
func NewIntegrationMetrics(integrationID string) *IntegrationMetrics {
	return &IntegrationMetrics{
		IntegrationID: integrationID,
	}
}

// SetRateLimit sets rate limit for an integration
func (rlm *RateLimitManager) SetRateLimit(integrationID string, limit *RateLimit) {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()
	rlm.limits[integrationID] = limit
}

// GetRateLimit gets rate limit for an integration
func (rlm *RateLimitManager) GetRateLimit(integrationID string) (*RateLimit, bool) {
	rlm.mu.RLock()
	defer rlm.mu.RUnlock()
	limit, exists := rlm.limits[integrationID]
	return limit, exists
}

// CheckRateLimit checks if rate limit is exceeded
func (rlm *RateLimitManager) CheckRateLimit(integrationID string) error {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	limit, exists := rlm.limits[integrationID]
	if !exists {
		return nil // No rate limit configured
	}

	now := time.Now()

	// Reset counters if window has passed
	if now.Sub(limit.LastReset) >= limit.WindowSize {
		limit.CurrentUsage = 0
		limit.LastReset = now
		limit.IsBlocked = false
	}

	// Check if currently blocked
	if limit.IsBlocked {
		if now.Before(limit.BlockUntil) {
			return fmt.Errorf("rate limit exceeded for %s, blocked until %v", integrationID, limit.BlockUntil)
		}
		limit.IsBlocked = false
	}

	// Check if limit would be exceeded
	if limit.CurrentUsage >= limit.BurstSize {
		limit.IsBlocked = true
		limit.BlockUntil = now.Add(limit.WindowSize)
		return fmt.Errorf("rate limit exceeded for %s", integrationID)
	}

	limit.CurrentUsage++
	return nil
}

// Execute runs an operation with circuit breaker protection
func (cb *CircuitBreaker) Execute(operation func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.State {
	case StateOpen:
		if time.Since(cb.LastFailureTime) > cb.Timeout {
			cb.State = StateHalfOpen
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	case StateHalfOpen:
		// Allow one request to test if service is back
		break
	case StateClosed:
		// Normal operation
		break
	}

	err := operation()
	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}
	return err
}

// recordFailure records a failure and updates circuit breaker state
func (cb *CircuitBreaker) recordFailure() {
	cb.FailureCount++
	cb.LastFailureTime = time.Now()

	if cb.FailureCount >= cb.Threshold {
		cb.State = StateOpen
	}
}

// recordSuccess records a success and resets circuit breaker
func (cb *CircuitBreaker) recordSuccess() {
	cb.FailureCount = 0
	if cb.State == StateHalfOpen {
		cb.State = StateClosed
	}
}

// GetState returns current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.State
}

// RetryOperation retries an operation with exponential backoff
func RetryOperation(operation func() error, config *RetryConfig) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err
		if !isRetryableError(err, config.RetryableErrors) {
			return err
		}

		if attempt < config.MaxAttempts {
			// Add jitter if enabled
			if config.Jitter {
				delay = addJitter(delay)
			}

			time.Sleep(delay)
			delay = time.Duration(float64(delay) * config.BackoffMultiplier)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %v", config.MaxAttempts, lastErr)
}

// isRetryableError checks if an error is retryable
func isRetryableError(err error, retryableErrors []string) bool {
	errStr := err.Error()
	for _, retryableError := range retryableErrors {
		if containsSubstring(errStr, retryableError) {
			return true
		}
	}
	return false
}

// containsSubstring checks if a string contains a substring
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// addJitter adds random jitter to delay
func addJitter(delay time.Duration) time.Duration {
	jitter := time.Duration(float64(delay) * 0.1) // 10% jitter
	return delay + jitter
}

// RecordRequest records a request in integration metrics
func (im *IntegrationMetrics) RecordRequest(success bool, responseTime time.Duration) {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.TotalRequests++
	im.LastRequestTime = time.Now()

	if success {
		im.SuccessfulRequests++
	} else {
		im.FailedRequests++
	}

	// Update average response time
	if im.TotalRequests == 1 {
		im.AverageResponseTime = responseTime
	} else {
		im.AverageResponseTime = time.Duration(
			(float64(im.AverageResponseTime) + float64(responseTime)) / 2,
		)
	}

	// Update success/error rates
	if im.TotalRequests > 0 {
		im.SuccessRate = float64(im.SuccessfulRequests) / float64(im.TotalRequests)
		im.ErrorRate = float64(im.FailedRequests) / float64(im.TotalRequests)
	}
}

// GetMetrics returns current metrics
func (im *IntegrationMetrics) GetMetrics() map[string]interface{} {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return map[string]interface{}{
		"integration_id":        im.IntegrationID,
		"total_requests":        im.TotalRequests,
		"successful_requests":   im.SuccessfulRequests,
		"failed_requests":       im.FailedRequests,
		"average_response_time": im.AverageResponseTime,
		"last_request_time":     im.LastRequestTime,
		"success_rate":          im.SuccessRate,
		"error_rate":            im.ErrorRate,
	}
}

// ResetMetrics resets all metrics
func (im *IntegrationMetrics) ResetMetrics() {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.TotalRequests = 0
	im.SuccessfulRequests = 0
	im.FailedRequests = 0
	im.AverageResponseTime = 0
	im.SuccessRate = 0
	im.ErrorRate = 0
}