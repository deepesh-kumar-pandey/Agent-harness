package mcp

import (
	"testing"

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
