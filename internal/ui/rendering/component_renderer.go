package rendering

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"kolajAi/internal/ui/components"
)

// ComponentData, UI bileşenlerine aktarılacak verilerin genel yapısı
type ComponentData map[string]interface{}

// ComponentRenderer, UI bileşenlerini işleyen yardımcı
type ComponentRenderer struct {
	templates       *template.Template
	notificationMgr *components.NotificationManager
}

// NewComponentRenderer, yeni bir bileşen işleyici oluşturur
func NewComponentRenderer(templates *template.Template) (*ComponentRenderer, error) {
	notifyMgr, err := components.NewNotificationManager(templates)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification manager: %w", err)
	}

	return &ComponentRenderer{
		templates:       templates,
		notificationMgr: notifyMgr,
	}, nil
}

// RenderComponent, belirtilen bileşeni verilen verilerle işler
func (r *ComponentRenderer) RenderComponent(name string, data interface{}) (template.HTML, error) {
	var buf bytes.Buffer
	err := r.templates.ExecuteTemplate(&buf, name, data)
	if err != nil {
		return "", fmt.Errorf("failed to render component %s: %w", name, err)
	}
	return template.HTML(buf.String()), nil
}

// RenderPartial renders a partial template to HTML
func (r *ComponentRenderer) RenderPartial(name string, data interface{}) (template.HTML, error) {
	var buf bytes.Buffer
	err := r.templates.ExecuteTemplate(&buf, name, data)
	if err != nil {
		return "", fmt.Errorf("failed to render partial %s: %w", name, err)
	}
	return template.HTML(buf.String()), nil
}

// InfoNotification creates and renders an info notification
func (r *ComponentRenderer) InfoNotification(title, message string) (template.HTML, error) {
	notification := r.notificationMgr.Info(title, message)
	html, err := notification.RenderToHTML(r.templates)
	if err != nil {
		return "", err
	}
	return template.HTML(html), nil
}

// SuccessNotification creates and renders a success notification
func (r *ComponentRenderer) SuccessNotification(title, message string) (template.HTML, error) {
	notification := r.notificationMgr.Success(title, message)
	html, err := notification.RenderToHTML(r.templates)
	if err != nil {
		return "", err
	}
	return template.HTML(html), nil
}

// WarningNotification creates and renders a warning notification
func (r *ComponentRenderer) WarningNotification(title, message string) (template.HTML, error) {
	notification := r.notificationMgr.Warning(title, message)
	html, err := notification.RenderToHTML(r.templates)
	if err != nil {
		return "", err
	}
	return template.HTML(html), nil
}

// ErrorNotification creates and renders an error notification
func (r *ComponentRenderer) ErrorNotification(title, message string) (template.HTML, error) {
	notification := r.notificationMgr.Error(title, message)
	html, err := notification.RenderToHTML(r.templates)
	if err != nil {
		return "", err
	}
	return template.HTML(html), nil
}

// CustomNotification creates and renders a custom notification
func (r *ComponentRenderer) CustomNotification(notificationType, title, message string) (template.HTML, error) {
	notification := r.notificationMgr.NewNotification(notificationType, title, message)
	html, err := notification.RenderToHTML(r.templates)
	if err != nil {
		return "", err
	}
	return template.HTML(html), nil
}

// GetNotificationManager returns the notification manager
func (r *ComponentRenderer) GetNotificationManager() *components.NotificationManager {
	return r.notificationMgr
}

// RenderBreadcrumb renders a breadcrumb component
func (r *ComponentRenderer) RenderBreadcrumb(items []map[string]interface{}) (template.HTML, error) {
	data := map[string]interface{}{
		"items": items,
	}
	return r.RenderComponent("components/breadcrumb", data)
}

// RenderCard renders a card component
func (r *ComponentRenderer) RenderCard(title, content string, options map[string]interface{}) (template.HTML, error) {
	data := map[string]interface{}{
		"title":   title,
		"content": template.HTML(content),
	}

	// Merge options
	for k, v := range options {
		data[k] = v
	}

	return r.RenderComponent("components/card", data)
}

