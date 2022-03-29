package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/print"
)

func describeServersCmd(config *dao.Config, configErr *error) *cobra.Command {
	var serverFlags core.ServerFlags

	cmd := cobra.Command{
		Aliases: []string{"server", "serv", "sv"},
		Use:     "servers [servers]",
		Short:   "Describe servers",
		Long:    "Describe servers.",
		Example: `  # Describe all servers
  sake describe servers

  # Describe servers that have tag <tag>
  sake describe servers --tags <tag>`,

		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			describeServers(config, args, serverFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			values := config.GetServerNameAndDesc()
			return values, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringSliceVarP(&serverFlags.Tags, "tags", "t", []string{}, "filter servers by their tag")
	err := cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetTags()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().BoolVarP(&serverFlags.Edit, "edit", "e", false, "edit server")

	return &cmd
}

func describeServers(
	config *dao.Config,
	args []string,
	serverFlags core.ServerFlags,
) {
	if serverFlags.Edit {
		if len(args) > 0 {
			err := config.EditServer(args[0])
			core.CheckIfError(err)
		} else {
			err := config.EditServer("")
			core.CheckIfError(err)
		}
	} else {
		allServers := false
		if len(args) == 0 &&
			len(serverFlags.Tags) == 0 {
			allServers = true
		}

		servers, err := config.FilterServers(allServers, args, serverFlags.Tags)
		core.CheckIfError(err)
		if len(servers) > 0 {
			print.PrintServerBlocks(servers)
		}
	}
}
