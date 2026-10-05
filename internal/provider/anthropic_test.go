package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAnthropicProvider_Validation verifies request validation and a valid request.
func TestAnthropicProvider_Validation(t *testing.T) {
	testCases := []struct {
		name        string
		request     ChatRequest
		expectError bool
	}{
		{
			name: "Missing API key",
			request: ChatRequest{
				Model: "claude-sonnet-4-5",
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
				Model:    "claude-sonnet-4-5",
				Messages: []Message{},
			},
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			provider := AnthropicProvider{}

			_, err := provider.Chat(testCase.request)

			if testCase.expectError && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}

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

			if request.URL.Path != "/messages" {
				t.Fatalf(
					"expected /messages, got %s",
					request.URL.Path,
				)
			}

			if request.Header.Get("x-api-key") != "test-key" {
				t.Fatal("expected Anthropic API key header")
			}

			if request.Header.Get("anthropic-version") != anthropicAPIVersion {
				t.Fatalf("expected Anthropic API version header")
			}

			writer.Header().Set(
				"Content-Type",
				"application/json",
			)

			_, _ = writer.Write([]byte(`{
				"content": [
					{
						"type": "text",
						"text": "Hello from Anthropic"
					}
				]
			}`))
		}))

		defer server.Close()

		provider := AnthropicProvider{
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

		if response.Content != "Hello from Anthropic" {
			t.Fatalf(
				"expected response %q, got %q",
				"Hello from Anthropic",
				response.Content,
			)
		}
	})
}

// TestAnthropicProvider_ToolCalls verifies Anthropic tool definitions and tool calls.
func TestAnthropicProvider_ToolCalls(t *testing.T) {
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

		var body AnthropicChatRequest

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

		if body.MaxTokens != 1024 {
			t.Fatalf(
				"expected max tokens %d, got %d",
				1024,
				body.MaxTokens,
			)
		}

		if len(body.Tools) != 1 {
			t.Fatalf(
				"expected 1 tool, got %d",
				len(body.Tools),
			)
		}

		if body.Tools[0].Name != "calculator" {
			t.Fatalf(
				"expected calculator tool, got %q",
				body.Tools[0].Name,
			)
		}

		if body.Tools[0].Description != "Performs calculations." {
			t.Fatalf(
				"expected calculator description, got %q",
				body.Tools[0].Description,
			)
		}

		writer.Header().Set(
			"Content-Type",
			"application/json",
		)

		_, _ = writer.Write([]byte(`{
			"content": [
				{
					"type": "tool_use",
					"id": "toolu_test",
					"name": "calculator",
					"input": {
						"operation": "add",
						"numbers": [5, 7]
					}
				}
			]
		}`))
	}))

	defer server.Close()

	provider := AnthropicProvider{
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

// TestAnthropicProvider_HTTPError verifies HTTP errors returned by Anthropic.
func TestAnthropicProvider_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		writer.WriteHeader(http.StatusUnauthorized)

		_, _ = writer.Write(
			[]byte(`{"error":{"type":"authentication_error","message":"invalid API key"}}`),
		)
	}))

	defer server.Close()

	provider := AnthropicProvider{
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

// TestAnthropicProvider_MalformedResponse verifies malformed responses return errors.
func TestAnthropicProvider_MalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		writer.Header().Set(
			"Content-Type",
			"application/json",
		)

		_, _ = writer.Write(
			[]byte(`{"content":`),
		)
	}))

	defer server.Close()

	provider := AnthropicProvider{
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

// TestConvertToAnthropicMessages verifies message and tool-call conversion.
func TestConvertToAnthropicMessages(t *testing.T) {
	testCases := []struct {
		name             string
		messages         []Message
		expectedMessages int
		expectedParts    int
		expectedRole     string
	}{
		{
			name: "converts user message",
			messages: []Message{
				{
					Role:    "user",
					Content: "Hello",
				},
			},
			expectedMessages: 1,
			expectedParts:    1,
			expectedRole:     "user",
		},
		{
			name: "converts assistant message",
			messages: []Message{
				{
					Role:    "assistant",
					Content: "Hello",
				},
			},
			expectedMessages: 1,
			expectedParts:    1,
			expectedRole:     "assistant",
		},
		{
			name: "converts assistant tool call",
			messages: []Message{
				{
					Role:    "assistant",
					Content: "",
					ToolCalls: []ToolCall{
						{
							Name: "calculator",
							Arguments: map[string]any{
								"operation": "add",
								"numbers":   []any{1.0, 2.0},
							},
						},
					},
				},
			},
			expectedMessages: 1,
			expectedParts:    1,
			expectedRole:     "assistant",
		},
		{
			name: "converts assistant message with text and tool call",
			messages: []Message{
				{
					Role:    "assistant",
					Content: "I will calculate that.",
					ToolCalls: []ToolCall{
						{
							Name: "calculator",
							Arguments: map[string]any{
								"operation": "add",
							},
						},
					},
				},
			},
			expectedMessages: 1,
			expectedParts:    2,
			expectedRole:     "assistant",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := convertToAnthropicMessages(
				testCase.messages,
			)

			if len(result) != testCase.expectedMessages {
				t.Fatalf(
					"expected %d messages, got %d",
					testCase.expectedMessages,
					len(result),
				)
			}

			if result[0].Role != testCase.expectedRole {
				t.Fatalf(
					"expected role %q, got %q",
					testCase.expectedRole,
					result[0].Role,
				)
			}

			if len(result[0].Content) != testCase.expectedParts {
				t.Fatalf(
					"expected %d content parts, got %d",
					testCase.expectedParts,
					len(result[0].Content),
				)
			}
		})
	}
}

