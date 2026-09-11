package tools

import (
	"fmt"

	"agent-harness/internal/provider"
)

type Conversation struct {
	Messages []provider.Message
}

func NewConversation() *Conversation {
	return &Conversation{
		Messages: make([]provider.Message, 0),
	}
}

func (c *Conversation) AddMessage(message provider.Message) error {
	if message.Role == "" {
		return fmt.Errorf("message role is required")
	}

	c.Messages = append(c.Messages, message)
	return nil
}

func (c *Conversation) GetMessages() []provider.Message {
	return c.Messages
}
