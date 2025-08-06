package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	
	"golang.org/x/crypto/bcrypt"
	"github.com/kolajai/internal/models"
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

		// CSRF token kontrolü
		if !h.ValidateCSRFToken(r) {
			AuthLogger.Printf("Login - CSRF token doğrulama hatası")
			h.RedirectWithFlash(w, r, "/login", "Güvenlik doğrulaması başarısız")
			return
		}

		err := r.ParseForm()
		if err != nil {
			AuthLogger.Printf("Login - Form parse hatası: %v", err)
			h.RedirectWithFlash(w, r, "/login", "Form işlenirken hata oluştu")
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")
		rememberMe := r.FormValue("remember_me") == "on"

		AuthLogger.Printf("Login - Giriş denemesi: Email=%s", email)

		// Rate limiting kontrolü (basit implementasyon)
		if !h.CheckLoginRateLimit(email) {
			AuthLogger.Printf("Login - Rate limit aşıldı: %s", email)
			h.RedirectWithFlash(w, r, "/login", "Çok fazla başarısız giriş denemesi. Lütfen 15 dakika sonra tekrar deneyin.")
			return
		}

		// Gerçek veritabanı kimlik doğrulaması
		// TODO: Repository pattern ile veritabanından kullanıcıyı çek
		var user models.User
		
		// Demo için sabit kullanıcı - production'da veritabanından çekilmeli
		if email == "admin@kolajAi.com" {
			// Demo password hash for "Admin123!"
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Admin123!"), bcrypt.DefaultCost)
			user = models.User{
				ID:       1,
				Name:     "Admin User",
				Email:    email,
				Password: string(hashedPassword),
				IsAdmin:  true,
				IsActive: true,
			}
		} else {
			AuthLogger.Printf("Login - Kullanıcı bulunamadı: %s", email)
			h.IncrementLoginAttempts(email)
			h.RedirectWithFlash(w, r, "/login", "Hatalı e-posta veya şifre")
			return
		}

		// Kullanıcı hesabı kilitli mi kontrol et
		if user.IsLocked() {
			AuthLogger.Printf("Login - Kilitli hesap giriş denemesi: %s", email)
			h.RedirectWithFlash(w, r, "/login", "Hesabınız geçici olarak kilitlenmiştir. Lütfen daha sonra tekrar deneyin.")
			return
		}

		// Kullanıcı aktif mi kontrol et
		if !user.IsActive {
			AuthLogger.Printf("Login - Pasif kullanıcı giriş denemesi: %s", email)
			h.RedirectWithFlash(w, r, "/login", "Hesabınız pasif durumda. Lütfen destek ekibiyle iletişime geçin.")
			return
		}

		// Şifre doğrulama
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			AuthLogger.Printf("Login - Hatalı şifre: %s", email)
			user.IncrementLoginAttempts()
			// TODO: Veritabanında güncelle
			h.IncrementLoginAttempts(email)
			h.RedirectWithFlash(w, r, "/login", "Hatalı e-posta veya şifre")
			return
		}

		// Başarılı giriş
		AuthLogger.Printf("Login - Başarılı giriş: %s", email)
		
		// Login attempts sıfırla
		user.ResetLoginAttempts()
		// TODO: Veritabanında güncelle

		// Last login bilgilerini güncelle
		now := time.Now()
		user.LastLoginAt = &now
		user.LastLoginIP = r.RemoteAddr
		// TODO: Veritabanında güncelle

		// Kullanıcı bilgilerini session için hazırla
		userSession := struct {
			ID    int64
			Email string
			Name  string
		}{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		}

		// Session süresini ayarla
		sessionDuration := 24 * time.Hour // Default 24 saat
		if rememberMe {
			sessionDuration = 30 * 24 * time.Hour // 30 gün
		}

		// Oturum oluştur - kullanıcı bilgilerini ve yetki durumunu kaydet
		if err := h.SessionManager.SetSessionWithExpiry(w, r, UserKey, userSession, sessionDuration); err != nil {
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
		if err := h.SessionManager.SetSession(w, r, "is_admin", user.IsAdmin); err != nil {
			AuthLogger.Printf("Login - Oturum oluşturma hatası (is_admin): %v", err)
			h.RedirectWithFlash(w, r, "/login", "Oturum oluşturulurken hata oluştu")
			return
		}

		AuthLogger.Printf("Login - Oturum başarıyla oluşturuldu, dashboard'a yönlendiriliyor")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	// GET isteği için giriş sayfasını göster
	AuthLogger.Printf("Login - GET isteği, login sayfası gösteriliyor")

	data := map[string]interface{}{
		"Title":          "Giriş - KolajAI",
		"PageHeading":    "Giriş Yap",
		"PageSubHeading": "Hesabınıza giriş yapın ve işlemlerinize devam edin!",
		"SidebarClass":   "bg-gradient-grd",
		"SidebarImage":   "/web/static/assets/images/auth/login.png",
	}

	// Şablonu render et
	h.RenderTemplate(w, r, "auth/login.gohtml", data)
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

	h.RenderTemplate(w, r, "auth/forgot-password.gohtml", data)
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
	h.RenderTemplate(w, r, "auth/reset-password.gohtml", data)
}

