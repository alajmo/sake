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
		fmt.Printf("Name: %s\n", server.Name)
		fmt.Printf("User: %s\n", server.User)
		fmt.Printf("Host: %s\n", server.Host)
		fmt.Printf("Port: %d\n", server.Port)
		fmt.Printf("Local: %t\n", server.Local)
		fmt.Printf("WorkDir: %s\n", server.WorkDir)
		fmt.Printf("Desc: %s\n", server.Desc)

		if len(server.Tags) > 0 {
			fmt.Printf("Tags: %s\n", server.GetValue("Tag", 0))
		}

		if len(server.Envs) > 0 {
			printEnv(server.Envs)
		}

		if i < len(servers)-1 {
			fmt.Printf("\n--\n\n")
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
		fmt.Printf("Task: %s\n", task.ID)
		fmt.Printf("Name: %s\n", task.Name)
		fmt.Printf("Desc: %s\n", task.Desc)
		fmt.Printf("Local: %t\n", task.Local)
		fmt.Printf("WorkDir: %s\n", task.WorkDir)
		fmt.Printf("Theme: %s\n", task.Theme.Name)
		fmt.Printf("Target: \n")
		fmt.Printf("%4sAll: %t\n", " ", task.Target.All)
		fmt.Printf("%4sServers: %s\n", " ", strings.Join(task.Target.Servers, ", "))
		fmt.Printf("%4sTags: %s", " ", strings.Join(task.Target.Tags, ", "))

		fmt.Println()

		fmt.Printf("Spec: \n")
		fmt.Printf("%4sOutput: %s\n", "", task.Spec.Output)
		fmt.Printf("%4sParallel: %t\n", "", task.Spec.Parallel)
		fmt.Printf("%4sAnyErrorsFatal: %t\n", "", task.Spec.AnyErrorsFatal)
		fmt.Printf("%4sIgnoreErrors: %t\n", "", task.Spec.IgnoreErrors)
		fmt.Printf("%4sIgnoreUnreachable: %t\n", "", task.Spec.IgnoreUnreachable)
		fmt.Printf("%4sOmitEmpty: %t", "", task.Spec.OmitEmpty)

		fmt.Println()

		if len(task.Envs) > 0 {
			printEnv(task.Envs)
		}

		if task.Cmd != "" {
			fmt.Printf("Cmd: \n")
			printCmd(task.Cmd)
		} else if len(task.Tasks) > 0 {
			fmt.Printf("Tasks: \n")
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
			fmt.Printf("\n--\n\n")
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
	fmt.Printf("Env: \n")
	for _, env := range env {
		fmt.Printf("%4s%s\n", " ", strings.Replace(strings.TrimSuffix(env, "\n"), "=", ": ", 1))
	}
}
