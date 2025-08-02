package security

import (
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"errors"
	"fmt"
	"image/png"
	"io"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"kolajAi/internal/models"
)

// TwoFAService handles two-factor authentication operations
type TwoFAService struct {
	db     *sql.DB
	issuer string
}

// TwoFASetup represents 2FA setup information
type TwoFASetup struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
	ManualEntry string   `json:"manual_entry"`
}

// TwoFAValidation represents 2FA validation result
type TwoFAValidation struct {
	Valid       bool   `json:"valid"`
	Used        bool   `json:"used"`
	Error       string `json:"error,omitempty"`
	BackupUsed  bool   `json:"backup_used"`
	TimeWindow  int    `json:"time_window,omitempty"`
}

// BackupCode represents a 2FA backup code
type BackupCode struct {
	ID        int       `json:"id"`
	UserID    int64     `json:"user_id"`
	Code      string    `json:"code"`
	Used      bool      `json:"used"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// NewTwoFAService creates a new 2FA service
func NewTwoFAService(db *sql.DB, issuer string) *TwoFAService {
	return &TwoFAService{
		db:     db,
		issuer: issuer,
	}
}

// GenerateSecret generates a new TOTP secret for user
func (t *TwoFAService) GenerateSecret(user *models.User) (*TwoFASetup, error) {
	if user == nil {
		return nil, errors.New("user cannot be nil")
	}

	// Generate secret key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      t.issuer,
		AccountName: user.Email,
		SecretSize:  32,
		Algorithm:   otp.AlgorithmSHA1,
		Period:      30,
		Digits:      6,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	// Generate backup codes
	backupCodes, err := t.generateBackupCodes(8)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Create manual entry string (formatted secret)
	manualEntry := t.formatSecretForManualEntry(key.Secret())

	return &TwoFASetup{
		Secret:      key.Secret(),
		QRCodeURL:   key.URL(),
		BackupCodes: backupCodes,
		ManualEntry: manualEntry,
	}, nil
}

// EnableTwoFA enables 2FA for user after verification
func (t *TwoFAService) EnableTwoFA(userID int64, secret string, verificationCode string, backupCodes []string) error {
	// Verify the code first
	if !t.validateTOTPCode(secret, verificationCode) {
		return errors.New("invalid verification code")
	}

	// Start transaction
	tx, err := t.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Update user's 2FA settings
	_, err = tx.Exec(`
		UPDATE users SET 
			two_factor_enabled = 1,
			two_factor_secret = ?,
			two_factor_enabled_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, secret, userID)
	if err != nil {
		return fmt.Errorf("failed to enable 2FA for user: %w", err)
	}

	// Clear existing backup codes
	_, err = tx.Exec("DELETE FROM two_factor_backup_codes WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("failed to clear old backup codes: %w", err)
	}

	// Insert new backup codes
	for _, code := range backupCodes {
		_, err = tx.Exec(`
			INSERT INTO two_factor_backup_codes (user_id, code, created_at)
			VALUES (?, ?, CURRENT_TIMESTAMP)
		`, userID, code)
		if err != nil {
			return fmt.Errorf("failed to save backup code: %w", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DisableTwoFA disables 2FA for user
func (t *TwoFAService) DisableTwoFA(userID int64, password string) error {
	// TODO: Verify user password before disabling 2FA
	
	// Start transaction
	tx, err := t.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Update user's 2FA settings
	_, err = tx.Exec(`
		UPDATE users SET 
			two_factor_enabled = 0,
			two_factor_secret = NULL,
			two_factor_enabled_at = NULL,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to disable 2FA for user: %w", err)
	}

	// Remove backup codes
	_, err = tx.Exec("DELETE FROM two_factor_backup_codes WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("failed to remove backup codes: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ValidateCode validates TOTP code or backup code
func (t *TwoFAService) ValidateCode(userID int64, code string) *TwoFAValidation {
	// Get user's 2FA settings
	var secret string
	var enabled bool
	err := t.db.QueryRow(`
		SELECT two_factor_secret, two_factor_enabled 
		FROM users 
		WHERE id = ? AND two_factor_enabled = 1
	`, userID).Scan(&secret, &enabled)
	
	if err != nil {
		return &TwoFAValidation{
			Valid: false,
			Error: "2FA not enabled for user",
		}
	}

	// Clean the code (remove spaces, dashes)
	cleanCode := t.cleanCode(code)

	// First, try TOTP validation
	if t.validateTOTPCode(secret, cleanCode) {
		return &TwoFAValidation{
			Valid: true,
		}
	}

	// If TOTP fails, try backup codes
	if t.validateAndUseBackupCode(userID, cleanCode) {
		return &TwoFAValidation{
			Valid:      true,
			BackupUsed: true,
		}
	}

	return &TwoFAValidation{
		Valid: false,
		Error: "invalid 2FA code",
	}
}

// GenerateQRCode generates QR code image for 2FA setup
func (t *TwoFAService) GenerateQRCode(secret string, user *models.User, writer io.Writer) error {
	key, err := otp.NewKeyFromURL(fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s",
		t.issuer, user.Email, secret, t.issuer,
	))
	if err != nil {
		return fmt.Errorf("failed to create OTP key: %w", err)
	}

	img, err := key.Image(256, 256)
	if err != nil {
		return fmt.Errorf("failed to generate QR code image: %w", err)
	}

	return png.Encode(writer, img)
}

// GetBackupCodes retrieves unused backup codes for user
func (t *TwoFAService) GetBackupCodes(userID int64) ([]string, error) {
	rows, err := t.db.Query(`
		SELECT code FROM two_factor_backup_codes 
		WHERE user_id = ? AND used = 0 
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query backup codes: %w", err)
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("failed to scan backup code: %w", err)
		}
		codes = append(codes, code)
	}

	return codes, nil
}

// RegenerateBackupCodes generates new backup codes for user
func (t *TwoFAService) RegenerateBackupCodes(userID int64) ([]string, error) {
	// Generate new codes
	newCodes, err := t.generateBackupCodes(8)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Start transaction
	tx, err := t.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Remove old codes
	_, err = tx.Exec("DELETE FROM two_factor_backup_codes WHERE user_id = ?", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to remove old backup codes: %w", err)
	}

	// Insert new codes
	for _, code := range newCodes {
		_, err = tx.Exec(`
			INSERT INTO two_factor_backup_codes (user_id, code, created_at)
			VALUES (?, ?, CURRENT_TIMESTAMP)
		`, userID, code)
		if err != nil {
			return nil, fmt.Errorf("failed to save backup code: %w", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newCodes, nil
}

// IsTwoFAEnabled checks if 2FA is enabled for user
func (t *TwoFAService) IsTwoFAEnabled(userID int64) (bool, error) {
	var enabled bool
	err := t.db.QueryRow(`
		SELECT two_factor_enabled FROM users WHERE id = ?
	`, userID).Scan(&enabled)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check 2FA status: %w", err)
	}

	return enabled, nil
}

// GetBackupCodeCount returns count of unused backup codes
func (t *TwoFAService) GetBackupCodeCount(userID int64) (int, error) {
	var count int
	err := t.db.QueryRow(`
		SELECT COUNT(*) FROM two_factor_backup_codes 
		WHERE user_id = ? AND used = 0
	`, userID).Scan(&count)
	
	if err != nil {
		return 0, fmt.Errorf("failed to count backup codes: %w", err)
	}

	return count, nil
}

// Helper methods

func (t *TwoFAService) validateTOTPCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

func (t *TwoFAService) validateAndUseBackupCode(userID int64, code string) bool {
	// Check if backup code exists and is unused
	var codeID int
	err := t.db.QueryRow(`
		SELECT id FROM two_factor_backup_codes 
		WHERE user_id = ? AND code = ? AND used = 0
	`, userID, code).Scan(&codeID)
	
	if err != nil {
		return false
	}

	// Mark code as used
	_, err = t.db.Exec(`
		UPDATE two_factor_backup_codes 
		SET used = 1, used_at = CURRENT_TIMESTAMP 
		WHERE id = ?
	`, codeID)
	
	return err == nil
}

func (t *TwoFAService) generateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)
	
	for i := 0; i < count; i++ {
		code, err := t.generateSingleBackupCode()
		if err != nil {
			return nil, err
		}
		codes[i] = code
	}
	
	return codes, nil
}

func (t *TwoFAService) generateSingleBackupCode() (string, error) {
	// Generate 8 random bytes
	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	
	// Convert to base32 and format
	code := base32.StdEncoding.EncodeToString(bytes)
	code = strings.TrimRight(code, "=") // Remove padding
	
	// Format as XXXX-XXXX
	if len(code) >= 8 {
		return fmt.Sprintf("%s-%s", code[:4], code[4:8]), nil
	}
	
	return code, nil
}

func (t *TwoFAService) cleanCode(code string) string {
	// Remove spaces, dashes, and convert to uppercase
	cleaned := strings.ReplaceAll(code, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	return strings.ToUpper(cleaned)
}

func (t *TwoFAService) formatSecretForManualEntry(secret string) string {
	// Format secret in groups of 4 characters for easier manual entry
	var formatted strings.Builder
	for i, char := range secret {
		if i > 0 && i%4 == 0 {
			formatted.WriteString(" ")
		}
		formatted.WriteRune(char)
	}
	return formatted.String()
}