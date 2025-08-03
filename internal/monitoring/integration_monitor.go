package monitoring

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"kolajAi/internal/integrations/registry"
)

// IntegrationMonitor provides comprehensive monitoring for all integrations
type IntegrationMonitor struct {
	registry         *registry.IntegrationRegistry
	healthCheckers   map[string]*HealthChecker
	metricsCollector *MetricsCollector
	alertManager     *AlertManager
	logger           Logger
	config           *MonitoringConfig
	mutex            sync.RWMutex
	isRunning        bool
	stopChan         chan struct{}
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	HealthCheckInterval    time.Duration `json:"health_check_interval"`
	MetricsInterval       time.Duration `json:"metrics_interval"`
	AlertThreshold        int           `json:"alert_threshold"`
	MaxFailures           int           `json:"max_failures"`
	ResponseTimeThreshold time.Duration `json:"response_time_threshold"`
	EnableDetailedLogs    bool          `json:"enable_detailed_logs"`
	EnableAlerts          bool          `json:"enable_alerts"`
	EnableMetrics         bool          `json:"enable_metrics"`
	RetentionPeriod       time.Duration `json:"retention_period"`
}

// HealthChecker manages health checking for a single integration
type HealthChecker struct {
	IntegrationID   string
	Provider        registry.IntegrationProvider
	Status          HealthStatus
	LastCheck       time.Time
	LastSuccess     time.Time
	FailureCount    int
	ResponseTime    time.Duration
	ErrorHistory    []HealthCheckError
	Metrics         *HealthMetrics
	Config          *HealthCheckConfig
	mutex           sync.RWMutex
}

// HealthStatus represents the health status of an integration
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
	HealthStatusMaintenance HealthStatus = "maintenance"
)

// HealthCheckConfig holds configuration for health checks
type HealthCheckConfig struct {
	Enabled         bool          `json:"enabled"`
	Interval        time.Duration `json:"interval"`
	Timeout         time.Duration `json:"timeout"`
	MaxFailures     int           `json:"max_failures"`
	RetryAttempts   int           `json:"retry_attempts"`
	RetryDelay      time.Duration `json:"retry_delay"`
	AlertOnFailure  bool          `json:"alert_on_failure"`
	AlertOnRecovery bool          `json:"alert_on_recovery"`
}

// HealthCheckError represents a health check error
type HealthCheckError struct {
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error"`
	Duration  time.Duration `json:"duration"`
	Attempt   int       `json:"attempt"`
}

// HealthMetrics holds health-related metrics
type HealthMetrics struct {
	TotalChecks      int64         `json:"total_checks"`
	SuccessfulChecks int64         `json:"successful_checks"`
	FailedChecks     int64         `json:"failed_checks"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	MaxResponseTime  time.Duration `json:"max_response_time"`
	MinResponseTime  time.Duration `json:"min_response_time"`
	Uptime           time.Duration `json:"uptime"`
	Downtime         time.Duration `json:"downtime"`
	LastFailure      *time.Time    `json:"last_failure,omitempty"`
	FailureRate      float64       `json:"failure_rate"`
}

// MetricsCollector collects and aggregates metrics from all integrations
type MetricsCollector struct {
	metrics          map[string]*IntegrationMetrics
	aggregatedMetrics *AggregatedMetrics
	mutex            sync.RWMutex
	config           *MetricsConfig
}

// MetricsConfig holds metrics collection configuration
type MetricsConfig struct {
	Enabled              bool          `json:"enabled"`
	CollectionInterval   time.Duration `json:"collection_interval"`
	RetentionPeriod      time.Duration `json:"retention_period"`
	EnablePerformanceMetrics bool      `json:"enable_performance_metrics"`
	EnableBusinessMetrics bool         `json:"enable_business_metrics"`
	MaxDataPoints        int           `json:"max_data_points"`
}

// IntegrationMetrics holds metrics for a single integration
type IntegrationMetrics struct {
	IntegrationID       string                    `json:"integration_id"`
	IntegrationName     string                    `json:"integration_name"`
	Category            string                    `json:"category"`
	Status              HealthStatus              `json:"status"`
	RequestCount        int64                     `json:"request_count"`
	ErrorCount          int64                     `json:"error_count"`
	AverageResponseTime time.Duration             `json:"average_response_time"`
	ThroughputRPS       float64                   `json:"throughput_rps"`
	ErrorRate           float64                   `json:"error_rate"`
	Availability        float64                   `json:"availability"`
	LastActivity        time.Time                 `json:"last_activity"`
	PerformanceMetrics  *PerformanceMetrics       `json:"performance_metrics,omitempty"`
	BusinessMetrics     *BusinessMetrics          `json:"business_metrics,omitempty"`
	TimeSeriesData      []MetricDataPoint         `json:"time_series_data"`
	Timestamp           time.Time                 `json:"timestamp"`
}

// PerformanceMetrics holds performance-related metrics
type PerformanceMetrics struct {
	CPUUsage        float64       `json:"cpu_usage"`
	MemoryUsage     int64         `json:"memory_usage"`
	NetworkIO       int64         `json:"network_io"`
	DiskIO          int64         `json:"disk_io"`
	CacheHitRate    float64       `json:"cache_hit_rate"`
	QueueSize       int           `json:"queue_size"`
	ActiveConnections int         `json:"active_connections"`
	ResponseTimes   []time.Duration `json:"response_times"`
}

// BusinessMetrics holds business-related metrics
type BusinessMetrics struct {
	ProductsSynced    int64     `json:"products_synced"`
	OrdersProcessed   int64     `json:"orders_processed"`
	TransactionVolume float64   `json:"transaction_volume"`
	Revenue           float64   `json:"revenue"`
	ConversionRate    float64   `json:"conversion_rate"`
	CustomerCount     int64     `json:"customer_count"`
	LastSyncTime      time.Time `json:"last_sync_time"`
}

// MetricDataPoint represents a single data point in time series
type MetricDataPoint struct {
	Timestamp time.Time   `json:"timestamp"`
	Value     interface{} `json:"value"`
	MetricType string     `json:"metric_type"`
}

// AggregatedMetrics holds aggregated metrics across all integrations
type AggregatedMetrics struct {
	TotalIntegrations    int                              `json:"total_integrations"`
	HealthyIntegrations  int                              `json:"healthy_integrations"`
	UnhealthyIntegrations int                             `json:"unhealthy_integrations"`
	DegradedIntegrations int                              `json:"degraded_integrations"`
	OverallAvailability  float64                          `json:"overall_availability"`
	TotalRequests        int64                            `json:"total_requests"`
	TotalErrors          int64                            `json:"total_errors"`
	OverallErrorRate     float64                          `json:"overall_error_rate"`
	CategoryMetrics      map[string]*CategoryMetrics      `json:"category_metrics"`
	RegionMetrics        map[string]*RegionMetrics        `json:"region_metrics"`
	Timestamp            time.Time                        `json:"timestamp"`
}

// CategoryMetrics holds metrics for a specific integration category
type CategoryMetrics struct {
	Category            string    `json:"category"`
	IntegrationCount    int       `json:"integration_count"`
	HealthyCount        int       `json:"healthy_count"`
	UnhealthyCount      int       `json:"unhealthy_count"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	TotalRequests       int64     `json:"total_requests"`
	TotalErrors         int64     `json:"total_errors"`
	ErrorRate           float64   `json:"error_rate"`
	Availability        float64   `json:"availability"`
}

