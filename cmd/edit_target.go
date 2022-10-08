package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
)

func editTarget(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Aliases: []string{"targets", "targ"},
		Use:     "target [target]",
		Short:   "Edit target",
		Long:    `Open up sake config file in $EDITOR and go to targets section.`,
		Example: `  # Edit targets
  sake edit target

  # Edit target <target>
  sake edit target <target>`,
		Run: func(cmd *cobra.Command, args []string) {
			err := *configErr
			switch e := err.(type) {
			case *core.ConfigNotFound:
				core.CheckIfError(e)
			default:
				runEditTarget(args, *config)
			}
		},
		Args: cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil || len(args) == 1 {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			return config.GetTargetNames(), cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	return &cmd
}

func runEditTarget(args []string, config dao.Config) {
	if len(args) > 0 {
		err := config.EditTarget(args[0])
		core.CheckIfError(err)
	} else {
		err := config.EditTarget("")
		core.CheckIfError(err)
	}
}
