// Package cli is the vloop cobra command tree.
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/mvelosop/vloop/internal/config"
	"github.com/spf13/cobra"
)

// Exit codes shared by every command.
const (
	ExitOK       = 0
	ExitProblems = 1 // the command ran and found problems or failed
	ExitUsage    = 2 // unknown command or flag, wrong argument count, invalid value for a flag or a settable field
)

// Build is the build-time identity stamped into the binary.
type Build struct {
	Version string
	Commit  string
}

// Globals holds the persistent flags every command inherits.
type Globals struct {
	Dir     string
	JSON    bool
	NoColor bool
	Quiet   bool
	Verbose bool

	ran bool // a command's RunE started: anything cobra rejected before that is usage
}

// Color reports whether output to w may be coloured: only when w is a
// terminal and neither --no-color nor a non-empty NO_COLOR disables it.
func (g *Globals) Color(w io.Writer) bool {
	if g.NoColor || os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	st, err := f.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}

// ProblemError marks an error as a problem the command found. Every error
// returned from a command exits 1 unless it is a usage error (UsageError, a
// config.UsageError, InvalidValueError or UnknownKeyError, or one cobra rejected
// before the command ran); Problem is kept for readability where the failure is
// the point.
type ProblemError struct{ Err error }

func (e *ProblemError) Error() string { return e.Err.Error() }
func (e *ProblemError) Unwrap() error { return e.Err }

// Problem wraps err so the command exits 1.
func Problem(err error) error { return &ProblemError{Err: err} }

// UsageError makes a command exit 2: the invocation, not the repository, is
// wrong — an invalid value for a flag or a settable field.
type UsageError struct{ Err error }

func (e *UsageError) Error() string { return e.Err.Error() }
func (e *UsageError) Unwrap() error { return e.Err }

// Usage wraps err so the command exits 2.
func Usage(err error) error { return &UsageError{Err: err} }

// ExitError makes a command exit with a specific code, as `vloop run` does for
// its own endings (R-3).
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string { return e.Err.Error() }
func (e *ExitError) Unwrap() error { return e.Err }

// Execute runs the command tree and returns the process exit code. Errors are
// printed to stderr as one line starting "vloop: "; nothing goes to stdout.
func Execute(b Build, args []string, stdout, stderr io.Writer) int {
	root, g := NewRoot(b)
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)
	err := root.Execute()
	if err == nil {
		return ExitOK
	}
	if err.Error() != "" { // an empty message exits non-zero having already reported
		fmt.Fprintf(stderr, "vloop: %s\n", oneLine(err.Error()))
	}
	var xe *ExitError
	if errors.As(err, &xe) {
		return xe.Code
	}
	var pe *ProblemError
	if errors.As(err, &pe) {
		return ExitProblems
	}
	var ue *UsageError
	var cu *config.UsageError
	var inv *config.InvalidValueError
	var unk *config.UnknownKeyError
	if errors.As(err, &ue) || errors.As(err, &cu) || errors.As(err, &inv) || errors.As(err, &unk) || !g.ran {
		return ExitUsage
	}
	return ExitProblems
}

// markRan wraps every RunE so Execute can tell cobra's own rejections (unknown
// command, flag-parse and argument-count errors), which happen before any RunE
// starts, from a command that ran and failed.
func markRan(c *cobra.Command, g *Globals) {
	if run := c.RunE; run != nil {
		c.RunE = func(cmd *cobra.Command, args []string) error {
			g.ran = true
			return run(cmd, args)
		}
	}
	for _, sub := range c.Commands() {
		markRan(sub, g)
	}
}

func oneLine(s string) string {
	for i, r := range s {
		if r == '\n' {
			return s[:i]
		}
	}
	return s
}

// NewRoot builds the command tree.
func NewRoot(b Build) (*cobra.Command, *Globals) {
	g := &Globals{}
	root := &cobra.Command{
		Use:           "vloop",
		Short:         "Plan, run and review autonomous Claude loops",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	pf := root.PersistentFlags()
	pf.StringVarP(&g.Dir, "dir", "C", "", "act as if started in `path`")
	pf.BoolVar(&g.JSON, "json", false, "machine-readable output on stdout")
	pf.BoolVar(&g.NoColor, "no-color", false, "disable colour (also: non-empty NO_COLOR)")
	pf.BoolVarP(&g.Quiet, "quiet", "q", false, "print less")
	pf.BoolVarP(&g.Verbose, "verbose", "v", false, "print more")
	root.AddCommand(newVersion(b, g))
	root.AddCommand(newConfig(g))
	root.AddCommand(newBrief(g))
	root.AddCommand(newSchema(g))
	root.AddCommand(newStatus(g))
	root.AddCommand(newTask(g))
	root.AddCommand(newMetrics(g))
	root.AddCommand(newDefect(g))
	root.AddCommand(newIntervention(g))
	root.AddCommand(newPlugin(b, g))
	root.AddCommand(newInit(b, g))
	root.AddCommand(newUpgrade(b, g))
	root.AddCommand(newDoctor(b, g))
	root.AddCommand(newRun(b, g))
	markRan(root, g)
	return root, g
}
