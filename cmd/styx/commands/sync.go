package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/killallservers/styx/pkg/installer"
	"github.com/killallservers/styx/pkg/lock"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Install tools from styx.lock (reproducible install)",
	Long: `Install tools from styx.lock on a fresh machine or CI environment.

This command reproduces the exact environment recorded in styx.lock:
same tool versions, same checksums, bit-for-bit reproducible.

Useful for:
- CI/CD pipelines (install identical tools from lock)
- New team members (get same environment as original developer)
- Fresh machines (reproduce environment without building config)

Examples:
  styx sync              # Install all tools from lock
  styx sync ripgrep fd   # Sync specific tools only`,
	RunE: runSync,
}

func runSync(cmd *cobra.Command, args []string) error {
	// Load lock file
	lockFile, err := lock.Load("styx.lock")
	if err != nil {
		return fmt.Errorf("failed to load lock file: %w", err)
	}

	if lockFile == nil {
		return fmt.Errorf("styx.lock not found (run 'styx lock' first to generate)")
	}

	// Validate lock file
	if err := lockFile.Validate(); err != nil {
		return fmt.Errorf("lock file validation failed: %w", err)
	}

	// Create directories
	homeDir := os.Getenv("HOME")
	storeDir := filepath.Join(homeDir, ".styx", "store")
	cacheDir := filepath.Join(homeDir, ".styx", "cache")
	binDir := filepath.Join(homeDir, ".styx", "bin")

	// Initialize downloader and installer (with empty registry since we're using lock)
	dl, err := installer.NewDownloader(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to initialize downloader: %w", err)
	}

	// For sync, we don't need the full registry - we use the lock file data directly
	// Create a minimal installer that can handle lock-based installs
	inst, err := installer.NewInstaller(nil, dl, storeDir, binDir)
	if err != nil {
		return fmt.Errorf("failed to initialize installer: %w", err)
	}

	// Build list of tools to sync
	// If specific tools requested, filter to those
	toolsToSync := make(map[string]bool)
	if len(args) > 0 {
		for _, toolName := range args {
			toolsToSync[toolName] = true
		}
	} else {
		// Sync all tools from lock
		for _, tool := range lockFile.Tools {
			toolsToSync[tool.Name] = true
		}
	}

	// Install each tool from lock file
	successCount := 0
	failCount := 0

	for _, toolEntry := range lockFile.Tools {
		if !toolsToSync[toolEntry.Name] {
			continue // Skip if not requested
		}

		fmt.Printf("Installing %s@%s... ", toolEntry.Name, toolEntry.Version)

		// Create the spec-like data needed for installation
		// We use the lock file data to reconstruct what we need
		result, err := inst.InstallFromLock(toolEntry, lockFile)
		if err != nil {
			fmt.Printf("✗ %v\n", err)
			failCount++
			continue
		}

		// Verify checksum matches lock
		if result.BinaryHash != toolEntry.BinaryHash {
			fmt.Printf("✗ checksum mismatch (expected: %s, got: %s)\n",
				toolEntry.BinaryHash, result.BinaryHash)
			failCount++
			continue
		}

		fmt.Printf("✓\n")
		successCount++
	}

	// Summary
	fmt.Printf("\n")
	if failCount == 0 {
		fmt.Printf("✓ Synced %d tools from lock file\n", successCount)
		return nil
	}

	fmt.Printf("✗ Sync incomplete: %d succeeded, %d failed\n", successCount, failCount)
	return fmt.Errorf("sync failed for %d tools", failCount)
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
