package components

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"time"

	"kolajAi/internal/config"
)

// Notification represents a notification in the system
type Notification struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Dismissible bool                   `json:"dismissible"`
	AutoDismiss bool                   `json:"auto_dismiss"`
	DismissTime time.Duration          `json:"dismiss_time"`
	Icon        string                 `json:"icon"`
	Color       string                 `json:"color"`
	CreatedAt   time.Time              `json:"created_at"`
	Target      string                 `json:"target"`
	Data        map[string]interface{} `json:"data"`
	Template    string                 `json:"template"`
}

// NotificationManager manages all notifications in the system
type NotificationManager struct {
	config      config.NotificationConfig
	templates   *template.Template
	typeConfigs map[string]config.NotificationType
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(templates *template.Template) (*NotificationManager, error) {
	cfg, ok := config.GetNotificationConfig()
	if !ok {
		return nil, fmt.Errorf("failed to get notification config")
	}

	return &NotificationManager{
		config:      cfg,
		templates:   templates,
		typeConfigs: cfg.Types,
	}, nil
}

// NewNotification creates a new notification with defaults based on type
func (m *NotificationManager) NewNotification(notificationType, title, message string) *Notification {
	typeConfig, exists := m.typeConfigs[notificationType]

	now := time.Now()
	id := fmt.Sprintf("notify-%d-%s", now.UnixNano(), strings.ToLower(notificationType))

	n := &Notification{
		ID:          id,
		Type:        notificationType,
		Title:       title,
		Message:     message,
		Dismissible: true,
		AutoDismiss: true,
		DismissTime: time.Duration(m.config.DefaultTTL) * time.Second,
		CreatedAt:   now,
		Data:        make(map[string]interface{}),
	}

	// Apply type-specific settings
	if exists {
		n.Icon = typeConfig.Icon
		n.Color = typeConfig.Color
		n.Template = typeConfig.Template
	} else {
		// Default values
		n.Icon = "info-circle"
		n.Color = "primary"
		n.Template = "notifications/default"
	}

	return n
}

// Info creates a new info notification
func (m *NotificationManager) Info(title, message string) *Notification {
	return m.NewNotification("info", title, message)
}

// Success creates a new success notification
func (m *NotificationManager) Success(title, message string) *Notification {
	return m.NewNotification("success", title, message)
}

// Warning creates a new warning notification
func (m *NotificationManager) Warning(title, message string) *Notification {
	return m.NewNotification("warning", title, message)
}

// Error creates a new error notification
func (m *NotificationManager) Error(title, message string) *Notification {
	return m.NewNotification("error", title, message)
}

// WithIcon sets the icon for the notification
func (n *Notification) WithIcon(icon string) *Notification {
	n.Icon = icon
	return n
}

// WithColor sets the color for the notification
func (n *Notification) WithColor(color string) *Notification {
	n.Color = color
	return n
}

// WithDismissible sets whether the notification is dismissible
func (n *Notification) WithDismissible(dismissible bool) *Notification {
	n.Dismissible = dismissible
	return n
}

// WithAutoDismiss sets whether the notification auto dismisses
func (n *Notification) WithAutoDismiss(autoDismiss bool) *Notification {
	n.AutoDismiss = autoDismiss
	return n
}

// WithDismissTime sets the time after which the notification auto dismisses
func (n *Notification) WithDismissTime(dismissTime time.Duration) *Notification {
	n.DismissTime = dismissTime
	return n
}

// WithTarget sets the target element for the notification
func (n *Notification) WithTarget(target string) *Notification {
	n.Target = target
	return n
}

// WithData adds custom data to the notification
func (n *Notification) WithData(key string, value interface{}) *Notification {
	n.Data[key] = value
	return n
}

// WithTemplate sets a custom template for the notification
func (n *Notification) WithTemplate(template string) *Notification {
	n.Template = template
	return n
}

// RenderToHTML renders the notification to HTML
func (n *Notification) RenderToHTML(templates *template.Template) (string, error) {
	var buf bytes.Buffer

	// Use specified template or fallback to a default template
	templateName := n.Template
	if templateName == "" {
		templateName = "notifications/default"
	}

	// Execute the template
	err := templates.ExecuteTemplate(&buf, templateName, n)
	if err != nil {
		return "", fmt.Errorf("failed to render notification template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

// ToMap converts the notification to a map for JSON serialization
func (n *Notification) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":           n.ID,
		"type":         n.Type,
		"title":        n.Title,
		"message":      n.Message,
		"dismissible":  n.Dismissible,
		"auto_dismiss": n.AutoDismiss,
		"dismiss_time": n.DismissTime.Milliseconds(),
		"icon":         n.Icon,
		"color":        n.Color,
		"created_at":   n.CreatedAt.Format(time.RFC3339),
		"target":       n.Target,
		"data":         n.Data,
	}
}

// ToScript converts the notification to a JavaScript snippet for client-side rendering
func (n *Notification) ToScript() string {
	// Simple implementation for now
	return fmt.Sprintf(`
	window.addEventListener('DOMContentLoaded', function() {
		var notification = %s;
		if (window.KolajAI && window.KolajAI.notifications) {
			window.KolajAI.notifications.show(notification);
		} else {
			console.warn('KolajAI notifications module not loaded');
			alert('%s: %s');
		}
	});
	`, toJSON(n.ToMap()), n.Title, n.Message)
}

// toJSON converts a value to a JSON string
func toJSON(v interface{}) string {
	if v == nil {
		return "null"
	}

	switch val := v.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, val)
	case int, int32, int64, float32, float64, bool:
		return fmt.Sprintf("%v", val)
	case []interface{}:
		items := make([]string, len(val))
		for i, item := range val {
			items[i] = toJSON(item)
		}
		return fmt.Sprintf("[%s]", strings.Join(items, ","))
	case map[string]interface{}:
		items := make([]string, 0, len(val))
		for k, v := range val {
			items = append(items, fmt.Sprintf(`"%s":%s`, k, toJSON(v)))
		}
		return fmt.Sprintf("{%s}", strings.Join(items, ","))
	}

	return "null"
}
