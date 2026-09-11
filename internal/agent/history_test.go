package tools

import (
	"reflect"
	"testing"

	providerpkg "agent-harness/internal/provider"
)

func TestConversation(t *testing.T) {
	testCases := []struct {
		name             string
		messages         []providerpkg.Message
		expectedMessages int
		expectError      bool
	}{
		{
			name:             "Creates empty conversation",
			messages:         nil,
			expectedMessages: 0,
		},
		{
			name: "Adds messages in order",
			messages: []providerpkg.Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi"},
			},
			expectedMessages: 2,
		},
		{
			name: "Allows empty content",
			messages: []providerpkg.Message{
				{Role: "user", Content: ""},
			},
			expectedMessages: 1,
		},
		{
			name: "Rejects message with empty role",
			messages: []providerpkg.Message{
				{Role: "", Content: "Invalid message"},
			},
			expectedMessages: 0,
			expectError:      true,
		},
		{
			name: "Rejects invalid message after valid messages",
			messages: []providerpkg.Message{
				{Role: "user", Content: "Hello"},
				{Role: "", Content: "Invalid message"},
			},
			expectedMessages: 1,
			expectError:      true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			conversation := NewConversation()

			if conversation == nil {
				t.Fatal("expected conversation, got nil")
			}

			for _, message := range testCase.messages {
				err := conversation.AddMessage(message)
				if err != nil {
					if !testCase.expectError {
						t.Fatalf("unexpected error: %v", err)
					}

					break
				}
			}

			messages := conversation.GetMessages()

			if len(messages) != testCase.expectedMessages {
				t.Fatalf(
					"expected %d messages, got %d",
					testCase.expectedMessages,
					len(messages),
				)
			}

			for index, expectedMessage := range testCase.messages {
				if index >= len(messages) {
					break
				}

				if !reflect.DeepEqual(messages[index], expectedMessage) {
					t.Fatalf(
						"expected message %v at index %d, got %v",
						expectedMessage,
						index,
						messages[index],
					)
				}
			}
		})
	}
}
