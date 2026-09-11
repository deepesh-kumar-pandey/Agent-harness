package orchestrator

import (
	"fmt"
	"strings"
	"testing"

	agentpkg "agent-harness/internal/agent"
	providerpkg "agent-harness/internal/provider"
	toolspkg "agent-harness/internal/tools"
)

// ─────────────────────────────────────────────
// Fake Provider
// ─────────────────────────────────────────────

type FakeProvider struct {
	Responses   []string
	ToolCalls   [][]providerpkg.ToolCall
	Index       int
	LastRequest providerpkg.ChatRequest
}

func (f *FakeProvider) Chat(
	request providerpkg.ChatRequest,
) (providerpkg.ChatResponse, error) {

	if f.Index >= len(f.Responses) {
		return providerpkg.ChatResponse{}, fmt.Errorf(
			"fake provider has no more responses",
		)
	}

	f.LastRequest = request

	response := providerpkg.ChatResponse{
		Content: f.Responses[f.Index],
	}

	if f.Index < len(f.ToolCalls) {
		response.ToolCalls = f.ToolCalls[f.Index]
	}

	f.Index++

	return response, nil
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
// Test Orchestrator AssignTool Unknown Tool
// ─────────────────────────────────────────────

func TestOrchestratorAssignToolUnknownTool(t *testing.T) {

	fmt.Println("Starting Orchestrator unknown tool tests...")

	registry := toolspkg.NewToolRegistry()
	testAgent := agentpkg.NewAgent(registry)

	testOrchestrator := NewOrchestrator(
		testAgent,
		&FakeProvider{},
	)

	testCases := []struct {
		name        string
		toolCall    ToolCall
		expectError bool
	}{
		{
			name: "Unknown tool",
			toolCall: ToolCall{
				Tool: "unknown",
				Args: map[string]any{},
			},
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			_, err := testOrchestrator.AssignTool(
				testCase.toolCall,
			)

			if testCase.expectError && err == nil {
				t.Fatalf("expected error, got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if err != nil && !strings.Contains(
				err.Error(),
				testCase.toolCall.Tool,
			) {
				t.Fatalf(
					"expected error to contain %q, got %q",
					testCase.toolCall.Tool,
					err.Error(),
				)
			}
		})
	}

	fmt.Println("Orchestrator unknown tool tests completed!")
}

// ─────────────────────────────────────────────
// Test Orchestrator AssignTool Execution Error
// ─────────────────────────────────────────────

func TestOrchestratorAssignToolExecutionError(t *testing.T) {

	fmt.Println("Starting Orchestrator tool execution error tests...")

	registry := toolspkg.NewToolRegistry()
	testAgent := agentpkg.NewAgent(registry)

	testOrchestrator := NewOrchestrator(
		testAgent,
		&FakeProvider{},
	)

	testCases := []struct {
		name        string
		toolCall    ToolCall
		expectError bool
	}{
		{
			name: "Calculator division by zero",
			toolCall: ToolCall{
				Tool: "calculator",
				Args: map[string]any{
					"operation": "divide",
					"numbers":   []any{10.0, 0.0},
				},
			},
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			_, err := testOrchestrator.AssignTool(
				testCase.toolCall,
			)

			if testCase.expectError && err == nil {
				t.Fatalf("expected error, got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if err != nil && !strings.Contains(
				err.Error(),
				testCase.toolCall.Tool,
			) {
				t.Fatalf(
					"expected error to contain %q, got %q",
					testCase.toolCall.Tool,
					err.Error(),
				)
			}
		})
	}

	fmt.Println("Orchestrator tool execution error tests completed!")
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
// Test Orchestrator Native Tool Call
// ─────────────────────────────────────────────

func TestOrchestratorRunAgentNativeToolCall(t *testing.T) {

	fmt.Println("Starting Orchestrator native tool-call tests...")

	registry := toolspkg.NewToolRegistry()
	testAgent := agentpkg.NewAgent(registry)

	fakeProvider := &FakeProvider{
		Responses: []string{
			"",
			"30",
		},
		ToolCalls: [][]providerpkg.ToolCall{
			{
				{
					Name: "calculator",
					Arguments: map[string]any{
						"operation": "add",
						"numbers":   []any{10.0, 20.0},
					},
				},
			},
			nil,
		},
	}

	testOrchestrator := NewOrchestrator(
		testAgent,
		fakeProvider,
	)

	request := providerpkg.ChatRequest{
		Model: "test-model",
		Messages: []providerpkg.Message{
			{
				Role:    "user",
				Content: "Calculate 10 + 20",
			},
		},
	}

	response, err := testOrchestrator.RunAgent(request)

	if err != nil {
		t.Fatalf(
			"expected no error, got: %v",
			err,
		)
	}

	if response.Content != "30" {
		t.Fatalf(
			"expected content %q, got %q",
			"30",
			response.Content,
		)
	}

	if response.ToolCall != nil {
		t.Fatalf(
			"expected no tool call after execution",
		)
	}

	fmt.Printf(
		"Native tool call executed successfully: %s\n",
		response.Content,
	)

	fmt.Println("Orchestrator native tool-call tests completed!")
}

// ─────────────────────────────────────────────
// Test Orchestrator Multiple Native Tool Calls
// ─────────────────────────────────────────────

func TestOrchestratorRunAgentMultipleNativeToolCalls(t *testing.T) {

	fmt.Println("Starting Orchestrator multiple native tool-call tests...")

	registry := toolspkg.NewToolRegistry()
	testAgent := agentpkg.NewAgent(registry)

	fakeProvider := &FakeProvider{
		Responses: []string{
			"",
			"Both calculations completed",
		},
		ToolCalls: [][]providerpkg.ToolCall{
			{
				{
					Name: "calculator",
					Arguments: map[string]any{
						"operation": "add",
						"numbers":   []any{10.0, 20.0},
					},
				},
				{
					Name: "calculator",
					Arguments: map[string]any{
						"operation": "multiply",
						"numbers":   []any{5.0, 6.0},
					},
				},
			},
			nil,
		},
	}

	testOrchestrator := NewOrchestrator(
		testAgent,
		fakeProvider,
	)

	request := providerpkg.ChatRequest{
		Model: "test-model",
		Messages: []providerpkg.Message{
			{
				Role:    "user",
				Content: "Calculate 10 + 20 and 5 * 6",
			},
		},
	}

	response, err := testOrchestrator.RunAgent(request)

	if err != nil {
		t.Fatalf(
			"expected no error, got: %v",
			err,
		)
	}

	if response.Content != "Both calculations completed" {
		t.Fatalf(
			"expected content %q, got %q",
			"Both calculations completed",
			response.Content,
		)
	}

	if response.ToolCall != nil {
		t.Fatalf(
			"expected no tool call after execution",
		)
	}

	// The conversation should contain:
	// 1 user message
	// 1 assistant message containing both tool calls
	// 2 tool result messages
	if len(fakeProvider.LastRequest.Messages) != 4 {
		t.Fatalf(
			"expected 4 messages, got %d",
			len(fakeProvider.LastRequest.Messages),
		)
	}

	assistantMessage := fakeProvider.LastRequest.Messages[1]

	if assistantMessage.Role != "assistant" {
		t.Fatalf(
			"expected assistant message, got %q",
			assistantMessage.Role,
		)
	}

	if len(assistantMessage.ToolCalls) != 2 {
		t.Fatalf(
			"expected 2 tool calls, got %d",
			len(assistantMessage.ToolCalls),
		)
	}

	firstToolResult := fakeProvider.LastRequest.Messages[2]

	if firstToolResult.Role != "tool" {
		t.Fatalf(
			"expected first tool result message, got %q",
			firstToolResult.Role,
		)
	}

	if firstToolResult.Content != "30" {
		t.Fatalf(
			"expected first tool result %q, got %q",
			"30",
			firstToolResult.Content,
		)
	}

	secondToolResult := fakeProvider.LastRequest.Messages[3]

	if secondToolResult.Role != "tool" {
		t.Fatalf(
			"expected second tool result message, got %q",
			secondToolResult.Role,
		)
	}

	if secondToolResult.Content != "30" {
		t.Fatalf(
			"expected second tool result %q, got %q",
			"30",
			secondToolResult.Content,
		)
	}

	fmt.Printf(
		"Multiple native tool calls executed successfully: %s\n",
		response.Content,
	)

	fmt.Println(
		"Orchestrator multiple native tool-call tests completed!",
	)
}

// ─────────────────────────────────────────────
// Test Orchestrator Multiple Native Tool Calls With Failure
// ─────────────────────────────────────────────

func TestOrchestratorRunAgentMultipleNativeToolCallsWithFailure(t *testing.T) {

	fmt.Println("Starting Orchestrator multiple native tool-call failure tests...")

	registry := toolspkg.NewToolRegistry()
	testAgent := agentpkg.NewAgent(registry)

	fakeProvider := &FakeProvider{
		Responses: []string{
			"",
		},
		ToolCalls: [][]providerpkg.ToolCall{
			{
				{
					Name: "calculator",
					Arguments: map[string]any{
						"operation": "add",
						"numbers":   []any{10.0, 20.0},
					},
				},
				{
					Name: "calculator",
					Arguments: map[string]any{
						"operation": "divide",
						"numbers":   []any{10.0, 0.0},
					},
				},
			},
		},
	}

	testOrchestrator := NewOrchestrator(
		testAgent,
		fakeProvider,
	)

	request := providerpkg.ChatRequest{
		Model: "test-model",
		Messages: []providerpkg.Message{
			{
				Role:    "user",
				Content: "Calculate 10 + 20 and 10 / 0",
			},
		},
	}

	_, err := testOrchestrator.RunAgent(request)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "calculator") {
		t.Fatalf(
			"expected error to contain %q, got %q",
			"calculator",
			err.Error(),
		)
	}

	fmt.Printf(
		"Multiple native tool-call failure handled successfully: %v\n",
		err,
	)

	fmt.Println(
		"Orchestrator multiple native tool-call failure tests completed!",
	)
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

// ─────────────────────────────────────────────
// Test Orchestrator Get Tool Schemas
// ─────────────────────────────────────────────

func TestOrchestratorGetToolSchemas(t *testing.T) {

	testCases := []struct {
		name          string
		registry      *toolspkg.ToolRegistry
		expectError   bool
		expectedCount int
	}{
		{
			name:          "Get registered tool schemas",
			registry:      toolspkg.NewToolRegistry(),
			expectError:   false,
			expectedCount: 3,
		},
		{
			name:        "Nil registry",
			registry:    nil,
			expectError: true,
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {

			testAgent := agentpkg.NewAgent(testCase.registry)

			orchestrator := NewOrchestrator(
				testAgent,
				&FakeProvider{},
			)

			schemas, err := orchestrator.GetToolSchemas()

			if testCase.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if schemas != nil {
					t.Fatalf(
						"expected no schemas, got %v",
						schemas,
					)
				}

				return
			}

			if err != nil {
				t.Fatalf(
					"expected no error, got: %v",
					err,
				)
			}

			if len(schemas) != testCase.expectedCount {
				t.Fatalf(
					"expected %d schemas, got %d",
					testCase.expectedCount,
					len(schemas),
				)
			}
		})
	}
}

// ─────────────────────────────────────────────
// Test Orchestrator Get Tool Definitions
// ─────────────────────────────────────────────

func TestOrchestratorGetToolDefinitions(t *testing.T) {

	testCases := []struct {
		name          string
		registry      *toolspkg.ToolRegistry
		expectError   bool
		expectedCount int
	}{
		{
			name:          "Get registered tool definitions",
			registry:      toolspkg.NewToolRegistry(),
			expectError:   false,
			expectedCount: 3,
		},
		{
			name:        "Nil registry",
			registry:    nil,
			expectError: true,
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {

			testAgent := agentpkg.NewAgent(testCase.registry)

			orchestrator := NewOrchestrator(
				testAgent,
				&FakeProvider{},
			)

			definitions, err := orchestrator.GetToolDefinitions()

			if testCase.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if definitions != nil {
					t.Fatalf(
						"expected no definitions, got %v",
						definitions,
					)
				}

				return
			}

			if err != nil {
				t.Fatalf(
					"expected no error, got: %v",
					err,
				)
			}

			if len(definitions) != testCase.expectedCount {
				t.Fatalf(
					"expected %d definitions, got %d",
					testCase.expectedCount,
					len(definitions),
				)
			}
		})
	}
}

