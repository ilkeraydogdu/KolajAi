package reporting

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ReportManager handles dynamic report generation
type ReportManager struct {
	db *sql.DB
}

// ReportConfig represents report configuration
type ReportConfig struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	DataSources []DataSource           `json:"data_sources"`
	Filters     []FilterConfig         `json:"filters"`
	Grouping    []GroupConfig          `json:"grouping"`
	Sorting     []SortConfig           `json:"sorting"`
	Columns     []ColumnConfig         `json:"columns"`
	Charts      []ChartConfig          `json:"charts"`
	Schedule    *ScheduleConfig        `json:"schedule,omitempty"`
	Permissions []string               `json:"permissions"`
	Parameters  map[string]interface{} `json:"parameters"`
	CreatedBy   int64                  `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// DataSource represents a data source for reports
type DataSource struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"` // "table", "query", "api"
	Source      string            `json:"source"`
	Joins       []JoinConfig      `json:"joins,omitempty"`
	Conditions  []ConditionConfig `json:"conditions,omitempty"`
	Aggregations []AggregateConfig `json:"aggregations,omitempty"`
}

// JoinConfig represents table join configuration
type JoinConfig struct {
	Type      string `json:"type"` // "INNER", "LEFT", "RIGHT"
	Table     string `json:"table"`
	Condition string `json:"condition"`
}

// ConditionConfig represents filter conditions
type ConditionConfig struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // "=", "!=", ">", "<", "LIKE", "IN"
	Value    interface{} `json:"value"`
	Logic    string      `json:"logic,omitempty"` // "AND", "OR"
}

// AggregateConfig represents aggregation configuration
type AggregateConfig struct {
	Function string `json:"function"` // "COUNT", "SUM", "AVG", "MIN", "MAX"
	Field    string `json:"field"`
	Alias    string `json:"alias"`
}

// FilterConfig represents dynamic filters
type FilterConfig struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Type        string      `json:"type"` // "text", "number", "date", "select", "multiselect"
	Field       string      `json:"field"`
	Operator    string      `json:"operator"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Options     []Option    `json:"options,omitempty"`
	Required    bool        `json:"required"`
}

// Option represents filter options
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// GroupConfig represents grouping configuration
type GroupConfig struct {
	Field string `json:"field"`
	Alias string `json:"alias,omitempty"`
}

// SortConfig represents sorting configuration
type SortConfig struct {
	Field string `json:"field"`
	Order string `json:"order"` // "ASC", "DESC"
}

// ColumnConfig represents column configuration
type ColumnConfig struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Field      string `json:"field"`
	Type       string `json:"type"` // "text", "number", "date", "currency", "percentage"
	Format     string `json:"format,omitempty"`
	Width      int    `json:"width,omitempty"`
	Sortable   bool   `json:"sortable"`
	Filterable bool   `json:"filterable"`
	Visible    bool   `json:"visible"`
}

// ChartConfig represents chart configuration
type ChartConfig struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"` // "bar", "line", "pie", "doughnut", "area"
	Title    string                 `json:"title"`
	XAxis    string                 `json:"x_axis"`
	YAxis    []string               `json:"y_axis"`
	Colors   []string               `json:"colors,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
	Position string                 `json:"position"` // "top", "bottom", "left", "right"
}

// ScheduleConfig represents report scheduling
type ScheduleConfig struct {
	Enabled   bool     `json:"enabled"`
	Frequency string   `json:"frequency"` // "daily", "weekly", "monthly"
	Time      string   `json:"time"`      // "HH:MM"
	Days      []string `json:"days,omitempty"` // For weekly: ["monday", "tuesday"]
	Recipients []string `json:"recipients"`
	Format    string   `json:"format"` // "pdf", "excel", "csv"
}

// ReportResult represents report execution result
type ReportResult struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Data        []map[string]interface{} `json:"data"`
	Charts      []ChartData              `json:"charts"`
	Summary     map[string]interface{}   `json:"summary"`
	Filters     map[string]interface{}   `json:"applied_filters"`
	TotalRows   int                      `json:"total_rows"`
	ExecutionTime time.Duration          `json:"execution_time"`
	GeneratedAt time.Time                `json:"generated_at"`
	GeneratedBy int64                    `json:"generated_by"`
}

// ChartData represents chart data
type ChartData struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"`
	Title  string                 `json:"title"`
	Labels []string               `json:"labels"`
	Data   []map[string]interface{} `json:"data"`
	Options map[string]interface{} `json:"options"`
}

