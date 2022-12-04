package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/print"
)

var tagHeaders = []string{"tag", "server"}

func listTagsCmd(config *dao.Config, configErr *error, listFlags *core.ListFlags) *cobra.Command {
	var tagFlags core.TagFlags

	cmd := cobra.Command{
		Aliases: []string{"tag"},
		Use:     "tags [tags]",
		Short:   "List tags",
		Long:    "List tags.",
		Example: `  # List all tags
  sake list tags`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			listTags(config, args, listFlags, &tagFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			tags := config.GetTags()
			return tags, cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().SortFlags = false

	cmd.Flags().StringSliceVar(&tagFlags.Headers, "headers", tagHeaders, "set headers")
	err := cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := tagHeaders
		return validHeaders, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func listTags(
	config *dao.Config,
	args []string,
	listFlags *core.ListFlags,
	tagFlags *core.TagFlags,
) {
	theme, err := config.GetTheme(listFlags.Theme)
	core.CheckIfError(err)

	options := print.PrintTableOptions{
		Output:           listFlags.Output,
		Theme:            *theme,
		OmitEmptyRows:    false,
		OmitEmptyColumns: true,
		Resource:         "tag",
	}

	allTags := config.GetTags()

	if len(args) > 0 {
		foundTags := core.Intersection(args, allTags)
		// Could not find one of the provided tags
		if len(foundTags) != len(args) {
			core.CheckIfError(&core.TagNotFound{Tags: args})
		}

		tags, err := config.GetTagAssocations(foundTags)
		core.CheckIfError(err)

		if len(tags) > 0 {
			err := print.PrintTable(tags, options, tagFlags.Headers, []string{}, true, true)
			core.CheckIfError(err)
		}
	} else {
		tags, err := config.GetTagAssocations(allTags)
		core.CheckIfError(err)
		if len(tags) > 0 {
			rows := dao.GetTableData(tags, tagFlags.Headers)
			err := print.PrintTable(rows, options, tagFlags.Headers, []string{}, true, true)
			core.CheckIfError(err)
		}
	}
}
