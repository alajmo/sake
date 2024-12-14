package run

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
)

type ServerTask struct {
	Server *dao.Server
	Task   *dao.Task
	Cmd    *dao.TaskCmd
	i      int
	j      int
}

func (run *Run) Table(dryRun bool) (dao.TableOutput, dao.ReportData, error) {
	task := run.Task
	servers := run.Servers
	uServers := run.UnreachableServers

	// TODO: data, reportData should be pointer?
	var data dao.TableOutput
	var reportData dao.ReportData
	data.Headers = append(reportData.Headers, "host")
	reportData.Headers = append(reportData.Headers, "host")
	// Append Command names if set
	for _, subTask := range task.Tasks {
		data.Headers = append(data.Headers, subTask.Name)
		reportData.Headers = append(reportData.Headers, subTask.Name)
	}

	// Populate the rows (server name is first cell, then commands and cmd output is set to empty string)
	for i, p := range servers {

		var client Client
		if p.Local {
			client = run.LocalClients[p.Name]
		} else {
			client = run.RemoteClients[p.Name]
		}

		title, err := getServerTitle(client, i, task.Theme.Table)
		if err != nil {
			return data, reportData, err
		}
		// p.Host
		data.Rows = append(data.Rows, dao.Row{Columns: []string{title}})
		reportData.Tasks = append(reportData.Tasks, dao.ReportRow{Name: title, Rows: []dao.Report{}})
		for range task.Tasks {
			data.Rows[i].Columns = append(data.Rows[i].Columns, "")
			reportData.Tasks[i].Rows = append(reportData.Tasks[i].Rows, dao.Report{})
		}
	}

	k := len(servers)
	for i, p := range uServers {
		reportData.Tasks = append(reportData.Tasks, dao.ReportRow{Name: p.Host, Rows: []dao.Report{}})
		for range task.Tasks {
			reportData.Tasks[k+i].Rows = append(reportData.Tasks[k+i].Rows, dao.Report{Status: dao.Unreachable})
		}
	}

	var err error
	switch task.Spec.Strategy {
	case "free":
		err = run.free(data, reportData, dryRun)
	case "host_pinned":
		err = run.hostPinned(data, reportData, dryRun)
	default:
		err = run.linear(data, reportData, dryRun)
	}

	reportData.Status = make(map[dao.TaskStatus]int, 5)
	for i := range reportData.Tasks {
		reportData.Tasks[i].Status = make(map[dao.TaskStatus]int, 5)
		for j := range reportData.Tasks[i].Rows {
			if reportData.Tasks[i].Rows[j].Status == dao.Unreachable {
				status := reportData.Tasks[i].Rows[j].Status
				reportData.Tasks[i].Status[status] = 1
				reportData.Status[status] += 1
				break
			} else {
				status := reportData.Tasks[i].Rows[j].Status
				reportData.Tasks[i].Status[status] += 1
				reportData.Status[status] += 1
			}
		}
	}

	if err != nil {
		switch err := err.(type) {
		case *ssh.ExitError:
			return data, reportData, &core.ExecError{Err: err, ExitCode: err.ExitStatus()}
		case *exec.ExitError:
			return data, reportData, &core.ExecError{Err: err, ExitCode: err.ExitCode()}
		default:
			return data, reportData, err
		}
	}

	return data, reportData, nil
}

