package repository

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"kolajAi/internal/core"
	"kolajAi/internal/database"
	"kolajAi/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// UserRepository handles user-related database operations
type UserRepository struct {
	db       *database.MySQLRepository
	baseRepo *BaseRepository
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *database.MySQLRepository) *UserRepository {
	return &UserRepository{
		db:       db,
		baseRepo: NewBaseRepository(db),
	}
}

// RegisterUser registers a new user
func (r *UserRepository) RegisterUser(name, email, password, phone string) (int64, error) {
	// Şifre kontrolü - eğer şifre zaten hash'lenmişse tekrar hash'leme
	if strings.HasPrefix(password, "$2a$") || strings.HasPrefix(password, "$2b$") || strings.HasPrefix(password, "$2y$") {
		log.Printf("INFO - RegisterUser: Şifre zaten hash'lenmiş, doğrudan kaydediliyor")
		hashedPasswordStr := password

		// Kullanıcı verilerini hazırla
		userData := map[string]interface{}{
			"name":       name,
			"email":      email,
			"password":   hashedPasswordStr, // String olarak kaydet
			"phone":      phone,
			"is_active":  true,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		}

		// Kullanıcıyı veritabanına kaydet
		userID, err := r.baseRepo.Create("users", userData)
		if err != nil {
			log.Printf("ERROR - RegisterUser: Kullanıcı kayıt hatası: %v", err)
			return 0, core.NewDatabaseError("error creating user", err)
		}

		return userID, nil
	}

	// Şifreyi hash'le
	log.Printf("INFO - RegisterUser: Şifre hash'leme başlıyor")
	// Use stronger bcrypt cost for production security (12 instead of default 10)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Printf("ERROR - RegisterUser: Şifre hash'leme hatası: %v", err)
		return 0, core.NewDatabaseError("error hashing password", err)
	}

	// Hash'lenmiş şifreyi string'e dönüştür
	hashedPasswordStr := string(hashedPassword)
	log.Printf("DEBUG - RegisterUser: Şifre başarıyla hash'lendi")
	log.Printf("DEBUG - RegisterUser: Hash bcrypt formatında: %v",
		strings.HasPrefix(hashedPasswordStr, "$2a$") ||
			strings.HasPrefix(hashedPasswordStr, "$2b$") ||
			strings.HasPrefix(hashedPasswordStr, "$2y$"))

	// Kullanıcı verilerini hazırla
	userData := map[string]interface{}{
		"name":       name,
		"email":      email,
		"password":   hashedPasswordStr, // String olarak kaydet
		"phone":      phone,
		"is_active":  true,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	// Kullanıcıyı veritabanına kaydet
	userID, err := r.baseRepo.Create("users", userData)
	if err != nil {
		log.Printf("ERROR - RegisterUser: Kullanıcı kayıt hatası: %v", err)
		return 0, core.NewDatabaseError("error creating user", err)
	}

	// Oluşturulan kullanıcıyı kontrol et
	createdUser, err := r.FindByEmail(email)
	if err != nil {
		log.Printf("WARN - RegisterUser: Oluşturulan kullanıcı kontrol edilemedi: %v", err)
		return userID, nil // Yine de başarılı kabul et
	}

	// Şifrenin doğru kaydedilip kaydedilmediğini kontrol et
	log.Printf("DEBUG - RegisterUser: Şifre veritabanına başarıyla kaydedildi")
	log.Printf("DEBUG - RegisterUser: Kaydedilen şifre bcrypt formatında: %v",
		strings.HasPrefix(createdUser.Password, "$2a$") ||
			strings.HasPrefix(createdUser.Password, "$2b$") ||
			strings.HasPrefix(createdUser.Password, "$2y$"))

	return userID, nil
}

