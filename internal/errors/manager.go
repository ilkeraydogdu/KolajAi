package errors

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// ErrorManager handles comprehensive error management
type ErrorManager struct {
	db         *sql.DB
	notifier   ErrorNotifier
	config     ErrorConfig
	handlers   map[ErrorType]ErrorHandlerInterface
	middleware []ErrorMiddleware
}

// ErrorType represents different types of errors
type ErrorType string

const (
	VALIDATION         ErrorType = "validation"
	DATABASE           ErrorType = "database"
	AUTHENTICATION     ErrorType = "authentication"
	AUTHORIZATION      ErrorType = "authorization"
	NETWORK            ErrorType = "network"
	SYSTEM             ErrorType = "system"
	APPLICATION        ErrorType = "application"
	SECURITY_VIOLATION ErrorType = "security_violation"
	PERFORMANCE        ErrorType = "performance"
	INTEGRATION        ErrorType = "integration"
	INTERNAL           ErrorType = "internal"
	FORBIDDEN          ErrorType = "forbidden"
	RATE_LIMITED       ErrorType = "rate_limited"
	WARNING            ErrorType = "warning"
	ERROR              ErrorType = "error"

	// Legacy constants for compatibility
	ErrorTypeValidation    ErrorType = "validation"
	ErrorTypeDatabase      ErrorType = "database"
	ErrorTypeAuth          ErrorType = "authentication"
	ErrorTypeAuthorization           = "authorization"
	ErrorTypeNetwork       ErrorType = "network"
	ErrorTypeSystem        ErrorType = "system"
	ErrorTypeApplication   ErrorType = "application"
	ErrorTypeSecurity      ErrorType = "security"
	ErrorTypePerformance   ErrorType = "performance"
	ErrorTypeIntegration   ErrorType = "integration"
)

// ErrorSeverity is already defined in integration_errors.go

// ApplicationError represents a comprehensive application error
type ApplicationError struct {
	ID            string                 `json:"id"`
	Type          ErrorType              `json:"type"`
	Severity      ErrorSeverity          `json:"severity"`
	Code          string                 `json:"code"`
	Message       string                 `json:"message"`
	UserMessage   string                 `json:"user_message"`
	Details       map[string]interface{} `json:"details"`
	Context       ErrorContext           `json:"context"`
	StackTrace    []StackFrame           `json:"stack_trace"`
	Cause         *ApplicationError      `json:"cause,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	Resolved      bool                   `json:"resolved"`
	ResolvedAt    *time.Time             `json:"resolved_at,omitempty"`
	ResolvedBy    string                 `json:"resolved_by,omitempty"`
	Occurrences   int                    `json:"occurrences"`
	FirstSeen     time.Time              `json:"first_seen"`
	LastSeen      time.Time              `json:"last_seen"`
	AffectedUsers []string               `json:"affected_users"`
	Tags          []string               `json:"tags"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ErrorContext provides context about where the error occurred
type ErrorContext struct {
	RequestID   string            `json:"request_id"`
	UserID      string            `json:"user_id"`
	SessionID   string            `json:"session_id"`
	IPAddress   string            `json:"ip_address"`
	UserAgent   string            `json:"user_agent"`
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	Parameters  map[string]string `json:"parameters"`
	Body        string            `json:"body,omitempty"`
	Function    string            `json:"function"`
	File        string            `json:"file"`
	Line        int               `json:"line"`
	StackTrace  string            `json:"stack_trace"`
	Version     string            `json:"version"`
	Environment string            `json:"environment"`
	Server      string            `json:"server"`
	Database    string            `json:"database,omitempty"`
	ExternalAPI string            `json:"external_api,omitempty"`
}

// StackFrame represents a single frame in the stack trace
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Package  string `json:"package"`
}

