package filesystem

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type FileSystem struct{}

// Read reads the contents of a file.
func (f *FileSystem) Read(path string) ([]byte, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", path, err)
	}

	return data, nil
}

// Write writes data to a file.
func (f *FileSystem) Write(path string, data []byte) error {
	err := os.WriteFile(path, data, 0644)

	if err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}

	return nil
}

// List returns the names of files and directories inside a directory.
func (f *FileSystem) List(path string) ([]string, error) {
	entries, err := os.ReadDir(path)

	if err != nil {
		return nil, fmt.Errorf("failed to read directory %q: %w", path, err)
	}

	names := make([]string, 0, len(entries))

	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return names, nil
}

// Search searches recursively for files matching the given pattern.
func (f *FileSystem) Search(path string, pattern string) ([]string, error) {
	var matches []string

	err := filepath.WalkDir(
		path,
		func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if entry.IsDir() {
				return nil
			}

			matched, err := filepath.Match(pattern, entry.Name())

			if err != nil {
				return err
			}

			if matched {
				matches = append(matches, path)
			}

			return nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to search path %q: %w",
			path,
			err,
		)
	}

	return matches, nil
}

// Delete deletes a file or an empty directory.
func (f *FileSystem) Delete(path string) error {
	err := os.Remove(path)

	if err != nil {
		return fmt.Errorf(
			"failed to delete %q: %w",
			path,
			err,
		)
	}

	return nil
}
