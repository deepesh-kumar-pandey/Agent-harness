package tools

import (
	"os"
	"path/filepath"
	"testing"
)

// Tests the basic filesystem operations.
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

// Tests the filesystem tool metadata and schema.
func TestFileSystemToolMethods(t *testing.T) {
	filesystemTool := &FilesystemTool{}

	if filesystemTool.Name() != "filesystem" {
		t.Fatalf("expected name %q, got %q", "filesystem", filesystemTool.Name())
	}

	if filesystemTool.Description() == "" {
		t.Fatal("expected description to be non-empty")
	}

	if filesystemTool.Schema() == nil {
		t.Fatal("expected schema to be non-nil")
	}
}

// Tests validation and filesystem error cases.
func TestFileSystemToolErrors(t *testing.T) {
	filesystemTool := &FilesystemTool{}

	testCases := []struct {
		name string
		args map[string]any
	}{
		{
			name: "missing operation",
			args: map[string]any{},
		},
		{
			name: "empty operation",
			args: map[string]any{
				"operation": "",
			},
		},
		{
			name: "missing read path",
			args: map[string]any{
				"operation": "read",
			},
		},
		{
			name: "empty read path",
			args: map[string]any{
				"operation": "read",
				"path":      "",
			},
		},
		{
			name: "read nonexistent file",
			args: map[string]any{
				"operation": "read",
				"path":      filepath.Join(t.TempDir(), "missing.txt"),
			},
		},
		{
			name: "missing write path",
			args: map[string]any{
				"operation": "write",
				"content":   "Hello World",
			},
		},
		{
			name: "write missing content",
			args: map[string]any{
				"operation": "write",
				"path":      filepath.Join(t.TempDir(), "test.txt"),
			},
		},
		{
			name: "list missing path",
			args: map[string]any{
				"operation": "list",
			},
		},
		{
			name: "list nonexistent directory",
			args: map[string]any{
				"operation": "list",
				"path":      filepath.Join(t.TempDir(), "missing"),
			},
		},
		{
			name: "exists missing path",
			args: map[string]any{
				"operation": "exists",
			},
		},
		{
			name: "delete missing path",
			args: map[string]any{
				"operation": "delete",
			},
		},
		{
			name: "delete nonexistent file",
			args: map[string]any{
				"operation": "delete",
				"path":      filepath.Join(t.TempDir(), "missing.txt"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := filesystemTool.Execute(testCase.args)

			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

// Tests that exists returns false for a nonexistent path.
func TestFileSystemToolExistsFalse(t *testing.T) {
	filesystemTool := &FilesystemTool{}

	result, err := filesystemTool.Execute(map[string]any{
		"operation": "exists",
		"path":      filepath.Join(t.TempDir(), "missing.txt"),
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	exists, ok := result.(bool)
	if !ok {
		t.Fatalf("expected bool result, got %T", result)
	}

	if exists {
		t.Fatal("expected nonexistent path to return false")
	}
}
