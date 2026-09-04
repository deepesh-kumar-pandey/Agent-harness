package tools

import (
	"fmt"

	"agent-harness/shell"
)

type ShellTool struct {
	shell shell.Shell
}

func (s *ShellTool) Name() string {
	return "shell"
}

func (s *ShellTool) Description() string {
	return "A tool for executing shell commands."
}

func (s *ShellTool) Execute(args map[string]any) (any, error) {
	command, ok := args["command"].(string)

	if !ok || command == "" {
		return nil, fmt.Errorf("command must be a non-empty string")
	}

	output, err := s.shell.Execute(command)

	if err != nil {
		return nil, fmt.Errorf("failed to execute shell command: %w", err)
	}

	return output, nil
}
