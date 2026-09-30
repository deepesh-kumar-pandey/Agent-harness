package mcp

import (
	"context"
	"fmt"

	toolspkg "agent-harness/internal/tools"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Runtime struct {
	sessions map[string]*mcpsdk.ClientSession
	tools    map[string][]toolspkg.Tool
}

func NewRuntime() *Runtime {
	return &Runtime{
		sessions: make(map[string]*mcpsdk.ClientSession),
		tools:    make(map[string][]toolspkg.Tool),
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
		return fmt.Errorf(
			"failed to connect to MCP server %q: %w",
			name,
			err,
		)
	}

	tools, err := RegisterTools(ctx, session, registry)
	if err != nil {
		_ = session.Close()

		return fmt.Errorf(
			"failed to register tools from MCP server %q: %w",
			name,
			err,
		)
	}

	r.sessions[name] = session
	r.tools[name] = tools

	return nil
}

func (r *Runtime) DisconnectServer(
	name string,
	registry *toolspkg.ToolRegistry,
) error {
	session, exists := r.sessions[name]
	if !exists {
		return fmt.Errorf(
			"MCP server %q is not connected",
			name,
		)
	}

	for _, tool := range r.tools[name] {
		registeredTool, err := registry.Get(tool.Name())
		if err != nil {
			continue
		}

		if registeredTool == tool {
			_ = registry.Remove(tool.Name())
		}
	}

	if err := session.Close(); err != nil {
		return fmt.Errorf(
			"failed to disconnect MCP server %q: %w",
			name,
			err,
		)
	}

	delete(r.sessions, name)
	delete(r.tools, name)

	return nil
}

func (r *Runtime) Close(
	registry *toolspkg.ToolRegistry,
) error {
	var firstErr error

	for name, session := range r.sessions {
		for _, tool := range r.tools[name] {
			registeredTool, err := registry.Get(tool.Name())
			if err != nil {
				continue
			}

			if registeredTool == tool {
				_ = registry.Remove(tool.Name())
			}
		}

		if err := session.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	r.sessions = make(map[string]*mcpsdk.ClientSession)
	r.tools = make(map[string][]toolspkg.Tool)

	return firstErr
}
