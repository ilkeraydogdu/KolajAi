package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"kolajAi/internal/models"
	"kolajAi/internal/services"
)

// AIAdvancedHandler handles advanced AI-related requests
type AIAdvancedHandler struct {
	*Handler
	aiAdvancedService *services.AIAdvancedService
}

// NewAIAdvancedHandler creates a new advanced AI handler
func NewAIAdvancedHandler(h *Handler, aiAdvancedService *services.AIAdvancedService) *AIAdvancedHandler {
	return &AIAdvancedHandler{
		Handler:           h,
		aiAdvancedService: aiAdvancedService,
	}
}

// GenerateProductImage handles AI image generation requests
func (h *AIAdvancedHandler) GenerateProductImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	user, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user has admin role for this feature
	if user.Role != "admin" && user.Role != "seller" {
		http.Error(w, "Forbidden: This feature is only available for admin and seller users", http.StatusForbidden)
		return
	}

	// Parse request body
	var req services.AIImageGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set user ID
	req.UserID = int(user.ID)

	// Check AI credits
	credits, err := h.aiAdvancedService.GetAICreditsBalance(int(user.ID))
	if err != nil {
		Logger.Printf("Error checking AI credits: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if credits < 40 { // Minimum credits needed for image generation
		http.Error(w, "Insufficient AI credits", http.StatusPaymentRequired)
		return
	}

	// Generate image
	resp, err := h.aiAdvancedService.GenerateProductImages(req)
	if err != nil {
		Logger.Printf("Error generating image: %v", err)
		http.Error(w, "Failed to generate image", http.StatusInternalServerError)
		return
	}

	// Deduct credits
	if err := h.aiAdvancedService.DeductAICredits(int(user.ID), resp.Credits); err != nil {
		Logger.Printf("Error deducting credits: %v", err)
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GenerateContent handles AI content generation requests
func (h *AIAdvancedHandler) GenerateContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	user, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req services.AIContentGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set user ID
	req.UserID = int(user.ID)

	// Check AI credits
	credits, err := h.aiAdvancedService.GetAICreditsBalance(int(user.ID))
	if err != nil {
		Logger.Printf("Error checking AI credits: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if credits < 10 { // Minimum credits needed for content generation
		http.Error(w, "Insufficient AI credits", http.StatusPaymentRequired)
		return
	}

	// Generate content
	resp, err := h.aiAdvancedService.GenerateContent(req)
	if err != nil {
		Logger.Printf("Error generating content: %v", err)
		http.Error(w, "Failed to generate content", http.StatusInternalServerError)
		return
	}

	// Deduct credits
	if err := h.aiAdvancedService.DeductAICredits(int(user.ID), resp.Credits); err != nil {
		Logger.Printf("Error deducting credits: %v", err)
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateAITemplate handles AI template creation requests
func (h *AIAdvancedHandler) CreateAITemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	user, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only admin users can create templates
	if user.Role != "admin" {
		http.Error(w, "Forbidden: Only admin users can create AI templates", http.StatusForbidden)
		return
	}

	// Parse request body
	var req struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create template
	template, err := h.aiAdvancedService.CreateAITemplate(int(user.ID), req.Type, req.Name)
	if err != nil {
		Logger.Printf("Error creating AI template: %v", err)
		http.Error(w, "Failed to create template", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

// StartAIChat handles AI chat session initialization
func (h *AIAdvancedHandler) StartAIChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	user, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		Context string `json:"context"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start chat session
	session, err := h.aiAdvancedService.StartChatSession(int(user.ID), req.Context)
	if err != nil {
		Logger.Printf("Error starting AI chat session: %v", err)
		http.Error(w, "Failed to start chat session", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// SendAIChatMessage handles AI chat messages
func (h *AIAdvancedHandler) SendAIChatMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	user, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		SessionID string `json:"session_id"`
		Message   string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check AI credits
	credits, err := h.aiAdvancedService.GetAICreditsBalance(int(user.ID))
	if err != nil {
		Logger.Printf("Error checking AI credits: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if credits < 5 { // Minimum credits needed for chat
		http.Error(w, "Insufficient AI credits", http.StatusPaymentRequired)
		return
	}

	// Send message and get response
	response, err := h.aiAdvancedService.SendChatMessage(req.SessionID, req.Message)
	if err != nil {
		Logger.Printf("Error sending AI chat message: %v", err)
		http.Error(w, "Failed to process message", http.StatusInternalServerError)
		return
	}

	// Deduct credits
	if err := h.aiAdvancedService.DeductAICredits(int(user.ID), 5); err != nil {
		Logger.Printf("Error deducting credits: %v", err)
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AnalyzeProductImage handles AI product image analysis
func (h *AIAdvancedHandler) AnalyzeProductImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	user, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		ImageURL string `json:"image_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check AI credits
	credits, err := h.aiAdvancedService.GetAICreditsBalance(int(user.ID))
	if err != nil {
		Logger.Printf("Error checking AI credits: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if credits < 15 { // Credits needed for image analysis
		http.Error(w, "Insufficient AI credits", http.StatusPaymentRequired)
		return
	}

	// Analyze image
	analysis, err := h.aiAdvancedService.AnalyzeProductImage(req.ImageURL)
	if err != nil {
		Logger.Printf("Error analyzing image: %v", err)
		http.Error(w, "Failed to analyze image", http.StatusInternalServerError)
		return
	}

	// Deduct credits
	if err := h.aiAdvancedService.DeductAICredits(int(user.ID), 15); err != nil {
		Logger.Printf("Error deducting credits: %v", err)
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysis)
}

// GetAICredits returns the user's AI credit balance
func (h *AIAdvancedHandler) GetAICredits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	user, err := h.getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get credits
	credits, err := h.aiAdvancedService.GetAICreditsBalance(int(user.ID))
	if err != nil {
		Logger.Printf("Error getting AI credits: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"credits": credits,
		"user_id": user.ID,
	})
}

// getUserFromSession is a helper function to get user from session
func (h *AIAdvancedHandler) getUserFromSession(r *http.Request) (*models.User, error) {
	if !h.IsAuthenticated(r) {
		return nil, fmt.Errorf("user not authenticated")
	}

	session, err := h.SessionManager.GetSession(r)
	if err != nil {
		return nil, fmt.Errorf("session error: %w", err)
	}

	userInterface, ok := session.Values[UserKey]
	if !ok {
		return nil, fmt.Errorf("user not found in session")
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		return nil, fmt.Errorf("invalid user data")
	}

	return user, nil
}