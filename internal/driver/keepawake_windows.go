//go:build windows

package driver

import (
	"runtime"
	"syscall"
)

const (
	esContinuous     = 0x80000000
	esSystemRequired = 0x00000001
)

// osHold sets ES_CONTINUOUS|ES_SYSTEM_REQUIRED on a thread kept for the life
// of the process; Windows drops the state when the thread, so the process, ends.
func osHold() error {
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("SetThreadExecutionState")
	res := make(chan error, 1)
	go func() {
		runtime.LockOSThread()
		if r, _, err := proc.Call(uintptr(esContinuous | esSystemRequired)); r == 0 {
			res <- err
			return
		}
		res <- nil
		select {} // keep the thread, and with it the state
	}()
	return <-res
}