// VerifyTempPassword verifies the temporary password for a user
// Bu fonksiyon kullanıcının geçici şifresini doğrular
// Parametreler:
//   - email: Kullanıcının e-posta adresi
//   - tempPassword: Kullanıcının girdiği geçici şifre
//
// Dönüş değerleri:
//   - bool: Doğrulama başarılı mı?
//   - error: Hata durumu
func (r *UserRepository) VerifyTempPassword(email, tempPassword string) (bool, error) {
	// Kullanıcıyı önce e-posta ile bulalım
	user, err := r.FindByEmail(email)
	if err != nil {
		log.Printf("ERROR - VerifyTempPassword: E-posta ile kullanıcı bulunamadı: %v", err)
		return false, core.NewDatabaseError("error finding user by email", err)
	}
	if user == nil {
		log.Printf("ERROR - VerifyTempPassword: Kullanıcı bulunamadı: %s", email)
		return false, nil
	}

	// Şifre karşılaştırma detaylı debug
	log.Printf("DEBUG - VerifyTempPassword: Şifre karşılaştırma başlıyor")
	log.Printf("DEBUG - VerifyTempPassword: Şifre doğrulama işlemi başlatıldı")
	log.Printf("DEBUG - VerifyTempPassword: Veritabanından hash alındı")

	// Veritabanındaki şifre hash'i boş veya geçersiz mi kontrol et
	if user.Password == "" {
		log.Printf("ERROR - VerifyTempPassword: Veritabanındaki şifre hash'i boş")

		// BaseRepo üzerinden tekrar kullanıcıyı sorgula
		var freshUser models.User
		conditions := map[string]interface{}{
			"email":     email,
			"is_active": true,
		}

		err = r.baseRepo.FindOne("users", &freshUser, conditions)
		if err != nil {
			log.Printf("ERROR - VerifyTempPassword: Tekrar sorgulama hatası: %v", err)
			return false, nil
		}

		log.Printf("DEBUG - VerifyTempPassword: Kullanıcı tekrar sorgulandı")

		// Şifre hala boş mu kontrol et
		if freshUser.Password == "" {
			log.Printf("ERROR - VerifyTempPassword: Tekrar sorgulanan kullanıcının şifresi de boş")
			return false, nil
		} else {
			// Şifre hash'ini güncelle ve tekrar dene
			user.Password = freshUser.Password
		}
	}

	// Hash formatı kontrolü
	if !strings.HasPrefix(user.Password, "$2a$") &&
		!strings.HasPrefix(user.Password, "$2b$") &&
		!strings.HasPrefix(user.Password, "$2y$") {
		log.Printf("ERROR - VerifyTempPassword: Veritabanındaki şifre geçerli bir bcrypt hash'i değil")
		return false, nil
	}

	// Panic durumlarını yakalamak için defer kullanımı
	defer func() {
		if r := recover(); r != nil {
			log.Printf("WARN - VerifyTempPassword: Şifre karşılaştırma hatası (panic): %v", r)
		}
	}()

	// Bcrypt ile şifre karşılaştırma
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(tempPassword))
	if err == nil {
		log.Printf("SUCCESS - VerifyTempPassword: Şifre doğrulandı (bcrypt)")
		return true, nil
	}

	log.Printf("ERROR - VerifyTempPassword: Şifre eşleşmedi: %v", err)
	return false, nil
}

