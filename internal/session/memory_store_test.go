package session

import (
	"testing"
)

// Test for Get function
func TestMemoryStoreGet(t *testing.T) {
	testCases := []struct {
		name        string
		id          string
		addSession  bool
		expectError bool
	}{
		{
			name:       "gets existing session",
			id:         "test-session",
			addSession: true,
		},
		{
			name:        "returns error for missing session",
			id:          "missing-session",
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := NewMemoryStore()

			if testCase.addSession {
				store.sessions[testCase.id] = NewSession(testCase.id)
			}

			session, err := store.Get(testCase.id)

			if testCase.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if session != nil {
					t.Fatal("expected nil session, got session")
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if session == nil {
				t.Fatal("expected session, got nil")
			}

			if session.ID != testCase.id {
				t.Errorf("expected ID %q, got %q", testCase.id, session.ID)
			}
		})
	}
}

// Test for Set function
func TestMemoryStoreSet(t *testing.T) {
	testCases := []struct {
		name string
		id   string
	}{
		{
			name: "sets session",
			id:   "test-session",
		},
		{
			name: "sets another session",
			id:   "session-123",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := NewMemoryStore()
			session := NewSession(testCase.id)

			err := store.Set(session)

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			storedSession, err := store.Get(testCase.id)

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if storedSession != session {
				t.Fatal("expected stored session to match original session")
			}
		})
	}
}

// Unit test for Delete
func TestMemoryStoreDelete(t *testing.T) {
	testCases := []struct {
		name        string
		id          string
		addSession  bool
		expectError bool
	}{
		{
			name:       "deletes existing session",
			id:         "test-session",
			addSession: true,
		},
		{
			name:        "returns error for missing session",
			id:          "missing-session",
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := NewMemoryStore()

			if testCase.addSession {
				store.sessions[testCase.id] = NewSession(testCase.id)
			}

			err := store.Delete(testCase.id)

			if testCase.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			_, err = store.Get(testCase.id)

			if err == nil {
				t.Fatal("expected session to be deleted")
			}
		})
	}
}

// Test for List function
func TestMemoryStoreList(t *testing.T) {
	store := NewMemoryStore()

	session1 := NewSession("session-1")
	session2 := NewSession("session-2")

	store.sessions[session1.ID] = session1
	store.sessions[session2.ID] = session2

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
