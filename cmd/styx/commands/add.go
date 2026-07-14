package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/killallservers/styx/pkg/config"
	"github.com/killallservers/styx/pkg/registry"
)

var addCmd = &cobra.Command{
	Use:   "add <tool>[@version]",
	Short: "Add a tool to your config",
	Long: `Add a tool to your local styx.toml configuration.

Resolves the tool from the registry and adds it with the specified version
(or latest if not specified). Updates .styx/styx.toml.

Examples:
  styx add ripgrep           # Add latest ripgrep
  styx add golang@1.22.0     # Add specific golang version
  styx add fd --force        # Add even if already present`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

var addForce bool

func runAdd(cmd *cobra.Command, args []string) error {
	toolArg := args[0]

	// Parse tool[@version]
	var toolName, requestedVersion string
	parts := strings.Split(toolArg, "@")
	toolName = parts[0]
	if len(parts) > 1 {
		requestedVersion = parts[1]
	}

	// Load registry
	reg, err := registry.LoadEmbeddedRegistry()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Check if tool exists
	spec, ok := reg[toolName]
	if !ok {
		return fmt.Errorf("tool %q not found in registry", toolName)
	}

	// Resolve version
	version := requestedVersion
	if version == "" {
		// Get latest version
		version = getLatestVersion(spec)
		if version == "" {
			return fmt.Errorf("no versions found for tool %q", toolName)
		}
	} else {
		// Validate requested version exists
		if spec.GetVersion(version) == nil {
			return fmt.Errorf("version %q not found for tool %q", version, toolName)
		}
	}

	// Load local config
	localConfig, err := config.LoadFromPath(".styx/styx.toml")
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load local config: %w", err)
	}
	if localConfig == nil {
		localConfig = config.NewConfig()
	}

	// Check if already present
	if existing, ok := localConfig.Tools[toolName]; ok && !addForce {
		return fmt.Errorf("tool %q already in config (version %s). Use --force to update", toolName, existing)
	}

	// Add to config
	localConfig.Tools[toolName] = version

	// Save config
	if err := os.MkdirAll(".styx", 0755); err != nil {
		return fmt.Errorf("failed to create .styx directory: %w", err)
	}

	if err := localConfig.Save(".styx/styx.toml"); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Added %s %s to .styx/styx.toml\n", toolName, version)
	fmt.Printf("\nRun 'styx install' to install or 'styx lock' to generate checksums\n")

	return nil
}

// getLatestVersion returns the highest version from the spec's versions map
func getLatestVersion(spec *registry.ToolSpec) string {
	if len(spec.Versions) == 0 {
		return ""
	}

	var latest string
	for version := range spec.Versions {
		if latest == "" || isGreaterVersion(version, latest) {
			latest = version
		}
	}
	return latest
}

// isGreaterVersion does simple semantic version comparison.
// Assumes versions are in MAJOR.MINOR.PATCH format.
func isGreaterVersion(v1, v2 string) bool {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			n1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			n2, _ = strconv.Atoi(parts2[i])
		}

		if n1 > n2 {
			return true
		}
		if n1 < n2 {
			return false
		}
	}

	return false
}

func init() {
	addCmd.Flags().BoolVar(&addForce, "force", false, "Update tool version even if already present")
	rootCmd.AddCommand(addCmd)
}
