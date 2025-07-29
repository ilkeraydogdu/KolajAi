package integrations

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// MockProvider implements IntegrationProvider for testing
type MockProvider struct {
	InitializeFunc   func(ctx context.Context, credentials Credentials, config map[string]interface{}) error
	HealthCheckFunc  func(ctx context.Context) error
	CloseFunc        func() error
	IsHealthy        bool
	CallCount        int
}

func (m *MockProvider) Initialize(ctx context.Context, credentials Credentials, config map[string]interface{}) error {
	m.CallCount++
	if m.InitializeFunc != nil {
		return m.InitializeFunc(ctx, credentials, config)
	}
	return nil
}

func (m *MockProvider) HealthCheck(ctx context.Context) error {
	m.CallCount++
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc(ctx)
	}
	if !m.IsHealthy {
		return errors.New("provider unhealthy")
	}
	return nil
}

func (m *MockProvider) GetCapabilities() []string {
	return []string{"test"}
}

func (m *MockProvider) GetRateLimit() RateLimitInfo {
	return RateLimitInfo{
		RequestsPerMinute: 100,
		RequestsPerSecond: 10,
		RequestsRemaining: 50,
		BurstSize:         20,
		ResetsAt:          time.Now().Add(time.Minute),
	}
}

func (m *MockProvider) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// MockLogger implements IntegrationLogger for testing
type MockLogger struct {
	Requests  []IntegrationRequest
	Responses []IntegrationResponse
	Errors    []error
	Webhooks  []WebhookEvent
}

func (m *MockLogger) LogRequest(integrationID string, request IntegrationRequest) {
	m.Requests = append(m.Requests, request)
}

func (m *MockLogger) LogResponse(integrationID string, response IntegrationResponse) {
	m.Responses = append(m.Responses, response)
}

func (m *MockLogger) LogError(integrationID string, err error) {
	m.Errors = append(m.Errors, err)
}

func (m *MockLogger) LogWebhook(integrationID string, event WebhookEvent) {
	m.Webhooks = append(m.Webhooks, event)
}

// MockMetrics implements IntegrationMetrics for testing
type MockMetrics struct {
	RequestCount int
	ErrorCount   int
	WebhookCount int
}

func (m *MockMetrics) RecordRequest(integrationID, method string, duration time.Duration, success bool) {
	m.RequestCount++
}

func (m *MockMetrics) RecordError(integrationID, errorCode string) {
	m.ErrorCount++
}

func (m *MockMetrics) RecordWebhook(integrationID, eventType string, success bool) {
	m.WebhookCount++
}

func (m *MockMetrics) GetMetrics(integrationID string) map[string]interface{} {
	return map[string]interface{}{
		"requests": m.RequestCount,
		"errors":   m.ErrorCount,
		"webhooks": m.WebhookCount,
	}
}

// MockCache implements IntegrationCache for testing
type MockCache struct {
	storage map[string]interface{}
}

func NewMockCache() *MockCache {
	return &MockCache{
		storage: make(map[string]interface{}),
	}
}

func (m *MockCache) Get(key string) (interface{}, bool) {
	val, exists := m.storage[key]
	return val, exists
}

func (m *MockCache) Set(key string, value interface{}, ttl time.Duration) error {
	m.storage[key] = value
	return nil
}

func (m *MockCache) Delete(key string) error {
	delete(m.storage, key)
	return nil
}

func (m *MockCache) Clear(pattern string) error {
	if pattern == "" {
		m.storage = make(map[string]interface{})
	} else {
		// Simple pattern matching - remove keys that contain the pattern
		for key := range m.storage {
			if strings.Contains(key, pattern) {
				delete(m.storage, key)
			}
		}
	}
	return nil
}

