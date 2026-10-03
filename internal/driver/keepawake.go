package driver

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

// AwakeWarning is the one run.log line for a hold that could not be taken.
func AwakeWarning(reason string) string {
	return "vloop: could not keep the machine awake (" + reason + ") — a sleeping machine stalls the run"
}

// KeepAwakeCommand is the command that holds the machine awake while process
// pid lives, for the given OS: nil for Windows, which holds with an API call
// instead. have reports whether an executable is on the path. It is an error
// when the OS has no way to hold.
func KeepAwakeCommand(goos string, pid int, have func(string) bool) ([]string, error) {
	p := strconv.Itoa(pid)
	switch goos {
	case "darwin":
		return []string{"caffeinate", "-i", "-w", p}, nil
	case "linux":
		if !have("systemd-inhibit") {
			return nil, errors.New("systemd-inhibit not found")
		}
		return []string{"systemd-inhibit", "--what=idle", "--why=vloop run", "--who=vloop",
			"tail", "--pid=" + p, "-f", "/dev/null"}, nil
	case "windows":
		return nil, nil
	}
	return nil, fmt.Errorf("no way to hold the machine awake on %s", goos)
}

// KeepAwake holds the machine awake until this process exits, however it
// exits: the helper watches vloop's pid, and the Windows flag dies with the
// process. A non-nil error says why no hold was taken.
func KeepAwake() error {
	return keepAwake(true, runtime.GOOS, os.Getpid(), haveExe, startHelper, osHold)
}

func haveExe(name string) bool { _, err := exec.LookPath(name); return err == nil }

func startHelper(argv []string) error {
	cmd := exec.Command(argv[0], argv[1:]...)
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait() // reap it when vloop's pid goes away
	return nil
}

// keepAwake is KeepAwake with its seams: with on false it attempts nothing.
func keepAwake(on bool, goos string, pid int, have func(string) bool, start func([]string) error, native func() error) error {
	if !on {
		return nil
	}
	argv, err := KeepAwakeCommand(goos, pid, have)
	if err != nil {
		return err
	}
	if argv == nil {
		return native()
	}
	return start(argv)
}
