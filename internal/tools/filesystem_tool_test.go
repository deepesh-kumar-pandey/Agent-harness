package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileSystemTool(t *testing.T) {
	testCases := []struct {
		name        string
		operation   string
		expectError bool
	}{
		{name: "Read File", operation: "read", expectError: false},
		{name: "Write File", operation: "write", expectError: false},
		{name: "List Directory", operation: "list", expectError: false},
		{name: "Check File Exists", operation: "exists", expectError: false},
		{name: "Delete File", operation: "delete", expectError: false},
		{name: "Invalid Operation", operation: "invalid", expectError: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, "test.txt")

			if testCase.operation == "read" ||
				testCase.operation == "exists" ||
				testCase.operation == "delete" {

				err := os.WriteFile(
					filePath,
					[]byte("Hello World"),
					0644,
				)

				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}

			filesystemTool := &FilesystemTool{}

			args := map[string]any{
				"operation": testCase.operation,
				"path":      filePath,
			}

			if testCase.operation == "write" {
				args["content"] = "Hello World"
			}

			if testCase.operation == "list" {
				args["path"] = tempDir
			}

			_, err := filesystemTool.Execute(args)

			if testCase.expectError && err == nil {
				t.Fatalf("expected error, got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("got error in %s: %v", testCase.name, err)
			}
		})
	}
}
