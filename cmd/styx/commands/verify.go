package commands

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/killallservers/styx/pkg/lock"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify tool checksums against styx.lock",
	Long: `Verify that installed tools match checksums recorded in styx.lock.

Useful for detecting corrupted downloads or tampered binaries.

Examples:
  styx verify             # Verify all installed tools
  styx verify ripgrep fd  # Verify specific tools`,
	RunE: runVerify,
}

func runVerify(cmd *cobra.Command, args []string) error {
	// Load lock file
	lockFile, err := lock.Load("styx.lock")
	if err != nil {
		return fmt.Errorf("failed to load lock file: %w", err)
	}

	// Validate lock file
	if err := lockFile.Validate(); err != nil {
		return fmt.Errorf("lock file validation failed: %w", err)
	}

	homeDir := os.Getenv("HOME")
	storeDir := filepath.Join(homeDir, ".styx", "store")

	// Calculate hashes of installed binaries
	actual := make(map[string]string)
	for _, tool := range lockFile.Tools {
		hash, err := calculateFileHash(filepath.Join(storeDir, tool.StoragePath))
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: failed to calculate hash: %v\n", tool.Name, err)
			continue
		}
		actual[tool.Name] = hash
	}

	// Verify against lock
	mismatches, err := lock.VerifyAgainstLock(lockFile, actual)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	if len(mismatches) > 0 {
		fmt.Fprintf(os.Stderr, "✗ Verification failed:\n")
		for _, mismatch := range mismatches {
			fmt.Fprintf(os.Stderr, "  - %s\n", mismatch)
		}
		return fmt.Errorf("checksums do not match")
	}

	fmt.Printf("✓ All %d tools verified successfully\n", len(lockFile.Tools))
	return nil
}

func calculateFileHash(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}
