package testing

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestManager handles comprehensive test management
type TestManager struct {
	db         *sql.DB
	config     TestConfig
	suites     map[string]*TestSuite
	runners    map[string]TestRunner
	reporters  []TestReporter
	hooks      TestHooks
	coverage   *CoverageCollector
}

// TestConfig holds test configuration
type TestConfig struct {
	Environment      string                 `json:"environment"`
	Parallel         bool                   `json:"parallel"`
	MaxWorkers       int                    `json:"max_workers"`
	Timeout          time.Duration          `json:"timeout"`
	RetryAttempts    int                    `json:"retry_attempts"`
	CoverageEnabled  bool                   `json:"coverage_enabled"`
	CoverageThreshold float64               `json:"coverage_threshold"`
	ReportFormats    []string               `json:"report_formats"`
	OutputDirectory  string                 `json:"output_directory"`
	DatabaseConfig   DatabaseTestConfig     `json:"database_config"`
	APIConfig        APITestConfig          `json:"api_config"`
	UIConfig         UITestConfig           `json:"ui_config"`
	PerformanceConfig PerformanceTestConfig `json:"performance_config"`
	SecurityConfig   SecurityTestConfig     `json:"security_config"`
	Tags             []string               `json:"tags"`
	ExcludeTags      []string               `json:"exclude_tags"`
	Filters          []TestFilter           `json:"filters"`
}

// DatabaseTestConfig holds database test configuration
type DatabaseTestConfig struct {
	TestDBURL        string            `json:"test_db_url"`
	MigrationsPath   string            `json:"migrations_path"`
	SeedDataPath     string            `json:"seed_data_path"`
	TransactionMode  string            `json:"transaction_mode"` // "rollback", "truncate", "recreate"
	IsolationLevel   string            `json:"isolation_level"`
	ConnectionPool   int               `json:"connection_pool"`
	QueryTimeout     time.Duration     `json:"query_timeout"`
	CustomSetup      []string          `json:"custom_setup"`
	CustomTeardown   []string          `json:"custom_teardown"`
}

// APITestConfig holds API test configuration
type APITestConfig struct {
	BaseURL          string            `json:"base_url"`
	Headers          map[string]string `json:"headers"`
	Timeout          time.Duration     `json:"timeout"`
	RetryAttempts    int               `json:"retry_attempts"`
	MockEnabled      bool              `json:"mock_enabled"`
	MockDataPath     string            `json:"mock_data_path"`
	AuthConfig       AuthTestConfig    `json:"auth_config"`
	RateLimitConfig  RateLimitConfig   `json:"rate_limit_config"`
}

// AuthTestConfig holds authentication test configuration
type AuthTestConfig struct {
	Enabled       bool              `json:"enabled"`
	Type          string            `json:"type"` // "jwt", "session", "api_key"
	TestUsers     []TestUser        `json:"test_users"`
	TokenEndpoint string            `json:"token_endpoint"`
	RefreshToken  bool              `json:"refresh_token"`
	Scopes        []string          `json:"scopes"`
}

