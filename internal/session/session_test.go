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
		})
	}
}
