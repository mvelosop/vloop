//go:build !windows

package state

import "os/exec"

// setCmdLine is a no-op off Windows: cmd only runs there.
func setCmdLine(*exec.Cmd, string) {}

// privateConsole is a no-op off Windows: there a code page belongs to the
// process, not to a shared console.
func privateConsole(*exec.Cmd) {}
