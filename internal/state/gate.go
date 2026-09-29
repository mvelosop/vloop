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

// RunGate runs verify from root in shell, streaming its output to stdout and
// stderr. It returns the command's own exit code (1 when it was killed by a
// signal) and how long it took. A shell that is not on PATH is a
// *ShellMissingError and nothing runs.
func RunGate(root, shell, verify string, stdout, stderr io.Writer) (int, time.Duration, error) {
	path, err := exec.LookPath(shell)
	if err != nil {
		return 0, 0, &ShellMissingError{Shell: shell}
	}
	c := exec.Command(path, GateArgs(shell, verify)...)
	c.Dir = root
	c.Stdout, c.Stderr = stdout, stderr
	start := time.Now()
	err = c.Run()
	d := time.Since(start)
	if err == nil {
		return 0, d, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		if code := ee.ExitCode(); code > 0 {
			return code, d, nil
		}
		return 1, d, nil
	}
	return 0, d, err
}

// ReplaceGate replaces a task's verify command, recording the old one, when
// and why in gate_history. Setting the command the task already has is an
// error.
func ReplaceGate(p *Plan, id, verify, reason string, at time.Time) error {
	t := p.Find(id)
	if t == nil {
		return &NoTaskError{id}
	}
	if verify == "" {
		return errors.New("verify command is empty")
	}
	if t.Verify == verify {
		return fmt.Errorf("%s already has that verify command", id)
	}
	t.GateHistory = append(t.GateHistory, GateReplace{
		Verify:     t.Verify,
		ReplacedAt: at.UTC().Format(time.RFC3339),
		Reason:     reason,
		By:         "operator",
	})
	t.Verify = verify
	return nil
}
