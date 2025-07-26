package database

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// DatabaseError represents a database error
type DatabaseError struct {
	Code    string
	Message string
	Err     error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
}

// Repository interface for database operations
type Repository interface {
	Create(table string, fields []string, values []interface{}) (int64, error)
	Update(table string, id interface{}, data interface{}) error
	Delete(table string, id interface{}) error
	FindByID(table string, id interface{}, result interface{}) error
	FindAll(table string, result interface{}, conditions map[string]interface{}, orderBy string, limit, offset int) error
	FindOne(table string, result interface{}, conditions map[string]interface{}) error
	Count(table string, conditions map[string]interface{}) (int64, error)
	Search(table string, fields []string, term string, limit, offset int, result interface{}) error
	FindByDateRange(table, dateField string, start, end time.Time, limit, offset int, result interface{}) error
	Transaction(fn func(*sql.Tx) error) error
	SetConnectionPool(maxOpen, maxIdle int, maxLifetime time.Duration)
	SoftDelete(table string, id interface{}) error
	BulkCreate(table string, data []interface{}) ([]int64, error)
	BulkUpdate(table string, ids []interface{}, data interface{}) error
	BulkDelete(table string, ids []interface{}) error
	Exists(table string, conditions map[string]interface{}) (bool, error)
}

// MySQLRepository represents a MySQL database repository
type MySQLRepository struct {
	db *sql.DB
}

// NewMySQLRepository creates a new MySQL repository
func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

// SetConnectionPool sets the connection pool parameters
func (r *MySQLRepository) SetConnectionPool(maxOpen, maxIdle int, maxLifetime time.Duration) {
	r.db.SetMaxOpenConns(maxOpen)
	r.db.SetMaxIdleConns(maxIdle)
	r.db.SetConnMaxLifetime(maxLifetime)
}

// validateTableName checks if the table name is valid
func validateTableName(table string) bool {
	// Basit bir regex kontrolü - sadece harf, rakam ve alt çizgi
	for _, c := range table {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

// Create inserts a new record
func (r *MySQLRepository) Create(table string, fields []string, values []interface{}) (int64, error) {
	if !validateTableName(table) {
		return 0, &DatabaseError{
			Code:    "INVALID_TABLE",
			Message: fmt.Sprintf("invalid table name: %s", table),
		}
	}

	qb := NewQueryBuilder(table)
	data := make(map[string]interface{})
	for i, field := range fields {
		// Byte array'i string'e dönüştür
		value := values[i]
		if byteArray, ok := value.([]byte); ok {
			value = string(byteArray)
		}
		data[field] = value
	}

	query, args := qb.BuildInsert(data)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return 0, &DatabaseError{
			Code:    "PREPARE_ERROR",
			Message: "error preparing statement",
			Err:     err,
		}
	}
	defer stmt.Close()

	result, err := stmt.Exec(args...)
	if err != nil {
		return 0, &DatabaseError{
			Code:    "EXEC_ERROR",
			Message: "error executing query",
			Err:     err,
		}
	}

	return result.LastInsertId()
}

// Update updates a record
func (r *MySQLRepository) Update(table string, id interface{}, data interface{}) error {
	if !validateTableName(table) {
		return &DatabaseError{
			Code:    "INVALID_TABLE",
			Message: fmt.Sprintf("invalid table name: %s", table),
		}
	}

	fields, values := getFieldsAndValues(data)
	qb := NewQueryBuilder(table)
	dataMap := make(map[string]interface{})
	for i, field := range fields {
		// Byte array'i string'e dönüştür
		value := values[i]
		if byteArray, ok := value.([]byte); ok {
			value = string(byteArray)
		}
		dataMap[field] = value
	}

	query, args := qb.Where("id", Equal, id).BuildUpdate(dataMap)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return &DatabaseError{
			Code:    "PREPARE_ERROR",
			Message: "error preparing statement",
			Err:     err,
		}
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return &DatabaseError{
			Code:    "EXEC_ERROR",
			Message: "error executing query",
			Err:     err,
		}
	}

	return nil
}

