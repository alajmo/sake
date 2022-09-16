//go:build windows
// +build windows

package run

func SSHToServer(server string, disableVerifyHost bool, knownHostFile string) error {
	return nil
}

func ExecTTY(cmd string, envs []string) error {
	return nil
}
