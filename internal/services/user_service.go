package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"kolajAi/internal/models"
	"kolajAi/internal/repository"
)

// UserService handles user-related operations
type UserService struct {
	repo     *repository.BaseRepository
	db       *sql.DB
	authSvc  *AuthService
}

// NewUserService creates a new user service
func NewUserService(repo *repository.BaseRepository, db *sql.DB, authSvc *AuthService) *UserService {
	return &UserService{
		repo:    repo,
		db:      db,
		authSvc: authSvc,
	}
}

// UserRegistrationRequest represents user registration data
type UserRegistrationRequest struct {
	Name            string `json:"name" validate:"required,min=2,max=100"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
	Phone           string `json:"phone" validate:"omitempty,e164"`
	AcceptTerms     bool   `json:"accept_terms" validate:"required"`
	NewsletterOptIn bool   `json:"newsletter_opt_in"`
}

// UserUpdateRequest represents user update data
type UserUpdateRequest struct {
	Name  string `json:"name" validate:"omitempty,min=2,max=100"`
	Phone string `json:"phone" validate:"omitempty,e164"`
}

// PasswordChangeRequest represents password change data
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

// UserProfile represents extended user profile information
type UserProfile struct {
	User     *models.User     `json:"user"`
	Customer *models.Customer `json:"customer,omitempty"`
	Stats    *UserStats       `json:"stats"`
}

// UserStats represents user statistics
type UserStats struct {
	TotalOrders    int     `json:"total_orders"`
	TotalSpent     float64 `json:"total_spent"`
	ReviewsCount   int     `json:"reviews_count"`
	WishlistCount  int     `json:"wishlist_count"`
	LoyaltyPoints  int     `json:"loyalty_points"`
	MemberSince    time.Time `json:"member_since"`
	LastActivity   time.Time `json:"last_activity"`
}

// RegisterUser creates a new user account
func (s *UserService) RegisterUser(req *UserRegistrationRequest) (*models.User, error) {
	// Validate input
	if err := s.validateRegistrationRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if user already exists
	existingUser, _ := s.GetUserByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Skip verification token for simplicity
	// verificationToken, err := s.generateToken()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to generate verification token: %w", err)
	// }

	// Create user
	user := &models.User{
		Name:      req.Name,
		Email:     strings.ToLower(strings.TrimSpace(req.Email)),
		Password:  hashedPassword,
		Phone:     req.Phone,
		IsActive:  true, // Active by default for simplicity
		IsAdmin:   false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	id, err := s.repo.Create("users", user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	user.ID = id

	// Create customer profile
	if err := s.createCustomerProfile(user, req.NewsletterOptIn); err != nil {
		// Log error but don't fail registration
		fmt.Printf("Warning: Failed to create customer profile for user %d: %v\n", user.ID, err)
	}

	return user, nil
}

// VerifyEmail verifies user email with token
func (s *UserService) VerifyEmail(token string) error {
	if token == "" {
		return errors.New("verification token is required")
	}

	// Find user by verification token
	query := "SELECT id, verification_token, token_expires_at FROM users WHERE verification_token = ? AND is_active = 0"
	var userID int
	var dbToken string
	var expiresAt time.Time

	err := s.db.QueryRow(query, token).Scan(&userID, &dbToken, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("invalid or expired verification token")
		}
		return fmt.Errorf("failed to verify token: %w", err)
	}

	// Check if token is expired
	if time.Now().After(expiresAt) {
		return errors.New("verification token has expired")
	}

	// Activate user
	updateQuery := `UPDATE users SET 
		is_active = 1, 
		verification_token = NULL, 
		token_expires_at = NULL, 
		updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`

	_, err = s.db.Exec(updateQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	return nil
}

// AuthenticateUser authenticates user with email and password
func (s *UserService) AuthenticateUser(email, password string) (*models.User, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	user, err := s.GetUserByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, errors.New("account is not activated")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Update last login
	s.updateLastLogin(int(user.ID))

	return user, nil
}

// GetUserByID retrieves user by ID
func (s *UserService) GetUserByID(id int) (*models.User, error) {
	if id <= 0 {
		return nil, errors.New("invalid user ID")
	}

	var user models.User
	err := s.repo.FindByID("users", id, &user)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves user by email
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}

	email = strings.ToLower(strings.TrimSpace(email))
	
	query := "SELECT id, name, email, password, phone, is_active, is_admin, created_at, updated_at FROM users WHERE email = ?"
	
	var user models.User
	err := s.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.Phone,
		&user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(userID int, req *UserUpdateRequest) (*models.User, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	// Get existing user
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	user.UpdatedAt = time.Now()

	// Save to database
	err = s.repo.Update("users", userID, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// ChangePassword changes user password
func (s *UserService) ChangePassword(userID int, req *PasswordChangeRequest) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}

	// Get user
	user, err := s.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	// Validate new password
	if err := s.validatePassword(req.NewPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := s.hashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	query := "UPDATE users SET password = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err = s.db.Exec(query, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// RequestPasswordReset generates password reset token
func (s *UserService) RequestPasswordReset(email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	user, err := s.GetUserByEmail(email)
	if err != nil {
		// Don't reveal if user exists
		return nil
	}

	// Generate reset token
	resetToken, err := s.generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Save reset token
	query := `UPDATE users SET 
		reset_token = ?, 
		token_expires_at = ?, 
		updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`

	expiresAt := time.Now().Add(1 * time.Hour)
	_, err = s.db.Exec(query, resetToken, expiresAt, user.ID)
	if err != nil {
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	// TODO: Send password reset email
	fmt.Printf("Password reset token for %s: %s\n", email, resetToken)

	return nil
}

// ResetPassword resets password with token
func (s *UserService) ResetPassword(token, newPassword string) error {
	if token == "" || newPassword == "" {
		return errors.New("token and new password are required")
	}

	// Validate new password
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// Find user by reset token
	query := "SELECT id, reset_token, token_expires_at FROM users WHERE reset_token = ?"
	var userID int
	var dbToken string
	var expiresAt time.Time

	err := s.db.QueryRow(query, token).Scan(&userID, &dbToken, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("invalid or expired reset token")
		}
		return fmt.Errorf("failed to verify token: %w", err)
	}

	// Check if token is expired
	if time.Now().After(expiresAt) {
		return errors.New("reset token has expired")
	}

	// Hash new password
	hashedPassword, err := s.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password and clear reset token
	updateQuery := `UPDATE users SET 
		password = ?, 
		reset_token = NULL, 
		token_expires_at = NULL, 
		updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`

	_, err = s.db.Exec(updateQuery, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	return nil
}

// GetUserProfile retrieves complete user profile
func (s *UserService) GetUserProfile(userID int) (*UserProfile, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	// Get user
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Get customer profile
	customer, _ := s.getCustomerByUserID(userID)

	// Get user stats
	stats, err := s.getUserStats(userID)
	if err != nil {
		// Log error but don't fail
		fmt.Printf("Warning: Failed to get user stats for user %d: %v\n", userID, err)
		stats = &UserStats{
			MemberSince: user.CreatedAt,
		}
	}

	return &UserProfile{
		User:     user,
		Customer: customer,
		Stats:    stats,
	}, nil
}

// DeactivateUser deactivates user account
func (s *UserService) DeactivateUser(userID int) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}

	query := "UPDATE users SET is_active = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := s.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	return nil
}

// DeleteUser soft deletes user account
func (s *UserService) DeleteUser(userID int) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}

	// Soft delete by anonymizing data
	anonymizedEmail := fmt.Sprintf("deleted_%d@deleted.com", userID)
	query := `UPDATE users SET 
		name = 'Deleted User', 
		email = ?, 
		phone = NULL, 
		is_active = 0, 
		verification_token = NULL, 
		reset_token = NULL, 
		updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`

	_, err := s.db.Exec(query, anonymizedEmail, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// Helper methods

func (s *UserService) validateRegistrationRequest(req *UserRegistrationRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	if len(req.Name) < 2 || len(req.Name) > 100 {
		return errors.New("name must be between 2 and 100 characters")
	}

	if req.Email == "" {
		return errors.New("email is required")
	}
	if !s.isValidEmail(req.Email) {
		return errors.New("invalid email format")
	}

	if err := s.validatePassword(req.Password); err != nil {
		return err
	}

	if req.Password != req.ConfirmPassword {
		return errors.New("passwords do not match")
	}

	if !req.AcceptTerms {
		return errors.New("you must accept the terms and conditions")
	}

	return nil
}

func (s *UserService) validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	hasSymbol := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\?]`).MatchString(password)

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return errors.New("password must contain at least one number")
	}
	if !hasSymbol {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

func (s *UserService) isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func (s *UserService) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (s *UserService) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *UserService) updateLastLogin(userID int) {
	query := "UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	s.db.Exec(query, userID)
}

