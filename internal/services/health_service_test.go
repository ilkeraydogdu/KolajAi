package services

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestNewHealthService(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	service := NewHealthService(db, nil)
	if service == nil {
		t.Fatal("NewHealthService returned nil")
	}

	if service.db != db {
		t.Error("Database not set correctly")
	}

	if len(service.checks) == 0 {
		t.Error("No health checks registered")
	}
}

func TestHealthService_RegisterCheck(t *testing.T) {
	service := NewHealthService(nil, nil)
	
	mockCheck := &MockHealthChecker{
		name:     "test-check",
		critical: true,
	}

	service.RegisterCheck(mockCheck)

	if _, exists := service.checks["test-check"]; !exists {
		t.Error("Health check not registered")
	}
}

func TestHealthService_UnregisterCheck(t *testing.T) {
	service := NewHealthService(nil, nil)
	
	mockCheck := &MockHealthChecker{
		name:     "test-check",
		critical: true,
	}

	service.RegisterCheck(mockCheck)
	service.UnregisterCheck("test-check")

	if _, exists := service.checks["test-check"]; exists {
		t.Error("Health check not unregistered")
	}
}

func TestHealthService_Check(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	service := NewHealthService(db, nil)
	
	// Add a mock healthy check
	service.RegisterCheck(&MockHealthChecker{
		name:     "healthy-check",
		critical: false,
		status:   HealthStatusHealthy,
	})

	// Add a mock degraded check
	service.RegisterCheck(&MockHealthChecker{
		name:     "degraded-check",
		critical: false,
		status:   HealthStatusDegraded,
	})

	ctx := context.Background()
	report := service.Check(ctx)

	if report == nil {
		t.Fatal("Health report is nil")
	}

	if report.Status != HealthStatusDegraded {
		t.Errorf("Expected degraded status, got %v", report.Status)
	}

	if report.Summary.Total < 2 {
		t.Errorf("Expected at least 2 checks, got %d", report.Summary.Total)
	}

	if report.Summary.Degraded == 0 {
		t.Error("Expected degraded checks")
	}
}

func TestHealthService_QuickCheck(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	service := NewHealthService(db, nil)
	
	// Add non-critical check (should be ignored in quick check)
	service.RegisterCheck(&MockHealthChecker{
		name:     "non-critical",
		critical: false,
		status:   HealthStatusUnhealthy,
	})

	ctx := context.Background()
	report := service.QuickCheck(ctx)

	if report == nil {
		t.Fatal("Health report is nil")
	}

	// Should only include critical checks (database check)
	if report.Summary.Total == 0 {
		t.Error("Expected at least one critical check")
	}

	// Non-critical unhealthy check should not affect quick check
	if report.Status == HealthStatusUnhealthy {
		t.Error("Quick check should ignore non-critical checks")
	}
}