func (run *Run) free(
	data dao.TableOutput,
	reportData dao.ReportData,
	dryRun bool,
) error {
	serverLen := len(run.Servers)
	taskLen := len(run.Task.Tasks)
	batch := int(run.Task.Spec.Batch)
	maxFailPercentage := run.Task.Spec.MaxFailPercentage
	var forks int
	if run.Task.Spec.Step {
		forks = 1
	} else {
		forks = CalcForks(batch, run.Task.Spec.Forks)
	}

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

	if remainder > 0 {
		quotient += 1
	}

	taskContinue := false
	failedHosts := make(chan bool, serverLen*taskLen)
	var mu sync.Mutex
	waitChan := make(chan struct{}, forks)
	for k := 0; k < quotient; k++ {
		var wg sync.WaitGroup
		errCh := make(chan error, batch*taskLen)

		start := k * batch * taskLen
		end := start + batch*taskLen

		if end > serverLen*taskLen {
			end = start + remainder*taskLen
		}

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

				if run.Task.Spec.Step && !taskContinue {
					taskOption, err := StepTaskExecute(r.Cmd.Name, r.Server.Host, &mu)
					if err != nil {
						errCh <- err
						failedHosts <- true
					}
					switch taskOption {
					case Yes:
					case No:
						<-waitChan
						return
					case Continue:
						taskContinue = true
					}
				}

				err := run.tableWork(r, r.j, register, data, reportData, dryRun)
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

func (run *Run) linear(
	data dao.TableOutput,
	reportData dao.ReportData,
	dryRun bool,
) error {
	serverLen := len(run.Servers)
	taskLen := len(run.Task.Tasks)
	batch := int(run.Task.Spec.Batch)
	var forks int
	if run.Task.Spec.Step {
		forks = 1
	} else {
		forks = CalcForks(batch, run.Task.Spec.Forks)
	}
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

	if remainder > 0 {
		quotient += 1
	}

	numFailed := 0
	taskContinue := false
	failedHosts := make(map[string]bool, serverLen)
	waitChan := make(chan struct{}, forks)
	var mu sync.Mutex
	for t := 0; t < taskLen; t++ {
		var wg sync.WaitGroup

		errCh := make(chan error, serverLen)

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

			for _, r := range runs[start:end] {
				if failedHosts[r.Server.Name] {
					continue
				}

				waitChan <- struct{}{}

				if run.Task.Spec.Step && !taskContinue {
					taskOption, err := StepTaskExecute(run.Task.Tasks[t].Name, r.Server.Host, &mu)
					if err != nil {
						return err
					}
					switch taskOption {
					case Yes:
					case No:
						<-waitChan
						continue
					case Continue:
						taskContinue = true
					}
				}

				wg.Add(1)

				go func(
					r ServerTask,
					register map[string]string,
					errCh chan<- error,
					wg *sync.WaitGroup,
				) {
					defer wg.Done()

					err := run.tableWork(r, 0, register, data, reportData, dryRun)
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
				}(r, register[r.Server.Name], errCh, &wg)
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

func (run *Run) hostPinned(
	data dao.TableOutput,
	reportData dao.ReportData,
	dryRun bool,
) error {
	serverLen := len(run.Servers)
	taskLen := len(run.Task.Tasks)
	batch := int(run.Task.Spec.Batch)
	var forks int
	if run.Task.Spec.Step {
		forks = 1
	} else {
		forks = CalcForks(batch, run.Task.Spec.Forks)
	}
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

	if remainder > 0 {
		quotient += 1
	}

	taskContinue := false
	failedHosts := make(chan bool, serverLen)
	waitChan := make(chan struct{}, forks)
	var mu sync.Mutex
	for k := 0; k < quotient; k++ {
		var wg sync.WaitGroup
		errCh := make(chan error, batch)

		start := k * batch * taskLen
		end := start + batch*taskLen

		if end > serverLen*taskLen {
			end = start + remainder*taskLen
		}

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
				for _, j := range r {
					waitChan <- struct{}{}

					if run.Task.Spec.Step && !taskContinue {
						taskOption, err := StepTaskExecute(j.Cmd.Name, j.Server.Host, &mu)
						if err != nil {
							errCh <- err
							failedHosts <- true
							break
						}
						switch taskOption {
						case Yes:
						case No:
							<-waitChan
							continue
						case Continue:
							taskContinue = true
						}
					}

					err := run.tableWork(j, 0, register[j.Server.Name], data, reportData, dryRun)
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

func (run *Run) tableWork(
	r ServerTask,
	si int,
	register map[string]string,
	data dao.TableOutput,
	reportData dao.ReportData,
	dryRun bool,
) error {
	var wg sync.WaitGroup

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

	shell := dao.SelectFirstNonEmpty((*r.Cmd).Shell, r.Task.Shell, r.Server.Shell, run.Config.Shell)
	shell = core.FormatShell(shell)
	workDir := getWorkDir((*r.Cmd).Local, (*r.Server).Local, (*r.Cmd).WorkDir, (*r.Server).WorkDir, (*r.Cmd).RootDir, (*r.Server).RootDir)
	t := TaskContext{
		rIndex:  r.i,
		cIndex:  r.j + 1, // first index (0) is server name
		client:  client,
		dryRun:  dryRun,
		env:     combinedEnvs,
		workDir: workDir,
		shell:   shell,
		cmd:     r.Cmd.Cmd,
		tty:     r.Cmd.TTY,
	}

	start := time.Now()
	out, stdout, stderr, err := runTableCmd(si, t, &wg)
	reportData.Tasks[r.i].Rows[r.j].Duration = time.Since(start)

	var errCode int
	switch err := err.(type) {
	case *ssh.ExitError:
		errCode = err.ExitStatus()
	case *exec.ExitError:
		errCode = err.ExitCode()
	}

	// TODO: Are mutex needed, perhaps if we're writing to the same buffer
	// dataMutex.Lock()
	// out, err := io.ReadAll(client.Stderr())
	// dataMutex.Unlock()
	if err != nil {
		switch r.Task.Spec.Print {
		case "stdout":
			data.Rows[t.rIndex].Columns[t.cIndex] = stdout
		case "stderr":
			data.Rows[t.rIndex].Columns[t.cIndex] = fmt.Sprintf("%s\n%s", stderr, err.Error())
		default:
			data.Rows[t.rIndex].Columns[t.cIndex] = fmt.Sprintf("%s\n%s", out, err.Error())
		}
	} else {
		switch r.Task.Spec.Print {
		case "stdout":
			data.Rows[t.rIndex].Columns[t.cIndex] = stdout
		case "stderr":
			data.Rows[t.rIndex].Columns[t.cIndex] = stderr
		default:
			data.Rows[t.rIndex].Columns[t.cIndex] = strings.TrimSuffix(out, "\n")
		}
	}

	reportData.Tasks[r.i].Rows[r.j].ReturnCode = errCode

	// TODO: Add skipped env variable
	if r.Cmd.Register != "" {
		register[r.Cmd.Register] = strings.TrimSuffix(out, "\n")
		register[r.Cmd.Register+"_stdout"] = stdout
		register[r.Cmd.Register+"_stderr"] = stderr
		register[r.Cmd.Register+"_rc"] = fmt.Sprint(reportData.Tasks[t.rIndex].Rows[r.j].ReturnCode)
		if err != nil {
			register[r.Cmd.Register+"_failed"] = "true"
			if r.Task.Spec.IgnoreErrors || r.Cmd.IgnoreErrors {
				register[r.Cmd.Register+"_status"] = "ignored"
			} else {
				register[r.Cmd.Register+"_status"] = "failed"
			}
		} else {
			register[r.Cmd.Register+"_failed"] = "false"
			register[r.Cmd.Register+"_status"] = "ok"
		}
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

func runTableCmd(i int, t TaskContext, wg *sync.WaitGroup) (string, string, string, error) {
	buf := new(bytes.Buffer)
	bufOut := new(bytes.Buffer)
	bufErr := new(bytes.Buffer)

	if t.dryRun {
		return t.cmd, bufOut.String(), bufErr.String(), nil
	}

	if t.tty {
		return buf.String(), bufOut.String(), bufErr.String(), ExecTTY(t.cmd, t.env)
	}

	err := t.client.Run(i, t.env, t.workDir, t.shell, t.cmd)
	if err != nil {
		return buf.String(), bufOut.String(), bufErr.String(), err
	}

	// Copy over commands STDOUT.
	var stdoutHandler = func(i int, client Client) {
		defer wg.Done()
		mw := io.MultiWriter(buf, bufOut)
		_, err = io.Copy(mw, client.Stdout(i))

		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
	}
	wg.Add(1)
	go stdoutHandler(i, t.client)

	// Copy over tasks's STDERR.
	var stderrHandler = func(i int, client Client) {
		defer wg.Done()
		mw := io.MultiWriter(buf, bufErr)
		_, err = io.Copy(mw, client.Stderr(i))
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
	}
	wg.Add(1)
	go stderrHandler(i, t.client)

	wg.Wait()

	if err := t.client.Wait(i); err != nil {
		return buf.String(), bufOut.String(), bufErr.String(), err
	}

	return buf.String(), bufOut.String(), bufErr.String(), nil
}

func getServerTitle(client Client, i int, ts dao.Table) (string, error) {
	if ts.Prefix == "" {
		return "", nil
	}

	name, host, user, port := client.Prefix()
	data := PrefixData{
		Name:  name,
		Host:  host,
		User:  user,
		Port:  port,
		Index: i,
	}
	prefix, err := PrefixTemplate(ts.Prefix, data)
	if err != nil {
		return "", err
	}

	return prefix, nil
}
