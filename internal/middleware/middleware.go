package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"kolajAi/internal/session"
	"kolajAi/internal/errors"
	"kolajAi/internal/security"
	"kolajAi/internal/cache"
	"compress/gzip"
)

// MiddlewareStack holds all middleware components
type MiddlewareStack struct {
	SecurityManager *security.SecurityManager
	SessionManager  *session.SessionManager
	ErrorManager    *errors.ErrorManager
	CacheManager    *cache.CacheManager
}

// NewMiddlewareStack creates a new middleware stack
func NewMiddlewareStack(
	securityManager *security.SecurityManager,
	sessionManager *session.SessionManager,
	errorManager *errors.ErrorManager,
	cacheManager *cache.CacheManager,
) *MiddlewareStack {
	return &MiddlewareStack{
		SecurityManager: securityManager,
		SessionManager:  sessionManager,
		ErrorManager:    errorManager,
		CacheManager:    cacheManager,
	}
}

// SecurityMiddleware applies security measures to all requests
func (ms *MiddlewareStack) SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set security headers
		ms.SecurityManager.SetSecurityHeaders(w)
		
		// Check IP whitelist/blacklist
		if blocked, reason := ms.SecurityManager.CheckIPAccess(r.RemoteAddr); blocked {
			ms.ErrorManager.HandleHTTPError(w, r, errors.NewApplicationError(
				errors.FORBIDDEN,
				"IP_BLOCKED",
				fmt.Sprintf("Access denied: %s", reason),
				nil,
			))
			return
		}
		
		// Rate limiting
		if limited, err := ms.SecurityManager.CheckRateLimit(r); limited {
			ms.ErrorManager.HandleHTTPError(w, r, errors.NewApplicationError(
				errors.RATE_LIMITED,
				"RATE_LIMIT_EXCEEDED",
				"Rate limit exceeded",
				err,
			))
			return
		}
		
		// Input validation for POST/PUT requests
		if r.Method == "POST" || r.Method == "PUT" {
			if err := ms.SecurityManager.ValidateInput(r); err != nil {
				ms.ErrorManager.HandleHTTPError(w, r, errors.NewApplicationError(
					errors.VALIDATION,
					"INPUT_VALIDATION_FAILED",
					"Input validation failed",
					err,
				))
				return
			}
		}
		
		// Vulnerability scanning
		if threats := ms.SecurityManager.ScanForThreats(r); len(threats) > 0 {
			// Log security event
			for _, threat := range threats {
				ms.SecurityManager.LogSecurityEvent(security.SecurityEvent{
					Type:      security.THREAT_DETECTED,
					Severity:  security.HIGH,
					Source:    r.RemoteAddr,
					UserAgent: r.UserAgent(),
					Method:    r.Method,
					URL:       r.URL.String(),
					Details:   map[string]interface{}{"threat": threat},
					Timestamp: time.Now(),
					Blocked:   true,
				})
			}
			
			ms.ErrorManager.HandleHTTPError(w, r, errors.NewApplicationError(
				errors.SECURITY_VIOLATION,
				"SECURITY_THREAT_DETECTED",
				"Security threat detected",
				fmt.Errorf("threats: %v", threats),
			))
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// SessionMiddleware handles session management
func (ms *MiddlewareStack) SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Load session
		sessionData, err := ms.SessionManager.GetSession(r)
		if err != nil && err != session.ErrSessionNotFound {
			log.Printf("Session error: %v", err)
		}
		
		// Add session to request context
		ctx := context.WithValue(r.Context(), "session", sessionData)
		r = r.WithContext(ctx)
		
		// Update session activity
		if sessionData != nil {
			sessionData.LastActivity = time.Now()
			if err := ms.SessionManager.UpdateSessionData(sessionData); err != nil {
				log.Printf("Failed to update session activity: %v", err)
			}
		}
		
		next.ServeHTTP(w, r)
	})
}

// ErrorHandlingMiddleware handles panics and errors
func (ms *MiddlewareStack) ErrorHandlingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				stack := debug.Stack()
				appError := errors.NewApplicationError(
					errors.INTERNAL,
					"PANIC",
					"Internal server error",
					fmt.Errorf("panic: %v", err),
				)
				appError.Context.StackTrace = string(stack)
				
				ms.ErrorManager.LogError(appError)
				ms.ErrorManager.HandleHTTPError(w, r, appError)
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