// UserBehaviorReport represents detailed user behavior analysis
type UserBehaviorReport struct {
	UserID              int64                  `json:"user_id"`
	Name                string                 `json:"name"`
	Email               string                 `json:"email"`
	RegistrationDate    time.Time              `json:"registration_date"`
	LastActivity        time.Time              `json:"last_activity"`
	TotalOrders         int                    `json:"total_orders"`
	TotalSpent          float64                `json:"total_spent"`
	AverageOrderValue   float64                `json:"average_order_value"`
	PreferredCategories []CategoryPreference   `json:"preferred_categories"`
	ShoppingPatterns    ShoppingPattern        `json:"shopping_patterns"`
	DeviceUsage         map[string]int         `json:"device_usage"`
	LocationData        LocationData           `json:"location_data"`
	PaymentMethods      []PaymentMethodUsage   `json:"payment_methods"`
	Recommendations     []ProductRecommendation `json:"recommendations"`
	RiskScore           float64                `json:"risk_score"`
	LifetimeValue       float64                `json:"lifetime_value"`
	Segmentation        UserSegment            `json:"segmentation"`
}

// CategoryPreference represents user's category preferences
type CategoryPreference struct {
	CategoryID   int     `json:"category_id"`
	CategoryName string  `json:"category_name"`
	OrderCount   int     `json:"order_count"`
	TotalSpent   float64 `json:"total_spent"`
	Percentage   float64 `json:"percentage"`
}

// ShoppingPattern represents user's shopping patterns
type ShoppingPattern struct {
	PreferredDays    []string               `json:"preferred_days"`
	PreferredHours   []int                  `json:"preferred_hours"`
	AverageSessionTime time.Duration        `json:"average_session_time"`
	PagesPerSession  float64                `json:"pages_per_session"`
	ConversionRate   float64                `json:"conversion_rate"`
	CartAbandonment  float64                `json:"cart_abandonment"`
	ReturnRate       float64                `json:"return_rate"`
	SeasonalTrends   map[string]interface{} `json:"seasonal_trends"`
}

// LocationData represents user's location information
type LocationData struct {
	Country     string  `json:"country"`
	City        string  `json:"city"`
	Region      string  `json:"region"`
	Timezone    string  `json:"timezone"`
	Coordinates []float64 `json:"coordinates,omitempty"`
}

// PaymentMethodUsage represents payment method usage
type PaymentMethodUsage struct {
	Method     string  `json:"method"`
	Usage      int     `json:"usage"`
	Percentage float64 `json:"percentage"`
}

// ProductRecommendation represents product recommendations
type ProductRecommendation struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Score       float64 `json:"score"`
	Reason      string  `json:"reason"`
}

// UserSegment represents user segmentation
type UserSegment struct {
	Primary   string                 `json:"primary"`
	Secondary []string               `json:"secondary"`
	Scores    map[string]float64     `json:"scores"`
	Attributes map[string]interface{} `json:"attributes"`
}

// NewReportManager creates a new report manager
func NewReportManager(db *sql.DB) *ReportManager {
	rm := &ReportManager{db: db}
	rm.createReportTables()
	return rm
}