// ResetUserPassword resets a user's password
func (r *UserRepository) ResetUserPassword(email, newPassword string) error {
	// Kullanıcıyı önce e-posta ile bulalım
	log.Printf("DEBUG - ResetUserPassword: Şifre değiştirme işlemi başlatılıyor. Email: %s", email)
	user, err := r.FindByEmail(email)
	if err != nil {
		log.Printf("ERROR - ResetUserPassword: Kullanıcı bulunamadı: %v", err)
		return core.NewDatabaseError("error finding user for reset password", err)
	}
	if user == nil {
		log.Printf("ERROR - ResetUserPassword: %s e-posta adresi ile kullanıcı bulunamadı", email)
		return core.NewDatabaseError("user not found", nil)
	}

	log.Printf("INFO - ResetUserPassword: Kullanıcı bulundu: %s (ID: %d, Aktif: %v)", email, user.ID, user.IsActive)

	// Şifre kontrolü - eğer şifre zaten hash'lenmişse tekrar hash'leme
	var hashedPasswordStr string
	if strings.HasPrefix(newPassword, "$2a$") || strings.HasPrefix(newPassword, "$2b$") || strings.HasPrefix(newPassword, "$2y$") {
		log.Printf("INFO - ResetUserPassword: Şifre zaten hash'lenmiş, doğrudan kaydediliyor")
		hashedPasswordStr = newPassword
	} else {
		// Yeni şifreyi hashle
		log.Printf("INFO - ResetUserPassword: Şifre hash'leme başlıyor. Şifre uzunluğu: %d", len(newPassword))
		// Use stronger bcrypt cost for production security (12 instead of default 10)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
		if err != nil {
			log.Printf("ERROR - ResetUserPassword: Şifre hash'leme hatası: %v", err)
			return core.NewDatabaseError("error hashing password", err)
		}

		// Hash'lenmiş şifreyi string'e dönüştür
		hashedPasswordStr = string(hashedPassword)
			log.Printf("DEBUG - ResetUserPassword: Şifre başarıyla hash'lendi")
	log.Printf("DEBUG - ResetUserPassword: Hash bcrypt formatında: %v",
		strings.HasPrefix(hashedPasswordStr, "$2a$") ||
			strings.HasPrefix(hashedPasswordStr, "$2b$") ||
			strings.HasPrefix(hashedPasswordStr, "$2y$"))
	}

	// Doğrudan SQL sorgusu ile şifre güncelleme ve hesabı aktifleştirme
	log.Printf("INFO - ResetUserPassword: Şifre ve hesap durumu güncelleniyor. Kullanıcı ID: %d", user.ID)
	updateQuery := "UPDATE users SET password = ?, is_active = true, updated_at = ? WHERE id = ?"
	stmt, err := database.DB.Prepare(updateQuery)
	if err != nil {
		log.Printf("ERROR - ResetUserPassword: SQL hazırlama hatası: %v", err)
		return core.NewDatabaseError("error preparing SQL statement", err)
	}
	defer stmt.Close()

	now := time.Now()
	result, err := stmt.Exec(hashedPasswordStr, now, user.ID)
	if err != nil {
		log.Printf("ERROR - ResetUserPassword: SQL çalıştırma hatası: %v", err)
		return core.NewDatabaseError("error executing SQL statement", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("ERROR - ResetUserPassword: Etkilenen satır sayısı alınamadı: %v", err)
	} else {
		log.Printf("INFO - ResetUserPassword: Etkilenen satır sayısı: %d", rowsAffected)
	}

	// Başarılı olup olmadığını kontrol et
	if rowsAffected == 0 {
		log.Printf("WARN - ResetUserPassword: Hiçbir satır etkilenmedi, kullanıcı bulunamadı veya zaten güncel: %d", user.ID)
		return core.NewDatabaseError("no rows affected", nil)
	}

	// Güncellenen kullanıcıyı kontrol et
	updatedUser, err := r.FindByEmail(email)
	if err != nil {
		log.Printf("WARN - ResetUserPassword: Güncellenen kullanıcı kontrol edilemedi: %v", err)
		return nil // Yine de başarılı kabul et
	}

	// Şifrenin doğru kaydedilip kaydedilmediğini kontrol et
	log.Printf("DEBUG - ResetUserPassword: Şifre veritabanına kaydedildi")
	log.Printf("DEBUG - ResetUserPassword: Kaydedilen şifre bcrypt formatında: %v",
		strings.HasPrefix(updatedUser.Password, "$2a$") ||
			strings.HasPrefix(updatedUser.Password, "$2b$") ||
			strings.HasPrefix(updatedUser.Password, "$2y$"))
	log.Printf("DEBUG - ResetUserPassword: Kullanıcı aktif mi: %v", updatedUser.IsActive)

	// Şifreleri karşılaştır
	if updatedUser.Password != hashedPasswordStr {
		log.Printf("WARN - ResetUserPassword: Kaydedilen şifre beklenen şifre ile eşleşmiyor")
		log.Printf("DEBUG - ResetUserPassword: Şifre karşılaştırma hatası")
		log.Printf("DEBUG - ResetUserPassword: Hash değerleri eşleşmiyor")
	}

	// Şifre doğrulama testi yap
	if !strings.HasPrefix(newPassword, "$2a$") && !strings.HasPrefix(newPassword, "$2b$") && !strings.HasPrefix(newPassword, "$2y$") {
		err = bcrypt.CompareHashAndPassword([]byte(updatedUser.Password), []byte(newPassword))
		if err != nil {
			log.Printf("ERROR - ResetUserPassword: Şifre doğrulama testi başarısız: %v", err)
		} else {
			log.Printf("SUCCESS - ResetUserPassword: Şifre doğrulama testi başarılı")
		}
	}

	log.Printf("SUCCESS - ResetUserPassword: Şifre başarıyla değiştirildi ve hesap aktifleştirildi: %s", email)
	return nil
}

// GetUserByEmail gets a user by email
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	conditions := map[string]interface{}{
		"email": email,
	}

	err := r.baseRepo.FindOne("users", &user, conditions)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, core.NewDatabaseError("error getting user", err)
	}

	// Şifre boş mu kontrol et ve log çıktısı ver
	log.Printf("DEBUG - GetUserByEmail: Kullanıcı bulundu: %s", email)
	if user.Password == "" {
		log.Printf("WARN - GetUserByEmail: Kullanıcının şifresi boş: %s", email)

		// Şifre boş ise doğrudan SQL sorgusu ile tekrar deneyelim
		var password string
		err := database.DB.QueryRow("SELECT password FROM users WHERE email = ?", email).Scan(&password)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("ERROR - GetUserByEmail: Doğrudan sorgu ile de kullanıcı bulunamadı: %s", email)
				return &user, nil
			}
			log.Printf("ERROR - GetUserByEmail: Doğrudan şifre sorgusu hatası: %v", err)
			return &user, nil
		}

		if password != "" {
			log.Printf("INFO - GetUserByEmail: Doğrudan sorgu ile şifre alındı")
			user.Password = password
		}
	}

	return &user, nil
}