// CacheMiddleware handles caching for GET requests
func (ms *MiddlewareStack) CacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only cache GET requests
		if r.Method != "GET" {
			next.ServeHTTP(w, r)
			return
		}
		
		// Skip caching for admin and API endpoints
		if strings.HasPrefix(r.URL.Path, "/admin") || 
		   strings.HasPrefix(r.URL.Path, "/api") ||
		   strings.Contains(r.URL.Path, "login") ||
		   strings.Contains(r.URL.Path, "logout") {
			next.ServeHTTP(w, r)
			return
		}
		
		// Generate cache key
		userID := getUserIDFromSession(r)
		userIDStr := ""
		if userID != nil {
			userIDStr = fmt.Sprintf("%v", userID)
		}
		cacheKey := ms.CacheManager.BuildKey(cache.CacheKey{
			Type:   "page",
			ID:     r.URL.Path,
			Params: map[string]string{
				"query":   r.URL.RawQuery,
				"user_id": userIDStr,
			},
		})
		
		// Try to get from cache
		if cached, err := ms.CacheManager.Get(r.Context(), "default", cacheKey); err == nil && cached != nil {
			// Serve from cache
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("X-Cache", "HIT")
			w.Write(cached)
			return
		}
		
		// Create response writer wrapper to capture response
		rw := &responseWriter{
			ResponseWriter: w,
			body:          make([]byte, 0),
		}
		
		next.ServeHTTP(rw, r)
		
		// Cache successful responses
		if rw.statusCode == 0 || rw.statusCode == 200 {
			// Cache for 30 minutes
			ms.CacheManager.Set(r.Context(), "default", cacheKey, rw.body, 30*time.Minute)
			w.Header().Set("X-Cache", "MISS")
		}
	})
}

// LoggingMiddleware logs all requests
func (ms *MiddlewareStack) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create response writer wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w}
		
		next.ServeHTTP(rw, r)
		
		duration := time.Since(start)
		
		// Log request details
		log.Printf(
			"%s %s %s %d %v %s %s",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			rw.statusCode,
			duration,
			r.UserAgent(),
			r.Referer(),
		)
		
		// Log slow requests
		if duration > 5*time.Second {
			ms.ErrorManager.LogError(errors.NewApplicationError(
				errors.WARNING,
				"SLOW_REQUEST",
				fmt.Sprintf("Slow request: %s %s took %v", r.Method, r.URL.Path, duration),
				nil,
			))
		}
	})
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func (ms *MiddlewareStack) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Set CORS headers
		if origin != "" {
			// Check if origin is allowed
			allowed := false
			allowedOrigins := []string{
				"http://localhost:3000",
				"http://localhost:8081",
				"https://kolajAi.com",
			}
			
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}
			
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// CompressionMiddleware compresses responses
func (ms *MiddlewareStack) CompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		
		// Skip compression for small responses and certain content types
		if strings.HasPrefix(r.URL.Path, "/static/") && 
		   (strings.HasSuffix(r.URL.Path, ".jpg") ||
		    strings.HasSuffix(r.URL.Path, ".png") ||
		    strings.HasSuffix(r.URL.Path, ".gif") ||
		    strings.HasSuffix(r.URL.Path, ".zip")) {
			next.ServeHTTP(w, r)
			return
		}
		
		// Create gzip response writer
		gzw := &gzipResponseWriter{
			ResponseWriter: w,
		}
		defer gzw.Close()
		
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		
		next.ServeHTTP(gzw, r)
	})
}

// CSRFMiddleware provides CSRF protection
func (ms *MiddlewareStack) CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF for GET, HEAD, OPTIONS requests
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}
		
		// Skip CSRF for login endpoint (for development/testing)
		if r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
			return
		}
		
		// Skip CSRF for API endpoints with proper authentication
		if strings.HasPrefix(r.URL.Path, "/api/") {
			// Check for API key or JWT token
			if r.Header.Get("Authorization") != "" || r.Header.Get("X-API-Key") != "" {
				next.ServeHTTP(w, r)
				return
			}
		}
		
		// Validate CSRF token
		token := r.Header.Get("X-CSRF-Token")
		if token == "" {
			token = r.FormValue("csrf_token")
		}
		
		if !ms.SecurityManager.ValidateCSRFToken(token, r) {
			ms.ErrorManager.HandleHTTPError(w, r, errors.NewApplicationError(
				errors.FORBIDDEN,
				"CSRF_TOKEN_INVALID",
				"CSRF token validation failed",
				nil,
			))
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture response data
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body = append(rw.body, b...)
	return rw.ResponseWriter.Write(b)
}

// gzipResponseWriter wraps http.ResponseWriter for gzip compression
type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (grw *gzipResponseWriter) Write(b []byte) (int, error) {
	if grw.writer == nil {
		grw.writer = gzip.NewWriter(grw.ResponseWriter)
	}
	return grw.writer.Write(b)
}

func (grw *gzipResponseWriter) Close() error {
	if grw.writer != nil {
		return grw.writer.Close()
	}
	return nil
}

// Helper function to get user ID from session
func getUserIDFromSession(r *http.Request) interface{} {
	if sessionValue := r.Context().Value("session"); sessionValue != nil {
		// The session value might be of different types depending on implementation
		return sessionValue
	}
	return nil
}