// RenderTable renders a table component
func (r *ComponentRenderer) RenderTable(headers []string, rows [][]string, options map[string]interface{}) (template.HTML, error) {
	data := map[string]interface{}{
		"headers": headers,
		"rows":    rows,
	}

	// Merge options
	for k, v := range options {
		data[k] = v
	}

	return r.RenderComponent("components/table", data)
}

// RenderButton renders a button component
func (r *ComponentRenderer) RenderButton(text, url string, options map[string]interface{}) (template.HTML, error) {
	data := map[string]interface{}{
		"text": text,
		"url":  url,
	}

	// Set defaults
	if _, ok := options["color"]; !ok {
		data["color"] = "primary"
	}

	if _, ok := options["size"]; !ok {
		data["size"] = "md"
	}

	// Merge options
	for k, v := range options {
		data[k] = v
	}

	return r.RenderComponent("components/button", data)
}

// RenderAlert renders an alert component
func (r *ComponentRenderer) RenderAlert(messageOrAlertType string, options interface{}) (template.HTML, error) {
	// Önce parametrelere göre fonksiyon davranışını belirle
	if options == nil {
		// Tek parametre durumu - yalnızca mesaj verildiyse
		return r.RenderComponent("alert-color", ComponentData{
			"Type":    "info",
			"Content": messageOrAlertType,
		})
	}

	// Options map[string]interface{} tipinde mi?
	if optionsMap, ok := options.(map[string]interface{}); ok {
		// Eski fonksiyon imzası için: RenderAlert(message string, options map[string]interface{})
		data := map[string]interface{}{
			"message": template.HTML(messageOrAlertType),
		}

		// Set defaults
		if _, ok := optionsMap["type"]; !ok {
			data["type"] = "info"
		}

		if _, ok := optionsMap["dismissible"]; !ok {
			data["dismissible"] = true
		}

		// Merge options
		for k, v := range optionsMap {
			data[k] = v
		}

		return r.RenderComponent("components/alert", data)
	}

	// Options string tipinde mi?
	if content, ok := options.(string); ok {
		// Yeni fonksiyon imzası için: RenderAlert(alertType string, content string)
		return r.RenderComponent("alert-color", ComponentData{
			"Type":    messageOrAlertType,
			"Content": content,
		})
	}

	// Desteklenmeyen parametre tipi
	return template.HTML(""), fmt.Errorf("unsupported parameter types in RenderAlert")
}

// RenderModal renders a modal component
func (r *ComponentRenderer) RenderModal(id, title, content string, options map[string]interface{}) (template.HTML, error) {
	data := map[string]interface{}{
		"id":      id,
		"title":   title,
		"content": template.HTML(content),
	}

	// Set defaults
	if _, ok := options["size"]; !ok {
		data["size"] = "medium"
	}

	// Merge options
	for k, v := range options {
		data[k] = v
	}

	return r.RenderComponent("components/modal", data)
}

// RenderPagination renders a pagination component
func (r *ComponentRenderer) RenderPagination(currentPage, totalPages int, baseURL string, options map[string]interface{}) (template.HTML, error) {
	data := map[string]interface{}{
		"current_page": currentPage,
		"total_pages":  totalPages,
		"base_url":     baseURL,
	}

	// Set defaults
	if _, ok := options["show_first_last"]; !ok {
		data["show_first_last"] = true
	}

	if _, ok := options["size"]; !ok {
		data["size"] = "md"
	}

	// Merge options
	for k, v := range options {
		data[k] = v
	}

	return r.RenderComponent("components/pagination", data)
}

