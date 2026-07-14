package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/killallservers/styx/pkg/config"
	"github.com/killallservers/styx/pkg/registry"
)

var infoCmd = &cobra.Command{
	Use:   "info [tool]",
	Short: "Show tool metadata",
	Long: `Show tool metadata from the registry.

Displays available versions, stability markers, and supported platforms.

Examples:
  styx info golang            # Show golang versions and platforms
  styx info                   # Show all tools in registry`,
	RunE: runInfo,
}

func runInfo(cmd *cobra.Command, args []string) error {
	// Load hierarchical config to get registry settings
	merged, err := config.LoadHierarchical()
	if err != nil {
		// Config might not exist, fall back to embedded
		merged = config.Merged{
			Tools:      make(map[string]string),
			Env:        make(map[string]string),
			Registries: []config.Registry{},
		}
	}

	// Load registry (with HTTP fallback if configured)
	reg, err := registry.ResolveRegistry(merged)
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// If no tool specified, show all
	if len(args) == 0 {
		return listAllTools(reg)
	}

	// Show specific tool
	toolName := args[0]
	return showToolInfo(reg, toolName)
}

func listAllTools(reg map[string]*registry.ToolSpec) error {
	if len(reg) == 0 {
		fmt.Println("Registry is empty")
		return nil
	}

	fmt.Println("Available tools in registry:")
	fmt.Println("")

	// Sort tools by name
	tools := make([]string, 0, len(reg))
	for name := range reg {
		tools = append(tools, name)
	}
	sort.Strings(tools)

	for _, toolName := range tools {
		spec := reg[toolName]
		versionCount := len(spec.Versions)
		fmt.Printf("  %-15s %d version(s)  %s\n", toolName, versionCount, spec.Repository)
	}

	fmt.Println("")
	fmt.Printf("Total: %d tools\n", len(reg))
	fmt.Println("")
	fmt.Println("Use 'styx info <tool>' to see details")

	return nil
}

func showToolInfo(reg map[string]*registry.ToolSpec, toolName string) error {
	spec, ok := reg[toolName]
	if !ok {
		return fmt.Errorf("tool %q not found in registry", toolName)
	}

	fmt.Printf("Tool: %s\n", spec.Name)
	fmt.Printf("Repository: %s\n", spec.Repository)
	fmt.Println("")

	// Sort versions (reverse semantic order would be nice, but alphabetical works)
	versions := make([]string, 0, len(spec.Versions))
	for v := range spec.Versions {
		versions = append(versions, v)
	}
	sort.Strings(versions)

	fmt.Printf("Versions (%d):\n", len(versions))
	for _, version := range versions {
		versionSpec := spec.Versions[version]
		stability := versionSpec.Stability
		released := versionSpec.Released

		// Determine stability marker
		stabilityMarker := "stable"
		if stability == "experimental" {
			stabilityMarker = "experimental"
		} else if stability == "legacy" {
			stabilityMarker = "legacy"
		}

		fmt.Printf("  • %-10s [%s] Released: %s\n",
			version, stabilityMarker, released)

		// Show platforms for this version
		if len(versionSpec.Methods) > 0 {
			method := versionSpec.Methods[0] // Binary method
			platforms := make([]string, 0, len(method.Platforms))
			for p := range method.Platforms {
				platforms = append(platforms, p)
			}
			sort.Strings(platforms)

			platformStr := strings.Join(platforms, ", ")
			fmt.Printf("      Platforms: %s\n", platformStr)
		}
	}

	fmt.Println("")
	fmt.Printf("Recommended version: %s\n", getLatestStableVersion(spec))

	return nil
}

func getLatestStableVersion(spec *registry.ToolSpec) string {
	// Find latest stable version (simplified: just return first stable, or first version)
	var stableVersions []string
	var allVersions []string

	for version, versionSpec := range spec.Versions {
		allVersions = append(allVersions, version)
		if versionSpec.Stability == "stable" {
			stableVersions = append(stableVersions, version)
		}
	}

	// Sort to find latest
	sort.Strings(stableVersions)
	if len(stableVersions) > 0 {
		return stableVersions[len(stableVersions)-1]
	}

	// Fallback to any version
	sort.Strings(allVersions)
	if len(allVersions) > 0 {
		return allVersions[len(allVersions)-1]
	}

	return "unknown"
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