// createReportTables creates necessary tables for reporting
func (rm *ReportManager) createReportTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS report_configs (
			id VARCHAR(128) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			category VARCHAR(100),
			config_json TEXT NOT NULL,
			created_by BIGINT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_category (category),
			INDEX idx_created_by (created_by)
		)`,
		`CREATE TABLE IF NOT EXISTS report_executions (
			id VARCHAR(128) PRIMARY KEY,
			report_id VARCHAR(128),
			executed_by BIGINT,
			execution_time_ms INT,
			row_count INT,
			filters_json TEXT,
			executed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_report_id (report_id),
			INDEX idx_executed_by (executed_by),
			INDEX idx_executed_at (executed_at)
		)`,
		`CREATE TABLE IF NOT EXISTS user_behavior_cache (
			user_id BIGINT PRIMARY KEY,
			behavior_data TEXT NOT NULL,
			last_updated DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_last_updated (last_updated)
		)`,
	}

	for _, query := range queries {
		if _, err := rm.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

// CreateReport creates a new report configuration
func (rm *ReportManager) CreateReport(config *ReportConfig) error {
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	query := `
		INSERT INTO report_configs (id, name, description, category, config_json, created_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = rm.db.Exec(query, config.ID, config.Name, config.Description, 
		config.Category, string(configJSON), config.CreatedBy)
	
	return err
}

// ExecuteReport executes a report and returns results
func (rm *ReportManager) ExecuteReport(reportID string, filters map[string]interface{}, userID int64) (*ReportResult, error) {
	startTime := time.Now()

	// Get report configuration
	config, err := rm.GetReportConfig(reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report config: %w", err)
	}

	// Build and execute query
	query, args := rm.buildQuery(config, filters)
	
	rows, err := rm.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Process results
	data, err := rm.processRows(rows, config.Columns)
	if err != nil {
		return nil, fmt.Errorf("failed to process results: %w", err)
	}

	// Generate charts
	charts := rm.generateCharts(data, config.Charts)

	// Generate summary
	summary := rm.generateSummary(data, config)

	executionTime := time.Since(startTime)

	result := &ReportResult{
		ID:            reportID,
		Name:          config.Name,
		Data:          data,
		Charts:        charts,
		Summary:       summary,
		Filters:       filters,
		TotalRows:     len(data),
		ExecutionTime: executionTime,
		GeneratedAt:   time.Now(),
		GeneratedBy:   userID,
	}

	// Log execution
	rm.logExecution(reportID, userID, executionTime, len(data), filters)

	return result, nil
}

// GetUserBehaviorReport generates comprehensive user behavior report
func (rm *ReportManager) GetUserBehaviorReport(userID int64) (*UserBehaviorReport, error) {
	// Check cache first
	if cached, err := rm.getCachedUserBehavior(userID); err == nil {
		return cached, nil
	}

	report := &UserBehaviorReport{UserID: userID}

	// Get basic user info
	if err := rm.getUserBasicInfo(userID, report); err != nil {
		return nil, err
	}

	// Get order statistics
	if err := rm.getUserOrderStats(userID, report); err != nil {
		return nil, err
	}

	// Get category preferences
	if err := rm.getUserCategoryPreferences(userID, report); err != nil {
		return nil, err
	}

	// Get shopping patterns
	if err := rm.getUserShoppingPatterns(userID, report); err != nil {
		return nil, err
	}

	// Get device usage
	if err := rm.getUserDeviceUsage(userID, report); err != nil {
		return nil, err
	}

	// Get location data
	if err := rm.getUserLocationData(userID, report); err != nil {
		return nil, err
	}

	// Get payment method preferences
	if err := rm.getUserPaymentMethods(userID, report); err != nil {
		return nil, err
	}

	// Generate recommendations
	if err := rm.generateUserRecommendations(userID, report); err != nil {
		return nil, err
	}

	// Calculate risk score and lifetime value
	rm.calculateUserMetrics(report)

	// Determine user segmentation
	rm.determineUserSegmentation(report)

	// Cache the result
	rm.cacheUserBehavior(userID, report)

	return report, nil
}

