package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/killallservers/styx/pkg/config"
	"github.com/killallservers/styx/pkg/installer"
	"github.com/killallservers/styx/pkg/lock"
	"github.com/killallservers/styx/pkg/registry"
)

var lockFromLock bool

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Generate or update styx.lock",
	Long: `Generate styx.lock from current configuration and registry.

Resolves all tool versions, fetches checksums, and writes a reproducible
lock file that can be committed to version control.

The lock file enables reproducible installs across machines via 'styx sync'.

Examples:
  styx lock               # Generate/update styx.lock
  styx lock --from-lock   # Reproduce from existing lock`,
	RunE: runLock,
}

func runLock(cmd *cobra.Command, args []string) error {
	// Load configuration from hierarchy (global + all parent dirs)
	merged, err := config.LoadHierarchical()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Load registry (with HTTP fallback if configured)
	reg, err := registry.ResolveRegistry(merged)
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Create directories
	homeDir := os.Getenv("HOME")
	storeDir := filepath.Join(homeDir, ".styx", "store")
	cacheDir := filepath.Join(homeDir, ".styx", "cache")
	binDir := filepath.Join(homeDir, ".styx", "bin")

	// Initialize downloader and installer
	dl, err := installer.NewDownloader(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to initialize downloader: %w", err)
	}

	inst, err := installer.NewInstaller(reg, dl, storeDir, binDir)
	if err != nil {
		return fmt.Errorf("failed to initialize installer: %w", err)
	}

	// Install tools to resolve versions and get hashes
	results, err := inst.InstallAll(merged.Tools)
	if err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	// Check if all succeeded
	for _, result := range results {
		if result.Error != nil {
			return fmt.Errorf("failed to resolve %s: %v", result.Name, result.Error)
		}
	}

	// Convert results to installation records for lock file
	toolRecords := make(map[string]lock.InstallationRecord)
	for _, result := range results {
		toolRecords[result.Name] = lock.InstallationRecord{
			Name:         result.Name,
			Version:      result.Version,
			Method:       result.Method,
			StorePath:    result.StorePath,
			BinaryHash:   result.BinaryHash,
			Executable:   result.Executable,
			SourceConfig: "config",
		}
	}

	// Generate lock file
	lockFile := lock.Generate(&lock.GeneratorConfig{
		RegistryURL:     "embedded",
		RegistryVersion: "0.1.0",
		Tools:           toolRecords,
		Env:             merged.Env,
	})

	// Validate lock file
	if err := lockFile.Validate(); err != nil {
		return fmt.Errorf("lock file validation failed: %w", err)
	}

	// Write lock file
	if err := lockFile.Save("styx.lock"); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}

	fmt.Printf("✓ Lock file generated: styx.lock (%d tools)\n", len(lockFile.Tools))
	return nil
}

func init() {
	lockCmd.Flags().BoolVar(&lockFromLock, "from-lock", false, "Reproduce from existing lock file")
	rootCmd.AddCommand(lockCmd)
}
