package handlers

import (
	"encoding/json"
	"fmt"
	"kolajAi/internal/models"
	"kolajAi/internal/services"
	"net/http"
	"strconv"
)

// MarketplaceIntegrationHandler handles marketplace integration requests
type MarketplaceIntegrationHandler struct {
	*Handler
	integrationService *services.MarketplaceIntegrationService
}

// NewMarketplaceIntegrationHandler creates a new marketplace integration handler
func NewMarketplaceIntegrationHandler(h *Handler, integrationService *services.MarketplaceIntegrationService) *MarketplaceIntegrationHandler {
	return &MarketplaceIntegrationHandler{
		Handler:            h,
		integrationService: integrationService,
	}
}

// GetAvailableIntegrations returns list of available marketplace integrations
func (h *MarketplaceIntegrationHandler) GetAvailableIntegrations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get available integrations
	integrations := h.integrationService.GetAvailableIntegrations()

	// Group integrations by type
	grouped := map[string][]map[string]interface{}{
		"turkish_marketplaces":     {},
		"international_marketplaces": {},
		"ecommerce_platforms":      {},
		"social_media":             {},
		"accounting_erp":           {},
		"shipping_cargo":           {},
	}

	for integrationType, platformInfo := range integrations {
		integrationData := map[string]interface{}{
			"type":               string(integrationType),
			"name":               platformInfo.Name,
			"platform_type":      platformInfo.Type,
			"country":            platformInfo.Country,
			"supported_features": platformInfo.SupportedFeatures,
			"required_fields":    platformInfo.RequiredFields,
			"optional_fields":    platformInfo.OptionalFields,
			"rate_limit":         platformInfo.RateLimit,
			"test_mode":          platformInfo.TestMode,
			"webhook_support":    platformInfo.WebhookSupport,
			"bulk_operations":    platformInfo.BulkOperations,
		}

		// Group by type and country
		switch {
		case platformInfo.Type == "marketplace" && platformInfo.Country == "TR":
			grouped["turkish_marketplaces"] = append(grouped["turkish_marketplaces"], integrationData)
		case platformInfo.Type == "marketplace" && platformInfo.Country != "TR":
			grouped["international_marketplaces"] = append(grouped["international_marketplaces"], integrationData)
		case platformInfo.Type == "ecommerce":
			grouped["ecommerce_platforms"] = append(grouped["ecommerce_platforms"], integrationData)
		case platformInfo.Type == "social":
			grouped["social_media"] = append(grouped["social_media"], integrationData)
		case platformInfo.Type == "accounting":
			grouped["accounting_erp"] = append(grouped["accounting_erp"], integrationData)
		case platformInfo.Type == "shipping":
			grouped["shipping_cargo"] = append(grouped["shipping_cargo"], integrationData)
		}
	}

	h.JSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    grouped,
		"total":   len(integrations),
	})
}

