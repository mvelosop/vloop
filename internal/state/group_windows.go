//go:build windows

package state

import (
	"os/exec"
	"unsafe"

	"golang.org/x/sys/windows"
)

// group is a Job object holding a started command and what it starts.
type group struct{ job windows.Handle }

// startGroup starts c and puts it in a new Job object that kills its
// processes when it closes.
func startGroup(c *exec.Cmd) (*group, error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, err
	}
	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
		},
	}
	if _, err := windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info))); err != nil {
		windows.CloseHandle(job)
		return nil, err
	}
	if err := c.Start(); err != nil {
		windows.CloseHandle(job)
		return nil, err
	}
	h, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(c.Process.Pid))
	if err == nil {
		err = windows.AssignProcessToJobObject(job, h)
		windows.CloseHandle(h)
	}
	if err != nil {
		_ = c.Process.Kill()
		_ = c.Wait()
		windows.CloseHandle(job)
		return nil, err
	}
	return &group{job: job}, nil
}

// kill terminates every process in the job.
func (g *group) kill() { _ = windows.TerminateJobObject(g.job, 1) }

func (g *group) close() { windows.CloseHandle(g.job) }
