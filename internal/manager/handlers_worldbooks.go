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
)

// Response types for world books
type WorldBookResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	UserID   *int64 `json:"user_id,omitempty"`
}

type WorldBooksListResponse struct {
	WorldBooks []*WorldBookResponse `json:"world_books"`
}

type WorldBookDetailResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	UserID   *int64 `json:"user_id,omitempty"`
	Data     string `json:"data"`
}

type WorldBookEntryResponse struct {
	ID            uint   `json:"id"`
	UID           string `json:"uid"`
	Keys          string `json:"keys"`
	SecondaryKeys string `json:"secondary_keys,omitempty"`
	Content       string `json:"content"`
	Comment       string `json:"comment,omitempty"`
	Constant      bool   `json:"constant"`
	Selective     bool   `json:"selective"`
	Order         int    `json:"order"`
	Position      string `json:"position"`
	Enabled       bool   `json:"enabled"`
	Extensions    string `json:"extensions,omitempty"`
}

type WorldBookEntriesListResponse struct {
	Entries []*WorldBookEntryResponse `json:"entries"`
}

// handleWorldBooksRoute routes world book requests to appropriate handlers
func (s *Server) handleWorldBooksRoute(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "user id not found in context")
		return
	}

	// Route based on method and path
	switch r.Method {
	case http.MethodGet:
		if strings.Contains(r.URL.Path, "/api/manager/worldbooks/") {
			// Check if it's an entries endpoint
			if strings.Contains(r.URL.Path, "/entries") {
				if strings.Contains(r.URL.Path, "/entries/") {
					// GET /api/manager/worldbooks/:id/entries/:eid - not used in this task
					writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				} else {
					// GET /api/manager/worldbooks/:id/entries
					s.handleListWorldBookEntries(w, r, userID)
				}
			} else {
				// GET /api/manager/worldbooks/:id
				s.handleGetWorldBook(w, r, userID)
			}
		} else {
			// GET /api/manager/worldbooks
			s.handleListWorldBooks(w, r, userID)
		}
	case http.MethodPost:
		if strings.Contains(r.URL.Path, "/activate") {
			// POST /api/manager/worldbooks/:id/activate
			s.handleActivateWorldBook(w, r, userID)
		} else if strings.Contains(r.URL.Path, "/entries/") && strings.Contains(r.URL.Path, "/toggle") {
			// POST /api/manager/worldbooks/:id/entries/:eid/toggle
			s.handleToggleWorldBookEntry(w, r, userID)
		} else {
			// POST /api/manager/worldbooks
			s.handleUploadWorldBook(w, r, userID)
		}
	case http.MethodPut:
		if strings.Contains(r.URL.Path, "/entries/") {
			// PUT /api/manager/worldbooks/:id/entries/:eid
			s.handleUpdateWorldBookEntry(w, r, userID)
		} else {
			// PUT /api/manager/worldbooks/:id
			s.handleUpdateWorldBook(w, r, userID)
		}
	case http.MethodDelete:
		// DELETE /api/manager/worldbooks/:id
		s.handleDeleteWorldBook(w, r, userID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleListWorldBooks lists all world books accessible to the user
func (s *Server) handleListWorldBooks(w http.ResponseWriter, r *http.Request, userID int64) {
	// Check if user can access personal world books
	canAccessPersonal := s.permission.CanModifyPersonal(userID)

	// List global world books (always accessible)
	globalBooks, err := s.storage.ListWorldBooks(nil)
	if err != nil {
		log.Printf("Error listing global world books: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list world books")
		return
	}

	var allBooks []*WorldBookResponse

	// Add global books
	for _, book := range globalBooks {
		allBooks = append(allBooks, &WorldBookResponse{
			ID:       book.ID,
			Name:     book.Name,
			IsActive: book.IsActive,
			UserID:   book.UserID,
		})
	}

	// Add personal books if user can access them
	if canAccessPersonal {
		personalBooks, err := s.storage.ListWorldBooks(&userID)
		if err != nil {
			log.Printf("Error listing personal world books for user %d: %v", userID, err)
			writeError(w, http.StatusInternalServerError, "failed to list world books")
			return
		}

		for _, book := range personalBooks {
			allBooks = append(allBooks, &WorldBookResponse{
				ID:       book.ID,
				Name:     book.Name,
				IsActive: book.IsActive,
				UserID:   book.UserID,
			})
		}
	}

	writeJSON(w, http.StatusOK, WorldBooksListResponse{WorldBooks: allBooks})
}

// handleGetWorldBook gets a specific world book
func (s *Server) handleGetWorldBook(w http.ResponseWriter, r *http.Request, userID int64) {
	bookID, err := parseIDFromPath(r.URL.Path, "/api/manager/worldbooks/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid world book id")
		return
	}

	book, err := s.storage.GetWorldBook(bookID)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "world book not found")
		} else {
			log.Printf("Error getting world book %d: %v", bookID, err)
			writeError(w, http.StatusInternalServerError, "failed to get world book")
		}
		return
	}

	// Check permission
	if !s.permission.CanAccessResource(userID, book.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	writeJSON(w, http.StatusOK, WorldBookDetailResponse{
		ID:       book.ID,
		Name:     book.Name,
		IsActive: book.IsActive,
		UserID:   book.UserID,
		Data:     book.Data,
	})
}