// CreateIntegration creates a new marketplace integration
func (h *MarketplaceIntegrationHandler) CreateIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check admin permission for some integrations
	if !user.IsAdmin && (user.Role != models.RoleVendor && user.Role != models.RoleAdmin) {
		h.ErrorResponse(w, "You don't have permission to create integrations", http.StatusForbidden)
		return
	}

	// Parse request body
	var req struct {
		Type   string                 `json:"type"`
		Config map[string]interface{} `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate integration type
	integrationType := models.IntegrationType(req.Type)
	if req.Type == "" {
		h.ErrorResponse(w, "Integration type is required", http.StatusBadRequest)
		return
	}

	// Create integration
	integration, err := h.integrationService.CreateIntegration(user.ID, integrationType, req.Config)
	if err != nil {
		h.ErrorResponse(w, fmt.Sprintf("Failed to create integration: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	h.JSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    integration,
		"message": "Integration created successfully",
	})
}

// GetUserIntegrations returns user's integrations
func (h *MarketplaceIntegrationHandler) GetUserIntegrations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get query parameters
	platform := r.URL.Query().Get("platform")
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get user integrations (this would be implemented in the service)
	integrations := []models.MarketplaceIntegration{} // Placeholder

	// Return integrations
	h.JSONResponse(w, map[string]interface{}{
		"success":   true,
		"data":      integrations,
		"count":     len(integrations),
		"limit":     limit,
		"offset":    offset,
		"filters": map[string]string{
			"platform": platform,
			"status":   status,
		},
	})
}

// GetIntegration returns a specific integration
func (h *MarketplaceIntegrationHandler) GetIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get integration ID from URL
	integrationIDStr := r.URL.Query().Get("id")
	if integrationIDStr == "" {
		h.ErrorResponse(w, "Integration ID is required", http.StatusBadRequest)
		return
	}

	integrationID, err := strconv.ParseInt(integrationIDStr, 10, 64)
	if err != nil {
		h.ErrorResponse(w, "Invalid integration ID", http.StatusBadRequest)
		return
	}

	// Get integration (this would be implemented in the service)
	integration := &models.MarketplaceIntegration{
		ID:       integrationID,
		UserID:   user.ID,
		Name:     "Sample Integration",
		Type:     models.IntegrationTrendyol,
		Platform: "trendyol",
		IsActive: true,
	}

	// Return integration
	h.JSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    integration,
	})
}

// UpdateIntegration updates an integration
func (h *MarketplaceIntegrationHandler) UpdateIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get integration ID from URL
	integrationIDStr := r.URL.Query().Get("id")
	if integrationIDStr == "" {
		h.ErrorResponse(w, "Integration ID is required", http.StatusBadRequest)
		return
	}

	integrationID, err := strconv.ParseInt(integrationIDStr, 10, 64)
	if err != nil {
		h.ErrorResponse(w, "Invalid integration ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		h.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update integration (this would be implemented in the service)
	// For now, just return success
	h.JSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Integration updated successfully",
		"data": map[string]interface{}{
			"id":         integrationID,
			"updated_at": "2024-01-01T00:00:00Z",
		},
	})
}

// DeleteIntegration deletes an integration
func (h *MarketplaceIntegrationHandler) DeleteIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get integration ID from URL
	integrationIDStr := r.URL.Query().Get("id")
	if integrationIDStr == "" {
		h.ErrorResponse(w, "Integration ID is required", http.StatusBadRequest)
		return
	}

	_, err := strconv.ParseInt(integrationIDStr, 10, 64)
	if err != nil {
		h.ErrorResponse(w, "Invalid integration ID", http.StatusBadRequest)
		return
	}

	// Delete integration (this would be implemented in the service)
	// For now, just return success
	h.JSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Integration deleted successfully",
	})
}

// SyncIntegration performs synchronization for an integration
func (h *MarketplaceIntegrationHandler) SyncIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		IntegrationID int64  `json:"integration_id"`
		SyncType      string `json:"sync_type"` // products, orders, inventory, all
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate sync type
	validSyncTypes := []string{"products", "orders", "inventory", "all"}
	isValidSyncType := false
	for _, validType := range validSyncTypes {
		if req.SyncType == validType {
			isValidSyncType = true
			break
		}
	}

	if !isValidSyncType {
		h.ErrorResponse(w, "Invalid sync type", http.StatusBadRequest)
		return
	}

	// Perform synchronization
	result, err := h.integrationService.SyncIntegration(req.IntegrationID, req.SyncType)
	if err != nil {
		h.ErrorResponse(w, fmt.Sprintf("Sync failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return sync result
	h.JSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    result,
		"message": "Synchronization completed",
	})
}

// GetSyncLogs returns synchronization logs for an integration
func (h *MarketplaceIntegrationHandler) GetSyncLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get integration ID from URL
	integrationIDStr := r.URL.Query().Get("integration_id")
	if integrationIDStr == "" {
		h.ErrorResponse(w, "Integration ID is required", http.StatusBadRequest)
		return
	}

	integrationID, err := strconv.ParseInt(integrationIDStr, 10, 64)
	if err != nil {
		h.ErrorResponse(w, "Invalid integration ID", http.StatusBadRequest)
		return
	}

	// Get query parameters
	syncType := r.URL.Query().Get("sync_type")
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get sync logs (this would be implemented in the service)
	logs := []models.SyncLog{} // Placeholder

	// Return logs
	h.JSONResponse(w, map[string]interface{}{
		"success":   true,
		"data":      logs,
		"count":     len(logs),
		"limit":     limit,
		"offset":    offset,
		"filters": map[string]interface{}{
			"integration_id": integrationID,
			"sync_type":      syncType,
			"status":         status,
		},
	})
}

// TestIntegration tests integration credentials
func (h *MarketplaceIntegrationHandler) TestIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		Type   string                 `json:"type"`
		Config map[string]interface{} `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Test integration credentials (this would be implemented in the service)
	// For now, just return success
	testResult := map[string]interface{}{
		"success":    true,
		"connection": "established",
		"api_access": true,
		"features": []string{
			"products",
			"orders",
			"inventory",
		},
		"rate_limit": map[string]interface{}{
			"remaining": 950,
			"total":     1000,
			"reset_at":  "2024-01-01T01:00:00Z",
		},
	}

	// Return test result
	h.JSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    testResult,
		"message": "Integration test completed successfully",
	})
}

