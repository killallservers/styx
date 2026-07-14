package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [shell]",
	Short: "Initialize shell integration",
	Long: `Initialize Styx shell integration for automatic environment loading.

When you cd into a project with .styx/styx.toml, Styx automatically loads
the configured environment (tools, paths, variables).

Supported shells:
  bash - Add to ~/.bashrc
  zsh  - Add to ~/.zshrc
  fish - Add to ~/.config/fish/conf.d/styx.fish

Examples:
  styx init bash              # Show bash setup code
  styx init zsh               # Show zsh setup code
  styx init fish              # Show fish setup code

Manual setup:
  1. Run: styx init bash
  2. Copy the output
  3. Paste into your ~/.bashrc at the end
  4. Restart your shell or run: source ~/.bashrc`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("must specify shell: bash, zsh, or fish")
	}

	shell := args[0]
	switch shell {
	case "bash":
		return initBash()
	case "zsh":
		return initZsh()
	case "fish":
		return initFish()
	default:
		return fmt.Errorf("unsupported shell: %s (use bash, zsh, or fish)", shell)
	}
}

func initBash() error {
	code := `# Styx: Auto-load environment when entering projects with .styx/styx.toml
_styx_prompt_command() {
	if command -v styx &> /dev/null; then
		# Look for .styx/styx.toml in current directory or parents
		local dir="$PWD"
		while [[ "$dir" != "/" ]]; do
			if [[ -f "$dir/.styx/styx.toml" ]]; then
				# Found a config, load the environment
				eval "$(styx env --export-format=shell)"
				return 0
			fi
			dir="$(dirname "$dir")"
		done
	fi
	return 0
}

# Hook into PROMPT_COMMAND to auto-load on each prompt
if [[ ":$PROMPT_COMMAND:" != *":_styx_prompt_command:"* ]]; then
	PROMPT_COMMAND="${PROMPT_COMMAND:+${PROMPT_COMMAND}$'\n'}_styx_prompt_command"
fi`

	fmt.Println("# Add this to your ~/.bashrc:")
	fmt.Print("\n")
	fmt.Print(code)
	fmt.Print("\n\n# After adding, run: source ~/.bashrc\n")
	return nil
}

func initZsh() error {
	code := `# Styx: Auto-load environment when entering projects with .styx/styx.toml
_styx_precmd() {
	if command -v styx &> /dev/null; then
		# Look for .styx/styx.toml in current directory or parents
		local dir="$PWD"
		while [[ "$dir" != "/" ]]; do
			if [[ -f "$dir/.styx/styx.toml" ]]; then
				# Found a config, load the environment
				eval "$(styx env --export-format=shell)"
				return 0
			fi
			dir="$(dirname "$dir")"
		done
	fi
	return 0
}

# Hook into precmd to auto-load on each prompt
autoload -Uz add-zsh-hook
add-zsh-hook precmd _styx_precmd`

	fmt.Println("# Add this to your ~/.zshrc:")
	fmt.Print("\n")
	fmt.Print(code)
	fmt.Print("\n\n# After adding, run: exec zsh\n")
	return nil
}

func initFish() error {
	// For fish, we create a config file in the fish config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	fishConfDir := filepath.Join(homeDir, ".config", "fish", "conf.d")
	fishConfFile := filepath.Join(fishConfDir, "styx.fish")

	code := `# Styx: Auto-load environment when entering projects with .styx/styx.toml
function _styx_pwd_change --on-variable PWD
	if command -q styx
		# Look for .styx/styx.toml in current directory or parents
		set -l dir $PWD
		while test "$dir" != "/"
			if test -f "$dir/.styx/styx.toml"
				# Found a config, load the environment
				eval (styx env --export-format=shell)
				return 0
			end
			set dir (dirname $dir)
		end
	end
	return 0
end

# Initial check when shell starts
_styx_pwd_change`

	fmt.Printf("# To install Fish integration, create this file:\n# %s\n\n", fishConfFile)
	fmt.Println(code)
	fmt.Printf("\n# Or run these commands:\n")
	fmt.Printf("mkdir -p '%s'\n", fishConfDir)
	fmt.Printf("cat > '%s' << 'EOF'\n%s\nEOF\n", fishConfFile, code)
	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
