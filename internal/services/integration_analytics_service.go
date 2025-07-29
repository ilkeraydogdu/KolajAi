package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"
)

// IntegrationAnalyticsService provides comprehensive analytics for integrations
type IntegrationAnalyticsService struct {
	db                 *sql.DB
	marketplaceService *MarketplaceIntegrationsService
	aiManager         *AIIntegrationManager
	metrics           *IntegrationMetrics
	alerts            *AlertManager
	reports           *ReportGenerator
	mu                sync.RWMutex
}

// IntegrationMetrics holds real-time metrics
type IntegrationMetrics struct {
	TotalIntegrations    int                            `json:"total_integrations"`
	ActiveIntegrations   int                            `json:"active_integrations"`
	SyncOperations       map[string]*SyncMetrics        `json:"sync_operations"`
	ErrorRates          map[string]float64             `json:"error_rates"`
	ResponseTimes       map[string]time.Duration       `json:"response_times"`
	ThroughputRates     map[string]float64             `json:"throughput_rates"`
	HealthScores        map[string]float64             `json:"health_scores"`
	LastUpdated         time.Time                      `json:"last_updated"`
	DailyStats          map[string]*DailyIntegrationStats `json:"daily_stats"`
	WeeklyTrends        map[string]*WeeklyTrend        `json:"weekly_trends"`
	MonthlyReports      map[string]*MonthlyReport      `json:"monthly_reports"`
}

// SyncMetrics holds sync operation metrics
type SyncMetrics struct {
	TotalSyncs       int64         `json:"total_syncs"`
	SuccessfulSyncs  int64         `json:"successful_syncs"`
	FailedSyncs      int64         `json:"failed_syncs"`
	AverageTime      time.Duration `json:"average_time"`
	LastSyncTime     time.Time     `json:"last_sync_time"`
	ProductsSynced   int64         `json:"products_synced"`
	OrdersProcessed  int64         `json:"orders_processed"`
	InventoryUpdates int64         `json:"inventory_updates"`
}

// DailyIntegrationStats holds daily statistics
type DailyIntegrationStats struct {
	Date             time.Time `json:"date"`
	SyncOperations   int       `json:"sync_operations"`
	ErrorCount       int       `json:"error_count"`
	ProductsSynced   int       `json:"products_synced"`
	OrdersProcessed  int       `json:"orders_processed"`
	Revenue          float64   `json:"revenue"`
	AverageResponse  float64   `json:"average_response"`
	UptimePercentage float64   `json:"uptime_percentage"`
}

// WeeklyTrend holds weekly trend data
type WeeklyTrend struct {
	WeekStarting     time.Time `json:"week_starting"`
	SyncGrowth       float64   `json:"sync_growth"`
	ErrorReduction   float64   `json:"error_reduction"`
	PerformanceGain  float64   `json:"performance_gain"`
	RevenueIncrease  float64   `json:"revenue_increase"`
}

// MonthlyReport holds monthly report data
type MonthlyReport struct {
	Month            time.Time                    `json:"month"`
	TotalOperations  int64                        `json:"total_operations"`
	SuccessRate      float64                      `json:"success_rate"`
	AverageResponse  time.Duration                `json:"average_response"`
	TotalRevenue     float64                      `json:"total_revenue"`
	TopPerformers    []IntegrationPerformance     `json:"top_performers"`
	Issues           []IntegrationIssue           `json:"issues"`
	Recommendations  []string                     `json:"recommendations"`
}

// IntegrationPerformance holds performance data for an integration
type IntegrationPerformance struct {
	IntegrationID   string        `json:"integration_id"`
	IntegrationName string        `json:"integration_name"`
	SuccessRate     float64       `json:"success_rate"`
	AverageResponse time.Duration `json:"average_response"`
	TotalOperations int64         `json:"total_operations"`
	Revenue         float64       `json:"revenue"`
	Score           float64       `json:"score"`
}

// IntegrationIssue represents an integration issue
type IntegrationIssue struct {
	IntegrationID string    `json:"integration_id"`
	IssueType     string    `json:"issue_type"`
	Description   string    `json:"description"`
	Severity      string    `json:"severity"`
	FirstSeen     time.Time `json:"first_seen"`
	LastSeen      time.Time `json:"last_seen"`
	Occurrences   int       `json:"occurrences"`
	Status        string    `json:"status"`
}

