package run

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
)

func (run *Run) Table(dryRun bool) (dao.TableOutput, error) {
	task := run.Task
	servers := run.Servers

	var data dao.TableOutput
	var dataExit dao.TableOutput
	var dataMutex = sync.RWMutex{}

	/**
	 ** Headers
	 **/
	data.Headers = append(data.Headers, "server")
	dataExit.Headers = append(dataExit.Headers, "server")

	// Append Command names if set
	for _, subTask := range task.Tasks {
		data.Headers = append(data.Headers, subTask.Name)
		dataExit.Headers = append(dataExit.Headers, subTask.Name)
	}

	// Populate the rows (server name is first cell, then commands and cmd output is set to empty string)
	for i, p := range servers {
		data.Rows = append(data.Rows, dao.Row{Columns: []string{p.Name}})
		dataExit.Rows = append(dataExit.Rows, dao.Row{Columns: []string{p.Name}})

		for range task.Tasks {
			data.Rows[i].Columns = append(data.Rows[i].Columns, "")
			dataExit.Rows[i].Columns = append(dataExit.Rows[i].Columns, "")
		}
	}

	var wg sync.WaitGroup
	/**
	 ** Values
	 **/

	waitChan := make(chan struct{}, 100)
	for i := range servers {
		wg.Add(1)
		waitChan <- struct{}{}

		if task.Spec.Parallel {
			go func(i int, wg *sync.WaitGroup) {
				defer wg.Done()
				// TODO: Handle errors when running tasks in parallel
				_ = run.TableWork(i, dryRun, data, dataExit, &dataMutex)
				<-waitChan
			}(i, &wg)
		} else {
			err := func(i int, wg *sync.WaitGroup) error {
				defer wg.Done()
				err := run.TableWork(i, dryRun, data, dataExit, &dataMutex)

				return err
			}(i, &wg)

			if err != nil && run.Task.Spec.AnyErrorsFatal {
				// Return proper exit code for failed tasks
				switch err := err.(type) {
				case *ssh.ExitError:
					return data, &core.ExecError{Err: err, ExitCode: err.ExitStatus()}
				case *exec.ExitError:
					return data, &core.ExecError{Err: err, ExitCode: err.ExitCode()}
				default:
					return data, err
				}
			}
		}

	}
	wg.Wait()

	return data, nil
}

func (run *Run) TableWork(rIndex int, dryRun bool, data dao.TableOutput, dataExit dao.TableOutput, dataMutex *sync.RWMutex) error {
	config := run.Config
	task := run.Task
	server := run.Servers[rIndex]
	var wg sync.WaitGroup

	register := make(map[string]string)
	var registers []string
	for j, cmd := range task.Tasks {
		combinedEnvs := dao.MergeEnvs(cmd.Envs, server.Envs, registers)
		var client Client
		if cmd.Local || server.Local {
			client = run.LocalClients[server.Name]
		} else {
			client = run.RemoteClients[server.Name]
		}

		shell := dao.SelectFirstNonEmpty(cmd.Shell, server.Shell, config.Shell)
		shell = core.FormatShell(shell)
		workDir := getWorkDir(cmd, server)
		tableCmd := TaskContext{
			rIndex:  rIndex,
			cIndex:  j + 1,
			client:  client,
			dryRun:  dryRun,
			env:     combinedEnvs,
			workDir: workDir,
			shell:   shell,
			cmd:     cmd.Cmd,
			tty:     cmd.TTY,
		}

		err := RunTableCmd(tableCmd, data, dataExit, dataMutex, &wg)

		if task.Tasks[j].Register != "" {
			register[task.Tasks[j].Register] = data.Rows[rIndex].Columns[j+1]
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

		if !task.Spec.IgnoreErrors && err != nil {
			return err
		}
	}

	wg.Wait()

	return nil
}

func RunTableCmd(t TaskContext, data dao.TableOutput, dataExit dao.TableOutput, dataMutex *sync.RWMutex, wg *sync.WaitGroup) error {
	if t.dryRun {
		data.Rows[t.rIndex].Columns[t.cIndex] = t.cmd
		return nil
	}

	if t.tty {
		return ExecTTY(t.cmd, t.env)
	}

	err := t.client.Run(t.env, t.workDir, t.shell, t.cmd)
	if err != nil {
		return err
	}

	// Copy over commands STDOUT.
	var stdoutHandler = func(client Client) {
		defer wg.Done()
		dataMutex.Lock()
		out, err := io.ReadAll(client.Stdout())

		data.Rows[t.rIndex].Columns[t.cIndex] = fmt.Sprintf("%s%s", data.Rows[t.rIndex].Columns[t.cIndex], strings.TrimSuffix(string(out), "\n"))
		dataMutex.Unlock()

		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
	}
	wg.Add(1)
	go stdoutHandler(t.client)

	// Copy over tasks's STDERR.
	var stderrHandler = func(client Client) {
		defer wg.Done()
		dataMutex.Lock()
		out, err := io.ReadAll(client.Stderr())
		data.Rows[t.rIndex].Columns[t.cIndex] = fmt.Sprintf("%s%s", data.Rows[t.rIndex].Columns[t.cIndex], strings.TrimSuffix(string(out), "\n"))
		dataMutex.Unlock()
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
	}
	wg.Add(1)
	go stderrHandler(t.client)

	wg.Wait()

	if err := t.client.Wait(); err != nil {
		// Attach the error at the end of stdout
		data.Rows[t.rIndex].Columns[t.cIndex] = fmt.Sprintf("%s\n%s", data.Rows[t.rIndex].Columns[t.cIndex], err.Error())

		// Add exit code to dataExit
		var errCode int
		switch err := err.(type) {
		case *ssh.ExitError:
			errCode = err.ExitStatus()
		case *exec.ExitError:
			errCode = err.ExitCode()
		}

		dataExit.Rows[t.rIndex].Columns[t.cIndex] = fmt.Sprintf("%v", errCode)

		return err
	} else {
		dataExit.Rows[t.rIndex].Columns[t.cIndex] = fmt.Sprintf("%v", 0)
	}

	return nil
}
