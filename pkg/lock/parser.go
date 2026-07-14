package lock

import (
	"encoding/json"
	"fmt"
	"os"
)

func Load(path string) (*LockFile, error) {
	if path == "" {
		path = "styx.lock"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	var lf LockFile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("failed to parse lock file: %w", err)
	}

	return &lf, nil
}

func (lf *LockFile) Save(path string) error {
	if path == "" {
		path = "styx.lock"
	}

	data, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}

	return nil
}

func (lf *LockFile) Validate() error {
	if lf.Version != "1.0" {
		return fmt.Errorf("unsupported lock file version: %s", lf.Version)
	}

	if len(lf.Tools) == 0 {
		return fmt.Errorf("lock file contains no tools")
	}

	for _, tool := range lf.Tools {
		if tool.Name == "" {
			return fmt.Errorf("tool entry missing name")
		}
		if tool.Version == "" {
			return fmt.Errorf("tool %s missing version", tool.Name)
		}
		if tool.BinaryHash == "" {
			return fmt.Errorf("tool %s missing binary hash", tool.Name)
		}
	}

	return nil
}
