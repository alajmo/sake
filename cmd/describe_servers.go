package cmd

import (
	"strings"

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
	cmd.Flags().SortFlags = false

	cmd.Flags().StringSliceVarP(&serverFlags.Tags, "tags", "t", []string{}, "filter servers by their tag")
	err := cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetTags()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringVarP(&serverFlags.Regex, "regex", "r", "", "filter servers on host regex")
	cmd.Flags().BoolVarP(&serverFlags.Invert, "invert", "v", false, "invert matching on servers")
	cmd.Flags().BoolVarP(&serverFlags.Edit, "edit", "e", false, "edit server")

	return &cmd
}

func describeServers(
	config *dao.Config,
	args []string,
	serverFlags core.ServerFlags,
) {
	var userArgs []string
	var serverArgs []string
	// Separate user arguments from task ids
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			userArgs = append(userArgs, arg)
		} else {
			serverArgs = append(serverArgs, arg)
		}
	}

	if serverFlags.Edit {
		if len(serverArgs) > 0 {
			err := config.EditServer(serverArgs[0])
			core.CheckIfError(err)
		} else {
			err := config.EditServer("")
			core.CheckIfError(err)
		}
	} else {
		allServers := false
		if len(serverArgs) == 0 &&
			len(serverFlags.Tags) == 0 {
			allServers = true
		}

		err := config.ParseInventory(userArgs)
		core.CheckIfError(err)

		servers, err := config.FilterServers(allServers, serverArgs, serverFlags.Tags, serverFlags.Regex, serverFlags.Invert)
		core.CheckIfError(err)

		if len(servers) > 0 {
			print.PrintServerBlocks(servers)
		}
	}
}
