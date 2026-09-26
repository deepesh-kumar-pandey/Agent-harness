package main

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	server := mcpsdk.NewServer(
		&mcpsdk.Implementation{
			Name:    "agent-harness-test-server",
			Version: "0.1.0",
		},
		nil,
	)

	mcpsdk.AddTool(
		server,
		&mcpsdk.Tool{
			Name:        "test_tool",
			Description: "A test tool for MCP command transport",
		},
		func(
			ctx context.Context,
			req *mcpsdk.CallToolRequest,
			args struct{},
		) (*mcpsdk.CallToolResult, any, error) {
			return &mcpsdk.CallToolResult{
				Content: []mcpsdk.Content{
					&mcpsdk.TextContent{
						Text: "MCP test tool executed successfully",
					},
				},
			}, nil, nil
		},
	)

	mcpsdk.AddTool(
		server,
		&mcpsdk.Tool{
			Name:        "error_tool",
			Description: "A test tool that always returns an error",
		},
		func(
			ctx context.Context,
			req *mcpsdk.CallToolRequest,
			args struct{},
		) (*mcpsdk.CallToolResult, any, error) {
			return nil, nil, fmt.Errorf("MCP test error")
		},
	)

	if err := server.Run(
		context.Background(),
		&mcpsdk.StdioTransport{},
	); err != nil {
		panic(err)
	}
}