func TestManager_RegisterIntegration(t *testing.T) {
	manager := NewManager(nil)
	
	integration := &Integration{
		ID:       "test-integration",
		Name:     "Test Integration",
		Type:     IntegrationTypeMarketplace,
		Provider: "test-provider",
		Status:   IntegrationStatusActive,
		Credentials: Credentials{
			APIKey:    "test-key",
			APISecret: "test-secret",
		},
		Config: map[string]interface{}{
			"environment": "test",
		},
	}
	
	provider := &MockProvider{
		IsHealthy: true,
	}
	
	err := manager.RegisterIntegration(integration, provider)
	if err != nil {
		t.Errorf("RegisterIntegration() error = %v, want nil", err)
	}
	
	// Verify integration was registered
	retrieved, err := manager.GetIntegration("test-integration")
	if err != nil {
		t.Errorf("GetIntegration() error = %v, want nil", err)
	}
	
	if retrieved.ID != integration.ID {
		t.Errorf("Integration ID = %v, want %v", retrieved.ID, integration.ID)
	}
}

func TestManager_RegisterIntegration_Duplicate(t *testing.T) {
	manager := NewManager(nil)
	
	integration := &Integration{
		ID:   "test-integration",
		Name: "Test Integration",
	}
	
	provider := &MockProvider{}
	
	// Register first time
	err := manager.RegisterIntegration(integration, provider)
	if err != nil {
		t.Errorf("First RegisterIntegration() error = %v, want nil", err)
	}
	
	// Try to register again
	err = manager.RegisterIntegration(integration, provider)
	if err == nil {
		t.Error("Expected error for duplicate registration, got nil")
	}
}

func TestManager_ExecuteRequest_WithCircuitBreaker(t *testing.T) {
	config := &ManagerConfig{
		EnableCircuitBreaker: true,
		EnableCaching:        false,
		EnableMetrics:        false,
		DefaultTimeout:       30 * time.Second,
	}
	
	manager := NewManager(config)
	
	// Set up logger and metrics
	logger := &MockLogger{}
	metrics := &MockMetrics{}
	manager.SetLogger(logger)
	manager.SetMetrics(metrics)
	
	integration := &Integration{
		ID:   "test-integration",
		Name: "Test Integration",
	}
	
	callCount := 0
	provider := &MockProvider{
		HealthCheckFunc: func(ctx context.Context) error {
			callCount++
			if callCount < 3 {
				return errors.New("provider error")
			}
			return nil
		},
	}
	
	err := manager.RegisterIntegration(integration, provider)
	if err != nil {
		t.Fatalf("RegisterIntegration() error = %v", err)
	}
	
	request := IntegrationRequest{
		ID:       "test-request",
		Method:   "GET",
		Endpoint: "/test",
	}
	
	ctx := context.Background()
	
	// First few requests should fail
	for i := 0; i < 2; i++ {
		_, err = manager.ExecuteRequest(ctx, "test-integration", request)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	}
	
	// Circuit breaker might be open now, so further requests might fail immediately
	// This is expected behavior
}

func TestManager_ExecuteRequest_WithCache(t *testing.T) {
	config := &ManagerConfig{
		EnableCircuitBreaker: false,
		EnableCaching:        true,
		EnableMetrics:        false,
		DefaultTimeout:       30 * time.Second,
	}
	
	manager := NewManager(config)
	
	// Set up cache
	cache := NewMockCache()
	manager.SetCache(cache)
	
	integration := &Integration{
		ID:   "test-integration",
		Name: "Test Integration",
	}
	
	provider := &MockProvider{
		IsHealthy: true,
	}
	
	err := manager.RegisterIntegration(integration, provider)
	if err != nil {
		t.Fatalf("RegisterIntegration() error = %v", err)
	}
	
	request := IntegrationRequest{
		ID:       "test-request",
		Method:   "GET",
		Endpoint: "/test",
	}
	
	ctx := context.Background()
	
	// First request should call provider
	response1, err := manager.ExecuteRequest(ctx, "test-integration", request)
	if err != nil {
		t.Errorf("First ExecuteRequest() error = %v, want nil", err)
	}
	
	// Second request should be served from cache
	response2, err := manager.ExecuteRequest(ctx, "test-integration", request)
	if err != nil {
		t.Errorf("Second ExecuteRequest() error = %v, want nil", err)
	}
	
	// Responses should be the same (from cache)
	if response1.ID != response2.ID {
		t.Error("Expected cached response")
	}
}