// RegionMetrics holds metrics for a specific region
type RegionMetrics struct {
	Region              string    `json:"region"`
	IntegrationCount    int       `json:"integration_count"`
	HealthyCount        int       `json:"healthy_count"`
	UnhealthyCount      int       `json:"unhealthy_count"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	TotalRequests       int64     `json:"total_requests"`
	TotalErrors         int64     `json:"total_errors"`
	ErrorRate           float64   `json:"error_rate"`
	Availability        float64   `json:"availability"`
}

// AlertManager manages alerts and notifications
type AlertManager struct {
	alerts         []Alert
	alertRules     []AlertRule
	notifications  []NotificationChannel
	mutex          sync.RWMutex
	config         *AlertConfig
}

// AlertConfig holds alert configuration
type AlertConfig struct {
	Enabled                bool          `json:"enabled"`
	EvaluationInterval     time.Duration `json:"evaluation_interval"`
	MaxAlerts              int           `json:"max_alerts"`
	AlertRetentionPeriod   time.Duration `json:"alert_retention_period"`
	DefaultSeverity        AlertSeverity `json:"default_severity"`
	EnableEmailNotifications bool        `json:"enable_email_notifications"`
	EnableSlackNotifications bool        `json:"enable_slack_notifications"`
	EnableWebhookNotifications bool      `json:"enable_webhook_notifications"`
}

// Alert represents an alert
type Alert struct {
	ID            string                 `json:"id"`
	IntegrationID string                 `json:"integration_id"`
	RuleID        string                 `json:"rule_id"`
	Severity      AlertSeverity          `json:"severity"`
	Status        AlertStatus            `json:"status"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Timestamp     time.Time              `json:"timestamp"`
	ResolvedAt    *time.Time             `json:"resolved_at,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
	Annotations   map[string]string      `json:"annotations"`
}

// AlertRule defines conditions for triggering alerts
type AlertRule struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Condition   string        `json:"condition"`
	Threshold   float64       `json:"threshold"`
	Duration    time.Duration `json:"duration"`
	Severity    AlertSeverity `json:"severity"`
	Enabled     bool          `json:"enabled"`
	Labels      map[string]string `json:"labels"`
}

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	SeverityCritical AlertSeverity = "critical"
	SeverityHigh     AlertSeverity = "high"
	SeverityMedium   AlertSeverity = "medium"
	SeverityLow      AlertSeverity = "low"
	SeverityInfo     AlertSeverity = "info"
)

// AlertStatus represents alert status
type AlertStatus string

const (
	AlertStatusFiring   AlertStatus = "firing"
	AlertStatusResolved AlertStatus = "resolved"
	AlertStatusSilenced AlertStatus = "silenced"
)

// NotificationChannel represents a notification channel
type NotificationChannel struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Name     string            `json:"name"`
	Config   map[string]string `json:"config"`
	Enabled  bool              `json:"enabled"`
}

// Logger interface for monitoring logs
type Logger interface {
	Debug(msg string, fields map[string]interface{})
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{})
	Fatal(msg string, fields map[string]interface{})
}

// DefaultLogger provides a default logger implementation
type DefaultLogger struct{}

func (l *DefaultLogger) Debug(msg string, fields map[string]interface{}) {
	log.Printf("[DEBUG] %s %v", msg, fields)
}

func (l *DefaultLogger) Info(msg string, fields map[string]interface{}) {
	log.Printf("[INFO] %s %v", msg, fields)
}

func (l *DefaultLogger) Warn(msg string, fields map[string]interface{}) {
	log.Printf("[WARN] %s %v", msg, fields)
}

func (l *DefaultLogger) Error(msg string, fields map[string]interface{}) {
	log.Printf("[ERROR] %s %v", msg, fields)
}

func (l *DefaultLogger) Fatal(msg string, fields map[string]interface{}) {
	log.Printf("[FATAL] %s %v", msg, fields)
	// In production, this should trigger alerts and potentially restart the service
}

// NewIntegrationMonitor creates a new integration monitor
func NewIntegrationMonitor(registry *registry.IntegrationRegistry, logger Logger) *IntegrationMonitor {
	if logger == nil {
		logger = &DefaultLogger{}
	}

	return &IntegrationMonitor{
		registry:       registry,
		healthCheckers: make(map[string]*HealthChecker),
		metricsCollector: &MetricsCollector{
			metrics: make(map[string]*IntegrationMetrics),
			config: &MetricsConfig{
				Enabled:              true,
				CollectionInterval:   30 * time.Second,
				RetentionPeriod:      24 * time.Hour,
				EnablePerformanceMetrics: true,
				EnableBusinessMetrics: true,
				MaxDataPoints:        1000,
			},
		},
		alertManager: &AlertManager{
			alerts:     []Alert{},
			alertRules: []AlertRule{},
			config: &AlertConfig{
				Enabled:                true,
				EvaluationInterval:     30 * time.Second,
				MaxAlerts:              1000,
				AlertRetentionPeriod:   7 * 24 * time.Hour,
				DefaultSeverity:        SeverityMedium,
				EnableEmailNotifications: true,
				EnableSlackNotifications: true,
				EnableWebhookNotifications: true,
			},
		},
		logger: logger,
		config: &MonitoringConfig{
			HealthCheckInterval:    60 * time.Second,
			MetricsInterval:       30 * time.Second,
			AlertThreshold:        3,
			MaxFailures:           5,
			ResponseTimeThreshold: 10 * time.Second,
			EnableDetailedLogs:    true,
			EnableAlerts:          true,
			EnableMetrics:         true,
			RetentionPeriod:       7 * 24 * time.Hour,
		},
		stopChan: make(chan struct{}),
	}
}

// Start starts the monitoring system
func (im *IntegrationMonitor) Start(ctx context.Context) error {
	im.mutex.Lock()
	if im.isRunning {
		im.mutex.Unlock()
		return fmt.Errorf("monitor is already running")
	}
	im.isRunning = true
	im.mutex.Unlock()

	// Initialize health checkers for all integrations
	integrations := im.registry.ListIntegrations()
	for _, integration := range integrations {
		if integration.IsActive {
			im.initializeHealthChecker(integration)
		}
	}

	// Initialize default alert rules
	im.initializeDefaultAlertRules()

	im.logger.Info("Integration monitor started", map[string]interface{}{
		"integrations_count": len(integrations),
		"health_check_interval": im.config.HealthCheckInterval,
		"metrics_interval": im.config.MetricsInterval,
	})

	// Start monitoring goroutines
	go im.runHealthChecks(ctx)
	go im.runMetricsCollection(ctx)
	go im.runAlertEvaluation(ctx)

	return nil
}

// Stop stops the monitoring system
func (im *IntegrationMonitor) Stop() error {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	if !im.isRunning {
		return fmt.Errorf("monitor is not running")
	}

	close(im.stopChan)
	im.isRunning = false

	im.logger.Info("Integration monitor stopped", nil)
	return nil
}

// initializeHealthChecker initializes a health checker for an integration
func (im *IntegrationMonitor) initializeHealthChecker(integration *registry.IntegrationDefinition) {
	provider, exists := im.registry.GetProvider(integration.ID)
	if !exists {
		im.logger.Warn("Provider not found for integration", map[string]interface{}{
			"integration_id": integration.ID,
			"integration_name": integration.DisplayName,
		})
		return
	}

	healthChecker := &HealthChecker{
		IntegrationID: integration.ID,
		Provider:      provider,
		Status:        HealthStatusUnknown,
		LastCheck:     time.Time{},
		LastSuccess:   time.Time{},
		FailureCount:  0,
		ErrorHistory:  []HealthCheckError{},
		Metrics: &HealthMetrics{
			MinResponseTime: time.Duration(^uint64(0) >> 1), // Max duration
		},
		Config: &HealthCheckConfig{
			Enabled:         true,
			Interval:        im.config.HealthCheckInterval,
			Timeout:         30 * time.Second,
			MaxFailures:     im.config.MaxFailures,
			RetryAttempts:   3,
			RetryDelay:      5 * time.Second,
			AlertOnFailure:  true,
			AlertOnRecovery: true,
		},
	}

	im.mutex.Lock()
	im.healthCheckers[integration.ID] = healthChecker
	im.mutex.Unlock()
}

// runHealthChecks runs health checks for all integrations
func (im *IntegrationMonitor) runHealthChecks(ctx context.Context) {
	ticker := time.NewTicker(im.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-im.stopChan:
			return
		case <-ticker.C:
			im.performHealthChecks(ctx)
		}
	}
}

// performHealthChecks performs health checks for all integrations
func (im *IntegrationMonitor) performHealthChecks(ctx context.Context) {
	im.mutex.RLock()
	checkers := make(map[string]*HealthChecker)
	for k, v := range im.healthCheckers {
		checkers[k] = v
	}
	im.mutex.RUnlock()

	var wg sync.WaitGroup
	for integrationID, checker := range checkers {
		if !checker.Config.Enabled {
			continue
		}

		wg.Add(1)
		go func(id string, hc *HealthChecker) {
			defer wg.Done()
			im.performSingleHealthCheck(ctx, id, hc)
		}(integrationID, checker)
	}

	wg.Wait()
}

// performSingleHealthCheck performs a health check for a single integration
func (im *IntegrationMonitor) performSingleHealthCheck(ctx context.Context, integrationID string, checker *HealthChecker) {
	checker.mutex.Lock()
	defer checker.mutex.Unlock()

	startTime := time.Now()
	checker.LastCheck = startTime

	// Create context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, checker.Config.Timeout)
	defer cancel()

	// Perform health check with retries
	var lastErr error
	success := false

	for attempt := 0; attempt <= checker.Config.RetryAttempts; attempt++ {
		err := checker.Provider.HealthCheck(checkCtx)
		if err == nil {
			success = true
			break
		}

		lastErr = err
		if attempt < checker.Config.RetryAttempts {
			time.Sleep(checker.Config.RetryDelay)
		}
	}

	// Calculate response time
	responseTime := time.Since(startTime)
	checker.ResponseTime = responseTime

	// Update metrics
	checker.Metrics.TotalChecks++
	if checker.Metrics.MinResponseTime > responseTime {
		checker.Metrics.MinResponseTime = responseTime
	}
	if checker.Metrics.MaxResponseTime < responseTime {
		checker.Metrics.MaxResponseTime = responseTime
	}

	// Calculate average response time
	totalTime := time.Duration(checker.Metrics.TotalChecks-1) * checker.Metrics.AverageResponseTime + responseTime
	checker.Metrics.AverageResponseTime = totalTime / time.Duration(checker.Metrics.TotalChecks)

	if success {
		// Health check succeeded
		checker.Metrics.SuccessfulChecks++
		previousStatus := checker.Status
		checker.Status = HealthStatusHealthy
		checker.LastSuccess = startTime
		checker.FailureCount = 0

		// Check for recovery
		if previousStatus == HealthStatusUnhealthy && checker.Config.AlertOnRecovery {
			im.triggerRecoveryAlert(integrationID, checker)
		}

		im.logger.Debug("Health check succeeded", map[string]interface{}{
			"integration_id": integrationID,
			"response_time": responseTime,
		})

	} else {
		// Health check failed
		checker.Metrics.FailedChecks++
		checker.FailureCount++

		// Add to error history
		errorEntry := HealthCheckError{
			Timestamp: startTime,
			Error:     lastErr.Error(),
			Duration:  responseTime,
			Attempt:   checker.Config.RetryAttempts + 1,
		}
		checker.ErrorHistory = append(checker.ErrorHistory, errorEntry)

		// Limit error history size
		if len(checker.ErrorHistory) > 100 {
			checker.ErrorHistory = checker.ErrorHistory[1:]
		}

		// Update status based on failure count
		previousStatus := checker.Status
		if checker.FailureCount >= checker.Config.MaxFailures {
			checker.Status = HealthStatusUnhealthy
		} else {
			checker.Status = HealthStatusDegraded
		}

		// Check for failure alert
		if previousStatus != HealthStatusUnhealthy && checker.Status == HealthStatusUnhealthy && checker.Config.AlertOnFailure {
			im.triggerFailureAlert(integrationID, checker, lastErr)
		}

		im.logger.Error("Health check failed", map[string]interface{}{
			"integration_id": integrationID,
			"error": lastErr.Error(),
			"failure_count": checker.FailureCount,
			"response_time": responseTime,
		})

		// Update failure time
		now := time.Now()
		checker.Metrics.LastFailure = &now
	}

	// Calculate failure rate
	if checker.Metrics.TotalChecks > 0 {
		checker.Metrics.FailureRate = float64(checker.Metrics.FailedChecks) / float64(checker.Metrics.TotalChecks) * 100
	}
}

// runMetricsCollection runs metrics collection
func (im *IntegrationMonitor) runMetricsCollection(ctx context.Context) {
	if !im.config.EnableMetrics {
		return
	}

	ticker := time.NewTicker(im.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-im.stopChan:
			return
		case <-ticker.C:
			im.collectMetrics()
		}
	}
}

// collectMetrics collects metrics from all integrations
func (im *IntegrationMonitor) collectMetrics() {
	im.metricsCollector.mutex.Lock()
	defer im.metricsCollector.mutex.Unlock()

	integrations := im.registry.ListIntegrations()
	timestamp := time.Now()

	// Collect metrics for each integration
	for _, integration := range integrations {
		if !integration.IsActive {
			continue
		}

		metrics := im.collectIntegrationMetrics(integration, timestamp)
		im.metricsCollector.metrics[integration.ID] = metrics
	}

	// Calculate aggregated metrics
	im.calculateAggregatedMetrics(timestamp)

	// Clean up old data points
	im.cleanupOldMetrics()
}

// collectIntegrationMetrics collects metrics for a single integration
func (im *IntegrationMonitor) collectIntegrationMetrics(integration *registry.IntegrationDefinition, timestamp time.Time) *IntegrationMetrics {
	// Get health checker
	im.mutex.RLock()
	healthChecker, exists := im.healthCheckers[integration.ID]
	im.mutex.RUnlock()

	metrics := &IntegrationMetrics{
		IntegrationID:   integration.ID,
		IntegrationName: integration.DisplayName,
		Category:        integration.Category,
		Status:          HealthStatusUnknown,
		LastActivity:    timestamp,
		TimeSeriesData:  []MetricDataPoint{},
		Timestamp:       timestamp,
	}

	if exists {
		healthChecker.mutex.RLock()
		metrics.Status = healthChecker.Status
		metrics.AverageResponseTime = healthChecker.Metrics.AverageResponseTime
		
		// Calculate availability
		if healthChecker.Metrics.TotalChecks > 0 {
			metrics.Availability = float64(healthChecker.Metrics.SuccessfulChecks) / float64(healthChecker.Metrics.TotalChecks) * 100
		}
		
		metrics.ErrorRate = healthChecker.Metrics.FailureRate
		healthChecker.mutex.RUnlock()
	}

	// Get provider metrics if available
	if provider, exists := im.registry.GetProvider(integration.ID); exists {
		providerMetrics := provider.GetMetrics()
		
		if requestCount, ok := providerMetrics["request_count"].(int64); ok {
			metrics.RequestCount = requestCount
		}
		if errorCount, ok := providerMetrics["error_count"].(int64); ok {
			metrics.ErrorCount = errorCount
		}
		if throughput, ok := providerMetrics["throughput_rps"].(float64); ok {
			metrics.ThroughputRPS = throughput
		}

		// Collect performance metrics if enabled
		if im.metricsCollector.config.EnablePerformanceMetrics {
			metrics.PerformanceMetrics = im.extractPerformanceMetrics(providerMetrics)
		}

		// Collect business metrics if enabled
		if im.metricsCollector.config.EnableBusinessMetrics {
			metrics.BusinessMetrics = im.extractBusinessMetrics(providerMetrics)
		}
	}

	// Add time series data point
	dataPoint := MetricDataPoint{
		Timestamp:  timestamp,
		Value:      metrics.Status,
		MetricType: "health_status",
	}
	metrics.TimeSeriesData = append(metrics.TimeSeriesData, dataPoint)

	return metrics
}

// extractPerformanceMetrics extracts performance metrics from provider metrics
func (im *IntegrationMonitor) extractPerformanceMetrics(providerMetrics map[string]interface{}) *PerformanceMetrics {
	perfMetrics := &PerformanceMetrics{}

	if cpuUsage, ok := providerMetrics["cpu_usage"].(float64); ok {
		perfMetrics.CPUUsage = cpuUsage
	}
	if memUsage, ok := providerMetrics["memory_usage"].(int64); ok {
		perfMetrics.MemoryUsage = memUsage
	}
	if networkIO, ok := providerMetrics["network_io"].(int64); ok {
		perfMetrics.NetworkIO = networkIO
	}
	if diskIO, ok := providerMetrics["disk_io"].(int64); ok {
		perfMetrics.DiskIO = diskIO
	}
	if cacheHitRate, ok := providerMetrics["cache_hit_rate"].(float64); ok {
		perfMetrics.CacheHitRate = cacheHitRate
	}
	if queueSize, ok := providerMetrics["queue_size"].(int); ok {
		perfMetrics.QueueSize = queueSize
	}
	if activeConns, ok := providerMetrics["active_connections"].(int); ok {
		perfMetrics.ActiveConnections = activeConns
	}

	return perfMetrics
}

// extractBusinessMetrics extracts business metrics from provider metrics
func (im *IntegrationMonitor) extractBusinessMetrics(providerMetrics map[string]interface{}) *BusinessMetrics {
	bizMetrics := &BusinessMetrics{}

	if productsSynced, ok := providerMetrics["products_synced"].(int64); ok {
		bizMetrics.ProductsSynced = productsSynced
	}
	if ordersProcessed, ok := providerMetrics["orders_processed"].(int64); ok {
		bizMetrics.OrdersProcessed = ordersProcessed
	}
	if transactionVolume, ok := providerMetrics["transaction_volume"].(float64); ok {
		bizMetrics.TransactionVolume = transactionVolume
	}
	if revenue, ok := providerMetrics["revenue"].(float64); ok {
		bizMetrics.Revenue = revenue
	}
	if conversionRate, ok := providerMetrics["conversion_rate"].(float64); ok {
		bizMetrics.ConversionRate = conversionRate
	}
	if customerCount, ok := providerMetrics["customer_count"].(int64); ok {
		bizMetrics.CustomerCount = customerCount
	}
	if lastSync, ok := providerMetrics["last_sync_time"].(time.Time); ok {
		bizMetrics.LastSyncTime = lastSync
	}

	return bizMetrics
}

// calculateAggregatedMetrics calculates aggregated metrics across all integrations
func (im *IntegrationMonitor) calculateAggregatedMetrics(timestamp time.Time) {
	aggregated := &AggregatedMetrics{
		CategoryMetrics: make(map[string]*CategoryMetrics),
		RegionMetrics:   make(map[string]*RegionMetrics),
		Timestamp:       timestamp,
	}

	categoryStats := make(map[string]*CategoryMetrics)
	regionStats := make(map[string]*RegionMetrics)

	for _, metrics := range im.metricsCollector.metrics {
		aggregated.TotalIntegrations++
		aggregated.TotalRequests += metrics.RequestCount
		aggregated.TotalErrors += metrics.ErrorCount

		// Count by status
		switch metrics.Status {
		case HealthStatusHealthy:
			aggregated.HealthyIntegrations++
		case HealthStatusUnhealthy:
			aggregated.UnhealthyIntegrations++
		case HealthStatusDegraded:
			aggregated.DegradedIntegrations++
		}

		// Category metrics
		if categoryStats[metrics.Category] == nil {
			categoryStats[metrics.Category] = &CategoryMetrics{
				Category: metrics.Category,
			}
		}
		catMetrics := categoryStats[metrics.Category]
		catMetrics.IntegrationCount++
		catMetrics.TotalRequests += metrics.RequestCount
		catMetrics.TotalErrors += metrics.ErrorCount

		if metrics.Status == HealthStatusHealthy {
			catMetrics.HealthyCount++
		} else if metrics.Status == HealthStatusUnhealthy {
			catMetrics.UnhealthyCount++
		}

		// Calculate average response time for category
		totalTime := time.Duration(catMetrics.IntegrationCount-1) * catMetrics.AverageResponseTime + metrics.AverageResponseTime
		catMetrics.AverageResponseTime = totalTime / time.Duration(catMetrics.IntegrationCount)
	}

	// Calculate overall metrics
	if aggregated.TotalIntegrations > 0 {
		aggregated.OverallAvailability = float64(aggregated.HealthyIntegrations) / float64(aggregated.TotalIntegrations) * 100
	}
	if aggregated.TotalRequests > 0 {
		aggregated.OverallErrorRate = float64(aggregated.TotalErrors) / float64(aggregated.TotalRequests) * 100
	}

	// Calculate category metrics
	for _, catMetrics := range categoryStats {
		if catMetrics.IntegrationCount > 0 {
			catMetrics.Availability = float64(catMetrics.HealthyCount) / float64(catMetrics.IntegrationCount) * 100
		}
		if catMetrics.TotalRequests > 0 {
			catMetrics.ErrorRate = float64(catMetrics.TotalErrors) / float64(catMetrics.TotalRequests) * 100
		}
	}

	aggregated.CategoryMetrics = categoryStats
	aggregated.RegionMetrics = regionStats

	im.metricsCollector.aggregatedMetrics = aggregated
}

// cleanupOldMetrics removes old metric data points
func (im *IntegrationMonitor) cleanupOldMetrics() {
	cutoffTime := time.Now().Add(-im.metricsCollector.config.RetentionPeriod)

	for _, metrics := range im.metricsCollector.metrics {
		// Remove old time series data points
		var newDataPoints []MetricDataPoint
		for _, dataPoint := range metrics.TimeSeriesData {
			if dataPoint.Timestamp.After(cutoffTime) {
				newDataPoints = append(newDataPoints, dataPoint)
			}
		}
		metrics.TimeSeriesData = newDataPoints
	}
}

// runAlertEvaluation runs alert evaluation
func (im *IntegrationMonitor) runAlertEvaluation(ctx context.Context) {
	if !im.config.EnableAlerts {
		return
	}

	ticker := time.NewTicker(im.alertManager.config.EvaluationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-im.stopChan:
			return
		case <-ticker.C:
			im.evaluateAlerts()
		}
	}
}

// evaluateAlerts evaluates all alert rules
func (im *IntegrationMonitor) evaluateAlerts() {
	im.alertManager.mutex.RLock()
	rules := make([]AlertRule, len(im.alertManager.alertRules))
	copy(rules, im.alertManager.alertRules)
	im.alertManager.mutex.RUnlock()

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		im.evaluateAlertRule(rule)
	}

	// Clean up old alerts
	im.cleanupOldAlerts()
}

// evaluateAlertRule evaluates a single alert rule
func (im *IntegrationMonitor) evaluateAlertRule(rule AlertRule) {
	// This is a simplified implementation
	// In a real system, you would have a more sophisticated rule evaluation engine
	
	switch rule.Condition {
	case "error_rate_high":
		im.evaluateErrorRateAlert(rule)
	case "response_time_high":
		im.evaluateResponseTimeAlert(rule)
	case "availability_low":
		im.evaluateAvailabilityAlert(rule)
	case "integration_down":
		im.evaluateIntegrationDownAlert(rule)
	}
}

// Helper methods for alert evaluation
func (im *IntegrationMonitor) evaluateErrorRateAlert(rule AlertRule) {
	im.metricsCollector.mutex.RLock()
	defer im.metricsCollector.mutex.RUnlock()

	for integrationID, metrics := range im.metricsCollector.metrics {
		if metrics.ErrorRate > rule.Threshold {
			im.createAlert(integrationID, rule, fmt.Sprintf("Error rate %.2f%% exceeds threshold %.2f%%", metrics.ErrorRate, rule.Threshold))
		}
	}
}

func (im *IntegrationMonitor) evaluateResponseTimeAlert(rule AlertRule) {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	for integrationID, checker := range im.healthCheckers {
		checker.mutex.RLock()
		responseTime := checker.Metrics.AverageResponseTime
		checker.mutex.RUnlock()

		if responseTime > time.Duration(rule.Threshold)*time.Millisecond {
			im.createAlert(integrationID, rule, fmt.Sprintf("Response time %v exceeds threshold %v", responseTime, time.Duration(rule.Threshold)*time.Millisecond))
		}
	}
}

func (im *IntegrationMonitor) evaluateAvailabilityAlert(rule AlertRule) {
	im.metricsCollector.mutex.RLock()
	defer im.metricsCollector.mutex.RUnlock()

	for integrationID, metrics := range im.metricsCollector.metrics {
		if metrics.Availability < rule.Threshold {
			im.createAlert(integrationID, rule, fmt.Sprintf("Availability %.2f%% below threshold %.2f%%", metrics.Availability, rule.Threshold))
		}
	}
}

func (im *IntegrationMonitor) evaluateIntegrationDownAlert(rule AlertRule) {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	for integrationID, checker := range im.healthCheckers {
		checker.mutex.RLock()
		status := checker.Status
		checker.mutex.RUnlock()

		if status == HealthStatusUnhealthy {
			im.createAlert(integrationID, rule, "Integration is unhealthy")
		}
	}
}

// createAlert creates a new alert
func (im *IntegrationMonitor) createAlert(integrationID string, rule AlertRule, description string) {
	alertID := fmt.Sprintf("%s_%s_%d", integrationID, rule.ID, time.Now().Unix())
	
	alert := Alert{
		ID:            alertID,
		IntegrationID: integrationID,
		RuleID:        rule.ID,
		Severity:      rule.Severity,
		Status:        AlertStatusFiring,
		Title:         rule.Name,
		Description:   description,
		Timestamp:     time.Now(),
		Metadata:      make(map[string]interface{}),
		Annotations:   rule.Labels,
	}

	im.alertManager.mutex.Lock()
	im.alertManager.alerts = append(im.alertManager.alerts, alert)
	im.alertManager.mutex.Unlock()

	im.logger.Warn("Alert triggered", map[string]interface{}{
		"alert_id": alertID,
		"integration_id": integrationID,
		"rule_id": rule.ID,
		"severity": rule.Severity,
		"description": description,
	})

	// Send notifications
	im.sendAlertNotifications(alert)
}

// sendAlertNotifications sends alert notifications
func (im *IntegrationMonitor) sendAlertNotifications(alert Alert) {
	// This would implement actual notification sending
	// For now, just log the notification
	im.logger.Info("Alert notification sent", map[string]interface{}{
		"alert_id": alert.ID,
		"severity": alert.Severity,
		"title": alert.Title,
	})
}

// triggerFailureAlert triggers an alert when an integration fails
func (im *IntegrationMonitor) triggerFailureAlert(integrationID string, checker *HealthChecker, err error) {
	rule := AlertRule{
		ID:        "integration_failure",
		Name:      "Integration Failure",
		Severity:  SeverityHigh,
		Condition: "integration_down",
	}

	im.createAlert(integrationID, rule, fmt.Sprintf("Integration failed: %v", err))
}

// triggerRecoveryAlert triggers an alert when an integration recovers
func (im *IntegrationMonitor) triggerRecoveryAlert(integrationID string, checker *HealthChecker) {
	im.logger.Info("Integration recovered", map[string]interface{}{
		"integration_id": integrationID,
		"downtime": time.Since(checker.LastSuccess),
	})
}

// cleanupOldAlerts removes old resolved alerts
func (im *IntegrationMonitor) cleanupOldAlerts() {
	im.alertManager.mutex.Lock()
	defer im.alertManager.mutex.Unlock()

	cutoffTime := time.Now().Add(-im.alertManager.config.AlertRetentionPeriod)
	var activeAlerts []Alert

	for _, alert := range im.alertManager.alerts {
		if alert.Status == AlertStatusFiring || (alert.ResolvedAt != nil && alert.ResolvedAt.After(cutoffTime)) {
			activeAlerts = append(activeAlerts, alert)
		}
	}

	im.alertManager.alerts = activeAlerts
}

// initializeDefaultAlertRules initializes default alert rules
func (im *IntegrationMonitor) initializeDefaultAlertRules() {
	defaultRules := []AlertRule{
		{
			ID:          "high_error_rate",
			Name:        "High Error Rate",
			Description: "Alert when error rate exceeds threshold",
			Condition:   "error_rate_high",
			Threshold:   5.0, // 5%
			Duration:    5 * time.Minute,
			Severity:    SeverityHigh,
			Enabled:     true,
			Labels:      map[string]string{"type": "error_rate"},
		},
		{
			ID:          "high_response_time",
			Name:        "High Response Time",
			Description: "Alert when response time exceeds threshold",
			Condition:   "response_time_high",
			Threshold:   10000, // 10 seconds in milliseconds
			Duration:    5 * time.Minute,
			Severity:    SeverityMedium,
			Enabled:     true,
			Labels:      map[string]string{"type": "performance"},
		},
		{
			ID:          "low_availability",
			Name:        "Low Availability",
			Description: "Alert when availability drops below threshold",
			Condition:   "availability_low",
			Threshold:   95.0, // 95%
			Duration:    10 * time.Minute,
			Severity:    SeverityHigh,
			Enabled:     true,
			Labels:      map[string]string{"type": "availability"},
		},
		{
			ID:          "integration_down",
			Name:        "Integration Down",
			Description: "Alert when integration is unhealthy",
			Condition:   "integration_down",
			Threshold:   0,
			Duration:    2 * time.Minute,
			Severity:    SeverityCritical,
			Enabled:     true,
			Labels:      map[string]string{"type": "health"},
		},
	}

	im.alertManager.mutex.Lock()
	im.alertManager.alertRules = defaultRules
	im.alertManager.mutex.Unlock()
}

// Public API methods

// GetIntegrationHealth returns health status for a specific integration
func (im *IntegrationMonitor) GetIntegrationHealth(integrationID string) (*HealthChecker, bool) {
	im.mutex.RLock()
	defer im.mutex.RUnlock()
	
	checker, exists := im.healthCheckers[integrationID]
	return checker, exists
}

// GetAllHealthStatus returns health status for all integrations
func (im *IntegrationMonitor) GetAllHealthStatus() map[string]*HealthChecker {
	im.mutex.RLock()
	defer im.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
	healthStatus := make(map[string]*HealthChecker)
	for k, v := range im.healthCheckers {
		healthStatus[k] = v
	}
	
	return healthStatus
}

// GetMetrics returns metrics for a specific integration
func (im *IntegrationMonitor) GetMetrics(integrationID string) (*IntegrationMetrics, bool) {
	im.metricsCollector.mutex.RLock()
	defer im.metricsCollector.mutex.RUnlock()
	
	metrics, exists := im.metricsCollector.metrics[integrationID]
	return metrics, exists
}

// GetAggregatedMetrics returns aggregated metrics across all integrations
func (im *IntegrationMonitor) GetAggregatedMetrics() *AggregatedMetrics {
	im.metricsCollector.mutex.RLock()
	defer im.metricsCollector.mutex.RUnlock()
	
	return im.metricsCollector.aggregatedMetrics
}

// GetActiveAlerts returns all active alerts
func (im *IntegrationMonitor) GetActiveAlerts() []Alert {
	im.alertManager.mutex.RLock()
	defer im.alertManager.mutex.RUnlock()
	
	var activeAlerts []Alert
	for _, alert := range im.alertManager.alerts {
		if alert.Status == AlertStatusFiring {
			activeAlerts = append(activeAlerts, alert)
		}
	}
	
	return activeAlerts
}

// GetMonitoringStatus returns the overall monitoring system status
func (im *IntegrationMonitor) GetMonitoringStatus() map[string]interface{} {
	im.mutex.RLock()
	isRunning := im.isRunning
	healthCheckersCount := len(im.healthCheckers)
	im.mutex.RUnlock()

	im.metricsCollector.mutex.RLock()
	metricsCount := len(im.metricsCollector.metrics)
	im.metricsCollector.mutex.RUnlock()

	im.alertManager.mutex.RLock()
	alertsCount := len(im.alertManager.alerts)
	alertRulesCount := len(im.alertManager.alertRules)
	im.alertManager.mutex.RUnlock()

	return map[string]interface{}{
		"is_running":            isRunning,
		"health_checkers_count": healthCheckersCount,
		"metrics_count":         metricsCount,
		"alerts_count":          alertsCount,
		"alert_rules_count":     alertRulesCount,
		"config":                im.config,
	}
}