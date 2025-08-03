package integrations

import (
	"context"
	"fmt"
	"sync"
	"time"
	"github.com/sony/gobreaker"
)

// Manager manages all integrations
type Manager struct {
	integrations    map[string]*Integration
	providers       map[string]IntegrationProvider
	webhookHandlers map[string]WebhookHandler
	circuitBreakers map[string]*gobreaker.CircuitBreaker
	logger          IntegrationLogger
	metrics         IntegrationMetrics
	cache           IntegrationCache
	eventBus        IntegrationEventBus
	mu              sync.RWMutex
	config          *ManagerConfig
}

// ManagerConfig holds configuration for the integration manager
type ManagerConfig struct {
	EnableCircuitBreaker bool
	EnableCaching        bool
	EnableMetrics        bool
	DefaultTimeout       time.Duration
	HealthCheckInterval  time.Duration
	MaxConcurrentRequests int
}

// NewManager creates a new integration manager
func NewManager(config *ManagerConfig) *Manager {
	if config == nil {
		config = &ManagerConfig{
			EnableCircuitBreaker:  true,
			EnableCaching:         true,
			EnableMetrics:         true,
			DefaultTimeout:        30 * time.Second,
			HealthCheckInterval:   5 * time.Minute,
			MaxConcurrentRequests: 100,
		}
	}

	return &Manager{
		integrations:    make(map[string]*Integration),
		providers:       make(map[string]IntegrationProvider),
		webhookHandlers: make(map[string]WebhookHandler),
		circuitBreakers: make(map[string]*gobreaker.CircuitBreaker),
		config:          config,
	}
}

// SetLogger sets the logger for the manager
func (m *Manager) SetLogger(logger IntegrationLogger) {
	m.logger = logger
}

// SetMetrics sets the metrics collector for the manager
func (m *Manager) SetMetrics(metrics IntegrationMetrics) {
	m.metrics = metrics
}

// SetCache sets the cache for the manager
func (m *Manager) SetCache(cache IntegrationCache) {
	m.cache = cache
}

// SetEventBus sets the event bus for the manager
func (m *Manager) SetEventBus(eventBus IntegrationEventBus) {
	m.eventBus = eventBus
}

// RegisterIntegration registers a new integration
func (m *Manager) RegisterIntegration(integration *Integration, provider IntegrationProvider) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.integrations[integration.ID]; exists {
		return fmt.Errorf("integration %s already registered", integration.ID)
	}

	// Initialize the provider with timeout
	ctx, cancel := context.WithTimeout(context.Background(), m.config.DefaultTimeout)
	defer cancel()

	if err := provider.Initialize(ctx, integration.Credentials, integration.Config); err != nil {
		return fmt.Errorf("failed to initialize provider: %w", err)
	}

	// Set up circuit breaker if enabled
	if m.config.EnableCircuitBreaker {
		settings := gobreaker.Settings{
			Name:        integration.ID,
			MaxRequests: uint32(DefaultCircuitBreakerConfig.HalfOpenMaxCalls),
			Interval:    DefaultCircuitBreakerConfig.Timeout,
			Timeout:     DefaultCircuitBreakerConfig.Timeout,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
				return counts.Requests >= uint32(DefaultCircuitBreakerConfig.FailureThreshold) && failureRatio >= 0.6
			},
			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
				if m.logger != nil {
					m.logger.LogError(name, fmt.Errorf("circuit breaker state changed from %s to %s", from, to))
				}
				if m.eventBus != nil {
					m.eventBus.Publish(IntegrationEvent{
						Type:          "circuit_breaker_state_change",
						IntegrationID: name,
						Timestamp:     time.Now(),
						Data: map[string]interface{}{
							"from": from.String(),
							"to":   to.String(),
						},
					})
				}
			},
		}
		m.circuitBreakers[integration.ID] = gobreaker.NewCircuitBreaker(settings)
	}

	// Store integration and provider
	m.integrations[integration.ID] = integration
	m.providers[integration.ID] = provider

	// Start health check routine
	go m.startHealthCheckRoutine(integration.ID)

	// Publish integration registered event
	if m.eventBus != nil {
		m.eventBus.Publish(IntegrationEvent{
			Type:          "integration_registered",
			IntegrationID: integration.ID,
			Timestamp:     time.Now(),
			Data: map[string]interface{}{
				"name":     integration.Name,
				"type":     integration.Type,
				"provider": integration.Provider,
			},
		})
	}

	return nil
}

// RegisterWebhookHandler registers a webhook handler for an integration
func (m *Manager) RegisterWebhookHandler(integrationID string, handler WebhookHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.integrations[integrationID]; !exists {
		return fmt.Errorf("integration %s not found", integrationID)
	}

	m.webhookHandlers[integrationID] = handler
	return nil
}