// AlertManager manages integration alerts
type AlertManager struct {
	rules         []IntegrationAlertRule
	notifications chan Alert
	subscribers   []AlertSubscriber
}

// IntegrationAlertRule defines conditions for triggering alerts (renamed to avoid conflict)
type IntegrationAlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Condition   string                 `json:"condition"`
	Threshold   float64                `json:"threshold"`
	Duration    time.Duration          `json:"duration"`
	Severity    string                 `json:"severity"`
	Enabled     bool                   `json:"enabled"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Alert represents an alert
type Alert struct {
	ID            string                 `json:"id"`
	RuleID        string                 `json:"rule_id"`
	IntegrationID string                 `json:"integration_id"`
	Message       string                 `json:"message"`
	Severity      string                 `json:"severity"`
	Timestamp     time.Time              `json:"timestamp"`
	Metadata      map[string]interface{} `json:"metadata"`
	Acknowledged  bool                   `json:"acknowledged"`
	Resolved      bool                   `json:"resolved"`
}

// AlertSubscriber represents an alert subscriber
type AlertSubscriber interface {
	Notify(alert Alert) error
	GetChannels() []string
}

// ReportGenerator generates integration reports
type ReportGenerator struct {
	templates map[string]*ReportTemplate
}

// ReportTemplate defines report structure
type ReportTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Schedule    string                 `json:"schedule"`
	Recipients  []string               `json:"recipients"`
	Sections    []ReportSection        `json:"sections"`
	Filters     map[string]interface{} `json:"filters"`
	Format      string                 `json:"format"`
}

// ReportSection defines a section of a report
type ReportSection struct {
	Title       string                 `json:"title"`
	Type        string                 `json:"type"`
	Query       string                 `json:"query"`
	Visualization string               `json:"visualization"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// NewIntegrationAnalyticsService creates a new analytics service
func NewIntegrationAnalyticsService(
	db *sql.DB,
	marketplaceService *MarketplaceIntegrationsService,
	aiManager *AIIntegrationManager,
) *IntegrationAnalyticsService {
	service := &IntegrationAnalyticsService{
		db:                 db,
		marketplaceService: marketplaceService,
		aiManager:         aiManager,
		metrics:           NewIntegrationMetrics(),
		alerts:            NewAlertManager(),
		reports:           NewReportGenerator(),
	}
	
	service.createAnalyticsTables()
	service.startMetricsCollection()
	service.startAlertMonitoring()
	
	return service
}

// NewIntegrationMetrics creates new integration metrics
func NewIntegrationMetrics() *IntegrationMetrics {
	return &IntegrationMetrics{
		SyncOperations:  make(map[string]*SyncMetrics),
		ErrorRates:      make(map[string]float64),
		ResponseTimes:   make(map[string]time.Duration),
		ThroughputRates: make(map[string]float64),
		HealthScores:    make(map[string]float64),
		DailyStats:      make(map[string]*DailyIntegrationStats),
		WeeklyTrends:    make(map[string]*WeeklyTrend),
		MonthlyReports:  make(map[string]*MonthlyReport),
		LastUpdated:     time.Now(),
	}
}

// NewAlertManager creates new alert manager
func NewAlertManager() *AlertManager {
	return &AlertManager{
		rules:         make([]IntegrationAlertRule, 0),
		notifications: make(chan Alert, 1000),
		subscribers:   make([]AlertSubscriber, 0),
	}
}

// NewReportGenerator creates new report generator
func NewReportGenerator() *ReportGenerator {
	return &ReportGenerator{
		templates: make(map[string]*ReportTemplate),
	}
}

