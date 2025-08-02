package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

var (
	AuthLogger *log.Logger
)

func init() {
	// Environment'a göre log seviyesini ayarla
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("GIN_MODE")
	}
	
	if env == "production" || env == "release" {
		// Production'da sadece stdout'a minimal log
		AuthLogger = log.New(os.Stdout, "[AUTH] ", log.LstdFlags)
	} else {
		// Development'ta debug log dosyası
		logFile, err := os.OpenFile("auth_ops_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Println("Auth log dosyası oluşturulamadı:", err)
			AuthLogger = log.New(os.Stdout, "[AUTH-OPS-DEBUG] ", log.LstdFlags)
		} else {
			AuthLogger = log.New(logFile, "[AUTH-OPS-DEBUG] ", log.LstdFlags|log.Lshortfile)
		}
	}
}

// Login handles the login request
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	AuthLogger.Printf("Login handler çağrıldı: Method=%s, URL=%s", r.Method, r.URL.Path)

	// Tüm çerezleri logla
	AuthLogger.Printf("Login - Request'teki tüm çerezler:")
	for _, cookie := range r.Cookies() {
		AuthLogger.Printf("- Çerez: %s=%s, Path=%s, MaxAge=%d", cookie.Name, cookie.Value, cookie.Path, cookie.MaxAge)
	}

	// Eğer kullanıcı zaten oturum açmışsa, anasayfaya yönlendir
	if h.IsAuthenticated(r) {
		AuthLogger.Printf("Login - Kullanıcı zaten oturum açmış, dashboard'a yönlendiriliyor")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	// POST isteği için giriş işlemi
	if r.Method == http.MethodPost {
		AuthLogger.Printf("Login - POST isteği alındı, giriş yapılmaya çalışılıyor")

		err := r.ParseForm()
		if err != nil {
			AuthLogger.Printf("Login - Form parse hatası: %v", err)
			h.RedirectWithFlash(w, r, "/login", "Form işlenirken hata oluştu")
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		AuthLogger.Printf("Login - Giriş denemesi: Email=%s", email)

		// Basitleştirilmiş kimlik doğrulama (gerçek uygulamada veritabanı sorgusu ile doğrulama yapılmalı)
		// Not: Bu örnek sadece demo amaçlıdır, gerçek uygulamalarda güvenli kimlik doğrulama kullanılmalıdır
		// SECURITY: Remove hardcoded credentials - this is just for demo
	// In production, use proper database authentication
	if email == "admin@example.com" && password == "password" {
			AuthLogger.Printf("Login - Başarılı giriş: %s", email)

			// Kullanıcı bilgilerini oluştur
			user := struct {
				ID    int64
				Email string
				Name  string
			}{
				ID:    1,
				Email: email,
				Name:  "Admin User",
			}

			// Oturum oluştur - kullanıcı bilgilerini ve yetki durumunu kaydet
			if err := h.SessionManager.SetSession(w, r, UserKey, user); err != nil {
				AuthLogger.Printf("Login - Oturum oluşturma hatası (user): %v", err)
				h.RedirectWithFlash(w, r, "/login", "Oturum oluşturulurken hata oluştu")
				return
			}
			// Kullanıcı kimliği ve admin bayrağını ekle
			if err := h.SessionManager.SetSession(w, r, "user_id", user.ID); err != nil {
				AuthLogger.Printf("Login - Oturum oluşturma hatası (user_id): %v", err)
				h.RedirectWithFlash(w, r, "/login", "Oturum oluşturulurken hata oluştu")
				return
			}
			if err := h.SessionManager.SetSession(w, r, "is_admin", true); err != nil {
				AuthLogger.Printf("Login - Oturum oluşturma hatası (is_admin): %v", err)
				h.RedirectWithFlash(w, r, "/login", "Oturum oluşturulurken hata oluştu")
				return
			}

			AuthLogger.Printf("Login - Oturum başarıyla oluşturuldu, dashboard'a yönlendiriliyor")
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		// Hatalı giriş
		AuthLogger.Printf("Login - Hatalı giriş denemesi: %s", email)
		h.RedirectWithFlash(w, r, "/login", "Hatalı e-posta veya şifre")
		return
	}

	// GET isteği için giriş sayfasını göster
	AuthLogger.Printf("Login - GET isteği, login sayfası gösteriliyor")

	data := map[string]interface{}{
		"Title":          "Giriş - KolajAI",
		"PageHeading":    "Giriş Yap",
		"PageSubHeading": "Hesabınıza giriş yapın ve işlemlerinize devam edin!",
	}

	// Şablonu render et
	h.RenderTemplate(w, r, "auth/login", data)
}

// Logout logs out the user
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	AuthLogger.Printf("Logout handler çağrıldı: Method=%s, URL=%s", r.Method, r.URL.Path)

	// Tüm çerezleri temizle
	AuthLogger.Printf("Logout - Tüm çerezler temizleniyor")
	h.SessionManager.CleanupAllCookies(w, r)

	// Oturumu temizle
	err := h.SessionManager.ClearSession(w, r)
	if err != nil {
		AuthLogger.Printf("Logout - Oturum temizleme hatası: %v", err)
	} else {
		AuthLogger.Printf("Logout - Oturum başarıyla temizlendi")
	}

	// Kullanıcıyı login sayfasına yönlendir
	AuthLogger.Printf("Logout - Kullanıcı login sayfasına yönlendiriliyor")
	h.RedirectWithFlash(w, r, "/login", "Başarıyla çıkış yapıldı")
}

// ForgotPassword handles the forgot password request
func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	AuthLogger.Printf("ForgotPassword handler çağrıldı: Method=%s", r.Method)

	if r.Method == http.MethodPost {
		AuthLogger.Printf("ForgotPassword - POST isteği alındı")

		// Form verilerini al
		email := r.FormValue("email")
		AuthLogger.Printf("ForgotPassword - İstek email: %s", email)

		// Kullanıcı var mı kontrol et (gerçek uygulamada veritabanı sorgusu ile yapılmalı)
		if email != "" {
			AuthLogger.Printf("ForgotPassword - Sıfırlama bağlantısı gönderildi (simüle): %s", email)
			h.RedirectWithFlash(w, r, "/login", "Şifre sıfırlama bağlantısı e-posta adresinize gönderildi")
			return
		}

		AuthLogger.Printf("ForgotPassword - Geçersiz e-posta")
		h.RedirectWithFlash(w, r, "/forgot-password", "Geçersiz e-posta adresi")
		return
	}

	AuthLogger.Printf("ForgotPassword - GET isteği, şifre sıfırlama sayfası gösteriliyor")

	data := map[string]interface{}{
		"Title":          "Şifremi Unuttum - KolajAI",
		"PageHeading":    "Şifremi Unuttum",
		"PageSubHeading": "Şifrenizi sıfırlamak için e-posta adresinizi girin!",
	}

	h.RenderTemplate(w, r, "auth/forgot-password", data)
}

