package main

import (
	"os"

	"github.com/mvelosop/vloop/internal/cli"
)

func main() {
	version := os.Getenv("VLOOP_GATE_VERSION")
	module := os.Getenv("VLOOP_GATE_MODULE")
	os.Exit(cli.Execute(cli.Build{Version: version, Commit: "gate", Module: module}, os.Args[1:], os.Stdout, os.Stderr))
}
