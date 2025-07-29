package security

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// SecurityManager handles comprehensive security management
type SecurityManager struct {
	db           *sql.DB
	config       SecurityConfig
	rateLimiter  RateLimiter
	ipWhitelist  map[string]bool
	ipBlacklist  map[string]bool
	validators   map[string]InputValidatorInterface
	scanners     []VulnerabilityScanner
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	MaxLoginAttempts     int           `json:"max_login_attempts"`
	LoginLockoutDuration time.Duration `json:"login_lockout_duration"`
	PasswordMinLength    int           `json:"password_min_length"`
	PasswordRequireUpper bool          `json:"password_require_upper"`
	PasswordRequireLower bool          `json:"password_require_lower"`
	PasswordRequireDigit bool          `json:"password_require_digit"`
	PasswordRequireSymbol bool         `json:"password_require_symbol"`
	SessionTimeout       time.Duration `json:"session_timeout"`
	CSRFTokenLength      int           `json:"csrf_token_length"`
	EnableIPWhitelist    bool          `json:"enable_ip_whitelist"`
	EnableIPBlacklist    bool          `json:"enable_ip_blacklist"`
	EnableRateLimit      bool          `json:"enable_rate_limit"`
	RateLimitRules       []RateLimitRule `json:"rate_limit_rules"`
	SecurityHeaders      SecurityHeaders `json:"security_headers"`
	SQLInjectionPatterns []string        `json:"sql_injection_patterns"`
	XSSPatterns          []string        `json:"xss_patterns"`
	FileUploadRules      FileUploadRules `json:"file_upload_rules"`
	EncryptionKey        string          `json:"encryption_key"`
	JWTSecret            string          `json:"jwt_secret"`
	TwoFactorEnabled     bool            `json:"two_factor_enabled"`
	AuditLogEnabled      bool            `json:"audit_log_enabled"`
}

// RateLimitRule defines rate limiting rules
type RateLimitRule struct {
	Path         string        `json:"path"`
	Method       string        `json:"method"`
	RequestsPerMinute int      `json:"requests_per_minute"`
	RequestsPerHour   int      `json:"requests_per_hour"`
	RequestsPerDay    int      `json:"requests_per_day"`
	BurstSize    int           `json:"burst_size"`
	WindowSize   time.Duration `json:"window_size"`
	Enabled      bool          `json:"enabled"`
}

// SecurityHeaders defines security headers configuration
type SecurityHeaders struct {
	ContentSecurityPolicy   string `json:"content_security_policy"`
	StrictTransportSecurity string `json:"strict_transport_security"`
	XFrameOptions          string `json:"x_frame_options"`
	XContentTypeOptions    string `json:"x_content_type_options"`
	XSSProtection          string `json:"x_xss_protection"`
	ReferrerPolicy         string `json:"referrer_policy"`
	PermissionsPolicy      string `json:"permissions_policy"`
	CrossOriginEmbedder    string `json:"cross_origin_embedder_policy"`
	CrossOriginOpener      string `json:"cross_origin_opener_policy"`
	CrossOriginResource    string `json:"cross_origin_resource_policy"`
}

// FileUploadRules defines file upload security rules
type FileUploadRules struct {
	MaxFileSize      int64    `json:"max_file_size"`
	AllowedMimeTypes []string `json:"allowed_mime_types"`
	AllowedExtensions []string `json:"allowed_extensions"`
	BlockedExtensions []string `json:"blocked_extensions"`
	ScanForMalware   bool     `json:"scan_for_malware"`
	QuarantinePath   string   `json:"quarantine_path"`
}

