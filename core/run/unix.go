// +build !windows

package run

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
)

func SSHToServer(host string, user string, port uint16, disableVerifyHost bool, knownHostFile string) error {
	sshBin, err := exec.LookPath("ssh")
	if err != nil {
		return err
	}

	sshConnStr := fmt.Sprintf("%s@%s", user, host)
	portStr := fmt.Sprintf("-p %d", port)

	if !disableVerifyHost {
		knownHostFileStr := fmt.Sprintf("-o UserKnownHostsFile=%s", knownHostFile)
		err = unix.Exec(sshBin, []string{"ssh", "-t", sshConnStr, portStr, knownHostFileStr}, os.Environ())
	} else {
		err = unix.Exec(sshBin, []string{"ssh", "-t", sshConnStr, portStr, "-o StrictHostKeyChecking=no"}, os.Environ())
	}

	if err != nil {
		return err
	}

	return nil
}

func ExecTTY(cmd string, envs []string) error {
	execBin, err := exec.LookPath("bash")
	if err != nil {
		return err
	}

	userEnv := append(os.Environ(), envs...)
	err = unix.Exec(execBin, []string{"bash", "-c", cmd}, userEnv)
	if err != nil {
		return err
	}

	return nil
}

