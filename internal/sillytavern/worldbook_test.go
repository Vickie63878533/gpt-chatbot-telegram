package sillytavern

import (
	"encoding/json"
	"testing"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// MockStorageWithWorldBook extends MockStorage with world book support
type MockStorageWithWorldBook struct {
	*MockStorage
	books       map[uint]*storage.WorldBook
	entries     map[uint]*storage.WorldBookEntry
	nextBookID  uint
	nextEntryID uint
	activeBooks map[int64]uint // userID -> bookID
}

func NewMockStorageWithWorldBook() *MockStorageWithWorldBook {
	return &MockStorageWithWorldBook{
		MockStorage: NewMockStorage(),
		books:       make(map[uint]*storage.WorldBook),
		entries:     make(map[uint]*storage.WorldBookEntry),
		nextBookID:  1,
		nextEntryID: 1,
		activeBooks: make(map[int64]uint),
	}
}

func (m *MockStorageWithWorldBook) CreateWorldBook(book *storage.WorldBook) error {
	book.ID = m.nextBookID
	m.nextBookID++
	m.books[book.ID] = book
	return nil
}

func (m *MockStorageWithWorldBook) GetWorldBook(id uint) (*storage.WorldBook, error) {
	book, ok := m.books[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return book, nil
}

func (m *MockStorageWithWorldBook) ListWorldBooks(userID *int64) ([]*storage.WorldBook, error) {
	var result []*storage.WorldBook
	for _, book := range m.books {
		// Include global books and user's own books
		if book.UserID == nil || (userID != nil && book.UserID != nil && *book.UserID == *userID) {
			result = append(result, book)
		}
	}
	return result, nil
}

func (m *MockStorageWithWorldBook) UpdateWorldBook(book *storage.WorldBook) error {
	if _, ok := m.books[book.ID]; !ok {
		return storage.ErrNotFound
	}
	m.books[book.ID] = book
	return nil
}

func (m *MockStorageWithWorldBook) DeleteWorldBook(id uint) error {
	delete(m.books, id)
	return nil
}

func (m *MockStorageWithWorldBook) ActivateWorldBook(userID *int64, bookID uint) error {
	// Deactivate all books for this user
	for _, book := range m.books {
		if book.UserID == nil || (userID != nil && book.UserID != nil && *book.UserID == *userID) {
			book.IsActive = false
		}
	}

	// Activate the specified book
	book, ok := m.books[bookID]
	if !ok {
		return storage.ErrNotFound
	}
	book.IsActive = true

	if userID != nil {
		m.activeBooks[*userID] = bookID
	} else {
		m.activeBooks[0] = bookID // Use 0 for global
	}

	return nil
}

func (m *MockStorageWithWorldBook) GetActiveWorldBook(userID *int64) (*storage.WorldBook, error) {
	var activeID uint
	var ok bool

	if userID != nil {
		activeID, ok = m.activeBooks[*userID]
	} else {
		activeID, ok = m.activeBooks[0]
	}

	if !ok {
		return nil, storage.ErrNotFound
	}

	return m.GetWorldBook(activeID)
}

func (m *MockStorageWithWorldBook) CreateWorldBookEntry(entry *storage.WorldBookEntry) error {
	entry.ID = m.nextEntryID
	m.nextEntryID++
	m.entries[entry.ID] = entry
	return nil
}

func (m *MockStorageWithWorldBook) GetWorldBookEntry(id uint) (*storage.WorldBookEntry, error) {
	entry, ok := m.entries[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return entry, nil
}

func (m *MockStorageWithWorldBook) ListWorldBookEntries(bookID uint) ([]*storage.WorldBookEntry, error) {
	var result []*storage.WorldBookEntry
	for _, entry := range m.entries {
		if entry.WorldBookID == bookID {
			result = append(result, entry)
		}
	}
	return result, nil
}

func (m *MockStorageWithWorldBook) UpdateWorldBookEntry(entry *storage.WorldBookEntry) error {
	if _, ok := m.entries[entry.ID]; !ok {
		return storage.ErrNotFound
	}
	m.entries[entry.ID] = entry
	return nil
}

func (m *MockStorageWithWorldBook) DeleteWorldBookEntry(id uint) error {
	delete(m.entries, id)
	return nil
}

func (m *MockStorageWithWorldBook) UpdateWorldBookEntryStatus(id uint, enabled bool) error {
	entry, ok := m.entries[id]
	if !ok {
		return storage.ErrNotFound
	}
	entry.Enabled = enabled
	return nil
}

func TestWorldBookManager_SaveAndLoad(t *testing.T) {
	mockStorage := NewMockStorageWithWorldBook()
	manager := NewWorldBookManager(mockStorage)

	// Create a valid world book
	bookData := WorldBookData{
		Name: "Test World",
		Entries: map[string]WorldBookEntryData{
			"entry1": {
				UID:      "entry1",
				Key:      []string{"dragon", "fire"},
				Content:  "Dragons breathe fire.",
				Constant: false,
				Order:    100,
				Position: 1,
				Disable:  false,
			},
		},
	}

	bookJSON, err := json.Marshal(bookData)
	if err != nil {
		t.Fatalf("Failed to marshal book data: %v", err)
	}

	userID := int64(123)
	book := &storage.WorldBook{
		UserID: &userID,
		Data:   string(bookJSON),
	}

	// Save the book
	err = manager.SaveBook(book)
	if err != nil {
		t.Fatalf("Failed to save book: %v", err)
	}

	// Verify the book was saved with correct name
	if book.Name != "Test World" {
		t.Errorf("Expected name 'Test World', got '%s'", book.Name)
	}

	// Verify entries were created
	entries, err := mockStorage.ListWorldBookEntries(book.ID)
	if err != nil {
		t.Fatalf("Failed to list entries: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	// Load the book
	loadedBook, err := manager.LoadBook(&userID, book.ID)
	if err != nil {
		t.Fatalf("Failed to load book: %v", err)
	}

	if loadedBook.Name != book.Name {
		t.Errorf("Expected name '%s', got '%s'", book.Name, loadedBook.Name)
	}
}

func TestWorldBookManager_ActivateBook(t *testing.T) {
	mockStorage := NewMockStorageWithWorldBook()
	manager := NewWorldBookManager(mockStorage)

	// Create and save a book
	bookData := WorldBookData{
		Name:    "Test World",
		Entries: map[string]WorldBookEntryData{},
	}

	bookJSON, _ := json.Marshal(bookData)
	userID := int64(123)
	book := &storage.WorldBook{
		UserID: &userID,
		Data:   string(bookJSON),
	}

	manager.SaveBook(book)

	// Activate the book
	err := manager.ActivateBook(&userID, book.ID)
	if err != nil {
		t.Fatalf("Failed to activate book: %v", err)
	}

	// Get active book
	activeBook, err := manager.GetActiveBook(&userID)
	if err != nil {
		t.Fatalf("Failed to get active book: %v", err)
	}

	if activeBook.ID != book.ID {
		t.Errorf("Expected active book ID %d, got %d", book.ID, activeBook.ID)
	}
}

func TestWorldBookManager_ListBooks(t *testing.T) {
	mockStorage := NewMockStorageWithWorldBook()
	manager := NewWorldBookManager(mockStorage)

	userID := int64(123)

	// Create multiple books
	for i := 0; i < 3; i++ {
		bookData := WorldBookData{
			Name:    "Test World",
			Entries: map[string]WorldBookEntryData{},
		}

		bookJSON, _ := json.Marshal(bookData)
		book := &storage.WorldBook{
			UserID: &userID,
			Data:   string(bookJSON),
		}

		manager.SaveBook(book)
	}

	// List books
	books, err := manager.ListBooks(&userID)
	if err != nil {
		t.Fatalf("Failed to list books: %v", err)
	}

	if len(books) != 3 {
		t.Errorf("Expected 3 books, got %d", len(books))
	}
}

func TestWorldBookManager_TriggerEntries_SimpleKeyword(t *testing.T) {
	mockStorage := NewMockStorageWithWorldBook()
	manager := NewWorldBookManager(mockStorage)

	// Create a world book with entries
	bookData := WorldBookData{
		Name: "Test World",
		Entries: map[string]WorldBookEntryData{
			"entry1": {
				UID:      "entry1",
				Key:      []string{"dragon"},
				Content:  "Dragons are powerful creatures.",
				Constant: false,
				Order:    100,
				Position: 1,
				Disable:  false,
			},
			"entry2": {
				UID:      "entry2",
				Key:      []string{"sword"},
				Content:  "Swords are weapons.",
				Constant: false,
				Order:    200,
				Position: 1,
				Disable:  false,
			},
		},
	}

	bookJSON, _ := json.Marshal(bookData)
	userID := int64(123)
	book := &storage.WorldBook{
		UserID: &userID,
		Data:   string(bookJSON),
	}

	manager.SaveBook(book)

	// Create messages that trigger "dragon"
	messages := []storage.HistoryItem{
		{Role: "user", Content: "Tell me about dragons"},
		{Role: "assistant", Content: "Dragons are mythical creatures."},
	}

	// Trigger entries
	triggered, err := manager.TriggerEntries(book.ID, messages)
	if err != nil {
		t.Fatalf("Failed to trigger entries: %v", err)
	}

	// Should trigger entry1 but not entry2
	if len(triggered) != 1 {
		t.Errorf("Expected 1 triggered entry, got %d", len(triggered))
	}

	if len(triggered) > 0 && triggered[0].UID != "entry1" {
		t.Errorf("Expected entry1 to be triggered, got %s", triggered[0].UID)
	}
}

func TestWorldBookManager_TriggerEntries_Constant(t *testing.T) {
	mockStorage := NewMockStorageWithWorldBook()
	manager := NewWorldBookManager(mockStorage)

	// Create a world book with a constant entry
	bookData := WorldBookData{
		Name: "Test World",
		Entries: map[string]WorldBookEntryData{
			"entry1": {
				UID:      "entry1",
				Key:      []string{"dragon"},
				Content:  "Dragons are powerful creatures.",
				Constant: true, // Always triggered
				Order:    100,
				Position: 1,
				Disable:  false,
			},
		},
	}

	bookJSON, _ := json.Marshal(bookData)
	userID := int64(123)
	book := &storage.WorldBook{
		UserID: &userID,
		Data:   string(bookJSON),
	}

	manager.SaveBook(book)

	// Create messages that don't mention "dragon"
	messages := []storage.HistoryItem{
		{Role: "user", Content: "Tell me about cats"},
	}

	// Trigger entries
	triggered, err := manager.TriggerEntries(book.ID, messages)
	if err != nil {
		t.Fatalf("Failed to trigger entries: %v", err)
	}

	// Should still trigger entry1 because it's constant
	if len(triggered) != 1 {
		t.Errorf("Expected 1 triggered entry (constant), got %d", len(triggered))
	}
}

func TestWorldBookManager_TriggerEntries_Disabled(t *testing.T) {
	mockStorage := NewMockStorageWithWorldBook()
	manager := NewWorldBookManager(mockStorage)

	// Create a world book with a disabled entry
	bookData := WorldBookData{
		Name: "Test World",
		Entries: map[string]WorldBookEntryData{
			"entry1": {
				UID:      "entry1",
				Key:      []string{"dragon"},
				Content:  "Dragons are powerful creatures.",
				Constant: false,
				Order:    100,
				Position: 1,
				Disable:  true, // Disabled
			},
		},
	}

	bookJSON, _ := json.Marshal(bookData)
	userID := int64(123)
	book := &storage.WorldBook{
		UserID: &userID,
		Data:   string(bookJSON),
	}

	manager.SaveBook(book)

	// Create messages that mention "dragon"
	messages := []storage.HistoryItem{
		{Role: "user", Content: "Tell me about dragons"},
	}

	// Trigger entries
	triggered, err := manager.TriggerEntries(book.ID, messages)
	if err != nil {
		t.Fatalf("Failed to trigger entries: %v", err)
	}

	// Should not trigger because entry is disabled
	if len(triggered) != 0 {
		t.Errorf("Expected 0 triggered entries (disabled), got %d", len(triggered))
	}
}

func TestWorldBookManager_TriggerEntries_Priority(t *testing.T) {
	mockStorage := NewMockStorageWithWorldBook()
	manager := NewWorldBookManager(mockStorage)

	// Create a world book with entries of different priorities
	bookData := WorldBookData{
		Name: "Test World",
		Entries: map[string]WorldBookEntryData{
			"entry1": {
				UID:      "entry1",
				Key:      []string{"dragon"},
				Content:  "Low priority dragon info.",
				Constant: false,
				Order:    200, // Lower priority
				Position: 1,
				Disable:  false,
			},
			"entry2": {
				UID:      "entry2",
				Key:      []string{"dragon"},
				Content:  "High priority dragon info.",
				Constant: false,
				Order:    50, // Higher priority
				Position: 1,
				Disable:  false,
			},
		},
	}

	bookJSON, _ := json.Marshal(bookData)
	userID := int64(123)
	book := &storage.WorldBook{
		UserID: &userID,
		Data:   string(bookJSON),
	}

	manager.SaveBook(book)

	// Create messages that mention "dragon"
	messages := []storage.HistoryItem{
		{Role: "user", Content: "Tell me about dragons"},
	}

	// Trigger entries
	triggered, err := manager.TriggerEntries(book.ID, messages)
	if err != nil {
		t.Fatalf("Failed to trigger entries: %v", err)
	}

	// Should trigger both entries
	if len(triggered) != 2 {
		t.Errorf("Expected 2 triggered entries, got %d", len(triggered))
	}

	// First entry should be entry2 (higher priority = lower order value)
	if len(triggered) > 0 && triggered[0].UID != "entry2" {
		t.Errorf("Expected entry2 to be first (higher priority), got %s", triggered[0].UID)
	}
}

func TestWorldBookManager_TriggerEntries_Regex(t *testing.T) {
	mockStorage := NewMockStorageWithWorldBook()
	manager := NewWorldBookManager(mockStorage)

	// Create a world book with regex pattern
	bookData := WorldBookData{
		Name: "Test World",
		Entries: map[string]WorldBookEntryData{
			"entry1": {
				UID:      "entry1",
				Key:      []string{"/drag.*/"}, // Regex pattern - matches anything starting with "drag"
				Content:  "Dragons and dragoons.",
				Constant: false,
				Order:    100,
				Position: 1,
				Disable:  false,
			},
		},
	}

	bookJSON, _ := json.Marshal(bookData)
	userID := int64(123)
	book := &storage.WorldBook{
		UserID: &userID,
		Data:   string(bookJSON),
	}

	manager.SaveBook(book)

	// Test with "dragon"
	messages1 := []storage.HistoryItem{
		{Role: "user", Content: "Tell me about dragons"},
	}

	triggered1, err := manager.TriggerEntries(book.ID, messages1)
	if err != nil {
		t.Fatalf("Failed to trigger entries: %v", err)
	}

	if len(triggered1) != 1 {
		t.Errorf("Expected 1 triggered entry for 'dragons', got %d", len(triggered1))
	}

	// Test with "dragoon"
	messages2 := []storage.HistoryItem{
		{Role: "user", Content: "Tell me about dragoons"},
	}

	triggered2, err := manager.TriggerEntries(book.ID, messages2)
	if err != nil {
		t.Fatalf("Failed to trigger entries: %v", err)
	}

	if len(triggered2) != 1 {
		t.Errorf("Expected 1 triggered entry for 'dragoons', got %d", len(triggered2))
	}
}

func TestWorldBookManager_UpdateEntryStatus(t *testing.T) {
	mockStorage := NewMockStorageWithWorldBook()
	manager := NewWorldBookManager(mockStorage)

	// Create a world book with an entry
	bookData := WorldBookData{
		Name: "Test World",
		Entries: map[string]WorldBookEntryData{
			"entry1": {
				UID:      "entry1",
				Key:      []string{"dragon"},
				Content:  "Dragons are powerful creatures.",
				Constant: false,
				Order:    100,
				Position: 1,
				Disable:  false,
			},
		},
	}

	bookJSON, _ := json.Marshal(bookData)
	userID := int64(123)
	book := &storage.WorldBook{
		UserID: &userID,
		Data:   string(bookJSON),
	}

	manager.SaveBook(book)

	// Get the entry
	entries, _ := mockStorage.ListWorldBookEntries(book.ID)
	if len(entries) == 0 {
		t.Fatal("No entries found")
	}

	entry := entries[0]

	// Disable the entry
	err := manager.UpdateEntryStatus(entry.ID, false)
	if err != nil {
		t.Fatalf("Failed to update entry status: %v", err)
	}

	// Verify the entry is disabled
	updatedEntry, _ := mockStorage.GetWorldBookEntry(entry.ID)
	if updatedEntry.Enabled {
		t.Error("Entry should be disabled")
	}

	// Enable the entry
	err = manager.UpdateEntryStatus(entry.ID, true)
	if err != nil {
		t.Fatalf("Failed to update entry status: %v", err)
	}

	// Verify the entry is enabled
	updatedEntry, _ = mockStorage.GetWorldBookEntry(entry.ID)
	if !updatedEntry.Enabled {
		t.Error("Entry should be enabled")
	}
}

func TestWorldBookManager_UpdateEntry(t *testing.T) {
	mockStorage := NewMockStorageWithWorldBook()
	manager := NewWorldBookManager(mockStorage)

	// Create a world book with an entry
	bookData := WorldBookData{
		Name: "Test World",
		Entries: map[string]WorldBookEntryData{
			"entry1": {
				UID:      "entry1",
				Key:      []string{"dragon"},
				Content:  "Dragons are powerful creatures.",
				Constant: false,
				Order:    100,
				Position: 1,
				Disable:  false,
			},
		},
	}

	bookJSON, _ := json.Marshal(bookData)
	userID := int64(123)
	book := &storage.WorldBook{
		UserID: &userID,
		Data:   string(bookJSON),
	}

	manager.SaveBook(book)

	// Get the entry
	entries, _ := mockStorage.ListWorldBookEntries(book.ID)
	if len(entries) == 0 {
		t.Fatal("No entries found")
	}

	entry := entries[0]

	// Update the entry content
	entry.Content = "Updated dragon information."
	err := manager.UpdateEntry(entry)
	if err != nil {
		t.Fatalf("Failed to update entry: %v", err)
	}

	// Verify the entry was updated
	updatedEntry, _ := mockStorage.GetWorldBookEntry(entry.ID)
	if updatedEntry.Content != "Updated dragon information." {
		t.Errorf("Expected updated content, got '%s'", updatedEntry.Content)
	}
}
