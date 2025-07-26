package database

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// DynamicRepository provides a completely dynamic database interface
type DynamicRepository struct {
	db *sql.DB
}

// NewDynamicRepository creates a new dynamic repository
func NewDynamicRepository(db *sql.DB) *DynamicRepository {
	return &DynamicRepository{db: db}
}

// DynamicQuery represents a dynamic query
type DynamicQuery struct {
	db            *sql.DB
	Table         string
	Operation     string // "SELECT", "INSERT", "UPDATE", "DELETE"
	Fields        []string
	Values        []interface{}
	Conditions    map[string]interface{}
	Joins         []Join
	GroupByClause string
	HavingClause  string
	OrderByClause string
	LimitValue    int
	OffsetValue   int
	ReturnID      bool
	Transaction   *sql.Tx
}

// Join represents a table join
type Join struct {
	Table     string
	Type      string // "INNER", "LEFT", "RIGHT"
	Condition string
}

// NewQuery creates a new dynamic query
func (r *DynamicRepository) NewQuery(table string) *DynamicQuery {
	return &DynamicQuery{
		db:         r.db,
		Table:      table,
		Operation:  "SELECT",
		Conditions: make(map[string]interface{}),
		Joins:      make([]Join, 0),
	}
}

// Select sets the fields to select
func (q *DynamicQuery) Select(fields ...string) *DynamicQuery {
	q.Operation = "SELECT"
	q.Fields = fields
	return q
}

// Insert prepares an insert operation
func (q *DynamicQuery) Insert(data interface{}) *DynamicQuery {
	q.Operation = "INSERT"

	// Get fields and values from the struct
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	fields := make([]string, 0)
	values := make([]interface{}, 0)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Interface()

		// Skip zero values and unexported fields
		if !reflect.DeepEqual(value, reflect.Zero(field.Type).Interface()) && field.IsExported() {
			fields = append(fields, field.Name)
			values = append(values, value)
		}
	}

	q.Fields = fields
	q.Values = values
	return q
}

// Update prepares an update operation
func (q *DynamicQuery) Update(data interface{}) *DynamicQuery {
	q.Operation = "UPDATE"

	// Get fields and values from the struct
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	fields := make([]string, 0)
	values := make([]interface{}, 0)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Interface()

		// Skip zero values and unexported fields
		if !reflect.DeepEqual(value, reflect.Zero(field.Type).Interface()) && field.IsExported() {
			fields = append(fields, field.Name)
			values = append(values, value)
		}
	}

	q.Fields = fields
	q.Values = values
	return q
}

// Delete prepares a delete operation
func (q *DynamicQuery) Delete() *DynamicQuery {
	q.Operation = "DELETE"
	return q
}

// Where adds a condition
func (q *DynamicQuery) Where(field string, value interface{}) *DynamicQuery {
	q.Conditions[field] = value
	return q
}

// WhereIn adds a WHERE IN condition
func (q *DynamicQuery) WhereIn(field string, values []interface{}) *DynamicQuery {
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = "?"
	}
	q.Conditions[field+" IN ("+strings.Join(placeholders, ",")+")"] = values
	return q
}

// Join adds a table join
func (q *DynamicQuery) Join(joinType, table, condition string) *DynamicQuery {
	q.Joins = append(q.Joins, Join{
		Type:      joinType,
		Table:     table,
		Condition: condition,
	})
	return q
}

// GroupBy sets the GROUP BY clause
func (q *DynamicQuery) GroupBy(groupBy string) *DynamicQuery {
	q.GroupByClause = groupBy
	return q
}

// Having sets the HAVING clause
func (q *DynamicQuery) Having(having string) *DynamicQuery {
	q.HavingClause = having
	return q
}

// OrderBy sets the ORDER BY clause
func (q *DynamicQuery) OrderBy(orderBy string) *DynamicQuery {
	q.OrderByClause = orderBy
	return q
}

// Limit sets the LIMIT clause
func (q *DynamicQuery) Limit(limit int) *DynamicQuery {
	q.LimitValue = limit
	return q
}

// Offset sets the OFFSET clause
func (q *DynamicQuery) Offset(offset int) *DynamicQuery {
	q.OffsetValue = offset
	return q
}

// WithTransaction sets the transaction
func (q *DynamicQuery) WithTransaction(tx *sql.Tx) *DynamicQuery {
	q.Transaction = tx
	return q
}

