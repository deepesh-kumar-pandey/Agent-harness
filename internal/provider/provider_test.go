package provider

import (
	"encoding/json"
	"testing"
)

func TestChatRequestJSON(t *testing.T) {
	testCases := []struct {
		name        string
		request     ChatRequest
		expectError bool
	}{
		{
			name: "Request without tools",
			request: ChatRequest{
				Model: "test-model",
				Messages: []Message{
					{
						Role:    "user",
						Content: "Hello",
					},
				},
			},
			expectError: false,
		},
		{
			name: "Request with tools",
			request: ChatRequest{
				Model: "test-model",
				Messages: []Message{
					{
						Role:    "user",
						Content: "Calculate 10 + 20",
					},
				},
				Tools: []ToolDefinition{
					{
						Name:        "calculator",
						Description: "A tool for performing mathematical calculations.",
						Parameters: map[string]any{
							"operation": "string",
							"numbers":   "array of numbers",
						},
					},
				},
			},
			expectError: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			data, err := json.Marshal(testCase.request)

			if testCase.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if data != nil {
					t.Fatalf("expected no JSON data, got %s", data)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if len(data) == 0 {
				t.Fatal("expected JSON data, got empty")
			}
		})
	}
}