// GetIntegration returns an integration by ID
func (m *Manager) GetIntegration(id string) (*Integration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	integration, exists := m.integrations[id]
	if !exists {
		return nil, fmt.Errorf("integration %s not found", id)
	}

	return integration, nil
}

// GetIntegrationsByType returns all integrations of a specific type
func (m *Manager) GetIntegrationsByType(integrationType IntegrationType) []*Integration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Integration
	for _, integration := range m.integrations {
		if integration.Type == integrationType {
			result = append(result, integration)
		}
	}

	return result
}

// GetAllIntegrations returns all registered integrations
func (m *Manager) GetAllIntegrations() []*Integration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Integration
	for _, integration := range m.integrations {
		result = append(result, integration)
	}

	return result
}

// ExecuteRequest executes a request through an integration with circuit breaker and retry logic
func (m *Manager) ExecuteRequest(ctx context.Context, integrationID string, request IntegrationRequest) (*IntegrationResponse, error) {
	integration, err := m.GetIntegration(integrationID)
	if err != nil {
		return nil, err
	}

	provider, exists := m.providers[integrationID]
	if !exists {
		return nil, fmt.Errorf("provider not found for integration %s", integrationID)
	}

	// Check cache if enabled
	if m.config.EnableCaching && m.cache != nil {
		cacheKey := fmt.Sprintf("%s:%s:%s", integrationID, request.Method, request.Endpoint)
		if cached, found := m.cache.Get(cacheKey); found {
			if response, ok := cached.(*IntegrationResponse); ok {
				return response, nil
			}
		}
	}

	// Execute with circuit breaker if enabled
	var response *IntegrationResponse
	var execErr error

	if m.config.EnableCircuitBreaker {
		cb, exists := m.circuitBreakers[integrationID]
		if !exists {
			return nil, fmt.Errorf("circuit breaker not found for integration %s", integrationID)
		}

		result, err := cb.Execute(func() (interface{}, error) {
			return m.executeWithRetry(ctx, integration, provider, request)
		})

		if err != nil {
			execErr = err
		} else if result != nil {
			response = result.(*IntegrationResponse)
		}
	} else {
		response, execErr = m.executeWithRetry(ctx, integration, provider, request)
	}

	// Log the request and response
	if m.logger != nil {
		m.logger.LogRequest(integrationID, request)
		if response != nil {
			m.logger.LogResponse(integrationID, *response)
		}
		if execErr != nil {
			m.logger.LogError(integrationID, execErr)
		}
	}

	// Record metrics
	if m.config.EnableMetrics && m.metrics != nil {
		duration := time.Duration(0)
		if response != nil {
			duration = response.Duration
		}
		m.metrics.RecordRequest(integrationID, request.Method, duration, execErr == nil)
		if execErr != nil {
			errorCode := "UNKNOWN"
			if integrationErr, ok := execErr.(*IntegrationError); ok {
				errorCode = integrationErr.Code
			}
			m.metrics.RecordError(integrationID, errorCode)
		}
	}

	// Cache successful responses if enabled
	if execErr == nil && response != nil && m.config.EnableCaching && m.cache != nil {
		cacheKey := fmt.Sprintf("%s:%s:%s", integrationID, request.Method, request.Endpoint)
		cacheTTL := 5 * time.Minute // Default cache TTL
		m.cache.Set(cacheKey, response, cacheTTL)
	}

	return response, execErr
}

// ProcessWebhook processes an incoming webhook
func (m *Manager) ProcessWebhook(ctx context.Context, integrationID string, event WebhookEvent) error {
	handler, exists := m.webhookHandlers[integrationID]
	if !exists {
		return fmt.Errorf("webhook handler not found for integration %s", integrationID)
	}

	// Validate webhook
	if err := handler.ValidateWebhook(event.Headers, []byte(event.Signature)); err != nil {
		return fmt.Errorf("webhook validation failed: %w", err)
	}

	// Process webhook
	if err := handler.ProcessWebhook(ctx, event); err != nil {
		return fmt.Errorf("webhook processing failed: %w", err)
	}

	// Log webhook
	if m.logger != nil {
		m.logger.LogWebhook(integrationID, event)
	}

	// Record metrics
	if m.config.EnableMetrics && m.metrics != nil {
		m.metrics.RecordWebhook(integrationID, event.Type, true)
	}

	// Publish webhook processed event
	if m.eventBus != nil {
		m.eventBus.Publish(IntegrationEvent{
			Type:          "webhook_processed",
			IntegrationID: integrationID,
			Timestamp:     time.Now(),
			Data: map[string]interface{}{
				"webhook_type": event.Type,
				"webhook_id":   event.ID,
			},
		})
	}

	return nil
}

