package services

import (
	"fmt"
	"log"
	"strings"

	"kolajAi/internal/config"
	"kolajAi/internal/core"
	"kolajAi/internal/email"
	"kolajAi/internal/models"
	"kolajAi/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication related business logic
type AuthService struct {
	userRepo *repository.UserRepository
	emailSvc *email.Service
	baseURL  string
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo *repository.UserRepository, emailSvc *email.Service) *AuthService {
	// Yapılandırmadan baseURL'i al, yoksa varsayılan değeri kullan
	baseURL := "http://localhost:8080" // Varsayılan değer
	appConfig, err := config.LoadConfig("config.yaml")
	if err == nil && appConfig != nil {
		// Server host ve port bilgisini kullan
		if appConfig.Server.Host != "" {
			scheme := "http"
			if appConfig.Server.Port == 443 {
				scheme = "https"
			}
			baseURL = fmt.Sprintf("%s://%s", scheme, appConfig.Server.Host)
			if appConfig.Server.Port != 80 && appConfig.Server.Port != 443 {
				baseURL = fmt.Sprintf("%s:%d", baseURL, appConfig.Server.Port)
			}
		}
	}

	return &AuthService{
		userRepo: userRepo,
		emailSvc: emailSvc,
		baseURL:  baseURL,
	}
}

// RegisterUser registers a new user
func (s *AuthService) RegisterUser(userData map[string]string) (int64, error) {
	// Check if email already exists
	email := userData["email"]
	exists, err := s.userRepo.EmailExists(email)
	if err != nil {
		log.Printf("Error checking email existence: %v", err)
		return 0, core.NewDatabaseError("Email kontrolü yapılırken hata oluştu", err)
	}
	if exists {
		log.Printf("Email already exists: %s", email)
		return 0, core.NewValidationError("Bu e-posta adresi zaten kullanılmakta", map[string][]string{
			"email": {"Bu e-posta adresi zaten kullanılmakta"},
		})
	}

	// Generate a random password
	randomPassword := s.GenerateRandomPassword(10)
	log.Printf("Generated random password for user %s", email)

	// Hash the password
	hashedPassword, err := s.CreateUserPassword(randomPassword)
	if err != nil {
		log.Printf("Error creating password hash: %v", err)
		return 0, core.NewAuthError("Şifre oluşturma hatası", err)
	}

	// Create user in database
	userID, err := s.userRepo.RegisterUser(
		userData["name"],
		email,
		hashedPassword,
		userData["phone"],
	)
	if err != nil {
		log.Printf("Error registering user: %v", err)
		return 0, core.NewDatabaseError("Kullanıcı kaydı yapılırken hata oluştu", err)
	}

	// Send welcome email with password
	err = s.SendWelcomeEmail(email, userData["name"], randomPassword)
	if err != nil {
		log.Printf("Failed to send welcome email: %v", err)
		// Don't fail the registration, just log the error
	}

	return userID, nil
}

// SendWelcomeEmail sends a welcome email with password to the user
func (s *AuthService) SendWelcomeEmail(to string, name string, password string) error {
	data := map[string]interface{}{
		"Name":         name,
		"Email":        to,
		"Password":     password,
		"AlertTitle":   "Hesabınız Oluşturuldu",
		"AlertContent": "Hesabınıza giriş yapmak için aşağıdaki bilgileri kullanabilirsiniz. Güvenliğiniz için lütfen ilk girişinizde şifrenizi değiştirin.",
		"ButtonLink":   s.baseURL + "/reset-password?email=" + to,
	}

	return s.emailSvc.SendTemplateEmail(to, "Hoş Geldiniz - KolajAI", "welcome", data)
}

// LoginUser logs in a user
func (s *AuthService) LoginUser(email, password string) (*models.User, error) {
	// Kullanıcıyı bul
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		log.Printf("Error finding user for login: %v", err)
		return nil, core.NewAuthError("Kullanıcı bulunamadı", err)
	}

	if user == nil {
		log.Printf("User not found: %s", email)
		return nil, core.NewAuthError("Kullanıcı bulunamadı", nil)
	}

	// Şifre boş mu kontrol et
	if user.Password == "" {
		log.Printf("ERROR - LoginUser: Password is empty for user: %s", email)
		return nil, core.NewAuthError("Şifre bulunamadı", nil)
	}

	// Şifre doğrulama
	log.Printf("DEBUG - LoginUser: Comparing passwords for user: %s", email)
	log.Printf("DEBUG - LoginUser: Password verification started")

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Printf("Invalid password for user %s: %v", email, err)
		return nil, core.NewAuthError("Geçersiz e-posta veya şifre", err)
	}

	// Hesap aktif mi kontrol et
	if !user.IsActive {
		log.Printf("Account not active: %s", email)
		return nil, core.NewAuthError("Hesabınız aktif değil. Lütfen e-postanızı kontrol edin veya yönetici ile iletişime geçin", nil)
	}

	log.Printf("User logged in successfully: %s", email)
	return user, nil
}

