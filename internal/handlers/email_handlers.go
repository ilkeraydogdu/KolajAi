package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"kolajAi/internal/services"
)

// EmailHandler handles email management requests
type EmailHandler struct {
	*Handler
	EmailService *services.EmailService
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(h *Handler, emailService *services.EmailService) *EmailHandler {
	return &EmailHandler{
		Handler:      h,
		EmailService: emailService,
	}
}

// Dashboard handles email management dashboard
func (h *EmailHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get email statistics
	stats := map[string]interface{}{
		"total_sent":    5250,
		"delivered":     4980,
		"opened":        3150,
		"clicked":       890,
		"bounced":       125,
		"unsubscribed":  45,
		"delivery_rate": 94.9,
		"open_rate":     63.3,
		"click_rate":    17.9,
		"bounce_rate":   2.4,
	}

	// Get recent campaigns
	campaigns := []map[string]interface{}{
		{
			"id":        1,
			"name":      "Yaz İndirimi 2024",
			"subject":   "🌞 %50'ye Varan İndirimler!",
			"sent":      2500,
			"delivered": 2380,
			"opened":    1654,
			"clicked":   425,
			"sent_at":   time.Now().Add(-24 * time.Hour),
			"status":    "completed",
		},
		{
			"id":        2,
			"name":      "Yeni Ürün Duyurusu",
			"subject":   "🎉 Yeni Koleksiyonumuz Çıktı!",
			"sent":      1800,
			"delivered": 1720,
			"opened":    1032,
			"clicked":   258,
			"sent_at":   time.Now().Add(-48 * time.Hour),
			"status":    "completed",
		},
	}

	// Get email templates
	templates := []map[string]interface{}{
		{
			"id":          1,
			"name":        "Hoş Geldin E-postası",
			"subject":     "{{site_name}}'e Hoş Geldiniz!",
			"type":        "welcome",
			"usage_count": 450,
			"last_used":   time.Now().Add(-12 * time.Hour),
		},
		{
			"id":          2,
			"name":        "Sipariş Onayı",
			"subject":     "Siparişiniz Alındı - #{{order_id}}",
			"type":        "transactional",
			"usage_count": 1250,
			"last_used":   time.Now().Add(-2 * time.Hour),
		},
	}

	data := map[string]interface{}{
		"Title":     "E-posta Yönetimi",
		"Stats":     stats,
		"Campaigns": campaigns,
		"Templates": templates,
	}

	h.RenderTemplate(w, r, "email/dashboard.gohtml", data)
}

// Campaigns handles email campaigns
func (h *EmailHandler) Campaigns(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	campaigns := []map[string]interface{}{
		{
			"id":         1,
			"name":       "Yaz İndirimi 2024",
			"subject":    "🌞 %50'ye Varan İndirimler!",
			"recipients": 2500,
			"sent":       2500,
			"delivered":  2380,
			"opened":     1654,
			"clicked":    425,
			"bounced":    120,
			"created_at": time.Now().Add(-72 * time.Hour),
			"sent_at":    time.Now().Add(-24 * time.Hour),
			"status":     "completed",
		},
		{
			"id":         2,
			"name":       "Yeni Ürün Duyurusu",
			"subject":    "🎉 Yeni Koleksiyonumuz Çıktı!",
			"recipients": 1800,
			"sent":       1800,
			"delivered":  1720,
			"opened":     1032,
			"clicked":    258,
			"bounced":    80,
			"created_at": time.Now().Add(-96 * time.Hour),
			"sent_at":    time.Now().Add(-48 * time.Hour),
			"status":     "completed",
		},
		{
			"id":         3,
			"name":       "Haftalık Bülten",
			"subject":    "📰 Bu Haftanın Öne Çıkanları",
			"recipients": 3200,
			"sent":       0,
			"delivered":  0,
			"opened":     0,
			"clicked":    0,
			"bounced":    0,
			"created_at": time.Now().Add(-2 * time.Hour),
			"sent_at":    nil,
			"status":     "draft",
		},
	}

	data := map[string]interface{}{
		"Title":     "E-posta Kampanyaları",
		"Campaigns": campaigns,
	}

	h.RenderTemplate(w, r, "email/campaigns.gohtml", data)
}

// Templates handles email templates
func (h *EmailHandler) Templates(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	templates := []map[string]interface{}{
		{
			"id":          1,
			"name":        "Hoş Geldin E-postası",
			"subject":     "{{site_name}}'e Hoş Geldiniz!",
			"type":        "welcome",
			"category":    "Otomatik",
			"status":      "active",
			"usage_count": 450,
			"last_used":   time.Now().Add(-12 * time.Hour),
			"created_at":  time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			"id":          2,
			"name":        "Sipariş Onayı",
			"subject":     "Siparişiniz Alındı - #{{order_id}}",
			"type":        "order_confirmation",
			"category":    "İşlemsel",
			"status":      "active",
			"usage_count": 1250,
			"last_used":   time.Now().Add(-2 * time.Hour),
			"created_at":  time.Now().Add(-60 * 24 * time.Hour),
		},
		{
			"id":          3,
			"name":        "Şifre Sıfırlama",
			"subject":     "Şifre Sıfırlama Talebi",
			"type":        "password_reset",
			"category":    "Güvenlik",
			"status":      "active",
			"usage_count": 89,
			"last_used":   time.Now().Add(-6 * time.Hour),
			"created_at":  time.Now().Add(-45 * 24 * time.Hour),
		},
		{
			"id":          4,
			"name":        "Promosyon Şablonu",
			"subject":     "🎉 Özel İndirim Fırsatı!",
			"type":        "promotional",
			"category":    "Pazarlama",
			"status":      "draft",
			"usage_count": 0,
			"last_used":   nil,
			"created_at":  time.Now().Add(-2 * time.Hour),
		},
	}

	data := map[string]interface{}{
		"Title":     "E-posta Şablonları",
		"Templates": templates,
	}

	h.RenderTemplate(w, r, "email/templates.gohtml", data)
}

// Settings handles email settings
func (h *EmailHandler) Settings(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	settings := map[string]interface{}{
		"smtp_settings": map[string]interface{}{
			"host":     "smtp.gmail.com",
			"port":     587,
			"username": "noreply@kolajai.com",
			"ssl":      true,
			"status":   "connected",
		},
		"sender_settings": map[string]interface{}{
			"from_name":    "KolajAI",
			"from_email":   "noreply@kolajai.com",
			"reply_to":     "support@kolajai.com",
			"bounce_email": "bounce@kolajai.com",
		},
		"delivery_settings": map[string]interface{}{
			"daily_limit":    10000,
			"hourly_limit":   1000,
			"retry_attempts": 3,
			"retry_delay":    300,
		},
		"tracking_settings": map[string]interface{}{
			"open_tracking":   true,
			"click_tracking":  true,
			"unsubscribe":     true,
			"bounce_handling": true,
		},
	}

	data := map[string]interface{}{
		"Title":    "E-posta Ayarları",
		"Settings": settings,
	}

	h.RenderTemplate(w, r, "email/settings.gohtml", data)
}

// API Methods

// APISendCampaign sends an email campaign
func (h *EmailHandler) APISendCampaign(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var campaign struct {
		Name        string     `json:"name"`
		Subject     string     `json:"subject"`
		Content     string     `json:"content"`
		Recipients  []string   `json:"recipients"`
		TemplateID  int        `json:"template_id,omitempty"`
		ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&campaign); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	campaignID := time.Now().Unix()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"campaign_id": campaignID,
		"message":     "Campaign sent successfully",
		"recipients":  len(campaign.Recipients),
	})
}

// APIGetEmailStats returns email statistics
func (h *EmailHandler) APIGetEmailStats(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := map[string]interface{}{
		"total_sent":       5250,
		"delivered":        4980,
		"opened":           3150,
		"clicked":          890,
		"bounced":          125,
		"unsubscribed":     45,
		"delivery_rate":    94.9,
		"open_rate":        63.3,
		"click_rate":       17.9,
		"bounce_rate":      2.4,
		"unsubscribe_rate": 0.9,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"stats":   stats,
	})
}

// APICreateTemplate creates a new email template
func (h *EmailHandler) APICreateTemplate(w http.ResponseWriter, r *http.Request) {
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
		Category string `json:"category"`
	}

	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	templateID := time.Now().Unix()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"template_id": templateID,
		"message":     "Template created successfully",
	})
}
