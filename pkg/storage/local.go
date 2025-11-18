package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// Storage defines the interface for file storage operations
type Storage interface {
	// Save saves a file and returns (filename, filePath, error)
	Save(file io.Reader, originalFilename string) (string, string, error)
	// Delete deletes a file
	Delete(filePath string) error
	// Get retrieves a file
	Get(filePath string) (io.ReadCloser, error)
}

// LocalStorage implements file storage on local filesystem
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local file storage
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
	}, nil
}

// Save saves a file to local storage and returns the file path
func (s *LocalStorage) Save(file io.Reader, originalFilename string) (string, string, error) {
	// Generate unique filename
	ext := filepath.Ext(originalFilename)
	filename := uuid.New().String() + ext

	// Create full path
	filePath := filepath.Join(s.basePath, filename)

	// Create file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath)
		return "", "", fmt.Errorf("failed to save file: %w", err)
	}

	// Return filename and relative path
	return filename, filename, nil
}

// Delete deletes a file from local storage
func (s *LocalStorage) Delete(filePath string) error {
	fullPath := filepath.Join(s.basePath, filePath)
	return os.Remove(fullPath)
}

// Get returns a file reader for the given file path
func (s *LocalStorage) Get(filePath string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, filePath)
	return os.Open(fullPath)
}

// GetPath returns the full path for a file
func (s *LocalStorage) GetPath(filePath string) string {
	return filepath.Join(s.basePath, filePath)
}
