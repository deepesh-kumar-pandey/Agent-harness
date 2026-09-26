package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIProvider_Validation(t *testing.T) {
	testCases := []struct {
		name        string
		request     ChatRequest
		expectError bool
	}{
		{
			name: "Missing API key",
			request: ChatRequest{
				Model: "gpt-4o",
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
				Model:    "gpt-4o",
				Messages: []Message{},
			},
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			provider := OpenAIProvider{}

			_, err := provider.Chat(testCase.request)

			if testCase.expectError && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}

	// Test a valid request separately using a local mock server.
	t.Run("Valid request", func(t *testing.T) {
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

			if request.URL.Path != "/chat/completions" {
				t.Fatalf(
					"expected /chat/completions, got %s",
					request.URL.Path,
				)
			}

			if request.Header.Get("Authorization") != "Bearer test-key" {
				t.Fatalf("expected OpenAI authorization header")
			}

			writer.Header().Set(
				"Content-Type",
				"application/json",
			)

			_, _ = writer.Write([]byte(`{
				"choices": [
					{
						"message": {
							"role": "assistant",
							"content": "Hello from OpenAI"
						}
					}
				]
			}`))
		}))

		defer server.Close()

		provider := OpenAIProvider{
			APIKey:  "test-key",
			BaseURL: server.URL,
			Client:  server.Client(),
		}

		response, err := provider.Chat(ChatRequest{
			Model: "test-model",
			Messages: []Message{
				{
					Role:    "user",
					Content: "Hello",
				},
			},
		})

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if response.Content != "Hello from OpenAI" {
			t.Fatalf(
				"expected response %q, got %q",
				"Hello from OpenAI",
				response.Content,
			)
		}
	})
}

func TestOpenAIProvider_ToolCalls(t *testing.T) {
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

		var body OpenAIChatRequest

		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf(
				"failed to decode request: %v",
				err,
			)
		}

		if body.Model != "test-model" {
			t.Fatalf(
				"expected model %q, got %q",
				"test-model",
				body.Model,
			)
		}

		if len(body.Tools) != 1 {
			t.Fatalf(
				"expected 1 tool, got %d",
				len(body.Tools),
			)
		}

		if body.Tools[0].Type != "function" {
			t.Fatalf(
				"expected function tool type, got %q",
				body.Tools[0].Type,
			)
		}

		if body.Tools[0].Function.Name != "calculator" {
			t.Fatalf(
				"expected calculator tool, got %q",
				body.Tools[0].Function.Name,
			)
		}

		writer.Header().Set(
			"Content-Type",
			"application/json",
		)

		_, _ = writer.Write([]byte(`{
			"choices": [
				{
					"message": {
						"role": "assistant",
						"content": "",
						"tool_calls": [
							{
								"function": {
									"name": "calculator",
									"arguments": "{\"operation\":\"add\",\"numbers\":[5,7]}"
								}
							}
						]
					}
				}
			]
		}`))
	}))

	defer server.Close()

	provider := OpenAIProvider{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	response, err := provider.Chat(ChatRequest{
		Model: "test-model",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Calculate 5 + 7",
			},
		},
		Tools: []ToolDefinition{
			{
				Name:        "calculator",
				Description: "Performs calculations.",
				Parameters: map[string]any{
					"type": "object",
				},
			},
		},
	})

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

	if toolCall.Name != "calculator" {
		t.Fatalf(
			"expected tool name %q, got %q",
			"calculator",
			toolCall.Name,
		)
	}

	operation, ok := toolCall.Arguments["operation"].(string)

	if !ok {
		t.Fatal("expected operation to be a string")
	}

	if operation != "add" {
		t.Fatalf(
			"expected operation %q, got %q",
			"add",
			operation,
		)
	}

	numbers, ok := toolCall.Arguments["numbers"].([]any)

	if !ok {
		t.Fatal("expected numbers to be []any")
	}

	if len(numbers) != 2 {
		t.Fatalf(
			"expected 2 numbers, got %d",
			len(numbers),
		)
	}

	if numbers[0] != float64(5) || numbers[1] != float64(7) {
		t.Fatalf(
			"expected numbers [5 7], got %v",
			numbers,
		)
	}
}

func TestOpenAIProvider_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		writer.WriteHeader(http.StatusUnauthorized)
		_, _ = writer.Write(
			[]byte(`{"error":{"message":"invalid API key"}}`),
		)
	}))

	defer server.Close()

	provider := OpenAIProvider{
		APIKey:  "invalid-key",
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	_, err := provider.Chat(ChatRequest{
		Model: "test-model",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	})

	if err == nil {
		t.Fatal("expected HTTP error, got nil")
	}
}

func TestOpenAIProvider_MalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		writer.Header().Set(
			"Content-Type",
			"application/json",
		)

		_, _ = writer.Write(
			[]byte(`{"choices":`),
		)
	}))

	defer server.Close()

	provider := OpenAIProvider{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	_, err := provider.Chat(ChatRequest{
		Model: "test-model",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	})

	if err == nil {
		t.Fatal("expected malformed response error, got nil")
	}
}