// TestUser represents a test user
type TestUser struct {
	ID          string            `json:"id"`
	Username    string            `json:"username"`
	Password    string            `json:"password"`
	Email       string            `json:"email"`
	Roles       []string          `json:"roles"`
	Permissions []string          `json:"permissions"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RateLimitConfig holds rate limiting test configuration
type RateLimitConfig struct {
	Enabled     bool          `json:"enabled"`
	RequestsPerSecond int     `json:"requests_per_second"`
	BurstSize   int           `json:"burst_size"`
	TestWindow  time.Duration `json:"test_window"`
}

// UITestConfig holds UI test configuration
type UITestConfig struct {
	Enabled         bool              `json:"enabled"`
	Browser         string            `json:"browser"`
	Headless        bool              `json:"headless"`
	WindowSize      WindowSize        `json:"window_size"`
	ImplicitWait    time.Duration     `json:"implicit_wait"`
	PageLoadTimeout time.Duration     `json:"page_load_timeout"`
	ScreenshotPath  string            `json:"screenshot_path"`
	VideoRecording  bool              `json:"video_recording"`
	Selectors       SelectorConfig    `json:"selectors"`
}

// WindowSize represents browser window size
type WindowSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// SelectorConfig holds selector configuration
type SelectorConfig struct {
	Strategy        string            `json:"strategy"` // "css", "xpath", "id", "class"
	Timeout         time.Duration     `json:"timeout"`
	RetryInterval   time.Duration     `json:"retry_interval"`
	CustomSelectors map[string]string `json:"custom_selectors"`
}

// PerformanceTestConfig holds performance test configuration
type PerformanceTestConfig struct {
	Enabled           bool              `json:"enabled"`
	LoadTestConfig    LoadTestConfig    `json:"load_test_config"`
	StressTestConfig  StressTestConfig  `json:"stress_test_config"`
	SpikeTestConfig   SpikeTestConfig   `json:"spike_test_config"`
	VolumeTestConfig  VolumeTestConfig  `json:"volume_test_config"`
	Thresholds        PerformanceThresholds `json:"thresholds"`
}

// LoadTestConfig holds load test configuration
type LoadTestConfig struct {
	VirtualUsers    int           `json:"virtual_users"`
	Duration        time.Duration `json:"duration"`
	RampUpTime      time.Duration `json:"ramp_up_time"`
	RampDownTime    time.Duration `json:"ramp_down_time"`
	RequestsPerSecond int         `json:"requests_per_second"`
}

// StressTestConfig holds stress test configuration
type StressTestConfig struct {
	MaxVirtualUsers int           `json:"max_virtual_users"`
	Duration        time.Duration `json:"duration"`
	IncrementStep   int           `json:"increment_step"`
	IncrementInterval time.Duration `json:"increment_interval"`
}

// SpikeTestConfig holds spike test configuration
type SpikeTestConfig struct {
	BaseUsers       int           `json:"base_users"`
	SpikeUsers      int           `json:"spike_users"`
	SpikeDuration   time.Duration `json:"spike_duration"`
	RecoveryTime    time.Duration `json:"recovery_time"`
	SpikeCount      int           `json:"spike_count"`
}

// VolumeTestConfig holds volume test configuration
type VolumeTestConfig struct {
	DataSize        int64         `json:"data_size"`
	RecordCount     int           `json:"record_count"`
	ConcurrentUsers int           `json:"concurrent_users"`
	Duration        time.Duration `json:"duration"`
}

// PerformanceThresholds holds performance thresholds
type PerformanceThresholds struct {
	ResponseTime    time.Duration `json:"response_time"`
	Throughput      float64       `json:"throughput"`
	ErrorRate       float64       `json:"error_rate"`
	CPUUsage        float64       `json:"cpu_usage"`
	MemoryUsage     float64       `json:"memory_usage"`
	DiskUsage       float64       `json:"disk_usage"`
}

// SecurityTestConfig holds security test configuration
type SecurityTestConfig struct {
	Enabled              bool                  `json:"enabled"`
	VulnerabilityScanning bool                 `json:"vulnerability_scanning"`
	PenetrationTesting   bool                  `json:"penetration_testing"`
	AuthenticationTests  AuthSecurityTests     `json:"authentication_tests"`
	AuthorizationTests   AuthorizationTests    `json:"authorization_tests"`
	InputValidationTests InputValidationTests  `json:"input_validation_tests"`
	SQLInjectionTests    SQLInjectionTests     `json:"sql_injection_tests"`
	XSSTests             XSSTests              `json:"xss_tests"`
	CSRFTests            CSRFTests             `json:"csrf_tests"`
	SecurityHeaders      SecurityHeaderTests   `json:"security_headers"`
}

// AuthSecurityTests holds authentication security tests
type AuthSecurityTests struct {
	BruteForceProtection bool     `json:"brute_force_protection"`
	WeakPasswordDetection bool    `json:"weak_password_detection"`
	SessionManagement    bool     `json:"session_management"`
	TokenSecurity        bool     `json:"token_security"`
	TwoFactorAuth        bool     `json:"two_factor_auth"`
	TestCases            []string `json:"test_cases"`
}

// AuthorizationTests holds authorization tests
type AuthorizationTests struct {
	RoleBasedAccess      bool     `json:"role_based_access"`
	PermissionChecks     bool     `json:"permission_checks"`
	PrivilegeEscalation  bool     `json:"privilege_escalation"`
	ResourceAccess       bool     `json:"resource_access"`
	TestCases            []string `json:"test_cases"`
}

// InputValidationTests holds input validation tests
type InputValidationTests struct {
	DataTypeValidation   bool     `json:"data_type_validation"`
	LengthValidation     bool     `json:"length_validation"`
	FormatValidation     bool     `json:"format_validation"`
	RangeValidation      bool     `json:"range_validation"`
	SpecialCharacters    bool     `json:"special_characters"`
	TestCases            []string `json:"test_cases"`
}

// SQLInjectionTests holds SQL injection tests
type SQLInjectionTests struct {
	ClassicSQLInjection  bool     `json:"classic_sql_injection"`
	BlindSQLInjection    bool     `json:"blind_sql_injection"`
	TimeBased            bool     `json:"time_based"`
	UnionBased           bool     `json:"union_based"`
	TestCases            []string `json:"test_cases"`
}

// XSSTests holds XSS tests
type XSSTests struct {
	ReflectedXSS         bool     `json:"reflected_xss"`
	StoredXSS            bool     `json:"stored_xss"`
	DOMBasedXSS          bool     `json:"dom_based_xss"`
	TestCases            []string `json:"test_cases"`
}

// CSRFTests holds CSRF tests
type CSRFTests struct {
	TokenValidation      bool     `json:"token_validation"`
	SameSiteProtection   bool     `json:"same_site_protection"`
	RefererValidation    bool     `json:"referer_validation"`
	TestCases            []string `json:"test_cases"`
}

// SecurityHeaderTests holds security header tests
type SecurityHeaderTests struct {
	ContentSecurityPolicy bool     `json:"content_security_policy"`
	StrictTransportSecurity bool   `json:"strict_transport_security"`
	XFrameOptions          bool     `json:"x_frame_options"`
	XContentTypeOptions    bool     `json:"x_content_type_options"`
	TestCases              []string `json:"test_cases"`
}

// TestFilter represents a test filter
type TestFilter struct {
	Type      string      `json:"type"`      // "name", "tag", "duration", "status"
	Operator  string      `json:"operator"`  // "equals", "contains", "greater_than", "less_than"
	Value     interface{} `json:"value"`
	Enabled   bool        `json:"enabled"`
}

// TestSuite represents a collection of tests
type TestSuite struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Tests       []*TestCase            `json:"tests"`
	SetupFunc   func() error           `json:"-"`
	TeardownFunc func() error          `json:"-"`
	BeforeEach  func(*TestCase) error  `json:"-"`
	AfterEach   func(*TestCase) error  `json:"-"`
	Config      map[string]interface{} `json:"config"`
	Parallel    bool                   `json:"parallel"`
	Timeout     time.Duration          `json:"timeout"`
	Retries     int                    `json:"retries"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TestCase represents a single test case
type TestCase struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Type          TestType               `json:"type"`
	Priority      TestPriority           `json:"priority"`
	Tags          []string               `json:"tags"`
	Prerequisites []string               `json:"prerequisites"`
	TestFunc      func(*TestContext) error `json:"-"`
	SetupFunc     func(*TestContext) error `json:"-"`
	TeardownFunc  func(*TestContext) error `json:"-"`
	Data          map[string]interface{} `json:"data"`
	Expected      interface{}            `json:"expected"`
	Timeout       time.Duration          `json:"timeout"`
	Retries       int                    `json:"retries"`
	Skip          bool                   `json:"skip"`
	SkipReason    string                 `json:"skip_reason"`
	Status        TestStatus             `json:"status"`
	Result        *TestResult            `json:"result"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// TestType represents different test types
type TestType string

const (
	TestTypeUnit         TestType = "unit"
	TestTypeIntegration  TestType = "integration"
	TestTypeAPI          TestType = "api"
	TestTypeUI           TestType = "ui"
	TestTypePerformance  TestType = "performance"
	TestTypeSecurity     TestType = "security"
	TestTypeDatabase     TestType = "database"
	TestTypeEndToEnd     TestType = "e2e"
	TestTypeSmoke        TestType = "smoke"
	TestTypeRegression   TestType = "regression"
)

// TestPriority represents test priority levels
type TestPriority string

const (
	PriorityLow      TestPriority = "low"
	PriorityMedium   TestPriority = "medium"
	PriorityHigh     TestPriority = "high"
	PriorityCritical TestPriority = "critical"
)

// TestStatus represents test execution status
type TestStatus string

const (
	StatusPending  TestStatus = "pending"
	StatusRunning  TestStatus = "running"
	StatusPassed   TestStatus = "passed"
	StatusFailed   TestStatus = "failed"
	StatusSkipped  TestStatus = "skipped"
	StatusError    TestStatus = "error"
	StatusTimeout  TestStatus = "timeout"
)

// TestResult represents test execution result
type TestResult struct {
	Status        TestStatus             `json:"status"`
	Duration      time.Duration          `json:"duration"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       time.Time              `json:"end_time"`
	Error         string                 `json:"error,omitempty"`
	StackTrace    string                 `json:"stack_trace,omitempty"`
	Assertions    []AssertionResult      `json:"assertions"`
	Logs          []TestLog              `json:"logs"`
	Screenshots   []string               `json:"screenshots,omitempty"`
	Videos        []string               `json:"videos,omitempty"`
	Metrics       map[string]interface{} `json:"metrics"`
	Coverage      *CoverageData          `json:"coverage,omitempty"`
	Artifacts     []TestArtifact         `json:"artifacts"`
	RetryAttempt  int                    `json:"retry_attempt"`
}

