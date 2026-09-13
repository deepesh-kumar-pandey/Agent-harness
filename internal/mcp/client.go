package mcp

import (
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
