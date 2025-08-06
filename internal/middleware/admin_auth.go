package middleware

import (
	"net/http"
	"log"
	"encoding/json"
	"time"
	
	"github.com/kolajai/internal/models"
)

// AdminAuthMiddleware checks if user is authenticated as admin
func AdminAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session
		session, err := GetSession(r)
		if err != nil {
			log.Printf("AdminAuth: Session error: %v", err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		
		// Check if user is logged in
		userID, ok := session.Values["user_id"].(int64)
		if !ok || userID == 0 {
			log.Printf("AdminAuth: No user_id in session")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		
		// Check if user is admin
		isAdmin, ok := session.Values["is_admin"].(bool)
		if !ok || !isAdmin {
			log.Printf("AdminAuth: User %d is not admin", userID)
			http.Error(w, "Forbidden - Admin access required", http.StatusForbidden)
			return
		}
		
		// Log admin access
		LogAdminAccess(userID, r)
		
		// Continue to next handler
		next(w, r)
	}
}

// RequirePermission checks if admin has specific permission
func RequirePermission(resource, action string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get user from session
			session, err := GetSession(r)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			userID, ok := session.Values["user_id"].(int64)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			// Check permission
			hasPermission := CheckUserPermission(userID, resource, action)
			if !hasPermission {
				log.Printf("Permission denied: User %d tried to %s on %s", userID, action, resource)
				http.Error(w, "Forbidden - Insufficient permissions", http.StatusForbidden)
				return
			}
			
			next(w, r)
		}
	}
}

// LogAdminAction logs admin actions for audit trail
func LogAdminAction(adminID int64, action, resource string, resourceID *int64, oldValue, newValue interface{}, r *http.Request) error {
	adminLog := models.AdminLog{
		AdminID:    adminID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  GetClientIP(r),
		UserAgent:  r.UserAgent(),
		Status:     "success",
		CreatedAt:  time.Now(),
	}
	
	// Convert values to JSON if provided
	if oldValue != nil {
		oldJSON, err := json.Marshal(oldValue)
		if err == nil {
			adminLog.OldValue = oldJSON
		}
	}
	
	if newValue != nil {
		newJSON, err := json.Marshal(newValue)
		if err == nil {
			adminLog.NewValue = newJSON
		}
	}
	
	// TODO: Save to database
	log.Printf("Admin Action: %+v", adminLog)
	
	return nil
}

// LogAdminAccess logs admin panel access
func LogAdminAccess(userID int64, r *http.Request) {
	log.Printf("Admin Access: User %d accessed %s from %s", userID, r.URL.Path, GetClientIP(r))
}

// CheckUserPermission checks if user has specific permission
func CheckUserPermission(userID int64, resource, action string) bool {
	// TODO: Implement actual permission check from database
	// For now, return true for demo
	return true
}

// GetClientIP gets the real client IP address
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}
	
	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// GetSession helper function
func GetSession(r *http.Request) (map[string]interface{}, error) {
	// TODO: Implement proper session retrieval
	// This is a placeholder
	return map[string]interface{}{
		"user_id": int64(1),
		"is_admin": true,
	}, nil
}