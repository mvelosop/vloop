// Command agent-eval replays the plugin's eval cases on Claude or Copilot with
// any model, run as `go run ./tools/agent-eval` from the repository root.
// It is a repository-owned program, not a vloop subcommand.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

func main() {
	os.Exit(run(os.Args[1:], ".", os.Stdout, os.Stderr))
}

type multi []string

func (m *multi) String() string     { return strings.Join(*m, ",") }
func (m *multi) Set(v string) error { *m = append(*m, v); return nil }

// options are the flags, as given.
type options struct {
	agent, model, out, casesDir, runs, judgeModel, timeout, vloop string
	patterns                                                      multi
	list                                                          bool
}

// run executes one invocation from dir and returns the exit code: 0 done,
// 1 failed, 2 refused or usage.
func run(args []string, dir string, stdout, stderr io.Writer) int {
	var o options
	fs := flag.NewFlagSet("agent-eval", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&o.agent, "agent", "", "")
	fs.StringVar(&o.model, "model", "", "")
	fs.StringVar(&o.out, "out", "", "")
	fs.StringVar(&o.casesDir, "cases", "plugin/evals", "")
	fs.Var(&o.patterns, "case", "")
	fs.StringVar(&o.runs, "runs", "1", "")
	fs.StringVar(&o.judgeModel, "judge-model", "sonnet", "")
	fs.StringVar(&o.timeout, "timeout", "20m", "")
	fs.StringVar(&o.vloop, "vloop", "", "")
	fs.BoolVar(&o.list, "list", false, "")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(stderr, "agent-eval: "+oneLine(err.Error()))
		return 2
	}
	if fs.NArg() > 0 {
		fmt.Fprintln(stderr, "agent-eval: unexpected argument "+fs.Arg(0))
		return 2
	}
	casesDir := o.casesDir
	if !o.list {
		if msg := checkFlags(&o); msg != "" {
			fmt.Fprintln(stderr, "agent-eval: "+msg)
			return 2
		}
	}
	names, err := caseNames(casesDir)
	if err != nil {
		fmt.Fprintln(stderr, "agent-eval: "+oneLine(reason(err)))
		return 1
	}
	var selected []string
	for _, n := range names {
		if len(o.patterns) == 0 {
			selected = append(selected, n)
			continue
		}
		for _, p := range o.patterns {
			ok, err := path.Match(p, n)
			if err != nil {
				fmt.Fprintf(stderr, "agent-eval: --case %s is not a pattern\n", p)
				return 2
			}
			if ok {
				selected = append(selected, n)
				break
			}
		}
	}
	if len(selected) == 0 {
		fmt.Fprintln(stderr, "agent-eval: no case matches")
		return 2
	}
	if !o.list && o.agent == "copilot" && findVloop(o.vloop) == "" {
		fmt.Fprintln(stderr, "agent-eval: no vloop binary for the fence hook")
		return 2
	}
	var cases []*evalCase
	for _, n := range selected {
		c, err := readCase(casesDir, n)
		if err != nil {
			fmt.Fprintln(stderr, "agent-eval: "+oneLine(err.Error()))
			return 1
		}
		cases = append(cases, c)
	}
	if o.list {
		tw := tabwriter.NewWriter(stdout, 0, 8, 2, ' ', 0)
		fmt.Fprintln(tw, "CASE\tPHASE\tGRADERS")
		for _, c := range cases {
			fmt.Fprintf(tw, "%s\t%s\t%d\n", c.name, c.phase, len(c.graders))
		}
		tw.Flush()
		return 0
	}
	runs, _ := strconv.Atoi(o.runs)
	timeout, _ := time.ParseDuration(o.timeout)
	return execute(&o, runs, timeout, cases, stdout, stderr)
}

// checkFlags returns the first refusal among the flags, in the order the brief fixes.
func checkFlags(o *options) string {
	for _, f := range []struct{ name, v string }{{"agent", o.agent}, {"model", o.model}, {"out", o.out}} {
		if f.v == "" {
			return "--" + f.name + " is required"
		}
	}
	if o.agent != "claude" && o.agent != "copilot" {
		return "unknown agent " + o.agent
	}
	if n, err := strconv.Atoi(o.runs); err != nil || n < 1 {
		return "--runs must be 1 or more"
	}
	if _, err := time.ParseDuration(o.timeout); err != nil {
		return "--timeout " + o.timeout + " is not a duration"
	}
	if entries, err := os.ReadDir(o.out); err == nil && len(entries) > 0 {
		return o.out + " is not empty"
	} else if err != nil {
		if st, serr := os.Stat(o.out); serr == nil && !st.IsDir() {
			return o.out + " is not empty"
		}
	}
	return ""
}

// findVloop returns the vloop binary: the --vloop file, else the one on PATH.
func findVloop(flagValue string) string {
	if flagValue != "" {
		if st, err := os.Stat(flagValue); err == nil && !st.IsDir() {
			return flagValue
		}
		return ""
	}
	p, err := exec.LookPath("vloop")
	if err != nil {
		return ""
	}
	return p
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
