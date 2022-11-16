package run

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"math"

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

func (run *Run) Table(dryRun bool) (dao.TableOutput, error) {
	task := run.Task
	servers := run.Servers

	var data dao.TableOutput
	var dataExit dao.TableOutput
	var dataMutex = sync.RWMutex{}

	data.Headers = append(data.Headers, "server")
	dataExit.Headers = append(dataExit.Headers, "server")
	// Append Command names if set
	for _, subTask := range task.Tasks {
		data.Headers = append(data.Headers, subTask.Name)
		dataExit.Headers = append(dataExit.Headers, subTask.Name)
	}
	// Populate rows (server name is first cell, then commands and cmd output is set to empty string)
	for i, p := range servers {
		data.Rows = append(data.Rows, dao.Row{Columns: []string{p.Name}})
		dataExit.Rows = append(dataExit.Rows, dao.Row{Columns: []string{p.Name}})
		for range task.Tasks {
			data.Rows[i].Columns = append(data.Rows[i].Columns, "")
			dataExit.Rows[i].Columns = append(dataExit.Rows[i].Columns, "")
		}
	}

	var err error
	switch task.Spec.Strategy {
	case "free":
		err = run.free(&run.Config, data, dataExit, &dataMutex, dryRun)
	case "column":
		err = run.column(&run.Config, data, dataExit, &dataMutex, dryRun)
	default:
		err = run.row(&run.Config, data, dataExit, &dataMutex, dryRun)
	}

	if err != nil && run.Task.Spec.AnyErrorsFatal {
		switch err := err.(type) {
		case *ssh.ExitError:
			return data, &core.ExecError{Err: err, ExitCode: err.ExitStatus()}
		case *exec.ExitError:
			return data, &core.ExecError{Err: err, ExitCode: err.ExitCode()}
		default:
			return data, err
		}
	}

	return data, nil
}

func (run *Run) free(
	config *dao.Config,
	data dao.TableOutput,
	dataExit dao.TableOutput,
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

	fmt.Printf("Batch: %v\n", batch)
	fmt.Printf("Quotient: %v\n", quotient)
	fmt.Printf("Remainder: %v\n\n", remainder)

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

				err := run.tableWork(r, r.j, register, data, dataExit, dataMutex, dryRun)
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

func (run *Run) row(
	config *dao.Config,
	data dao.TableOutput,
	dataExit dao.TableOutput,
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

	fmt.Printf("Batch: %v\n", batch)
	fmt.Printf("Quotient: %v\n", quotient)
	fmt.Printf("Remainder: %v\n\n", remainder)

	if remainder > 0 {
		quotient += 1
	}

	numFailed := 0
	failedHosts := make(map[string]bool, serverLen)
	waitChan := make(chan struct{}, forks)
	for t := 0; t < taskLen; t++ {
		var wg sync.WaitGroup

		errCh := make(chan error, serverLen)

		for k := 0; k < quotient; k++ {
			failedHostsCh := make(chan struct {string; bool}, batch)

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
				wg.Add(1)

				go func(
					r ServerTask,
					register map[string]string,
					errCh chan<- error,
					failedHosts chan<- struct {string; bool},
					wg *sync.WaitGroup,
				) {
					defer wg.Done()

					err := run.tableWork(r, 0, register, data, dataExit, dataMutex, dryRun)
					<-waitChan
					if err != nil {
						errCh <- err
						failedHostsCh <- struct {string; bool} {r.Server.Name, true}
					} else {
						failedHostsCh <- struct {string; bool} {r.Server.Name, false}
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

func (run *Run) column(
	config *dao.Config,
	data dao.TableOutput,
	dataExit dao.TableOutput,
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

	fmt.Printf("Batch: %v\n", batch)
	fmt.Printf("Quotient: %v\n", quotient)
	fmt.Printf("Remainder: %v\n\n", remainder)

	if remainder > 0 {
		quotient += 1
	}

	failedHosts := make(chan bool, serverLen)
	waitChan := make(chan struct{}, forks)
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

					err := run.tableWork(j, 0, register[j.Server.Name], data, dataExit, dataMutex, dryRun)
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
	dataExit dao.TableOutput,
	dataMutex *sync.RWMutex,
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

	shell := dao.SelectFirstNonEmpty(r.Task.Shell, r.Server.Shell, run.Config.Shell)
	shell = core.FormatShell(shell)
	workDir := getWorkDir(*r.Cmd, *r.Server)
	t := TaskContext{
		rIndex:  r.i,
		cIndex:  r.j + 1, // first index (0) is server name
		client:  client,
		dryRun:  dryRun,
		env:     combinedEnvs,
		workDir: workDir,
		shell:   shell,
		cmd:     r.Cmd.Cmd,
		tty:     r.Task.TTY,
	}

	out, stdout, stderr, err := runTableCmd(si, t, &wg)

	// Add exit code to dataExit
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
		data.Rows[t.rIndex].Columns[t.cIndex] = fmt.Sprintf("%s\n%s", out, err.Error())
	} else {
		data.Rows[t.rIndex].Columns[t.cIndex] = strings.TrimSuffix(out, "\n")
	}

	dataExit.Rows[r.i].Columns[r.j+1] = fmt.Sprint(errCode)

	if r.Cmd.Register != "" {
		register[r.Cmd.Register] = strings.TrimSuffix(out, "\n")
		register[r.Cmd.Register+"_stdout"] = stdout
		register[r.Cmd.Register+"_stderr"] = stderr
		register[r.Cmd.Register+"_rc"] = dataExit.Rows[t.rIndex].Columns[r.j+1]
		if err != nil {
			register[r.Cmd.Register+"_failed"] = "true"
		} else {
			register[r.Cmd.Register+"_failed"] = "false"
		}
		// TODO: Add skipped env variable
	}

	if !r.Task.Spec.IgnoreErrors && err != nil {
		return err
	}

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
