//go:build !windows

package state

import (
	"os/exec"
	"syscall"
)

// group is a started command's process group.
type group struct{ pgid int }

// startGroup starts c as the leader of a new process group.
func startGroup(c *exec.Cmd) (*group, error) {
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.Setpgid = true
	if err := c.Start(); err != nil {
		return nil, err
	}
	return &group{pgid: c.Process.Pid}, nil
}

// kill kills every process left in the group.
func (g *group) kill() { _ = syscall.Kill(-g.pgid, syscall.SIGKILL) }

func (g *group) close() {}
