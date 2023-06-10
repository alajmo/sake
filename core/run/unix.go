//go:build !windows
// +build !windows

package run

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/alajmo/sake/core/dao"
	"golang.org/x/sys/unix"
)

// func SSHToServer(host string, user string, port uint16, bastion string, disableVerifyHost bool, knownHostFile string) error {
func SSHToServer(server dao.Server, disableVerifyHost bool, knownHostFile string) error {
	sshBin, err := exec.LookPath("ssh")
	if err != nil {
		return err
	}

	sshConnStr := fmt.Sprintf("%s@%s", server.User, server.Host)
	portStr := fmt.Sprintf("-p %d", server.Port)

	args := []string{"ssh", "-t", sshConnStr, portStr}
	if disableVerifyHost {
		args = append(args, "-o StrictHostKeyChecking=no")
	} else {
		args = append(args, fmt.Sprintf("-o UserKnownHostsFile=%s", knownHostFile))
	}

	// TODO:
	if len(server.Bastions) > 0 {
		jumphosts := []string{}
		for _, bastion := range server.Bastions {
			jumphosts = append(jumphosts, fmt.Sprintf("%s@%s:%d", bastion.User, bastion.Host, bastion.Port))
		}

		args = append(args, fmt.Sprintf("-J %s", strings.Join(jumphosts, ",")))
	}

	err = unix.Exec(sshBin, args, os.Environ())

	if err != nil {
		return err
	}

	return nil
}

func ExecTTY(cmd string, envs []string) error {
	shell := "bash"
	foundShell, found := os.LookupEnv("SHELL")
	if found {
		shell = foundShell
	}

	execBin, err := exec.LookPath(shell)
	if err != nil {
		return err
	}

	userEnv := append(os.Environ(), envs...)
	err = unix.Exec(execBin, []string{shell, "-c", cmd}, userEnv)
	if err != nil {
		return err
	}

	return nil
}
