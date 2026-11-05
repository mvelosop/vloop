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
		fmt.Fprintln(stderr, "kit: no command (want report)")
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
	}
	fmt.Fprintf(stderr, "kit: unknown command %q (want report)\n", oneLine(args[0]))
	return 1
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func sortStrings(s []string) { sort.Strings(s) }
