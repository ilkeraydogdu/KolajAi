package config

// NotificationType represents a notification type configuration
type NotificationType struct {
	Icon     string `json:"icon"`
	Color    string `json:"color"`
	Template string `json:"template"`
}

// UpdateNotificationConfig updates the NotificationConfig struct to include required fields
func UpdateNotificationConfig() {
	// Update the NotificationConfig struct in config.go to include these fields
	// This is just a placeholder function to document the changes needed
}

// GetDefaultNotificationTypes returns default notification types
func GetDefaultNotificationTypes() map[string]NotificationType {
	return map[string]NotificationType{
		"info": {
			Icon:     "info-circle",
			Color:    "primary",
			Template: "notifications/info",
		},
		"success": {
			Icon:     "check-circle",
			Color:    "success",
			Template: "notifications/success",
		},
		"warning": {
			Icon:     "exclamation-triangle",
			Color:    "warning",
			Template: "notifications/warning",
		},
		"error": {
			Icon:     "exclamation-circle",
			Color:    "danger",
			Template: "notifications/error",
		},
	}
}

// GetNotificationTypes returns configured notification types
func GetNotificationTypes() map[string]NotificationType {
	// In a real implementation, this would load from configuration
	return GetDefaultNotificationTypes()
}
