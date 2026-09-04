package tools

import (
	"testing"

	"agent-harness/internal/tools"
)

func TestNewAgent(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "Create agent with registry",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := tools.NewToolRegistry()

			agent := NewAgent(registry)

			if agent == nil {
				t.Fatalf("got nil agent in test case: %s", testCase.name)
			}

			if agent.registry != registry {
				t.Fatalf(
					"registry mismatch in test case: %s",
					testCase.name,
				)
			}
		})
	}
}

func TestExecuteTool(t *testing.T) {
	testCases := []struct {
		name        string
		toolName    string
		args        map[string]any
		expected    any
		expectError bool
	}{
		{
			name:     "Execute calculator addition",
			toolName: "calculator",
			args: map[string]any{
				"operation": "add",
				"numbers":   []float64{10, 20, 30},
			},
			expected:    float64(60),
			expectError: false,
		},
		{
			name:     "Execute calculator multiplication",
			toolName: "calculator",
			args: map[string]any{
				"operation": "multiply",
				"numbers":   []float64{2, 3, 4},
			},
			expected:    float64(24),
			expectError: false,
		},
		{
			name:     "Execute shell command",
			toolName: "shell",
			args: map[string]any{
				"command": "echo hello",
			},
			expected:    "hello\n",
			expectError: false,
		},
		{
			name:        "Execute unregistered tool",
			toolName:    "unknown",
			args:        map[string]any{},
			expected:    nil,
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := tools.NewToolRegistry()

			agent := NewAgent(registry)

			result, err := agent.ExecuteTool(
				testCase.toolName,
				testCase.args,
			)

			if testCase.expectError {
				if err == nil {
					t.Fatalf(
						"expected error in test case: %s",
						testCase.name,
					)
				}

				if result != nil {
					t.Fatalf(
						"expected nil result, got %v in test case: %s",
						result,
						testCase.name,
					)
				}

				return
			}

			if err != nil {
				t.Fatalf(
					"unexpected error in test case: %v",
					err,
				)
			}

			if result != testCase.expected {
				t.Fatalf(
					"expected %v, got %v in test case: %s",
					testCase.expected,
					result,
					testCase.name,
				)
			}
		})
	}
}
