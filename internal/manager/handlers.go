package manager

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/sillytavern"
)

// Response types
type ErrorResponse struct {
	Error string `json:"error"`
}

type CharacterCardResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar,omitempty"`
	IsActive bool   `json:"is_active"`
	UserID   *int64 `json:"user_id,omitempty"`
}

type CharacterCardsListResponse struct {
	Characters []*CharacterCardResponse `json:"characters"`
}

type CharacterCardDetailResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar,omitempty"`
	IsActive bool   `json:"is_active"`
	UserID   *int64 `json:"user_id,omitempty"`
	Data     string `json:"data"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// Helper functions
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

func parseIDFromPath(path string, prefix string) (uint, error) {
	parts := strings.Split(strings.TrimPrefix(path, prefix), "/")
	if len(parts) == 0 {
		return 0, fmt.Errorf("invalid path")
	}

	id, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid id: %w", err)
	}

	return uint(id), nil
}

// Character Card Handlers

// handleCharactersRoute routes character card requests to appropriate handlers
func (s *Server) handleCharactersRoute(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "user id not found in context")
		return
	}

	// Route based on method and path
	switch r.Method {
	case http.MethodGet:
		if strings.Contains(r.URL.Path, "/api/manager/characters/") {
			// GET /api/manager/characters/:id
			s.handleGetCharacter(w, r, userID)
		} else {
			// GET /api/manager/characters
			s.handleListCharacters(w, r, userID)
		}
	case http.MethodPost:
		if strings.Contains(r.URL.Path, "/activate") {
			// POST /api/manager/characters/:id/activate
			s.handleActivateCharacter(w, r, userID)
		} else {
			// POST /api/manager/characters
			s.handleUploadCharacter(w, r, userID)
		}
	case http.MethodPut:
		// PUT /api/manager/characters/:id
		s.handleUpdateCharacter(w, r, userID)
	case http.MethodDelete:
		// DELETE /api/manager/characters/:id
		s.handleDeleteCharacter(w, r, userID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleListCharacters lists all character cards accessible to the user
func (s *Server) handleListCharacters(w http.ResponseWriter, r *http.Request, userID int64) {
	// Check if user can access personal cards
	canAccessPersonal := s.permission.CanModifyPersonal(userID)

	// List global cards (always accessible)
	globalCards, err := s.storage.ListCharacterCards(nil)
	if err != nil {
		log.Printf("Error listing global character cards: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list character cards")
		return
	}

	var allCards []*CharacterCardResponse

	// Add global cards
	for _, card := range globalCards {
		allCards = append(allCards, &CharacterCardResponse{
			ID:       card.ID,
			Name:     card.Name,
			Avatar:   card.Avatar,
			IsActive: card.IsActive,
			UserID:   card.UserID,
		})
	}

	// Add personal cards if user can access them
	if canAccessPersonal {
		personalCards, err := s.storage.ListCharacterCards(&userID)
		if err != nil {
			log.Printf("Error listing personal character cards for user %d: %v", userID, err)
			writeError(w, http.StatusInternalServerError, "failed to list character cards")
			return
		}

		for _, card := range personalCards {
			allCards = append(allCards, &CharacterCardResponse{
				ID:       card.ID,
				Name:     card.Name,
				Avatar:   card.Avatar,
				IsActive: card.IsActive,
				UserID:   card.UserID,
			})
		}
	}

	writeJSON(w, http.StatusOK, CharacterCardsListResponse{Characters: allCards})
}

// handleGetCharacter gets a specific character card
func (s *Server) handleGetCharacter(w http.ResponseWriter, r *http.Request, userID int64) {
	cardID, err := parseIDFromPath(r.URL.Path, "/api/manager/characters/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid character id")
		return
	}

	card, err := s.storage.GetCharacterCard(cardID)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "character card not found")
		} else {
			log.Printf("Error getting character card %d: %v", cardID, err)
			writeError(w, http.StatusInternalServerError, "failed to get character card")
		}
		return
	}

	// Check permission
	if !s.permission.CanAccessResource(userID, card.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	writeJSON(w, http.StatusOK, CharacterCardDetailResponse{
		ID:       card.ID,
		Name:     card.Name,
		Avatar:   card.Avatar,
		IsActive: card.IsActive,
		UserID:   card.UserID,
		Data:     card.Data,
	})
}

// handleUploadCharacter uploads a new character card
func (s *Server) handleUploadCharacter(w http.ResponseWriter, r *http.Request, userID int64) {
	// Check if user can modify personal settings
	if !s.permission.CanModifyPersonal(userID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil { // 10MB limit
		writeError(w, http.StatusBadRequest, "failed to parse form")
		return
	}

	// Get the file
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file not provided")
		return
	}
	defer file.Close()

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read file")
		return
	}

	// Parse character card from file
	var cardData sillytavern.CharacterCardV2
	if err := json.Unmarshal(fileData, &cardData); err != nil {
		writeError(w, http.StatusBadRequest, "invalid character card format")
		return
	}

	// Validate it's a V2 card
	if cardData.Spec != "chara_card_v2" {
		writeError(w, http.StatusBadRequest, "only SillyTavern V2 format is supported")
		return
	}

	// Create character card
	card := &storage.CharacterCard{
		UserID: &userID,
		Name:   cardData.Data.Name,
		Avatar: "", // Could extract from cardData if available
		Data:   string(fileData),
	}

	if err := s.storage.CreateCharacterCard(card); err != nil {
		log.Printf("Error creating character card: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create character card")
		return
	}

	writeJSON(w, http.StatusCreated, CharacterCardResponse{
		ID:       card.ID,
		Name:     card.Name,
		Avatar:   card.Avatar,
		IsActive: card.IsActive,
		UserID:   card.UserID,
	})
}

// handleUpdateCharacter updates an existing character card
func (s *Server) handleUpdateCharacter(w http.ResponseWriter, r *http.Request, userID int64) {
	cardID, err := parseIDFromPath(r.URL.Path, "/api/manager/characters/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid character id")
		return
	}

	// Get existing card
	card, err := s.storage.GetCharacterCard(cardID)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "character card not found")
		} else {
			log.Printf("Error getting character card %d: %v", cardID, err)
			writeError(w, http.StatusInternalServerError, "failed to get character card")
		}
		return
	}

	// Check permission
	if !s.permission.CanModifyResource(userID, card.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	// Parse request body
	var updateReq struct {
		Name   string `json:"name,omitempty"`
		Avatar string `json:"avatar,omitempty"`
		Data   string `json:"data,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Update fields
	if updateReq.Name != "" {
		card.Name = updateReq.Name
	}
	if updateReq.Avatar != "" {
		card.Avatar = updateReq.Avatar
	}
	if updateReq.Data != "" {
		// Validate the data is valid JSON
		var cardData sillytavern.CharacterCardV2
		if err := json.Unmarshal([]byte(updateReq.Data), &cardData); err != nil {
			writeError(w, http.StatusBadRequest, "invalid character card data")
			return
		}
		if cardData.Spec != "chara_card_v2" {
			writeError(w, http.StatusBadRequest, "only SillyTavern V2 format is supported")
			return
		}
		card.Data = updateReq.Data
	}

	if err := s.storage.UpdateCharacterCard(card); err != nil {
		log.Printf("Error updating character card %d: %v", cardID, err)
		writeError(w, http.StatusInternalServerError, "failed to update character card")
		return
	}

	writeJSON(w, http.StatusOK, CharacterCardResponse{
		ID:       card.ID,
		Name:     card.Name,
		Avatar:   card.Avatar,
		IsActive: card.IsActive,
		UserID:   card.UserID,
	})
}

