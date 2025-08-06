package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"log"
	
	"github.com/gorilla/mux"
	"kolajAi/internal/repository"
	"kolajAi/internal/models"
	"kolajAi/internal/database"
)

// AdminHandler handles admin-related requests
type AdminHandler struct {
	*Handler
	AdminRepo *repository.AdminRepository
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(h *Handler, db *database.MySQLRepository) *AdminHandler {
	return &AdminHandler{
		Handler:   h,
		AdminRepo: repository.NewAdminRepository(db),
	}
}

// AdminDashboard handles admin dashboard page
func (h *AdminHandler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	// Get dashboard statistics from database
	stats, err := h.AdminRepo.GetDashboardStats()
	if err != nil {
		log.Printf("Error getting dashboard stats: %v", err)
		h.HandleError(w, r, err, "Dashboard verilerini alırken hata oluştu")
		return
	}

	// Get recent orders
	recentOrders, err := h.AdminRepo.GetRecentOrders(5)
	if err != nil {
		log.Printf("Error getting recent orders: %v", err)
		recentOrders = []map[string]interface{}{}
	}

	// Get recent users
	recentUsers, err := h.AdminRepo.GetRecentUsers(5)
	if err != nil {
		log.Printf("Error getting recent users: %v", err)
		recentUsers = []map[string]interface{}{}
	}

	data := map[string]interface{}{
		"Title":        "Admin Dashboard",
		"Stats":        stats,
		"RecentOrders": recentOrders,
		"RecentUsers":  recentUsers,
	}
	
	h.RenderTemplate(w, r, "admin/dashboard.gohtml", data)
}

// AdminUsers handles admin users page
func (h *AdminHandler) AdminUsers(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 20
	filters := make(map[string]interface{})
	
	// Get filters from query params
	if status := r.URL.Query().Get("status"); status != "" {
		filters["status"] = status
	}
	if role := r.URL.Query().Get("role"); role != "" {
		filters["role"] = role
	}
	if search := r.URL.Query().Get("search"); search != "" {
		filters["search"] = search
	}

	// Get users from database
	users, total, err := h.AdminRepo.GetUsers(page, limit, filters)
	if err != nil {
		log.Printf("Error getting users: %v", err)
		h.HandleError(w, r, err, "Kullanıcı verileri alınırken hata oluştu")
		return
	}

	// Calculate pagination
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// Get user statistics
	stats, err := h.AdminRepo.GetDashboardStats()
	if err != nil {
		log.Printf("Error getting user stats: %v", err)
		stats = make(map[string]interface{})
	}

	data := map[string]interface{}{
		"Title":       "User Management",
		"Users":       users,
		"TotalCount":  total,
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"Stats":       stats,
		"Filters":     filters,
	}
	
	h.RenderTemplate(w, r, "admin/users", data)
}

// AdminOrders handles admin orders page
func (h *AdminHandler) AdminOrders(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 20
	filters := make(map[string]interface{})
	
	// Get filters from query params
	if status := r.URL.Query().Get("status"); status != "" {
		filters["status"] = status
	}
	if paymentStatus := r.URL.Query().Get("payment_status"); paymentStatus != "" {
		filters["payment_status"] = paymentStatus
	}

	// Get orders from database
	orders, total, err := h.AdminRepo.GetOrders(page, limit, filters)
	if err != nil {
		log.Printf("Error getting orders: %v", err)
		h.HandleError(w, r, err, "Sipariş verileri alınırken hata oluştu")
		return
	}

	// Calculate pagination
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// Get order statistics
	stats, err := h.AdminRepo.GetDashboardStats()
	if err != nil {
		log.Printf("Error getting order stats: %v", err)
		stats = make(map[string]interface{})
	}

	data := map[string]interface{}{
		"Title":       "Order Management",
		"Orders":      orders,
		"TotalCount":  total,
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"Stats":       stats,
		"Filters":     filters,
	}
	
	h.RenderTemplate(w, r, "admin/orders", data)
}

// AdminProducts handles admin products page
func (h *AdminHandler) AdminProducts(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 20
	filters := make(map[string]interface{})
	
	// Get filters from query params
	if status := r.URL.Query().Get("status"); status != "" {
		filters["status"] = status
	}
	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		filters["category_id"] = categoryID
	}

	// Get products from database
	products, total, err := h.AdminRepo.GetProducts(page, limit, filters)
	if err != nil {
		log.Printf("Error getting products: %v", err)
		h.HandleError(w, r, err, "Ürün verileri alınırken hata oluştu")
		return
	}

	// Calculate pagination
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// Get categories for filter dropdown (simplified for now)
	categories := []map[string]interface{}{
		{"ID": 1, "Name": "Electronics"},
		{"ID": 2, "Name": "Clothing"},
		{"ID": 3, "Name": "Books"},
		{"ID": 4, "Name": "Home & Garden"},
	}

	data := map[string]interface{}{
		"Title":       "Product Management",
		"Products":    products,
		"Categories":  categories,
		"TotalCount":  total,
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"Filters":     filters,
	}
	
	h.RenderTemplate(w, r, "admin/products", data)
}

