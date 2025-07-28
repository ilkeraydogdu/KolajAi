package router

import (
	"net/http"
	"strings"
	"time"

	"kolajAi/internal/middleware"
)

// Router represents the application router
type Router struct {
	mux        *http.ServeMux
	middleware *middleware.MiddlewareStack
	routes     map[string]RouteInfo
}

// RouteInfo holds information about a route
type RouteInfo struct {
	Pattern     string
	Method      string
	Handler     http.HandlerFunc
	Middleware  []string
	Protected   bool
	CacheEnabled bool
	RateLimit   int
	Description string
	Tags        []string
}

// NewRouter creates a new router with middleware stack
func NewRouter(middlewareStack *middleware.MiddlewareStack) *Router {
	return &Router{
		mux:        http.NewServeMux(),
		middleware: middlewareStack,
		routes:     make(map[string]RouteInfo),
	}
}

// ServeHTTP implements http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware stack
	handler := r.applyMiddleware(r.mux)
	handler.ServeHTTP(w, req)
}

// HandleFunc registers a handler function for the given pattern
func (r *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	r.HandleFuncWithOptions(pattern, handler, RouteOptions{})
}

// RouteOptions holds options for route registration
type RouteOptions struct {
	Method       string
	Protected    bool
	CacheEnabled bool
	RateLimit    int
	Middleware   []string
	Description  string
	Tags         []string
}

// HandleFuncWithOptions registers a handler function with options
func (r *Router) HandleFuncWithOptions(pattern string, handler http.HandlerFunc, options RouteOptions) {
	// Store route information
	r.routes[pattern] = RouteInfo{
		Pattern:     pattern,
		Method:      options.Method,
		Handler:     handler,
		Middleware:  options.Middleware,
		Protected:   options.Protected,
		CacheEnabled: options.CacheEnabled,
		RateLimit:   options.RateLimit,
		Description: options.Description,
		Tags:        options.Tags,
	}

	// Apply route-specific middleware
	finalHandler := r.applyRouteMiddleware(handler, options)
	
	// Register with mux
	r.mux.HandleFunc(pattern, finalHandler)
}

// Handle registers a handler for the given pattern
func (r *Router) Handle(pattern string, handler http.Handler) {
	r.mux.Handle(pattern, handler)
}

// applyMiddleware applies the middleware stack to the handler
func (r *Router) applyMiddleware(handler http.Handler) http.Handler {
	// Apply middleware in reverse order (last middleware wraps first)
	middlewares := []func(http.Handler) http.Handler{
		r.middleware.LoggingMiddleware,
		r.middleware.ErrorHandlingMiddleware,
		r.middleware.CompressionMiddleware,
		r.middleware.CORSMiddleware,
		r.middleware.SecurityMiddleware,
		r.middleware.SessionMiddleware,
		r.middleware.CSRFMiddleware,
		r.middleware.CacheMiddleware,
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}

// applyRouteMiddleware applies route-specific middleware
func (r *Router) applyRouteMiddleware(handler http.HandlerFunc, options RouteOptions) http.HandlerFunc {
	finalHandler := handler

	// Apply rate limiting if specified
	if options.RateLimit > 0 {
		finalHandler = r.rateLimitMiddleware(finalHandler, options.RateLimit)
	}

	// Apply authentication if protected
	if options.Protected {
		finalHandler = r.authMiddleware(finalHandler)
	}

	// Apply custom middleware if specified
	for _, middlewareName := range options.Middleware {
		switch middlewareName {
		case "admin":
			finalHandler = r.adminMiddleware(finalHandler)
		case "api":
			finalHandler = r.apiMiddleware(finalHandler)
		case "vendor":
			finalHandler = r.vendorMiddleware(finalHandler)
		}
	}

	return finalHandler
}

// authMiddleware checks if user is authenticated
func (r *Router) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Get session from context
		session := req.Context().Value("session")
		if session == nil {
			r.redirectToLogin(w, req)
			return
		}

		// Check if session is valid and active
		// This would use the session manager to validate
		
		next.ServeHTTP(w, req)
	})
}

// adminMiddleware checks if user has admin privileges
func (r *Router) adminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Get session from context
		session := req.Context().Value("session")
		if session == nil {
			r.redirectToLogin(w, req)
			return
		}

		// Check admin privileges
		// This would use the session manager to check user role
		
		next.ServeHTTP(w, req)
	})
}

// vendorMiddleware checks if user has vendor privileges
func (r *Router) vendorMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Get session from context
		session := req.Context().Value("session")
		if session == nil {
			r.redirectToLogin(w, req)
			return
		}

		// Check vendor privileges
		// This would use the session manager to check user role
		
		next.ServeHTTP(w, req)
	})
}

