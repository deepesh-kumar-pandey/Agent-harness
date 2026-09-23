package mcp

import (
	"context"
	"testing"

	toolspkg "agent-harness/internal/tools"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tool Adapter creation unit test
func TestNewToolAdapter(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "creates tool adapter",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("Creating MCP tool adapter")

			tool := NewTool(mcpsdk.Tool{
				Name:        "test-tool",
				Description: "test description",
			})

			ctx := context.Background()

			adapter := NewToolAdapter(ctx, tool, nil)

			if adapter == nil {
				t.Fatal("expected tool adapter, got nil")
			}

			if adapter.tool != tool {
				t.Fatal("expected adapter to contain MCP tool")
			}

			if adapter.ctx != ctx {
				t.Fatal("expected adapter to contain context")
			}

			t.Log("MCP tool adapter created successfully")
		})
	}
}

// Tool Adapter Name and Description unit test
func TestToolAdapterNameAndDescription(t *testing.T) {
	tool := NewTool(mcpsdk.Tool{
		Name:        "test-tool",
		Description: "test description",
	})

	adapter := NewToolAdapter(context.Background(), tool, nil)

	if got := adapter.Name(); got != "test-tool" {
		t.Fatalf("expected name %q, got %q", "test-tool", got)
	}

	if got := adapter.Description(); got != "test description" {
		t.Fatalf("expected description %q, got %q", "test description", got)
	}

	t.Log("MCP tool adapter information retrieved successfully")
}

// Tool Adapter Schema unit test
func TestToolAdapterSchema(t *testing.T) {
	tool := NewTool(mcpsdk.Tool{
		Name:        "test-tool",
		Description: "test description",
	})

	adapter := NewToolAdapter(context.Background(), tool, nil)

	schema := adapter.Schema()

	if schema != nil {
		t.Fatalf("expected nil schema, got %v", schema)
	}

	t.Log("MCP tool adapter schema retrieved successfully")
}

// Tool Adapter Execute unit test
func TestToolAdapterExecute(t *testing.T) {
	clientTransport, serverTransport := mcpsdk.NewInMemoryTransports()

	server := mcpsdk.NewServer(
		&mcpsdk.Implementation{
			Name:    "test-server",
			Version: "0.1.0",
		},
		nil,
	)

	mcpsdk.AddTool(
		server,
		&mcpsdk.Tool{
			Name:        "test-tool",
			Description: "test tool",
		},
		func(
			ctx context.Context,
			req *mcpsdk.CallToolRequest,
			args struct{},
		) (*mcpsdk.CallToolResult, any, error) {
			return &mcpsdk.CallToolResult{
				Content: []mcpsdk.Content{
					&mcpsdk.TextContent{
						Text: "adapter executed successfully",
					},
				},
			}, nil, nil
		},
	)

	ctx := context.Background()

	serverErr := make(chan error, 1)

	go func() {
		serverErr <- server.Run(ctx, serverTransport)
	}()

	client := NewClient()

	session, err := client.Connect(ctx, clientTransport)
	if err != nil {
		t.Fatalf("expected connection to succeed, got error: %v", err)
	}

	tool := NewTool(mcpsdk.Tool{
		Name:        "test-tool",
		Description: "test tool",
	})

	adapter := NewToolAdapter(ctx, tool, session)

	result, err := adapter.Execute(map[string]any{})
	if err != nil {
		t.Fatalf("expected tool execution to succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("expected tool result, got nil")
	}

	textResult, ok := result.(string)
	if !ok {
		t.Fatalf("expected string, got %T", result)
	}

	if textResult != "adapter executed successfully" {
		t.Fatalf(
			"expected result %q, got %q",
			"adapter executed successfully",
			textResult,
		)
	}

	if err := session.Close(); err != nil {
		t.Fatalf("failed to close session: %v", err)
	}

	select {
	case err := <-serverErr:
		if err != nil {
			t.Fatalf("server returned error: %v", err)
		}
	default:
	}

	t.Log("MCP tool adapter executed successfully")
}

// Register Tools unit test
func TestRegisterTools(t *testing.T) {
	clientTransport, serverTransport := mcpsdk.NewInMemoryTransports()

	server := mcpsdk.NewServer(
		&mcpsdk.Implementation{
			Name:    "test-server",
			Version: "0.1.0",
		},
		nil,
	)

	mcpsdk.AddTool(
		server,
		&mcpsdk.Tool{
			Name:        "test-tool",
			Description: "test tool",
		},
		func(
			ctx context.Context,
			req *mcpsdk.CallToolRequest,
			args struct{},
		) (*mcpsdk.CallToolResult, any, error) {
			return &mcpsdk.CallToolResult{}, nil, nil
		},
	)

	ctx := context.Background()

	serverErr := make(chan error, 1)

	go func() {
		serverErr <- server.Run(ctx, serverTransport)
	}()

	client := NewClient()

	session, err := client.Connect(ctx, clientTransport)
	if err != nil {
		t.Fatalf("expected connection to succeed, got error: %v", err)
	}

	registry := toolspkg.NewToolRegistry()

	if err := RegisterTools(ctx, session, registry); err != nil {
		t.Fatalf("expected tools to register successfully, got error: %v", err)
	}

	if !registry.Has("test-tool") {
		t.Fatal("expected test-tool to be registered")
	}

	tool, err := registry.Get("test-tool")
	if err != nil {
		t.Fatalf("expected registered tool to be retrieved, got error: %v", err)
	}

	if tool.Name() != "test-tool" {
		t.Fatalf("expected tool name %q, got %q", "test-tool", tool.Name())
	}

	if tool.Description() != "test tool" {
		t.Fatalf(
			"expected tool description %q, got %q",
			"test tool",
			tool.Description(),
		)
	}

	if err := session.Close(); err != nil {
		t.Fatalf("failed to close session: %v", err)
	}

	select {
	case err := <-serverErr:
		if err != nil {
			t.Fatalf("server returned error: %v", err)
		}
	default:
	}

	t.Log("MCP tools registered successfully")
}
