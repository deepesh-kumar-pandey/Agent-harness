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

// Test Set and Get preserve session messages, timestamps, and metadata.
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

			session.Rename("Renamed Session")
			session.Metadata["description"] = "Session metadata test"

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

			if !loadedSession.CreatedAt.Equal(session.CreatedAt) {
				t.Errorf(
					"expected CreatedAt %v, got %v",
					session.CreatedAt,
					loadedSession.CreatedAt,
				)
			}

			if !loadedSession.UpdatedAt.Equal(session.UpdatedAt) {
				t.Errorf(
					"expected UpdatedAt %v, got %v",
					session.UpdatedAt,
					loadedSession.UpdatedAt,
				)
			}

			if loadedSession.Metadata["name"] != "Renamed Session" {
				t.Errorf(
					"expected session name %q, got %q",
					"Renamed Session",
					loadedSession.Metadata["name"],
				)
			}

			if loadedSession.Metadata["description"] != "Session metadata test" {
				t.Errorf(
					"expected metadata description %q, got %q",
					"Session metadata test",
					loadedSession.Metadata["description"],
				)
			}

			fmt.Printf(
				"Session messages, timestamps, and metadata stored and restored successfully: %s\n",
				testCase.name,
			)
		})
	}
}

// Test Get returns an error for a missing session.
func TestFileSessionStoreGetMissing(t *testing.T) {
	store := NewFileSessionStore(t.TempDir())

	session, err := store.Get("missing-session")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if session != nil {
		t.Fatal("expected nil session, got session")
	}
}

// Test Delete removes an existing session and rejects missing sessions.
func TestFileSessionStoreDelete(t *testing.T) {
	testCases := []struct {
		name        string
		addSession  bool
		expectError bool
	}{
		{
			name:       "deletes existing session",
			addSession: true,
		},
		{
			name:        "returns error for missing session",
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			dir := t.TempDir()
			store := NewFileSessionStore(dir)

			if testCase.addSession {
				if err := store.Set(NewSession("test-session")); err != nil {
					t.Fatalf(
						"expected no error while setting session, got %v",
						err,
					)
				}
			}

			err := store.Delete("test-session")

			if testCase.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			_, err = store.Get("test-session")

			if err == nil {
				t.Fatal("expected session to be deleted")
			}
		})
	}
}

// Test List returns all stored sessions.
func TestFileSessionStoreList(t *testing.T) {
	dir := t.TempDir()
	store := NewFileSessionStore(dir)

	session1 := NewSession("session-1")
	session2 := NewSession("session-2")

	if err := store.Set(session1); err != nil {
		t.Fatalf("expected no error while setting session, got %v", err)
	}

	if err := store.Set(session2); err != nil {
		t.Fatalf("expected no error while setting session, got %v", err)
	}

	sessions := store.List()

	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}

	found := make(map[string]bool)

	for _, session := range sessions {
		found[session.ID] = true
	}

	if !found["session-1"] {
		t.Fatal("expected session-1 in list")
	}

	if !found["session-2"] {
		t.Fatal("expected session-2 in list")
	}
}

// Test List returns an empty list for an empty directory.
func TestFileSessionStoreListEmpty(t *testing.T) {
	store := NewFileSessionStore(t.TempDir())

	sessions := store.List()

	if sessions == nil {
		t.Fatal("expected empty session list, got nil")
	}

	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(sessions))
	}
}
