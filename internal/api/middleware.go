package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"kolajAi/internal/cache"
	"kolajAi/internal/errors"
	"kolajAi/internal/security"
	"kolajAi/internal/session"
)

// APIMiddleware provides comprehensive API middleware stack
type APIMiddleware struct {
	SecurityManager *security.SecurityManager
	SessionManager  *session.SessionManager
	ErrorManager    *errors.ErrorManager
	CacheManager    *cache.CacheManager
	Config          *APIConfig
}

// APIConfig holds API configuration
type APIConfig struct {
	Version           string        `json:"version"`
	RateLimitEnabled  bool          `json:"rate_limit_enabled"`
	CacheEnabled      bool          `json:"cache_enabled"`
	CompressionEnabled bool         `json:"compression_enabled"`
	CORSEnabled       bool          `json:"cors_enabled"`
	RequestTimeout    time.Duration `json:"request_timeout"`
	MaxRequestSize    int64         `json:"max_request_size"`
	AllowedOrigins    []string      `json:"allowed_origins"`
	AllowedMethods    []string      `json:"allowed_methods"`
	AllowedHeaders    []string      `json:"allowed_headers"`
	ExposedHeaders    []string      `json:"exposed_headers"`
	AllowCredentials  bool          `json:"allow_credentials"`
	MaxAge            int           `json:"max_age"`
}

// APIResponse represents standardized API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Meta      *APIMeta    `json:"meta,omitempty"`
	Timestamp string      `json:"timestamp"`
	RequestID string      `json:"request_id"`
	Version   string      `json:"version"`
}

// APIError represents API error details
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Field   string `json:"field,omitempty"`
}

// APIMeta represents API response metadata
type APIMeta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// NewAPIMiddleware creates new API middleware
func NewAPIMiddleware(
	securityManager *security.SecurityManager,
	sessionManager *session.SessionManager,
	errorManager *errors.ErrorManager,
	cacheManager *cache.CacheManager,
	config *APIConfig,
) *APIMiddleware {
	return &APIMiddleware{
		SecurityManager: securityManager,
		SessionManager:  sessionManager,
		ErrorManager:    errorManager,
		CacheManager:    cacheManager,
		Config:          config,
	}
}

// APIHandler wraps HTTP handlers with API middleware
func (m *APIMiddleware) APIHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Generate request ID
		requestID := generateRequestID()
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		r = r.WithContext(ctx)

		// Set standard headers
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("X-API-Version", m.Config.Version)

		// Apply middleware chain
		handler := m.applyMiddlewareChain(next)
		handler(w, r)
	}
}

// applyMiddlewareChain applies all middleware in order
func (m *APIMiddleware) applyMiddlewareChain(next http.HandlerFunc) http.HandlerFunc {
	// Apply middleware in reverse order
	handler := next

	// Recovery middleware (last)
	handler = m.recoveryMiddleware(handler)

	// Error handling middleware
	handler = m.errorHandlingMiddleware(handler)

	// Request timeout middleware
	handler = m.timeoutMiddleware(handler)

	// Request size limiting middleware
	handler = m.requestSizeLimitMiddleware(handler)

	// Compression middleware
	if m.Config.CompressionEnabled {
		handler = m.compressionMiddleware(handler)
	}

	// Cache middleware
	if m.Config.CacheEnabled {
		handler = m.cacheMiddleware(handler)
	}

	// Rate limiting middleware
	if m.Config.RateLimitEnabled {
		handler = m.rateLimitMiddleware(handler)
	}

	// Security middleware
	handler = m.securityMiddleware(handler)

	// CORS middleware
	if m.Config.CORSEnabled {
		handler = m.corsMiddleware(handler)
	}

	// Logging middleware (first)
	handler = m.loggingMiddleware(handler)

	return handler
}

// loggingMiddleware logs API requests
func (m *APIMiddleware) loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create response writer wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next(rw, r)
		
		duration := time.Since(start)
		requestID := r.Context().Value("request_id").(string)
		
		log.Printf("API Request: %s %s | Status: %d | Duration: %v | RequestID: %s | IP: %s | UserAgent: %s",
			r.Method, r.URL.Path, rw.statusCode, duration, requestID, 
			getClientIP(r), r.UserAgent())
	}
}

// corsMiddleware handles CORS
func (m *APIMiddleware) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Check if origin is allowed
		if m.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(m.Config.AllowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(m.Config.AllowedHeaders, ", "))
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(m.Config.ExposedHeaders, ", "))
		w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", m.Config.MaxAge))
		
		if m.Config.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		next(w, r)
	}
}

