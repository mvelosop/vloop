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

// createNoWindow is CREATE_NO_WINDOW: the process gets a console of its own,
// with no window, instead of sharing vloop's.
const createNoWindow = 0x08000000

// privateConsole starts c on a console of its own, so the code pages a pwsh
// gate sets for itself stay with it and never reach the operator's terminal.
func privateConsole(c *exec.Cmd) {
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.CreationFlags |= createNoWindow
}