// ─────────────────────────────────────────────
// Test Orchestrator RunAgent Tool Definitions
// ─────────────────────────────────────────────

func TestOrchestratorRunAgentToolDefinitions(t *testing.T) {

	registry := toolspkg.NewToolRegistry()
	testAgent := agentpkg.NewAgent(registry)

	fakeProvider := &FakeProvider{
		Responses: []string{
			"Hello",
		},
	}

	testOrchestrator := NewOrchestrator(
		testAgent,
		fakeProvider,
	)

	request := providerpkg.ChatRequest{
		Model: "test-model",
		Messages: []providerpkg.Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	response, err := testOrchestrator.RunAgent(request)

	if err != nil {
		t.Fatalf(
			"expected no error, got: %v",
			err,
		)
	}

	if response.Content != "Hello" {
		t.Fatalf(
			"expected response %q, got %q",
			"Hello",
			response.Content,
		)
	}

	if len(fakeProvider.LastRequest.Tools) != 3 {
		t.Fatalf(
			"expected 3 tool definitions, got %d",
			len(fakeProvider.LastRequest.Tools),
		)
	}

	for _, tool := range fakeProvider.LastRequest.Tools {

		if tool.Name == "" {
			t.Fatal("expected tool definition to have a name")
		}

		if tool.Description == "" {
			t.Fatal("expected tool definition to have a description")
		}

		if tool.Parameters == nil {
			t.Fatal("expected tool definition to have parameters")
		}
	}
}

// ─────────────────────────────────────────────
// Test Orchestrator RunAgent Provider Error
// ─────────────────────────────────────────────

func TestOrchestratorRunAgentProviderError(t *testing.T) {

	fmt.Println("Starting Orchestrator RunAgent provider error tests...")

	registry := toolspkg.NewToolRegistry()
	testAgent := agentpkg.NewAgent(registry)

	fakeProvider := &FakeProvider{
		Responses: []string{},
	}

	testOrchestrator := NewOrchestrator(
		testAgent,
		fakeProvider,
	)

	request := providerpkg.ChatRequest{
		Model: "test-model",
		Messages: []providerpkg.Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	_, err := testOrchestrator.RunAgent(request)

	if err == nil {
		t.Fatal("expected provider error, got nil")
	}

	expectedError := "fake provider has no more responses"

	if !strings.Contains(err.Error(), expectedError) {
		t.Fatalf(
			"expected error to contain %q, got %q",
			expectedError,
			err.Error(),
		)
	}

	fmt.Printf(
		"Provider error correctly propagated: %v\n",
		err,
	)

	fmt.Println(
		"Orchestrator RunAgent provider error tests completed!",
	)
}

// ─────────────────────────────────────────────
// Test Orchestrator RunAgent Provider Exhausted
// ─────────────────────────────────────────────

func TestOrchestratorRunAgentProviderExhausted(t *testing.T) {

	fmt.Println("Starting Orchestrator RunAgent provider exhaustion tests...")

	registry := toolspkg.NewToolRegistry()
	testAgent := agentpkg.NewAgent(registry)

	fakeProvider := &FakeProvider{
		Responses: []string{
			"",
		},
		ToolCalls: [][]providerpkg.ToolCall{
			{
				{
					Name: "calculator",
					Arguments: map[string]any{
						"operation": "add",
						"numbers":   []any{10.0, 20.0},
					},
				},
			},
		},
	}

	testOrchestrator := NewOrchestrator(
		testAgent,
		fakeProvider,
	)

	request := providerpkg.ChatRequest{
		Model: "test-model",
		Messages: []providerpkg.Message{
			{
				Role:    "user",
				Content: "Calculate 10 + 20",
			},
		},
	}

	_, err := testOrchestrator.RunAgent(request)

	if err == nil {
		t.Fatal("expected provider exhaustion error, got nil")
	}

	expectedError := "fake provider has no more responses"

	if !strings.Contains(err.Error(), expectedError) {
		t.Fatalf(
			"expected error to contain %q, got %q",
			expectedError,
			err.Error(),
		)
	}

	fmt.Printf(
		"Provider exhaustion correctly handled: %v\n",
		err,
	)

	fmt.Println(
		"Orchestrator RunAgent provider exhaustion tests completed!",
	)
}

// ─────────────────────────────────────────────
// Test Orchestrator Conversation History
// ─────────────────────────────────────────────

func TestOrchestratorRunAgentConversationHistory(t *testing.T) {

	fmt.Println("Starting Orchestrator conversation history tests...")

	registry := toolspkg.NewToolRegistry()
	testAgent := agentpkg.NewAgent(registry)

	fakeProvider := &FakeProvider{
		Responses: []string{
			"Your name is Alice.",
			"Your name is Alice.",
		},
	}

	testOrchestrator := NewOrchestrator(
		testAgent,
		fakeProvider,
	)

	firstRequest := providerpkg.ChatRequest{
		Model: "test-model",
		Messages: []providerpkg.Message{
			{
				Role:    "user",
				Content: "My name is Alice.",
			},
		},
	}

	firstResponse, err := testOrchestrator.RunAgent(firstRequest)

	if err != nil {
		t.Fatalf(
			"expected no error on first request, got: %v",
			err,
		)
	}

	if firstResponse.Content != "Your name is Alice." {
		t.Fatalf(
			"expected first response %q, got %q",
			"Your name is Alice.",
			firstResponse.Content,
		)
	}

	secondRequest := providerpkg.ChatRequest{
		Model: "test-model",
		Messages: []providerpkg.Message{
			{
				Role:    "user",
				Content: "What is my name?",
			},
		},
	}

	secondResponse, err := testOrchestrator.RunAgent(secondRequest)

	if err != nil {
		t.Fatalf(
			"expected no error on second request, got: %v",
			err,
		)
	}

	if secondResponse.Content != "Your name is Alice." {
		t.Fatalf(
			"expected second response %q, got %q",
			"Your name is Alice.",
			secondResponse.Content,
		)
	}

	messages := fakeProvider.LastRequest.Messages

	if len(messages) != 3 {
		t.Fatalf(
			"expected 3 conversation messages, got %d",
			len(messages),
		)
	}

	expectedMessages := []providerpkg.Message{
		{
			Role:    "user",
			Content: "My name is Alice.",
		},
		{
			Role:    "assistant",
			Content: "Your name is Alice.",
		},
		{
			Role:    "user",
			Content: "What is my name?",
		},
	}

	for _, expectedMessage := range expectedMessages {

		found := false

		for _, actualMessage := range messages {

			if actualMessage.Role == expectedMessage.Role &&
				actualMessage.Content == expectedMessage.Content {

				found = true
				break
			}
		}

		if !found {
			t.Fatalf(
				"expected message %+v to be present in conversation",
				expectedMessage,
			)
		}
	}

	fmt.Println("Conversation history preserved successfully.")
	fmt.Println("Orchestrator conversation history tests completed!")
}