// AssertionResult represents assertion result
type AssertionResult struct {
	Name      string      `json:"name"`
	Expected  interface{} `json:"expected"`
	Actual    interface{} `json:"actual"`
	Passed    bool        `json:"passed"`
	Message   string      `json:"message"`
	Location  string      `json:"location"`
	Timestamp time.Time   `json:"timestamp"`
}

// TestLog represents a test log entry
type TestLog struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Context   map[string]interface{} `json:"context"`
}

// TestArtifact represents a test artifact
type TestArtifact struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
}

// CoverageData represents code coverage data
type CoverageData struct {
	LineCoverage       float64            `json:"line_coverage"`
	BranchCoverage     float64            `json:"branch_coverage"`
	FunctionCoverage   float64            `json:"function_coverage"`
	StatementCoverage  float64            `json:"statement_coverage"`
	CoveredLines       []int              `json:"covered_lines"`
	UncoveredLines     []int              `json:"uncovered_lines"`
	FilesCoverage      map[string]float64 `json:"files_coverage"`
	PackagesCoverage   map[string]float64 `json:"packages_coverage"`
}

// TestContext provides context for test execution
type TestContext struct {
	TestCase    *TestCase
	Config      *TestConfig
	DB          *sql.DB
	Logger      TestLogger
	Assertions  *AssertionHelper
	HTTP        *HTTPTestHelper
	UI          *UITestHelper
	Performance *PerformanceHelper
	Security    *SecurityHelper
	Data        map[string]interface{}
}

