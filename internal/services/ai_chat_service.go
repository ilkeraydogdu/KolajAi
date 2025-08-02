package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"


	"kolajAi/internal/repository"
)

// AIChatService provides intelligent chat capabilities
type AIChatService struct {
	db           *sql.DB
	repo         *repository.BaseRepository
	aiService    *AIService
	openAIAPIKey string
	model        string
}

// ChatMessage represents a chat message
type ChatMessage struct {
	ID          string                 `json:"id"`
	SessionID   string                 `json:"session_id"`
	UserID      int64                  `json:"user_id"`
	Role        string                 `json:"role"` // user, assistant, system
	Content     string                 `json:"content"`
	MessageType string                 `json:"message_type"` // text, image, file, product_recommendation
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Tokens      int                    `json:"tokens,omitempty"`
}

// ChatSession represents a chat conversation session
type ChatSession struct {
	ID          string                 `json:"id"`
	UserID      int64                  `json:"user_id"`
	Title       string                 `json:"title"`
	Context     string                 `json:"context"` // shopping, support, general
	Status      string                 `json:"status"`  // active, archived, closed
	Messages    []ChatMessage          `json:"messages,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	LastMessage *time.Time             `json:"last_message,omitempty"`
}

// ChatRequest represents a chat request
type ChatRequest struct {
	SessionID   string                 `json:"session_id,omitempty"`
	UserID      int64                  `json:"user_id"`
	Message     string                 `json:"message"`
	Context     string                 `json:"context,omitempty"`
	MessageType string                 `json:"message_type,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	SessionID     string                 `json:"session_id"`
	MessageID     string                 `json:"message_id"`
	Content       string                 `json:"content"`
	MessageType   string                 `json:"message_type"`
	Suggestions   []string               `json:"suggestions,omitempty"`
	Products      []ProductSuggestion    `json:"products,omitempty"`
	Actions       []ChatAction           `json:"actions,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	TokensUsed    int                    `json:"tokens_used"`
	ResponseTime  time.Duration          `json:"response_time"`
}

// ProductSuggestion represents a product recommendation
type ProductSuggestion struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
	Rating      float64 `json:"rating"`
	Reason      string  `json:"reason"`
	Confidence  float64 `json:"confidence"`
}

// ChatAction represents an actionable item from chat
type ChatAction struct {
	Type        string                 `json:"type"` // add_to_cart, view_product, search, contact_support
	Label       string                 `json:"label"`
	URL         string                 `json:"url,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Priority    int                    `json:"priority"`
}

// ChatAnalytics represents chat analytics data
type ChatAnalytics struct {
	TotalSessions      int                    `json:"total_sessions"`
	ActiveSessions     int                    `json:"active_sessions"`
	AvgResponseTime    time.Duration          `json:"avg_response_time"`
	AvgSessionLength   time.Duration          `json:"avg_session_length"`
	TopIntents         []IntentAnalytics      `json:"top_intents"`
	SatisfactionScore  float64                `json:"satisfaction_score"`
	ConversionRate     float64                `json:"conversion_rate"`
	TokensUsed         int                    `json:"tokens_used"`
	CostAnalysis       map[string]interface{} `json:"cost_analysis"`
}

// IntentAnalytics represents intent analysis data
type IntentAnalytics struct {
	Intent     string  `json:"intent"`
	Count      int     `json:"count"`
	Confidence float64 `json:"confidence"`
	Success    float64 `json:"success_rate"`
}

// NewAIChatService creates a new AI chat service
func NewAIChatService(db *sql.DB, repo *repository.BaseRepository, aiService *AIService, openAIAPIKey string) *AIChatService {
	return &AIChatService{
		db:           db,
		repo:         repo,
		aiService:    aiService,
		openAIAPIKey: openAIAPIKey,
		model:        "gpt-4",
	}
}

