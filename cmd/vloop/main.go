// Command vloop is the vloop CLI.
package main

import (
	"os"

	"github.com/mvelosop/vloop/internal/cli"
)

//go:generate go run ../gendocs ../../docs/guide/commands.md

// Stamped at build time with -ldflags "-X main.version=… -X main.commit=…".
var (
	version = "1.0.0"
	commit  = "unknown"
)

func main() {
	os.Exit(cli.Execute(cli.Build{Version: version, Commit: commit}, os.Args[1:], os.Stdout, os.Stderr))
}
