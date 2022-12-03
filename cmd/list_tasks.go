package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/print"
)

var taskHeaders = []string{"task", "desc", "local", "tty", "attach", "work_dir", "shell", "spec", "target", "theme"}

func listTasksCmd(config *dao.Config, configErr *error, listFlags *core.ListFlags) *cobra.Command {
	var taskFlags core.TaskFlags

	cmd := cobra.Command{
		Aliases: []string{"task", "tsk", "tsks"},
		Use:     "tasks [tasks]",
		Short:   "List tasks",
		Long:    "List tasks.",
		Example: `  # List all tasks
  sake list tasks

  # List task <task>
  sake list task <task>`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			listTasks(config, args, listFlags, &taskFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			return config.GetTaskIDAndDesc(), cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().SortFlags = false

	cmd.Flags().BoolVarP(&taskFlags.AllHeaders, "all-headers", "H", false, "select all task headers")
	cmd.Flags().StringSliceVar(&taskFlags.Headers, "headers", []string{"task", "desc"}, "set headers")
	err := cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := taskHeaders
		return validHeaders, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)
	cmd.MarkFlagsMutuallyExclusive("all-headers", "headers")

	return &cmd
}

func listTasks(
	config *dao.Config,
	args []string,
	listFlags *core.ListFlags,
	taskFlags *core.TaskFlags,
) {
	tasks, err := config.GetTasksByIDs(args)
	core.CheckIfError(err)

	theme, err := config.GetTheme(listFlags.Theme)
	core.CheckIfError(err)

	if len(tasks) > 0 {
		options := print.PrintTableOptions{
			Output:           listFlags.Output,
			Theme:            *theme,
			OmitEmptyRows:    false,
			OmitEmptyColumns: true,
			Resource:         "task",
		}

		var headers []string
		if taskFlags.AllHeaders {
			headers = taskHeaders
		} else {
			headers = taskFlags.Headers
		}

		rows := dao.GetTableData(tasks, headers)
		err := print.PrintTable(rows, options, headers, []string{}, true, true)
		core.CheckIfError(err)
	}
}
