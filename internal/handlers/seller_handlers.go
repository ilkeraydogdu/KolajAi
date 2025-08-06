package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	
	"kolajAi/internal/models"
	"kolajAi/internal/services"
)

// SellerHandler handles seller-related requests
type SellerHandler struct {
	*Handler
	VendorService  *services.VendorService
	ProductService *services.ProductService
	OrderService   *services.OrderService
}

// getUserIDFromSession gets user ID from session
func (h *SellerHandler) getUserIDFromSession(w http.ResponseWriter, r *http.Request) (int, error) {
	session, err := h.SessionManager.GetSession(r)
	if err != nil {
		return 0, fmt.Errorf("oturum bilgisi alınamadı: %w", err)
	}
	
	userIDInterface, exists := session.Values[UserKey]
	if !exists {
		return 0, fmt.Errorf("kullanıcı bilgisi bulunamadı")
	}
	
	userID, ok := userIDInterface.(int)
	if !ok {
		return 0, fmt.Errorf("geçersiz kullanıcı bilgisi")
	}
	
	return userID, nil
}

// NewSellerHandler creates a new seller handler
func NewSellerHandler(h *Handler, vendorService *services.VendorService, productService *services.ProductService, orderService *services.OrderService) *SellerHandler {
	return &SellerHandler{
		Handler:        h,
		VendorService:  vendorService,
		ProductService: productService,
		OrderService:   orderService,
	}
}

// Dashboard handles seller dashboard page
func (h *SellerHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get current user ID from session
	userID, err := h.getUserIDFromSession(w, r)
	if err != nil {
		h.RedirectWithFlash(w, r, "/login", err.Error())
		return
	}

	// Get vendor information
	vendor, err := h.VendorService.GetVendorByUserID(userID)
	if err != nil {
		log.Printf("Error getting vendor: %v", err)
		h.HandleError(w, r, err, "Satıcı bilgileri alınamadı")
		return
	}

	// Get vendor statistics
	stats, err := h.VendorService.GetVendorStats(vendor.ID)
	if err != nil {
		log.Printf("Error getting vendor stats: %v", err)
		stats = map[string]interface{}{
			"total_products": 0,
			"total_orders":   0,
			"total_sales":    0.0,
			"rating":         0.0,
		}
	}

	// Get recent orders from database
	recentOrders, err := h.OrderService.GetOrdersByVendor(vendor.ID, 5, 0)
	if err != nil {
		log.Printf("Error getting recent orders: %v", err)
		recentOrders = []models.Order{} // Empty slice on error
	}

	data := map[string]interface{}{
		"Title":        "Satıcı Paneli",
		"Vendor":       vendor,
		"Stats":        stats,
		"RecentOrders": recentOrders,
	}

	h.RenderTemplate(w, r, "seller/dashboard.gohtml", data)
}

// Products handles seller products page
func (h *SellerHandler) Products(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get current user ID from session
	userID, err := h.getUserIDFromSession(w, r)
	if err != nil {
		h.RedirectWithFlash(w, r, "/login", err.Error())
		return
	}

	// Get vendor information
	vendor, err := h.VendorService.GetVendorByUserID(userID)
	if err != nil {
		log.Printf("Error getting vendor: %v", err)
		h.HandleError(w, r, err, "Satıcı bilgileri alınamadı")
		return
	}

	// Get vendor products from database
	products, err := h.ProductService.GetProductsByVendor(vendor.ID, 50, 0)
	if err != nil {
		log.Printf("Error getting vendor products: %v", err)
		products = []models.Product{} // Empty slice on error
	}

	data := map[string]interface{}{
		"Title":    "Ürünlerim",
		"Vendor":   vendor,
		"Products": products,
	}

	h.RenderTemplate(w, r, "seller/products.gohtml", data)
}

// Orders handles seller orders page
func (h *SellerHandler) Orders(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get current user ID from session
	userID, err := h.getUserIDFromSession(w, r)
	if err != nil {
		h.RedirectWithFlash(w, r, "/login", err.Error())
		return
	}

	// Get vendor information
	vendor, err := h.VendorService.GetVendorByUserID(userID)
	if err != nil {
		log.Printf("Error getting vendor: %v", err)
		h.HandleError(w, r, err, "Satıcı bilgileri alınamadı")
		return
	}

	// Get vendor orders from database
	orders, err := h.OrderService.GetOrdersByVendor(vendor.ID, 50, 0)
	if err != nil {
		log.Printf("Error getting vendor orders: %v", err)
		orders = []models.Order{} // Empty slice on error
	}

	data := map[string]interface{}{
		"Title":  "Siparişlerim",
		"Vendor": vendor,
		"Orders": orders,
	}

	h.RenderTemplate(w, r, "seller/orders.gohtml", data)
}

// API Methods

// APIGetProducts handles API request for vendor products
func (h *SellerHandler) APIGetProducts(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get current user ID from session
	userID, err := h.getUserIDFromSession(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	vendor, err := h.VendorService.GetVendorByUserID(userID)
	if err != nil {
		http.Error(w, "Vendor not found", http.StatusNotFound)
		return
	}

	// Get products from database
	products, err := h.ProductService.GetProductsByVendor(vendor.ID, 50, 0)
	if err != nil {
		log.Printf("Error getting vendor products: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"products": products,
	})
}

// APIUpdateProductStatus handles product status updates
func (h *SellerHandler) APIUpdateProductStatus(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse product ID from URL or form
	productIDStr := r.FormValue("product_id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	status := r.FormValue("status")
	if status == "" {
		http.Error(w, "Status is required", http.StatusBadRequest)
		return
	}

	// In real implementation, update product status in database
	log.Printf("Updating product %d status to %s", productID, status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Product status updated successfully",
	})
}

// APIGetOrders handles API request for vendor orders
func (h *SellerHandler) APIGetOrders(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get current user ID from session
	userID, err := h.getUserIDFromSession(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	vendor, err := h.VendorService.GetVendorByUserID(userID)
	if err != nil {
		http.Error(w, "Vendor not found", http.StatusNotFound)
		return
	}

	// Get orders from database
	orders, err := h.OrderService.GetOrdersByVendor(vendor.ID, 50, 0)
	if err != nil {
		log.Printf("Error getting vendor orders: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"orders":  orders,
	})
}

// APIUpdateOrderStatus handles order status updates
func (h *SellerHandler) APIUpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse order ID from URL or form
	orderIDStr := r.FormValue("order_id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	status := r.FormValue("status")
	if status == "" {
		http.Error(w, "Status is required", http.StatusBadRequest)
		return
	}

	// In real implementation, update order status in database
	log.Printf("Updating order %d status to %s", orderID, status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Order status updated successfully",
	})
}