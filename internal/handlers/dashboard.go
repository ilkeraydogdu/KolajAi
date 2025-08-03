package handlers

import (
	"log"
	"net/http"
	"os"
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

	// Tüm çerezleri logla
	DashboardLogger.Printf("Dashboard - Request'teki tüm çerezler:")
	for _, cookie := range r.Cookies() {
		DashboardLogger.Printf("- Çerez: %s=%s, Path=%s, MaxAge=%d", cookie.Name, cookie.Value, cookie.Path, cookie.MaxAge)
	}

	// Kimlik doğrulaması kontrolü
	if !h.IsAuthenticated(r) {
		DashboardLogger.Printf("Dashboard - Kullanıcı kimliği doğrulanmamış, login sayfasına yönlendiriliyor")
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Şablonu render et
	DashboardLogger.Printf("Dashboard - Kimlik doğrulanmış, dashboard sayfası gösteriliyor")
	h.RenderTemplate(w, r, "dashboard/index", map[string]interface{}{
		"Title": "Dashboard - KolajAI",
	})
}
