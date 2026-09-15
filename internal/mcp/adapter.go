package mcp

import (
	"context"

	toolspkg "agent-harness/internal/tools"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type ToolAdapter struct {
	tool    *Tool
	session *mcpsdk.ClientSession
	ctx     context.Context
}

func NewToolAdapter(
	ctx context.Context,
	tool *Tool,
	session *mcpsdk.ClientSession,
) *ToolAdapter {
	return &ToolAdapter{
		tool:    tool,
		session: session,
		ctx:     ctx,
	}
}

func (t *ToolAdapter) Name() string {
	return t.tool.Name()
}

func (t *ToolAdapter) Description() string {
	return t.tool.Description()
}

func (t *ToolAdapter) Schema() map[string]any {
	return t.tool.Schema()
}

func (t *ToolAdapter) Execute(args map[string]any) (any, error) {
	return t.tool.Execute(t.ctx, t.session, args)
}

var _ toolspkg.Tool = (*ToolAdapter)(nil)

func RegisterTools(
	ctx context.Context,
	session *mcpsdk.ClientSession,
	registry *toolspkg.ToolRegistry,
) error {
	client := NewClient()

	mcpTools, err := client.ListTools(ctx, session)
	if err != nil {
		return err
	}

	for _, mcpTool := range mcpTools {
		tool := NewTool(*mcpTool)

		adapter := NewToolAdapter(ctx, tool, session)

		registry.Register(adapter.Name(), adapter)
	}

	return nil
}
