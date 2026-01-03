package command

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/config"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/storage"
	"github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/telegram/api"
)

// AdminCacheTTL is the time-to-live for cached admin lists (120 seconds)
const AdminCacheTTL = 120

// IsGroupAdmin checks if a user is an admin in a group
// It first checks the cache, and if not found, fetches from Telegram API
func IsGroupAdmin(chatID int64, userID int64, ctx *config.WorkerContext) (bool, error) {
	// Get cached admin list
	admins, err := ctx.DB.GetGroupAdmins(chatID)
	if err == nil && len(admins) > 0 {
		// Cache hit - check if user is in admin list
		for _, admin := range admins {
			if admin.User.ID == userID {
				return true, nil
			}
		}
		return false, nil
	}

	// Cache miss or error - fetch from Telegram API
	admins, err = fetchGroupAdmins(chatID, ctx)
	if err != nil {
		return false, fmt.Errorf("failed to fetch group admins: %w", err)
	}

	// Cache the admin list
	if err := ctx.DB.SaveGroupAdmins(chatID, admins, AdminCacheTTL); err != nil {
		log.Printf("Warning: failed to cache group admins: %v", err)
		// Continue even if caching fails
	}

	// Check if user is in admin list
	for _, admin := range admins {
		if admin.User.ID == userID {
			return true, nil
		}
	}

	return false, nil
}

// fetchGroupAdmins fetches the list of administrators from Telegram API
func fetchGroupAdmins(chatID int64, ctx *config.WorkerContext) ([]storage.ChatMember, error) {
	// Get the bot API client from context
	botAPI, ok := ctx.Bot.(*api.Client)
	if !ok {
		return nil, fmt.Errorf("invalid bot API client type")
	}

	// Fetch administrators from Telegram
	tgAdmins, err := botAPI.GetChatAdministrators(chatID)
	if err != nil {
		return nil, fmt.Errorf("telegram API error: %w", err)
	}

	// Convert to our storage format
	admins := make([]storage.ChatMember, 0, len(tgAdmins))
	for _, tgAdmin := range tgAdmins {
		// Only include actual administrators and creators
		// Exclude restricted users and members
		if tgAdmin.Status == "administrator" || tgAdmin.Status == "creator" {
			admins = append(admins, storage.ChatMember{
				User: storage.User{
					ID: tgAdmin.User.ID,
				},
				Status: tgAdmin.Status,
			})
		}
	}

	return admins, nil
}

// ClearAdminCache clears the cached admin list for a group
// This can be called when admin permissions change
func ClearAdminCache(chatID int64, ctx *config.WorkerContext) error {
	// Save an empty list with 0 TTL to effectively clear the cache
	return ctx.DB.SaveGroupAdmins(chatID, []storage.ChatMember{}, 0)
}

// RefreshAdminCache forces a refresh of the admin cache for a group
func RefreshAdminCache(chatID int64, ctx *config.WorkerContext) error {
	admins, err := fetchGroupAdmins(chatID, ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch group admins: %w", err)
	}

	return ctx.DB.SaveGroupAdmins(chatID, admins, AdminCacheTTL)
}

// IsUserAuthorized checks if a user is authorized to use the bot
// This checks whitelist, generous mode, and group settings
func IsUserAuthorized(message *tgbotapi.Message, ctx *config.WorkerContext, globalConfig *config.Config) (bool, string) {
	chatID := message.Chat.ID

	// Check if generous mode is enabled (open access)
	if globalConfig.IAmAGenerousPerson {
		return true, ""
	}

	// Private chat authorization
	if message.Chat.IsPrivate() {
		// Check private chat whitelist
		for _, allowedID := range globalConfig.ChatWhiteList {
			if allowedID == fmt.Sprintf("%d", chatID) {
				return true, ""
			}
		}
		return false, fmt.Sprintf("Unauthorized. Your chat_id: %d", chatID)
	}

	// Group chat authorization
	if message.Chat.IsGroup() || message.Chat.IsSuperGroup() {
		// Check if group bot is enabled
		if !globalConfig.GroupChatBotEnable {
			return false, "Group chat bot is disabled"
		}

		// Check group whitelist
		for _, allowedID := range globalConfig.ChatGroupWhiteList {
			if allowedID == fmt.Sprintf("%d", chatID) {
				return true, ""
			}
		}
		return false, fmt.Sprintf("Unauthorized group. Group chat_id: %d", chatID)
	}

	// Unknown chat type
	return false, fmt.Sprintf("Unsupported chat type. Your chat_id: %d", chatID)
}
