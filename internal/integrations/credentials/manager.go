package credentials

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
	"kolajAi/internal/integrations"
)

// Manager handles secure storage and retrieval of integration credentials
type Manager struct {
	encryptionKey []byte
	store         Store
	cache         map[string]*cachedCredential
	cacheMutex    sync.RWMutex
	cacheTTL      time.Duration
}

// Store interface for credential storage backend
type Store interface {
	Get(integrationID string) ([]byte, error)
	Set(integrationID string, data []byte) error
	Delete(integrationID string) error
	List() ([]string, error)
}

// cachedCredential holds a credential with expiry time
type cachedCredential struct {
	credential integrations.Credentials
	expiresAt  time.Time
}

// NewManager creates a new credential manager
func NewManager(encryptionKey []byte, store Store) (*Manager, error) {
	if len(encryptionKey) != 32 {
		return nil, errors.New("encryption key must be 32 bytes")
	}
	
	return &Manager{
		encryptionKey: encryptionKey,
		store:         store,
		cache:         make(map[string]*cachedCredential),
		cacheTTL:      5 * time.Minute,
	}, nil
}

// GetCredentials retrieves credentials for an integration
func (m *Manager) GetCredentials(integrationID string) (*integrations.Credentials, error) {
	// Check cache first
	m.cacheMutex.RLock()
	if cached, exists := m.cache[integrationID]; exists && cached.expiresAt.After(time.Now()) {
		m.cacheMutex.RUnlock()
		return &cached.credential, nil
	}
	m.cacheMutex.RUnlock()
	
	// Retrieve from store
	encryptedData, err := m.store.Get(integrationID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credentials: %w", err)
	}
	
	// Decrypt
	decryptedData, err := m.decrypt(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credentials: %w", err)
	}
	
	// Unmarshal
	var creds integrations.Credentials
	if err := json.Unmarshal(decryptedData, &creds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}
	
	// Update cache
	m.cacheMutex.Lock()
	m.cache[integrationID] = &cachedCredential{
		credential: creds,
		expiresAt:  time.Now().Add(m.cacheTTL),
	}
	m.cacheMutex.Unlock()
	
	return &creds, nil
}

// SetCredentials stores credentials for an integration
func (m *Manager) SetCredentials(integrationID string, creds *integrations.Credentials) error {
	// Validate credentials
	if err := m.validateCredentials(creds); err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}
	
	// Marshal
	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}
	
	// Encrypt
	encryptedData, err := m.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt credentials: %w", err)
	}
	
	// Store
	if err := m.store.Set(integrationID, encryptedData); err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}
	
	// Update cache
	m.cacheMutex.Lock()
	m.cache[integrationID] = &cachedCredential{
		credential: *creds,
		expiresAt:  time.Now().Add(m.cacheTTL),
	}
	m.cacheMutex.Unlock()
	
	return nil
}

// DeleteCredentials removes credentials for an integration
func (m *Manager) DeleteCredentials(integrationID string) error {
	// Delete from store
	if err := m.store.Delete(integrationID); err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}
	
	// Remove from cache
	m.cacheMutex.Lock()
	delete(m.cache, integrationID)
	m.cacheMutex.Unlock()
	
	return nil
}

// RotateCredentials updates credentials with new values
func (m *Manager) RotateCredentials(integrationID string, newCreds *integrations.Credentials) error {
	// Get existing credentials for audit
	oldCreds, err := m.GetCredentials(integrationID)
	if err != nil {
		// If credentials don't exist, just set new ones
		return m.SetCredentials(integrationID, newCreds)
	}
	
	// Log rotation (in production, this would go to audit log)
	fmt.Printf("Rotating credentials for integration %s\n", integrationID)
	
	// Set new credentials
	if err := m.SetCredentials(integrationID, newCreds); err != nil {
		return fmt.Errorf("failed to rotate credentials: %w", err)
	}
	
	// In production, you might want to:
	// 1. Keep old credentials for a grace period
	// 2. Send notifications about rotation
	// 3. Update audit logs
	_ = oldCreds // Prevent unused variable warning
	
	return nil
}

