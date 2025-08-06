package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"kolajAi/internal/services"
)

// AnalyticsHandler handles advanced analytics requests
type AnalyticsHandler struct {
	*Handler
	AnalyticsService *services.AdvancedAnalyticsService
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(h *Handler, analyticsService *services.AdvancedAnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		Handler:          h,
		AnalyticsService: analyticsService,
	}
}

// Dashboard handles analytics dashboard
func (h *AnalyticsHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get business metrics
	metrics := map[string]interface{}{
		"revenue": map[string]interface{}{
			"total_revenue":     125000.50,
			"revenue_growth":    15.8,
			"monthly_recurring": 45000.00,
			"average_order":     89.50,
		},
		"sales": map[string]interface{}{
			"total_orders":     1450,
			"completed_orders": 1380,
			"conversion_rate":  3.2,
			"sales_growth":     12.5,
		},
		"customers": map[string]interface{}{
			"total_customers": 2850,
			"new_customers":   185,
			"retention_rate":  78.5,
			"lifetime_value":  450.75,
		},
		"products": map[string]interface{}{
			"total_products": 850,
			"best_selling":   "Test Ürün 1",
			"low_stock":      25,
			"out_of_stock":   8,
		},
	}

	// Get revenue trend data
	revenueTrend := []map[string]interface{}{
		{"month": "Ocak", "revenue": 95000, "orders": 1200},
		{"month": "Şubat", "revenue": 105000, "orders": 1350},
		{"month": "Mart", "revenue": 115000, "orders": 1400},
		{"month": "Nisan", "revenue": 125000, "orders": 1450},
	}

	// Get top products
	topProducts := []map[string]interface{}{
		{"name": "Test Ürün 1", "sales": 450, "revenue": 22500},
		{"name": "Test Ürün 2", "sales": 380, "revenue": 19000},
		{"name": "Test Ürün 3", "sales": 320, "revenue": 16000},
		{"name": "Test Ürün 4", "sales": 280, "revenue": 14000},
		{"name": "Test Ürün 5", "sales": 250, "revenue": 12500},
	}

	// Get customer segments
	customerSegments := []map[string]interface{}{
		{"segment": "Premium", "count": 450, "revenue": 67500, "percentage": 54.0},
		{"segment": "Regular", "count": 1200, "revenue": 48000, "percentage": 38.4},
		{"segment": "New", "count": 1200, "revenue": 9500, "percentage": 7.6},
	}

	data := map[string]interface{}{
		"Title":            "İleri Düzey Analitik",
		"Metrics":          metrics,
		"RevenueTrend":     revenueTrend,
		"TopProducts":      topProducts,
		"CustomerSegments": customerSegments,
	}

	h.RenderTemplate(w, r, "analytics/dashboard.gohtml", data)
}

// Revenue handles revenue analytics
func (h *AnalyticsHandler) Revenue(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get detailed revenue data
	revenueData := map[string]interface{}{
		"total_revenue":       125000.50,
		"revenue_growth":      15.8,
		"monthly_recurring":   45000.00,
		"average_order_value": 89.50,
		"profit_margin":       35.2,
		"revenue_by_category": map[string]float64{
			"Elektronik": 45000,
			"Giyim":      35000,
			"Kitap":      25000,
			"Ev & Yaşam": 20000,
		},
		"revenue_by_channel": map[string]float64{
			"Web":    75000,
			"Mobile": 35000,
			"API":    15000,
		},
	}

	// Get revenue forecast
	revenueForecast := []map[string]interface{}{
		{"month": "Mayıs", "predicted": 135000, "confidence": 85},
		{"month": "Haziran", "predicted": 145000, "confidence": 80},
		{"month": "Temmuz", "predicted": 155000, "confidence": 75},
		{"month": "Ağustos", "predicted": 165000, "confidence": 70},
	}

	data := map[string]interface{}{
		"Title":           "Gelir Analizi",
		"RevenueData":     revenueData,
		"RevenueForecast": revenueForecast,
	}

	h.RenderTemplate(w, r, "analytics/revenue.gohtml", data)
}

// Customers handles customer analytics
func (h *AnalyticsHandler) Customers(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get customer analytics
	customerData := map[string]interface{}{
		"total_customers":  2850,
		"new_customers":    185,
		"active_customers": 2100,
		"retention_rate":   78.5,
		"churn_rate":       21.5,
		"lifetime_value":   450.75,
		"acquisition_cost": 25.50,
		"customer_segments": []map[string]interface{}{
			{"name": "VIP", "count": 150, "revenue": 45000, "avg_order": 300},
			{"name": "Premium", "count": 450, "revenue": 67500, "avg_order": 150},
			{"name": "Regular", "count": 1200, "revenue": 48000, "avg_order": 40},
			{"name": "New", "count": 1050, "revenue": 21000, "avg_order": 20},
		},
	}

	// Get customer behavior
	customerBehavior := []map[string]interface{}{
		{"behavior": "Repeat Purchase", "percentage": 65.2, "trend": "up"},
		{"behavior": "Cart Abandonment", "percentage": 28.5, "trend": "down"},
		{"behavior": "Email Engagement", "percentage": 42.8, "trend": "up"},
		{"behavior": "Social Sharing", "percentage": 15.3, "trend": "stable"},
	}

	data := map[string]interface{}{
		"Title":            "Müşteri Analizi",
		"CustomerData":     customerData,
		"CustomerBehavior": customerBehavior,
	}

	h.RenderTemplate(w, r, "analytics/customers.gohtml", data)
}

