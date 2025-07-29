package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

// CredentialManager manages secure storage and retrieval of credentials
type CredentialManager struct {
	masterKey    []byte
	credentials  map[string]*SecureCredential
	gcm          cipher.AEAD
	mutex        sync.RWMutex
	rotationTime time.Duration
}

// SecureCredential represents an encrypted credential
type SecureCredential struct {
	ID            string            `json:"id"`
	ProviderType  string            `json:"provider_type"`
	ProviderName  string            `json:"provider_name"`
	EncryptedData []byte            `json:"encrypted_data"`
	Nonce         []byte            `json:"nonce"`
	Salt          []byte            `json:"salt"`
	KeyVersion    int               `json:"key_version"`
	CreatedAt     time.Time         `json:"created_at"`
	LastRotated   time.Time         `json:"last_rotated"`
	ExpiresAt     *time.Time        `json:"expires_at,omitempty"`
	Metadata      map[string]string `json:"metadata"`
}

// CredentialData represents the actual credential data
type CredentialData struct {
	APIKey        string            `json:"api_key,omitempty"`
	APISecret     string            `json:"api_secret,omitempty"`
	AccessToken   string            `json:"access_token,omitempty"`
	RefreshToken  string            `json:"refresh_token,omitempty"`
	Username      string            `json:"username,omitempty"`
	Password      string            `json:"password,omitempty"`
	ClientID      string            `json:"client_id,omitempty"`
	ClientSecret  string            `json:"client_secret,omitempty"`
	Additional    map[string]string `json:"additional,omitempty"`
	Environment   string            `json:"environment"`
	Region        string            `json:"region,omitempty"`
}

// NewCredentialManager creates a new secure credential manager
func NewCredentialManager(masterPassword string) (*CredentialManager, error) {
	// Generate master key from password using PBKDF2
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	masterKey := pbkdf2.Key([]byte(masterPassword), salt, 100000, 32, sha256.New)

	// Create AES-GCM cipher
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &CredentialManager{
		masterKey:    masterKey,
		credentials:  make(map[string]*SecureCredential),
		gcm:          gcm,
		rotationTime: 30 * 24 * time.Hour, // 30 days default rotation
	}, nil
}

// StoreCredential securely stores a credential
func (cm *CredentialManager) StoreCredential(id, providerType, providerName string, data *CredentialData) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Validate input
	if err := cm.validateCredentialData(data); err != nil {
		return fmt.Errorf("credential validation failed: %w", err)
	}

	// Marshal credential data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal credential data: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, cm.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data
	encryptedData := cm.gcm.Seal(nil, nonce, jsonData, nil)

	// Generate salt for this credential
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	now := time.Now()
	credential := &SecureCredential{
		ID:            id,
		ProviderType:  providerType,
		ProviderName:  providerName,
		EncryptedData: encryptedData,
		Nonce:         nonce,
		Salt:          salt,
		KeyVersion:    1,
		CreatedAt:     now,
		LastRotated:   now,
		Metadata:      make(map[string]string),
	}

	// Set expiration if specified
	if data.Environment == "production" {
		expiresAt := now.Add(cm.rotationTime)
		credential.ExpiresAt = &expiresAt
	}

	cm.credentials[id] = credential
	return nil
}

// GetCredential retrieves and decrypts a credential
func (cm *CredentialManager) GetCredential(id string) (*CredentialData, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	credential, exists := cm.credentials[id]
	if !exists {
		return nil, fmt.Errorf("credential not found: %s", id)
	}

	// Check if credential is expired
	if credential.ExpiresAt != nil && time.Now().After(*credential.ExpiresAt) {
		return nil, fmt.Errorf("credential expired: %s", id)
	}

	// Decrypt data
	plaintext, err := cm.gcm.Open(nil, credential.Nonce, credential.EncryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Unmarshal credential data
	var data CredentialData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credential data: %w", err)
	}

	return &data, nil
}

