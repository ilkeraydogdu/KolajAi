package session

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

// SessionManager handles advanced session management
type SessionManager struct {
	db           *sql.DB
	cookieName   string
	secure       bool
	httpOnly     bool
	sameSite     http.SameSite
	maxAge       int
	domain       string
	path         string
	encryptKey   []byte
	mu           sync.RWMutex
	cleanupTimer *time.Timer
}

// SessionData represents session data structure
type SessionData struct {
	ID           string                 `json:"id"`
	UserID       int64                  `json:"user_id"`
	UserAgent    string                 `json:"user_agent"`
	IPAddress    string                 `json:"ip_address"`
	LoginTime    time.Time              `json:"login_time"`
	LastActivity time.Time              `json:"last_activity"`
	ExpiresAt    time.Time              `json:"expires_at"`
	IsActive     bool                   `json:"is_active"`
	DeviceInfo   map[string]interface{} `json:"device_info"`
	Permissions  []string               `json:"permissions"`
	Preferences  map[string]interface{} `json:"preferences"`
	Data         map[string]interface{} `json:"data"`
}

// SessionConfig holds session configuration
type SessionConfig struct {
	CookieName     string
	Secure         bool
	HTTPOnly       bool
	SameSite       http.SameSite
	MaxAge         int
	Domain         string
	Path           string
	EncryptionKey  string
	CleanupInterval time.Duration
}

// NewSessionManager creates a new advanced session manager
func NewSessionManager(db *sql.DB, config SessionConfig) (*SessionManager, error) {
	// Generate encryption key if not provided
	var encryptKey []byte
	if config.EncryptionKey != "" {
		hash := sha256.Sum256([]byte(config.EncryptionKey))
		encryptKey = hash[:]
	} else {
		encryptKey = make([]byte, 32)
		if _, err := rand.Read(encryptKey); err != nil {
			return nil, fmt.Errorf("failed to generate encryption key: %w", err)
		}
	}

	sm := &SessionManager{
		db:         db,
		cookieName: config.CookieName,
		secure:     config.Secure,
		httpOnly:   config.HTTPOnly,
		sameSite:   config.SameSite,
		maxAge:     config.MaxAge,
		domain:     config.Domain,
		path:       config.Path,
		encryptKey: encryptKey,
	}

	// Create sessions table if not exists
	if err := sm.createSessionsTable(); err != nil {
		return nil, fmt.Errorf("failed to create sessions table: %w", err)
	}

	// Start cleanup routine
	sm.startCleanupRoutine(config.CleanupInterval)

	return sm, nil
}

