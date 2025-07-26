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

	"github.com/gorilla/sessions"
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

// SessionManager oturum yönetimi için kullanılır
type SessionManager struct {
	store  *sessions.CookieStore
	mutex  sync.Mutex
	Logger *log.Logger
}

// NewSessionManager yeni bir session manager oluşturur
func NewSessionManager(secret string) *SessionManager {
	return &SessionManager{
		store:  sessions.NewCookieStore([]byte(secret)),
		Logger: Logger,
	}
}

// GetSession mevcut HTTP isteği için oturum bilgisini getirir
func (sm *SessionManager) GetSession(r *http.Request) (*sessions.Session, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, err := sm.store.Get(r, SessionCookieName)

	sm.Logger.Printf("GetSession çağrıldı - Cookie Adı: %s, Hata: %v", SessionCookieName, err)
	if err != nil {
		sm.Logger.Printf("Oturum çerezini okuma hatası: %v", err)
		return nil, err
	}

	// Session bilgilerini detaylı logla
	sm.Logger.Printf("Oturum Bilgileri: IsNew=%v, Values=%+v", session.IsNew, session.Values)

	// UserKey kontrolü
	if user, ok := session.Values[UserKey]; ok {
		sm.Logger.Printf("Kullanıcı oturumda bulundu: %+v", user)
	} else {
		sm.Logger.Printf("Kullanıcı oturumda bulunamadı")
	}

	return session, nil
}

// SetSession HTTP yanıtı için oturum bilgilerini günceller
func (sm *SessionManager) SetSession(w http.ResponseWriter, r *http.Request, key, val interface{}) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, err := sm.store.Get(r, SessionCookieName)
	if err != nil {
		sm.Logger.Printf("SetSession - Oturum çerezini okuma hatası: %v", err)
		return err
	}

	session.Values[key] = val
	sm.Logger.Printf("Oturum güncellendi - Key: %v, Value: %+v", key, val)

	return session.Save(r, w)
}

// ClearSession HTTP yanıtı için oturum bilgilerini temizler
func (sm *SessionManager) ClearSession(w http.ResponseWriter, r *http.Request) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, err := sm.store.Get(r, SessionCookieName)
	if err != nil {
		sm.Logger.Printf("ClearSession - Oturum çerezini okuma hatası: %v", err)
		return err
	}

	// Tüm session değerlerini temizle
	for k := range session.Values {
		sm.Logger.Printf("Oturum değeri siliniyor: %v", k)
		delete(session.Values, k)
	}

	// Çerezi geçersiz kılmak için
	session.Options.MaxAge = -1

	sm.Logger.Printf("Oturum tamamen temizlendi")
	return session.Save(r, w)
}

// cleanupAllCookies istemcideki tüm çerezleri temizler
func (sm *SessionManager) CleanupAllCookies(w http.ResponseWriter, r *http.Request) {
	sm.Logger.Printf("CleanupAllCookies çağrıldı - Tüm çerezler temizleniyor")

	// Session çerezini temizle
	session, err := sm.store.Get(r, SessionCookieName)
	if err == nil {
		session.Options.MaxAge = -1
		session.Save(r, w)
		sm.Logger.Printf("Session çerezi temizlendi: %s", SessionCookieName)
	} else {
		sm.Logger.Printf("Session çerezi temizlenirken hata: %v", err)
	}

	// Request'teki tüm çerezleri al ve temizle
	for _, cookie := range r.Cookies() {
		expiredCookie := &http.Cookie{
			Name:    cookie.Name,
			Value:   "",
			Path:    "/",
			Expires: time.Unix(0, 0),
			MaxAge:  -1,
		}
		http.SetCookie(w, expiredCookie)
		sm.Logger.Printf("Çerez temizlendi: %s", cookie.Name)
	}
}

// AddFlash oturum için flash mesajı ekler
func (sm *SessionManager) AddFlash(w http.ResponseWriter, r *http.Request, message string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, err := sm.store.Get(r, SessionCookieName)
	if err != nil {
		return err
	}

	session.AddFlash(message, FlashKey)
	sm.Logger.Printf("Flash mesajı eklendi: %s", message)

	return session.Save(r, w)
}

// GetFlashes oturumdaki flash mesajlarını getirir
func (sm *SessionManager) GetFlashes(w http.ResponseWriter, r *http.Request) ([]interface{}, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, err := sm.store.Get(r, SessionCookieName)
	if err != nil {
		return nil, err
	}

	flashes := session.Flashes(FlashKey)
	sm.Logger.Printf("Flash mesajları alındı: %+v", flashes)

	err = session.Save(r, w)
	if err != nil {
		return nil, err
	}

	return flashes, nil
}

// Handler temel handler yapısı
type Handler struct {
	Templates       *template.Template
	SessionManager  *SessionManager
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
	session, err := h.SessionManager.GetSession(r)
	if err != nil {
		h.SessionManager.Logger.Printf("IsAuthenticated - Oturum alınamadı: %v", err)
		return false
	}

	_, ok := session.Values[UserKey]
	h.SessionManager.Logger.Printf("IsAuthenticated sonucu: %v", ok)
	return ok
}

// RenderTemplate şablon render işlemini gerçekleştirir
func (h *Handler) RenderTemplate(w http.ResponseWriter, r *http.Request, name string, data map[string]interface{}) {
	h.SessionManager.Logger.Printf("RenderTemplate çağrıldı - Şablon: %s", name)

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

	// Oturumdaki flash mesajlarını al
	flashes, err := h.SessionManager.GetFlashes(w, r)
	if err == nil {
		templateContext["flashes"] = flashes
	}

	// Kimlik doğrulaması durumunu kontrol et
	if h.IsAuthenticated(r) {
		session, _ := h.SessionManager.GetSession(r)
		if user, ok := session.Values[UserKey]; ok {
			templateContext["currentUser"] = user
			templateContext["isAuthenticated"] = true
		}
	} else {
		templateContext["isAuthenticated"] = false
	}

	h.SessionManager.Logger.Printf("Şablon verileri: %+v", templateContext)

	// Şablonu render et
	err = h.Templates.ExecuteTemplate(w, name, templateContext)
	if err != nil {
		h.SessionManager.Logger.Printf("Şablon render hatası: %v", err)
		http.Error(w, fmt.Sprintf("Template rendering error: %v", err), http.StatusInternalServerError)
		return
	}
}

// RedirectWithFlash kullanıcıyı flash mesajı ile birlikte yönlendirir
func (h *Handler) RedirectWithFlash(w http.ResponseWriter, r *http.Request, url, message string) {
	h.SessionManager.Logger.Printf("RedirectWithFlash - URL: %s, Mesaj: %s", url, message)

	if message != "" {
		err := h.SessionManager.AddFlash(w, r, message)
		if err != nil {
			h.SessionManager.Logger.Printf("Flash mesajı eklenirken hata: %v", err)
		}
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
