package config

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	testCases := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			// Tests loading a valid configuration file.
			name:        "Valid config",
			path:        "config.json",
			expectError: false,
		},
		{
			// Tests handling a configuration file that does not exist.
			name:        "Config file does not exist",
			path:        "nonexistent.json",
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			config, err := Load(testCase.path)

			if testCase.expectError {
				if err == nil {
					t.Errorf("expected an error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if config == nil {
				t.Errorf(
					"Please provide a valid config file for the test case: %s",
					testCase.path,
				)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			// Tests a valid provider configuration.
			name: "Valid config",
			config: Config{
				Provider: Provider{
					Name:     "ollama",
					Model:    "llama3.1",
					BaseURL:  "http://localhost:11434",
					Endpoint: "/api/chat",
				},
			},
			expectError: false,
		},
		{
			// Tests a valid configuration containing an MCP server.
			name: "Valid config with MCP server",
			config: Config{
				Provider: Provider{
					Name:     "ollama",
					Model:    "llama3.1",
					BaseURL:  "http://localhost:11434",
					Endpoint: "/api/chat",
				},
				MCP: MCPConfig{
					Servers: []MCPServer{
						{
							Name:    "example",
							Command: "example-mcp-server",
						},
					},
				},
			},
			expectError: false,
		},
		{
			// Tests validation when the provider name is missing.
			name: "Missing provider name",
			config: Config{
				Provider: Provider{
					Model:    "llama3.1",
					BaseURL:  "http://localhost:11434",
					Endpoint: "/api/chat",
				},
			},
			expectError: true,
		},
		{
			// Tests validation when the provider model is missing.
			name: "Missing model",
			config: Config{
				Provider: Provider{
					Name:     "ollama",
					BaseURL:  "http://localhost:11434",
					Endpoint: "/api/chat",
				},
			},
			expectError: true,
		},
		{
			// Tests validation when the provider base URL is missing.
			name: "Missing base URL",
			config: Config{
				Provider: Provider{
					Name:     "ollama",
					Model:    "llama3.1",
					Endpoint: "/api/chat",
				},
			},
			expectError: true,
		},
		{
			// Tests validation when the provider endpoint is missing.
			name: "Missing endpoint",
			config: Config{
				Provider: Provider{
					Name:    "ollama",
					Model:   "llama3.1",
					BaseURL: "http://localhost:11434",
				},
			},
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.config.Validate()

			if testCase.expectError {
				if err == nil {
					t.Errorf("expected an error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
