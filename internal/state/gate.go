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

// RunGate runs verify from root in shell, streaming its output to stdout and
// stderr. It returns the command's own exit code (1 when it was killed by a
// signal) and how long it took. An unknown shell or one that is not on PATH is
// an error and nothing runs.
func RunGate(root, shell, verify string, stdout, stderr io.Writer) (int, time.Duration, error) {
	c, err := GateCommand(root, shell, verify, nil)
	if err != nil {
		return 0, 0, err
	}
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