// ResetPassword handles the password reset process
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	AuthLogger.Printf("ResetPassword handler çağrıldı: Method=%s", r.Method)

	if r.Method == http.MethodPost {
		AuthLogger.Printf("ResetPassword - POST isteği alındı")

		// Form verilerini al
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		AuthLogger.Printf("ResetPassword - İstek email: %s", email)

		// Şifrelerin eşleştiğini kontrol et
		if password != confirmPassword {
			AuthLogger.Printf("ResetPassword - Şifreler eşleşmiyor")
			h.RedirectWithFlash(w, r, fmt.Sprintf("/reset-password?email=%s", email), "Şifreler eşleşmiyor")
			return
		}

		// Şifre değiştirme işlemi (gerçek uygulamada veritabanı güncellemesi yapılmalı)
		AuthLogger.Printf("ResetPassword - Şifre başarıyla değiştirildi (simüle): %s", email)
		h.RedirectWithFlash(w, r, "/login", "Şifreniz başarıyla değiştirildi. Şimdi giriş yapabilirsiniz.")
		return
	}

	// GET isteği için şifre sıfırlama sayfasını göster
	email := r.URL.Query().Get("email")
	token := r.URL.Query().Get("token")

	AuthLogger.Printf("ResetPassword - GET isteği, email=%s, token=%s", email, token)

	// Email parametresi kontrolü
	if email == "" {
		AuthLogger.Printf("ResetPassword - Email parametresi eksik, giriş sayfasına yönlendiriliyor")
		h.RedirectWithFlash(w, r, "/login", "Geçersiz şifre sıfırlama bağlantısı")
		return
	}

	// Token kontrolü (gerçek uygulamada veritabanı sorgusu ile doğrulanmalı)
	// Bu örnekte token kontrolünü atlıyoruz

	data := map[string]interface{}{
		"Title":          "Şifre Sıfırla - KolajAI",
		"PageHeading":    "Yeni Şifre Oluştur",
		"PageSubHeading": "Şifre sıfırlama talebinizi aldık. Lütfen yeni şifrenizi girin!",
		"Email":          email,
	}

	AuthLogger.Printf("ResetPassword - Şifre sıfırlama sayfası gösteriliyor: %s", email)
	h.RenderTemplate(w, r, "auth/reset-password", data)
}

