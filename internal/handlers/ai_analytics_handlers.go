package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"kolajAi/internal/services"
)

// AIAnalyticsHandler handles advanced AI analytics requests
type AIAnalyticsHandler struct {
	*Handler
	analyticsService *services.AIAnalyticsService
}

// NewAIAnalyticsHandler creates a new AI analytics handler
func NewAIAnalyticsHandler(h *Handler, analyticsService *services.AIAnalyticsService) *AIAnalyticsHandler {
	return &AIAnalyticsHandler{
		Handler:          h,
		analyticsService: analyticsService,
	}
}

// GetMarketTrends returns market trend analysis
func (h *AIAnalyticsHandler) GetMarketTrends(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get market trends
	trends, err := h.analyticsService.AnalyzeMarketTrends()
	if err != nil {
		Logger.Printf("Error getting market trends: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"trends":  trends,
		"count":   len(trends),
	}); err != nil {
		Logger.Printf("Error encoding market trends response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetProductInsights returns detailed product insights
func (h *AIAnalyticsHandler) GetProductInsights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract product ID from URL path
	productIDStr := r.URL.Path[len("/api/ai/product-insights/"):]
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Get product insights
	insights, err := h.analyticsService.AnalyzeProductInsights(productID)
	if err != nil {
		Logger.Printf("Error getting product insights for product %d: %v", productID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"insights": insights,
	}); err != nil {
		Logger.Printf("Error encoding product insights response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetCustomerSegments returns customer segmentation analysis
func (h *AIAnalyticsHandler) GetCustomerSegments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get customer segments
	segments, err := h.analyticsService.AnalyzeCustomerSegments()
	if err != nil {
		Logger.Printf("Error getting customer segments: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"segments": segments,
		"count":    len(segments),
	}); err != nil {
		Logger.Printf("Error encoding customer segments response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetPricingStrategy returns AI-powered pricing recommendations
func (h *AIAnalyticsHandler) GetPricingStrategy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract product ID from URL path
	productIDStr := r.URL.Path[len("/api/ai/pricing-strategy/"):]
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Get pricing strategy
	strategy, err := h.analyticsService.GeneratePricingStrategy(productID)
	if err != nil {
		Logger.Printf("Error getting pricing strategy for product %d: %v", productID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"strategy": strategy,
	}); err != nil {
		Logger.Printf("Error encoding pricing strategy response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetAnalyticsDashboard renders the AI analytics dashboard page
func (h *AIAnalyticsHandler) GetAnalyticsDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Prepare template data
	data := h.GetTemplateData()
	data["PageTitle"] = "AI Analytics Dashboard"
	data["PageDescription"] = "Gelişmiş AI analitikleri ve pazar içgörüleri"

	// Get some sample data for the dashboard
	trends, err := h.analyticsService.AnalyzeMarketTrends()
	if err != nil {
		Logger.Printf("Error getting market trends for dashboard: %v", err)
		trends = make([]*services.MarketTrend, 0)
	}

	segments, err := h.analyticsService.AnalyzeCustomerSegments()
	if err != nil {
		Logger.Printf("Error getting customer segments for dashboard: %v", err)
		segments = make([]*services.CustomerSegment, 0)
	}

	data["MarketTrends"] = trends
	data["CustomerSegments"] = segments

	// Render template
	if err := h.Templates.ExecuteTemplate(w, "ai/analytics-dashboard", data); err != nil {
		Logger.Printf("Error rendering analytics dashboard template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetMarketTrendsPage renders the market trends analysis page
func (h *AIAnalyticsHandler) GetMarketTrendsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Prepare template data
	data := h.GetTemplateData()
	data["PageTitle"] = "Pazar Trend Analizi"
	data["PageDescription"] = "AI destekli pazar trend analizi ve tahminleri"

	// Render template
	if err := h.Templates.ExecuteTemplate(w, "ai/market-trends", data); err != nil {
		Logger.Printf("Error rendering market trends template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetProductInsightsPage renders the product insights page
func (h *AIAnalyticsHandler) GetProductInsightsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Prepare template data
	data := h.GetTemplateData()
	data["PageTitle"] = "Ürün İçgörüleri"
	data["PageDescription"] = "AI destekli ürün performans analizi ve öneriler"

	// Render template
	if err := h.Templates.ExecuteTemplate(w, "ai/product-insights", data); err != nil {
		Logger.Printf("Error rendering product insights template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetCustomerSegmentsPage renders the customer segments analysis page
func (h *AIAnalyticsHandler) GetCustomerSegmentsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Prepare template data
	data := h.GetTemplateData()
	data["PageTitle"] = "Müşteri Segmentasyonu"
	data["PageDescription"] = "AI destekli müşteri davranış analizi ve segmentasyon"

	// Render template
	if err := h.Templates.ExecuteTemplate(w, "ai/customer-segments", data); err != nil {
		Logger.Printf("Error rendering customer segments template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetPricingStrategyPage renders the pricing strategy page
func (h *AIAnalyticsHandler) GetPricingStrategyPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication
	if !h.IsAuthenticated(r) {
		h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
		return
	}

	// Prepare template data
	data := h.GetTemplateData()
	data["PageTitle"] = "Fiyatlandırma Stratejisi"
	data["PageDescription"] = "AI destekli fiyat optimizasyonu ve strateji önerileri"

	// Render template
	if err := h.Templates.ExecuteTemplate(w, "ai/pricing-strategy", data); err != nil {
		Logger.Printf("Error rendering pricing strategy template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
