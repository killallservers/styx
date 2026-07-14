package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/killallservers/styx/pkg/config"
	"github.com/killallservers/styx/pkg/lock"
)

var envExportFormat string

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Show environment variables and tool paths",
	Long: `Show loaded environment variables and tool paths from styx.toml and styx.lock.

Displays the merged environment (global + local) that would be set when
running tools or spawning a shell.

Export formats:
  human   - Human-readable (default): KEY=value
  shell   - Shell-compatible: export KEY='value' (for eval)
  json    - JSON format (for programmatic use)

Examples:
  styx env                              # Show all env vars (human format)
  styx env --export-format=shell        # Output for eval "$(styx env ...)"
  styx env --export-format=json         # JSON output
  styx env RUST_BACKTRACE               # Show specific variable`,
	RunE: runEnv,
}

func init() {
	envCmd.Flags().StringVar(&envExportFormat, "export-format", "human", "Export format: human, shell, or json")
	rootCmd.AddCommand(envCmd)
}

func runEnv(cmd *cobra.Command, args []string) error {
	// Load configuration from hierarchy (global + all parent dirs)
	merged, err := config.LoadHierarchical()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Collect environment variables to display
	env := make(map[string]string)

	// Add configured env vars
	for k, v := range merged.Env {
		env[k] = v
	}

	// Add tool paths from lock file or bin directory
	homeDir := os.Getenv("HOME")
	binDir := filepath.Join(homeDir, ".styx", "bin")

	lockFile, err := lock.Load("styx.lock")
	if err == nil && lockFile != nil {
		// Lock file exists, use paths from there
		for _, tool := range lockFile.Tools {
			// Add TOOL_NAME_PATH style variable
			varName := strings.ToUpper(tool.Name) + "_PATH"
			toolPath := filepath.Join(binDir, tool.Executable)
			env[varName] = toolPath
		}
	} else {
		// No lock file, just add tool names that are configured
		for toolName := range merged.Tools {
			varName := strings.ToUpper(toolName) + "_PATH"
			// Assume executable name matches tool name for common tools
			toolPath := filepath.Join(binDir, toolName)
			env[varName] = toolPath
		}
	}

	// Filter if specific variables requested
	if len(args) > 0 {
		filtered := make(map[string]string)
		for _, varName := range args {
			if val, ok := env[varName]; ok {
				filtered[varName] = val
			}
		}
		env = filtered
	}

	// Sort keys
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Output in requested format
	switch envExportFormat {
	case "shell":
		return outputShellFormat(env, keys)
	case "json":
		return outputJSONFormat(env, keys)
	case "human":
		return outputHumanFormat(env, keys)
	default:
		return fmt.Errorf("invalid export format: %s (use human, shell, or json)", envExportFormat)
	}
}

func outputHumanFormat(env map[string]string, keys []string) error {
	for _, k := range keys {
		fmt.Printf("%s=%s\n", k, env[k])
	}
	return nil
}

func outputShellFormat(env map[string]string, keys []string) error {
	for _, k := range keys {
		// Quote value for shell safety
		quoted := shellQuote(env[k])
		fmt.Printf("export %s=%s\n", k, quoted)
	}
	return nil
}

func outputJSONFormat(env map[string]string, keys []string) error {
	// Output as JSON object
	data := make(map[string]string)
	for k, v := range env {
		data[k] = v
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func shellQuote(s string) string {
	// If string contains special chars, use single quotes (but escape embedded single quotes)
	if strings.ContainsAny(s, " \t\n$`\\\"'") {
		// Escape single quotes by ending quote, adding escaped quote, starting quote again
		escaped := strings.ReplaceAll(s, "'", "'\\''")
		return fmt.Sprintf("'%s'", escaped)
	}
	return s
}
