package orchestrator

import (
	"fmt"
	"testing"

	agentpkg "agent-harness/internal/agent"
	providerpkg "agent-harness/internal/provider"
	toolspkg "agent-harness/internal/tools"
)

// ─────────────────────────────────────────────
// Fake Provider
// ─────────────────────────────────────────────

type FakeProvider struct {
	Responses []string
	Index     int
}

func (f *FakeProvider) Chat(
	request providerpkg.ChatRequest,
) (providerpkg.ChatResponse, error) {

	if f.Index >= len(f.Responses) {
		return providerpkg.ChatResponse{}, fmt.Errorf(
			"fake provider has no more responses",
		)
	}

	response := f.Responses[f.Index]
	f.Index++

	return providerpkg.ChatResponse{
		Content: response,
	}, nil
}

// ─────────────────────────────────────────────
// Test Orchestrator Run
// ─────────────────────────────────────────────

func TestOrchestratorRun(t *testing.T) {

	fmt.Println("Starting Orchestrator Run tests...")

	registry := toolspkg.NewToolRegistry()
	testAgent := agentpkg.NewAgent(registry)

	fakeProvider := &FakeProvider{
		Responses: []string{
			"Fake response",
		},
	}

	testOrchestrator := NewOrchestrator(
		testAgent,
		fakeProvider,
	)

	testCases := []struct {
		name        string
		toolName    string
		args        map[string]any
		expectError bool
	}{
		{
			name:     "Execute calculator",
			toolName: "calculator",
			args: map[string]any{
				"operation": "add",
				"numbers":   []float64{10, 20},
			},
			expectError: false,
		},
		{
			name:     "Execute shell",
			toolName: "shell",
			args: map[string]any{
				"command": "echo",
				"args":    []string{"Hello"},
			},
			expectError: false,
		},
		{
			name:        "Unknown tool",
			toolName:    "unknown",
			args:        map[string]any{},
			expectError: true,
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {

			fmt.Printf("Running test: %s\n", testCase.name)

			result, err := testOrchestrator.Run(
				testCase.toolName,
				testCase.args,
			)

			if testCase.expectError && err == nil {
				t.Fatalf("Expected an error, but got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !testCase.expectError {
				fmt.Printf(
					"Tool executed successfully: %v\n",
					result,
				)
			} else {
				fmt.Printf(
					"Error correctly returned: %v\n",
					err,
				)
			}
		})
	}

	fmt.Println("Orchestrator Run tests completed!")
}

// ─────────────────────────────────────────────
// Test Orchestrator Chat
// ─────────────────────────────────────────────

func TestOrchestratorChat(t *testing.T) {

	fmt.Println("Starting Orchestrator Chat tests...")

	registry := toolspkg.NewToolRegistry()
	testAgent := agentpkg.NewAgent(registry)

	fakeProvider := &FakeProvider{
		Responses: []string{
			"Fake response",
		},
	}

	testOrchestrator := NewOrchestrator(
		testAgent,
		fakeProvider,
	)

	testCases := []struct {
		name        string
		request     providerpkg.ChatRequest
		expectError bool
		expected    string
	}{
		{
			name: "Simple chat",
			request: providerpkg.ChatRequest{
				Model: "test-model",
				Messages: []providerpkg.Message{
					{
						Role:    "user",
						Content: "Hello",
					},
				},
			},
			expectError: false,
			expected:    "Fake response",
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {

			fmt.Printf("Running test: %s\n", testCase.name)

			response, err := testOrchestrator.Chat(
				testCase.request,
			)

			if testCase.expectError && err == nil {
				t.Fatalf("Expected an error, but got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !testCase.expectError &&
				response.Content != testCase.expected {

				t.Fatalf(
					"Expected response %q, got %q",
					testCase.expected,
					response.Content,
				)
			}

			if !testCase.expectError {
				fmt.Printf(
					"Provider response: %s\n",
					response.Content,
				)
			} else {
				fmt.Printf(
					"Error correctly returned: %v\n",
					err,
				)
			}
		})
	}

	fmt.Println("Orchestrator Chat tests completed!")
}

// ─────────────────────────────────────────────
// Test Orchestrator AssignTool
// ─────────────────────────────────────────────

func TestOrchestratorAssignTool(t *testing.T) {

	fmt.Println("Starting Orchestrator AssignTool tests...")

	registry := toolspkg.NewToolRegistry()
	testAgent := agentpkg.NewAgent(registry)

	fakeProvider := &FakeProvider{
		Responses: []string{
			"Fake response",
		},
	}

	testOrchestrator := NewOrchestrator(
		testAgent,
		fakeProvider,
	)

	testCases := []struct {
		name        string
		toolCall    ToolCall
		expectError bool
	}{
		{
			name: "Assign calculator tool",
			toolCall: ToolCall{
				Tool: "calculator",
				Args: map[string]any{
					"operation": "add",
					"numbers":   []float64{10, 20},
				},
			},
			expectError: false,
		},
		{
			name: "Assign shell tool",
			toolCall: ToolCall{
				Tool: "shell",
				Args: map[string]any{
					"command": "echo",
					"args":    []string{"Hello"},
				},
			},
			expectError: false,
		},
		{
			name: "Assign unknown tool",
			toolCall: ToolCall{
				Tool: "unknown",
				Args: map[string]any{},
			},
			expectError: true,
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {

			fmt.Printf("Running test: %s\n", testCase.name)

			result, err := testOrchestrator.AssignTool(
				testCase.toolCall,
			)

			if testCase.expectError && err == nil {
				t.Fatalf("Expected an error, but got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !testCase.expectError {
				fmt.Printf(
					"Tool assigned successfully: %v\n",
					result,
				)
			} else {
				fmt.Printf(
					"Error correctly returned: %v\n",
					err,
				)
			}
		})
	}

	fmt.Println("Orchestrator AssignTool tests completed!")
}

// ─────────────────────────────────────────────
// Test Orchestrator RunAgent
// ─────────────────────────────────────────────

func TestOrchestratorRunAgent(t *testing.T) {

	fmt.Println("Starting Orchestrator RunAgent tests...")

	registry := toolspkg.NewToolRegistry()
	testAgent := agentpkg.NewAgent(registry)

	testCases := []struct {
		name            string
		responses       []string
		request         providerpkg.ChatRequest
		expectError     bool
		expectedContent string
	}{
		{
			name: "Plain text response",
			responses: []string{
				"Hello",
			},
			request: providerpkg.ChatRequest{
				Model: "test-model",
				Messages: []providerpkg.Message{
					{
						Role:    "user",
						Content: "Hello",
					},
				},
			},
			expectError:     false,
			expectedContent: "Hello",
		},
		{
			name: "JSON tool call response",
			responses: []string{
				`{
					"content": "",
					"tool_call": {
						"tool": "calculator",
						"args": {
							"operation": "add",
							"numbers": [10, 20]
						}
					}
				}`,
				"30",
			},
			request: providerpkg.ChatRequest{
				Model: "test-model",
				Messages: []providerpkg.Message{
					{
						Role:    "user",
						Content: "Calculate 10 + 20",
					},
				},
			},
			expectError:     false,
			expectedContent: "30",
		},
		{
			name: "Invalid JSON response",
			responses: []string{
				`{"content": "Hello"`,
			},
			request: providerpkg.ChatRequest{
				Model: "test-model",
				Messages: []providerpkg.Message{
					{
						Role:    "user",
						Content: "Hello",
					},
				},
			},
			expectError:     false,
			expectedContent: `{"content": "Hello"`,
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {

			fmt.Printf("Running test: %s\n", testCase.name)

			fakeProvider := &FakeProvider{
				Responses: testCase.responses,
			}

			testOrchestrator := NewOrchestrator(
				testAgent,
				fakeProvider,
			)

			response, err := testOrchestrator.RunAgent(
				testCase.request,
			)

			if testCase.expectError && err == nil {
				t.Fatalf("Expected an error, but got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !testCase.expectError {

				if response.Content != testCase.expectedContent {
					t.Fatalf(
						"Expected content %q, got %q",
						testCase.expectedContent,
						response.Content,
					)
				}

				if response.ToolCall != nil {
					t.Fatalf("Expected no tool call after tool execution")
				}

				fmt.Printf(
					"Agent response: %s\n",
					response.Content,
				)
			}

			if testCase.expectError {
				fmt.Printf(
					"Error correctly returned: %v\n",
					err,
				)
			}
		})
	}

	fmt.Println("Orchestrator RunAgent tests completed!")
}

// ─────────────────────────────────────────────
// Test Orchestrator Max Tool Calls
// ─────────────────────────────────────────────

func TestOrchestratorMaxToolCalls(t *testing.T) {

	fmt.Println("Starting Orchestrator Max Tool Calls tests...")

	registry := toolspkg.NewToolRegistry()
	testAgent := agentpkg.NewAgent(registry)

	fakeProvider := &FakeProvider{
		Responses: []string{
			`{
				"content": "",
				"tool_call": {
					"tool": "calculator",
					"args": {
						"operation": "add",
						"numbers": [10, 20]
					}
				}
			}`,
			`{
				"content": "",
				"tool_call": {
					"tool": "calculator",
					"args": {
						"operation": "add",
						"numbers": [30, 40]
					}
				}
			}`,
			`{
				"content": "",
				"tool_call": {
					"tool": "calculator",
					"args": {
						"operation": "add",
						"numbers": [50, 60]
					}
				}
			}`,
		},
	}

	testOrchestrator := NewOrchestrator(
		testAgent,
		fakeProvider,
		WithMaxToolCalls(2),
	)

	request := providerpkg.ChatRequest{
		Model: "test-model",
		Messages: []providerpkg.Message{
			{
				Role:    "user",
				Content: "Keep calculating",
			},
		},
	}

	_, err := testOrchestrator.RunAgent(request)

	if err == nil {
		t.Fatalf("Expected maximum tool-call error, but got nil")
	}

	expectedError := "maximum tool-call limit (2) exceeded"

	if err.Error() != expectedError {
		t.Fatalf(
			"Expected error %q, got %q",
			expectedError,
			err.Error(),
		)
	}

	fmt.Printf(
		"Max tool-call limit correctly enforced: %v\n",
		err,
	)

	fmt.Println("Orchestrator Max Tool Calls tests completed!")
}
