package mcp

import (
	"context"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestNewClient(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "creates client",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("creating MCP client")

			client := NewClient()

			if client == nil {
				t.Fatal("expected client, got nil")
			}

			if client.client == nil {
				t.Fatal("expected SDK client, got nil")
			}

			t.Log("MCP client created successfully")
		})
	}
}

func TestClientConnect(t *testing.T) {
	clientTransport, serverTransport := mcpsdk.NewInMemoryTransports()

	server := mcpsdk.NewServer(
		&mcpsdk.Implementation{
			Name:    "test-server",
			Version: "0.1.0",
		},
		nil,
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

	if session == nil {
		t.Fatal("expected client session, got nil")
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

// Unit test for list tool function
func TestClientListTools(t *testing.T) {
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
			Name:        "test_tool",
			Description: "test MCP tool",
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

	tools, err := client.ListTools(ctx, session)
	if err != nil {
		t.Fatalf("expected tool listing to succeed, got error: %v", err)
	}

	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}

	if tools[0].Name != "test_tool" {
		t.Fatalf("expected tool name %q, got %q", "test_tool", tools[0].Name)
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
