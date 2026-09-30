package session

import (
	"path/filepath"
	"testing"
	"time"

	providerpkg "agent-harness/internal/provider"
)

// TestNewDatabaseSessionStore tests database store creation and table initialization.
func TestNewDatabaseSessionStore(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "creates database store",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			dir := t.TempDir()
			dbPath := filepath.Join(dir, "sessions.db")

			store, err := NewDatabaseSessionStore(dbPath)
			if err != nil {
				t.Fatalf(
					"expected no error while creating database store, got %v",
					err,
				)
			}

			if store == nil {
				t.Fatal("expected database store, got nil")
			}

			if store.db == nil {
				t.Fatal("expected database connection, got nil")
			}

			var tableCount int

			err = store.db.QueryRow(`
				SELECT COUNT(*)
				FROM sqlite_master
				WHERE type = 'table'
				AND name IN ('sessions', 'messages')
			`).Scan(&tableCount)

			if err != nil {
				t.Fatalf(
					"expected tables query to succeed, got %v",
					err,
				)
			}

			if tableCount != 2 {
				t.Fatalf(
					"expected 2 tables, got %d",
					tableCount,
				)
			}

			if err := store.db.Close(); err != nil {
				t.Fatalf(
					"expected database to close successfully, got %v",
					err,
				)
			}
		})
	}
}

// TestDatabaseSessionStoreSet tests storing a session and rejecting invalid sessions.
func TestDatabaseSessionStoreSet(t *testing.T) {
	testCases := []struct {
		name        string
		session     *Session
		expectError bool
	}{
		{
			name: "stores session and messages",
			session: &Session{
				ID:        "test-session",
				CreatedAt: time.Now().Add(-time.Hour),
				UpdatedAt: time.Now(),
				Metadata: map[string]string{
					"name": "Test Session",
				},
				Messages: []providerpkg.Message{
					{
						Role:    "user",
						Content: "Hello",
					},
					{
						Role:    "assistant",
						Content: "Hello! How can I help you?",
					},
				},
			},
		},
		{
			name:        "rejects nil session",
			session:     nil,
			expectError: true,
		},
		{
			name: "rejects empty session ID",
			session: &Session{
				ID: "",
			},
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			dir := t.TempDir()
			dbPath := filepath.Join(dir, "sessions.db")

			store, err := NewDatabaseSessionStore(dbPath)
			if err != nil {
				t.Fatalf(
					"expected no error while creating database store, got %v",
					err,
				)
			}

			defer store.db.Close()

			err = store.Set(testCase.session)

			if testCase.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf(
					"expected no error while setting session, got %v",
					err,
				)
			}

			var sessionID string
			var metadata string

			err = store.db.QueryRow(`
				SELECT id, metadata
				FROM sessions
				WHERE id = ?
			`, testCase.session.ID).Scan(&sessionID, &metadata)

			if err != nil {
				t.Fatalf(
					"expected session query to succeed, got %v",
					err,
				)
			}

			if sessionID != testCase.session.ID {
				t.Fatalf(
					"expected session ID %s, got %s",
					testCase.session.ID,
					sessionID,
				)
			}

			if metadata != `{"name":"Test Session"}` {
				t.Fatalf(
					"expected metadata to be stored, got %s",
					metadata,
				)
			}

			var messageCount int

			err = store.db.QueryRow(`
				SELECT COUNT(*)
				FROM messages
				WHERE session_id = ?
			`, testCase.session.ID).Scan(&messageCount)

			if err != nil {
				t.Fatalf(
					"expected message query to succeed, got %v",
					err,
				)
			}

			if messageCount != 2 {
				t.Fatalf(
					"expected 2 messages, got %d",
					messageCount,
				)
			}
		})
	}
}

// TestDatabaseSessionStoreGet tests retrieving a session, metadata, and messages from the database.
func TestDatabaseSessionStoreGet(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "gets session and messages",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			dir := t.TempDir()
			dbPath := filepath.Join(dir, "sessions.db")

			store, err := NewDatabaseSessionStore(dbPath)
			if err != nil {
				t.Fatalf(
					"expected no error while creating database store, got %v",
					err,
				)
			}

			defer store.db.Close()

			createdAt := time.Now().Add(-time.Hour)
			updatedAt := time.Now()

			expectedSession := &Session{
				ID:        "test-session",
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				Metadata: map[string]string{
					"name": "Test Session",
				},
				Messages: []providerpkg.Message{
					{
						Role:    "user",
						Content: "Hello",
					},
					{
						Role:    "assistant",
						Content: "Hello! How can I help you?",
					},
				},
			}

			if err := store.Set(expectedSession); err != nil {
				t.Fatalf(
					"expected no error while setting session, got %v",
					err,
				)
			}

			actualSession, err := store.Get(expectedSession.ID)
			if err != nil {
				t.Fatalf(
					"expected no error while getting session, got %v",
					err,
				)
			}

			if actualSession == nil {
				t.Fatal("expected session, got nil")
			}

			if actualSession.ID != expectedSession.ID {
				t.Fatalf(
					"expected session ID %s, got %s",
					expectedSession.ID,
					actualSession.ID,
				)
			}

			if actualSession.Metadata["name"] != expectedSession.Metadata["name"] {
				t.Fatalf(
					"expected session name %s, got %s",
					expectedSession.Metadata["name"],
					actualSession.Metadata["name"],
				)
			}

			if len(actualSession.Messages) != len(expectedSession.Messages) {
				t.Fatalf(
					"expected %d messages, got %d",
					len(expectedSession.Messages),
					len(actualSession.Messages),
				)
			}

			for index, expectedMessage := range expectedSession.Messages {
				actualMessage := actualSession.Messages[index]

				if actualMessage.Role != expectedMessage.Role {
					t.Fatalf(
						"expected message role %s, got %s",
						expectedMessage.Role,
						actualMessage.Role,
					)
				}

				if actualMessage.Content != expectedMessage.Content {
					t.Fatalf(
						"expected message content %s, got %s",
						expectedMessage.Content,
						actualMessage.Content,
					)
				}
			}
		})
	}
}

