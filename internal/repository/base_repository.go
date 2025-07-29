package repository

import (
	"database/sql"
	"reflect"
	"time"

	"kolajAi/internal/database"
)

// BaseRepository provides common database operations
type BaseRepository struct {
	db *database.MySQLRepository
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *database.MySQLRepository) *BaseRepository {
	return &BaseRepository{db: db}
}

// Create inserts a new record
func (r *BaseRepository) Create(table string, data interface{}) (int64, error) {
	fields, values := getFieldsAndValues(data)
	return r.db.Create(table, fields, values)
}

// Update updates a record
func (r *BaseRepository) Update(table string, id interface{}, data interface{}) error {
	return r.db.Update(table, id, data)
}

// Delete removes a record
func (r *BaseRepository) Delete(table string, id interface{}) error {
	return r.db.Delete(table, id)
}

// FindByID retrieves a record by its ID
func (r *BaseRepository) FindByID(table string, id interface{}, result interface{}) error {
	return r.db.FindByID(table, id, result)
}

// FindAll retrieves multiple records
func (r *BaseRepository) FindAll(table string, result interface{}, conditions map[string]interface{}, orderBy string, limit, offset int) error {
	return r.db.FindAll(table, orderBy, limit, offset, result)
}

// FindOne retrieves a single record
func (r *BaseRepository) FindOne(table string, result interface{}, conditions map[string]interface{}) error {
	return r.db.FindOne(table, conditions, result)
}

// Count returns the number of records
func (r *BaseRepository) Count(table string, conditions map[string]interface{}) (int64, error) {
	return r.db.Count(table, conditions)
}

// Search searches records
func (r *BaseRepository) Search(table string, fields []string, term string, limit, offset int, result interface{}) error {
	return r.db.Search(table, fields, term, limit, offset, result)
}

// FindByDateRange finds records within a date range
func (r *BaseRepository) FindByDateRange(table, dateField string, start, end time.Time, limit, offset int, result interface{}) error {
	return r.db.FindByDateRange(table, dateField, start, end, limit, offset, result)
}

// Transaction executes a function within a transaction
func (r *BaseRepository) Transaction(fn func(*sql.Tx) error) error {
	return r.db.Transaction(fn)
}

// Exists checks if a record exists
func (r *BaseRepository) Exists(table string, conditions map[string]interface{}) (bool, error) {
	return r.db.Exists(table, conditions)
}

// Helper functions

// getFieldsAndValues extracts fields and values from a struct or map
func getFieldsAndValues(data interface{}) ([]string, []interface{}) {
	fields := make([]string, 0)
	values := make([]interface{}, 0)

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Map tipini kontrol et
	if v.Kind() == reflect.Map {
		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			value := iter.Value().Interface()

			// Byte array'i string'e dönüştür
			if byteArray, ok := value.([]byte); ok {
				value = string(byteArray)
			}

			fields = append(fields, key)
			values = append(values, value)
		}
		return fields, values
	}

	// Struct tipini işle
	if v.Kind() == reflect.Struct {
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			value := v.Field(i).Interface()

			// Skip zero values and unexported fields
			if !reflect.DeepEqual(value, reflect.Zero(field.Type).Interface()) && field.IsExported() {
				// Byte array'i string'e dönüştür
				if byteArray, ok := value.([]byte); ok {
					value = string(byteArray)
				}

				// Önce db tag'ini kontrol et, yoksa field adını kullan
				dbTag := field.Tag.Get("db")
				if dbTag != "" {
					fields = append(fields, dbTag)
				} else {
					fields = append(fields, field.Name)
				}
				values = append(values, value)
			}
		}
	}

	return fields, values
}

// CreateStruct creates a record from a struct (for SimpleRepository compatibility)
func (r *BaseRepository) CreateStruct(table string, data interface{}) (int64, error) {
	fields, values := getFieldsAndValues(data)
	return r.db.Create(table, fields, values)
}

// SetConnectionPool sets database connection pool parameters
func (r *BaseRepository) SetConnectionPool(maxOpen, maxIdle int, maxLifetime time.Duration) {
	r.db.SetConnectionPool(maxOpen, maxIdle, maxLifetime)
}

// SoftDelete performs a soft delete by setting a deleted_at timestamp
func (r *BaseRepository) SoftDelete(table string, id interface{}) error {
	return r.db.SoftDelete(table, id)
}

// BulkCreate creates multiple records at once
func (r *BaseRepository) BulkCreate(table string, data []interface{}) ([]int64, error) {
	return r.db.BulkCreate(table, data)
}

// BulkUpdate updates multiple records at once
func (r *BaseRepository) BulkUpdate(table string, ids []interface{}, data interface{}) error {
	return r.db.BulkUpdate(table, ids, data)
}

// BulkDelete deletes multiple records at once
func (r *BaseRepository) BulkDelete(table string, ids []interface{}) error {
	return r.db.BulkDelete(table, ids)
}

// Begin starts a transaction
func (r *BaseRepository) Begin() (database.Transaction, error) {
	return r.db.Begin()
}

// Exec executes a query
func (r *BaseRepository) Exec(query string, args ...interface{}) (database.Result, error) {
	return r.db.Exec(query, args...)
}

// Query executes a query and returns rows
func (r *BaseRepository) Query(query string, args ...interface{}) (database.Rows, error) {
	return r.db.Query(query, args...)
}

// QueryRow executes a query and returns a single row
func (r *BaseRepository) QueryRow(query string, args ...interface{}) database.Row {
	return r.db.QueryRow(query, args...)
}
