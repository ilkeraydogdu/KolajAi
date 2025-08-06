package handlers

import (
	"encoding/json"
	"fmt"
	"kolajAi/internal/services"
	"net/http"
	"strconv"
	"time"
)

// AIVisionHandler handles AI vision-related HTTP requests
type AIVisionHandler struct {
	*Handler
	aiVisionService *services.AIVisionService
}

// NewAIVisionHandler creates a new AI vision handler
func NewAIVisionHandler(base *Handler, aiVisionService *services.AIVisionService) *AIVisionHandler {
	return &AIVisionHandler{
		Handler:         base,
		aiVisionService: aiVisionService,
	}
}

// GetVisionDashboard displays the AI vision dashboard
func (h *AIVisionHandler) GetVisionDashboard(w http.ResponseWriter, r *http.Request) {
	// Get user ID from session
	userID, err := h.getUserIDFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user's image library
	library, err := h.aiVisionService.GetUserImageLibrary(userID)
	if err != nil {
		h.HandleError(w, r, fmt.Errorf("failed to get image library: %w", err), "Görsel kütüphanesi yüklenemedi")
		return
	}

	// Get user statistics
	stats, err := h.aiVisionService.GetUserImageStats(userID)
	if err != nil {
		h.HandleError(w, r, fmt.Errorf("failed to get user stats: %w", err), "İstatistikler yüklenemedi")
		return
	}

	// Get recent images (last 20)
	recentImages, err := h.aiVisionService.SmartImageSearch(services.SmartSearchQuery{
		UserID: userID,
		SortBy: "date",
		Limit:  20,
	})
	if err != nil {
		h.HandleError(w, r, fmt.Errorf("failed to get recent images: %w", err), "Son görseller yüklenemedi")
		return
	}

	data := h.GetTemplateData()
	data["Library"] = library
	data["Stats"] = stats
	data["RecentImages"] = recentImages.Images
	data["Title"] = "AI Görsel Yönetimi"
	data["PageName"] = "ai_vision_dashboard"

	if err := h.Templates.ExecuteTemplate(w, "ai_vision_dashboard.html", data); err != nil {
		h.HandleError(w, r, err, "Şablon yüklenemedi")
	}
}

// UploadImage handles image upload and AI analysis
func (h *AIVisionHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserIDFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if userID == 0 {
		h.WriteJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		h.WriteJSONError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("image")
	if err != nil {
		h.WriteJSONError(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Process the uploaded image with AI analysis
	result, err := h.aiVisionService.ProcessUploadedImage(userID, file, header)
	if err != nil {
		h.WriteJSONError(w, fmt.Sprintf("Failed to process image: %v", err), http.StatusBadRequest)
		return
	}

	// Log successful upload
	h.logAIOperation(userID, result.ImageID, "upload", "success", result.ProcessingTime, "")

	h.WriteJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Görsel başarıyla yüklendi ve analiz edildi",
		"data":    result,
	})
}

// SearchImages handles intelligent image search
func (h *AIVisionHandler) SearchImages(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if userID == 0 {
		h.WriteJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse search parameters
	query := services.SmartSearchQuery{
		UserID:        userID,
		Query:         r.URL.Query().Get("q"),
		Categories:    r.URL.Query()["category"],
		Tags:          r.URL.Query()["tag"],
		Colors:        r.URL.Query()["color"],
		QualityFilter: "any", // Default quality filter
		SizeFilter:    "any", // Default size filter
		SortBy:        "relevance",
		Limit:         50, // Default limit
		Offset:        0,
	}

	// Parse numeric parameters
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			query.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			query.Offset = offset
		}
	}

	if qualityStr := r.URL.Query().Get("quality"); qualityStr != "" {
		query.QualityFilter = qualityStr
	}

	if sizeStr := r.URL.Query().Get("size"); sizeStr != "" {
		query.SizeFilter = sizeStr
	}

	if sortStr := r.URL.Query().Get("sort"); sortStr != "" {
		query.SortBy = sortStr
	}

	// Parse date range
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			query.DateRange.StartDate = &startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			query.DateRange.EndDate = &endDate
		}
	}

	// Perform smart search
	results, err := h.aiVisionService.SmartImageSearch(query)
	if err != nil {
		h.WriteJSONError(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Log search operation
	h.logAIOperation(userID, "", "search", "success", results.ProcessTime, query.Query)

	h.WriteJSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    results,
	})
}

