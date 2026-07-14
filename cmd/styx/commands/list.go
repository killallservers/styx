package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/killallservers/styx/pkg/lock"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed tools",
	Long: `List all installed tools with their versions and paths.

Shows tools from styx.lock if available, otherwise scans ~/.styx/bin.

Examples:
  styx list              # Show all installed tools
  styx list --verify     # Verify checksums too`,
	RunE: runList,
}

var listVerify bool

func runList(cmd *cobra.Command, args []string) error {
	homeDir := os.Getenv("HOME")
	binDir := filepath.Join(homeDir, ".styx", "bin")

	// Try to load lock file first
	lockFile, err := lock.Load("styx.lock")
	if err == nil && lockFile != nil {
		return listFromLock(lockFile, binDir)
	}

	// Fallback: scan bin directory
	return listFromBinDir(binDir)
}

func listFromLock(lockFile *lock.LockFile, binDir string) error {
	if len(lockFile.Tools) == 0 {
		fmt.Println("No tools in lock file")
		return nil
	}

	// Create tabwriter for formatted output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TOOL\tVERSION\tEXECUTABLE\tPATH\tSTATUS")
	fmt.Fprintln(w, "----\t-------\t----------\t----\t------")

	for _, tool := range lockFile.Tools {
		status := "✓"
		executable := tool.Executable
		toolPath := filepath.Join(binDir, tool.Executable)

		// Check if executable exists and verify checksum if requested
		if listVerify {
			hash, err := calculateFileHash(toolPath)
			if err != nil {
				status = "✗ (not found)"
			} else if hash != tool.BinaryHash {
				status = "✗ (checksum mismatch)"
			}
		} else {
			// Just check if file exists
			if _, err := os.Stat(toolPath); err != nil {
				status = "✗ (not found)"
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			tool.Name,
			tool.Version,
			executable,
			toolPath,
			status)
	}

	w.Flush()
	return nil
}

func listFromBinDir(binDir string) error {
	// Scan bin directory for symlinks
	entries, err := os.ReadDir(binDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No tools installed (bin directory not found)")
			return nil
		}
		return fmt.Errorf("failed to read bin directory: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No tools installed")
		return nil
	}

	// Create tabwriter for formatted output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TOOL\tPATH\tSTATUS")
	fmt.Fprintln(w, "----\t----\t------")

	// Sort entries by name
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		toolPath := filepath.Join(binDir, entry.Name())
		status := "✓"

		// Check if it's a valid symlink or file
		if entry.Type()&os.ModeSymlink != 0 {
			// It's a symlink - check if target exists
			target, err := os.Readlink(toolPath)
			if err != nil {
				status = "✗ (broken symlink)"
			} else if _, err := os.Stat(target); err != nil {
				status = "✗ (target missing)"
			}
		} else {
			// Regular file
			if _, err := os.Stat(toolPath); err != nil {
				status = "✗ (not found)"
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n",
			entry.Name(),
			toolPath,
			status)
	}

	w.Flush()
	return nil
}

func init() {
	listCmd.Flags().BoolVar(&listVerify, "verify", false, "Verify checksums against lock file")
	rootCmd.AddCommand(listCmd)
}
