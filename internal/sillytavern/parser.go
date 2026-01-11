package sillytavern

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"io"
	"strings"

	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
)

// PNGParser handles parsing of SillyTavern character cards from PNG images
type PNGParser struct{}

// NewPNGParser creates a new PNG parser
func NewPNGParser() *PNGParser {
	return &PNGParser{}
}

// ParseCharacterCardFromPNG extracts character card data from a PNG image
// SillyTavern stores character card data in PNG tEXt chunks with key "chara"
func (p *PNGParser) ParseCharacterCardFromPNG(imageData []byte) (*storage.CharacterCard, error) {
	reader := bytes.NewReader(imageData)

	// Decode PNG to access chunks
	_, err := png.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("invalid PNG image: %w", err)
	}

	// Reset reader to start
	reader.Seek(0, io.SeekStart)

	// Extract character data from PNG chunks
	cardJSON, err := p.extractCharaChunk(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to extract character data: %w", err)
	}

	// Validate the character card format
	var cardData CharacterCardV2
	if err := json.Unmarshal([]byte(cardJSON), &cardData); err != nil {
		return nil, fmt.Errorf("invalid character card JSON: %w", err)
	}

	// Validate it's a V2 card
	if cardData.Spec != "chara_card_v2" {
		return nil, errors.New("only SillyTavern V2 format is supported")
	}

	// Create the character card model
	card := &storage.CharacterCard{
		Name:   cardData.Data.Name,
		Avatar: base64.StdEncoding.EncodeToString(imageData),
		Data:   cardJSON,
	}

	return card, nil
}

// extractCharaChunk extracts the "chara" tEXt chunk from a PNG image
func (p *PNGParser) extractCharaChunk(reader io.Reader) (string, error) {
	// Read PNG signature
	signature := make([]byte, 8)
	if _, err := io.ReadFull(reader, signature); err != nil {
		return "", fmt.Errorf("failed to read PNG signature: %w", err)
	}

	// Verify PNG signature
	expectedSig := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	if !bytes.Equal(signature, expectedSig) {
		return "", errors.New("invalid PNG signature")
	}

	// Read chunks until we find "tEXt" with key "chara"
	for {
		// Read chunk length (4 bytes)
		lengthBytes := make([]byte, 4)
		if _, err := io.ReadFull(reader, lengthBytes); err != nil {
			if err == io.EOF {
				return "", errors.New("character data not found in PNG")
			}
			return "", fmt.Errorf("failed to read chunk length: %w", err)
		}

		length := uint32(lengthBytes[0])<<24 | uint32(lengthBytes[1])<<16 |
			uint32(lengthBytes[2])<<8 | uint32(lengthBytes[3])

		// Read chunk type (4 bytes)
		chunkType := make([]byte, 4)
		if _, err := io.ReadFull(reader, chunkType); err != nil {
			return "", fmt.Errorf("failed to read chunk type: %w", err)
		}

		// Read chunk data
		chunkData := make([]byte, length)
		if _, err := io.ReadFull(reader, chunkData); err != nil {
			return "", fmt.Errorf("failed to read chunk data: %w", err)
		}

		// Read CRC (4 bytes) - we don't validate it
		crc := make([]byte, 4)
		if _, err := io.ReadFull(reader, crc); err != nil {
			return "", fmt.Errorf("failed to read chunk CRC: %w", err)
		}

		// Check if this is a tEXt chunk
		if string(chunkType) == "tEXt" {
			// tEXt format: keyword\0text
			nullIndex := bytes.IndexByte(chunkData, 0)
			if nullIndex == -1 {
				continue
			}

			keyword := string(chunkData[:nullIndex])
			text := chunkData[nullIndex+1:]

			// Check if this is the "chara" chunk
			if keyword == "chara" {
				// The text might be base64 encoded
				decoded, err := base64.StdEncoding.DecodeString(string(text))
				if err != nil {
					// If it's not base64, use it as-is
					return string(text), nil
				}
				return string(decoded), nil
			}
		}

		// Check for IEND chunk (end of PNG)
		if string(chunkType) == "IEND" {
			return "", errors.New("character data not found in PNG")
		}
	}
}

// ValidateCharacterCardV2 validates a character card in V2 format
func (p *PNGParser) ValidateCharacterCardV2(cardJSON string) error {
	var cardData CharacterCardV2
	if err := json.Unmarshal([]byte(cardJSON), &cardData); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Check spec
	if cardData.Spec != "chara_card_v2" {
		return fmt.Errorf("invalid spec: expected 'chara_card_v2', got '%s'", cardData.Spec)
	}

	// Check spec version
	if cardData.SpecVersion == "" {
		return errors.New("spec_version is required")
	}

	// Check required fields
	if cardData.Data.Name == "" {
		return errors.New("character name is required")
	}

	return nil
}

// UploadCard is a convenience method that combines parsing and saving
func (m *CharacterCardManager) UploadCard(userID *int64, imageData []byte) (*storage.CharacterCard, error) {
	parser := NewPNGParser()

	// Parse the PNG image
	card, err := parser.ParseCharacterCardFromPNG(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse character card: %w", err)
	}

	// Set the owner
	card.UserID = userID

	// Save the card
	if err := m.SaveCard(card); err != nil {
		return nil, fmt.Errorf("failed to save character card: %w", err)
	}

	return card, nil
}

// UploadCardFromJSON uploads a character card from raw JSON data
func (m *CharacterCardManager) UploadCardFromJSON(userID *int64, cardJSON string, avatarBase64 string) (*storage.CharacterCard, error) {
	parser := NewPNGParser()

	// Validate the JSON
	if err := parser.ValidateCharacterCardV2(cardJSON); err != nil {
		return nil, fmt.Errorf("invalid character card: %w", err)
	}

	// Parse to get the name
	var cardData CharacterCardV2
	if err := json.Unmarshal([]byte(cardJSON), &cardData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Create the card
	card := &storage.CharacterCard{
		UserID: userID,
		Name:   cardData.Data.Name,
		Avatar: avatarBase64,
		Data:   cardJSON,
	}

	// Save the card
	if err := m.SaveCard(card); err != nil {
		return nil, fmt.Errorf("failed to save character card: %w", err)
	}

	return card, nil
}

// ExportCardToPNG exports a character card to PNG format (creates a simple PNG with tEXt chunk)
func (p *PNGParser) ExportCardToPNG(card *storage.CharacterCard) ([]byte, error) {
	// If the card already has an avatar, decode and use it
	if card.Avatar != "" && strings.HasPrefix(card.Avatar, "data:image/png;base64,") {
		// Extract base64 data
		base64Data := strings.TrimPrefix(card.Avatar, "data:image/png;base64,")
		imageData, err := base64.StdEncoding.DecodeString(base64Data)
		if err == nil {
			return imageData, nil
		}
	} else if card.Avatar != "" {
		// Try direct base64 decode
		imageData, err := base64.StdEncoding.DecodeString(card.Avatar)
		if err == nil {
			// Verify it's a valid PNG
			if bytes.HasPrefix(imageData, []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
				return imageData, nil
			}
		}
	}

	// If no valid avatar, return error (we need the original PNG to preserve it)
	return nil, errors.New("cannot export card without original PNG avatar data")
}
