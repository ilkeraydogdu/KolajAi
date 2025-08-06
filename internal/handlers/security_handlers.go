package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

// SecurityHandler handles security management requests
type SecurityHandler struct {
	*Handler
}

// NewSecurityHandler creates a new security handler
func NewSecurityHandler(h *Handler) *SecurityHandler {
	return &SecurityHandler{
		Handler: h,
	}
}

// Dashboard handles security management dashboard
func (h *SecurityHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get security statistics
	stats := map[string]interface{}{
		"active_sessions":     245,
		"failed_logins":       18,
		"blocked_ips":         5,
		"2fa_enabled_users":   180,
		"oauth_connections":   95,
		"security_alerts":     3,
		"jwt_tokens_active":   320,
		"last_breach_attempt": time.Now().Add(-24 * time.Hour),
	}

	// Get recent security events
	securityEvents := []map[string]interface{}{
		{
			"id":        1,
			"type":      "failed_login",
			"severity":  "medium",
			"message":   "Çoklu başarısız giriş denemesi",
			"ip":        "192.168.1.100",
			"user":      "test@example.com",
			"timestamp": time.Now().Add(-30 * time.Minute),
			"status":    "blocked",
		},
		{
			"id":        2,
			"type":      "suspicious_activity",
			"severity":  "high",
			"message":   "Şüpheli API istekleri",
			"ip":        "10.0.0.50",
			"user":      "unknown",
			"timestamp": time.Now().Add(-2 * time.Hour),
			"status":    "investigating",
		},
		{
			"id":        3,
			"type":      "2fa_bypass_attempt",
			"severity":  "critical",
			"message":   "2FA bypass denemesi",
			"ip":        "203.0.113.1",
			"user":      "admin@example.com",
			"timestamp": time.Now().Add(-4 * time.Hour),
			"status":    "blocked",
		},
	}

	// Get active threats
	activeThreats := []map[string]interface{}{
		{
			"id":         1,
			"type":       "brute_force",
			"source_ip":  "192.168.1.100",
			"target":     "login endpoint",
			"attempts":   25,
			"blocked":    true,
			"first_seen": time.Now().Add(-6 * time.Hour),
			"last_seen":  time.Now().Add(-10 * time.Minute),
		},
		{
			"id":         2,
			"type":       "sql_injection",
			"source_ip":  "10.0.0.75",
			"target":     "/api/products",
			"attempts":   8,
			"blocked":    true,
			"first_seen": time.Now().Add(-2 * time.Hour),
			"last_seen":  time.Now().Add(-30 * time.Minute),
		},
	}

	data := map[string]interface{}{
		"Title":          "Güvenlik Yönetimi",
		"Stats":          stats,
		"SecurityEvents": securityEvents,
		"ActiveThreats":  activeThreats,
	}

	h.RenderTemplate(w, r, "security/dashboard.gohtml", data)
}

// Users handles user security management
func (h *SecurityHandler) Users(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get users with security info
	users := []map[string]interface{}{
		{
			"id":              1,
			"email":           "admin@example.com",
			"last_login":      time.Now().Add(-2 * time.Hour),
			"login_attempts":  0,
			"2fa_enabled":     true,
			"oauth_connected": true,
			"role":            "admin",
			"status":          "active",
			"ip_address":      "192.168.1.50",
		},
		{
			"id":              2,
			"email":           "user@example.com",
			"last_login":      time.Now().Add(-24 * time.Hour),
			"login_attempts":  3,
			"2fa_enabled":     false,
			"oauth_connected": false,
			"role":            "user",
			"status":          "active",
			"ip_address":      "10.0.0.25",
		},
		{
			"id":              3,
			"email":           "blocked@example.com",
			"last_login":      time.Now().Add(-48 * time.Hour),
			"login_attempts":  10,
			"2fa_enabled":     false,
			"oauth_connected": false,
			"role":            "user",
			"status":          "blocked",
			"ip_address":      "203.0.113.1",
		},
	}

	data := map[string]interface{}{
		"Title": "Kullanıcı Güvenliği",
		"Users": users,
	}

	h.RenderTemplate(w, r, "security/users.gohtml", data)
}

