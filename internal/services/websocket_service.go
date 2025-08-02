package services

import (

	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketService manages real-time WebSocket connections
type WebSocketService struct {
	clients    map[string]*Client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
	upgrader   websocket.Upgrader
}

// Client represents a WebSocket client connection
type Client struct {
	ID       string
	UserID   int64
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *WebSocketService
	Channels map[string]bool // Subscribed channels
	Metadata map[string]interface{}
	LastSeen time.Time
}

// Message represents a WebSocket message
type Message struct {
	Type      string                 `json:"type"`
	Channel   string                 `json:"channel,omitempty"`
	UserID    int64                  `json:"user_id,omitempty"`
	Data      interface{}            `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	MessageID string                 `json:"message_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MessageType constants
const (
	MessageTypeChat         = "chat"
	MessageTypeNotification = "notification"
	MessageTypeOrderUpdate  = "order_update"
	MessageTypeProductView  = "product_view"
	MessageTypeUserOnline   = "user_online"
	MessageTypeUserOffline  = "user_offline"
	MessageTypeTyping       = "typing"
	MessageTypeHeartbeat    = "heartbeat"
	MessageTypeSubscribe    = "subscribe"
	MessageTypeUnsubscribe  = "unsubscribe"
	MessageTypeError        = "error"
	MessageTypeSuccess      = "success"
)

// Channel constants
const (
	ChannelGlobal        = "global"
	ChannelUserPrefix    = "user_"
	ChannelChatPrefix    = "chat_"
	ChannelOrderPrefix   = "order_"
	ChannelProductPrefix = "product_"
	ChannelAdminPrefix   = "admin_"
)

// ConnectionStats represents WebSocket connection statistics
type ConnectionStats struct {
	TotalConnections    int                    `json:"total_connections"`
	ActiveConnections   int                    `json:"active_connections"`
	ConnectionsByUser   map[int64]int          `json:"connections_by_user"`
	MessagesSent        int64                  `json:"messages_sent"`
	MessagesReceived    int64                  `json:"messages_received"`
	ChannelSubscriptions map[string]int        `json:"channel_subscriptions"`
	Uptime              time.Duration          `json:"uptime"`
	LastActivity        time.Time              `json:"last_activity"`
	ErrorCount          int64                  `json:"error_count"`
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService() *WebSocketService {
	return &WebSocketService{
		clients:    make(map[string]*Client),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
		},
	}
}

// Start starts the WebSocket service
func (ws *WebSocketService) Start() {
	go ws.run()
	log.Println("WebSocket service started")
}

// HandleWebSocket handles WebSocket upgrade and client management
func (ws *WebSocketService) HandleWebSocket(w http.ResponseWriter, r *http.Request, userID int64) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		ID:       ws.generateClientID(),
		UserID:   userID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      ws,
		Channels: make(map[string]bool),
		Metadata: make(map[string]interface{}),
		LastSeen: time.Now(),
	}

	// Register client
	ws.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump()
}

// SendToUser sends a message to a specific user
func (ws *WebSocketService) SendToUser(userID int64, message *Message) error {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	sent := false
	for _, client := range ws.clients {
		if client.UserID == userID {
			select {
			case client.Send <- messageBytes:
				sent = true
			default:
				close(client.Send)
				delete(ws.clients, client.ID)
			}
		}
	}

	if !sent {
		return fmt.Errorf("user %d not connected", userID)
	}

	return nil
}

// SendToChannel sends a message to all clients subscribed to a channel
func (ws *WebSocketService) SendToChannel(channel string, message *Message) error {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	for _, client := range ws.clients {
		if client.Channels[channel] {
			select {
			case client.Send <- messageBytes:
			default:
				close(client.Send)
				delete(ws.clients, client.ID)
			}
		}
	}

	return nil
}

// BroadcastToAll sends a message to all connected clients
func (ws *WebSocketService) BroadcastToAll(message *Message) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	select {
	case ws.broadcast <- messageBytes:
		return nil
	default:
		return fmt.Errorf("broadcast channel full")
	}
}

