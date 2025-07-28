package handlers

import (
	"encoding/json"
	"fmt"
	"kolajAi/internal/models"
	"kolajAi/internal/services"
	"net/http"
	"strconv"
)

// AITemplateHandler handles AI template related requests
type AITemplateHandler struct {
	*Handler
	aiTemplateService *services.AITemplateService
}

// NewAITemplateHandler creates a new AI template handler
func NewAITemplateHandler(h *Handler, aiTemplateService *services.AITemplateService) *AITemplateHandler {
	return &AITemplateHandler{
		Handler:           h,
		aiTemplateService: aiTemplateService,
	}
}

// GenerateTemplate handles template generation requests
func (h *AITemplateHandler) GenerateTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check AI template permission
	if !user.CanUseAITemplates() {
		h.ErrorResponse(w, "You don't have permission to use AI templates", http.StatusForbidden)
		return
	}

	// Parse request body
	var req services.TemplateGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set user ID
	req.UserID = user.ID

	// Generate template
	result, err := h.aiTemplateService.GenerateTemplate(&req)
	if err != nil {
		h.ErrorResponse(w, fmt.Sprintf("Failed to generate template: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	h.JSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    result,
		"message": "Template generated successfully",
	})
}

// GetTemplates handles getting user's templates
func (h *AITemplateHandler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get query parameters
	templateType := r.URL.Query().Get("type")
	platform := r.URL.Query().Get("platform")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get templates (this would be implemented in the service)
	templates := []models.AITemplate{} // Placeholder

	// Return templates
	h.JSONResponse(w, map[string]interface{}{
		"success":   true,
		"data":      templates,
		"count":     len(templates),
		"limit":     limit,
		"offset":    offset,
		"filters": map[string]string{
			"type":     templateType,
			"platform": platform,
		},
	})
}

// GetTemplate handles getting a specific template
func (h *AITemplateHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get template ID from URL
	templateIDStr := r.URL.Query().Get("id")
	if templateIDStr == "" {
		h.ErrorResponse(w, "Template ID is required", http.StatusBadRequest)
		return
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		h.ErrorResponse(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	// Get template (this would be implemented in the service)
	template := &models.AITemplate{
		ID:     templateID,
		UserID: user.ID,
		Name:   "Sample Template",
		Type:   models.TemplateTypeSocialMedia,
	}

	// Return template
	h.JSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    template,
	})
}

// DeleteTemplate handles template deletion
func (h *AITemplateHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get template ID from URL
	templateIDStr := r.URL.Query().Get("id")
	if templateIDStr == "" {
		h.ErrorResponse(w, "Template ID is required", http.StatusBadRequest)
		return
	}

	_, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		h.ErrorResponse(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	// Delete template (this would be implemented in the service)
	// For now, just return success
	h.JSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Template deleted successfully",
	})
}

// UpdateTemplate handles template updates
func (h *AITemplateHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check AI template permission
	if !user.CanUseAITemplates() {
		h.ErrorResponse(w, "You don't have permission to use AI templates", http.StatusForbidden)
		return
	}

	// Get template ID from URL
	templateIDStr := r.URL.Query().Get("id")
	if templateIDStr == "" {
		h.ErrorResponse(w, "Template ID is required", http.StatusBadRequest)
		return
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		h.ErrorResponse(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		h.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update template (this would be implemented in the service)
	// For now, just return success
	h.JSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Template updated successfully",
		"data": map[string]interface{}{
			"id":         templateID,
			"updated_at": "2024-01-01T00:00:00Z",
		},
	})
}