// buildQuery builds SQL query from report configuration
func (rm *ReportManager) buildQuery(config *ReportConfig, filters map[string]interface{}) (string, []interface{}) {
	var query strings.Builder
	var args []interface{}

	// Build SELECT clause
	query.WriteString("SELECT ")
	
	// Add columns
	columnParts := make([]string, 0)
	for _, col := range config.Columns {
		if col.Visible {
			columnParts = append(columnParts, fmt.Sprintf("%s AS %s", col.Field, col.ID))
		}
	}
	query.WriteString(strings.Join(columnParts, ", "))

	// Build FROM clause
	if len(config.DataSources) > 0 {
		query.WriteString(" FROM ")
		query.WriteString(config.DataSources[0].Source)

		// Add joins
		for _, join := range config.DataSources[0].Joins {
			query.WriteString(fmt.Sprintf(" %s JOIN %s ON %s", 
				join.Type, join.Table, join.Condition))
		}
	}

	// Build WHERE clause
	whereClauses := make([]string, 0)
	
	// Add data source conditions
	for _, ds := range config.DataSources {
		for _, condition := range ds.Conditions {
			whereClauses = append(whereClauses, fmt.Sprintf("%s %s ?", 
				condition.Field, condition.Operator))
			args = append(args, condition.Value)
		}
	}

	// Add filter conditions
	for _, filter := range config.Filters {
		if value, exists := filters[filter.ID]; exists && value != nil {
			whereClauses = append(whereClauses, fmt.Sprintf("%s %s ?", 
				filter.Field, filter.Operator))
			args = append(args, value)
		}
	}

	if len(whereClauses) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(whereClauses, " AND "))
	}

	// Build GROUP BY clause
	if len(config.Grouping) > 0 {
		query.WriteString(" GROUP BY ")
		groupFields := make([]string, 0)
		for _, group := range config.Grouping {
			groupFields = append(groupFields, group.Field)
		}
		query.WriteString(strings.Join(groupFields, ", "))
	}

	// Build ORDER BY clause
	if len(config.Sorting) > 0 {
		query.WriteString(" ORDER BY ")
		sortParts := make([]string, 0)
		for _, sort := range config.Sorting {
			sortParts = append(sortParts, fmt.Sprintf("%s %s", sort.Field, sort.Order))
		}
		query.WriteString(strings.Join(sortParts, ", "))
	}

	return query.String(), args
}

// processRows processes SQL rows into result data
func (rm *ReportManager) processRows(rows *sql.Rows, columns []ColumnConfig) ([]map[string]interface{}, error) {
	data := make([]map[string]interface{}, 0)

	// Get column names
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Create value containers
	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Process each row
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]interface{})
		for i, col := range cols {
			val := values[i]
			if val != nil {
				// Format value based on column type
				if len(columns) > i {
					val = rm.formatValue(val, columns[i].Type, columns[i].Format)
				}
				row[col] = val
			}
		}

		data = append(data, row)
	}

	return data, rows.Err()
}

// formatValue formats value based on column type
func (rm *ReportManager) formatValue(value interface{}, colType, format string) interface{} {
	switch colType {
	case "currency":
		if val, ok := value.(float64); ok {
			return fmt.Sprintf("%.2f TL", val)
		}
	case "percentage":
		if val, ok := value.(float64); ok {
			return fmt.Sprintf("%.1f%%", val*100)
		}
	case "date":
		if val, ok := value.(time.Time); ok {
			if format != "" {
				return val.Format(format)
			}
			return val.Format("2006-01-02")
		}
	}
	return value
}

// generateCharts generates chart data from result data
func (rm *ReportManager) generateCharts(data []map[string]interface{}, chartConfigs []ChartConfig) []ChartData {
	charts := make([]ChartData, 0)

	for _, config := range chartConfigs {
		chartData := ChartData{
			ID:      config.ID,
			Type:    config.Type,
			Title:   config.Title,
			Options: config.Options,
		}

		// Extract labels and data based on chart type
		switch config.Type {
		case "pie", "doughnut":
			chartData.Labels, chartData.Data = rm.generatePieChartData(data, config)
		case "bar", "line", "area":
			chartData.Labels, chartData.Data = rm.generateBarChartData(data, config)
		}

		charts = append(charts, chartData)
	}

	return charts
}

// generatePieChartData generates data for pie/doughnut charts
func (rm *ReportManager) generatePieChartData(data []map[string]interface{}, config ChartConfig) ([]string, []map[string]interface{}) {
	labels := make([]string, 0)
	values := make([]interface{}, 0)

	for _, row := range data {
		if label, exists := row[config.XAxis]; exists {
			labels = append(labels, fmt.Sprintf("%v", label))
		}
		if len(config.YAxis) > 0 {
			if value, exists := row[config.YAxis[0]]; exists {
				values = append(values, value)
			}
		}
	}

	chartData := []map[string]interface{}{
		{
			"data":            values,
			"backgroundColor": config.Colors,
		},
	}

	return labels, chartData
}

