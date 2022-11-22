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
	"math"
	"os"
	"os/exec"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/print"
)

func (run *Run) Text(dryRun bool) (dao.ReportData, error) {
	task := run.Task
	servers := run.Servers
	prefixMaxLen := calcMaxPrefixLength(run.LocalClients)

	// TODO: reportData should be pointer?
	var reportData dao.ReportData
	var dataMutex = sync.RWMutex{}
	reportData.Headers = append(reportData.Headers, "server")
	// Append Command names if set
	for _, subTask := range task.Tasks {
		reportData.Headers = append(reportData.Headers, subTask.Name)
	}
	// Populate the rows (server name is first cell, then commands and cmd output is set to empty string)
	for i, p := range servers {
		reportData.Tasks = append(reportData.Tasks, dao.ReportRow{Name: p.Name, Rows: []dao.Report{}})
		for range task.Tasks {
			reportData.Tasks[i].Rows = append(reportData.Tasks[i].Rows, dao.Report{})
		}
	}

	var err error
	switch task.Spec.Strategy {
	case "free":
		err = run.freeText(&run.Config, prefixMaxLen, reportData, &dataMutex, dryRun)
	case "column":
		err = run.columnText(&run.Config, prefixMaxLen, reportData, &dataMutex, dryRun)
	default:
		err = run.rowText(&run.Config, prefixMaxLen, reportData, &dataMutex, dryRun)
	}

	reportData.Status = make(map[dao.TaskStatus]int, 5)
	for i := range reportData.Tasks {
		reportData.Tasks[i].Status = make(map[dao.TaskStatus]int, 5)
		for j := range reportData.Tasks[i].Rows {
			status := reportData.Tasks[i].Rows[j].Status
			reportData.Tasks[i].Status[status] += 1
			reportData.Status[status] += 1
		}
	}

	if err != nil && run.Task.Spec.AnyErrorsFatal {
		switch err := err.(type) {
		case *ssh.ExitError:
			return reportData, &core.ExecError{Err: err, ExitCode: err.ExitStatus()}
		case *exec.ExitError:
			return reportData, &core.ExecError{Err: err, ExitCode: err.ExitCode()}
		default:
			return reportData, err
		}
	}

	return reportData, nil
}

func (run *Run) freeText(
	config *dao.Config,
	prefixMaxLen int,
	reportData dao.ReportData,
	dataMutex *sync.RWMutex,
	dryRun bool,
) error {
	serverLen := len(run.Servers)
	taskLen := len(run.Task.Tasks)
	batch := int(run.Task.Spec.Batch)
	forks := CalcFreeForks(batch, taskLen, run.Task.Spec.Forks)
	maxFailPercentage := run.Task.Spec.MaxFailPercentage

	register := make(map[string]map[string]string)
	var runs []ServerTask
	for i := range run.Servers {
		register[run.Servers[i].Name] = map[string]string{}
		for j := range run.Task.Tasks {
			runs = append(runs, ServerTask{
				Server: &run.Servers[i],
				Task:   run.Task,
				Cmd:    &run.Task.Tasks[j],
				i:      i,
				j:      j,
			})
		}
	}

	// calculate how many total tasks
	quotient, remainder := serverLen/batch, serverLen%batch

	// TODO: Remove
	// fmt.Printf("Batch: %v\n", batch)
	// fmt.Printf("Quotient: %v\n", quotient)
	// fmt.Printf("Remainder: %v\n\n", remainder)

	if remainder > 0 {
		quotient += 1
	}

	failedHosts := make(chan bool, serverLen*taskLen)
	waitChan := make(chan struct{}, forks)
	for k := 0; k < quotient; k++ {
		var wg sync.WaitGroup
		errCh := make(chan error, batch*taskLen)

		start := k * batch * taskLen
		end := start + batch*taskLen

		if end > serverLen*taskLen {
			end = start + remainder*taskLen
		}

		// For each server task
		for i := range runs[start:end] {
			wg.Add(1)
			go func(
				r ServerTask,
				register map[string]string,
				errCh chan<- error,
				failedHosts chan<- bool,
				wg *sync.WaitGroup,
			) {
				defer wg.Done()
				waitChan <- struct{}{}

				err := run.textWork(r, r.j, register, prefixMaxLen, reportData, dataMutex, dryRun)
				<-waitChan
				if err != nil {
					errCh <- err
					failedHosts <- true
				}
			}(runs[start+i], register[runs[start+i].Server.Name], errCh, failedHosts, &wg)
		}

		wg.Wait()

		percentageFailed := uint8(math.Floor(float64(len(failedHosts)) / float64(serverLen) * 100))
		if percentageFailed > maxFailPercentage {
			close(errCh)
			return <-errCh
		}

		close(errCh)
	}

	return nil
}

