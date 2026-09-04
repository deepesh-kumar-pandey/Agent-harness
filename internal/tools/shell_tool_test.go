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
