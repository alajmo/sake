package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/print"
)

var specHeaders = []string{"spec", "output", "parallel", "any_errors_fatal", "ignore_errors", "ignore_unreachable", "omit_empty"}

func listSpecsCmd(config *dao.Config, configErr *error, listFlags *core.ListFlags) *cobra.Command {
	var specFlags core.SpecFlags

	cmd := cobra.Command{
		Aliases: []string{"spec"},
		Use:     "specs [specs]",
		Short:   "List specs",
		Long:    "List specs.",
		Example: `  # List all specs
  sake list specs`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			listSpecs(config, args, listFlags, &specFlags)
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

	cmd.Flags().StringSliceVar(&specFlags.Headers, "headers", specHeaders, "set headers")
	err := cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := specHeaders
		return validHeaders, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func listSpecs(
	config *dao.Config,
	args []string,
	listFlags *core.ListFlags,
	specFlags *core.SpecFlags,
) {
	theme, err := config.GetTheme(listFlags.Theme)
	core.CheckIfError(err)

	options := print.PrintTableOptions{
		Output:               listFlags.Output,
		Theme:                *theme,
		OmitEmpty:            false,
		SuppressEmptyColumns: true,
		Resource:             "spec",
	}

	var specs []dao.Spec
	if len(args) > 0 {
		s, err := config.GetSpecsByName(args)
		core.CheckIfError(err)
		specs = s
	} else {
		specs = config.Specs
	}

	if len(specs) > 0 {
		rows := dao.GetTableData(specs, specFlags.Headers)
		print.PrintTable(rows, options, specFlags.Headers)
	}
}