// TestLogger interface for test logging
type TestLogger interface {
	Debug(message string, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message string, args ...interface{})
	Fatal(message string, args ...interface{})
}

// AssertionHelper provides assertion methods
type AssertionHelper struct {
	t       *testing.T
	results []AssertionResult
}

// HTTPClient represents an HTTP client for testing
type HTTPClient struct {
	BaseURL string
	Timeout time.Duration
}

// WebDriver represents a web driver for UI testing  
type WebDriver struct {
	BrowserType string
	Headless    bool
}

// PerformanceMetrics holds performance test metrics
type PerformanceMetrics struct {
	ResponseTime time.Duration
	Throughput   float64
	ErrorRate    float64
}

// MetricsCollector collects performance metrics
type MetricsCollector struct {
	Metrics []PerformanceMetrics
}

// VulnerabilityScanner scans for security vulnerabilities
type VulnerabilityScanner struct {
	ScanTypes []string
}

// SecurityAnalyzer analyzes security test results
type SecurityAnalyzer struct {
	Rules []string
}

// HTTPTestHelper provides HTTP testing utilities
type HTTPTestHelper struct {
	client  *HTTPClient
	config  *APITestConfig
	baseURL string
}

// UITestHelper provides UI testing utilities
type UITestHelper struct {
	driver WebDriver
	config *UITestConfig
}

// PerformanceHelper provides performance testing utilities
type PerformanceHelper struct {
	config    *PerformanceTestConfig
	metrics   *PerformanceMetrics
	collector *MetricsCollector
}

// SecurityHelper provides security testing utilities
type SecurityHelper struct {
	config   *SecurityTestConfig
	scanner  VulnerabilityScanner
	analyzer SecurityAnalyzer
}

// TestRunner interface for different test runners
type TestRunner interface {
	Run(ctx context.Context, suite *TestSuite) (*TestSuiteResult, error)
	GetName() string
	GetSupportedTypes() []TestType
	Configure(config map[string]interface{}) error
}

// TestReporter interface for test reporting
type TestReporter interface {
	GenerateReport(results []*TestSuiteResult) error
	GetFormat() string
	SetOutputPath(path string)
}

// TestHooks provides hooks for test lifecycle events
type TestHooks struct {
	BeforeAll  func() error
	AfterAll   func() error
	BeforeSuite func(*TestSuite) error
	AfterSuite  func(*TestSuite, *TestSuiteResult) error
	BeforeTest  func(*TestCase) error
	AfterTest   func(*TestCase, *TestResult) error
}

