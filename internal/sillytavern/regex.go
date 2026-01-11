package sillytavern

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// RegexProcessor processes input and output text using regex patterns
type RegexProcessor struct {
	storage storage.Storage
}

// NewRegexProcessor creates a new RegexProcessor
func NewRegexProcessor(storage storage.Storage) *RegexProcessor {
	return &RegexProcessor{
		storage: storage,
	}
}

// ProcessInput applies input regex patterns to the text
func (p *RegexProcessor) ProcessInput(userID *int64, text string) (string, error) {
	patterns, err := p.storage.ListRegexPatterns(userID, "input")
	if err != nil {
		return text, fmt.Errorf("failed to list input patterns: %w", err)
	}

	return p.applyPatterns(patterns, text)
}

// ProcessOutput applies output regex patterns to the text
func (p *RegexProcessor) ProcessOutput(userID *int64, text string) (string, error) {
	patterns, err := p.storage.ListRegexPatterns(userID, "output")
	if err != nil {
		return text, fmt.Errorf("failed to list output patterns: %w", err)
	}

	return p.applyPatterns(patterns, text)
}

// applyPatterns applies a list of regex patterns to text in order
func (p *RegexProcessor) applyPatterns(patterns []*storage.RegexPattern, text string) (string, error) {
	// Filter enabled patterns
	enabledPatterns := make([]*storage.RegexPattern, 0)
	for _, pattern := range patterns {
		if pattern.Enabled {
			enabledPatterns = append(enabledPatterns, pattern)
		}
	}

	// Sort by order
	sort.Slice(enabledPatterns, func(i, j int) bool {
		return enabledPatterns[i].Order < enabledPatterns[j].Order
	})

	// Apply each pattern
	result := text
	for _, pattern := range enabledPatterns {
		// Validate pattern before compiling
		if err := validateRegexPattern(pattern.Pattern); err != nil {
			// Skip invalid patterns but log the error
			continue
		}

		re, err := regexp.Compile(pattern.Pattern)
		if err != nil {
			// Skip patterns that fail to compile
			continue
		}

		result = re.ReplaceAllString(result, pattern.Replace)
	}

	return result, nil
}

// ListPatterns lists all regex patterns for a user and type
func (p *RegexProcessor) ListPatterns(userID *int64, patternType string) ([]*storage.RegexPattern, error) {
	patterns, err := p.storage.ListRegexPatterns(userID, patternType)
	if err != nil {
		return nil, fmt.Errorf("failed to list patterns: %w", err)
	}

	return patterns, nil
}

// UpdatePatternStatus updates the enabled status of a pattern
func (p *RegexProcessor) UpdatePatternStatus(patternID uint, enabled bool) error {
	if err := p.storage.UpdateRegexPatternStatus(patternID, enabled); err != nil {
		return fmt.Errorf("failed to update pattern status: %w", err)
	}

	return nil
}

// UpdatePattern updates a regex pattern
func (p *RegexProcessor) UpdatePattern(pattern *storage.RegexPattern) error {
	// Validate the pattern before saving
	if err := validateRegexPattern(pattern.Pattern); err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Validate pattern type
	if pattern.Type != "input" && pattern.Type != "output" {
		return errors.New("pattern type must be 'input' or 'output'")
	}

	if err := p.storage.UpdateRegexPattern(pattern); err != nil {
		return fmt.Errorf("failed to update pattern: %w", err)
	}

	return nil
}

// validateRegexPattern validates a regex pattern for safety (prevents ReDoS)
func validateRegexPattern(pattern string) error {
	// Check for empty pattern
	if strings.TrimSpace(pattern) == "" {
		return errors.New("pattern cannot be empty")
	}

	// Check pattern length (prevent extremely long patterns)
	if len(pattern) > 1000 {
		return errors.New("pattern is too long (max 1000 characters)")
	}

	// Try to compile the pattern
	_, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex syntax: %w", err)
	}

	// Check for potentially dangerous patterns that could cause ReDoS
	dangerousPatterns := []string{
		`(\w+\s*)+`,           // Nested quantifiers
		`(a+)+`,               // Nested quantifiers
		`(a*)*`,               // Nested quantifiers
		`(a+)*`,               // Nested quantifiers
		`(a|a)*`,              // Alternation with overlap
		`(a|ab)*`,             // Alternation with overlap
		`(\d+\s*)+`,           // Nested quantifiers with digits
		`([a-zA-Z]+\s*)+`,     // Nested quantifiers with character classes
	}

	for _, dangerous := range dangerousPatterns {
		if strings.Contains(pattern, dangerous) {
			return fmt.Errorf("pattern contains potentially dangerous construct: %s", dangerous)
		}
	}

	// Test the pattern with a timeout to detect catastrophic backtracking
	testString := strings.Repeat("a", 100) + "b"
	done := make(chan bool, 1)
	go func() {
		re, _ := regexp.Compile(pattern)
		re.MatchString(testString)
		done <- true
	}()

	select {
	case <-done:
		// Pattern executed successfully
		return nil
	case <-time.After(100 * time.Millisecond):
		// Pattern took too long - likely ReDoS
		return errors.New("pattern execution timeout - possible ReDoS vulnerability")
	}
}
