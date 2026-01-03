package command

import (
	"testing"
)

func TestFormatTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		timestamp int64
		wantEmpty bool
	}{
		{
			name:      "zero timestamp",
			timestamp: 0,
			wantEmpty: false, // Should return "unknown"
		},
		{
			name:      "valid timestamp",
			timestamp: 1704067200, // 2024-01-01 00:00:00 UTC
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTimestamp(tt.timestamp)
			if tt.wantEmpty && result != "" {
				t.Errorf("Expected empty string, got '%s'", result)
			}
			if !tt.wantEmpty && result == "" {
				t.Error("Expected non-empty string")
			}
			if tt.timestamp == 0 && result != "unknown" {
				t.Errorf("Expected 'unknown' for zero timestamp, got '%s'", result)
			}
		})
	}
}
