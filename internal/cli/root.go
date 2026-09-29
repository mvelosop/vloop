// Package cli is the vloop cobra command tree.
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// Exit codes shared by every command.
const (
	ExitOK       = 0
	ExitProblems = 1 // the command ran and found problems or failed
	ExitUsage    = 2 // unknown command, flag or config key, missing argument, invalid value
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

// ProblemError makes a command exit 1: it ran and found problems. Any other
// error returned from a command is a usage error and exits 2.
type ProblemError struct{ Err error }

func (e *ProblemError) Error() string { return e.Err.Error() }
func (e *ProblemError) Unwrap() error { return e.Err }

// Problem wraps err so the command exits 1.
func Problem(err error) error { return &ProblemError{Err: err} }

// Execute runs the command tree and returns the process exit code. Errors are
// printed to stderr as one line starting "vloop: "; nothing goes to stdout.
func Execute(b Build, args []string, stdout, stderr io.Writer) int {
	root, _ := NewRoot(b)
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
	var pe *ProblemError
	if errors.As(err, &pe) {
		return ExitProblems
	}
	return ExitUsage
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
	return root, g
}
