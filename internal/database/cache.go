package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

// CacheRepository is a repository that caches database operations
type CacheRepository struct {
	repo  Repository
	cache *cache.Cache
	stats struct {
		hits   int64
		misses int64
		items  int64
	}
}

// NewCacheRepository creates a new cache repository
func NewCacheRepository(repo Repository, defaultExpiration, cleanupInterval time.Duration) *CacheRepository {
	return &CacheRepository{
		repo:  repo,
		cache: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Create creates a record and caches it
func (r *CacheRepository) Create(table string, fields []string, values []interface{}) (int64, error) {
	id, err := r.repo.Create(table, fields, values)
	if err != nil {
		return 0, err
	}

	// Cache the new record - create a map from fields and values
	cacheKey := fmt.Sprintf("%s:%d", table, id)
	data := make(map[string]interface{})
	for i, field := range fields {
		if i < len(values) {
			data[field] = values[i]
		}
	}
	if err := r.cacheRecord(cacheKey, data); err != nil {
		// Log cache error but don't fail the operation
		fmt.Printf("Cache error: %v\n", err)
	}

	return id, nil
}

// Update updates a record and refreshes its cache
func (r *CacheRepository) Update(table string, id interface{}, data interface{}) error {
	if err := r.repo.Update(table, id, data); err != nil {
		return err
	}

	// Update cache
	cacheKey := fmt.Sprintf("%s:%v", table, id)
	if err := r.cacheRecord(cacheKey, data); err != nil {
		// Log cache error but don't fail the operation
		fmt.Printf("Cache error: %v\n", err)
	}

	return nil
}

// Delete deletes a record and removes it from cache
func (r *CacheRepository) Delete(table string, id interface{}) error {
	if err := r.repo.Delete(table, id); err != nil {
		return err
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("%s:%v", table, id)
	r.cache.Delete(cacheKey)

	return nil
}

// FindByID finds a record by ID, using cache if available
func (r *CacheRepository) FindByID(table string, id interface{}, result interface{}) error {
	cacheKey := fmt.Sprintf("%s:%v", table, id)

	// Try to get from cache first
	if cached, found := r.cache.Get(cacheKey); found {
		r.stats.hits++
		return json.Unmarshal(cached.([]byte), result)
	}

	// If not in cache, get from database
	if err := r.repo.FindByID(table, id, result); err != nil {
		return err
	}

	// Cache the result
	if err := r.cacheRecord(cacheKey, result); err != nil {
		// Log cache error but don't fail the operation
		fmt.Printf("Cache error: %v\n", err)
	}

	return nil
}

// FindAll finds all records with pagination
func (r *CacheRepository) FindAll(table string, result interface{}, conditions map[string]interface{}, orderBy string, limit, offset int) error {
	// For list operations, we don't use cache as the data might be stale
	return r.repo.FindAll(table, result, conditions, orderBy, limit, offset)
}

// FindOne finds a single record
func (r *CacheRepository) FindOne(table string, result interface{}, conditions map[string]interface{}) error {
	// For find operations with conditions, we don't use cache
	return r.repo.FindOne(table, result, conditions)
}

// Count returns the number of records
func (r *CacheRepository) Count(table string, conditions map[string]interface{}) (int64, error) {
	// For count operations, we don't use cache
	return r.repo.Count(table, conditions)
}

// Search searches records
func (r *CacheRepository) Search(table string, fields []string, term string, limit, offset int, result interface{}) error {
	// For search operations, we don't use cache
	return r.repo.Search(table, fields, term, limit, offset, result)
}

// FindByDateRange finds records within a date range
func (r *CacheRepository) FindByDateRange(table, dateField string, start, end time.Time, limit, offset int, result interface{}) error {
	// For date range queries, we don't use cache
	return r.repo.FindByDateRange(table, dateField, start, end, limit, offset, result)
}

// Transaction executes a function within a transaction
func (r *CacheRepository) Transaction(fn func(*sql.Tx) error) error {
	return r.repo.Transaction(fn)
}

// Helper functions

// cacheRecord caches a record
func (r *CacheRepository) cacheRecord(key string, data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling data for cache: %v", err)
	}

	r.cache.Set(key, bytes, cache.DefaultExpiration)
	r.stats.items++
	return nil
}

// ClearCache clears the entire cache
func (r *CacheRepository) ClearCache() {
	r.cache.Flush()
}

// DeleteCache deletes a specific cache entry
func (r *CacheRepository) DeleteCache(key string) {
	r.cache.Delete(key)
}

// GetStats returns cache statistics
func (r *CacheRepository) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"items":  r.stats.items,
		"hits":   r.stats.hits,
		"misses": r.stats.misses,
	}
}

// Get retrieves a value from cache or database
func (r *CacheRepository) Get(table string, id interface{}, result interface{}) error {
	cacheKey := fmt.Sprintf("%s:%v", table, id)
	if cached, found := r.cache.Get(cacheKey); found {
		r.stats.hits++
		return json.Unmarshal(cached.([]byte), result)
	}

	r.stats.misses++
	if err := r.repo.FindByID(table, id, result); err != nil {
		return err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	r.cache.Set(cacheKey, data, cache.DefaultExpiration)
	r.stats.items++
	return nil
}

// Set stores a value in cache and database
func (r *CacheRepository) Set(table string, id interface{}, value interface{}) (int64, error) {
	// This method needs to be refactored to use proper fields and values
	// For now, we'll return an error to indicate it needs implementation
	return 0, fmt.Errorf("Set method needs to be properly implemented with field/value mapping")

	data, err := json.Marshal(value)
	if err != nil {
		return 0, err
	}

	cacheKey := fmt.Sprintf("%s:%v", table, id)
	r.cache.Set(cacheKey, data, cache.DefaultExpiration)
	r.stats.items++
	return id.(int64), nil
}

// Flush clears the cache
func (r *CacheRepository) Flush() {
	r.cache.Flush()
	r.stats.items = 0
}
