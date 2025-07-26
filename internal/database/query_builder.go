package database

import (
	"fmt"
	"strings"
	"time"
)

// QueryType represents the type of query
type QueryType string

const (
	Select QueryType = "SELECT"
	Insert QueryType = "INSERT"
	Update QueryType = "UPDATE"
	Delete QueryType = "DELETE"
	Count  QueryType = "COUNT"
)

// JoinType represents the type of join
type JoinType string

const (
	InnerJoin JoinType = "INNER JOIN"
	LeftJoin  JoinType = "LEFT JOIN"
	RightJoin JoinType = "RIGHT JOIN"
	FullJoin  JoinType = "FULL JOIN"
)

// Operator represents the comparison operator
type Operator string

const (
	Equal        Operator = "="
	NotEqual     Operator = "!="
	GreaterThan  Operator = ">"
	LessThan     Operator = "<"
	GreaterEqual Operator = ">="
	LessEqual    Operator = "<="
	Like         Operator = "LIKE"
	In           Operator = "IN"
	NotIn        Operator = "NOT IN"
	Between      Operator = "BETWEEN"
	IsNull       Operator = "IS NULL"
	IsNotNull    Operator = "IS NOT NULL"
)

// SortDirection represents the sort direction
type SortDirection string

const (
	Ascending  SortDirection = "ASC"
	Descending SortDirection = "DESC"
)

// Condition represents a where condition
type Condition struct {
	Field    string
	Operator Operator
	Value    interface{}
	Or       bool // If true, condition will be joined with OR instead of AND
}

// Subquery represents a subquery in the query
type Subquery struct {
	Query *QueryBuilder
	Alias string
}

// Case represents a CASE-WHEN expression
type Case struct {
	Field        string
	Cases        map[interface{}]interface{}
	DefaultValue interface{}
}

// QueryBuilder helps build SQL queries safely
type QueryBuilder struct {
	queryType  QueryType
	table      string
	selects    []string
	joins      []string
	conditions []Condition
	orderBy    []string
	groupBy    []string
	having     string
	limit      int
	offset     int
	values     map[string]interface{}
	subqueries []Subquery
	cases      []Case
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(table string) *QueryBuilder {
	return &QueryBuilder{
		table:      table,
		queryType:  Select,
		selects:    []string{"*"},
		conditions: make([]Condition, 0),
		orderBy:    make([]string, 0),
		groupBy:    make([]string, 0),
		values:     make(map[string]interface{}),
		subqueries: make([]Subquery, 0),
		cases:      make([]Case, 0),
	}
}

// Select sets the columns to select
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.queryType = Select
	qb.selects = columns
	return qb
}

// Insert prepares an insert query
func (qb *QueryBuilder) Insert(data map[string]interface{}) *QueryBuilder {
	qb.queryType = Insert
	qb.values = data
	return qb
}

// Update prepares an update query
func (qb *QueryBuilder) Update(data map[string]interface{}) *QueryBuilder {
	qb.queryType = Update
	qb.values = data
	return qb
}

// Delete prepares a delete query
func (qb *QueryBuilder) Delete() *QueryBuilder {
	qb.queryType = Delete
	return qb
}

// Count prepares a count query
func (qb *QueryBuilder) Count() *QueryBuilder {
	qb.queryType = Count
	return qb
}

// Join adds a join clause
func (qb *QueryBuilder) Join(joinType JoinType, table, condition string) *QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("%s %s ON %s", joinType, table, condition))
	return qb
}

// Where adds a where condition
func (qb *QueryBuilder) Where(field string, operator Operator, value interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return qb
}

// OrWhere adds a where condition with OR
func (qb *QueryBuilder) OrWhere(field string, operator Operator, value interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		Field:    field,
		Operator: operator,
		Value:    value,
		Or:       true,
	})
	return qb
}

// WhereIn adds a WHERE IN condition
func (qb *QueryBuilder) WhereIn(field string, values []interface{}) *QueryBuilder {
	return qb.Where(field, In, values)
}

// WhereNotIn adds a WHERE NOT IN condition
func (qb *QueryBuilder) WhereNotIn(field string, values []interface{}) *QueryBuilder {
	return qb.Where(field, NotIn, values)
}

// WhereLike adds a WHERE LIKE condition
func (qb *QueryBuilder) WhereLike(field, pattern string) *QueryBuilder {
	return qb.Where(field, Like, "%"+pattern+"%")
}

// WhereStartsWith adds a WHERE LIKE condition for prefix match
func (qb *QueryBuilder) WhereStartsWith(field, prefix string) *QueryBuilder {
	return qb.Where(field, Like, prefix+"%")
}

// WhereEndsWith adds a WHERE LIKE condition for suffix match
func (qb *QueryBuilder) WhereEndsWith(field, suffix string) *QueryBuilder {
	return qb.Where(field, Like, "%"+suffix)
}

// WhereBetween adds a WHERE BETWEEN condition
func (qb *QueryBuilder) WhereBetween(field string, start, end interface{}) *QueryBuilder {
	return qb.Where(field, Between, []interface{}{start, end})
}

