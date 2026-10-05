package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadConfig verifies that configurations can be loaded from JSON files.
func TestLoadConfig(t *testing.T) {
	validConfig := `{
		"provider": {
			"name": "ollama",
			"model": "llama3.1",
			"base_url": "http://localhost:11434",
			"endpoint": "/api/chat"
		}
	}`

	invalidJSON := `{
		"provider": {
			"name": "ollama",
			"model": "llama3.1",
	}`

	invalidConfig := `{
		"provider": {
			"name": "ollama"
		}
	}`

	dir := t.TempDir()

	validPath := filepath.Join(dir, "config.json")
	invalidJSONPath := filepath.Join(dir, "invalid.json")
	invalidConfigPath := filepath.Join(dir, "invalid-config.json")

	if err := os.WriteFile(validPath, []byte(validConfig), 0600); err != nil {
		t.Fatalf("failed to create valid config: %v", err)
	}

	if err := os.WriteFile(invalidJSONPath, []byte(invalidJSON), 0600); err != nil {
		t.Fatalf("failed to create invalid JSON config: %v", err)
	}

	if err := os.WriteFile(invalidConfigPath, []byte(invalidConfig), 0600); err != nil {
		t.Fatalf("failed to create invalid config: %v", err)
	}

	testCases := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name: "Valid config",
			path: validPath,
		},
		{
			name:        "Config file does not exist",
			path:        filepath.Join(dir, "nonexistent.json"),
			expectError: true,
		},
		{
			name:        "Invalid JSON",
			path:        invalidJSONPath,
			expectError: true,
		},
		{
			name:        "Config fails validation",
			path:        invalidConfigPath,
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
				t.Fatalf("expected a valid config")
			}

			if config.Provider.Name != "ollama" {
				t.Errorf(
					"expected provider name %q, got %q",
					"ollama",
					config.Provider.Name,
				)
			}

			if config.Provider.Model != "llama3.1" {
				t.Errorf(
					"expected model %q, got %q",
					"llama3.1",
					config.Provider.Model,
				)
			}

			if config.Provider.BaseURL != "http://localhost:11434" {
				t.Errorf(
					"expected base URL %q, got %q",
					"http://localhost:11434",
					config.Provider.BaseURL,
				)
			}

			if config.Provider.Endpoint != "/api/chat" {
				t.Errorf(
					"expected endpoint %q, got %q",
					"/api/chat",
					config.Provider.Endpoint,
				)
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

	testPath := filepath.Join(t.TempDir(), "saved-config.json")

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

	if savedConfig.Provider.Model != testConfig.Provider.Model {
		t.Errorf(
			"expected model %q, got %q",
			testConfig.Provider.Model,
			savedConfig.Provider.Model,
		)
	}

	if savedConfig.Provider.BaseURL != testConfig.Provider.BaseURL {
		t.Errorf(
			"expected base URL %q, got %q",
			testConfig.Provider.BaseURL,
			savedConfig.Provider.BaseURL,
		)
	}

	if savedConfig.Provider.Endpoint != testConfig.Provider.Endpoint {
		t.Errorf(
			"expected endpoint %q, got %q",
			testConfig.Provider.Endpoint,
			savedConfig.Provider.Endpoint,
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

// TestSaveConfigError verifies that saving to an invalid path returns an error.
func TestSaveConfigError(t *testing.T) {
	testConfig := &Config{
		Provider: Provider{
			Name:     "ollama",
			Model:    "llama3.1",
			BaseURL:  "http://localhost:11434",
			Endpoint: "/api/chat",
		},
	}

	path := filepath.Join(t.TempDir(), "missing", "config.json")

	err := Save(path, testConfig)

	if err == nil {
		t.Fatal("expected error, got nil")
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
			name: "Valid Ollama config",
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
			name: "Valid OpenAI config",
			config: Config{
				Provider: Provider{
					Name:     "openai",
					Model:    "gpt-4o",
					BaseURL:  "https://api.openai.com/v1",
					Endpoint: "/chat/completions",
				},
			},
			expectError: false,
		},
		{
			name: "Valid Gemini config",
			config: Config{
				Provider: Provider{
					Name:     "gemini",
					Model:    "gemini-2.5-flash",
					BaseURL:  "https://generativelanguage.googleapis.com/v1beta",
					Endpoint: "/models/gemini-2.5-flash:generateContent",
				},
			},
			expectError: false,
		},
		{
			name: "Valid Anthropic config",
			config: Config{
				Provider: Provider{
					Name:     "anthropic",
					Model:    "claude-sonnet-4-5",
					BaseURL:  "https://api.anthropic.com/v1",
					Endpoint: "/messages",
				},
			},
			expectError: false,
		},
		{
			name: "Valid Mistral config",
			config: Config{
				Provider: Provider{
					Name:     "mistral",
					Model:    "mistral-large-latest",
					BaseURL:  "https://api.mistral.ai/v1",
					Endpoint: "/chat/completions",
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
