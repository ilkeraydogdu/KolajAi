package security

import (
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// InputValidator provides comprehensive input validation and sanitization
type InputValidator struct {
	maxStringLength int
	allowedDomains  map[string]bool
	sqlInjectionRegex *regexp.Regexp
	xssRegex          *regexp.Regexp
	emailRegex        *regexp.Regexp
	phoneRegex        *regexp.Regexp
	urlRegex          *regexp.Regexp
}

// ValidationResult contains validation result and sanitized value
type ValidationResult struct {
	IsValid      bool     `json:"is_valid"`
	SanitizedValue interface{} `json:"sanitized_value"`
	Errors       []string `json:"errors"`
	Warnings     []string `json:"warnings"`
}

// ValidationRule defines validation rules for different input types
type ValidationRule struct {
	Required     bool   `json:"required"`
	MinLength    int    `json:"min_length"`
	MaxLength    int    `json:"max_length"`
	Pattern      string `json:"pattern"`
	AllowedValues []string `json:"allowed_values"`
	Type         string `json:"type"` // string, int, float, email, url, phone, etc.
}

// NewInputValidator creates a new input validator
func NewInputValidator() *InputValidator {
	// Common SQL injection patterns
	sqlPattern := `(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|vbscript|onload|onerror|onclick)`
	
	// Common XSS patterns
	xssPattern := `(?i)(<script|javascript:|vbscript:|onload=|onerror=|onclick=|onmouseover=|<iframe|<object|<embed)`
	
	// Email validation pattern (RFC 5322 compliant)
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	
	// Phone number pattern (international format)
	phonePattern := `^\+?[1-9]\d{1,14}$`
	
	// URL validation pattern
	urlPattern := `^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(/.*)?$`

	return &InputValidator{
		maxStringLength:   10000, // Default max string length
		allowedDomains:    make(map[string]bool),
		sqlInjectionRegex: regexp.MustCompile(sqlPattern),
		xssRegex:          regexp.MustCompile(xssPattern),
		emailRegex:        regexp.MustCompile(emailPattern),
		phoneRegex:        regexp.MustCompile(phonePattern),
		urlRegex:          regexp.MustCompile(urlPattern),
	}
}

// ValidateString validates and sanitizes string input
func (v *InputValidator) ValidateString(input string, rule ValidationRule) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []string{},
		Warnings: []string{},
	}

	// Check if required
	if rule.Required && strings.TrimSpace(input) == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "field is required")
		return result
	}

	// Skip validation if empty and not required
	if strings.TrimSpace(input) == "" {
		result.SanitizedValue = ""
		return result
	}

	// Length validation
	if rule.MinLength > 0 && len(input) < rule.MinLength {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("minimum length is %d characters", rule.MinLength))
	}

	maxLen := rule.MaxLength
	if maxLen == 0 {
		maxLen = v.maxStringLength
	}
	if len(input) > maxLen {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("maximum length is %d characters", maxLen))
	}

	// Pattern validation
	if rule.Pattern != "" {
		if matched, err := regexp.MatchString(rule.Pattern, input); err != nil || !matched {
			result.IsValid = false
			result.Errors = append(result.Errors, "input does not match required pattern")
		}
	}

	// Allowed values validation
	if len(rule.AllowedValues) > 0 {
		allowed := false
		for _, allowedValue := range rule.AllowedValues {
			if input == allowedValue {
				allowed = true
				break
			}
		}
		if !allowed {
			result.IsValid = false
			result.Errors = append(result.Errors, "value is not in allowed list")
		}
	}

	// Security checks
	if v.containsSQLInjection(input) {
		result.IsValid = false
		result.Errors = append(result.Errors, "potential SQL injection detected")
	}

	if v.containsXSS(input) {
		result.IsValid = false
		result.Errors = append(result.Errors, "potential XSS attack detected")
	}

	// Sanitize the input
	sanitized := v.sanitizeString(input)
	result.SanitizedValue = sanitized

	// Add warning if sanitization changed the input
	if sanitized != input {
		result.Warnings = append(result.Warnings, "input was sanitized")
	}

	return result
}

// ValidateEmail validates email addresses
func (v *InputValidator) ValidateEmail(email string) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []string{},
		Warnings: []string{},
	}

	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "email is required")
		return result
	}

	// Length check
	if len(email) > 254 {
		result.IsValid = false
		result.Errors = append(result.Errors, "email is too long")
	}

	// Format validation
	if !v.emailRegex.MatchString(email) {
		result.IsValid = false
		result.Errors = append(result.Errors, "invalid email format")
	}

	// Domain validation
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		domain := parts[1]
		if len(v.allowedDomains) > 0 && !v.allowedDomains[domain] {
			result.IsValid = false
			result.Errors = append(result.Errors, "email domain not allowed")
		}
	}

	result.SanitizedValue = email
	return result
}