// Build constructs the SQL query
func (q *DynamicQuery) Build() (string, []interface{}) {
	var query strings.Builder
	args := make([]interface{}, 0)

	switch q.Operation {
	case "SELECT":
		query.WriteString("SELECT ")
		if len(q.Fields) > 0 {
			query.WriteString(strings.Join(q.Fields, ", "))
		} else {
			query.WriteString("*")
		}
		query.WriteString(" FROM ")
		query.WriteString(q.Table)

	case "INSERT":
		query.WriteString("INSERT INTO ")
		query.WriteString(q.Table)
		query.WriteString(" (")
		query.WriteString(strings.Join(q.Fields, ", "))
		query.WriteString(") VALUES (")
		placeholders := make([]string, len(q.Fields))
		for i := range q.Fields {
			placeholders[i] = "?"
		}
		query.WriteString(strings.Join(placeholders, ", "))
		query.WriteString(")")
		args = append(args, q.Values...)
		return query.String(), args

	case "UPDATE":
		query.WriteString("UPDATE ")
		query.WriteString(q.Table)
		query.WriteString(" SET ")
		setClause := make([]string, len(q.Fields))
		for i, field := range q.Fields {
			setClause[i] = fmt.Sprintf("%s = ?", field)
		}
		query.WriteString(strings.Join(setClause, ", "))
		args = append(args, q.Values...)

	case "DELETE":
		query.WriteString("DELETE FROM ")
		query.WriteString(q.Table)
	}

	// Add joins
	if len(q.Joins) > 0 {
		for _, join := range q.Joins {
			query.WriteString(fmt.Sprintf(" %s JOIN %s ON %s", join.Type, join.Table, join.Condition))
		}
	}

	// Add where conditions
	if len(q.Conditions) > 0 {
		query.WriteString(" WHERE ")
		whereClause := make([]string, 0)
		for field, value := range q.Conditions {
			if strings.Contains(field, " IN ") {
				whereClause = append(whereClause, field)
				if values, ok := value.([]interface{}); ok {
					args = append(args, values...)
				}
			} else {
				whereClause = append(whereClause, fmt.Sprintf("%s = ?", field))
				args = append(args, value)
			}
		}
		query.WriteString(strings.Join(whereClause, " AND "))
	}

	// Add group by
	if q.GroupByClause != "" {
		query.WriteString(" GROUP BY ")
		query.WriteString(q.GroupByClause)
	}

	// Add having
	if q.HavingClause != "" {
		query.WriteString(" HAVING ")
		query.WriteString(q.HavingClause)
	}

	// Add order by
	if q.OrderByClause != "" {
		query.WriteString(" ORDER BY ")
		query.WriteString(q.OrderByClause)
	}

	// Add limit and offset
	if q.LimitValue > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", q.LimitValue))
		if q.OffsetValue > 0 {
			query.WriteString(fmt.Sprintf(" OFFSET %d", q.OffsetValue))
		}
	}

	return query.String(), args
}

// Execute executes the query
func (q *DynamicQuery) Execute() (sql.Result, error) {
	query, args := q.Build()

	if q.Transaction != nil {
		return q.Transaction.Exec(query, args...)
	}
	return q.db.Exec(query, args...)
}

// Query executes the query and returns rows
func (q *DynamicQuery) Query() (*sql.Rows, error) {
	query, args := q.Build()

	if q.Transaction != nil {
		return q.Transaction.Query(query, args...)
	}
	return q.db.Query(query, args...)
}

// QueryRow executes the query and returns a single row
func (q *DynamicQuery) QueryRow() *sql.Row {
	query, args := q.Build()

	if q.Transaction != nil {
		return q.Transaction.QueryRow(query, args...)
	}
	return q.db.QueryRow(query, args...)
}

// Scan scans the result into the provided interface
func (q *DynamicQuery) Scan(result interface{}) error {
	rows, err := q.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	v := reflect.ValueOf(result)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("result must be a pointer to a slice")
	}
	v = v.Elem()
	t := v.Type().Elem()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return err
		}

		item := reflect.New(t).Elem()
		for i, col := range columns {
			field := item.FieldByName(col)
			if field.IsValid() && field.CanSet() {
				val := values[i]
				if val != nil {
					field.Set(reflect.ValueOf(val))
				}
			}
		}

		v.Set(reflect.Append(v, item))
	}

	return rows.Err()
}

// NewDynamicQuery creates a new dynamic query
func NewDynamicQuery(db *sql.DB, table string) *DynamicQuery {
	return &DynamicQuery{
		db:         db,
		Table:      table,
		Operation:  "SELECT",
		Conditions: make(map[string]interface{}),
		Joins:      make([]Join, 0),
	}
}