// GetImageAnalysis returns detailed analysis for a specific image
func (h *AIVisionHandler) GetImageAnalysis(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if userID == 0 {
		h.WriteJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	imageID := r.URL.Query().Get("image_id")
	if imageID == "" {
		h.WriteJSONError(w, "Image ID is required", http.StatusBadRequest)
		return
	}

	analysis, err := h.aiVisionService.GetImageAnalysis(userID, imageID)
	if err != nil {
		h.WriteJSONError(w, fmt.Sprintf("Failed to get analysis: %v", err), http.StatusNotFound)
		return
	}

	h.WriteJSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    analysis,
	})
}

// DeleteImage handles image deletion
func (h *AIVisionHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" && r.Method != "POST" {
		h.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserIDFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if userID == 0 {
		h.WriteJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	imageID := r.URL.Query().Get("image_id")
	if imageID == "" {
		h.WriteJSONError(w, "Image ID is required", http.StatusBadRequest)
		return
	}

	err = h.aiVisionService.DeleteImage(userID, imageID)
	if err != nil {
		h.WriteJSONError(w, fmt.Sprintf("Failed to delete image: %v", err), http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Görsel başarıyla silindi",
	})
}

// GetImagesByCategory returns images filtered by category
func (h *AIVisionHandler) GetImagesByCategory(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if userID == 0 {
		h.WriteJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	category := r.URL.Query().Get("category")
	if category == "" {
		h.WriteJSONError(w, "Category is required", http.StatusBadRequest)
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	images, err := h.aiVisionService.GetImagesByCategory(userID, category, limit, 0)
	if err != nil {
		h.WriteJSONError(w, fmt.Sprintf("Failed to get images: %v", err), http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    images,
	})
}

// GetImagesByTag returns images filtered by tag
func (h *AIVisionHandler) GetImagesByTag(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if userID == 0 {
		h.WriteJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tag := r.URL.Query().Get("tag")
	if tag == "" {
		h.WriteJSONError(w, "Tag is required", http.StatusBadRequest)
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	images, err := h.aiVisionService.GetImagesByTag(userID, tag, limit, 0)
	if err != nil {
		h.WriteJSONError(w, fmt.Sprintf("Failed to get images: %v", err), http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    images,
	})
}

// CreateCollection creates a new image collection
func (h *AIVisionHandler) CreateCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserIDFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if userID == 0 {
		h.WriteJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		ImageIDs    []string `json:"image_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	collection, err := h.aiVisionService.CreateImageCollection(userID, request.Name, request.Description, request.ImageIDs, false)
	if err != nil {
		h.WriteJSONError(w, fmt.Sprintf("Failed to create collection: %v", err), http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Koleksiyon başarıyla oluşturuldu",
		"data":    collection,
	})
}

// UpdateCollection updates an existing image collection
func (h *AIVisionHandler) UpdateCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" && r.Method != "POST" {
		h.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserIDFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if userID == 0 {
		h.WriteJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	collectionID := r.URL.Query().Get("collection_id")
	if collectionID == "" {
		h.WriteJSONError(w, "Collection ID is required", http.StatusBadRequest)
		return
	}

	var request struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		ImageIDs    []string `json:"image_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = h.aiVisionService.UpdateImageCollection(userID, collectionID, request.Name, request.Description, request.ImageIDs, false)
	if err != nil {
		h.WriteJSONError(w, fmt.Sprintf("Failed to update collection: %v", err), http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Koleksiyon başarıyla güncellendi",
	})
}

// DeleteCollection deletes an image collection
func (h *AIVisionHandler) DeleteCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" && r.Method != "POST" {
		h.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserIDFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if userID == 0 {
		h.WriteJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	collectionID := r.URL.Query().Get("collection_id")
	if collectionID == "" {
		h.WriteJSONError(w, "Collection ID is required", http.StatusBadRequest)
		return
	}

	err = h.aiVisionService.DeleteImageCollection(userID, collectionID)
	if err != nil {
		h.WriteJSONError(w, fmt.Sprintf("Failed to delete collection: %v", err), http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Koleksiyon başarıyla silindi",
	})
}

// SuggestCategories suggests product categories based on image analysis
func (h *AIVisionHandler) SuggestCategories(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if userID == 0 {
		h.WriteJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	imageID := r.URL.Query().Get("image_id")
	if imageID == "" {
		h.WriteJSONError(w, "Image ID is required", http.StatusBadRequest)
		return
	}

	suggestions, err := h.aiVisionService.SuggestProductCategories(imageID, userID)
	if err != nil {
		h.WriteJSONError(w, fmt.Sprintf("Failed to get suggestions: %v", err), http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    suggestions,
	})
}

// GetUserStats returns user's image statistics
func (h *AIVisionHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if userID == 0 {
		h.WriteJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats, err := h.aiVisionService.GetUserImageStats(userID)
	if err != nil {
		h.WriteJSONError(w, fmt.Sprintf("Failed to get stats: %v", err), http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// GetImageLibrary returns user's organized image library
func (h *AIVisionHandler) GetImageLibrary(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if userID == 0 {
		h.WriteJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	library, err := h.aiVisionService.GetUserImageLibrary(userID)
	if err != nil {
		h.WriteJSONError(w, fmt.Sprintf("Failed to get library: %v", err), http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, map[string]interface{}{
		"success": true,
		"data":    library,
	})
}

// RenderVisionUploadPage renders the image upload page
func (h *AIVisionHandler) RenderVisionUploadPage(w http.ResponseWriter, r *http.Request) {
	data := h.GetTemplateData()
	data["Title"] = "Görsel Yükle"
	data["PageName"] = "ai_vision_upload"

	if err := h.Templates.ExecuteTemplate(w, "ai_vision_upload.html", data); err != nil {
		h.HandleError(w, r, err, "Şablon yüklenemedi")
	}
}

// RenderVisionSearchPage renders the image search page
func (h *AIVisionHandler) RenderVisionSearchPage(w http.ResponseWriter, r *http.Request) {
	data := h.GetTemplateData()
	data["Title"] = "Görsel Arama"
	data["PageName"] = "ai_vision_search"

	if err := h.Templates.ExecuteTemplate(w, "ai_vision_search.html", data); err != nil {
		h.HandleError(w, r, err, "Şablon yüklenemedi")
	}
}

// RenderVisionGalleryPage renders the image gallery page
func (h *AIVisionHandler) RenderVisionGalleryPage(w http.ResponseWriter, r *http.Request) {
	data := h.GetTemplateData()
	data["Title"] = "Görsel Galerisi"
	data["PageName"] = "ai_vision_gallery"

	if err := h.Templates.ExecuteTemplate(w, "ai_vision_gallery.html", data); err != nil {
		h.HandleError(w, r, err, "Şablon yüklenemedi")
	}
}

// Helper method to log AI operations
func (h *AIVisionHandler) logAIOperation(userID int, imageID, operation, status string, processingTime time.Duration, metadata string) {
	// Log AI operation for monitoring and analytics
	// This would typically write to a log file or database
	// For now, we'll just use a simple log
	fmt.Printf("AI Operation: UserID=%d, ImageID=%s, Operation=%s, Status=%s, Time=%v, Metadata=%s\n",
		userID, imageID, operation, status, processingTime, metadata)
}

// Helper method to write JSON responses
func (h *AIVisionHandler) WriteJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// Helper method to write JSON error responses
func (h *AIVisionHandler) WriteJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

// getUserIDFromSession extracts user ID from session
func (h *AIVisionHandler) getUserIDFromSession(r *http.Request) (int, error) {
	session, err := h.SessionManager.GetSession(r)
	if err != nil || session.Values["user_id"] == nil {
		return 0, fmt.Errorf("no valid session")
	}
	
	userID, ok := session.Values["user_id"].(int)
	if !ok || userID == 0 {
		return 0, fmt.Errorf("invalid user ID")
	}
	
	return userID, nil
}