// TestSuiteResult represents the result of a test suite execution
type TestSuiteResult struct {
	SuiteID       string        `json:"suite_id"`
	SuiteName     string        `json:"suite_name"`
	Status        TestStatus    `json:"status"`
	Duration      time.Duration `json:"duration"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	TestResults   []*TestResult `json:"test_results"`
	TotalTests    int           `json:"total_tests"`
	PassedTests   int           `json:"passed_tests"`
	FailedTests   int           `json:"failed_tests"`
	SkippedTests  int           `json:"skipped_tests"`
	ErrorTests    int           `json:"error_tests"`
	Coverage      *CoverageData `json:"coverage,omitempty"`
	Metrics       map[string]interface{} `json:"metrics"`
	Artifacts     []TestArtifact `json:"artifacts"`
}

// CoverageCollector collects code coverage data
type CoverageCollector struct {
	enabled     bool
	threshold   float64
	files       map[string]*FileCoverage
	packages    map[string]*PackageCoverage
	totalLines  int
	coveredLines int
}

// FileCoverage represents coverage data for a file
type FileCoverage struct {
	Path           string    `json:"path"`
	Lines          int       `json:"lines"`
	CoveredLines   int       `json:"covered_lines"`
	Coverage       float64   `json:"coverage"`
	Functions      int       `json:"functions"`
	CoveredFuncs   int       `json:"covered_functions"`
	Branches       int       `json:"branches"`
	CoveredBranches int      `json:"covered_branches"`
	Statements     int       `json:"statements"`
	CoveredStmts   int       `json:"covered_statements"`
}

// PackageCoverage represents coverage data for a package
type PackageCoverage struct {
	Name         string         `json:"name"`
	Files        []*FileCoverage `json:"files"`
	Coverage     float64        `json:"coverage"`
	TotalLines   int            `json:"total_lines"`
	CoveredLines int            `json:"covered_lines"`
}

// NewTestManager creates a new test manager
func NewTestManager(db *sql.DB, config TestConfig) *TestManager {
	tm := &TestManager{
		db:        db,
		config:    config,
		suites:    make(map[string]*TestSuite),
		runners:   make(map[string]TestRunner),
		reporters: make([]TestReporter, 0),
		coverage:  NewCoverageCollector(config.CoverageEnabled, config.CoverageThreshold),
	}

	tm.createTestTables()
	tm.initializeRunners()
	tm.initializeReporters()

	return tm
}

// createTestTables creates necessary tables for test management
func (tm *TestManager) createTestTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS test_suites (
			id VARCHAR(128) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			tags TEXT,
			config TEXT,
			parallel BOOLEAN DEFAULT FALSE,
			timeout_seconds INT DEFAULT 300,
			retries INT DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_name (name),
			INDEX idx_created_at (created_at)
		)`,
		`CREATE TABLE IF NOT EXISTS test_cases (
			id VARCHAR(128) PRIMARY KEY,
			suite_id VARCHAR(128) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			type VARCHAR(50) NOT NULL,
			priority VARCHAR(20) DEFAULT 'medium',
			tags TEXT,
			prerequisites TEXT,
			data TEXT,
			expected TEXT,
			timeout_seconds INT DEFAULT 60,
			retries INT DEFAULT 0,
			skip_test BOOLEAN DEFAULT FALSE,
			skip_reason TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_suite_id (suite_id),
			INDEX idx_type (type),
			INDEX idx_priority (priority),
			FOREIGN KEY (suite_id) REFERENCES test_suites(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS test_executions (
			id VARCHAR(128) PRIMARY KEY,
			suite_id VARCHAR(128) NOT NULL,
			case_id VARCHAR(128) NOT NULL,
			status VARCHAR(20) NOT NULL,
			duration_ms INT DEFAULT 0,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			error_message TEXT,
			stack_trace TEXT,
			assertions TEXT,
			logs TEXT,
			screenshots TEXT,
			videos TEXT,
			metrics TEXT,
			coverage TEXT,
			artifacts TEXT,
			retry_attempt INT DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_suite_id (suite_id),
			INDEX idx_case_id (case_id),
			INDEX idx_status (status),
			INDEX idx_start_time (start_time),
			FOREIGN KEY (suite_id) REFERENCES test_suites(id) ON DELETE CASCADE,
			FOREIGN KEY (case_id) REFERENCES test_cases(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS test_coverage (
			id VARCHAR(128) PRIMARY KEY,
			execution_id VARCHAR(128) NOT NULL,
			file_path VARCHAR(512) NOT NULL,
			total_lines INT DEFAULT 0,
			covered_lines INT DEFAULT 0,
			line_coverage DECIMAL(5,2) DEFAULT 0.00,
			branch_coverage DECIMAL(5,2) DEFAULT 0.00,
			function_coverage DECIMAL(5,2) DEFAULT 0.00,
			statement_coverage DECIMAL(5,2) DEFAULT 0.00,
			covered_line_numbers TEXT,
			uncovered_line_numbers TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_execution_id (execution_id),
			INDEX idx_file_path (file_path),
			FOREIGN KEY (execution_id) REFERENCES test_executions(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS test_metrics (
			id VARCHAR(128) PRIMARY KEY,
			execution_id VARCHAR(128) NOT NULL,
			metric_name VARCHAR(100) NOT NULL,
			metric_value DECIMAL(10,4) NOT NULL,
			metric_unit VARCHAR(20),
			threshold_value DECIMAL(10,4),
			passed BOOLEAN DEFAULT TRUE,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_execution_id (execution_id),
			INDEX idx_metric_name (metric_name),
			INDEX idx_timestamp (timestamp),
			FOREIGN KEY (execution_id) REFERENCES test_executions(id) ON DELETE CASCADE
		)`,
	}

	for _, query := range queries {
		if _, err := tm.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create test table: %w", err)
		}
	}

	return nil
}

