package validation

import (
	"fmt"
	"strings"
)

// Validator defines the basic validation operations
type Validator interface {
	Validate(field string, value string, rules ...string) (bool, []string)
	ValidateMap(data map[string]string, rules map[string][]string) (bool, map[string][]string)
}

// DefaultValidator is the standard implementation of Validator
type DefaultValidator struct {
	rules   map[string]ValidationRule
	errors  map[string][]string
	options map[string]interface{}
}

// ValidationRule defines a validation function
type ValidationRule func(value string, options ...interface{}) (bool, string)

// New creates a new DefaultValidator with standard rules
func New(options ...map[string]interface{}) *DefaultValidator {
	v := &DefaultValidator{
		rules:   make(map[string]ValidationRule),
		errors:  make(map[string][]string),
		options: make(map[string]interface{}),
	}

	// Add default options if provided
	if len(options) > 0 {
		for k, val := range options[0] {
			v.options[k] = val
		}
	}

	// Register default rules
	v.RegisterRule("required", RequiredRule)
	v.RegisterRule("email", EmailRule)
	v.RegisterRule("min", MinLengthRule)
	v.RegisterRule("max", MaxLengthRule)
	v.RegisterRule("numeric", NumericRule)
	v.RegisterRule("alpha", AlphaRule)
	v.RegisterRule("alphanumeric", AlphaNumericRule)
	v.RegisterRule("phone", PhoneRule)
	v.RegisterRule("url", URLRule)
	v.RegisterRule("date", DateRule)
	v.RegisterRule("match", MatchRule)

	return v
}

// RegisterRule adds a new validation rule
func (v *DefaultValidator) RegisterRule(name string, rule ValidationRule) {
	v.rules[name] = rule
}

// Validate checks a single field against the given rules
func (v *DefaultValidator) Validate(field string, value string, rules ...string) (bool, []string) {
	v.errors[field] = []string{}
	valid := true

	for _, ruleStr := range rules {
		// Parse rule and options
		ruleParts := strings.Split(ruleStr, ":")
		ruleName := ruleParts[0]

		var ruleOptions []interface{}
		if len(ruleParts) > 1 {
			// Has options
			for _, opt := range strings.Split(ruleParts[1], ",") {
				ruleOptions = append(ruleOptions, opt)
			}
		}

		// Get the rule function
		ruleFunc, exists := v.rules[ruleName]
		if !exists {
			v.errors[field] = append(v.errors[field], fmt.Sprintf("Unknown validation rule: %s", ruleName))
			valid = false
			continue
		}

		// Apply the rule
		ruleValid, errorMsg := ruleFunc(value, ruleOptions...)
		if !ruleValid {
			v.errors[field] = append(v.errors[field], errorMsg)
			valid = false
		}
	}

	return valid, v.errors[field]
}

// ValidateMap validates multiple fields at once
func (v *DefaultValidator) ValidateMap(data map[string]string, rules map[string][]string) (bool, map[string][]string) {
	v.errors = make(map[string][]string)
	valid := true

	for field, fieldRules := range rules {
		value, exists := data[field]
		if !exists {
			value = "" // Field doesn't exist, treat as empty
		}

		fieldValid, _ := v.Validate(field, value, fieldRules...)
		if !fieldValid {
			valid = false
		}
	}

	return valid, v.errors
}

// GetErrors returns all validation errors
func (v *DefaultValidator) GetErrors() map[string][]string {
	return v.errors
}

// GetFirstError returns the first error for a field, or empty string if none
func (v *DefaultValidator) GetFirstError(field string) string {
	if errors, exists := v.errors[field]; exists && len(errors) > 0 {
		return errors[0]
	}
	return ""
}
