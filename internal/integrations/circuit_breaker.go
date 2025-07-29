package integrations

import (
	"sync"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	// StateClosed allows requests to pass through
	StateClosed CircuitBreakerState = iota
	// StateOpen blocks all requests
	StateOpen
	// StateHalfOpen allows limited requests to test if the service has recovered
	StateHalfOpen
)

// String returns the string representation of the state
func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name            string
	maxFailures     int
	resetTimeout    time.Duration
	halfOpenCalls   int
	onStateChange   func(name string, from, to CircuitBreakerState)
	
	mu              sync.Mutex
	state           CircuitBreakerState
	failures        int
	successCount    int
	lastFailureTime time.Time
	halfOpenAllowed int
}

// CircuitBreakerConfigNew holds configuration for circuit breaker
type CircuitBreakerConfigNew struct {
	Name            string
	MaxFailures     int
	ResetTimeout    time.Duration
	HalfOpenCalls   int
	OnStateChange   func(name string, from, to CircuitBreakerState)
}

// DefaultCircuitBreakerConfigNew returns default circuit breaker configuration
var DefaultCircuitBreakerConfigNew = CircuitBreakerConfigNew{
	MaxFailures:     5,
	ResetTimeout:    60 * time.Second,
	HalfOpenCalls:   3,
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfigNew) *CircuitBreaker {
	if config.MaxFailures <= 0 {
		config.MaxFailures = DefaultCircuitBreakerConfigNew.MaxFailures
	}
	if config.ResetTimeout <= 0 {
		config.ResetTimeout = DefaultCircuitBreakerConfigNew.ResetTimeout
	}
	if config.HalfOpenCalls <= 0 {
		config.HalfOpenCalls = DefaultCircuitBreakerConfigNew.HalfOpenCalls
	}
	
	return &CircuitBreaker{
		name:          config.Name,
		maxFailures:   config.MaxFailures,
		resetTimeout:  config.ResetTimeout,
		halfOpenCalls: config.HalfOpenCalls,
		onStateChange: config.OnStateChange,
		state:         StateClosed,
	}
}

// Execute runs the given function if the circuit breaker allows it
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	if !cb.canExecute() {
		return nil, &IntegrationError{
			Code:      "CIRCUIT_BREAKER_OPEN",
			Message:   "circuit breaker is open",
			Provider:  cb.name,
			Retryable: false,
			Timestamp: time.Now(),
		}
	}
	
	result, err := fn()
	cb.recordResult(err)
	
	return result, err
}

// canExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	now := time.Now()
	
	switch cb.state {
	case StateClosed:
		return true
		
	case StateOpen:
		// Check if we should transition to half-open
		if now.Sub(cb.lastFailureTime) > cb.resetTimeout {
			cb.changeState(StateHalfOpen)
			cb.halfOpenAllowed = cb.halfOpenCalls
			return true
		}
		return false
		
	case StateHalfOpen:
		// Allow limited requests in half-open state
		if cb.halfOpenAllowed > 0 {
			cb.halfOpenAllowed--
			return true
		}
		return false
		
	default:
		return false
	}
}

// recordResult records the result of an execution
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
}

// onSuccess handles successful execution
func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		// Reset failure count on success
		cb.failures = 0
		
	case StateHalfOpen:
		cb.successCount++
		// If we've had enough successes in half-open state, close the circuit
		if cb.successCount >= cb.halfOpenCalls {
			cb.changeState(StateClosed)
			cb.failures = 0
			cb.successCount = 0
		}
		
	case StateOpen:
		// This shouldn't happen, but handle it anyway
		cb.failures = 0
	}
}

// onFailure handles failed execution
func (cb *CircuitBreaker) onFailure() {
	cb.lastFailureTime = time.Now()
	
	switch cb.state {
	case StateClosed:
		cb.failures++
		if cb.failures >= cb.maxFailures {
			cb.changeState(StateOpen)
		}
		
	case StateHalfOpen:
		// Any failure in half-open state opens the circuit again
		cb.changeState(StateOpen)
		cb.successCount = 0
		
	case StateOpen:
		// Already open, just update the failure time
	}
}

// changeState changes the circuit breaker state
func (cb *CircuitBreaker) changeState(newState CircuitBreakerState) {
	if cb.state == newState {
		return
	}
	
	oldState := cb.state
	cb.state = newState
	
	if cb.onStateChange != nil {
		// Call the callback in a goroutine to avoid blocking
		go cb.onStateChange(cb.name, oldState, newState)
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// GetStats returns statistics about the circuit breaker
func (cb *CircuitBreaker) GetStats() CircuitBreakerStats {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	return CircuitBreakerStats{
		Name:            cb.name,
		State:           cb.state.String(),
		Failures:        cb.failures,
		SuccessCount:    cb.successCount,
		LastFailureTime: cb.lastFailureTime,
		HalfOpenAllowed: cb.halfOpenAllowed,
	}
}

// CircuitBreakerStats holds statistics about a circuit breaker
type CircuitBreakerStats struct {
	Name            string    `json:"name"`
	State           string    `json:"state"`
	Failures        int       `json:"failures"`
	SuccessCount    int       `json:"success_count"`
	LastFailureTime time.Time `json:"last_failure_time"`
	HalfOpenAllowed int       `json:"half_open_allowed"`
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.changeState(StateClosed)
	cb.failures = 0
	cb.successCount = 0
	cb.halfOpenAllowed = 0
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
	config   CircuitBreakerConfigNew
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(defaultConfig CircuitBreakerConfigNew) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		config:   defaultConfig,
	}
}

// GetBreaker gets or creates a circuit breaker for the given name
func (cbm *CircuitBreakerManager) GetBreaker(name string) *CircuitBreaker {
	cbm.mu.RLock()
	breaker, exists := cbm.breakers[name]
	cbm.mu.RUnlock()
	
	if exists {
		return breaker
	}
	
	cbm.mu.Lock()
	defer cbm.mu.Unlock()
	
	// Double-check after acquiring write lock
	breaker, exists = cbm.breakers[name]
	if exists {
		return breaker
	}
	
	// Create new breaker with default config
	config := cbm.config
	config.Name = name
	breaker = NewCircuitBreaker(config)
	cbm.breakers[name] = breaker
	
	return breaker
}

// GetAllStats returns statistics for all circuit breakers
func (cbm *CircuitBreakerManager) GetAllStats() map[string]CircuitBreakerStats {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()
	
	stats := make(map[string]CircuitBreakerStats)
	for name, breaker := range cbm.breakers {
		stats[name] = breaker.GetStats()
	}
	
	return stats
}

// ResetAll resets all circuit breakers
func (cbm *CircuitBreakerManager) ResetAll() {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()
	
	for _, breaker := range cbm.breakers {
		breaker.Reset()
	}
}