func (run *Run) rowText(
	config *dao.Config,
	prefixMaxLen int,
	reportData dao.ReportData,
	dataMutex *sync.RWMutex,
	dryRun bool,
) error {
	serverLen := len(run.Servers)
	taskLen := len(run.Task.Tasks)
	batch := int(run.Task.Spec.Batch)
	forks := CalcForks(batch, run.Task.Spec.Forks)
	maxFailPercentage := run.Task.Spec.MaxFailPercentage

	register := make(map[string]map[string]string)
	for i := range run.Servers {
		register[run.Servers[i].Name] = map[string]string{}
	}
	var runs []ServerTask
	for i := range run.Task.Tasks {
		for j := range run.Servers {
			runs = append(runs, ServerTask{
				Server: &run.Servers[j],
				Task:   run.Task,
				Cmd:    &run.Task.Tasks[i],
				i:      j,
				j:      i,
			})
		}
	}

	// calculate how many total tasks
	quotient, remainder := serverLen/batch, serverLen%batch

	// TODO: Remove
	// fmt.Printf("Batch: %v\n", batch)
	// fmt.Printf("Quotient: %v\n", quotient)
	// fmt.Printf("Remainder: %v\n\n", remainder)

	if remainder > 0 {
		quotient += 1
	}

	numFailed := 0
	failedHosts := make(map[string]bool, serverLen)
	waitChan := make(chan struct{}, forks)
	for t := 0; t < taskLen; t++ {
		var wg sync.WaitGroup

		errCh := make(chan error, serverLen)

		if run.Task.Theme.Text.Header != "" {
			err := printHeader(t, taskLen, run.Task.Tasks[t].Name, run.Task.Tasks[t].Desc, run.Task.Theme.Text)
			if err != nil {
				return err
			}
		}

		// Per batch
		for k := 0; k < quotient; k++ {
			failedHostsCh := make(chan struct {
				string
				bool
			}, batch)

			start := t*serverLen + k*batch
			end := start + batch

			if end > (t+1)*serverLen {
				end = start + remainder
			}

			// Per task
			for _, r := range runs[start:end] {
				if failedHosts[r.Server.Name] {
					continue
				}

				waitChan <- struct{}{}
				wg.Add(1)

				go func(
					r ServerTask,
					register map[string]string,
					errCh chan<- error,
					failedHosts chan<- struct {
						string
						bool
					},
					wg *sync.WaitGroup,
				) {
					defer wg.Done()

					err := run.textWork(r, 0, register, prefixMaxLen, reportData, dataMutex, dryRun)
					<-waitChan
					if err != nil {
						errCh <- err
						failedHostsCh <- struct {
							string
							bool
						}{r.Server.Name, true}
					} else {
						failedHostsCh <- struct {
							string
							bool
						}{r.Server.Name, false}
					}
				}(r, register[r.Server.Name], errCh, failedHostsCh, &wg)
			}

			wg.Wait()

			close(failedHostsCh)
			for p := range failedHostsCh {
				failedHosts[p.string] = p.bool
				if p.bool {
					numFailed += 1
				}
			}

			percentageFailed := uint8(math.Floor(float64(numFailed) / float64(serverLen) * 100))
			if percentageFailed > maxFailPercentage {
				close(errCh)
				return <-errCh
			}
		}

		close(errCh)
	}

	return nil
}

