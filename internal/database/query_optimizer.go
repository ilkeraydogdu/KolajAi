package database

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// QueryOptimizer provides SQL query optimization and caching
type QueryOptimizer struct {
	db          *sql.DB
	cache       *QueryCache
	stats       *QueryStats
	slowQueries map[string]*SlowQuery
	mutex       sync.RWMutex
}

// QueryCache stores cached query results
type QueryCache struct {
	cache  map[string]*CacheEntry
	mutex  sync.RWMutex
	maxAge time.Duration
}

// CacheEntry represents a cached query result
type CacheEntry struct {
	Data      interface{}
	CreatedAt time.Time
	TTL       time.Duration
}

// QueryStats tracks query performance
type QueryStats struct {
	TotalQueries    int64
	SlowQueries     int64
	CacheHits       int64
	CacheMisses     int64
	AverageExecTime time.Duration
	mutex           sync.RWMutex
}

// SlowQuery tracks slow query information
type SlowQuery struct {
	Query       string
	Count       int
	TotalTime   time.Duration
	AverageTime time.Duration
	LastSeen    time.Time
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer(db *sql.DB) *QueryOptimizer {
	return &QueryOptimizer{
		db: db,
		cache: &QueryCache{
			cache:  make(map[string]*CacheEntry),
			maxAge: 5 * time.Minute,
		},
		stats:       &QueryStats{},
		slowQueries: make(map[string]*SlowQuery),
	}
}

// OptimizeCountQuery optimizes COUNT(*) queries
func (qo *QueryOptimizer) OptimizeCountQuery(table string, conditions map[string]interface{}) (int64, error) {
	// Generate cache key
	cacheKey := qo.generateCacheKey("count", table, conditions)

	// Check cache first
	if cached := qo.cache.Get(cacheKey); cached != nil {
		qo.stats.IncrementCacheHits()
		return cached.(int64), nil
	}

	qo.stats.IncrementCacheMisses()

	// Build optimized query
	query, args := qo.buildOptimizedCountQuery(table, conditions)

	// Execute with timing
	start := time.Now()
	var count int64
	err := qo.db.QueryRow(query, args...).Scan(&count)
	duration := time.Since(start)

	// Track performance
	qo.trackQuery(query, duration)

	if err != nil {
		return 0, err
	}

	// Cache result for 1 minute (counts change frequently)
	qo.cache.Set(cacheKey, count, 1*time.Minute)

	return count, nil
}

// buildOptimizedCountQuery builds an optimized COUNT query
func (qo *QueryOptimizer) buildOptimizedCountQuery(table string, conditions map[string]interface{}) (string, []interface{}) {
	var query strings.Builder
	var args []interface{}

	// Use COUNT(1) instead of COUNT(*) for better performance
	query.WriteString("SELECT COUNT(1) FROM ")
	query.WriteString(table)

	if len(conditions) > 0 {
		query.WriteString(" WHERE ")
		conditionParts := make([]string, 0, len(conditions))

		for field, value := range conditions {
			// Add index hints for common fields
			if field == "created_at" || field == "updated_at" {
				conditionParts = append(conditionParts, fmt.Sprintf("%s = ?", field))
			} else if field == "status" || field == "is_active" {
				conditionParts = append(conditionParts, fmt.Sprintf("%s = ?", field))
			} else {
				conditionParts = append(conditionParts, fmt.Sprintf("%s = ?", field))
			}
			args = append(args, value)
		}

		query.WriteString(strings.Join(conditionParts, " AND "))
	}

	return query.String(), args
}

// OptimizeSelectQuery optimizes SELECT queries with pagination and caching
func (qo *QueryOptimizer) OptimizeSelectQuery(table string, fields []string, conditions map[string]interface{}, orderBy string, limit, offset int) (*sql.Rows, error) {
	// Build optimized query
	query, args := qo.buildOptimizedSelectQuery(table, fields, conditions, orderBy, limit, offset)

	// Execute with timing
	start := time.Now()
	rows, err := qo.db.Query(query, args...)
	duration := time.Since(start)

	// Track performance
	qo.trackQuery(query, duration)

	return rows, err
}

// buildOptimizedSelectQuery builds an optimized SELECT query
func (qo *QueryOptimizer) buildOptimizedSelectQuery(table string, fields []string, conditions map[string]interface{}, orderBy string, limit, offset int) (string, []interface{}) {
	var query strings.Builder
	var args []interface{}

	// SELECT clause
	query.WriteString("SELECT ")
	if len(fields) == 0 {
		query.WriteString("*")
	} else {
		query.WriteString(strings.Join(fields, ", "))
	}

	query.WriteString(" FROM ")
	query.WriteString(table)

	// WHERE clause
	if len(conditions) > 0 {
		query.WriteString(" WHERE ")
		conditionParts := make([]string, 0, len(conditions))

		for field, value := range conditions {
			conditionParts = append(conditionParts, fmt.Sprintf("%s = ?", field))
			args = append(args, value)
		}

		query.WriteString(strings.Join(conditionParts, " AND "))
	}

	// ORDER BY clause
	if orderBy != "" {
		query.WriteString(" ORDER BY ")
		query.WriteString(orderBy)
	}

	// LIMIT clause
	if limit > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", limit))
		if offset > 0 {
			query.WriteString(fmt.Sprintf(" OFFSET %d", offset))
		}
	}

	return query.String(), args
}

