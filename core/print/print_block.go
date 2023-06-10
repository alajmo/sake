package print

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
)

func PrintServerList(servers []dao.Server) error {
	theme := dao.DEFAULT_THEME
	theme.Table.Options.DrawBorder = core.Ptr(false)
	theme.Table.Options.SeparateColumns = core.Ptr(false)
	theme.Table.Options.SeparateRows = core.Ptr(false)
	theme.Table.Options.SeparateHeader = core.Ptr(true)
	theme.Table.Options.SeparateFooter = core.Ptr(false)
	options := PrintTableOptions{
		Theme:            theme,
		Output:           "table",
		OmitEmptyRows:    true,
		OmitEmptyColumns: true,
	}

	headers := []string{"Server", "Host", "Bastion", "Tags"}
	rows := dao.GetTableData(servers, headers)
	err := PrintTable(rows, options, headers, []string{}, true, false)

	return err
}

func PrintServerBlocks(servers []dao.Server) {
	if len(servers) == 0 {
		return
	}

	fmt.Println()

	for i, server := range servers {
		output := ""

		output += printStringField("name", server.Name, false)
		if server.Name != server.Group {
			output += printStringField("group", server.Group, false)
		}
		output += printStringField("desc", server.Desc, false)
		output += printStringField("user", server.User, false)
		output += printStringField("host", server.Host, false)
		output += printNumberField("port", int(server.Port), false)
		if len(server.Bastions) > 0 {
			output += printBastion(server.Bastions)
		}

		output += printBoolField("local", server.Local, false)
		output += printStringField("shell", server.Shell, false)
		output += printStringField("work_dir", server.WorkDir, false)
		output += printSliceField("tags", server.Tags, false)

		fmt.Print(output)

		envs := server.GetNonDefaultEnvs()
		if envs != nil {
			printEnv(envs)
		}

		if i < len(servers)-1 {
			fmt.Print("\n--\n\n")
		}
	}

	fmt.Println()
}

func PrintTaskBlock(tasks []dao.Task) {
	if len(tasks) == 0 {
		return
	}

	fmt.Println()

	for i, task := range tasks {
		output := ""

		if task.ID == task.Name {
			output += printStringField("name", task.Name, false)
		} else {
			output += printStringField("task", task.ID, false)
			output += printStringField("name", task.Name, false)
		}
		output += printStringField("desc", task.Desc, false)
		output += printStringField("theme", task.Theme.Name, false)
		output += printStringField("shell", task.Shell, false)
		output += printStringField("work_dir", task.WorkDir, false)
		output += printBoolField("local", task.Local, false)
		output += printBoolField("tty", task.TTY, false)
		output += printBoolField("attach", task.Attach, false)

		fmt.Print(output)

		PrintSpecBlocks([]dao.Spec{task.Spec}, true)
		PrintTargetBlocks([]dao.Target{task.Target}, true)

		if task.Envs != nil {
			printEnv(task.Envs)
		}

		if task.Cmd != "" {
			fmt.Printf("cmd: \n")
			printCmd(task.Cmd)
		} else if len(task.Tasks) > 0 {
			fmt.Printf("tasks: \n")
			for i, st := range task.Tasks {
				if st.Name != "" {
					if st.Desc != "" {
						fmt.Printf("%3s - %s: %s\n", " ", st.Name, st.Desc)
					} else {
						fmt.Printf("%3s - %s\n", " ", st.Name)
					}
				} else {
					fmt.Printf("%3s - %s-%d\n", " ", "task", i)
				}
			}
		}

		if i < len(tasks)-1 {
			fmt.Print("\n--\n\n")
		}
	}

	if len(tasks) != 1 {
		fmt.Println()
	}
}

func PrintTargetBlocks(targets []dao.Target, indent bool) {
	if len(targets) == 0 {
		return
	}

	for i, target := range targets {
		output := ""
		output += printStringField("desc", target.Desc, indent)
		output += printBoolField("all", target.All, indent)
		output += printBoolField("invert", target.Invert, indent)
		output += printSliceField("servers", target.Servers, indent)
		output += printStringField("regex", target.Regex, indent)
		output += printSliceField("tags", target.Tags, indent)
		output += printNumberField("limit", int(target.Limit), indent)
		output += printNumberField("limit_p", int(target.LimitP), indent)

		if output == "" {
			continue
		}

		if indent {
			fmt.Printf("target:\n%s", output)
		} else {
			name := printStringField("name", target.Name, indent)
			fmt.Printf("%s%s", name, output)
		}

		if i < len(targets)-1 {
			fmt.Printf("\n--\n\n")
		}
	}
}

