package security

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"kolajAi/internal/models"
)

// OAuth2Service handles OAuth2 authentication
type OAuth2Service struct {
	db        *sql.DB
	providers map[string]*OAuth2Provider
}

// OAuth2Provider represents an OAuth2 provider configuration
type OAuth2Provider struct {
	Name         string   `json:"name"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	AuthURL      string   `json:"auth_url"`
	TokenURL     string   `json:"token_url"`
	UserInfoURL  string   `json:"user_info_url"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
}

// OAuth2State represents OAuth2 state information
type OAuth2State struct {
	State     string    `json:"state"`
	Provider  string    `json:"provider"`
	ReturnURL string    `json:"return_url,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
}

// OAuth2UserInfo represents user information from OAuth2 provider
type OAuth2UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture,omitempty"`
	Verified bool   `json:"verified,omitempty"`
}

// OAuth2TokenResponse represents OAuth2 token response
type OAuth2TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// OAuth2Account represents linked OAuth2 account
type OAuth2Account struct {
	ID           int        `json:"id"`
	UserID       int64      `json:"user_id"`
	Provider     string     `json:"provider"`
	ProviderID   string     `json:"provider_id"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	Picture      string     `json:"picture,omitempty"`
	AccessToken  string     `json:"-"`
	RefreshToken string     `json:"-"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// NewOAuth2Service creates a new OAuth2 service
func NewOAuth2Service(db *sql.DB) *OAuth2Service {
	service := &OAuth2Service{
		db:        db,
		providers: make(map[string]*OAuth2Provider),
	}

	// Initialize default providers
	service.initializeDefaultProviders()

	return service
}

// AddProvider adds a new OAuth2 provider
func (o *OAuth2Service) AddProvider(name string, provider *OAuth2Provider) {
	o.providers[name] = provider
}

// GetAuthURL generates OAuth2 authorization URL
func (o *OAuth2Service) GetAuthURL(providerName, returnURL string) (string, error) {
	provider, exists := o.providers[providerName]
	if !exists {
		return "", fmt.Errorf("provider %s not found", providerName)
	}

	// Generate state
	state, err := o.generateState(providerName, returnURL)
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Build authorization URL
	authURL, err := url.Parse(provider.AuthURL)
	if err != nil {
		return "", fmt.Errorf("invalid auth URL: %w", err)
	}

	params := url.Values{}
	params.Add("client_id", provider.ClientID)
	params.Add("redirect_uri", provider.RedirectURL)
	params.Add("response_type", "code")
	params.Add("scope", strings.Join(provider.Scopes, " "))
	params.Add("state", state)

	authURL.RawQuery = params.Encode()

	return authURL.String(), nil
}

// HandleCallback handles OAuth2 callback
func (o *OAuth2Service) HandleCallback(providerName, code, state string) (*OAuth2UserInfo, error) {
	// Validate state
	if !o.validateState(state, providerName) {
		return nil, errors.New("invalid state parameter")
	}

	provider, exists := o.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerName)
	}

	// Exchange code for token
	token, err := o.exchangeCodeForToken(provider, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user info
	userInfo, err := o.getUserInfo(provider, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return userInfo, nil
}

// LinkAccount links OAuth2 account to existing user
func (o *OAuth2Service) LinkAccount(userID int64, providerName string, userInfo *OAuth2UserInfo, token *OAuth2TokenResponse) error {
	// Check if account is already linked
	exists, err := o.isAccountLinked(userID, providerName, userInfo.ID)
	if err != nil {
		return fmt.Errorf("failed to check if account is linked: %w", err)
	}
	if exists {
		return errors.New("account is already linked")
	}

	// Calculate token expiration
	var expiresAt *time.Time
	if token.ExpiresIn > 0 {
		expiry := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
		expiresAt = &expiry
	}

	// Insert OAuth2 account
	_, err = o.db.Exec(`
		INSERT INTO oauth2_accounts (user_id, provider, provider_id, email, name, picture, 
									access_token, refresh_token, expires_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, userID, providerName, userInfo.ID, userInfo.Email, userInfo.Name, userInfo.Picture,
		token.AccessToken, token.RefreshToken, expiresAt)

	if err != nil {
		return fmt.Errorf("failed to link OAuth2 account: %w", err)
	}

	return nil
}

// UnlinkAccount unlinks OAuth2 account from user
func (o *OAuth2Service) UnlinkAccount(userID int64, providerName string) error {
	result, err := o.db.Exec(`
		DELETE FROM oauth2_accounts 
		WHERE user_id = ? AND provider = ?
	`, userID, providerName)

	if err != nil {
		return fmt.Errorf("failed to unlink OAuth2 account: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("OAuth2 account not found")
	}

	return nil
}

// GetLinkedAccounts returns linked OAuth2 accounts for user
func (o *OAuth2Service) GetLinkedAccounts(userID int64) ([]*OAuth2Account, error) {
	rows, err := o.db.Query(`
		SELECT id, user_id, provider, provider_id, email, name, picture, 
			   expires_at, created_at, updated_at
		FROM oauth2_accounts 
		WHERE user_id = ?
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query linked accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*OAuth2Account
	for rows.Next() {
		account := &OAuth2Account{}
		err := rows.Scan(
			&account.ID, &account.UserID, &account.Provider, &account.ProviderID,
			&account.Email, &account.Name, &account.Picture, &account.ExpiresAt,
			&account.CreatedAt, &account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan OAuth2 account: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// FindUserByOAuth2Account finds user by OAuth2 account
func (o *OAuth2Service) FindUserByOAuth2Account(providerName, providerID string) (*models.User, error) {
	var userID int64
	err := o.db.QueryRow(`
		SELECT user_id FROM oauth2_accounts 
		WHERE provider = ? AND provider_id = ?
	`, providerName, providerID).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to find user by OAuth2 account: %w", err)
	}

	// Get user details
	var user models.User
	err = o.db.QueryRow(`
		SELECT id, name, email, phone, role, is_active, is_admin, created_at, updated_at
		FROM users WHERE id = ?
	`, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone, &user.Role,
		&user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user details: %w", err)
	}

	return &user, nil
}

// CreateUserFromOAuth2 creates a new user from OAuth2 account
func (o *OAuth2Service) CreateUserFromOAuth2(providerName string, userInfo *OAuth2UserInfo, token *OAuth2TokenResponse) (*models.User, error) {
	// Start transaction
	tx, err := o.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Create user
	result, err := tx.Exec(`
		INSERT INTO users (name, email, is_active, created_at, updated_at)
		VALUES (?, ?, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, userInfo.Name, userInfo.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID: %w", err)
	}

	// Calculate token expiration
	var expiresAt *time.Time
	if token.ExpiresIn > 0 {
		expiry := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
		expiresAt = &expiry
	}

	// Link OAuth2 account
	_, err = tx.Exec(`
		INSERT INTO oauth2_accounts (user_id, provider, provider_id, email, name, picture,
									access_token, refresh_token, expires_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, userID, providerName, userInfo.ID, userInfo.Email, userInfo.Name, userInfo.Picture,
		token.AccessToken, token.RefreshToken, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to link OAuth2 account: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return created user
	user := &models.User{
		ID:        userID,
		Name:      userInfo.Name,
		Email:     userInfo.Email,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return user, nil
}

// Helper methods

func (o *OAuth2Service) initializeDefaultProviders() {
	// Google OAuth2
	o.providers["google"] = &OAuth2Provider{
		Name:        "google",
		AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:    "https://oauth2.googleapis.com/token",
		UserInfoURL: "https://www.googleapis.com/oauth2/v2/userinfo",
		Scopes:      []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	}

	// GitHub OAuth2
	o.providers["github"] = &OAuth2Provider{
		Name:        "github",
		AuthURL:     "https://github.com/login/oauth/authorize",
		TokenURL:    "https://github.com/login/oauth/access_token",
		UserInfoURL: "https://api.github.com/user",
		Scopes:      []string{"user:email"},
	}

	// Facebook OAuth2
	o.providers["facebook"] = &OAuth2Provider{
		Name:        "facebook",
		AuthURL:     "https://www.facebook.com/v18.0/dialog/oauth",
		TokenURL:    "https://graph.facebook.com/v18.0/oauth/access_token",
		UserInfoURL: "https://graph.facebook.com/v18.0/me?fields=id,name,email,picture",
		Scopes:      []string{"email", "public_profile"},
	}
}

func (o *OAuth2Service) generateState(provider, returnURL string) (string, error) {
	// Generate random state
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	state := hex.EncodeToString(bytes)

	// Store state in database with expiration
	_, err := o.db.Exec(`
		INSERT INTO oauth2_states (state, provider, return_url, expires_at, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, state, provider, returnURL, time.Now().Add(10*time.Minute))

	return state, err
}

func (o *OAuth2Service) validateState(state, provider string) bool {
	var count int
	err := o.db.QueryRow(`
		SELECT COUNT(*) FROM oauth2_states 
		WHERE state = ? AND provider = ? AND expires_at > CURRENT_TIMESTAMP
	`, state, provider).Scan(&count)

	if err != nil || count == 0 {
		return false
	}

	// Clean up used state
	o.db.Exec("DELETE FROM oauth2_states WHERE state = ?", state)

	return true
}

func (o *OAuth2Service) exchangeCodeForToken(provider *OAuth2Provider, code string) (*OAuth2TokenResponse, error) {
	// Prepare token request
	data := url.Values{}
	data.Set("client_id", provider.ClientID)
	data.Set("client_secret", provider.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", provider.RedirectURL)

	// Make token request
	resp, err := http.PostForm(provider.TokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to make token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	var tokenResp OAuth2TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

func (o *OAuth2Service) getUserInfo(provider *OAuth2Provider, accessToken string) (*OAuth2UserInfo, error) {
	// Create request
	req, err := http.NewRequest("GET", provider.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user info request: %w", err)
	}

	// Add authorization header
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Make request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make user info request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user info response: %w", err)
	}

	var userInfo OAuth2UserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info response: %w", err)
	}

	return &userInfo, nil
}

func (o *OAuth2Service) isAccountLinked(userID int64, provider, providerID string) (bool, error) {
	var count int
	err := o.db.QueryRow(`
		SELECT COUNT(*) FROM oauth2_accounts 
		WHERE user_id = ? AND provider = ? AND provider_id = ?
	`, userID, provider, providerID).Scan(&count)

	return count > 0, err
}
