package sillytavern

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// CharacterCardManager manages SillyTavern character cards
type CharacterCardManager struct {
	storage storage.Storage
}

// NewCharacterCardManager creates a new CharacterCardManager
func NewCharacterCardManager(storage storage.Storage) *CharacterCardManager {
	return &CharacterCardManager{
		storage: storage,
	}
}

// CharacterCardV2 represents the SillyTavern V2 character card format
type CharacterCardV2 struct {
	Spec        string              `json:"spec"`
	SpecVersion string              `json:"spec_version"`
	Data        CharacterCardV2Data `json:"data"`
}

// CharacterCardV2Data contains the actual character data
type CharacterCardV2Data struct {
	Name                     string                 `json:"name"`
	Description              string                 `json:"description"`
	Personality              string                 `json:"personality"`
	Scenario                 string                 `json:"scenario"`
	FirstMes                 string                 `json:"first_mes"`
	MesExample               string                 `json:"mes_example"`
	CreatorNotes             string                 `json:"creator_notes"`
	SystemPrompt             string                 `json:"system_prompt"`
	PostHistoryInstructions  string                 `json:"post_history_instructions"`
	AlternateGreetings       []string               `json:"alternate_greetings"`
	CharacterBook            *CharacterBook         `json:"character_book,omitempty"`
	Tags                     []string               `json:"tags"`
	Creator                  string                 `json:"creator"`
	CharacterVersion         string                 `json:"character_version"`
	Extensions               map[string]interface{} `json:"extensions"`
}

// CharacterBook represents the embedded world book in a character card
type CharacterBook struct {
	Entries []CharacterBookEntry `json:"entries"`
}

// CharacterBookEntry represents a single entry in the character book
type CharacterBookEntry struct {
	Keys           []string `json:"keys"`
	Content        string   `json:"content"`
	Enabled        bool     `json:"enabled"`
	InsertionOrder int      `json:"insertion_order"`
	Position       string   `json:"position"`
}

// LoadCard loads a character card by ID
func (m *CharacterCardManager) LoadCard(userID *int64, cardID uint) (*storage.CharacterCard, error) {
	card, err := m.storage.GetCharacterCard(cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to load character card: %w", err)
	}

	// Check if user has access to this card
	if card.UserID != nil && userID != nil && *card.UserID != *userID {
		return nil, errors.New("access denied: card belongs to another user")
	}

	return card, nil
}

// SaveCard saves a character card
func (m *CharacterCardManager) SaveCard(card *storage.CharacterCard) error {
	// Validate the card data is valid JSON
	var cardData CharacterCardV2
	if err := json.Unmarshal([]byte(card.Data), &cardData); err != nil {
		return fmt.Errorf("invalid character card data: %w", err)
	}

	// Validate it's a V2 card
	if cardData.Spec != "chara_card_v2" {
		return errors.New("only SillyTavern V2 format is supported")
	}

	// Update the name field from the data
	card.Name = cardData.Data.Name

	// Create or update the card
	if card.ID == 0 {
		if err := m.storage.CreateCharacterCard(card); err != nil {
			return fmt.Errorf("failed to create character card: %w", err)
		}
	} else {
		if err := m.storage.UpdateCharacterCard(card); err != nil {
			return fmt.Errorf("failed to update character card: %w", err)
		}
	}

	return nil
}

// ListCards lists all character cards accessible to the user
func (m *CharacterCardManager) ListCards(userID *int64) ([]*storage.CharacterCard, error) {
	cards, err := m.storage.ListCharacterCards(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list character cards: %w", err)
	}

	return cards, nil
}

// ActivateCard activates a character card for the user
func (m *CharacterCardManager) ActivateCard(userID *int64, cardID uint) error {
	// First verify the card exists and user has access
	card, err := m.LoadCard(userID, cardID)
	if err != nil {
		return err
	}

	// Check if this is a global card or user's own card
	if card.UserID != nil && userID != nil && *card.UserID != *userID {
		return errors.New("cannot activate card belonging to another user")
	}

	if err := m.storage.ActivateCharacterCard(userID, cardID); err != nil {
		return fmt.Errorf("failed to activate character card: %w", err)
	}

	return nil
}

// GetActiveCard gets the currently active character card for the user
func (m *CharacterCardManager) GetActiveCard(userID *int64) (*storage.CharacterCard, error) {
	card, err := m.storage.GetActiveCharacterCard(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active character card: %w", err)
	}

	return card, nil
}

// ParseCardData parses the JSON data from a character card
func (m *CharacterCardManager) ParseCardData(data string) (*CharacterCardV2, error) {
	var cardData CharacterCardV2
	if err := json.Unmarshal([]byte(data), &cardData); err != nil {
		return nil, fmt.Errorf("failed to parse character card data: %w", err)
	}

	// Validate it's a V2 card
	if cardData.Spec != "chara_card_v2" {
		return nil, errors.New("only SillyTavern V2 format is supported")
	}

	return &cardData, nil
}
