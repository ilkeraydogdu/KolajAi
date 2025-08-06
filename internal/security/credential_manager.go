package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// CredentialManager handles secure credential storage and rotation
type CredentialManager struct {
	encryptionKey []byte
	credentials   map[string]*SecureCredential
	rotationRules map[string]*RotationRule
	mu            sync.RWMutex
	vault         VaultInterface
}

// SecureCredential represents an encrypted credential
type SecureCredential struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	EncryptedData  string                 `json:"encrypted_data"`
	Type           string                 `json:"type"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	LastRotated    time.Time              `json:"last_rotated"`
	RotationPeriod time.Duration          `json:"rotation_period"`
	Version        int                    `json:"version"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// RotationRule defines credential rotation rules
type RotationRule struct {
	CredentialID   string        `json:"credential_id"`
	RotationPeriod time.Duration `json:"rotation_period"`
	AutoRotate     bool          `json:"auto_rotate"`
	NotifyBefore   time.Duration `json:"notify_before"`
	LastCheck      time.Time     `json:"last_check"`
}

// VaultInterface defines the interface for external vault systems
type VaultInterface interface {
	Store(key string, value []byte) error
	Retrieve(key string) ([]byte, error)
	Delete(key string) error
	Rotate(key string) error
}

// NewCredentialManager creates a new credential manager
func NewCredentialManager(encryptionKey string, vault VaultInterface) (*CredentialManager, error) {
	if len(encryptionKey) != 32 {
		return nil, errors.New("encryption key must be 32 bytes")
	}

	return &CredentialManager{
		encryptionKey: []byte(encryptionKey),
		credentials:   make(map[string]*SecureCredential),
		rotationRules: make(map[string]*RotationRule),
		vault:         vault,
	}, nil
}

// StoreCredential securely stores a credential
func (cm *CredentialManager) StoreCredential(id, name string, data map[string]interface{}, credType string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal credential data: %w", err)
	}

	// Encrypt the data
	encryptedData, err := cm.encrypt(jsonData)
	if err != nil {
		return fmt.Errorf("failed to encrypt credential: %w", err)
	}

	// Create secure credential
	credential := &SecureCredential{
		ID:            id,
		Name:          name,
		EncryptedData: base64.StdEncoding.EncodeToString(encryptedData),
		Type:          credType,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		LastRotated:   time.Now(),
		Version:       1,
		Metadata:      make(map[string]interface{}),
	}

	// Store in memory
	cm.credentials[id] = credential

	// Store in vault if available
	if cm.vault != nil {
		if err := cm.vault.Store(id, encryptedData); err != nil {
			return fmt.Errorf("failed to store in vault: %w", err)
		}
	}

	return nil
}

// RetrieveCredential retrieves and decrypts a credential
func (cm *CredentialManager) RetrieveCredential(id string) (map[string]interface{}, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	credential, exists := cm.credentials[id]
	if !exists {
		// Try to retrieve from vault
		if cm.vault != nil {
			encryptedData, err := cm.vault.Retrieve(id)
			if err != nil {
				return nil, fmt.Errorf("credential not found: %s", id)
			}

			// Decrypt the data
			decryptedData, err := cm.decrypt(encryptedData)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt credential: %w", err)
			}

			var data map[string]interface{}
			if err := json.Unmarshal(decryptedData, &data); err != nil {
				return nil, fmt.Errorf("failed to unmarshal credential data: %w", err)
			}

			return data, nil
		}
		return nil, fmt.Errorf("credential not found: %s", id)
	}

	// Decode from base64
	encryptedData, err := base64.StdEncoding.DecodeString(credential.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode credential: %w", err)
	}

	// Decrypt the data
	decryptedData, err := cm.decrypt(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(decryptedData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credential data: %w", err)
	}

	return data, nil
}

// RotateCredential rotates a credential
func (cm *CredentialManager) RotateCredential(id string, newData map[string]interface{}) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	credential, exists := cm.credentials[id]
	if !exists {
		return fmt.Errorf("credential not found: %s", id)
	}

	// Convert new data to JSON
	jsonData, err := json.Marshal(newData)
	if err != nil {
		return fmt.Errorf("failed to marshal new credential data: %w", err)
	}

	// Encrypt the new data
	encryptedData, err := cm.encrypt(jsonData)
	if err != nil {
		return fmt.Errorf("failed to encrypt new credential: %w", err)
	}

	// Update credential
	credential.EncryptedData = base64.StdEncoding.EncodeToString(encryptedData)
	credential.UpdatedAt = time.Now()
	credential.LastRotated = time.Now()
	credential.Version++

	// Update in vault if available
	if cm.vault != nil {
		if err := cm.vault.Store(id, encryptedData); err != nil {
			return fmt.Errorf("failed to update in vault: %w", err)
		}
	}

	return nil
}

// SetRotationRule sets a rotation rule for a credential
func (cm *CredentialManager) SetRotationRule(credentialID string, period time.Duration, autoRotate bool) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.credentials[credentialID]; !exists {
		return fmt.Errorf("credential not found: %s", credentialID)
	}

	cm.rotationRules[credentialID] = &RotationRule{
		CredentialID:   credentialID,
		RotationPeriod: period,
		AutoRotate:     autoRotate,
		NotifyBefore:   24 * time.Hour,
		LastCheck:      time.Now(),
	}

	return nil
}

// CheckRotationRequired checks if any credentials need rotation
func (cm *CredentialManager) CheckRotationRequired() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var needsRotation []string

	for id, credential := range cm.credentials {
		if rule, exists := cm.rotationRules[id]; exists {
			if time.Since(credential.LastRotated) >= rule.RotationPeriod {
				needsRotation = append(needsRotation, id)
			}
		}
	}

	return needsRotation
}

// encrypt encrypts data using AES-GCM
func (cm *CredentialManager) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(cm.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-GCM
func (cm *CredentialManager) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(cm.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GenerateSecureKey generates a secure encryption key
func GenerateSecureKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// HashPassword hashes a password using SHA256
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// StartRotationMonitor starts a background goroutine to monitor credential rotation
func (cm *CredentialManager) StartRotationMonitor(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			needsRotation := cm.CheckRotationRequired()
			for _, id := range needsRotation {
				if rule, exists := cm.rotationRules[id]; exists && rule.AutoRotate {
					// In a real implementation, this would trigger automatic rotation
					// For now, we'll just log it
					fmt.Printf("Credential %s needs rotation\n", id)
				}
			}
		}
	}()
}