func TestManager_GetIntegrationsByType(t *testing.T) {
	manager := NewManager(nil)
	
	// Register multiple integrations
	integrations := []*Integration{
		{
			ID:   "marketplace-1",
			Type: IntegrationTypeMarketplace,
		},
		{
			ID:   "marketplace-2",
			Type: IntegrationTypeMarketplace,
		},
		{
			ID:   "payment-1",
			Type: IntegrationTypePayment,
		},
	}
	
	provider := &MockProvider{}
	
	for _, integration := range integrations {
		err := manager.RegisterIntegration(integration, provider)
		if err != nil {
			t.Errorf("RegisterIntegration() error = %v", err)
		}
	}
	
	// Get marketplace integrations
	marketplaceIntegrations := manager.GetIntegrationsByType(IntegrationTypeMarketplace)
	if len(marketplaceIntegrations) != 2 {
		t.Errorf("Marketplace integrations count = %v, want %v", len(marketplaceIntegrations), 2)
	}
	
	// Get payment integrations
	paymentIntegrations := manager.GetIntegrationsByType(IntegrationTypePayment)
	if len(paymentIntegrations) != 1 {
		t.Errorf("Payment integrations count = %v, want %v", len(paymentIntegrations), 1)
	}
}

func TestManager_DisableEnableIntegration(t *testing.T) {
	manager := NewManager(nil)
	
	integration := &Integration{
		ID:     "test-integration",
		Status: IntegrationStatusActive,
	}
	
	provider := &MockProvider{}
	
	err := manager.RegisterIntegration(integration, provider)
	if err != nil {
		t.Fatalf("RegisterIntegration() error = %v", err)
	}
	
	// Disable integration
	err = manager.DisableIntegration("test-integration")
	if err != nil {
		t.Errorf("DisableIntegration() error = %v, want nil", err)
	}
	
	// Check status
	retrieved, _ := manager.GetIntegration("test-integration")
	if retrieved.Status != IntegrationStatusInactive {
		t.Errorf("Status = %v, want %v", retrieved.Status, IntegrationStatusInactive)
	}
	
	// Enable integration
	err = manager.EnableIntegration("test-integration")
	if err != nil {
		t.Errorf("EnableIntegration() error = %v, want nil", err)
	}
	
	// Check status again
	retrieved, _ = manager.GetIntegration("test-integration")
	if retrieved.Status != IntegrationStatusActive {
		t.Errorf("Status = %v, want %v", retrieved.Status, IntegrationStatusActive)
	}
}

func TestManager_UpdateIntegrationConfig(t *testing.T) {
	manager := NewManager(nil)
	
	integration := &Integration{
		ID: "test-integration",
		Config: map[string]interface{}{
			"environment": "test",
		},
	}
	
	provider := &MockProvider{}
	
	err := manager.RegisterIntegration(integration, provider)
	if err != nil {
		t.Fatalf("RegisterIntegration() error = %v", err)
	}
	
	// Update config
	newConfig := map[string]interface{}{
		"environment": "production",
		"timeout":     30,
	}
	
	err = manager.UpdateIntegrationConfig("test-integration", newConfig)
	if err != nil {
		t.Errorf("UpdateIntegrationConfig() error = %v, want nil", err)
	}
	
	// Verify config was updated
	retrieved, _ := manager.GetIntegration("test-integration")
	if retrieved.Config["environment"] != "production" {
		t.Errorf("Config environment = %v, want %v", retrieved.Config["environment"], "production")
	}
}