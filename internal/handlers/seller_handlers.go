package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"log"
	
	"kolajAi/internal/services"
)

// SellerHandler handles seller-related requests
type SellerHandler struct {
	*Handler
	VendorService  *services.VendorService
	ProductService *services.ProductService
	OrderService   *services.OrderService
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

	// Get current user ID (mock implementation)
	userID := 1 // Mock user ID - in real implementation, get from session
	if userID == 0 {
		h.RedirectWithFlash(w, r, "/login", "Kullanıcı bilgisi bulunamadı")
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

	// Get recent orders (mock data for now)
	recentOrders := []map[string]interface{}{
		{
			"id":          1,
			"customer":    "Ahmet Yılmaz",
			"product":     "Test Ürün",
			"amount":      99.99,
			"status":      "completed",
			"created_at":  time.Now().Add(-24 * time.Hour),
		},
		{
			"id":          2,
			"customer":    "Ayşe Demir",
			"product":     "Test Ürün 2",
			"amount":      149.99,
			"status":      "pending",
			"created_at":  time.Now().Add(-48 * time.Hour),
		},
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

	userID := 1 // Mock user ID
	if userID == 0 {
		h.RedirectWithFlash(w, r, "/login", "Kullanıcı bilgisi bulunamadı")
		return
	}

	// Get vendor information
	vendor, err := h.VendorService.GetVendorByUserID(userID)
	if err != nil {
		log.Printf("Error getting vendor: %v", err)
		h.HandleError(w, r, err, "Satıcı bilgileri alınamadı")
		return
	}

	// Get vendor products (mock data for now)
	products := []map[string]interface{}{
		{
			"id":          1,
			"name":        "Test Ürün 1",
			"price":       99.99,
			"stock":       50,
			"status":      "active",
			"created_at":  time.Now().Add(-72 * time.Hour),
		},
		{
			"id":          2,
			"name":        "Test Ürün 2",
			"price":       149.99,
			"stock":       25,
			"status":      "active",
			"created_at":  time.Now().Add(-96 * time.Hour),
		},
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

	userID := 1 // Mock user ID
	if userID == 0 {
		h.RedirectWithFlash(w, r, "/login", "Kullanıcı bilgisi bulunamadı")
		return
	}

	// Get vendor information
	vendor, err := h.VendorService.GetVendorByUserID(userID)
	if err != nil {
		log.Printf("Error getting vendor: %v", err)
		h.HandleError(w, r, err, "Satıcı bilgileri alınamadı")
		return
	}

	// Get vendor orders (mock data for now)
	orders := []map[string]interface{}{
		{
			"id":          1,
			"customer":    "Ahmet Yılmaz",
			"product":     "Test Ürün 1",
			"quantity":    2,
			"amount":      199.98,
			"status":      "completed",
			"created_at":  time.Now().Add(-24 * time.Hour),
		},
		{
			"id":          2,
			"customer":    "Ayşe Demir",
			"product":     "Test Ürün 2",
			"quantity":    1,
			"amount":      149.99,
			"status":      "pending",
			"created_at":  time.Now().Add(-48 * time.Hour),
		},
		{
			"id":          3,
			"customer":    "Mehmet Kaya",
			"product":     "Test Ürün 1",
			"quantity":    1,
			"amount":      99.99,
			"status":      "shipped",
			"created_at":  time.Now().Add(-72 * time.Hour),
		},
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

	userID := 1 // Mock user ID
	vendor, err := h.VendorService.GetVendorByUserID(userID)
	if err != nil {
		http.Error(w, "Vendor not found", http.StatusNotFound)
		return
	}

	// Mock products data
	products := []map[string]interface{}{
		{
			"id":          1,
			"name":        "Test Ürün 1",
			"price":       99.99,
			"stock":       50,
			"status":      "active",
			"vendor_id":   vendor.ID,
		},
		{
			"id":          2,
			"name":        "Test Ürün 2",
			"price":       149.99,
			"stock":       25,
			"status":      "active",
			"vendor_id":   vendor.ID,
		},
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

	userID := 1 // Mock user ID
	vendor, err := h.VendorService.GetVendorByUserID(userID)
	if err != nil {
		http.Error(w, "Vendor not found", http.StatusNotFound)
		return
	}

	// Mock orders data
	orders := []map[string]interface{}{
		{
			"id":          1,
			"customer":    "Ahmet Yılmaz",
			"product":     "Test Ürün 1",
			"quantity":    2,
			"amount":      199.98,
			"status":      "completed",
			"vendor_id":   vendor.ID,
			"created_at":  time.Now().Add(-24 * time.Hour),
		},
		{
			"id":          2,
			"customer":    "Ayşe Demir",
			"product":     "Test Ürün 2",
			"quantity":    1,
			"amount":      149.99,
			"status":      "pending",
			"vendor_id":   vendor.ID,
			"created_at":  time.Now().Add(-48 * time.Hour),
		},
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