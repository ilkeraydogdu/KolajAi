package router

import (
	"log"
	"net/http"
	"os"

	"kolajAi/internal/handlers"
	"kolajAi/internal/middleware"
)

var (
	RouterLogger *log.Logger
)

func init() {
	// Router için log dosyası oluştur
	logFile, err := os.OpenFile("router_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Router log dosyası oluşturulamadı:", err)
		RouterLogger = log.New(os.Stdout, "[ROUTER-DEBUG] ", log.LstdFlags)
	} else {
		RouterLogger = log.New(logFile, "[ROUTER-DEBUG] ", log.LstdFlags|log.Lshortfile)
	}
}

// NewRouter creates a new router instance
func NewRouter(h *handlers.Handler, static interface{}) http.Handler {
	// Middleware zinciri oluştur
	middlewareChain := middleware.Chain(
		middleware.RequestLogger, // İstek logları
		middleware.Secure,        // Güvenlik başlıkları
	)

	// Ana handler, tüm istekleri önce logla
	mux := http.NewServeMux()

	// Statik dosyalar için handler
	staticDir := "./web/static"
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	
	// Giriş sayfası
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		RouterLogger.Printf("Login route: %s %s", r.Method, r.URL.Path)
		h.Login(w, r)
	})

	// Kayıt sayfası
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		RouterLogger.Printf("Register route: %s %s", r.Method, r.URL.Path)
		h.Register(w, r)
	})

	// Şifremi unuttum sayfası
	mux.HandleFunc("/forgot-password", func(w http.ResponseWriter, r *http.Request) {
		RouterLogger.Printf("ForgotPassword route: %s %s", r.Method, r.URL.Path)
		h.ForgotPassword(w, r)
	})

	// Şifre sıfırlama sayfası
	mux.HandleFunc("/reset-password", func(w http.ResponseWriter, r *http.Request) {
		RouterLogger.Printf("ResetPassword route: %s %s", r.Method, r.URL.Path)
		h.ResetPassword(w, r)
	})

	// Auth gerektiren rotalar için middleware ekle
	authMux := http.NewServeMux()

	// Dashboard sayfası
	authMux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		RouterLogger.Printf("Dashboard route: %s %s", r.Method, r.URL.Path)
		h.Dashboard(w, r)
	})

	// Çıkış yapma
	authMux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		RouterLogger.Printf("Logout route: %s %s", r.Method, r.URL.Path)
		h.Logout(w, r)
	})

	// Auth middleware ile korunan rotaları ana mux'e ekle
	mux.Handle("/dashboard", middleware.Auth(authMux))
	mux.Handle("/logout", middleware.Auth(authMux))

	// Ana sayfa - yönlendirme yap
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		RouterLogger.Printf("Root route: %s %s -> redirecting to /login", r.Method, r.URL.Path)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	// Favicon
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/static/assets/images/favicon-32x32.png")
	})

	// Log middleware ile sarılmış mux'i döndür
	return middlewareChain(mux)
}