// ListIntegrations returns all integration IDs that have stored credentials
func (m *Manager) ListIntegrations() ([]string, error) {
	return m.store.List()
}

// ClearCache removes all cached credentials
func (m *Manager) ClearCache() {
	m.cacheMutex.Lock()
	m.cache = make(map[string]*cachedCredential)
	m.cacheMutex.Unlock()
}

// encrypt encrypts data using AES-GCM
func (m *Manager) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(m.encryptionKey)
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
func (m *Manager) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(m.encryptionKey)
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

// validateCredentials validates that credentials are properly formed
func (m *Manager) validateCredentials(creds *integrations.Credentials) error {
	if creds == nil {
		return errors.New("credentials cannot be nil")
	}
	
	// At least one credential field should be non-empty
	hasCredential := false
	if creds.APIKey != "" || creds.APISecret != "" || 
	   creds.AccessToken != "" || creds.RefreshToken != "" ||
	   len(creds.Extra) > 0 {
		hasCredential = true
	}
	
	if !hasCredential {
		return errors.New("at least one credential field must be provided")
	}
	
	return nil
}

// DatabaseStore implements Store interface using database
type DatabaseStore struct {
	tableName string
	// In real implementation, this would have a database connection
}

// NewDatabaseStore creates a new database-backed credential store
func NewDatabaseStore(tableName string) *DatabaseStore {
	return &DatabaseStore{
		tableName: tableName,
	}
}

// Get retrieves encrypted credentials from database
func (s *DatabaseStore) Get(integrationID string) ([]byte, error) {
	// In real implementation:
	// query := "SELECT encrypted_data FROM " + s.tableName + " WHERE integration_id = ?"
	// return result
	
	// For now, return a placeholder
	return nil, errors.New("not implemented")
}

// Set stores encrypted credentials in database
func (s *DatabaseStore) Set(integrationID string, data []byte) error {
	// In real implementation:
	// query := "INSERT INTO " + s.tableName + " (integration_id, encrypted_data, updated_at) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE encrypted_data = ?, updated_at = ?"
	// execute query
	
	return errors.New("not implemented")
}

// Delete removes credentials from database
func (s *DatabaseStore) Delete(integrationID string) error {
	// In real implementation:
	// query := "DELETE FROM " + s.tableName + " WHERE integration_id = ?"
	// execute query
	
	return errors.New("not implemented")
}

// List returns all integration IDs from database
func (s *DatabaseStore) List() ([]string, error) {
	// In real implementation:
	// query := "SELECT integration_id FROM " + s.tableName
	// return results
	
	return nil, errors.New("not implemented")
}

// MemoryStore implements Store interface using in-memory storage (for testing)
type MemoryStore struct {
	data map[string][]byte
	mu   sync.RWMutex
}

// NewMemoryStore creates a new in-memory credential store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string][]byte),
	}
}

// Get retrieves encrypted credentials from memory
func (s *MemoryStore) Get(integrationID string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	data, exists := s.data[integrationID]
	if !exists {
		return nil, errors.New("credentials not found")
	}
	
	return data, nil
}

// Set stores encrypted credentials in memory
func (s *MemoryStore) Set(integrationID string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.data[integrationID] = data
	return nil
}

// Delete removes credentials from memory
func (s *MemoryStore) Delete(integrationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.data, integrationID)
	return nil
}

// List returns all integration IDs from memory
func (s *MemoryStore) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	ids := make([]string, 0, len(s.data))
	for id := range s.data {
		ids = append(ids, id)
	}
	
	return ids, nil
}

// GenerateEncryptionKey generates a secure random encryption key
func GenerateEncryptionKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

// EncodeKey encodes a key to base64 for storage
func EncodeKey(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}

// DecodeKey decodes a base64 encoded key
func DecodeKey(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}