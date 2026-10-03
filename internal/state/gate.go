package state

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"
)

// ShellMissingError is returned when the plan's shell is not on PATH.
type ShellMissingError struct{ Shell string }

func (e *ShellMissingError) Error() string {
	return fmt.Sprintf("this plan's gates are %s commands and %s is not on PATH", e.Shell, e.Shell)
}

// GateArgs returns the arguments that run cmd in shell: sh and bash take -c,
// pwsh and powershell take -NoProfile -NonInteractive -Command, cmd takes /C.
func GateArgs(shell, cmd string) []string {
	switch shell {
	case "pwsh", "powershell":
		return []string{"-NoProfile", "-NonInteractive", "-Command", cmd}
	case "cmd":
		return []string{"/C", cmd}
	default:
		return []string{"-c", cmd}
	}
}

// UnknownShellError is returned when the plan's shell is not one of the five
// gate shells.
type UnknownShellError struct{ Shell string }

func (e *UnknownShellError) Error() string {
	return fmt.Sprintf("shell %s is not one of sh, bash, pwsh, powershell, cmd", e.Shell)
}

// GateCmdLine is the command line cmd.exe receives for verify: /C and the
// command verbatim, so cmd sees the quotes the plan wrote.
func GateCmdLine(verify string) string {
	return "/C " + verify
}

// GateCommand builds the one command that runs verify from root in shell; the
// driver and vloop task gate both start gates from it. An empty shell is sh.
// A shell outside the five known ones is an *UnknownShellError and one that is
// not on PATH a *ShellMissingError; neither builds a command. A nil env keeps
// the process environment.
func GateCommand(root, shell, verify string, env []string) (*exec.Cmd, error) {
	if shell == "" {
		shell = "sh"
	}
	switch shell {
	case "sh", "bash", "pwsh", "powershell", "cmd":
	default:
		return nil, &UnknownShellError{Shell: shell}
	}
	path, err := exec.LookPath(shell)
	if err != nil {
		return nil, &ShellMissingError{Shell: shell}
	}
	c := exec.Command(path, GateArgs(shell, verify)...)
	if shell == "cmd" {
		setCmdLine(c, GateCmdLine(verify))
	}
	c.Dir = root
	c.Env = env
	return c, nil
}

// RunGate runs verify from root in shell with no deadline; see RunGateWithin.
func RunGate(root, shell, verify string, stdout, stderr io.Writer) (int, time.Duration, error) {
	code, d, _, err := RunGateWithin(root, shell, verify, stdout, stderr, 0)
	return code, d, err
}

// RunGateWithin runs verify from root in shell, in a process group of its own,
// streaming its output to stdout and stderr. A gate still running after
// timeout (none when it is zero) is killed with everything it started and
// timedOut is true; so is whatever a finished gate left running. It returns
// the command's own exit code (1 when it was killed by a signal) and how long
// it took. An unknown shell or one that is not on PATH is an error and nothing
// runs.
func RunGateWithin(root, shell, verify string, stdout, stderr io.Writer, timeout time.Duration) (code int, d time.Duration, timedOut bool, err error) {
	c, err := GateCommand(root, shell, verify, nil)
	if err != nil {
		return 0, 0, false, err
	}
	c.Stdout, c.Stderr = stdout, stderr
	start := time.Now()
	timedOut, err = RunGroup(c, timeout)
	d = time.Since(start)
	if err == nil {
		return 0, d, timedOut, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		if code := ee.ExitCode(); code > 0 {
			return code, d, timedOut, nil
		}
		return 1, d, timedOut, nil
	}
	return 0, d, timedOut, err
}

// GateTimedOutLine is the line that ends a gate's log when it timed out.
func GateTimedOutLine(id string, minutes int) string {
	return fmt.Sprintf("vloop: gate %s timed out after %d min", id, minutes)
}
