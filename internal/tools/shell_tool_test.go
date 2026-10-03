package tools

import (
	"testing"
)

func TestShellTool(t *testing.T) {
	testCases := []struct {
		name        string
		args        map[string]any
		expectError bool
	}{
		{
			name: "Valid Echo Command",
			args: map[string]any{
				"command": "echo",
			},
			expectError: false,
		},
		{
			name:        "Missing Command",
			args:        map[string]any{},
			expectError: true,
		},
		{
			name: "Invalid Command",
			args: map[string]any{
				"command": "this-command-does-not-exist",
			},
			expectError: true,
		},
	}

	shellTool := &ShellTool{}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			_, err := shellTool.Execute(testCase.args)

			if testCase.expectError && err == nil {
				t.Fatalf("expected error, got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
		})
	}
}

func TestShellToolMethods(t *testing.T) {
	shellTool := &ShellTool{}

	if shellTool.Name() != "shell" {
		t.Fatalf("expected name %q, got %q", "shell", shellTool.Name())
	}

	if shellTool.Description() == "" {
		t.Fatal("expected description to be non-empty")
	}

	if shellTool.Schema() == nil {
		t.Fatal("expected schema to be non-nil")
	}
}

// Tests valid and invalid shell argument formats.
func TestShellArgs(t *testing.T) {
	testCases := []struct {
		name        string
		input       any
		expected    []string
		expectError bool
	}{
		{
			name:     "nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "string slice",
			input:    []string{"hello", "world"},
			expected: []string{"hello", "world"},
		},
		{
			name:     "any slice",
			input:    []any{"hello", "world"},
			expected: []string{"hello", "world"},
		},
		{
			name:        "invalid element",
			input:       []any{"hello", 123},
			expectError: true,
		},
		{
			name:        "invalid type",
			input:       "invalid",
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := shellArgs(testCase.input)

			if testCase.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(testCase.expected) {
				t.Fatalf("expected %v, got %v", testCase.expected, result)
			}

			for i := range result {
				if result[i] != testCase.expected[i] {
					t.Fatalf("expected %v, got %v", testCase.expected, result)
				}
			}
		})
	}
}
