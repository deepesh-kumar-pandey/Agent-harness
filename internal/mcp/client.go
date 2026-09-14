package mcp

import (
	"context"

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
