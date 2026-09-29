package main

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	configpkg "agent-harness/config"
	agentpkg "agent-harness/internal/agent"
	mcppkg "agent-harness/internal/mcp"
	providerpkg "agent-harness/internal/provider"
	sessionpkg "agent-harness/internal/session"
	toolspkg "agent-harness/internal/tools"
)

type fakeLocalProvider struct {
	models     []string
	listErr    error
	pullErrors []error
	pullCalls  []string
}

func (f *fakeLocalProvider) Chat(
	request providerpkg.ChatRequest,
) (providerpkg.ChatResponse, error) {
	return providerpkg.ChatResponse{}, nil
}

func (f *fakeLocalProvider) ListModels() ([]string, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}

	return f.models, nil
}

func (f *fakeLocalProvider) PullModel(name string) error {
	f.pullCalls = append(f.pullCalls, name)

	if len(f.pullErrors) == 0 {
		return nil
	}

	err := f.pullErrors[0]
	f.pullErrors = f.pullErrors[1:]

	return err
}

type fakeCloudProvider struct{}

func (fakeCloudProvider) Chat(
	request providerpkg.ChatRequest,
) (providerpkg.ChatResponse, error) {
	return providerpkg.ChatResponse{}, nil
}

// Test for resolving the model using the environment variable.
func TestResolveModelUsesEnvOverride(t *testing.T) {
	t.Setenv("OLLAMA_MODEL", "custom-model")

	got, source := resolveModel("config-model")

	if got != "custom-model" || source != "environment" {
		t.Fatalf(
			"resolveModel() = (%q, %q), want (%q, %q)",
			got,
			source,
			"custom-model",
			"environment",
		)
	}
}

// Test for resolving the model using the configuration.
func TestResolveModelUsesConfigModel(t *testing.T) {
	t.Setenv("OLLAMA_MODEL", "")

	got, source := resolveModel("config-model")

	if got != "config-model" || source != "config" {
		t.Fatalf(
			"resolveModel() = (%q, %q), want (%q, %q)",
			got,
			source,
			"config-model",
			"config",
		)
	}
}

// Test for resolving the default model.
func TestResolveModelDefaultsToInstalledModel(t *testing.T) {
	t.Setenv("OLLAMA_MODEL", "")

	got, source := resolveModel("")

	if got != "kirito1/qwen3-coder:4b" || source != "default" {
		t.Fatalf(
			"resolveModel() = (%q, %q), want (%q, %q)",
			got,
			source,
			"kirito1/qwen3-coder:4b",
			"default",
		)
	}
}

