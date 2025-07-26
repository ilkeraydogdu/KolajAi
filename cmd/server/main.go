package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"kolajAi/internal/handlers"
)

var (
	MainLogger *log.Logger
)

func init() {
	// Ana uygulama için log dosyası oluştur
	logFile, err := os.OpenFile("main_app_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Ana uygulama log dosyası oluşturulamadı:", err)
		MainLogger = log.New(os.Stdout, "[MAIN-APP-DEBUG] ", log.LstdFlags)
	} else {
		MainLogger = log.New(logFile, "[MAIN-APP-DEBUG] ", log.LstdFlags|log.Lshortfile)
	}
}

func main() {
	MainLogger.Println("KolajAI uygulaması başlatılıyor...")

	// Şablonları yükle
	MainLogger.Println("Şablonlar yükleniyor...")
	
	// Template fonksiyonlarını tanımla
	funcMap := template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict fonksiyonu için çift sayıda parametre gerekli")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict fonksiyonu için anahtarlar string olmalı")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"rand": func() int {
			return time.Now().Nanosecond() % 1000
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
	
	tmpl, err := template.New("").Funcs(funcMap).ParseGlob("web/templates/**/*.gohtml")
	if err != nil {
		MainLogger.Fatalf("Şablonlar yüklenemedi: %v", err)
	}
	MainLogger.Printf("Şablonlar başarıyla yüklendi!")

	// Handler'ları oluştur
	MainLogger.Println("Handler'lar oluşturuluyor...")

	// Session manager oluştur - güvenli bir anahtar kullan
	sessionManager := handlers.NewSessionManager("supersecretkey123")

	h := &handlers.Handler{
		Templates:      tmpl,
		SessionManager: sessionManager,
		TemplateContext: map[string]interface{}{
			"AppName": "KolajAI",
			"Year":    time.Now().Year(),
		},
	}

	// Router oluştur ve handler'ları ekle
	router := http.NewServeMux()

	// Statik dosyalar
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Ana sayfa
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	// Auth işlemleri
	router.HandleFunc("/login", h.Login)
	router.HandleFunc("/register", h.Register)
	router.HandleFunc("/forgot-password", h.ForgotPassword)
	router.HandleFunc("/reset-password", h.ResetPassword)

	// Auth gerektiren sayfalar
	router.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if !h.IsAuthenticated(r) {
			h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
			return
		}
		h.Dashboard(w, r)
	})

	router.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if !h.IsAuthenticated(r) {
			h.RedirectWithFlash(w, r, "/login", "Zaten çıkış yapılmış")
			return
		}
		h.Logout(w, r)
	})

	// Favicon
	router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/static/assets/images/favicon-32x32.png")
	})

	// Sunucuyu başlat
	addr := ":8081"
	MainLogger.Printf("Sunucu başlatılıyor: %s", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	MainLogger.Printf("KolajAI uygulaması başlatıldı. %s adresinde dinleniyor...", addr)
	MainLogger.Fatal(server.ListenAndServe())
}