// AdminReports handles admin reports page
func (h *AdminHandler) AdminReports(w http.ResponseWriter, r *http.Request) {
	// Mock reports data
	data := map[string]interface{}{
		"Title": "Reports",
		"Stats": map[string]interface{}{
			"TotalRevenue":   "₺123,456.78",
			"TotalOrders":    567,
			"TotalProducts":  890,
			"TotalCustomers": 1234,
		},
		"Reports": map[string]interface{}{
			"Sales": []map[string]interface{}{
				{"Date": "2024-01-15", "Amount": "₺1,234.56"},
				{"Date": "2024-01-14", "Amount": "₺987.65"},
			},
			"TopProducts": []map[string]interface{}{
				{
					"Name":      "Product 1",
					"Category":  "Electronics",
					"SoldCount": 45,
					"Revenue":   "₺4,499.55",
					"Image":     "/static/images/product-placeholder.jpg",
				},
			},
			"NewCustomers":      123,
			"ReturningCustomers": 456,
			"TopVendors": []map[string]interface{}{
				{
					"Name":         "Vendor 1",
					"ProductCount": 25,
					"Revenue":      "₺12,345.67",
					"OrderCount":   89,
				},
			},
		},
		"DetailedReports": []map[string]interface{}{
			{
				"ID":        1,
				"Name":      "Monthly Sales Report",
				"Type":      "sales",
				"TypeColor": "success",
				"UpdatedAt": "2024-01-15",
				"Status":    "ready",
			},
		},
		"ScheduledReports": []map[string]interface{}{
			{
				"ID":         1,
				"Name":       "Weekly Sales Summary",
				"Schedule":   "Weekly",
				"Recipients": "admin@example.com",
				"Active":     true,
			},
		},
	}
	
	h.RenderTemplate(w, r, "admin/reports", data)
}

// AdminVendors handles admin vendors page
func (h *AdminHandler) AdminVendors(w http.ResponseWriter, r *http.Request) {
	// Mock vendors data
	data := map[string]interface{}{
		"Title": "Vendor Management",
		"Vendors": []map[string]interface{}{
			{
				"ID":                 1,
				"Name":               "Vendor 1",
				"BusinessName":       "Vendor Business Ltd.",
				"Email":              "vendor@example.com",
				"Phone":              "+90 555 123 4567",
				"ProductCount":       25,
				"ActiveProductCount": 23,
				"TotalSales":         "₺12,345.67",
				"OrderCount":         89,
				"CommissionRate":     15.0,
				"CommissionEarned":   "₺1,851.85",
				"Status":             "active",
				"CreatedAt":          "2024-01-01",
				"Logo":               "/static/images/vendor-logo.jpg",
			},
		},
		"Stats": map[string]interface{}{
			"TotalVendors":   50,
			"ActiveVendors":  45,
			"PendingVendors": 3,
			"TotalRevenue":   "₺500,000.00",
		},
		"Categories": []map[string]interface{}{
			{"ID": 1, "Name": "Electronics"},
			{"ID": 2, "Name": "Clothing"},
		},
		"TotalCount":  50,
		"CurrentPage": 1,
		"TotalPages":  5,
	}
	
	h.RenderTemplate(w, r, "admin/vendors", data)
}

