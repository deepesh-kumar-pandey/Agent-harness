package session

import (
	"os"
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
		{
			name: "returns error when database initialization fails",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			dir := t.TempDir()

			if testCase.name == "returns error when database initialization fails" {
				filePath := filepath.Join(dir, "database-file")

				if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
					t.Fatalf(
						"expected test file creation to succeed, got %v",
						err,
					)
				}

				dbPath := filepath.Join(filePath, "sessions.db")

				store, err := NewDatabaseSessionStore(dbPath)

				if err == nil {
					t.Fatal("expected error when database initialization fails, got nil")
				}

				if store != nil {
					t.Fatal("expected database store to be nil, got non-nil")
				}

				return
			}

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
			name: "updates existing session",
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
		{
			name: "returns error when database transaction fails",
		},
		{
			name: "returns error when message insertion fails",
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

			if testCase.name == "updates existing session" {
				firstSession := &Session{
					ID:        "test-session",
					CreatedAt: time.Now().Add(-2 * time.Hour),
					UpdatedAt: time.Now().Add(-2 * time.Hour),
					Metadata: map[string]string{
						"name": "Original Session",
					},
					Messages: []providerpkg.Message{
						{
							Role:    "user",
							Content: "Old message",
						},
						{
							Role:    "assistant",
							Content: "Old response",
						},
					},
				}

				if err := store.Set(firstSession); err != nil {
					t.Fatalf(
						"expected no error while setting first session, got %v",
						err,
					)
				}

				updatedSession := &Session{
					ID:        "test-session",
					CreatedAt: firstSession.CreatedAt,
					UpdatedAt: time.Now(),
					Metadata: map[string]string{
						"name": "Updated Session",
					},
					Messages: []providerpkg.Message{
						{
							Role:    "user",
							Content: "New message",
						},
					},
				}

				if err := store.Set(updatedSession); err != nil {
					t.Fatalf(
						"expected no error while updating session, got %v",
						err,
					)
				}

				var metadata string

				err := store.db.QueryRow(`
					SELECT metadata
					FROM sessions
					WHERE id = ?
				`, updatedSession.ID).Scan(&metadata)

				if err != nil {
					t.Fatalf(
						"expected session query to succeed, got %v",
						err,
					)
				}

				if metadata != `{"name":"Updated Session"}` {
					t.Fatalf(
						"expected updated metadata, got %s",
						metadata,
					)
				}

				var messageCount int

				err = store.db.QueryRow(`
					SELECT COUNT(*)
					FROM messages
					WHERE session_id = ?
				`, updatedSession.ID).Scan(&messageCount)

				if err != nil {
					t.Fatalf(
						"expected message count query to succeed, got %v",
						err,
					)
				}

				if messageCount != 1 {
					t.Fatalf(
						"expected old messages to be replaced with 1 new message, got %d",
						messageCount,
					)
				}

				var messageContent string

				err = store.db.QueryRow(`
					SELECT content
					FROM messages
					WHERE session_id = ?
					ORDER BY id
				`, updatedSession.ID).Scan(&messageContent)

				if err != nil {
					t.Fatalf(
						"expected message query to succeed, got %v",
						err,
					)
				}

				if messageContent != "New message" {
					t.Fatalf(
						"expected new message content, got %s",
						messageContent,
					)
				}

				return
			}

			if testCase.name == "returns error when database transaction fails" {
				session := &Session{
					ID:        "transaction-error-session",
					CreatedAt: time.Now().Add(-time.Hour),
					UpdatedAt: time.Now(),
					Metadata: map[string]string{
						"name": "Transaction Error Session",
					},
				}

				if err := store.db.Close(); err != nil {
					t.Fatalf(
						"expected database close to succeed, got %v",
						err,
					)
				}

				err := store.Set(session)

				if err == nil {
					t.Fatal("expected error when database transaction fails, got nil")
				}

				return
			}

			if testCase.name == "returns error when message insertion fails" {
				_, err := store.db.Exec(`
					CREATE TRIGGER fail_message_insert
					BEFORE INSERT ON messages
					WHEN NEW.session_id = 'message-error-session'
					BEGIN
						SELECT RAISE(ABORT, 'forced message insert failure');
					END;
				`)
				if err != nil {
					t.Fatalf(
						"expected trigger creation to succeed, got %v",
						err,
					)
				}

				session := &Session{
					ID:        "message-error-session",
					CreatedAt: time.Now().Add(-time.Hour),
					UpdatedAt: time.Now(),
					Metadata: map[string]string{
						"name": "Message Error Session",
					},
					Messages: []providerpkg.Message{
						{
							Role:    "user",
							Content: "test message",
						},
					},
				}

				err = store.Set(session)

				if err == nil {
					t.Fatal("expected error when message insertion fails, got nil")
				}

				var sessionCount int

				err = store.db.QueryRow(`
					SELECT COUNT(*)
					FROM sessions
					WHERE id = ?
				`, session.ID).Scan(&sessionCount)

				if err != nil {
					t.Fatalf(
						"expected session query to succeed, got %v",
						err,
					)
				}

				if sessionCount != 0 {
					t.Fatalf(
						"expected transaction rollback to remove session, got %d sessions",
						sessionCount,
					)
				}

				return
			}

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
		{
			name: "gets message with tool calls",
		},
		{
			name: "returns error when session does not exist",
		},
		{
			name: "returns error when metadata is invalid",
		},
		{
			name: "returns error when tool calls are invalid",
		},
		{
			name: "returns error when message query fails",
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

			if testCase.name == "returns error when session does not exist" {
				_, err := store.Get("missing-session")

				if err == nil {
					t.Fatal("expected error when getting missing session, got nil")
				}

				return
			}

			if testCase.name == "returns error when metadata is invalid" {
				_, err := store.db.Exec(`
					INSERT INTO sessions (
						id,
						created_at,
						updated_at,
						metadata
					)
					VALUES (?, ?, ?, ?)
				`,
					"invalid-session",
					time.Now().Add(-time.Hour),
					time.Now(),
					`invalid-json`,
				)

				if err != nil {
					t.Fatalf(
						"expected session insert to succeed, got %v",
						err,
					)
				}

				_, err = store.Get("invalid-session")

				if err == nil {
					t.Fatal("expected error for invalid metadata, got nil")
				}

				return
			}

			if testCase.name == "returns error when tool calls are invalid" {
				_, err := store.db.Exec(`
					INSERT INTO sessions (
						id,
						created_at,
						updated_at,
						metadata
					)
					VALUES (?, ?, ?, ?)
				`,
					"invalid-tool-calls-session",
					time.Now().Add(-time.Hour),
					time.Now(),
					`{}`,
				)

				if err != nil {
					t.Fatalf(
						"expected session insert to succeed, got %v",
						err,
					)
				}

				_, err = store.db.Exec(`
					INSERT INTO messages (
						session_id,
						role,
						content,
						tool_calls
					)
					VALUES (?, ?, ?, ?)
				`,
					"invalid-tool-calls-session",
					"assistant",
					"Hello",
					`invalid-json`,
				)

				if err != nil {
					t.Fatalf(
						"expected message insert to succeed, got %v",
						err,
					)
				}

				_, err = store.Get("invalid-tool-calls-session")

				if err == nil {
					t.Fatal("expected error for invalid tool calls, got nil")
				}

				return
			}

			if testCase.name == "returns error when message query fails" {
				session := &Session{
					ID:        "query-error-session",
					CreatedAt: time.Now().Add(-time.Hour),
					UpdatedAt: time.Now(),
					Metadata: map[string]string{
						"name": "Query Error Session",
					},
				}

				if err := store.Set(session); err != nil {
					t.Fatalf(
						"expected no error while setting session, got %v",
						err,
					)
				}

				if err := store.db.Close(); err != nil {
					t.Fatalf(
						"expected database close to succeed, got %v",
						err,
					)
				}

				_, err := store.Get(session.ID)

				if err == nil {
					t.Fatal("expected error when message query fails, got nil")
				}

				return
			}

			if testCase.name == "gets message with tool calls" {
				expectedSession := &Session{
					ID:        "tool-calls-session",
					CreatedAt: time.Now().Add(-time.Hour),
					UpdatedAt: time.Now(),
					Metadata: map[string]string{
						"name": "Tool Calls Session",
					},
					Messages: []providerpkg.Message{
						{
							Role:    "assistant",
							Content: "I will calculate that.",
							ToolCalls: []providerpkg.ToolCall{
								{
									Name: "calculator",
									Arguments: map[string]any{
										"operation": "add",
										"numbers":   []any{1.0, 2.0},
									},
								},
							},
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

				if len(actualSession.Messages) != 1 {
					t.Fatalf(
						"expected 1 message, got %d",
						len(actualSession.Messages),
					)
				}

				if len(actualSession.Messages[0].ToolCalls) != 1 {
					t.Fatalf(
						"expected 1 tool call, got %d",
						len(actualSession.Messages[0].ToolCalls),
					)
				}

				toolCall := actualSession.Messages[0].ToolCalls[0]

				if toolCall.Name != "calculator" {
					t.Fatalf(
						"expected tool name %q, got %q",
						"calculator",
						toolCall.Name,
					)
				}

				if toolCall.Arguments["operation"] != "add" {
					t.Fatalf(
						"expected operation %q, got %v",
						"add",
						toolCall.Arguments["operation"],
					)
				}

				numbers, ok := toolCall.Arguments["numbers"].([]any)
				if !ok {
					t.Fatalf(
						"expected numbers to be []any, got %T",
						toolCall.Arguments["numbers"],
					)
				}

				if len(numbers) != 2 {
					t.Fatalf(
						"expected 2 numbers, got %d",
						len(numbers),
					)
				}

				if numbers[0] != 1.0 || numbers[1] != 2.0 {
					t.Fatalf(
						"expected numbers [1 2], got %v",
						numbers,
					)
				}

				return
			}

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
		{
			name: "returns error when session does not exist",
		},
		{
			name: "returns error when database query fails",
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

			if testCase.name == "returns error when session does not exist" {
				err := store.Delete("missing-session")

				if err == nil {
					t.Fatal("expected error when deleting missing session, got nil")
				}

				return
			}

			if testCase.name == "returns error when database query fails" {
				if err := store.db.Close(); err != nil {
					t.Fatalf(
						"expected database close to succeed, got %v",
						err,
					)
				}

				err := store.Delete("test-session")

				if err == nil {
					t.Fatal("expected error when database query fails, got nil")
				}

				return
			}

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
		{
			name: "returns empty list when no sessions exist",
		},
		{
			name: "returns empty list when database query fails",
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

			if testCase.name == "returns empty list when no sessions exist" {
				sessions := store.List()

				if len(sessions) != 0 {
					t.Fatalf(
						"expected empty session list, got %d sessions",
						len(sessions),
					)
				}

				return
			}

			if testCase.name == "returns empty list when database query fails" {
				if err := store.db.Close(); err != nil {
					t.Fatalf(
						"expected database close to succeed, got %v",
						err,
					)
				}

				sessions := store.List()

				if len(sessions) != 0 {
					t.Fatalf(
						"expected empty list, got %d sessions",
						len(sessions),
					)
				}

				return
			}

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
