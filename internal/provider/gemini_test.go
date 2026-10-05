package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Tests Gemini Chat validation for missing API key, model, messages, and valid requests.
func TestGeminiProvider_Validation(t *testing.T) {
	testCases := []struct {
		name        string
		provider    GeminiProvider
		request     ChatRequest
		expectError bool
	}{
		{
			name:     "Missing API key",
			provider: GeminiProvider{},
			request: ChatRequest{
				Model: "gemini-2.5-flash",
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
			provider: GeminiProvider{
				APIKey: "test-key",
			},
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
			provider: GeminiProvider{
				APIKey: "test-key",
			},
			request: ChatRequest{
				Model:    "gemini-2.5-flash",
				Messages: []Message{},
			},
			expectError: true,
		},
		{
			name: "Valid request",
			provider: GeminiProvider{
				APIKey: "test-key",
			},
			request: ChatRequest{
				Model: "gemini-2.5-flash",
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
			testProvider := testCase.provider

			if testCase.name == "Valid request" {
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

					if request.URL.Path != "/v1beta/models/gemini-2.5-flash:generateContent" {
						t.Fatalf(
							"expected Gemini generateContent path, got %s",
							request.URL.Path,
						)
					}

					if request.Header.Get("x-goog-api-key") != "test-key" {
						t.Fatalf("expected API key header")
					}

					writer.Header().Set(
						"Content-Type",
						"application/json",
					)

					_, _ = writer.Write(
						[]byte(`{
							"candidates": [
								{
									"content": {
										"role": "model",
										"parts": [
											{
												"text": "Hello"
											}
										]
									}
								}
							]
						}`),
					)
				}))

				defer server.Close()

				testProvider = GeminiProvider{
					APIKey:  "test-key",
					BaseURL: server.URL + "/v1beta",
					Client:  server.Client(),
				}
			}

			_, err := testProvider.Chat(testCase.request)

			if testCase.expectError && err == nil {
				t.Fatalf("expected error, got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// Tests conversion of Gemini function-call responses into generic tool calls.
func TestGeminiProvider_ToolCalls(t *testing.T) {
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
				"candidates": [
					{
						"content": {
							"role": "model",
							"parts": [
								{
									"functionCall": {
										"name": "calculator",
										"args": {
											"operation": "add",
											"numbers": [5, 7]
										}
									}
								}
							]
						}
					}
				]
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

			provider := GeminiProvider{
				APIKey:  "test-key",
				BaseURL: server.URL,
				Client:  server.Client(),
			}

			request := ChatRequest{
				Model: "gemini-2.5-flash",
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

// Tests conversion of a generic ChatRequest into a Gemini request.
func TestConvertToGeminiRequest(t *testing.T) {
	request := ChatRequest{
		Model: "gemini-2.5-flash",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Calculate 5 + 7",
				ToolCalls: []ToolCall{
					{
						Name: "calculator",
						Arguments: map[string]any{
							"operation": "add",
							"numbers":   []any{5, 7},
						},
					},
				},
			},
		},
		Tools: []ToolDefinition{
			{
				Name:        "calculator",
				Description: "Perform calculations",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"operation": map[string]any{
							"type": "string",
						},
					},
				},
			},
		},
	}

	result := convertToGeminiRequest(request)

	if len(result.Contents) != 1 {
		t.Fatalf(
			"expected 1 content, got %d",
			len(result.Contents),
		)
	}

	content := result.Contents[0]

	if content.Role != "user" {
		t.Fatalf(
			"expected role user, got %q",
			content.Role,
		)
	}

	if len(content.Parts) != 2 {
		t.Fatalf(
			"expected 2 parts, got %d",
			len(content.Parts),
		)
	}

	if content.Parts[0].Text != "Calculate 5 + 7" {
		t.Fatalf(
			"expected content %q, got %q",
			"Calculate 5 + 7",
			content.Parts[0].Text,
		)
	}

	if content.Parts[1].FunctionCall == nil {
		t.Fatal("expected function call")
	}

	if content.Parts[1].FunctionCall.Name != "calculator" {
		t.Fatalf(
			"expected function name calculator, got %q",
			content.Parts[1].FunctionCall.Name,
		)
	}

	if content.Parts[1].FunctionCall.Args["operation"] != "add" {
		t.Fatalf(
			"expected operation add, got %v",
			content.Parts[1].FunctionCall.Args["operation"],
		)
	}

	if len(result.Tools) != 1 {
		t.Fatalf(
			"expected 1 tool, got %d",
			len(result.Tools),
		)
	}

	tool := result.Tools[0]

	if len(tool.FunctionDeclarations) != 1 {
		t.Fatalf(
			"expected 1 function declaration, got %d",
			len(tool.FunctionDeclarations),
		)
	}

	function := tool.FunctionDeclarations[0]

	if function.Name != "calculator" {
		t.Fatalf(
			"expected function name calculator, got %q",
			function.Name,
		)
	}

	if function.Description != "Perform calculations" {
		t.Fatalf(
			"expected description %q, got %q",
			"Perform calculations",
			function.Description,
		)
	}

	if function.Parameters["type"] != "object" {
		t.Fatalf(
			"expected parameter type object, got %v",
			function.Parameters["type"],
		)
	}
}

// Tests successful Gemini chat response conversion.
func TestGeminiProvider_Chat(t *testing.T) {
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

		if request.URL.Path != "/models/gemini-2.5-flash:generateContent" {
			t.Fatalf(
				"expected generateContent path, got %s",
				request.URL.Path,
			)
		}

		if request.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("expected application/json content type")
		}

		if request.Header.Get("x-goog-api-key") != "test-key" {
			t.Fatalf("expected Gemini API key")
		}

		var body GeminiChatRequest

		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf(
				"failed to decode request: %v",
				err,
			)
		}

		if len(body.Contents) != 1 {
			t.Fatalf(
				"expected 1 content, got %d",
				len(body.Contents),
			)
		}

		if body.Contents[0].Parts[0].Text != "Hello" {
			t.Fatalf(
				"expected Hello, got %q",
				body.Contents[0].Parts[0].Text,
			)
		}

		writer.Header().Set(
			"Content-Type",
			"application/json",
		)

		_, _ = writer.Write(
			[]byte(`{
				"candidates": [
					{
						"content": {
							"role": "model",
							"parts": [
								{
									"text": "Hello from Gemini"
								}
							]
						}
					}
				]
			}`),
		)
	}))

	defer server.Close()

	provider := GeminiProvider{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	request := ChatRequest{
		Model: "gemini-2.5-flash",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
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

	if response.Content != "Hello from Gemini" {
		t.Fatalf(
			"expected content %q, got %q",
			"Hello from Gemini",
			response.Content,
		)
	}
}

// Tests Gemini HTTP error handling.
func TestGeminiProvider_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		writer.WriteHeader(http.StatusInternalServerError)

		_, _ = writer.Write(
			[]byte(`{"error":"internal server error"}`),
		)
	}))

	defer server.Close()

	provider := GeminiProvider{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	request := ChatRequest{
		Model: "gemini-2.5-flash",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	_, err := provider.Chat(request)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// Tests malformed Gemini responses.
func TestGeminiProvider_MalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		writer.Header().Set(
			"Content-Type",
			"application/json",
		)

		_, _ = writer.Write(
			[]byte(`{"candidates":`),
		)
	}))

	defer server.Close()

	provider := GeminiProvider{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	request := ChatRequest{
		Model: "gemini-2.5-flash",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	_, err := provider.Chat(request)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// Tests Gemini responses without candidates.
func TestGeminiProvider_NoCandidates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		writer.Header().Set(
			"Content-Type",
			"application/json",
		)

		_, _ = writer.Write(
			[]byte(`{"candidates":[]}`),
		)
	}))

	defer server.Close()

	provider := GeminiProvider{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	request := ChatRequest{
		Model: "gemini-2.5-flash",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	_, err := provider.Chat(request)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