// StartChat starts a new chat session or continues existing one
func (s *AIChatService) StartChat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	startTime := time.Now()

	// Get or create session
	session, err := s.getOrCreateSession(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create session: %w", err)
	}

	// Save user message
	userMessage := &ChatMessage{
		ID:          s.generateMessageID(),
		SessionID:   session.ID,
		UserID:      req.UserID,
		Role:        "user",
		Content:     req.Message,
		MessageType: s.getMessageType(req.MessageType),
		Metadata:    req.Metadata,
		Timestamp:   time.Now(),
	}

	if err := s.saveMessage(ctx, userMessage); err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// Get conversation context
	context, err := s.buildConversationContext(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to build context: %w", err)
	}

	// Generate AI response
	aiResponse, err := s.generateAIResponse(ctx, context, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AI response: %w", err)
	}

	// Save AI message
	aiMessage := &ChatMessage{
		ID:          s.generateMessageID(),
		SessionID:   session.ID,
		UserID:      req.UserID,
		Role:        "assistant",
		Content:     aiResponse.Content,
		MessageType: aiResponse.MessageType,
		Metadata:    aiResponse.Metadata,
		Timestamp:   time.Now(),
		Tokens:      aiResponse.TokensUsed,
	}

	if err := s.saveMessage(ctx, aiMessage); err != nil {
		return nil, fmt.Errorf("failed to save AI message: %w", err)
	}

	// Update session
	if err := s.updateSession(ctx, session.ID); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	// Add response time
	aiResponse.ResponseTime = time.Since(startTime)
	aiResponse.SessionID = session.ID
	aiResponse.MessageID = aiMessage.ID

	return aiResponse, nil
}

// GetChatHistory retrieves chat history for a session
func (s *AIChatService) GetChatHistory(ctx context.Context, sessionID string, limit int) (*ChatSession, error) {
	session, err := s.getSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	messages, err := s.getSessionMessages(ctx, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	session.Messages = messages
	return session, nil
}

// GetUserSessions retrieves all chat sessions for a user
func (s *AIChatService) GetUserSessions(ctx context.Context, userID int64, limit int) ([]ChatSession, error) {
	query := `
		SELECT id, user_id, title, context, status, metadata, created_at, updated_at, last_message
		FROM chat_sessions 
		WHERE user_id = ? 
		ORDER BY updated_at DESC 
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []ChatSession
	for rows.Next() {
		var session ChatSession
		var metadataJSON string
		var lastMessage sql.NullTime

		err := rows.Scan(
			&session.ID, &session.UserID, &session.Title, &session.Context,
			&session.Status, &metadataJSON, &session.CreatedAt,
			&session.UpdatedAt, &lastMessage,
		)
		if err != nil {
			continue
		}

		if metadataJSON != "" {
			json.Unmarshal([]byte(metadataJSON), &session.Metadata)
		}

		if lastMessage.Valid {
			session.LastMessage = &lastMessage.Time
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// AnalyzeIntent analyzes user message intent
func (s *AIChatService) AnalyzeIntent(ctx context.Context, message string, context string) (*IntentAnalytics, error) {
	// Prepare intent analysis prompt
	prompt := s.buildIntentAnalysisPrompt(message, context)

	// Call AI service for intent analysis
	response, err := s.callOpenAI(ctx, prompt, "gpt-3.5-turbo", 150)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze intent: %w", err)
	}

	// Parse intent response
	intent, err := s.parseIntentResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse intent: %w", err)
	}

	return intent, nil
}

// GetProductRecommendations gets AI-powered product recommendations
func (s *AIChatService) GetProductRecommendations(ctx context.Context, userID int64, message string, limit int) ([]ProductSuggestion, error) {
	// Get user preferences and history
	userProfile, err := s.getUserProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Build recommendation prompt
	prompt := s.buildRecommendationPrompt(message, userProfile)

	// Get AI recommendations
	response, err := s.callOpenAI(ctx, prompt, s.model, 500)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations: %w", err)
	}

	// Parse and fetch products
	recommendations, err := s.parseRecommendations(ctx, response, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to parse recommendations: %w", err)
	}

	return recommendations, nil
}

// GetChatAnalytics retrieves chat analytics
func (s *AIChatService) GetChatAnalytics(ctx context.Context, startDate, endDate time.Time) (*ChatAnalytics, error) {
	analytics := &ChatAnalytics{}

	// Get total sessions
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM chat_sessions 
		WHERE created_at BETWEEN ? AND ?
	`, startDate, endDate).Scan(&analytics.TotalSessions)
	if err != nil {
		return nil, fmt.Errorf("failed to get total sessions: %w", err)
	}

	// Get active sessions
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM chat_sessions 
		WHERE status = 'active' AND updated_at BETWEEN ? AND ?
	`, startDate, endDate).Scan(&analytics.ActiveSessions)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}

	// Calculate average response time and session length
	analytics.AvgResponseTime, analytics.AvgSessionLength = s.calculateAverages(ctx, startDate, endDate)

	// Get top intents
	analytics.TopIntents, err = s.getTopIntents(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get top intents: %w", err)
	}

	// Calculate satisfaction and conversion rates
	analytics.SatisfactionScore = s.calculateSatisfactionScore(ctx, startDate, endDate)
	analytics.ConversionRate = s.calculateConversionRate(ctx, startDate, endDate)

	// Get token usage
	err = s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(tokens), 0) FROM chat_messages 
		WHERE timestamp BETWEEN ? AND ? AND role = 'assistant'
	`, startDate, endDate).Scan(&analytics.TokensUsed)
	if err != nil {
		return nil, fmt.Errorf("failed to get token usage: %w", err)
	}

	// Calculate cost analysis
	analytics.CostAnalysis = s.calculateCostAnalysis(analytics.TokensUsed)

	return analytics, nil
}

