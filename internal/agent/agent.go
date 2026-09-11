package tools

import (
	"agent-harness/internal/provider"
	"agent-harness/internal/tools"
)

type Agent struct {
	registry     *tools.ToolRegistry
	conversation *Conversation
}

func NewAgent(registry *tools.ToolRegistry) *Agent {
	return &Agent{
		registry:     registry,
		conversation: NewConversation(),
	}
}

func (a *Agent) AddMessage(message provider.Message) error {
	return a.conversation.AddMessage(message)
}

func (a *Agent) GetMessages() []provider.Message {
	return a.conversation.GetMessages()
}

func (a *Agent) GetToolSchemas() ([]map[string]any, error) {
	return a.registry.Schemas()
}

func (a *Agent) ExecuteTool(name string, args map[string]any) (any, error) {
	tool, err := a.registry.Get(name)

	if err != nil {
		return nil, err
	}

	return tool.Execute(args)
}

func (a *Agent) Run(
	name string,
	args map[string]any,
) (any, error) {
	return a.ExecuteTool(name, args)
}