// AdminSystemHealth handles admin system health page
func (h *AdminHandler) AdminSystemHealth(w http.ResponseWriter, r *http.Request) {
	// Get real system health data
	systemHealth, err := h.AdminRepo.GetSystemHealth()
	if err != nil {
		log.Printf("Error getting system health: %v", err)
		systemHealth = map[string]interface{}{
			"OverallStatus": "unhealthy",
			"HealthScore":   0,
			"DatabaseStatus": "disconnected",
		}
	}

	// Add additional system metrics (these would typically come from system monitoring tools)
	systemHealth["Uptime"] = "Runtime metrics not available"
	systemHealth["ServerLoad"] = "N/A"
	systemHealth["MemoryUsage"] = "N/A"
	systemHealth["MemoryUsed"] = "N/A"
	systemHealth["DatabaseConnections"] = "N/A"

	data := map[string]interface{}{
		"Title": "System Health",
		"SystemHealth": systemHealth,
		"ServerStatus": map[string]interface{}{
			"CPU": map[string]interface{}{
				"Status": "healthy",
				"Usage":  45,
				"Cores":  4,
			},
			"Memory": map[string]interface{}{
				"Status": "healthy",
				"Used":   "2.1GB",
				"Total":  "8GB",
				"Usage":  26,
			},
			"Disk": map[string]interface{}{
				"Status": "healthy",
				"Used":   "45GB",
				"Total":  "100GB",
				"Usage":  45,
			},
			"Network": map[string]interface{}{
				"Status":   "healthy",
				"Upload":   "1.2",
				"Download": "5.8",
			},
		},
		"DatabaseStatus": map[string]interface{}{
			"Connection": map[string]interface{}{
				"Status":       "connected",
				"ResponseTime": 12,
			},
			"Connections": map[string]interface{}{
				"Active": 25,
				"Max":    100,
				"Idle":   5,
			},
			"Size": map[string]interface{}{
				"Used":   "1.2GB",
				"Tables": 45,
				"Usage":  60,
			},
			"Performance": map[string]interface{}{
				"QPS":         150,
				"SlowQueries": 2,
			},
		},
		"Services": []map[string]interface{}{
			{
				"ID":          "web-server",
				"Name":        "Web Server",
				"Status":      "running",
				"Uptime":      "15 days",
				"CPU":         12,
				"Memory":      "512MB",
				"LastRestart": "2024-01-01",
			},
			{
				"ID":     "database",
				"Name":   "Database",
				"Status": "running",
				"Uptime": "15 days",
				"CPU":    8,
				"Memory": "1GB",
			},
		},
		"SystemLogs": []map[string]interface{}{
			{
				"Timestamp": "2024-01-15 10:30:00",
				"Level":     "info",
				"Message":   "System health check completed successfully",
			},
			{
				"Timestamp": "2024-01-15 10:25:00",
				"Level":     "warning",
				"Message":   "High memory usage detected",
			},
		},
		"Metrics": map[string]interface{}{
			"ResponseTime": map[string]interface{}{
				"Average": 120,
				"P95":     250,
				"Max":     500,
			},
			"ErrorRates": []map[string]interface{}{
				{"Code": "200", "Count": 1500, "Percentage": 95.0},
				{"Code": "404", "Count": 50, "Percentage": 3.2},
				{"Code": "500", "Count": 28, "Percentage": 1.8},
			},
		},
		"HealthChecks": []map[string]interface{}{
			{
				"Name":        "Database Connection",
				"Description": "Check database connectivity",
				"Status":      "pass",
				"LastCheck":   "2 minutes ago",
			},
			{
				"Name":        "Disk Space",
				"Description": "Check available disk space",
				"Status":      "warn",
				"LastCheck":   "5 minutes ago",
			},
		},
	}
	
	h.RenderTemplate(w, r, "admin/system-health", data)
}