// RegisterSuite registers a test suite
func (tm *TestManager) RegisterSuite(suite *TestSuite) {
	suite.ID = tm.generateSuiteID(suite.Name)
	suite.CreatedAt = time.Now()
	suite.UpdatedAt = time.Now()
	tm.suites[suite.ID] = suite
}

// RegisterRunner registers a test runner
func (tm *TestManager) RegisterRunner(runner TestRunner) {
	tm.runners[runner.GetName()] = runner
}

// RegisterReporter registers a test reporter
func (tm *TestManager) RegisterReporter(reporter TestReporter) {
	tm.reporters = append(tm.reporters, reporter)
}

// RunSuite runs a specific test suite
func (tm *TestManager) RunSuite(ctx context.Context, suiteID string) (*TestSuiteResult, error) {
	suite, exists := tm.suites[suiteID]
	if !exists {
		return nil, fmt.Errorf("test suite not found: %s", suiteID)
	}

	// Find appropriate runner
	runner := tm.findRunner(suite)
	if runner == nil {
		return nil, fmt.Errorf("no suitable runner found for suite: %s", suite.Name)
	}

	// Execute hooks
	if tm.hooks.BeforeSuite != nil {
		if err := tm.hooks.BeforeSuite(suite); err != nil {
			return nil, fmt.Errorf("before suite hook failed: %w", err)
		}
	}

	// Run the suite
	result, err := runner.Run(ctx, suite)
	if err != nil {
		return nil, fmt.Errorf("suite execution failed: %w", err)
	}

	// Execute hooks
	if tm.hooks.AfterSuite != nil {
		tm.hooks.AfterSuite(suite, result)
	}

	// Store results
	tm.storeTestResults(result)

	return result, nil
}

// RunAllSuites runs all registered test suites
func (tm *TestManager) RunAllSuites(ctx context.Context) ([]*TestSuiteResult, error) {
	results := make([]*TestSuiteResult, 0)

	// Execute hooks
	if tm.hooks.BeforeAll != nil {
		if err := tm.hooks.BeforeAll(); err != nil {
			return nil, fmt.Errorf("before all hook failed: %w", err)
		}
	}

	// Run suites
	for _, suite := range tm.suites {
		if tm.shouldRunSuite(suite) {
			result, err := tm.RunSuite(ctx, suite.ID)
			if err != nil {
				tm.logError(fmt.Sprintf("Suite %s failed: %v", suite.Name, err))
				continue
			}
			results = append(results, result)
		}
	}

	// Execute hooks
	if tm.hooks.AfterAll != nil {
		tm.hooks.AfterAll()
	}

	// Generate reports
	tm.generateReports(results)

	return results, nil
}

// RunTestsByTag runs tests filtered by tags
func (tm *TestManager) RunTestsByTag(ctx context.Context, tags []string) ([]*TestSuiteResult, error) {
	results := make([]*TestSuiteResult, 0)

	for _, suite := range tm.suites {
		if tm.hasMatchingTags(suite.Tags, tags) {
			result, err := tm.RunSuite(ctx, suite.ID)
			if err != nil {
				continue
			}
			results = append(results, result)
		}
	}

	return results, nil
}

