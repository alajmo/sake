package run

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/alajmo/sake/core/dao"
)

func (run *Run) Table(dryRun bool) dao.TableOutput {
	task := run.Task
	servers := run.Servers

	var data dao.TableOutput
	var dataMutex = sync.RWMutex{}

	/**
	 ** Headers
	 **/
	data.Headers = append(data.Headers, "server")

	// Append Command names if set
	for _, subTask := range task.Tasks {
		if subTask.Name != "" {
			data.Headers = append(data.Headers, subTask.Name)
		} else {
			data.Headers = append(data.Headers, "output")
		}
	}

	// Populate the rows (server name is first cell, then commands and cmd output is set to empty string)
	for i, p := range servers {
		data.Rows = append(data.Rows, dao.Row{Columns: []string{p.Name}})

		for range task.Tasks {
			data.Rows[i].Columns = append(data.Rows[i].Columns, "")
		}
	}

	var wg sync.WaitGroup
	/**
	 ** Values
	 **/
	for i := range servers {
		wg.Add(1)
		if task.Spec.Parallel {
			go func(i int, wg *sync.WaitGroup) {
				defer wg.Done()
				// TODO
				_ = run.TableWork(i, dryRun, data, &dataMutex)
			}(i, &wg)
		} else {
			err := func(i int, wg *sync.WaitGroup) error {
				defer wg.Done()
				err := run.TableWork(i, dryRun, data, &dataMutex)

				return err
			}(i, &wg)

			if run.Task.Spec.AnyErrorsFatal && err != nil {
				break
			}
		}

	}
	wg.Wait()

	return data
}

func (run *Run) TableWork(rIndex int, dryRun bool, data dao.TableOutput, dataMutex *sync.RWMutex) error {
	task := run.Task
	server := run.Servers[rIndex]
	var wg sync.WaitGroup

	for j, cmd := range task.Tasks {
		combinedEnvs := dao.MergeEnvs(server.Envs, cmd.Envs)
		var client Client
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

		tableCmd := TaskContext{
			rIndex: rIndex,
			cIndex: j + 1,
			client: client,
			dryRun: dryRun,
			env:    combinedEnvs,
			cmd:    cmdString,
		}

		err := RunTableCmd(tableCmd, data, dataMutex, &wg)
		if err != nil && !task.Spec.IgnoreErrors {
			return err
		}
	}

	wg.Wait()

	return nil
}

func RunTableCmd(t TaskContext, data dao.TableOutput, dataMutex *sync.RWMutex, wg *sync.WaitGroup) error {
	combinedEnvs := dao.MergeEnvs([]string{}, t.env)
	// combinedEnvs := dao.MergeEnvs(t.client.GetEnv(), t.env)

	if t.dryRun {
		data.Rows[t.rIndex].Columns[t.cIndex] = t.cmd
		return nil
	}

	err := t.client.Run(combinedEnvs, t.cmd)
	if err != nil {
		return err
	}

	// Copy over commands STDOUT.
	var stdoutHandler = func(client Client) {
		defer wg.Done()
		dataMutex.Lock()
		out, err := ioutil.ReadAll(client.Stdout())

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
		out, err := ioutil.ReadAll(client.Stderr())
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
		data.Rows[t.rIndex].Columns[t.cIndex] = fmt.Sprintf("%s\n%s", data.Rows[t.rIndex].Columns[t.cIndex], err.Error())
		return err
	}

	return nil
}