// securityMiddleware applies security checks
func (m *APIMiddleware) securityMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Apply security headers
		m.SecurityManager.SetSecurityHeaders(w)
		
		// Check IP access
		if blocked, reason := m.SecurityManager.CheckIPAccess(getClientIP(r)); blocked {
			m.sendErrorResponse(w, r, http.StatusForbidden, "IP_BLOCKED", 
				fmt.Sprintf("Access denied: %s", reason))
			return
		}
		
		// Input validation for POST/PUT/PATCH requests
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			if err := m.SecurityManager.ValidateInput(r); err != nil {
				m.sendErrorResponse(w, r, http.StatusBadRequest, "INVALID_INPUT", 
					"Request validation failed")
				return
			}
		}
		
		// Vulnerability scanning
		if threats := m.SecurityManager.ScanForThreats(r); len(threats) > 0 {
			m.sendErrorResponse(w, r, http.StatusBadRequest, "SECURITY_THREAT", 
				"Security threat detected")
			return
		}
		
		next(w, r)
	}
}

// rateLimitMiddleware applies rate limiting
func (m *APIMiddleware) rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if limited, _ := m.SecurityManager.CheckRateLimit(r); limited {
			m.sendErrorResponse(w, r, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", 
				"Rate limit exceeded")
			return
		}
		
		next(w, r)
	}
}

// cacheMiddleware implements API response caching
func (m *APIMiddleware) cacheMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only cache GET requests
		if r.Method != "GET" {
			next(w, r)
			return
		}
		
		cacheKey := fmt.Sprintf("api_cache:%s:%s", r.Method, r.URL.String())
		
		// Try to get from cache
		if cached, found := m.CacheManager.Get(cacheKey); found {
			w.Header().Set("X-Cache", "HIT")
			w.Write(cached.([]byte))
			return
		}
		
		// Create response recorder
		recorder := &responseRecorder{ResponseWriter: w, body: make([]byte, 0)}
		
		next(recorder, r)
		
		// Cache successful responses
		if recorder.statusCode >= 200 && recorder.statusCode < 300 {
			m.CacheManager.Set(cacheKey, recorder.body, 5*time.Minute)
			w.Header().Set("X-Cache", "MISS")
		}
		
		w.Write(recorder.body)
	}
}

// compressionMiddleware implements response compression
func (m *APIMiddleware) compressionMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next(w, r)
			return
		}
		
		// Implementation would go here - using gzip compression
		next(w, r)
	}
}

// requestSizeLimitMiddleware limits request size
func (m *APIMiddleware) requestSizeLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > m.Config.MaxRequestSize {
			m.sendErrorResponse(w, r, http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", 
				"Request entity too large")
			return
		}
		
		r.Body = http.MaxBytesReader(w, r.Body, m.Config.MaxRequestSize)
		next(w, r)
	}
}

// timeoutMiddleware implements request timeout
func (m *APIMiddleware) timeoutMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), m.Config.RequestTimeout)
		defer cancel()
		
		r = r.WithContext(ctx)
		next(w, r)
	}
}

// errorHandlingMiddleware handles API errors
func (m *APIMiddleware) errorHandlingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.sendErrorResponse(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", 
					"Internal server error")
			}
		}()
		
		next(w, r)
	}
}

// recoveryMiddleware recovers from panics
func (m *APIMiddleware) recoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("API Panic: %v\n%s", err, debug.Stack())
				m.sendErrorResponse(w, r, http.StatusInternalServerError, "PANIC_RECOVERED", 
					"Server error recovered")
			}
		}()
		
		next(w, r)
	}
}

// Helper methods

// sendErrorResponse sends standardized error response
func (m *APIMiddleware) sendErrorResponse(w http.ResponseWriter, r *http.Request, 
	statusCode int, code string, message string) {
	
	requestID := r.Context().Value("request_id").(string)
	
	response := APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: requestID,
		Version:   m.Config.Version,
	}
	
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// SendSuccessResponse sends standardized success response
func (m *APIMiddleware) SendSuccessResponse(w http.ResponseWriter, r *http.Request, 
	data interface{}, meta *APIMeta) {
	
	requestID := r.Context().Value("request_id").(string)
	
	response := APIResponse{
		Success:   true,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: requestID,
		Version:   m.Config.Version,
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// isOriginAllowed checks if origin is allowed
func (m *APIMiddleware) isOriginAllowed(origin string) bool {
	for _, allowed := range m.Config.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// generateRequestID generates unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// getClientIP gets client IP address
func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}
	return r.RemoteAddr
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// responseRecorder records response for caching
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
}

func (rr *responseRecorder) Write(data []byte) (int, error) {
	rr.body = append(rr.body, data...)
	return len(data), nil
}