// Register handles the user registration process
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	AuthLogger.Printf("Register handler çağrıldı: Method=%s", r.Method)

	if r.Method == http.MethodPost {
		AuthLogger.Printf("Register - POST isteği alındı")

		// Form verilerini al
		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")

		AuthLogger.Printf("Register - Kayıt denemesi: Name=%s, Email=%s", name, email)

		// Basit doğrulama
		if name == "" || email == "" || password == "" {
			AuthLogger.Printf("Register - Eksik form verileri")
			h.RedirectWithFlash(w, r, "/register", "Lütfen tüm alanları doldurun")
			return
		}

		// Kullanıcı kaydı (gerçek uygulamada veritabanına kayıt yapılmalı)
		// Bu örnek sadece demo amaçlıdır
		AuthLogger.Printf("Register - Kullanıcı başarıyla kaydedildi (simüle): %s", email)
		h.RedirectWithFlash(w, r, "/login", "Kaydınız başarıyla tamamlandı. Şimdi giriş yapabilirsiniz.")
		return
	}

	// GET isteği için kayıt sayfasını göster
	AuthLogger.Printf("Register - GET isteği, kayıt sayfası gösteriliyor")

	data := map[string]interface{}{
		"Title":          "Kayıt Ol - KolajAI",
		"PageHeading":    "Hesap Oluştur",
		"PageSubHeading": "Yeni bir hesap oluşturmak için lütfen bilgilerinizi girin!",
	}

	h.RenderTemplate(w, r, "auth/register", data)
}

// VerifyTempPassword handles temporary password verification
func (h *Handler) VerifyTempPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// JSON response için header ayarla
	w.Header().Set("Content-Type", "application/json")

	// Request body'yi parse et
	var req struct {
		Email        string `json:"email"`
		TempPassword string `json:"temp_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		AuthLogger.Printf("VerifyTempPassword - JSON parse hatası: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Geçersiz istek formatı",
		})
		return
	}

	// Basit doğrulama - production'da daha güçlü bir sistem olmalı
	// Şimdilik sadece başarılı response döndürelim
	AuthLogger.Printf("VerifyTempPassword - Email: %s, TempPassword: %s", req.Email, req.TempPassword)
	
	// Basit kontrol: temp password boş değilse geçerli kabul et
	if req.TempPassword != "" && req.Email != "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Geçici şifre doğrulandı",
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Geçersiz geçici şifre",
		})
	}
}
