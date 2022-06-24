package run

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/exp/slices"
	"golang.org/x/term"
	"io"
	"os"
	"strings"
	"sync"
	"text/template"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/print"
)

func (run *Run) Text(dryRun bool) error {
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

			switch err.(type) {
			case *template.ExecError:
				return err
			case *core.TemplateParseError:
				return err
			default:
				if run.Task.Spec.AnyErrorsFatal && err != nil {
					// The error is printed for each server in method RunTextCmd.
					// We just return early so other tasks are not executed.
					return nil
				}
			}
		}
	}

	wg.Wait()

	return nil
}

func (run *Run) TextWork(rIndex int, prefixMaxLen int, dryRun bool) error {
	task := run.Task
	server := run.Servers[rIndex]
	prefix := getPrefixer(run.LocalClients[server.Host], rIndex, prefixMaxLen, task.Theme.Text, task.Spec.Parallel)

	numTasks := len(task.Tasks)

	var wg sync.WaitGroup
	for j, cmd := range task.Tasks {
		var client Client
		combinedEnvs := dao.MergeEnvs(cmd.Envs, server.Envs)
		if cmd.Local || server.Local {
			client = run.LocalClients[server.Host]
		} else {
			client = run.RemoteClients[server.Host]
		}

		workDir := getWorkDir(cmd, server)
		args := TaskContext{
			rIndex:   rIndex,
			cIndex:   j,
			client:   client,
			dryRun:   dryRun,
			env:      combinedEnvs,
			workDir:  workDir,
			cmd:      cmd.Cmd,
			desc:     cmd.Desc,
			name:     cmd.Name,
			numTasks: numTasks,
			tty:      cmd.TTY,
		}

		err := RunTextCmd(args, task.Theme.Text, prefix, task.Spec.Parallel, &wg)
		switch err.(type) {
		case *template.ExecError:
			return err
		case *core.TemplateParseError:
			return err
		default:
			if err != nil && !task.Spec.IgnoreErrors {
				return err
			}
		}
	}

	wg.Wait()

	return nil
}

func RunTextCmd(t TaskContext, textStyle dao.Text, prefix string, parallel bool, wg *sync.WaitGroup) error {
	if textStyle.Header != "" && !parallel {
		err := printHeader(t.cIndex, t.numTasks, t.name, t.desc, textStyle)
		if err != nil {
			return err
		}
	}

	if t.dryRun {
		printCmd(prefix, t.cmd)
		return nil
	}

	if t.tty {
		return ExecTTY(t.cmd, t.env)
	}

	err := t.client.Run(t.env, t.workDir, t.cmd)
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

func HeaderTemplate(header string, data HeaderData) (string, error) {
	tmpl, err := template.New("header.tmpl").Parse(header)
	if err != nil {
		return "", &core.TemplateParseError{Msg: err.Error()}
	}

	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, data)
	if err != nil {
		return "", &core.TemplateParseError{Msg: err.Error()}
	}

	s := buf.String()

	return s, nil
}

type HeaderData struct {
	Name     string
	Desc     string
	Index    int
	NumTasks int
}

func (h HeaderData) Style(s any, args ...string) string {
	v := core.AnyToString(s)
	colors := text.Colors{}

	for _, k := range args {
		switch {
		case strings.Contains(k, "fg_"):
			fg := print.GetFg(strings.TrimPrefix(k, "fg_"))
			colors = append(colors, *fg)
		case strings.Contains(k, "bg_"):
			bg := print.GetBg(strings.TrimPrefix(k, "bg_"))
			colors = append(colors, *bg)
		case slices.Contains([]string{"normal", "bold", "faint", "italic", "underline", "crossed_out"}, k):
			attr := print.GetAttr(k)
			colors = append(colors, *attr)
		}
	}

	return colors.Sprintf(v)
}

func printHeader(i int, numTasks int, name string, desc string, ts dao.Text) error {
	data := HeaderData{
		Name:     name,
		Desc:     desc,
		Index:    i + 1,
		NumTasks: numTasks,
	}
	header, err := HeaderTemplate(ts.Header, data)
	if err != nil {
		return err
	}

	width, _, _ := term.GetSize(0)
	headerLength := len(core.Strip(header))
	if width > 0 && ts.HeaderFiller != "" {
		header = fmt.Sprintf("\n%s%s\n", header, strings.Repeat(ts.HeaderFiller, width-headerLength-1))
	} else {
		header = fmt.Sprintf("\n%s\n", header)
	}

	fmt.Println(header)

	return nil
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

	if (textStyle.Header == "" || parallel) && len(prefix) < prefixMaxLen { // Left padding.
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
