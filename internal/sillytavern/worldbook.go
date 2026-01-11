package sillytavern

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// WorldBookManager manages SillyTavern world books
type WorldBookManager struct {
	storage storage.Storage
}

// NewWorldBookManager creates a new WorldBookManager
func NewWorldBookManager(storage storage.Storage) *WorldBookManager {
	return &WorldBookManager{
		storage: storage,
	}
}

// WorldBookData represents the SillyTavern world book format
type WorldBookData struct {
	Name       string                       `json:"name"`
	Entries    map[string]WorldBookEntryData `json:"entries"`
	Extensions map[string]interface{}       `json:"extensions,omitempty"`
}

// WorldBookEntryData represents a single entry in the world book
type WorldBookEntryData struct {
	UID               string                 `json:"uid"`
	Key               []string               `json:"key"`
	KeySecondary      []string               `json:"keysecondary,omitempty"`
	Comment           string                 `json:"comment,omitempty"`
	Content           string                 `json:"content"`
	Constant          bool                   `json:"constant"`
	Selective         bool                   `json:"selective"`
	Order             int                    `json:"order"`
	Position          int                    `json:"position"` // 0 = before_char, 1 = after_char
	Disable           bool                   `json:"disable"`
	ExcludeRecursion  bool                   `json:"excludeRecursion,omitempty"`
	Probability       int                    `json:"probability,omitempty"`
	UseProbability    bool                   `json:"useProbability,omitempty"`
	Depth             int                    `json:"depth,omitempty"`
	SelectiveLogic    int                    `json:"selectiveLogic,omitempty"`
	Extensions        map[string]interface{} `json:"extensions,omitempty"`
}

// LoadBook loads a world book by ID
func (m *WorldBookManager) LoadBook(userID *int64, bookID uint) (*storage.WorldBook, error) {
	book, err := m.storage.GetWorldBook(bookID)
	if err != nil {
		return nil, fmt.Errorf("failed to load world book: %w", err)
	}

	// Check if user has access to this book
	if book.UserID != nil && userID != nil && *book.UserID != *userID {
		return nil, errors.New("access denied: world book belongs to another user")
	}

	return book, nil
}

// SaveBook saves a world book
func (m *WorldBookManager) SaveBook(book *storage.WorldBook) error {
	// Validate the book data is valid JSON
	var bookData WorldBookData
	if err := json.Unmarshal([]byte(book.Data), &bookData); err != nil {
		return fmt.Errorf("invalid world book data: %w", err)
	}

	// Update the name field from the data
	book.Name = bookData.Name

	// Create or update the book
	if book.ID == 0 {
		if err := m.storage.CreateWorldBook(book); err != nil {
			return fmt.Errorf("failed to create world book: %w", err)
		}

		// Create entries in the database
		for _, entryData := range bookData.Entries {
			entry := m.convertToStorageEntry(book.ID, entryData)
			if err := m.storage.CreateWorldBookEntry(entry); err != nil {
				return fmt.Errorf("failed to create world book entry: %w", err)
			}
		}
	} else {
		if err := m.storage.UpdateWorldBook(book); err != nil {
			return fmt.Errorf("failed to update world book: %w", err)
		}

		// Update entries - for simplicity, we'll delete and recreate
		// In production, you might want a more sophisticated sync
		existingEntries, err := m.storage.ListWorldBookEntries(book.ID)
		if err != nil {
			return fmt.Errorf("failed to list existing entries: %w", err)
		}

		// Delete existing entries
		for _, entry := range existingEntries {
			if err := m.storage.DeleteWorldBookEntry(entry.ID); err != nil {
				return fmt.Errorf("failed to delete old entry: %w", err)
			}
		}

		// Create new entries
		for _, entryData := range bookData.Entries {
			entry := m.convertToStorageEntry(book.ID, entryData)
			if err := m.storage.CreateWorldBookEntry(entry); err != nil {
				return fmt.Errorf("failed to create world book entry: %w", err)
			}
		}
	}

	return nil
}

// ListBooks lists all world books accessible to the user
func (m *WorldBookManager) ListBooks(userID *int64) ([]*storage.WorldBook, error) {
	books, err := m.storage.ListWorldBooks(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list world books: %w", err)
	}

	return books, nil
}

// ActivateBook activates a world book for the user
func (m *WorldBookManager) ActivateBook(userID *int64, bookID uint) error {
	// First verify the book exists and user has access
	book, err := m.LoadBook(userID, bookID)
	if err != nil {
		return err
	}

	// Check if this is a global book or user's own book
	if book.UserID != nil && userID != nil && *book.UserID != *userID {
		return errors.New("cannot activate world book belonging to another user")
	}

	if err := m.storage.ActivateWorldBook(userID, bookID); err != nil {
		return fmt.Errorf("failed to activate world book: %w", err)
	}

	return nil
}

// GetActiveBook gets the currently active world book for the user
func (m *WorldBookManager) GetActiveBook(userID *int64) (*storage.WorldBook, error) {
	book, err := m.storage.GetActiveWorldBook(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active world book: %w", err)
	}

	return book, nil
}

