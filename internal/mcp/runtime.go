package mcp

import (
	"context"
	"fmt"

	toolspkg "agent-harness/internal/tools"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Runtime struct {
	sessions []*mcpsdk.ClientSession
}

func NewRuntime() *Runtime {
	return &Runtime{
		sessions: make([]*mcpsdk.ClientSession, 0),
	}
}

func (r *Runtime) ConnectServer(
	ctx context.Context,
	name string,
	command string,
	args []string,
	registry *toolspkg.ToolRegistry,
) error {
	client := NewClient()

	session, err := client.ConnectCommand(ctx, command, args)
	if err != nil {
		return fmt.Errorf("failed to connect to MCP server %q: %w", name, err)
	}

	if err := RegisterTools(ctx, session, registry); err != nil {
		_ = session.Close()
		return fmt.Errorf("failed to register tools from MCP server %q: %w", name, err)
	}

	r.sessions = append(r.sessions, session)

	return nil
}

func (r *Runtime) Close() error {
	var firstErr error

	for _, session := range r.sessions {
		if err := session.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	r.sessions = nil

	return firstErr
}
