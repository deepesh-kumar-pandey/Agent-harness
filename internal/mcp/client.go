package mcp

import (
	"context"
	"os/exec"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Client struct {
	client *mcpsdk.Client
}

func NewClient() *Client {
	return &Client{
		client: mcpsdk.NewClient(
			&mcpsdk.Implementation{
				Name:    "agent-harness",
				Version: "0.1.0",
			},
			nil,
		),
	}
}

func (c *Client) Connect(
	ctx context.Context,
	transport mcpsdk.Transport,
) (*mcpsdk.ClientSession, error) {
	return c.client.Connect(ctx, transport, nil)
}

func (c *Client) ConnectCommand(
	ctx context.Context,
	command string,
	args []string,
) (*mcpsdk.ClientSession, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	transport := &mcpsdk.CommandTransport{
		Command: cmd,
	}

	return c.Connect(ctx, transport)
}

func (c *Client) ListTools(
	ctx context.Context,
	session *mcpsdk.ClientSession,
) ([]*mcpsdk.Tool, error) {
	result, err := session.ListTools(ctx, &mcpsdk.ListToolsParams{})
	if err != nil {
		return nil, err
	}

	return result.Tools, nil
}
