package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	
	"github.com/gorilla/mux"
)

// AdminHandler handles admin-related requests
type AdminHandler struct {
	*Handler
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(h *Handler) *AdminHandler {
	return &AdminHandler{
		Handler: h,
	}
}

// AdminDashboard handles admin dashboard page
func (h *AdminHandler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	// Mock admin dashboard data
	data := map[string]interface{}{
		"Title": "Admin Dashboard",
		"Stats": map[string]interface{}{
			"TotalUsers":    1234,
			"TotalOrders":   567,
			"TotalProducts": 890,
			"TotalRevenue":  "₺123,456.78",
		},
		"RecentOrders": []map[string]interface{}{
			{
				"ID":          "ORD-001",
				"CustomerName": "John Doe",
				"Amount":      "₺299.99",
				"Status":      "completed",
				"Date":        "2024-01-15",
			},
		},
		"RecentUsers": []map[string]interface{}{
			{
				"ID":    1,
				"Name":  "Jane Smith",
				"Email": "jane@example.com",
				"Date":  "2024-01-15",
			},
		},
	}
	
	h.RenderTemplate(w, r, "admin/dashboard", data)
}

// AdminUsers handles admin users page
func (h *AdminHandler) AdminUsers(w http.ResponseWriter, r *http.Request) {
	// Mock users data
	data := map[string]interface{}{
		"Title": "User Management",
		"Users": []map[string]interface{}{
			{
				"ID":        1,
				"Name":      "John Doe",
				"Email":     "john@example.com",
				"Status":    "active",
				"CreatedAt": "2024-01-01",
				"LastLogin": "2024-01-15",
			},
		},
		"TotalCount":  100,
		"CurrentPage": 1,
		"TotalPages":  10,
	}
	
	h.RenderTemplate(w, r, "admin/users", data)
}

// AdminOrders handles admin orders page
func (h *AdminHandler) AdminOrders(w http.ResponseWriter, r *http.Request) {
	// Mock orders data
	data := map[string]interface{}{
		"Title": "Order Management",
		"Orders": []map[string]interface{}{
			{
				"ID":            "ORD-001",
				"CustomerName":  "John Doe",
				"CustomerEmail": "john@example.com",
				"Amount":        "₺299.99",
				"Status":        "pending",
				"CreatedAt":     "2024-01-15",
				"Items": []map[string]interface{}{
					{
						"Name":     "Product 1",
						"Quantity": 2,
						"Price":    "₺149.99",
					},
				},
			},
		},
		"Stats": map[string]interface{}{
			"TotalOrders":     567,
			"PendingOrders":   23,
			"CompletedOrders": 544,
			"TotalRevenue":    "₺123,456.78",
		},
		"TotalCount":  567,
		"CurrentPage": 1,
		"TotalPages":  57,
	}
	
	h.RenderTemplate(w, r, "admin/orders", data)
}

// AdminProducts handles admin products page
func (h *AdminHandler) AdminProducts(w http.ResponseWriter, r *http.Request) {
	// Mock products data
	data := map[string]interface{}{
		"Title": "Product Management",
		"Products": []map[string]interface{}{
			{
				"ID":          1,
				"Name":        "Sample Product",
				"SKU":         "SKU-001",
				"Price":       "₺99.99",
				"Stock":       50,
				"Status":      "active",
				"Category":    "Electronics",
				"CreatedAt":   "2024-01-01",
				"Image":       "/static/images/product-placeholder.jpg",
			},
		},
		"Categories": []map[string]interface{}{
			{"ID": 1, "Name": "Electronics"},
			{"ID": 2, "Name": "Clothing"},
		},
		"TotalCount":  890,
		"CurrentPage": 1,
		"TotalPages":  89,
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
	// Mock system health data
	data := map[string]interface{}{
		"Title": "System Health",
		"SystemHealth": map[string]interface{}{
			"OverallStatus":       "healthy",
			"HealthScore":         95,
			"ServerLoad":          75,
			"Uptime":             "15 days",
			"DatabaseStatus":     "connected",
			"DatabaseConnections": 25,
			"MemoryUsage":        68,
			"MemoryUsed":         "2.1GB",
		},
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
	stats := map[string]interface{}{
		"totalUsers":    1234,
		"activeUsers":   1100,
		"newUsers":      134,
		"bannedUsers":   0,
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
	userID := vars["id"]
	
	var request struct {
		Status string `json:"status"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	// Mock update
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("User %s status updated to %s", userID, request.Status),
	})
}

// APIUpdateOrderStatus updates order status
func (h *AdminHandler) APIUpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]
	
	var request struct {
		Status string `json:"status"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	// Mock update
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Order %s status updated to %s", orderID, request.Status),
	})
}

// APIDeleteUser deletes a user
func (h *AdminHandler) APIDeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	
	// Mock deletion
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("User %s deleted successfully", userID),
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
	// Mock health check
	healthData := map[string]interface{}{
		"status":       "healthy",
		"score":        95,
		"timestamp":    time.Now().Format(time.RFC3339),
		"checks": []map[string]interface{}{
			{"name": "database", "status": "pass", "responseTime": "12ms"},
			{"name": "cache", "status": "pass", "responseTime": "5ms"},
			{"name": "storage", "status": "pass", "usage": "45%"},
		},
	}
	
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


