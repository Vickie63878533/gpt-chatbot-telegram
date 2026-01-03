# i18n Package

This package provides internationalization (i18n) support for the Telegram bot.

## Supported Languages

- **English** (`en`, `en-us`)
- **Simplified Chinese** (`zh-cn`, `zh-hans`, `cn`)
- **Traditional Chinese** (`zh-tw`, `zh-hk`, `zh-mo`, `zh-hant`)
- **Portuguese** (`pt`, `pt-br`)

## Usage

```go
import "github.com/tbxark/ChatGPT-Telegram-Workers/go_version/internal/i18n"

// Load translations for a specific language
i18n := i18n.LoadI18n("zh-cn")

// Access translated strings
systemMessage := i18n.Env.SystemInitMessage
helpText := i18n.Command.Help.Summary
newChatMessage := i18n.Command.New.NewChatStart
```

## Structure

The `I18n` struct contains the following sections:

### Env
- `SystemInitMessage`: Default system initialization message for AI

### Command.Help
Command descriptions for the help menu:
- `Summary`: Help menu header
- `Help`: /help command description
- `New`: /new command description
- `Start`: /start command description
- `Img`: /img command description
- `Setenv`: /setenv command description
- `Setenvs`: /setenvs command description
- `Delenv`: /delenv command description
- `Clearenv`: /clearenv command description
- `System`: /system command description
- `Redo`: /redo command description
- `Models`: /models command description
- `Echo`: /echo command description

### Command.New
- `NewChatStart`: Message shown when starting a new conversation

### CallbackQuery
Callback query related messages:
- `OpenModelList`: Text for opening model list
- `SelectProvider`: Prompt to select AI provider
- `SelectModel`: Prompt to select model
- `ChangeModel`: Confirmation message when model is changed

## Adding a New Language

To add support for a new language:

1. Create a new file `<language_code>.go` in this directory
2. Implement a function that returns a populated `*I18n` struct
3. Add the language code mapping in `LoadI18n()` function in `i18n.go`
4. Add test cases in `i18n_test.go`

Example:

```go
// fr.go
package i18n

func fr() *I18n {
    i := &I18n{}
    
    i.Env.SystemInitMessage = "Vous êtes un assistant utile"
    i.Command.Help.Summary = "Les commandes suivantes sont actuellement prises en charge:\n"
    // ... fill in all other fields
    
    return i
}
```

Then update `LoadI18n()`:

```go
func LoadI18n(lang string) *I18n {
    switch lang {
    // ... existing cases
    case "fr", "fr-fr":
        return fr()
    default:
        return en()
    }
}
```

## Testing

Run tests with:

```bash
go test ./internal/i18n/...
```

The test suite verifies:
- All language codes map correctly
- All required fields are populated for each language
- Language-specific translations are unique and correct
