package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/manager"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	config     *config.Config
	storage    storage.Storage
	router     *http.ServeMux
}

// New creates a new HTTP server
func New(cfg *config.Config, db storage.Storage) *Server {
	s := &Server{
		config:  cfg,
		storage: db,
		router:  http.NewServeMux(),
	}

	// Setup routes
	s.setupRoutes()

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Starting HTTP server on port %d", s.config.Port)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down HTTP server...")
	return s.httpServer.Shutdown(ctx)
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	s.router.HandleFunc("/", s.handleRoot)
	s.router.HandleFunc("/init", s.handleInit)
	s.router.HandleFunc("/telegram/", s.handleTelegram)

	// Register manager routes if enabled
	if s.config.ManagerEnabled {
		log.Println("Manager is enabled, registering manager routes...")
		managerServer := manager.New(s.config, s.storage)
		managerServer.RegisterRoutes(s.router)
	} else {
		log.Println("Manager is disabled")
	}
}
