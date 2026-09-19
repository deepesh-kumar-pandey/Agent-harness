package session

import providerpkg "agent-harness/internal/provider"

type Session struct {
	ID       string
	Messages []providerpkg.Message
}

func NewSession(id string) *Session {
	return &Session{
		ID:       id,
		Messages: make([]providerpkg.Message, 0),
	}
}
