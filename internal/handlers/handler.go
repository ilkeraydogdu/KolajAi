package handlers

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"kolajAi/internal/session"
)

// contextKey, context değerleri için özel anahtar tipi
type contextKey string

const (
	SessionCookieName = "kolaj-session"
	UserKey           = contextKey("user")
	FlashKey          = "flash"
)

var (
	Logger *log.Logger
)

func init() {
	// Detaylı log için logger oluştur
	logFile, err := os.OpenFile("auth_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Log dosyası oluşturulamadı:", err)
		Logger = log.New(os.Stdout, "[AUTH-DEBUG] ", log.LstdFlags)
	} else {
		Logger = log.New(logFile, "[AUTH-DEBUG] ", log.LstdFlags|log.Lshortfile)
	}
}



// Handler temel handler yapısı
type Handler struct {
	Templates       *template.Template
	SessionManager  *session.SessionManager
	TemplateContext map[string]interface{}
}

// WithUser kullanıcı bilgisini context'e ekler
func WithUser(ctx context.Context, user interface{}) context.Context {
	Logger.Printf("WithUser çağrıldı - User: %+v", user)
	return context.WithValue(ctx, UserKey, user)
}

// UserFromContext context'ten kullanıcı bilgisini alır
func UserFromContext(ctx context.Context) (interface{}, bool) {
	user := ctx.Value(UserKey)
	Logger.Printf("UserFromContext çağrıldı - User: %+v, Exists: %v", user, user != nil)
	return user, user != nil
}

// IsAuthenticated kullanıcının oturum açmış olup olmadığını kontrol eder
func (h *Handler) IsAuthenticated(r *http.Request) bool {
	sessionData, err := h.SessionManager.GetSession(r)
	if err != nil {
		Logger.Printf("IsAuthenticated - Oturum alınamadı: %v", err)
		return false
	}

	// Session data varsa ve user ID pozitifse authenticated
	if sessionData != nil {
		authenticated := sessionData.UserID > 0 && sessionData.IsActive
		Logger.Printf("IsAuthenticated sonucu: %v (UserID: %d)", authenticated, sessionData.UserID)
		return authenticated
	}
	
	Logger.Printf("IsAuthenticated sonucu: false (session data nil)")
	return false
}

// RenderTemplate şablon render işlemini gerçekleştirir
func (h *Handler) RenderTemplate(w http.ResponseWriter, r *http.Request, name string, data map[string]interface{}) {
	Logger.Printf("RenderTemplate çağrıldı - Şablon: %s", name)

	// Şablon bağlamını oluştur
	templateContext := make(map[string]interface{})

	// Global bağlamdan değerleri al
	for k, v := range h.TemplateContext {
		templateContext[k] = v
	}

	// Gelen veriyi bağlama ekle
	for k, v := range data {
		templateContext[k] = v
	}

	// Flash mesajları için şimdilik boş bırak (advanced session manager'da flash desteği eklenebilir)
	templateContext["flashes"] = []interface{}{}

	// Kimlik doğrulaması durumunu kontrol et
	if h.IsAuthenticated(r) {
		sessionData, _ := h.SessionManager.GetSession(r)
		if sessionData != nil {
			// Kullanıcı bilgilerini oluştur
			user := map[string]interface{}{
				"ID":    sessionData.UserID,
				"Email": "admin@example.com", // Gerçek uygulamada veritabanından al
				"Name":  "Admin User",
			}
			templateContext["currentUser"] = user
			templateContext["isAuthenticated"] = true
		}
	} else {
		templateContext["isAuthenticated"] = false
	}

	Logger.Printf("Şablon verileri: %+v", templateContext)

	// Şablonu render et
	err := h.Templates.ExecuteTemplate(w, name, templateContext)
	if err != nil {
		Logger.Printf("Şablon render hatası: %v", err)
		http.Error(w, fmt.Sprintf("Template rendering error: %v", err), http.StatusInternalServerError)
		return
	}
}

// RedirectWithFlash kullanıcıyı flash mesajı ile birlikte yönlendirir
func (h *Handler) RedirectWithFlash(w http.ResponseWriter, r *http.Request, url, message string) {
	Logger.Printf("RedirectWithFlash - URL: %s, Mesaj: %s", url, message)

	// TODO: Flash mesajları için advanced session manager'a destek eklenebilir
	if message != "" {
		Logger.Printf("Flash mesajı: %s", message)
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

// GetTemplateData returns base template data
func (h *Handler) GetTemplateData() map[string]interface{} {
	data := make(map[string]interface{})

	// Copy template context
	for k, v := range h.TemplateContext {
		data[k] = v
	}

	return data
}

// HandleError handles errors and renders error page
func (h *Handler) HandleError(w http.ResponseWriter, r *http.Request, err error, message string) {
	Logger.Printf("Error: %v", err)

	data := h.GetTemplateData()
	data["Error"] = message
	data["ErrorDetails"] = err.Error()

	w.WriteHeader(http.StatusInternalServerError)
	h.RenderTemplate(w, r, "error", data)
}
