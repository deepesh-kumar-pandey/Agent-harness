package session

import (
	providerpkg "agent-harness/internal/provider"
	"time"
)

type Session struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Messages  []providerpkg.Message
}

func NewSession(id string) *Session {
	now := time.Now()

	return &Session{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		Messages:  make([]providerpkg.Message, 0),
	}
}
