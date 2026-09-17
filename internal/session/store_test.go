package session

import (
	"fmt"
	"testing"
)

// Test for New Session Creation
func TestNewSessionStore(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "creates session store",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fmt.Printf("Running test case: %s\n", testCase.name)

			store := NewSessionStore()

			if store == nil {
				t.Fatal("expected session store, got nil")
			}

			if store.sessions == nil {
				t.Fatal("expected sessions map, got nil")
			}

			fmt.Printf("Session store created successfully: %s\n", testCase.name)
		})
	}
}

// Test for Set function
func TestSessionStoreSet(t *testing.T) {
	testCases := []struct {
		name        string
		session     *Session
		expectError bool
	}{
		{
			name: "sets session",
			session: &Session{
				ID: "test-session",
			},
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

			store := NewSessionStore()

			err := store.Set(testCase.session)

			if testCase.expectError && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !testCase.expectError {
				storedSession, exists := store.sessions[testCase.session.ID]

				if !exists {
					t.Fatal("expected session to be stored")
				}

				if storedSession != testCase.session {
					t.Fatal("expected stored session to match input session")
				}
			}

			fmt.Printf("Set test completed successfully: %s\n", testCase.name)
		})
	}
}

// Test for Get function
func TestSessionStoreGet(t *testing.T) {
	testCases := []struct {
		name        string
		id          string
		expectError bool
	}{
		{
			name:        "gets existing session",
			id:          "test-session",
			expectError: false,
		},
		{
			name:        "returns error for missing session",
			id:          "missing-session",
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fmt.Printf("Running test case: %s\n", testCase.name)

			store := NewSessionStore()
			session := NewSession(testCase.id)

			if !testCase.expectError {
				err := store.Set(session)

				if err != nil {
					t.Fatalf("expected no error while setting session, got %v", err)
				}
			}

			result, err := store.Get(testCase.id)

			if testCase.expectError && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !testCase.expectError && result != session {
				t.Fatal("expected returned session to match stored session")
			}

			fmt.Printf("Get test completed successfully: %s\n", testCase.name)
		})
	}
}

// Test for Delete function
func TestSessionStoreDelete(t *testing.T) {
	testCases := []struct {
		name        string
		id          string
		expectError bool
	}{
		{
			name:        "deletes existing session",
			id:          "test-session",
			expectError: false,
		},
		{
			name:        "returns error for missing session",
			id:          "missing-session",
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fmt.Printf("Running test case: %s\n", testCase.name)

			store := NewSessionStore()
			session := NewSession(testCase.id)

			if !testCase.expectError {
				err := store.Set(session)

				if err != nil {
					t.Fatalf("expected no error while setting session, got %v", err)
				}
			}

			err := store.Delete(testCase.id)

			if testCase.expectError && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !testCase.expectError {
				_, exists := store.sessions[testCase.id]

				if exists {
					t.Fatal("expected session to be deleted")
				}
			}

			fmt.Printf("Delete test completed successfully: %s\n", testCase.name)
		})
	}
}
