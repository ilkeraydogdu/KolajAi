package security

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"kolajAi/internal/models"
)

// JWTService handles JWT token operations
type JWTService struct {
	secretKey       []byte
	issuer          string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// JWTClaims represents JWT claims structure
type JWTClaims struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsAdmin   bool   `json:"is_admin"`
	SessionID string `json:"session_id"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// TokenValidationResult represents token validation result
type TokenValidationResult struct {
	Valid     bool       `json:"valid"`
	Claims    *JWTClaims `json:"claims,omitempty"`
	Error     string     `json:"error,omitempty"`
	ExpiresAt time.Time  `json:"expires_at,omitempty"`
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, issuer string) *JWTService {
	return &JWTService{
		secretKey:       []byte(secretKey),
		issuer:          issuer,
		accessTokenTTL:  15 * time.Minute,   // Short-lived access tokens
		refreshTokenTTL: 7 * 24 * time.Hour, // 7 days refresh tokens
	}
}

// GenerateTokenPair generates access and refresh token pair
func (j *JWTService) GenerateTokenPair(user *models.User) (*TokenPair, error) {
	if user == nil {
		return nil, errors.New("user cannot be nil")
	}

	sessionID, err := j.generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	now := time.Now()
	accessExpiresAt := now.Add(j.accessTokenTTL)
	refreshExpiresAt := now.Add(j.refreshTokenTTL)

	// Generate access token
	accessClaims := &JWTClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		IsAdmin:   user.IsAdmin,
		SessionID: sessionID,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   fmt.Sprintf("%d", user.ID),
			Audience:  []string{"kolajAI"},
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        sessionID,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(j.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := &JWTClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		IsAdmin:   user.IsAdmin,
		SessionID: sessionID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   fmt.Sprintf("%d", user.ID),
			Audience:  []string{"kolajAI"},
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        sessionID,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(j.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int64(j.accessTokenTTL.Seconds()),
		ExpiresAt:    accessExpiresAt,
	}, nil
}

// ValidateToken validates a JWT token and returns claims
func (j *JWTService) ValidateToken(tokenString string) *TokenValidationResult {
	if tokenString == "" {
		return &TokenValidationResult{
			Valid: false,
			Error: "token is empty",
		}
	}

	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return &TokenValidationResult{
			Valid: false,
			Error: fmt.Sprintf("failed to parse token: %v", err),
		}
	}

	// Extract claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return &TokenValidationResult{
			Valid: false,
			Error: "invalid token claims",
		}
	}

	// Validate token
	if !token.Valid {
		return &TokenValidationResult{
			Valid: false,
			Error: "token is invalid",
		}
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return &TokenValidationResult{
			Valid: false,
			Error: "token is expired",
		}
	}

	// Check issuer
	if claims.Issuer != j.issuer {
		return &TokenValidationResult{
			Valid: false,
			Error: "invalid token issuer",
		}
	}

	return &TokenValidationResult{
		Valid:     true,
		Claims:    claims,
		ExpiresAt: claims.ExpiresAt.Time,
	}
}

// RefreshTokenPair generates new token pair using refresh token
func (j *JWTService) RefreshTokenPair(refreshTokenString string) (*TokenPair, error) {
	// Validate refresh token
	result := j.ValidateToken(refreshTokenString)
	if !result.Valid {
		return nil, fmt.Errorf("invalid refresh token: %s", result.Error)
	}

	// Check if it's a refresh token
	if result.Claims.TokenType != "refresh" {
		return nil, errors.New("token is not a refresh token")
	}

	// Create user object from claims
	user := &models.User{
		ID:      result.Claims.UserID,
		Email:   result.Claims.Email,
		Role:    result.Claims.Role,
		IsAdmin: result.Claims.IsAdmin,
	}

	// Generate new token pair
	return j.GenerateTokenPair(user)
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func (j *JWTService) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is empty")
	}

	// Check if header starts with "Bearer "
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", errors.New("authorization header must start with 'Bearer '")
	}

	// Extract token
	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if token == "" {
		return "", errors.New("token is empty")
	}

	return token, nil
}

// RevokeToken revokes a token (adds to blacklist)
func (j *JWTService) RevokeToken(tokenString string) error {
	result := j.ValidateToken(tokenString)
	if !result.Valid {
		return fmt.Errorf("cannot revoke invalid token: %s", result.Error)
	}

	// TODO: Add to token blacklist in database or cache
	// For now, we'll just validate the token structure

	return nil
}

// GetTokenClaims extracts claims without validation (for debugging)
func (j *JWTService) GetTokenClaims(tokenString string) (*JWTClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &JWTClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// IsTokenExpired checks if token is expired without full validation
func (j *JWTService) IsTokenExpired(tokenString string) (bool, error) {
	claims, err := j.GetTokenClaims(tokenString)
	if err != nil {
		return true, err
	}

	if claims.ExpiresAt == nil {
		return false, nil
	}

	return time.Now().After(claims.ExpiresAt.Time), nil
}

// GetTokenTimeToExpiry returns time until token expires
func (j *JWTService) GetTokenTimeToExpiry(tokenString string) (time.Duration, error) {
	claims, err := j.GetTokenClaims(tokenString)
	if err != nil {
		return 0, err
	}

	if claims.ExpiresAt == nil {
		return 0, errors.New("token has no expiration")
	}

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl < 0 {
		return 0, nil // Already expired
	}

	return ttl, nil
}

// ValidateAccessToken validates specifically access tokens
func (j *JWTService) ValidateAccessToken(tokenString string) *TokenValidationResult {
	result := j.ValidateToken(tokenString)
	if !result.Valid {
		return result
	}

	if result.Claims.TokenType != "access" {
		return &TokenValidationResult{
			Valid: false,
			Error: "token is not an access token",
		}
	}

	return result
}

// ValidateRefreshToken validates specifically refresh tokens
func (j *JWTService) ValidateRefreshToken(tokenString string) *TokenValidationResult {
	result := j.ValidateToken(tokenString)
	if !result.Valid {
		return result
	}

	if result.Claims.TokenType != "refresh" {
		return &TokenValidationResult{
			Valid: false,
			Error: "token is not a refresh token",
		}
	}

	return result
}

// Helper methods

func (j *JWTService) generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// SetAccessTokenTTL sets access token time to live
func (j *JWTService) SetAccessTokenTTL(ttl time.Duration) {
	j.accessTokenTTL = ttl
}

// SetRefreshTokenTTL sets refresh token time to live
func (j *JWTService) SetRefreshTokenTTL(ttl time.Duration) {
	j.refreshTokenTTL = ttl
}

// GetAccessTokenTTL returns access token time to live
func (j *JWTService) GetAccessTokenTTL() time.Duration {
	return j.accessTokenTTL
}

// GetRefreshTokenTTL returns refresh token time to live
func (j *JWTService) GetRefreshTokenTTL() time.Duration {
	return j.refreshTokenTTL
}
