package manager

import (
	"log"
	"net/http"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// Server represents the manager HTTP server (integrated into main server)
type Server struct {
	config     *config.Config
	storage    storage.Storage
	auth       *AuthMiddleware
	permission *PermissionChecker
}

// New creates a new manager server
func New(cfg *config.Config, db storage.Storage) *Server {
	return &Server{
		config:     cfg,
		storage:    db,
		auth:       NewAuthMiddleware(db),
		permission: NewPermissionChecker(cfg),
	}
}

// RegisterRoutes registers all manager routes to the given mux
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	// Health check endpoint (no auth required)
	mux.HandleFunc("/api/manager/health", s.handleHealth)

	// Character card endpoints - use new routing handler
	mux.HandleFunc("/api/manager/characters", s.withAuth(s.handleCharactersRoute))
	mux.HandleFunc("/api/manager/characters/", s.withAuth(s.handleCharactersRoute))

	// World book endpoints - use new routing handler
	mux.HandleFunc("/api/manager/worldbooks", s.withAuth(s.handleWorldBooksRoute))
	mux.HandleFunc("/api/manager/worldbooks/", s.withAuth(s.handleWorldBooksRoute))

	// Preset endpoints - use new routing handler
	mux.HandleFunc("/api/manager/presets", s.withAuth(s.handlePresetsRoute))
	mux.HandleFunc("/api/manager/presets/", s.withAuth(s.handlePresetsRoute))

	// Regex endpoints - use new routing handler
	mux.HandleFunc("/api/manager/regex", s.withAuth(s.handleRegexRoute))
	mux.HandleFunc("/api/manager/regex/", s.withAuth(s.handleRegexRoute))

	log.Println("Manager routes registered")
}

// withAuth wraps a handler with authentication middleware
func (s *Server) withAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.auth.Authenticate(http.HandlerFunc(handler)).ServeHTTP(w, r)
	}
}

// handleHealth returns a health check response
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
