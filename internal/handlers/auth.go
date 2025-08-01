package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	
	"kolajAi/internal/services"
)

var (
	AuthLogger *log.Logger
)

func init() {
	// Auth işlemleri için log dosyası oluştur
	logFile, err := os.OpenFile("auth_ops_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Auth log dosyası oluşturulamadı:", err)
		AuthLogger = log.New(os.Stdout, "[AUTH-OPS-DEBUG] ", log.LstdFlags)
	} else {
		AuthLogger = log.New(logFile, "[AUTH-OPS-DEBUG] ", log.LstdFlags|log.Lshortfile)
	}
}

// AuthHandler handles authentication related requests
type AuthHandler struct {
	*Handler
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(h *Handler, authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		Handler:     h,
		authService: authService,
	}
}

// Login handles the login request
func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	AuthLogger.Printf("Login handler çağrıldı: Method=%s, URL=%s", r.Method, r.URL.Path)

	// Eğer kullanıcı zaten oturum açmışsa, anasayfaya yönlendir
	if ah.IsAuthenticated(r) {
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
			ah.RedirectWithFlash(w, r, "/login", "Form işlenirken hata oluştu")
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		AuthLogger.Printf("Login - Giriş denemesi: Email=%s", email)

		// AuthService kullanarak doğrulama yap
		user, err := ah.authService.LoginUser(email, password)
		if err != nil {
			AuthLogger.Printf("Login - Giriş hatası: %v", err)
			ah.RedirectWithFlash(w, r, "/login", "Hatalı e-posta veya şifre")
			return
		}

		// Kullanıcı permissions'larını al
		permissions := user.GetPermissions()
		
		AuthLogger.Printf("Login - Başarılı giriş: UserID=%d, Email=%s, Role=%s", user.ID, user.Email, user.Role)

		// Gelişmiş session manager ile oturum oluştur
		sessionData, err := ah.SessionManager.CreateSession(w, r, user.ID, permissions)
		if err != nil {
			AuthLogger.Printf("Login - Oturum oluşturma hatası: %v", err)
			ah.RedirectWithFlash(w, r, "/login", "Oturum oluşturulurken hata oluştu")
			return
		}

		AuthLogger.Printf("Login - Oturum başarıyla oluşturuldu: SessionID=%s, UserID=%d", sessionData.ID, sessionData.UserID)
		
		// Dashboard'a yönlendir
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
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
	ah.RenderTemplate(w, r, "auth/login", data)
}

// Logout logs out the user
func (ah *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	AuthLogger.Printf("Logout handler çağrıldı: Method=%s, URL=%s", r.Method, r.URL.Path)

	// Session'ı sonlandır
	err := ah.SessionManager.DestroySession(w, r)
	if err != nil {
		AuthLogger.Printf("Logout - Session sonlandırma hatası: %v", err)
	} else {
		AuthLogger.Printf("Logout - Session başarıyla sonlandırıldı")
	}

	// Kullanıcıyı login sayfasına yönlendir
	AuthLogger.Printf("Logout - Kullanıcı login sayfasına yönlendiriliyor")
	ah.RedirectWithFlash(w, r, "/login", "Başarıyla çıkış yapıldı")
}

// ForgotPassword handles the forgot password request
func (ah *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	AuthLogger.Printf("ForgotPassword handler çağrıldı: Method=%s", r.Method)

	if r.Method == http.MethodPost {
		AuthLogger.Printf("ForgotPassword - POST isteği alındı")

		// Form verilerini al
		email := r.FormValue("email")
		AuthLogger.Printf("ForgotPassword - İstek email: %s", email)

		// AuthService kullanarak şifre sıfırlama işlemini başlat
		err := ah.authService.ForgotPassword(email)
		if err != nil {
			AuthLogger.Printf("ForgotPassword - Hata: %v", err)
			ah.RedirectWithFlash(w, r, "/forgot-password", "İşlem sırasında bir hata oluştu")
			return
		}

		AuthLogger.Printf("ForgotPassword - Sıfırlama e-postası gönderildi: %s", email)
		ah.RedirectWithFlash(w, r, "/login", "Şifre sıfırlama bağlantısı e-posta adresinize gönderildi")
		return
	}

	AuthLogger.Printf("ForgotPassword - GET isteği, şifre sıfırlama sayfası gösteriliyor")

	data := map[string]interface{}{
		"Title":          "Şifremi Unuttum - KolajAI",
		"PageHeading":    "Şifremi Unuttum",
		"PageSubHeading": "Şifrenizi sıfırlamak için e-posta adresinizi girin!",
	}

	ah.RenderTemplate(w, r, "auth/forgot-password", data)
}