// handleDeleteCharacter deletes a character card
func (s *Server) handleDeleteCharacter(w http.ResponseWriter, r *http.Request, userID int64) {
	cardID, err := parseIDFromPath(r.URL.Path, "/api/manager/characters/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid character id")
		return
	}

	// Get existing card
	card, err := s.storage.GetCharacterCard(cardID)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "character card not found")
		} else {
			log.Printf("Error getting character card %d: %v", cardID, err)
			writeError(w, http.StatusInternalServerError, "failed to get character card")
		}
		return
	}

	// Check permission
	if !s.permission.CanModifyResource(userID, card.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	if err := s.storage.DeleteCharacterCard(cardID); err != nil {
		log.Printf("Error deleting character card %d: %v", cardID, err)
		writeError(w, http.StatusInternalServerError, "failed to delete character card")
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Success: true, Message: "character card deleted"})
}

// handleActivateCharacter activates a character card
func (s *Server) handleActivateCharacter(w http.ResponseWriter, r *http.Request, userID int64) {
	// Extract card ID from path like /api/manager/characters/:id/activate
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/manager/characters/"), "/")
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	cardID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid character id")
		return
	}

	// Get existing card
	card, err := s.storage.GetCharacterCard(uint(cardID))
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "character card not found")
		} else {
			log.Printf("Error getting character card %d: %v", cardID, err)
			writeError(w, http.StatusInternalServerError, "failed to get character card")
		}
		return
	}

	// Check permission
	if !s.permission.CanModifyResource(userID, card.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	if err := s.storage.ActivateCharacterCard(&userID, uint(cardID)); err != nil {
		log.Printf("Error activating character card %d: %v", cardID, err)
		writeError(w, http.StatusInternalServerError, "failed to activate character card")
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Success: true, Message: "character card activated"})
}
