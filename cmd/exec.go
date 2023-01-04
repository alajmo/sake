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
			setRunFlags.AnyErrorsFatal = cmd.Flags().Changed("any-errors-fatal")
			setRunFlags.Attach = cmd.Flags().Changed("attach")
			setRunFlags.Forks = cmd.Flags().Changed("forks")
			setRunFlags.Batch = cmd.Flags().Changed("batch")
			setRunFlags.BatchP = cmd.Flags().Changed("batch-p")
			setRunFlags.IgnoreErrors = cmd.Flags().Changed("ignore-error")
			setRunFlags.IgnoreUnreachable = cmd.Flags().Changed("ignore-unreachable")
			setRunFlags.Invert = cmd.Flags().Changed("invert")
			setRunFlags.Limit = cmd.Flags().Changed("limit")
			setRunFlags.LimitP = cmd.Flags().Changed("limit-p")
			setRunFlags.ListHosts = cmd.Flags().Changed("describe-hosts")
			setRunFlags.Order = cmd.Flags().Changed("order")
			setRunFlags.Local = cmd.Flags().Changed("local")
			setRunFlags.OmitEmptyRows = cmd.Flags().Changed("omit-empty-rows")
			setRunFlags.OmitEmptyColumns = cmd.Flags().Changed("omit-empty-columns")
			setRunFlags.Regex = cmd.Flags().Changed("regex")
			setRunFlags.Report = cmd.Flags().Changed("report")
			setRunFlags.Servers = cmd.Flags().Changed("servers")
			setRunFlags.Silent = cmd.Flags().Changed("silent")
			setRunFlags.Confirm = cmd.Flags().Changed("confirm")
			setRunFlags.Step = cmd.Flags().Changed("step")
			setRunFlags.TTY = cmd.Flags().Changed("tty")
			setRunFlags.Tags = cmd.Flags().Changed("tags")
			setRunFlags.Verbose = cmd.Flags().Changed("verbose")
			setRunFlags.MaxFailPercentage = cmd.Flags().Changed("max-fail-percentage")

			if setRunFlags.MaxFailPercentage {
				maxFailPercentage, err := cmd.Flags().GetUint8("max-fail-percentage")
				core.CheckIfError(err)
				if maxFailPercentage > 100 {
					core.Exit(&core.InvalidPercentInput{Name: "max-fail-percentage"})
				}
				runFlags.MaxFailPercentage = maxFailPercentage
			}

			if setRunFlags.Forks {
				forks, err := cmd.Flags().GetUint32("forks")
				core.CheckIfError(err)
				if forks == 0 {
					core.Exit(&core.ZeroNotAllowed{Name: "forks"})
				}
				runFlags.Forks = forks
			}

			if setRunFlags.Batch {
				batch, err := cmd.Flags().GetUint32("batch")
				core.CheckIfError(err)
				if batch == 0 {
					core.Exit(&core.ZeroNotAllowed{Name: "batch"})
				}
				runFlags.Batch = batch
			}

			if setRunFlags.BatchP {
				batchp, err := cmd.Flags().GetUint8("batch-p")
				core.CheckIfError(err)
				if batchp == 0 || batchp > 100 {
					core.Exit(&core.InvalidPercentInput2{Name: "batch-p"})
				}
				runFlags.BatchP = batchp
			}

			if setRunFlags.Limit {
				limit, err := cmd.Flags().GetUint32("limit")
				core.CheckIfError(err)
				if limit == 0 {
					core.Exit(&core.ZeroNotAllowed{Name: "limit"})
				}
				runFlags.Limit = limit
			}

			// Min-limit-p 1
			if setRunFlags.LimitP {
				limitp, err := cmd.Flags().GetUint8("limit-p")
				core.CheckIfError(err)
				if limitp == 0 || limitp > 100 {
					core.Exit(&core.InvalidPercentInput2{Name: "limit-p"})
				}
				runFlags.LimitP = limitp
			}

			execTask(args, config, &runFlags, &setRunFlags)
		},
		DisableAutoGenTag: true,
	}

	cmd.PersistentFlags().SortFlags = false
	cmd.Flags().SortFlags = false

	cmd.Flags().BoolVar(&runFlags.DryRun, "dry-run", false, "prints the command to see what will be executed")
	cmd.Flags().BoolVar(&runFlags.Describe, "describe", false, "print task information")
	cmd.Flags().BoolVar(&runFlags.ListHosts, "list-hosts", false, "print hosts that will be targetted")
	cmd.Flags().BoolVarP(&runFlags.Verbose, "verbose", "V", false, "enable all diagnostics")

	cmd.Flags().StringVarP(&runFlags.Strategy, "strategy", "S", "", "set execution strategy [linear|host_pinned|free]")
	err := cmd.RegisterFlagCompletionFunc("strategy", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		return strategies, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().Uint32P("forks", "f", 10000, "max number of concurrent processes")
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

	cmd.Flags().StringVar(&runFlags.Order, "order", "", "order hosts")
	err = cmd.RegisterFlagCompletionFunc("order", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		return orders, cobra.ShellCompDirectiveDefault
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

	cmd.Flags().StringVarP(&runFlags.Output, "output", "o", "", "set task output [text|table|table-2|table-3|table-4|html|markdown|json|csv|none]")
	err = cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		valid := []string{"text", "table", "table-2", "table-3", "table-4", "html", "markdown", "json", "csv", "none"}
		return valid, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringVarP(&runFlags.Print, "print", "p", "", "set print [all|stdout|stderr]")
	err = cmd.RegisterFlagCompletionFunc("print", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		valid := []string{"all", "stdout", "stderr"}
		return valid, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().BoolVar(&runFlags.OmitEmptyRows, "omit-empty-rows", false, "omit empty row for table output")
	cmd.Flags().BoolVar(&runFlags.OmitEmptyColumns, "omit-empty-columns", false, "omit empty column for table output")
	cmd.Flags().BoolVarP(&runFlags.Silent, "silent", "q", false, "omit showing loader when running tasks")
	cmd.Flags().BoolVar(&runFlags.Confirm, "confirm", false, "confirm root task before running")
	cmd.Flags().BoolVar(&runFlags.Step, "step", false, "confirm each task before running")
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

	cmd.Flags().StringSliceVarP(&runFlags.Report, "report", "R", []string{"recap"}, "reports to show")
	err = cmd.RegisterFlagCompletionFunc("report", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		return reports, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringVarP(&runFlags.IdentityFile, "identity-file", "i", "", "set identity file for all servers")
	cmd.Flags().StringVarP(&runFlags.User, "user", "U", "", "set ssh user")
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

		spec, err := config.GetSpec("default")
		core.CheckIfError(err)
		tt, err := config.GetTarget("default")
		core.CheckIfError(err)

		task := dao.Task{Spec: *spec, Target: *tt, Tasks: []dao.TaskCmd{cmd}, ID: "output", Name: "output"}
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
		err = target.RunTask([]string{}, runFlags, setRunFlags)
		core.CheckIfError(err)
	}
}