// Test for selecting a local model.
func TestSelectModel(t *testing.T) {
	testCases := []struct {
		name           string
		provider       providerpkg.Provider
		configured     string
		input          string
		expected       string
		expectedSource string
		expectError    bool
		expectOutput   string
		expectPulls    int
	}{
		{
			name:           "Configured model is available",
			provider:       &fakeLocalProvider{models: []string{"llama3.1"}},
			configured:     "llama3.1",
			expected:       "llama3.1",
			expectedSource: "config",
		},
		{
			name:           "Selects installed alternative",
			provider:       &fakeLocalProvider{models: []string{"mistral:latest"}},
			configured:     "llama3.1",
			input:          "1\n",
			expected:       "mistral:latest",
			expectedSource: "local-selection",
			expectOutput:   "Available local models:",
		},
		{
			name: "Invalid choice then selects alternative",
			provider: &fakeLocalProvider{
				models: []string{"mistral:latest"},
			},
			configured:     "llama3.1",
			input:          "x\n1\n",
			expected:       "mistral:latest",
			expectedSource: "local-selection",
			expectOutput:   "Invalid choice",
		},
		{
			name: "Pulls configured model",
			provider: &fakeLocalProvider{
				models: []string{},
			},
			configured:     "llama3.1",
			input:          "p\n",
			expected:       "llama3.1",
			expectedSource: "pulled",
			expectOutput:   "Model \"llama3.1\" pulled successfully.",
			expectPulls:    1,
		},
		{
			name: "Retries after pull failure",
			provider: &fakeLocalProvider{
				pullErrors: []error{errors.New("download failed")},
			},
			configured:     "llama3.1",
			input:          "p\np\n",
			expected:       "llama3.1",
			expectedSource: "pulled",
			expectOutput:   "Failed to pull model \"llama3.1\": download failed",
			expectPulls:    2,
		},
		{
			name: "Quit after pull failure",
			provider: &fakeLocalProvider{
				pullErrors: []error{errors.New("download failed")},
			},
			configured:   "llama3.1",
			input:        "p\nq\n",
			expectError:  true,
			expectOutput: "Failed to pull model \"llama3.1\": download failed",
			expectPulls:  1,
		},
		{
			name: "Reports Ollama unavailable",
			provider: &fakeLocalProvider{
				listErr: errors.New("connection refused"),
			},
			configured:   "llama3.1",
			expectError:  true,
			expectOutput: "Unable to connect to Ollama.",
		},
		{
			name:        "Cancels when input ends",
			provider:    &fakeLocalProvider{},
			configured:  "llama3.1",
			input:       "",
			expectError: true,
		},
		{
			name:           "Skips discovery for cloud provider",
			provider:       fakeCloudProvider{},
			configured:     "deepseek-chat",
			input:          "",
			expected:       "deepseek-chat",
			expectedSource: "config",
			expectOutput:   "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var output strings.Builder
			input := bufio.NewScanner(strings.NewReader(testCase.input))

			got, source, err := selectModel(
				testCase.provider,
				testCase.configured,
				"config",
				input,
				&output,
			)

			if testCase.expectError && err == nil {
				t.Fatalf("expected error, got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if !testCase.expectError && got != testCase.expected {
				t.Fatalf(
					"expected model %q, got %q",
					testCase.expected,
					got,
				)
			}

			if !testCase.expectError && source != testCase.expectedSource {
				t.Fatalf(
					"expected source %q, got %q",
					testCase.expectedSource,
					source,
				)
			}

			if testCase.expectOutput != "" &&
				!strings.Contains(output.String(), testCase.expectOutput) {
				t.Fatalf(
					"expected output to contain %q, got %q",
					testCase.expectOutput,
					output.String(),
				)
			}

			if localProvider, ok := testCase.provider.(*fakeLocalProvider); ok {
				if len(localProvider.pullCalls) != testCase.expectPulls {
					t.Fatalf(
						"expected %d pull calls, got %d",
						testCase.expectPulls,
						len(localProvider.pullCalls),
					)
				}

				for _, model := range localProvider.pullCalls {
					if model != testCase.configured {
						t.Fatalf(
							"expected pull call for %q, got %q",
							testCase.configured,
							model,
						)
					}
				}
			}
		})
	}
}