// TestDatabaseSessionStoreDelete tests deleting a session and its messages.
func TestDatabaseSessionStoreDelete(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "deletes session and messages",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			dir := t.TempDir()
			dbPath := filepath.Join(dir, "sessions.db")

			store, err := NewDatabaseSessionStore(dbPath)
			if err != nil {
				t.Fatalf(
					"expected no error while creating database store, got %v",
					err,
				)
			}

			defer store.db.Close()

			session := &Session{
				ID:        "test-session",
				CreatedAt: time.Now().Add(-time.Hour),
				UpdatedAt: time.Now(),
				Metadata: map[string]string{
					"name": "Test Session",
				},
				Messages: []providerpkg.Message{
					{
						Role:    "user",
						Content: "Hello",
					},
					{
						Role:    "assistant",
						Content: "Hello! How can I help you?",
					},
				},
			}

			if err := store.Set(session); err != nil {
				t.Fatalf(
					"expected no error while setting session, got %v",
					err,
				)
			}

			if err := store.Delete(session.ID); err != nil {
				t.Fatalf(
					"expected no error while deleting session, got %v",
					err,
				)
			}

			_, err = store.Get(session.ID)
			if err == nil {
				t.Fatal("expected session to be deleted")
			}

			var messageCount int

			err = store.db.QueryRow(`
				SELECT COUNT(*)
				FROM messages
				WHERE session_id = ?
			`, session.ID).Scan(&messageCount)

			if err != nil {
				t.Fatalf(
					"expected message query to succeed, got %v",
					err,
				)
			}

			if messageCount != 0 {
				t.Fatalf(
					"expected 0 messages after deleting session, got %d",
					messageCount,
				)
			}
		})
	}
}

// TestDatabaseSessionStoreList tests listing all sessions from the database.
func TestDatabaseSessionStoreList(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "lists all sessions",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			dir := t.TempDir()
			dbPath := filepath.Join(dir, "sessions.db")

			store, err := NewDatabaseSessionStore(dbPath)
			if err != nil {
				t.Fatalf(
					"expected no error while creating database store, got %v",
					err,
				)
			}

			defer store.db.Close()

			firstSession := &Session{
				ID:        "first-session",
				CreatedAt: time.Now().Add(-2 * time.Hour),
				UpdatedAt: time.Now().Add(-2 * time.Hour),
				Metadata: map[string]string{
					"name": "First Session",
				},
				Messages: []providerpkg.Message{
					{
						Role:    "user",
						Content: "First message",
					},
				},
			}

			secondSession := &Session{
				ID:        "second-session",
				CreatedAt: time.Now().Add(-time.Hour),
				UpdatedAt: time.Now().Add(-time.Hour),
				Metadata: map[string]string{
					"name": "Second Session",
				},
				Messages: []providerpkg.Message{
					{
						Role:    "user",
						Content: "Second message",
					},
				},
			}

			if err := store.Set(firstSession); err != nil {
				t.Fatalf(
					"expected no error while setting first session, got %v",
					err,
				)
			}

			if err := store.Set(secondSession); err != nil {
				t.Fatalf(
					"expected no error while setting second session, got %v",
					err,
				)
			}

			sessions := store.List()

			if len(sessions) != 2 {
				t.Fatalf(
					"expected 2 sessions, got %d",
					len(sessions),
				)
			}

			if sessions[0].ID != firstSession.ID {
				t.Fatalf(
					"expected first session ID %s, got %s",
					firstSession.ID,
					sessions[0].ID,
				)
			}

			if sessions[1].ID != secondSession.ID {
				t.Fatalf(
					"expected second session ID %s, got %s",
					secondSession.ID,
					sessions[1].ID,
				)
			}

			if len(sessions[0].Messages) != 1 {
				t.Fatalf(
					"expected first session to have 1 message, got %d",
					len(sessions[0].Messages),
				)
			}

			if len(sessions[1].Messages) != 1 {
				t.Fatalf(
					"expected second session to have 1 message, got %d",
					len(sessions[1].Messages),
				)
			}
		})
	}
}
