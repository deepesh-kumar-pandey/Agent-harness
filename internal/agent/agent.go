package tools

import (
	"fmt"

	providerpkg "agent-harness/internal/provider"
	sessionpkg "agent-harness/internal/session"
	toolspkg "agent-harness/internal/tools"
)

type Agent struct {
	registry     *toolspkg.ToolRegistry
	conversation *Conversation
	session      *sessionpkg.Session
}

func NewAgent(registry *toolspkg.ToolRegistry) *Agent {
	return &Agent{
		registry:     registry,
		conversation: NewConversation(),
	}
}

func (a *Agent) SetSession(session *sessionpkg.Session) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	a.session = session
	return nil
}

func (a *Agent) AddMessage(message providerpkg.Message) error {
	if a.session != nil {
		if message.Role == "" {
			return fmt.Errorf("message role is required")
		}

		a.session.Messages = append(
			a.session.Messages,
			message,
		)

		return nil
	}

	return a.conversation.AddMessage(message)
}

func (a *Agent) GetMessages() []providerpkg.Message {
	if a.session != nil {
		return a.session.Messages
	}

	return a.conversation.GetMessages()
}

func (a *Agent) GetToolSchemas() ([]map[string]any, error) {
	return a.registry.Schemas()
}

func (a *Agent) ExecuteTool(
	name string,
	args map[string]any,
) (any, error) {
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
