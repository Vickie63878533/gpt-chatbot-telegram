package manager

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// AuthMiddleware handles authentication for manager endpoints
type AuthMiddleware struct {
	storage storage.Storage
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(db storage.Storage) *AuthMiddleware {
	return &AuthMiddleware{
		storage: db,
	}
}

// Authenticate validates user ID and token from request headers
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user ID and token from headers
		userIDStr := r.Header.Get("X-User-ID")
		token := r.Header.Get("X-Auth-Token")

		// Validate headers are present
		if userIDStr == "" || token == "" {
			log.Printf("Missing authentication headers: userID=%s, hasToken=%v", userIDStr, token != "")
			http.Error(w, `{"error":"missing authentication headers"}`, http.StatusUnauthorized)
			return
		}

		// Parse user ID
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			log.Printf("Invalid user ID format: %s", userIDStr)
			http.Error(w, `{"error":"invalid user id"}`, http.StatusUnauthorized)
			return
		}

		// Validate token
		valid, err := m.storage.ValidateLoginToken(userID, token)
		if err != nil {
			log.Printf("Error validating token for user %d: %v", userID, err)
			http.Error(w, `{"error":"authentication failed"}`, http.StatusUnauthorized)
			return
		}

		if !valid {
			log.Printf("Invalid or expired token for user %d", userID)
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext extracts the user ID from request context
func GetUserIDFromContext(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value("userID").(int64)
	return userID, ok
}
