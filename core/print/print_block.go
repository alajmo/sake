package print

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/alajmo/sake/core/dao"
)

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
		output += printStringField("bastion", server.BastionHost, false)
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

		PrintTargetBlocks([]dao.Target{task.Target}, true)
		PrintSpecBlocks([]dao.Spec{task.Spec}, true, false)

		envs := task.GetNonDefaultEnvs()
		if envs != nil {
			printEnv(envs)
		}

		if task.Cmd != "" {
			fmt.Printf("cmd: \n")
			printCmd(task.Cmd)
		} else if len(task.Tasks) > 0 {
			fmt.Printf("tasks: \n")
			for _, st := range task.Tasks {
				if st.Name != "" {
					if st.Desc != "" {
						fmt.Printf("%3s - %s: %s\n", " ", st.Name, st.Desc)
					} else {
						fmt.Printf("%3s - %s\n", " ", st.Name)
					}
				} else {
					fmt.Printf("%3s - %s\n", " ", "cmd")
				}
			}
		}

		if i < len(tasks)-1 {
			fmt.Print("\n--\n\n")
		}
	}
	fmt.Println()
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

func PrintTargetBlocks(targets []dao.Target, indent bool) {
	if len(targets) == 0 {
		return
	}

	for i, target := range targets {
		output := ""
		output += printBoolField("all", target.All, indent)
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

func PrintSpecBlocks(specs []dao.Spec, indent bool, name bool) {
	if len(specs) == 0 {
		return
	}

	for i, spec := range specs {
		output := ""
		if name {
			printStringField("name", spec.Name, indent)
		}
		output += printStringField("output", spec.Output, indent)
		output += printBoolField("parallel", spec.Parallel, indent)
		output += printBoolField("any_errors_fatal", spec.AnyErrorsFatal, indent)
		output += printBoolField("ignore_errors", spec.IgnoreErrors, indent)
		output += printBoolField("ignore_unreachable", spec.IgnoreUnreachable, indent)
		output += printBoolField("omit_empty", spec.OmitEmpty, indent)

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
