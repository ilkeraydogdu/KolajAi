package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"kolajAi/internal/models"
	"kolajAi/internal/services"
	"kolajAi/internal/session"
	"kolajAi/internal/reporting"
	"kolajAi/internal/notifications"
	"kolajAi/internal/seo"
	"kolajAi/internal/errors"
)

// AdminHandler handles comprehensive admin operations
type AdminHandler struct {
	*Handler
	DB                *sql.DB
	productService    *services.ProductService
	vendorService     *services.VendorService
	orderService      *services.OrderService
	auctionService    *services.AuctionService
	sessionManager    *session.SessionManager
	reportManager     *reporting.ReportManager
	notificationMgr   *notifications.NotificationManager
	seoManager        *seo.SEOManager
	errorManager      *errors.ErrorManager
}

// NewAdminHandler creates a new enhanced admin handler
func NewAdminHandler(h *Handler, productService *services.ProductService, vendorService *services.VendorService, 
	orderService *services.OrderService, auctionService *services.AuctionService, sessionMgr *session.SessionManager,
	reportMgr *reporting.ReportManager, notificationMgr *notifications.NotificationManager, 
	seoMgr *seo.SEOManager, errorMgr *errors.ErrorManager) *AdminHandler {
	return &AdminHandler{
		Handler:           h,
		productService:    productService,
		vendorService:     vendorService,
		orderService:      orderService,
		auctionService:    auctionService,
		sessionManager:    sessionMgr,
		reportManager:     reportMgr,
		notificationMgr:   notificationMgr,
		seoManager:        seoMgr,
		errorManager:      errorMgr,
	}
}

// AdminDashboard shows comprehensive admin dashboard
func (h *AdminHandler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	if !h.IsAdmin(r) {
		h.RedirectWithFlash(w, r, "/login", "Admin yetkisi gerekli")
		return
	}

	data := h.GetTemplateData()

	// Comprehensive dashboard statistics
	stats := map[string]interface{}{
		"users": map[string]interface{}{
			"total":          h.getTotalUsers(),
			"active_today":   h.getActiveUsersToday(),
			"new_this_week":  h.getNewUsersThisWeek(),
			"growth_rate":    h.getUserGrowthRate(),
		},
		"products": map[string]interface{}{
			"total":           h.getTotalProducts(),
			"published":       h.getPublishedProducts(),
			"out_of_stock":    h.getOutOfStockProducts(),
			"low_stock":       h.getLowStockProducts(),
		},
		"orders": map[string]interface{}{
			"total":           h.getTotalOrders(),
			"pending":         h.getPendingOrders(),
			"completed_today": h.getCompletedOrdersToday(),
			"revenue_today":   h.getRevenueToday(),
		},
		"revenue": map[string]interface{}{
			"total":         h.getTotalRevenue(),
			"this_month":    h.getRevenueThisMonth(),
			"last_month":    h.getRevenueLastMonth(),
			"growth_rate":   h.getRevenueGrowthRate(),
		},
		"sessions":   h.getSessionStats(),
		"errors":     h.getErrorStats(),
		"seo":        h.getSEOStats(),
		"system":     h.getSystemHealth(),
	}

	// Recent activities
	activities := h.getRecentActivities(10)
	
	// Performance metrics
	performanceMetrics := h.getPerformanceMetrics()

	data["Stats"] = stats
	data["Activities"] = activities
	data["Performance"] = performanceMetrics
	data["LastUpdated"] = time.Now()

	h.RenderTemplate(w, r, "admin/dashboard", data)
}

