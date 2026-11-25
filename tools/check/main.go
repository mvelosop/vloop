// Command check runs this repository's toolchain checks (CLAUDE.md →
// Toolchain) in any shell: the tests, vet, the gofmt check, go mod tidy -diff
// and the Linux and Windows builds. It stops at the first that fails and exits 1.
// .vloop/config.toml's [[check]] runs it as `go run ./tools/check`.
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type step struct {
	name  string
	args  []string
	env   []string
	quiet bool // fails when it prints anything
}

var steps = []step{
	{name: "test", args: []string{"go", "test", "-timeout", "30m", "./..."}},
	{name: "vet", args: []string{"go", "vet", "./..."}},
	{name: "gofmt", args: []string{"gofmt", "-l", "."}, quiet: true},
	{name: "tidy", args: []string{"go", "mod", "tidy", "-diff"}},
	{name: "linux build", args: []string{"go", "build", "./..."}, env: []string{"GOOS=linux"}},
	{name: "windows build", args: []string{"go", "build", "./..."}, env: []string{"GOOS=windows"}},
}

func main() {
	for _, s := range steps {
		if err := run(s); err != nil {
			fmt.Fprintf(os.Stderr, "check: %s failed: %v\n", s.name, err)
			os.Exit(1)
		}
	}
}

func run(s step) error {
	cmd := exec.Command(s.args[0], s.args[1:]...)
	cmd.Env = append(os.Environ(), s.env...)
	cmd.Stderr = os.Stderr
	if !s.quiet {
		cmd.Stdout = os.Stdout
		return cmd.Run()
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return err
	}
	if listed := strings.TrimSpace(out.String()); listed != "" {
		return fmt.Errorf("it lists\n%s", listed)
	}
	return nil
}
