package database

import (
	"database/sql"
	"reflect"
	"time"
)

// RepositoryWrapper wraps MySQLRepository to provide struct-based operations
type RepositoryWrapper struct {
	*MySQLRepository
}

// NewRepositoryWrapper creates a new repository wrapper
func NewRepositoryWrapper(repo *MySQLRepository) *RepositoryWrapper {
	return &RepositoryWrapper{MySQLRepository: repo}
}

// CreateStruct creates a record from a struct
func (r *RepositoryWrapper) CreateStruct(table string, data interface{}) (int64, error) {
	fields, values := r.structToFieldsAndValues(data)
	return r.MySQLRepository.Create(table, fields, values)
}

// structToFieldsAndValues converts a struct to fields and values
func (r *RepositoryWrapper) structToFieldsAndValues(data interface{}) ([]string, []interface{}) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	var fields []string
	var values []interface{}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip ID field for creation
		if field.Name == "ID" {
			continue
		}

		// Get db tag or use field name
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			dbTag = field.Name
		}

		fields = append(fields, dbTag)
		values = append(values, value.Interface())
	}

	return fields, values
}

// Exec executes a query without returning any rows
func (r *RepositoryWrapper) Exec(query string, args ...interface{}) (Result, error) {
	result, err := r.MySQLRepository.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return &resultWrapper{result: result}, nil
}

// Query executes a query that returns rows
func (r *RepositoryWrapper) Query(query string, args ...interface{}) (Rows, error) {
	rows, err := r.MySQLRepository.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return &rowsWrapper{rows: rows}, nil
}

// QueryRow executes a query that returns at most one row
func (r *RepositoryWrapper) QueryRow(query string, args ...interface{}) Row {
	row := r.MySQLRepository.db.QueryRow(query, args...)
	return &rowWrapper{row: row}
}

// Begin starts a transaction
func (r *RepositoryWrapper) Begin() (Transaction, error) {
	tx, err := r.MySQLRepository.db.Begin()
	if err != nil {
		return nil, err
	}
	return &txWrapper{tx: tx}, nil
}

// Wrapper types - using implementations from db.go

type txWrapper struct {
	tx *sql.Tx
}

func (t *txWrapper) Exec(query string, args ...interface{}) (Result, error) {
	result, err := t.tx.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return &resultWrapper{result: result}, nil
}

func (t *txWrapper) Commit() error {
	return t.tx.Commit()
}

func (t *txWrapper) Rollback() error {
	return t.tx.Rollback()
}

// SimpleRepository interface for struct-based operations
type SimpleRepository interface {
	CreateStruct(table string, data interface{}) (int64, error)
	Update(table string, id interface{}, data interface{}) error
	Delete(table string, id interface{}) error
	FindByID(table string, id interface{}, result interface{}) error
	FindAll(table string, result interface{}, conditions map[string]interface{}, orderBy string, limit, offset int) error
	FindOne(table string, result interface{}, conditions map[string]interface{}) error
	Count(table string, conditions map[string]interface{}) (int64, error)
	Search(table string, fields []string, term string, limit, offset int, result interface{}) error
	FindByDateRange(table, dateField string, start, end time.Time, limit, offset int, result interface{}) error
	SetConnectionPool(maxOpen, maxIdle int, maxLifetime time.Duration)
	SoftDelete(table string, id interface{}) error
	BulkCreate(table string, data []interface{}) ([]int64, error)
	BulkUpdate(table string, ids []interface{}, data interface{}) error
	BulkDelete(table string, ids []interface{}) error
	Exists(table string, conditions map[string]interface{}) (bool, error)
	Exec(query string, args ...interface{}) (Result, error)
	Begin() (Transaction, error)
	Query(query string, args ...interface{}) (Rows, error)
	QueryRow(query string, args ...interface{}) Row
}

// Result interface for SQL result
type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

// Rows interface for SQL rows
type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
}

// Row interface for SQL row
type Row interface {
	Scan(dest ...interface{}) error
}

// Transaction interface for SQL transaction
type Transaction interface {
	Exec(query string, args ...interface{}) (Result, error)
	Commit() error
	Rollback() error
}

// FindAll implements SimpleRepository interface
func (r *RepositoryWrapper) FindAll(table string, result interface{}, conditions map[string]interface{}, orderBy string, limit, offset int) error {
	// Simple implementation - just call the underlying FindAll with basic parameters
	return r.MySQLRepository.FindAll(table, orderBy, limit, offset, result)
}

// FindOne implements SimpleRepository interface
func (r *RepositoryWrapper) FindOne(table string, result interface{}, conditions map[string]interface{}) error {
	// Simple implementation - just call the underlying FindOne
	return r.MySQLRepository.FindOne(table, conditions, result)
}