// Products handles product analytics
func (h *AnalyticsHandler) Products(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Get product analytics
	productData := map[string]interface{}{
		"total_products":  850,
		"active_products": 780,
		"best_seller":     "Test Ürün 1",
		"worst_performer": "Test Ürün 50",
		"average_rating":  4.2,
		"return_rate":     5.8,
	}

	// Get product performance
	productPerformance := []map[string]interface{}{
		{"name": "Test Ürün 1", "sales": 450, "revenue": 22500, "rating": 4.8, "stock": 25},
		{"name": "Test Ürün 2", "sales": 380, "revenue": 19000, "rating": 4.5, "stock": 15},
		{"name": "Test Ürün 3", "sales": 320, "revenue": 16000, "rating": 4.3, "stock": 8},
		{"name": "Test Ürün 4", "sales": 280, "revenue": 14000, "rating": 4.1, "stock": 32},
		{"name": "Test Ürün 5", "sales": 250, "revenue": 12500, "rating": 4.0, "stock": 12},
	}

	// Get category performance
	categoryPerformance := []map[string]interface{}{
		{"category": "Elektronik", "products": 150, "sales": 1200, "revenue": 45000},
		{"category": "Giyim", "products": 250, "sales": 800, "revenue": 35000},
		{"category": "Kitap", "products": 200, "sales": 600, "revenue": 25000},
		{"category": "Ev & Yaşam", "products": 250, "sales": 400, "revenue": 20000},
	}

	data := map[string]interface{}{
		"Title":               "Ürün Analizi",
		"ProductData":         productData,
		"ProductPerformance":  productPerformance,
		"CategoryPerformance": categoryPerformance,
	}

	h.RenderTemplate(w, r, "analytics/products.gohtml", data)
}

// API Methods

// APIGetBusinessMetrics returns comprehensive business metrics
func (h *AnalyticsHandler) APIGetBusinessMetrics(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "month"
	}

	metrics := map[string]interface{}{
		"revenue": map[string]interface{}{
			"total":     125000.50,
			"growth":    15.8,
			"recurring": 45000.00,
			"avg_order": 89.50,
		},
		"sales": map[string]interface{}{
			"orders":     1450,
			"completed":  1380,
			"conversion": 3.2,
			"growth":     12.5,
		},
		"customers": map[string]interface{}{
			"total":          2850,
			"new":            185,
			"retention":      78.5,
			"lifetime_value": 450.75,
		},
		"period":       period,
		"generated_at": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"metrics": metrics,
	})
}

// APIGetRevenueForecast returns revenue forecasting data
func (h *AnalyticsHandler) APIGetRevenueForecast(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	months := r.URL.Query().Get("months")
	if months == "" {
		months = "6"
	}

	forecast := []map[string]interface{}{
		{"month": "Mayıs", "predicted": 135000, "confidence": 85, "factors": []string{"seasonal", "trend"}},
		{"month": "Haziran", "predicted": 145000, "confidence": 80, "factors": []string{"marketing", "trend"}},
		{"month": "Temmuz", "predicted": 155000, "confidence": 75, "factors": []string{"summer", "promotion"}},
		{"month": "Ağustos", "predicted": 165000, "confidence": 70, "factors": []string{"back-to-school"}},
		{"month": "Eylül", "predicted": 175000, "confidence": 65, "factors": []string{"autumn", "new-products"}},
		{"month": "Ekim", "predicted": 185000, "confidence": 60, "factors": []string{"holiday-prep"}},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"forecast": forecast,
		"months":   months,
	})
}

// APIGetCustomerSegments returns customer segmentation data
func (h *AnalyticsHandler) APIGetCustomerSegments(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	segments := []map[string]interface{}{
		{
			"name":        "VIP",
			"count":       150,
			"revenue":     45000,
			"avg_order":   300,
			"retention":   95.2,
			"description": "En yüksek değerli müşteriler",
		},
		{
			"name":        "Premium",
			"count":       450,
			"revenue":     67500,
			"avg_order":   150,
			"retention":   85.8,
			"description": "Düzenli alışveriş yapan müşteriler",
		},
		{
			"name":        "Regular",
			"count":       1200,
			"revenue":     48000,
			"avg_order":   40,
			"retention":   65.5,
			"description": "Orta düzey müşteriler",
		},
		{
			"name":        "New",
			"count":       1050,
			"revenue":     21000,
			"avg_order":   20,
			"retention":   25.0,
			"description": "Yeni kayıt olan müşteriler",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"segments": segments,
	})
}

// APIGetProductInsights returns product performance insights
func (h *AnalyticsHandler) APIGetProductInsights(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	insights := map[string]interface{}{
		"top_performers": []map[string]interface{}{
			{"name": "Test Ürün 1", "sales": 450, "revenue": 22500, "growth": 25.5},
			{"name": "Test Ürün 2", "sales": 380, "revenue": 19000, "growth": 18.2},
			{"name": "Test Ürün 3", "sales": 320, "revenue": 16000, "growth": 12.8},
		},
		"trending_up": []map[string]interface{}{
			{"name": "Yeni Ürün A", "growth": 45.2, "potential": "high"},
			{"name": "Yeni Ürün B", "growth": 38.5, "potential": "medium"},
		},
		"needs_attention": []map[string]interface{}{
			{"name": "Eski Ürün X", "decline": -15.5, "action": "promotion"},
			{"name": "Eski Ürün Y", "decline": -22.8, "action": "discontinue"},
		},
		"recommendations": []string{
			"Test Ürün 1 stokunu artırın",
			"Yeni Ürün A için pazarlama kampanyası başlatın",
			"Eski Ürün X için indirim yapın",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"insights": insights,
	})
}
