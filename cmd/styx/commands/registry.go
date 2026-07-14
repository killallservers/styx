package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/killallservers/styx/pkg/registry"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage the tool registry",
	Long: `Registry operations for browsing, listing, and exporting tool specs.

Subcommands:
  list       - List all tools in registry
  info       - Show versions for a tool
  export     - Export registry as JSON (for HTTP server)`,
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tools in registry",
	Long: `List all available tools in the registry with version counts.

Shows tool name, description (repository), and number of versions.`,
	RunE: runRegistryList,
}

var registryInfoCmd = &cobra.Command{
	Use:   "info <tool>",
	Short: "Show versions for a tool",
	Long: `Display all available versions for a specific tool.

Shows version, release date, and stability level.`,
	Args: cobra.ExactArgs(1),
	RunE: runRegistryInfo,
}

var registryExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export registry as JSON",
	Long: `Export the entire registry as JSON for use with HTTP registry server.

Output can be written to a file and served from registry.styx.sh`,
	RunE: runRegistryExport,
}

func runRegistryList(cmd *cobra.Command, args []string) error {
	reg, err := registry.LoadEmbeddedRegistry()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Collect and sort tools
	tools := make([]string, 0, len(reg))
	for toolName := range reg {
		tools = append(tools, toolName)
	}
	sort.Strings(tools)

	// Display
	fmt.Printf("Registry: %d tools\n\n", len(tools))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Tool\tRepository\tVersions")
	fmt.Fprintln(w, "----\t----------\t--------")

	for _, toolName := range tools {
		spec := reg[toolName]
		versionCount := len(spec.Versions)
		fmt.Fprintf(w, "%s\t%s\t%d\n", toolName, spec.Repository, versionCount)
	}

	w.Flush()

	return nil
}

func runRegistryInfo(cmd *cobra.Command, args []string) error {
	toolName := args[0]

	reg, err := registry.LoadEmbeddedRegistry()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	spec, ok := reg[toolName]
	if !ok {
		return fmt.Errorf("tool %q not found in registry", toolName)
	}

	// Collect and sort versions
	versions := make([]string, 0, len(spec.Versions))
	for v := range spec.Versions {
		versions = append(versions, v)
	}
	sort.Slice(versions, func(i, j int) bool {
		return isGreaterVersion(versions[i], versions[j])
	})

	// Display
	fmt.Printf("Tool: %s\n", spec.Name)
	fmt.Printf("Repository: %s\n\n", spec.Repository)
	fmt.Printf("Versions: %d\n\n", len(versions))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Version\tReleased\tStability")
	fmt.Fprintln(w, "-------\t--------\t---------")

	for _, v := range versions {
		versionSpec := spec.Versions[v]
		fmt.Fprintf(w, "%s\t%s\t%s\n", v, versionSpec.Released, versionSpec.Stability)
	}

	w.Flush()

	return nil
}

func runRegistryExport(cmd *cobra.Command, args []string) error {
	reg, err := registry.LoadEmbeddedRegistry()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Convert to JSON-serializable format
	output := make(map[string]interface{})
	output["version"] = "1.0.0"
	output["generated"] = "2026-07-14"
	output["tools"] = reg

	// Marshal to JSON
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))

	return nil
}

func init() {
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryInfoCmd)
	registryCmd.AddCommand(registryExportCmd)
	rootCmd.AddCommand(registryCmd)
}
