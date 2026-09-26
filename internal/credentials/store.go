package credentials

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrCredentialNotFound is returned when the credentials file is missing
// or the requested provider has no stored credential.
var ErrCredentialNotFound = errors.New("credential not found")

type FileStore struct {
	path string
}

type credentialsData map[string]string

func NewFileStore() (*FileStore, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return nil, err
	}

	dir := filepath.Join(homeDir, ".agent-harness")

	return &FileStore{
		path: filepath.Join(dir, "credentials.json"),
	}, nil
}

func (s *FileStore) Set(provider string, key string) error {
	if provider == "" {
		return fmt.Errorf("provider is required")
	}

	if key == "" {
		return fmt.Errorf("credential key is required")
	}

	dir := filepath.Dir(s.path)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data := credentialsData{}

	file, err := os.ReadFile(s.path)
	if err == nil {
		if err := json.Unmarshal(file, &data); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	data[provider] = key

	file, err = json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, file, 0600)
}

func (s *FileStore) Get(provider string) (string, error) {
	if provider == "" {
		return "", fmt.Errorf("provider is required")
	}

	file, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%w for provider %q", ErrCredentialNotFound, provider)
		}
		return "", err
	}

	data := credentialsData{}

	if err := json.Unmarshal(file, &data); err != nil {
		return "", err
	}

	key, exists := data[provider]
	if !exists {
		return "", fmt.Errorf("%w for provider %q", ErrCredentialNotFound, provider)
	}

	return key, nil
}

func (s *FileStore) Delete(provider string) error {
	if provider == "" {
		return fmt.Errorf("provider is required")
	}

	file, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	data := credentialsData{}

	if err := json.Unmarshal(file, &data); err != nil {
		return err
	}

	if _, exists := data[provider]; !exists {
		return fmt.Errorf("credential not found for provider %q", provider)
	}

	delete(data, provider)

	file, err = json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, file, 0600)
}
