// Command kit is the Copilot acceptance kit's program, run as
// `go run ./tools/copilot-acceptance/kit <command>` from the repository root.
// It is a repository-owned program, not a vloop subcommand.
package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes one invocation and returns the exit code: 0 done, 1 refused
// or failed with a one-line message on stderr.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "kit: no command (want report, probe, summary, hook or calls)")
		return 1
	}
	switch args[0] {
	case "report":
		if len(args) < 2 {
			fmt.Fprintln(stderr, "kit report: no kind (want run, second, work, verify, gate-review or plan)")
			return 1
		}
		if !reportKinds[args[1]] {
			fmt.Fprintf(stderr, "kit report: unknown kind %q (want run, second, work, verify, gate-review or plan)\n", oneLine(args[1]))
			return 1
		}
		if len(args) < 5 {
			fmt.Fprintln(stderr, "kit report: want <kind> <repo> <vloop> <exit code>")
			return 1
		}
		out, err := report(args[1], args[2], args[3], args[4])
		if err != nil {
			fmt.Fprintln(stderr, "kit report: "+oneLine(err.Error()))
			return 1
		}
		io.WriteString(stdout, out)
		return 0
	case "probe":
		if len(args) < 3 {
			fmt.Fprintln(stderr, "kit probe: want <dir> <copilot>")
			return 1
		}
		return finish("kit probe", probe(args[1], args[2]), stderr)
	case "summary":
		if len(args) < 2 {
			fmt.Fprintln(stderr, "kit summary: want <dir>")
			return 1
		}
		out, err := summaryProbe(args[1])
		return emit("kit summary", out, err, stdout, stderr)
	case "calls":
		if len(args) < 2 {
			fmt.Fprintln(stderr, "kit calls: want <out.jsonl>")
			return 1
		}
		return finish("kit calls", calls(args[1], stdout), stderr)
	case "hook":
		return hook(args[1:], stdout, stderr)
	}
	fmt.Fprintf(stderr, "kit: unknown command %q (want report, probe, summary, hook or calls)\n", oneLine(args[0]))
	return 1
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func sortStrings(s []string) { sort.Strings(s) }

// finish returns the exit code of a command that printed nothing of its own:
// 1 with the error as one line on stderr, else 0.
func finish(name string, err error, stderr io.Writer) int {
	if err != nil {
		if err == errNoReview {
			fmt.Fprintln(stderr, err)
		} else {
			fmt.Fprintln(stderr, name+": "+oneLine(err.Error()))
		}
		return 1
	}
	return 0
}

func emit(name, out string, err error, stdout, stderr io.Writer) int {
	if err := finish(name, err, stderr); err != 0 {
		return err
	}
	io.WriteString(stdout, out)
	return 0
}

func hook(args []string, stdout, stderr io.Writer) int {
	const want = "want forms, failure or summary"
	if len(args) == 0 {
		fmt.Fprintln(stderr, "kit hook: no subcommand ("+want+")")
		return 1
	}
	switch args[0] {
	case "forms":
		if len(args) < 3 {
			fmt.Fprintln(stderr, "kit hook forms: want <dir> <copilot>")
			return 1
		}
		return finish("kit hook forms", hookFormsRun(args[1], args[2]), stderr)
	case "failure":
		if len(args) < 4 {
			fmt.Fprintln(stderr, "kit hook failure: want <dir> <copilot> <mode>")
			return 1
		}
		return finish("kit hook failure", hookFailureRun(args[1], args[2], args[3]), stderr)
	case "summary":
		if len(args) < 2 {
			fmt.Fprintln(stderr, "kit hook summary: want <dir>")
			return 1
		}
		out, err := hookSummary(args[1])
		return emit("kit hook summary", out, err, stdout, stderr)
	}
	fmt.Fprintf(stderr, "kit hook: unknown subcommand %q (%s)\n", oneLine(args[0]), want)
	return 1
}
