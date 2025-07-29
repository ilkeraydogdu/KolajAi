package services

import (
	"fmt"
	"sync"
	"time"
)

// MonitoringService provides comprehensive monitoring for integrations
type MonitoringService struct {
	alerts      map[string]*Alert
	metrics     map[string]*IntegrationMetrics
	healthChecks map[string]*HealthCheck
	notifications []Notification
	mu          sync.RWMutex
}

// Alert represents a monitoring alert
type Alert struct {
	ID            string                 `json:"id"`
	IntegrationID string                 `json:"integration_id"`
	Type          AlertType              `json:"type"`
	Severity      AlertSeverity          `json:"severity"`
	Message       string                 `json:"message"`
	Details       map[string]interface{} `json:"details"`
	CreatedAt     time.Time              `json:"created_at"`
	ResolvedAt    *time.Time             `json:"resolved_at"`
	IsActive      bool                   `json:"is_active"`
}

// AlertType represents different types of alerts
type AlertType string

const (
	AlertTypeHighErrorRate    AlertType = "high_error_rate"
	AlertTypeHighResponseTime AlertType = "high_response_time"
	AlertTypeCircuitBreaker   AlertType = "circuit_breaker"
	AlertTypeRateLimit        AlertType = "rate_limit"
	AlertTypeConnectionFailed AlertType = "connection_failed"
	AlertTypeLowSuccessRate   AlertType = "low_success_rate"
)

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "info"
	SeverityWarning  AlertSeverity = "warning"
	SeverityError    AlertSeverity = "error"
	SeverityCritical AlertSeverity = "critical"
)

// HealthCheck represents a health check for an integration
type HealthCheck struct {
	IntegrationID string
	Status        HealthStatus
	LastCheck     time.Time
	ResponseTime  time.Duration
	ErrorCount    int
	SuccessCount  int
	Details       map[string]interface{}
}