// createSessionsTable creates the sessions table
func (sm *SessionManager) createSessionsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS sessions (
		id VARCHAR(128) PRIMARY KEY,
		user_id BIGINT,
		user_agent TEXT,
		ip_address VARCHAR(45),
		login_time DATETIME,
		last_activity DATETIME,
		expires_at DATETIME,
		is_active BOOLEAN DEFAULT TRUE,
		device_info TEXT,
		permissions TEXT,
		preferences TEXT,
		data TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_user_id (user_id),
		INDEX idx_expires_at (expires_at),
		INDEX idx_is_active (is_active)
	)`

	_, err := sm.db.Exec(query)
	return err
}

// GenerateSessionID generates a cryptographically secure session ID
func (sm *SessionManager) GenerateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	
	// Add timestamp for uniqueness
	timestamp := time.Now().UnixNano()
	sessionData := fmt.Sprintf("%s:%d", base64.URLEncoding.EncodeToString(bytes), timestamp)
	
	// Hash the session data
	hash := sha256.Sum256([]byte(sessionData))
	return base64.URLEncoding.EncodeToString(hash[:]), nil
}

// CreateSession creates a new session
func (sm *SessionManager) CreateSession(w http.ResponseWriter, r *http.Request, userID int64, permissions []string) (*SessionData, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessionID, err := sm.GenerateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(sm.maxAge) * time.Second)

	// Get device info
	deviceInfo := sm.extractDeviceInfo(r)

	session := &SessionData{
		ID:           sessionID,
		UserID:       userID,
		UserAgent:    r.UserAgent(),
		IPAddress:    sm.getRealIP(r),
		LoginTime:    now,
		LastActivity: now,
		ExpiresAt:    expiresAt,
		IsActive:     true,
		DeviceInfo:   deviceInfo,
		Permissions:  permissions,
		Preferences:  make(map[string]interface{}),
		Data:         make(map[string]interface{}),
	}

	// Save to database
	if err := sm.saveSessionToDB(session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Set cookie
	sm.setCookie(w, sessionID, expiresAt)

	return session, nil
}

// GetSession retrieves session data
func (sm *SessionManager) GetSession(r *http.Request) (*SessionData, error) {
	cookie, err := r.Cookie(sm.cookieName)
	if err != nil {
		return nil, fmt.Errorf("session cookie not found: %w", err)
	}

	return sm.getSessionFromDB(cookie.Value)
}

// UpdateSession updates session data
func (sm *SessionManager) UpdateSession(sessionID string, updates map[string]interface{}) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, err := sm.getSessionFromDB(sessionID)
	if err != nil {
		return err
	}

	// Update fields
	for key, value := range updates {
		switch key {
		case "preferences":
			if prefs, ok := value.(map[string]interface{}); ok {
				session.Preferences = prefs
			}
		case "data":
			if data, ok := value.(map[string]interface{}); ok {
				session.Data = data
			}
		case "permissions":
			if perms, ok := value.([]string); ok {
				session.Permissions = perms
			}
		}
	}

	session.LastActivity = time.Now()
	return sm.saveSessionToDB(session)
}

// UpdateSessionData updates session data by SessionData
func (sm *SessionManager) UpdateSessionData(sessionData *SessionData) error {
	return sm.saveSessionToDB(sessionData)
}

// DestroySession destroys a session
func (sm *SessionManager) DestroySession(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(sm.cookieName)
	if err != nil {
		return nil // Cookie doesn't exist, nothing to destroy
	}

	// Remove from database
	query := "UPDATE sessions SET is_active = FALSE WHERE id = ?"
	_, err = sm.db.Exec(query, cookie.Value)
	if err != nil {
		return fmt.Errorf("failed to destroy session: %w", err)
	}

	// Clear cookie
	sm.clearCookie(w)
	return nil
}

// GetUserSessions gets all active sessions for a user
func (sm *SessionManager) GetUserSessions(userID int64) ([]*SessionData, error) {
	query := `
		SELECT id, user_id, user_agent, ip_address, login_time, last_activity, 
		       expires_at, is_active, device_info, permissions, preferences, data
		FROM sessions 
		WHERE user_id = ? AND is_active = TRUE AND expires_at > NOW()
		ORDER BY last_activity DESC
	`

	rows, err := sm.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*SessionData
	for rows.Next() {
		session := &SessionData{}
		var deviceInfoJSON, permissionsJSON, preferencesJSON, dataJSON string

		err := rows.Scan(
			&session.ID, &session.UserID, &session.UserAgent, &session.IPAddress,
			&session.LoginTime, &session.LastActivity, &session.ExpiresAt,
			&session.IsActive, &deviceInfoJSON, &permissionsJSON,
			&preferencesJSON, &dataJSON,
		)
		if err != nil {
			continue
		}

		// Parse JSON fields
		json.Unmarshal([]byte(deviceInfoJSON), &session.DeviceInfo)
		json.Unmarshal([]byte(permissionsJSON), &session.Permissions)
		json.Unmarshal([]byte(preferencesJSON), &session.Preferences)
		json.Unmarshal([]byte(dataJSON), &session.Data)

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// saveSessionToDB saves session to database
func (sm *SessionManager) saveSessionToDB(session *SessionData) error {
	deviceInfoJSON, _ := json.Marshal(session.DeviceInfo)
	permissionsJSON, _ := json.Marshal(session.Permissions)
	preferencesJSON, _ := json.Marshal(session.Preferences)
	dataJSON, _ := json.Marshal(session.Data)

	query := `
		INSERT INTO sessions (id, user_id, user_agent, ip_address, login_time, 
		                     last_activity, expires_at, is_active, device_info, 
		                     permissions, preferences, data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		last_activity = VALUES(last_activity),
		device_info = VALUES(device_info),
		permissions = VALUES(permissions),
		preferences = VALUES(preferences),
		data = VALUES(data)
	`

	_, err := sm.db.Exec(query,
		session.ID, session.UserID, session.UserAgent, session.IPAddress,
		session.LoginTime, session.LastActivity, session.ExpiresAt,
		session.IsActive, deviceInfoJSON, permissionsJSON,
		preferencesJSON, dataJSON,
	)

	return err
}

// getSessionFromDB retrieves session from database
func (sm *SessionManager) getSessionFromDB(sessionID string) (*SessionData, error) {
	query := `
		SELECT id, user_id, user_agent, ip_address, login_time, last_activity,
		       expires_at, is_active, device_info, permissions, preferences, data
		FROM sessions 
		WHERE id = ? AND is_active = TRUE AND expires_at > NOW()
	`

	row := sm.db.QueryRow(query, sessionID)
	
	session := &SessionData{}
	var deviceInfoJSON, permissionsJSON, preferencesJSON, dataJSON string

	err := row.Scan(
		&session.ID, &session.UserID, &session.UserAgent, &session.IPAddress,
		&session.LoginTime, &session.LastActivity, &session.ExpiresAt,
		&session.IsActive, &deviceInfoJSON, &permissionsJSON,
		&preferencesJSON, &dataJSON,
	)
	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	json.Unmarshal([]byte(deviceInfoJSON), &session.DeviceInfo)
	json.Unmarshal([]byte(permissionsJSON), &session.Permissions)
	json.Unmarshal([]byte(preferencesJSON), &session.Preferences)
	json.Unmarshal([]byte(dataJSON), &session.Data)

	return session, nil
}

// setCookie sets the session cookie
func (sm *SessionManager) setCookie(w http.ResponseWriter, sessionID string, expiresAt time.Time) {
	cookie := &http.Cookie{
		Name:     sm.cookieName,
		Value:    sessionID,
		Path:     sm.path,
		Domain:   sm.domain,
		Expires:  expiresAt,
		MaxAge:   sm.maxAge,
		Secure:   sm.secure,
		HttpOnly: sm.httpOnly,
		SameSite: sm.sameSite,
	}
	http.SetCookie(w, cookie)
}

// clearCookie clears the session cookie
func (sm *SessionManager) clearCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     sm.cookieName,
		Value:    "",
		Path:     sm.path,
		Domain:   sm.domain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   sm.secure,
		HttpOnly: sm.httpOnly,
		SameSite: sm.sameSite,
	}
	http.SetCookie(w, cookie)
}

// extractDeviceInfo extracts device information from request
func (sm *SessionManager) extractDeviceInfo(r *http.Request) map[string]interface{} {
	userAgent := r.UserAgent()
	
	deviceInfo := map[string]interface{}{
		"user_agent": userAgent,
		"platform":   sm.detectPlatform(userAgent),
		"browser":    sm.detectBrowser(userAgent),
		"device":     sm.detectDevice(userAgent),
		"language":   r.Header.Get("Accept-Language"),
		"encoding":   r.Header.Get("Accept-Encoding"),
	}

	return deviceInfo
}

// detectPlatform detects platform from user agent
func (sm *SessionManager) detectPlatform(userAgent string) string {
	ua := strings.ToLower(userAgent)
	
	if strings.Contains(ua, "windows") {
		return "Windows"
	} else if strings.Contains(ua, "mac") {
		return "macOS"
	} else if strings.Contains(ua, "linux") {
		return "Linux"
	} else if strings.Contains(ua, "android") {
		return "Android"
	} else if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
		return "iOS"
	}
	
	return "Unknown"
}

// detectBrowser detects browser from user agent
func (sm *SessionManager) detectBrowser(userAgent string) string {
	ua := strings.ToLower(userAgent)
	
	if strings.Contains(ua, "chrome") && !strings.Contains(ua, "edge") {
		return "Chrome"
	} else if strings.Contains(ua, "firefox") {
		return "Firefox"
	} else if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		return "Safari"
	} else if strings.Contains(ua, "edge") {
		return "Edge"
	} else if strings.Contains(ua, "opera") {
		return "Opera"
	}
	
	return "Unknown"
}

// detectDevice detects device type from user agent
func (sm *SessionManager) detectDevice(userAgent string) string {
	ua := strings.ToLower(userAgent)
	
	if strings.Contains(ua, "mobile") {
		return "Mobile"
	} else if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		return "Tablet"
	}
	
	return "Desktop"
}

// getRealIP gets the real IP address from request
func (sm *SessionManager) getRealIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// startCleanupRoutine starts the session cleanup routine
func (sm *SessionManager) startCleanupRoutine(interval time.Duration) {
	if interval <= 0 {
		interval = 1 * time.Hour // Default cleanup interval
	}

	sm.cleanupTimer = time.AfterFunc(interval, func() {
		sm.cleanupExpiredSessions()
		sm.startCleanupRoutine(interval) // Reschedule
	})
}

// cleanupExpiredSessions removes expired sessions
func (sm *SessionManager) cleanupExpiredSessions() {
	query := "DELETE FROM sessions WHERE expires_at < NOW() OR is_active = FALSE"
	_, err := sm.db.Exec(query)
	if err != nil {
		// Log error but don't stop the cleanup routine
		fmt.Printf("Session cleanup error: %v\n", err)
	}
}

// GetSessionStats returns session statistics
func (sm *SessionManager) GetSessionStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total active sessions
	var totalActive int
	err := sm.db.QueryRow("SELECT COUNT(*) FROM sessions WHERE is_active = TRUE AND expires_at > NOW()").Scan(&totalActive)
	if err == nil {
		stats["total_active"] = totalActive
	}

	// Sessions by platform
	platformQuery := `
		SELECT JSON_EXTRACT(device_info, '$.platform') as platform, COUNT(*) as count
		FROM sessions 
		WHERE is_active = TRUE AND expires_at > NOW()
		GROUP BY platform
	`
	rows, err := sm.db.Query(platformQuery)
	if err == nil {
		platforms := make(map[string]int)
		for rows.Next() {
			var platform string
			var count int
			if err := rows.Scan(&platform, &count); err == nil {
				platforms[platform] = count
			}
		}
		rows.Close()
		stats["by_platform"] = platforms
	}

	return stats, nil
}

// Close closes the session manager
func (sm *SessionManager) Close() {
	if sm.cleanupTimer != nil {
		sm.cleanupTimer.Stop()
	}
}