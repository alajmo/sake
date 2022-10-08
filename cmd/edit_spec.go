package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
)

func editSpec(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Aliases: []string{"specs", "sp"},
		Use:     "spec [spec]",
		Short:   "Edit spec",
		Long:    `Open up sake config file in $EDITOR and go to specs section.`,
		Example: `  # Edit specs
  sake edit spec

  # Edit spec <spec>
  sake edit spec <spec>`,
		Run: func(cmd *cobra.Command, args []string) {
			err := *configErr
			switch e := err.(type) {
			case *core.ConfigNotFound:
				core.CheckIfError(e)
			default:
				runEditSpec(args, *config)
			}
		},
		Args: cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil || len(args) == 1 {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			return config.GetSpecNames(), cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	return &cmd
}

func runEditSpec(args []string, config dao.Config) {
	if len(args) > 0 {
		err := config.EditSpec(args[0])
		core.CheckIfError(err)
	} else {
		err := config.EditSpec("")
		core.CheckIfError(err)
	}
}
