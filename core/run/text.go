package run

import (
	"bufio"
	"fmt"
	"golang.org/x/term"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/print"
)

func (run *Run) Text(dryRun bool) {
	servers := run.Servers
	prefixMaxLen := calcMaxPrefixLength(run.LocalClients)

	var wg sync.WaitGroup
	for i := range servers {
		wg.Add(1)

		if run.Task.Spec.Parallel {
			go func(i int, wg *sync.WaitGroup) {
				defer wg.Done()
				// TODO
				_ = run.TextWork(i, prefixMaxLen, dryRun)
			}(i, &wg)
		} else {
			err := func(i int, wg *sync.WaitGroup) error {
				defer wg.Done()
				err := run.TextWork(i, prefixMaxLen, dryRun)
				return err
			}(i, &wg)

			if run.Task.Spec.AnyErrorsFatal && err != nil {
				break
			}
		}
	}

	wg.Wait()
}

func (run *Run) TextWork(rIndex int, prefixMaxLen int, dryRun bool) error {
	task := run.Task
	server := run.Servers[rIndex]
	prefix := getPrefixer(run.LocalClients[server.Host], rIndex, prefixMaxLen, task.Theme.Text, task.Spec.Parallel)

	numTasks := len(task.Tasks)

	var wg sync.WaitGroup
	for j, cmd := range task.Tasks {
		var client Client
		combinedEnvs := dao.MergeEnvs(server.Envs, cmd.Envs)
		if cmd.Local || server.Local {
			client = run.LocalClients[server.Host]
		} else {
			client = run.RemoteClients[server.Host]
		}

		var cmdString string
		if cmd.WorkDir != "" {
			cmdString = fmt.Sprintf("cd %s; %s", cmd.WorkDir, cmd.Cmd)
		} else if server.WorkDir != "" {
			cmdString = fmt.Sprintf("cd %s; %s", server.WorkDir, cmd.Cmd)
		} else {
			cmdString = cmd.Cmd
		}

		args := TaskContext{
			rIndex:   rIndex,
			cIndex:   j,
			client:   client,
			dryRun:   dryRun,
			env:      combinedEnvs,
			cmd:      cmdString,
			desc:     cmd.Desc,
			name:     cmd.Name,
			numTasks: numTasks,
		}

		err := RunTextCmd(args, task.Theme.Text, prefix, task.Spec.Parallel, &wg)
		if err != nil && !task.Spec.IgnoreErrors {
			return err
		}
	}

	wg.Wait()

	return nil
}

func RunTextCmd(t TaskContext, textStyle dao.Text, prefix string, parallel bool, wg *sync.WaitGroup) error {
	if textStyle.Header && !parallel {
		printHeader(t.cIndex, t.numTasks, t.name, t.desc, textStyle)
	}

	if t.dryRun {
		printCmd(prefix, t.cmd)
		return nil
	}

	err := t.client.Run(t.env, t.cmd)
	if err != nil {
		return err
	}

	// Copy over commands STDOUT.
	go func(client Client) {
		defer wg.Done()
		var err error
		if prefix != "" {
			_, err = io.Copy(os.Stdout, core.NewPrefixer(client.Stdout(), prefix))
		} else {
			_, err = io.Copy(os.Stdout, client.Stdout())
		}

		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
	}(t.client)
	wg.Add(1)

	// Copy over tasks's STDERR.
	go func(client Client) {
		defer wg.Done()
		var err error
		if prefix != "" {
			_, err = io.Copy(os.Stderr, core.NewPrefixer(client.Stderr(), prefix))
		} else {
			_, err = io.Copy(os.Stderr, client.Stderr())
		}
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
	}(t.client)
	wg.Add(1)

	wg.Wait()

	if err := t.client.Wait(); err != nil {
		if prefix != "" {
			fmt.Printf("%s%s\n", prefix, err.Error())
		} else {
			fmt.Printf("%s\n", err.Error())
		}

		return err
	}

	return nil
}

func printHeader(i int, numTasks int, name string, desc string, ts dao.Text) {
	var header string

	var prefixName string
	if name == "" {
		prefixName = "Command"
	} else {
		prefixName = name
	}

	var prefixPart1 string
	if numTasks > 1 {
		prefixPart1 = fmt.Sprintf("%s (%d/%d)", text.Bold.Sprintf(ts.HeaderPrefix), i+1, numTasks)
	} else {
		prefixPart1 = text.Bold.Sprintf(ts.HeaderPrefix)
	}

	var prefixPart2 string
	if desc != "" {
		prefixPart2 = fmt.Sprintf("%s: %s", text.Bold.Sprintf(prefixName), desc)
	} else {
		prefixPart2 = fmt.Sprintf("%s", text.Bold.Sprintf(prefixName))
	}

	width, _, err := term.GetSize(0)
	// Simply don't use width if there's an error
	if err != nil {
	}

	if prefixPart1 != "" {
		header = fmt.Sprintf("%s %s", prefixPart1, prefixPart2)
	} else {
		header = fmt.Sprintf("%s", prefixPart2)
	}
	headerLength := len(core.Strip(header))

	if width > 0 && ts.HeaderChar != "" {
		header = fmt.Sprintf("\n%s %s\n", header, strings.Repeat(ts.HeaderChar, width-headerLength-1))
	} else {
		header = fmt.Sprintf("\n%s\n", header)
	}
	fmt.Println(header)
}

func getPrefixer(client Client, i, prefixMaxLen int, textStyle dao.Text, parallel bool) string {
	if !textStyle.Prefix {
		return ""
	}

	prefix := client.Prefix()
	prefixLen := len(prefix)
	var prefixColor *text.Color
	if len(textStyle.PrefixColors) < 1 {
		prefixColor = print.GetFg("")
	} else {
		prefixColor = print.GetFg(textStyle.PrefixColors[i%len(textStyle.PrefixColors)])
	}

	if (!textStyle.Header || parallel) && len(prefix) < prefixMaxLen { // Left padding.
		prefixString := prefix + strings.Repeat(" ", prefixMaxLen-prefixLen) + " | "
		if prefixColor != nil {
			prefix = prefixColor.Sprintf(prefixString)
		} else {
			prefix = prefixString
		}
	} else {
		prefixString := prefix + " | "
		if prefixColor != nil {
			prefix = prefixColor.Sprintf(prefixString)
		} else {
			prefix = prefixString
		}
	}

	return prefix
}

func calcMaxPrefixLength(clients map[string]Client) int {
	var prefixMaxLen int = 0
	for _, c := range clients {
		prefix := c.Prefix()
		if len(prefix) > prefixMaxLen {
			prefixMaxLen = len(prefix)
		}
	}

	return prefixMaxLen
}

func printCmd(prefix string, cmd string) {
	scanner := bufio.NewScanner(strings.NewReader(cmd))
	for scanner.Scan() {
		fmt.Printf("%s%s\n", prefix, scanner.Text())
	}
}
