package manager

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// Response types for presets
type PresetResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	APIType  string `json:"api_type"`
	IsActive bool   `json:"is_active"`
	UserID   *int64 `json:"user_id,omitempty"`
}

type PresetsListResponse struct {
	Presets []*PresetResponse `json:"presets"`
}

type PresetDetailResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	APIType  string `json:"api_type"`
	IsActive bool   `json:"is_active"`
	UserID   *int64 `json:"user_id,omitempty"`
	Data     string `json:"data"`
}

// handlePresetsRoute routes preset requests to appropriate handlers
func (s *Server) handlePresetsRoute(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "user id not found in context")
		return
	}

	// Route based on method and path
	switch r.Method {
	case http.MethodGet:
		if strings.Contains(r.URL.Path, "/api/manager/presets/") {
			// GET /api/manager/presets/:id
			s.handleGetPreset(w, r, userID)
		} else {
			// GET /api/manager/presets
			s.handleListPresets(w, r, userID)
		}
	case http.MethodPost:
		if strings.Contains(r.URL.Path, "/activate") {
			// POST /api/manager/presets/:id/activate
			s.handleActivatePreset(w, r, userID)
		} else {
			// POST /api/manager/presets
			s.handleCreatePreset(w, r, userID)
		}
	case http.MethodPut:
		// PUT /api/manager/presets/:id
		s.handleUpdatePreset(w, r, userID)
	case http.MethodDelete:
		// DELETE /api/manager/presets/:id
		s.handleDeletePreset(w, r, userID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleListPresets lists all presets accessible to the user
func (s *Server) handleListPresets(w http.ResponseWriter, r *http.Request, userID int64) {
	// Get API type from query parameter
	apiType := r.URL.Query().Get("api_type")
	if apiType == "" {
		apiType = "openai" // Default to openai
	}

	// Check if user can access personal presets
	canAccessPersonal := s.permission.CanModifyPersonal(userID)

	// List global presets (always accessible)
	globalPresets, err := s.storage.ListPresets(nil, apiType)
	if err != nil {
		log.Printf("Error listing global presets: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list presets")
		return
	}

	var allPresets []*PresetResponse

	// Add global presets
	for _, preset := range globalPresets {
		allPresets = append(allPresets, &PresetResponse{
			ID:       preset.ID,
			Name:     preset.Name,
			APIType:  preset.APIType,
			IsActive: preset.IsActive,
			UserID:   preset.UserID,
		})
	}

	// Add personal presets if user can access them
	if canAccessPersonal {
		personalPresets, err := s.storage.ListPresets(&userID, apiType)
		if err != nil {
			log.Printf("Error listing personal presets for user %d: %v", userID, err)
			writeError(w, http.StatusInternalServerError, "failed to list presets")
			return
		}

		for _, preset := range personalPresets {
			allPresets = append(allPresets, &PresetResponse{
				ID:       preset.ID,
				Name:     preset.Name,
				APIType:  preset.APIType,
				IsActive: preset.IsActive,
				UserID:   preset.UserID,
			})
		}
	}

	writeJSON(w, http.StatusOK, PresetsListResponse{Presets: allPresets})
}

// handleGetPreset gets a specific preset
func (s *Server) handleGetPreset(w http.ResponseWriter, r *http.Request, userID int64) {
	presetID, err := parseIDFromPath(r.URL.Path, "/api/manager/presets/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid preset id")
		return
	}

	preset, err := s.storage.GetPreset(presetID)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "preset not found")
		} else {
			log.Printf("Error getting preset %d: %v", presetID, err)
			writeError(w, http.StatusInternalServerError, "failed to get preset")
		}
		return
	}

	// Check permission
	if !s.permission.CanAccessResource(userID, preset.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	writeJSON(w, http.StatusOK, PresetDetailResponse{
		ID:       preset.ID,
		Name:     preset.Name,
		APIType:  preset.APIType,
		IsActive: preset.IsActive,
		UserID:   preset.UserID,
		Data:     preset.Data,
	})
}

