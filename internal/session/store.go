package session

import (
	"fmt"
)

type Store interface {
	Get(id string) (*Session, error)
	Set(session *Session) error
	Delete(id string) error
}

type SessionStore struct {
	sessions map[string]*Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*Session),
	}
}

func (store *SessionStore) Set(session *Session) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	store.sessions[session.ID] = session
	return nil
}

func (store *SessionStore) Get(id string) (*Session, error) {
	session, exists := store.sessions[id]

	if !exists {
		return nil, fmt.Errorf("session not found %s", id)
	}

	return session, nil
}

func (store *SessionStore) Delete(id string) error {
	_, exists := store.sessions[id]

	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	delete(store.sessions, id)
	return nil
}
