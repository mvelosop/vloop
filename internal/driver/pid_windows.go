//go:build windows

package driver

import "os"

// On Windows FindProcess opens the process and fails when it is gone.
func processAlive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	p.Release()
	return true
}
