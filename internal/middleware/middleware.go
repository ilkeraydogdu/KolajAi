package middleware

import (
	"log"
	"net/http"
	"os"
)

var (
	Logger *log.Logger
)

func init() {
	// Middleware için log dosyası oluştur
	logFile, err := os.OpenFile("middleware_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Middleware log dosyası oluşturulamadı:", err)
		Logger = log.New(os.Stdout, "[MIDDLEWARE-DEBUG] ", log.LstdFlags)
	} else {
		Logger = log.New(logFile, "[MIDDLEWARE-DEBUG] ", log.LstdFlags|log.Lshortfile)
	}
}

// Middleware represents a middleware
type Middleware func(http.Handler) http.Handler

// Chain chains multiple middlewares
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Auth middleware for authentication
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Logger.Printf("Auth middleware çalışıyor: URL=%s, Method=%s", r.URL.Path, r.Method)

		// Tüm çerezleri logla
		Logger.Printf("Request'teki tüm çerezler:")
		for _, cookie := range r.Cookies() {
			Logger.Printf("- Çerez: %s=%s, Path=%s, MaxAge=%d", cookie.Name, cookie.Value, cookie.Path, cookie.MaxAge)
		}

		// Session kontrolü (örnek implementasyon, gerçek uygulamada oturum mantığına göre değiştirilmeli)
		authToken, err := r.Cookie("kolaj-session")

		if err != nil {
			Logger.Printf("Auth hatası: Session çerezi bulunamadı - %v", err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		Logger.Printf("Auth token bulundu: %s", authToken.Value)

		// Public URL'ler için erişimi kontrol etmiyoruz
		if isPublicURL(r.URL.Path) {
			Logger.Printf("Bu bir public URL, auth kontrolü atlanıyor: %s", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		// Token geçerli mi kontrol et
		if authToken.Value == "" {
			Logger.Printf("Auth token değeri boş, login sayfasına yönlendiriliyor")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Token geçerli, isteği işlemeye devam et
		Logger.Printf("Auth başarılı, isteği handler'a iletiyorum")
		next.ServeHTTP(w, r)
	})
}

// isPublicURL checks if a URL path is public
func isPublicURL(path string) bool {
	publicPaths := []string{
		"/login",
		"/register",
		"/forgot-password",
		"/reset-password",
		"/static/",
		"/favicon.ico",
	}

	for _, p := range publicPaths {
		if p == path || (len(p) > 0 && p[len(p)-1] == '/' && len(path) >= len(p) && path[:len(p)] == p) {
			return true
		}
	}

	return false
}

// Secure middleware for security headers
func Secure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Logger.Printf("Secure middleware çalışıyor: %s", r.URL.Path)

		// Set security headers
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:;")

		next.ServeHTTP(w, r)
	})
}

// Logger middleware for logging requests
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Logger.Printf("Request: %s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
