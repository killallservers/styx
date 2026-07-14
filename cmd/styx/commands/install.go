package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/killallservers/styx/pkg/config"
	"github.com/killallservers/styx/pkg/installer"
	"github.com/killallservers/styx/pkg/registry"
)

var installCmd = &cobra.Command{
	Use:   "install [tool]...",
	Short: "Install tools from styx.toml",
	Long: `Install tools specified in styx.toml (and global config).

Resolves tool specs from registry, downloads binaries, verifies checksums,
extracts to content-addressable storage, and creates symlinks.

Examples:
  styx install              # Install all tools from config
  styx install ripgrep fd   # Install specific tools`,
	RunE: runInstall,
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Load configuration from hierarchy (global + all parent dirs)
	merged, err := config.LoadHierarchical()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override with command-line args if provided
	if len(args) > 0 {
		for _, toolName := range args {
			if _, ok := merged.Tools[toolName]; ok {
				// Tool already in config, keep its version
				continue
			}
			// If tool name provided without version, we'll need a default
			// For now, error out
			return fmt.Errorf("tool %q not found in config", toolName)
		}
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

	// Install tools
	results, err := inst.InstallAll(merged.Tools)
	if err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	// Report results
	for _, result := range results {
		if result.Error != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: %v\n", result.Name, result.Error)
		} else {
			fmt.Printf("✓ %s@%s installed to %s\n", result.Name, result.Version, binDir)
		}
	}

	// Check if all succeeded
	for _, result := range results {
		if result.Error != nil {
			return fmt.Errorf("one or more installations failed")
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(installCmd)
}
