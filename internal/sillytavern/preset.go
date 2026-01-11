package sillytavern

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// PresetManager manages SillyTavern presets
type PresetManager struct {
	storage storage.Storage
}

// NewPresetManager creates a new PresetManager
func NewPresetManager(storage storage.Storage) *PresetManager {
	return &PresetManager{
		storage: storage,
	}
}

// PresetData represents the preset configuration parameters
type PresetData struct {
	Name               string   `json:"name"`
	Temperature        float64  `json:"temperature,omitempty"`
	TopP               float64  `json:"top_p,omitempty"`
	TopK               int      `json:"top_k,omitempty"`
	MaxTokens          int      `json:"max_tokens,omitempty"`
	PresencePenalty    float64  `json:"presence_penalty,omitempty"`
	FrequencyPenalty   float64  `json:"frequency_penalty,omitempty"`
	RepetitionPenalty  float64  `json:"repetition_penalty,omitempty"`
	StopSequences      []string `json:"stop_sequences,omitempty"`
	// Additional provider-specific parameters can be stored in extensions
	Extensions         map[string]interface{} `json:"extensions,omitempty"`
}

// LoadPreset loads a preset by ID
func (m *PresetManager) LoadPreset(userID *int64, presetID uint) (*storage.Preset, error) {
	preset, err := m.storage.GetPreset(presetID)
	if err != nil {
		return nil, fmt.Errorf("failed to load preset: %w", err)
	}

	// Check if user has access to this preset
	if preset.UserID != nil && userID != nil && *preset.UserID != *userID {
		return nil, errors.New("access denied: preset belongs to another user")
	}

	return preset, nil
}

// SavePreset saves a preset
func (m *PresetManager) SavePreset(preset *storage.Preset) error {
	// Validate the preset data is valid JSON
	var presetData PresetData
	if err := json.Unmarshal([]byte(preset.Data), &presetData); err != nil {
		return fmt.Errorf("invalid preset data: %w", err)
	}

	// Update the name field from the data
	preset.Name = presetData.Name

	// Validate API type
	if preset.APIType == "" {
		return errors.New("API type is required")
	}

	// Create or update the preset
	if preset.ID == 0 {
		if err := m.storage.CreatePreset(preset); err != nil {
			return fmt.Errorf("failed to create preset: %w", err)
		}
	} else {
		if err := m.storage.UpdatePreset(preset); err != nil {
			return fmt.Errorf("failed to update preset: %w", err)
		}
	}

	return nil
}

// ListPresets lists all presets accessible to the user for a specific API type
func (m *PresetManager) ListPresets(userID *int64, apiType string) ([]*storage.Preset, error) {
	presets, err := m.storage.ListPresets(userID, apiType)
	if err != nil {
		return nil, fmt.Errorf("failed to list presets: %w", err)
	}

	return presets, nil
}

// ActivatePreset activates a preset for the user
func (m *PresetManager) ActivatePreset(userID *int64, presetID uint) error {
	// First verify the preset exists and user has access
	preset, err := m.LoadPreset(userID, presetID)
	if err != nil {
		return err
	}

	// Check if this is a global preset or user's own preset
	if preset.UserID != nil && userID != nil && *preset.UserID != *userID {
		return errors.New("cannot activate preset belonging to another user")
	}

	if err := m.storage.ActivatePreset(userID, presetID); err != nil {
		return fmt.Errorf("failed to activate preset: %w", err)
	}

	return nil
}

// GetActivePreset gets the currently active preset for the user and API type
func (m *PresetManager) GetActivePreset(userID *int64, apiType string) (*storage.Preset, error) {
	preset, err := m.storage.GetActivePreset(userID, apiType)
	if err != nil {
		return nil, fmt.Errorf("failed to get active preset: %w", err)
	}

	return preset, nil
}

// ParsePresetData parses the JSON data from a preset
func (m *PresetManager) ParsePresetData(data string) (*PresetData, error) {
	var presetData PresetData
	if err := json.Unmarshal([]byte(data), &presetData); err != nil {
		return nil, fmt.Errorf("failed to parse preset data: %w", err)
	}

	return &presetData, nil
}