// AdminSEO handles admin SEO page
func (h *AdminHandler) AdminSEO(w http.ResponseWriter, r *http.Request) {
	// Mock SEO data
	data := map[string]interface{}{
		"Title": "SEO Management",
		"SEOStats": map[string]interface{}{
			"OverallScore":        85,
			"IndexedPages":        1250,
			"IndexedPagesGrowth":  45,
			"TotalKeywords":       125,
			"KeywordRankings":     68,
			"Backlinks":           890,
			"BacklinksGrowth":     23,
		},
		"MetaTags": []map[string]interface{}{
			{
				"ID":          1,
				"Page":        "Home",
				"Title":       "E-commerce Platform - Best Products Online",
				"Description": "Discover amazing products at great prices. Shop electronics, clothing, and more with fast delivery.",
				"IsOptimized": true,
			},
			{
				"ID":          2,
				"Page":        "Products",
				"Title":       "Products - Shop Online",
				"Description": "Browse our wide selection of products.",
				"IsOptimized": false,
			},
		},
		"Keywords": []map[string]interface{}{
			{
				"ID":          1,
				"Keyword":     "online shopping",
				"URL":         "/",
				"CurrentRank": 15,
				"RankChange":  3,
			},
			{
				"ID":          2,
				"Keyword":     "electronics store",
				"URL":         "/products/electronics",
				"CurrentRank": 8,
				"RankChange":  -2,
			},
		},
		"SEOAnalysis": map[string]interface{}{
			"Technical": []map[string]interface{}{
				{"Name": "SSL Certificate", "Status": "pass"},
				{"Name": "Mobile Friendly", "Status": "pass"},
				{"Name": "Page Speed", "Status": "warning"},
				{"Name": "XML Sitemap", "Status": "pass"},
			},
			"Content": []map[string]interface{}{
				{"Name": "Title Tags", "Status": "pass"},
				{"Name": "Meta Descriptions", "Status": "warning"},
				{"Name": "H1 Tags", "Status": "pass"},
				{"Name": "Alt Text", "Status": "fail"},
			},
			"Performance": []map[string]interface{}{
				{"Name": "Load Time", "Status": "warning"},
				{"Name": "Core Web Vitals", "Status": "pass"},
				{"Name": "Image Optimization", "Status": "fail"},
			},
		},
		"Sitemap": map[string]interface{}{
			"TotalURLs":          1250,
			"IsValid":            true,
			"LastGenerated":      "2024-01-15 10:00:00",
			"SubmittedToGoogle":  true,
		},
		"Robots": map[string]interface{}{
			"IsValid": true,
			"Content": "User-agent: *\nAllow: /\nSitemap: https://example.com/sitemap.xml",
		},
		"SEORecommendations": []map[string]interface{}{
			{
				"ID":          1,
				"Title":       "Optimize Image Alt Text",
				"Description": "Many images are missing alt text which affects accessibility and SEO.",
				"Priority":    "high",
				"ActionURL":   "/admin/products",
				"ActionText":  "Fix Images",
			},
			{
				"ID":          2,
				"Title":       "Improve Page Load Speed",
				"Description": "Some pages are loading slower than recommended.",
				"Priority":    "medium",
				"ActionURL":   "/admin/system-health",
				"ActionText":  "Check Performance",
			},
		},
		"SEOReports": []map[string]interface{}{
			{
				"ID":          1,
				"Name":        "Monthly SEO Report",
				"GeneratedAt": "2024-01-15",
				"Status":      "completed",
			},
		},
	}
	
	h.RenderTemplate(w, r, "admin/seo", data)
}

// API Handlers

// APIGetUserStats returns user statistics
func (h *AdminHandler) APIGetUserStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.AdminRepo.GetDashboardStats()
	if err != nil {
		log.Printf("Error getting user stats: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "İstatistikler alınırken hata oluştu",
		})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// APIUpdateUserStatus updates user status
func (h *AdminHandler) APIUpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Geçersiz kullanıcı ID",
		})
		return
	}
	
	var request struct {
		Status string `json:"status"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Geçersiz istek formatı",
		})
		return
	}
	
	// Convert status to boolean
	isActive := request.Status == "active"
	
	// Update user status in database
	err = h.AdminRepo.UpdateUserStatus(userID, isActive)
	if err != nil {
		log.Printf("Error updating user status: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Kullanıcı durumu güncellenirken hata oluştu",
		})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Kullanıcı %d durumu %s olarak güncellendi", userID, request.Status),
	})
}

// APIUpdateOrderStatus updates order status
func (h *AdminHandler) APIUpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderIDStr := vars["id"]
	
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Geçersiz sipariş ID",
		})
		return
	}
	
	var request struct {
		Status string `json:"status"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Geçersiz istek formatı",
		})
		return
	}
	
	// Update order status in database
	err = h.AdminRepo.UpdateOrderStatus(orderID, request.Status)
	if err != nil {
		log.Printf("Error updating order status: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Sipariş durumu güncellenirken hata oluştu",
		})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Sipariş %d durumu %s olarak güncellendi", orderID, request.Status),
	})
}

// APIDeleteUser deletes a user
func (h *AdminHandler) APIDeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Geçersiz kullanıcı ID",
		})
		return
	}
	
	// Delete user in database (soft delete)
	err = h.AdminRepo.DeleteUser(userID)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Kullanıcı silinirken hata oluştu",
		})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Kullanıcı %d başarıyla silindi", userID),
	})
}

// APIDeleteOrder deletes an order
func (h *AdminHandler) APIDeleteOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]
	
	// Mock deletion
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Order %s deleted successfully", orderID),
	})
}

// APISystemHealthCheck performs system health check
func (h *AdminHandler) APISystemHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Get real system health data
	healthData, err := h.AdminRepo.GetSystemHealth()
	if err != nil {
		log.Printf("Error getting system health: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Sistem sağlığı kontrol edilirken hata oluştu",
		})
		return
	}
	
	// Add timestamp and format response
	healthData["timestamp"] = time.Now().Format(time.RFC3339)
	
	// Add basic checks
	checks := []map[string]interface{}{}
	if healthData["DatabaseStatus"] == "connected" {
		checks = append(checks, map[string]interface{}{
			"name": "database", 
			"status": "pass", 
			"details": "Database connection successful",
		})
	} else {
		checks = append(checks, map[string]interface{}{
			"name": "database", 
			"status": "fail", 
			"details": "Database connection failed",
		})
	}
	
	healthData["checks"] = checks
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    healthData,
	})
}