func TestHealthService_ServeHTTP(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	service := NewHealthService(db, nil)

	// Test normal health check
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	service.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected JSON content type")
	}

	// Test quick health check
	req = httptest.NewRequest("GET", "/health?quick=true", nil)
	w = httptest.NewRecorder()

	service.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHealthService_UnhealthyResponse(t *testing.T) {
	service := NewHealthService(nil, nil)
	
	// Add critical unhealthy check
	service.RegisterCheck(&MockHealthChecker{
		name:     "critical-unhealthy",
		critical: true,
		status:   HealthStatusUnhealthy,
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	service.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
}

func TestDatabaseHealthCheck(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	check := &DatabaseHealthCheck{db: db}

	if check.Name() != "database" {
		t.Errorf("Expected name 'database', got %s", check.Name())
	}

	if !check.IsCritical() {
		t.Error("Database check should be critical")
	}

	ctx := context.Background()
	result := check.Check(ctx)

	if result.Status != HealthStatusHealthy {
		t.Errorf("Expected healthy status, got %v", result.Status)
	}

	if result.Critical != true {
		t.Error("Database check should be marked as critical")
	}
}

func TestDatabaseHealthCheck_NilDB(t *testing.T) {
	check := &DatabaseHealthCheck{db: nil}

	ctx := context.Background()
	result := check.Check(ctx)

	if result.Status != HealthStatusUnhealthy {
		t.Errorf("Expected unhealthy status, got %v", result.Status)
	}

	if result.Error == "" {
		t.Error("Expected error message for nil database")
	}
}

func TestRedisHealthCheck(t *testing.T) {
	// Test with nil client
	check := &RedisHealthCheck{client: nil}
	ctx := context.Background()
	result := check.Check(ctx)

	if result.Status != HealthStatusDegraded {
		t.Errorf("Expected degraded status for nil client, got %v", result.Status)
	}

	if check.IsCritical() {
		t.Error("Redis check should not be critical")
	}

	if check.Name() != "redis" {
		t.Errorf("Expected name 'redis', got %s", check.Name())
	}
}

func TestMemoryHealthCheck(t *testing.T) {
	check := &MemoryHealthCheck{}

	if check.Name() != "memory" {
		t.Errorf("Expected name 'memory', got %s", check.Name())
	}

	if check.IsCritical() {
		t.Error("Memory check should not be critical")
	}

	ctx := context.Background()
	result := check.Check(ctx)

	// Memory check should always return a status
	if result.Status == "" {
		t.Error("Memory check returned empty status")
	}

	if result.Metadata == nil {
		t.Error("Memory check should include metadata")
	}

	if _, exists := result.Metadata["heap_usage_percent"]; !exists {
		t.Error("Memory check should include heap usage percentage")
	}
}

func TestDiskHealthCheck(t *testing.T) {
	check := &DiskHealthCheck{}

	if check.Name() != "disk" {
		t.Errorf("Expected name 'disk', got %s", check.Name())
	}

	if check.IsCritical() {
		t.Error("Disk check should not be critical")
	}

	ctx := context.Background()
	result := check.Check(ctx)

	// Disk check is not implemented, should return healthy
	if result.Status != HealthStatusHealthy {
		t.Errorf("Expected healthy status, got %v", result.Status)
	}
}

// Mock health checker for testing
type MockHealthChecker struct {
	name     string
	critical bool
	status   HealthStatus
	err      error
	delay    time.Duration
}

func (m *MockHealthChecker) Name() string {
	return m.name
}

func (m *MockHealthChecker) IsCritical() bool {
	return m.critical
}

func (m *MockHealthChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()

	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	check := HealthCheck{
		Name:      m.name,
		Timestamp: start,
		Critical:  m.critical,
		Duration:  time.Since(start),
	}

	if m.err != nil {
		check.Status = HealthStatusUnhealthy
		check.Error = m.err.Error()
	} else {
		check.Status = m.status
		if m.status == HealthStatusHealthy {
			check.Message = "Mock check passed"
		}
	}

	return check
}

func TestCalculateOverallStatus(t *testing.T) {
	service := NewHealthService(nil, nil)

	tests := []struct {
		name     string
		checks   map[string]HealthCheck
		expected HealthStatus
	}{
		{
			name: "all healthy",
			checks: map[string]HealthCheck{
				"check1": {Status: HealthStatusHealthy, Critical: false},
				"check2": {Status: HealthStatusHealthy, Critical: true},
			},
			expected: HealthStatusHealthy,
		},
		{
			name: "one degraded",
			checks: map[string]HealthCheck{
				"check1": {Status: HealthStatusHealthy, Critical: false},
				"check2": {Status: HealthStatusDegraded, Critical: false},
			},
			expected: HealthStatusDegraded,
		},
		{
			name: "critical unhealthy",
			checks: map[string]HealthCheck{
				"check1": {Status: HealthStatusHealthy, Critical: false},
				"check2": {Status: HealthStatusUnhealthy, Critical: true},
			},
			expected: HealthStatusUnhealthy,
		},
		{
			name: "non-critical unhealthy",
			checks: map[string]HealthCheck{
				"check1": {Status: HealthStatusHealthy, Critical: true},
				"check2": {Status: HealthStatusUnhealthy, Critical: false},
			},
			expected: HealthStatusUnhealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculateOverallStatus(tt.checks)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}