// convertToStorageEntry converts WorldBookEntryData to storage.WorldBookEntry
func (m *WorldBookManager) convertToStorageEntry(bookID uint, data WorldBookEntryData) *storage.WorldBookEntry {
	// Convert keys to JSON
	keysJSON, _ := json.Marshal(data.Key)
	secondaryKeysJSON, _ := json.Marshal(data.KeySecondary)
	extensionsJSON, _ := json.Marshal(data.Extensions)

	// Convert position int to string
	position := "after_char"
	if data.Position == 0 {
		position = "before_char"
	}

	return &storage.WorldBookEntry{
		WorldBookID:   bookID,
		UID:           data.UID,
		Keys:          string(keysJSON),
		SecondaryKeys: string(secondaryKeysJSON),
		Content:       data.Content,
		Comment:       data.Comment,
		Constant:      data.Constant,
		Selective:     data.Selective,
		Order:         data.Order,
		Position:      position,
		Enabled:       !data.Disable,
		Extensions:    string(extensionsJSON),
	}
}

// TriggerEntries finds world book entries that should be triggered based on message content
// This implements the core world book logic: keyword matching, priority sorting, etc.
func (m *WorldBookManager) TriggerEntries(bookID uint, messages []storage.HistoryItem) ([]*storage.WorldBookEntry, error) {
	// Get all entries for this book
	entries, err := m.storage.ListWorldBookEntries(bookID)
	if err != nil {
		return nil, fmt.Errorf("failed to list world book entries: %w", err)
	}

	// Combine all message content into a single string for matching
	var messageText strings.Builder
	for _, msg := range messages {
		if contentStr, ok := msg.Content.(string); ok {
			messageText.WriteString(contentStr)
			messageText.WriteString(" ")
		}
	}
	combinedText := strings.ToLower(messageText.String())

	var triggered []*storage.WorldBookEntry

	for _, entry := range entries {
		// Skip disabled entries
		if !entry.Enabled {
			continue
		}

		// Constant entries are always included
		if entry.Constant {
			triggered = append(triggered, entry)
			continue
		}

		// Parse keys from JSON
		var keys []string
		if err := json.Unmarshal([]byte(entry.Keys), &keys); err != nil {
			continue // Skip entries with invalid keys
		}

		// Check if any key matches
		matched := false
		for _, key := range keys {
			key = strings.ToLower(strings.TrimSpace(key))
			if key == "" {
				continue
			}

			// Try as regex first (if it looks like a regex pattern)
			if strings.HasPrefix(key, "/") && strings.HasSuffix(key, "/") {
				// Extract regex pattern (remove slashes)
				pattern := key[1 : len(key)-1]
				if re, err := regexp.Compile(pattern); err == nil {
					if re.MatchString(combinedText) {
						matched = true
						break
					}
				}
			} else {
				// Simple substring match
				if strings.Contains(combinedText, key) {
					matched = true
					break
				}
			}
		}

		// If selective mode, also check secondary keys
		if entry.Selective && !matched {
			var secondaryKeys []string
			if entry.SecondaryKeys != "" {
				if err := json.Unmarshal([]byte(entry.SecondaryKeys), &secondaryKeys); err == nil {
					for _, key := range secondaryKeys {
						key = strings.ToLower(strings.TrimSpace(key))
						if key == "" {
							continue
						}

						// Try as regex first
						if strings.HasPrefix(key, "/") && strings.HasSuffix(key, "/") {
							pattern := key[1 : len(key)-1]
							if re, err := regexp.Compile(pattern); err == nil {
								if re.MatchString(combinedText) {
									matched = true
									break
								}
							}
						} else {
							// Simple substring match
							if strings.Contains(combinedText, key) {
								matched = true
								break
							}
						}
					}
				}
			}
		}

		if matched {
			triggered = append(triggered, entry)
		}
	}

	// Sort by order (priority)
	// Lower order values have higher priority
	for i := 0; i < len(triggered); i++ {
		for j := i + 1; j < len(triggered); j++ {
			if triggered[i].Order > triggered[j].Order {
				triggered[i], triggered[j] = triggered[j], triggered[i]
			}
		}
	}

	return triggered, nil
}

// UpdateEntryStatus updates the enabled status of a world book entry
func (m *WorldBookManager) UpdateEntryStatus(entryID uint, enabled bool) error {
	if err := m.storage.UpdateWorldBookEntryStatus(entryID, enabled); err != nil {
		return fmt.Errorf("failed to update entry status: %w", err)
	}
	return nil
}

// UpdateEntry updates a world book entry
func (m *WorldBookManager) UpdateEntry(entry *storage.WorldBookEntry) error {
	if err := m.storage.UpdateWorldBookEntry(entry); err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}
	return nil
}

// ParseBookData parses the JSON data from a world book
func (m *WorldBookManager) ParseBookData(data string) (*WorldBookData, error) {
	var bookData WorldBookData
	if err := json.Unmarshal([]byte(data), &bookData); err != nil {
		return nil, fmt.Errorf("failed to parse world book data: %w", err)
	}

	return &bookData, nil
}