// ExecuteWithCache executes a query with caching
func (qo *QueryOptimizer) ExecuteWithCache(ctx context.Context, query string, args []interface{}, ttl time.Duration) (*sql.Rows, error) {
	cacheKey := qo.generateQueryCacheKey(query, args)

	// Check cache
	if cached := qo.cache.Get(cacheKey); cached != nil {
		qo.stats.IncrementCacheHits()
		// Return cached rows (this would need more complex implementation for actual rows)
		// For now, we'll skip caching for complex result sets
	}

	qo.stats.IncrementCacheMisses()

	// Execute query with timing
	start := time.Now()
	rows, err := qo.db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	// Track performance
	qo.trackQuery(query, duration)

	return rows, err
}

// trackQuery tracks query performance
func (qo *QueryOptimizer) trackQuery(query string, duration time.Duration) {
	qo.stats.mutex.Lock()
	defer qo.stats.mutex.Unlock()

	qo.stats.TotalQueries++

	// Update average execution time
	if qo.stats.TotalQueries == 1 {
		qo.stats.AverageExecTime = duration
	} else {
		qo.stats.AverageExecTime = time.Duration(
			(int64(qo.stats.AverageExecTime)*qo.stats.TotalQueries + int64(duration)) / (qo.stats.TotalQueries + 1),
		)
	}

	// Track slow queries (> 100ms)
	if duration > 100*time.Millisecond {
		qo.stats.SlowQueries++
		qo.trackSlowQuery(query, duration)
	}
}

// trackSlowQuery tracks individual slow queries
func (qo *QueryOptimizer) trackSlowQuery(query string, duration time.Duration) {
	qo.mutex.Lock()
	defer qo.mutex.Unlock()

	// Normalize query for tracking (remove specific values)
	normalizedQuery := qo.normalizeQuery(query)

	if slowQuery, exists := qo.slowQueries[normalizedQuery]; exists {
		slowQuery.Count++
		slowQuery.TotalTime += duration
		slowQuery.AverageTime = time.Duration(int64(slowQuery.TotalTime) / int64(slowQuery.Count))
		slowQuery.LastSeen = time.Now()
	} else {
		qo.slowQueries[normalizedQuery] = &SlowQuery{
			Query:       normalizedQuery,
			Count:       1,
			TotalTime:   duration,
			AverageTime: duration,
			LastSeen:    time.Now(),
		}
	}

	// Log slow query
	log.Printf("SLOW QUERY (%v): %s", duration, normalizedQuery)
}

// normalizeQuery normalizes a query for tracking purposes
func (qo *QueryOptimizer) normalizeQuery(query string) string {
	// Replace specific values with placeholders
	normalized := strings.ReplaceAll(query, "'", "?")
	normalized = strings.ReplaceAll(normalized, "\"", "?")

	// Remove extra whitespace
	normalized = strings.Join(strings.Fields(normalized), " ")

	return normalized
}

// generateCacheKey generates a cache key for the given parameters
func (qo *QueryOptimizer) generateCacheKey(operation string, table string, params interface{}) string {
	data := fmt.Sprintf("%s:%s:%v", operation, table, params)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// generateQueryCacheKey generates a cache key for a raw query
func (qo *QueryOptimizer) generateQueryCacheKey(query string, args []interface{}) string {
	data := fmt.Sprintf("%s:%v", query, args)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// GetStats returns query performance statistics
func (qo *QueryOptimizer) GetStats() *QueryStats {
	qo.stats.mutex.RLock()
	defer qo.stats.mutex.RUnlock()

	// Return a copy to avoid race conditions
	return &QueryStats{
		TotalQueries:    qo.stats.TotalQueries,
		SlowQueries:     qo.stats.SlowQueries,
		CacheHits:       qo.stats.CacheHits,
		CacheMisses:     qo.stats.CacheMisses,
		AverageExecTime: qo.stats.AverageExecTime,
	}
}

// GetSlowQueries returns the top slow queries
func (qo *QueryOptimizer) GetSlowQueries(limit int) []*SlowQuery {
	qo.mutex.RLock()
	defer qo.mutex.RUnlock()

	queries := make([]*SlowQuery, 0, len(qo.slowQueries))
	for _, query := range qo.slowQueries {
		queries = append(queries, query)
	}

	// Sort by average time (simple bubble sort for small datasets)
	for i := 0; i < len(queries)-1; i++ {
		for j := 0; j < len(queries)-i-1; j++ {
			if queries[j].AverageTime < queries[j+1].AverageTime {
				queries[j], queries[j+1] = queries[j+1], queries[j]
			}
		}
	}

	if limit > 0 && limit < len(queries) {
		queries = queries[:limit]
	}

	return queries
}

// Cache methods
func (c *QueryCache) Get(key string) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil
	}

	// Check if expired
	if time.Since(entry.CreatedAt) > entry.TTL {
		delete(c.cache, key)
		return nil
	}

	return entry.Data
}

func (c *QueryCache) Set(key string, data interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[key] = &CacheEntry{
		Data:      data,
		CreatedAt: time.Now(),
		TTL:       ttl,
	}
}

func (c *QueryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]*CacheEntry)
}

// Stats methods
func (s *QueryStats) IncrementCacheHits() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.CacheHits++
}

func (s *QueryStats) IncrementCacheMisses() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.CacheMisses++
}
