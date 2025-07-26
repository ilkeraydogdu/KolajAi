package handlers

import (
	"net/http"
	"strconv"

	"kolajAi/internal/models"
	"kolajAi/internal/services"
)

// AdminHandler handles admin-related requests
type AdminHandler struct {
	*Handler
	productService *services.ProductService
	vendorService  *services.VendorService
	orderService   *services.OrderService
	auctionService *services.AuctionService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(h *Handler, productService *services.ProductService, vendorService *services.VendorService, orderService *services.OrderService, auctionService *services.AuctionService) *AdminHandler {
	return &AdminHandler{
		Handler:        h,
		productService: productService,
		vendorService:  vendorService,
		orderService:   orderService,
		auctionService: auctionService,
	}
}

// AdminDashboard shows the main admin dashboard
func (h *AdminHandler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	if !h.IsAdmin(r) {
		h.RedirectWithFlash(w, r, "/login", "Admin yetkisi gerekli")
		return
	}

	data := h.GetTemplateData()

	// İstatistikler
	stats := make(map[string]interface{})

	// Ürün sayısı
	if productCount, err := h.productService.GetProductCount(); err == nil {
		stats["ProductCount"] = productCount
	}

	// Satıcı sayısı
	if vendorCount, err := h.vendorService.GetVendorCount(); err == nil {
		stats["VendorCount"] = vendorCount
	}

	// Aktif açık artırma sayısı
	if auctionCount, err := h.auctionService.GetActiveAuctionCount(); err == nil {
		stats["ActiveAuctionCount"] = auctionCount
	}

	// Son eklenen ürünler
	if recentProducts, err := h.productService.GetRecentProducts(5); err == nil {
		stats["RecentProducts"] = recentProducts
	}

	// Bekleyen satıcılar
	if pendingVendors, err := h.vendorService.GetPendingVendors(); err == nil {
		stats["PendingVendors"] = pendingVendors
	}

	data["Stats"] = stats
	h.RenderTemplate(w, r, "admin/dashboard", data)
}

// AdminProducts shows product management page
func (h *AdminHandler) AdminProducts(w http.ResponseWriter, r *http.Request) {
	if !h.IsAdmin(r) {
		h.RedirectWithFlash(w, r, "/login", "Admin yetkisi gerekli")
		return
	}

	data := h.GetTemplateData()

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	limit := 20
	offset := (page - 1) * limit

	// Ürünleri getir
	products, err := h.productService.GetAllProducts(limit, offset)
	if err != nil {
		h.HandleError(w, r, err, "Ürünler yüklenirken hata oluştu")
		return
	}

	data["Products"] = products
	data["CurrentPage"] = page
	h.RenderTemplate(w, r, "admin/products", data)
}

// AdminProductEdit shows product edit form
func (h *AdminHandler) AdminProductEdit(w http.ResponseWriter, r *http.Request) {
	if !h.IsAdmin(r) {
		h.RedirectWithFlash(w, r, "/login", "Admin yetkisi gerekli")
		return
	}

	if r.Method == "GET" {
		h.showProductEditForm(w, r)
	} else if r.Method == "POST" {
		h.updateProduct(w, r)
	}
}

func (h *AdminHandler) showProductEditForm(w http.ResponseWriter, r *http.Request) {
	data := h.GetTemplateData()

	// Ürün ID'sini al
	idStr := r.URL.Path[len("/admin/products/edit/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.HandleError(w, r, err, "Geçersiz ürün ID")
		return
	}

	// Ürünü getir
	product, err := h.productService.GetProductByID(id)
	if err != nil {
		h.HandleError(w, r, err, "Ürün bulunamadı")
		return
	}

	// Kategorileri getir
	categories, err := h.productService.GetAllCategories()
	if err == nil {
		data["Categories"] = categories
	}

	data["Product"] = product
	h.RenderTemplate(w, r, "admin/product_edit", data)
}

func (h *AdminHandler) updateProduct(w http.ResponseWriter, r *http.Request) {
	// Ürün ID'sini al
	idStr := r.URL.Path[len("/admin/products/edit/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.HandleError(w, r, err, "Geçersiz ürün ID")
		return
	}

	// Form verilerini al
	product := &models.Product{
		ID:          id,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		ShortDesc:   r.FormValue("short_desc"),
		SKU:         r.FormValue("sku"),
		Status:      r.FormValue("status"),
		Tags:        r.FormValue("tags"),
	}

	// Fiyat
	if priceStr := r.FormValue("price"); priceStr != "" {
		if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
			product.Price = price
		}
	}

	// Karşılaştırma fiyatı
	if comparePriceStr := r.FormValue("compare_price"); comparePriceStr != "" {
		if comparePrice, err := strconv.ParseFloat(comparePriceStr, 64); err == nil {
			product.ComparePrice = comparePrice
		}
	}

	// Stok
	if stockStr := r.FormValue("stock"); stockStr != "" {
		if stock, err := strconv.Atoi(stockStr); err == nil {
			product.Stock = stock
		}
	}

	// Kategori ID
	if categoryIDStr := r.FormValue("category_id"); categoryIDStr != "" {
		if categoryID, err := strconv.Atoi(categoryIDStr); err == nil {
			product.CategoryID = categoryID
		}
	}

	// Öne çıkan ürün
	product.IsFeatured = r.FormValue("is_featured") == "on"

	// Ürünü güncelle
	err = h.productService.UpdateProduct(id, product)
	if err != nil {
		h.HandleError(w, r, err, "Ürün güncellenirken hata oluştu")
		return
	}

	h.RedirectWithFlash(w, r, "/admin/products", "Ürün başarıyla güncellendi")
}

// AdminVendors shows vendor management page
func (h *AdminHandler) AdminVendors(w http.ResponseWriter, r *http.Request) {
	if !h.IsAdmin(r) {
		h.RedirectWithFlash(w, r, "/login", "Admin yetkisi gerekli")
		return
	}

	data := h.GetTemplateData()

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	limit := 20
	offset := (page - 1) * limit

	// Satıcıları getir
	vendors, err := h.vendorService.GetAllVendors(limit, offset)
	if err != nil {
		h.HandleError(w, r, err, "Satıcılar yüklenirken hata oluştu")
		return
	}

	data["Vendors"] = vendors
	data["CurrentPage"] = page
	h.RenderTemplate(w, r, "admin/vendors", data)
}

// AdminVendorApprove approves a vendor
func (h *AdminHandler) AdminVendorApprove(w http.ResponseWriter, r *http.Request) {
	if !h.IsAdmin(r) {
		h.RedirectWithFlash(w, r, "/login", "Admin yetkisi gerekli")
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Satıcı ID'sini al
	idStr := r.FormValue("vendor_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.HandleError(w, r, err, "Geçersiz satıcı ID")
		return
	}

	// Satıcıyı onayla
	err = h.vendorService.ApproveVendor(id)
	if err != nil {
		h.HandleError(w, r, err, "Satıcı onaylanırken hata oluştu")
		return
	}

	h.RedirectWithFlash(w, r, "/admin/vendors", "Satıcı başarıyla onaylandı")
}

// AdminSettings shows system settings
func (h *AdminHandler) AdminSettings(w http.ResponseWriter, r *http.Request) {
	if !h.IsAdmin(r) {
		h.RedirectWithFlash(w, r, "/login", "Admin yetkisi gerekli")
		return
	}

	data := h.GetTemplateData()

	// Sistem ayarları burada yüklenebilir
	settings := map[string]interface{}{
		"SiteName":        "KolajAI Marketplace",
		"MaintenanceMode": false,
		"MaxUploadSize":   "10MB",
		"DefaultCurrency": "TRY",
	}

	data["Settings"] = settings
	h.RenderTemplate(w, r, "admin/settings", data)
}

// IsAdmin checks if the current user is an admin
func (h *Handler) IsAdmin(r *http.Request) bool {
	session, _ := h.SessionManager.GetSession(r)
	userID, ok := session.Values["user_id"]
	if !ok {
		return false
	}

	isAdmin, ok := session.Values["is_admin"]
	if !ok {
		return false
	}

	return userID != nil && isAdmin == true
}