// ResetPassword handles the password reset process
func (ah *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
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
			ah.RedirectWithFlash(w, r, fmt.Sprintf("/reset-password?email=%s", email), "Şifreler eşleşmiyor")
			return
		}

		// AuthService kullanarak şifre değiştir
		err := ah.authService.ResetPassword(email, password)
		if err != nil {
			AuthLogger.Printf("ResetPassword - Hata: %v", err)
			ah.RedirectWithFlash(w, r, fmt.Sprintf("/reset-password?email=%s", email), "Şifre değiştirme işlemi başarısız")
			return
		}

		AuthLogger.Printf("ResetPassword - Şifre başarıyla değiştirildi: %s", email)
		ah.RedirectWithFlash(w, r, "/login", "Şifreniz başarıyla değiştirildi. Şimdi giriş yapabilirsiniz.")
		return
	}

	// GET isteği için şifre sıfırlama sayfasını göster
	email := r.URL.Query().Get("email")
	token := r.URL.Query().Get("token")

	AuthLogger.Printf("ResetPassword - GET isteği, email=%s, token=%s", email, token)

	// Email parametresi kontrolü
	if email == "" {
		AuthLogger.Printf("ResetPassword - Email parametresi eksik, giriş sayfasına yönlendiriliyor")
		ah.RedirectWithFlash(w, r, "/login", "Geçersiz şifre sıfırlama bağlantısı")
		return
	}

	data := map[string]interface{}{
		"Title":          "Şifre Sıfırla - KolajAI",
		"PageHeading":    "Yeni Şifre Oluştur",
		"PageSubHeading": "Şifre sıfırlama talebinizi aldık. Lütfen yeni şifrenizi girin!",
		"Email":          email,
	}

	AuthLogger.Printf("ResetPassword - Şifre sıfırlama sayfası gösteriliyor: %s", email)
	ah.RenderTemplate(w, r, "auth/reset-password", data)
}

// Register handles the user registration process
func (ah *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	AuthLogger.Printf("Register handler çağrıldı: Method=%s", r.Method)

	if r.Method == http.MethodPost {
		AuthLogger.Printf("Register - POST isteği alındı")

		// Form verilerini al
		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")
		phone := r.FormValue("phone")

		AuthLogger.Printf("Register - Kayıt denemesi: Name=%s, Email=%s", name, email)

		// Basit doğrulama
		if name == "" || email == "" || password == "" {
			AuthLogger.Printf("Register - Eksik form verileri")
			ah.RedirectWithFlash(w, r, "/register", "Lütfen tüm alanları doldurun")
			return
		}

		// AuthService kullanarak kullanıcı kaydı yap
		userData := map[string]string{
			"name":     name,
			"email":    email,
			"phone":    phone,
			"password": password, // Kullanıcının verdiği şifreyi ekle
		}
		
		userID, err := ah.authService.RegisterUser(userData)
		if err != nil {
			AuthLogger.Printf("Register - Kayıt hatası: %v", err)
			ah.RedirectWithFlash(w, r, "/register", "Kayıt işlemi başarısız: " + err.Error())
			return
		}

		AuthLogger.Printf("Register - Kullanıcı başarıyla kaydedildi: UserID=%d, Email=%s", userID, email)
		ah.RedirectWithFlash(w, r, "/login", "Kaydınız başarıyla tamamlandı. Şimdi giriş yapabilirsiniz.")
		return
	}

	// GET isteği için kayıt sayfasını göster
	AuthLogger.Printf("Register - GET isteği, kayıt sayfası gösteriliyor")

	data := map[string]interface{}{
		"Title":          "Kayıt Ol - KolajAI",
		"PageHeading":    "Hesap Oluştur",
		"PageSubHeading": "Yeni bir hesap oluşturmak için lütfen bilgilerinizi girin!",
	}

	ah.RenderTemplate(w, r, "auth/register", data)
}
