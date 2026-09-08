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

	commandArgs, err := shellArgs(args["args"])
	if err != nil {
		return nil, err
	}

	output, err := s.shell.Execute(command, commandArgs...)

	if err != nil {
		return nil, fmt.Errorf("failed to execute shell command: %w", err)
	}

	return output, nil
}

func shellArgs(value any) ([]string, error) {
	if value == nil {
		return nil, nil
	}

	values, ok := value.([]any)
	if !ok {
		stringArgs, ok := value.([]string)
		if ok {
			return stringArgs, nil
		}

		return nil, fmt.Errorf("args must be an array of strings")
	}

	result := make([]string, len(values))

	for index, value := range values {
		argument, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("args must be an array of strings")
		}

		result[index] = argument
	}

	return result, nil
}

func (s *ShellTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type": "string",
			},
			"args": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []string{
			"command",
		},
	}
}
