package run

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/crypto/ssh"
	"golang.org/x/exp/slices"
	"golang.org/x/term"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"text/template"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/print"
)

func (run *Run) Text(dryRun bool) error {
	task := run.Task
	servers := run.Servers
	prefixMaxLen := calcMaxPrefixLength(run.LocalClients)

	var dataExit dao.TableOutput
	var dataMutex = sync.RWMutex{}
	dataExit.Headers = append(dataExit.Headers, "server")
	// Append Command names if set
	for _, subTask := range task.Tasks {
		dataExit.Headers = append(dataExit.Headers, subTask.Name)
	}
	// Populate the rows (server name is first cell, then commands and cmd output is set to empty string)
	for i, p := range servers {
		dataExit.Rows = append(dataExit.Rows, dao.Row{Columns: []string{p.Name}})

		for range task.Tasks {
			dataExit.Rows[i].Columns = append(dataExit.Rows[i].Columns, "")
		}
	}

	waitChan := make(chan struct{}, 100)
	var wg sync.WaitGroup
	for i := range servers {
		wg.Add(1)
		waitChan <- struct{}{}

		if run.Task.Spec.Parallel {
			go func(i int, wg *sync.WaitGroup) {
				defer wg.Done()
				// TODO: Handle errors when running tasks in parallel
				_ = run.TextWork(i, prefixMaxLen, dryRun, dataExit, &dataMutex)
				<-waitChan
			}(i, &wg)
		} else {
			err := func(i int, wg *sync.WaitGroup) error {
				defer wg.Done()
				err := run.TextWork(i, prefixMaxLen, dryRun, dataExit, &dataMutex)
				return err
			}(i, &wg)

			if err != nil {
				switch err.(type) {
				case *template.ExecError:
					return err
				case *core.TemplateParseError:
					return err
				default:
					if run.Task.Spec.AnyErrorsFatal {
						// Return proper exit code for failed tasks
						switch err := err.(type) {
						case *ssh.ExitError:
							return &core.ExecError{Err: err, ExitCode: err.ExitStatus()}
						case *exec.ExitError:
							return &core.ExecError{Err: err, ExitCode: err.ExitCode()}
						default:
							return err
						}
					}
				}
			}
		}
	}

	wg.Wait()

	return nil
}

func (run *Run) TextWork(rIndex int, prefixMaxLen int, dryRun bool, dataExit dao.TableOutput, dataMutex *sync.RWMutex) error {
	config := run.Config
	task := run.Task
	server := run.Servers[rIndex]
	prefix := getPrefixer(run.LocalClients[server.Name], rIndex, prefixMaxLen, task.Theme.Text, task.Spec.Parallel)

	numTasks := len(task.Tasks)

	var wg sync.WaitGroup
	register := make(map[string]string)
	var registers []string
	for j, cmd := range task.Tasks {
		var client Client
		combinedEnvs := dao.MergeEnvs(cmd.Envs, server.Envs, registers)
		if cmd.Local || server.Local {
			client = run.LocalClients[server.Name]
		} else {
			client = run.RemoteClients[server.Name]
		}

		shell := dao.SelectFirstNonEmpty(cmd.Shell, server.Shell, config.Shell)
		shell = core.FormatShell(shell)
		workDir := getWorkDir(cmd, server)
		args := TaskContext{
			rIndex:   rIndex,
			cIndex:   j,
			client:   client,
			dryRun:   dryRun,
			env:      combinedEnvs,
			workDir:  workDir,
			shell:    shell,
			cmd:      cmd.Cmd,
			desc:     cmd.Desc,
			name:     cmd.Name,
			numTasks: numTasks,
			tty:      cmd.TTY,
		}

		buf, bufOut, bufErr, err := RunTextCmd(args, task.Theme.Text, prefix, task.Spec.Parallel, dataExit, task.Tasks[j].Register, dataMutex, &wg)

		// Add exit code to dataExit
		var errCode int
		switch err := err.(type) {
		case *ssh.ExitError:
			errCode = err.ExitStatus()
		case *exec.ExitError:
			errCode = err.ExitCode()
		}

		dataExit.Rows[rIndex].Columns[j+1] = fmt.Sprint(errCode)

		// variable
		// stdout
		// stderr
		// rc
		// failed
		if task.Tasks[j].Register != "" {
			register[task.Tasks[j].Register] = buf
			register[task.Tasks[j].Register + "_stdout"] = bufOut
			register[task.Tasks[j].Register + "_stderr"] = bufErr
			register[task.Tasks[j].Register + "_rc"] = dataExit.Rows[rIndex].Columns[j+1]
			if err != nil {
				register[task.Tasks[j].Register + "_failed"] = "true"
			} else {
				register[task.Tasks[j].Register + "_failed"] = "false"
			}
			// TODO: Add skipped env variable

			registers = []string{}
			for  k, v := range register {
				envStdout := fmt.Sprintf("%v=%v", k, v)
				registers = append(registers, envStdout)
			}
		}

		switch err.(type) {
		case *template.ExecError:
			return err
		case *core.TemplateParseError:
			return err
		default:
			if !task.Spec.IgnoreErrors && err != nil {
				return err
			}
		}
	}

	wg.Wait()

	return nil
}