// createAnalyticsTables creates necessary tables for analytics
func (ias *IntegrationAnalyticsService) createAnalyticsTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS integration_metrics (
			id VARCHAR(128) PRIMARY KEY,
			integration_id VARCHAR(128) NOT NULL,
			metric_type VARCHAR(50) NOT NULL,
			metric_value DECIMAL(15,4) NOT NULL,
			timestamp DATETIME NOT NULL,
			metadata TEXT,
			INDEX idx_integration_id (integration_id),
			INDEX idx_metric_type (metric_type),
			INDEX idx_timestamp (timestamp)
		)`,
		`CREATE TABLE IF NOT EXISTS integration_sync_logs (
			id VARCHAR(128) PRIMARY KEY,
			integration_id VARCHAR(128) NOT NULL,
			operation_type VARCHAR(50) NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			status VARCHAR(20) NOT NULL,
			products_count INT DEFAULT 0,
			success_count INT DEFAULT 0,
			error_count INT DEFAULT 0,
			error_details TEXT,
			response_time_ms INT,
			INDEX idx_integration_id (integration_id),
			INDEX idx_operation_type (operation_type),
			INDEX idx_start_time (start_time),
			INDEX idx_status (status)
		)`,
		`CREATE TABLE IF NOT EXISTS integration_alerts (
			id VARCHAR(128) PRIMARY KEY,
			rule_id VARCHAR(128) NOT NULL,
			integration_id VARCHAR(128) NOT NULL,
			alert_type VARCHAR(50) NOT NULL,
			severity VARCHAR(20) NOT NULL,
			message TEXT NOT NULL,
			triggered_at DATETIME NOT NULL,
			acknowledged_at DATETIME,
			resolved_at DATETIME,
			metadata TEXT,
			INDEX idx_integration_id (integration_id),
			INDEX idx_triggered_at (triggered_at),
			INDEX idx_severity (severity)
		)`,
		`CREATE TABLE IF NOT EXISTS integration_reports (
			id VARCHAR(128) PRIMARY KEY,
			report_type VARCHAR(50) NOT NULL,
			period_start DATETIME NOT NULL,
			period_end DATETIME NOT NULL,
			data TEXT NOT NULL,
			generated_at DATETIME NOT NULL,
			INDEX idx_report_type (report_type),
			INDEX idx_period_start (period_start)
		)`,
	}

	for _, query := range queries {
		if _, err := ias.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create analytics table: %w", err)
		}
	}

	return nil
}

// startMetricsCollection starts collecting metrics in background
func (ias *IntegrationAnalyticsService) startMetricsCollection() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		
		for range ticker.C {
			ias.collectMetrics()
		}
	}()
}

// startAlertMonitoring starts monitoring for alerts
func (ias *IntegrationAnalyticsService) startAlertMonitoring() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		
		for range ticker.C {
			ias.checkAlertRules()
		}
	}()
}

// collectMetrics collects current metrics from all integrations
func (ias *IntegrationAnalyticsService) collectMetrics() {
	ias.mu.Lock()
	defer ias.mu.Unlock()
	
	// Get all integrations
	integrations := ias.marketplaceService.GetAllIntegrations()
	
	ias.metrics.TotalIntegrations = len(integrations)
	ias.metrics.ActiveIntegrations = 0
	
	for integrationID, integration := range integrations {
		if integration.IsActive {
			ias.metrics.ActiveIntegrations++
		}
		
		// Collect sync metrics
		syncMetrics := ias.collectSyncMetrics(integrationID)
		ias.metrics.SyncOperations[integrationID] = syncMetrics
		
		// Calculate error rates
		ias.metrics.ErrorRates[integrationID] = ias.calculateErrorRate(integrationID)
		
		// Measure response times
		ias.metrics.ResponseTimes[integrationID] = ias.measureResponseTime(integrationID)
		
		// Calculate throughput
		ias.metrics.ThroughputRates[integrationID] = ias.calculateThroughput(integrationID)
		
		// Calculate health scores
		ias.metrics.HealthScores[integrationID] = ias.calculateHealthScore(integrationID)
	}
	
	ias.metrics.LastUpdated = time.Now()
	
	// Store metrics to database
	ias.storeMetrics()
}

// collectSyncMetrics collects sync metrics for an integration
func (ias *IntegrationAnalyticsService) collectSyncMetrics(integrationID string) *SyncMetrics {
	// Query database for sync statistics
	query := `
		SELECT 
			COUNT(*) as total_syncs,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as successful_syncs,
			SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END) as failed_syncs,
			AVG(response_time_ms) as avg_response_time,
			MAX(start_time) as last_sync_time,
			SUM(products_count) as products_synced
		FROM integration_sync_logs 
		WHERE integration_id = ? AND start_time > DATE_SUB(NOW(), INTERVAL 24 HOUR)
	`
	
	var metrics SyncMetrics
	var avgResponseMs sql.NullFloat64
	var lastSyncTime sql.NullTime
	
	err := ias.db.QueryRow(query, integrationID).Scan(
		&metrics.TotalSyncs,
		&metrics.SuccessfulSyncs,
		&metrics.FailedSyncs,
		&avgResponseMs,
		&lastSyncTime,
		&metrics.ProductsSynced,
	)
	
	if err == nil {
		if avgResponseMs.Valid {
			metrics.AverageTime = time.Duration(avgResponseMs.Float64) * time.Millisecond
		}
		if lastSyncTime.Valid {
			metrics.LastSyncTime = lastSyncTime.Time
		}
	}
	
	return &metrics
}

// calculateErrorRate calculates error rate for an integration
func (ias *IntegrationAnalyticsService) calculateErrorRate(integrationID string) float64 {
	query := `
		SELECT 
			COUNT(*) as total_operations,
			SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END) as error_operations
		FROM integration_sync_logs 
		WHERE integration_id = ? AND start_time > DATE_SUB(NOW(), INTERVAL 1 HOUR)
	`
	
	var totalOps, errorOps int
	err := ias.db.QueryRow(query, integrationID).Scan(&totalOps, &errorOps)
	if err != nil || totalOps == 0 {
		return 0.0
	}
	
	return float64(errorOps) / float64(totalOps) * 100.0
}

// measureResponseTime measures average response time for an integration
func (ias *IntegrationAnalyticsService) measureResponseTime(integrationID string) time.Duration {
	query := `
		SELECT AVG(response_time_ms) 
		FROM integration_sync_logs 
		WHERE integration_id = ? AND start_time > DATE_SUB(NOW(), INTERVAL 1 HOUR)
	`
	
	var avgMs sql.NullFloat64
	err := ias.db.QueryRow(query, integrationID).Scan(&avgMs)
	if err != nil || !avgMs.Valid {
		return 0
	}
	
	return time.Duration(avgMs.Float64) * time.Millisecond
}

// calculateThroughput calculates throughput rate for an integration
func (ias *IntegrationAnalyticsService) calculateThroughput(integrationID string) float64 {
	query := `
		SELECT COUNT(*) 
		FROM integration_sync_logs 
		WHERE integration_id = ? AND start_time > DATE_SUB(NOW(), INTERVAL 1 HOUR)
	`
	
	var operations int
	err := ias.db.QueryRow(query, integrationID).Scan(&operations)
	if err != nil {
		return 0.0
	}
	
	return float64(operations) / 60.0 // operations per minute
}

// calculateHealthScore calculates overall health score for an integration
func (ias *IntegrationAnalyticsService) calculateHealthScore(integrationID string) float64 {
	errorRate := ias.metrics.ErrorRates[integrationID]
	responseTime := ias.metrics.ResponseTimes[integrationID]
	throughput := ias.metrics.ThroughputRates[integrationID]
	
	// Calculate health score based on multiple factors
	errorScore := math.Max(0, 100-errorRate*2)
	responseScore := math.Max(0, 100-float64(responseTime.Milliseconds())/10)
	throughputScore := math.Min(100, throughput*10)
	
	return (errorScore + responseScore + throughputScore) / 3.0
}

// storeMetrics stores current metrics to database
func (ias *IntegrationAnalyticsService) storeMetrics() {
	for integrationID, errorRate := range ias.metrics.ErrorRates {
		ias.storeMetric(integrationID, "error_rate", errorRate)
	}
	
	for integrationID, responseTime := range ias.metrics.ResponseTimes {
		ias.storeMetric(integrationID, "response_time", float64(responseTime.Milliseconds()))
	}
	
	for integrationID, throughput := range ias.metrics.ThroughputRates {
		ias.storeMetric(integrationID, "throughput", throughput)
	}
	
	for integrationID, healthScore := range ias.metrics.HealthScores {
		ias.storeMetric(integrationID, "health_score", healthScore)
	}
}

// storeMetric stores a single metric to database
func (ias *IntegrationAnalyticsService) storeMetric(integrationID, metricType string, value float64) {
	query := `
		INSERT INTO integration_metrics (id, integration_id, metric_type, metric_value, timestamp)
		VALUES (?, ?, ?, ?, NOW())
	`
	
	metricID := fmt.Sprintf("%s_%s_%d", integrationID, metricType, time.Now().Unix())
	ias.db.Exec(query, metricID, integrationID, metricType, value)
}

// checkAlertRules checks all alert rules and triggers alerts if necessary
func (ias *IntegrationAnalyticsService) checkAlertRules() {
	for _, rule := range ias.alerts.rules {
		if !rule.Enabled {
			continue
		}
		
		if ias.evaluateAlertRule(rule) {
			alert := Alert{
				ID:            ias.generateAlertID(),
				RuleID:        rule.ID,
				Message:       fmt.Sprintf("Alert rule '%s' triggered", rule.Name),
				Severity:      rule.Severity,
				Timestamp:     time.Now(),
				Metadata:      rule.Metadata,
			}
			
			ias.triggerAlert(alert)
		}
	}
}

// evaluateAlertRule evaluates whether an alert rule should trigger
func (ias *IntegrationAnalyticsService) evaluateAlertRule(rule IntegrationAlertRule) bool {
	// Implement rule evaluation logic based on rule.Condition
	// This is a simplified version
	return false
}

// triggerAlert triggers an alert
func (ias *IntegrationAnalyticsService) triggerAlert(alert Alert) {
	// Store alert in database
	ias.storeAlert(alert)
	
	// Send to notification channel
	select {
	case ias.alerts.notifications <- alert:
	default:
		// Channel full, log error
	}
	
	// Notify subscribers
	for _, subscriber := range ias.alerts.subscribers {
		go subscriber.Notify(alert)
	}
}

// storeAlert stores alert in database
func (ias *IntegrationAnalyticsService) storeAlert(alert Alert) {
	query := `
		INSERT INTO integration_alerts (id, rule_id, integration_id, alert_type, severity, message, triggered_at, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	metadataJSON, _ := json.Marshal(alert.Metadata)
	ias.db.Exec(query, alert.ID, alert.RuleID, alert.IntegrationID, "system", alert.Severity, alert.Message, alert.Timestamp, string(metadataJSON))
}