// WhereNull adds a WHERE IS NULL condition
func (qb *QueryBuilder) WhereNull(field string) *QueryBuilder {
	return qb.Where(field, IsNull, nil)
}

// WhereNotNull adds a WHERE IS NOT NULL condition
func (qb *QueryBuilder) WhereNotNull(field string) *QueryBuilder {
	return qb.Where(field, IsNotNull, nil)
}

// WhereDate adds a WHERE condition for date comparison
func (qb *QueryBuilder) WhereDate(field string, operator Operator, date time.Time) *QueryBuilder {
	return qb.Where(field, operator, date.Format("2006-01-02"))
}

// WhereDateBetween adds a WHERE BETWEEN condition for date range
func (qb *QueryBuilder) WhereDateBetween(field string, start, end time.Time) *QueryBuilder {
	return qb.WhereBetween(field, start.Format("2006-01-02"), end.Format("2006-01-02"))
}

// OrderBy adds an ORDER BY clause
func (qb *QueryBuilder) OrderBy(field string, direction SortDirection) *QueryBuilder {
	qb.orderBy = append(qb.orderBy, fmt.Sprintf("%s %s", field, direction))
	return qb
}

// GroupBy adds a GROUP BY clause
func (qb *QueryBuilder) GroupBy(fields ...string) *QueryBuilder {
	qb.groupBy = append(qb.groupBy, fields...)
	return qb
}

// Having adds a HAVING clause
func (qb *QueryBuilder) Having(having string) *QueryBuilder {
	qb.having = having
	return qb
}

// Limit sets the limit
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset sets the offset
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Subquery adds a subquery to the query
func (qb *QueryBuilder) Subquery(subquery *QueryBuilder, alias string) *QueryBuilder {
	qb.subqueries = append(qb.subqueries, Subquery{
		Query: subquery,
		Alias: alias,
	})
	return qb
}

// Case adds a CASE-WHEN expression to the query
func (qb *QueryBuilder) Case(field string, cases map[interface{}]interface{}, defaultValue interface{}) *QueryBuilder {
	qb.cases = append(qb.cases, Case{
		Field:        field,
		Cases:        cases,
		DefaultValue: defaultValue,
	})
	return qb
}

// Validate validates the query builder
func (qb *QueryBuilder) Validate() error {
	if qb.table == "" {
		return fmt.Errorf("table name is required")
	}

	if qb.queryType == "" {
		return fmt.Errorf("query type is required")
	}

	if qb.queryType == Insert && len(qb.values) == 0 {
		return fmt.Errorf("values are required for insert query")
	}

	if qb.queryType == Update && len(qb.values) == 0 {
		return fmt.Errorf("values are required for update query")
	}

	return nil
}

// Build constructs the SQL query
func (qb *QueryBuilder) Build() (string, []interface{}) {
	if err := qb.Validate(); err != nil {
		return "", nil
	}

	var query strings.Builder
	args := make([]interface{}, 0)

	switch qb.queryType {
	case Select:
		query.WriteString("SELECT ")
		if len(qb.selects) > 0 {
			query.WriteString(strings.Join(qb.selects, ", "))
		} else {
			query.WriteString("*")
		}
		query.WriteString(" FROM ")
		query.WriteString(qb.table)

		// Add subqueries
		for _, subquery := range qb.subqueries {
			subquerySQL, subqueryArgs := subquery.Query.Build()
			query.WriteString(fmt.Sprintf(" (%s) AS %s", subquerySQL, subquery.Alias))
			args = append(args, subqueryArgs...)
		}

		// Add CASE expressions
		for _, caseExpr := range qb.cases {
			query.WriteString(fmt.Sprintf(" CASE %s", caseExpr.Field))
			for when, then := range caseExpr.Cases {
				query.WriteString(fmt.Sprintf(" WHEN %v THEN %v", when, then))
			}
			if caseExpr.DefaultValue != nil {
				query.WriteString(fmt.Sprintf(" ELSE %v", caseExpr.DefaultValue))
			}
			query.WriteString(" END")
		}

	case Count:
		query.WriteString("SELECT COUNT(*) FROM ")
		query.WriteString(qb.table)

	case Insert:
		query.WriteString("INSERT INTO ")
		query.WriteString(qb.table)
		query.WriteString(" (")
		columns := make([]string, 0, len(qb.values))
		values := make([]interface{}, 0, len(qb.values))
		for column, value := range qb.values {
			columns = append(columns, column)
			values = append(values, value)
		}
		query.WriteString(strings.Join(columns, ", "))
		query.WriteString(") VALUES (")
		placeholders := make([]string, len(columns))
		for i := range columns {
			placeholders[i] = "?"
		}
		query.WriteString(strings.Join(placeholders, ", "))
		query.WriteString(")")
		args = append(args, values...)
		return query.String(), args

	case Update:
		query.WriteString("UPDATE ")
		query.WriteString(qb.table)
		query.WriteString(" SET ")
		setClause := make([]string, 0, len(qb.values))
		for field, value := range qb.values {
			setClause = append(setClause, fmt.Sprintf("%s = ?", field))
			args = append(args, value)
		}
		query.WriteString(strings.Join(setClause, ", "))

	case Delete:
		query.WriteString("DELETE FROM ")
		query.WriteString(qb.table)
	}

	// Add joins
	if len(qb.joins) > 0 {
		for _, join := range qb.joins {
			query.WriteString(fmt.Sprintf(" %s", join))
		}
	}

	// Add where conditions
	if len(qb.conditions) > 0 {
		query.WriteString(" WHERE ")
		whereClause := make([]string, 0)
		for i, condition := range qb.conditions {
			// İlk koşul değilse ve OR değilse, AND ekle
			if i > 0 && !condition.Or {
				whereClause = append(whereClause, "AND")
			} else if i > 0 && condition.Or {
				whereClause = append(whereClause, "OR")
			}

			// Operatör tipine göre koşul oluştur
			if condition.Operator == IsNull || condition.Operator == IsNotNull {
				whereClause = append(whereClause, fmt.Sprintf("%s %s", condition.Field, condition.Operator))
			} else {
				whereClause = append(whereClause, fmt.Sprintf("%s %s ?", condition.Field, condition.Operator))
				args = append(args, condition.Value)
			}
		}
		query.WriteString(strings.Join(whereClause, " "))
	}

	// Add group by
	if len(qb.groupBy) > 0 {
		query.WriteString(" GROUP BY ")
		query.WriteString(strings.Join(qb.groupBy, ", "))
	}

	// Add having
	if qb.having != "" {
		query.WriteString(" HAVING ")
		query.WriteString(qb.having)
	}

	// Add order by
	if len(qb.orderBy) > 0 {
		query.WriteString(" ORDER BY ")
		query.WriteString(strings.Join(qb.orderBy, ", "))
	}

	// Add limit and offset
	if qb.limit > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", qb.limit))
		if qb.offset > 0 {
			query.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
		}
	}

	return query.String(), args
}

