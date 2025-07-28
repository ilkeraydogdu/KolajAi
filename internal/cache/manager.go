package cache

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// CacheManager handles comprehensive caching
type CacheManager struct {
	stores    map[string]CacheStore
	config    CacheConfig
	stats     *CacheStats
	mu        sync.RWMutex
	db        *sql.DB
	metrics   *MetricsCollector
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	DefaultTTL       time.Duration            `json:"default_ttl"`
	MaxMemoryUsage   int64                    `json:"max_memory_usage"`
	EvictionPolicy   EvictionPolicy           `json:"eviction_policy"`
	CompressionEnabled bool                   `json:"compression_enabled"`
	EncryptionEnabled bool                    `json:"encryption_enabled"`
	EncryptionKey    string                   `json:"encryption_key"`
	Stores           map[string]StoreConfig   `json:"stores"`
	Clusters         []ClusterConfig          `json:"clusters"`
	Replication      ReplicationConfig        `json:"replication"`
	Monitoring       MonitoringConfig         `json:"monitoring"`
	Persistence      PersistenceConfig        `json:"persistence"`
}

// StoreConfig holds individual store configuration
type StoreConfig struct {
	Type           StoreType     `json:"type"`
	MaxSize        int64         `json:"max_size"`
	TTL            time.Duration `json:"ttl"`
	EvictionPolicy EvictionPolicy `json:"eviction_policy"`
	Shards         int           `json:"shards"`
	Enabled        bool          `json:"enabled"`
	Settings       map[string]interface{} `json:"settings"`
}

// ClusterConfig holds cluster configuration
type ClusterConfig struct {
	Name     string   `json:"name"`
	Nodes    []string `json:"nodes"`
	Primary  string   `json:"primary"`
	Replicas []string `json:"replicas"`
	Enabled  bool     `json:"enabled"`
}

// ReplicationConfig holds replication configuration
type ReplicationConfig struct {
	Enabled       bool          `json:"enabled"`
	Factor        int           `json:"factor"`
	SyncMode      string        `json:"sync_mode"` // "sync", "async"
	HealthCheck   time.Duration `json:"health_check"`
	FailoverTime  time.Duration `json:"failover_time"`
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	Enabled         bool          `json:"enabled"`
	MetricsInterval time.Duration `json:"metrics_interval"`
	AlertThresholds AlertThresholds `json:"alert_thresholds"`
	LogLevel        string        `json:"log_level"`
}

// AlertThresholds holds alert threshold configuration
type AlertThresholds struct {
	MemoryUsage    float64 `json:"memory_usage"`
	HitRatio       float64 `json:"hit_ratio"`
	ResponseTime   time.Duration `json:"response_time"`
	ErrorRate      float64 `json:"error_rate"`
}

// PersistenceConfig holds persistence configuration
type PersistenceConfig struct {
	Enabled       bool          `json:"enabled"`
	BackupInterval time.Duration `json:"backup_interval"`
	BackupPath    string        `json:"backup_path"`
	Compression   bool          `json:"compression"`
	Encryption    bool          `json:"encryption"`
}

// StoreType represents different cache store types
type StoreType string

const (
	StoreTypeMemory     StoreType = "memory"
	StoreTypeRedis      StoreType = "redis"
	StoreTypeMemcached  StoreType = "memcached"
	StoreTypeDatabase   StoreType = "database"
	StoreTypeFile       StoreType = "file"
	StoreTypeDistributed StoreType = "distributed"
)

// EvictionPolicy represents cache eviction policies
type EvictionPolicy string

const (
	EvictionLRU    EvictionPolicy = "lru"
	EvictionLFU    EvictionPolicy = "lfu"
	EvictionFIFO   EvictionPolicy = "fifo"
	EvictionRandom EvictionPolicy = "random"
	EvictionTTL    EvictionPolicy = "ttl"
)

