package manager

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// Response types for regex patterns
type RegexPatternResponse struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
	Replace string `json:"replace"`
	Type    string `json:"type"`
	Order   int    `json:"order"`
	Enabled bool   `json:"enabled"`
	UserID  *int64 `json:"user_id,omitempty"`
}

type RegexPatternsListResponse struct {
	Patterns []*RegexPatternResponse `json:"patterns"`
}

// handleRegexRoute routes regex pattern requests to appropriate handlers
func (s *Server) handleRegexRoute(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "user id not found in context")
		return
	}

	// Route based on method and path
	switch r.Method {
	case http.MethodGet:
		if strings.Contains(r.URL.Path, "/api/manager/regex/") {
			// GET /api/manager/regex/:id
			s.handleGetRegexPattern(w, r, userID)
		} else {
			// GET /api/manager/regex
			s.handleListRegexPatterns(w, r, userID)
		}
	case http.MethodPost:
		if strings.Contains(r.URL.Path, "/toggle") {
			// POST /api/manager/regex/:id/toggle
			s.handleToggleRegexPattern(w, r, userID)
		} else {
			// POST /api/manager/regex
			s.handleCreateRegexPattern(w, r, userID)
		}
	case http.MethodPut:
		// PUT /api/manager/regex/:id
		s.handleUpdateRegexPattern(w, r, userID)
	case http.MethodDelete:
		// DELETE /api/manager/regex/:id
		s.handleDeleteRegexPattern(w, r, userID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleListRegexPatterns lists all regex patterns accessible to the user
func (s *Server) handleListRegexPatterns(w http.ResponseWriter, r *http.Request, userID int64) {
	// Get pattern type from query parameter
	patternType := r.URL.Query().Get("type")
	if patternType == "" {
		patternType = "input" // Default to input
	}

	// Check if user can access personal patterns
	canAccessPersonal := s.permission.CanModifyPersonal(userID)

	// List global patterns (always accessible)
	globalPatterns, err := s.storage.ListRegexPatterns(nil, patternType)
	if err != nil {
		log.Printf("Error listing global regex patterns: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list regex patterns")
		return
	}

	var allPatterns []*RegexPatternResponse

	// Add global patterns
	for _, pattern := range globalPatterns {
		allPatterns = append(allPatterns, &RegexPatternResponse{
			ID:      pattern.ID,
			Name:    pattern.Name,
			Pattern: pattern.Pattern,
			Replace: pattern.Replace,
			Type:    pattern.Type,
			Order:   pattern.Order,
			Enabled: pattern.Enabled,
			UserID:  pattern.UserID,
		})
	}

	// Add personal patterns if user can access them
	if canAccessPersonal {
		personalPatterns, err := s.storage.ListRegexPatterns(&userID, patternType)
		if err != nil {
			log.Printf("Error listing personal regex patterns for user %d: %v", userID, err)
			writeError(w, http.StatusInternalServerError, "failed to list regex patterns")
			return
		}

		for _, pattern := range personalPatterns {
			allPatterns = append(allPatterns, &RegexPatternResponse{
				ID:      pattern.ID,
				Name:    pattern.Name,
				Pattern: pattern.Pattern,
				Replace: pattern.Replace,
				Type:    pattern.Type,
				Order:   pattern.Order,
				Enabled: pattern.Enabled,
				UserID:  pattern.UserID,
			})
		}
	}

	writeJSON(w, http.StatusOK, RegexPatternsListResponse{Patterns: allPatterns})
}

// handleGetRegexPattern gets a specific regex pattern
func (s *Server) handleGetRegexPattern(w http.ResponseWriter, r *http.Request, userID int64) {
	patternID, err := parseIDFromPath(r.URL.Path, "/api/manager/regex/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid regex pattern id")
		return
	}

	pattern, err := s.storage.GetRegexPattern(patternID)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "regex pattern not found")
		} else {
			log.Printf("Error getting regex pattern %d: %v", patternID, err)
			writeError(w, http.StatusInternalServerError, "failed to get regex pattern")
		}
		return
	}

	// Check permission
	if !s.permission.CanAccessResource(userID, pattern.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	writeJSON(w, http.StatusOK, RegexPatternResponse{
		ID:      pattern.ID,
		Name:    pattern.Name,
		Pattern: pattern.Pattern,
		Replace: pattern.Replace,
		Type:    pattern.Type,
		Order:   pattern.Order,
		Enabled: pattern.Enabled,
		UserID:  pattern.UserID,
	})
}

