package handlers

import (
	"encoding/json"
	"fmt"
	"kolajAi/internal/models"
	"kolajAi/internal/services"
	"net/http"
	"time"
)

// MarketplaceHandler handles marketplace integration requests
type MarketplaceHandler struct {
	*Handler
	marketplaceService *services.MarketplaceIntegrationsService
}

// NewMarketplaceHandler creates a new marketplace handler
func NewMarketplaceHandler(h *Handler, marketplaceService *services.MarketplaceIntegrationsService) *MarketplaceHandler {
	return &MarketplaceHandler{
		Handler:            h,
		marketplaceService: marketplaceService,
	}
}

// GetAllIntegrations returns all available marketplace integrations
func (h *MarketplaceHandler) GetAllIntegrations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get integration type filter if provided
	integrationType := r.URL.Query().Get("type")

	var integrations interface{}
	if integrationType != "" {
		integrations = h.marketplaceService.GetIntegrationsByType(integrationType)
	} else {
		integrations = h.marketplaceService.GetAllIntegrations()
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"integrations": integrations,
	})
}

// GetIntegration returns a specific integration
func (h *MarketplaceHandler) GetIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get integration ID from query
	integrationID := r.URL.Query().Get("id")
	if integrationID == "" {
		http.Error(w, "Integration ID required", http.StatusBadRequest)
		return
	}

	// Get integration
	integration, err := h.marketplaceService.GetIntegration(integrationID)
	if err != nil {
		http.Error(w, "Integration not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"integration": integration,
	})
}

// ConfigureIntegration configures a marketplace integration
func (h *MarketplaceHandler) ConfigureIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated and is admin
	user, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if user.Role != "admin" {
		http.Error(w, "Forbidden: Only admin users can configure integrations", http.StatusForbidden)
		return
	}

	// Parse request body
	var req struct {
		IntegrationID string            `json:"integration_id"`
		Credentials   map[string]string `json:"credentials"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Configure integration
	if err := h.marketplaceService.ConfigureIntegration(req.IntegrationID, req.Credentials); err != nil {
		Logger.Printf("Error configuring integration: %v", err)
		http.Error(w, "Failed to configure integration", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Integration configured successfully",
	})
}

// SyncProducts syncs products with a marketplace
func (h *MarketplaceHandler) SyncProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	user, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user has permission (admin or seller)
	if user.Role != "admin" && user.Role != "seller" {
		http.Error(w, "Forbidden: Only admin and seller users can sync products", http.StatusForbidden)
		return
	}

	// Parse request body
	var req struct {
		IntegrationID string        `json:"integration_id"`
		ProductIDs    []int         `json:"product_ids"`
		Products      []interface{} `json:"products,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Sync products
	if err := h.marketplaceService.SyncProducts(req.IntegrationID, req.Products); err != nil {
		Logger.Printf("Error syncing products: %v", err)
		http.Error(w, "Failed to sync products", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Products synced successfully",
	})
}

// GetMarketplaceOrders retrieves orders from a marketplace
func (h *MarketplaceHandler) GetMarketplaceOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	user, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user has permission
	if user.Role != "admin" && user.Role != "seller" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get integration ID from query
	integrationID := r.URL.Query().Get("integration_id")
	if integrationID == "" {
		http.Error(w, "Integration ID required", http.StatusBadRequest)
		return
	}

	// Get orders
	orders, err := h.marketplaceService.GetMarketplaceOrders(integrationID, h.parseTimeParam(r, "since"))
	if err != nil {
		Logger.Printf("Error getting marketplace orders: %v", err)
		http.Error(w, "Failed to get orders", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"orders":  orders,
		"count":   len(orders),
	})
}

// CreateShipment creates a shipment with cargo integration
func (h *MarketplaceHandler) CreateShipment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	user, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user has permission
	if user.Role != "admin" && user.Role != "seller" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse request body
	var req struct {
		CargoID      string      `json:"cargo_id"`
		ShipmentData interface{} `json:"shipment_data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create shipment
	trackingNumber, err := h.marketplaceService.CreateShipment(req.CargoID, req.ShipmentData)
	if err != nil {
		Logger.Printf("Error creating shipment: %v", err)
		http.Error(w, "Failed to create shipment", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":         true,
		"tracking_number": trackingNumber,
	})
}

// GenerateInvoice generates an invoice using e-fatura integration
func (h *MarketplaceHandler) GenerateInvoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	user, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user has permission
	if user.Role != "admin" && user.Role != "seller" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse request body
	var req struct {
		EFaturaID   string      `json:"efatura_id"`
		InvoiceData interface{} `json:"invoice_data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate invoice
	invoiceNumber, err := h.marketplaceService.GenerateInvoice(req.EFaturaID, req.InvoiceData)
	if err != nil {
		Logger.Printf("Error generating invoice: %v", err)
		http.Error(w, "Failed to generate invoice", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"invoice_number": invoiceNumber,
	})
}

// UpdateInventory updates inventory across marketplaces
func (h *MarketplaceHandler) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	user, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user has permission
	if user.Role != "admin" && user.Role != "seller" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse request body
	var req struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update inventory
	if err := h.marketplaceService.UpdateInventory(req.ProductID, req.Quantity); err != nil {
		Logger.Printf("Error updating inventory: %v", err)
		http.Error(w, "Failed to update inventory", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Inventory updated successfully",
	})
}

// getUserFromSession is a helper function to get user from session
func (h *MarketplaceHandler) getUserFromSession(r *http.Request) (*models.User, error) {
	if !h.IsAuthenticated(r) {
		return nil, fmt.Errorf("user not authenticated")
	}

	session, err := h.SessionManager.GetSession(r)
	if err != nil {
		return nil, fmt.Errorf("session error: %w", err)
	}

	userInterface, ok := session.Values[UserKey]
	if !ok {
		return nil, fmt.Errorf("user not found in session")
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		return nil, fmt.Errorf("invalid user data")
	}

	return user, nil
}

// parseTimeParam parses time parameter from request
func (h *MarketplaceHandler) parseTimeParam(r *http.Request, param string) time.Time {
	timeStr := r.URL.Query().Get(param)
	if timeStr == "" {
		// Default to 24 hours ago
		return time.Now().Add(-24 * time.Hour)
	}

	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		// If parsing fails, default to 24 hours ago
		return time.Now().Add(-24 * time.Hour)
	}

	return t
}
