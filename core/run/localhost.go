package run

import (
	"fmt"
	"github.com/alajmo/sake/core/dao"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// Client is a wrapper over the SSH connection/Sessions.
type LocalhostClient struct {
	Name string
	User string
	Host string
	Port uint16

	Sessions []LocalSession
}

type LocalSession struct {
	stdin   io.WriteCloser
	cmd     *exec.Cmd
	stdout  io.Reader
	stderr  io.Reader
	running bool
}

func (c *LocalhostClient) Connect(dialer SSHDialFunc, _ bool, _ string, _ uint, mu *sync.Mutex) *ErrConnect {
	return nil
}

func (c *LocalhostClient) Run(i int, env []string, workDir string, shell string, cmdStr string) error {
	var err error

	if c.Sessions[i].running {
		return fmt.Errorf("command already running")
	}

	userEnv := os.Environ()

	if shell == "" {
		shell = dao.DEFAULT_SHELL
	}

	var cmdString string
	if workDir != "" {
		cmdString = fmt.Sprintf("cd %s; %s", workDir, cmdStr)
	} else {
		cmdString = cmdStr
	}

	args := strings.SplitN(shell, " ", 2)
	shellProgram := args[0]
	shellArgs := append(args[1:], cmdString)

	cmd := exec.Command(shellProgram, shellArgs...)
	cmd.Env = append(userEnv, env...)
	c.Sessions[i].cmd = cmd

	c.Sessions[i].stdout, err = cmd.StdoutPipe()
	if err != nil {
		return err
	}

	c.Sessions[i].stderr, err = cmd.StderrPipe()
	if err != nil {
		return err
	}

	c.Sessions[i].stdin, err = cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := c.Sessions[i].cmd.Start(); err != nil {
		return err
	}

	c.Sessions[i].running = true
	return nil
}

func (c *LocalhostClient) Wait(i int) error {
	if !c.Sessions[i].running {
		return fmt.Errorf("trying to wait on stopped command")
	}
	err := c.Sessions[i].cmd.Wait()
	c.Sessions[i].running = false
	return err
}

func (c *LocalhostClient) Close(i int) error {
	return nil
}

func (c *LocalhostClient) Stdin(i int) io.WriteCloser {
	return c.Sessions[i].stdin
}

func (c *LocalhostClient) Write(i int, p []byte) (n int, err error) {
	return c.Sessions[i].stdin.Write(p)
}

func (c *LocalhostClient) WriteClose(i int) error {
	return c.Sessions[i].stdin.Close()
}

func (c *LocalhostClient) Stderr(i int) io.Reader {
	return c.Sessions[i].stderr
}

func (c *LocalhostClient) Stdout(i int) io.Reader {
	return c.Sessions[i].stdout
}

func (c *LocalhostClient) Prefix() (string, string, string, uint16) {
	return c.Name, c.Host, c.User, c.Port
}

func (c *LocalhostClient) Signal(i int, sig os.Signal) error {
	return c.Sessions[i].cmd.Process.Signal(sig)
}

func (c *LocalhostClient) GetName() string {
	return c.Name
}

func (c *LocalhostClient) Connected() bool {
	return true
}
