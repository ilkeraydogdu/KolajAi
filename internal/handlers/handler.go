package handlers

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"kolajAi/internal/session"
	"kolajAi/internal/errors"
	"kolajAi/internal/models"
)

// contextKey, context değerleri için özel anahtar tipi
type contextKey string

const (
	// UserKey kullanıcı bilgilerini saklamak için kullanılan anahtar
	UserKey = contextKey("user")
	// SessionCookieName oturum çerezi için kullanılan isim
	SessionCookieName = "kolajAI_session"
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

// Handler tüm handler'lar için temel yapı
type Handler struct {
	Templates       *template.Template
	SessionManager  *session.SessionManager // Gelişmiş session manager
	TemplateContext map[string]interface{}
	ErrorManager    *errors.ErrorManager
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
	
	// Session var mı ve aktif mi kontrol et
	if sessionData == nil || !sessionData.IsActive {
		Logger.Printf("IsAuthenticated - Session yok veya aktif değil")
		return false
	}
	
	// Session süresi dolmuş mu kontrol et
	if time.Now().After(sessionData.ExpiresAt) {
		Logger.Printf("IsAuthenticated - Session süresi dolmuş")
		return false
	}
	
	// UserID var mı kontrol et
	if sessionData.UserID <= 0 {
		Logger.Printf("IsAuthenticated - Geçersiz UserID: %d", sessionData.UserID)
		return false
	}
	
	Logger.Printf("IsAuthenticated - Kullanıcı doğrulandı: UserID=%d", sessionData.UserID)
	return true
}

// GetCurrentUser mevcut kullanıcı bilgilerini döner
func (h *Handler) GetCurrentUser(r *http.Request) (*models.User, error) {
	sessionData, err := h.SessionManager.GetSession(r)
	if err != nil {
		return nil, err
	}
	
	if sessionData == nil || sessionData.UserID <= 0 {
		return nil, errors.NewApplicationError(errors.AUTHENTICATION, "NO_USER", "Kullanıcı bulunamadı", nil)
	}
	
	// TODO: Veritabanından kullanıcı bilgilerini çek
	// Şimdilik dummy data dönüyoruz
	user := &models.User{
		ID:    sessionData.UserID,
		Email: "user@example.com", // Bu bilgiler veritabanından gelecek
		Name:  "User",
	}
	
	return user, nil
}

// HasPermission kullanıcının belirtilen yetkiye sahip olup olmadığını kontrol eder
func (h *Handler) HasPermission(r *http.Request, permission string) bool {
	sessionData, err := h.SessionManager.GetSession(r)
	if err != nil || sessionData == nil {
		return false
	}
	
	for _, perm := range sessionData.Permissions {
		if perm == permission {
			return true
		}
	}
	
	return false
}

// RenderTemplate şablon render işlemini gerçekleştirir
func (h *Handler) RenderTemplate(w http.ResponseWriter, r *http.Request, name string, data map[string]interface{}) {
	Logger.Printf("RenderTemplate çağrıldı - Şablon: %s", name)

	// Temel template context'i kopyala
	templateContext := make(map[string]interface{})
	for k, v := range h.TemplateContext {
		templateContext[k] = v
	}

	// Data'yı template context'e ekle
	for k, v := range data {
		templateContext[k] = v
	}

	// Kullanıcı bilgilerini ekle
	if h.IsAuthenticated(r) {
		user, err := h.GetCurrentUser(r)
		if err == nil && user != nil {
			templateContext["user"] = user
			templateContext["isAuthenticated"] = true
			
			// Admin kontrolü
			if h.HasPermission(r, "admin") {
				templateContext["isAdmin"] = true
			}
		} else {
			templateContext["isAuthenticated"] = false
		}
	} else {
		templateContext["isAuthenticated"] = false
	}

	// Flash mesajlarını al - TODO: Implement flash messages with new session manager
	templateContext["flash"] = []string{}

	// Şablonu render et
	err := h.Templates.ExecuteTemplate(w, name, templateContext)
	if err != nil {
		Logger.Printf("Template render hatası: %v", err)
		http.Error(w, "Sayfa yüklenirken bir hata oluştu", http.StatusInternalServerError)
	}
}

// RedirectWithFlash flash mesajı ile yönlendirme yapar
func (h *Handler) RedirectWithFlash(w http.ResponseWriter, r *http.Request, url string, message string) {
	// TODO: Implement flash messages with new session manager
	// Şimdilik sadece yönlendir
	Logger.Printf("RedirectWithFlash: URL=%s, Message=%s", url, message)
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