// RotateCredential rotates a credential with new data
func (cm *CredentialManager) RotateCredential(id string, newData *CredentialData) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	credential, exists := cm.credentials[id]
	if !exists {
		return fmt.Errorf("credential not found: %s", id)
	}

	// Validate new credential data
	if err := cm.validateCredentialData(newData); err != nil {
		return fmt.Errorf("new credential validation failed: %w", err)
	}

	// Marshal new credential data
	jsonData, err := json.Marshal(newData)
	if err != nil {
		return fmt.Errorf("failed to marshal new credential data: %w", err)
	}

	// Generate new nonce
	nonce := make([]byte, cm.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt new data
	encryptedData := cm.gcm.Seal(nil, nonce, jsonData, nil)

	// Update credential
	now := time.Now()
	credential.EncryptedData = encryptedData
	credential.Nonce = nonce
	credential.KeyVersion++
	credential.LastRotated = now

	// Update expiration
	if newData.Environment == "production" {
		expiresAt := now.Add(cm.rotationTime)
		credential.ExpiresAt = &expiresAt
	}

	return nil
}

// DeleteCredential securely deletes a credential
func (cm *CredentialManager) DeleteCredential(id string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if _, exists := cm.credentials[id]; !exists {
		return fmt.Errorf("credential not found: %s", id)
	}

	// Securely wipe the credential data
	credential := cm.credentials[id]
	for i := range credential.EncryptedData {
		credential.EncryptedData[i] = 0
	}
	for i := range credential.Nonce {
		credential.Nonce[i] = 0
	}
	for i := range credential.Salt {
		credential.Salt[i] = 0
	}

	delete(cm.credentials, id)
	return nil
}

// ListCredentials returns a list of credential metadata (without sensitive data)
func (cm *CredentialManager) ListCredentials() []CredentialMetadata {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	metadata := make([]CredentialMetadata, 0, len(cm.credentials))
	for _, cred := range cm.credentials {
		meta := CredentialMetadata{
			ID:           cred.ID,
			ProviderType: cred.ProviderType,
			ProviderName: cred.ProviderName,
			KeyVersion:   cred.KeyVersion,
			CreatedAt:    cred.CreatedAt,
			LastRotated:  cred.LastRotated,
			ExpiresAt:    cred.ExpiresAt,
			IsExpired:    cred.ExpiresAt != nil && time.Now().After(*cred.ExpiresAt),
		}
		metadata = append(metadata, meta)
	}

	return metadata
}

