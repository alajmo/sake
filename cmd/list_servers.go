package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/print"
)

func listServersCmd(config *dao.Config, configErr *error, listFlags *core.ListFlags) *cobra.Command {
	var serverFlags core.ServerFlags

	cmd := cobra.Command{
		Aliases: []string{"server", "serv", "s"},
		Use:     "servers [servers]",
		Short:   "List servers",
		Long:    "List servers.",
		Example: `  # List all servers
  sake list servers

  # List servers <server>
  sake list servers <server>

  # List servers that have tag <tag>
  sake list servers --tags <tag>`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			listServers(config, args, listFlags, &serverFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			values := config.GetServerNameAndDesc()
			return values, cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().StringSliceVarP(&serverFlags.Tags, "tags", "t", []string{}, "filter servers by tags")
	err := cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetTags()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVar(&serverFlags.Headers, "headers", []string{"server", "host", "tag", "description"}, "set headers. Available headers: server, local, user, host, port, tag, description")
	err = cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := []string{"server", "local", "user", "host", "port", "tag", "description"}
		return validHeaders, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func listServers(config *dao.Config, args []string, listFlags *core.ListFlags, serverFlags *core.ServerFlags) {
	theme, err := config.GetTheme(listFlags.Theme)
	core.CheckIfError(err)

	allServers := false
	if len(args) == 0 &&
		len(serverFlags.Tags) == 0 {
		allServers = true
	}

	servers, err := config.FilterServers(allServers, args, serverFlags.Tags)
	core.CheckIfError(err)

	if len(servers) > 0 {
		options := print.PrintTableOptions{
			Output:               listFlags.Output,
			Theme:                *theme,
			OmitEmpty:            false,
			SuppressEmptyColumns: true,
		}

		print.PrintTable("", servers, options, serverFlags.Headers, []string{})
	}
}
