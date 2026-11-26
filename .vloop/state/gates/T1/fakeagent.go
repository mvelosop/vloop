package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	agent := strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]))
	args := os.Args[1:]
	if log := os.Getenv("FAKE_AGENT_LOG"); log != "" {
		f, _ := os.OpenFile(log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o666)
		if f != nil {
			fmt.Fprintln(f, agent+" "+strings.Join(args, " "))
			_ = f.Close()
		}
	}
	key := strings.ToUpper(agent)
	if len(args) >= 3 && args[0] == "plugin" && args[1] == "list" && args[2] == "--json" {
		if fail := os.Getenv("FAKE_" + key + "_LIST_FAIL"); fail != "" {
			fmt.Fprintln(os.Stderr, fail)
			os.Exit(1)
		}
		raw := os.Getenv("FAKE_" + key + "_LIST")
		if raw == "" {
			raw = "[]"
		}
		fmt.Println(raw)
		return
	}
	if len(args) >= 3 && args[0] == "plugin" && args[1] == "marketplace" && args[2] == "update" {
		if fail := os.Getenv("FAKE_" + key + "_MARKETPLACE_FAIL"); fail != "" {
			fmt.Fprintln(os.Stderr, "marketplace detail")
			fmt.Fprintln(os.Stderr, fail)
			os.Exit(1)
		}
		return
	}
	if len(args) >= 2 && args[0] == "plugin" && args[1] == "update" {
		if fail := os.Getenv("FAKE_" + key + "_UPDATE_FAIL"); fail != "" {
			fmt.Fprintln(os.Stderr, "update detail")
			fmt.Fprintln(os.Stderr, fail)
			os.Exit(1)
		}
		return
	}
	fmt.Fprintln(os.Stderr, "unexpected fake "+agent+" arguments:", strings.Join(args, " "))
	os.Exit(1)
}
