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

// AIHandler handles AI-related requests
type AIHandler struct {
	*Handler
	aiService *services.AIService
}

// NewAIHandler creates a new AI handler
func NewAIHandler(h *Handler, aiService *services.AIService) *AIHandler {
	return &AIHandler{
		Handler:   h,
		aiService: aiService,
	}
}

// getUserFromSession is a helper function to get user from session
func (h *AIHandler) getUserFromSession(r *http.Request) (*models.User, error) {
	if !h.IsAuthenticated(r) {
		return nil, fmt.Errorf("user not authenticated")
	}

	sessionData, err := h.SessionManager.GetSession(r)
	if err != nil {
		return nil, fmt.Errorf("session error: %w", err)
	}

	if sessionData == nil {
		return nil, fmt.Errorf("session data not found")
	}

	// Create a basic user object from session data
	// In a real application, you would fetch full user details from database
	user := &models.User{
		ID:    sessionData.UserID,
		Email: "admin@example.com", // Placeholder - should be fetched from DB
		Name:  "Admin User",        // Placeholder - should be fetched from DB
	}

	return user, nil
}

// GetRecommendations returns personalized product recommendations for a user
func (h *AIHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 12 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 50 {
			limit = parsedLimit
		}
	}

	// Get recommendations
	recommendations, err := h.aiService.GetPersonalizedRecommendations(int(user.ID), limit)
	if err != nil {
		Logger.Printf("Error getting recommendations for user %d: %v", int(user.ID), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success":         true,
		"recommendations": recommendations,
		"count":           len(recommendations),
	}); err != nil {
		Logger.Printf("Error encoding recommendations response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// OptimizePrice provides AI-powered price optimization for a product
func (h *AIHandler) OptimizePrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	_, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse product ID from URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Product ID required", http.StatusBadRequest)
		return
	}

	productID, err := strconv.Atoi(pathParts[2])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Get price optimization
	optimization, err := h.aiService.OptimizeProductPricing(productID)
	if err != nil {
		Logger.Printf("Error optimizing price for product %d: %v", productID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"optimization": optimization,
	}); err != nil {
		Logger.Printf("Error encoding price optimization response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// PredictCategory predicts product category based on name and description
func (h *AIHandler) PredictCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	_, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var request struct {
		ProductName string `json:"product_name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(request.ProductName) == "" {
		http.Error(w, "Product name is required", http.StatusBadRequest)
		return
	}

	// Get category predictions
	predictions, err := h.aiService.PredictProductCategory(request.ProductName, request.Description)
	if err != nil {
		Logger.Printf("Error predicting category for product '%s': %v", request.ProductName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"predictions": predictions,
		"count":       len(predictions),
	}); err != nil {
		Logger.Printf("Error encoding category prediction response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// SmartSearch performs AI-enhanced product search
func (h *AIHandler) SmartSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get search parameters
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		http.Error(w, "Search query is required", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Perform smart search
	searchResult, err := h.aiService.SmartSearch(query, limit, offset)
	if err != nil {
		Logger.Printf("Error performing smart search for query '%s': %v", query, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"result":  searchResult,
	}); err != nil {
		Logger.Printf("Error encoding smart search response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetAIDashboard shows the AI dashboard page
func (h *AIHandler) GetAIDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user, err := h.getUserFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Prepare template data
	data := h.GetTemplateData()
	data["User"] = user
	data["PageTitle"] = "AI Dashboard - KolajAI"
	data["PageDescription"] = "AI-powered insights and recommendations for your marketplace experience"

	// Get some sample recommendations for display
	recommendations, err := h.aiService.GetPersonalizedRecommendations(int(user.ID), 6)
	if err != nil {
		Logger.Printf("Error getting recommendations for dashboard: %v", err)
		recommendations = make([]*services.ProductRecommendation, 0)
	}
	data["Recommendations"] = recommendations

	// Render template
	if err := h.Templates.ExecuteTemplate(w, "ai/dashboard", data); err != nil {
		Logger.Printf("Error rendering AI dashboard template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetRecommendationsPage shows the recommendations page
func (h *AIHandler) GetRecommendationsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user, err := h.getUserFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 24 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// Get recommendations
	recommendations, err := h.aiService.GetPersonalizedRecommendations(int(user.ID), limit)
	if err != nil {
		Logger.Printf("Error getting recommendations: %v", err)
		recommendations = make([]*services.ProductRecommendation, 0)
	}

	// Prepare template data
	data := h.GetTemplateData()
	data["User"] = user
	data["PageTitle"] = "Kişisel Öneriler - KolajAI"
	data["PageDescription"] = "Size özel AI destekli ürün önerileri"
	data["Recommendations"] = recommendations
	data["RecommendationCount"] = len(recommendations)

	// Render template
	if err := h.Templates.ExecuteTemplate(w, "ai/recommendations", data); err != nil {
		Logger.Printf("Error rendering recommendations template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetPriceOptimizationPage shows the price optimization page for vendors
func (h *AIHandler) GetPriceOptimizationPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user, err := h.getUserFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Prepare template data
	data := h.GetTemplateData()
	data["User"] = user
	data["PageTitle"] = "Fiyat Optimizasyonu - KolajAI"
	data["PageDescription"] = "AI destekli fiyat optimizasyon önerileri"

	// Render template
	if err := h.Templates.ExecuteTemplate(w, "ai/price-optimization", data); err != nil {
		Logger.Printf("Error rendering price optimization template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetSmartSearchPage shows the enhanced search page
func (h *AIHandler) GetSmartSearchPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get search query
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	// Prepare template data
	data := h.GetTemplateData()
	data["PageTitle"] = "Akıllı Arama - KolajAI"
	data["PageDescription"] = "AI destekli gelişmiş ürün arama"
	data["SearchQuery"] = query

	// If there's a query, perform search
	if query != "" {
		searchResult, err := h.aiService.SmartSearch(query, 24, 0)
		if err != nil {
			Logger.Printf("Error performing smart search: %v", err)
			data["SearchError"] = "Arama sırasında bir hata oluştu"
		} else {
			data["SearchResult"] = searchResult
		}
	}

	// Render template
	if err := h.Templates.ExecuteTemplate(w, "ai/smart-search", data); err != nil {
		Logger.Printf("Error rendering smart search template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