// handleUploadWorldBook uploads a new world book
func (s *Server) handleUploadWorldBook(w http.ResponseWriter, r *http.Request, userID int64) {
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

	// Parse world book from file
	var bookData map[string]interface{}
	if err := json.Unmarshal(fileData, &bookData); err != nil {
		writeError(w, http.StatusBadRequest, "invalid world book format")
		return
	}

	// Validate it has required fields
	if _, hasName := bookData["name"]; !hasName {
		writeError(w, http.StatusBadRequest, "world book must have a name field")
		return
	}

	// Extract name
	name, ok := bookData["name"].(string)
	if !ok || name == "" {
		writeError(w, http.StatusBadRequest, "world book name must be a non-empty string")
		return
	}

	// Create world book
	book := &storage.WorldBook{
		UserID: &userID,
		Name:   name,
		Data:   string(fileData),
	}

	if err := s.storage.CreateWorldBook(book); err != nil {
		log.Printf("Error creating world book: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create world book")
		return
	}

	writeJSON(w, http.StatusCreated, WorldBookResponse{
		ID:       book.ID,
		Name:     book.Name,
		IsActive: book.IsActive,
		UserID:   book.UserID,
	})
}

// handleUpdateWorldBook updates an existing world book
func (s *Server) handleUpdateWorldBook(w http.ResponseWriter, r *http.Request, userID int64) {
	bookID, err := parseIDFromPath(r.URL.Path, "/api/manager/worldbooks/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid world book id")
		return
	}

	// Get existing book
	book, err := s.storage.GetWorldBook(bookID)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "world book not found")
		} else {
			log.Printf("Error getting world book %d: %v", bookID, err)
			writeError(w, http.StatusInternalServerError, "failed to get world book")
		}
		return
	}

	// Check permission
	if !s.permission.CanModifyResource(userID, book.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	// Parse request body
	var updateReq struct {
		Name string `json:"name,omitempty"`
		Data string `json:"data,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Update fields
	if updateReq.Name != "" {
		book.Name = updateReq.Name
	}
	if updateReq.Data != "" {
		// Validate the data is valid JSON
		var bookData map[string]interface{}
		if err := json.Unmarshal([]byte(updateReq.Data), &bookData); err != nil {
			writeError(w, http.StatusBadRequest, "invalid world book data")
			return
		}
		book.Data = updateReq.Data
	}

	if err := s.storage.UpdateWorldBook(book); err != nil {
		log.Printf("Error updating world book %d: %v", bookID, err)
		writeError(w, http.StatusInternalServerError, "failed to update world book")
		return
	}

	writeJSON(w, http.StatusOK, WorldBookResponse{
		ID:       book.ID,
		Name:     book.Name,
		IsActive: book.IsActive,
		UserID:   book.UserID,
	})
}

// handleDeleteWorldBook deletes a world book
func (s *Server) handleDeleteWorldBook(w http.ResponseWriter, r *http.Request, userID int64) {
	bookID, err := parseIDFromPath(r.URL.Path, "/api/manager/worldbooks/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid world book id")
		return
	}

	// Get existing book
	book, err := s.storage.GetWorldBook(bookID)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "world book not found")
		} else {
			log.Printf("Error getting world book %d: %v", bookID, err)
			writeError(w, http.StatusInternalServerError, "failed to get world book")
		}
		return
	}

	// Check permission
	if !s.permission.CanModifyResource(userID, book.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	if err := s.storage.DeleteWorldBook(bookID); err != nil {
		log.Printf("Error deleting world book %d: %v", bookID, err)
		writeError(w, http.StatusInternalServerError, "failed to delete world book")
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Success: true, Message: "world book deleted"})
}

// handleActivateWorldBook activates a world book
func (s *Server) handleActivateWorldBook(w http.ResponseWriter, r *http.Request, userID int64) {
	// Extract book ID from path like /api/manager/worldbooks/:id/activate
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/manager/worldbooks/"), "/")
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	bookID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid world book id")
		return
	}

	// Get existing book
	book, err := s.storage.GetWorldBook(uint(bookID))
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "world book not found")
		} else {
			log.Printf("Error getting world book %d: %v", bookID, err)
			writeError(w, http.StatusInternalServerError, "failed to get world book")
		}
		return
	}

	// Check permission
	if !s.permission.CanModifyResource(userID, book.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	if err := s.storage.ActivateWorldBook(&userID, uint(bookID)); err != nil {
		log.Printf("Error activating world book %d: %v", bookID, err)
		writeError(w, http.StatusInternalServerError, "failed to activate world book")
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Success: true, Message: "world book activated"})
}

// handleListWorldBookEntries lists all entries in a world book
func (s *Server) handleListWorldBookEntries(w http.ResponseWriter, r *http.Request, userID int64) {
	// Extract book ID from path like /api/manager/worldbooks/:id/entries
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/manager/worldbooks/"), "/")
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	bookID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid world book id")
		return
	}

	// Get the world book to check permissions
	book, err := s.storage.GetWorldBook(uint(bookID))
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "world book not found")
		} else {
			log.Printf("Error getting world book %d: %v", bookID, err)
			writeError(w, http.StatusInternalServerError, "failed to get world book")
		}
		return
	}

	// Check permission
	if !s.permission.CanAccessResource(userID, book.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	// List entries
	entries, err := s.storage.ListWorldBookEntries(uint(bookID))
	if err != nil {
		log.Printf("Error listing world book entries for book %d: %v", bookID, err)
		writeError(w, http.StatusInternalServerError, "failed to list entries")
		return
	}

	var entryResponses []*WorldBookEntryResponse
	for _, entry := range entries {
		entryResponses = append(entryResponses, &WorldBookEntryResponse{
			ID:            entry.ID,
			UID:           entry.UID,
			Keys:          entry.Keys,
			SecondaryKeys: entry.SecondaryKeys,
			Content:       entry.Content,
			Comment:       entry.Comment,
			Constant:      entry.Constant,
			Selective:     entry.Selective,
			Order:         entry.Order,
			Position:      entry.Position,
			Enabled:       entry.Enabled,
			Extensions:    entry.Extensions,
		})
	}

	writeJSON(w, http.StatusOK, WorldBookEntriesListResponse{Entries: entryResponses})
}

// handleUpdateWorldBookEntry updates a world book entry
func (s *Server) handleUpdateWorldBookEntry(w http.ResponseWriter, r *http.Request, userID int64) {
	// Extract book ID and entry ID from path like /api/manager/worldbooks/:id/entries/:eid
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/manager/worldbooks/"), "/")
	if len(parts) < 4 {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	bookID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid world book id")
		return
	}

	entryID, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid entry id")
		return
	}

	// Get the world book to check permissions
	book, err := s.storage.GetWorldBook(uint(bookID))
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "world book not found")
		} else {
			log.Printf("Error getting world book %d: %v", bookID, err)
			writeError(w, http.StatusInternalServerError, "failed to get world book")
		}
		return
	}

	// Check permission
	if !s.permission.CanModifyResource(userID, book.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	// Get the entry
	entry, err := s.storage.GetWorldBookEntry(uint(entryID))
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "entry not found")
		} else {
			log.Printf("Error getting world book entry %d: %v", entryID, err)
			writeError(w, http.StatusInternalServerError, "failed to get entry")
		}
		return
	}

	// Verify entry belongs to the book
	if entry.WorldBookID != uint(bookID) {
		writeError(w, http.StatusBadRequest, "entry does not belong to this world book")
		return
	}

	// Parse request body
	var updateReq struct {
		Keys          string `json:"keys,omitempty"`
		SecondaryKeys string `json:"secondary_keys,omitempty"`
		Content       string `json:"content,omitempty"`
		Comment       string `json:"comment,omitempty"`
		Constant      *bool  `json:"constant,omitempty"`
		Selective     *bool  `json:"selective,omitempty"`
		Order         *int   `json:"order,omitempty"`
		Position      string `json:"position,omitempty"`
		Extensions    string `json:"extensions,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Update fields
	if updateReq.Keys != "" {
		entry.Keys = updateReq.Keys
	}
	if updateReq.SecondaryKeys != "" {
		entry.SecondaryKeys = updateReq.SecondaryKeys
	}
	if updateReq.Content != "" {
		entry.Content = updateReq.Content
	}
	if updateReq.Comment != "" {
		entry.Comment = updateReq.Comment
	}
	if updateReq.Constant != nil {
		entry.Constant = *updateReq.Constant
	}
	if updateReq.Selective != nil {
		entry.Selective = *updateReq.Selective
	}
	if updateReq.Order != nil {
		entry.Order = *updateReq.Order
	}
	if updateReq.Position != "" {
		entry.Position = updateReq.Position
	}
	if updateReq.Extensions != "" {
		entry.Extensions = updateReq.Extensions
	}

	if err := s.storage.UpdateWorldBookEntry(entry); err != nil {
		log.Printf("Error updating world book entry %d: %v", entryID, err)
		writeError(w, http.StatusInternalServerError, "failed to update entry")
		return
	}

	writeJSON(w, http.StatusOK, WorldBookEntryResponse{
		ID:            entry.ID,
		UID:           entry.UID,
		Keys:          entry.Keys,
		SecondaryKeys: entry.SecondaryKeys,
		Content:       entry.Content,
		Comment:       entry.Comment,
		Constant:      entry.Constant,
		Selective:     entry.Selective,
		Order:         entry.Order,
		Position:      entry.Position,
		Enabled:       entry.Enabled,
		Extensions:    entry.Extensions,
	})
}

// handleToggleWorldBookEntry toggles the enabled status of a world book entry
func (s *Server) handleToggleWorldBookEntry(w http.ResponseWriter, r *http.Request, userID int64) {
	// Extract book ID and entry ID from path like /api/manager/worldbooks/:id/entries/:eid/toggle
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/manager/worldbooks/"), "/")
	if len(parts) < 4 {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	bookID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid world book id")
		return
	}

	entryID, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid entry id")
		return
	}

	// Get the world book to check permissions
	book, err := s.storage.GetWorldBook(uint(bookID))
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "world book not found")
		} else {
			log.Printf("Error getting world book %d: %v", bookID, err)
			writeError(w, http.StatusInternalServerError, "failed to get world book")
		}
		return
	}

	// Check permission
	if !s.permission.CanModifyResource(userID, book.UserID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	// Get the entry
	entry, err := s.storage.GetWorldBookEntry(uint(entryID))
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "entry not found")
		} else {
			log.Printf("Error getting world book entry %d: %v", entryID, err)
			writeError(w, http.StatusInternalServerError, "failed to get entry")
		}
		return
	}

	// Verify entry belongs to the book
	if entry.WorldBookID != uint(bookID) {
		writeError(w, http.StatusBadRequest, "entry does not belong to this world book")
		return
	}

	// Toggle enabled status
	if err := s.storage.UpdateWorldBookEntryStatus(uint(entryID), !entry.Enabled); err != nil {
		log.Printf("Error toggling world book entry %d: %v", entryID, err)
		writeError(w, http.StatusInternalServerError, "failed to toggle entry status")
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("entry status toggled to %v", !entry.Enabled),
	})
}
