package storage

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type ContentStore struct {
	baseDir string
}

func NewContentStore(baseDir string) (*ContentStore, error) {
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}

	return &ContentStore{baseDir: baseDir}, nil
}

func (cs *ContentStore) Write(filename string, data io.Reader) (hash string, err error) {
	h := sha256.New()
	tempDir := filepath.Join(cs.baseDir, ".tmp")
	if err := os.MkdirAll(tempDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	tempFile := filepath.Join(tempDir, filename)
	f, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(io.MultiWriter(f, h), data); err != nil {
		os.Remove(tempFile)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	hash = fmt.Sprintf("%x", h.Sum(nil))
	finalPath := filepath.Join(cs.baseDir, hash[:8], filename)
	finalDir := filepath.Dir(finalPath)

	if err := os.MkdirAll(finalDir, 0700); err != nil {
		os.Remove(tempFile)
		return "", fmt.Errorf("failed to create final directory: %w", err)
	}

	if err := os.Rename(tempFile, finalPath); err != nil {
		os.Remove(tempFile)
		return "", fmt.Errorf("failed to move file to store: %w", err)
	}

	return hash, nil
}

func (cs *ContentStore) Path(hash string) string {
	return filepath.Join(cs.baseDir, hash[:8])
}

func (cs *ContentStore) Exists(hash string) bool {
	_, err := os.Stat(cs.Path(hash))
	return err == nil
}

func (cs *ContentStore) Verify(hash string, expectedHash string) bool {
	return hash == expectedHash
}