func (s *UserService) createCustomerProfile(user *models.User, newsletterOptIn bool) error {
	customer := &models.Customer{
		UserID:           uint(user.ID),
		FirstName:        user.Name,
		LastName:         "",
		Language:         "tr",
		Currency:         "TRY",
		Newsletter:       newsletterOptIn,
		SMSNotifications: false,
		Status:           models.CustomerStatusActive,
		Tier:             models.CustomerTierBronze,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	_, err := s.repo.Create("customers", customer)
	return err
}

func (s *UserService) getCustomerByUserID(userID int) (*models.Customer, error) {
	query := "SELECT id, user_id, first_name, last_name, phone, language, currency, newsletter, sms_notifications, status, tier, total_orders, total_spent, created_at, updated_at FROM customers WHERE user_id = ?"
	
	var customer models.Customer
	err := s.db.QueryRow(query, userID).Scan(
		&customer.ID, &customer.UserID, &customer.FirstName, &customer.LastName,
		&customer.Phone, &customer.Language, &customer.Currency, &customer.Newsletter,
		&customer.SMSNotifications, &customer.Status, &customer.Tier,
		&customer.TotalOrders, &customer.TotalSpent, &customer.CreatedAt, &customer.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}

	return &customer, nil
}

func (s *UserService) getUserStats(userID int) (*UserStats, error) {
	// Get user creation date
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	stats := &UserStats{
		MemberSince:  user.CreatedAt,
		LastActivity: user.UpdatedAt,
	}

	// Get customer stats if exists
	customer, err := s.getCustomerByUserID(userID)
	if err == nil {
		stats.TotalOrders = customer.TotalOrders
		stats.TotalSpent = customer.TotalSpent
	}

	// Get reviews count
	reviewsQuery := "SELECT COUNT(*) FROM reviews WHERE customer_id = (SELECT id FROM customers WHERE user_id = ?)"
	s.db.QueryRow(reviewsQuery, userID).Scan(&stats.ReviewsCount)

	// TODO: Get wishlist count, loyalty points when those systems are implemented

	return stats, nil
}