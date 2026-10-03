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
				"calculator",
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

// TestOpenAIProvider_HTTPError verifies HTTP errors returned by the provider.
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

// TestOpenAIProvider_MalformedResponse verifies malformed JSON responses fail.
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

// TestOpenAIProvider_NoChoices verifies an empty choices response fails.
func TestOpenAIProvider_NoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		writer.Header().Set(
			"Content-Type",
			"application/json",
		)

		_, _ = writer.Write([]byte(`{"choices":[]}`))
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
		t.Fatal("expected no-choices error, got nil")
	}
}

// TestConvertToOpenAIMessages verifies messages and tool calls are converted correctly.
func TestConvertToOpenAIMessages(t *testing.T) {
	testCases := []struct {
		name              string
		messages          []Message
		expectedMessages  int
		expectedToolCalls int
	}{
		{
			name: "converts message without tool calls",
			messages: []Message{
				{
					Role:    "user",
					Content: "Hello",
				},
			},
			expectedMessages:  1,
			expectedToolCalls: 0,
		},
		{
			name: "converts message with tool calls",
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
			expectedMessages:  1,
			expectedToolCalls: 1,
		},
		{
			name: "skips tool call with invalid arguments",
			messages: []Message{
				{
					Role:    "assistant",
					Content: "",
					ToolCalls: []ToolCall{
						{
							Name: "invalid-tool",
							Arguments: map[string]any{
								"invalid": func() {},
							},
						},
					},
				},
			},
			expectedMessages:  1,
			expectedToolCalls: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := convertToOpenAIMessages(testCase.messages)

			if len(result) != testCase.expectedMessages {
				t.Fatalf(
					"expected %d messages, got %d",
					testCase.expectedMessages,
					len(result),
				)
			}

			if len(result[0].ToolCalls) != testCase.expectedToolCalls {
				t.Fatalf(
					"expected %d tool calls, got %d",
					testCase.expectedToolCalls,
					len(result[0].ToolCalls),
				)
			}

			if result[0].Role != testCase.messages[0].Role {
				t.Fatalf(
					"expected role %q, got %q",
					testCase.messages[0].Role,
					result[0].Role,
				)
			}

			if result[0].Content != testCase.messages[0].Content {
				t.Fatalf(
					"expected content %q, got %q",
					testCase.messages[0].Content,
					result[0].Content,
				)
			}
		})
	}
}

// TestConvertFromOpenAIMessage verifies OpenAI tool calls are converted correctly.
func TestConvertFromOpenAIMessage(t *testing.T) {
	testCases := []struct {
		name          string
		message       OpenAIMessage
		expectedArgs  map[string]any
		expectedCalls int
	}{
		{
			name: "valid arguments",
			message: OpenAIMessage{
				Role: "assistant",
				ToolCalls: []OpenAIToolCall{
					{
						Function: OpenAIFunction{
							Name:      "calculator",
							Arguments: `{"operation":"add"}`,
						},
					},
				},
			},
			expectedArgs: map[string]any{
				"operation": "add",
			},
			expectedCalls: 1,
		},
		{
			name: "empty arguments",
			message: OpenAIMessage{
				Role: "assistant",
				ToolCalls: []OpenAIToolCall{
					{
						Function: OpenAIFunction{
							Name:      "calculator",
							Arguments: "",
						},
					},
				},
			},
			expectedArgs:  nil,
			expectedCalls: 1,
		},
		{
			name: "malformed arguments",
			message: OpenAIMessage{
				Role: "assistant",
				ToolCalls: []OpenAIToolCall{
					{
						Function: OpenAIFunction{
							Name:      "calculator",
							Arguments: `{"operation":`,
						},
					},
				},
			},
			expectedArgs:  map[string]any{},
			expectedCalls: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := convertFromOpenAIMessage(testCase.message)

			if len(result.ToolCalls) != testCase.expectedCalls {
				t.Fatalf(
					"expected %d tool calls, got %d",
					testCase.expectedCalls,
					len(result.ToolCalls),
				)
			}

			if testCase.expectedCalls > 0 {
				if len(result.ToolCalls[0].Arguments) != len(testCase.expectedArgs) {
					t.Fatalf(
						"expected %d arguments, got %d",
						len(testCase.expectedArgs),
						len(result.ToolCalls[0].Arguments),
					)
				}
			}
		})
	}
}
