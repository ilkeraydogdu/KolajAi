package security

import (
	"strings"
	"testing"
	"time"

	"kolajAi/internal/models"
)

func TestNewJWTService(t *testing.T) {
	secretKey := "test-secret-key-at-least-32-chars"
	issuer := "test-issuer"

	service := NewJWTService(secretKey, issuer)

	if service == nil {
		t.Fatal("NewJWTService returned nil")
	}

	if service.issuer != issuer {
		t.Errorf("Expected issuer %s, got %s", issuer, service.issuer)
	}

	if string(service.secretKey) != secretKey {
		t.Error("Secret key not set correctly")
	}
}

func TestJWTService_GenerateTokenPair(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-chars", "test-issuer")

	user := &models.User{
		ID:    123,
		Email: "test@example.com",
		Role:  "user",
	}

	tokenPair, err := service.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	if tokenPair.AccessToken == "" {
		t.Error("Access token is empty")
	}

	if tokenPair.RefreshToken == "" {
		t.Error("Refresh token is empty")
	}

	if tokenPair.TokenType != "Bearer" {
		t.Errorf("Expected token type 'Bearer', got %s", tokenPair.TokenType)
	}

	if tokenPair.ExpiresIn <= 0 {
		t.Error("ExpiresIn should be positive")
	}
}

func TestJWTService_ValidateToken(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-chars", "test-issuer")

	user := &models.User{
		ID:    123,
		Email: "test@example.com",
		Role:  "user",
	}

	tokenPair, err := service.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	// Test valid access token
	result := service.ValidateToken(tokenPair.AccessToken)
	if result.Error != "" {
		t.Fatalf("ValidateToken failed: %v", result.Error)
	}

	if !result.Valid {
		t.Error("Token should be valid")
	}

	if result.Claims.UserID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, result.Claims.UserID)
	}

	if result.Claims.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, result.Claims.Email)
	}

	if result.Claims.Role != user.Role {
		t.Errorf("Expected role %s, got %s", user.Role, result.Claims.Role)
	}
}

func TestJWTService_ValidateInvalidToken(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-chars", "test-issuer")

	// Test invalid token
	result := service.ValidateToken("invalid-token")
	if result.Error == "" {
		t.Error("Expected error for invalid token")
	}

	if result.Valid {
		t.Error("Invalid token should not be valid")
	}
}

func TestJWTService_ValidateExpiredToken(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-chars", "test-issuer")
	// Set very short TTL for testing
	service.accessTokenTTL = 1 * time.Millisecond

	user := &models.User{
		ID:    123,
		Email: "test@example.com",
		Role:  "user",
	}

	tokenPair, err := service.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	result := service.ValidateToken(tokenPair.AccessToken)
	if result.Error == "" {
		t.Error("Expected error for expired token")
	}

	if result.Valid {
		t.Error("Expired token should not be valid")
	}
}

func TestJWTService_RefreshTokenPair(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-chars", "test-issuer")

	user := &models.User{
		ID:    123,
		Email: "test@example.com",
		Role:  "user",
	}

	originalPair, err := service.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	// Wait a moment to ensure new timestamps
	time.Sleep(10 * time.Millisecond)

	newPair, err := service.RefreshTokenPair(originalPair.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshTokenPair failed: %v", err)
	}

	if newPair.AccessToken == originalPair.AccessToken {
		t.Error("New access token should be different")
	}

	if newPair.RefreshToken == originalPair.RefreshToken {
		t.Error("New refresh token should be different")
	}

	// Validate new access token
	result := service.ValidateToken(newPair.AccessToken)
	if result.Error != "" {
		t.Fatalf("ValidateToken failed for new token: %v", result.Error)
	}

	if !result.Valid {
		t.Error("New token should be valid")
	}
}

func TestJWTService_ExtractTokenFromHeader(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-chars", "test-issuer")

	tests := []struct {
		name        string
		header      string
		expected    string
		expectError bool
	}{
		{
			name:        "valid bearer token",
			header:      "Bearer abc123",
			expected:    "abc123",
			expectError: false,
		},
		{
			name:        "missing bearer prefix",
			header:      "abc123",
			expected:    "",
			expectError: true,
		},
		{
			name:        "empty header",
			header:      "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "only bearer",
			header:      "Bearer",
			expected:    "",
			expectError: true,
		},
		{
			name:        "bearer with extra spaces",
			header:      "Bearer   abc123   ",
			expected:    "abc123",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.ExtractTokenFromHeader(tt.header)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if token != tt.expected {
					t.Errorf("Expected token %s, got %s", tt.expected, token)
				}
			}
		})
	}
}

func TestJWTService_IsTokenExpired(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-chars", "test-issuer")

	user := &models.User{
		ID:    123,
		Email: "test@example.com",
		Role:  "user",
	}

	tokenPair, err := service.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	// Fresh token should not be expired
	expired, err := service.IsTokenExpired(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("IsTokenExpired failed: %v", err)
	}
	if expired {
		t.Error("Fresh token should not be expired")
	}

	// Test with expired token
	service.accessTokenTTL = 1 * time.Millisecond
	expiredTokenPair, err := service.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	expired, err = service.IsTokenExpired(expiredTokenPair.AccessToken)
	if err != nil {
		t.Fatalf("IsTokenExpired failed: %v", err)
	}
	if !expired {
		t.Error("Expired token should be detected as expired")
	}
}