// GetConnectionStats returns current connection statistics
func (ws *WebSocketService) GetConnectionStats() *ConnectionStats {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	stats := &ConnectionStats{
		TotalConnections:     len(ws.clients),
		ActiveConnections:    len(ws.clients),
		ConnectionsByUser:    make(map[int64]int),
		ChannelSubscriptions: make(map[string]int),
		LastActivity:         time.Now(),
	}

	// Count connections by user
	for _, client := range ws.clients {
		stats.ConnectionsByUser[client.UserID]++
	}

	// Count channel subscriptions
	for _, client := range ws.clients {
		for channel := range client.Channels {
			stats.ChannelSubscriptions[channel]++
		}
	}

	return stats
}

// GetConnectedUsers returns list of connected user IDs
func (ws *WebSocketService) GetConnectedUsers() []int64 {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	userMap := make(map[int64]bool)
	for _, client := range ws.clients {
		userMap[client.UserID] = true
	}

	users := make([]int64, 0, len(userMap))
	for userID := range userMap {
		users = append(users, userID)
	}

	return users
}

// IsUserOnline checks if a user is currently online
func (ws *WebSocketService) IsUserOnline(userID int64) bool {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	for _, client := range ws.clients {
		if client.UserID == userID {
			return true
		}
	}

	return false
}

// SendChatMessage sends a chat message
func (ws *WebSocketService) SendChatMessage(sessionID string, fromUserID, toUserID int64, content string, messageType string) error {
	message := &Message{
		Type:      MessageTypeChat,
		Channel:   ChannelChatPrefix + sessionID,
		UserID:    fromUserID,
		Timestamp: time.Now(),
		MessageID: ws.generateMessageID(),
		Data: map[string]interface{}{
			"session_id":   sessionID,
			"from_user_id": fromUserID,
			"to_user_id":   toUserID,
			"content":      content,
			"message_type": messageType,
		},
	}

	// Send to both users
	if err := ws.SendToUser(fromUserID, message); err != nil {
		log.Printf("Failed to send message to sender %d: %v", fromUserID, err)
	}

	if err := ws.SendToUser(toUserID, message); err != nil {
		log.Printf("Failed to send message to recipient %d: %v", toUserID, err)
	}

	return nil
}

// SendNotification sends a real-time notification
func (ws *WebSocketService) SendNotification(userID int64, notification interface{}) error {
	message := &Message{
		Type:      MessageTypeNotification,
		Channel:   ChannelUserPrefix + fmt.Sprintf("%d", userID),
		UserID:    userID,
		Data:      notification,
		Timestamp: time.Now(),
		MessageID: ws.generateMessageID(),
	}

	return ws.SendToUser(userID, message)
}

// SendOrderUpdate sends an order status update
func (ws *WebSocketService) SendOrderUpdate(userID int64, orderID uint, status string, details interface{}) error {
	message := &Message{
		Type:      MessageTypeOrderUpdate,
		Channel:   ChannelOrderPrefix + fmt.Sprintf("%d", orderID),
		UserID:    userID,
		Timestamp: time.Now(),
		MessageID: ws.generateMessageID(),
		Data: map[string]interface{}{
			"order_id": orderID,
			"status":   status,
			"details":  details,
		},
	}

	return ws.SendToUser(userID, message)
}

// SendTypingIndicator sends typing indicator
func (ws *WebSocketService) SendTypingIndicator(sessionID string, userID int64, isTyping bool) error {
	message := &Message{
		Type:      MessageTypeTyping,
		Channel:   ChannelChatPrefix + sessionID,
		UserID:    userID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"session_id": sessionID,
			"user_id":    userID,
			"is_typing":  isTyping,
		},
	}

	return ws.SendToChannel(ChannelChatPrefix+sessionID, message)
}

// Private methods

func (ws *WebSocketService) run() {
	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-ws.register:
			ws.registerClient(client)

		case client := <-ws.unregister:
			ws.unregisterClient(client)

		case message := <-ws.broadcast:
			ws.broadcastToClients(message)

		case <-ticker.C:
			ws.pingClients()
		}
	}
}

