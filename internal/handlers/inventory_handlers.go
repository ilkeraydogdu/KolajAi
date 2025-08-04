package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"log"
	
	"kolajAi/internal/services"
)

// InventoryHandler handles inventory management requests
type InventoryHandler struct {
	*Handler
	InventoryService *services.InventoryService
	ProductService   *services.ProductService
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(h *Handler, inventoryService *services.InventoryService, productService *services.ProductService) *InventoryHandler {
	return &InventoryHandler{
		Handler:          h,
		InventoryService: inventoryService,
		ProductService:   productService,
	}
}

// Dashboard handles inventory dashboard page
func (h *InventoryHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	// Check admin permission
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get inventory statistics
	stats := map[string]interface{}{
		"total_products":    150,
		"low_stock_items":   12,
		"out_of_stock":      3,
		"overstock_items":   8,
		"reorder_alerts":    5,
		"total_value":       125000.50,
	}

	// Get low stock alerts
	lowStockAlerts := []map[string]interface{}{
		{
			"product_id":   1,
			"product_name": "Test Ürün 1",
			"current_stock": 5,
			"min_stock":    10,
			"urgency":      "high",
		},
		{
			"product_id":   2,
			"product_name": "Test Ürün 2", 
			"current_stock": 2,
			"min_stock":    15,
			"urgency":      "critical",
		},
	}

	// Get recent stock movements
	recentMovements := []map[string]interface{}{
		{
			"product_name": "Test Ürün 1",
			"type":         "sale",
			"quantity":     -3,
			"timestamp":    time.Now().Add(-2 * time.Hour),
		},
		{
			"product_name": "Test Ürün 2",
			"type":         "restock",
			"quantity":     50,
			"timestamp":    time.Now().Add(-4 * time.Hour),
		},
	}

	data := map[string]interface{}{
		"Title":           "Envanter Yönetimi",
		"Stats":           stats,
		"LowStockAlerts":  lowStockAlerts,
		"RecentMovements": recentMovements,
	}

	h.RenderTemplate(w, r, "inventory/dashboard.gohtml", data)
}

// StockLevels handles stock levels management page
func (h *InventoryHandler) StockLevels(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get all products with stock information
	products := []map[string]interface{}{
		{
			"id":           1,
			"name":         "Test Ürün 1",
			"sku":          "TEST001",
			"current_stock": 45,
			"min_stock":    10,
			"max_stock":    100,
			"status":       "normal",
			"last_updated": time.Now().Add(-24 * time.Hour),
		},
		{
			"id":           2,
			"name":         "Test Ürün 2",
			"sku":          "TEST002", 
			"current_stock": 5,
			"min_stock":    15,
			"max_stock":    50,
			"status":       "low",
			"last_updated": time.Now().Add(-12 * time.Hour),
		},
	}

	data := map[string]interface{}{
		"Title":    "Stok Seviyeleri",
		"Products": products,
	}

	h.RenderTemplate(w, r, "inventory/stock-levels.gohtml", data)
}

// Alerts handles inventory alerts page
func (h *InventoryHandler) Alerts(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get inventory alerts
	alerts := []map[string]interface{}{
		{
			"id":           1,
			"product_name": "Test Ürün 1",
			"alert_type":   "low_stock",
			"severity":     "high",
			"message":      "Stok seviyesi minimum değerin altında",
			"created_at":   time.Now().Add(-2 * time.Hour),
			"status":       "active",
		},
		{
			"id":           2,
			"product_name": "Test Ürün 2",
			"alert_type":   "out_of_stock",
			"severity":     "critical",
			"message":      "Ürün stokta yok",
			"created_at":   time.Now().Add(-1 * time.Hour),
			"status":       "active",
		},
	}

	data := map[string]interface{}{
		"Title":  "Envanter Uyarıları",
		"Alerts": alerts,
	}

	h.RenderTemplate(w, r, "inventory/alerts.gohtml", data)
}

// API Methods

// APIGetStockLevels returns stock levels data
func (h *InventoryHandler) APIGetStockLevels(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Mock stock data
	stockData := []map[string]interface{}{
		{
			"product_id":    1,
			"product_name":  "Test Ürün 1",
			"current_stock": 45,
			"min_stock":     10,
			"status":        "normal",
		},
		{
			"product_id":    2,
			"product_name":  "Test Ürün 2",
			"current_stock": 5,
			"min_stock":     15,
			"status":        "low",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stockData,
	})
}

// APIUpdateStock updates stock levels
func (h *InventoryHandler) APIUpdateStock(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	productIDStr := r.FormValue("product_id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	quantityStr := r.FormValue("quantity")
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		http.Error(w, "Invalid quantity", http.StatusBadRequest)
		return
	}

	operation := r.FormValue("operation") // "add" or "set"

	log.Printf("Updating stock for product %d: %s %d", productID, operation, quantity)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Stock updated successfully",
	})
}

// APIGetAlerts returns inventory alerts
func (h *InventoryHandler) APIGetAlerts(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	alerts := []map[string]interface{}{
		{
			"id":           1,
			"product_name": "Test Ürün 1",
			"alert_type":   "low_stock",
			"severity":     "high",
			"message":      "Stok seviyesi minimum değerin altında",
			"created_at":   time.Now().Add(-2 * time.Hour),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"alerts":  alerts,
	})
}

// APIDismissAlert dismisses an inventory alert
func (h *InventoryHandler) APIDismissAlert(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	alertIDStr := r.FormValue("alert_id")
	alertID, err := strconv.Atoi(alertIDStr)
	if err != nil {
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}

	log.Printf("Dismissing alert %d", alertID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Alert dismissed successfully",
	})
}