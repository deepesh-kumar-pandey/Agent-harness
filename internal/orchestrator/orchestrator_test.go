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

type FakeProvider struct{}

func (f *FakeProvider) Chat(
	request providerpkg.ChatRequest,
) (providerpkg.ChatResponse, error) {

	return providerpkg.ChatResponse{
		Content: "Fake response",
	}, nil
}

// ─────────────────────────────────────────────
// Test Orchestrator Run
// ─────────────────────────────────────────────

func TestOrchestratorRun(t *testing.T) {

	fmt.Println("Starting Orchestrator Run tests...")

	registry := toolspkg.NewToolRegistry()
	testAgent := agentpkg.NewAgent(registry)

	fakeProvider := &FakeProvider{}

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

	fakeProvider := &FakeProvider{}

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
