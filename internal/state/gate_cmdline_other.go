//go:build !windows

package state

import "os/exec"

// setCmdLine is a no-op off Windows: cmd only runs there.
func setCmdLine(*exec.Cmd, string) {}
