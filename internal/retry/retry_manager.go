package retry

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"time"
)

// RetryManager handles retry logic for operations
type RetryManager struct {
	config RetryConfig
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxAttempts     int           `json:"max_attempts"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	JitterEnabled   bool          `json:"jitter_enabled"`
	RetryableErrors []string      `json:"retryable_errors"`
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		JitterEnabled: true,
		RetryableErrors: []string{
			"TIMEOUT",
			"TEMPORARY_FAILURE",
			"RATE_LIMIT",
			"SERVICE_UNAVAILABLE",
		},
	}
}

// NewRetryManager creates a new retry manager
func NewRetryManager(config RetryConfig) *RetryManager {
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = 3
	}
	if config.InitialDelay <= 0 {
		config.InitialDelay = 1 * time.Second
	}
	if config.MaxDelay <= 0 {
		config.MaxDelay = 30 * time.Second
	}
	if config.BackoffFactor <= 1 {
		config.BackoffFactor = 2.0
	}

	return &RetryManager{
		config: config,
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func(ctx context.Context) error

// RetryResult contains the result of a retry operation
type RetryResult struct {
	Attempts      int           `json:"attempts"`
	LastError     error         `json:"last_error"`
	Success       bool          `json:"success"`
	TotalDuration time.Duration `json:"total_duration"`
}

// Execute executes a function with retry logic
func (rm *RetryManager) Execute(ctx context.Context, fn RetryableFunc) *RetryResult {
	startTime := time.Now()
	result := &RetryResult{
		Attempts: 0,
		Success:  false,
	}

	for attempt := 0; attempt < rm.config.MaxAttempts; attempt++ {
		result.Attempts++

		// Check context before attempting
		if err := ctx.Err(); err != nil {
			result.LastError = err
			break
		}

		// Execute the function
		err := fn(ctx)
		if err == nil {
			result.Success = true
			break
		}

		result.LastError = err

		// Check if error is retryable
		if !rm.isRetryableError(err) {
			break
		}

		// Don't sleep after the last attempt
		if attempt < rm.config.MaxAttempts-1 {
			delay := rm.calculateDelay(attempt)

			select {
			case <-ctx.Done():
				result.LastError = ctx.Err()
				break
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}

	result.TotalDuration = time.Since(startTime)
	return result
}

// ExecuteWithCustomRetry executes a function with custom retry logic
func (rm *RetryManager) ExecuteWithCustomRetry(ctx context.Context, fn RetryableFunc, isRetryable func(error) bool) *RetryResult {
	startTime := time.Now()
	result := &RetryResult{
		Attempts: 0,
		Success:  false,
	}

	for attempt := 0; attempt < rm.config.MaxAttempts; attempt++ {
		result.Attempts++

		// Check context before attempting
		if err := ctx.Err(); err != nil {
			result.LastError = err
			break
		}

		// Execute the function
		err := fn(ctx)
		if err == nil {
			result.Success = true
			break
		}

		result.LastError = err

		// Check if error is retryable using custom function
		if !isRetryable(err) {
			break
		}

		// Don't sleep after the last attempt
		if attempt < rm.config.MaxAttempts-1 {
			delay := rm.calculateDelay(attempt)

			select {
			case <-ctx.Done():
				result.LastError = ctx.Err()
				break
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}

	result.TotalDuration = time.Since(startTime)
	return result
}

// calculateDelay calculates the delay for the given attempt
func (rm *RetryManager) calculateDelay(attempt int) time.Duration {
	// Exponential backoff
	delay := float64(rm.config.InitialDelay) * math.Pow(rm.config.BackoffFactor, float64(attempt))

	// Apply max delay cap
	if delay > float64(rm.config.MaxDelay) {
		delay = float64(rm.config.MaxDelay)
	}

	// Apply jitter if enabled
	if rm.config.JitterEnabled {
		// Use cryptographically secure random for jitter
		maxJitterNanos := int64(0.3 * delay) // delay is already in nanoseconds as float64
		if maxJitterNanos > 0 {
			maxJitter := big.NewInt(maxJitterNanos)
			jitterBig, err := rand.Int(rand.Reader, maxJitter)
			if err != nil {
				// Fallback to no jitter if crypto/rand fails
				// Keep original delay unchanged
			} else {
				jitterNanos := jitterBig.Int64()
				delay = delay + float64(jitterNanos)
			}
		}
	}

	return time.Duration(delay)
}

// isRetryableError checks if an error is retryable
func (rm *RetryManager) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// Check against configured retryable errors
	for _, retryableErr := range rm.config.RetryableErrors {
		if contains(errMsg, retryableErr) {
			return true
		}
	}

	// Check for common retryable patterns
	retryablePatterns := []string{
		"timeout",
		"temporary",
		"unavailable",
		"rate limit",
		"too many requests",
		"connection refused",
		"connection reset",
		"EOF",
	}

	for _, pattern := range retryablePatterns {
		if contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(containsIgnoreCase(s, substr)))
}

// containsIgnoreCase performs case-insensitive substring search
func containsIgnoreCase(s, substr string) bool {
	sLower := toLowerCase(s)
	substrLower := toLowerCase(substr)
	return indexOfSubstring(sLower, substrLower) >= 0
}

// toLowerCase converts string to lowercase
func toLowerCase(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

// indexOfSubstring finds the index of substring in string
func indexOfSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// RetryableOperation represents an operation that can be retried
type RetryableOperation struct {
	Name        string
	MaxAttempts int
	Timeout     time.Duration
	OnRetry     func(attempt int, err error)
	OnSuccess   func(attempts int)
	OnFailure   func(attempts int, err error)
}

// ExecuteOperation executes a retryable operation with callbacks
func (rm *RetryManager) ExecuteOperation(ctx context.Context, op RetryableOperation, fn RetryableFunc) error {
	// Override max attempts if specified
	originalMaxAttempts := rm.config.MaxAttempts
	if op.MaxAttempts > 0 {
		rm.config.MaxAttempts = op.MaxAttempts
	}
	defer func() {
		rm.config.MaxAttempts = originalMaxAttempts
	}()

	// Apply timeout if specified
	if op.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, op.Timeout)
		defer cancel()
	}

	// Execute with retry
	var result *RetryResult
	result = rm.Execute(ctx, func(ctx context.Context) error {
		err := fn(ctx)
		if err != nil && op.OnRetry != nil && result != nil && result.Attempts < rm.config.MaxAttempts {
			op.OnRetry(result.Attempts, err)
		}
		return err
	})

	// Call appropriate callback
	if result.Success {
		if op.OnSuccess != nil {
			op.OnSuccess(result.Attempts)
		}
		return nil
	} else {
		if op.OnFailure != nil {
			op.OnFailure(result.Attempts, result.LastError)
		}
		return fmt.Errorf("operation %s failed after %d attempts: %w", op.Name, result.Attempts, result.LastError)
	}
}