// CreateUserPassword creates a hashed password for a user
func (s *AuthService) CreateUserPassword(password string) (string, error) {
	// Şifreyi hashle
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return "", core.NewAuthError("Şifre hash'leme hatası", err)
	}

	// Hash'lenmiş şifreyi string'e dönüştür
	hashedPasswordStr := string(hashedPassword)
	log.Printf("DEBUG - CreateUserPassword: Şifre başarıyla hash'lendi")
	log.Printf("DEBUG - CreateUserPassword: Hash bcrypt formatında: %v",
		strings.HasPrefix(hashedPasswordStr, "$2a$") ||
			strings.HasPrefix(hashedPasswordStr, "$2b$") ||
			strings.HasPrefix(hashedPasswordStr, "$2y$"))

	return hashedPasswordStr, nil
}

// VerifyPassword verifies a password against a hash
func (s *AuthService) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GenerateRandomPassword generates a random password
func (s *AuthService) GenerateRandomPassword(length int) string {
	if length < 8 {
		length = 8 // Minimum 8 karakter
	}

	return uuid.New().String()[:length]
}

// ActivateAccount activates a user account
func (s *AuthService) ActivateAccount(userID int64) error {
	// Activate the account
	err := s.userRepo.ActivateAccount(userID)
	if err != nil {
		return core.NewDatabaseError("Hesap aktifleştirme sırasında hata oluştu", err)
	}

	return nil
}

// ForgotPassword initiates the password reset process
func (s *AuthService) ForgotPassword(email string) error {
	// Check if user exists
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		// Don't reveal if email exists or not
		log.Printf("Forgot password request for non-existing email: %s", err)
		return nil
	}

	if user == nil {
		log.Printf("User not found for forgot password: %s", email)
		return nil // Kullanıcı bulunamadı hatası verme, güvenlik nedeniyle
	}

	// Generate temporary password
	tempPassword := s.GenerateRandomPassword(10)
	log.Printf("Generated temporary password for user %s", email)

	// Hash the password
	hashedPassword, err := s.CreateUserPassword(tempPassword)
	if err != nil {
		log.Printf("Error creating password hash: %v", err)
		return core.NewAuthError("Şifre oluşturma hatası", err)
	}

	// Update user's password
	err = s.userRepo.ResetUserPassword(email, hashedPassword)
	if err != nil {
		log.Printf("Error resetting user password: %v", err)
		return core.NewDatabaseError("Şifre sıfırlama işlemi sırasında hata oluştu", err)
	}

	// Send reset email with temp password
	data := map[string]interface{}{
		"Name":         user.Name,
		"Email":        email,
		"Password":     tempPassword,
		"AlertTitle":   "Geçici Şifre",
		"AlertContent": "Aşağıdaki geçici şifre ile giriş yapabilirsiniz. Güvenliğiniz için lütfen giriş yaptıktan sonra şifrenizi değiştirin.",
		"ButtonLink":   s.baseURL + "/reset-password?email=" + email,
	}

	err = s.emailSvc.SendTemplateEmail(email, "Şifre Sıfırlama - KolajAI", "welcome", data)
	if err != nil {
		log.Printf("Error sending password reset email: %v", err)
		return core.NewAuthError("Şifre sıfırlama e-postası gönderilemedi", err)
	}

	log.Printf("Password reset email sent to: %s", email)
	return nil
}

// ResetPassword resets a user's password
func (s *AuthService) ResetPassword(email, newPassword string) error {
	// Find user by email
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		log.Printf("Error finding user for reset password: %v", err)
		return core.NewAuthError("Kullanıcı bulunamadı", err)
	}

	if user == nil {
		log.Printf("User not found for reset password: %s", email)
		return core.NewAuthError("Kullanıcı bulunamadı", nil)
	}

	// Hash the new password
	hashedPassword, err := s.CreateUserPassword(newPassword)
	if err != nil {
		log.Printf("Error creating password hash: %v", err)
		return core.NewAuthError("Şifre oluşturma hatası", err)
	}

	// Update password
	err = s.userRepo.ResetUserPassword(email, hashedPassword)
	if err != nil {
		log.Printf("Error updating user password: %v", err)
		return core.NewDatabaseError("Şifre güncelleme işlemi sırasında hata oluştu", err)
	}

	log.Printf("Password successfully reset for user: %s", email)
	return nil
}
