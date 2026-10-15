// Command vloop is the vloop CLI.
package main

import (
	"fmt"
	"os"

	"github.com/mvelosop/vloop/internal/cli"
)

//go:generate go run ../gendocs ../../docs/guide/commands.md

// Stamped at build time with -ldflags "-X main.commit=…". The version is the
// embedded plugin.json's, never stamped.
var commit = "unknown"

func main() {
	version, err := cli.PluginVersion()
	if err != nil {
		fmt.Fprintln(os.Stderr, "vloop:", err)
		os.Exit(1)
	}
	os.Exit(cli.Execute(cli.Resolve(version, commit), os.Args[1:], os.Stdout, os.Stderr))
}
