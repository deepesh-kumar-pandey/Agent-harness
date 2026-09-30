package session

import (
	"encoding/json"
	"fmt"
	"time"

	providerpkg "agent-harness/internal/provider"
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
	if session.Metadata == nil {
		session.Metadata = make(map[string]string)
	}

	session.Metadata["name"] = name
	session.UpdatedAt = time.Now()
}

// Export returns the session as JSON.
func (session *Session) Export() ([]byte, error) {
	return json.MarshalIndent(session, "", "  ")
}

// Import restores a session from JSON.
func Import(data []byte) (*Session, error) {
	var session Session

	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	if session.ID == "" {
		return nil, fmt.Errorf("session ID must not be empty")
	}

	return &session, nil
}
