package main

import (
	"bufio"
	"errors"
	"strings"
	"testing"

	providerpkg "agent-harness/internal/provider"
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
		expectMessage  string
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
			name:           "Invalid choice then selects alternative",
			provider:       &fakeLocalProvider{models: []string{"mistral:latest"}},
			configured:     "llama3.1",
			input:          "x\n1\n",
			expected:       "mistral:latest",
			expectedSource: "local-selection",
			expectOutput:   "Invalid choice",
			expectMessage:  "selection flow",
		},
		{
			name:           "Pulls configured model",
			provider:       &fakeLocalProvider{},
			configured:     "llama3.1",
			input:          "p\n",
			expected:       "llama3.1",
			expectedSource: "pulled",
			expectOutput:   "Model \"llama3.1\" pulled successfully.",
			expectPulls:    1,
			expectMessage:  "pull success",
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
			expectMessage:  "pull retry",
		},
		{
			name: "Quit after pull failure",
			provider: &fakeLocalProvider{
				pullErrors: []error{errors.New("download failed")},
			},
			configured:    "llama3.1",
			input:         "p\nq\n",
			expectError:   true,
			expectOutput:  "Failed to pull model \"llama3.1\": download failed",
			expectPulls:   1,
			expectMessage: "pull quit",
		},
		{
			name:          "Reports Ollama unavailable",
			provider:      &fakeLocalProvider{listErr: errors.New("connection refused")},
			configured:    "llama3.1",
			expectError:   true,
			expectOutput:  "Unable to connect to Ollama.",
			expectMessage: "connection error",
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
				t.Fatalf("expected model %q, got %q", testCase.expected, got)
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

			if localProvider, ok := testCase.provider.(*fakeLocalProvider); ok &&
				len(localProvider.pullCalls) != testCase.expectPulls {
				t.Fatalf(
					"expected %d pull calls, got %d",
					testCase.expectPulls,
					len(localProvider.pullCalls),
				)
			}
		})
	}
}

func TestHandleCommand(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		model          string
		modelSource    string
		expectedResult CommandResult
		expectedOutput string
	}{
		{
			name:           "Help command",
			input:          "help",
			model:          "test-model",
			modelSource:    "test-source",
			expectedResult: CommandHandled,
			expectedOutput: "Available commands: help, exit, model, clear",
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

			got := handleCommand(
				testCase.input,
				testCase.model,
				testCase.modelSource,
				&output,
			)

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
		})
	}
}

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
