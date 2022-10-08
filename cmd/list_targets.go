package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/print"
)

func listTargetsCmd(config *dao.Config, configErr *error, listFlags *core.ListFlags) *cobra.Command {
	var targetFlags core.TargetFlags

	cmd := cobra.Command{
		Aliases: []string{"target"},
		Use:     "targets [targets]",
		Short:   "List targets",
		Long:    "List targets.",
		Example: `  # List all targets
  sake list targets`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			listTargets(config, args, listFlags, &targetFlags)
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

	cmd.Flags().StringSliceVar(&targetFlags.Headers, "headers", []string{"name", "regex"}, "set headers. Available headers: name, regex")
	err := cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := []string{"name", "regex"}
		return validHeaders, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func listTargets(
	config *dao.Config,
	args []string,
	listFlags *core.ListFlags,
	targetFlags *core.TargetFlags,
) {
	theme, err := config.GetTheme(listFlags.Theme)
	core.CheckIfError(err)

	options := print.PrintTableOptions{
		Output:               listFlags.Output,
		Theme:                *theme,
		OmitEmpty:            false,
		SuppressEmptyColumns: true,
	}

	var targets []dao.Target
	if len(args) > 0 {
		t, err := config.GetTargetsByName(args)
		core.CheckIfError(err)
		targets = t
	} else {
		targets = config.Targets
	}

	if len(targets) > 0 {
		print.PrintTable("", targets, options, targetFlags.Headers, []string{})
	}
}
