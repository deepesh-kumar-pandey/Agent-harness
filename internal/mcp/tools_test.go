package mcp

import (
	"context"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestNewTool(t *testing.T) {
	testCases := []struct {
		name string
		tool mcpsdk.Tool
	}{
		{
			name: "creates tool",
			tool: mcpsdk.Tool{
				Name:        "test-tool",
				Description: "test-tool",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("Creating MCP tool")

			tool := NewTool(testCase.tool)

			if tool == nil {
				t.Fatalf("expected tool %q, got nil", testCase.tool.Name)
			}

			if tool.tool.Name != testCase.tool.Name {
				t.Fatalf(
					"expected tool name %q, got %q",
					testCase.tool.Name,
					tool.tool.Name,
				)
			}

			t.Log("MCP tool created successfully")
		})
	}
}

// Tool Name and Description unit test
func TestToolNameAndDescription(t *testing.T) {
	testCases := []struct {
		name     string
		tool     mcpsdk.Tool
		wantName string
		wantDesc string
	}{
		{
			name: "returns tool information",
			tool: mcpsdk.Tool{
				Name:        "test-tool",
				Description: "test description",
			},
			wantName: "test-tool",
			wantDesc: "test description",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("Creating MCP tool")

			tool := NewTool(testCase.tool)

			if got := tool.Name(); got != testCase.wantName {
				t.Fatalf(
					"expected name %q, got %q",
					testCase.wantName,
					got,
				)
			}

			if got := tool.Description(); got != testCase.wantDesc {
				t.Fatalf(
					"expected description %q, got %q",
					testCase.wantDesc,
					got,
				)
			}

			t.Log("MCP tool information retrieved successfully")
		})
	}
}

// Tool Execute unit test
func TestToolExecute(t *testing.T) {
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
						Text: "tool executed successfully",
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

	result, err := tool.Execute(ctx, session, map[string]any{})
	if err != nil {
		t.Fatalf("expected tool execution to succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("expected tool result, got nil")
	}

	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}

	textContent, ok := result.Content[0].(*mcpsdk.TextContent)
	if !ok {
		t.Fatalf("expected text content, got %T", result.Content[0])
	}

	if textContent.Text != "tool executed successfully" {
		t.Fatalf(
			"expected result %q, got %q",
			"tool executed successfully",
			textContent.Text,
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
}

// Tool Schema unit test
func TestToolSchema(t *testing.T) {
	testCases := []struct {
		name        string
		tool        mcpsdk.Tool
		expectNil   bool
		expectedKey string
		expectedVal any
	}{
		{
			name: "returns input schema",
			tool: mcpsdk.Tool{
				Name:        "test-tool",
				Description: "test description",
				InputSchema: &jsonschema.Schema{
					Type: "object",
				},
			},
			expectNil:   false,
			expectedKey: "type",
			expectedVal: "object",
		},
		{
			name: "returns nil when input schema is nil",
			tool: mcpsdk.Tool{
				Name:        "test-tool",
				Description: "test description",
			},
			expectNil: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("Creating MCP tool")

			tool := NewTool(testCase.tool)
			schema := tool.Schema()

			if testCase.expectNil {
				if schema != nil {
					t.Fatalf("expected nil schema, got %v", schema)
				}
				return
			}

			if schema == nil {
				t.Fatal("expected schema, got nil")
			}

			got, exists := schema[testCase.expectedKey]
			if !exists {
				t.Fatalf(
					"expected schema to contain key %q",
					testCase.expectedKey,
				)
			}

			if got != testCase.expectedVal {
				t.Fatalf(
					"expected schema value %v, got %v",
					testCase.expectedVal,
					got,
				)
			}

			t.Log("MCP tool schema retrieved successfully")
		})
	}
}
