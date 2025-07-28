package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"kolajAi/internal/models"
)

// Validator provides comprehensive validation functionality
type Validator struct {
	rules map[string][]ValidationRule
}

// ValidationRule represents a validation rule
type ValidationRule struct {
	Field     string
	Rule      string
	Parameter string
	Message   string
	Required  bool
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		rules: make(map[string][]ValidationRule),
	}
}

// ValidateStruct validates a struct using reflection and validation tags
func (v *Validator) ValidateStruct(s interface{}) error {
	var errors []ValidationError

	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("validation requires a struct")
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		
		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Get validation tags
		validationTag := fieldType.Tag.Get("validate")
		if validationTag == "" {
			continue
		}

		fieldName := fieldType.Name
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			// Use JSON field name if available
			if idx := strings.Index(jsonTag, ","); idx != -1 {
				fieldName = jsonTag[:idx]
			} else {
				fieldName = jsonTag
			}
		}

		// Parse and apply validation rules
		rules := v.parseValidationTag(validationTag)
		for _, rule := range rules {
			if err := v.validateField(fieldName, field.Interface(), rule); err != nil {
				errors = append(errors, *err)
			}
		}
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidateProduct validates product data with business rules
func (v *Validator) ValidateProduct(product *models.Product) error {
	var errors []ValidationError

	// Required fields
	if strings.TrimSpace(product.Name) == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Rule:    "required",
			Message: "Product name is required",
		})
	}

	if len(product.Name) > 255 {
		errors = append(errors, ValidationError{
			Field:   "name",
			Rule:    "max_length",
			Message: "Product name cannot exceed 255 characters",
		})
	}

	if product.Price < 0 {
		errors = append(errors, ValidationError{
			Field:   "price",
			Rule:    "min",
			Message: "Product price cannot be negative",
			Value:   product.Price,
		})
	}

	if product.Stock < 0 {
		errors = append(errors, ValidationError{
			Field:   "stock",
			Rule:    "min",
			Message: "Product stock cannot be negative",
			Value:   product.Stock,
		})
	}

	if product.VendorID <= 0 {
		errors = append(errors, ValidationError{
			Field:   "vendor_id",
			Rule:    "required",
			Message: "Valid vendor ID is required",
			Value:   product.VendorID,
		})
	}

	if product.CategoryID <= 0 {
		errors = append(errors, ValidationError{
			Field:   "category_id",
			Rule:    "required",
			Message: "Valid category ID is required",
			Value:   product.CategoryID,
		})
	}

	// SKU validation
	if product.SKU != "" {
		if !v.isValidSKU(product.SKU) {
			errors = append(errors, ValidationError{
				Field:   "sku",
				Rule:    "format",
				Message: "SKU format is invalid",
				Value:   product.SKU,
			})
		}
	}

	// Status validation
	validStatuses := []string{"draft", "active", "inactive", "out_of_stock"}
	if product.Status != "" && !v.contains(validStatuses, product.Status) {
		errors = append(errors, ValidationError{
			Field:   "status",
			Rule:    "in",
			Message: "Invalid product status",
			Value:   product.Status,
		})
	}

	// Weight validation
	if product.Weight < 0 {
		errors = append(errors, ValidationError{
			Field:   "weight",
			Rule:    "min",
			Message: "Product weight cannot be negative",
			Value:   product.Weight,
		})
	}

	// Wholesale validation
	if product.WholesalePrice > 0 && product.MinWholesaleQty <= 0 {
		errors = append(errors, ValidationError{
			Field:   "min_wholesale_qty",
			Rule:    "required_with",
			Message: "Minimum wholesale quantity is required when wholesale price is set",
		})
	}

	// Price comparison validation
	if product.ComparePrice > 0 && product.ComparePrice <= product.Price {
		errors = append(errors, ValidationError{
			Field:   "compare_price",
			Rule:    "greater_than",
			Message: "Compare price must be greater than regular price",
		})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidateUser validates user data
func (v *Validator) ValidateUser(user *models.User) error {
	var errors []ValidationError

	// Name validation
	if strings.TrimSpace(user.Name) == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Rule:    "required",
			Message: "Name is required",
		})
	}

	if len(user.Name) < 2 {
		errors = append(errors, ValidationError{
			Field:   "name",
			Rule:    "min_length",
			Message: "Name must be at least 2 characters long",
		})
	}

	if len(user.Name) > 100 {
		errors = append(errors, ValidationError{
			Field:   "name",
			Rule:    "max_length",
			Message: "Name cannot exceed 100 characters",
		})
	}

	// Email validation
	if strings.TrimSpace(user.Email) == "" {
		errors = append(errors, ValidationError{
			Field:   "email",
			Rule:    "required",
			Message: "Email is required",
		})
	}

	if !v.isValidEmail(user.Email) {
		errors = append(errors, ValidationError{
			Field:   "email",
			Rule:    "email",
			Message: "Invalid email format",
			Value:   user.Email,
		})
	}

	// Password validation
	if strings.TrimSpace(user.Password) == "" {
		errors = append(errors, ValidationError{
			Field:   "password",
			Rule:    "required",
			Message: "Password is required",
		})
	}

	if err := v.validatePassword(user.Password); err != nil {
		errors = append(errors, ValidationError{
			Field:   "password",
			Rule:    "password_strength",
			Message: err.Error(),
		})
	}

	// Phone validation
	if user.Phone != "" && !v.isValidPhone(user.Phone) {
		errors = append(errors, ValidationError{
			Field:   "phone",
			Rule:    "phone",
			Message: "Invalid phone number format",
			Value:   user.Phone,
		})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidateOrder validates order data
func (v *Validator) ValidateOrder(order *models.Order) error {
	var errors []ValidationError

	if order.UserID <= 0 {
		errors = append(errors, ValidationError{
			Field:   "user_id",
			Rule:    "required",
			Message: "Valid user ID is required",
		})
	}

	if order.TotalAmount < 0 {
		errors = append(errors, ValidationError{
			Field:   "total_amount",
			Rule:    "min",
			Message: "Total amount cannot be negative",
		})
	}

	// Status validation
	validStatuses := []string{"pending", "confirmed", "processing", "shipped", "delivered", "cancelled"}
	if order.Status != "" && !v.contains(validStatuses, order.Status) {
		errors = append(errors, ValidationError{
			Field:   "status",
			Rule:    "in",
			Message: "Invalid order status",
		})
	}

	// Shipping address validation
	if strings.TrimSpace(order.ShippingAddress) == "" {
		errors = append(errors, ValidationError{
			Field:   "shipping_address",
			Rule:    "required",
			Message: "Shipping address is required",
		})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// parseValidationTag parses validation tag string
func (v *Validator) parseValidationTag(tag string) []string {
	return strings.Split(tag, ",")
}

// validateField validates a single field with a rule
func (v *Validator) validateField(fieldName string, value interface{}, rule string) *ValidationError {
	parts := strings.Split(rule, ":")
	ruleName := parts[0]
	var parameter string
	if len(parts) > 1 {
		parameter = parts[1]
	}

	switch ruleName {
	case "required":
		if v.isEmpty(value) {
			return &ValidationError{
				Field:   fieldName,
				Rule:    ruleName,
				Message: fmt.Sprintf("%s is required", fieldName),
			}
		}

	case "min":
		if param, err := strconv.ParseFloat(parameter, 64); err == nil {
			if num, ok := v.toFloat(value); ok && num < param {
				return &ValidationError{
					Field:   fieldName,
					Rule:    ruleName,
					Message: fmt.Sprintf("%s must be at least %v", fieldName, param),
					Value:   value,
				}
			}
		}

	case "max":
		if param, err := strconv.ParseFloat(parameter, 64); err == nil {
			if num, ok := v.toFloat(value); ok && num > param {
				return &ValidationError{
					Field:   fieldName,
					Rule:    ruleName,
					Message: fmt.Sprintf("%s cannot exceed %v", fieldName, param),
					Value:   value,
				}
			}
		}

	case "min_length":
		if param, err := strconv.Atoi(parameter); err == nil {
			if str, ok := value.(string); ok && len(str) < param {
				return &ValidationError{
					Field:   fieldName,
					Rule:    ruleName,
					Message: fmt.Sprintf("%s must be at least %d characters long", fieldName, param),
					Value:   value,
				}
			}
		}

	case "max_length":
		if param, err := strconv.Atoi(parameter); err == nil {
			if str, ok := value.(string); ok && len(str) > param {
				return &ValidationError{
					Field:   fieldName,
					Rule:    ruleName,
					Message: fmt.Sprintf("%s cannot exceed %d characters", fieldName, param),
					Value:   value,
				}
			}
		}

	case "email":
		if str, ok := value.(string); ok && !v.isValidEmail(str) {
			return &ValidationError{
				Field:   fieldName,
				Rule:    ruleName,
				Message: fmt.Sprintf("%s must be a valid email address", fieldName),
				Value:   value,
			}
		}

	case "url":
		if str, ok := value.(string); ok && !v.isValidURL(str) {
			return &ValidationError{
				Field:   fieldName,
				Rule:    ruleName,
				Message: fmt.Sprintf("%s must be a valid URL", fieldName),
				Value:   value,
			}
		}

	case "phone":
		if str, ok := value.(string); ok && !v.isValidPhone(str) {
			return &ValidationError{
				Field:   fieldName,
				Rule:    ruleName,
				Message: fmt.Sprintf("%s must be a valid phone number", fieldName),
				Value:   value,
			}
		}

	case "date":
		if str, ok := value.(string); ok && !v.isValidDate(str) {
			return &ValidationError{
				Field:   fieldName,
				Rule:    ruleName,
				Message: fmt.Sprintf("%s must be a valid date", fieldName),
				Value:   value,
			}
		}

	case "in":
		values := strings.Split(parameter, "|")
		if str, ok := value.(string); ok && !v.contains(values, str) {
			return &ValidationError{
				Field:   fieldName,
				Rule:    ruleName,
				Message: fmt.Sprintf("%s must be one of: %s", fieldName, parameter),
				Value:   value,
			}
		}
	}

	return nil
}

// Helper validation methods

func (v *Validator) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case int, int8, int16, int32, int64:
		return v == 0
	case uint, uint8, uint16, uint32, uint64:
		return v == 0
	case float32, float64:
		return v == 0
	case bool:
		return !v
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	}

	return false
}

func (v *Validator) toFloat(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func (v *Validator) isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func (v *Validator) isValidURL(url string) bool {
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return urlRegex.MatchString(url)
}

func (v *Validator) isValidPhone(phone string) bool {
	// Remove common separators
	cleaned := regexp.MustCompile(`[\s\-\(\)]+`).ReplaceAllString(phone, "")
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(cleaned)
}

func (v *Validator) isValidDate(date string) bool {
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"02/01/2006",
		"01-02-2006",
	}

	for _, format := range formats {
		if _, err := time.Parse(format, date); err == nil {
			return true
		}
	}
	return false
}

func (v *Validator) isValidSKU(sku string) bool {
	// SKU should contain only alphanumeric characters, hyphens, and underscores
	skuRegex := regexp.MustCompile(`^[A-Za-z0-9\-_]+$`)
	return len(sku) >= 3 && len(sku) <= 50 && skuRegex.MatchString(sku)
}

func (v *Validator) validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return fmt.Errorf("password cannot exceed 128 characters")
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

func (v *Validator) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Custom validation methods for business logic

// ValidateProductInventory validates product inventory constraints
func (v *Validator) ValidateProductInventory(product *models.Product, requestedQuantity int) error {
	var errors []ValidationError

	if !product.IsAvailable() {
		errors = append(errors, ValidationError{
			Field:   "product",
			Rule:    "availability",
			Message: "Product is not available",
		})
	}

	if product.Stock < requestedQuantity {
		errors = append(errors, ValidationError{
			Field:   "quantity",
			Rule:    "stock_limit",
			Message: fmt.Sprintf("Requested quantity (%d) exceeds available stock (%d)", requestedQuantity, product.Stock),
		})
	}

	if product.MinStock > 0 && (product.Stock-requestedQuantity) < product.MinStock {
		errors = append(errors, ValidationError{
			Field:   "quantity",
			Rule:    "min_stock",
			Message: fmt.Sprintf("Order would reduce stock below minimum threshold (%d)", product.MinStock),
		})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidateWholesaleOrder validates wholesale order requirements
func (v *Validator) ValidateWholesaleOrder(product *models.Product, quantity int) error {
	var errors []ValidationError

	if product.WholesalePrice <= 0 {
		errors = append(errors, ValidationError{
			Field:   "product",
			Rule:    "wholesale_not_available",
			Message: "Wholesale pricing is not available for this product",
		})
	}

	if quantity < product.MinWholesaleQty {
		errors = append(errors, ValidationError{
			Field:   "quantity",
			Rule:    "min_wholesale_qty",
			Message: fmt.Sprintf("Minimum wholesale quantity is %d", product.MinWholesaleQty),
		})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}