package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.1.0-dev"

var rootCmd = &cobra.Command{
	Use:   "styx",
	Short: "Styx - Unified dev environment manager",
	Long: `Styx is a unified dev environment manager for Modo Ventures and portfolio companies.

One TOML config file, one lock file, identical environments everywhere—no Nix language, no runtime AI dependency.

Reference: https://github.com/killallservers/styx`,
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
