package session

import (
	providerpkg "agent-harness/internal/provider"
	"time"
)

type Session struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Metadata  map[string]string
	Messages  []providerpkg.Message
}

func NewSession(id string) *Session {
	now := time.Now()

	return &Session{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  make(map[string]string),
		Messages:  make([]providerpkg.Message, 0),
	}
}

// Rename updates the session display name.
func (session *Session) Rename(name string) {
	session.Metadata["name"] = name
	session.UpdatedAt = time.Now()
}