func RunTextCmd(t TaskContext, textStyle dao.Text, prefix string, parallel bool, dataExit dao.TableOutput, register string, dataMutex *sync.RWMutex, wg *sync.WaitGroup) (string, string, string, error) {
	buf := new(bytes.Buffer)
	bufOut := new(bytes.Buffer)
	bufErr := new(bytes.Buffer)

	if textStyle.Header != "" && !parallel {
		err := printHeader(t.cIndex, t.numTasks, t.name, t.desc, textStyle)
		if err != nil {
			return buf.String(), bufOut.String(), bufErr.String(), err
		}
	}

	if t.dryRun {
		printCmd(prefix, t.cmd)
		return buf.String(), bufOut.String(), bufErr.String(), nil
	}

	if t.tty {
		return buf.String(), bufOut.String(), bufErr.String(), ExecTTY(t.cmd, t.env)
	}

	err := t.client.Run(t.env, t.workDir, t.shell, t.cmd)
	if err != nil {
		return buf.String(), bufOut.String(), bufErr.String(), err
	}

	// Copy over commands STDOUT.
	go func(client Client) {
		defer wg.Done()
		var err error

		if register == "" {
			if prefix != "" {
				_, err = io.Copy(os.Stdout, core.NewPrefixer(client.Stdout(), prefix))
			} else {
				_, err = io.Copy(os.Stdout, client.Stdout())
			}
		} else {
			mw := io.MultiWriter(buf, bufOut)
			r := io.TeeReader(client.Stdout(), mw)
			// TODO: Refactor to NewReader: https://pkg.go.dev/golang.org/x/text/transform?utm_source=godoc#NewReader
			if prefix != "" {
				_, err = io.Copy(os.Stdout, core.NewPrefixer(r, prefix))
			} else {
				_, err = io.Copy(os.Stdout, r)
			}
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

		if register == "" {
			if prefix != "" {
				_, err = io.Copy(os.Stderr, core.NewPrefixer(client.Stderr(), prefix))
			} else {
				_, err = io.Copy(os.Stderr, client.Stderr())
			}
		} else {
			mw := io.MultiWriter(buf, bufErr)
			r := io.TeeReader(client.Stderr(), mw)
			if prefix != "" {
				_, err = io.Copy(os.Stderr, core.NewPrefixer(r, prefix))
			} else {
				_, err = io.Copy(os.Stderr, r)
			}
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

		return buf.String(), bufOut.String(), bufErr.String(), nil
	}

	return buf.String(), bufOut.String(), bufErr.String(), nil
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
