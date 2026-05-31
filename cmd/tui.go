package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/tui"
)

func tuiCmd(config *dao.Config, configErr *error) *cobra.Command {
	var reload bool

	cmd := cobra.Command{
		Use:   "tui",
		Short: "Open interactive TUI",
		Long:  `Open an interactive terminal user interface for browsing servers, tasks, and executing commands.`,
		Example: `  # Open TUI
  sake tui

  # Open TUI with auto-reload on config changes
  sake tui --reload`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)

			tui.RunTui(config, reload)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().BoolVar(&reload, "reload", false, "auto-reload on config file changes")

	return &cmd
}