// Helper functions for common queries

// FindByID constructs a query to find a record by ID
func (qb *QueryBuilder) FindByID(id interface{}) (string, []interface{}) {
	return qb.Where("id", Equal, id).Build()
}

// FindByField constructs a query to find records by a field value
func (qb *QueryBuilder) FindByField(field string, value interface{}) (string, []interface{}) {
	return qb.Where(field, Equal, value).Build()
}

// FindByFields constructs a query to find records by multiple field values
func (qb *QueryBuilder) FindByFields(fields map[string]interface{}) (string, []interface{}) {
	for field, value := range fields {
		qb.Where(field, Equal, value)
	}
	return qb.Build()
}

// Search constructs a query to search records
func (qb *QueryBuilder) Search(fields []string, term string) (string, []interface{}) {
	if len(fields) == 0 {
		return qb.Build()
	}

	conditions := make([]Condition, 0, len(fields))
	for _, field := range fields {
		conditions = append(conditions, Condition{
			Field:    field,
			Operator: Like,
			Value:    "%" + term + "%",
		})
	}

	qb.conditions = append(qb.conditions, conditions...)
	return qb.Build()
}

// DateRange constructs a query to find records within a date range
func (qb *QueryBuilder) DateRange(field string, start, end time.Time) (string, []interface{}) {
	return qb.WhereBetween(field, start, end).Build()
}

// Paginate constructs a query with pagination
func (qb *QueryBuilder) Paginate(page, perPage int) (string, []interface{}) {
	offset := (page - 1) * perPage
	return qb.Limit(perPage).Offset(offset).Build()
}

// Filter constructs a query with filters
func (qb *QueryBuilder) Filter(filters map[string]interface{}) *QueryBuilder {
	if filters == nil {
		return qb
	}

	for field, value := range filters {
		switch v := value.(type) {
		case []interface{}:
			qb.WhereIn(field, v)
		case string:
			if strings.HasPrefix(v, "%") || strings.HasSuffix(v, "%") {
				qb.Where(field, Like, v)
			} else {
				qb.Where(field, Equal, v)
			}
		default:
			qb.Where(field, Equal, v)
		}
	}
	return qb
}

// Sort constructs a query with sorting
func (qb *QueryBuilder) Sort(field string, direction SortDirection) (string, []interface{}) {
	return qb.OrderBy(field, direction).Build()
}

// BuildCount constructs a count query
func (qb *QueryBuilder) BuildCount() (string, []interface{}) {
	qb.queryType = Count
	return qb.Build()
}

// BuildInsert constructs an insert query
func (qb *QueryBuilder) BuildInsert(data map[string]interface{}) (string, []interface{}) {
	qb.queryType = Insert
	qb.values = data
	return qb.Build()
}

// BuildUpdate constructs an update query
func (qb *QueryBuilder) BuildUpdate(data map[string]interface{}) (string, []interface{}) {
	qb.queryType = Update
	qb.values = data
	return qb.Build()
}

// BuildDelete constructs a delete query
func (qb *QueryBuilder) BuildDelete() (string, []interface{}) {
	qb.queryType = Delete
	return qb.Build()
}
