package marketplace

import (
	"fmt"
	"strconv"
)

// getString safely extracts a string value from a map
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", val)
	}
	return ""
}

// getStringWithDefault safely extracts a string value from a map with a default
func getStringWithDefault(data map[string]interface{}, key string, defaultValue string) string {
	if val := getString(data, key); val != "" {
		return val
	}
	return defaultValue
}

// getInt safely extracts an int value from a map
func getInt(data map[string]interface{}, key string) int {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return 0
}

// getIntWithDefault safely extracts an int value from a map with a default
func getIntWithDefault(data map[string]interface{}, key string, defaultValue int) int {
	if val := getInt(data, key); val != 0 {
		return val
	}
	return defaultValue
}

// getFloat64 safely extracts a float64 value from a map
func getFloat64(data map[string]interface{}, key string) float64 {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return 0.0
}

// getFloat64WithDefault safely extracts a float64 value from a map with a default
func getFloat64WithDefault(data map[string]interface{}, key string, defaultValue float64) float64 {
	if val := getFloat64(data, key); val != 0.0 {
		return val
	}
	return defaultValue
}

// getBool safely extracts a bool value from a map
func getBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			return v == "true" || v == "1" || v == "yes"
		case int:
			return v != 0
		}
	}
	return false
}

// getBoolWithDefault safely extracts a bool value from a map with a default
func getBoolWithDefault(data map[string]interface{}, key string, defaultValue bool) bool {
	if _, ok := data[key]; ok {
		return getBool(data, key)
	}
	return defaultValue
}