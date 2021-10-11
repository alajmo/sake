package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/yac/core"
	"github.com/alajmo/yac/core/dao"
	"github.com/alajmo/yac/core/print"
)

func listNetworksCmd(config *dao.Config, configErr *error, listFlags *print.ListFlags) *cobra.Command {
	var networkFlags print.ListNetworkFlags

	cmd := cobra.Command{
		Aliases: []string{"network", "net", "n"},
		Use:     "networks [flags]",
		Short:   "List networks",
		Long:    "List networks",
		Example: `  # List networks
  yac list networks`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			listNetworks(config, args, listFlags, &networkFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			names := config.GetDirNames()
			return names, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringSliceVarP(&networkFlags.Tags, "tags", "t", []string{}, "filter networks by their tag")
	err := cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetTags()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVar(&networkFlags.Headers, "headers", []string{"name", "hosts", "tags", "description"}, "Specify headers, defaults to name, tags, description")
	err = cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := []string{"name", "hosts", "tags", "description"}
		return validHeaders, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func listNetworks(
	config *dao.Config,
	args []string,
	listFlags *print.ListFlags,
	networkFlags *print.ListNetworkFlags,
) {
	// Table Style
	// switch config.Theme.Table {
	// case "ascii":
	// 	core.YacList.Box = core.StyleBoxASCII
	// default:
	// 	core.YacList.Box = core.StyleBoxDefault
	// }

	networksName := config.GetNetworksByName(args)
	networksTag := config.GetNetworksByTag(networkFlags.Tags)

	filteredNetworks := dao.GetIntersectNetworks(networksName, networksTag)

	print.PrintNetworks(filteredNetworks, *listFlags, *networkFlags)
}
