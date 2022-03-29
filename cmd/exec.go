package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/run"
)

func execCmd(config *dao.Config, configErr *error) *cobra.Command {
	var runFlags core.RunFlags
	var setRunFlags core.SetRunFlags

	cmd := cobra.Command{
		Use:   "exec <command>",
		Short: "Execute arbitrary commands",
		Long: `Execute arbitrary commands.

Single quote your command if you don't want the
file globbing and environments variables expansion to take place
before the command gets executed in each directory.`,
		Example: `  # List files in all servers
  sake exec --all ls

  # List git files that have markdown suffix for all servers
  sake exec --all 'git ls-files | grep -e ".md"'`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)

			// This is necessary since cobra doesn't support pointers for bools
			// (that would allow us to use nil as default value)
			setRunFlags.Local = cmd.Flags().Changed("local")
			setRunFlags.Parallel = cmd.Flags().Changed("parallel")
			setRunFlags.OmitEmpty = cmd.Flags().Changed("omit-empty")
			setRunFlags.AnyErrorsFatal = cmd.Flags().Changed("any-errors-fatal")
			setRunFlags.IgnoreErrors = cmd.Flags().Changed("ignore-error")
			setRunFlags.IgnoreUnreachable = cmd.Flags().Changed("ignore_unreachable")

			execTask(args, config, &runFlags, &setRunFlags)
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().BoolVar(&runFlags.TTY, "tty", false, "replace the currenty process")
	cmd.Flags().BoolVar(&runFlags.Attach, "attach", false, "ssh to server after command")
	cmd.Flags().BoolVar(&runFlags.Local, "local", false, "run command on localhost")
	cmd.Flags().BoolVar(&runFlags.DryRun, "dry-run", false, "prints the command to see what will be executed")
	cmd.Flags().BoolVar(&runFlags.Debug, "debug", false, "enable debug mode")
	cmd.Flags().BoolVar(&runFlags.AnyErrorsFatal, "any-errors-fatal", false, "stop task execution on all servers on error")
	cmd.Flags().BoolVar(&runFlags.IgnoreErrors, "ignore-errors", false, "continue task execution on errors")
	cmd.Flags().BoolVar(&runFlags.IgnoreUnreachable, "ignore-unreachable", false, "ignore unreachable hosts")
	cmd.Flags().BoolVar(&runFlags.OmitEmpty, "omit-empty", false, "omit empty results for table output")
	cmd.Flags().BoolVarP(&runFlags.Parallel, "parallel", "p", false, "run server tasks in parallel")
	cmd.Flags().StringVarP(&runFlags.IdentityFile, "identity-file", "i", "", "set identity file for all servers")
	cmd.Flags().StringVar(&runFlags.Password, "password", "", "set ssh password for all servers")
	cmd.Flags().StringVar(&runFlags.KnownHostsFile, "known-hosts-file", "", "set known hosts file")

	cmd.Flags().StringVarP(&runFlags.Output, "output", "o", "", "set task output [text|table|markdown|html]")
	err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		valid := []string{"table", "markdown", "html"}
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

	cmd.PersistentFlags().StringVar(&runFlags.Theme, "theme", "default", "set theme")
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

func execTask(
	args []string,
	config *dao.Config,
	runFlags *core.RunFlags,
	setRunFlags *core.SetRunFlags,
) {
	servers, err := config.FilterServers(runFlags.All, runFlags.Servers, runFlags.Tags)
	core.CheckIfError(err)

	if len(servers) == 0 {
		fmt.Println("No targets")
	} else {
		cmdStr := strings.Join(args[0:], " ")

		cmd := dao.TaskCmd{
			Cmd: cmdStr,
		}

		task := dao.Task{Tasks: []dao.TaskCmd{cmd}, ID: "output", Name: "output"}
		taskErrors := make([]dao.ResourceErrors[dao.Task], 1)

		var configErr = ""
		for _, taskError := range taskErrors {
			if len(taskError.Errors) > 0 {
				configErr = fmt.Sprintf("%s%s", configErr, dao.FormatErrors(taskError.Resource, taskError.Errors))
			}
		}
		if configErr != "" {
			core.CheckIfError(errors.New(configErr))
		}

		target := run.Run{Servers: servers, Task: &task, Config: *config}
		err := target.RunTask([]string{}, runFlags, setRunFlags)
		core.CheckIfError(err)
	}
}
