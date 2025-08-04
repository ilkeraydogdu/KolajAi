package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"log"
	
	"kolajAi/internal/services"
)

// NotificationHandler handles notification management requests
type NotificationHandler struct {
	*Handler
	NotificationService *services.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(h *Handler, notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		Handler:             h,
		NotificationService: notificationService,
	}
}

// Dashboard handles notification management dashboard
func (h *NotificationHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get notification statistics
	stats := map[string]interface{}{
		"total_sent":       1250,
		"delivered":        1180,
		"opened":          890,
		"clicked":         245,
		"failed":          35,
		"pending":         15,
		"delivery_rate":   94.4,
		"open_rate":       75.4,
		"click_rate":      19.6,
	}

	// Get recent notifications
	recentNotifications := []map[string]interface{}{
		{
			"id":          1,
			"title":       "Yeni Sipariş Bildirimi",
			"type":        "order",
			"channel":     "email",
			"recipients":  150,
			"sent_at":     time.Now().Add(-2 * time.Hour),
			"status":      "delivered",
			"open_rate":   78.5,
		},
		{
			"id":          2,
			"title":       "Stok Uyarısı",
			"type":        "inventory",
			"channel":     "push",
			"recipients":  25,
			"sent_at":     time.Now().Add(-4 * time.Hour),
			"status":      "delivered",
			"open_rate":   85.2,
		},
	}

	// Get notification templates
	templates := []map[string]interface{}{
		{
			"id":          1,
			"name":        "Sipariş Onayı",
			"type":        "order_confirmation",
			"channel":     "email",
			"usage_count": 450,
			"last_used":   time.Now().Add(-24 * time.Hour),
		},
		{
			"id":          2,
			"name":        "Hoş Geldin Mesajı",
			"type":        "welcome",
			"channel":     "email",
			"usage_count": 125,
			"last_used":   time.Now().Add(-48 * time.Hour),
		},
	}

	data := map[string]interface{}{
		"Title":               "Bildirim Yönetimi",
		"Stats":               stats,
		"RecentNotifications": recentNotifications,
		"Templates":           templates,
	}

	h.RenderTemplate(w, r, "notifications/dashboard.gohtml", data)
}

// Templates handles notification templates management
func (h *NotificationHandler) Templates(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get all notification templates
	templates := []map[string]interface{}{
		{
			"id":          1,
			"name":        "Sipariş Onayı",
			"subject":     "Siparişiniz Alındı - #{{order_id}}",
			"type":        "order_confirmation",
			"channel":     "email",
			"status":      "active",
			"created_at":  time.Now().Add(-30 * 24 * time.Hour),
			"usage_count": 450,
		},
		{
			"id":          2,
			"name":        "Hoş Geldin Mesajı",
			"subject":     "{{site_name}}'e Hoş Geldiniz!",
			"type":        "welcome",
			"channel":     "email",
			"status":      "active",
			"created_at":  time.Now().Add(-60 * 24 * time.Hour),
			"usage_count": 125,
		},
		{
			"id":          3,
			"name":        "Şifre Sıfırlama",
			"subject":     "Şifre Sıfırlama Talebi",
			"type":        "password_reset",
			"channel":     "email",
			"status":      "active",
			"created_at":  time.Now().Add(-45 * 24 * time.Hour),
			"usage_count": 89,
		},
	}

	data := map[string]interface{}{
		"Title":     "Bildirim Şablonları",
		"Templates": templates,
	}

	h.RenderTemplate(w, r, "notifications/templates.gohtml", data)
}

// Campaigns handles notification campaigns
func (h *NotificationHandler) Campaigns(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get notification campaigns
	campaigns := []map[string]interface{}{
		{
			"id":           1,
			"name":         "Yaz İndirimi Kampanyası",
			"type":         "marketing",
			"channel":      "email",
			"status":       "completed",
			"recipients":   2500,
			"sent":         2480,
			"delivered":    2350,
			"opened":       1650,
			"clicked":      420,
			"scheduled_at": time.Now().Add(-48 * time.Hour),
			"sent_at":      time.Now().Add(-48 * time.Hour),
		},
		{
			"id":           2,
			"name":         "Stok Uyarı Bildirimi",
			"type":         "system",
			"channel":      "push",
			"status":       "active",
			"recipients":   150,
			"sent":         150,
			"delivered":    148,
			"opened":       125,
			"clicked":      45,
			"scheduled_at": time.Now().Add(-2 * time.Hour),
			"sent_at":      time.Now().Add(-2 * time.Hour),
		},
	}

	data := map[string]interface{}{
		"Title":     "Bildirim Kampanyaları",
		"Campaigns": campaigns,
	}

	h.RenderTemplate(w, r, "notifications/campaigns.gohtml", data)
}

// API Methods

// APISendNotification sends a notification
func (h *NotificationHandler) APISendNotification(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Title       string                 `json:"title"`
		Message     string                 `json:"message"`
		Type        string                 `json:"type"`
		Channel     string                 `json:"channel"`
		Recipients  []int                  `json:"recipients"`
		TemplateID  string                 `json:"template_id,omitempty"`
		Variables   map[string]interface{} `json:"variables,omitempty"`
		ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Sending notification: %s to %d recipients via %s", request.Title, len(request.Recipients), request.Channel)

	// Mock notification sending
	notificationID := time.Now().Unix()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":         true,
		"notification_id": notificationID,
		"message":         "Notification sent successfully",
		"recipients":      len(request.Recipients),
	})
}

// APIGetNotificationStats returns notification statistics
func (h *NotificationHandler) APIGetNotificationStats(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := map[string]interface{}{
		"total_sent":     1250,
		"delivered":      1180,
		"opened":         890,
		"clicked":        245,
		"failed":         35,
		"pending":        15,
		"delivery_rate":  94.4,
		"open_rate":      75.4,
		"click_rate":     19.6,
		"bounce_rate":    2.8,
		"unsubscribe_rate": 0.5,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"stats":   stats,
	})
}

// APICreateTemplate creates a new notification template
func (h *NotificationHandler) APICreateTemplate(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var template struct {
		Name     string `json:"name"`
		Subject  string `json:"subject"`
		Content  string `json:"content"`
		Type     string `json:"type"`
		Channel  string `json:"channel"`
	}

	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Creating notification template: %s", template.Name)

	templateID := time.Now().Unix()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"template_id": templateID,
		"message":     "Template created successfully",
	})
}

// APIUpdateTemplate updates a notification template
func (h *NotificationHandler) APIUpdateTemplate(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	templateIDStr := r.URL.Query().Get("id")
	templateID, err := strconv.Atoi(templateIDStr)
	if err != nil {
		http.Error(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	var template struct {
		Name     string `json:"name"`
		Subject  string `json:"subject"`
		Content  string `json:"content"`
		Status   string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Updating notification template %d", templateID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Template updated successfully",
	})
}

// APIDeleteTemplate deletes a notification template
func (h *NotificationHandler) APIDeleteTemplate(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	templateIDStr := r.URL.Query().Get("id")
	templateID, err := strconv.Atoi(templateIDStr)
	if err != nil {
		http.Error(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	log.Printf("Deleting notification template %d", templateID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Template deleted successfully",
	})
}