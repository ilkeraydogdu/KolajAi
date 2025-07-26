package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// AuditLog represents a database audit log entry
type AuditLog struct {
	TableName string                 `json:"table_name"`
	RecordID  interface{}            `json:"record_id"`
	Action    string                 `json:"action"`
	OldValues map[string]interface{} `json:"old_values,omitempty"`
	NewValues map[string]interface{} `json:"new_values,omitempty"`
	UserID    interface{}            `json:"user_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
}

// QueryLogger represents a query logger
type QueryLogger struct {
	logger *log.Logger
}

// NewQueryLogger creates a new query logger
func NewQueryLogger(logger *log.Logger) *QueryLogger {
	return &QueryLogger{logger: logger}
}

// LogQuery logs a database query
func (l *QueryLogger) LogQuery(query string, args []interface{}, duration time.Duration) {
	l.logger.Printf("Query: %s\nArgs: %v\nDuration: %v", query, args, duration)
}

// LogError logs a database error
func (l *QueryLogger) LogError(err error, query string, args []interface{}) {
	l.logger.Printf("Error: %v\nQuery: %s\nArgs: %v", err, query, args)
}

// AuditLogger represents an audit logger
type AuditLogger struct {
	logger *log.Logger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger *log.Logger) *AuditLogger {
	return &AuditLogger{logger: logger}
}

// LogAudit logs an audit entry
func (l *AuditLogger) LogAudit(log *AuditLog) error {
	data, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("error marshaling audit log: %v", err)
	}

	l.logger.Printf("Audit: %s", string(data))
	return nil
}

// AuditRepository represents a repository with audit logging
type AuditRepository struct {
	repo        Repository
	auditLogger *AuditLogger
	queryLogger *QueryLogger
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(repo Repository, auditLogger *AuditLogger, queryLogger *QueryLogger) *AuditRepository {
	return &AuditRepository{
		repo:        repo,
		auditLogger: auditLogger,
		queryLogger: queryLogger,
	}
}

// Create creates a record with audit logging
func (r *AuditRepository) Create(table string, fields []string, values []interface{}) (int64, error) {
	start := time.Now()
	id, err := r.repo.Create(table, fields, values)
	duration := time.Since(start)

	r.queryLogger.LogQuery("CREATE", []interface{}{table, fields, values}, duration)

	if err != nil {
		r.queryLogger.LogError(err, "CREATE", []interface{}{table, fields, values})
		return 0, err
	}

	// Create a map from fields and values for audit logging
	newValues := make(map[string]interface{})
	for i, field := range fields {
		if i < len(values) {
			newValues[field] = values[i]
		}
	}
	
	auditLog := &AuditLog{
		TableName: table,
		RecordID:  id,
		Action:    "CREATE",
		NewValues: newValues,
		Timestamp: time.Now(),
	}

	if err := r.auditLogger.LogAudit(auditLog); err != nil {
		r.queryLogger.LogError(err, "AUDIT_CREATE", []interface{}{auditLog})
	}

	return id, nil
}

// Update updates a record with audit logging
func (r *AuditRepository) Update(table string, id interface{}, data interface{}) error {
	start := time.Now()
	err := r.repo.Update(table, id, data)
	duration := time.Since(start)

	r.queryLogger.LogQuery("UPDATE", []interface{}{table, id, data}, duration)

	if err != nil {
		r.queryLogger.LogError(err, "UPDATE", []interface{}{table, id, data})
		return err
	}

	auditLog := &AuditLog{
		TableName: table,
		RecordID:  id,
		Action:    "UPDATE",
		NewValues: data.(map[string]interface{}),
		Timestamp: time.Now(),
	}

	if err := r.auditLogger.LogAudit(auditLog); err != nil {
		r.queryLogger.LogError(err, "AUDIT_UPDATE", []interface{}{auditLog})
	}

	return nil
}

// Delete deletes a record with audit logging
func (r *AuditRepository) Delete(table string, id interface{}) error {
	start := time.Now()
	err := r.repo.Delete(table, id)
	duration := time.Since(start)

	r.queryLogger.LogQuery("DELETE", []interface{}{table, id}, duration)

	if err != nil {
		r.queryLogger.LogError(err, "DELETE", []interface{}{table, id})
		return err
	}

	auditLog := &AuditLog{
		TableName: table,
		RecordID:  id,
		Action:    "DELETE",
		Timestamp: time.Now(),
	}

	if err := r.auditLogger.LogAudit(auditLog); err != nil {
		r.queryLogger.LogError(err, "AUDIT_DELETE", []interface{}{auditLog})
	}

	return nil
}

// FindByID finds a record by ID with audit logging
func (r *AuditRepository) FindByID(table string, id interface{}, result interface{}) error {
	start := time.Now()
	err := r.repo.FindByID(table, id, result)
	duration := time.Since(start)

	r.queryLogger.LogQuery("FIND_BY_ID", []interface{}{table, id}, duration)

	if err != nil {
		r.queryLogger.LogError(err, "FIND_BY_ID", []interface{}{table, id})
	}

	return err
}

// FindAll finds all records with audit logging
func (r *AuditRepository) FindAll(table string, result interface{}, conditions map[string]interface{}, orderBy string, limit, offset int) error {
	start := time.Now()
	err := r.repo.FindAll(table, result, conditions, orderBy, limit, offset)
	duration := time.Since(start)

	r.queryLogger.LogQuery("FIND_ALL", []interface{}{table, conditions, orderBy, limit, offset}, duration)

	if err != nil {
		r.queryLogger.LogError(err, "FIND_ALL", []interface{}{table, conditions, orderBy, limit, offset})
	}

	return err
}

// FindOne finds a single record with audit logging
func (r *AuditRepository) FindOne(table string, result interface{}, conditions map[string]interface{}) error {
	start := time.Now()
	err := r.repo.FindOne(table, result, conditions)
	duration := time.Since(start)

	r.queryLogger.LogQuery("FIND_ONE", []interface{}{table, conditions}, duration)

	if err != nil {
		r.queryLogger.LogError(err, "FIND_ONE", []interface{}{table, conditions})
	}

	return err
}

// Count returns the number of records with audit logging
func (r *AuditRepository) Count(table string, conditions map[string]interface{}) (int64, error) {
	start := time.Now()
	count, err := r.repo.Count(table, conditions)
	duration := time.Since(start)

	r.queryLogger.LogQuery("COUNT", []interface{}{table, conditions}, duration)

	if err != nil {
		r.queryLogger.LogError(err, "COUNT", []interface{}{table, conditions})
	}

	return count, err
}

// Search searches records with audit logging
func (r *AuditRepository) Search(table string, fields []string, term string, limit, offset int, result interface{}) error {
	start := time.Now()
	err := r.repo.Search(table, fields, term, limit, offset, result)
	duration := time.Since(start)

	r.queryLogger.LogQuery("SEARCH", []interface{}{table, fields, term, limit, offset}, duration)

	if err != nil {
		r.queryLogger.LogError(err, "SEARCH", []interface{}{table, fields, term, limit, offset})
	}

	return err
}

// FindByDateRange finds records within a date range with audit logging
func (r *AuditRepository) FindByDateRange(table, dateField string, startTime, endTime time.Time, limit, offset int, result interface{}) error {
	start := time.Now()
	err := r.repo.FindByDateRange(table, dateField, startTime, endTime, limit, offset, result)
	duration := time.Since(start)

	r.queryLogger.LogQuery("FIND_BY_DATE_RANGE", []interface{}{table, dateField, startTime, endTime, limit, offset}, duration)

	if err != nil {
		r.queryLogger.LogError(err, "FIND_BY_DATE_RANGE", []interface{}{table, dateField, startTime, endTime, limit, offset})
	}

	return err
}

// Transaction executes a function within a transaction with audit logging
func (r *AuditRepository) Transaction(fn func(*sql.Tx) error) error {
	start := time.Now()
	err := r.repo.Transaction(fn)
	duration := time.Since(start)

	r.queryLogger.LogQuery("TRANSACTION", nil, duration)

	if err != nil {
		r.queryLogger.LogError(err, "TRANSACTION", nil)
	}

	return err
}

// ... implement other Repository interface methods ...