func (run *Run) columnText(
	config *dao.Config,
	prefixMaxLen int,
	reportData dao.ReportData,
	dataMutex *sync.RWMutex,
	dryRun bool,
) error {
	serverLen := len(run.Servers)
	taskLen := len(run.Task.Tasks)
	batch := int(run.Task.Spec.Batch)
	forks := CalcForks(batch, run.Task.Spec.Forks)
	maxFailPercentage := run.Task.Spec.MaxFailPercentage

	register := make(map[string]map[string]string)
	var runs []ServerTask
	for i := range run.Servers {
		register[run.Servers[i].Name] = map[string]string{}
		for j := range run.Task.Tasks {
			runs = append(runs, ServerTask{
				Server: &run.Servers[i],
				Task:   run.Task,
				Cmd:    &run.Task.Tasks[j],
				i:      i,
				j:      j,
			})
		}
	}

	// calculate how many total tasks
	quotient, remainder := serverLen/batch, serverLen%batch

	// TODO: Remove
	// fmt.Printf("Batch: %v\n", batch)
	// fmt.Printf("Quotient: %v\n", quotient)
	// fmt.Printf("Remainder: %v\n\n", remainder)

	if remainder > 0 {
		quotient += 1
	}

	failedHosts := make(chan bool, serverLen)
	waitChan := make(chan struct{}, forks)

	// Per batch
	for k := 0; k < quotient; k++ {
		var wg sync.WaitGroup
		errCh := make(chan error, batch)

		start := k * batch * taskLen
		end := start + batch*taskLen

		if end > serverLen*taskLen {
			end = start + remainder*taskLen
		}

		// Per server
		for t := start; t < end; t = t + taskLen {
			wg.Add(1)
			go func(
				r []ServerTask,
				register map[string]map[string]string,
				errCh chan<- error,
				failedHosts chan<- bool,
				wg *sync.WaitGroup,
			) {
				defer wg.Done()
				for i, j := range r {
					waitChan <- struct{}{}

					if run.Task.Theme.Text.Header != "" && batch == 1 {
						err := printHeader(i, taskLen, j.Cmd.Name, j.Cmd.Desc, run.Task.Theme.Text)
						if err != nil {
							fmt.Println(err)
							// TODO: What to do here
							// return err
						}
					}

					err := run.textWork(j, 0, register[j.Server.Name], prefixMaxLen, reportData, dataMutex, dryRun)
					<-waitChan
					if err != nil {
						errCh <- err
						failedHosts <- true
						break
					}
				}
			}(runs[t:t+taskLen], register, errCh, failedHosts, &wg)
		}

		wg.Wait()

		percentageFailed := uint8(math.Floor(float64(len(failedHosts)) / float64(serverLen) * 100))
		if percentageFailed > maxFailPercentage {
			close(errCh)
			return <-errCh
		}

		close(errCh)
	}

	return nil
}

