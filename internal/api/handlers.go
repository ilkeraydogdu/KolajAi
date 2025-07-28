package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kolajAi/internal/models"
	"kolajAi/internal/services"
	"kolajAi/internal/validation"
)

// APIHandlers provides comprehensive REST API endpoints
type APIHandlers struct {
	middleware      *APIMiddleware
	productService  *services.ProductService
	orderService    *services.OrderService
	vendorService   *services.VendorService
	aiService       *services.AIService
	authService     *services.AuthService
	inventoryService *services.InventoryService
	validator       *validation.Validator
}

// NewAPIHandlers creates new API handlers
func NewAPIHandlers(
	middleware *APIMiddleware,
	productService *services.ProductService,
	orderService *services.OrderService,
	vendorService *services.VendorService,
	aiService *services.AIService,
	authService *services.AuthService,
	inventoryService *services.InventoryService,
	validator *validation.Validator,
) *APIHandlers {
	return &APIHandlers{
		middleware:       middleware,
		productService:   productService,
		orderService:     orderService,
		vendorService:    vendorService,
		aiService:        aiService,
		authService:      authService,
		inventoryService: inventoryService,
		validator:        validator,
	}
}

// RegisterRoutes registers all API routes
func (h *APIHandlers) RegisterRoutes(mux *http.ServeMux) {
	// API v1 routes
	apiV1 := "/api/v1"

	// Product endpoints
	mux.HandleFunc(apiV1+"/products", h.middleware.APIHandler(h.handleProducts))
	mux.HandleFunc(apiV1+"/products/", h.middleware.APIHandler(h.handleProductByID))
	mux.HandleFunc(apiV1+"/products/search", h.middleware.APIHandler(h.handleProductSearch))
	mux.HandleFunc(apiV1+"/products/categories", h.middleware.APIHandler(h.handleProductCategories))
	mux.HandleFunc(apiV1+"/products/featured", h.middleware.APIHandler(h.handleFeaturedProducts))
	mux.HandleFunc(apiV1+"/products/trending", h.middleware.APIHandler(h.handleTrendingProducts))

	// Order endpoints
	mux.HandleFunc(apiV1+"/orders", h.middleware.APIHandler(h.handleOrders))
	mux.HandleFunc(apiV1+"/orders/", h.middleware.APIHandler(h.handleOrderByID))
	mux.HandleFunc(apiV1+"/orders/status", h.middleware.APIHandler(h.handleOrderStatus))

	// Cart endpoints
	mux.HandleFunc(apiV1+"/cart", h.middleware.APIHandler(h.handleCart))
	mux.HandleFunc(apiV1+"/cart/add", h.middleware.APIHandler(h.handleAddToCart))
	mux.HandleFunc(apiV1+"/cart/update", h.middleware.APIHandler(h.handleUpdateCart))
	mux.HandleFunc(apiV1+"/cart/remove", h.middleware.APIHandler(h.handleRemoveFromCart))
	mux.HandleFunc(apiV1+"/cart/clear", h.middleware.APIHandler(h.handleClearCart))

	// User endpoints
	mux.HandleFunc(apiV1+"/users/profile", h.middleware.APIHandler(h.handleUserProfile))
	mux.HandleFunc(apiV1+"/users/orders", h.middleware.APIHandler(h.handleUserOrders))
	mux.HandleFunc(apiV1+"/users/wishlist", h.middleware.APIHandler(h.handleUserWishlist))

	// Vendor endpoints
	mux.HandleFunc(apiV1+"/vendors", h.middleware.APIHandler(h.handleVendors))
	mux.HandleFunc(apiV1+"/vendors/", h.middleware.APIHandler(h.handleVendorByID))
	mux.HandleFunc(apiV1+"/vendors/products", h.middleware.APIHandler(h.handleVendorProducts))
	mux.HandleFunc(apiV1+"/vendors/orders", h.middleware.APIHandler(h.handleVendorOrders))

	// AI endpoints
	mux.HandleFunc(apiV1+"/ai/recommendations", h.middleware.APIHandler(h.handleAIRecommendations))
	mux.HandleFunc(apiV1+"/ai/search", h.middleware.APIHandler(h.handleAISearch))
	mux.HandleFunc(apiV1+"/ai/price-optimization", h.middleware.APIHandler(h.handlePriceOptimization))
	mux.HandleFunc(apiV1+"/ai/analytics", h.middleware.APIHandler(h.handleAIAnalytics))

	// Inventory endpoints
	mux.HandleFunc(apiV1+"/inventory", h.middleware.APIHandler(h.handleInventory))
	mux.HandleFunc(apiV1+"/inventory/stock", h.middleware.APIHandler(h.handleStockLevels))
	mux.HandleFunc(apiV1+"/inventory/alerts", h.middleware.APIHandler(h.handleInventoryAlerts))

	// Authentication endpoints
	mux.HandleFunc(apiV1+"/auth/login", h.middleware.APIHandler(h.handleLogin))
	mux.HandleFunc(apiV1+"/auth/register", h.middleware.APIHandler(h.handleRegister))
	mux.HandleFunc(apiV1+"/auth/refresh", h.middleware.APIHandler(h.handleRefreshToken))
	mux.HandleFunc(apiV1+"/auth/logout", h.middleware.APIHandler(h.handleLogout))

	// Admin endpoints
	mux.HandleFunc(apiV1+"/admin/dashboard", h.middleware.APIHandler(h.handleAdminDashboard))
	mux.HandleFunc(apiV1+"/admin/users", h.middleware.APIHandler(h.handleAdminUsers))
	mux.HandleFunc(apiV1+"/admin/reports", h.middleware.APIHandler(h.handleAdminReports))
	mux.HandleFunc(apiV1+"/admin/system", h.middleware.APIHandler(h.handleSystemHealth))

	// Health check
	mux.HandleFunc(apiV1+"/health", h.middleware.APIHandler(h.handleHealthCheck))
}