// TestConvertToAnthropicTools verifies tool definitions are converted correctly.
func TestConvertToAnthropicTools(t *testing.T) {
	testCases := []struct {
		name          string
		tools         []ToolDefinition
		expectedTools int
	}{
		{
			name:          "no tools",
			tools:         nil,
			expectedTools: 0,
		},
		{
			name: "converts tools",
			tools: []ToolDefinition{
				{
					Name:        "calculator",
					Description: "Performs calculations.",
					Parameters: map[string]any{
						"type": "object",
					},
				},
			},
			expectedTools: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := convertToAnthropicTools(
				testCase.tools,
			)

			if len(result) != testCase.expectedTools {
				t.Fatalf(
					"expected %d tools, got %d",
					testCase.expectedTools,
					len(result),
				)
			}

			if testCase.expectedTools > 0 {
				if result[0].Name != "calculator" {
					t.Fatalf(
						"expected calculator tool, got %q",
						result[0].Name,
					)
				}

				if result[0].Description != "Performs calculations." {
					t.Fatalf(
						"expected description, got %q",
						result[0].Description,
					)
				}
			}
		})
	}
}

// TestConvertFromAnthropicResponse verifies Anthropic responses are converted correctly.
func TestConvertFromAnthropicResponse(t *testing.T) {
	response := AnthropicResponse{
		Content: []AnthropicContentBlock{
			{
				Type: "text",
				Text: "I will calculate that.",
			},
			{
				Type: "tool_use",
				ID:   "toolu_test",
				Name: "calculator",
				Input: map[string]any{
					"operation": "add",
					"numbers":   []any{1.0, 2.0},
				},
			},
		},
	}

	result := convertFromAnthropicResponse(response)

	if result.Content != "I will calculate that." {
		t.Fatalf(
			"expected content %q, got %q",
			"I will calculate that.",
			result.Content,
		)
	}

	if len(result.ToolCalls) != 1 {
		t.Fatalf(
			"expected 1 tool call, got %d",
			len(result.ToolCalls),
		)
	}

	if result.ToolCalls[0].Name != "calculator" {
		t.Fatalf(
			"expected tool name %q, got %q",
			"calculator",
			result.ToolCalls[0].Name,
		)
	}
}

// TestConvertToAnthropicRequest verifies the complete request conversion.
func TestConvertToAnthropicRequest(t *testing.T) {
	request := ChatRequest{
		Model: "test-model",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
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
	}

	result := convertToAnthropicRequest(request)

	if result.Model != "test-model" {
		t.Fatalf(
			"expected model %q, got %q",
			"test-model",
			result.Model,
		)
	}

	if result.MaxTokens != 1024 {
		t.Fatalf(
			"expected max tokens %d, got %d",
			1024,
			result.MaxTokens,
		)
	}

	if len(result.Messages) != 1 {
		t.Fatalf(
			"expected 1 message, got %d",
			len(result.Messages),
		)
	}

	if result.Messages[0].Role != "user" {
		t.Fatalf(
			"expected user role, got %q",
			result.Messages[0].Role,
		)
	}

	if len(result.Tools) != 1 {
		t.Fatalf(
			"expected 1 tool, got %d",
			len(result.Tools),
		)
	}
}