// CacheStore interface for different cache implementations
type CacheStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
	Keys(ctx context.Context, pattern string) ([]string, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Size(ctx context.Context) (int64, error)
	Stats(ctx context.Context) (*StoreStats, error)
	Close() error
}

// CacheItem represents a cached item
type CacheItem struct {
	Key        string                 `json:"key"`
	Value      []byte                 `json:"value"`
	TTL        time.Duration          `json:"ttl"`
	CreatedAt  time.Time              `json:"created_at"`
	AccessedAt time.Time              `json:"accessed_at"`
	AccessCount int64                 `json:"access_count"`
	Size       int64                  `json:"size"`
	Compressed bool                   `json:"compressed"`
	Encrypted  bool                   `json:"encrypted"`
	Tags       []string               `json:"tags"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	TotalHits        int64                 `json:"total_hits"`
	TotalMisses      int64                 `json:"total_misses"`
	TotalSets        int64                 `json:"total_sets"`
	TotalDeletes     int64                 `json:"total_deletes"`
	TotalEvictions   int64                 `json:"total_evictions"`
	TotalMemoryUsage int64                 `json:"total_memory_usage"`
	HitRatio         float64               `json:"hit_ratio"`
	StoreStats       map[string]*StoreStats `json:"store_stats"`
	LastReset        time.Time             `json:"last_reset"`
	mu               sync.RWMutex
}

// StoreStats represents individual store statistics
type StoreStats struct {
	Hits          int64         `json:"hits"`
	Misses        int64         `json:"misses"`
	Sets          int64         `json:"sets"`
	Deletes       int64         `json:"deletes"`
	Evictions     int64         `json:"evictions"`
	MemoryUsage   int64         `json:"memory_usage"`
	ItemCount     int64         `json:"item_count"`
	HitRatio      float64       `json:"hit_ratio"`
	AvgSetTime    time.Duration `json:"avg_set_time"`
	AvgGetTime    time.Duration `json:"avg_get_time"`
	LastAccess    time.Time     `json:"last_access"`
}

// CacheOperation represents cache operation types
type CacheOperation string

const (
	OperationGet    CacheOperation = "get"
	OperationSet    CacheOperation = "set"
	OperationDelete CacheOperation = "delete"
	OperationClear  CacheOperation = "clear"
)

// CacheEvent represents a cache event
type CacheEvent struct {
	ID         string         `json:"id"`
	Operation  CacheOperation `json:"operation"`
	Store      string         `json:"store"`
	Key        string         `json:"key"`
	Size       int64          `json:"size"`
	TTL        time.Duration  `json:"ttl"`
	Success    bool           `json:"success"`
	Duration   time.Duration  `json:"duration"`
	Error      string         `json:"error,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`
	ClientInfo ClientInfo     `json:"client_info"`
}

// ClientInfo represents client information
type ClientInfo struct {
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
}

// MetricsCollector collects cache metrics
type MetricsCollector struct {
	enabled     bool
	interval    time.Duration
	events      chan CacheEvent
	aggregates  map[string]*MetricAggregate
	mu          sync.RWMutex
}

// MetricAggregate represents aggregated metrics
type MetricAggregate struct {
	Count       int64         `json:"count"`
	TotalTime   time.Duration `json:"total_time"`
	MinTime     time.Duration `json:"min_time"`
	MaxTime     time.Duration `json:"max_time"`
	AvgTime     time.Duration `json:"avg_time"`
	Errors      int64         `json:"errors"`
	LastUpdated time.Time     `json:"last_updated"`
}

// CacheKey represents a structured cache key
type CacheKey struct {
	Prefix    string            `json:"prefix"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"`
	ID        string            `json:"id"`
	Version   string            `json:"version"`
	Tags      []string          `json:"tags"`
	Params    map[string]string `json:"params"`
}

// NewCacheManager creates a new cache manager
func NewCacheManager(db *sql.DB, config CacheConfig) *CacheManager {
	cm := &CacheManager{
		stores:  make(map[string]CacheStore),
		config:  config,
		stats:   NewCacheStats(),
		db:      db,
		metrics: NewMetricsCollector(config.Monitoring),
	}

	cm.createCacheTables()
	cm.initializeStores()
	cm.startMonitoring()

	return cm
}