// Delete deletes a record
func (r *MySQLRepository) Delete(table string, id interface{}) error {
	if !validateTableName(table) {
		return &DatabaseError{
			Code:    "INVALID_TABLE",
			Message: fmt.Sprintf("invalid table name: %s", table),
		}
	}

	qb := NewQueryBuilder(table)
	query, args := qb.Where("id", Equal, id).BuildDelete()
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return &DatabaseError{
			Code:    "PREPARE_ERROR",
			Message: "error preparing statement",
			Err:     err,
		}
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return &DatabaseError{
			Code:    "EXEC_ERROR",
			Message: "error executing query",
			Err:     err,
		}
	}

	return nil
}

// FindByID finds a record by ID
func (r *MySQLRepository) FindByID(table string, id interface{}, result interface{}) error {
	if !validateTableName(table) {
		return &DatabaseError{
			Code:    "INVALID_TABLE",
			Message: fmt.Sprintf("invalid table name: %s", table),
		}
	}

	qb := NewQueryBuilder(table)
	query, args := qb.FindByID(id)
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return &DatabaseError{
			Code:    "PREPARE_ERROR",
			Message: "error preparing statement",
			Err:     err,
		}
	}
	defer stmt.Close()

	row := stmt.QueryRow(args...)
	return r.scanRowToStruct(row, result)
}

// scanRowToStruct scans a single row into a struct using reflection
func (r *MySQLRepository) scanRowToStruct(row *sql.Row, dest interface{}) error {
	// Get reflection values
	structValue := reflect.ValueOf(dest)
	if structValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer to struct")
	}
	
	structValue = structValue.Elem()
	if structValue.Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to struct")
	}

	// Create scan destinations based on struct fields
	scanDests := make([]interface{}, structValue.NumField())
	
	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		if field.CanSet() {
			scanDests[i] = field.Addr().Interface()
		} else {
			var dummy interface{}
			scanDests[i] = &dummy
		}
	}

	return row.Scan(scanDests...)
}

// FindAll finds all records with pagination
func (r *MySQLRepository) FindAll(table string, orderBy string, limit, offset int, dest interface{}) error {
	if !validateTableName(table) {
		return fmt.Errorf("invalid table name: %s", table)
	}

	qb := NewQueryBuilder(table)
	
	// Parse orderBy string (e.g., "created_at DESC" -> field: "created_at", direction: DESC)
	field := "id"
	direction := Descending
	
	if orderBy != "" {
		parts := strings.Fields(orderBy)
		if len(parts) >= 1 {
			field = parts[0]
		}
		if len(parts) >= 2 {
			if strings.ToUpper(parts[1]) == "ASC" {
				direction = Ascending
			} else {
				direction = Descending
			}
		}
	}
	
	query, args := qb.OrderBy(field, direction).Limit(limit).Offset(offset).Build()
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	// Use reflection to populate the slice
	return r.scanRowsToSlice(rows, dest)
}

// scanRowsToSlice scans multiple rows into a slice using reflection
func (r *MySQLRepository) scanRowsToSlice(rows *sql.Rows, dest interface{}) error {
	// Get reflection values
	sliceValue := reflect.ValueOf(dest)
	if sliceValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer to slice")
	}
	
	sliceValue = sliceValue.Elem()
	if sliceValue.Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to slice")
	}

	// Get element type
	elemType := sliceValue.Type().Elem()
	
	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("error getting columns: %v", err)
	}

	// Scan rows
	for rows.Next() {
		// Create new element
		elemPtr := reflect.New(elemType)
		elem := elemPtr.Elem()

		// Create scan destinations
		scanDests := make([]interface{}, len(columns))
		for i, col := range columns {
			// Find field by db tag or name
			field := r.findFieldByColumn(elem, col)
			if field.IsValid() && field.CanSet() {
				scanDests[i] = field.Addr().Interface()
			} else {
				// Use a dummy variable for unknown columns
				var dummy interface{}
				scanDests[i] = &dummy
			}
		}

		// Scan the row
		if err := rows.Scan(scanDests...); err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}

		// Append to slice
		sliceValue.Set(reflect.Append(sliceValue, elem))
	}

	return rows.Err()
}

