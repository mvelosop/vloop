// Command gendocs writes docs/guide/commands.md from the vloop command tree.
// It is run by `go generate ./...` (see cmd/vloop/main.go).
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mvelosop/vloop/internal/cli"
)

func main() {
	if len(os.Args) != 2 || strings.HasPrefix(os.Args[1], "-") {
		fmt.Fprintln(os.Stderr, "usage: gendocs <output file>")
		os.Exit(2)
	}
	if err := os.WriteFile(os.Args[1], []byte(cli.CommandsReference()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "gendocs:", err)
		os.Exit(1)
	}
}
