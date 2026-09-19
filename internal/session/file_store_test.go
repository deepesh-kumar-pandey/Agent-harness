package session

import (
	"fmt"
	"os"
	"testing"

	providerpkg "agent-harness/internal/provider"
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

			fmt.Printf(
				"File session store created successfully: %s\n",
				testCase.name,
			)
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

				fmt.Printf(
					"Error correctly returned: %s\n",
					testCase.name,
				)
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			sessionPath := store.sessionPath(testCase.session.ID)

			if _, err := os.Stat(sessionPath); err != nil {
				t.Fatalf(
					"expected session file to exist, got %v",
					err,
				)
			}

			fmt.Printf(
				"Session stored successfully: %s\n",
				testCase.name,
			)
		})
	}
}

// Test Set and Get preserve session messages.
func TestFileSessionStoreSetAndGetMessages(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "stores and restores messages",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fmt.Printf("Running test case: %s\n", testCase.name)

			dir := t.TempDir()
			store := NewFileSessionStore(dir)

			session := NewSession("test-session")

			session.Messages = append(
				session.Messages,
				providerpkg.Message{
					Role:    "user",
					Content: "Hello",
				},
				providerpkg.Message{
					Role:    "assistant",
					Content: "Hi there",
				},
			)

			if err := store.Set(session); err != nil {
				t.Fatalf(
					"expected no error while setting session, got %v",
					err,
				)
			}

			loadedSession, err := store.Get("test-session")
			if err != nil {
				t.Fatalf(
					"expected no error while getting session, got %v",
					err,
				)
			}

			if len(loadedSession.Messages) != 2 {
				t.Fatalf(
					"expected 2 messages, got %d",
					len(loadedSession.Messages),
				)
			}

			if loadedSession.Messages[0].Content != "Hello" {
				t.Errorf(
					"expected first message content %q, got %q",
					"Hello",
					loadedSession.Messages[0].Content,
				)
			}

			if loadedSession.Messages[1].Content != "Hi there" {
				t.Errorf(
					"expected second message content %q, got %q",
					"Hi there",
					loadedSession.Messages[1].Content,
				)
			}

			fmt.Printf(
				"Session messages stored and restored successfully: %s\n",
				testCase.name,
			)
		})
	}
}