// ValidateURL validates and sanitizes URLs
func (v *InputValidator) ValidateURL(inputURL string) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []string{},
		Warnings: []string{},
	}

	inputURL = strings.TrimSpace(inputURL)

	if inputURL == "" {
		result.SanitizedValue = ""
		return result
	}

	// Parse URL
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "invalid URL format")
		return result
	}

	// Scheme validation
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		result.IsValid = false
		result.Errors = append(result.Errors, "only HTTP and HTTPS URLs are allowed")
	}

	// Host validation
	if parsedURL.Host == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "URL must have a valid host")
	}

	// Domain validation
	if len(v.allowedDomains) > 0 {
		host := strings.ToLower(parsedURL.Host)
		// Remove port if present
		if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
			host = host[:colonIndex]
		}
		
		if !v.allowedDomains[host] {
			result.IsValid = false
			result.Errors = append(result.Errors, "URL domain not allowed")
		}
	}

	result.SanitizedValue = parsedURL.String()
	return result
}

// ValidatePhone validates phone numbers
func (v *InputValidator) ValidatePhone(phone string) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []string{},
		Warnings: []string{},
	}

	// Remove all non-digit characters except +
	sanitized := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")

	if sanitized == "" {
		result.SanitizedValue = ""
		return result
	}

	// Validate format
	if !v.phoneRegex.MatchString(sanitized) {
		result.IsValid = false
		result.Errors = append(result.Errors, "invalid phone number format")
	}

	result.SanitizedValue = sanitized
	
	if sanitized != phone {
		result.Warnings = append(result.Warnings, "phone number was sanitized")
	}

	return result
}

// ValidateInteger validates integer input
func (v *InputValidator) ValidateInteger(input string, min, max int64) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []string{},
		Warnings: []string{},
	}

	input = strings.TrimSpace(input)

	if input == "" {
		result.SanitizedValue = nil
		return result
	}

	// Parse integer
	value, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "invalid integer format")
		return result
	}

	// Range validation
	if min != 0 && value < min {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("value must be at least %d", min))
	}

	if max != 0 && value > max {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("value must be at most %d", max))
	}

	result.SanitizedValue = value
	return result
}

// ValidateFloat validates float input
func (v *InputValidator) ValidateFloat(input string, min, max float64) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []string{},
		Warnings: []string{},
	}

	input = strings.TrimSpace(input)

	if input == "" {
		result.SanitizedValue = nil
		return result
	}

	// Parse float
	value, err := strconv.ParseFloat(input, 64)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "invalid number format")
		return result
	}

	// Range validation
	if min != 0 && value < min {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("value must be at least %f", min))
	}

	if max != 0 && value > max {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("value must be at most %f", max))
	}

	result.SanitizedValue = value
	return result
}

// ValidateDate validates date input
func (v *InputValidator) ValidateDate(input, format string) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []string{},
		Warnings: []string{},
	}

	input = strings.TrimSpace(input)

	if input == "" {
		result.SanitizedValue = nil
		return result
	}

	// Use default format if not provided
	if format == "" {
		format = "2006-01-02"
	}

	// Parse date
	parsedTime, err := time.Parse(format, input)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("invalid date format, expected: %s", format))
		return result
	}

	result.SanitizedValue = parsedTime
	return result
}

// ValidateJSON validates JSON input
func (v *InputValidator) ValidateJSON(input string) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []string{},
		Warnings: []string{},
	}

	input = strings.TrimSpace(input)

	if input == "" {
		result.SanitizedValue = ""
		return result
	}

	// Basic JSON structure validation
	if !strings.HasPrefix(input, "{") && !strings.HasPrefix(input, "[") {
		result.IsValid = false
		result.Errors = append(result.Errors, "JSON must start with { or [")
		return result
	}

	if !strings.HasSuffix(input, "}") && !strings.HasSuffix(input, "]") {
		result.IsValid = false
		result.Errors = append(result.Errors, "JSON must end with } or ]")
		return result
	}

	// Security checks
	if v.containsXSS(input) {
		result.IsValid = false
		result.Errors = append(result.Errors, "potential XSS in JSON detected")
	}

	result.SanitizedValue = input
	return result
}

