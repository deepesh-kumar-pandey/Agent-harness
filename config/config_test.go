package config

import (
	"os"
	"testing"
)

// TestLoadConfig verifies that a configuration can be loaded from a JSON file.
func TestLoadConfig(t *testing.T) {
	validConfig := `{
		"provider": {
			"name": "ollama",
			"model": "llama3.1",
			"base_url": "http://localhost:11434",
			"endpoint": "/api/chat"
		}
	}`

	testCases := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "Valid config",
			path:        "config.json",
			expectError: false,
		},
		{
			name:        "Config file does not exist",
			path:        "nonexistent.json",
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.name == "Valid config" {
				if err := os.WriteFile(testCase.path, []byte(validConfig), 0600); err != nil {
					t.Fatalf("failed to create test config: %v", err)
				}
				defer os.Remove(testCase.path)
			}

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
				t.Errorf("expected a valid config")
			}
		})
	}
}

// TestSaveConfig verifies that a configuration can be saved and loaded back correctly.
func TestSaveConfig(t *testing.T) {
	testConfig := &Config{
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
					Args:    []string{"--port", "8080"},
				},
			},
		},
	}

	testPath := "saved-config.json"
	defer os.Remove(testPath)

	if err := Save(testPath, testConfig); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	savedConfig, err := Load(testPath)
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	if savedConfig.Provider.Name != testConfig.Provider.Name {
		t.Errorf(
			"expected provider name %q, got %q",
			testConfig.Provider.Name,
			savedConfig.Provider.Name,
		)
	}

	if len(savedConfig.MCP.Servers) != 1 {
		t.Fatalf(
			"expected 1 MCP server, got %d",
			len(savedConfig.MCP.Servers),
		)
	}

	server := savedConfig.MCP.Servers[0]

	if server.Name != "example" {
		t.Errorf(
			"expected server name %q, got %q",
			"example",
			server.Name,
		)
	}

	if server.Command != "example-mcp-server" {
		t.Errorf(
			"expected command %q, got %q",
			"example-mcp-server",
			server.Command,
		)
	}

	if len(server.Args) != 2 ||
		server.Args[0] != "--port" ||
		server.Args[1] != "8080" {
		t.Errorf("unexpected server args: %v", server.Args)
	}
}

// TestValidateConfig verifies that valid and invalid configurations are handled correctly.
func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
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
