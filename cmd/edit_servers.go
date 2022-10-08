package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
)

func editServer(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Aliases: []string{"servers", "serv"},
		Use:     "server [server]",
		Short:   "Edit server",
		Long:    `Open up sake config file in $EDITOR and go to servers section.`,
		Example: `  # Edit servers
  sake edit server

  # Edit server <server>
  sake edit server <server>`,
		Run: func(cmd *cobra.Command, args []string) {
			err := *configErr
			switch e := err.(type) {
			case *core.ConfigNotFound:
				core.CheckIfError(e)
			default:
				runEditServer(args, *config)
			}
		},
		Args: cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil || len(args) == 1 {
				return []string{}, cobra.ShellCompDirectiveDefault
			}
			values := config.GetServerGroupsAndDesc()
			return values, cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	return &cmd
}

func runEditServer(args []string, config dao.Config) {
	if len(args) > 0 {
		err := config.EditServer(args[0])
		core.CheckIfError(err)
	} else {
		err := config.EditServer("")
		core.CheckIfError(err)
	}
}
