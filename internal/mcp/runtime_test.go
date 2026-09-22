package mcp

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	toolspkg "agent-harness/internal/tools"
)

// Tests creating a new MCP runtime.
func TestNewRuntime(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "creates runtime",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("creating MCP runtime")

			runtime := NewRuntime()

			if runtime == nil {
				t.Fatal("expected runtime, got nil")
			}

			if runtime.sessions == nil {
				t.Fatal("expected sessions slice, got nil")
			}

			if len(runtime.sessions) != 0 {
				t.Fatalf("expected 0 sessions, got %d", len(runtime.sessions))
			}

			t.Log("MCP runtime created successfully")
		})
	}
}

// Tests connecting an MCP server and registering its tools.
func TestRuntimeConnectServer(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine current test file")
	}

	projectRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")
	command := filepath.Join(projectRoot, "cmd", "mcp-test-server", "mcp-test-server")

	if _, err := os.Stat(command); err != nil {
		t.Fatalf("test MCP server executable not found: %v", err)
	}

	ctx := context.Background()
	registry := toolspkg.NewToolRegistry()
	mcpRuntime := NewRuntime()

	err := mcpRuntime.ConnectServer(
		ctx,
		"test-server",
		command,
		[]string{},
		registry,
	)
	if err != nil {
		t.Fatalf("expected server connection to succeed, got error: %v", err)
	}

	if len(mcpRuntime.sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(mcpRuntime.sessions))
	}

	if !registry.Has("test_tool") {
		t.Fatal("expected test_tool to be registered")
	}

	if err := mcpRuntime.Close(); err != nil {
		t.Fatalf("failed to close runtime: %v", err)
	}

	if len(mcpRuntime.sessions) != 0 {
		t.Fatalf("expected 0 sessions after close, got %d", len(mcpRuntime.sessions))
	}
}

// Tests connecting multiple MCP servers.
func TestRuntimeConnectMultipleServers(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine current test file")
	}

	projectRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")
	command := filepath.Join(projectRoot, "cmd", "mcp-test-server", "mcp-test-server")

	if _, err := os.Stat(command); err != nil {
		t.Fatalf("test MCP server executable not found: %v", err)
	}

	ctx := context.Background()
	registry := toolspkg.NewToolRegistry()
	mcpRuntime := NewRuntime()

	err := mcpRuntime.ConnectServer(
		ctx,
		"test-server-1",
		command,
		[]string{},
		registry,
	)
	if err != nil {
		t.Fatalf("expected first server connection to succeed, got error: %v", err)
	}

	err = mcpRuntime.ConnectServer(
		ctx,
		"test-server-2",
		command,
		[]string{},
		registry,
	)
	if err != nil {
		t.Fatalf("expected second server connection to succeed, got error: %v", err)
	}

	if len(mcpRuntime.sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(mcpRuntime.sessions))
	}

	if !registry.Has("test_tool") {
		t.Fatal("expected test_tool to be registered")
	}

	if err := mcpRuntime.Close(); err != nil {
		t.Fatalf("failed to close runtime: %v", err)
	}

	if len(mcpRuntime.sessions) != 0 {
		t.Fatalf("expected 0 sessions after close, got %d", len(mcpRuntime.sessions))
	}
}

// Tests handling an invalid MCP server command.
func TestRuntimeConnectServerInvalidCommand(t *testing.T) {
	ctx := context.Background()
	registry := toolspkg.NewToolRegistry()
	mcpRuntime := NewRuntime()

	err := mcpRuntime.ConnectServer(
		ctx,
		"invalid-server",
		"invalid-mcp-command",
		[]string{},
		registry,
	)

	if err == nil {
		t.Fatal("expected connection error, got nil")
	}

	if len(mcpRuntime.sessions) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(mcpRuntime.sessions))
	}
}

// Tests closing an MCP runtime.
func TestRuntimeClose(t *testing.T) {
	mcpRuntime := NewRuntime()

	if err := mcpRuntime.Close(); err != nil {
		t.Fatalf("expected close to succeed, got error: %v", err)
	}

	if mcpRuntime.sessions != nil {
		t.Fatal("expected sessions to be nil after close")
	}
}