// UpdateResetToken fonksiyonu kaldırıldı - token sistemi artık kullanılmıyor

// Create inserts a new user
func (r *UserRepository) Create(user *models.User) (int64, error) {
	data := map[string]interface{}{
		"name":       user.Name,
		"email":      user.Email,
		"password":   user.Password,
		"phone":      user.Phone,
		"is_active":  user.IsActive,
		"is_admin":   user.IsAdmin,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	id, err := r.baseRepo.Create("users", data)
	if err != nil {
		return 0, core.NewDatabaseError("error creating user", err)
	}
	return id, nil
}

// Update modifies an existing user
func (r *UserRepository) Update(user *models.User) error {
	data := map[string]interface{}{
		"name":       user.Name,
		"email":      user.Email,
		"password":   user.Password,
		"phone":      user.Phone,
		"is_active":  user.IsActive,
		"is_admin":   user.IsAdmin,
		"updated_at": time.Now(),
	}

	err := r.baseRepo.Update("users", user.ID, data)
	if err != nil {
		return core.NewDatabaseError("error updating user", err)
	}
	return nil
}

// Delete removes a user
func (r *UserRepository) Delete(id int64) error {
	err := r.baseRepo.Delete("users", id)
	if err != nil {
		return core.NewDatabaseError("error deleting user", err)
	}
	return nil
}

// FindByID retrieves a user by ID
func (r *UserRepository) FindByID(id int64) (*models.User, error) {
	var user models.User
	err := r.baseRepo.FindByID("users", id, &user)
	if err != nil {
		return nil, core.NewDatabaseError("error finding user by ID", err)
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User

	// Doğrudan SQL sorgusu ile kullanıcıyı bul
	query := `SELECT id, name, email, password, phone, is_active, is_admin, created_at, updated_at 
			  FROM users WHERE email = ?`

	err := database.DB.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.Phone,
		&user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("DEBUG - FindByEmail: Kullanıcı bulunamadı: %s", email)
			return nil, nil
		}
		log.Printf("ERROR - FindByEmail: Veritabanı hatası: %v", err)
		return nil, core.NewDatabaseError("error finding user by email", err)
	}

	log.Printf("DEBUG - FindByEmail: Kullanıcı bulundu: %s", email)

	// Şifre boş mu kontrol et
	if user.Password == "" {
		log.Printf("WARN - FindByEmail: Kullanıcının şifresi boş: %s", email)

		// Şifre boş ise doğrudan SQL sorgusu ile tekrar deneyelim
		var password string
		err := database.DB.QueryRow("SELECT password FROM users WHERE email = ?", email).Scan(&password)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("ERROR - FindByEmail: Doğrudan sorgu ile de kullanıcı bulunamadı: %s", email)
				return &user, nil
			}
			log.Printf("ERROR - FindByEmail: Doğrudan şifre sorgusu hatası: %v", err)
			return &user, nil
		}

		if password != "" {
			log.Printf("INFO - FindByEmail: Doğrudan sorgu ile şifre alındı")
			user.Password = password
		}
	}

	return &user, nil
}

// FindByUsername finds a user by username
func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	conditions := map[string]interface{}{
		"username": username,
	}

	err := r.baseRepo.FindOne("users", &user, conditions)
	if err != nil {
		return nil, core.NewDatabaseError("error finding user by username", err)
	}
	return &user, nil
}

