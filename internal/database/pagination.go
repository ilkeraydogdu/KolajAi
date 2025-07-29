package database

import (
	"database/sql"
	"fmt"
	"math"
)

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	OrderBy  string `json:"order_by"`
	Order    string `json:"order"`
}

// PaginationResult holds pagination result metadata
type PaginationResult struct {
	Page         int `json:"page"`
	PageSize     int `json:"page_size"`
	TotalPages   int `json:"total_pages"`
	TotalRecords int `json:"total_records"`
	HasNext      bool `json:"has_next"`
	HasPrev      bool `json:"has_prev"`
}

// DefaultPaginationParams returns default pagination parameters
func DefaultPaginationParams() PaginationParams {
	return PaginationParams{
		Page:     1,
		PageSize: 20,
		OrderBy:  "id",
		Order:    "DESC",
	}
}

// Validate validates and normalizes pagination parameters
func (p *PaginationParams) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100 // Max page size to prevent abuse
	}
	if p.Order != "ASC" && p.Order != "DESC" {
		p.Order = "DESC"
	}
	if p.OrderBy == "" {
		p.OrderBy = "id"
	}
}

// GetOffset calculates the offset for SQL queries
func (p *PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit returns the limit for SQL queries
func (p *PaginationParams) GetLimit() int {
	return p.PageSize
}

// BuildPaginationQuery builds a paginated SQL query
func BuildPaginationQuery(baseQuery string, params PaginationParams, allowedOrderFields []string) string {
	params.Validate()
	
	// Validate order by field to prevent SQL injection
	orderByValid := false
	for _, field := range allowedOrderFields {
		if params.OrderBy == field {
			orderByValid = true
			break
		}
	}
	if !orderByValid && len(allowedOrderFields) > 0 {
		params.OrderBy = allowedOrderFields[0]
	}
	
	return fmt.Sprintf("%s ORDER BY %s %s LIMIT %d OFFSET %d",
		baseQuery,
		params.OrderBy,
		params.Order,
		params.GetLimit(),
		params.GetOffset(),
	)
}

// CountTotal counts total records for pagination
func CountTotal(db *sql.DB, countQuery string, args ...interface{}) (int, error) {
	var total int
	err := db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// CalculatePaginationResult calculates pagination metadata
func CalculatePaginationResult(params PaginationParams, totalRecords int) PaginationResult {
	params.Validate()
	
	totalPages := int(math.Ceil(float64(totalRecords) / float64(params.PageSize)))
	
	return PaginationResult{
		Page:         params.Page,
		PageSize:     params.PageSize,
		TotalPages:   totalPages,
		TotalRecords: totalRecords,
		HasNext:      params.Page < totalPages,
		HasPrev:      params.Page > 1,
	}
}

// PaginatedQuery executes a paginated query
type PaginatedQuery struct {
	DB               *sql.DB
	BaseQuery        string
	CountQuery       string
	Args             []interface{}
	AllowedOrderBy   []string
}

// Execute executes the paginated query
func (pq *PaginatedQuery) Execute(params PaginationParams, scanFunc func(*sql.Rows) (interface{}, error)) (*PaginatedResponse, error) {
	// Count total records
	total, err := CountTotal(pq.DB, pq.CountQuery, pq.Args...)
	if err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}
	
	// Build paginated query
	query := BuildPaginationQuery(pq.BaseQuery, params, pq.AllowedOrderBy)
	
	// Execute query
	rows, err := pq.DB.Query(query, pq.Args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
	
	// Scan results
	var results []interface{}
	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, item)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	
	// Calculate pagination result
	paginationResult := CalculatePaginationResult(params, total)
	
	return &PaginatedResponse{
		Data:       results,
		Pagination: paginationResult,
	}, nil
}

// PaginatedResponse holds paginated data and metadata
type PaginatedResponse struct {
	Data       []interface{}    `json:"data"`
	Pagination PaginationResult `json:"pagination"`
}

// CursorPaginationParams holds cursor-based pagination parameters
type CursorPaginationParams struct {
	Cursor   string `json:"cursor"`
	Limit    int    `json:"limit"`
	OrderBy  string `json:"order_by"`
	Order    string `json:"order"`
}

// CursorPaginationResult holds cursor pagination metadata
type CursorPaginationResult struct {
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Count      int    `json:"count"`
}

// BuildCursorQuery builds a cursor-based pagination query
func BuildCursorQuery(baseQuery string, params CursorPaginationParams, cursorField string) string {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	
	query := baseQuery
	if params.Cursor != "" {
		operator := ">"
		if params.Order == "DESC" {
			operator = "<"
		}
		query = fmt.Sprintf("%s WHERE %s %s '%s'", baseQuery, cursorField, operator, params.Cursor)
	}
	
	return fmt.Sprintf("%s ORDER BY %s %s LIMIT %d",
		query,
		params.OrderBy,
		params.Order,
		params.Limit+1, // Fetch one extra to determine if there are more
	)
}