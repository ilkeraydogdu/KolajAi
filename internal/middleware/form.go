package middleware

import (
	"context"
	"net/http"

	"kolajAi/internal/core"
	"kolajAi/internal/validation"
)

// FormContext key types
type formContextKey string

// Context keys
const (
	FormDataKey   formContextKey = "form_data"
	FormErrorsKey formContextKey = "form_errors"
)

// FormData stores the form data
type FormData struct {
	Data      map[string]string
	Errors    map[string][]string
	IsValid   bool
	Schema    string
	Validator *validation.FormValidator
}

// ProcessFormMiddleware middleware for processing form data
func ProcessFormMiddleware(validator *validation.FormValidator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Process only POST, PUT requests
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				next.ServeHTTP(w, r)
				return
			}

			// Get schema name from the path or query parameter
			schema := r.URL.Query().Get("schema")
			if schema == "" {
				// Try to extract from path
				parts := SplitPath(r.URL.Path)
				if len(parts) > 0 {
					schema = parts[len(parts)-1]
				}
			}

			// If still no schema, continue without processing
			if schema == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Parse form
			if err := r.ParseForm(); err != nil {
				appErr := core.NewFormParseError(err)
				if r.Header.Get("Content-Type") == "application/json" {
					core.RespondWithError(w, appErr)
					return
				}
				// Store error in context and continue
				ctx := context.WithValue(r.Context(), FormErrorsKey, appErr)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Convert form data to map
			formData := make(map[string]string)
			for key, values := range r.Form {
				if len(values) > 0 {
					formData[key] = values[0]
				}
			}

			// Validate form
			isValid, validationErrors := validator.ValidateForm(schema, formData)

			// Create form data structure
			data := &FormData{
				Data:      formData,
				Errors:    validationErrors,
				IsValid:   isValid,
				Schema:    schema,
				Validator: validator,
			}

			// Store form data in context
			ctx := context.WithValue(r.Context(), FormDataKey, data)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetFormData retrieves form data from the request context
func GetFormData(r *http.Request) *FormData {
	data, ok := r.Context().Value(FormDataKey).(*FormData)
	if !ok {
		return nil
	}
	return data
}

// GetFormErrors retrieves form errors from the request context
func GetFormErrors(r *http.Request) *core.AppError {
	err, ok := r.Context().Value(FormErrorsKey).(*core.AppError)
	if !ok {
		return nil
	}
	return err
}

// SplitPath splits a URL path into segments
func SplitPath(path string) []string {
	var parts []string
	var current string

	for _, c := range path {
		if c == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}