// handleCreateRegexPattern creates a new regex pattern
func (s *Server) handleCreateRegexPattern(w http.ResponseWriter, r *http.Request, userID int64) {
	// Check if user can modify personal settings
	if !s.permission.CanModifyPersonal(userID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	// Parse request body
	var createReq struct {
		Name    string `json:"name"`
		Pattern string `json:"pattern"`
		Replace string `json:"replace"`
		Type    string `json:"type"`
		Order   *int   `json:"order,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if createReq.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if createReq.Pattern == "" {
		writeError(w, http.StatusBadRequest, "pattern is required")
		return
	}
	if createReq.Replace == "" {
		writeError(w, http.StatusBadRequest, "replace is required")
		return
	}
	if createReq.Type != "input" && createReq.Type != "output" {
		writeError(w, http.StatusBadRequest, "type must be 'input' or 'output'")
		return
	}

	// Validate regex pattern (prevent ReDoS)
	if _, err := regexp.Compile(createReq.Pattern); err != nil {
		writeError(w, http.StatusBadRequest, "invalid regex pattern: "+err.Error())
		return
	}

	// Set default order if not provided
	order := 100
	if createReq.Order != nil {
		order = *createReq.Order
	}

	// Create regex pattern
	pattern := &storage.RegexPattern{
		UserID:  &userID,
		Name:    createReq.Name,
		Pattern: createReq.Pattern,
		Replace: createReq.Replace,
		Type:    createReq.Type,
		Order:   order,
		Enabled: true,
	}

	if err := s.storage.CreateRegexPattern(pattern); err != nil {
		log.Printf("Error creating regex pattern: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create regex pattern")
		return
	}

	writeJSON(w, http.StatusCreated, RegexPatternResponse{
		ID:      pattern.ID,
		Name:    pattern.Name,
		Pattern: pattern.Pattern,
		Replace: pattern.Replace,
		Type:    pattern.Type,
		Order:   pattern.Order,
		Enabled: pattern.Enabled,
		UserID:  pattern.UserID,
	})
}

// handleUpdateRegexPattern updates an existing regex pattern
func (s *Server) handleUpdateRegexPattern(w http.ResponseWriter, r *http.Request, userID int64) {
	patternID, err := parseIDFromPath(r.URL.Path, "/api/manager/regex/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid regex pattern id")
		return
	}

	// Get existing pattern
	pattern, err := s.storage.GetRegexPattern(patternID)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "regex pattern not found")
		} else {
			log.Printf("Error getting regex pattern %d: %v", patternID, err)
			writeError(w, http.StatusInternalServerError, "failed to get regex pattern")
		}
		return
	}

	// Check permission
	if !s.permission.CanModifyResource(userID, pattern.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	// Parse request body
	var updateReq struct {
		Name    string `json:"name,omitempty"`
		Pattern string `json:"pattern,omitempty"`
		Replace string `json:"replace,omitempty"`
		Type    string `json:"type,omitempty"`
		Order   *int   `json:"order,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Update fields
	if updateReq.Name != "" {
		pattern.Name = updateReq.Name
	}
	if updateReq.Pattern != "" {
		// Validate regex pattern (prevent ReDoS)
		if _, err := regexp.Compile(updateReq.Pattern); err != nil {
			writeError(w, http.StatusBadRequest, "invalid regex pattern: "+err.Error())
			return
		}
		pattern.Pattern = updateReq.Pattern
	}
	if updateReq.Replace != "" {
		pattern.Replace = updateReq.Replace
	}
	if updateReq.Type != "" {
		if updateReq.Type != "input" && updateReq.Type != "output" {
			writeError(w, http.StatusBadRequest, "type must be 'input' or 'output'")
			return
		}
		pattern.Type = updateReq.Type
	}
	if updateReq.Order != nil {
		pattern.Order = *updateReq.Order
	}

	if err := s.storage.UpdateRegexPattern(pattern); err != nil {
		log.Printf("Error updating regex pattern %d: %v", patternID, err)
		writeError(w, http.StatusInternalServerError, "failed to update regex pattern")
		return
	}

	writeJSON(w, http.StatusOK, RegexPatternResponse{
		ID:      pattern.ID,
		Name:    pattern.Name,
		Pattern: pattern.Pattern,
		Replace: pattern.Replace,
		Type:    pattern.Type,
		Order:   pattern.Order,
		Enabled: pattern.Enabled,
		UserID:  pattern.UserID,
	})
}

// handleDeleteRegexPattern deletes a regex pattern
func (s *Server) handleDeleteRegexPattern(w http.ResponseWriter, r *http.Request, userID int64) {
	patternID, err := parseIDFromPath(r.URL.Path, "/api/manager/regex/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid regex pattern id")
		return
	}

	// Get existing pattern
	pattern, err := s.storage.GetRegexPattern(patternID)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "regex pattern not found")
		} else {
			log.Printf("Error getting regex pattern %d: %v", patternID, err)
			writeError(w, http.StatusInternalServerError, "failed to get regex pattern")
		}
		return
	}

	// Check permission
	if !s.permission.CanModifyResource(userID, pattern.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	if err := s.storage.DeleteRegexPattern(patternID); err != nil {
		log.Printf("Error deleting regex pattern %d: %v", patternID, err)
		writeError(w, http.StatusInternalServerError, "failed to delete regex pattern")
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Success: true, Message: "regex pattern deleted"})
}

// handleToggleRegexPattern toggles the enabled status of a regex pattern
func (s *Server) handleToggleRegexPattern(w http.ResponseWriter, r *http.Request, userID int64) {
	// Extract pattern ID from path like /api/manager/regex/:id/toggle
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/manager/regex/"), "/")
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	patternID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid regex pattern id")
		return
	}

	// Get existing pattern
	pattern, err := s.storage.GetRegexPattern(uint(patternID))
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "regex pattern not found")
		} else {
			log.Printf("Error getting regex pattern %d: %v", patternID, err)
			writeError(w, http.StatusInternalServerError, "failed to get regex pattern")
		}
		return
	}

	// Check permission
	if !s.permission.CanModifyResource(userID, pattern.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	// Toggle enabled status
	if err := s.storage.UpdateRegexPatternStatus(uint(patternID), !pattern.Enabled); err != nil {
		log.Printf("Error toggling regex pattern %d: %v", patternID, err)
		writeError(w, http.StatusInternalServerError, "failed to toggle regex pattern status")
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "regex pattern status toggled",
	})
}