func TestJWTService_GetTokenTimeToExpiry(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-chars", "test-issuer")

	user := &models.User{
		ID:    123,
		Email: "test@example.com",
		Role:  "user",
	}

	tokenPair, err := service.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	ttl, err := service.GetTokenTimeToExpiry(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("GetTokenTimeToExpiry failed: %v", err)
	}
	
	// TTL should be close to the configured access token TTL
	expectedTTL := service.accessTokenTTL
	tolerance := 5 * time.Second

	if ttl < expectedTTL-tolerance || ttl > expectedTTL+tolerance {
		t.Errorf("Expected TTL around %v, got %v", expectedTTL, ttl)
	}
}

func TestJWTService_ValidateAccessToken(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-chars", "test-issuer")

	user := &models.User{
		ID:    123,
		Email: "test@example.com",
		Role:  "user",
	}

	tokenPair, err := service.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	// Test valid access token
	result := service.ValidateAccessToken(tokenPair.AccessToken)
	if result.Error != "" {
		t.Fatalf("ValidateAccessToken failed: %v", result.Error)
	}
	claims := result.Claims

	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, claims.UserID)
	}

	if claims.TokenType != "access" {
		t.Errorf("Expected token type 'access', got %s", claims.TokenType)
	}

	// Test refresh token (should fail)
	result = service.ValidateAccessToken(tokenPair.RefreshToken)
	if result.Error == "" {
		t.Error("Refresh token should not be valid as access token")
	}
}

func TestJWTService_ValidateRefreshToken(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-chars", "test-issuer")

	user := &models.User{
		ID:    123,
		Email: "test@example.com",
		Role:  "user",
	}

	tokenPair, err := service.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	// Test valid refresh token
	result := service.ValidateRefreshToken(tokenPair.RefreshToken)
	if result.Error != "" {
		t.Fatalf("ValidateRefreshToken failed: %v", result.Error)
	}
	claims := result.Claims

	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, claims.UserID)
	}

	if claims.TokenType != "refresh" {
		t.Errorf("Expected token type 'refresh', got %s", claims.TokenType)
	}

	// Test access token (should fail)
	result = service.ValidateRefreshToken(tokenPair.AccessToken)
	if result.Error == "" {
		t.Error("Access token should not be valid as refresh token")
	}
}

func TestJWTService_GenerateSessionID(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-chars", "test-issuer")

	sessionID1, err := service.generateSessionID()
	if err != nil {
		t.Fatalf("generateSessionID failed: %v", err)
	}
	sessionID2, err := service.generateSessionID()
	if err != nil {
		t.Fatalf("generateSessionID failed: %v", err)
	}

	if sessionID1 == "" {
		t.Error("Session ID should not be empty")
	}

	if sessionID2 == "" {
		t.Error("Session ID should not be empty")
	}

	if sessionID1 == sessionID2 {
		t.Error("Session IDs should be unique")
	}

	// Session ID should be hex encoded (64 characters for 32 bytes)
	if len(sessionID1) != 64 {
		t.Errorf("Expected session ID length 64, got %d", len(sessionID1))
	}

	// Should only contain hex characters
	for _, char := range sessionID1 {
		if !strings.Contains("0123456789abcdef", string(char)) {
			t.Errorf("Session ID contains non-hex character: %c", char)
		}
	}
}

func TestJWTService_GetTokenClaims(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-chars", "test-issuer")

	user := &models.User{
		ID:    123,
		Email: "test@example.com",
		Role:  "admin",
	}

	tokenPair, err := service.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	claims, err := service.GetTokenClaims(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("GetTokenClaims failed: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, claims.UserID)
	}

	if claims.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, claims.Email)
	}

	if claims.Role != user.Role {
		t.Errorf("Expected role %s, got %s", user.Role, claims.Role)
	}

	if claims.Issuer != service.issuer {
		t.Errorf("Expected issuer %s, got %s", service.issuer, claims.Issuer)
	}
}

func BenchmarkJWTService_GenerateTokenPair(b *testing.B) {
	service := NewJWTService("test-secret-key-at-least-32-chars", "test-issuer")

	user := &models.User{
		ID:    123,
		Email: "test@example.com",
		Role:  "user",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GenerateTokenPair(user)
		if err != nil {
			b.Fatalf("GenerateTokenPair failed: %v", err)
		}
	}
}

func BenchmarkJWTService_ValidateToken(b *testing.B) {
	service := NewJWTService("test-secret-key-at-least-32-chars", "test-issuer")

	user := &models.User{
		ID:    123,
		Email: "test@example.com",
		Role:  "user",
	}

	tokenPair, err := service.GenerateTokenPair(user)
	if err != nil {
		b.Fatalf("GenerateTokenPair failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := service.ValidateToken(tokenPair.AccessToken)
		if result.Error != "" {
			b.Fatalf("ValidateToken failed: %v", result.Error)
		}
	}
}