package database

import (
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
}