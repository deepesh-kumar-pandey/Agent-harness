package tools

import (
	"context"
	"testing"

	mcppkg "agent-harness/internal/mcp"
	providerpkg "agent-harness/internal/provider"
	sessionpkg "agent-harness/internal/session"
	toolspkg "agent-harness/internal/tools"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestNewAgent verifies that an agent is created with the provided registry.
func TestNewAgent(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "Create agent with registry",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := toolspkg.NewToolRegistry()

			agent := NewAgent(registry)

			if agent == nil {
				t.Fatalf("got nil agent in test case: %s", testCase.name)
			}

			if agent.registry != registry {
				t.Fatalf(
					"registry mismatch in test case: %s",
					testCase.name,
				)
			}
		})
	}
}

// TestAgentSetSession verifies that an agent can attach a session.
func TestAgentSetSession(t *testing.T) {
	testCases := []struct {
		name        string
		session     *sessionpkg.Session
		expectError bool
	}{
		{
			name:        "sets session",
			session:     sessionpkg.NewSession("test-session"),
			expectError: false,
		},
		{
			name:        "rejects nil session",
			session:     nil,
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := toolspkg.NewToolRegistry()
			agent := NewAgent(registry)

			err := agent.SetSession(testCase.session)

			if testCase.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if agent.session != testCase.session {
				t.Fatal("expected agent session to match provided session")
			}
		})
	}
}

// TestAgentSessionMessages verifies that messages are stored in the active session.
func TestAgentSessionMessages(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "stores messages in session",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := toolspkg.NewToolRegistry()
			agent := NewAgent(registry)
			session := sessionpkg.NewSession("test-session")

			if err := agent.SetSession(session); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			message := providerpkg.Message{
				Role:    "user",
				Content: "Hello",
			}

			if err := agent.AddMessage(message); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			messages := agent.GetMessages()

			if len(messages) != 1 {
				t.Fatalf("expected 1 message, got %d", len(messages))
			}

			if messages[0].Role != "user" {
				t.Errorf(
					"expected role %q, got %q",
					"user",
					messages[0].Role,
				)
			}

			if messages[0].Content != "Hello" {
				t.Errorf(
					"expected content %q, got %q",
					"Hello",
					messages[0].Content,
				)
			}

			if len(session.Messages) != 1 {
				t.Fatalf(
					"expected session to contain 1 message, got %d",
					len(session.Messages),
				)
			}
		})
	}
}

// TestExecuteTool verifies that registered tools can be executed.
func TestExecuteTool(t *testing.T) {
	testCases := []struct {
		name        string
		toolName    string
		args        map[string]any
		expected    any
		expectError bool
	}{
		{
			name:     "Execute calculator addition",
			toolName: "calculator",
			args: map[string]any{
				"operation": "add",
				"numbers":   []float64{10, 20, 30},
			},
			expected:    float64(60),
			expectError: false,
		},
		{
			name:     "Execute calculator multiplication",
			toolName: "calculator",
			args: map[string]any{
				"operation": "multiply",
				"numbers":   []float64{2, 3, 4},
			},
			expected:    float64(24),
			expectError: false,
		},
		{
			name:     "Execute shell command",
			toolName: "shell",
			args: map[string]any{
				"command": "echo",
				"args":    []string{"hello"},
			},
			expected:    "hello\n",
			expectError: false,
		},
		{
			name:        "Execute unregistered tool",
			toolName:    "unknown",
			args:        map[string]any{},
			expected:    nil,
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := toolspkg.NewToolRegistry()

			agent := NewAgent(registry)

			result, err := agent.ExecuteTool(
				testCase.toolName,
				testCase.args,
			)

			if testCase.expectError {
				if err == nil {
					t.Fatalf(
						"expected error in test case: %s",
						testCase.name,
					)
				}

				if result != nil {
					t.Fatalf(
						"expected nil result, got %v in test case: %s",
						result,
						testCase.name,
					)
				}

				return
			}

			if err != nil {
				t.Fatalf(
					"unexpected error in test case: %v",
					err,
				)
			}

			if result != testCase.expected {
				t.Fatalf(
					"expected %v, got %v in test case: %s",
					testCase.expected,
					result,
					testCase.name,
				)
			}
		})
	}
}

