package tools

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOllamaProvider_Validation(t *testing.T) {

	provider := OllamaProvider{}

	testCases := []struct {
		name        string
		request     ChatRequest
		expectError bool
	}{
		{
			name: "Missing model",
			request: ChatRequest{
				Model: "",
				Messages: []Message{
					{
						Role:    "user",
						Content: "Hello",
					},
				},
			},
			expectError: true,
		},
		{
			name: "Missing messages",
			request: ChatRequest{
				Model:    "llama3.1",
				Messages: []Message{},
			},
			expectError: true,
		},
		{
			name: "Valid request",
			request: ChatRequest{
				Model: "llama3.1",
				Messages: []Message{
					{
						Role:    "user",
						Content: "Hello",
					},
				},
			},
			expectError: false,
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {
			testProvider := provider
			if testCase.name == "Valid request" {
				server := httptest.NewServer(http.HandlerFunc(func(
					writer http.ResponseWriter,
					request *http.Request,
				) {
					if request.Method != http.MethodPost {
						t.Fatalf("expected POST request, got %s", request.Method)
					}
					writer.Header().Set("Content-Type", "application/json")
					_, _ = writer.Write([]byte(`{"message":{"content":"Hello"}}`))
				}))
				defer server.Close()
				testProvider = OllamaProvider{
					BaseURL: server.URL,
					Client:  server.Client(),
				}
			}

			_, err := testProvider.Chat(testCase.request)

			if testCase.expectError && err == nil {
				t.Fatalf("❌ expected error, got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("❌ unexpected error: %v", err)
			}

			t.Logf("✅ Test passed: %s", testCase.name)
		})
	}
}

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
