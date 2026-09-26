package session

import (
	"encoding/json"
	"testing"
	"time"

	providerpkg "agent-harness/internal/provider"
)

// TestNewSession verifies a new session is initialized correctly.
func TestNewSession(t *testing.T) {
	testCases := []struct {
		name string
		id   string
	}{
		{
			name: "creates session",
			id:   "test-session",
		},
		{
			name: "creates another session",
			id:   "session-123",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			session := NewSession(testCase.id)

			if session == nil {
				t.Fatal("expected session, got nil")
			}

			if session.ID != testCase.id {
				t.Errorf("expected ID %q, got %q", testCase.id, session.ID)
			}

			if session.CreatedAt.IsZero() {
				t.Fatal("expected CreatedAt to be set")
			}

			if session.UpdatedAt.IsZero() {
				t.Fatal("expected UpdatedAt to be set")
			}

			if !session.CreatedAt.Equal(session.UpdatedAt) {
				t.Fatal("expected CreatedAt and UpdatedAt to match")
			}

			if session.Metadata == nil {
				t.Fatal("expected metadata, got nil")
			}

			if len(session.Metadata) != 0 {
				t.Errorf(
					"expected empty metadata, got %d",
					len(session.Metadata),
				)
			}

			if session.Messages == nil {
				t.Fatal("expected messages, got nil")
			}

			if len(session.Messages) != 0 {
				t.Errorf(
					"expected empty messages, got %d",
					len(session.Messages),
				)
			}
		})
	}
}

// TestSessionRename verifies renaming updates metadata and the timestamp.
func TestSessionRename(t *testing.T) {
	session := NewSession("test-session")
	previousUpdatedAt := session.UpdatedAt

	time.Sleep(time.Millisecond)

	session.Rename("My Session")

	if session.Metadata["name"] != "My Session" {
		t.Errorf(
			"expected session name %q, got %q",
			"My Session",
			session.Metadata["name"],
		)
	}

	if !session.UpdatedAt.After(previousUpdatedAt) {
		t.Fatal("expected UpdatedAt to be updated")
	}
}

// TestSessionExport verifies a session can be exported as JSON.
func TestSessionExport(t *testing.T) {
	session := NewSession("test-session")

	session.Metadata["name"] = "My Session"
	session.Metadata["description"] = "Export test"

	session.Messages = append(
		session.Messages,
		providerpkg.Message{
			Role:    "user",
			Content: "Hello",
		},
	)

	data, err := session.Export()
	if err != nil {
		t.Fatalf("expected no error while exporting session, got %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected exported data, got empty")
	}

	var exported Session

	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("expected valid JSON, got %v", err)
	}

	if exported.ID != "test-session" {
		t.Errorf(
			"expected exported ID %q, got %q",
			"test-session",
			exported.ID,
		)
	}

	if exported.Metadata["name"] != "My Session" {
		t.Errorf(
			"expected exported metadata name %q, got %q",
			"My Session",
			exported.Metadata["name"],
		)
	}

	if exported.Metadata["description"] != "Export test" {
		t.Errorf(
			"expected exported metadata description %q, got %q",
			"Export test",
			exported.Metadata["description"],
		)
	}

	if len(exported.Messages) != 1 {
		t.Fatalf(
			"expected 1 exported message, got %d",
			len(exported.Messages),
		)
	}

	if exported.Messages[0].Content != "Hello" {
		t.Errorf(
			"expected exported message content %q, got %q",
			"Hello",
			exported.Messages[0].Content,
		)
	}
}

// TestSessionImport verifies a session can be restored from JSON.
func TestSessionImport(t *testing.T) {
	session := NewSession("test-session")

	session.Metadata["name"] = "My Session"
	session.Metadata["description"] = "Import test"

	session.Messages = append(
		session.Messages,
		providerpkg.Message{
			Role:    "user",
			Content: "Hello",
		},
	)

	data, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("expected no error while creating session JSON, got %v", err)
	}

	imported, err := Import(data)
	if err != nil {
		t.Fatalf("expected no error while importing session, got %v", err)
	}

	if imported == nil {
		t.Fatal("expected imported session, got nil")
	}

	if imported.ID != session.ID {
		t.Errorf(
			"expected imported ID %q, got %q",
			session.ID,
			imported.ID,
		)
	}

	if !imported.CreatedAt.Equal(session.CreatedAt) {
		t.Errorf(
			"expected imported CreatedAt %v, got %v",
			session.CreatedAt,
			imported.CreatedAt,
		)
	}

	if !imported.UpdatedAt.Equal(session.UpdatedAt) {
		t.Errorf(
			"expected imported UpdatedAt %v, got %v",
			session.UpdatedAt,
			imported.UpdatedAt,
		)
	}

	if imported.Metadata["name"] != "My Session" {
		t.Errorf(
			"expected imported metadata name %q, got %q",
			"My Session",
			imported.Metadata["name"],
		)
	}

	if imported.Metadata["description"] != "Import test" {
		t.Errorf(
			"expected imported metadata description %q, got %q",
			"Import test",
			imported.Metadata["description"],
		)
	}

	if len(imported.Messages) != 1 {
		t.Fatalf(
			"expected 1 imported message, got %d",
			len(imported.Messages),
		)
	}

	if imported.Messages[0].Role != "user" {
		t.Errorf(
			"expected imported message role %q, got %q",
			"user",
			imported.Messages[0].Role,
		)
	}

	if imported.Messages[0].Content != "Hello" {
		t.Errorf(
			"expected imported message content %q, got %q",
			"Hello",
			imported.Messages[0].Content,
		)
	}
}
