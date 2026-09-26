package provider

type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ChatRequest struct {
	Model    string           `json:"model"`
	Messages []Message        `json:"messages"`
	Tools    []ToolDefinition `json:"tools,omitempty"`
}

type ToolCall struct {
	Name      string
	Arguments map[string]any
}

type ChatResponse struct {
	Content   string
	ToolCalls []ToolCall
}

type Provider interface {
	Chat(request ChatRequest) (ChatResponse, error)
}

type LocalModelManager interface {
	ListModels() ([]string, error)
	PullModel(name string) error
}