// ErrorConfig holds error management configuration
type ErrorConfig struct {
	Environment         string             `json:"environment"`
	EnableStackTrace    bool               `json:"enable_stack_trace"`
	EnableNotifications bool               `json:"enable_notifications"`
	MaxStackDepth       int                `json:"max_stack_depth"`
	SamplingRate        float64            `json:"sampling_rate"`
	RetentionDays       int                `json:"retention_days"`
	NotificationRules   []NotificationRule `json:"notification_rules"`
	IgnorePatterns      []string           `json:"ignore_patterns"`
	GroupingRules       []GroupingRule     `json:"grouping_rules"`
}

// NotificationRule defines when to send notifications
type NotificationRule struct {
	Severity   ErrorSeverity `json:"severity"`
	Type       ErrorType     `json:"type"`
	Threshold  int           `json:"threshold"`
	TimeWindow time.Duration `json:"time_window"`
	Recipients []string      `json:"recipients"`
	Channels   []string      `json:"channels"` // email, slack, webhook
	Template   string        `json:"template"`
	Enabled    bool          `json:"enabled"`
}

// GroupingRule defines how to group similar errors
type GroupingRule struct {
	Fields     []string      `json:"fields"`
	TimeWindow time.Duration `json:"time_window"`
	MaxGroup   int           `json:"max_group"`
}

// ErrorHandlerInterface handles specific error types
type ErrorHandlerInterface interface {
	Handle(ctx context.Context, err *ApplicationError) error
	CanHandle(errorType ErrorType) bool
	Priority() int
}

// ErrorMiddleware processes errors before handling
type ErrorMiddleware interface {
	Process(ctx context.Context, err *ApplicationError) (*ApplicationError, error)
}

// ErrorNotifier sends error notifications
type ErrorNotifier interface {
	Notify(ctx context.Context, err *ApplicationError, rule NotificationRule) error
}

// ErrorManagerStats represents error statistics
type ErrorManagerStats struct {
	TotalErrors      int                       `json:"total_errors"`
	ErrorsByType     map[ErrorType]int         `json:"errors_by_type"`
	ErrorsBySeverity map[ErrorSeverity]int     `json:"errors_by_severity"`
	TopErrors        []ErrorSummary            `json:"top_errors"`
	TrendData        []ErrorTrendPoint         `json:"trend_data"`
	ResolutionTime   map[ErrorSeverity]float64 `json:"avg_resolution_time"`
	AffectedUsers    int                       `json:"affected_users"`
	ErrorRate        float64                   `json:"error_rate"`
	Uptime           float64                   `json:"uptime"`
}

// ErrorSummary represents a summary of similar errors
type ErrorSummary struct {
	ID            string        `json:"id"`
	Message       string        `json:"message"`
	Type          ErrorType     `json:"type"`
	Severity      ErrorSeverity `json:"severity"`
	Count         int           `json:"count"`
	FirstSeen     time.Time     `json:"first_seen"`
	LastSeen      time.Time     `json:"last_seen"`
	Resolved      bool          `json:"resolved"`
	AffectedUsers int           `json:"affected_users"`
}

// ErrorTrendPoint represents a point in error trend data
type ErrorTrendPoint struct {
	Timestamp time.Time     `json:"timestamp"`
	Count     int           `json:"count"`
	Type      ErrorType     `json:"type"`
	Severity  ErrorSeverity `json:"severity"`
}

// NewErrorManager creates a new error manager
func NewErrorManager(db *sql.DB, notifier ErrorNotifier, config ErrorConfig) *ErrorManager {
	em := &ErrorManager{
		db:         db,
		notifier:   notifier,
		config:     config,
		handlers:   make(map[ErrorType]ErrorHandlerInterface),
		middleware: make([]ErrorMiddleware, 0),
	}

	em.createErrorTables()
	em.registerDefaultHandlers()

	return em
}

