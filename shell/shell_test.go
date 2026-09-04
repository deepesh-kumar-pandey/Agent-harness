package shell

import (
	"testing"
)

func TestExecute(t *testing.T) {
	testCases := []struct {
		name        string
		command     string
		args        []string
		expectError bool
	}{
		{
			name:        "Valid Echo Command",
			command:     "echo",
			args:        []string{"hello"},
			expectError: false,
		},
		{
			name:        "Valid Pwd Command",
			command:     "pwd",
			args:        []string{},
			expectError: false,
		},
		{
			name:        "Command Not Found",
			command:     "this-command-does-not-exist",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "Command Execution Failed",
			command:     "false",
			args:        []string{},
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			shell := &Shell{}

			_, err := shell.Execute(
				testCase.command,
				testCase.args...,
			)

			if testCase.expectError && err == nil {
				t.Fatalf("expected error, got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
		})
	}
}