// FindAll retrieves all users with pagination
func (r *UserRepository) FindAll(page, perPage int) ([]*models.User, error) {
	var users []*models.User
	err := r.baseRepo.FindAll("users", &users, nil, "created_at", perPage, (page-1)*perPage)
	if err != nil {
		return nil, core.NewDatabaseError("error finding all users", err)
	}
	return users, nil
}

// FindByRole retrieves users by role with pagination
func (r *UserRepository) FindByRole(role string, page, perPage int) ([]*models.User, error) {
	var users []*models.User
	conditions := map[string]interface{}{
		"role": role,
	}

	err := r.baseRepo.FindAll("users", &users, conditions, "created_at", perPage, (page-1)*perPage)
	if err != nil {
		return nil, core.NewDatabaseError("error finding users by role", err)
	}
	return users, nil
}

// Search searches users
func (r *UserRepository) Search(term string, page, perPage int) ([]*models.User, error) {
	var users []*models.User
	err := r.baseRepo.Search("users", []string{"username", "email", "name"}, term, perPage, (page-1)*perPage, &users)
	if err != nil {
		return nil, core.NewDatabaseError("error searching users", err)
	}
	return users, nil
}

// Count returns the total number of users
func (r *UserRepository) Count() (int64, error) {
	count, err := r.baseRepo.Count("users", nil)
	if err != nil {
		return 0, core.NewDatabaseError("error counting users", err)
	}
	return count, nil
}

// CountByRole returns the number of users with a specific role
func (r *UserRepository) CountByRole(role string) (int64, error) {
	conditions := map[string]interface{}{
		"role": role,
	}

	count, err := r.baseRepo.Count("users", conditions)
	if err != nil {
		return 0, core.NewDatabaseError("error counting users by role", err)
	}
	return count, nil
}

// Transaction executes a function within a transaction
func (r *UserRepository) Transaction(fn func(*sql.Tx) error) error {
	err := r.baseRepo.Transaction(fn)
	if err != nil {
		return core.NewDatabaseError("error in transaction", err)
	}
	return nil
}

// FindByVerificationToken fonksiyonu kaldırıldı - token sistemi artık kullanılmıyor

// ActivateAccount activates a user account
func (r *UserRepository) ActivateAccount(userID int64) error {
	// Doğrudan SQL sorgusu ile hesabı aktifleştir
	updateQuery := "UPDATE users SET is_active = true, updated_at = ? WHERE id = ?"
	stmt, err := database.DB.Prepare(updateQuery)
	if err != nil {
		log.Printf("ERROR - ActivateAccount: SQL hazırlama hatası: %v", err)
		return core.NewDatabaseError("error preparing SQL statement", err)
	}
	defer stmt.Close()

	now := time.Now()
	result, err := stmt.Exec(now, userID)
	if err != nil {
		log.Printf("ERROR - ActivateAccount: SQL çalıştırma hatası: %v", err)
		return core.NewDatabaseError("error executing SQL statement", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("ERROR - ActivateAccount: Etkilenen satır sayısı alınamadı: %v", err)
	} else {
		log.Printf("INFO - ActivateAccount: Etkilenen satır sayısı: %d", rowsAffected)
	}

	// Başarılı olup olmadığını kontrol et
	if rowsAffected == 0 {
		log.Printf("WARN - ActivateAccount: Hiçbir satır etkilenmedi, kullanıcı zaten aktif olabilir: %d", userID)
	} else {
		log.Printf("SUCCESS - ActivateAccount: Kullanıcı hesabı başarıyla aktifleştirildi: %d", userID)
	}

	return nil
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(userID int64, hashedPassword string) error {
	updateData := map[string]interface{}{
		"password":   hashedPassword,
		"updated_at": time.Now(),
	}

	err := r.baseRepo.Update("users", userID, updateData)
	if err != nil {
		return core.NewDatabaseError("error updating password", err)
	}
	return nil
}

// EmailExists checks if an email exists
func (r *UserRepository) EmailExists(email string) (bool, error) {
	conditions := map[string]interface{}{
		"email": email,
	}

	exists, err := r.baseRepo.Exists("users", conditions)
	if err != nil {
		return false, core.NewDatabaseError("error checking if email exists", err)
	}
	return exists, nil
}

// SaveResetToken fonksiyonu kaldırıldı - token sistemi artık kullanılmıyor