// generateBarChartData generates data for bar/line/area charts
func (rm *ReportManager) generateBarChartData(data []map[string]interface{}, config ChartConfig) ([]string, []map[string]interface{}) {
	labels := make([]string, 0)
	datasets := make([]map[string]interface{}, len(config.YAxis))

	// Initialize datasets
	for i, yAxis := range config.YAxis {
		datasets[i] = map[string]interface{}{
			"label": yAxis,
			"data":  make([]interface{}, 0),
		}
		if i < len(config.Colors) {
			datasets[i]["backgroundColor"] = config.Colors[i]
			datasets[i]["borderColor"] = config.Colors[i]
		}
	}

	// Populate data
	for _, row := range data {
		if label, exists := row[config.XAxis]; exists {
			labels = append(labels, fmt.Sprintf("%v", label))
		}

		for i, yAxis := range config.YAxis {
			if value, exists := row[yAxis]; exists {
				data := datasets[i]["data"].([]interface{})
				datasets[i]["data"] = append(data, value)
			}
		}
	}

	return labels, datasets
}

// generateSummary generates summary statistics
func (rm *ReportManager) generateSummary(data []map[string]interface{}, config *ReportConfig) map[string]interface{} {
	summary := make(map[string]interface{})
	
	summary["total_rows"] = len(data)
	summary["generated_at"] = time.Now().Format("2006-01-02 15:04:05")
	
	// Calculate numeric summaries
	for _, col := range config.Columns {
		if col.Type == "number" || col.Type == "currency" {
			values := make([]float64, 0)
			for _, row := range data {
				if val, exists := row[col.ID]; exists {
					if numVal, ok := val.(float64); ok {
						values = append(values, numVal)
					}
				}
			}
			
			if len(values) > 0 {
				summary[col.ID+"_sum"] = rm.sum(values)
				summary[col.ID+"_avg"] = rm.average(values)
				summary[col.ID+"_min"] = rm.min(values)
				summary[col.ID+"_max"] = rm.max(values)
			}
		}
	}
	
	return summary
}

// Helper functions for calculations
func (rm *ReportManager) sum(values []float64) float64 {
	total := 0.0
	for _, v := range values {
		total += v
	}
	return total
}

func (rm *ReportManager) average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return rm.sum(values) / float64(len(values))
}