// CredentialMetadata contains non-sensitive credential information
type CredentialMetadata struct {
	ID           string     `json:"id"`
	ProviderType string     `json:"provider_type"`
	ProviderName string     `json:"provider_name"`
	KeyVersion   int        `json:"key_version"`
	CreatedAt    time.Time  `json:"created_at"`
	LastRotated  time.Time  `json:"last_rotated"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	IsExpired    bool       `json:"is_expired"`
}

// GetExpiredCredentials returns credentials that need rotation
func (cm *CredentialManager) GetExpiredCredentials() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	var expired []string
	now := time.Now()

	for id, cred := range cm.credentials {
		if cred.ExpiresAt != nil && now.After(*cred.ExpiresAt) {
			expired = append(expired, id)
		}
	}

	return expired
}

// ExportCredentials exports encrypted credentials for backup
func (cm *CredentialManager) ExportCredentials() (string, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	export := struct {
		Version     string                        `json:"version"`
		ExportedAt  time.Time                     `json:"exported_at"`
		Credentials map[string]*SecureCredential `json:"credentials"`
	}{
		Version:     "1.0",
		ExportedAt:  time.Now(),
		Credentials: cm.credentials,
	}

	data, err := json.Marshal(export)
	if err != nil {
		return "", fmt.Errorf("failed to marshal export data: %w", err)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// ImportCredentials imports encrypted credentials from backup
func (cm *CredentialManager) ImportCredentials(exportData string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	data, err := base64.StdEncoding.DecodeString(exportData)
	if err != nil {
		return fmt.Errorf("failed to decode export data: %w", err)
	}

	var importData struct {
		Version     string                        `json:"version"`
		ExportedAt  time.Time                     `json:"exported_at"`
		Credentials map[string]*SecureCredential `json:"credentials"`
	}

	if err := json.Unmarshal(data, &importData); err != nil {
		return fmt.Errorf("failed to unmarshal import data: %w", err)
	}

	// Validate version compatibility
	if importData.Version != "1.0" {
		return fmt.Errorf("unsupported export version: %s", importData.Version)
	}

	// Import credentials
	for id, cred := range importData.Credentials {
		cm.credentials[id] = cred
	}

	return nil
}

// validateCredentialData validates credential data
func (cm *CredentialManager) validateCredentialData(data *CredentialData) error {
	if data == nil {
		return fmt.Errorf("credential data cannot be nil")
	}

	// Check that at least one credential field is provided
	hasCredential := data.APIKey != "" || data.APISecret != "" ||
		data.AccessToken != "" || data.RefreshToken != "" ||
		data.Username != "" || data.Password != "" ||
		data.ClientID != "" || data.ClientSecret != ""

	if !hasCredential {
		return fmt.Errorf("at least one credential field must be provided")
	}

	// Validate environment
	validEnvironments := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
		"sandbox":     true,
	}

	if data.Environment == "" {
		data.Environment = "development"
	} else if !validEnvironments[data.Environment] {
		return fmt.Errorf("invalid environment: %s", data.Environment)
	}

	return nil
}

// SetRotationPeriod sets the automatic rotation period for credentials
func (cm *CredentialManager) SetRotationPeriod(duration time.Duration) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.rotationTime = duration
}

// HealthCheck verifies the credential manager is working properly
func (cm *CredentialManager) HealthCheck() error {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// Test encryption/decryption with a dummy credential
	testData := &CredentialData{
		APIKey:      "test_key",
		APISecret:   "test_secret",
		Environment: "development",
	}

	// Create temporary credential
	testID := "health_check_test"
	if err := cm.storeCredentialUnsafe(testID, "test", "test", testData); err != nil {
		return fmt.Errorf("health check failed during store: %w", err)
	}

	// Retrieve and verify
	retrieved, err := cm.getCredentialUnsafe(testID)
	if err != nil {
		return fmt.Errorf("health check failed during retrieve: %w", err)
	}

	if retrieved.APIKey != testData.APIKey || retrieved.APISecret != testData.APISecret {
		return fmt.Errorf("health check failed: data mismatch")
	}

	// Clean up
	delete(cm.credentials, testID)

	return nil
}

// storeCredentialUnsafe stores credential without locking (for internal use)
func (cm *CredentialManager) storeCredentialUnsafe(id, providerType, providerName string, data *CredentialData) error {
	// Marshal credential data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal credential data: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, cm.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data
	encryptedData := cm.gcm.Seal(nil, nonce, jsonData, nil)

	now := time.Now()
	credential := &SecureCredential{
		ID:            id,
		ProviderType:  providerType,
		ProviderName:  providerName,
		EncryptedData: encryptedData,
		Nonce:         nonce,
		KeyVersion:    1,
		CreatedAt:     now,
		LastRotated:   now,
		Metadata:      make(map[string]string),
	}

	cm.credentials[id] = credential
	return nil
}

// getCredentialUnsafe retrieves credential without locking (for internal use)
func (cm *CredentialManager) getCredentialUnsafe(id string) (*CredentialData, error) {
	credential, exists := cm.credentials[id]
	if !exists {
		return nil, fmt.Errorf("credential not found: %s", id)
	}

	// Decrypt data
	plaintext, err := cm.gcm.Open(nil, credential.Nonce, credential.EncryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Unmarshal credential data
	var data CredentialData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credential data: %w", err)
	}

	return &data, nil
}