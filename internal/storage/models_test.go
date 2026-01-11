package storage

import (
	"testing"
)

// TestSillyTavernModels_TableNames verifies that all SillyTavern models have correct table names
func TestSillyTavernModels_TableNames(t *testing.T) {
	tests := []struct {
		name      string
		model     interface{ TableName() string }
		wantTable string
	}{
		{
			name:      "CharacterCard",
			model:     CharacterCard{},
			wantTable: "character_cards",
		},
		{
			name:      "WorldBook",
			model:     WorldBook{},
			wantTable: "world_books",
		},
		{
			name:      "WorldBookEntry",
			model:     WorldBookEntry{},
			wantTable: "world_book_entries",
		},
		{
			name:      "Preset",
			model:     Preset{},
			wantTable: "presets",
		},
		{
			name:      "RegexPattern",
			model:     RegexPattern{},
			wantTable: "regex_patterns",
		},
		{
			name:      "LoginToken",
			model:     LoginToken{},
			wantTable: "login_tokens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.TableName(); got != tt.wantTable {
				t.Errorf("%s.TableName() = %v, want %v", tt.name, got, tt.wantTable)
			}
		})
	}
}

// TestSillyTavernModels_Instantiation verifies that all models can be instantiated
func TestSillyTavernModels_Instantiation(t *testing.T) {
	// Test that models can be created without panics
	_ = CharacterCard{}
	_ = WorldBook{}
	_ = WorldBookEntry{}
	_ = Preset{}
	_ = RegexPattern{}
	_ = LoginToken{}
}