// Product handlers

func (h *APIHandlers) handleProducts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.getProducts(w, r)
	case "POST":
		h.createProduct(w, r)
	default:
		h.sendError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

func (h *APIHandlers) getProducts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	category := r.URL.Query().Get("category")
	vendor := r.URL.Query().Get("vendor")
	minPrice, _ := strconv.ParseFloat(r.URL.Query().Get("min_price"), 64)
	maxPrice, _ := strconv.ParseFloat(r.URL.Query().Get("max_price"), 64)
	sortBy := r.URL.Query().Get("sort_by")
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := r.URL.Query().Get("sort_order")
	if sortOrder == "" {
		sortOrder = "desc"
	}

	offset := (page - 1) * limit

	// Build filters
	filters := map[string]interface{}{}
	if category != "" {
		filters["category_id"] = category
	}
	if vendor != "" {
		filters["vendor_id"] = vendor
	}
	if minPrice > 0 {
		filters["min_price"] = minPrice
	}
	if maxPrice > 0 {
		filters["max_price"] = maxPrice
	}

	// Get products
	products, err := h.productService.GetProductsWithFilters(filters, sortBy, sortOrder, limit, offset)
	if err != nil {
		h.sendError(w, r, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch products")
		return
	}

	// Get total count for pagination
	total, err := h.productService.GetProductCount(filters)
	if err != nil {
		h.sendError(w, r, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to count products")
		return
	}

	totalPages := (int(total) + limit - 1) / limit

	meta := &APIMeta{
		Page:       page,
		PerPage:    limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	h.middleware.SendSuccessResponse(w, r, products, meta)
}

func (h *APIHandlers) createProduct(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated and authorized
	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		h.sendError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		h.sendError(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
		return
	}

	// Validate product data
	if err := h.validator.ValidateStruct(&product); err != nil {
		h.sendError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Set user as vendor (or check if user is vendor)
	product.VendorID = int(userID)
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	// Create product
	productID, err := h.productService.CreateProduct(&product)
	if err != nil {
		h.sendError(w, r, http.StatusInternalServerError, "CREATE_ERROR", "Failed to create product")
		return
	}

	product.ID = int(productID)
	h.middleware.SendSuccessResponse(w, r, product, nil)
}

func (h *APIHandlers) handleProductByID(w http.ResponseWriter, r *http.Request) {
	// Extract product ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/products/")
	productID, err := strconv.Atoi(path)
	if err != nil {
		h.sendError(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	switch r.Method {
	case "GET":
		h.getProductByID(w, r, productID)
	case "PUT":
		h.updateProduct(w, r, productID)
	case "DELETE":
		h.deleteProduct(w, r, productID)
	default:
		h.sendError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

func (h *APIHandlers) getProductByID(w http.ResponseWriter, r *http.Request, productID int) {
	product, err := h.productService.GetProductByID(productID)
	if err != nil {
		h.sendError(w, r, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found")
		return
	}

	// Increment view count
	go h.productService.IncrementViewCount(productID)

	h.middleware.SendSuccessResponse(w, r, product, nil)
}

func (h *APIHandlers) updateProduct(w http.ResponseWriter, r *http.Request, productID int) {
	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		h.sendError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	// Check if user owns the product or is admin
	product, err := h.productService.GetProductByID(productID)
	if err != nil {
		h.sendError(w, r, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found")
		return
	}

	if product.VendorID != int(userID) && !h.isAdmin(r) {
		h.sendError(w, r, http.StatusForbidden, "FORBIDDEN", "Access denied")
		return
	}

	var updateData models.Product
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		h.sendError(w, r, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
		return
	}

	// Validate update data
	if err := h.validator.ValidateStruct(&updateData); err != nil {
		h.sendError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	updateData.ID = productID
	updateData.UpdatedAt = time.Now()

	err = h.productService.UpdateProduct(&updateData)
	if err != nil {
		h.sendError(w, r, http.StatusInternalServerError, "UPDATE_ERROR", "Failed to update product")
		return
	}

	h.middleware.SendSuccessResponse(w, r, updateData, nil)
}

func (h *APIHandlers) deleteProduct(w http.ResponseWriter, r *http.Request, productID int) {
	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		h.sendError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	// Check if user owns the product or is admin
	product, err := h.productService.GetProductByID(productID)
	if err != nil {
		h.sendError(w, r, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found")
		return
	}

	if product.VendorID != int(userID) && !h.isAdmin(r) {
		h.sendError(w, r, http.StatusForbidden, "FORBIDDEN", "Access denied")
		return
	}

	err = h.productService.DeleteProduct(productID)
	if err != nil {
		h.sendError(w, r, http.StatusInternalServerError, "DELETE_ERROR", "Failed to delete product")
		return
	}

	h.middleware.SendSuccessResponse(w, r, map[string]string{"message": "Product deleted successfully"}, nil)
}

func (h *APIHandlers) handleProductSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.sendError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		h.sendError(w, r, http.StatusBadRequest, "MISSING_QUERY", "Search query is required")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Use AI service for enhanced search
	searchResult, err := h.aiService.EnhancedSearch(query, limit, offset)
	if err != nil {
		h.sendError(w, r, http.StatusInternalServerError, "SEARCH_ERROR", "Search failed")
		return
	}

	totalPages := (searchResult.TotalCount + limit - 1) / limit

	meta := &APIMeta{
		Page:       page,
		PerPage:    limit,
		Total:      searchResult.TotalCount,
		TotalPages: totalPages,
	}

	response := map[string]interface{}{
		"products":       searchResult.Products,
		"suggestions":    searchResult.Suggestions,
		"processed_time": searchResult.ProcessedTime.String(),
	}

	h.middleware.SendSuccessResponse(w, r, response, meta)
}

// AI handlers

func (h *APIHandlers) handleAIRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.sendError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		h.sendError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	recommendations, err := h.aiService.GetPersonalizedRecommendations(int(userID), limit)
	if err != nil {
		h.sendError(w, r, http.StatusInternalServerError, "AI_ERROR", "Failed to get recommendations")
		return
	}

	h.middleware.SendSuccessResponse(w, r, recommendations, nil)
}

func (h *APIHandlers) handlePriceOptimization(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.sendError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		h.sendError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	// Check if user is vendor
	if !h.isVendor(r) {
		h.sendError(w, r, http.StatusForbidden, "FORBIDDEN", "Vendor access required")
		return
	}

	optimizations, err := h.aiService.GetPriceOptimizations(int(userID))
	if err != nil {
		h.sendError(w, r, http.StatusInternalServerError, "AI_ERROR", "Failed to get price optimizations")
		return
	}

	h.middleware.SendSuccessResponse(w, r, optimizations, nil)
}

// Health check handler

func (h *APIHandlers) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.sendError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   h.middleware.Config.Version,
		"services": map[string]string{
			"database": "healthy",
			"cache":    "healthy",
			"ai":       "healthy",
		},
	}

	h.middleware.SendSuccessResponse(w, r, health, nil)
}

// Helper methods

func (h *APIHandlers) sendError(w http.ResponseWriter, r *http.Request, statusCode int, code string, message string) {
	h.middleware.sendErrorResponse(w, r, statusCode, code, message)
}

func (h *APIHandlers) getUserIDFromContext(r *http.Request) int64 {
	// This would extract user ID from JWT token or session
	// Implementation depends on authentication method
	return 0 // Placeholder
}

func (h *APIHandlers) isAdmin(r *http.Request) bool {
	// Check if user has admin role
	return false // Placeholder
}

func (h *APIHandlers) isVendor(r *http.Request) bool {
	// Check if user has vendor role
	return false // Placeholder
}

// Additional handlers for cart, orders, etc. would be implemented similarly
func (h *APIHandlers) handleCart(w http.ResponseWriter, r *http.Request) {
	// Cart implementation
}

func (h *APIHandlers) handleAddToCart(w http.ResponseWriter, r *http.Request) {
	// Add to cart implementation
}

func (h *APIHandlers) handleOrders(w http.ResponseWriter, r *http.Request) {
	// Orders implementation
}

func (h *APIHandlers) handleLogin(w http.ResponseWriter, r *http.Request) {
	// Login implementation
}

func (h *APIHandlers) handleRegister(w http.ResponseWriter, r *http.Request) {
	// Register implementation
}

func (h *APIHandlers) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	// Admin dashboard API implementation
}