package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
	
	"github.com/kolajai/internal/models"
)

var (
	DashboardLogger *log.Logger
)

func init() {
	// Environment'a göre log seviyesini ayarla
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("GIN_MODE")
	}
	
	if env == "production" || env == "release" {
		// Production'da sadece stdout'a minimal log
		DashboardLogger = log.New(os.Stdout, "[DASHBOARD] ", log.LstdFlags)
	} else {
		// Development'ta debug log dosyası
		logFile, err := os.OpenFile("dashboard_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Println("Dashboard log dosyası oluşturulamadı:", err)
			DashboardLogger = log.New(os.Stdout, "[DASHBOARD-DEBUG] ", log.LstdFlags)
		} else {
			DashboardLogger = log.New(logFile, "[DASHBOARD-DEBUG] ", log.LstdFlags|log.Lshortfile)
		}
	}
}

// Dashboard handles the dashboard page
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	DashboardLogger.Printf("Dashboard handler çağrıldı: Method=%s, URL=%s", r.Method, r.URL.Path)

	// Kimlik doğrulaması kontrolü
	if !h.IsAuthenticated(r) {
		DashboardLogger.Printf("Dashboard - Kullanıcı kimliği doğrulanmamış, login sayfasına yönlendiriliyor")
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get user info from session
	userInfo := h.GetUserFromSession(r)
	if userInfo == nil {
		DashboardLogger.Printf("Dashboard - Kullanıcı bilgisi alınamadı")
		h.RedirectWithFlash(w, r, "/login", "Oturum süresi dolmuş, lütfen tekrar giriş yapın")
		return
	}

	// Get dashboard stats
	stats := h.GetDashboardStats(userInfo.ID)
	
	// Get recent notifications
	notifications := h.GetRecentNotifications(userInfo.ID, 5)
	
	// Get pending tasks
	tasks := h.GetPendingTasks(userInfo.ID, 5)
	
	// Get quick actions based on user role
	quickActions := h.GetQuickActions(userInfo)
	
	// Prepare chart data
	activityChart := h.GetActivityChartData(7) // Last 7 days
	
	// Check if it's admin user
	isAdmin := h.IsAdminUser(r)

	data := map[string]interface{}{
		"Title":          "Dashboard - KolajAI",
		"PageTitle":      "Dashboard",
		"UserID":         userInfo.ID,
		"UserName":       userInfo.Name,
		"UserEmail":      userInfo.Email,
		"CurrentTime":    time.Now().Format("02.01.2006 15:04"),
		"Stats":          stats,
		"Notifications":  notifications,
		"Tasks":          tasks,
		"QuickActions":   quickActions,
		"ActivityChart":  activityChart,
		"IsAdmin":        isAdmin,
		"ShowWelcome":    h.IsFirstLogin(userInfo.ID),
	}

	// Şablonu render et
	DashboardLogger.Printf("Dashboard - Kimlik doğrulanmış, dashboard sayfası gösteriliyor")
	h.RenderTemplate(w, r, "dashboard/index", data)
}

// GetDashboardStats retrieves dashboard statistics
func (h *Handler) GetDashboardStats(userID int64) *models.DashboardStats {
	// TODO: Get real stats from database
	// Demo data for now
	return &models.DashboardStats{
		TotalUsers:          1250,
		ActiveUsers:         987,
		TotalProducts:       3456,
		TotalOrders:         789,
		TotalRevenue:        125678.90,
		NewUsersToday:       23,
		OrdersToday:         45,
		RevenueToday:        3456.78,
		PendingTasks:        5,
		UnreadNotifications: 3,
	}
}

