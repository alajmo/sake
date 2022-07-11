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

// Client is a wrapper over the SSH connection/sessions.
type LocalhostClient struct {
	Name string
	User string
	Host string

	stdin   io.WriteCloser
	cmd     *exec.Cmd
	stdout  io.Reader
	stderr  io.Reader
	running bool
}

func (c *LocalhostClient) Connect(_ bool, _ string, mu *sync.Mutex) *ErrConnect {
	return nil
}

func (c *LocalhostClient) Run(env []string, workDir string, shell string, cmdStr string) error {
	var err error

	if c.running {
		return fmt.Errorf("Command already running")
	}

	userEnv := os.Environ()

	if shell == "" {
		shell = dao.DEFAULT_SHELL
	}

	args := strings.SplitN(shell, " ", 2)
	shellProgram := args[0]
	shellFlag := append(args[1:], cmdStr)

	cmd := exec.Command(shellProgram, shellFlag...)
	cmd.Env = append(userEnv, env...)
	cmd.Dir = workDir
	c.cmd = cmd

	c.stdout, err = cmd.StdoutPipe()
	if err != nil {
		return err
	}

	c.stderr, err = cmd.StderrPipe()
	if err != nil {
		return err
	}

	c.stdin, err = cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := c.cmd.Start(); err != nil {
		return err
	}

	c.running = true
	return nil
}

func (c *LocalhostClient) Wait() error {
	if !c.running {
		return fmt.Errorf("Trying to wait on stopped command")
	}
	err := c.cmd.Wait()
	c.running = false
	return err
}

func (c *LocalhostClient) Close() error {
	return nil
}

func (c *LocalhostClient) Stdin() io.WriteCloser {
	return c.stdin
}

func (c *LocalhostClient) Write(p []byte) (n int, err error) {
	return c.stdin.Write(p)
}

func (c *LocalhostClient) WriteClose() error {
	return c.stdin.Close()
}

func (c *LocalhostClient) Stderr() io.Reader {
	return c.stderr
}

func (c *LocalhostClient) Stdout() io.Reader {
	return c.stdout
}

func (c *LocalhostClient) Prefix() string {
	return c.Host
}

func (c *LocalhostClient) Signal(sig os.Signal) error {
	return c.cmd.Process.Signal(sig)
}

func (c *LocalhostClient) GetName() string {
	return c.Name
}
