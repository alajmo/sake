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
			setRunFlags.All = cmd.Flags().Changed("all")
			setRunFlags.Invert = cmd.Flags().Changed("invert")
			setRunFlags.Local = cmd.Flags().Changed("local")
			setRunFlags.TTY = cmd.Flags().Changed("tty")
			setRunFlags.OmitEmpty = cmd.Flags().Changed("omit-empty")
			setRunFlags.AnyErrorsFatal = cmd.Flags().Changed("any-errors-fatal")
			setRunFlags.IgnoreErrors = cmd.Flags().Changed("ignore-error")
			setRunFlags.IgnoreUnreachable = cmd.Flags().Changed("ignore_unreachable")

			maxFailPercentage, err := cmd.Flags().GetUint8("max-fail-percentage")
			core.CheckIfError(err)
			runFlags.MaxFailPercentage = maxFailPercentage

			forks, err := cmd.Flags().GetUint32("forks")
			core.CheckIfError(err)
			runFlags.Forks = forks

			batch, err := cmd.Flags().GetUint32("batch")
			core.CheckIfError(err)
			batchp, err := cmd.Flags().GetUint8("batch-p")
			core.CheckIfError(err)
			runFlags.Batch = batch
			runFlags.BatchP = batchp

			limit, err := cmd.Flags().GetUint32("limit")
			core.CheckIfError(err)
			limitp, err := cmd.Flags().GetUint8("limit-p")
			core.CheckIfError(err)

			runFlags.Limit = limit
			runFlags.LimitP = limitp

			execTask(args, config, &runFlags, &setRunFlags)
		},
		DisableAutoGenTag: true,
	}

	cmd.PersistentFlags().SortFlags = false
	cmd.Flags().SortFlags = false

	cmd.Flags().BoolVar(&runFlags.DryRun, "dry-run", false, "prints the command to see what will be executed")

	cmd.Flags().StringVarP(&runFlags.Strategy, "strategy", "S", "", "set execution strategy [free|row|column]")
	err := cmd.RegisterFlagCompletionFunc("strategy", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		valid := []string{"free", "row", "column"}
		return valid, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().Uint32P("forks", "f", 10000, "set maximal number of processes to run in parallel")
	cmd.Flags().Uint32P("batch", "b", 0, "set number of hosts to run in parallel")
	cmd.Flags().Uint8P("batch-p", "B", 0, "set percentage of servers to run in parallel [0-100]")
	cmd.MarkFlagsMutuallyExclusive("batch", "batch-p")

	cmd.Flags().BoolVarP(&runFlags.All, "all", "a", false, "target all servers")
	cmd.Flags().BoolVarP(&runFlags.Invert, "invert", "v", false, "invert matching on servers")
	cmd.Flags().StringVarP(&runFlags.Regex, "regex", "r", "", "filter servers on host regex")

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

	cmd.Flags().StringVarP(&runFlags.Target, "target", "T", "", "target servers by target name")
	err = cmd.RegisterFlagCompletionFunc("target", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		values := config.GetTargetNames()
		return values, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().Uint32P("limit", "l", 0, "set limit of servers to target")
	cmd.Flags().Uint8P("limit-p", "L", 0, "set percentage of servers to target")
	cmd.MarkFlagsMutuallyExclusive("limit", "limit-p")

	cmd.Flags().BoolVar(&runFlags.IgnoreUnreachable, "ignore-unreachable", false, "ignore unreachable hosts")
	cmd.Flags().Uint8P("max-fail-percentage", "M", 0, "stop task execution on all servers when threshold reached")
	cmd.Flags().BoolVar(&runFlags.AnyErrorsFatal, "any-errors-fatal", false, "stop task execution on all servers on error")
	cmd.MarkFlagsMutuallyExclusive("any-errors-fatal", "max-fail-percentage")
	cmd.Flags().BoolVar(&runFlags.IgnoreErrors, "ignore-errors", false, "continue task execution on errors")

	cmd.Flags().StringVarP(&runFlags.Spec, "spec", "J", "", "set spec")
	err = cmd.RegisterFlagCompletionFunc("spec", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		values := config.GetSpecNames()
		return values, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringVarP(&runFlags.Output, "output", "o", "", "set task output [text|table|table-2|table-3|table-4|html|markdown|json|csv]")
	err = cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		valid := []string{"text", "table", "table-2", "table-3", "table-4", "html", "markdown", "json", "csv"}
		return valid, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().BoolVar(&runFlags.OmitEmpty, "omit-empty", false, "omit empty results for table output")
	cmd.Flags().BoolVarP(&runFlags.Silent, "silent", "q", false, "omit showing loader when running tasks")
	cmd.PersistentFlags().StringVar(&runFlags.Theme, "theme", "default", "set theme")
	err = cmd.RegisterFlagCompletionFunc("theme", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		names := config.GetThemeNames()
		return names, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().BoolVar(&runFlags.TTY, "tty", false, "replace the current process")
	cmd.Flags().BoolVar(&runFlags.Attach, "attach", false, "ssh to server after command")
	cmd.Flags().BoolVar(&runFlags.Local, "local", false, "run command on localhost")
	cmd.MarkFlagsMutuallyExclusive("tty", "attach", "local")

	cmd.Flags().StringVarP(&runFlags.IdentityFile, "identity-file", "i", "", "set identity file for all servers")
	cmd.Flags().StringVar(&runFlags.Password, "password", "", "set ssh password for all servers")
	cmd.Flags().StringVar(&runFlags.KnownHostsFile, "known-hosts-file", "", "set known hosts file")

	return &cmd
}

func execTask(
	args []string,
	config *dao.Config,
	runFlags *core.RunFlags,
	setRunFlags *core.SetRunFlags,
) {
	err := config.ParseInventory([]string{})
	core.CheckIfError(err)

	servers, err := config.FilterServers(runFlags.All, runFlags.Servers, runFlags.Tags, runFlags.Regex, runFlags.Invert)
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
