//go:build windows
// +build windows

package run

func SSHToServer(host string, user string, port uint16, disableVerifyHost bool, knownHostFile string) error {
	return nil
}

func ExecTTY(cmd string, envs []string) error {
	return nil
}
