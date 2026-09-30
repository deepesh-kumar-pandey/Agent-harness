package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	if session.ID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	if err := ensureDir(store.dir); err != nil {
		return err
	}

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	tempFile, err := os.CreateTemp(store.dir, session.ID+".tmp-*")
	if err != nil {
		return err
	}

	tempPath := tempFile.Name()

	defer os.Remove(tempPath)

	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		return err
	}

	if err := tempFile.Close(); err != nil {
		return err
	}

	if err := os.Rename(tempPath, store.sessionPath(session.ID)); err != nil {
		return err
	}

	return nil
}

func (store *FileSessionStore) Get(id string) (*Session, error) {
	data, err := os.ReadFile(store.sessionPath(id))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("session not found: %s", id)
		}

		return nil, err
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

func (store *FileSessionStore) List() []*Session {
	entries, err := os.ReadDir(store.dir)
	if err != nil {
		return []*Session{}
	}

	sessions := make([]*Session, 0)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".json")

		session, err := store.Get(id)
		if err != nil {
			continue
		}

		sessions = append(sessions, session)
	}

	return sessions
}

func (store *FileSessionStore) sessionPath(id string) string {
	return filepath.Join(store.dir, id+".json")
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0700)
}