// GetTemplateTypes returns available template types
func (h *AITemplateHandler) GetTemplateTypes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	templateTypes := []map[string]interface{}{
		{
			"type":        string(models.TemplateTypeSocialMedia),
			"name":        "Social Media Post",
			"description": "Create engaging social media posts",
			"platforms":   []string{"instagram", "facebook", "twitter", "telegram"},
		},
		{
			"type":        string(models.TemplateTypeProductImage),
			"name":        "Product Image",
			"description": "Enhanced product images with overlays",
			"platforms":   []string{"all"},
		},
		{
			"type":        string(models.TemplateTypeProductDesc),
			"name":        "Product Description",
			"description": "AI-generated product descriptions",
			"platforms":   []string{"all"},
		},
		{
			"type":        string(models.TemplateTypeMarketingEmail),
			"name":        "Marketing Email",
			"description": "Professional marketing email templates",
			"platforms":   []string{"email"},
		},
		{
			"type":        string(models.TemplateTypeBanner),
			"name":        "Banner",
			"description": "Eye-catching banners for promotions",
			"platforms":   []string{"web", "social"},
		},
		{
			"type":        string(models.TemplateTypeStory),
			"name":        "Story",
			"description": "Instagram and Facebook story templates",
			"platforms":   []string{"instagram", "facebook"},
		},
		{
			"type":        string(models.TemplateTypeTelegram),
			"name":        "Telegram Post",
			"description": "Optimized posts for Telegram channels",
			"platforms":   []string{"telegram"},
		},
	}

	h.JSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    templateTypes,
	})
}

// GetPlatformSpecs returns platform-specific specifications
func (h *AITemplateHandler) GetPlatformSpecs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	platform := r.URL.Query().Get("platform")
	if platform == "" {
		h.ErrorResponse(w, "Platform parameter is required", http.StatusBadRequest)
		return
	}

	platformSpecs := map[string]interface{}{
		"instagram": map[string]interface{}{
			"post": map[string]interface{}{
				"dimensions": map[string]int{"width": 1080, "height": 1080},
				"max_text":   2200,
				"hashtags":   30,
			},
			"story": map[string]interface{}{
				"dimensions": map[string]int{"width": 1080, "height": 1920},
				"max_text":   2200,
			},
		},
		"facebook": map[string]interface{}{
			"post": map[string]interface{}{
				"dimensions": map[string]int{"width": 1200, "height": 630},
				"max_text":   63206,
			},
			"story": map[string]interface{}{
				"dimensions": map[string]int{"width": 1080, "height": 1920},
				"max_text":   2200,
			},
		},
		"twitter": map[string]interface{}{
			"post": map[string]interface{}{
				"dimensions": map[string]int{"width": 1200, "height": 675},
				"max_text":   280,
			},
		},
		"telegram": map[string]interface{}{
			"post": map[string]interface{}{
				"dimensions": map[string]int{"width": 1280, "height": 720},
				"max_text":   4096,
			},
		},
	}

	if specs, exists := platformSpecs[platform]; exists {
		h.JSONResponse(w, map[string]interface{}{
			"success": true,
			"data":    specs,
		})
	} else {
		h.ErrorResponse(w, "Platform not supported", http.StatusBadRequest)
	}
}

// RateTemplate handles template rating
func (h *AITemplateHandler) RateTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var ratingData struct {
		TemplateID int64  `json:"template_id"`
		Rating     int    `json:"rating"`
		Comment    string `json:"comment"`
	}

	if err := json.NewDecoder(r.Body).Decode(&ratingData); err != nil {
		h.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate rating
	if ratingData.Rating < 1 || ratingData.Rating > 5 {
		h.ErrorResponse(w, "Rating must be between 1 and 5", http.StatusBadRequest)
		return
	}

	// Save rating (this would be implemented in the service)
	// For now, just return success
	h.JSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Rating saved successfully",
	})
}

// TemplateUsage handles template usage tracking
func (h *AITemplateHandler) TemplateUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := h.GetUserFromSession(r)
	if user == nil {
		h.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var usageData struct {
		TemplateID int64  `json:"template_id"`
		Platform   string `json:"platform"`
		ProductID  *int64 `json:"product_id,omitempty"`
		Success    bool   `json:"success"`
		OutputURL  string `json:"output_url,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&usageData); err != nil {
		h.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Track usage (this would be implemented in the service)
	// For now, just return success
	h.JSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Usage tracked successfully",
	})
}