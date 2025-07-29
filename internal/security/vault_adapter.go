package security

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/vault/api"
)

// HashiCorpVaultAdapter implements VaultInterface for HashiCorp Vault
type HashiCorpVaultAdapter struct {
	client *api.Client
	path   string
}

// NewHashiCorpVaultAdapter creates a new HashiCorp Vault adapter
func NewHashiCorpVaultAdapter(address, token, path string) (*HashiCorpVaultAdapter, error) {
	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	client.SetToken(token)

	return &HashiCorpVaultAdapter{
		client: client,
		path:   path,
	}, nil
}

// Store stores a value in Vault
func (hv *HashiCorpVaultAdapter) Store(key string, value []byte) error {
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"value": value,
		},
	}

	_, err := hv.client.Logical().Write(fmt.Sprintf("%s/data/%s", hv.path, key), data)
	if err != nil {
		return fmt.Errorf("failed to write to vault: %w", err)
	}

	return nil
}

// Retrieve retrieves a value from Vault
func (hv *HashiCorpVaultAdapter) Retrieve(key string) ([]byte, error) {
	secret, err := hv.client.Logical().Read(fmt.Sprintf("%s/data/%s", hv.path, key))
	if err != nil {
		return nil, fmt.Errorf("failed to read from vault: %w", err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	value, ok := data["value"].([]byte)
	if !ok {
		// Try to convert from string
		if strValue, ok := data["value"].(string); ok {
			return []byte(strValue), nil
		}
		return nil, fmt.Errorf("invalid value format")
	}

	return value, nil
}

// Delete deletes a value from Vault
func (hv *HashiCorpVaultAdapter) Delete(key string) error {
	_, err := hv.client.Logical().Delete(fmt.Sprintf("%s/data/%s", hv.path, key))
	if err != nil {
		return fmt.Errorf("failed to delete from vault: %w", err)
	}

	return nil
}

// Rotate rotates a key in Vault
func (hv *HashiCorpVaultAdapter) Rotate(key string) error {
	// For HashiCorp Vault, rotation would typically involve:
	// 1. Creating a new version of the secret
	// 2. Updating the metadata
	// This is a simplified implementation
	
	// Read current value
	currentValue, err := hv.Retrieve(key)
	if err != nil {
		return fmt.Errorf("failed to retrieve current value: %w", err)
	}

	// Store as new version (Vault KV v2 handles versioning automatically)
	return hv.Store(key, currentValue)
}

// LocalVaultAdapter implements VaultInterface for local development
type LocalVaultAdapter struct {
	storage map[string][]byte
}

// NewLocalVaultAdapter creates a new local vault adapter
func NewLocalVaultAdapter() *LocalVaultAdapter {
	return &LocalVaultAdapter{
		storage: make(map[string][]byte),
	}
}

// Store stores a value locally
func (lv *LocalVaultAdapter) Store(key string, value []byte) error {
	lv.storage[key] = value
	return nil
}

// Retrieve retrieves a value locally
func (lv *LocalVaultAdapter) Retrieve(key string) ([]byte, error) {
	value, exists := lv.storage[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return value, nil
}

// Delete deletes a value locally
func (lv *LocalVaultAdapter) Delete(key string) error {
	delete(lv.storage, key)
	return nil
}

// Rotate rotates a key locally (no-op for local storage)
func (lv *LocalVaultAdapter) Rotate(key string) error {
	// No rotation needed for local storage
	return nil
}