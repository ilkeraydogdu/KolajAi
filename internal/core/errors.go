package core

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// ErrorCode represents a standardized error code
type ErrorCode string

// AppError standardizes application errors
type AppError struct {
	Code      ErrorCode
	Message   string
	UserMsg   string // User-friendly message
	Err       error  // Original error
	HTTPCode  int    // HTTP response code
	LogLevel  LogLevel
	RequestID string
}

// LogLevel indicates how an error should be logged
type LogLevel int

const (
	// Error codes
	ErrInternal     ErrorCode = "INTERNAL_ERROR"
	ErrValidation   ErrorCode = "VALIDATION_ERROR"
	ErrAuth         ErrorCode = "AUTH_ERROR"
	ErrNotFound     ErrorCode = "NOT_FOUND"
	ErrDatabase     ErrorCode = "DATABASE_ERROR"
	ErrFormParse    ErrorCode = "FORM_PARSE_ERROR"
	ErrUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrForbidden    ErrorCode = "FORBIDDEN"
	ErrBadRequest   ErrorCode = "BAD_REQUEST"

	// Log levels
	LogSilent LogLevel = iota
	LogError
	LogWarn
	LogInfo
	LogDebug
)

// Error returns the error string
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Err.Error())
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the original error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithRequestID adds a request ID to the error
func (e *AppError) WithRequestID(id string) *AppError {
	e.RequestID = id
	return e
}

// WithHTTPCode sets the HTTP response code
func (e *AppError) WithHTTPCode(code int) *AppError {
	e.HTTPCode = code
	return e
}

// WithLogLevel sets the log level
func (e *AppError) WithLogLevel(level LogLevel) *AppError {
	e.LogLevel = level
	return e
}

// Log logs the error according to its log level
func (e *AppError) Log() {
	switch e.LogLevel {
	case LogDebug:
		log.Printf("[DEBUG] [%s] %s", e.Code, e.Error())
	case LogInfo:
		log.Printf("[INFO] [%s] %s", e.Code, e.Error())
	case LogWarn:
		log.Printf("[WARN] [%s] %s", e.Code, e.Error())
	case LogError:
		log.Printf("[ERROR] [%s] %s", e.Code, e.Error())
	}
}

// RespondWithError sends a JSON error response
func RespondWithError(w http.ResponseWriter, err error) {
	var appErr *AppError
	var httpCode int
	var userMsg string

	// Convert to AppError if not already
	if e, ok := err.(*AppError); ok {
		appErr = e
		httpCode = e.HTTPCode
		userMsg = e.UserMsg
		appErr.Log()
	} else {
		// Create a generic error
		appErr = NewError(ErrInternal, "Bir hata oluştu", err).
			WithHTTPCode(http.StatusInternalServerError).
			WithLogLevel(LogError)
		httpCode = http.StatusInternalServerError
		userMsg = "Bir hata oluştu, lütfen daha sonra tekrar deneyin."
		appErr.Log()
	}

	// Default HTTP code if not set
	if httpCode == 0 {
		httpCode = http.StatusInternalServerError
	}

	// Create response
	response := map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    appErr.Code,
			"message": userMsg,
		},
	}

	// Add request ID if available
	if appErr.RequestID != "" {
		response["request_id"] = appErr.RequestID
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	json.NewEncoder(w).Encode(response)
}

// NewError creates a new AppError
func NewError(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:     code,
		Message:  message,
		UserMsg:  message, // Default user message to the same
		Err:      err,
		LogLevel: LogError,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string, validationErrors map[string][]string) *AppError {
	data, _ := json.Marshal(validationErrors)
	err := fmt.Errorf("validation errors: %s", string(data))

	return &AppError{
		Code:     ErrValidation,
		Message:  message,
		UserMsg:  message,
		Err:      err,
		HTTPCode: http.StatusBadRequest,
		LogLevel: LogWarn,
	}
}

// NewDatabaseError creates a database error
func NewDatabaseError(message string, err error) *AppError {
	return &AppError{
		Code:     ErrDatabase,
		Message:  message,
		UserMsg:  "Veritabanı işlemi sırasında bir hata oluştu",
		Err:      err,
		HTTPCode: http.StatusInternalServerError,
		LogLevel: LogError,
	}
}

// NewAuthError creates an authentication error
func NewAuthError(message string, err error) *AppError {
	return &AppError{
		Code:     ErrAuth,
		Message:  message,
		UserMsg:  message,
		Err:      err,
		HTTPCode: http.StatusUnauthorized,
		LogLevel: LogWarn,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Code:     ErrNotFound,
		Message:  message,
		UserMsg:  message,
		HTTPCode: http.StatusNotFound,
		LogLevel: LogInfo,
	}
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string, err error) *AppError {
	return &AppError{
		Code:     ErrBadRequest,
		Message:  message,
		UserMsg:  message,
		Err:      err,
		HTTPCode: http.StatusBadRequest,
		LogLevel: LogWarn,
	}
}

// NewFormParseError creates a form parsing error
func NewFormParseError(err error) *AppError {
	return &AppError{
		Code:     ErrFormParse,
		Message:  "Form verileri işlenemedi",
		UserMsg:  "Form verileri işlenemedi, lütfen tekrar deneyin",
		Err:      err,
		HTTPCode: http.StatusBadRequest,
		LogLevel: LogWarn,
	}
}
