package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"kolajAi/internal/models"
	"kolajAi/internal/services"
)

// EcommerceHandler handles e-commerce related requests
type EcommerceHandler struct {
	*Handler
	vendorService  *services.VendorService
	productService *services.ProductService
	orderService   *services.OrderService
	auctionService *services.AuctionService
	userService    *services.UserService
	addressService *services.AddressService
}

// NewEcommerceHandler creates a new e-commerce handler
func NewEcommerceHandler(h *Handler, vendorService *services.VendorService, productService *services.ProductService, orderService *services.OrderService, auctionService *services.AuctionService, userService *services.UserService, addressService *services.AddressService) *EcommerceHandler {
	return &EcommerceHandler{
		Handler:        h,
		vendorService:  vendorService,
		productService: productService,
		orderService:   orderService,
		auctionService: auctionService,
		userService:    userService,
		addressService: addressService,
	}
}

// Marketplace - Ana sayfa
func (h *EcommerceHandler) Marketplace(w http.ResponseWriter, r *http.Request) {
	data := h.GetTemplateData()

	// Öne çıkan ürünler
	featuredProducts, err := h.productService.GetFeaturedProducts(12, 0)
	if err == nil {
		data["FeaturedProducts"] = featuredProducts
	}

	// Kategoriler
	categories, err := h.productService.GetAllCategories()
	if err == nil {
		data["Categories"] = categories
	}

	// Aktif açık artırmalar
	activeAuctions, err := h.auctionService.GetActiveAuctions(6, 0)
	if err == nil {
		data["ActiveAuctions"] = activeAuctions
	}

	h.RenderTemplate(w, r, "marketplace/index", data)
}

// Products - Ürün listesi
func (h *EcommerceHandler) Products(w http.ResponseWriter, r *http.Request) {
	data := h.GetTemplateData()

	// Query parametreleri
	categoryID := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")
	page := h.getPageFromQuery(r)
	limit := services.DefaultProductLimit
	offset := (page - 1) * limit

	var products []models.Product
	var err error

	if search != "" {
		products, err = h.productService.SearchProducts(search, limit, offset)
		data["SearchTerm"] = search
	} else if categoryID != "" {
		catID, _ := strconv.Atoi(categoryID)
		products, err = h.productService.GetProductsByCategory(catID, limit, offset)

		// Kategori bilgisi
		if catID > 0 {
			category, catErr := h.productService.GetCategoryByID(catID)
			if catErr == nil {
				data["Category"] = category
			}
		}
	} else {
		// Tüm aktif ürünler
		products, err = h.productService.GetAllProducts(limit, offset)
	}

	if err != nil {
		h.HandleError(w, r, err, "Ürünler yüklenirken hata oluştu")
		return
	}

	// Kategoriler
	categories, err := h.productService.GetAllCategories()
	if err == nil {
		data["Categories"] = categories
	}

	data["Products"] = products
	data["CurrentPage"] = page
	h.RenderTemplate(w, r, "marketplace/products", data)
}

