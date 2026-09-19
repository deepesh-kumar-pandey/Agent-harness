package tools

import (
	"fmt"

	providerpkg "agent-harness/internal/provider"
)

type Conversation struct {
	Messages []providerpkg.Message
}

func NewConversation() *Conversation {
	return &Conversation{
		Messages: make([]providerpkg.Message, 0),
	}
}

func (c *Conversation) AddMessage(message providerpkg.Message) error {
	if message.Role == "" {
		return fmt.Errorf("message role is required")
	}

	c.Messages = append(c.Messages, message)
	return nil
}

func (c *Conversation) GetMessages() []providerpkg.Message {
	return c.Messages
}