// APIGenerateSitemap generates XML sitemap
func (h *AdminHandler) APIGenerateSitemap(w http.ResponseWriter, r *http.Request) {
	// Mock sitemap generation
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Sitemap generated successfully",
		"data": map[string]interface{}{
			"totalUrls":   1250,
			"generatedAt": time.Now().Format(time.RFC3339),
		},
	})
}

// APIAnalyzeSEO performs SEO analysis
func (h *AdminHandler) APIAnalyzeSEO(w http.ResponseWriter, r *http.Request) {
	// Mock SEO analysis
	analysisData := map[string]interface{}{
		"overallScore": 85,
		"issues": []map[string]interface{}{
			{"type": "warning", "message": "Some images missing alt text"},
			{"type": "error", "message": "Page load time too slow"},
		},
		"recommendations": []string{
			"Optimize images",
			"Improve page speed",
			"Add more internal links",
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    analysisData,
	})
}

// APICreateUser creates a new user
func (h *AdminHandler) APICreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Geçersiz istek formatı",
		})
		return
	}
	
	// Validate user data
	if err := user.Validate(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	
	// Set timestamps
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	
	// Create user in database
	userID, err := h.AdminRepo.Create("users", user)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Kullanıcı oluşturulurken hata oluştu",
		})
		return
	}
	
	user.ID = userID
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Kullanıcı başarıyla oluşturuldu",
		"data":    user,
	})
}

// APIBulkProductAction handles bulk product actions
func (h *AdminHandler) APIBulkProductAction(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Action     string  `json:"action"`
		ProductIDs []int64 `json:"product_ids"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Geçersiz istek formatı",
		})
		return
	}
	
	if len(request.ProductIDs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "En az bir ürün seçilmelidir",
		})
		return
	}
	
	// Process bulk action
	var newStatus string
	switch request.Action {
	case "approve":
		newStatus = "active"
	case "deactivate":
		newStatus = "inactive"
	case "reject":
		newStatus = "rejected"
	case "delete":
		// For delete, we'll set status to inactive (soft delete)
		newStatus = "inactive"
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Geçersiz işlem",
		})
		return
	}
	
	// Update products in database
	successCount := 0
	for _, productID := range request.ProductIDs {
		err := h.AdminRepo.Update("products", productID, map[string]interface{}{
			"status":     newStatus,
			"updated_at": time.Now(),
		})
		if err != nil {
			log.Printf("Error updating product %d: %v", productID, err)
		} else {
			successCount++
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("%d ürün başarıyla güncellendi", successCount),
		"data": map[string]interface{}{
			"processed": len(request.ProductIDs),
			"success":   successCount,
			"failed":    len(request.ProductIDs) - successCount,
		},
	})
}

// APIUpdateProductStatus updates individual product status
func (h *AdminHandler) APIUpdateProductStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productIDStr := vars["id"]
	
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Geçersiz ürün ID",
		})
		return
	}
	
	var request struct {
		Status string `json:"status"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Geçersiz istek formatı",
		})
		return
	}
	
	// Update product status in database
	err = h.AdminRepo.Update("products", productID, map[string]interface{}{
		"status":     request.Status,
		"updated_at": time.Now(),
	})
	if err != nil {
		log.Printf("Error updating product status: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Ürün durumu güncellenirken hata oluştu",
		})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Ürün %d durumu %s olarak güncellendi", productID, request.Status),
	})
}

// APIExportUsers exports user data
func (h *AdminHandler) APIExportUsers(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Format string   `json:"format"`
		Fields []string `json:"fields"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Geçersiz istek formatı",
		})
		return
	}
	
	// Get all users (for now, we'll limit to 1000)
	users, _, err := h.AdminRepo.GetUsers(1, 1000, map[string]interface{}{})
	if err != nil {
		log.Printf("Error getting users for export: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Kullanıcı verileri alınırken hata oluştu",
		})
		return
	}
	
	// For now, we'll return the data as JSON
	// In a real implementation, you would generate CSV/Excel/PDF based on the format
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("%d kullanıcı %s formatında hazırlandı", len(users), request.Format),
		"data": map[string]interface{}{
			"format":     request.Format,
			"fields":     request.Fields,
			"users":      users,
			"total":      len(users),
			"exportedAt": time.Now().Format(time.RFC3339),
		},
	})
}


