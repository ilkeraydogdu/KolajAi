package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"kolajAi/internal/errors"
)

// RetryStrategy defines different retry strategies
type RetryStrategy string

const (
	StrategyFixed        RetryStrategy = "fixed"
	StrategyLinear       RetryStrategy = "linear"
	StrategyExponential  RetryStrategy = "exponential"
	StrategyFibonacci    RetryStrategy = "fibonacci"
	StrategyCustom       RetryStrategy = "custom"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts     int           `json:"max_attempts"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	Strategy        RetryStrategy `json:"strategy"`
	BackoffFactor   float64       `json:"backoff_factor"`
	Jitter          bool          `json:"jitter"`
	JitterFactor    float64       `json:"jitter_factor"`
	RetryableErrors []errors.ErrorCode `json:"retryable_errors"`
	NonRetryableErrors []errors.ErrorCode `json:"non_retryable_errors"`
	OnRetry         func(attempt int, err error) `json:"-"`
	CustomDelayFunc func(attempt int) time.Duration `json:"-"`
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		Strategy:      StrategyExponential,
		BackoffFactor: 2.0,
		Jitter:        true,
		JitterFactor:  0.1,
		RetryableErrors: []errors.ErrorCode{
			errors.ErrorCodeNetworkTimeout,
			errors.ErrorCodeConnectionFailed,
			errors.ErrorCodeServiceUnavailable,
			errors.ErrorCodeTooManyRequests,
			errors.ErrorCodeRateLimitExceeded,
			errors.ErrorCodeAPIError,
		},
		NonRetryableErrors: []errors.ErrorCode{
			errors.ErrorCodeAuthenticationFailed,
			errors.ErrorCodeInvalidCredentials,
			errors.ErrorCodeAccessDenied,
			errors.ErrorCodeValidationFailed,
			errors.ErrorCodeInvalidInput,
		},
	}
}

// RetryManager manages retry operations
type RetryManager struct {
	config  *RetryConfig
	stats   *RetryStats
	mutex   sync.RWMutex
	rand    *rand.Rand
}

// RetryStats tracks retry statistics
type RetryStats struct {
	TotalAttempts    int64                       `json:"total_attempts"`
	TotalRetries     int64                       `json:"total_retries"`
	SuccessfulRetries int64                      `json:"successful_retries"`
	FailedRetries    int64                       `json:"failed_retries"`
	AverageAttempts  float64                     `json:"average_attempts"`
	RetriesByError   map[errors.ErrorCode]int64  `json:"retries_by_error"`
	RetriesByProvider map[string]int64           `json:"retries_by_provider"`
	LastRetry        time.Time                   `json:"last_retry"`
}

// NewRetryManager creates a new retry manager
func NewRetryManager(config *RetryConfig) *RetryManager {
	if config == nil {
		config = DefaultRetryConfig()
	}

	return &RetryManager{
		config: config,
		stats: &RetryStats{
			RetriesByError:    make(map[errors.ErrorCode]int64),
			RetriesByProvider: make(map[string]int64),
		},
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// RetryableFunc represents a function that can be retried
type RetryableFunc func() error

// RetryableFuncWithContext represents a function that can be retried with context
type RetryableFuncWithContext func(ctx context.Context) error

// Execute executes a function with retry logic
func (rm *RetryManager) Execute(fn RetryableFunc) error {
	return rm.ExecuteWithContext(context.Background(), func(ctx context.Context) error {
		return fn()
	})
}

// ExecuteWithContext executes a function with retry logic and context
func (rm *RetryManager) ExecuteWithContext(ctx context.Context, fn RetryableFuncWithContext) error {
	var lastErr error
	
	for attempt := 1; attempt <= rm.config.MaxAttempts; attempt++ {
		rm.recordAttempt()
		
		// Execute the function
		err := fn(ctx)
		if err == nil {
			// Success
			if attempt > 1 {
				rm.recordSuccessfulRetry()
			}
			return nil
		}
		
		lastErr = err
		
		// Check if we should retry
		if !rm.shouldRetry(err, attempt) {
			rm.recordFailedRetry(err)
			return err
		}
		
		// Don't wait after the last attempt
		if attempt == rm.config.MaxAttempts {
			rm.recordFailedRetry(err)
			break
		}
		
		// Calculate delay
		delay := rm.calculateDelay(attempt)
		
		// Call retry callback if provided
		if rm.config.OnRetry != nil {
			rm.config.OnRetry(attempt, err)
		}
		
		rm.recordRetry(err)
		
		// Wait before next attempt
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}
	
	return lastErr
}

// shouldRetry determines if an error should be retried
func (rm *RetryManager) shouldRetry(err error, attempt int) bool {
	// Check if we've exceeded max attempts
	if attempt >= rm.config.MaxAttempts {
		return false
	}
	
	// Check if it's an integration error
	integrationErr, ok := err.(*errors.IntegrationError)
	if !ok {
		// For non-integration errors, only retry network-related errors
		return rm.isNetworkError(err)
	}
	
	// Check if explicitly marked as non-retryable
	for _, nonRetryableCode := range rm.config.NonRetryableErrors {
		if integrationErr.Code == nonRetryableCode {
			return false
		}
	}
	
	// Check if explicitly marked as retryable
	for _, retryableCode := range rm.config.RetryableErrors {
		if integrationErr.Code == retryableCode {
			return true
		}
	}
	
	// Use the error's retryable flag
	return integrationErr.IsRetryable()
}

// isNetworkError checks if an error is network-related
func (rm *RetryManager) isNetworkError(err error) bool {
	errorMsg := err.Error()
	networkKeywords := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"network unreachable",
		"no route to host",
		"temporary failure",
	}
	
	for _, keyword := range networkKeywords {
		if contains(errorMsg, keyword) {
			return true
		}
	}
	
	return false
}

// calculateDelay calculates the delay for the next retry attempt
func (rm *RetryManager) calculateDelay(attempt int) time.Duration {
	var delay time.Duration
	
	switch rm.config.Strategy {
	case StrategyFixed:
		delay = rm.config.InitialDelay
		
	case StrategyLinear:
		delay = time.Duration(attempt) * rm.config.InitialDelay
		
	case StrategyExponential:
		delay = time.Duration(float64(rm.config.InitialDelay) * math.Pow(rm.config.BackoffFactor, float64(attempt-1)))
		
	case StrategyFibonacci:
		delay = time.Duration(fibonacci(attempt)) * rm.config.InitialDelay
		
	case StrategyCustom:
		if rm.config.CustomDelayFunc != nil {
			delay = rm.config.CustomDelayFunc(attempt)
		} else {
			delay = rm.config.InitialDelay
		}
		
	default:
		delay = rm.config.InitialDelay
	}
	
	// Apply maximum delay limit
	if delay > rm.config.MaxDelay {
		delay = rm.config.MaxDelay
	}
	
	// Apply jitter if enabled
	if rm.config.Jitter {
		delay = rm.applyJitter(delay)
	}
	
	return delay
}

// applyJitter applies jitter to the delay
func (rm *RetryManager) applyJitter(delay time.Duration) time.Duration {
	if rm.config.JitterFactor <= 0 {
		return delay
	}
	
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	jitter := time.Duration(float64(delay) * rm.config.JitterFactor * (rm.rand.Float64()*2 - 1))
	result := delay + jitter
	
	// Ensure the result is positive
	if result < 0 {
		result = delay / 2
	}
	
	return result
}

// fibonacci calculates the nth Fibonacci number
func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	
	return b
}

// recordAttempt records an attempt
func (rm *RetryManager) recordAttempt() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.stats.TotalAttempts++
}

// recordRetry records a retry
func (rm *RetryManager) recordRetry(err error) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	rm.stats.TotalRetries++
	rm.stats.LastRetry = time.Now()
	
	// Record by error type
	if integrationErr, ok := err.(*errors.IntegrationError); ok {
		rm.stats.RetriesByError[integrationErr.Code]++
		rm.stats.RetriesByProvider[integrationErr.Provider]++
	}
}

// recordSuccessfulRetry records a successful retry
func (rm *RetryManager) recordSuccessfulRetry() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.stats.SuccessfulRetries++
}

// recordFailedRetry records a failed retry
func (rm *RetryManager) recordFailedRetry(err error) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.stats.FailedRetries++
}

// GetStats returns retry statistics
func (rm *RetryManager) GetStats() *RetryStats {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	// Calculate average attempts
	avgAttempts := float64(0)
	if rm.stats.TotalRetries > 0 {
		avgAttempts = float64(rm.stats.TotalAttempts) / float64(rm.stats.TotalRetries)
	}
	
	// Create a copy of stats
	statsCopy := &RetryStats{
		TotalAttempts:     rm.stats.TotalAttempts,
		TotalRetries:      rm.stats.TotalRetries,
		SuccessfulRetries: rm.stats.SuccessfulRetries,
		FailedRetries:     rm.stats.FailedRetries,
		AverageAttempts:   avgAttempts,
		RetriesByError:    make(map[errors.ErrorCode]int64),
		RetriesByProvider: make(map[string]int64),
		LastRetry:         rm.stats.LastRetry,
	}
	
	// Copy maps
	for k, v := range rm.stats.RetriesByError {
		statsCopy.RetriesByError[k] = v
	}
	for k, v := range rm.stats.RetriesByProvider {
		statsCopy.RetriesByProvider[k] = v
	}
	
	return statsCopy
}

// ResetStats resets retry statistics
func (rm *RetryManager) ResetStats() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	rm.stats = &RetryStats{
		RetriesByError:    make(map[errors.ErrorCode]int64),
		RetriesByProvider: make(map[string]int64),
	}
}

// UpdateConfig updates the retry configuration
func (rm *RetryManager) UpdateConfig(config *RetryConfig) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.config = config
}

// GetConfig returns the current retry configuration
func (rm *RetryManager) GetConfig() *RetryConfig {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	return rm.config
}

// RetryWithBackoff is a convenience function for simple retry with exponential backoff
func RetryWithBackoff(ctx context.Context, maxAttempts int, initialDelay time.Duration, fn RetryableFuncWithContext) error {
	config := &RetryConfig{
		MaxAttempts:   maxAttempts,
		InitialDelay:  initialDelay,
		MaxDelay:      5 * time.Minute,
		Strategy:      StrategyExponential,
		BackoffFactor: 2.0,
		Jitter:        true,
		JitterFactor:  0.1,
	}
	
	manager := NewRetryManager(config)
	return manager.ExecuteWithContext(ctx, fn)
}

// RetryWithCustomStrategy creates a retry manager with custom strategy
func RetryWithCustomStrategy(config *RetryConfig, fn RetryableFunc) error {
	manager := NewRetryManager(config)
	return manager.Execute(fn)
}

// CircuitBreakerRetryManager combines circuit breaker with retry logic
type CircuitBreakerRetryManager struct {
	retryManager    *RetryManager
	circuitBreaker  *CircuitBreaker
}

// CircuitBreaker represents a simple circuit breaker
type CircuitBreaker struct {
	maxFailures   int64
	resetTimeout  time.Duration
	failures      int64
	lastFailTime  time.Time
	state         CircuitBreakerState
	mutex         sync.RWMutex
}

// CircuitBreakerState represents circuit breaker states
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int64, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

// Execute executes a function through the circuit breaker
func (cb *CircuitBreaker) Execute(fn RetryableFunc) error {
	if !cb.canExecute() {
		return errors.NewIntegrationError(
			errors.ErrorCodeServiceUnavailable,
			"circuit breaker is open",
			"circuit_breaker",
		)
	}
	
	err := fn()
	cb.recordResult(err)
	return err
}

// canExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) canExecute() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// recordResult records the result of an execution
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()
		
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}
	} else {
		// Success
		cb.failures = 0
		cb.state = StateClosed
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// NewCircuitBreakerRetryManager creates a new circuit breaker retry manager
func NewCircuitBreakerRetryManager(retryConfig *RetryConfig, maxFailures int64, resetTimeout time.Duration) *CircuitBreakerRetryManager {
	return &CircuitBreakerRetryManager{
		retryManager:   NewRetryManager(retryConfig),
		circuitBreaker: NewCircuitBreaker(maxFailures, resetTimeout),
	}
}

// Execute executes a function with both circuit breaker and retry logic
func (cbrm *CircuitBreakerRetryManager) Execute(fn RetryableFunc) error {
	return cbrm.circuitBreaker.Execute(func() error {
		return cbrm.retryManager.Execute(fn)
	})
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
			len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr ||
			 containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}