func (run *Run) textWork(
	r ServerTask,
	si int,
	register map[string]string,
	prefixMaxLen int,
	reportData dao.ReportData,
	dataMutex *sync.RWMutex,
	dryRun bool,
) error {
	prefix := getPrefixer(run.LocalClients[r.Server.Name], r.i, prefixMaxLen, r.Task.Theme.Text)

	numTasks := len(r.Task.Tasks)

	var registerEnvs []string
	for k, v := range register {
		envStdout := fmt.Sprintf("%v=%v", k, v)
		registerEnvs = append(registerEnvs, envStdout)
	}
	combinedEnvs := dao.MergeEnvs(r.Cmd.Envs, r.Server.Envs, registerEnvs)
	var client Client
	if r.Cmd.Local || r.Server.Local {
		client = run.LocalClients[r.Server.Name]
	} else {
		client = run.RemoteClients[r.Server.Name]
	}

	shell := dao.SelectFirstNonEmpty(r.Task.Shell, r.Server.Shell, run.Config.Shell)
	shell = core.FormatShell(shell)
	workDir := getWorkDir(*r.Cmd, *r.Server)
	t := TaskContext{
		rIndex:   r.i,
		cIndex:   r.j,
		client:   client,
		dryRun:   dryRun,
		env:      combinedEnvs,
		workDir:  workDir,
		shell:    shell,
		cmd:      r.Cmd.Cmd,
		desc:     r.Cmd.Desc,
		name:     r.Cmd.Name,
		numTasks: numTasks,
		tty:      r.Cmd.TTY,
	}

	start := time.Now()
	var wg sync.WaitGroup
	out, stdout, stderr, err := runTextCmd(si, t, r.Task.Theme.Text, prefix, r.Cmd.Register, &wg)
	reportData.Tasks[r.i].Rows[r.j].Duration = time.Since(start)

	// Add exit code to reportData
	var errCode int
	switch err := err.(type) {
	case *ssh.ExitError:
		errCode = err.ExitStatus()
	case *exec.ExitError:
		errCode = err.ExitCode()
	case *template.ExecError:
		return err
	case *core.TemplateParseError:
		return err
	}

	reportData.Tasks[r.i].Rows[r.j].ReturnCode = errCode

	if r.Cmd.Register != "" {
		register[r.Cmd.Register] = strings.TrimSuffix(out, "\n")
		register[r.Cmd.Register+"_stdout"] = stdout
		register[r.Cmd.Register+"_stderr"] = stderr
		register[r.Cmd.Register+"_rc"] = fmt.Sprint(reportData.Tasks[t.rIndex].Rows[r.j+1].ReturnCode)
		if err != nil {
			register[r.Cmd.Register+"_failed"] = "true"
		} else {
			register[r.Cmd.Register+"_failed"] = "false"
		}
		// TODO: Add skipped env variable
	}

	if err != nil {
		if r.Task.Spec.IgnoreErrors || r.Cmd.IgnoreErrors {
			reportData.Tasks[r.i].Rows[r.j].Status = dao.Ignored
			return nil
		} else {
			reportData.Tasks[r.i].Rows[r.j].Status = dao.Failed
			return err
		}
	}

	reportData.Tasks[r.i].Rows[r.j].Status = dao.Ok

	return nil
}

func runTextCmd(
	i int,
	t TaskContext,
	textStyle dao.Text,
	prefix string,
	register string,
	wg *sync.WaitGroup,
) (string, string, string, error) {
	buf := new(bytes.Buffer)
	bufOut := new(bytes.Buffer)
	bufErr := new(bytes.Buffer)

	if t.dryRun {
		printCmd(prefix, t.cmd)
		return buf.String(), bufOut.String(), bufErr.String(), nil
	}

	if t.tty {
		return buf.String(), bufOut.String(), bufErr.String(), ExecTTY(t.cmd, t.env)
	}

	err := t.client.Run(i, t.env, t.workDir, t.shell, t.cmd)
	if err != nil {
		return buf.String(), bufOut.String(), bufErr.String(), err
	}

	// Copy over commands STDOUT.
	go func(client Client) {
		defer wg.Done()
		var err error

		if register == "" {
			if prefix != "" {
				_, err = io.Copy(os.Stdout, core.NewPrefixer(client.Stdout(i), prefix))
			} else {
				_, err = io.Copy(os.Stdout, client.Stdout(i))
			}
		} else {
			mw := io.MultiWriter(buf, bufOut)
			r := io.TeeReader(client.Stdout(i), mw)
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
				_, err = io.Copy(os.Stderr, core.NewPrefixer(client.Stderr(i), prefix))
			} else {
				_, err = io.Copy(os.Stderr, client.Stderr(i))
			}
		} else {
			mw := io.MultiWriter(buf, bufErr)
			r := io.TeeReader(client.Stderr(i), mw)
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

	if err := t.client.Wait(i); err != nil {
		if prefix != "" {
			fmt.Printf("%s%s\n", prefix, err.Error())
		} else {
			fmt.Printf("%s\n", err.Error())
		}

		return buf.String(), bufOut.String(), bufErr.String(), err
	}

	return buf.String(), bufOut.String(), bufErr.String(), nil
}

func headerTemplate(header string, data HeaderData) (string, error) {
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
	header, err := headerTemplate(ts.Header, data)
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

func getPrefixer(client Client, i, prefixMaxLen int, textStyle dao.Text) string {
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

	if textStyle.Header == "" && len(prefix) < prefixMaxLen { // Left padding.
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