// NewApplicationError creates a new application error
func NewApplicationError(errorType ErrorType, code, message string, cause error) *ApplicationError {
	return &ApplicationError{
		Type:        errorType,
		Code:        code,
		Message:     message,
		UserMessage: message,
		Cause:       nil, // would set properly in real implementation
		Timestamp:   time.Now(),
		Details:     make(map[string]interface{}),
	}
}

// HandleHTTPError handles HTTP errors and sends appropriate response
func (em *ErrorManager) HandleHTTPError(w http.ResponseWriter, r *http.Request, err *ApplicationError) {
	// Log the error
	em.LogError(err)

	// Set appropriate HTTP status code
	statusCode := em.getHTTPStatusCode(err.Type)
	w.WriteHeader(statusCode)

	// Send JSON error response
	w.Header().Set("Content-Type", "application/json")
	// In real implementation, would use proper JSON encoding
	fmt.Fprintf(w, `{"error":{"type":"%s","code":"%s","message":"%s"}}`, err.Type, err.Code, err.UserMessage)
}

// LogError logs an error
func (em *ErrorManager) LogError(err *ApplicationError) {
	// In a real implementation, this would store to database
	fmt.Printf("Error: %+v\n", err)
}

// createErrorTables creates necessary tables for error management
func (em *ErrorManager) createErrorTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS errors (
			id VARCHAR(128) PRIMARY KEY,
			type VARCHAR(50) NOT NULL,
			severity VARCHAR(20) NOT NULL,
			code VARCHAR(100),
			message TEXT NOT NULL,
			user_message TEXT,
			details TEXT,
			context_json TEXT,
			stack_trace TEXT,
			cause_id VARCHAR(128),
			timestamp DATETIME NOT NULL,
			resolved BOOLEAN DEFAULT FALSE,
			resolved_at DATETIME,
			resolved_by VARCHAR(255),
			occurrences INT DEFAULT 1,
			first_seen DATETIME NOT NULL,
			last_seen DATETIME NOT NULL,
			affected_users TEXT,
			tags TEXT,
			metadata TEXT,
			INDEX idx_type (type),
			INDEX idx_severity (severity),
			INDEX idx_timestamp (timestamp),
			INDEX idx_resolved (resolved),
			INDEX idx_first_seen (first_seen)
		)`,
		`CREATE TABLE IF NOT EXISTS error_occurrences (
			id VARCHAR(128) PRIMARY KEY,
			error_id VARCHAR(128) NOT NULL,
			request_id VARCHAR(128),
			user_id VARCHAR(100),
			session_id VARCHAR(128),
			ip_address VARCHAR(45),
			user_agent TEXT,
			url TEXT,
			method VARCHAR(10),
			headers TEXT,
			parameters TEXT,
			body TEXT,
			timestamp DATETIME NOT NULL,
			INDEX idx_error_id (error_id),
			INDEX idx_timestamp (timestamp),
			INDEX idx_user_id (user_id),
			FOREIGN KEY (error_id) REFERENCES errors(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS error_notifications (
			id VARCHAR(128) PRIMARY KEY,
			error_id VARCHAR(128) NOT NULL,
			rule_name VARCHAR(100) NOT NULL,
			channel VARCHAR(50) NOT NULL,
			recipient VARCHAR(255) NOT NULL,
			sent_at DATETIME NOT NULL,
			status VARCHAR(20) NOT NULL,
			response TEXT,
			INDEX idx_error_id (error_id),
			INDEX idx_sent_at (sent_at),
			FOREIGN KEY (error_id) REFERENCES errors(id) ON DELETE CASCADE
		)`,
	}

	for _, query := range queries {
		if _, err := em.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create error table: %w", err)
		}
	}

	return nil
}

// RegisterHandler registers an error handler for specific error types
func (em *ErrorManager) RegisterHandler(handler ErrorHandlerInterface) {
	// Implementation would register handlers based on their supported types
}

// RegisterMiddleware registers error processing middleware
func (em *ErrorManager) RegisterMiddleware(middleware ErrorMiddleware) {
	em.middleware = append(em.middleware, middleware)
}

// HandleError processes and handles an error
func (em *ErrorManager) HandleError(ctx context.Context, err error, errorType ErrorType, severity ErrorSeverity) *ApplicationError {
	appError := em.createApplicationError(err, errorType, severity, ctx)

	// Process through middleware
	for _, middleware := range em.middleware {
		processedError, middlewareErr := middleware.Process(ctx, appError)
		if middlewareErr != nil {
			// Log middleware error but continue
			continue
		}
		appError = processedError
	}

	// Check if error should be ignored
	if em.shouldIgnoreError(appError) {
		return appError
	}

	// Store error in database
	em.storeError(appError)

	// Handle error with registered handlers
	if handler, exists := em.handlers[errorType]; exists {
		handler.Handle(ctx, appError)
	}

	// Check notification rules
	em.checkNotificationRules(ctx, appError)

	return appError
}

// createApplicationError creates a comprehensive application error
func (em *ErrorManager) createApplicationError(err error, errorType ErrorType, severity ErrorSeverity, ctx context.Context) *ApplicationError {
	now := time.Now()

	appError := &ApplicationError{
		ID:          em.generateErrorID(),
		Type:        errorType,
		Severity:    severity,
		Message:     err.Error(),
		UserMessage: em.generateUserMessage(err, errorType),
		Details:     make(map[string]interface{}),
		Context:     em.extractContext(ctx),
		Timestamp:   now,
		FirstSeen:   now,
		LastSeen:    now,
		Occurrences: 1,
		Tags:        make([]string, 0),
		Metadata:    make(map[string]interface{}),
	}

	// Generate stack trace if enabled
	if em.config.EnableStackTrace {
		appError.StackTrace = em.captureStackTrace()
	}

	// Set error code based on type and error
	appError.Code = em.generateErrorCode(errorType, err)

	// Extract additional details based on error type
	appError.Details = em.extractErrorDetails(err, errorType)

	return appError
}

// extractContext extracts error context from the request context
func (em *ErrorManager) extractContext(ctx context.Context) ErrorContext {
	errorCtx := ErrorContext{
		Environment: "production", // This would come from config
		Version:     "1.0.0",      // This would come from build info
	}

	// Extract request ID if available
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			errorCtx.RequestID = id
		}
	}

	// Extract user ID if available
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			errorCtx.UserID = id
		}
	}

	// Extract HTTP request info if available
	if req := ctx.Value("http_request"); req != nil {
		if r, ok := req.(*http.Request); ok {
			errorCtx.URL = r.URL.String()
			errorCtx.Method = r.Method
			errorCtx.UserAgent = r.UserAgent()
			errorCtx.IPAddress = em.extractIPAddress(r)

			// Extract headers (filtering sensitive ones)
			errorCtx.Headers = em.extractSafeHeaders(r.Header)

			// Extract parameters
			errorCtx.Parameters = em.extractParameters(r)
		}
	}

	// Add function context
	if pc, file, line, ok := runtime.Caller(3); ok {
		errorCtx.Function = runtime.FuncForPC(pc).Name()
		errorCtx.File = file
		errorCtx.Line = line
	}

	return errorCtx
}

// captureStackTrace captures the current stack trace
func (em *ErrorManager) captureStackTrace() []StackFrame {
	frames := make([]StackFrame, 0)
	maxDepth := em.config.MaxStackDepth
	if maxDepth <= 0 {
		maxDepth = 32
	}

	for i := 2; i < maxDepth; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		frame := StackFrame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
			Package:  em.extractPackageName(fn.Name()),
		}

		frames = append(frames, frame)
	}

	return frames
}

// extractPackageName extracts package name from function name
func (em *ErrorManager) extractPackageName(funcName string) string {
	parts := strings.Split(funcName, ".")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], ".")
	}
	return ""
}

// generateErrorID generates a unique error ID
func (em *ErrorManager) generateErrorID() string {
	return fmt.Sprintf("err_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond())
}

// generateErrorCode generates an error code based on type and error
func (em *ErrorManager) generateErrorCode(errorType ErrorType, err error) string {
	// This would implement logic to generate meaningful error codes
	return fmt.Sprintf("%s_001", strings.ToUpper(string(errorType)))
}

// generateUserMessage generates a user-friendly error message
func (em *ErrorManager) generateUserMessage(err error, errorType ErrorType) string {
	switch errorType {
	case ErrorTypeValidation:
		return "Lütfen girdiğiniz bilgileri kontrol edin."
	case ErrorTypeAuth:
		return "Giriş yapmanız gerekiyor."
	case ErrorTypeAuthorization:
		return "Bu işlem için yetkiniz bulunmuyor."
	case ErrorTypeNetwork:
		return "Bağlantı sorunu yaşanıyor. Lütfen daha sonra tekrar deneyin."
	case ErrorTypeDatabase:
		return "Sistem geçici olarak kullanılamıyor. Lütfen daha sonra tekrar deneyin."
	default:
		return "Bir hata oluştu. Lütfen daha sonra tekrar deneyin."
	}
}

// extractErrorDetails extracts additional details based on error type
func (em *ErrorManager) extractErrorDetails(err error, errorType ErrorType) map[string]interface{} {
	details := make(map[string]interface{})

	// Add error type specific details
	switch errorType {
	case ErrorTypeDatabase:
		// Extract SQL-specific details if available
		details["database_error"] = err.Error()
	case ErrorTypeValidation:
		// Extract validation-specific details
		details["validation_type"] = "field_validation"
	}

	return details
}

// extractIPAddress extracts the real IP address from request
func (em *ErrorManager) extractIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	return r.RemoteAddr
}

// extractSafeHeaders extracts safe headers (filtering sensitive ones)
func (em *ErrorManager) extractSafeHeaders(headers http.Header) map[string]string {
	safeHeaders := make(map[string]string)
	sensitiveHeaders := map[string]bool{
		"authorization": true,
		"cookie":        true,
		"x-api-key":     true,
		"x-auth-token":  true,
	}

	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		if !sensitiveHeaders[lowerKey] && len(values) > 0 {
			safeHeaders[key] = values[0]
		}
	}

	return safeHeaders
}

// extractParameters extracts request parameters
func (em *ErrorManager) extractParameters(r *http.Request) map[string]string {
	params := make(map[string]string)

	// Extract query parameters
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	// Extract form parameters (if applicable)
	if r.Method == "POST" || r.Method == "PUT" {
		r.ParseForm()
		for key, values := range r.PostForm {
			if len(values) > 0 && !em.isSensitiveParam(key) {
				params[key] = values[0]
			}
		}
	}

	return params
}

// isSensitiveParam checks if a parameter contains sensitive data
func (em *ErrorManager) isSensitiveParam(param string) bool {
	sensitiveParams := []string{"password", "token", "secret", "key", "auth"}
	lowerParam := strings.ToLower(param)

	for _, sensitive := range sensitiveParams {
		if strings.Contains(lowerParam, sensitive) {
			return true
		}
	}

	return false
}

// shouldIgnoreError checks if an error should be ignored based on patterns
func (em *ErrorManager) shouldIgnoreError(err *ApplicationError) bool {
	for _, pattern := range em.config.IgnorePatterns {
		if strings.Contains(err.Message, pattern) {
			return true
		}
	}
	return false
}

// storeError stores the error in the database
func (em *ErrorManager) storeError(err *ApplicationError) error {
	// Check if similar error exists and group them
	existingError := em.findSimilarError(err)
	if existingError != nil {
		return em.updateExistingError(existingError, err)
	}

	// Store new error
	return em.insertNewError(err)
}

// findSimilarError finds similar errors for grouping
func (em *ErrorManager) findSimilarError(err *ApplicationError) *ApplicationError {
	// Implementation would use grouping rules to find similar errors
	// This is a simplified version
	query := `
		SELECT id, occurrences FROM errors 
		WHERE type = ? AND message = ? AND resolved = FALSE 
		AND first_seen > DATE_SUB(NOW(), INTERVAL 1 HOUR)
		LIMIT 1
	`

	var existingID string
	var occurrences int
	dbErr := em.db.QueryRow(query, err.Type, err.Message).Scan(&existingID, &occurrences)
	if dbErr != nil {
		return nil
	}

	return &ApplicationError{ID: existingID, Occurrences: occurrences}
}

// updateExistingError updates an existing error with new occurrence
func (em *ErrorManager) updateExistingError(existing, new *ApplicationError) error {
	query := `
		UPDATE errors 
		SET occurrences = occurrences + 1, last_seen = NOW()
		WHERE id = ?
	`

	_, err := em.db.Exec(query, existing.ID)
	return err
}

// insertNewError inserts a new error into the database
func (em *ErrorManager) insertNewError(err *ApplicationError) error {
	detailsJSON, _ := json.Marshal(err.Details)
	contextJSON, _ := json.Marshal(err.Context)
	stackTraceJSON, _ := json.Marshal(err.StackTrace)
	affectedUsersJSON, _ := json.Marshal(err.AffectedUsers)
	tagsJSON, _ := json.Marshal(err.Tags)
	metadataJSON, _ := json.Marshal(err.Metadata)

	query := `
		INSERT INTO errors (
			id, type, severity, code, message, user_message, details,
			context_json, stack_trace, timestamp, first_seen, last_seen,
			occurrences, affected_users, tags, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, dbErr := em.db.Exec(query,
		err.ID, err.Type, err.Severity, err.Code, err.Message, err.UserMessage,
		string(detailsJSON), string(contextJSON), string(stackTraceJSON),
		err.Timestamp, err.FirstSeen, err.LastSeen, err.Occurrences,
		string(affectedUsersJSON), string(tagsJSON), string(metadataJSON),
	)

	return dbErr
}

// checkNotificationRules checks if error matches notification rules
func (em *ErrorManager) checkNotificationRules(ctx context.Context, err *ApplicationError) {
	if !em.config.EnableNotifications {
		return
	}

	for _, rule := range em.config.NotificationRules {
		if em.matchesNotificationRule(err, rule) {
			em.sendNotification(ctx, err, rule)
		}
	}
}

// matchesNotificationRule checks if error matches a notification rule
func (em *ErrorManager) matchesNotificationRule(err *ApplicationError, rule NotificationRule) bool {
	if !rule.Enabled {
		return false
	}

	if rule.Severity != "" && err.Severity != rule.Severity {
		return false
	}

	if rule.Type != "" && err.Type != rule.Type {
		return false
	}

	// Check threshold within time window
	if rule.Threshold > 1 {
		count := em.getErrorCountInTimeWindow(err, rule.TimeWindow)
		return count >= rule.Threshold
	}

	return true
}

// getErrorCountInTimeWindow gets error count within time window
func (em *ErrorManager) getErrorCountInTimeWindow(err *ApplicationError, window time.Duration) int {
	query := `
		SELECT COUNT(*) FROM errors 
		WHERE type = ? AND severity = ? AND message = ?
		AND timestamp > DATE_SUB(NOW(), INTERVAL ? SECOND)
	`

	var count int
	em.db.QueryRow(query, err.Type, err.Severity, err.Message, int(window.Seconds())).Scan(&count)
	return count
}

// sendNotification sends error notification
func (em *ErrorManager) sendNotification(ctx context.Context, err *ApplicationError, rule NotificationRule) {
	if em.notifier != nil {
		em.notifier.Notify(ctx, err, rule)
	}
}

// GetErrorStats returns error statistics
func (em *ErrorManager) GetErrorStats(timeRange time.Duration) (*ErrorManagerStats, error) {
	stats := &ErrorManagerStats{
		ErrorsByType:     make(map[ErrorType]int),
		ErrorsBySeverity: make(map[ErrorSeverity]int),
		ResolutionTime:   make(map[ErrorSeverity]float64),
	}

	// Get total errors
	query := "SELECT COUNT(*) FROM errors WHERE timestamp > DATE_SUB(NOW(), INTERVAL ? SECOND)"
	em.db.QueryRow(query, int(timeRange.Seconds())).Scan(&stats.TotalErrors)

	// Get errors by type
	typeQuery := `
		SELECT type, COUNT(*) FROM errors 
		WHERE timestamp > DATE_SUB(NOW(), INTERVAL ? SECOND)
		GROUP BY type
	`
	rows, err := em.db.Query(typeQuery, int(timeRange.Seconds()))
	if err == nil {
		for rows.Next() {
			var errorType string
			var count int
			if err := rows.Scan(&errorType, &count); err == nil {
				stats.ErrorsByType[ErrorType(errorType)] = count
			}
		}
		rows.Close()
	}

	// Similar queries for other statistics...

	return stats, nil
}

// ResolveError marks an error as resolved
func (em *ErrorManager) ResolveError(errorID, resolvedBy string) error {
	query := `
		UPDATE errors 
		SET resolved = TRUE, resolved_at = NOW(), resolved_by = ?
		WHERE id = ?
	`

	_, err := em.db.Exec(query, resolvedBy, errorID)
	return err
}

// registerDefaultHandlers registers default error handlers
func (em *ErrorManager) registerDefaultHandlers() {
	// Implementation would register default handlers for each error type
}

// HTTPErrorHandler converts application errors to HTTP responses
func (em *ErrorManager) HTTPErrorHandler(w http.ResponseWriter, r *http.Request, err *ApplicationError) {
	w.Header().Set("Content-Type", "application/json")

	// Determine HTTP status code based on error type
	statusCode := em.getHTTPStatusCode(err.Type)
	w.WriteHeader(statusCode)

	// Create error response
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    err.Code,
			"message": err.UserMessage,
			"type":    err.Type,
		},
		"request_id": err.Context.RequestID,
		"timestamp":  err.Timestamp.Format(time.RFC3339),
	}

	// Add details in development mode
	if em.config.Environment == "development" {
		response["debug"] = map[string]interface{}{
			"error_id":    err.ID,
			"message":     err.Message,
			"stack_trace": err.StackTrace,
			"context":     err.Context,
		}
	}

	json.NewEncoder(w).Encode(response)
}

// getHTTPStatusCode returns appropriate HTTP status code for error type
func (em *ErrorManager) getHTTPStatusCode(errorType ErrorType) int {
	switch errorType {
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeAuth:
		return http.StatusUnauthorized
	case ErrorTypeAuthorization:
		return http.StatusForbidden
	case ErrorTypeNetwork:
		return http.StatusServiceUnavailable
	case ErrorTypeDatabase, ErrorTypeSystem:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// Error implements the error interface
func (ae *ApplicationError) Error() string {
	return ae.Message
}

// WithCause adds a cause to the error
func (ae *ApplicationError) WithCause(cause error) *ApplicationError {
	if appErr, ok := cause.(*ApplicationError); ok {
		ae.Cause = appErr
	}
	return ae
}

// WithTag adds a tag to the error
func (ae *ApplicationError) WithTag(tag string) *ApplicationError {
	ae.Tags = append(ae.Tags, tag)
	return ae
}

// WithMetadata adds metadata to the error
func (ae *ApplicationError) WithMetadata(key string, value interface{}) *ApplicationError {
	ae.Metadata[key] = value
	return ae
}