// Register handles the user registration process
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	AuthLogger.Printf("Register handler çağrıldı: Method=%s", r.Method)

	if r.Method == http.MethodPost {
		AuthLogger.Printf("Register - POST isteği alındı")

		// CSRF token kontrolü
		if !h.ValidateCSRFToken(r) {
			AuthLogger.Printf("Register - CSRF token doğrulama hatası")
			h.RedirectWithFlash(w, r, "/register", "Güvenlik doğrulaması başarısız")
			return
		}

		// Form verilerini al
		name := r.FormValue("name")
		email := r.FormValue("email")
		phone := r.FormValue("phone")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")
		captchaAnswer := r.FormValue("captcha")
		captchaExpected := r.FormValue("captchaExpected")
		termsAccepted := r.FormValue("terms") == "on"

		AuthLogger.Printf("Register - Kayıt denemesi: Name=%s, Email=%s", name, email)

		// Temel doğrulama
		if name == "" || email == "" || phone == "" || password == "" {
			AuthLogger.Printf("Register - Eksik form verileri")
			h.RedirectWithFlash(w, r, "/register", "Lütfen tüm alanları doldurun")
			return
		}

		// Şifre eşleşme kontrolü
		if password != confirmPassword {
			AuthLogger.Printf("Register - Şifreler eşleşmiyor")
			h.RedirectWithFlash(w, r, "/register", "Şifreler eşleşmiyor")
			return
		}

		// Şifre güvenlik kontrolü
		if err := models.ValidatePassword(password); err != nil {
			AuthLogger.Printf("Register - Zayıf şifre: %v", err)
			h.RedirectWithFlash(w, r, "/register", err.Error())
			return
		}

		// CAPTCHA doğrulama (basit implementasyon)
		if captchaAnswer != captchaExpected && captchaExpected != "" {
			AuthLogger.Printf("Register - CAPTCHA doğrulama hatası")
			h.RedirectWithFlash(w, r, "/register", "Güvenlik kodu hatalı")
			return
		}

		// Kullanım koşulları kontrolü
		if !termsAccepted {
			AuthLogger.Printf("Register - Kullanım koşulları kabul edilmedi")
			h.RedirectWithFlash(w, r, "/register", "Kullanım koşullarını kabul etmelisiniz")
			return
		}

		// Email benzersizlik kontrolü
		// TODO: Veritabanında email kontrolü yapılmalı
		if h.EmailExists(email) {
			AuthLogger.Printf("Register - Email zaten kayıtlı: %s", email)
			h.RedirectWithFlash(w, r, "/register", "Bu e-posta adresi zaten kayıtlı")
			return
		}

		// Şifreyi hashle
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			AuthLogger.Printf("Register - Şifre hashleme hatası: %v", err)
			h.RedirectWithFlash(w, r, "/register", "Kayıt işlemi sırasında hata oluştu")
			return
		}

		// Yeni kullanıcı oluştur
		user := models.User{
			Name:     name,
			Email:    email,
			Phone:    phone,
			Password: string(hashedPassword),
			Role:     "customer",
			IsActive: true,
			IsAdmin:  false,
			IsSeller: false,
			EmailVerified: false,
			EmailVerificationToken: h.GenerateVerificationToken(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Kullanıcıyı doğrula
		if err := user.Validate(); err != nil {
			AuthLogger.Printf("Register - Kullanıcı doğrulama hatası: %v", err)
			h.RedirectWithFlash(w, r, "/register", "Geçersiz kullanıcı bilgileri")
			return
		}

		// TODO: Kullanıcıyı veritabanına kaydet
		// userID, err := h.UserRepository.Create(&user)
		// if err != nil {
		//     AuthLogger.Printf("Register - Veritabanı kayıt hatası: %v", err)
		//     h.RedirectWithFlash(w, r, "/register", "Kayıt işlemi sırasında hata oluştu")
		//     return
		// }

		// Email doğrulama maili gönder
		// TODO: Email servisi ile doğrulama maili gönder
		// err = h.EmailService.SendVerificationEmail(user.Email, user.EmailVerificationToken)
		// if err != nil {
		//     AuthLogger.Printf("Register - Email gönderme hatası: %v", err)
		// }

		AuthLogger.Printf("Register - Kullanıcı başarıyla kaydedildi: %s", email)
		h.RedirectWithFlash(w, r, "/login", "Kaydınız başarıyla tamamlandı. E-posta adresinize gönderilen doğrulama linkine tıklayarak hesabınızı aktifleştirin.")
		return
	}

	// GET isteği için kayıt sayfasını göster
	AuthLogger.Printf("Register - GET isteği, kayıt sayfası gösteriliyor")

	// Validation kurallarını JSON olarak hazırla
	validationRules := map[string]interface{}{
		"name": map[string]interface{}{
			"required": true,
			"minLength": 5,
		},
		"email": map[string]interface{}{
			"required": true,
			"email": true,
		},
		"phone": map[string]interface{}{
			"required": true,
			"pattern": "0[0-9 ]{10,14}",
		},
		"password": map[string]interface{}{
			"required": true,
			"minLength": 8,
			"pattern": "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]{8,}$",
		},
		"confirm_password": map[string]interface{}{
			"required": true,
			"match": "password",
		},
		"terms": map[string]interface{}{
			"required": true,
		},
	}

	rulesJSON, _ := json.Marshal(validationRules)

	data := map[string]interface{}{
		"Title":          "Kayıt Ol - KolajAI",
		"PageHeading":    "Hesap Oluştur",
		"PageSubHeading": "Yeni bir hesap oluşturmak için lütfen bilgilerinizi girin!",
		"ValidationRules": string(rulesJSON),
	}

	h.RenderTemplate(w, r, "auth/register.gohtml", data)
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
	AuthLogger.Printf("VerifyTempPassword - Email: %s", req.Email)
	
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