// generateAlertID generates a unique alert ID
func (ias *IntegrationAnalyticsService) generateAlertID() string {
	return fmt.Sprintf("alert_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond())
}

// GetMetrics returns current integration metrics
func (ias *IntegrationAnalyticsService) GetMetrics() *IntegrationMetrics {
	ias.mu.RLock()
	defer ias.mu.RUnlock()
	
	// Return a copy of metrics
	metricsCopy := *ias.metrics
	return &metricsCopy
}

// GetIntegrationHealth returns health status for a specific integration
func (ias *IntegrationAnalyticsService) GetIntegrationHealth(integrationID string) map[string]interface{} {
	ias.mu.RLock()
	defer ias.mu.RUnlock()
	
	return map[string]interface{}{
		"integration_id":  integrationID,
		"health_score":    ias.metrics.HealthScores[integrationID],
		"error_rate":      ias.metrics.ErrorRates[integrationID],
		"response_time":   ias.metrics.ResponseTimes[integrationID],
		"throughput":      ias.metrics.ThroughputRates[integrationID],
		"sync_metrics":    ias.metrics.SyncOperations[integrationID],
		"last_updated":    ias.metrics.LastUpdated,
	}
}

// GenerateReport generates a comprehensive integration report
func (ias *IntegrationAnalyticsService) GenerateReport(reportType string, startTime, endTime time.Time) (map[string]interface{}, error) {
	switch reportType {
	case "daily":
		return ias.generateDailyReport(startTime, endTime)
	case "weekly":
		return ias.generateWeeklyReport(startTime, endTime)
	case "monthly":
		return ias.generateMonthlyReport(startTime, endTime)
	default:
		return nil, fmt.Errorf("unsupported report type: %s", reportType)
	}
}

// generateDailyReport generates daily integration report
func (ias *IntegrationAnalyticsService) generateDailyReport(startTime, endTime time.Time) (map[string]interface{}, error) {
	// Implementation for daily report
	return map[string]interface{}{
		"report_type": "daily",
		"period":      fmt.Sprintf("%s to %s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02")),
		"metrics":     ias.GetMetrics(),
	}, nil
}

// generateWeeklyReport generates weekly integration report
func (ias *IntegrationAnalyticsService) generateWeeklyReport(startTime, endTime time.Time) (map[string]interface{}, error) {
	// Implementation for weekly report
	return map[string]interface{}{
		"report_type": "weekly",
		"period":      fmt.Sprintf("%s to %s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02")),
		"metrics":     ias.GetMetrics(),
	}, nil
}

// generateMonthlyReport generates monthly integration report
func (ias *IntegrationAnalyticsService) generateMonthlyReport(startTime, endTime time.Time) (map[string]interface{}, error) {
	// Implementation for monthly report
	return map[string]interface{}{
		"report_type": "monthly",
		"period":      fmt.Sprintf("%s to %s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02")),
		"metrics":     ias.GetMetrics(),
	}, nil
}