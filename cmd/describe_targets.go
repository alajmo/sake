package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/print"
)

func describeTargetsCmd(config *dao.Config, configErr *error) *cobra.Command {
	var targetFlags core.TargetFlags

	cmd := cobra.Command{
		Aliases: []string{"target"},
		Use:     "targets [targets]",
		Short:   "Describe targets",
		Long:    "Describe targets.",
		Example: `  # Describe all targets
  sake describe targets`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			describeTargets(config, args, &targetFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			targets := config.GetTargetNames()
			return targets, cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}
	cmd.Flags().SortFlags = false

	cmd.Flags().BoolVarP(&targetFlags.Edit, "edit", "e", false, "edit target")

	return &cmd
}

func describeTargets(
	config *dao.Config,
	args []string,
	targetFlags *core.TargetFlags,
) {
	if targetFlags.Edit {
		if len(args) > 0 {
			err := config.EditTarget(args[0])
			core.CheckIfError(err)
		} else {
			err := config.EditTarget("")
			core.CheckIfError(err)
		}
	}

	var targets []dao.Target
	if len(args) > 0 {
		t, err := config.GetTargetsByName(args)
		core.CheckIfError(err)
		targets = t
	} else {
		targets = config.Targets
	}

	print.PrintTargetBlocks(targets, false)
}
