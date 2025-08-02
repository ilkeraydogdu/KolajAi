package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error     ErrorDetail `json:"error"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp string      `json:"timestamp"`
	Path      string      `json:"path"`
}

// ErrorDetail contains detailed error information
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Stack   string                 `json:"stack,omitempty"`
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// Error implements the error interface
func (v *ValidationError) Error() string {
	return v.Message
}

// ErrorHandler handles all application errors
type ErrorHandler struct {
	logger *log.Logger
	debug  bool
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *log.Logger, debug bool) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
		debug:  debug,
	}
}

// HandleError handles different types of errors and returns appropriate HTTP responses
func (eh *ErrorHandler) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	// Generate request ID for tracking
	requestID := eh.generateRequestID()
	
	// Log the error
	eh.logError(requestID, r, err)
	
	// Determine error type and create appropriate response
	errorResponse := eh.createErrorResponse(err, requestID, r.URL.Path)
	
	// Set appropriate HTTP status code
	statusCode := eh.getHTTPStatusCode(err)
	
	// Send JSON response for API requests
	if eh.isAPIRequest(r) {
		eh.sendJSONError(w, statusCode, errorResponse)
		return
	}
	
	// Send HTML error page for web requests
	eh.sendHTMLError(w, r, statusCode, errorResponse)
}

// HandleValidationErrors handles validation errors specifically
func (eh *ErrorHandler) HandleValidationErrors(w http.ResponseWriter, r *http.Request, errors []ValidationError) {
	requestID := eh.generateRequestID()
	
	errorResponse := ErrorResponse{
		Error: ErrorDetail{
			Code:    "VALIDATION_ERROR",
			Message: "Request validation failed",
			Details: map[string]interface{}{
				"validation_errors": errors,
			},
		},
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:      r.URL.Path,
	}
	
	if eh.isAPIRequest(r) {
		eh.sendJSONError(w, http.StatusBadRequest, errorResponse)
		return
	}
	
	eh.sendHTMLError(w, r, http.StatusBadRequest, errorResponse)
}

// HandleNotFound handles 404 errors
func (eh *ErrorHandler) HandleNotFound(w http.ResponseWriter, r *http.Request) {
	requestID := eh.generateRequestID()
	
	errorResponse := ErrorResponse{
		Error: ErrorDetail{
			Code:    "NOT_FOUND",
			Message: "The requested resource was not found",
		},
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:      r.URL.Path,
	}
	
	if eh.isAPIRequest(r) {
		eh.sendJSONError(w, http.StatusNotFound, errorResponse)
		return
	}
	
	eh.sendHTMLError(w, r, http.StatusNotFound, errorResponse)
}

// HandleMethodNotAllowed handles 405 errors
func (eh *ErrorHandler) HandleMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	requestID := eh.generateRequestID()
	
	errorResponse := ErrorResponse{
		Error: ErrorDetail{
			Code:    "METHOD_NOT_ALLOWED",
			Message: fmt.Sprintf("Method %s is not allowed for this resource", r.Method),
		},
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:      r.URL.Path,
	}
	
	if eh.isAPIRequest(r) {
		eh.sendJSONError(w, http.StatusMethodNotAllowed, errorResponse)
		return
	}
	
	eh.sendHTMLError(w, r, http.StatusMethodNotAllowed, errorResponse)
}

// HandleInternalError handles 500 errors
func (eh *ErrorHandler) HandleInternalError(w http.ResponseWriter, r *http.Request, err error) {
	requestID := eh.generateRequestID()
	
	// Log the internal error with stack trace
	eh.logError(requestID, r, err)
	
	errorResponse := ErrorResponse{
		Error: ErrorDetail{
			Code:    "INTERNAL_ERROR",
			Message: "An internal server error occurred",
		},
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:      r.URL.Path,
	}
	
	// Include stack trace in debug mode
	if eh.debug && err != nil {
		errorResponse.Error.Stack = eh.getStackTrace()
		errorResponse.Error.Details = map[string]interface{}{
			"error_message": err.Error(),
		}
	}
	
	if eh.isAPIRequest(r) {
		eh.sendJSONError(w, http.StatusInternalServerError, errorResponse)
		return
	}
	
	eh.sendHTMLError(w, r, http.StatusInternalServerError, errorResponse)
}

// HandleUnauthorized handles 401 errors
func (eh *ErrorHandler) HandleUnauthorized(w http.ResponseWriter, r *http.Request, message string) {
	requestID := eh.generateRequestID()
	
	if message == "" {
		message = "Authentication required"
	}
	
	errorResponse := ErrorResponse{
		Error: ErrorDetail{
			Code:    "UNAUTHORIZED",
			Message: message,
		},
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:      r.URL.Path,
	}
	
	if eh.isAPIRequest(r) {
		eh.sendJSONError(w, http.StatusUnauthorized, errorResponse)
		return
	}
	
	// Redirect to login page for web requests
	http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
}

// HandleForbidden handles 403 errors
func (eh *ErrorHandler) HandleForbidden(w http.ResponseWriter, r *http.Request, message string) {
	requestID := eh.generateRequestID()
	
	if message == "" {
		message = "Access denied"
	}
	
	errorResponse := ErrorResponse{
		Error: ErrorDetail{
			Code:    "FORBIDDEN",
			Message: message,
		},
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:      r.URL.Path,
	}
	
	if eh.isAPIRequest(r) {
		eh.sendJSONError(w, http.StatusForbidden, errorResponse)
		return
	}
	
	eh.sendHTMLError(w, r, http.StatusForbidden, errorResponse)
}

// createErrorResponse creates a standardized error response
func (eh *ErrorHandler) createErrorResponse(err error, requestID, path string) ErrorResponse {
	errorResponse := ErrorResponse{
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:      path,
	}
	
	// Determine error type and create appropriate error detail
	switch e := err.(type) {
	case *ValidationError:
		errorResponse.Error = ErrorDetail{
			Code:    "VALIDATION_ERROR",
			Message: e.Message,
			Details: map[string]interface{}{
				"field": e.Field,
				"value": e.Value,
			},
		}
	default:
		errorResponse.Error = ErrorDetail{
			Code:    "UNKNOWN_ERROR",
			Message: err.Error(),
		}
		
		if eh.debug {
			errorResponse.Error.Stack = eh.getStackTrace()
		}
	}
	
	return errorResponse
}

// getHTTPStatusCode determines the appropriate HTTP status code for an error
func (eh *ErrorHandler) getHTTPStatusCode(err error) int {
	errMsg := strings.ToLower(err.Error())
	
	switch {
	case strings.Contains(errMsg, "not found"):
		return http.StatusNotFound
	case strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "authentication"):
		return http.StatusUnauthorized
	case strings.Contains(errMsg, "forbidden") || strings.Contains(errMsg, "access denied"):
		return http.StatusForbidden
	case strings.Contains(errMsg, "validation") || strings.Contains(errMsg, "invalid"):
		return http.StatusBadRequest
	case strings.Contains(errMsg, "conflict"):
		return http.StatusConflict
	case strings.Contains(errMsg, "timeout"):
		return http.StatusRequestTimeout
	default:
		return http.StatusInternalServerError
	}
}

// isAPIRequest checks if the request is an API request
func (eh *ErrorHandler) isAPIRequest(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/") || 
		   r.Header.Get("Accept") == "application/json" ||
		   r.Header.Get("Content-Type") == "application/json"
}

// sendJSONError sends a JSON error response
func (eh *ErrorHandler) sendJSONError(w http.ResponseWriter, statusCode int, errorResponse ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		eh.logger.Printf("Failed to encode error response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// sendHTMLError sends an HTML error page
func (eh *ErrorHandler) sendHTMLError(w http.ResponseWriter, r *http.Request, statusCode int, errorResponse ErrorResponse) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(statusCode)
	
	// Simple HTML error page
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Error %d</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .error-container { max-width: 600px; margin: 0 auto; }
        .error-code { font-size: 48px; color: #e74c3c; margin-bottom: 20px; }
        .error-message { font-size: 18px; color: #333; margin-bottom: 20px; }
        .error-details { background: #f8f9fa; padding: 15px; border-radius: 5px; }
        .request-id { font-size: 12px; color: #666; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="error-container">
        <div class="error-code">%d</div>
        <div class="error-message">%s</div>
        <div class="error-details">
            <p><strong>Path:</strong> %s</p>
            <p><strong>Time:</strong> %s</p>
        </div>
        <div class="request-id">Request ID: %s</div>
        <p><a href="/">← Back to Home</a></p>
    </div>
</body>
</html>`, 
		statusCode, statusCode, errorResponse.Error.Message, 
		errorResponse.Path, errorResponse.Timestamp, errorResponse.RequestID)
	
	w.Write([]byte(html))
}

// logError logs the error with context
func (eh *ErrorHandler) logError(requestID string, r *http.Request, err error) {
	if eh.logger == nil {
		return
	}
	
	eh.logger.Printf("[%s] Error in %s %s: %v", 
		requestID, r.Method, r.URL.Path, err)
	
	if eh.debug {
		eh.logger.Printf("[%s] Stack trace: %s", requestID, eh.getStackTrace())
	}
}

// generateRequestID generates a unique request ID
func (eh *ErrorHandler) generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond()%1000)
}

// getStackTrace returns the current stack trace
func (eh *ErrorHandler) getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// RecoverMiddleware recovers from panics and handles them as errors
func (eh *ErrorHandler) RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Convert panic to error
				var panicErr error
				if e, ok := err.(error); ok {
					panicErr = e
				} else {
					panicErr = fmt.Errorf("panic: %v", err)
				}
				
				eh.HandleInternalError(w, r, panicErr)
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}