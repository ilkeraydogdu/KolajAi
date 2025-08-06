package handlers

import (
	"log"
	"net/http"
	"strconv"

	"kolajAi/internal/models"
	"kolajAi/internal/services"
)

// MarketplacePageHandler handles marketplace web page requests
type MarketplacePageHandler struct {
	*Handler
	productService *services.ProductService
	auctionService *services.AuctionService
	orderService   *services.OrderService
	vendorService  *services.VendorService
}

// NewMarketplacePageHandler creates a new marketplace page handler
func NewMarketplacePageHandler(h *Handler, productService *services.ProductService, auctionService *services.AuctionService, orderService *services.OrderService, vendorService *services.VendorService) *MarketplacePageHandler {
	return &MarketplacePageHandler{
		Handler:        h,
		productService: productService,
		auctionService: auctionService,
		orderService:   orderService,
		vendorService:  vendorService,
	}
}

// Index handles marketplace home page
func (h *MarketplacePageHandler) Index(w http.ResponseWriter, r *http.Request) {
	// Get categories from database
	categories, err := h.productService.GetAllCategories()
	if err != nil {
		log.Printf("Error loading categories: %v", err)
		categories = []models.Category{} // Empty slice on error
	}

	// Get featured products
	featuredProducts, err := h.productService.GetFeaturedProducts(8, 0)
	if err != nil {
		log.Printf("Error loading featured products: %v", err)
		featuredProducts = []models.Product{} // Empty slice on error
	}

	// Get active auctions
	activeAuctions, err := h.auctionService.GetActiveAuctions(6)
	if err != nil {
		log.Printf("Error loading active auctions: %v", err)
		activeAuctions = []models.Auction{} // Empty slice on error
	}

	data := map[string]interface{}{
		"Title":            "KolajAI Marketplace",
		"Categories":       categories,
		"FeaturedProducts": featuredProducts,
		"ActiveAuctions":   activeAuctions,
		"AppName":          "KolajAI",
	}

	h.RenderTemplate(w, r, "marketplace/index", data)
}

// Products handles marketplace products page
func (h *MarketplacePageHandler) Products(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")
	page := 1
	limit := 20

	// Parse page from query params
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Get products from database
	products, err := h.productService.GetProducts(category, search, page, limit)
	if err != nil {
		log.Printf("Error loading products: %v", err)
		products = []models.Product{} // Empty slice on error
	}

	// Get categories for filter
	categories, err := h.productService.GetAllCategories()
	if err != nil {
		log.Printf("Error loading categories: %v", err)
		categories = []models.Category{} // Empty slice on error
	}

	data := map[string]interface{}{
		"Title":      "Ürünler - KolajAI Marketplace",
		"Products":   products,
		"Categories": categories,
		"Search":     search,
		"Category":   category,
		"Page":       page,
		"AppName":    "KolajAI",
	}

	h.RenderTemplate(w, r, "marketplace/products", data)
}

// ProductDetail handles single product page
func (h *MarketplacePageHandler) ProductDetail(w http.ResponseWriter, r *http.Request) {
	// Extract product ID from URL path
	idStr := r.URL.Path[len("/marketplace/product/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.HandleError(w, r, err, "Geçersiz ürün ID")
		return
	}

	// Get product from database
	product, err := h.productService.GetProductByID(id)
	if err != nil {
		h.HandleError(w, r, err, "Ürün bulunamadı")
		return
	}

	// Get related products
	relatedProducts, err := h.productService.GetProductsByCategory(product.CategoryID, 4, 0)
	if err != nil {
		log.Printf("Error loading related products: %v", err)
		relatedProducts = []models.Product{} // Empty slice on error
	}

	data := map[string]interface{}{
		"Title":           product.Name + " - KolajAI Marketplace",
		"Product":         product,
		"RelatedProducts": relatedProducts,
		"AppName":         "KolajAI",
	}

	h.RenderTemplate(w, r, "marketplace/product-detail", data)
}

// Auctions handles marketplace auctions page
func (h *MarketplacePageHandler) Auctions(w http.ResponseWriter, r *http.Request) {
	// Get active auctions
	auctions, err := h.auctionService.GetActiveAuctions(20)
	if err != nil {
		log.Printf("Error loading auctions: %v", err)
		auctions = []models.Auction{} // Empty slice on error
	}

	data := map[string]interface{}{
		"Title":    "Açık Artırmalar - KolajAI Marketplace",
		"Auctions": auctions,
		"AppName":  "KolajAI",
	}

	h.RenderTemplate(w, r, "marketplace/auctions", data)
}

// AuctionDetail handles single auction page
func (h *MarketplacePageHandler) AuctionDetail(w http.ResponseWriter, r *http.Request) {
	// Extract auction ID from URL path
	idStr := r.URL.Path[len("/marketplace/auction/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.HandleError(w, r, err, "Geçersiz açık artırma ID")
		return
	}

	// Get auction from database
	auction, err := h.auctionService.GetAuctionByID(id)
	if err != nil {
		h.HandleError(w, r, err, "Açık artırma bulunamadı")
		return
	}

	// Get auction bids
	bids, err := h.auctionService.GetAuctionBids(id, 10, 0)
	if err != nil {
		log.Printf("Error loading auction bids: %v", err)
		bids = []models.AuctionBid{} // Empty slice on error
	}

	data := map[string]interface{}{
		"Title":   auction.Title + " - KolajAI Marketplace",
		"Auction": auction,
		"Bids":    bids,
		"AppName": "KolajAI",
	}

	h.RenderTemplate(w, r, "marketplace/auction-detail", data)
}

// Categories handles marketplace categories page
func (h *MarketplacePageHandler) Categories(w http.ResponseWriter, r *http.Request) {
	// Get all categories
	categories, err := h.productService.GetAllCategories()
	if err != nil {
		log.Printf("Error loading categories: %v", err)
		categories = []models.Category{} // Empty slice on error
	}

	data := map[string]interface{}{
		"Title":      "Kategoriler - KolajAI Marketplace",
		"Categories": categories,
		"AppName":    "KolajAI",
	}

	h.RenderTemplate(w, r, "marketplace/categories", data)
}

// Search handles marketplace search page
func (h *MarketplacePageHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	page := 1
	limit := 20

	// Parse page from query params
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	var products []models.Product
	if query != "" {
		// Search products
		var err error
		products, err = h.productService.GetProducts("", query, page, limit)
		if err != nil {
			log.Printf("Error searching products: %v", err)
			products = []models.Product{} // Empty slice on error
		}
	}

	data := map[string]interface{}{
		"Title":    "Arama Sonuçları - KolajAI Marketplace",
		"Products": products,
		"Query":    query,
		"Page":     page,
		"AppName":  "KolajAI",
	}

	h.RenderTemplate(w, r, "marketplace/search", data)
}
