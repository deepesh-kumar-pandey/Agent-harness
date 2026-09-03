package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileSystem(t *testing.T) {
	testCases := []struct {
		name        string
		path        string
		content     string
		expectError bool
	}{
		{
			name:        "Read Existing File",
			path:        "test.txt",
			content:     "Hello, Agent Harness!",
			expectError: false,
		},
		{
			name:        "Read Empty File",
			path:        "empty.txt",
			content:     "",
			expectError: false,
		},
		{
			name:        "Read Missing File",
			path:        "missing.txt",
			content:     "",
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if !testCase.expectError {
				file, err := os.Create(testCase.path)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				if _, err := file.WriteString(testCase.content); err != nil {
					file.Close()
					t.Fatalf("failed to write test file: %v", err)
				}
				if err := file.Close(); err != nil {
					t.Fatalf("failed to close test file: %v", err)
				}
				defer os.Remove(testCase.path)
			}

			data, err := os.ReadFile(testCase.path)
			if (err != nil) != testCase.expectError {
				t.Fatalf("expected error: %v, got: %v", testCase.expectError, err)
			}
			if err == nil && string(data) != testCase.content {
				t.Errorf("expected content %q, got %q", testCase.content, data)
			}
		})
	}
}

func TestFileSystemWrite(t *testing.T) {
	testCases := []struct {
		name        string
		path        string
		data        []byte
		expectError bool
	}{
		{
			name:        "Write New File",
			path:        "test.txt",
			data:        []byte("Hello, Agent Harness!"),
			expectError: false,
		},
		{
			name:        "Overwrite Existing File",
			path:        "test.txt",
			data:        []byte("New Content"),
			expectError: false,
		},
		{
			name:        "Invalid Path",
			path:        "/invalid/path/test.txt",
			data:        []byte("Hello"),
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := os.WriteFile(testCase.path, testCase.data, 0644)

			if err != nil && !testCase.expectError {
				t.Fatalf("expected no error, got: %v", err)
			}

			if err == nil && testCase.expectError {
				t.Fatalf("expected error for path %q, but got none", testCase.path)
			}

			if !testCase.expectError {
				defer os.Remove(testCase.path)
				readData, err := os.ReadFile(testCase.path)
				if err != nil {
					t.Fatalf("failed to read file after write: %v", err)
				}
				if string(readData) != string(testCase.data) {
					t.Errorf("expected data %q, got %q", testCase.data, readData)
				}
			}
		})
	}
}

func TestFileSystemList(t *testing.T) {
	dir := t.TempDir()
	emptyDir := t.TempDir()

	testCases := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "Valid Directory",
			path:        dir,
			expectError: false,
		},
		{
			name:        "Empty Directory",
			path:        emptyDir,
			expectError: false,
		},
		{
			name:        "Missing Directory",
			path:        "does-not-exist",
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := os.ReadDir(testCase.path)

			if err == nil && testCase.expectError {
				t.Fatalf("expected error for path %q, but got none", testCase.path)
			}

			if err != nil && !testCase.expectError {
				t.Fatalf("expected no error, got: %v", err)
			}
		})
	}
}

func TestFileSystemSearch(t *testing.T) {
	dir := t.TempDir()

	// Create test files
	files := map[string]string{
		"main.go":     "package main",
		"test.go":     "package main",
		"readme.txt":  "hello",
		"config.json": "{}",
	}

	for name, content := range files {
		path := filepath.Join(dir, name)

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", name, err)
		}
	}

	testCases := []struct {
		name        string
		path        string
		pattern     string
		expectError bool
	}{
		{
			name:        "Find Go Files",
			path:        dir,
			pattern:     "*.go",
			expectError: false,
		},
		{
			name:        "No Matching Files",
			path:        dir,
			pattern:     "*.txt",
			expectError: false,
		},
		{
			name:        "Missing Directory",
			path:        filepath.Join(dir, "does-not-exist"),
			pattern:     "*.go",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matches := make([]string, 0)

			entries, err := os.ReadDir(tc.path)
			if err == nil {
				for _, entry := range entries {
					matched, matchErr := filepath.Match(tc.pattern, entry.Name())
					if matchErr != nil {
						err = matchErr
						break
					}
					if matched {
						matches = append(matches, filepath.Join(tc.path, entry.Name()))
					}
				}
			}

			if tc.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			t.Logf("matches: %v", matches)
		})
	}
}

func TestFileSystemDelete(t *testing.T) {
	dir := t.TempDir()

	// Create a file that we will delete
	filePath := filepath.Join(dir, "test.txt")

	if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create an empty directory that we will delete
	emptyDir := filepath.Join(dir, "empty")
	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	testCases := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "Delete File",
			path:        filePath,
			expectError: false,
		},
		{
			name:        "Delete Empty Directory",
			path:        emptyDir,
			expectError: false,
		},
		{
			name:        "Delete Missing File",
			path:        filepath.Join(dir, "does-not-exist.txt"),
			expectError: true,
		},
	}

	filesystem := FileSystem{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			err := filesystem.Delete(tc.path)

			if tc.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
		})
	}
}