// SecurityEvent represents a security event
type SecurityEvent struct {
	ID          string                 `json:"id"`
	Type        SecurityEventType      `json:"type"`
	Severity    SecuritySeverity       `json:"severity"`
	Source      string                 `json:"source"`
	Target      string                 `json:"target"`
	UserID      string                 `json:"user_id"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Method      string                 `json:"method"`
	URL         string                 `json:"url"`
	Payload     string                 `json:"payload"`
	Headers     map[string]string      `json:"headers"`
	Timestamp   time.Time              `json:"timestamp"`
	Blocked     bool                   `json:"blocked"`
	RiskScore   float64                `json:"risk_score"`
	Details     map[string]interface{} `json:"details"`
	Resolution  string                 `json:"resolution"`
	ResolvedAt  *time.Time             `json:"resolved_at"`
	ResolvedBy  string                 `json:"resolved_by"`
}

// SecurityEventType represents different types of security events
type SecurityEventType string

const (
	EventTypeLoginAttempt       SecurityEventType = "login_attempt"
	EventTypeLoginFailure       SecurityEventType = "login_failure"
	EventTypeLoginSuccess       SecurityEventType = "login_success"
	EventTypeBruteForce         SecurityEventType = "brute_force"
	EventTypeSQLInjection       SecurityEventType = "sql_injection"
	EventTypeXSS                SecurityEventType = "xss"
	EventTypeCSRF               SecurityEventType = "csrf"
	EventTypeFileUpload         SecurityEventType = "file_upload"
	EventTypeUnauthorizedAccess SecurityEventType = "unauthorized_access"
	EventTypePrivilegeEscalation SecurityEventType = "privilege_escalation"
	EventTypeDataBreach         SecurityEventType = "data_breach"
	EventTypeSuspiciousActivity SecurityEventType = "suspicious_activity"
	EventTypeRateLimitExceeded  SecurityEventType = "rate_limit_exceeded"
	EventTypeIPBlocked          SecurityEventType = "ip_blocked"
	EventTypeMalwareDetected    SecurityEventType = "malware_detected"
	THREAT_DETECTED             SecurityEventType = "threat_detected"
)

// SecuritySeverity represents security event severity levels
type SecuritySeverity string

const (
	SeverityLow      SecuritySeverity = "low"
	SeverityMedium   SecuritySeverity = "medium"
	SeverityHigh     SecuritySeverity = "high"
	SeverityCritical SecuritySeverity = "critical"
	HIGH             SecuritySeverity = "high"
)

// RateLimiter interface for rate limiting
type RateLimiter interface {
	Allow(key string, rule RateLimitRule) bool
	GetUsage(key string) (int, error)
	Reset(key string) error
}

// InputValidatorInterface interface for input validation
type InputValidatorInterface interface {
	Validate(input string, rules ValidationRules) []ValidationError
	SanitizeInput(input string) string
}

// ValidationRules defines input validation rules
type ValidationRules struct {
	MinLength    int      `json:"min_length"`
	MaxLength    int      `json:"max_length"`
	AllowedChars string   `json:"allowed_chars"`
	BlockedChars string   `json:"blocked_chars"`
	Patterns     []string `json:"patterns"`
	Required     bool     `json:"required"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// VulnerabilityScanner interface for vulnerability scanning
type VulnerabilityScanner interface {
	Scan(target string, scanType ScanType) (*ScanResult, error)
	GetName() string
	IsEnabled() bool
}

// ScanType represents different types of vulnerability scans
type ScanType string

const (
	ScanTypeSQL      ScanType = "sql"
	ScanTypeXSS      ScanType = "xss"
	ScanTypeCSRF     ScanType = "csrf"
	ScanTypeMalware  ScanType = "malware"
	ScanTypeGeneral  ScanType = "general"
)

// ScanResult represents vulnerability scan result
type ScanResult struct {
	ScanID        string                 `json:"scan_id"`
	Target        string                 `json:"target"`
	ScanType      ScanType               `json:"scan_type"`
	Status        string                 `json:"status"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       time.Time              `json:"end_time"`
	Duration      time.Duration          `json:"duration"`
	Vulnerabilities []Vulnerability      `json:"vulnerabilities"`
	RiskScore     float64                `json:"risk_score"`
	Recommendations []string             `json:"recommendations"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// Vulnerability represents a detected vulnerability
type Vulnerability struct {
	ID          string           `json:"id"`
	Type        string           `json:"type"`
	Severity    SecuritySeverity `json:"severity"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Location    string           `json:"location"`
	Evidence    string           `json:"evidence"`
	Impact      string           `json:"impact"`
	Solution    string           `json:"solution"`
	References  []string         `json:"references"`
	CVSS        float64          `json:"cvss"`
	CWE         string           `json:"cwe"`
	OWASP       string           `json:"owasp"`
}

// SecurityReport represents a comprehensive security report
type SecurityReport struct {
	ID              string                    `json:"id"`
	GeneratedAt     time.Time                 `json:"generated_at"`
	Period          string                    `json:"period"`
	TotalEvents     int                       `json:"total_events"`
	EventsByType    map[SecurityEventType]int `json:"events_by_type"`
	EventsBySeverity map[SecuritySeverity]int `json:"events_by_severity"`
	TopThreats      []ThreatSummary           `json:"top_threats"`
	VulnerabilityStats VulnerabilityStats     `json:"vulnerability_stats"`
	ComplianceStatus ComplianceStatus         `json:"compliance_status"`
	Recommendations []SecurityRecommendation  `json:"recommendations"`
	TrendData       []SecurityTrendPoint      `json:"trend_data"`
}

// ThreatSummary represents a threat summary
type ThreatSummary struct {
	Type        SecurityEventType `json:"type"`
	Count       int               `json:"count"`
	Severity    SecuritySeverity  `json:"severity"`
	Sources     []string          `json:"sources"`
	Targets     []string          `json:"targets"`
	FirstSeen   time.Time         `json:"first_seen"`
	LastSeen    time.Time         `json:"last_seen"`
}

// VulnerabilityStats represents vulnerability statistics
type VulnerabilityStats struct {
	Total        int                       `json:"total"`
	BySeverity   map[SecuritySeverity]int  `json:"by_severity"`
	ByType       map[string]int            `json:"by_type"`
	Fixed        int                       `json:"fixed"`
	Open         int                       `json:"open"`
	InProgress   int                       `json:"in_progress"`
	AverageFixTime time.Duration           `json:"average_fix_time"`
}

// ComplianceStatus represents compliance status
type ComplianceStatus struct {
	GDPR    ComplianceItem `json:"gdpr"`
	PCI     ComplianceItem `json:"pci"`
	SOX     ComplianceItem `json:"sox"`
	HIPAA   ComplianceItem `json:"hipaa"`
	ISO27001 ComplianceItem `json:"iso27001"`
}

// ComplianceItem represents a compliance item
type ComplianceItem struct {
	Status      string    `json:"status"`
	Score       float64   `json:"score"`
	LastChecked time.Time `json:"last_checked"`
	Issues      []string  `json:"issues"`
}

// SecurityRecommendation represents a security recommendation
type SecurityRecommendation struct {
	ID          string           `json:"id"`
	Type        string           `json:"type"`
	Priority    string           `json:"priority"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Impact      string           `json:"impact"`
	Effort      string           `json:"effort"`
	Category    string           `json:"category"`
	Status      string           `json:"status"`
	DueDate     *time.Time       `json:"due_date"`
}

// SecurityTrendPoint represents a point in security trend data
type SecurityTrendPoint struct {
	Timestamp time.Time                 `json:"timestamp"`
	Events    map[SecurityEventType]int `json:"events"`
	RiskScore float64                   `json:"risk_score"`
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(db *sql.DB, config SecurityConfig) *SecurityManager {
	sm := &SecurityManager{
		db:          db,
		config:      config,
		ipWhitelist: make(map[string]bool),
		ipBlacklist: make(map[string]bool),
		validators:  make(map[string]InputValidatorInterface),
		scanners:    make([]VulnerabilityScanner, 0),
	}

	sm.createSecurityTables()
	sm.initializeRateLimiter()
	sm.loadIPLists()
	sm.initializeValidators()
	sm.initializeScanners()

	return sm
}

// createSecurityTables creates necessary tables for security management
func (sm *SecurityManager) createSecurityTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS security_events (
			id VARCHAR(128) PRIMARY KEY,
			type VARCHAR(50) NOT NULL,
			severity VARCHAR(20) NOT NULL,
			source VARCHAR(255),
			target VARCHAR(255),
			user_id VARCHAR(128),
			ip_address VARCHAR(45),
			user_agent TEXT,
			method VARCHAR(10),
			url TEXT,
			payload TEXT,
			headers TEXT,
			timestamp DATETIME NOT NULL,
			blocked BOOLEAN DEFAULT FALSE,
			risk_score DECIMAL(5,2) DEFAULT 0.00,
			details TEXT,
			resolution TEXT,
			resolved_at DATETIME,
			resolved_by VARCHAR(128),
			INDEX idx_type (type),
			INDEX idx_severity (severity),
			INDEX idx_timestamp (timestamp),
			INDEX idx_ip_address (ip_address),
			INDEX idx_user_id (user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS ip_whitelist (
			id INT AUTO_INCREMENT PRIMARY KEY,
			ip_address VARCHAR(45) NOT NULL,
			cidr_range VARCHAR(50),
			description TEXT,
			enabled BOOLEAN DEFAULT TRUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY unique_ip (ip_address),
			INDEX idx_enabled (enabled)
		)`,
		`CREATE TABLE IF NOT EXISTS ip_blacklist (
			id INT AUTO_INCREMENT PRIMARY KEY,
			ip_address VARCHAR(45) NOT NULL,
			cidr_range VARCHAR(50),
			reason TEXT,
			blocked_until DATETIME,
			auto_blocked BOOLEAN DEFAULT FALSE,
			enabled BOOLEAN DEFAULT TRUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY unique_ip (ip_address),
			INDEX idx_enabled (enabled),
			INDEX idx_blocked_until (blocked_until)
		)`,
		`CREATE TABLE IF NOT EXISTS rate_limit_tracking (
			id VARCHAR(128) PRIMARY KEY,
			key_hash VARCHAR(64) NOT NULL,
			rule_path VARCHAR(255) NOT NULL,
			request_count INT DEFAULT 0,
			window_start DATETIME NOT NULL,
			window_end DATETIME NOT NULL,
			last_request DATETIME DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_key_hash (key_hash),
			INDEX idx_window_end (window_end)
		)`,
		`CREATE TABLE IF NOT EXISTS vulnerability_scans (
			id VARCHAR(128) PRIMARY KEY,
			target VARCHAR(512) NOT NULL,
			scan_type VARCHAR(50) NOT NULL,
			status VARCHAR(20) NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			duration_seconds INT DEFAULT 0,
			vulnerabilities_found INT DEFAULT 0,
			risk_score DECIMAL(5,2) DEFAULT 0.00,
			scan_data TEXT,
			recommendations TEXT,
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_target (target),
			INDEX idx_scan_type (scan_type),
			INDEX idx_status (status),
			INDEX idx_start_time (start_time)
		)`,
		`CREATE TABLE IF NOT EXISTS security_audit_log (
			id VARCHAR(128) PRIMARY KEY,
			user_id VARCHAR(128),
			action VARCHAR(100) NOT NULL,
			resource VARCHAR(255),
			resource_id VARCHAR(128),
			ip_address VARCHAR(45),
			user_agent TEXT,
			details TEXT,
			success BOOLEAN DEFAULT TRUE,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_user_id (user_id),
			INDEX idx_action (action),
			INDEX idx_timestamp (timestamp),
			INDEX idx_success (success)
		)`,
	}

	for _, query := range queries {
		if _, err := sm.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create security table: %w", err)
		}
	}

	return nil
}

// SecurityMiddleware provides security middleware for HTTP requests
func (sm *SecurityManager) SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Add security headers
		sm.addSecurityHeaders(w)

		// Check IP whitelist/blacklist
		blocked, _ := sm.CheckIPAccess(r.RemoteAddr)
	if blocked {
			sm.logSecurityEvent(ctx, EventTypeIPBlocked, SeverityHigh, r, "IP blocked", nil)
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Rate limiting
		if sm.config.EnableRateLimit {
			if !sm.checkRateLimit(r) {
				sm.logSecurityEvent(ctx, EventTypeRateLimitExceeded, SeverityMedium, r, "Rate limit exceeded", nil)
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
		}

		// Input validation and sanitization
		if err := sm.validateRequest(r); err != nil {
			sm.logSecurityEvent(ctx, EventTypeSuspiciousActivity, SeverityMedium, r, "Invalid input detected", map[string]interface{}{
				"error": err.Error(),
			})
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		// SQL injection detection
		if sm.detectSQLInjection(r) {
			sm.logSecurityEvent(ctx, EventTypeSQLInjection, SeverityCritical, r, "SQL injection attempt detected", nil)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// XSS detection
		if sm.detectXSS(r) {
			sm.logSecurityEvent(ctx, EventTypeXSS, SeverityHigh, r, "XSS attempt detected", nil)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// CSRF protection
		if sm.requiresCSRFProtection(r) && !sm.validateCSRFToken(r) {
			sm.logSecurityEvent(ctx, EventTypeCSRF, SeverityHigh, r, "CSRF token validation failed", nil)
			http.Error(w, "CSRF token invalid", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// addSecurityHeaders adds security headers to response
func (sm *SecurityManager) addSecurityHeaders(w http.ResponseWriter) {
	headers := sm.config.SecurityHeaders

	if headers.ContentSecurityPolicy != "" {
		w.Header().Set("Content-Security-Policy", headers.ContentSecurityPolicy)
	}
	if headers.StrictTransportSecurity != "" {
		w.Header().Set("Strict-Transport-Security", headers.StrictTransportSecurity)
	}
	if headers.XFrameOptions != "" {
		w.Header().Set("X-Frame-Options", headers.XFrameOptions)
	}
	if headers.XContentTypeOptions != "" {
		w.Header().Set("X-Content-Type-Options", headers.XContentTypeOptions)
	}
	if headers.XSSProtection != "" {
		w.Header().Set("X-XSS-Protection", headers.XSSProtection)
	}
	if headers.ReferrerPolicy != "" {
		w.Header().Set("Referrer-Policy", headers.ReferrerPolicy)
	}
	if headers.PermissionsPolicy != "" {
		w.Header().Set("Permissions-Policy", headers.PermissionsPolicy)
	}
	if headers.CrossOriginEmbedder != "" {
		w.Header().Set("Cross-Origin-Embedder-Policy", headers.CrossOriginEmbedder)
	}
	if headers.CrossOriginOpener != "" {
		w.Header().Set("Cross-Origin-Opener-Policy", headers.CrossOriginOpener)
	}
	if headers.CrossOriginResource != "" {
		w.Header().Set("Cross-Origin-Resource-Policy", headers.CrossOriginResource)
	}
}

// CheckIPAccess checks if IP is allowed
func (sm *SecurityManager) CheckIPAccess(remoteAddr string) (bool, string) {
	ip := sm.extractIP(remoteAddr)

	// Check blacklist first
	if sm.config.EnableIPBlacklist {
		if sm.isIPBlacklisted(ip) {
			return true, "IP is blacklisted"
		}
	}

	// Check whitelist
	if sm.config.EnableIPWhitelist {
		if !sm.isIPWhitelisted(ip) {
			return true, "IP not in whitelist"
		}
	}

	return false, ""
}

// extractIP extracts IP from remote address
func (sm *SecurityManager) extractIP(remoteAddr string) string {
	// Extract IP from remote address (format: "IP:port")
	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		return remoteAddr[:idx]
	}
	return remoteAddr
}

// SetSecurityHeaders sets security headers on response
func (sm *SecurityManager) SetSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("Content-Security-Policy", "default-src 'self'")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
}

// CheckRateLimit checks if request exceeds rate limits
func (sm *SecurityManager) CheckRateLimit(r *http.Request) (bool, error) {
	if !sm.config.EnableRateLimit {
		return false, nil
	}
	
	// Simple rate limiting implementation
	// In a real implementation, this would use a proper rate limiter
	return false, nil
}

// ValidateInput validates request input for security threats
func (sm *SecurityManager) ValidateInput(r *http.Request) error {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}
	
	// Check for SQL injection patterns
	for _, values := range r.Form {
		for _, value := range values {
			if sm.containsSQLInjection(value) {
				return fmt.Errorf("potential SQL injection detected")
			}
			if sm.containsXSS(value) {
				return fmt.Errorf("potential XSS detected")
			}
		}
	}
	
	return nil
}

// ScanForThreats scans request for security threats
func (sm *SecurityManager) ScanForThreats(r *http.Request) []string {
	var threats []string
	
	// Check URL for threats
	if sm.containsSQLInjection(r.URL.String()) {
		threats = append(threats, "SQL injection in URL")
	}
	
	// Check headers for threats
	for name, values := range r.Header {
		for _, value := range values {
			if sm.containsXSS(value) {
				threats = append(threats, fmt.Sprintf("XSS in header %s", name))
			}
		}
	}
	
	return threats
}

// LogSecurityEvent logs a security event
func (sm *SecurityManager) LogSecurityEvent(event SecurityEvent) error {
	// In a real implementation, this would store to database
	fmt.Printf("Security Event: %+v\n", event)
	return nil
}

// ValidateCSRFToken validates CSRF token
func (sm *SecurityManager) ValidateCSRFToken(token string, r *http.Request) bool {
	// Simple CSRF validation - in production, use proper CSRF tokens
	return token != ""
}

// containsSQLInjection checks for SQL injection patterns
func (sm *SecurityManager) containsSQLInjection(input string) bool {
	patterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_", "DROP", "SELECT", "INSERT", "UPDATE", "DELETE",
	}
	
	input = strings.ToUpper(input)
	for _, pattern := range patterns {
		if strings.Contains(input, strings.ToUpper(pattern)) {
			return true
		}
	}
	return false
}

// containsXSS checks for XSS patterns
func (sm *SecurityManager) containsXSS(input string) bool {
	patterns := []string{
		"<script", "</script>", "javascript:", "onload=", "onerror=", "onclick=",
	}
	
	input = strings.ToLower(input)
	for _, pattern := range patterns {
		if strings.Contains(input, pattern) {
			return true
		}
	}
	return false
}

// checkRateLimit checks if request exceeds rate limits
func (sm *SecurityManager) checkRateLimit(r *http.Request) bool {
	if sm.rateLimiter == nil {
		return true
	}

	// Find applicable rate limit rule
	rule := sm.findRateLimitRule(r)
	if rule == nil {
		return true
	}

	// Create rate limit key
	key := sm.createRateLimitKey(r, rule)

	return sm.rateLimiter.Allow(key, *rule)
}

// validateRequest validates the entire request
func (sm *SecurityManager) validateRequest(r *http.Request) error {
	// Validate query parameters
	for key, values := range r.URL.Query() {
		for _, value := range values {
			if err := sm.validateInput(key, value); err != nil {
				return err
			}
		}
	}

	// Validate form data
	if r.Method == "POST" || r.Method == "PUT" {
		r.ParseForm()
		for key, values := range r.PostForm {
			for _, value := range values {
				if err := sm.validateInput(key, value); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// validateInput validates individual input
func (sm *SecurityManager) validateInput(field, value string) error {
	// Basic validation rules
	rules := ValidationRules{
		MaxLength: 10000, // Prevent extremely long inputs
	}

	validator := sm.validators["default"]
	if validator != nil {
		errors := validator.Validate(value, rules)
		if len(errors) > 0 {
			return fmt.Errorf("validation failed for field %s: %s", field, errors[0].Message)
		}
	}

	return nil
}

// detectSQLInjection detects SQL injection attempts
func (sm *SecurityManager) detectSQLInjection(r *http.Request) bool {
	patterns := sm.config.SQLInjectionPatterns
	if len(patterns) == 0 {
		// Default SQL injection patterns
		patterns = []string{
			`(?i)(union\s+select)`,
			`(?i)(select\s+.*\s+from)`,
			`(?i)(insert\s+into)`,
			`(?i)(delete\s+from)`,
			`(?i)(drop\s+table)`,
			`(?i)(update\s+.*\s+set)`,
			`(?i)(\'\s*or\s*\'\s*=\s*\')`,
			`(?i)(\'\s*or\s*1\s*=\s*1)`,
			`(?i)(--\s)`,
			`(?i)(/\*.*\*/)`,
		}
	}

	// Check all input sources
	inputs := sm.extractAllInputs(r)
	
	for _, input := range inputs {
		for _, pattern := range patterns {
			if matched, _ := regexp.MatchString(pattern, input); matched {
				return true
			}
		}
	}

	return false
}

// detectXSS detects XSS attempts
func (sm *SecurityManager) detectXSS(r *http.Request) bool {
	patterns := sm.config.XSSPatterns
	if len(patterns) == 0 {
		// Default XSS patterns
		patterns = []string{
			`(?i)<script[^>]*>.*?</script>`,
			`(?i)<iframe[^>]*>.*?</iframe>`,
			`(?i)<object[^>]*>.*?</object>`,
			`(?i)<embed[^>]*>`,
			`(?i)<link[^>]*>`,
			`(?i)javascript:`,
			`(?i)vbscript:`,
			`(?i)onload\s*=`,
			`(?i)onerror\s*=`,
			`(?i)onclick\s*=`,
			`(?i)onmouseover\s*=`,
		}
	}

	inputs := sm.extractAllInputs(r)
	
	for _, input := range inputs {
		for _, pattern := range patterns {
			if matched, _ := regexp.MatchString(pattern, input); matched {
				return true
			}
		}
	}

	return false
}

// validateCSRFToken validates CSRF token
func (sm *SecurityManager) validateCSRFToken(r *http.Request) bool {
	// Get token from header or form
	token := r.Header.Get("X-CSRF-Token")
	if token == "" {
		token = r.FormValue("csrf_token")
	}

	if token == "" {
		return false
	}

	// Validate token (implementation would verify against session)
	return sm.isValidCSRFToken(token, r)
}

// requiresCSRFProtection checks if request requires CSRF protection
func (sm *SecurityManager) requiresCSRFProtection(r *http.Request) bool {
	// CSRF protection for state-changing methods
	return r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" || r.Method == "PATCH"
}

// logSecurityEvent logs a security event
func (sm *SecurityManager) logSecurityEvent(ctx context.Context, eventType SecurityEventType, severity SecuritySeverity, r *http.Request, message string, details map[string]interface{}) {
	event := &SecurityEvent{
		ID:        sm.generateEventID(),
		Type:      eventType,
		Severity:  severity,
		IPAddress: sm.getRealIP(r),
		UserAgent: r.UserAgent(),
		Method:    r.Method,
		URL:       r.URL.String(),
		Timestamp: time.Now(),
		Details:   details,
	}

	// Extract user ID from context if available
	if userID := ctx.Value("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			event.UserID = uid
		}
	}

	// Extract headers (filtered)
	event.Headers = sm.extractSafeHeaders(r)

	// Calculate risk score
	event.RiskScore = sm.calculateRiskScore(event)

	// Store in database
	sm.storeSecurityEvent(event)

	// Check if automatic blocking is needed
	if sm.shouldAutoBlock(event) {
		sm.autoBlockIP(event.IPAddress, fmt.Sprintf("Auto-blocked due to %s", eventType))
	}
}

// GenerateSecurityReport generates a comprehensive security report
func (sm *SecurityManager) GenerateSecurityReport(startDate, endDate time.Time) (*SecurityReport, error) {
	report := &SecurityReport{
		ID:               sm.generateReportID(),
		GeneratedAt:      time.Now(),
		Period:           fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		EventsByType:     make(map[SecurityEventType]int),
		EventsBySeverity: make(map[SecuritySeverity]int),
	}

	// Get total events
	err := sm.db.QueryRow(`
		SELECT COUNT(*) FROM security_events 
		WHERE timestamp BETWEEN ? AND ?
	`, startDate, endDate).Scan(&report.TotalEvents)
	if err != nil {
		return nil, err
	}

	// Get events by type
	rows, err := sm.db.Query(`
		SELECT type, COUNT(*) FROM security_events 
		WHERE timestamp BETWEEN ? AND ?
		GROUP BY type
	`, startDate, endDate)
	if err == nil {
		for rows.Next() {
			var eventType string
			var count int
			if err := rows.Scan(&eventType, &count); err == nil {
				report.EventsByType[SecurityEventType(eventType)] = count
			}
		}
		rows.Close()
	}

	// Get events by severity
	rows, err = sm.db.Query(`
		SELECT severity, COUNT(*) FROM security_events 
		WHERE timestamp BETWEEN ? AND ?
		GROUP BY severity
	`, startDate, endDate)
	if err == nil {
		for rows.Next() {
			var severity string
			var count int
			if err := rows.Scan(&severity, &count); err == nil {
				report.EventsBySeverity[SecuritySeverity(severity)] = count
			}
		}
		rows.Close()
	}

	// Get top threats
	report.TopThreats = sm.getTopThreats(startDate, endDate)

	// Get vulnerability statistics
	report.VulnerabilityStats = sm.getVulnerabilityStats()

	// Get compliance status
	report.ComplianceStatus = sm.getComplianceStatus()

	// Generate recommendations
	report.Recommendations = sm.generateSecurityRecommendations()

	// Get trend data
	report.TrendData = sm.getSecurityTrendData(startDate, endDate)

	return report, nil
}

// Helper methods

func (sm *SecurityManager) getRealIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

func (sm *SecurityManager) isIPBlacklisted(ip string) bool {
	// Check in-memory cache first
	if sm.ipBlacklist[ip] {
		return true
	}

	// Check database
	var count int
	err := sm.db.QueryRow(`
		SELECT COUNT(*) FROM ip_blacklist 
		WHERE ip_address = ? AND enabled = TRUE 
		AND (blocked_until IS NULL OR blocked_until > NOW())
	`, ip).Scan(&count)
	
	return err == nil && count > 0
}

func (sm *SecurityManager) isIPWhitelisted(ip string) bool {
	// Check in-memory cache first
	if sm.ipWhitelist[ip] {
		return true
	}

	// Check database
	var count int
	err := sm.db.QueryRow(`
		SELECT COUNT(*) FROM ip_whitelist 
		WHERE ip_address = ? AND enabled = TRUE
	`, ip).Scan(&count)
	
	return err == nil && count > 0
}

func (sm *SecurityManager) extractAllInputs(r *http.Request) []string {
	inputs := make([]string, 0)

	// Query parameters
	for _, values := range r.URL.Query() {
		for _, value := range values {
			inputs = append(inputs, value)
		}
	}

	// Form data
	if r.Method == "POST" || r.Method == "PUT" {
		r.ParseForm()
		for _, values := range r.PostForm {
			for _, value := range values {
				inputs = append(inputs, value)
			}
		}
	}

	// Headers (selected ones)
	safeHeaders := []string{"User-Agent", "Referer", "Accept", "Accept-Language"}
	for _, header := range safeHeaders {
		if value := r.Header.Get(header); value != "" {
			inputs = append(inputs, value)
		}
	}

	return inputs
}

func (sm *SecurityManager) extractSafeHeaders(r *http.Request) map[string]string {
	safeHeaders := make(map[string]string)
	sensitiveHeaders := map[string]bool{
		"authorization": true,
		"cookie":        true,
		"x-api-key":     true,
		"x-auth-token":  true,
	}

	for key, values := range r.Header {
		lowerKey := strings.ToLower(key)
		if !sensitiveHeaders[lowerKey] && len(values) > 0 {
			safeHeaders[key] = values[0]
		}
	}

	return safeHeaders
}

func (sm *SecurityManager) generateEventID() string {
	return fmt.Sprintf("sec_event_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond())
}

func (sm *SecurityManager) generateReportID() string {
	return fmt.Sprintf("sec_report_%d", time.Now().UnixNano())
}

func (sm *SecurityManager) calculateRiskScore(event *SecurityEvent) float64 {
	score := 0.0

	// Base score by event type
	switch event.Type {
	case EventTypeSQLInjection, EventTypeDataBreach:
		score += 9.0
	case EventTypeXSS, EventTypeCSRF, EventTypeMalwareDetected:
		score += 7.0
	case EventTypeBruteForce, EventTypeUnauthorizedAccess:
		score += 5.0
	case EventTypeRateLimitExceeded, EventTypeSuspiciousActivity:
		score += 3.0
	default:
		score += 1.0
	}

	// Adjust by severity
	switch event.Severity {
	case SeverityCritical:
		score *= 1.5
	case SeverityHigh:
		score *= 1.2
	case SeverityMedium:
		score *= 1.0
	case SeverityLow:
		score *= 0.8
	}

	// Cap at 10.0
	if score > 10.0 {
		score = 10.0
	}

	return score
}

func (sm *SecurityManager) storeSecurityEvent(event *SecurityEvent) {
	detailsJSON, _ := json.Marshal(event.Details)
	headersJSON, _ := json.Marshal(event.Headers)

	query := `
		INSERT INTO security_events (
			id, type, severity, source, target, user_id, ip_address, user_agent,
			method, url, payload, headers, timestamp, blocked, risk_score, details
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	sm.db.Exec(query,
		event.ID, event.Type, event.Severity, event.Source, event.Target,
		event.UserID, event.IPAddress, event.UserAgent, event.Method, event.URL,
		event.Payload, string(headersJSON), event.Timestamp, event.Blocked,
		event.RiskScore, string(detailsJSON),
	)
}

func (sm *SecurityManager) shouldAutoBlock(event *SecurityEvent) bool {
	// Auto-block for critical threats
	if event.Severity == SeverityCritical {
		return true
	}

	// Auto-block for repeated high-severity events from same IP
	if event.Severity == SeverityHigh {
		var count int
		sm.db.QueryRow(`
			SELECT COUNT(*) FROM security_events 
			WHERE ip_address = ? AND severity = 'high' 
			AND timestamp > DATE_SUB(NOW(), INTERVAL 1 HOUR)
		`, event.IPAddress).Scan(&count)
		
		return count >= 3
	}

	return false
}

func (sm *SecurityManager) autoBlockIP(ip, reason string) {
	query := `
		INSERT INTO ip_blacklist (ip_address, reason, blocked_until, auto_blocked, enabled)
		VALUES (?, ?, DATE_ADD(NOW(), INTERVAL 24 HOUR), TRUE, TRUE)
		ON DUPLICATE KEY UPDATE
		reason = VALUES(reason), blocked_until = VALUES(blocked_until),
		auto_blocked = TRUE, enabled = TRUE, updated_at = NOW()
	`

	sm.db.Exec(query, ip, reason)
	
	// Update in-memory cache
	sm.ipBlacklist[ip] = true
}

// Placeholder implementations for missing methods
func (sm *SecurityManager) initializeRateLimiter() {}
func (sm *SecurityManager) loadIPLists() {}
func (sm *SecurityManager) initializeValidators() {}
func (sm *SecurityManager) initializeScanners() {}
func (sm *SecurityManager) findRateLimitRule(r *http.Request) *RateLimitRule { return nil }
func (sm *SecurityManager) createRateLimitKey(r *http.Request, rule *RateLimitRule) string { return "" }
func (sm *SecurityManager) isValidCSRFToken(token string, r *http.Request) bool { return true }
func (sm *SecurityManager) getTopThreats(startDate, endDate time.Time) []ThreatSummary { return []ThreatSummary{} }
func (sm *SecurityManager) getVulnerabilityStats() VulnerabilityStats { return VulnerabilityStats{} }
func (sm *SecurityManager) getComplianceStatus() ComplianceStatus { return ComplianceStatus{} }
func (sm *SecurityManager) generateSecurityRecommendations() []SecurityRecommendation { return []SecurityRecommendation{} }
func (sm *SecurityManager) getSecurityTrendData(startDate, endDate time.Time) []SecurityTrendPoint { return []SecurityTrendPoint{} }