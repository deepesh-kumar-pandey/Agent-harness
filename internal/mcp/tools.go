package mcp

import (
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