// handleCreatePreset creates a new preset
func (s *Server) handleCreatePreset(w http.ResponseWriter, r *http.Request, userID int64) {
	// Check if user can modify personal settings
	if !s.permission.CanModifyPersonal(userID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	// Parse request body
	var createReq struct {
		Name    string `json:"name"`
		APIType string `json:"api_type"`
		Data    string `json:"data"`
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
	if createReq.APIType == "" {
		writeError(w, http.StatusBadRequest, "api_type is required")
		return
	}
	if createReq.Data == "" {
		writeError(w, http.StatusBadRequest, "data is required")
		return
	}

	// Validate data is valid JSON
	var presetData map[string]interface{}
	if err := json.Unmarshal([]byte(createReq.Data), &presetData); err != nil {
		writeError(w, http.StatusBadRequest, "invalid preset data format")
		return
	}

	// Create preset
	preset := &storage.Preset{
		UserID:  &userID,
		Name:    createReq.Name,
		APIType: createReq.APIType,
		Data:    createReq.Data,
	}

	if err := s.storage.CreatePreset(preset); err != nil {
		log.Printf("Error creating preset: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create preset")
		return
	}

	writeJSON(w, http.StatusCreated, PresetResponse{
		ID:       preset.ID,
		Name:     preset.Name,
		APIType:  preset.APIType,
		IsActive: preset.IsActive,
		UserID:   preset.UserID,
	})
}

// handleUpdatePreset updates an existing preset
func (s *Server) handleUpdatePreset(w http.ResponseWriter, r *http.Request, userID int64) {
	presetID, err := parseIDFromPath(r.URL.Path, "/api/manager/presets/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid preset id")
		return
	}

	// Get existing preset
	preset, err := s.storage.GetPreset(presetID)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "preset not found")
		} else {
			log.Printf("Error getting preset %d: %v", presetID, err)
			writeError(w, http.StatusInternalServerError, "failed to get preset")
		}
		return
	}

	// Check permission
	if !s.permission.CanModifyResource(userID, preset.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	// Parse request body
	var updateReq struct {
		Name    string `json:"name,omitempty"`
		APIType string `json:"api_type,omitempty"`
		Data    string `json:"data,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Update fields
	if updateReq.Name != "" {
		preset.Name = updateReq.Name
	}
	if updateReq.APIType != "" {
		preset.APIType = updateReq.APIType
	}
	if updateReq.Data != "" {
		// Validate the data is valid JSON
		var presetData map[string]interface{}
		if err := json.Unmarshal([]byte(updateReq.Data), &presetData); err != nil {
			writeError(w, http.StatusBadRequest, "invalid preset data format")
			return
		}
		preset.Data = updateReq.Data
	}

	if err := s.storage.UpdatePreset(preset); err != nil {
		log.Printf("Error updating preset %d: %v", presetID, err)
		writeError(w, http.StatusInternalServerError, "failed to update preset")
		return
	}

	writeJSON(w, http.StatusOK, PresetResponse{
		ID:       preset.ID,
		Name:     preset.Name,
		APIType:  preset.APIType,
		IsActive: preset.IsActive,
		UserID:   preset.UserID,
	})
}

// handleDeletePreset deletes a preset
func (s *Server) handleDeletePreset(w http.ResponseWriter, r *http.Request, userID int64) {
	presetID, err := parseIDFromPath(r.URL.Path, "/api/manager/presets/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid preset id")
		return
	}

	// Get existing preset
	preset, err := s.storage.GetPreset(presetID)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "preset not found")
		} else {
			log.Printf("Error getting preset %d: %v", presetID, err)
			writeError(w, http.StatusInternalServerError, "failed to get preset")
		}
		return
	}

	// Check permission
	if !s.permission.CanModifyResource(userID, preset.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	if err := s.storage.DeletePreset(presetID); err != nil {
		log.Printf("Error deleting preset %d: %v", presetID, err)
		writeError(w, http.StatusInternalServerError, "failed to delete preset")
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Success: true, Message: "preset deleted"})
}

// handleActivatePreset activates a preset
func (s *Server) handleActivatePreset(w http.ResponseWriter, r *http.Request, userID int64) {
	// Extract preset ID from path like /api/manager/presets/:id/activate
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/manager/presets/"), "/")
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	presetID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid preset id")
		return
	}

	// Get existing preset
	preset, err := s.storage.GetPreset(uint(presetID))
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "preset not found")
		} else {
			log.Printf("Error getting preset %d: %v", presetID, err)
			writeError(w, http.StatusInternalServerError, "failed to get preset")
		}
		return
	}

	// Check permission
	if !s.permission.CanModifyResource(userID, preset.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	if err := s.storage.ActivatePreset(&userID, uint(presetID)); err != nil {
		log.Printf("Error activating preset %d: %v", presetID, err)
		writeError(w, http.StatusInternalServerError, "failed to activate preset")
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Success: true, Message: "preset activated"})
}
