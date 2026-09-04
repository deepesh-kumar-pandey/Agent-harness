package tools

import (
	"fmt"
	"os"
)

type FilesystemTool struct{}

func (f *FilesystemTool) Name() string {
	return "filesystem"
}

func (f *FilesystemTool) Description() string {
	return "A tool for performing filesystem operations."
}

func (f *FilesystemTool) Execute(args map[string]any) (any, error) {
	operation, ok := args["operation"].(string)

	if !ok || operation == "" {
		return nil, fmt.Errorf("operation must be a non-empty string")
	}

	switch operation {

	case "read":
		path, ok := args["path"].(string)

		if !ok || path == "" {
			return nil, fmt.Errorf("path must be a non-empty string")
		}

		data, err := os.ReadFile(path)

		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		return string(data), nil

	case "write":
		path, ok := args["path"].(string)

		if !ok || path == "" {
			return nil, fmt.Errorf("path must be a non-empty string")
		}

		content, ok := args["content"].(string)

		if !ok {
			return nil, fmt.Errorf("content must be a string")
		}

		err := os.WriteFile(path, []byte(content), 0644)

		if err != nil {
			return nil, fmt.Errorf("failed to write file: %w", err)
		}

		return fmt.Sprintf("Successfully wrote to %s", path), nil

	case "list":
		path, ok := args["path"].(string)

		if !ok || path == "" {
			return nil, fmt.Errorf("path must be a non-empty string")
		}

		files, err := os.ReadDir(path)

		if err != nil {
			return nil, fmt.Errorf("failed to list directory: %w", err)
		}

		result := make([]string, 0, len(files))

		for _, file := range files {
			result = append(result, file.Name())
		}

		return result, nil

	case "exists":
		path, ok := args["path"].(string)

		if !ok || path == "" {
			return nil, fmt.Errorf("path must be a non-empty string")
		}

		_, err := os.Stat(path)

		if err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}

			return nil, fmt.Errorf("failed to check path: %w", err)
		}

		return true, nil

	case "delete":
		path, ok := args["path"].(string)

		if !ok || path == "" {
			return nil, fmt.Errorf("path must be a non-empty string")
		}

		err := os.Remove(path)

		if err != nil {
			return nil, fmt.Errorf("failed to delete file: %w", err)
		}

		return fmt.Sprintf("Successfully deleted %s", path), nil

	default:
		return nil, fmt.Errorf("unsupported filesystem operation: %s", operation)
	}
}
