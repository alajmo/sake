package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/run"
)

func runCmd(config *dao.Config, configErr *error) *cobra.Command {
	var runFlags core.RunFlags
	var setRunFlags core.SetRunFlags

	cmd := cobra.Command{
		Use:   "run <task> [flags]",
		Short: "Run tasks",
		Long: `Run tasks specified in a sake.yaml file.`,
		Example: `  # Run task <task> for all servers
  sake run <task> --all

  # Run task <task> for servers <server>
  sake run <task> --servers <server>

  # Run task <task> for all servers that have tags <tag>
  sake run <task> --tags <tag>`,
		Args:                  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)

			// This is necessary since cobra doesn't support pointers for bools
			// (that would allow us to use nil as default value)
			setRunFlags.Local = cmd.Flags().Changed("local")
			setRunFlags.Parallel = cmd.Flags().Changed("parallel")
			setRunFlags.OmitEmpty = cmd.Flags().Changed("omit-empty")
			setRunFlags.AnyErrorsFatal = cmd.Flags().Changed("any-errors-fatal")
			setRunFlags.IgnoreErrors = cmd.Flags().Changed("ignore-errors")
			setRunFlags.IgnoreUnreachable = cmd.Flags().Changed("ignore-unreachable")

			runTask(args, config, &runFlags, &setRunFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			return config.GetTaskIDAndDesc(), cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().BoolVar(&runFlags.TTY, "tty", false, "replace the current process")
	cmd.Flags().BoolVar(&runFlags.Attach, "attach", false, "ssh to server after command")
	cmd.Flags().BoolVar(&runFlags.Local, "local", false, "run task on localhost")
	cmd.Flags().BoolVar(&runFlags.Describe, "describe", false, "print task information")
	cmd.Flags().BoolVar(&runFlags.DryRun, "dry-run", false, "print the task to see what will be executed")
	cmd.Flags().BoolVar(&runFlags.Debug, "debug", false, "enable debug mode")
	cmd.Flags().BoolVar(&runFlags.AnyErrorsFatal, "any-errors-fatal", false, "stop task execution on all servers on error")
	cmd.Flags().BoolVar(&runFlags.IgnoreErrors, "ignore-errors", false, "continue task execution on errors")
	cmd.Flags().BoolVar(&runFlags.IgnoreUnreachable, "ignore-unreachable", false, "ignore unreachable hosts")
	cmd.Flags().BoolVar(&runFlags.OmitEmpty, "omit-empty", false, "omit empty results for table output")
	cmd.Flags().BoolVarP(&runFlags.Parallel, "parallel", "p", false, "run server tasks in parallel")
	cmd.Flags().BoolVarP(&runFlags.Edit, "edit", "e", false, "edit task")
	cmd.Flags().StringVarP(&runFlags.IdentityFile, "identity-file", "i", "", "set identity file for all servers")
	cmd.Flags().StringVar(&runFlags.Password, "password", "", "set ssh password for all servers")
	cmd.Flags().StringVar(&runFlags.KnownHostsFile, "known-hosts-file", "", "set known hosts file")

	cmd.Flags().StringVarP(&runFlags.Output, "output", "o", "", "set task output [text|table|html|markdown]")
	err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		valid := []string{"text", "table", "html", "markdown"}
		return valid, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().BoolVarP(&runFlags.All, "all", "a", false, "target all servers")

	cmd.Flags().StringSliceVarP(&runFlags.Servers, "servers", "s", []string{}, "target servers by names")
	err = cmd.RegisterFlagCompletionFunc("servers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		servers := config.GetServerNameAndDesc()
		return servers, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVarP(&runFlags.Tags, "tags", "t", []string{}, "target servers by tags")
	err = cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		tags := config.GetTags()
		return tags, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.PersistentFlags().StringVar(&runFlags.Theme, "theme", "", "set theme")
	err = cmd.RegisterFlagCompletionFunc("theme", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		names := config.GetThemeNames()

		return names, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func runTask(
	args []string,
	config *dao.Config,
	runFlags *core.RunFlags,
	setRunFlags *core.SetRunFlags,
) {
	var taskIDs []string
	var userArgs []string
	// Seperate user arguments from task ids
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			userArgs = append(userArgs, arg)
		} else {
			taskIDs = append(taskIDs, arg)
		}
	}

	if runFlags.Edit {
		if len(args) > 0 {
			err := config.EditTask(taskIDs[0])
			core.CheckIfError(err)
		} else {
			err := config.EditTask("")
			core.CheckIfError(err)
		}
	}

	for _, taskID := range taskIDs {
		task, err := config.GetTask(taskID)
		core.CheckIfError(err)

		servers, err := config.GetTaskServers(task, runFlags)
		core.CheckIfError(err)

		if len(servers) == 0 {
			fmt.Println("No targets")
		} else {
			target := run.Run{Servers: servers, Task: task, Config: *config}
			err := target.RunTask(userArgs, runFlags, setRunFlags)
			core.CheckIfError(err)
		}
	}
}
