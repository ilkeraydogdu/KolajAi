package middleware

import (
	"context"
	"net/http"
	"kolajAi/internal/session"
	"kolajAi/internal/errors"
)

// AdminMiddleware checks if the user is an admin
func (ms *MiddlewareStack) AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session
		sessionData, err := ms.SessionManager.GetSession(r)
		if err != nil || sessionData == nil {
			// Redirect to login if no session
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}

		// Check if user exists in session
		userValue, exists := sessionData.Values["user"]
		if !exists {
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}

		// Check admin status
		isAdminValue, adminExists := sessionData.Values["is_admin"]
		if !adminExists {
			ms.ErrorManager.HandleHTTPError(w, r, errors.NewApplicationError(
				errors.FORBIDDEN,
				"ACCESS_DENIED",
				"Admin access required",
				nil,
			))
			return
		}

		isAdmin, ok := isAdminValue.(bool)
		if !ok || !isAdmin {
			ms.ErrorManager.HandleHTTPError(w, r, errors.NewApplicationError(
				errors.FORBIDDEN,
				"ACCESS_DENIED",
				"Admin access required",
				nil,
			))
			return
		}

		// Add user to context for handlers to use
		ctx := context.WithValue(r.Context(), "admin_user", userValue)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// RequireAdmin is a helper function for route-specific admin requirement
func RequireAdmin(ms *MiddlewareStack, handler http.HandlerFunc) http.HandlerFunc {
	return ms.AdminMiddleware(handler).ServeHTTP
}