// TestRun verifies that the agent Run method executes registered tools.
func TestRun(t *testing.T) {
	testCases := []struct {
		name        string
		toolName    string
		args        map[string]any
		expected    any
		expectError bool
	}{
		{
			name:     "Run calculator",
			toolName: "calculator",
			args: map[string]any{
				"operation": "add",
				"numbers":   []float64{10, 20},
			},
			expected:    float64(30),
			expectError: false,
		},
		{
			name:     "Run shell",
			toolName: "shell",
			args: map[string]any{
				"command": "echo",
				"args":    []string{"hello"},
			},
			expected:    "hello\n",
			expectError: false,
		},
		{
			name:        "Run unknown tool",
			toolName:    "unknown",
			args:        map[string]any{},
			expected:    nil,
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := toolspkg.NewToolRegistry()
			agent := NewAgent(registry)

			result, err := agent.Run(
				testCase.toolName,
				testCase.args,
			)

			if testCase.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}

				if result != nil {
					t.Fatalf(
						"expected nil result, got %v",
						result,
					)
				}

				return
			}

			if err != nil {
				t.Fatalf(
					"unexpected error: %v",
					err,
				)
			}

			if result != testCase.expected {
				t.Fatalf(
					"expected %v, got %v",
					testCase.expected,
					result,
				)
			}
		})
	}
}

// TestAgentGetToolSchemas verifies that the agent returns tool schemas.
func TestAgentGetToolSchemas(t *testing.T) {
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
			agent := NewAgent(testCase.registry)

			schemas, err := agent.GetToolSchemas()

			if testCase.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if schemas != nil {
					t.Fatalf("expected no schemas, got %v", schemas)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
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

// TestAgentRunMCPTool verifies that MCP tools can be executed by the agent.
func TestAgentRunMCPTool(t *testing.T) {
	ctx := context.Background()

	server := mcpsdk.NewServer(
		&mcpsdk.Implementation{
			Name:    "test-server",
			Version: "1.0.0",
		},
		nil,
	)

	mcpsdk.AddTool(
		server,
		&mcpsdk.Tool{
			Name:        "mcp-test-tool",
			Description: "Test MCP tool",
		},
		func(
			ctx context.Context,
			req *mcpsdk.CallToolRequest,
			args map[string]any,
		) (*mcpsdk.CallToolResult, map[string]any, error) {
			return &mcpsdk.CallToolResult{
					Content: []mcpsdk.Content{
						&mcpsdk.TextContent{
							Text: "MCP tool executed successfully",
						},
					},
				},
				nil,
				nil
		},
	)

	clientTransport, serverTransport := mcpsdk.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("failed to connect MCP server: %v", err)
	}

	client := mcppkg.NewClient()

	clientSession, err := client.Connect(ctx, clientTransport)
	if err != nil {
		t.Fatalf("failed to connect MCP client: %v", err)
	}

	defer clientSession.Close()
	defer serverSession.Close()

	registry := toolspkg.NewToolRegistry()

	err = mcppkg.RegisterTools(
		ctx,
		clientSession,
		registry,
	)
	if err != nil {
		t.Fatalf("failed to register MCP tools: %v", err)
	}

	agent := NewAgent(registry)

	result, err := agent.Run(
		"mcp-test-tool",
		map[string]any{},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	callResult, ok := result.(*mcpsdk.CallToolResult)
	if !ok {
		t.Fatalf(
			"expected *mcpsdk.CallToolResult, got %T",
			result,
		)
	}

	if len(callResult.Content) != 1 {
		t.Fatalf(
			"expected 1 content item, got %d",
			len(callResult.Content),
		)
	}

	textContent, ok := callResult.Content[0].(*mcpsdk.TextContent)
	if !ok {
		t.Fatalf(
			"expected *mcpsdk.TextContent, got %T",
			callResult.Content[0],
		)
	}

	expected := "MCP tool executed successfully"

	if textContent.Text != expected {
		t.Fatalf(
			"expected %q, got %q",
			expected,
			textContent.Text,
		)
	}
}