func PrintSpecBlocks(specs []dao.Spec, indent bool) {
	if len(specs) == 0 {
		return
	}

	for i, spec := range specs {
		output := ""
		output += printStringField("desc", spec.Desc, indent)
		output += printBoolField("describe", spec.Describe, indent)
		output += printBoolField("list_hosts", spec.ListHosts, indent)
		output += printStringField("order", spec.Order, indent)
		output += printBoolField("Silent", spec.Silent, indent)
		output += printBoolField("Hidden", spec.Hidden, indent)
		output += printStringField("strategy", spec.Strategy, indent)
		output += printNumberField("batch", int(spec.Batch), indent)
		output += printNumberField("batch_p", int(spec.BatchP), indent)
		output += printNumberField("forks", int(spec.Forks), indent)
		output += printStringField("output", spec.Output, indent)
		output += printStringField("print", spec.Print, indent)
		output += printNumberField("max_fail_percentage", int(spec.MaxFailPercentage), indent)
		output += printBoolField("any_errors_fatal", spec.AnyErrorsFatal, indent)
		output += printBoolField("ignore_errors", spec.IgnoreErrors, indent)
		output += printBoolField("ignore_unreachable", spec.IgnoreUnreachable, indent)
		output += printBoolField("omit_empty_rows", spec.OmitEmptyRows, indent)
		output += printBoolField("omit_empty_columns", spec.OmitEmptyColumns, indent)
		output += printSliceField("report", spec.Report, indent)
		output += printBoolField("verbose", spec.Verbose, indent)
		output += printBoolField("confirm", spec.Confirm, indent)
		output += printBoolField("step", spec.Step, indent)

		if output == "" {
			continue
		}

		if indent {
			fmt.Printf("spec:\n%s", output)
		} else {
			name := printStringField("name", spec.Name, indent)
			fmt.Printf("%s%s", name, output)
		}

		if i < len(specs)-1 {
			fmt.Printf("\n--\n\n")
		}
	}
}

func printCmd(cmd string) {
	scanner := bufio.NewScanner(strings.NewReader(cmd))
	for scanner.Scan() {
		fmt.Printf("%4s%s\n", " ", scanner.Text())
	}
}

func printEnv(env []string) {
	fmt.Printf("env: \n")
	for _, env := range env {
		fmt.Printf("%4s%s\n", " ", strings.Replace(strings.TrimSuffix(env, "\n"), "=", ": ", 1))
	}
}

func printBastion(bastions []dao.Bastion) string {
	if len(bastions) == 1 {
		return fmt.Sprintf("bastion: %s\n", bastions[0].GetPrint())
	}

	output := "bastions: \n"
	for _, bastion := range bastions {
		output += fmt.Sprintf("%4s- %s\n", " ", bastion.GetPrint())
	}
	return output
}

func printStringField(key string, value string, indent bool) string {
	if value != "" {
		if indent {
			return fmt.Sprintf("%4s%s: %s\n", " ", key, value)
		} else {
			return fmt.Sprintf("%s: %s\n", key, value)
		}
	}

	return ""
}

func printBoolField(key string, value bool, indent bool) string {
	if value {
		if indent {
			return fmt.Sprintf("%4s%s: %t\n", " ", key, value)
		} else {
			return fmt.Sprintf("%s: %t\n", key, value)
		}
	}
	return ""
}

func printSliceField(key string, value []string, indent bool) string {
	if len(value) > 0 {
		if indent {
			return fmt.Sprintf("%4s%s: %s\n", " ", key, strings.Join(value, ", "))
		} else {
			return fmt.Sprintf("%s: %s\n", key, strings.Join(value, ", "))
		}
	}
	return ""
}

func printNumberField(key string, value int, indent bool) string {
	if value > 0 {
		if indent {
			return fmt.Sprintf("%4s%s: %d\n", " ", key, value)
		} else {
			return fmt.Sprintf("%s: %d\n", key, value)
		}
	}
	return ""
}
