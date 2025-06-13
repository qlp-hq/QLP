package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// LocalStorage implements the Store interface for the local filesystem.
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local storage instance.
// It uses the LOCAL_STORAGE_PATH environment variable, defaulting to "./output".
func NewLocalStorage() (*LocalStorage, error) {
	path := os.Getenv("LOCAL_STORAGE_PATH")
	if path == "" {
		path = "./output"
	}

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("could not create base storage directory: %w", err)
	}

	return &LocalStorage{basePath: path}, nil
}

// Save writes content to a file on the local disk.
func (l *LocalStorage) Save(ctx context.Context, path string, content []byte) (string, error) {
	fullPath := filepath.Join(l.basePath, path)

	// Ensure the directory for the file exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", fmt.Errorf("could not create directory for artifact: %w", err)
	}

	err := ioutil.WriteFile(fullPath, content, 0644)
	if err != nil {
		return "", fmt.Errorf("could not write artifact to disk: %w", err)
	}
	return fullPath, nil
}

// Get reads content from a file on the local disk.
func (l *LocalStorage) Get(ctx context.Context, path string) ([]byte, error) {
	fullPath := filepath.Join(l.basePath, path)
	return ioutil.ReadFile(fullPath)
}

// GetURL returns a file URI for the given path.
func (l *LocalStorage) GetURL(ctx context.Context, path string) (string, error) {
	absPath, err := filepath.Abs(filepath.Join(l.basePath, path))
	if err != nil {
		return "", err
	}
	return "file://" + absPath, nil
}
