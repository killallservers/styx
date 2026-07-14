package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/killallservers/styx/pkg/config"
	"github.com/killallservers/styx/pkg/installer"
	"github.com/killallservers/styx/pkg/lock"
	"github.com/killallservers/styx/pkg/registry"
)

var updateCmd = &cobra.Command{
	Use:   "update [tool@version...]",
	Short: "Update tool version(s)",
	Long: `Update one or more tools to a specific version.

Updates the local config (.styx/styx.toml) with new versions,
then regenerates the lock file with updated checksums.

Examples:
  styx update golang@1.24.0           # Update golang
  styx update golang@1.24.0 node@21   # Update multiple tools
  styx update ripgrep@14.2.0 --lock   # Also regenerate lock`,
	RunE: runUpdate,
}

var updateLock bool

func runUpdate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("must specify at least one tool@version")
	}

	// Parse tool@version arguments
	updates := make(map[string]string)
	for _, arg := range args {
		parts := strings.Split(arg, "@")
		if len(parts) != 2 {
			return fmt.Errorf("invalid format: %s (use tool@version)", arg)
		}
		toolName := parts[0]
		version := parts[1]

		if toolName == "" || version == "" {
			return fmt.Errorf("invalid format: %s (use tool@version)", arg)
		}

		updates[toolName] = version
	}

	// Load local config (just the current dir's .styx/styx.toml for editing)
	localConfig, err := config.LoadFromPath(".styx/styx.toml")
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load local config: %w", err)
	}
	if localConfig == nil {
		localConfig = config.NewConfig()
	}

	// Load registry to validate versions exist
	reg, err := registry.LoadEmbeddedRegistry()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Validate all requested versions exist in registry
	for toolName, version := range updates {
		spec, ok := reg[toolName]
		if !ok {
			return fmt.Errorf("tool %q not found in registry", toolName)
		}

		versionSpec := spec.GetVersion(version)
		if versionSpec == nil {
			return fmt.Errorf("version %q not found for tool %q", version, toolName)
		}
	}

	// Update local config
	for toolName, version := range updates {
		localConfig.Tools[toolName] = version
		fmt.Printf("Updated %s to %s\n", toolName, version)
	}

	// Save updated config
	configPath := ".styx/styx.toml"
	if err := os.MkdirAll(".styx", 0755); err != nil {
		return fmt.Errorf("failed to create .styx directory: %w", err)
	}

	if err := localConfig.Save(configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\n✓ Updated .styx/styx.toml\n")

	// Optionally regenerate lock file
	if updateLock {
		fmt.Println("\nRegenerating styx.lock...")

		// Load hierarchical config (includes all parent dirs)
		merged, err := config.LoadHierarchical()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
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

		// Install updated tools to resolve versions and get hashes
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
	} else {
		fmt.Println("\nRun 'styx lock' to regenerate styx.lock with updated checksums")
	}

	return nil
}

func init() {
	updateCmd.Flags().BoolVar(&updateLock, "lock", false, "Regenerate styx.lock after updating config")
	rootCmd.AddCommand(updateCmd)
}
