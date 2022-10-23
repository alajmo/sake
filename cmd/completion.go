package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
)

func completionCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "Generate completion script",
		Long: `To load completions:
Bash:

  $ source <(sake completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ sake completion bash > /etc/bash_completion.d/sake
  # macOS:
  $ sake completion bash > /usr/local/etc/bash_completion.d/sake

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ sake completion zsh > "${fpath[1]}/_sake"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ sake completion fish | source

  # To load completions for each session, execute once:
  $ sake completion fish > ~/.config/fish/completions/sake.fish

PowerShell:

  PS> mani completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> mani completion powershell > mani.ps1
  # and source this file from your PowerShell profile.
		`,
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run:       generateCompletion,
	}

	return &cmd
}

func generateCompletion(cmd *cobra.Command, args []string) {
	switch args[0] {
	case "bash":
		err := cmd.Root().GenBashCompletion(os.Stdout)
		core.CheckIfError(err)
	case "zsh":
		err := cmd.Root().GenZshCompletion(os.Stdout)
		core.CheckIfError(err)
	case "fish":
		err := cmd.Root().GenFishCompletion(os.Stdout, true)
		core.CheckIfError(err)
	case "powershell":
		err := cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		core.CheckIfError(err)
	}
}
