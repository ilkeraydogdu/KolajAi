package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
)

// AssetManager manages webpack built assets
type AssetManager struct {
	manifest map[string]string
	mu       sync.RWMutex
}

// NewAssetManager creates a new asset manager
func NewAssetManager(manifestPath string) *AssetManager {
	am := &AssetManager{
		manifest: make(map[string]string),
	}

	if err := am.LoadManifest(manifestPath); err != nil {
		log.Printf("Warning: Could not load asset manifest: %v", err)
	}

	return am
}

// LoadManifest loads the webpack manifest file
func (am *AssetManager) LoadManifest(manifestPath string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest file: %w", err)
	}

	if err := json.Unmarshal(data, &am.manifest); err != nil {
		return fmt.Errorf("failed to parse manifest JSON: %w", err)
	}

	return nil
}

// GetAssetURL returns the URL for a given asset key
func (am *AssetManager) GetAssetURL(key string) string {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if url, exists := am.manifest[key]; exists {
		return url
	}

	// Fallback to original key if not found in manifest
	return "/static/" + key
}

// GetCSSAssets returns all CSS asset URLs
func (am *AssetManager) GetCSSAssets() []string {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var assets []string
	for key, url := range am.manifest {
		if len(key) > 4 && key[len(key)-4:] == ".css" {
			assets = append(assets, url)
		}
	}

	return assets
}

// GetJSAssets returns JavaScript asset URLs in the correct order
func (am *AssetManager) GetJSAssets() []string {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// Define the correct loading order
	order := []string{"runtime.js", "vendors.js", "main.js"}
	var assets []string

	for _, key := range order {
		if url, exists := am.manifest[key]; exists {
			assets = append(assets, url)
		}
	}

	return assets
}

// GetAssetMap returns the full asset map for template usage
func (am *AssetManager) GetAssetMap() map[string]string {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// Return a copy to prevent external modifications
	result := make(map[string]string)
	for k, v := range am.manifest {
		result[k] = v
	}

	return result
}
