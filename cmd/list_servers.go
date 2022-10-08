package cmd

import (
	"strings"

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

	cmd.Flags().StringVarP(&serverFlags.Regex, "regex", "r", "", "filter servers on host regex")

	cmd.Flags().BoolVarP(&serverFlags.Invert, "invert", "v", false, "invert matching on servers")

	cmd.Flags().StringSliceVarP(&serverFlags.Tags, "tags", "t", []string{}, "filter servers by tags")
	err := cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetTags()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVar(&serverFlags.Headers, "headers", []string{"server", "host", "tag", "description"}, "set headers. Available headers: server, local, user, host, port, bastion, tag, description")
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

	theme, err := config.GetTheme(listFlags.Theme)
	core.CheckIfError(err)

	allServers := false
	if len(serverArgs) == 0 &&
		len(serverFlags.Tags) == 0 {
		allServers = true
	}

	err = config.ParseInventory(userArgs)
	core.CheckIfError(err)

	servers, err := config.FilterServers(allServers, serverArgs, serverFlags.Tags, serverFlags.Regex, serverFlags.Invert)
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
