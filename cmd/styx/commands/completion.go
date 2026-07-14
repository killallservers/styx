package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate shell completion",
	Long: `Generate shell completion script for bash, zsh, or fish.

To use bash completion, source the generated script in ~/.bashrc:
  styx completion bash >> ~/.bashrc

To use zsh completion, source the generated script in ~/.zshrc:
  styx completion zsh >> ~/.zshrc

To use fish completion, copy the generated script to fish conf.d:
  styx completion fish > ~/.config/fish/conf.d/styx.fish

Examples:
  styx completion bash   # Generate bash completion
  styx completion zsh    # Generate zsh completion
  styx completion fish   # Generate fish completion`,
	ValidArgs: []string{"bash", "zsh", "fish"},
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := args[0]
		switch shell {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		default:
			return fmt.Errorf("unsupported shell: %s", shell)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
