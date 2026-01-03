# Authentication and Authorization

This module provides authentication and authorization functionality for the Telegram bot.

## Features

### Group Admin Verification

The `IsGroupAdmin` function checks if a user is an administrator in a Telegram group:

- **Caching**: Admin lists are cached for 120 seconds to reduce API calls
- **Automatic Fetching**: If cache misses, automatically fetches from Telegram API
- **Status Filtering**: Only includes users with "administrator" or "creator" status

```go
isAdmin, err := IsGroupAdmin(chatID, userID, ctx)
if err != nil {
    // Handle error
}
if isAdmin {
    // User is an admin
}
```

### Admin Cache Management

#### Cache TTL
- Admin lists are cached for **120 seconds** (defined by `AdminCacheTTL` constant)
- After expiration, the next check will fetch fresh data from Telegram API

#### Manual Cache Control

```go
// Clear the cache for a specific group
err := ClearAdminCache(chatID, ctx)

// Force refresh the cache
err := RefreshAdminCache(chatID, ctx)
```

### User Authorization

The `IsUserAuthorized` function checks if a user is authorized to use the bot:

```go
authorized, message := IsUserAuthorized(message, ctx, globalConfig)
if !authorized {
    // Send rejection message to user
    // message contains the reason (including chat_id)
}
```

#### Authorization Logic

1. **Generous Mode**: If `I_AM_A_GENEROUS_PERSON` is enabled, all users are authorized
2. **Private Chats**: Checks against `CHAT_WHITE_LIST`
3. **Group Chats**: 
   - Checks if `GROUP_CHAT_BOT_ENABLE` is enabled
   - Checks against `CHAT_GROUP_WHITE_LIST`

#### Rejection Messages

- Private chat: `"Unauthorized. Your chat_id: {chatID}"`
- Group chat: `"Unauthorized group. Group chat_id: {chatID}"`
- Disabled group bot: `"Group chat bot is disabled"`

## Auth Checkers

Pre-defined auth checkers for commands:

### NoAuthRequired
Always allows access (for public commands like `/start`, `/help`)

```go
func (c *MyCommand) NeedAuth() AuthChecker {
    return NoAuthRequired
}
```

### AdminOnly
Requires admin permissions in groups, always allows in private chats

```go
func (c *MyCommand) NeedAuth() AuthChecker {
    return AdminOnly
}
```

### ShareModeGroup
For commands that work in share mode groups (same as AdminOnly)

```go
func (c *MyCommand) NeedAuth() AuthChecker {
    return ShareModeGroup
}
```

## Implementation Details

### Telegram API Integration

The module uses the `api.Client` wrapper to fetch admin lists:

```go
botAPI, ok := ctx.Bot.(*api.Client)
tgAdmins, err := botAPI.GetChatAdministrators(chatID)
```

### Storage Format

Admin lists are stored in the `group_admins` table with:
- `chat_id`: The group chat ID
- `admins`: JSON array of ChatMember objects
- `expires_at`: Unix timestamp for expiration
- `updated_at`: Last update timestamp

### Error Handling

- **Cache Errors**: If caching fails, the function continues (logs warning)
- **API Errors**: Returns error to caller for proper handling
- **Type Errors**: Returns error if bot API client type is invalid

## Requirements Validation

This implementation validates:

- **Requirement 8.6**: Admin permission verification for commands
- **Requirement 8.7**: Admin cache with 120-second TTL

## Testing

The auth module integrates with existing command tests. To test:

```bash
go test ./internal/telegram/command/...
```

## Usage Example

```go
// In a command handler
func (c *SetEnvCommand) Handle(message *tgbotapi.Message, args string, ctx *config.WorkerContext) error {
    // Check if user is admin (for group chats)
    if message.Chat.IsGroup() || message.Chat.IsSuperGroup() {
        isAdmin, err := IsGroupAdmin(message.Chat.ID, message.From.ID, ctx)
        if err != nil {
            return fmt.Errorf("failed to check admin status: %w", err)
        }
        if !isAdmin {
            return fmt.Errorf("only administrators can use this command")
        }
    }
    
    // Process command...
    return nil
}
```
