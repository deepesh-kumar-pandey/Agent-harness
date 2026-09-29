package mcp

import (
	"context"
	"fmt"

	toolspkg "agent-harness/internal/tools"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Runtime struct {
	sessions map[string]*mcpsdk.ClientSession
}

func NewRuntime() *Runtime {
	return &Runtime{
		sessions: make(map[string]*mcpsdk.ClientSession),
	}
}

func (r *Runtime) ConnectServer(
	ctx context.Context,
	name string,
	command string,
	args []string,
	registry *toolspkg.ToolRegistry,
) error {
	if _, exists := r.sessions[name]; exists {
		return fmt.Errorf("MCP server %q is already connected", name)
	}

	client := NewClient()

	session, err := client.ConnectCommand(ctx, command, args)
	if err != nil {
		return fmt.Errorf("failed to connect to MCP server %q: %w", name, err)
	}

	if err := RegisterTools(ctx, session, registry); err != nil {
		_ = session.Close()
		return fmt.Errorf("failed to register tools from MCP server %q: %w", name, err)
	}

	r.sessions[name] = session

	return nil
}

func (r *Runtime) DisconnectServer(name string) error {
	session, exists := r.sessions[name]
	if !exists {
		return fmt.Errorf("MCP server %q is not connected", name)
	}

	if err := session.Close(); err != nil {
		return fmt.Errorf("failed to disconnect MCP server %q: %w", name, err)
	}

	delete(r.sessions, name)

	return nil
}

func (r *Runtime) Close() error {
	var firstErr error

	for _, session := range r.sessions {
		if err := session.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	r.sessions = make(map[string]*mcpsdk.ClientSession)

	return firstErr
}
