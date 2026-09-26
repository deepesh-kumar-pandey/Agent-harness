package session

import (
	"testing"
)

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