// HealthStatus represents health check status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// Notification represents a notification
type Notification struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Severity  AlertSeverity          `json:"severity"`
	Channel   string                 `json:"channel"` // email, slack, webhook
	Recipient string                 `json:"recipient"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	SentAt    *time.Time             `json:"sent_at"`
	IsSent    bool                   `json:"is_sent"`
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	CheckInterval     time.Duration `json:"check_interval"`
	AlertThresholds   AlertThresholds `json:"alert_thresholds"`
	NotificationChannels []string    `json:"notification_channels"`
	RetentionPeriod   time.Duration `json:"retention_period"`
}

// AlertThresholds holds alert threshold configuration
type AlertThresholds struct {
	ErrorRateThreshold     float64 `json:"error_rate_threshold"`
	ResponseTimeThreshold  time.Duration `json:"response_time_threshold"`
	SuccessRateThreshold   float64 `json:"success_rate_threshold"`
	CircuitBreakerThreshold int    `json:"circuit_breaker_threshold"`
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(config *MonitoringConfig) *MonitoringService {
	return &MonitoringService{
		alerts:        make(map[string]*Alert),
		metrics:       make(map[string]*IntegrationMetrics),
		healthChecks:  make(map[string]*HealthCheck),
		notifications: []Notification{},
	}
}

// MonitorIntegration monitors an integration
func (ms *MonitoringService) MonitorIntegration(integrationID string, metrics *IntegrationMetrics) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.metrics[integrationID] = metrics

	// Check for alerts
	ms.checkForAlerts(integrationID, metrics)
}

// checkForAlerts checks if alerts should be triggered
func (ms *MonitoringService) checkForAlerts(integrationID string, metrics *IntegrationMetrics) {
	// Check error rate
	if metrics.ErrorRate > 0.1 { // 10% error rate threshold
		ms.createAlert(integrationID, AlertTypeHighErrorRate, SeverityError, 
			fmt.Sprintf("High error rate detected: %.2f%%", metrics.ErrorRate*100))
	}

	// Check response time
	if metrics.AverageResponseTime > 5*time.Second {
		ms.createAlert(integrationID, AlertTypeHighResponseTime, SeverityWarning,
			fmt.Sprintf("High response time detected: %v", metrics.AverageResponseTime))
	}

	// Check success rate
	if metrics.SuccessRate < 0.9 { // 90% success rate threshold
		ms.createAlert(integrationID, AlertTypeLowSuccessRate, SeverityWarning,
			fmt.Sprintf("Low success rate detected: %.2f%%", metrics.SuccessRate*100))
	}
}

// createAlert creates a new alert
func (ms *MonitoringService) createAlert(integrationID string, alertType AlertType, severity AlertSeverity, message string) {
	alert := &Alert{
		ID:            generateAlertID(),
		IntegrationID: integrationID,
		Type:          alertType,
		Severity:      severity,
		Message:       message,
		Details:       make(map[string]interface{}),
		CreatedAt:     time.Now(),
		IsActive:      true,
	}

	ms.alerts[alert.ID] = alert

	// Create notification
	ms.createNotification(alert)
}

// createNotification creates a notification for an alert
func (ms *MonitoringService) createNotification(alert *Alert) {
	notification := Notification{
		ID:        generateNotificationID(),
		Type:      "alert",
		Title:     fmt.Sprintf("Alert: %s", alert.Type),
		Message:   alert.Message,
		Severity:  alert.Severity,
		Channel:   "email", // Default channel
		Recipient: "admin@kolajai.com",
		Data: map[string]interface{}{
			"alert_id":       alert.ID,
			"integration_id": alert.IntegrationID,
			"alert_type":     alert.Type,
		},
		CreatedAt: time.Now(),
		IsSent:    false,
	}

	ms.notifications = append(ms.notifications, notification)
}

// GetAlerts returns all active alerts
func (ms *MonitoringService) GetAlerts() []*Alert {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var activeAlerts []*Alert
	for _, alert := range ms.alerts {
		if alert.IsActive {
			activeAlerts = append(activeAlerts, alert)
		}
	}

	return activeAlerts
}

// GetAlertsByIntegration returns alerts for a specific integration
func (ms *MonitoringService) GetAlertsByIntegration(integrationID string) []*Alert {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var integrationAlerts []*Alert
	for _, alert := range ms.alerts {
		if alert.IntegrationID == integrationID && alert.IsActive {
			integrationAlerts = append(integrationAlerts, alert)
		}
	}

	return integrationAlerts
}

// ResolveAlert resolves an alert
func (ms *MonitoringService) ResolveAlert(alertID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	alert, exists := ms.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	now := time.Now()
	alert.IsActive = false
	alert.ResolvedAt = &now

	return nil
}

// GetHealthStatus returns health status for an integration
func (ms *MonitoringService) GetHealthStatus(integrationID string) *HealthCheck {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	healthCheck, exists := ms.healthChecks[integrationID]
	if !exists {
		return &HealthCheck{
			IntegrationID: integrationID,
			Status:        HealthStatusUnknown,
			LastCheck:     time.Now(),
		}
	}

	return healthCheck
}

// UpdateHealthCheck updates health check for an integration
func (ms *MonitoringService) UpdateHealthCheck(integrationID string, status HealthStatus, responseTime time.Duration, err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	healthCheck, exists := ms.healthChecks[integrationID]
	if !exists {
		healthCheck = &HealthCheck{
			IntegrationID: integrationID,
			Details:       make(map[string]interface{}),
		}
		ms.healthChecks[integrationID] = healthCheck
	}

	healthCheck.Status = status
	healthCheck.LastCheck = time.Now()
	healthCheck.ResponseTime = responseTime

	if err != nil {
		healthCheck.ErrorCount++
		healthCheck.Details["last_error"] = err.Error()
	} else {
		healthCheck.SuccessCount++
	}

	// Update status based on error rate
	totalChecks := healthCheck.SuccessCount + healthCheck.ErrorCount
	if totalChecks > 0 {
		errorRate := float64(healthCheck.ErrorCount) / float64(totalChecks)
		if errorRate > 0.5 {
			healthCheck.Status = HealthStatusUnhealthy
		} else if errorRate > 0.1 {
			healthCheck.Status = HealthStatusDegraded
		} else {
			healthCheck.Status = HealthStatusHealthy
		}
	}
}

// GetNotifications returns all notifications
func (ms *MonitoringService) GetNotifications() []Notification {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return ms.notifications
}

// MarkNotificationSent marks a notification as sent
func (ms *MonitoringService) MarkNotificationSent(notificationID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for i, notification := range ms.notifications {
		if notification.ID == notificationID {
			now := time.Now()
			ms.notifications[i].IsSent = true
			ms.notifications[i].SentAt = &now
			return nil
		}
	}

	return fmt.Errorf("notification not found: %s", notificationID)
}

// GetMetrics returns metrics for an integration
func (ms *MonitoringService) GetMetrics(integrationID string) (*IntegrationMetrics, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	metrics, exists := ms.metrics[integrationID]
	return metrics, exists
}

// GetAllMetrics returns all metrics
func (ms *MonitoringService) GetAllMetrics() map[string]*IntegrationMetrics {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	result := make(map[string]*IntegrationMetrics)
	for id, metrics := range ms.metrics {
		result[id] = metrics
	}

	return result
}

// generateAlertID generates a unique alert ID
func generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

// generateNotificationID generates a unique notification ID
func generateNotificationID() string {
	return fmt.Sprintf("notification_%d", time.Now().UnixNano())
}