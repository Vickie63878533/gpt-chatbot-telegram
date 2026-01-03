# Plugin System

The plugin system allows you to extend the bot with custom commands using JSON templates. Plugins can make HTTP requests, process responses, and format output for Telegram.

## Features

- **Template-based**: Define plugins using JSON templates
- **HTTP requests**: Make GET/POST requests to external APIs
- **Variable interpolation**: Use `{{variable}}` syntax for dynamic values
- **Conditional logic**: Use `{{#if}}` blocks for conditional rendering
- **Loops**: Use `{{#each}}` blocks to iterate over arrays
- **Multiple output formats**: Support for text, HTML, Markdown, and images

## Configuration

Plugins are configured using environment variables:

### Plugin Command
```bash
PLUGIN_COMMAND_<name>=<template_json_or_url>
```
The plugin template as JSON string or URL to fetch the template from.

### Plugin Description
```bash
PLUGIN_DESCRIPTION_<name>=<description>
```
Optional description shown in the help command.

### Plugin Scope
```bash
PLUGIN_SCOPE_<name>=<scope1>,<scope2>
```
Optional comma-separated list of scopes (e.g., `all_private_chats,all_chat_administrators`).

### Plugin Environment Variables
```bash
PLUGIN_ENV_<variable>=<value>
```
Environment variables accessible in plugin templates via `{{ENV.variable}}`.

## Template Format

A plugin template is a JSON object with the following structure:

```json
{
  "url": "https://api.example.com/{{DATA}}",
  "method": "GET",
  "headers": {
    "Authorization": "Bearer {{ENV.API_KEY}}"
  },
  "input": {
    "type": "text",
    "required": true
  },
  "query": {
    "param1": "{{DATA}}",
    "param2": "value"
  },
  "body": {
    "type": "json",
    "content": {
      "key": "{{DATA}}"
    }
  },
  "response": {
    "content": {
      "input_type": "json",
      "output_type": "html",
      "output": "<b>{{result}}</b>"
    },
    "error": {
      "input_type": "text",
      "output_type": "text",
      "output": "Error: {{.}}"
    }
  }
}
```

### Template Fields

#### `url` (required)
The API endpoint URL. Supports variable interpolation.

#### `method` (required)
HTTP method: `GET`, `POST`, `PUT`, `DELETE`, etc.

#### `headers` (optional)
HTTP headers as key-value pairs. Values support interpolation.

#### `input` (required)
Defines how user input is processed:
- `type`: Input format
  - `text`: Plain text (default)
  - `json`: Parse as JSON
  - `space-separated`: Split by spaces into array
  - `comma-separated`: Split by commas into array
- `required`: Whether input is required (boolean)

#### `query` (optional)
URL query parameters as key-value pairs. Values support interpolation.

#### `body` (optional)
Request body configuration:
- `type`: Body format
  - `json`: JSON object
  - `form`: URL-encoded form data
  - `text`: Plain text
- `content`: Body content (object or string). Supports interpolation.

#### `response` (required)
Response handling configuration:

##### `content` (required)
Success response handling:
- `input_type`: How to parse response
  - `json`: Parse as JSON
  - `text`: Plain text
  - `blob`: Binary data (for images)
- `output_type`: Output format for Telegram
  - `text`: Plain text
  - `html`: HTML formatting
  - `markdown`: Markdown formatting
  - `image`: Send as image
- `output`: Template for formatting output. Supports interpolation.

##### `error` (required)
Error response handling (same structure as `content`).

## Variable Interpolation

### Simple Variables
```
{{variable}}
```

### Nested Variables
```
{{user.name}}
{{data.items[0]}}
```

### Conditionals
```
{{#if condition}}
  Content when true
{{#else}}
  Content when false
{{/if}}
```

### Loops
```
{{#each item in items}}
  {{item.name}}
{{/each}}
```

### Special Variables

- `{{DATA}}`: User input (formatted according to `input.type`)
- `{{ENV.variable}}`: Plugin environment variable
- `{{.}}`: Current context (useful in loops)

## Examples

### Simple API Call

```bash
export PLUGIN_COMMAND_weather='{"url":"https://api.weather.com/v1/current?q={{DATA}}","method":"GET","input":{"type":"text","required":true},"response":{"content":{"input_type":"json","output_type":"text","output":"Temperature: {{temp}}°C"},"error":{"input_type":"text","output_type":"text","output":"Error: {{.}}"}}}'
export PLUGIN_DESCRIPTION_weather="Get current weather"
```

Usage: `/weather London`

### Dictionary Lookup

See `plugins/dicten.json` for a complete example that:
- Fetches word definitions from an API
- Parses JSON response
- Formats output as HTML with loops and conditionals

### DNS Query

See `plugins/dns.json` for an example that:
- Makes DNS queries via Cloudflare API
- Uses query parameters
- Formats DNS records as HTML

### Image Generation

See `plugins/pollinations.json` for an example that:
- Generates images from text prompts
- Uses environment variables for configuration
- Returns image URLs

## Testing

Run plugin tests:
```bash
go test ./internal/plugin/
```

## Implementation Details

The plugin system consists of:

- `types.go`: Type definitions for templates and configs
- `interpolate.go`: Variable interpolation engine
- `template.go`: HTTP request execution and response handling
- `loader.go`: Plugin loading from environment and files
- `plugin.go` (in command package): Command handler for plugins

Plugins are automatically registered as commands when the bot starts, and they appear in the `/help` command output.