// Private helper methods

func (s *AIChatService) getOrCreateSession(ctx context.Context, req *ChatRequest) (*ChatSession, error) {
	if req.SessionID != "" {
		session, err := s.getSession(ctx, req.SessionID)
		if err == nil {
			return session, nil
		}
	}

	// Create new session
	session := &ChatSession{
		ID:        s.generateSessionID(),
		UserID:    req.UserID,
		Title:     s.generateSessionTitle(req.Message),
		Context:   s.getContext(req.Context),
		Status:    "active",
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.saveSession(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *AIChatService) getSession(ctx context.Context, sessionID string) (*ChatSession, error) {
	query := `
		SELECT id, user_id, title, context, status, metadata, created_at, updated_at, last_message
		FROM chat_sessions WHERE id = ?
	`

	var session ChatSession
	var metadataJSON string
	var lastMessage sql.NullTime

	err := s.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID, &session.UserID, &session.Title, &session.Context,
		&session.Status, &metadataJSON, &session.CreatedAt,
		&session.UpdatedAt, &lastMessage,
	)
	if err != nil {
		return nil, err
	}

	if metadataJSON != "" {
		json.Unmarshal([]byte(metadataJSON), &session.Metadata)
	}

	if lastMessage.Valid {
		session.LastMessage = &lastMessage.Time
	}

	return &session, nil
}

func (s *AIChatService) saveSession(ctx context.Context, session *ChatSession) error {
	metadataJSON, _ := json.Marshal(session.Metadata)

	query := `
		INSERT INTO chat_sessions (id, user_id, title, context, status, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.Title, session.Context,
		session.Status, string(metadataJSON), session.CreatedAt, session.UpdatedAt,
	)

	return err
}

func (s *AIChatService) saveMessage(ctx context.Context, message *ChatMessage) error {
	metadataJSON, _ := json.Marshal(message.Metadata)

	query := `
		INSERT INTO chat_messages (id, session_id, user_id, role, content, message_type, metadata, timestamp, tokens)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		message.ID, message.SessionID, message.UserID, message.Role,
		message.Content, message.MessageType, string(metadataJSON),
		message.Timestamp, message.Tokens,
	)

	return err
}

func (s *AIChatService) updateSession(ctx context.Context, sessionID string) error {
	query := `UPDATE chat_sessions SET updated_at = ?, last_message = ? WHERE id = ?`
	now := time.Now()
	_, err := s.db.ExecContext(ctx, query, now, now, sessionID)
	return err
}

func (s *AIChatService) buildConversationContext(ctx context.Context, session *ChatSession) (string, error) {
	messages, err := s.getSessionMessages(ctx, session.ID, 10) // Last 10 messages
	if err != nil {
		return "", err
	}

	var contextBuilder strings.Builder
	contextBuilder.WriteString(fmt.Sprintf("Chat Context: %s\n", session.Context))
	contextBuilder.WriteString("Recent conversation:\n")

	for _, msg := range messages {
		contextBuilder.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
	}

	return contextBuilder.String(), nil
}

func (s *AIChatService) generateAIResponse(ctx context.Context, context string, req *ChatRequest) (*ChatResponse, error) {
	// Build comprehensive prompt
	prompt := s.buildChatPrompt(context, req.Message, req.Context)

	// Call OpenAI
	response, err := s.callOpenAI(ctx, prompt, s.model, 1000)
	if err != nil {
		return nil, err
	}

	// Parse response and add enhancements
	chatResponse := &ChatResponse{
		Content:     response,
		MessageType: "text",
		TokensUsed:  s.estimateTokens(prompt + response),
		Metadata:    make(map[string]interface{}),
	}

	// Add product recommendations if shopping context
	if req.Context == "shopping" {
		products, _ := s.GetProductRecommendations(ctx, req.UserID, req.Message, 3)
		chatResponse.Products = products
	}

	// Add suggested actions
	chatResponse.Actions = s.generateChatActions(req.Message, req.Context)

	// Add conversation suggestions
	chatResponse.Suggestions = s.generateSuggestions(req.Message, req.Context)

	return chatResponse, nil
}

