package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/print"
)

func describeSpecsCmd(config *dao.Config, configErr *error) *cobra.Command {
	var specFlags core.SpecFlags

	cmd := cobra.Command{
		Aliases: []string{"spec"},
		Use:     "specs [specs]",
		Short:   "Describe specs",
		Long:    "Describe specs.",
		Example: `  # Describe all specs
  sake describe specs`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			describeSpecs(config, args, &specFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			specs := config.GetSpecNames()
			return specs, cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}
	cmd.Flags().SortFlags = false

	cmd.Flags().BoolVarP(&specFlags.Edit, "edit", "e", false, "edit spec")

	return &cmd
}

func describeSpecs(
	config *dao.Config,
	args []string,
	specFlags *core.SpecFlags,
) {
	if specFlags.Edit {
		if len(args) > 0 {
			err := config.EditSpec(args[0])
			core.CheckIfError(err)
		} else {
			err := config.EditSpec("")
			core.CheckIfError(err)
		}
	}

	var specs []dao.Spec
	if len(args) > 0 {
		t, err := config.GetSpecsByName(args)
		core.CheckIfError(err)
		specs = t
	} else {
		specs = config.Specs
	}

	print.PrintSpecBlocks(specs, false)
}
