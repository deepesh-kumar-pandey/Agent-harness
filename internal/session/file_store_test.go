package session

import (
	"fmt"
	"os"
	"testing"
)

// Test constructor creates a file session store.
func TestNewFileSessionStore(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "creates file session store",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fmt.Printf("Running test case: %s\n", testCase.name)

			store := NewFileSessionStore(t.TempDir())

			if store == nil {
				t.Fatal("expected file session store, got nil")
			}

			if store.dir == "" {
				t.Fatal("expected store directory, got empty")
			}

			fmt.Printf("File session store created successfully: %s\n", testCase.name)
		})
	}
}

// Test Set stores a session and rejects nil sessions.
func TestFileSessionStoreSet(t *testing.T) {
	testCases := []struct {
		name        string
		session     *Session
		expectError bool
	}{
		{
			name:        "sets session",
			session:     &Session{ID: "test-session"},
			expectError: false,
		},
		{
			name:        "rejects nil session",
			session:     nil,
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fmt.Printf("Running test case: %s\n", testCase.name)

			dir := t.TempDir()
			store := NewFileSessionStore(dir)

			err := store.Set(testCase.session)

			if testCase.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				fmt.Printf("Error correctly returned: %s\n", testCase.name)
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			sessionPath := store.sessionPath(testCase.session.ID)

			if _, err := os.Stat(sessionPath); err != nil {
				t.Fatalf("expected session file to exist, got %v", err)
			}

			fmt.Printf("Session stored successfully: %s\n", testCase.name)
		})
	}
}
