package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
)

func editTask(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Aliases: []string{"tasks", "tsk"},
		Use:     "task [task]",
		Short:   "Open up sake config file in $EDITOR and go to tasks section",
		Long:    `Open up sake config file in $EDITOR and go to tasks section.`,
		Example: `  # Edit tasks
  sake edit task

  # Edit task <task>
  sake edit task <task>`,
		Run: func(cmd *cobra.Command, args []string) {
			err := *configErr
			switch e := err.(type) {
			case *core.ConfigNotFound:
				core.CheckIfError(e)
			default:
				runEditTask(args, *config)
			}
		},
		Args: cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil || len(args) == 1 {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			return config.GetTaskIDAndDesc(), cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	return &cmd
}

func runEditTask(args []string, config dao.Config) {
	if len(args) > 0 {
		err := config.EditTask(args[0])
		core.CheckIfError(err)
	} else {
		err := config.EditTask("")
		core.CheckIfError(err)
	}
}