// NewCacheStats creates new cache statistics
func NewCacheStats() *CacheStats {
	return &CacheStats{
		StoreStats: make(map[string]*StoreStats),
		LastReset:  time.Now(),
	}
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(config MonitoringConfig) *MetricsCollector {
	mc := &MetricsCollector{
		enabled:    config.Enabled,
		interval:   config.MetricsInterval,
		events:     make(chan CacheEvent, 1000),
		aggregates: make(map[string]*MetricAggregate),
	}

	if mc.enabled {
		go mc.processEvents()
	}

	return mc
}

// createCacheTables creates necessary tables for cache management
func (cm *CacheManager) createCacheTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS cache_items (
			cache_key VARCHAR(512) PRIMARY KEY,
			store_name VARCHAR(100) NOT NULL,
			value LONGBLOB,
			ttl_seconds INT DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			access_count BIGINT DEFAULT 0,
			size_bytes BIGINT DEFAULT 0,
			compressed BOOLEAN DEFAULT FALSE,
			encrypted BOOLEAN DEFAULT FALSE,
			tags TEXT,
			metadata TEXT,
			expires_at DATETIME,
			INDEX idx_store_name (store_name),
			INDEX idx_expires_at (expires_at),
			INDEX idx_accessed_at (accessed_at)
		)`,
		`CREATE TABLE IF NOT EXISTS cache_stats (
			id INT AUTO_INCREMENT PRIMARY KEY,
			store_name VARCHAR(100) NOT NULL,
			operation VARCHAR(20) NOT NULL,
			hits BIGINT DEFAULT 0,
			misses BIGINT DEFAULT 0,
			sets BIGINT DEFAULT 0,
			deletes BIGINT DEFAULT 0,
			evictions BIGINT DEFAULT 0,
			memory_usage BIGINT DEFAULT 0,
			avg_response_time_ms INT DEFAULT 0,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_store_name (store_name),
			INDEX idx_timestamp (timestamp)
		)`,
		`CREATE TABLE IF NOT EXISTS cache_events (
			id VARCHAR(128) PRIMARY KEY,
			operation VARCHAR(20) NOT NULL,
			store_name VARCHAR(100) NOT NULL,
			cache_key VARCHAR(512),
			size_bytes BIGINT DEFAULT 0,
			ttl_seconds INT DEFAULT 0,
			success BOOLEAN DEFAULT TRUE,
			duration_ms INT DEFAULT 0,
			error_message TEXT,
			client_ip VARCHAR(45),
			user_agent TEXT,
			user_id VARCHAR(128),
			session_id VARCHAR(128),
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_operation (operation),
			INDEX idx_store_name (store_name),
			INDEX idx_timestamp (timestamp),
			INDEX idx_success (success)
		)`,
	}

	for _, query := range queries {
		if _, err := cm.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create cache table: %w", err)
		}
	}

	return nil
}

// Get retrieves a value from cache
func (cm *CacheManager) Get(ctx context.Context, storeName, key string) ([]byte, error) {
	start := time.Now()
	
	store, exists := cm.getStore(storeName)
	if !exists {
		return nil, fmt.Errorf("store %s not found", storeName)
	}

	value, err := store.Get(ctx, key)
	duration := time.Since(start)

	// Record metrics
	event := CacheEvent{
		ID:        cm.generateEventID(),
		Operation: OperationGet,
		Store:     storeName,
		Key:       key,
		Success:   err == nil,
		Duration:  duration,
		Timestamp: time.Now(),
	}
	
	if err != nil {
		event.Error = err.Error()
		cm.stats.RecordMiss(storeName)
	} else {
		event.Size = int64(len(value))
		cm.stats.RecordHit(storeName)
	}

	cm.recordEvent(event)
	
	return value, err
}

// Set stores a value in cache
func (cm *CacheManager) Set(ctx context.Context, storeName, key string, value []byte, ttl time.Duration) error {
	start := time.Now()
	
	store, exists := cm.getStore(storeName)
	if !exists {
		return fmt.Errorf("store %s not found", storeName)
	}

	// Apply compression if enabled
	if cm.config.CompressionEnabled {
		compressed, err := cm.compress(value)
		if err == nil && len(compressed) < len(value) {
			value = compressed
		}
	}

	// Apply encryption if enabled
	if cm.config.EncryptionEnabled {
		encrypted, err := cm.encrypt(value)
		if err == nil {
			value = encrypted
		}
	}

	err := store.Set(ctx, key, value, ttl)
	duration := time.Since(start)

	// Record metrics
	event := CacheEvent{
		ID:        cm.generateEventID(),
		Operation: OperationSet,
		Store:     storeName,
		Key:       key,
		Size:      int64(len(value)),
		TTL:       ttl,
		Success:   err == nil,
		Duration:  duration,
		Timestamp: time.Now(),
	}
	
	if err != nil {
		event.Error = err.Error()
	} else {
		cm.stats.RecordSet(storeName)
	}

	cm.recordEvent(event)
	
	return err
}

// Delete removes a value from cache
func (cm *CacheManager) Delete(ctx context.Context, storeName, key string) error {
	start := time.Now()
	
	store, exists := cm.getStore(storeName)
	if !exists {
		return fmt.Errorf("store %s not found", storeName)
	}

	err := store.Delete(ctx, key)
	duration := time.Since(start)

	// Record metrics
	event := CacheEvent{
		ID:        cm.generateEventID(),
		Operation: OperationDelete,
		Store:     storeName,
		Key:       key,
		Success:   err == nil,
		Duration:  duration,
		Timestamp: time.Now(),
	}
	
	if err != nil {
		event.Error = err.Error()
	} else {
		cm.stats.RecordDelete(storeName)
	}

	cm.recordEvent(event)
	
	return err
}

// GetOrSet retrieves a value or sets it if not found
func (cm *CacheManager) GetOrSet(ctx context.Context, storeName, key string, ttl time.Duration, generator func() ([]byte, error)) ([]byte, error) {
	// Try to get from cache first
	value, err := cm.Get(ctx, storeName, key)
	if err == nil {
		return value, nil
	}

	// Generate new value
	value, err = generator()
	if err != nil {
		return nil, err
	}

	// Set in cache
	if setErr := cm.Set(ctx, storeName, key, value, ttl); setErr != nil {
		// Log error but return the generated value
		cm.logError(fmt.Sprintf("Failed to set cache key %s: %v", key, setErr))
	}

	return value, nil
}

// GetMulti retrieves multiple values from cache
func (cm *CacheManager) GetMulti(ctx context.Context, storeName string, keys []string) (map[string][]byte, error) {
	results := make(map[string][]byte)
	
	for _, key := range keys {
		if value, err := cm.Get(ctx, storeName, key); err == nil {
			results[key] = value
		}
	}
	
	return results, nil
}

// SetMulti stores multiple values in cache
func (cm *CacheManager) SetMulti(ctx context.Context, storeName string, items map[string][]byte, ttl time.Duration) error {
	for key, value := range items {
		if err := cm.Set(ctx, storeName, key, value, ttl); err != nil {
			return err
		}
	}
	
	return nil
}

// DeleteMulti removes multiple values from cache
func (cm *CacheManager) DeleteMulti(ctx context.Context, storeName string, keys []string) error {
	for _, key := range keys {
		if err := cm.Delete(ctx, storeName, key); err != nil {
			return err
		}
	}
	
	return nil
}

// InvalidateByTags invalidates cache items by tags
func (cm *CacheManager) InvalidateByTags(ctx context.Context, storeName string, tags []string) error {
	_, exists := cm.getStore(storeName)
	if !exists {
		return fmt.Errorf("store %s not found", storeName)
	}

	// Get all keys matching tags (implementation depends on store type)
	keys, err := cm.getKeysByTags(ctx, storeName, tags)
	if err != nil {
		return err
	}

	// Delete matching keys
	return cm.DeleteMulti(ctx, storeName, keys)
}

// Flush clears all items from a store
func (cm *CacheManager) Flush(ctx context.Context, storeName string) error {
	start := time.Now()
	
	store, exists := cm.getStore(storeName)
	if !exists {
		return fmt.Errorf("store %s not found", storeName)
	}

	err := store.Clear(ctx)
	duration := time.Since(start)

	// Record metrics
	event := CacheEvent{
		ID:        cm.generateEventID(),
		Operation: OperationClear,
		Store:     storeName,
		Success:   err == nil,
		Duration:  duration,
		Timestamp: time.Now(),
	}
	
	if err != nil {
		event.Error = err.Error()
	}

	cm.recordEvent(event)
	
	return err
}

// GetStats returns cache statistics
func (cm *CacheManager) GetStats() *CacheStats {
	cm.stats.mu.RLock()
	defer cm.stats.mu.RUnlock()
	
	// Calculate hit ratio
	totalRequests := cm.stats.TotalHits + cm.stats.TotalMisses
	if totalRequests > 0 {
		cm.stats.HitRatio = float64(cm.stats.TotalHits) / float64(totalRequests)
	}
	
	return cm.stats
}

// GetStoreStats returns statistics for a specific store
func (cm *CacheManager) GetStoreStats(storeName string) (*StoreStats, error) {
	store, exists := cm.getStore(storeName)
	if !exists {
		return nil, fmt.Errorf("store %s not found", storeName)
	}

	return store.Stats(context.Background())
}

// BuildKey builds a structured cache key
func (cm *CacheManager) BuildKey(keyData CacheKey) string {
	parts := make([]string, 0)
	
	if keyData.Prefix != "" {
		parts = append(parts, keyData.Prefix)
	}
	if keyData.Namespace != "" {
		parts = append(parts, keyData.Namespace)
	}
	if keyData.Type != "" {
		parts = append(parts, keyData.Type)
	}
	if keyData.ID != "" {
		parts = append(parts, keyData.ID)
	}
	if keyData.Version != "" {
		parts = append(parts, "v"+keyData.Version)
	}
	
	// Add sorted parameters
	if len(keyData.Params) > 0 {
		paramKeys := make([]string, 0, len(keyData.Params))
		for k := range keyData.Params {
			paramKeys = append(paramKeys, k)
		}
		sort.Strings(paramKeys)
		
		for _, k := range paramKeys {
			parts = append(parts, k+"="+keyData.Params[k])
		}
	}
	
	key := strings.Join(parts, ":")
	
	// Add hash if key is too long
	if len(key) > 250 {
		hash := sha256.Sum256([]byte(key))
		return fmt.Sprintf("hash:%x", hash)
	}
	
	return key
}

// HashKey creates a hash of the key for consistent distribution
func (cm *CacheManager) HashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", hash)
}

// GetHealthStatus returns health status of all stores
func (cm *CacheManager) GetHealthStatus() map[string]bool {
	status := make(map[string]bool)
	
	for name, store := range cm.stores {
		// Simple health check by trying to get stats
		_, err := store.Stats(context.Background())
		status[name] = err == nil
	}
	
	return status
}

// Helper methods

func (cm *CacheManager) getStore(name string) (CacheStore, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	store, exists := cm.stores[name]
	return store, exists
}

func (cm *CacheManager) recordEvent(event CacheEvent) {
	if cm.metrics != nil && cm.metrics.enabled {
		select {
		case cm.metrics.events <- event:
		default:
			// Channel full, skip event
		}
	}
}

func (cm *CacheManager) generateEventID() string {
	return fmt.Sprintf("cache_event_%d", time.Now().UnixNano())
}

func (cm *CacheManager) compress(data []byte) ([]byte, error) {
	// Implementation would use compression library (gzip, lz4, etc.)
	return data, nil
}

func (cm *CacheManager) encrypt(data []byte) ([]byte, error) {
	// Implementation would use encryption library (AES, etc.)
	return data, nil
}

func (cm *CacheManager) getKeysByTags(ctx context.Context, storeName string, tags []string) ([]string, error) {
	// Implementation would query database for keys with matching tags
	return []string{}, nil
}

func (cm *CacheManager) logError(message string) {
	// Implementation would log errors
	fmt.Printf("Cache Error: %s\n", message)
}

func (cm *CacheManager) initializeStores() {
	// Implementation would initialize configured stores
}

func (cm *CacheManager) startMonitoring() {
	// Implementation would start monitoring routines
}

// CacheStats methods

func (cs *CacheStats) RecordHit(storeName string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	cs.TotalHits++
	
	if cs.StoreStats[storeName] == nil {
		cs.StoreStats[storeName] = &StoreStats{}
	}
	cs.StoreStats[storeName].Hits++
	cs.StoreStats[storeName].LastAccess = time.Now()
}

func (cs *CacheStats) RecordMiss(storeName string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	cs.TotalMisses++
	
	if cs.StoreStats[storeName] == nil {
		cs.StoreStats[storeName] = &StoreStats{}
	}
	cs.StoreStats[storeName].Misses++
	cs.StoreStats[storeName].LastAccess = time.Now()
}

func (cs *CacheStats) RecordSet(storeName string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	cs.TotalSets++
	
	if cs.StoreStats[storeName] == nil {
		cs.StoreStats[storeName] = &StoreStats{}
	}
	cs.StoreStats[storeName].Sets++
}

func (cs *CacheStats) RecordDelete(storeName string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	cs.TotalDeletes++
	
	if cs.StoreStats[storeName] == nil {
		cs.StoreStats[storeName] = &StoreStats{}
	}
	cs.StoreStats[storeName].Deletes++
}

func (cs *CacheStats) Reset() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	cs.TotalHits = 0
	cs.TotalMisses = 0
	cs.TotalSets = 0
	cs.TotalDeletes = 0
	cs.TotalEvictions = 0
	cs.TotalMemoryUsage = 0
	cs.HitRatio = 0
	cs.StoreStats = make(map[string]*StoreStats)
	cs.LastReset = time.Now()
}

// MetricsCollector methods

func (mc *MetricsCollector) processEvents() {
	ticker := time.NewTicker(mc.interval)
	defer ticker.Stop()

	for {
		select {
		case event := <-mc.events:
			mc.processEvent(event)
		case <-ticker.C:
			mc.aggregateMetrics()
		}
	}
}

func (mc *MetricsCollector) processEvent(event CacheEvent) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	key := fmt.Sprintf("%s:%s", event.Store, event.Operation)
	
	if mc.aggregates[key] == nil {
		mc.aggregates[key] = &MetricAggregate{
			MinTime: event.Duration,
			MaxTime: event.Duration,
		}
	}

	agg := mc.aggregates[key]
	agg.Count++
	agg.TotalTime += event.Duration
	agg.AvgTime = time.Duration(int64(agg.TotalTime) / agg.Count)
	
	if event.Duration < agg.MinTime {
		agg.MinTime = event.Duration
	}
	if event.Duration > agg.MaxTime {
		agg.MaxTime = event.Duration
	}
	
	if !event.Success {
		agg.Errors++
	}
	
	agg.LastUpdated = time.Now()
}

func (mc *MetricsCollector) aggregateMetrics() {
	// Implementation would aggregate and persist metrics
}

// Close closes the cache manager and all stores
func (cm *CacheManager) Close() error {
	for _, store := range cm.stores {
		if err := store.Close(); err != nil {
			cm.logError(fmt.Sprintf("Failed to close store: %v", err))
		}
	}
	return nil
}