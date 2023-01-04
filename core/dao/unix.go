//go:build !windows
// +build !windows

package dao

import (
	"golang.org/x/sys/unix"
	"os/exec"
)

func ExecEditor(editor string, args []string, env []string) error {
	editorBin, err := exec.LookPath(editor)
	if err != nil {
		return err
	}

	err = unix.Exec(editorBin, args, env)
	if err != nil {
		return err
	}

	return nil
}
