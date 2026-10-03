package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

// Tests reading existing, empty, and missing files.
func TestFileSystemRead(t *testing.T) {
	filesystem := FileSystem{}
	dir := t.TempDir()

	existingFile := filepath.Join(dir, "test.txt")
	emptyFile := filepath.Join(dir, "empty.txt")
	missingFile := filepath.Join(dir, "missing.txt")

	if err := os.WriteFile(existingFile, []byte("Hello, Agent Harness!"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create empty file: %v", err)
	}

	testCases := []struct {
		name        string
		path        string
		content     string
		expectError bool
	}{
		{
			name:    "Read Existing File",
			path:    existingFile,
			content: "Hello, Agent Harness!",
		},
		{
			name:    "Read Empty File",
			path:    emptyFile,
			content: "",
		},
		{
			name:        "Read Missing File",
			path:        missingFile,
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			data, err := filesystem.Read(testCase.path)

			if testCase.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(data) != testCase.content {
				t.Errorf("expected content %q, got %q", testCase.content, data)
			}
		})
	}
}

// Tests writing new files, overwriting files, and invalid paths.
func TestFileSystemWrite(t *testing.T) {
	filesystem := FileSystem{}
	dir := t.TempDir()

	testCases := []struct {
		name        string
		path        string
		data        []byte
		expectError bool
	}{
		{
			name: "Write New File",
			path: filepath.Join(dir, "test.txt"),
			data: []byte("Hello, Agent Harness!"),
		},
		{
			name: "Overwrite Existing File",
			path: filepath.Join(dir, "existing.txt"),
			data: []byte("New Content"),
		},
		{
			name:        "Invalid Path",
			path:        filepath.Join(dir, "missing", "test.txt"),
			data:        []byte("Hello"),
			expectError: true,
		},
	}

	if err := os.WriteFile(
		testCases[1].path,
		[]byte("Old Content"),
		0644,
	); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := filesystem.Write(testCase.path, testCase.data)

			if testCase.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			data, err := os.ReadFile(testCase.path)
			if err != nil {
				t.Fatalf("failed to read written file: %v", err)
			}

			if string(data) != string(testCase.data) {
				t.Errorf("expected data %q, got %q", testCase.data, data)
			}
		})
	}
}

// Tests listing valid, empty, and missing directories.
func TestFileSystemList(t *testing.T) {
	filesystem := FileSystem{}
	dir := t.TempDir()
	emptyDir := t.TempDir()

	filePath := filepath.Join(dir, "test.txt")
	subDir := filepath.Join(dir, "subdir")

	if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	testCases := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name: "Valid Directory",
			path: dir,
		},
		{
			name: "Empty Directory",
			path: emptyDir,
		},
		{
			name:        "Missing Directory",
			path:        filepath.Join(dir, "does-not-exist"),
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			names, err := filesystem.List(testCase.path)

			if testCase.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if testCase.name == "Empty Directory" && len(names) != 0 {
				t.Fatalf("expected empty directory, got %v", names)
			}

			if testCase.name == "Valid Directory" {
				if len(names) != 2 {
					t.Fatalf("expected 2 entries, got %d", len(names))
				}
			}
		})
	}
}

// Tests recursive file search and error handling.
func TestFileSystemSearch(t *testing.T) {
	filesystem := FileSystem{}
	dir := t.TempDir()
	nestedDir := filepath.Join(dir, "nested")

	if err := os.Mkdir(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}

	files := map[string]string{
		filepath.Join(dir, "main.go"):           "package main",
		filepath.Join(dir, "readme.txt"):        "hello",
		filepath.Join(nestedDir, "test.go"):     "package main",
		filepath.Join(nestedDir, "config.json"): "{}",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", path, err)
		}
	}

	testCases := []struct {
		name          string
		path          string
		pattern       string
		expectedCount int
		expectError   bool
	}{
		{
			name:          "Find Go Files",
			path:          dir,
			pattern:       "*.go",
			expectedCount: 2,
		},
		{
			name:          "Find JSON Files",
			path:          dir,
			pattern:       "*.json",
			expectedCount: 1,
		},
		{
			name:          "No Matching Files",
			path:          dir,
			pattern:       "*.md",
			expectedCount: 0,
		},
		{
			name:        "Missing Directory",
			path:        filepath.Join(dir, "does-not-exist"),
			pattern:     "*.go",
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			matches, err := filesystem.Search(testCase.path, testCase.pattern)

			if testCase.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(matches) != testCase.expectedCount {
				t.Fatalf(
					"expected %d matches, got %d: %v",
					testCase.expectedCount,
					len(matches),
					matches,
				)
			}
		})
	}
}

// Tests deleting files, empty directories, and missing paths.
func TestFileSystemDelete(t *testing.T) {
	filesystem := FileSystem{}
	dir := t.TempDir()

	filePath := filepath.Join(dir, "test.txt")
	emptyDir := filepath.Join(dir, "empty")
	missingPath := filepath.Join(dir, "does-not-exist")

	if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	testCases := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name: "Delete File",
			path: filePath,
		},
		{
			name: "Delete Empty Directory",
			path: emptyDir,
		},
		{
			name:        "Delete Missing File",
			path:        missingPath,
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := filesystem.Delete(testCase.path)

			if testCase.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