// executeWithRetry executes a request with retry logic
func (m *Manager) executeWithRetry(ctx context.Context, integration *Integration, provider IntegrationProvider, request IntegrationRequest) (*IntegrationResponse, error) {
	retryPolicy := DefaultRetryPolicy
	if request.Retries > 0 {
		retryPolicy.MaxAttempts = request.Retries
	}

	var lastErr error
	delay := retryPolicy.InitialDelay

	for attempt := 0; attempt < retryPolicy.MaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				// Continue with retry
			}
		}

		// Execute the actual request (this would be implemented by specific providers)
		response := &IntegrationResponse{
			ID:        request.ID,
			Duration:  time.Since(time.Now()),
		}

		// For now, just do a health check as example
		err := provider.HealthCheck(ctx)
		if err == nil {
			response.StatusCode = 200
			return response, nil
		}

		lastErr = err

		// Check if error is retryable
		integrationErr, ok := err.(*IntegrationError)
		if !ok || !integrationErr.Retryable {
			return response, err
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * retryPolicy.BackoffFactor)
		if delay > retryPolicy.MaxDelay {
			delay = retryPolicy.MaxDelay
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// startHealthCheckRoutine starts a routine to periodically check integration health
func (m *Manager) startHealthCheckRoutine(integrationID string) {
	ticker := time.NewTicker(m.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.performHealthCheck(integrationID)
		}
	}
}

// performHealthCheck performs a health check on an integration
func (m *Manager) performHealthCheck(integrationID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	provider, exists := m.providers[integrationID]
	if !exists {
		return
	}

	err := provider.HealthCheck(ctx)
	
	m.mu.Lock()
	integration, exists := m.integrations[integrationID]
	if !exists {
		m.mu.Unlock()
		return
	}

	// Update integration status
	if err != nil {
		integration.Status = IntegrationStatusError
		integration.Metadata.ErrorCount++
	} else {
		integration.Status = IntegrationStatusActive
		integration.Metadata.SuccessCount++
	}
	integration.Metadata.LastHealthCheck = time.Now()
	m.mu.Unlock()

	// Publish health check event
	if m.eventBus != nil {
		m.eventBus.Publish(IntegrationEvent{
			Type:          "health_check_completed",
			IntegrationID: integrationID,
			Timestamp:     time.Now(),
			Data: map[string]interface{}{
				"status": integration.Status,
				"error":  err != nil,
			},
		})
	}
}

// UpdateIntegrationConfig updates the configuration of an integration
func (m *Manager) UpdateIntegrationConfig(integrationID string, config map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	integration, exists := m.integrations[integrationID]
	if !exists {
		return fmt.Errorf("integration %s not found", integrationID)
	}

	provider, exists := m.providers[integrationID]
	if !exists {
		return fmt.Errorf("provider not found for integration %s", integrationID)
	}

	// Re-initialize provider with new config
	ctx, cancel := context.WithTimeout(context.Background(), m.config.DefaultTimeout)
	defer cancel()

	if err := provider.Initialize(ctx, integration.Credentials, config); err != nil {
		return fmt.Errorf("failed to update integration config: %w", err)
	}

	integration.Config = config
	integration.UpdatedAt = time.Now()

	return nil
}

// DisableIntegration disables an integration
func (m *Manager) DisableIntegration(integrationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	integration, exists := m.integrations[integrationID]
	if !exists {
		return fmt.Errorf("integration %s not found", integrationID)
	}

	integration.Status = IntegrationStatusInactive
	integration.UpdatedAt = time.Now()

	// Close the provider
	if provider, exists := m.providers[integrationID]; exists {
		provider.Close()
	}

	return nil
}

// EnableIntegration enables a disabled integration
func (m *Manager) EnableIntegration(integrationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	integration, exists := m.integrations[integrationID]
	if !exists {
		return fmt.Errorf("integration %s not found", integrationID)
	}

	provider, exists := m.providers[integrationID]
	if !exists {
		return fmt.Errorf("provider not found for integration %s", integrationID)
	}

	// Re-initialize the provider
	ctx, cancel := context.WithTimeout(context.Background(), m.config.DefaultTimeout)
	defer cancel()

	if err := provider.Initialize(ctx, integration.Credentials, integration.Config); err != nil {
		return fmt.Errorf("failed to enable integration: %w", err)
	}

	integration.Status = IntegrationStatusActive
	integration.UpdatedAt = time.Now()

	return nil
}

// GetIntegrationMetrics returns metrics for a specific integration
func (m *Manager) GetIntegrationMetrics(integrationID string) (map[string]interface{}, error) {
	if !m.config.EnableMetrics || m.metrics == nil {
		return nil, fmt.Errorf("metrics not enabled")
	}

	return m.metrics.GetMetrics(integrationID), nil
}

// Close closes all integrations and cleans up resources
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, provider := range m.providers {
		if err := provider.Close(); err != nil {
			if m.logger != nil {
				m.logger.LogError(id, fmt.Errorf("failed to close provider: %w", err))
			}
		}
	}

	return nil
}