// findFieldByColumn finds a struct field by column name using db tag
func (r *MySQLRepository) findFieldByColumn(elem reflect.Value, columnName string) reflect.Value {
	elemType := elem.Type()
	
	for i := 0; i < elem.NumField(); i++ {
		field := elemType.Field(i)
		dbTag := field.Tag.Get("db")
		
		// Check db tag first, then field name
		if dbTag == columnName || (dbTag == "" && strings.ToLower(field.Name) == strings.ToLower(columnName)) {
			return elem.Field(i)
		}
	}
	
	return reflect.Value{}
}

// FindOne finds a single record
func (r *MySQLRepository) FindOne(table string, conditions map[string]interface{}, dest interface{}) error {
	if !validateTableName(table) {
		return &DatabaseError{
			Code:    "INVALID_TABLE",
			Message: fmt.Sprintf("invalid table name: %s", table),
		}
	}

	qb := NewQueryBuilder(table)
	qb.Filter(conditions)
	query, args := qb.Limit(1).Build()
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return &DatabaseError{
			Code:    "PREPARE_ERROR",
			Message: "error preparing statement",
			Err:     err,
		}
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return &DatabaseError{
			Code:    "QUERY_ERROR",
			Message: "error executing query",
			Err:     err,
		}
	}
	defer rows.Close()

	if !rows.Next() {
		return sql.ErrNoRows
	}

	// Struct tipini kontrol et
	v := reflect.ValueOf(dest)
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		// Struct için özel işlem
		columns, err := rows.Columns()
		if err != nil {
			return &DatabaseError{
				Code:    "COLUMN_ERROR",
				Message: "error getting columns",
				Err:     err,
			}
		}

		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		err = rows.Scan(valuePtrs...)
		if err != nil {
			return &DatabaseError{
				Code:    "SCAN_ERROR",
				Message: "error scanning row",
				Err:     err,
			}
		}

		// Struct'ı doldur
		elem := v.Elem()
		for i, col := range columns {
			field := elem.FieldByName(col)
			if field.IsValid() && field.CanSet() {
				val := values[i]
				if val != nil {
					// Byte array'i string'e dönüştür
					if byteArray, ok := val.([]byte); ok {
						// Eğer şifre alanıysa ve bcrypt formatındaysa, doğrudan byte array'i stringe çevir
						if col == "password" && (strings.HasPrefix(string(byteArray), "$2a$") ||
							strings.HasPrefix(string(byteArray), "$2b$") ||
							strings.HasPrefix(string(byteArray), "$2y$")) {
							val = string(byteArray)
						} else {
							val = string(byteArray)
						}
					}

					// Değeri set et
					if field.Type().Kind() == reflect.String {
						field.SetString(fmt.Sprintf("%v", val))
					} else {
						field.Set(reflect.ValueOf(val))
					}
				}
			}
		}

		return nil
	}

	// Basit tip için doğrudan Scan kullan
	var result interface{}
	err = rows.Scan(&result)
	if err != nil {
		return &DatabaseError{
			Code:    "SCAN_ERROR",
			Message: "error scanning row",
			Err:     err,
		}
	}

	// Byte array'i string'e dönüştür
	if byteArray, ok := result.([]byte); ok {
		result = string(byteArray)
	}

	reflect.ValueOf(dest).Elem().Set(reflect.ValueOf(result))
	return nil
}

// Count counts records
func (r *MySQLRepository) Count(table string, conditions map[string]interface{}) (int64, error) {
	if !validateTableName(table) {
		return 0, fmt.Errorf("invalid table name: %s", table)
	}

	qb := NewQueryBuilder(table)
	qb.Filter(conditions)
	query, args := qb.BuildCount()
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("error preparing statement: %v", err)
	}
	defer stmt.Close()

	var count int64
	err = stmt.QueryRow(args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error executing query: %v", err)
	}

	return count, nil
}

// Search searches records
func (r *MySQLRepository) Search(table string, fields []string, term string, limit, offset int, dest interface{}) error {
	if !validateTableName(table) {
		return fmt.Errorf("invalid table name: %s", table)
	}

	qb := NewQueryBuilder(table)
	qb.Search(fields, term)
	query, args := qb.Limit(limit).Offset(offset).Build()
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	return rows.Scan(dest)
}

