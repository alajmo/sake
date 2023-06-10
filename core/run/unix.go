//go:build !windows
// +build !windows

package run

import (
	"fmt"
	"github.com/alajmo/sake/core/dao"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
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
	// if server.BastionHost != "" {
	// 	jumphost := fmt.Sprintf("%s@%s:%d", server.BastionUser, server.BastionHost, server.BastionPort)
	// 	args = append(args, fmt.Sprintf("-J %s", jumphost))
	// }

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