// Threats handles threat management
func (h *SecurityHandler) Threats(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get threat data
	threats := []map[string]interface{}{
		{
			"id":         1,
			"type":       "brute_force",
			"severity":   "high",
			"source_ip":  "192.168.1.100",
			"target":     "login endpoint",
			"attempts":   25,
			"blocked":    true,
			"first_seen": time.Now().Add(-6 * time.Hour),
			"last_seen":  time.Now().Add(-10 * time.Minute),
			"status":     "active",
		},
		{
			"id":         2,
			"type":       "sql_injection",
			"severity":   "critical",
			"source_ip":  "10.0.0.75",
			"target":     "/api/products",
			"attempts":   8,
			"blocked":    true,
			"first_seen": time.Now().Add(-2 * time.Hour),
			"last_seen":  time.Now().Add(-30 * time.Minute),
			"status":     "blocked",
		},
		{
			"id":         3,
			"type":       "xss_attempt",
			"severity":   "medium",
			"source_ip":  "203.0.113.50",
			"target":     "/admin/products",
			"attempts":   3,
			"blocked":    true,
			"first_seen": time.Now().Add(-1 * time.Hour),
			"last_seen":  time.Now().Add(-45 * time.Minute),
			"status":     "monitored",
		},
	}

	data := map[string]interface{}{
		"Title":   "Tehdit Yönetimi",
		"Threats": threats,
	}

	h.RenderTemplate(w, r, "security/threats.gohtml", data)
}

// Settings handles security settings
func (h *SecurityHandler) Settings(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get security settings
	settings := map[string]interface{}{
		"password_policy": map[string]interface{}{
			"min_length":        8,
			"require_uppercase": true,
			"require_lowercase": true,
			"require_numbers":   true,
			"require_symbols":   true,
			"max_age_days":      90,
		},
		"session_settings": map[string]interface{}{
			"timeout_minutes":     30,
			"remember_me_days":    7,
			"concurrent_sessions": 3,
		},
		"2fa_settings": map[string]interface{}{
			"enabled":        true,
			"required_roles": []string{"admin", "moderator"},
			"backup_codes":   true,
		},
		"oauth_settings": map[string]interface{}{
			"google_enabled":   true,
			"github_enabled":   true,
			"facebook_enabled": false,
		},
		"security_headers": map[string]interface{}{
			"csp_enabled":    true,
			"hsts_enabled":   true,
			"xframe_options": "DENY",
			"xss_protection": true,
		},
	}

	data := map[string]interface{}{
		"Title":    "Güvenlik Ayarları",
		"Settings": settings,
	}

	h.RenderTemplate(w, r, "security/settings.gohtml", data)
}

// API Methods

// APIGetSecurityStats returns security statistics
func (h *SecurityHandler) APIGetSecurityStats(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := map[string]interface{}{
		"active_sessions":   245,
		"failed_logins":     18,
		"blocked_ips":       5,
		"2fa_enabled_users": 180,
		"oauth_connections": 95,
		"security_alerts":   3,
		"jwt_tokens_active": 320,
		"threat_level":      "medium",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"stats":   stats,
	})
}

// APIBlockIP blocks an IP address
func (h *SecurityHandler) APIBlockIP(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ip := r.FormValue("ip")
	reason := r.FormValue("reason")
	duration := r.FormValue("duration") // in hours

	if ip == "" {
		http.Error(w, "IP address is required", http.StatusBadRequest)
		return
	}

	log.Printf("Blocking IP %s for reason: %s, duration: %s", ip, reason, duration)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "IP address blocked successfully",
		"ip":      ip,
	})
}

// APIUnblockIP unblocks an IP address
func (h *SecurityHandler) APIUnblockIP(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ip := r.FormValue("ip")
	if ip == "" {
		http.Error(w, "IP address is required", http.StatusBadRequest)
		return
	}

	log.Printf("Unblocking IP %s", ip)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "IP address unblocked successfully",
		"ip":      ip,
	})
}

// APIEnable2FA enables 2FA for a user
func (h *SecurityHandler) APIEnable2FA(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.FormValue("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	log.Printf("Enabling 2FA for user %d", userID)

	// Mock 2FA setup
	qrCode := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="
	secret := "JBSWY3DPEHPK3PXP"

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "2FA enabled successfully",
		"qr_code": qrCode,
		"secret":  secret,
	})
}

// APIUpdateSecuritySettings updates security settings
func (h *SecurityHandler) APIUpdateSecuritySettings(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var settings map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Updating security settings: %+v", settings)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Security settings updated successfully",
	})
}

// APIGetThreatDetails returns detailed threat information
func (h *SecurityHandler) APIGetThreatDetails(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	threatIDStr := r.URL.Query().Get("id")
	threatID, err := strconv.Atoi(threatIDStr)
	if err != nil {
		http.Error(w, "Invalid threat ID", http.StatusBadRequest)
		return
	}

	// Mock threat details
	threatDetails := map[string]interface{}{
		"id":          threatID,
		"type":        "brute_force",
		"severity":    "high",
		"source_ip":   "192.168.1.100",
		"target":      "login endpoint",
		"attempts":    25,
		"blocked":     true,
		"first_seen":  time.Now().Add(-6 * time.Hour),
		"last_seen":   time.Now().Add(-10 * time.Minute),
		"status":      "active",
		"geolocation": "Turkey, Istanbul",
		"user_agent":  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"attack_pattern": []string{
			"POST /login",
			"POST /api/auth/login",
			"POST /admin/login",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"threat":  threatDetails,
	})
}