func (s *AIChatService) getSessionMessages(ctx context.Context, sessionID string, limit int) ([]ChatMessage, error) {
	query := `
		SELECT id, session_id, user_id, role, content, message_type, metadata, timestamp, tokens
		FROM chat_messages 
		WHERE session_id = ? 
		ORDER BY timestamp DESC 
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var message ChatMessage
		var metadataJSON string

		err := rows.Scan(
			&message.ID, &message.SessionID, &message.UserID, &message.Role,
			&message.Content, &message.MessageType, &metadataJSON,
			&message.Timestamp, &message.Tokens,
		)
		if err != nil {
			continue
		}

		if metadataJSON != "" {
			json.Unmarshal([]byte(metadataJSON), &message.Metadata)
		}

		messages = append(messages, message)
	}

	// Reverse to get chronological order
	for i := len(messages)/2 - 1; i >= 0; i-- {
		opp := len(messages) - 1 - i
		messages[i], messages[opp] = messages[opp], messages[i]
	}

	return messages, nil
}

// Utility methods
func (s *AIChatService) generateSessionID() string {
	return fmt.Sprintf("chat_%d", time.Now().UnixNano())
}

func (s *AIChatService) generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

func (s *AIChatService) generateSessionTitle(message string) string {
	if len(message) > 50 {
		return message[:47] + "..."
	}
	return message
}

func (s *AIChatService) getMessageType(messageType string) string {
	if messageType != "" {
		return messageType
	}
	return "text"
}

func (s *AIChatService) getContext(context string) string {
	if context != "" {
		return context
	}
	return "general"
}

func (s *AIChatService) estimateTokens(text string) int {
	// Rough estimation: ~4 characters per token
	return len(text) / 4
}

// AI prompt builders and other helper methods would continue here...
// This is a comprehensive foundation for the AI chat system

func (s *AIChatService) buildChatPrompt(context, message, chatContext string) string {
	return fmt.Sprintf(`
You are KolajAI, an intelligent shopping assistant for an enterprise marketplace.

Context: %s
Chat Context: %s

User Message: %s

Provide helpful, accurate, and engaging responses. If discussing products, be specific about features and benefits. Always maintain a professional yet friendly tone.

Response:`, context, chatContext, message)
}

func (s *AIChatService) callOpenAI(_ context.Context, _, _ string, _ int) (string, error) {
	// This would integrate with OpenAI API
	// For now, return a mock response
	return "I'm KolajAI, your intelligent shopping assistant. How can I help you today?", nil
}

func (s *AIChatService) generateChatActions(message, context string) []ChatAction {
	actions := []ChatAction{}
	
	if strings.Contains(strings.ToLower(message), "product") || context == "shopping" {
		actions = append(actions, ChatAction{
			Type:     "search",
			Label:    "Search Products",
			URL:      "/products/search",
			Priority: 1,
		})
	}

	return actions
}

func (s *AIChatService) generateSuggestions(_, _ string) []string {
	suggestions := []string{
		"Can you recommend similar products?",
		"What are the best deals today?",
		"Help me find products in my budget",
		"I need help with my order",
		"How do I return a product?",
		"Contact customer support",
	}

	return suggestions
}

// Additional helper methods would be implemented here...
func (s *AIChatService) buildIntentAnalysisPrompt(_, _ string) string { return "" }
func (s *AIChatService) parseIntentResponse(response string) (*IntentAnalytics, error) { return nil, nil }
func (s *AIChatService) getUserProfile(ctx context.Context, userID int64) (map[string]interface{}, error) { return nil, nil }
func (s *AIChatService) buildRecommendationPrompt(message string, profile map[string]interface{}) string { return "" }
func (s *AIChatService) parseRecommendations(ctx context.Context, response string, limit int) ([]ProductSuggestion, error) { return nil, nil }
func (s *AIChatService) calculateAverages(ctx context.Context, start, end time.Time) (time.Duration, time.Duration) { return 0, 0 }
func (s *AIChatService) getTopIntents(ctx context.Context, start, end time.Time) ([]IntentAnalytics, error) { return nil, nil }
func (s *AIChatService) calculateSatisfactionScore(ctx context.Context, start, end time.Time) float64 { return 0.0 }
func (s *AIChatService) calculateConversionRate(ctx context.Context, start, end time.Time) float64 { return 0.0 }
func (s *AIChatService) calculateCostAnalysis(tokens int) map[string]interface{} { return nil }