// Test for handling CLI commands.
func TestHandleCommand(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		model          string
		modelSource    string
		expectedResult CommandResult
		expectedOutput string
		mcpServers     []configpkg.MCPServer
	}{
		{
			name:           "Help command",
			input:          "help",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Available commands: help, exit, model, session, clear, mcp",
		},
		{
			name:           "Model command",
			input:          "model",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Current model : test-model\nSource: test-source",
		},
		{
			name:           "Session command",
			input:          "session",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Current session: test-session",
		},
		{
			name:           "Session list command",
			input:          "session list",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Sessions:\n- test-session",
		},
		{
			name:           "Session load command",
			input:          "session load loaded-session",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Loaded session: loaded-session",
		},
		{
			name:           "Session load command missing ID",
			input:          "session load",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Usage: session load <id>",
		},
		{
			name:           "Session load command session not found",
			input:          "session load missing-session",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Failed to load session: session not found: missing-session",
		},
		{
			name:           "Session create command",
			input:          "session create new-session",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Created and switched to session: new-session",
		},
		{
			name:           "Session create command missing ID",
			input:          "session create",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Usage: session create <id>",
		},
		{
			name:           "Session delete command",
			input:          "session delete delete-session",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Deleted session: delete-session",
		},
		{
			name:           "Session delete command missing ID",
			input:          "session delete",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Usage: session delete <id>",
		},
		{
			name:           "Session delete current session",
			input:          "session delete test-session",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Cannot delete current session.",
		},
		{
			name:           "Session delete command session not found",
			input:          "session delete missing-session",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Failed to delete session: session not found: missing-session",
		},
		{
			name:           "MCP list command",
			input:          "mcp list",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "MCP servers:\n- test-server: ./cmd/mcp-test-server/mcp-test-server --port 8080\n- filesystem: npx -y @modelcontextprotocol/server-filesystem /tmp",
			mcpServers: []configpkg.MCPServer{
				{
					Name:    "test-server",
					Command: "./cmd/mcp-test-server/mcp-test-server",
					Args:    []string{"--port", "8080"},
				},
				{
					Name:    "filesystem",
					Command: "npx",
					Args: []string{
						"-y",
						"@modelcontextprotocol/server-filesystem",
						"/tmp",
					},
				},
			},
		},
		{
			name:           "MCP list with no servers",
			input:          "mcp list",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "No MCP servers configured.",
		},
		{
			name:           "MCP command missing subcommand",
			input:          "mcp",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Usage: mcp list | mcp add <name> <command> [args...] | mcp remove <name>",
		},
		{
			name:           "MCP unknown subcommand",
			input:          "mcp test",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Unknown mcp command.",
		},
		{
			name:           "MCP add command",
			input:          "mcp add test-server ./cmd/mcp-test-server/mcp-test-server",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Added MCP server: test-server",
		},
		{
			name:           "MCP add command missing arguments",
			input:          "mcp add test-server",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Usage: mcp add <name> <command> [args...]",
		},
		{
			name:           "MCP add duplicate server",
			input:          "mcp add test-server ./test-server",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "MCP server \"test-server\" already exists.",
			mcpServers: []configpkg.MCPServer{
				{
					Name:    "test-server",
					Command: "./test-server",
				},
			},
		},
		{
			name:           "MCP remove command",
			input:          "mcp remove test-server",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Removed MCP server: test-server",
			mcpServers: []configpkg.MCPServer{
				{
					Name:    "test-server",
					Command: "./test-server",
				},
				{
					Name:    "filesystem",
					Command: "npx",
					Args: []string{
						"-y",
						"@modelcontextprotocol/server-filesystem",
						"/tmp",
					},
				},
			},
		},
		{
			name:           "MCP remove command missing name",
			input:          "mcp remove",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Usage: mcp remove <name>",
		},
		{
			name:           "MCP remove server not found",
			input:          "mcp remove missing-server",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "MCP server \"missing-server\" not found.",
			mcpServers: []configpkg.MCPServer{
				{
					Name:    "test-server",
					Command: "./test-server",
				},
			},
		},
		{
			name:           "Clear command",
			input:          "clear",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "\033[H\033[2J",
		},
		{
			name:           "Exit command",
			input:          "exit",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandExit,
		},
		{
			name:           "Normal input",
			input:          "hello",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandNotHandled,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var output strings.Builder

			sessionStore := sessionpkg.NewSessionStore()

			registry := toolspkg.NewToolRegistry()
			mcpRuntime := mcppkg.NewRuntime()

			agentClient := agentpkg.NewAgent(registry)

			t.Cleanup(func() {
				if err := mcpRuntime.Close(); err != nil {
					t.Fatalf("failed to close MCP runtime: %v", err)
				}
			})

			currentSession := sessionpkg.NewSession("test-session")

			err := sessionStore.Set(currentSession)
			if err != nil {
				t.Fatalf(
					"expected no error while setting session, got %v",
					err,
				)
			}

			if testCase.input == "session load loaded-session" {
				err := sessionStore.Set(
					sessionpkg.NewSession("loaded-session"),
				)

				if err != nil {
					t.Fatalf(
						"expected no error while setting loaded session, got %v",
						err,
					)
				}
			}

			if testCase.input == "session delete delete-session" {
				err := sessionStore.Set(
					sessionpkg.NewSession("delete-session"),
				)

				if err != nil {
					t.Fatalf(
						"expected no error while setting delete session, got %v",
						err,
					)
				}
			}

			err = agentClient.SetSession(currentSession)
			if err != nil {
				t.Fatalf(
					"expected no error while setting agent session, got %v",
					err,
				)
			}

			appConfig := &configpkg.Config{
				Provider: configpkg.Provider{
					Name:     "ollama",
					Model:    "llama3.1",
					BaseURL:  "http://localhost:11434",
					Endpoint: "/api/chat",
				},
				MCP: configpkg.MCPConfig{
					Servers: testCase.mcpServers,
				},
			}

			isMCPAddTest := testCase.input ==
				"mcp add test-server ./cmd/mcp-test-server/mcp-test-server"

			testServerCommand := ""

			if isMCPAddTest {
				originalDir, err := os.Getwd()
				if err != nil {
					t.Fatalf("failed to get working directory: %v", err)
				}

				tempDir := t.TempDir()

				if err := os.Chdir(tempDir); err != nil {
					t.Fatalf("failed to change working directory: %v", err)
				}

				t.Cleanup(func() {
					if err := os.Chdir(originalDir); err != nil {
						t.Fatalf(
							"failed to restore working directory: %v",
							err,
						)
					}
				})

				if err := os.MkdirAll("config", 0755); err != nil {
					t.Fatalf(
						"failed to create config directory: %v",
						err,
					)
				}

				_, currentFile, _, ok := runtime.Caller(0)
				if !ok {
					t.Fatal("failed to determine current test file")
				}

				projectRoot := filepath.Join(
					filepath.Dir(currentFile),
					"..",
				)

				testServerCommand = filepath.Join(
					projectRoot,
					"cmd",
					"mcp-test-server",
					"mcp-test-server",
				)

				if _, err := os.Stat(testServerCommand); err != nil {
					t.Fatalf(
						"test MCP server executable not found: %v",
						err,
					)
				}

				testCase.input = "mcp add test-server " + testServerCommand

				initialConfig := &configpkg.Config{
					Provider: configpkg.Provider{
						Name:     "ollama",
						Model:    "llama3.1",
						BaseURL:  "http://localhost:11434",
						Endpoint: "/api/chat",
					},
				}

				if err := configpkg.Save(
					filepath.Join("config", "config.json"),
					initialConfig,
				); err != nil {
					t.Fatalf(
						"failed to create initial config: %v",
						err,
					)
				}
			}

			isMCPRemoveTest := testCase.input == "mcp remove test-server"

			if isMCPRemoveTest {
				originalDir, err := os.Getwd()
				if err != nil {
					t.Fatalf("failed to get working directory: %v", err)
				}

				tempDir := t.TempDir()

				if err := os.Chdir(tempDir); err != nil {
					t.Fatalf("failed to change working directory: %v", err)
				}

				t.Cleanup(func() {
					if err := os.Chdir(originalDir); err != nil {
						t.Fatalf(
							"failed to restore working directory: %v",
							err,
						)
					}
				})

				if err := os.MkdirAll("config", 0755); err != nil {
					t.Fatalf(
						"failed to create config directory: %v",
						err,
					)
				}

				initialConfig := &configpkg.Config{
					Provider: configpkg.Provider{
						Name:     "ollama",
						Model:    "llama3.1",
						BaseURL:  "http://localhost:11434",
						Endpoint: "/api/chat",
					},
					MCP: configpkg.MCPConfig{
						Servers: testCase.mcpServers,
					},
				}

				if err := configpkg.Save(
					filepath.Join("config", "config.json"),
					initialConfig,
				); err != nil {
					t.Fatalf(
						"failed to create initial config: %v",
						err,
					)
				}
			}

			got := handleCommand(
				testCase.input,
				testCase.model,
				testCase.modelSource,
				&currentSession,
				sessionStore,
				agentClient,
				appConfig,
				mcpRuntime,
				registry,
				&output,
			)

			if isMCPAddTest {
				savedConfig, err := configpkg.Load(
					filepath.Join("config", "config.json"),
				)
				if err != nil {
					t.Fatalf(
						"failed to load saved config: %v",
						err,
					)
				}

				if len(savedConfig.MCP.Servers) != 1 {
					t.Fatalf(
						"expected 1 MCP server, got %d",
						len(savedConfig.MCP.Servers),
					)
				}

				server := savedConfig.MCP.Servers[0]

				if server.Name != "test-server" {
					t.Fatalf(
						"expected server name %q, got %q",
						"test-server",
						server.Name,
					)
				}

				if server.Command != testServerCommand {
					t.Fatalf(
						"expected server command %q, got %q",
						testServerCommand,
						server.Command,
					)
				}

				if !registry.Has("test_tool") {
					t.Fatalf(
						"expected MCP tool %q to be registered immediately",
						"test_tool",
					)
				}
			}

			if isMCPRemoveTest {
				savedConfig, err := configpkg.Load(
					filepath.Join("config", "config.json"),
				)
				if err != nil {
					t.Fatalf(
						"failed to load saved config: %v",
						err,
					)
				}

				if len(savedConfig.MCP.Servers) != 1 {
					t.Fatalf(
						"expected 1 MCP server, got %d",
						len(savedConfig.MCP.Servers),
					)
				}

				if savedConfig.MCP.Servers[0].Name != "filesystem" {
					t.Fatalf(
						"expected remaining server %q, got %q",
						"filesystem",
						savedConfig.MCP.Servers[0].Name,
					)
				}
			}

			if got != testCase.expectedResult {
				t.Fatalf(
					"handleCommand() = %v, expectedResult = %v",
					got,
					testCase.expectedResult,
				)
			}

			if testCase.expectedOutput != "" &&
				!strings.Contains(output.String(), testCase.expectedOutput) {
				t.Fatalf(
					"expected output to contain %q, got %q",
					testCase.expectedOutput,
					output.String(),
				)
			}

			if testCase.input == "session create new-session" {
				_, err := sessionStore.Get("new-session")

				if err != nil {
					t.Fatalf(
						"expected new-session to exist in store, got error: %v",
						err,
					)
				}

				if currentSession.ID != "new-session" {
					t.Fatalf(
						"expected current session to be new-session, got %q",
						currentSession.ID,
					)
				}

				if agentClient.GetMessages() == nil {
					t.Fatalf(
						"expected agent to have active session",
					)
				}
			}

			if testCase.input == "session load loaded-session" {
				if currentSession.ID != "loaded-session" {
					t.Fatalf(
						"expected current session to be loaded-session, got %q",
						currentSession.ID,
					)
				}

				if agentClient.GetMessages() == nil {
					t.Fatalf(
						"expected agent to have active session",
					)
				}
			}

			if testCase.input == "session delete delete-session" {
				_, err := sessionStore.Get("delete-session")

				if err == nil {
					t.Fatalf(
						"expected delete-session to be removed from store",
					)
				}
			}
		})
	}
}

// Test for clearing the terminal.
func TestClearTerminal(t *testing.T) {
	var output strings.Builder

	clearTerminal(&output)

	expected := "\033[H\033[2J"

	if output.String() != expected {
		t.Fatalf(
			"clearTerminal() = %q, expected %q",
			output.String(),
			expected,
		)
	}
}

// Test for checking whether a model exists.
func TestContainsModel(t *testing.T) {
	testCases := []struct {
		name     string
		models   []string
		target   string
		expected bool
	}{
		{
			name:     "Model exists",
			models:   []string{"llama3.1", "mistral"},
			target:   "mistral",
			expected: true,
		},
		{
			name:     "Model does not exist",
			models:   []string{"llama3.1", "mistral"},
			target:   "qwen",
			expected: false,
		},
		{
			name:     "Empty model list",
			models:   []string{},
			target:   "llama3.1",
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := containsModel(testCase.models, testCase.target)

			if got != testCase.expected {
				t.Fatalf(
					"containsModel() = %v, expected %v",
					got,
					testCase.expected,
				)
			}
		})
	}
}
