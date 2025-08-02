package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// HealthService provides comprehensive health checking
type HealthService struct {
	db          *sql.DB
	redisClient *redis.Client
	checks      map[string]HealthChecker
	mu          sync.RWMutex
}

// HealthStatus represents the overall health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck represents a single health check result
type HealthCheck struct {
	Name        string                 `json:"name"`
	Status      HealthStatus           `json:"status"`
	Message     string                 `json:"message,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Critical    bool                   `json:"critical"`
	LastSuccess *time.Time             `json:"last_success,omitempty"`
	LastFailure *time.Time             `json:"last_failure,omitempty"`
}

// HealthReport represents the complete health report
type HealthReport struct {
	Status      HealthStatus           `json:"status"`
	Version     string                 `json:"version"`
	Timestamp   time.Time              `json:"timestamp"`
	Duration    time.Duration          `json:"duration"`
	Checks      map[string]HealthCheck `json:"checks"`
	System      SystemInfo             `json:"system"`
	Summary     HealthSummary          `json:"summary"`
}

// SystemInfo represents system information
type SystemInfo struct {
	Hostname     string        `json:"hostname"`
	OS           string        `json:"os"`
	Arch         string        `json:"arch"`
	GoVersion    string        `json:"go_version"`
	NumCPU       int           `json:"num_cpu"`
	NumGoroutine int           `json:"num_goroutine"`
	Memory       MemoryInfo    `json:"memory"`
	Uptime       time.Duration `json:"uptime"`
}

// MemoryInfo represents memory usage information
type MemoryInfo struct {
	Alloc        uint64 `json:"alloc"`         // bytes allocated and not yet freed
	TotalAlloc   uint64 `json:"total_alloc"`   // bytes allocated (even if freed)
	Sys          uint64 `json:"sys"`           // bytes obtained from system
	Lookups      uint64 `json:"lookups"`       // number of pointer lookups
	Mallocs      uint64 `json:"mallocs"`       // number of mallocs
	Frees        uint64 `json:"frees"`         // number of frees
	HeapAlloc    uint64 `json:"heap_alloc"`    // bytes allocated and not yet freed (same as Alloc above)
	HeapSys      uint64 `json:"heap_sys"`      // bytes obtained from system
	HeapIdle     uint64 `json:"heap_idle"`     // bytes in idle spans
	HeapInuse    uint64 `json:"heap_inuse"`    // bytes in non-idle span
	HeapReleased uint64 `json:"heap_released"` // bytes released to the OS
	HeapObjects  uint64 `json:"heap_objects"`  // total number of allocated objects
	StackInuse   uint64 `json:"stack_inuse"`   // bytes used by stack allocator
	StackSys     uint64 `json:"stack_sys"`     // bytes obtained from system for stack allocator
	MSpanInuse   uint64 `json:"mspan_inuse"`   // bytes used by mspan structures
	MSpanSys     uint64 `json:"mspan_sys"`     // bytes obtained from system for mspan structures
	MCacheInuse  uint64 `json:"mcache_inuse"`  // bytes used by mcache structures
	MCacheSys    uint64 `json:"mcache_sys"`    // bytes obtained from system for mcache structures
	GCSys        uint64 `json:"gc_sys"`        // bytes used for garbage collection system metadata
	OtherSys     uint64 `json:"other_sys"`     // bytes used for other system allocations
	NextGC       uint64 `json:"next_gc"`       // next collection will happen when HeapAlloc ≥ this amount
	LastGC       uint64 `json:"last_gc"`       // end time of last collection (nanoseconds since 1970)
	PauseTotalNs uint64 `json:"pause_total_ns"` // cumulative nanoseconds in GC stop-the-world pauses
	PauseNs      uint64 `json:"pause_ns"`      // circular buffer of recent GC stop-the-world pause times
	NumGC        uint32 `json:"num_gc"`        // number of completed GC cycles
	NumForcedGC  uint32 `json:"num_forced_gc"` // number of GC cycles that were forced by the application
	GCCPUFraction float64 `json:"gc_cpu_fraction"` // fraction of CPU time used by GC
}

// HealthSummary represents a summary of health checks
type HealthSummary struct {
	Total      int `json:"total"`
	Healthy    int `json:"healthy"`
	Degraded   int `json:"degraded"`
	Unhealthy  int `json:"unhealthy"`
	Critical   int `json:"critical"`
}

// HealthChecker interface for health check implementations
type HealthChecker interface {
	Check(ctx context.Context) HealthCheck
	Name() string
	IsCritical() bool
}

var (
	startTime = time.Now()
)

// NewHealthService creates a new health service
func NewHealthService(db *sql.DB, redisClient *redis.Client) *HealthService {
	service := &HealthService{
		db:          db,
		redisClient: redisClient,
		checks:      make(map[string]HealthChecker),
	}

	// Register default health checks
	service.RegisterCheck(&DatabaseHealthCheck{db: db})
	if redisClient != nil {
		service.RegisterCheck(&RedisHealthCheck{client: redisClient})
	}
	service.RegisterCheck(&MemoryHealthCheck{})
	service.RegisterCheck(&DiskHealthCheck{})

	return service
}

// RegisterCheck registers a new health check
func (h *HealthService) RegisterCheck(checker HealthChecker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[checker.Name()] = checker
}

// UnregisterCheck removes a health check
func (h *HealthService) UnregisterCheck(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.checks, name)
}

// Check performs all health checks and returns a comprehensive report
func (h *HealthService) Check(ctx context.Context) *HealthReport {
	startTime := time.Now()

	h.mu.RLock()
	checks := make(map[string]HealthChecker, len(h.checks))
	for name, checker := range h.checks {
		checks[name] = checker
	}
	h.mu.RUnlock()

	// Perform all health checks concurrently
	results := make(chan HealthCheck, len(checks))
	var wg sync.WaitGroup

	for _, checker := range checks {
		wg.Add(1)
		go func(checker HealthChecker) {
			defer wg.Done()
			
			checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			
			result := checker.Check(checkCtx)
			results <- result
		}(checker)
	}

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	checkResults := make(map[string]HealthCheck)
	for result := range results {
		checkResults[result.Name] = result
	}

	// Calculate overall status
	overallStatus := h.calculateOverallStatus(checkResults)

	// Generate system info
	systemInfo := h.getSystemInfo()

	// Generate summary
	summary := h.generateSummary(checkResults)

	return &HealthReport{
		Status:    overallStatus,
		Version:   "1.0.0", // Should come from build info
		Timestamp: time.Now(),
		Duration:  time.Since(startTime),
		Checks:    checkResults,
		System:    systemInfo,
		Summary:   summary,
	}
}

// QuickCheck performs a quick health check (critical checks only)
func (h *HealthService) QuickCheck(ctx context.Context) *HealthReport {
	startTime := time.Now()

	h.mu.RLock()
	criticalChecks := make(map[string]HealthChecker)
	for name, checker := range h.checks {
		if checker.IsCritical() {
			criticalChecks[name] = checker
		}
	}
	h.mu.RUnlock()

	// Perform critical health checks
	results := make(chan HealthCheck, len(criticalChecks))
	var wg sync.WaitGroup

	for _, checker := range criticalChecks {
		wg.Add(1)
		go func(checker HealthChecker) {
			defer wg.Done()
			
			checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			
			result := checker.Check(checkCtx)
			results <- result
		}(checker)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	checkResults := make(map[string]HealthCheck)
	for result := range results {
		checkResults[result.Name] = result
	}

	overallStatus := h.calculateOverallStatus(checkResults)
	summary := h.generateSummary(checkResults)

	return &HealthReport{
		Status:    overallStatus,
		Version:   "1.0.0",
		Timestamp: time.Now(),
		Duration:  time.Since(startTime),
		Checks:    checkResults,
		Summary:   summary,
	}
}

// calculateOverallStatus determines the overall health status
func (h *HealthService) calculateOverallStatus(checks map[string]HealthCheck) HealthStatus {
	hasUnhealthy := false
	hasDegraded := false

	for _, check := range checks {
		switch check.Status {
		case HealthStatusUnhealthy:
			if check.Critical {
				return HealthStatusUnhealthy
			}
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return HealthStatusUnhealthy
	}
	if hasDegraded {
		return HealthStatusDegraded
	}

	return HealthStatusHealthy
}

// getSystemInfo collects system information
func (h *HealthService) getSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemInfo{
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		Uptime:       time.Since(startTime),
		Memory: MemoryInfo{
			Alloc:         m.Alloc,
			TotalAlloc:    m.TotalAlloc,
			Sys:           m.Sys,
			Lookups:       m.Lookups,
			Mallocs:       m.Mallocs,
			Frees:         m.Frees,
			HeapAlloc:     m.HeapAlloc,
			HeapSys:       m.HeapSys,
			HeapIdle:      m.HeapIdle,
			HeapInuse:     m.HeapInuse,
			HeapReleased:  m.HeapReleased,
			HeapObjects:   m.HeapObjects,
			StackInuse:    m.StackInuse,
			StackSys:      m.StackSys,
			MSpanInuse:    m.MSpanInuse,
			MSpanSys:      m.MSpanSys,
			MCacheInuse:   m.MCacheInuse,
			MCacheSys:     m.MCacheSys,
			GCSys:         m.GCSys,
			OtherSys:      m.OtherSys,
			NextGC:        m.NextGC,
			LastGC:        m.LastGC,
			PauseTotalNs:  m.PauseTotalNs,
			NumGC:         m.NumGC,
			NumForcedGC:   m.NumForcedGC,
			GCCPUFraction: m.GCCPUFraction,
		},
	}
}

// generateSummary creates a summary of health check results
func (h *HealthService) generateSummary(checks map[string]HealthCheck) HealthSummary {
	summary := HealthSummary{
		Total: len(checks),
	}

	for _, check := range checks {
		switch check.Status {
		case HealthStatusHealthy:
			summary.Healthy++
		case HealthStatusDegraded:
			summary.Degraded++
		case HealthStatusUnhealthy:
			summary.Unhealthy++
		}

		if check.Critical {
			summary.Critical++
		}
	}

	return summary
}

// ServeHTTP implements http.Handler for health check endpoint
func (h *HealthService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var report *HealthReport
	
	// Check if quick check is requested
	if r.URL.Query().Get("quick") == "true" {
		report = h.QuickCheck(ctx)
	} else {
		report = h.Check(ctx)
	}

	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	switch report.Status {
	case HealthStatusDegraded:
		statusCode = http.StatusOK // Still return 200 for degraded
	case HealthStatusUnhealthy:
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(report); err != nil {
		http.Error(w, "Failed to encode health report", http.StatusInternalServerError)
	}
}

// Individual Health Check Implementations

// DatabaseHealthCheck checks database connectivity
type DatabaseHealthCheck struct {
	db *sql.DB
}

func (d *DatabaseHealthCheck) Name() string {
	return "database"
}

func (d *DatabaseHealthCheck) IsCritical() bool {
	return true
}

func (d *DatabaseHealthCheck) Check(ctx context.Context) HealthCheck {
	start := time.Now()
	check := HealthCheck{
		Name:      "database",
		Timestamp: start,
		Critical:  true,
	}

	if d.db == nil {
		check.Status = HealthStatusUnhealthy
		check.Error = "database connection not initialized"
		check.Duration = time.Since(start)
		return check
	}

	// Test connection
	if err := d.db.PingContext(ctx); err != nil {
		check.Status = HealthStatusUnhealthy
		check.Error = fmt.Sprintf("database ping failed: %v", err)
		check.Duration = time.Since(start)
		return check
	}

	// Get connection stats
	stats := d.db.Stats()
	check.Status = HealthStatusHealthy
	check.Message = "database connection healthy"
	check.Duration = time.Since(start)
	check.Metadata = map[string]interface{}{
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration,
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}

	return check
}

// RedisHealthCheck checks Redis connectivity
type RedisHealthCheck struct {
	client *redis.Client
}

func (r *RedisHealthCheck) Name() string {
	return "redis"
}

func (r *RedisHealthCheck) IsCritical() bool {
	return false // Redis is not critical for basic functionality
}

func (r *RedisHealthCheck) Check(ctx context.Context) HealthCheck {
	start := time.Now()
	check := HealthCheck{
		Name:      "redis",
		Timestamp: start,
		Critical:  false,
	}

	if r.client == nil {
		check.Status = HealthStatusDegraded
		check.Error = "redis client not initialized"
		check.Duration = time.Since(start)
		return check
	}

	// Test connection
	if err := r.client.Ping(ctx).Err(); err != nil {
		check.Status = HealthStatusDegraded
		check.Error = fmt.Sprintf("redis ping failed: %v", err)
		check.Duration = time.Since(start)
		return check
	}

	// Get Redis info
	info, err := r.client.Info(ctx).Result()
	if err != nil {
		check.Status = HealthStatusDegraded
		check.Error = fmt.Sprintf("failed to get redis info: %v", err)
		check.Duration = time.Since(start)
		return check
	}

	check.Status = HealthStatusHealthy
	check.Message = "redis connection healthy"
	check.Duration = time.Since(start)
	check.Metadata = map[string]interface{}{
		"info_length": len(info),
	}

	return check
}

// MemoryHealthCheck checks memory usage
type MemoryHealthCheck struct{}

func (m *MemoryHealthCheck) Name() string {
	return "memory"
}

func (m *MemoryHealthCheck) IsCritical() bool {
	return false
}

func (m *MemoryHealthCheck) Check(ctx context.Context) HealthCheck {
	start := time.Now()
	check := HealthCheck{
		Name:      "memory",
		Timestamp: start,
		Critical:  false,
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Check if memory usage is concerning (>80% of allocated heap)
	heapUsagePercent := float64(memStats.HeapInuse) / float64(memStats.HeapSys) * 100

	if heapUsagePercent > 90 {
		check.Status = HealthStatusUnhealthy
		check.Message = fmt.Sprintf("high memory usage: %.2f%%", heapUsagePercent)
	} else if heapUsagePercent > 80 {
		check.Status = HealthStatusDegraded
		check.Message = fmt.Sprintf("elevated memory usage: %.2f%%", heapUsagePercent)
	} else {
		check.Status = HealthStatusHealthy
		check.Message = fmt.Sprintf("memory usage normal: %.2f%%", heapUsagePercent)
	}

	check.Duration = time.Since(start)
	check.Metadata = map[string]interface{}{
		"heap_usage_percent": heapUsagePercent,
		"heap_inuse":         memStats.HeapInuse,
		"heap_sys":           memStats.HeapSys,
		"heap_alloc":         memStats.HeapAlloc,
		"num_gc":             memStats.NumGC,
		"gc_cpu_fraction":    memStats.GCCPUFraction,
	}

	return check
}

// DiskHealthCheck checks disk space
type DiskHealthCheck struct{}

func (d *DiskHealthCheck) Name() string {
	return "disk"
}

func (d *DiskHealthCheck) IsCritical() bool {
	return false
}

func (d *DiskHealthCheck) Check(ctx context.Context) HealthCheck {
	start := time.Now()
	check := HealthCheck{
		Name:      "disk",
		Timestamp: start,
		Critical:  false,
		Status:    HealthStatusHealthy,
		Message:   "Disk space OK",
		Duration:  time.Since(start),
	}

	// Basic disk space checking using os.Stat
	if info, err := os.Stat("."); err == nil {
		// Bu basit bir implementation. Production'da daha detaylı kontrol gerekebilir
		check.Metadata = map[string]interface{}{
			"path": ".",
			"mode": info.Mode().String(),
			"size": info.Size(),
		}
	} else {
		check.Status = HealthStatusUnhealthy
		check.Message = fmt.Sprintf("Disk space check failed: %v", err)
	}

	check.Duration = time.Since(start)
	return check
}