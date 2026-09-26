package credentials

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewFileStore tests the creation of a file-based credential store.
func TestNewFileStore(t *testing.T) {
	store, err := NewFileStore()
	if err != nil {
		t.Fatalf("NewFileStore() returned error: %v", err)
	}

	if store == nil {
		t.Fatal("expected file store, got nil")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("expected home directory, got error: %v", err)
	}

	expectedPath := filepath.Join(
		homeDir,
		".agent-harness",
		"credentials.json",
	)

	if store.path != expectedPath {
		t.Fatalf(
			"expected store path %q, got %q",
			expectedPath,
			store.path,
		)
	}
}

// TestFileStoreSet tests storing credentials in the file store.
func TestFileStoreSet(t *testing.T) {
	testCases := []struct {
		name        string
		provider    string
		key         string
		expectError bool
	}{
		{
			name:        "sets credential",
			provider:    "openai",
			key:         "test-key",
			expectError: false,
		},
		{
			name:        "empty provider",
			provider:    "",
			key:         "test-key",
			expectError: true,
		},
		{
			name:        "empty key",
			provider:    "openai",
			key:         "",
			expectError: true,
		},
		{
			name:        "updates existing credential",
			provider:    "openai",
			key:         "updated-key",
			expectError: false,
		},
		{
			name:        "preserves multiple providers",
			provider:    "anthropic",
			key:         "anthropic-key",
			expectError: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tempDir := t.TempDir()

			store := &FileStore{
				path: filepath.Join(tempDir, "credentials.json"),
			}

			if testCase.name == "updates existing credential" {
				err := store.Set("openai", "old-key")
				if err != nil {
					t.Fatalf("failed to set initial credential: %v", err)
				}
			}

			if testCase.name == "preserves multiple providers" {
				err := store.Set("openai", "openai-key")
				if err != nil {
					t.Fatalf("failed to set OpenAI credential: %v", err)
				}
			}

			err := store.Set(testCase.provider, testCase.key)

			if (err != nil) != testCase.expectError {
				t.Fatalf(
					"expected error: %v, got error: %v",
					testCase.expectError,
					err != nil,
				)
			}

			if err != nil {
				return
			}

			file, err := os.ReadFile(store.path)
			if err != nil {
				t.Fatalf(
					"expected credentials file, got error: %v",
					err,
				)
			}

			if !strings.Contains(string(file), testCase.provider) {
				t.Fatalf(
					"expected provider %q in credentials file",
					testCase.provider,
				)
			}

			if testCase.name == "updates existing credential" {
				var data map[string]string

				err := json.Unmarshal(file, &data)
				if err != nil {
					t.Fatalf(
						"expected valid credentials JSON, got error: %v",
						err,
					)
				}

				if data["openai"] != "updated-key" {
					t.Fatalf(
						"expected updated key %q, got %q",
						"updated-key",
						data["openai"],
					)
				}
			}

			if testCase.name == "preserves multiple providers" {
				var data map[string]string

				err := json.Unmarshal(file, &data)
				if err != nil {
					t.Fatalf(
						"expected valid credentials JSON, got error: %v",
						err,
					)
				}

				if data["openai"] != "openai-key" {
					t.Fatalf(
						"expected OpenAI key %q, got %q",
						"openai-key",
						data["openai"],
					)
				}

				if data["anthropic"] != "anthropic-key" {
					t.Fatalf(
						"expected Anthropic key %q, got %q",
						"anthropic-key",
						data["anthropic"],
					)
				}
			}
		})
	}
}

// TestFileStoreGet tests retrieving credentials from the file store.
func TestFileStoreGet(t *testing.T) {
	testCases := []struct {
		name              string
		provider          string
		expectKey         string
		expectError       bool
		expectErrNotFound bool
		missingFile       bool
		malformedFile     bool
	}{
		{
			name:        "gets credential",
			provider:    "openai",
			expectKey:   "test-key",
			expectError: false,
		},
		{
			name:              "provider not found",
			provider:          "anthropic",
			expectKey:         "",
			expectError:       true,
			expectErrNotFound: true,
		},
		{
			name:        "empty provider",
			provider:    "",
			expectKey:   "",
			expectError: true,
		},
		{
			name:              "missing credentials file",
			provider:          "openai",
			expectKey:         "",
			expectError:       true,
			expectErrNotFound: true,
			missingFile:       true,
		},
		{
			name:              "malformed JSON preserves parsing error",
			provider:          "openai",
			expectKey:         "",
			expectError:       true,
			expectErrNotFound: false,
			malformedFile:     true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tempDir := t.TempDir()

			store := &FileStore{
				path: filepath.Join(tempDir, "credentials.json"),
			}

			switch {
			case testCase.missingFile:
				// Leave the file absent.

			case testCase.malformedFile:
				if err := os.WriteFile(
					store.path,
					[]byte("{invalid"),
					0600,
				); err != nil {
					t.Fatalf(
						"failed to write malformed JSON: %v",
						err,
					)
				}

			default:
				if err := store.Set("openai", "test-key"); err != nil {
					t.Fatalf(
						"failed to set test credential: %v",
						err,
					)
				}
			}

			key, err := store.Get(testCase.provider)

			if (err != nil) != testCase.expectError {
				t.Fatalf(
					"expected error: %v, got error: %v",
					testCase.expectError,
					err != nil,
				)
			}

			if testCase.expectError {
				isNotFound := errors.Is(err, ErrCredentialNotFound)

				if isNotFound != testCase.expectErrNotFound {
					t.Fatalf(
						"ErrCredentialNotFound: expected %v, got %v (error: %v)",
						testCase.expectErrNotFound,
						isNotFound,
						err,
					)
				}
			}

			if key != testCase.expectKey {
				t.Fatalf(
					"expected key %q, got %q",
					testCase.expectKey,
					key,
				)
			}
		})
	}
}

// TestFileStoreDelete tests deleting credentials from the file store.
func TestFileStoreDelete(t *testing.T) {
	testCases := []struct {
		name        string
		provider    string
		expectError bool
	}{
		{
			name:        "deletes credential",
			provider:    "openai",
			expectError: false,
		},
		{
			name:        "provider not found",
			provider:    "anthropic",
			expectError: true,
		},
		{
			name:        "empty provider",
			provider:    "",
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tempDir := t.TempDir()

			store := &FileStore{
				path: filepath.Join(tempDir, "credentials.json"),
			}

			err := store.Set("openai", "test-key")
			if err != nil {
				t.Fatalf(
					"failed to set test credential: %v",
					err,
				)
			}

			err = store.Delete(testCase.provider)

			if (err != nil) != testCase.expectError {
				t.Fatalf(
					"expected error: %v, got error: %v",
					testCase.expectError,
					err != nil,
				)
			}

			if err != nil {
				return
			}

			_, err = store.Get(testCase.provider)
			if err == nil {
				t.Fatalf(
					"expected credential for provider %q to be deleted",
					testCase.provider,
				)
			}
		})
	}
}