// GetTestResults retrieves test results
func (tm *TestManager) GetTestResults(suiteID string, limit, offset int) ([]*TestResult, error) {
	query := `
		SELECT id, case_id, status, duration_ms, start_time, end_time, 
		       error_message, stack_trace, assertions, logs, screenshots,
		       videos, metrics, coverage, artifacts, retry_attempt
		FROM test_executions 
		WHERE suite_id = ?
		ORDER BY start_time DESC
		LIMIT ? OFFSET ?
	`

	rows, err := tm.db.Query(query, suiteID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]*TestResult, 0)
	for rows.Next() {
		result, err := tm.scanTestResult(rows)
		if err != nil {
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// GetTestStats retrieves test statistics
func (tm *TestManager) GetTestStats(startDate, endDate time.Time) (*TestStats, error) {
	stats := &TestStats{
		ByType:     make(map[TestType]int),
		ByPriority: make(map[TestPriority]int),
		ByStatus:   make(map[TestStatus]int),
	}

	// Get basic counts
	query := `
		SELECT 
		    COUNT(*) as total_tests,
		    SUM(CASE WHEN status = 'passed' THEN 1 ELSE 0 END) as passed_tests,
		    SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_tests,
		    SUM(CASE WHEN status = 'skipped' THEN 1 ELSE 0 END) as skipped_tests,
		    SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END) as error_tests,
		    AVG(duration_ms) as avg_duration
		FROM test_executions te
		JOIN test_cases tc ON te.case_id = tc.id
		WHERE te.start_time BETWEEN ? AND ?
	`

	err := tm.db.QueryRow(query, startDate, endDate).Scan(
		&stats.TotalTests, &stats.PassedTests, &stats.FailedTests,
		&stats.SkippedTests, &stats.ErrorTests, &stats.AvgDuration,
	)
	if err != nil {
		return nil, err
	}

	// Calculate success rate
	if stats.TotalTests > 0 {
		stats.SuccessRate = float64(stats.PassedTests) / float64(stats.TotalTests)
	}

	return stats, nil
}

// TestStats represents test statistics
type TestStats struct {
	TotalTests    int                        `json:"total_tests"`
	PassedTests   int                        `json:"passed_tests"`
	FailedTests   int                        `json:"failed_tests"`
	SkippedTests  int                        `json:"skipped_tests"`
	ErrorTests    int                        `json:"error_tests"`
	SuccessRate   float64                    `json:"success_rate"`
	AvgDuration   time.Duration              `json:"avg_duration"`
	ByType        map[TestType]int           `json:"by_type"`
	ByPriority    map[TestPriority]int       `json:"by_priority"`
	ByStatus      map[TestStatus]int         `json:"by_status"`
	TrendData     []TestTrendPoint           `json:"trend_data"`
}

// TestTrendPoint represents a point in test trend data
type TestTrendPoint struct {
	Date      time.Time `json:"date"`
	Passed    int       `json:"passed"`
	Failed    int       `json:"failed"`
	Total     int       `json:"total"`
	Duration  time.Duration `json:"duration"`
}

// Helper methods

func (tm *TestManager) generateSuiteID(name string) string {
	return fmt.Sprintf("suite_%s_%d", strings.ReplaceAll(name, " ", "_"), time.Now().Unix())
}

func (tm *TestManager) findRunner(suite *TestSuite) TestRunner {
	// Find the most appropriate runner based on test types in the suite
	testTypes := tm.getTestTypes(suite)
	
	for _, runner := range tm.runners {
		supportedTypes := runner.GetSupportedTypes()
		if tm.supportsAllTypes(supportedTypes, testTypes) {
			return runner
		}
	}
	
	return nil
}

func (tm *TestManager) getTestTypes(suite *TestSuite) []TestType {
	types := make(map[TestType]bool)
	for _, test := range suite.Tests {
		types[test.Type] = true
	}
	
	result := make([]TestType, 0, len(types))
	for t := range types {
		result = append(result, t)
	}
	
	return result
}

func (tm *TestManager) supportsAllTypes(supported, required []TestType) bool {
	supportedMap := make(map[TestType]bool)
	for _, t := range supported {
		supportedMap[t] = true
	}
	
	for _, t := range required {
		if !supportedMap[t] {
			return false
		}
	}
	
	return true
}

func (tm *TestManager) shouldRunSuite(suite *TestSuite) bool {
	// Apply filters
	for _, filter := range tm.config.Filters {
		if !tm.applyFilter(suite, filter) {
			return false
		}
	}
	
	// Check tags
	if len(tm.config.Tags) > 0 {
		if !tm.hasMatchingTags(suite.Tags, tm.config.Tags) {
			return false
		}
	}
	
	// Check exclude tags
	if len(tm.config.ExcludeTags) > 0 {
		if tm.hasMatchingTags(suite.Tags, tm.config.ExcludeTags) {
			return false
		}
	}
	
	return true
}

func (tm *TestManager) hasMatchingTags(suiteTags, filterTags []string) bool {
	suiteTagMap := make(map[string]bool)
	for _, tag := range suiteTags {
		suiteTagMap[tag] = true
	}
	
	for _, tag := range filterTags {
		if suiteTagMap[tag] {
			return true
		}
	}
	
	return false
}

func (tm *TestManager) applyFilter(suite *TestSuite, filter TestFilter) bool {
	if !filter.Enabled {
		return true
	}
	
	switch filter.Type {
	case "name":
		return tm.applyStringFilter(suite.Name, filter.Operator, filter.Value.(string))
	case "tag":
		return tm.hasMatchingTags(suite.Tags, []string{filter.Value.(string)})
	default:
		return true
	}
}

func (tm *TestManager) applyStringFilter(value, operator, filterValue string) bool {
	switch operator {
	case "equals":
		return value == filterValue
	case "contains":
		return strings.Contains(value, filterValue)
	default:
		return true
	}
}

func (tm *TestManager) storeTestResults(result *TestSuiteResult) {
	// Store suite result and individual test results
	// Implementation would save to database
}

func (tm *TestManager) scanTestResult(scanner interface{}) (*TestResult, error) {
	// Implementation would scan database row to TestResult
	return &TestResult{}, nil
}

func (tm *TestManager) generateReports(results []*TestSuiteResult) {
	for _, reporter := range tm.reporters {
		if err := reporter.GenerateReport(results); err != nil {
			tm.logError(fmt.Sprintf("Failed to generate %s report: %v", reporter.GetFormat(), err))
		}
	}
}

func (tm *TestManager) initializeRunners() {
	// Initialize default test runners
}

func (tm *TestManager) initializeReporters() {
	// Initialize default test reporters
}

func (tm *TestManager) logError(message string) {
	fmt.Printf("Test Manager Error: %s\n", message)
}

// NewCoverageCollector creates a new coverage collector
func NewCoverageCollector(enabled bool, threshold float64) *CoverageCollector {
	return &CoverageCollector{
		enabled:   enabled,
		threshold: threshold,
		files:     make(map[string]*FileCoverage),
		packages:  make(map[string]*PackageCoverage),
	}
}

// Assertion helper methods

func (ah *AssertionHelper) Equal(expected, actual interface{}, message string) bool {
	passed := reflect.DeepEqual(expected, actual)
	
	result := AssertionResult{
		Name:      "Equal",
		Expected:  expected,
		Actual:    actual,
		Passed:    passed,
		Message:   message,
		Location:  ah.getLocation(),
		Timestamp: time.Now(),
	}
	
	ah.results = append(ah.results, result)
	
	if !passed && ah.t != nil {
		ah.t.Errorf("Assertion failed: %s. Expected %v, got %v", message, expected, actual)
	}
	
	return passed
}

func (ah *AssertionHelper) NotEqual(expected, actual interface{}, message string) bool {
	passed := !reflect.DeepEqual(expected, actual)
	
	result := AssertionResult{
		Name:      "NotEqual",
		Expected:  expected,
		Actual:    actual,
		Passed:    passed,
		Message:   message,
		Location:  ah.getLocation(),
		Timestamp: time.Now(),
	}
	
	ah.results = append(ah.results, result)
	
	if !passed && ah.t != nil {
		ah.t.Errorf("Assertion failed: %s. Expected not %v, got %v", message, expected, actual)
	}
	
	return passed
}

func (ah *AssertionHelper) True(value bool, message string) bool {
	return ah.Equal(true, value, message)
}

func (ah *AssertionHelper) False(value bool, message string) bool {
	return ah.Equal(false, value, message)
}

func (ah *AssertionHelper) Nil(value interface{}, message string) bool {
	return ah.Equal(nil, value, message)
}

func (ah *AssertionHelper) NotNil(value interface{}, message string) bool {
	return ah.NotEqual(nil, value, message)
}

func (ah *AssertionHelper) getLocation() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func (ah *AssertionHelper) GetResults() []AssertionResult {
	return ah.results
}