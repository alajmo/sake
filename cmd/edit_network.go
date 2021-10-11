package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/yac/core"
	"github.com/alajmo/yac/core/dao"
)

func editNetwork(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Use:   "network",
		Short: "Edit yac network",
		Long:  `Edit yac network`,

		Example: `  # Edit a network called server
  yac edit network server

  # Edit network in specific yac config
  yac edit --config path/to/yac/config`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			runEditNetwork(args, *config)
		},
		Args: cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil || len(args) == 1 {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			values := config.GetNetworkNames()
			return values, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return &cmd
}

func runEditNetwork(args []string, config dao.Config) {
	if len(args) > 0 {
		config.EditNetworks(args[0])
	} else {
		config.EditNetworks("")
	}
}