// API endpoint for dashboard data
func (h *AdminHandler) AdminDashboardAPI(w http.ResponseWriter, r *http.Request) {
	if !h.IsAdmin(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := map[string]interface{}{
		"users": map[string]interface{}{
			"total":          h.getTotalUsers(),
			"active_today":   h.getActiveUsersToday(),
			"new_this_week":  h.getNewUsersThisWeek(),
			"growth_rate":    h.getUserGrowthRate(),
		},
		"products": map[string]interface{}{
			"total":           h.getTotalProducts(),
			"published":       h.getPublishedProducts(),
			"out_of_stock":    h.getOutOfStockProducts(),
			"low_stock":       h.getLowStockProducts(),
		},
		"orders": map[string]interface{}{
			"total":           h.getTotalOrders(),
			"pending":         h.getPendingOrders(),
			"completed_today": h.getCompletedOrdersToday(),
			"revenue_today":   h.getRevenueToday(),
		},
		"revenue": map[string]interface{}{
			"total":         h.getTotalRevenue(),
			"this_month":    h.getRevenueThisMonth(),
			"last_month":    h.getRevenueLastMonth(),
			"growth_rate":   h.getRevenueGrowthRate(),
		},
		"sessions":   h.getSessionStats(),
		"errors":     h.getErrorStats(),
		"seo":        h.getSEOStats(),
		"system":     h.getSystemHealth(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
		"timestamp": time.Now(),
	})
}

// AdminProducts shows enhanced product management
func (h *AdminHandler) AdminProducts(w http.ResponseWriter, r *http.Request) {
	if !h.IsAdmin(r) {
		h.RedirectWithFlash(w, r, "/login", "Admin yetkisi gerekli")
		return
	}

	switch r.Method {
	case "GET":
		h.showProductsPage(w, r)
	case "POST":
		h.createProduct(w, r)
	case "PUT":
		h.updateProduct(w, r)
	case "DELETE":
		h.deleteProduct(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *AdminHandler) showProductsPage(w http.ResponseWriter, r *http.Request) {
	data := h.GetTemplateData()

	// Pagination
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	limit := 20
	offset := (page - 1) * limit

	// Filters
	search := r.URL.Query().Get("search")
	category := r.URL.Query().Get("category")
	status := r.URL.Query().Get("status")
	sortBy := r.URL.Query().Get("sort_by")
	sortOrder := r.URL.Query().Get("sort_order")

	// Get products with filters
	products, total, err := h.getProductsWithFilters(limit, offset, search, category, status, sortBy, sortOrder)
	if err != nil {
		h.errorManager.HandleError(r.Context(), err, errors.ErrorTypeDatabase, errors.SeverityHigh)
		h.HandleError(w, r, err, "Ürünler yüklenirken hata oluştu")
		return
	}

	// Get categories for filter dropdown
	categories, _ := h.productService.GetAllCategories()

	data["Products"] = products
	data["Categories"] = categories
	data["CurrentPage"] = page
	data["TotalPages"] = (total + limit - 1) / limit
	data["Total"] = total
	data["Filters"] = map[string]string{
		"search":     search,
		"category":   category,
		"status":     status,
		"sort_by":    sortBy,
		"sort_order": sortOrder,
	}

	h.RenderTemplate(w, r, "admin/products", data)
}

// AdminProductEdit shows enhanced product edit form
func (h *AdminHandler) AdminProductEdit(w http.ResponseWriter, r *http.Request) {
	if !h.IsAdmin(r) {
		h.RedirectWithFlash(w, r, "/login", "Admin yetkisi gerekli")
		return
	}

	if r.Method == "GET" {
		h.showProductEditForm(w, r)
	} else if r.Method == "POST" {
		h.updateProductFromForm(w, r)
	}
}

func (h *AdminHandler) showProductEditForm(w http.ResponseWriter, r *http.Request) {
	data := h.GetTemplateData()

	// Get product ID
	idStr := r.URL.Path[len("/admin/products/edit/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.HandleError(w, r, err, "Geçersiz ürün ID")
		return
	}

	// Get product
	product, err := h.productService.GetProductByID(id)
	if err != nil {
		h.HandleError(w, r, err, "Ürün bulunamadı")
		return
	}

	// Get categories
	categories, err := h.productService.GetAllCategories()
	if err == nil {
		data["Categories"] = categories
	}

	// Get product SEO data
	if h.seoManager != nil {
		seoData := h.getProductSEOData(id)
		data["SEO"] = seoData
	}

	// Get product analytics
	analytics := h.getProductAnalytics(id)
	data["Analytics"] = analytics

	data["Product"] = product
	h.RenderTemplate(w, r, "admin/product_edit", data)
}

// AdminUsers shows comprehensive user management
func (h *AdminHandler) AdminUsers(w http.ResponseWriter, r *http.Request) {
	if !h.IsAdmin(r) {
		h.RedirectWithFlash(w, r, "/login", "Admin yetkisi gerekli")
		return
	}

	data := h.GetTemplateData()

	// Pagination and filters
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	limit := 20
	offset := (page - 1) * limit

	search := r.URL.Query().Get("search")
	role := r.URL.Query().Get("role")
	status := r.URL.Query().Get("status")

	// Get users with filters
	users, total, err := h.getUsersWithFilters(limit, offset, search, role, status)
	if err != nil {
		h.HandleError(w, r, err, "Kullanıcılar yüklenirken hata oluştu")
		return
	}

	data["Users"] = users
	data["CurrentPage"] = page
	data["TotalPages"] = (total + limit - 1) / limit
	data["Total"] = total
	data["Filters"] = map[string]string{
		"search": search,
		"role":   role,
		"status": status,
	}

	h.RenderTemplate(w, r, "admin/users", data)
}

// AdminUserDetail shows detailed user information
func (h *AdminHandler) AdminUserDetail(w http.ResponseWriter, r *http.Request) {
	if !h.IsAdmin(r) {
		h.RedirectWithFlash(w, r, "/login", "Admin yetkisi gerekli")
		return
	}

	data := h.GetTemplateData()

	// Get user ID
	idStr := r.URL.Path[len("/admin/users/detail/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.HandleError(w, r, err, "Geçersiz kullanıcı ID")
		return
	}

	// Get user behavior report
	if h.reportManager != nil {
		userReport, err := h.reportManager.GetUserBehaviorReport(id)
		if err != nil {
			log.Printf("Error getting user behavior report: %v", err)
		} else {
			data["UserReport"] = userReport
		}
	}

	// Get user sessions
	if h.sessionManager != nil {
		sessions, err := h.sessionManager.GetUserSessions(id)
		if err != nil {
			log.Printf("Error getting user sessions: %v", err)
		} else {
			data["Sessions"] = sessions
		}
	}

	h.RenderTemplate(w, r, "admin/user_detail", data)
}

// AdminReports shows comprehensive reporting dashboard
func (h *AdminHandler) AdminReports(w http.ResponseWriter, r *http.Request) {
	if !h.IsAdmin(r) {
		h.RedirectWithFlash(w, r, "/login", "Admin yetkisi gerekli")
		return
	}

	data := h.GetTemplateData()

	reportType := r.URL.Query().Get("type")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if reportType != "" {
		// Generate specific report
		report := h.generateReport(reportType, startDate, endDate)
		data["Report"] = report
		data["ReportType"] = reportType
	}

	// Available report types
	data["ReportTypes"] = []map[string]string{
		{"value": "sales", "label": "Satış Raporu"},
		{"value": "users", "label": "Kullanıcı Raporu"},
		{"value": "products", "label": "Ürün Raporu"},
		{"value": "inventory", "label": "Envanter Raporu"},
		{"value": "user_behavior", "label": "Kullanıcı Davranış Raporu"},
		{"value": "seo", "label": "SEO Raporu"},
		{"value": "performance", "label": "Performans Raporu"},
	}

	h.RenderTemplate(w, r, "admin/reports", data)
}

// AdminSEO shows SEO management interface
func (h *AdminHandler) AdminSEO(w http.ResponseWriter, r *http.Request) {
	if !h.IsAdmin(r) {
		h.RedirectWithFlash(w, r, "/login", "Admin yetkisi gerekli")
		return
	}

	data := h.GetTemplateData()

	if h.seoManager != nil {
		// Get SEO overview
		seoOverview := map[string]interface{}{
			"total_pages":     h.getTotalSEOPages(),
			"indexed_pages":   h.getIndexedPages(),
			"avg_seo_score":   h.getAverageSEOScore(),
			"issues_count":    h.getSEOIssuesCount(),
			"sitemap_status":  h.getSitemapStatus(),
			"robots_status":   h.getRobotsStatus(),
		}

		// Get recent SEO activities
		seoActivities := h.getRecentSEOActivities(10)

		data["SEOOverview"] = seoOverview
		data["SEOActivities"] = seoActivities
	}

	h.RenderTemplate(w, r, "admin/seo", data)
}

// AdminNotifications shows notification management
func (h *AdminHandler) AdminNotifications(w http.ResponseWriter, r *http.Request) {
	if !h.IsAdmin(r) {
		h.RedirectWithFlash(w, r, "/login", "Admin yetkisi gerekli")
		return
	}

	data := h.GetTemplateData()

	if h.notificationMgr != nil {
		// Get notification statistics
		startDate := time.Now().AddDate(0, 0, -30)
		endDate := time.Now()
		
		stats, err := h.notificationMgr.GetNotificationStats(startDate, endDate)
		if err != nil {
			log.Printf("Failed to get notification stats: %v", err)
		} else {
			data["NotificationStats"] = stats
		}
	}

	h.RenderTemplate(w, r, "admin/notifications", data)
}

// AdminSystem shows system management interface
func (h *AdminHandler) AdminSystem(w http.ResponseWriter, r *http.Request) {
	if !h.IsAdmin(r) {
		h.RedirectWithFlash(w, r, "/login", "Admin yetkisi gerekli")
		return
	}

	data := h.GetTemplateData()

	// System health
	systemHealth := h.getSystemHealth()
	data["SystemHealth"] = systemHealth

	// Error statistics
	if h.errorManager != nil {
		errorStats, err := h.errorManager.GetErrorStats(24 * time.Hour)
		if err != nil {
			log.Printf("Failed to get error stats: %v", err)
		} else {
			data["ErrorStats"] = errorStats
		}
	}

	// System logs (recent)
	systemLogs := h.getRecentSystemLogs(50)
	data["SystemLogs"] = systemLogs

	h.RenderTemplate(w, r, "admin/system", data)
}

// AdminSettings shows enhanced system settings
func (h *AdminHandler) AdminSettings(w http.ResponseWriter, r *http.Request) {
	if !h.IsAdmin(r) {
		h.RedirectWithFlash(w, r, "/login", "Admin yetkisi gerekli")
		return
	}

	data := h.GetTemplateData()

	// Get all system settings
	settings := h.getAllSystemSettings()
	data["Settings"] = settings

	// Get language settings
	languages := h.getSupportedLanguages()
	data["Languages"] = languages

	// Get SEO settings
	if h.seoManager != nil {
		seoSettings := h.getSEOSettings()
		data["SEOSettings"] = seoSettings
	}

	h.RenderTemplate(w, r, "admin/settings", data)
}

// Helper methods for statistics
func (h *AdminHandler) getTotalUsers() int {
	var count int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Printf("Error getting total users: %v", err)
		return 0
	}
	return count
}

func (h *AdminHandler) getActiveUsersToday() int {
	var count int
	err := h.DB.QueryRow(`
		SELECT COUNT(DISTINCT user_id) 
		FROM sessions 
		WHERE DATE(last_activity) = CURDATE() AND is_active = TRUE
	`).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (h *AdminHandler) getNewUsersThisWeek() int {
	var count int
	err := h.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM users 
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
	`).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (h *AdminHandler) getTotalProducts() int {
	if count, err := h.productService.GetProductCount(); err == nil {
		return int(count)
	}
	return 0
}

func (h *AdminHandler) getPublishedProducts() int {
	var count int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM products WHERE status = 'published'").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (h *AdminHandler) getTotalOrders() int {
	var count int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM orders").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (h *AdminHandler) getTotalRevenue() float64 {
	var revenue sql.NullFloat64
	err := h.DB.QueryRow("SELECT SUM(total_amount) FROM orders WHERE status = 'completed'").Scan(&revenue)
	if err != nil || !revenue.Valid {
		return 0.0
	}
	return revenue.Float64
}

func (h *AdminHandler) getRevenueThisMonth() float64 {
	var revenue sql.NullFloat64
	err := h.DB.QueryRow(`
		SELECT SUM(total_amount) 
		FROM orders 
		WHERE status = 'completed' AND MONTH(created_at) = MONTH(NOW()) AND YEAR(created_at) = YEAR(NOW())
	`).Scan(&revenue)
	if err != nil || !revenue.Valid {
		return 0.0
	}
	return revenue.Float64
}

func (h *AdminHandler) getSessionStats() map[string]interface{} {
	if h.sessionManager == nil {
		return map[string]interface{}{}
	}
	
	stats, err := h.sessionManager.GetSessionStats()
	if err != nil {
		log.Printf("Error getting session stats: %v", err)
		return map[string]interface{}{}
	}
	return stats
}

func (h *AdminHandler) getErrorStats() map[string]interface{} {
	if h.errorManager == nil {
		return map[string]interface{}{}
	}
	
	stats, err := h.errorManager.GetErrorStats(24 * time.Hour)
	if err != nil {
		log.Printf("Error getting error stats: %v", err)
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"total_errors": stats.TotalErrors,
		"by_type":      stats.ErrorsByType,
		"by_severity":  stats.ErrorsBySeverity,
	}
}

func (h *AdminHandler) getSEOStats() map[string]interface{} {
	return map[string]interface{}{
		"total_pages":   h.getTotalSEOPages(),
		"avg_score":     h.getAverageSEOScore(),
		"issues_count":  h.getSEOIssuesCount(),
	}
}

func (h *AdminHandler) getSystemHealth() map[string]interface{} {
	return map[string]interface{}{
		"database":    h.checkDatabaseHealth(),
		"memory":      h.getMemoryUsage(),
		"disk":        h.getDiskUsage(),
		"cpu":         h.getCPUUsage(),
		"uptime":      h.getSystemUptime(),
	}
}

// Placeholder implementations for new methods
func (h *AdminHandler) getUserGrowthRate() float64 { return 0.0 }
func (h *AdminHandler) getOutOfStockProducts() int { return 0 }
func (h *AdminHandler) getLowStockProducts() int { return 0 }
func (h *AdminHandler) getPendingOrders() int { return 0 }
func (h *AdminHandler) getCompletedOrdersToday() int { return 0 }
func (h *AdminHandler) getRevenueToday() float64 { return 0.0 }
func (h *AdminHandler) getRevenueLastMonth() float64 { return 0.0 }
func (h *AdminHandler) getRevenueGrowthRate() float64 { return 0.0 }
func (h *AdminHandler) getRecentActivities(limit int) []map[string]interface{} { return []map[string]interface{}{} }
func (h *AdminHandler) getPerformanceMetrics() map[string]interface{} { return map[string]interface{}{} }
func (h *AdminHandler) getProductsWithFilters(limit, offset int, search, category, status, sortBy, sortOrder string) ([]*models.Product, int, error) {
	products, err := h.productService.GetAllProducts(limit, offset)
	// Convert []models.Product to []*models.Product
	productPtrs := make([]*models.Product, len(products))
	for i := range products {
		productPtrs[i] = &products[i]
	}
	return productPtrs, len(products), err
}
func (h *AdminHandler) getUsersWithFilters(limit, offset int, search, role, status string) ([]*models.User, int, error) {
	return []*models.User{}, 0, nil
}
func (h *AdminHandler) getProductSEOData(productID int) map[string]interface{} { return map[string]interface{}{} }
func (h *AdminHandler) getProductAnalytics(productID int) map[string]interface{} { return map[string]interface{}{} }
func (h *AdminHandler) generateReport(reportType, startDate, endDate string) map[string]interface{} { return map[string]interface{}{} }
func (h *AdminHandler) getTotalSEOPages() int { return 0 }
func (h *AdminHandler) getIndexedPages() int { return 0 }
func (h *AdminHandler) getAverageSEOScore() float64 { return 0.0 }
func (h *AdminHandler) getSEOIssuesCount() int { return 0 }
func (h *AdminHandler) getSitemapStatus() string { return "active" }
func (h *AdminHandler) getRobotsStatus() string { return "active" }
func (h *AdminHandler) getRecentSEOActivities(limit int) []map[string]interface{} { return []map[string]interface{}{} }
func (h *AdminHandler) getRecentSystemLogs(limit int) []map[string]interface{} { return []map[string]interface{}{} }
func (h *AdminHandler) getAllSystemSettings() map[string]interface{} { return map[string]interface{}{} }
func (h *AdminHandler) getSupportedLanguages() []map[string]string { return []map[string]string{} }
func (h *AdminHandler) getSEOSettings() map[string]interface{} { return map[string]interface{}{} }
func (h *AdminHandler) checkDatabaseHealth() string { return "healthy" }
func (h *AdminHandler) getMemoryUsage() float64 { return 0.0 }
func (h *AdminHandler) getDiskUsage() float64 { return 0.0 }
func (h *AdminHandler) getCPUUsage() float64 { return 0.0 }
func (h *AdminHandler) getSystemUptime() string { return "0h" }

// CRUD operations for products
func (h *AdminHandler) createProduct(w http.ResponseWriter, r *http.Request) {
	// Implementation for creating product
}

func (h *AdminHandler) updateProduct(w http.ResponseWriter, r *http.Request) {
	// Implementation for updating product via API
}

func (h *AdminHandler) updateProductFromForm(w http.ResponseWriter, r *http.Request) {
	// Get product ID
	idStr := r.URL.Path[len("/admin/products/edit/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.HandleError(w, r, err, "Geçersiz ürün ID")
		return
	}

	// Form data processing
	product := &models.Product{
		ID:          id,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		ShortDesc:   r.FormValue("short_desc"),
		SKU:         r.FormValue("sku"),
		Status:      r.FormValue("status"),
		Tags:        r.FormValue("tags"),
	}

	// Parse numeric fields
	if priceStr := r.FormValue("price"); priceStr != "" {
		if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
			product.Price = price
		}
	}

	if comparePriceStr := r.FormValue("compare_price"); comparePriceStr != "" {
		if comparePrice, err := strconv.ParseFloat(comparePriceStr, 64); err == nil {
			product.ComparePrice = comparePrice
		}
	}

	if stockStr := r.FormValue("stock"); stockStr != "" {
		if stock, err := strconv.Atoi(stockStr); err == nil {
			product.Stock = stock
		}
	}

	if categoryIDStr := r.FormValue("category_id"); categoryIDStr != "" {
		if categoryID, err := strconv.Atoi(categoryIDStr); err == nil {
			product.CategoryID = categoryID
		}
	}

	product.IsFeatured = r.FormValue("is_featured") == "on"

	// Update product
	err = h.productService.UpdateProduct(id, product)
	if err != nil {
		h.errorManager.HandleError(r.Context(), err, errors.ErrorTypeApplication, errors.SeverityHigh)
		h.HandleError(w, r, err, "Ürün güncellenirken hata oluştu")
		return
	}

	// Update SEO data if available
	if h.seoManager != nil {
		h.updateProductSEO(id, r)
	}

	h.RedirectWithFlash(w, r, "/admin/products", "Ürün başarıyla güncellendi")
}

func (h *AdminHandler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	// Implementation for deleting product
}

func (h *AdminHandler) updateProductSEO(productID int, r *http.Request) {
	// Update SEO data for product
	seoTitle := r.FormValue("seo_title")
	seoDescription := r.FormValue("seo_description")
	seoKeywords := r.FormValue("seo_keywords")
	
	if seoTitle != "" {
		h.seoManager.SetTranslation("product", productID, "tr", "title", seoTitle)
	}
	if seoDescription != "" {
		h.seoManager.SetTranslation("product", productID, "tr", "description", seoDescription)
	}
	if seoKeywords != "" {
		h.seoManager.SetTranslation("product", productID, "tr", "keywords", seoKeywords)
	}
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

	idStr := r.FormValue("vendor_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.HandleError(w, r, err, "Geçersiz satıcı ID")
		return
	}

	err = h.vendorService.ApproveVendor(id)
	if err != nil {
		h.HandleError(w, r, err, "Satıcı onaylanırken hata oluştu")
		return
	}

	h.RedirectWithFlash(w, r, "/admin/vendors", "Satıcı başarıyla onaylandı")
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
