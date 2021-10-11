package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/yac/core"
	"github.com/alajmo/yac/core/dao"
)

func editCmd(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Aliases: []string{"e", "ed"},
		Use:     "edit",
		Short:   "Edit yac config",
		Long:    `Edit yac config`,

		Example: `  # Edit current context
  yac edit

  # Edit specific yac config
  edit edit --config path/to/yac/config`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			runEdit(args, *config)
		},
	}

	cmd.AddCommand(
		editDir(config, configErr),
		editTask(config, configErr),
		editProject(config, configErr),
		editNetwork(config, configErr),
	)

	return &cmd
}

func runEdit(args []string, config dao.Config) {
	config.EditConfig()
}