// ValidateProductData validates marketplace product data
func (v *InputValidator) ValidateProductData(data map[string]interface{}) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []string{},
		Warnings: []string{},
		SanitizedValue: make(map[string]interface{}),
	}

	sanitizedData := make(map[string]interface{})

	// Required fields validation
	requiredFields := []string{"title", "description", "price", "category_id"}
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("required field missing: %s", field))
		}
	}

	// Validate title
	if title, ok := data["title"].(string); ok {
		titleResult := v.ValidateString(title, ValidationRule{
			Required:  true,
			MinLength: 5,
			MaxLength: 200,
		})
		if !titleResult.IsValid {
			result.IsValid = false
			result.Errors = append(result.Errors, "title validation failed")
		}
		sanitizedData["title"] = titleResult.SanitizedValue
	}

	// Validate description
	if description, ok := data["description"].(string); ok {
		descResult := v.ValidateString(description, ValidationRule{
			Required:  true,
			MinLength: 10,
			MaxLength: 5000,
		})
		if !descResult.IsValid {
			result.IsValid = false
			result.Errors = append(result.Errors, "description validation failed")
		}
		sanitizedData["description"] = descResult.SanitizedValue
	}

	// Validate price
	if priceStr, ok := data["price"].(string); ok {
		priceResult := v.ValidateFloat(priceStr, 0.01, 999999.99)
		if !priceResult.IsValid {
			result.IsValid = false
			result.Errors = append(result.Errors, "price validation failed")
		}
		sanitizedData["price"] = priceResult.SanitizedValue
	}

	// Validate category_id
	if categoryStr, ok := data["category_id"].(string); ok {
		categoryResult := v.ValidateInteger(categoryStr, 1, 999999)
		if !categoryResult.IsValid {
			result.IsValid = false
			result.Errors = append(result.Errors, "category_id validation failed")
		}
		sanitizedData["category_id"] = categoryResult.SanitizedValue
	}

	// Validate images if present
	if images, ok := data["images"].([]interface{}); ok {
		var sanitizedImages []interface{}
		for i, img := range images {
			if imgMap, ok := img.(map[string]interface{}); ok {
				if url, ok := imgMap["url"].(string); ok {
					urlResult := v.ValidateURL(url)
					if !urlResult.IsValid {
						result.Warnings = append(result.Warnings, fmt.Sprintf("image %d URL validation failed", i))
					} else {
						sanitizedImages = append(sanitizedImages, map[string]interface{}{
							"url": urlResult.SanitizedValue,
						})
					}
				}
			}
		}
		sanitizedData["images"] = sanitizedImages
	}

	result.SanitizedValue = sanitizedData
	return result
}

// containsSQLInjection checks for SQL injection patterns
func (v *InputValidator) containsSQLInjection(input string) bool {
	return v.sqlInjectionRegex.MatchString(input)
}

// containsXSS checks for XSS patterns
func (v *InputValidator) containsXSS(input string) bool {
	return v.xssRegex.MatchString(input)
}

// sanitizeString sanitizes string input
func (v *InputValidator) sanitizeString(input string) string {
	// HTML escape
	sanitized := html.EscapeString(input)
	
	// Remove control characters
	sanitized = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, sanitized)
	
	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)
	
	return sanitized
}

// SetMaxStringLength sets the maximum allowed string length
func (v *InputValidator) SetMaxStringLength(length int) {
	v.maxStringLength = length
}

// AddAllowedDomain adds a domain to the allowed domains list
func (v *InputValidator) AddAllowedDomain(domain string) {
	v.allowedDomains[strings.ToLower(domain)] = true
}

// RemoveAllowedDomain removes a domain from the allowed domains list
func (v *InputValidator) RemoveAllowedDomain(domain string) {
	delete(v.allowedDomains, strings.ToLower(domain))
}

// ClearAllowedDomains clears the allowed domains list
func (v *InputValidator) ClearAllowedDomains() {
	v.allowedDomains = make(map[string]bool)
}

// ValidateBatch validates multiple inputs at once
func (v *InputValidator) ValidateBatch(inputs map[string]interface{}, rules map[string]ValidationRule) map[string]*ValidationResult {
	results := make(map[string]*ValidationResult)

	for field, input := range inputs {
		rule, hasRule := rules[field]
		if !hasRule {
			rule = ValidationRule{Type: "string"} // Default rule
		}

		switch inputStr := input.(type) {
		case string:
			switch rule.Type {
			case "email":
				results[field] = v.ValidateEmail(inputStr)
			case "url":
				results[field] = v.ValidateURL(inputStr)
			case "phone":
				results[field] = v.ValidatePhone(inputStr)
			case "int":
				results[field] = v.ValidateInteger(inputStr, 0, 0)
			case "float":
				results[field] = v.ValidateFloat(inputStr, 0, 0)
			case "date":
				results[field] = v.ValidateDate(inputStr, "")
			case "json":
				results[field] = v.ValidateJSON(inputStr)
			default:
				results[field] = v.ValidateString(inputStr, rule)
			}
		default:
			// For non-string inputs, convert to string first
			if inputStr := fmt.Sprintf("%v", input); inputStr != "" {
				results[field] = v.ValidateString(inputStr, rule)
			}
		}
	}

	return results
}

// IsValidationPassed checks if all validation results passed
func IsValidationPassed(results map[string]*ValidationResult) bool {
	for _, result := range results {
		if !result.IsValid {
			return false
		}
	}
	return true
}

// GetValidationErrors extracts all validation errors
func GetValidationErrors(results map[string]*ValidationResult) map[string][]string {
	errors := make(map[string][]string)
	
	for field, result := range results {
		if len(result.Errors) > 0 {
			errors[field] = result.Errors
		}
	}
	
	return errors
}

// GetSanitizedData extracts all sanitized values
func GetSanitizedData(results map[string]*ValidationResult) map[string]interface{} {
	sanitized := make(map[string]interface{})
	
	for field, result := range results {
		if result.IsValid && result.SanitizedValue != nil {
			sanitized[field] = result.SanitizedValue
		}
	}
	
	return sanitized
}