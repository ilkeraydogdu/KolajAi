package validation

import (
	"encoding/json"
	"fmt"
)

// FormValidator is a specialized validator for web forms
type FormValidator struct {
	validator *DefaultValidator
	schemas   map[string]FormSchema
}

// NewFormValidator creates a new form validator
func NewFormValidator() *FormValidator {
	return &FormValidator{
		validator: New(),
		schemas:   RegisterSchemas(),
	}
}

// ValidateForm validates a form against its schema
func (v *FormValidator) ValidateForm(formName string, data map[string]string) (bool, map[string][]string) {
	schema, exists := v.schemas[formName]
	if !exists {
		return false, map[string][]string{
			"_form": {"Form şeması bulunamadı: " + formName},
		}
	}

	// Convert schema to validation rules
	rules := make(map[string][]string)
	for fieldName, fieldRules := range schema.Fields {
		fieldValidations := []string{}

		if fieldRules.Required {
			fieldValidations = append(fieldValidations, "required")
		}

		if fieldRules.MinLength > 0 {
			fieldValidations = append(fieldValidations, fmt.Sprintf("min:%d", fieldRules.MinLength))
		}

		if fieldRules.MaxLength > 0 {
			fieldValidations = append(fieldValidations, fmt.Sprintf("max:%d", fieldRules.MaxLength))
		}

		if fieldRules.Pattern != "" {
			// Special handling for known patterns
			if fieldRules.Pattern == `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$` {
				fieldValidations = append(fieldValidations, "email")
			} else if fieldRules.Pattern == `^0[0-9]{3} [0-9]{3} [0-9]{4}$` {
				fieldValidations = append(fieldValidations, "phone")
			} else {
				// TODO: Add custom pattern validation
			}
		}

		rules[fieldName] = fieldValidations
	}

	return v.validator.ValidateMap(data, rules)
}

// GetSchemaJSON returns the JSON representation of a form schema
func (v *FormValidator) GetSchemaJSON(formName string) (string, error) {
	schema, exists := v.schemas[formName]
	if !exists {
		return "", nil
	}

	// Convert to client-side friendly format
	clientSchema := make(map[string]map[string]interface{})

	for fieldName, fieldRules := range schema.Fields {
		clientSchema[fieldName] = map[string]interface{}{
			"required": fieldRules.Required,
			"message":  fieldRules.Message,
		}

		if fieldRules.MinLength > 0 {
			clientSchema[fieldName]["minLength"] = fieldRules.MinLength
		}

		if fieldRules.MaxLength > 0 {
			clientSchema[fieldName]["maxLength"] = fieldRules.MaxLength
		}

		if fieldRules.Pattern != "" {
			clientSchema[fieldName]["pattern"] = fieldRules.Pattern
		}
	}

	jsonData, err := json.Marshal(clientSchema)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// RegisterSchema adds a new form schema
func (v *FormValidator) RegisterSchema(name string, schema FormSchema) {
	v.schemas[name] = schema
}
