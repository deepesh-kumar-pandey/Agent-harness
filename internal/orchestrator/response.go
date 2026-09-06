package orchestrator

type ToolCall struct {
	Tool string         `json:"tool"`
	Args map[string]any `json:"args"`
}

type AgentResponse struct {
	Content  string    `json:"content"`
	ToolCall *ToolCall `json:"tool_call"`
}
