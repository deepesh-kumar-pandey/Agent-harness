package session

import (
	"fmt"
)

type MemoryStore struct {
	sessions map[string]*Session
}

var _ Store = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions: make(map[string]*Session),
	}
}

func (s *MemoryStore) Get(id string) (*Session, error) {
	session, ok := s.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session %q not found", id)
	}

	return session, nil
}

func (s *MemoryStore) Set(session *Session) error {
	s.sessions[session.ID] = session
	return nil
}

func (s *MemoryStore) Delete(id string) error {
	_, ok := s.sessions[id]
	if !ok {
		return fmt.Errorf("session %q not found", id)
	}

	delete(s.sessions, id)

	return nil
}

func (s *MemoryStore) List() []*Session {
	sessions := make([]*Session, 0, len(s.sessions))

	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}