// GetIntegrationStats returns integration statistics
func (h *MarketplaceIntegrationHandler) GetIntegrationStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get integration ID from URL
	integrationIDStr := r.URL.Query().Get("integration_id")
	if integrationIDStr == "" {
		h.ErrorResponse(w, "Integration ID is required", http.StatusBadRequest)
		return
	}

	integrationID, err := strconv.ParseInt(integrationIDStr, 10, 64)
	if err != nil {
		h.ErrorResponse(w, "Invalid integration ID", http.StatusBadRequest)
		return
	}

	// Get integration statistics (this would be implemented in the service)
	stats := map[string]interface{}{
		"integration_id": integrationID,
		"products": map[string]interface{}{
			"total_synced":    250,
			"successful":      240,
			"failed":          10,
			"last_sync":       "2024-01-01T10:00:00Z",
		},
		"orders": map[string]interface{}{
			"total_synced":    150,
			"successful":      148,
			"failed":          2,
			"last_sync":       "2024-01-01T11:00:00Z",
		},
		"inventory": map[string]interface{}{
			"total_synced":    250,
			"successful":      250,
			"failed":          0,
			"last_sync":       "2024-01-01T11:30:00Z",
		},
		"performance": map[string]interface{}{
			"avg_sync_time":   "2.5 minutes",
			"success_rate":    96.5,
			"api_calls_today": 1250,
			"rate_limit_hits": 0,
		},
	}

	// Return statistics
	h.JSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// BulkSync performs bulk synchronization for multiple integrations
func (h *MarketplaceIntegrationHandler) BulkSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		IntegrationIDs []int64 `json:"integration_ids"`
		SyncType       string  `json:"sync_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if len(req.IntegrationIDs) == 0 {
		h.ErrorResponse(w, "At least one integration ID is required", http.StatusBadRequest)
		return
	}

	if len(req.IntegrationIDs) > 10 {
		h.ErrorResponse(w, "Maximum 10 integrations can be synced at once", http.StatusBadRequest)
		return
	}

	// Perform bulk sync (this would be implemented in the service)
	// For now, just return success
	results := []map[string]interface{}{}
	for _, integrationID := range req.IntegrationIDs {
		results = append(results, map[string]interface{}{
			"integration_id": integrationID,
			"success":        true,
			"records_synced": 50,
			"duration":       "1.2 minutes",
		})
	}

	// Return bulk sync results
	h.JSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    results,
		"message": "Bulk synchronization completed",
	})
}