// CreateTemplateFuncs returns a map of template functions for rendering components
func (r *ComponentRenderer) CreateTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"component": func(name string, data interface{}) (template.HTML, error) {
			return r.RenderComponent(name, data)
		},
		"partial": func(name string, data interface{}) (template.HTML, error) {
			return r.RenderPartial(name, data)
		},
		"notification": func(type_, title, message string) (template.HTML, error) {
			return r.CustomNotification(type_, title, message)
		},
		"infoNotification": func(title, message string) (template.HTML, error) {
			return r.InfoNotification(title, message)
		},
		"successNotification": func(title, message string) (template.HTML, error) {
			return r.SuccessNotification(title, message)
		},
		"warningNotification": func(title, message string) (template.HTML, error) {
			return r.WarningNotification(title, message)
		},
		"errorNotification": func(title, message string) (template.HTML, error) {
			return r.ErrorNotification(title, message)
		},
		"breadcrumb": func(items []map[string]interface{}) (template.HTML, error) {
			return r.RenderBreadcrumb(items)
		},
		"card": func(title, content string, options map[string]interface{}) (template.HTML, error) {
			return r.RenderCard(title, content, options)
		},
		"table": func(headers []string, rows [][]string, options map[string]interface{}) (template.HTML, error) {
			return r.RenderTable(headers, rows, options)
		},
		"button": func(text, url string, options map[string]interface{}) (template.HTML, error) {
			return r.RenderButton(text, url, options)
		},
		"alert": func(message string, options map[string]interface{}) (template.HTML, error) {
			return r.RenderAlert(message, options)
		},
		"modal": func(id, title, content string, options map[string]interface{}) (template.HTML, error) {
			return r.RenderModal(id, title, content, options)
		},
		"pagination": func(currentPage, totalPages int, baseURL string, options map[string]interface{}) (template.HTML, error) {
			return r.RenderPagination(currentPage, totalPages, baseURL, options)
		},
		// Helper functions
		"join": strings.Join,
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call, needs even number of arguments")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}
}

// AlertData, uyarı bileşeni için veri yapısı
type AlertData struct {
	Type    string // primary, secondary, success, danger, warning, info, dark
	Icon    string // Material icon name
	Title   string
	Content string
}

// RenderAlertWithIcon, ikonlu alert bileşenini işler
func (r *ComponentRenderer) RenderAlertWithIcon(alertType string, icon string, title string, content string) (template.HTML, error) {
	data := ComponentData{
		"Type":    alertType,
		"Icon":    icon,
		"Title":   title,
		"Content": content,
	}
	return r.RenderComponent("alert-color-with-icon", data)
}

// RenderBorderAlert, kenarlıklı alert bileşenini işler
func (r *ComponentRenderer) RenderBorderAlert(alertType string, content string) (template.HTML, error) {
	data := ComponentData{
		"Type":    alertType,
		"Content": content,
	}
	return r.RenderComponent("alert-border", data)
}

// RenderBorderAlertWithIcon, ikonlu kenarlıklı alert bileşenini işler
func (r *ComponentRenderer) RenderBorderAlertWithIcon(alertType string, icon string, title string, content string) (template.HTML, error) {
	data := ComponentData{
		"Type":    alertType,
		"Icon":    icon,
		"Title":   title,
		"Content": content,
	}
	return r.RenderComponent("alert-border-with-icon", data)
}

// NotificationCardData, bildirim kartı bileşeni için veri yapısı
type NotificationCardData struct {
	Type     string // primary, info, warning, danger, success
	Icon     string // Material icon name
	Title    string
	Function string // JavaScript function to call
}

// RenderNotificationCard, bildirim kartı bileşenini işler
func (r *ComponentRenderer) RenderNotificationCard(notifType string, icon string, title string, function string) (template.HTML, error) {
	data := ComponentData{
		"Type":     notifType,
		"Icon":     icon,
		"Title":    title,
		"Function": function,
	}
	return r.RenderComponent("notification-card", data)
}