// FindByDateRange finds records within a date range
func (r *MySQLRepository) FindByDateRange(table, dateField string, start, end time.Time, limit, offset int, dest interface{}) error {
	if !validateTableName(table) {
		return fmt.Errorf("invalid table name: %s", table)
	}

	qb := NewQueryBuilder(table)
	query, args := qb.WhereDateBetween(dateField, start, end).Limit(limit).Offset(offset).Build()
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	return rows.Scan(dest)
}

// Transaction executes a function within a transaction
func (r *MySQLRepository) Transaction(fn func(*sql.Tx) error) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("error rolling back transaction: %v (original error: %v)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

// Close closes the database connection
func (r *MySQLRepository) Close() error {
	return r.db.Close()
}

// SoftDelete performs a soft delete operation
func (r *MySQLRepository) SoftDelete(table string, id interface{}) error {
	return r.Update(table, id, map[string]interface{}{
		"deleted_at": time.Now(),
		"is_deleted": true,
	})
}

// BulkCreate performs bulk insert operations
func (r *MySQLRepository) BulkCreate(table string, data []interface{}) ([]int64, error) {
	if !validateTableName(table) {
		return nil, &DatabaseError{
			Code:    "INVALID_TABLE",
			Message: fmt.Sprintf("invalid table name: %s", table),
		}
	}

	ids := make([]int64, 0, len(data))
	for _, item := range data {
		fields, values := getFieldsAndValues(item)
		id, err := r.Create(table, fields, values)
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// BulkUpdate performs bulk update operations
func (r *MySQLRepository) BulkUpdate(table string, ids []interface{}, data interface{}) error {
	if !validateTableName(table) {
		return &DatabaseError{
			Code:    "INVALID_TABLE",
			Message: fmt.Sprintf("invalid table name: %s", table),
		}
	}

	for _, id := range ids {
		if err := r.Update(table, id, data); err != nil {
			return err
		}
	}
	return nil
}

// BulkDelete performs bulk delete operations
func (r *MySQLRepository) BulkDelete(table string, ids []interface{}) error {
	if !validateTableName(table) {
		return &DatabaseError{
			Code:    "INVALID_TABLE",
			Message: fmt.Sprintf("invalid table name: %s", table),
		}
	}

	for _, id := range ids {
		if err := r.Delete(table, id); err != nil {
			return err
		}
	}
	return nil
}

// Exists checks if a record exists
func (r *MySQLRepository) Exists(table string, conditions map[string]interface{}) (bool, error) {
	if !validateTableName(table) {
		return false, &DatabaseError{
			Code:    "INVALID_TABLE",
			Message: fmt.Sprintf("invalid table name: %s", table),
		}
	}

	qb := NewQueryBuilder(table)
	qb.Filter(conditions)
	query, args := qb.BuildCount()
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return false, &DatabaseError{
			Code:    "PREPARE_ERROR",
			Message: "error preparing statement",
			Err:     err,
		}
	}
	defer stmt.Close()

	var count int64
	err = stmt.QueryRow(args...).Scan(&count)
	if err != nil {
		return false, &DatabaseError{
			Code:    "EXEC_ERROR",
			Message: "error executing query",
			Err:     err,
		}
	}

	return count > 0, nil
}

// Helper functions

// getFieldsAndValues extracts fields and values from a struct
func getFieldsAndValues(data interface{}) ([]string, []interface{}) {
	fields := make([]string, 0)
	values := make([]interface{}, 0)

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Map {
		mapType := val.Type()
		if mapType.Key().Kind() == reflect.String {
			iter := val.MapRange()
			for iter.Next() {
				key := iter.Key().String()
				value := iter.Value().Interface()
				fields = append(fields, key)
				values = append(values, value)
			}
		}
		return fields, values
	}

	if val.Kind() != reflect.Struct {
		return fields, values
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Get field tag or use field name
		tag := field.Tag.Get("db")
		if tag == "-" {
			continue
		}
		if tag == "" {
			tag = field.Name
		}

		// Handle embedded structs
		if field.Anonymous && val.Field(i).Kind() == reflect.Struct {
			embeddedFields, embeddedValues := getFieldsAndValues(val.Field(i).Interface())
			fields = append(fields, embeddedFields...)
			values = append(values, embeddedValues...)
			continue
		}

		fields = append(fields, tag)
		values = append(values, val.Field(i).Interface())
	}

	return fields, values
}