func (rm *ReportManager) min(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

func (rm *ReportManager) max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

// GetReportConfig retrieves report configuration
func (rm *ReportManager) GetReportConfig(reportID string) (*ReportConfig, error) {
	query := "SELECT config_json FROM report_configs WHERE id = ?"
	
	var configJSON string
	err := rm.db.QueryRow(query, reportID).Scan(&configJSON)
	if err != nil {
		return nil, err
	}

	var config ReportConfig
	err = json.Unmarshal([]byte(configJSON), &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// logExecution logs report execution
func (rm *ReportManager) logExecution(reportID string, userID int64, executionTime time.Duration, rowCount int, filters map[string]interface{}) {
	filtersJSON, _ := json.Marshal(filters)
	
	query := `
		INSERT INTO report_executions (id, report_id, executed_by, execution_time_ms, row_count, filters_json)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	
	executionID := fmt.Sprintf("%s_%d_%d", reportID, userID, time.Now().Unix())
	
	rm.db.Exec(query, executionID, reportID, userID, 
		executionTime.Milliseconds(), rowCount, string(filtersJSON))
}

// Additional methods for user behavior analysis would continue here...
// (getUserBasicInfo, getUserOrderStats, etc.)

// getUserBasicInfo gets basic user information
func (rm *ReportManager) getUserBasicInfo(userID int64, report *UserBehaviorReport) error {
	query := `
		SELECT name, email, created_at,
		       (SELECT MAX(last_activity) FROM sessions WHERE user_id = ? AND is_active = TRUE) as last_activity
		FROM users WHERE id = ?
	`
	
	var lastActivity sql.NullTime
	err := rm.db.QueryRow(query, userID, userID).Scan(
		&report.Name, &report.Email, &report.RegistrationDate, &lastActivity)
	
	if err != nil {
		return err
	}
	
	if lastActivity.Valid {
		report.LastActivity = lastActivity.Time
	}
	
	return nil
}

// getUserOrderStats gets user order statistics
func (rm *ReportManager) getUserOrderStats(userID int64, report *UserBehaviorReport) error {
	query := `
		SELECT COUNT(*) as total_orders, 
		       COALESCE(SUM(total_amount), 0) as total_spent,
		       COALESCE(AVG(total_amount), 0) as average_order_value
		FROM orders 
		WHERE user_id = ? AND status != 'cancelled'
	`
	
	return rm.db.QueryRow(query, userID).Scan(
		&report.TotalOrders, &report.TotalSpent, &report.AverageOrderValue)
}

// getCachedUserBehavior retrieves cached user behavior data
func (rm *ReportManager) getCachedUserBehavior(userID int64) (*UserBehaviorReport, error) {
	query := "SELECT behavior_data FROM user_behavior_cache WHERE user_id = ? AND last_updated > DATE_SUB(NOW(), INTERVAL 1 HOUR)"
	
	var behaviorJSON string
	err := rm.db.QueryRow(query, userID).Scan(&behaviorJSON)
	if err != nil {
		return nil, err
	}
	
	var report UserBehaviorReport
	err = json.Unmarshal([]byte(behaviorJSON), &report)
	return &report, err
}

// cacheUserBehavior caches user behavior data
func (rm *ReportManager) cacheUserBehavior(userID int64, report *UserBehaviorReport) {
	behaviorJSON, err := json.Marshal(report)
	if err != nil {
		return
	}
	
	query := `
		INSERT INTO user_behavior_cache (user_id, behavior_data) 
		VALUES (?, ?) 
		ON DUPLICATE KEY UPDATE behavior_data = VALUES(behavior_data)
	`
	
	rm.db.Exec(query, userID, string(behaviorJSON))
}

// Placeholder methods for additional user behavior analysis
func (rm *ReportManager) getUserCategoryPreferences(userID int64, report *UserBehaviorReport) error {
	// Implementation would analyze user's category preferences
	report.PreferredCategories = []CategoryPreference{}
	return nil
}

func (rm *ReportManager) getUserShoppingPatterns(userID int64, report *UserBehaviorReport) error {
	// Implementation would analyze shopping patterns
	report.ShoppingPatterns = ShoppingPattern{}
	return nil
}

func (rm *ReportManager) getUserDeviceUsage(userID int64, report *UserBehaviorReport) error {
	// Implementation would analyze device usage from sessions
	report.DeviceUsage = make(map[string]int)
	return nil
}

func (rm *ReportManager) getUserLocationData(userID int64, report *UserBehaviorReport) error {
	// Implementation would analyze location data
	report.LocationData = LocationData{}
	return nil
}

func (rm *ReportManager) getUserPaymentMethods(userID int64, report *UserBehaviorReport) error {
	// Implementation would analyze payment method preferences
	report.PaymentMethods = []PaymentMethodUsage{}
	return nil
}

func (rm *ReportManager) generateUserRecommendations(userID int64, report *UserBehaviorReport) error {
	// Implementation would generate product recommendations
	report.Recommendations = []ProductRecommendation{}
	return nil
}

func (rm *ReportManager) calculateUserMetrics(report *UserBehaviorReport) {
	// Calculate risk score and lifetime value
	report.RiskScore = 0.1 // Low risk by default
	report.LifetimeValue = report.TotalSpent * 1.2 // Simple LTV calculation
}

func (rm *ReportManager) determineUserSegmentation(report *UserBehaviorReport) {
	// Determine user segmentation based on behavior
	report.Segmentation = UserSegment{
		Primary:   "Regular Customer",
		Secondary: []string{"Online Shopper"},
		Scores:    make(map[string]float64),
		Attributes: make(map[string]interface{}),
	}
}