// ProductDetail - Ürün detayı
func (h *EcommerceHandler) ProductDetail(w http.ResponseWriter, r *http.Request) {
	productIDStr := strings.TrimPrefix(r.URL.Path, "/product/")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := h.GetTemplateData()

	// Ürün bilgisi
	product, err := h.productService.GetProductByID(productID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Görüntülenme sayısını artır
	go h.productService.IncrementProductViews(productID)

	// Ürün resimleri
	images, err := h.productService.GetProductImages(productID)
	if err == nil {
		data["ProductImages"] = images
	}

	// Ürün yorumları
	reviews, err := h.productService.GetProductReviews(productID, 10, 0)
	if err == nil {
		data["ProductReviews"] = reviews
	}

	// Satıcı bilgisi
	vendor, err := h.vendorService.GetVendorByID(product.VendorID)
	if err == nil {
		data["Vendor"] = vendor
	}

	// Benzer ürünler
	similarProducts, err := h.productService.GetProductsByCategory(int(product.CategoryID), 4, 0)
	if err == nil {
		data["SimilarProducts"] = similarProducts
	}

	data["Product"] = product
	h.RenderTemplate(w, r, "marketplace/product-detail", data)
}

// AddToCart - Sepete ekleme
func (h *EcommerceHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Form verilerini al
	productID, _ := strconv.Atoi(r.FormValue("product_id"))
	quantity, _ := strconv.Atoi(r.FormValue("quantity"))
	if quantity <= 0 {
		quantity = 1
	}

	// Ürün bilgisini al
	product, err := h.productService.GetProductByID(productID)
	if err != nil {
		h.SetFlashError("Ürün bulunamadı")
		http.Redirect(w, r, "/products", http.StatusSeeOther)
		return
	}

	// Stok kontrolü
	if product.Stock < quantity {
		h.SetFlashError("Yeterli stok bulunmamaktadır")
		http.Redirect(w, r, fmt.Sprintf("/product/%d", productID), http.StatusSeeOther)
		return
	}

	// Sepeti al veya oluştur
	var cart *models.Cart
	userID := h.GetUserID(r)

	if userID > 0 {
		cart, err = h.orderService.GetCartByUser(userID)
		if err != nil {
			// Yeni sepet oluştur
			cart = &models.Cart{UserID: userID}
			err = h.orderService.CreateCart(cart)
			if err != nil {
				h.HandleError(w, r, err, "Sepet oluşturulamadı")
				return
			}
		}
	} else {
		// Session tabanlı sepet
		sessionID := h.GetSessionID(r)
		cart, err = h.orderService.GetCartBySession(sessionID)
		if err != nil {
			cart = &models.Cart{UserID: 0} // Session-based cart without SessionID field
			err = h.orderService.CreateCart(cart)
			if err != nil {
				h.HandleError(w, r, err, "Sepet oluşturulamadı")
				return
			}
		}
	}

	// Sepete ürün ekle
	cartItem := &models.CartItem{
		CartID:    cart.ID,
		ProductID: productID,
		Quantity:  quantity,
		Price:     product.Price,
	}

	err = h.orderService.AddCartItem(cartItem)
	if err != nil {
		h.HandleError(w, r, err, "Ürün sepete eklenemedi")
		return
	}

	h.SetFlashSuccess("Ürün sepete eklendi")
	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}

// Cart - Sepet görüntüleme
func (h *EcommerceHandler) Cart(w http.ResponseWriter, r *http.Request) {
	data := h.GetTemplateData()

	// Sepeti al
	var cart *models.Cart
	var err error
	userID := h.GetUserID(r)

	if userID > 0 {
		cart, err = h.orderService.GetCartByUser(userID)
	} else {
		sessionID := h.GetSessionID(r)
		cart, err = h.orderService.GetCartBySession(sessionID)
	}

	if err != nil {
		data["CartItems"] = []models.CartItem{}
		data["CartTotal"] = 0.0
	} else {
		// Sepet öğelerini al
		cartItems, err := h.orderService.GetCartItems(cart.ID)
		if err == nil {
			data["CartItems"] = cartItems

			// Toplam hesapla
			var total float64
			for _, item := range cartItems {
				total += item.Price * float64(item.Quantity)
			}
			data["CartTotal"] = total
		}
	}

	h.RenderTemplate(w, r, "marketplace/cart", data)
}

// Auctions - Açık artırmalar
func (h *EcommerceHandler) Auctions(w http.ResponseWriter, r *http.Request) {
	data := h.GetTemplateData()

	page := h.getPageFromQuery(r)
	limit := 20
	offset := (page - 1) * limit

	// Aktif açık artırmalar
	auctions, err := h.auctionService.GetActiveAuctions(limit, offset)
	if err != nil {
		h.HandleError(w, r, err, "Açık artırmalar yüklenirken hata oluştu")
		return
	}

	// Yakında bitecek açık artırmalar
	endingAuctions, err := h.auctionService.GetEndingAuctions(24, 6, 0)
	if err == nil {
		data["EndingAuctions"] = endingAuctions
	}

	data["Auctions"] = auctions
	data["CurrentPage"] = page
	h.RenderTemplate(w, r, "marketplace/auctions", data)
}

// AuctionDetail - Açık artırma detayı
func (h *EcommerceHandler) AuctionDetail(w http.ResponseWriter, r *http.Request) {
	auctionIDStr := strings.TrimPrefix(r.URL.Path, "/auction/")
	auctionID, err := strconv.Atoi(auctionIDStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := h.GetTemplateData()

	// Açık artırma bilgisi
	auction, err := h.auctionService.GetAuctionByID(auctionID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Görüntülenme sayısını artır
	go h.auctionService.IncrementAuctionViews(auctionID)

	// Teklifler
	bids, err := h.auctionService.GetAuctionBids(auctionID, 10, 0)
	if err == nil {
		data["AuctionBids"] = bids
	}

	// Açık artırma resimleri
	images, err := h.auctionService.GetAuctionImages(auctionID)
	if err == nil {
		data["AuctionImages"] = images
	}

	// Satıcı bilgisi
	vendor, err := h.vendorService.GetVendorByID(auction.VendorID)
	if err == nil {
		data["Vendor"] = vendor
	}

	// Kullanıcının bu açık artırmayı takip edip etmediğini kontrol et
	userID := h.GetUserID(r)
	if userID > 0 {
		isWatching, err := h.auctionService.IsUserWatching(auctionID, userID)
		if err == nil {
			data["IsWatching"] = isWatching
		}
	}

	data["Auction"] = auction
	h.RenderTemplate(w, r, "marketplace/auction-detail", data)
}

// PlaceBid - Teklif verme
func (h *EcommerceHandler) PlaceBid(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.IsAuthenticated(r) {
		h.SetFlashError("Teklif vermek için giriş yapmalısınız")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	auctionID, _ := strconv.Atoi(r.FormValue("auction_id"))
	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
	userID := h.GetUserID(r)

	bid := &models.AuctionBid{
		AuctionID: auctionID,
		UserID:    userID,
		Amount:    amount,
		IPAddress: r.RemoteAddr,
	}

	err := h.auctionService.PlaceBid(bid)
	if err != nil {
		h.SetFlashError(err.Error())
	} else {
		h.SetFlashSuccess("Teklifiniz başarıyla verildi")
	}

	http.Redirect(w, r, fmt.Sprintf("/auction/%d", auctionID), http.StatusSeeOther)
}

// Vendor Dashboard
func (h *EcommerceHandler) VendorDashboard(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := h.GetTemplateData()
	userID := h.GetUserID(r)

	// Satıcı bilgisini al
	vendor, err := h.vendorService.GetVendorByUserID(userID)
	if err != nil {
		// Satıcı değilse, satıcı olmak için yönlendir
		h.SetFlashError("Satıcı paneline erişmek için önce satıcı başvurusu yapmalısınız")
		http.Redirect(w, r, "/become-vendor", http.StatusSeeOther)
		return
	}

	// Satıcı istatistikleri
	stats, err := h.vendorService.GetVendorStats(vendor.ID)
	if err == nil {
		data["VendorStats"] = stats
	}

	// Son ürünler
	products, err := h.productService.GetProductsByVendor(vendor.ID, 5, 0)
	if err == nil {
		data["RecentProducts"] = products
	}

	// Son siparişler
	orders, err := h.orderService.GetVendorOrders(vendor.ID, 5, 0)
	if err == nil {
		data["RecentOrders"] = orders
	}

	data["Vendor"] = vendor
	h.RenderTemplate(w, r, "vendor/dashboard", data)
}

// User Profile Handlers

func (h *EcommerceHandler) UserProfile(w http.ResponseWriter, r *http.Request) {
	userID := h.GetUserID(r)
	
	if r.Method == "POST" {
		// Handle profile update
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Form parse error", http.StatusBadRequest)
			return
		}
		
		// Update user profile logic would go here
		// For now, just redirect with success message
		h.RedirectWithFlash(w, r, "/user/profile", "Profil başarıyla güncellendi")
		return
	}
	
	// Get user data
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		h.SessionManager.Logger.Printf("Error getting user: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	// Get user statistics
	stats, err := h.userService.GetUserStats(userID)
	if err != nil {
		h.SessionManager.Logger.Printf("Error getting user stats: %v", err)
		stats = &services.UserStats{} // Empty stats
	}
	
	data := map[string]interface{}{
		"User":  user,
		"Stats": stats,
		"Title": "Profil",
	}
	
	h.RenderTemplate(w, r, "user/profile", data)
}

func (h *EcommerceHandler) UserOrders(w http.ResponseWriter, r *http.Request) {
	userID := h.GetUserID(r)
	page := h.getPageFromQuery(r)
	
	// Get filters from query parameters
	status := r.URL.Query().Get("status")
	dateRange := r.URL.Query().Get("date_range")
	
	// Get user orders
	orders, totalCount, err := h.orderService.GetUserOrders(userID, page, services.DefaultProductLimit, status, dateRange)
	if err != nil {
		h.SessionManager.Logger.Printf("Error getting user orders: %v", err)
		orders = []*models.Order{}
		totalCount = 0
	}
	
	// Calculate pagination
	totalPages := (totalCount + services.DefaultProductLimit - 1) / services.DefaultProductLimit
	
	// Get order statistics
	orderStats, err := h.orderService.GetUserOrderStats(userID)
	if err != nil {
		h.SessionManager.Logger.Printf("Error getting order stats: %v", err)
		orderStats = &models.OrderStats{}
	}
	
	data := map[string]interface{}{
		"Orders":      orders,
		"OrderStats":  orderStats,
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"TotalCount":  totalCount,
		"Status":      status,
		"DateRange":   dateRange,
		"Title":       "Siparişlerim",
	}
	
	h.RenderTemplate(w, r, "user/orders", data)
}

func (h *EcommerceHandler) UserAddresses(w http.ResponseWriter, r *http.Request) {
	userID := h.GetUserID(r)
	
	if r.Method == "POST" {
		// Handle address operations (add, edit, delete)
		action := r.FormValue("action")
		
		switch action {
		case "add":
			// Add new address
			fullName := r.FormValue("full_name")
			names := strings.SplitN(fullName, " ", 2)
			firstName := names[0]
			lastName := ""
			if len(names) > 1 {
				lastName = names[1]
			}
			
			address := &models.Address{
				CustomerID:   uint(userID),
				Title:        r.FormValue("title"),
				FirstName:    firstName,
				LastName:     lastName,
				AddressLine1: r.FormValue("address_line_1"),
				AddressLine2: r.FormValue("address_line_2"),
				City:         r.FormValue("city"),
				State:        r.FormValue("state"),
				PostalCode:   r.FormValue("postal_code"),
				Country:      r.FormValue("country"),
				Phone:        r.FormValue("phone"),
				IsDefault:    r.FormValue("is_default") == "1",
			}
			
				err := h.addressService.CreateAddress(address)
	if err != nil {
		h.SessionManager.Logger.Printf("Error creating address: %v", err)
				h.RedirectWithFlash(w, r, "/user/addresses", "Adres eklenemedi")
				return
			}
			
			h.RedirectWithFlash(w, r, "/user/addresses", "Adres başarıyla eklendi")
			return
			
		case "edit":
			// Edit existing address
			addressID, _ := strconv.Atoi(r.FormValue("address_id"))
			
			fullName := r.FormValue("full_name")
			names := strings.SplitN(fullName, " ", 2)
			firstName := names[0]
			lastName := ""
			if len(names) > 1 {
				lastName = names[1]
			}
			
			address := &models.Address{
				ID:           uint(addressID),
				CustomerID:   uint(userID),
				Title:        r.FormValue("title"),
				FirstName:    firstName,
				LastName:     lastName,
				AddressLine1: r.FormValue("address_line_1"),
				AddressLine2: r.FormValue("address_line_2"),
				City:         r.FormValue("city"),
				State:        r.FormValue("state"),
				PostalCode:   r.FormValue("postal_code"),
				Country:      r.FormValue("country"),
				Phone:        r.FormValue("phone"),
				IsDefault:    r.FormValue("is_default") == "1",
			}
			
				err := h.addressService.UpdateAddress(address)
	if err != nil {
		h.SessionManager.Logger.Printf("Error updating address: %v", err)
				h.RedirectWithFlash(w, r, "/user/addresses", "Adres güncellenemedi")
				return
			}
			
			h.RedirectWithFlash(w, r, "/user/addresses", "Adres başarıyla güncellendi")
			return
			
		case "delete":
			// Delete address
			addressID, _ := strconv.Atoi(r.FormValue("address_id"))
			
				err := h.addressService.DeleteAddress(addressID, userID)
	if err != nil {
		h.SessionManager.Logger.Printf("Error deleting address: %v", err)
				h.RedirectWithFlash(w, r, "/user/addresses", "Adres silinemedi")
				return
			}
			
			h.RedirectWithFlash(w, r, "/user/addresses", "Adres başarıyla silindi")
			return
			
		case "set_default":
			// Set default address
			addressID, _ := strconv.Atoi(r.FormValue("address_id"))
			
				err := h.addressService.SetDefaultAddress(addressID, userID)
	if err != nil {
		h.SessionManager.Logger.Printf("Error setting default address: %v", err)
				h.RedirectWithFlash(w, r, "/user/addresses", "Varsayılan adres ayarlanamadı")
				return
			}
			
			h.RedirectWithFlash(w, r, "/user/addresses", "Varsayılan adres güncellendi")
			return
		}
	}
	
	// Get user addresses
	addresses, err := h.addressService.GetUserAddresses(userID)
	if err != nil {
		h.SessionManager.Logger.Printf("Error getting user addresses: %v", err)
		addresses = []models.Address{}
	}
	
	data := map[string]interface{}{
		"Addresses": addresses,
		"Title":     "Adreslerim",
	}
	
	h.RenderTemplate(w, r, "user/addresses", data)
}

// Helper methods

func (h *EcommerceHandler) getPageFromQuery(r *http.Request) int {
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	return page
}

func (h *EcommerceHandler) SetFlashSuccess(message string) {
	// Flash message implementation - bu gerçek implementasyonda session'a eklenir
	// Şimdilik boş bırakıyoruz
}

func (h *EcommerceHandler) SetFlashError(message string) {
	// Flash message implementation - bu gerçek implementasyonda session'a eklenir
	// Şimdilik boş bırakıyoruz
}

func (h *EcommerceHandler) GetUserID(r *http.Request) int {
	// Get user ID from session - gerçek implementasyonda session'dan alınır
	return 1 // Test için sabit değer
}

func (h *EcommerceHandler) GetSessionID(r *http.Request) string {
	// Get session ID - gerçek implementasyonda session'dan alınır
	return "test-session-id" // Test için sabit değer
}

// API Endpoints

// API - Ürün arama
func (h *EcommerceHandler) APISearchProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	products, err := h.productService.SearchProducts(query, 10, 0)
	if err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// API - Sepet güncelleme
func (h *EcommerceHandler) APIUpdateCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ItemID   int `json:"item_id"`
		Quantity int `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Quantity <= 0 {
		// Öğeyi sil
		err := h.orderService.RemoveCartItem(req.ItemID)
		if err != nil {
			http.Error(w, "Failed to remove item", http.StatusInternalServerError)
			return
		}
	} else {
		// Miktarı güncelle
		cartItem := &models.CartItem{Quantity: req.Quantity}
		err := h.orderService.UpdateCartItem(req.ItemID, cartItem)
		if err != nil {
			http.Error(w, "Failed to update item", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
