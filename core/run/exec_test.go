package run

import (
	"testing"

	"github.com/alajmo/sake/core/test"
)

func TestWorkDir(t *testing.T) {
	// Remote

	test.CheckEqS(t, getWorkDir(false, false, "", "", "cmd-root", "server-root"), "")
	test.CheckEqS(t, getWorkDir(false, false, "cmd", "", "cmd-root", "server-root"), "cmd")
	test.CheckEqS(t, getWorkDir(false, false, "", "server", "", "server-root"), "server")
	test.CheckEqS(t, getWorkDir(false, false, "cmd", "server", "cmd-root", "server-root"), "server/cmd")

	// Local

	test.CheckEqS(t, getWorkDir(true, false, "", "", "cmd-root", "server-root"), "cmd-root")
	test.CheckEqS(t, getWorkDir(true, true, "", "", "cmd-root", "server-root"), "cmd-root")
	test.CheckEqS(t, getWorkDir(true, false, "", "server", "cmd-root", "server-root"), "cmd-root")
	test.CheckEqS(t, getWorkDir(false, true, "", "", "cmd-root", "server-root"), "cmd-root")

	test.CheckEqS(t, getWorkDir(false, true, "cmd", "server", "cmd-root", "server-root"), "server-root/server/cmd")
	test.CheckEqS(t, getWorkDir(true, true, "cmd", "server", "", "server-root"), "server-root/server/cmd")

	test.CheckEqS(t, getWorkDir(true, false, "cmd", "", "cmd-root", "server-root"), "cmd-root/cmd")
	test.CheckEqS(t, getWorkDir(true, false, "cmd", "server", "cmd-root", "server-root"), "cmd-root/cmd")
	test.CheckEqS(t, getWorkDir(true, true, "cmd", "", "cmd-root", "server-root"), "cmd-root/cmd")
	test.CheckEqS(t, getWorkDir(false, true, "cmd", "", "cmd-root", "server-root"), "cmd-root/cmd")

	test.CheckEqS(t, getWorkDir(false, true, "", "server", "cmd-root", "server-root"), "server-root/server")
	test.CheckEqS(t, getWorkDir(true, true, "", "server", "cmd-root", "server-root"), "server-root/server")
}
