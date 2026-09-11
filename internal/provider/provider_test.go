package provider

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

func TestOllamaProvider_ToolCalls(t *testing.T) {
	testCases := []struct {
		name              string
		response          string
		expectedToolName  string
		expectedOperation string
		expectedNumbers   []any
	}{
		{
			name: "Calculator tool call",
			response: `{
				"message": {
					"role": "assistant",
					"content": "",
					"tool_calls": [
						{
							"function": {
								"name": "calculator",
								"arguments": {
									"operation": "add",
									"numbers": [5, 7]
								}
							}
						}
					]
				}
			}`,
			expectedToolName:  "calculator",
			expectedOperation: "add",
			expectedNumbers:   []any{float64(5), float64(7)},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(
				writer http.ResponseWriter,
				request *http.Request,
			) {
				if request.Method != http.MethodPost {
					t.Fatalf(
						"expected POST request, got %s",
						request.Method,
					)
				}

				writer.Header().Set(
					"Content-Type",
					"application/json",
				)

				_, _ = writer.Write(
					[]byte(testCase.response),
				)
			}))

			defer server.Close()

			provider := OllamaProvider{
				BaseURL: server.URL,
				Client:  server.Client(),
			}

			request := ChatRequest{
				Model: "test-model",
				Messages: []Message{
					{
						Role:    "user",
						Content: "Calculate 5 + 7",
					},
				},
			}

			response, err := provider.Chat(request)

			if err != nil {
				t.Fatalf(
					"expected no error, got: %v",
					err,
				)
			}

			if len(response.ToolCalls) != 1 {
				t.Fatalf(
					"expected 1 tool call, got %d",
					len(response.ToolCalls),
				)
			}

			toolCall := response.ToolCalls[0]

			if toolCall.Name != testCase.expectedToolName {
				t.Fatalf(
					"expected tool name %q, got %q",
					testCase.expectedToolName,
					toolCall.Name,
				)
			}

			operation, ok := toolCall.Arguments["operation"].(string)

			if !ok {
				t.Fatal("expected operation to be a string")
			}

			if operation != testCase.expectedOperation {
				t.Fatalf(
					"expected operation %q, got %q",
					testCase.expectedOperation,
					operation,
				)
			}

			numbers, ok := toolCall.Arguments["numbers"].([]any)

			if !ok {
				t.Fatal("expected numbers to be []any")
			}

			if len(numbers) != len(testCase.expectedNumbers) {
				t.Fatalf(
					"expected %d numbers, got %d",
					len(testCase.expectedNumbers),
					len(numbers),
				)
			}

			for index, number := range testCase.expectedNumbers {
				if numbers[index] != number {
					t.Fatalf(
						"expected number %v at index %d, got %v",
						number,
						index,
						numbers[index],
					)
				}
			}
		})
	}
}

func TestOllamaProvider_ListModels(t *testing.T) {
	testCases := []struct {
		name        string
		response    string
		statusCode  int
		serverError bool
		expectError bool
		expected    []string
	}{
		{
			name:       "Returns model names",
			response:   `{"models":[{"name":"kirito1/qwen3-coder:4b"},{"name":"mistral:latest"}]}`,
			statusCode: http.StatusOK,
			expected:   []string{"kirito1/qwen3-coder:4b", "mistral:latest"},
		},
		{
			name:        "Rejects malformed response",
			response:    `{"models":`,
			statusCode:  http.StatusOK,
			expectError: true,
		},
		{
			name:        "Reports HTTP error",
			response:    `{"error":"unavailable"}`,
			statusCode:  http.StatusInternalServerError,
			expectError: true,
		},
		{
			name:        "Reports server failure",
			serverError: true,
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var provider OllamaProvider

			if testCase.serverError {
				provider = OllamaProvider{
					BaseURL: "http://127.0.0.1:1",
				}
			} else {
				server := httptest.NewServer(http.HandlerFunc(func(
					writer http.ResponseWriter,
					request *http.Request,
				) {
					if request.Method != http.MethodGet {
						t.Fatalf("expected GET request, got %s", request.Method)
					}

					if request.URL.Path != "/api/tags" {
						t.Fatalf("expected /api/tags, got %s", request.URL.Path)
					}

					writer.WriteHeader(testCase.statusCode)
					_, _ = writer.Write([]byte(testCase.response))
				}))
				defer server.Close()

				provider = OllamaProvider{
					BaseURL: server.URL,
					Client:  server.Client(),
				}
			}

			models, err := provider.ListModels()

			if testCase.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if len(models) != len(testCase.expected) {
				t.Fatalf("expected %d models, got %d", len(testCase.expected), len(models))
			}

			for index, expectedModel := range testCase.expected {
				if models[index] != expectedModel {
					t.Fatalf("expected model %q, got %q", expectedModel, models[index])
				}
			}
		})
	}
}

func TestOllamaProvider_HasModel(t *testing.T) {
	testCases := []struct {
		name          string
		model         string
		expectedFound bool
	}{
		{
			name:          "Configured model is present",
			model:         "llama3.1",
			expectedFound: true,
		},
		{
			name:          "Configured model is missing",
			model:         "missing-model",
			expectedFound: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(
				writer http.ResponseWriter,
				request *http.Request,
			) {
				_, _ = writer.Write([]byte(
					`{"models":[{"name":"llama3.1"},{"name":"mistral:latest"}]}`,
				))
			}))
			defer server.Close()

			provider := OllamaProvider{
				BaseURL: server.URL,
				Client:  server.Client(),
			}

			found, err := provider.HasModel(testCase.model)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if found != testCase.expectedFound {
				t.Fatalf("expected found=%t, got %t", testCase.expectedFound, found)
			}
		})
	}
}

func TestOllamaProvider_PullModel(t *testing.T) {
	testCases := []struct {
		name        string
		statusCode  int
		expectError bool
	}{
		{
			name:       "Pulls model successfully",
			statusCode: http.StatusOK,
		},
		{
			name:        "Reports pull error",
			statusCode:  http.StatusInternalServerError,
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(
				writer http.ResponseWriter,
				request *http.Request,
			) {
				if request.Method != http.MethodPost {
					t.Fatalf("expected POST request, got %s", request.Method)
				}

				if request.URL.Path != "/api/pull" {
					t.Fatalf("expected /api/pull, got %s", request.URL.Path)
				}

				var body struct {
					Name string `json:"name"`
				}
				if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode request: %v", err)
				}

				if body.Name != "llama3.1" {
					t.Fatalf("expected model llama3.1, got %q", body.Name)
				}

				writer.WriteHeader(testCase.statusCode)
				_, _ = writer.Write([]byte(`{"status":"success"}`))
			}))
			defer server.Close()

			provider := OllamaProvider{
				BaseURL: server.URL,
				Client:  server.Client(),
			}

			err := provider.PullModel("llama3.1")
			if testCase.expectError && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
		})
	}
}
