package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type FileSessionStore struct {
	dir string
}

func NewFileSessionStore(dir string) *FileSessionStore {
	return &FileSessionStore{
		dir: dir,
	}
}

func (store *FileSessionStore) Set(session *Session) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	if err := ensureDir(store.dir); err != nil {
		return err
	}

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return os.WriteFile(store.sessionPath(session.ID), data, 0600)
}

func (store *FileSessionStore) Get(id string) (*Session, error) {
	data, err := os.ReadFile(store.sessionPath(id))
	if err != nil {
		return nil, fmt.Errorf("session not found: %s", id)
	}

	var session Session

	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (store *FileSessionStore) Delete(id string) error {
	err := os.Remove(store.sessionPath(id))

	if err != nil {
		return fmt.Errorf("session not found: %s", id)
	}

	return nil
}

func (store *FileSessionStore) sessionPath(id string) string {
	return filepath.Join(store.dir, id+".json")
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0700)
}
