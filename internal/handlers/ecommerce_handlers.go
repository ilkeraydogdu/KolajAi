package handlers

import (
	"encoding/json"
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
	// Get products from service - simplified
	products := []map[string]interface{}{
		{"id": 1, "name": "Sample Product", "price": 99.99},
	}
	
	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"products": products,
		"success":  true,
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
	
	// Get product from service - simplified
	product := map[string]interface{}{
		"id":    id,
		"name":  "Sample Product",
		"price": 99.99,
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
	
	// Search products - simplified
	products := []map[string]interface{}{
		{"id": 1, "name": "Found Product", "price": 49.99},
	}
	
	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"products": products,
		"query":    query,
		"success":  true,
	})
}

// GetCategories handles category listing
func (h *EcommerceHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	// Get categories from service - simplified
	categories := []map[string]interface{}{
		{"id": 1, "name": "Electronics", "slug": "electronics"},
		{"id": 2, "name": "Clothing", "slug": "clothing"},
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