// GetRecentNotifications retrieves recent notifications for user
func (h *Handler) GetRecentNotifications(userID int64, limit int) []models.Notification {
	// TODO: Get real notifications from database
	// Demo data for now
	return []models.Notification{
		{
			ID:        1,
			UserID:    userID,
			Title:     "Hoş Geldiniz",
			Message:   "KolajAI platformuna hoş geldiniz! Başlamak için profil bilgilerinizi tamamlayın.",
			Type:      "info",
			IsRead:    false,
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:        2,
			UserID:    userID,
			Title:     "Güvenlik Bildirimi",
			Message:   "Hesabınıza yeni bir cihazdan giriş yapıldı.",
			Type:      "warning",
			IsRead:    false,
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
	}
}

// GetPendingTasks retrieves pending tasks for user
func (h *Handler) GetPendingTasks(userID int64, limit int) []models.Task {
	// TODO: Get real tasks from database
	// Demo data for now
	dueDate := time.Now().Add(24 * time.Hour)
	return []models.Task{
		{
			ID:          1,
			UserID:      userID,
			Title:       "Profil bilgilerini güncelle",
			Description: "Profil sayfanızdan kişisel bilgilerinizi güncelleyin",
			Status:      "pending",
			Priority:    "medium",
			DueDate:     &dueDate,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
		},
		{
			ID:          2,
			UserID:      userID,
			Title:       "Güvenlik ayarlarını kontrol et",
			Description: "İki faktörlü kimlik doğrulamayı etkinleştirin",
			Status:      "pending",
			Priority:    "high",
			CreatedAt:   time.Now().Add(-12 * time.Hour),
		},
	}
}

// GetQuickActions returns quick actions based on user role
func (h *Handler) GetQuickActions(user *UserInfo) []models.QuickAction {
	actions := []models.QuickAction{
		{
			Title:       "Profil",
			Description: "Profil bilgilerinizi yönetin",
			Icon:        "person",
			URL:         "/profile",
			Color:       "primary",
		},
		{
			Title:       "Bildirimler",
			Description: "Tüm bildirimlerinizi görün",
			Icon:        "notifications",
			URL:         "/notifications",
			Color:       "info",
			Badge:       "3",
		},
		{
			Title:       "Ayarlar",
			Description: "Hesap ayarlarınızı yönetin",
			Icon:        "settings",
			URL:         "/settings",
			Color:       "secondary",
		},
	}
	
	// Add admin-specific actions
	if user.IsAdmin {
		actions = append(actions, models.QuickAction{
			Title:       "Admin Panel",
			Description: "Sistem yönetimi",
			Icon:        "admin_panel_settings",
			URL:         "/admin/dashboard",
			Color:       "danger",
		})
	}
	
	return actions
}

// GetActivityChartData returns user activity chart data
func (h *Handler) GetActivityChartData(days int) *models.DashboardChart {
	labels := []string{}
	data := []interface{}{}
	
	// Generate last N days
	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		labels = append(labels, date.Format("02 Jan"))
		// TODO: Get real data from database
		data = append(data, 10 + i*5) // Demo data
	}
	
	return &models.DashboardChart{
		Labels: labels,
		Data:   data,
		Type:   "line",
	}
}

// IsFirstLogin checks if this is user's first login
func (h *Handler) IsFirstLogin(userID int64) bool {
	// TODO: Check from database
	return false
}

// DashboardAPI handles AJAX requests for dashboard data
func (h *Handler) DashboardAPI(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get action from query
	action := r.URL.Query().Get("action")
	
	switch action {
	case "refresh-stats":
		h.RefreshDashboardStats(w, r)
	case "mark-notification-read":
		h.MarkNotificationRead(w, r)
	case "complete-task":
		h.CompleteTask(w, r)
	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
	}
}

// RefreshDashboardStats returns updated dashboard stats via AJAX
func (h *Handler) RefreshDashboardStats(w http.ResponseWriter, r *http.Request) {
	userInfo := h.GetUserFromSession(r)
	if userInfo == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	stats := h.GetDashboardStats(userInfo.ID)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// MarkNotificationRead marks a notification as read
func (h *Handler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		NotificationID int64 `json:"notification_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	// TODO: Update notification in database
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Bildirim okundu olarak işaretlendi",
	})
}

// CompleteTask marks a task as completed
func (h *Handler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		TaskID int64 `json:"task_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	// TODO: Update task in database
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Görev tamamlandı",
	})
}