// NotificationDropdownItemData, bildirim dropdown öğesi için veri yapısı
type NotificationDropdownItemData struct {
	ImageURL    string
	InitialsBg  string // primary, danger, etc.
	Initials    string // e.g. "RS" for user initials
	Title       string
	Description string
	Time        string
}

// NotificationDropdownOptionData, bildirim dropdown seçeneği için veri yapısı
type NotificationDropdownOptionData struct {
	Icon string
	Text string
}

// NotificationDropdownData, bildirim dropdown bileşeni için veri yapısı
type NotificationDropdownData struct {
	Title   string
	Options []NotificationDropdownOptionData
	Divider bool
	Items   []NotificationDropdownItemData
}

// RenderNotificationDropdown, bildirim dropdown bileşenini işler
func (r *ComponentRenderer) RenderNotificationDropdown(data NotificationDropdownData) (template.HTML, error) {
	componentData := ComponentData{
		"Title":   data.Title,
		"Options": data.Options,
		"Divider": data.Divider,
		"Items":   data.Items,
	}
	return r.RenderComponent("notification-dropdown", componentData)
}

// GetNotificationJS, bildirim JS kodunu döndürür
func (r *ComponentRenderer) GetNotificationJS() (template.HTML, error) {
	return r.RenderComponent("notification-js", nil)
}

// RenderAlertToResponse, uyarı bileşenini doğrudan HTTP yanıtına işler
func (r *ComponentRenderer) RenderAlertToResponse(w http.ResponseWriter, data AlertData, withIcon bool, withBorder bool) error {
	var templateName string
	if withIcon {
		if withBorder {
			templateName = "alert-border-with-icon"
		} else {
			templateName = "alert-color-with-icon"
		}
	} else {
		if withBorder {
			templateName = "alert-border"
		} else {
			templateName = "alert-color"
		}
	}

	// ComponentData oluştur
	componentData := ComponentData{
		"Type":    data.Type,
		"Icon":    data.Icon,
		"Title":   data.Title,
		"Content": data.Content,
	}

	// Şablonu işle
	html, err := r.RenderComponent(templateName, componentData)
	if err != nil {
		return err
	}

	// HTML yanıtı gönder
	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write([]byte(html))
	return err
}

// RenderNotificationCardToResponse, bildirim kartı bileşenini doğrudan HTTP yanıtına işler
func (r *ComponentRenderer) RenderNotificationCardToResponse(w http.ResponseWriter, data NotificationCardData) error {
	componentData := ComponentData{
		"Type":     data.Type,
		"Icon":     data.Icon,
		"Title":    data.Title,
		"Function": data.Function,
	}

	html, err := r.RenderComponent("notification-card", componentData)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write([]byte(html))
	return err
}

// RenderNotificationDropdownToResponse, bildirim dropdown bileşenini doğrudan HTTP yanıtına işler
func (r *ComponentRenderer) RenderNotificationDropdownToResponse(w http.ResponseWriter, data NotificationDropdownData) error {
	componentData := ComponentData{
		"Title":   data.Title,
		"Options": data.Options,
		"Divider": data.Divider,
		"Items":   data.Items,
	}

	html, err := r.RenderComponent("notification-dropdown", componentData)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write([]byte(html))
	return err
}

// GetAlertHTML, bir uyarı bileşeni için HTML döndürür
func (r *ComponentRenderer) GetAlertHTML(data AlertData, withIcon bool, withBorder bool) (template.HTML, error) {
	var templateName string
	if withIcon {
		if withBorder {
			templateName = "alert-border-with-icon"
		} else {
			templateName = "alert-color-with-icon"
		}
	} else {
		if withBorder {
			templateName = "alert-border"
		} else {
			templateName = "alert-color"
		}
	}

	// ComponentData oluştur
	componentData := ComponentData{
		"Type":    data.Type,
		"Icon":    data.Icon,
		"Title":   data.Title,
		"Content": data.Content,
	}

	return r.RenderComponent(templateName, componentData)
}
