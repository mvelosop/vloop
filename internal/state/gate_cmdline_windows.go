//go:build windows

package state

import (
	"os/exec"
	"syscall"
)

// setCmdLine hands cmd.exe its command line as written; exec's own quoting
// would escape the quotes cmd does not unquote.
func setCmdLine(c *exec.Cmd, args string) {
	c.SysProcAttr = &syscall.SysProcAttr{CmdLine: syscall.EscapeArg(c.Path) + " " + args}
}