// apiMiddleware applies API-specific middleware
func (r *Router) apiMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Set API headers
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-API-Version", "1.0")
		
		// Check API key or JWT token
		apiKey := req.Header.Get("X-API-Key")
		authHeader := req.Header.Get("Authorization")
		
		if apiKey == "" && authHeader == "" {
			http.Error(w, "API key or authorization token required", http.StatusUnauthorized)
			return
		}
		
		next.ServeHTTP(w, req)
	})
}

// rateLimitMiddleware applies rate limiting
func (r *Router) rateLimitMiddleware(next http.HandlerFunc, limit int) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// This would use the security manager to check rate limits
		// For now, just pass through
		next.ServeHTTP(w, req)
	})
}

// redirectToLogin redirects to login page
func (r *Router) redirectToLogin(w http.ResponseWriter, req *http.Request) {
	// Check if it's an API request
	if strings.HasPrefix(req.URL.Path, "/api/") || 
	   req.Header.Get("Accept") == "application/json" ||
	   req.Header.Get("Content-Type") == "application/json" {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	
	// Redirect to login page
	http.Redirect(w, req, "/login?redirect="+req.URL.Path, http.StatusSeeOther)
}

// GetRoutes returns all registered routes
func (r *Router) GetRoutes() map[string]RouteInfo {
	return r.routes
}

// RouteGroup allows grouping routes with common middleware
type RouteGroup struct {
	router     *Router
	prefix     string
	middleware []string
	protected  bool
}

// Group creates a new route group
func (r *Router) Group(prefix string, options RouteOptions) *RouteGroup {
	return &RouteGroup{
		router:     r,
		prefix:     strings.TrimSuffix(prefix, "/"),
		middleware: options.Middleware,
		protected:  options.Protected,
	}
}

// HandleFunc registers a handler function in the group
func (rg *RouteGroup) HandleFunc(pattern string, handler http.HandlerFunc) {
	rg.HandleFuncWithOptions(pattern, handler, RouteOptions{})
}

// HandleFuncWithOptions registers a handler function with options in the group
func (rg *RouteGroup) HandleFuncWithOptions(pattern string, handler http.HandlerFunc, options RouteOptions) {
	// Combine group middleware with route middleware
	combinedMiddleware := append(rg.middleware, options.Middleware...)
	options.Middleware = combinedMiddleware
	
	// Apply group protection if not overridden
	if rg.protected && !options.Protected {
		options.Protected = true
	}
	
	// Add prefix to pattern
	fullPattern := rg.prefix + pattern
	
	rg.router.HandleFuncWithOptions(fullPattern, handler, options)
}

// Static serves static files
func (r *Router) Static(prefix, dir string) {
	fileServer := http.FileServer(http.Dir(dir))
	r.Handle(prefix, http.StripPrefix(prefix, fileServer))
}

// NotFound sets a custom 404 handler
func (r *Router) NotFound(handler http.HandlerFunc) {
	// This would require a custom mux implementation
	// For now, we'll use the default behavior
}

// MethodNotAllowed sets a custom 405 handler
func (r *Router) MethodNotAllowed(handler http.HandlerFunc) {
	// This would require a custom mux implementation
	// For now, we'll use the default behavior
}

// Use adds middleware to the router
func (r *Router) Use(middleware func(http.Handler) http.Handler) {
	// This would require modifying the middleware stack
	// For now, middleware is applied through the stack
}

// Health check endpoint
func (r *Router) SetupHealthCheck() {
	r.HandleFuncWithOptions("/health", r.healthCheckHandler, RouteOptions{
		Method:      "GET",
		Description: "Health check endpoint",
		Tags:        []string{"system", "health"},
	})
}

// healthCheckHandler handles health check requests
func (r *Router) healthCheckHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	// This would use JSON encoder
	w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
}

// Metrics endpoint
func (r *Router) SetupMetrics() {
	r.HandleFuncWithOptions("/metrics", r.metricsHandler, RouteOptions{
		Method:      "GET",
		Description: "Application metrics endpoint",
		Tags:        []string{"system", "metrics"},
		Protected:   true,
		Middleware:  []string{"admin"},
	})
}

// metricsHandler handles metrics requests
func (r *Router) metricsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	// This would use JSON encoder
	w.Write([]byte(`{"requests_total":0,"memory_usage":0}`))
}

// Global variables
var startTime = time.Now()

// CORS options
type CORSOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// SetCORSOptions sets CORS options for the router
func (r *Router) SetCORSOptions(options CORSOptions) {
	// This would configure CORS middleware
	// For now, CORS is handled in the middleware stack
}
