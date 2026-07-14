package lock

import (
	"fmt"
	"time"

	"github.com/killallservers/styx/pkg/config"
)

// GeneratorConfig holds configuration for lock generation.
type GeneratorConfig struct {
	RegistryURL     string
	RegistryVersion string
	Tools           map[string]InstallationRecord
	Env             map[string]string
}

// InstallationRecord holds details about an installed tool.
type InstallationRecord struct {
	Name         string
	Version      string
	Method       string
	StorePath    string
	BinaryHash   string
	Executable   string
	SourceConfig string
}

// Generate creates a new LockFile from the provided configuration.
func Generate(cfg *GeneratorConfig) *LockFile {
	lock := &LockFile{
		Version:     "1.0",
		GeneratedAt: time.Now(),
		RegistrySnapshot: Registry{
			URL:     cfg.RegistryURL,
			Version: cfg.RegistryVersion,
		},
		Tools: make([]ToolEntry, 0),
		Env:   make(map[string]string),
	}

	// Add tools
	for _, record := range cfg.Tools {
		lock.Tools = append(lock.Tools, ToolEntry{
			Name:          record.Name,
			Version:       record.Version,
			InstallMethod: record.Method,
			StoragePath:   record.StorePath,
			BinaryHash:    record.BinaryHash,
			Executable:    record.Executable,
			SourceConfig:  record.SourceConfig,
		})
	}

	// Add environment variables
	for k, v := range cfg.Env {
		lock.Env[k] = v
	}

	return lock
}

// FromMergedConfig generates a lock file from a merged configuration and installation results.
// This is the high-level API used by commands.
func FromMergedConfig(merged config.Merged, registryVersion string, installResults []interface{}) (*LockFile, error) {
	tools := make(map[string]InstallationRecord)

	for _, result := range installResults {
		// Type assertion for installation results
		if ir, ok := result.(InstallationRecord); ok {
			if ir.Name != "" {
				tools[ir.Name] = ir
			}
		}
	}

	return Generate(&GeneratorConfig{
		RegistryURL:     "embedded",
		RegistryVersion: registryVersion,
		Tools:           tools,
		Env:             merged.Env,
	}), nil
}

// VerifyAgainstLock checks that installed tools match what's recorded in a lock file.
func VerifyAgainstLock(lockFile *LockFile, actual map[string]string) ([]string, error) {
	mismatches := make([]string, 0)

	for _, lockTool := range lockFile.Tools {
		actualHash, ok := actual[lockTool.Name]
		if !ok {
			mismatches = append(mismatches, fmt.Sprintf("tool %q is missing", lockTool.Name))
			continue
		}

		if actualHash != lockTool.BinaryHash {
			mismatches = append(mismatches,
				fmt.Sprintf("tool %q hash mismatch: expected %s, got %s",
					lockTool.Name, lockTool.BinaryHash, actualHash))
		}
	}

	return mismatches, nil
}
