package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"kolajAi/internal/services"
)

// EcommerceHandler handles e-commerce related requests
type EcommerceHandler struct {
	*Handler
	vendorService  *services.VendorService
	productService *services.ProductService
	orderService   *services.OrderService
	auctionService *services.AuctionService
}

// NewEcommerceHandler creates a new e-commerce handler
func NewEcommerceHandler(h *Handler, vendorService *services.VendorService, productService *services.ProductService, orderService *services.OrderService, auctionService *services.AuctionService) *EcommerceHandler {
	return &EcommerceHandler{
		Handler:        h,
		vendorService:  vendorService,
		productService: productService,
		orderService:   orderService,
		auctionService: auctionService,
	}
}

// GetProducts handles product listing
func (h *EcommerceHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")
	page := 1
	limit := 20
	
	// Parse page and limit from query params
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	// Get products from service
	products, err := h.productService.GetProducts(category, search, page, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get products: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"products": products,
		"success":  true,
		"page":     page,
		"limit":    limit,
	})
}

// GetProduct handles single product view
func (h *EcommerceHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	// Get product ID from URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/product/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	
	// Get product from service
	product, err := h.productService.GetProductByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get product: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"product": product,
		"success": true,
	})
}

// SearchProducts handles product search
func (h *EcommerceHandler) SearchProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Search query required", http.StatusBadRequest)
		return
	}
	
	// Parse pagination parameters
	page := 1
	limit := 20
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	// Search products using the service
	products, err := h.productService.GetProducts("", query, page, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to search products: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"products": products,
		"query":    query,
		"page":     page,
		"limit":    limit,
		"success":  true,
	})
}

// GetCategories handles category listing
func (h *EcommerceHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	// Get categories from service
	categories, err := h.productService.GetAllCategories()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get categories: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"categories": categories,
		"success":    true,
	})
}

// HealthCheck provides a simple health check endpoint
func (h *EcommerceHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"service": "ecommerce",
		"success": true,
	})
}