func (ws *WebSocketService) registerClient(client *Client) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	ws.clients[client.ID] = client

	// Subscribe to user's personal channel
	client.Channels[ChannelUserPrefix+fmt.Sprintf("%d", client.UserID)] = true

	// Subscribe to global channel
	client.Channels[ChannelGlobal] = true

	log.Printf("Client %s (User %d) connected. Total clients: %d", client.ID, client.UserID, len(ws.clients))

	// Send welcome message
	welcomeMessage := &Message{
		Type:      MessageTypeSuccess,
		Data:      map[string]interface{}{"message": "Connected successfully"},
		Timestamp: time.Now(),
	}

	messageBytes, _ := json.Marshal(welcomeMessage)
	select {
	case client.Send <- messageBytes:
	default:
		close(client.Send)
		delete(ws.clients, client.ID)
	}

	// Notify others about user online status
	ws.notifyUserStatus(client.UserID, true)
}

func (ws *WebSocketService) unregisterClient(client *Client) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if _, ok := ws.clients[client.ID]; ok {
		delete(ws.clients, client.ID)
		close(client.Send)

		log.Printf("Client %s (User %d) disconnected. Total clients: %d", client.ID, client.UserID, len(ws.clients))

		// Check if user is still online with other connections
		userStillOnline := false
		for _, c := range ws.clients {
			if c.UserID == client.UserID {
				userStillOnline = true
				break
			}
		}

		if !userStillOnline {
			ws.notifyUserStatus(client.UserID, false)
		}
	}
}

func (ws *WebSocketService) broadcastToClients(message []byte) {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	for clientID, client := range ws.clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(ws.clients, clientID)
		}
	}
}

func (ws *WebSocketService) pingClients() {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	pingMessage := &Message{
		Type:      MessageTypeHeartbeat,
		Timestamp: time.Now(),
	}

	messageBytes, _ := json.Marshal(pingMessage)

	for clientID, client := range ws.clients {
		select {
		case client.Send <- messageBytes:
		default:
			close(client.Send)
			delete(ws.clients, clientID)
		}
	}
}

func (ws *WebSocketService) notifyUserStatus(userID int64, isOnline bool) {
	messageType := MessageTypeUserOnline
	if !isOnline {
		messageType = MessageTypeUserOffline
	}

	statusMessage := &Message{
		Type:      messageType,
		Channel:   ChannelGlobal,
		UserID:    userID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"user_id":   userID,
			"is_online": isOnline,
		},
	}

	ws.SendToChannel(ChannelGlobal, statusMessage)
}

func (ws *WebSocketService) generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}

func (ws *WebSocketService) generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// Client methods

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		c.LastSeen = time.Now()
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.LastSeen = time.Now()
		c.handleMessage(messageBytes)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(messageBytes []byte) {
	var message Message
	if err := json.Unmarshal(messageBytes, &message); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	switch message.Type {
	case MessageTypeSubscribe:
		c.handleSubscribe(&message)
	case MessageTypeUnsubscribe:
		c.handleUnsubscribe(&message)
	case MessageTypeChat:
		c.handleChatMessage(&message)
	case MessageTypeTyping:
		c.handleTypingIndicator(&message)
	default:
		log.Printf("Unknown message type: %s", message.Type)
	}
}

func (c *Client) handleSubscribe(message *Message) {
	if channel, ok := message.Data.(string); ok {
		c.Channels[channel] = true
		log.Printf("Client %s subscribed to channel %s", c.ID, channel)
	}
}

func (c *Client) handleUnsubscribe(message *Message) {
	if channel, ok := message.Data.(string); ok {
		delete(c.Channels, channel)
		log.Printf("Client %s unsubscribed from channel %s", c.ID, channel)
	}
}

func (c *Client) handleChatMessage(message *Message) {
	// Process chat message
	if data, ok := message.Data.(map[string]interface{}); ok {
		if sessionID, ok := data["session_id"].(string); ok {
			// Broadcast to chat channel
			c.Hub.SendToChannel(ChannelChatPrefix+sessionID, message)
		}
	}
}

func (c *Client) handleTypingIndicator(message *Message) {
	// Process typing indicator
	if data, ok := message.Data.(map[string]interface{}); ok {
		if sessionID, ok := data["session_id"].(string); ok {
			// Broadcast typing indicator to chat channel
			c.Hub.SendToChannel(ChannelChatPrefix+sessionID, message)
		}
	}
}