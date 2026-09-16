package mcp

import (
	"context"
	"encoding/json"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Tool struct {
	tool mcpsdk.Tool
}

func NewTool(tool mcpsdk.Tool) *Tool {
	return &Tool{
		tool: tool,
	}
}

func (t *Tool) Name() string {
	return t.tool.Name
}

func (t *Tool) Description() string {
	return t.tool.Description
}

func (t *Tool) Execute(
	ctx context.Context,
	session *mcpsdk.ClientSession,
	args map[string]any,
) (*mcpsdk.CallToolResult, error) {
	return session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name:      t.tool.Name,
		Arguments: args,
	})
}

func (t *Tool) Schema() map[string]any {
	if t.tool.InputSchema == nil {
		return nil
	}

	data, err := json.Marshal(t.tool.InputSchema)
	if err != nil {
		return nil
	}

	var schema map[string]any

	if err := json.Unmarshal(data, &schema); err != nil {
		return nil
	}

	return schema
}
