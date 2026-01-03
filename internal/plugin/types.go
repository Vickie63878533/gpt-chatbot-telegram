package plugin

// TemplateInputType defines how input data should be parsed
type TemplateInputType string

const (
	InputTypeJSON           TemplateInputType = "json"
	InputTypeSpaceSeparated TemplateInputType = "space-separated"
	InputTypeCommaSeparated TemplateInputType = "comma-separated"
	InputTypeText           TemplateInputType = "text"
)

// TemplateBodyType defines the request body format
type TemplateBodyType string

const (
	BodyTypeJSON TemplateBodyType = "json"
	BodyTypeForm TemplateBodyType = "form"
	BodyTypeText TemplateBodyType = "text"
)

// TemplateResponseType defines how response should be parsed
type TemplateResponseType string

const (
	ResponseTypeJSON TemplateResponseType = "json"
	ResponseTypeText TemplateResponseType = "text"
	ResponseTypeBlob TemplateResponseType = "blob"
)

// TemplateOutputType defines how output should be sent to Telegram
type TemplateOutputType string

const (
	OutputTypeText     TemplateOutputType = "text"
	OutputTypeImage    TemplateOutputType = "image"
	OutputTypeHTML     TemplateOutputType = "html"
	OutputTypeMarkdown TemplateOutputType = "markdown"
)

// RequestTemplate defines a plugin template
type RequestTemplate struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Input   struct {
		Type     TemplateInputType `json:"type"`
		Required bool              `json:"required"`
	} `json:"input"`
	Query map[string]string `json:"query,omitempty"`
	Body  *struct {
		Type    TemplateBodyType `json:"type"`
		Content interface{}      `json:"content"` // can be map[string]string or string
	} `json:"body,omitempty"`
	Response struct {
		Content struct {
			InputType  TemplateResponseType `json:"input_type"`
			OutputType TemplateOutputType   `json:"output_type"`
			Output     string               `json:"output"`
		} `json:"content"`
		Error struct {
			InputType  TemplateResponseType `json:"input_type"`
			OutputType TemplateOutputType   `json:"output_type"`
			Output     string               `json:"output"`
		} `json:"error"`
	} `json:"response"`
}

// ExecuteResult represents the result of executing a plugin
type ExecuteResult struct {
	Type    TemplateOutputType
	Content string
}

// PluginConfig represents a plugin configuration
type PluginConfig struct {
	Value       string   `json:"value"`
	Description string   `json:"description,omitempty"`
	Scope       []string `json:"scope,omitempty"`
}
