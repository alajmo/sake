package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/print"
)

func describeTasksCmd(config *dao.Config, configErr *error) *cobra.Command {
	var taskFlags core.TaskFlags

	cmd := cobra.Command{
		Aliases: []string{"task", "tsk"},
		Use:     "tasks [tasks]",
		Short:   "Describe tasks",
		Long:    "Describe tasks.",
		Example: `  # Describe all tasks
  sake describe tasks

  # Describe task <task>
  sake describe task <task>`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			describe(config, args, taskFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			values := config.GetTaskIDAndDesc()
			return values, cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().SortFlags = false

	cmd.Flags().BoolVarP(&taskFlags.Edit, "edit", "e", false, "edit task")

	return &cmd
}

func describe(config *dao.Config, args []string, taskFlags core.TaskFlags) {
	if taskFlags.Edit {
		if len(args) > 0 {
			err := config.EditTask(args[0])
			core.CheckIfError(err)
		} else {
			err := config.EditTask("")
			core.CheckIfError(err)
		}
	} else {
		tasks, err := config.GetTasksByIDs(args)
		core.CheckIfError(err)

		if len(tasks) > 0 {
			for i := range tasks {
				for j := range tasks[i].Tasks {
					envs, err := dao.ParseTaskEnv(tasks[i].Tasks[j].Envs, []string{}, []string{}, []string{})
					core.CheckIfError(err)

					tasks[i].Tasks[j].Envs = envs
				}
			}

			print.PrintTaskBlock(tasks)
		}
	}
}
