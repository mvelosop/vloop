package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	args := os.Args[1:]
	if len(args) >= 2 && args[0] == "list" && args[1] == "-m" {
		if os.Getenv("FAKE_GO_LIST_FAIL") != "" {
			fmt.Fprintln(os.Stderr, "first line")
			fmt.Fprintln(os.Stderr, os.Getenv("FAKE_GO_LIST_FAIL"))
			os.Exit(1)
		}
		if len(args) > 2 && args[2] == "-json" {
			fmt.Printf("{\"Path\":\"github.com/mvelosop/vloop\",\"Version\":%q}\n", os.Getenv("FAKE_GO_LATEST"))
		} else {
			fmt.Printf("github.com/mvelosop/vloop %s\n", os.Getenv("FAKE_GO_LATEST"))
		}
		return
	}
	if len(args) >= 2 && args[0] == "env" {
		if os.Getenv("FAKE_GO_ENV_FAIL") != "" {
			fmt.Fprintln(os.Stderr, "env detail")
			fmt.Fprintln(os.Stderr, os.Getenv("FAKE_GO_ENV_FAIL"))
			os.Exit(1)
		}
		switch args[1] {
		case "GOBIN":
			fmt.Println(os.Getenv("FAKE_GOBIN"))
		case "GOPATH":
			fmt.Println(os.Getenv("FAKE_GOPATH"))
		default:
			os.Exit(1)
		}
		return
	}
	if len(args) >= 2 && args[0] == "install" {
		if log := os.Getenv("FAKE_GO_LOG"); log != "" {
			_ = os.WriteFile(log, []byte(strings.Join(args, " ")+"\n"), 0o666)
		}
		if os.Getenv("FAKE_GO_INSTALL_FAIL") != "" {
			fmt.Fprintln(os.Stderr, "install detail")
			fmt.Fprintln(os.Stderr, os.Getenv("FAKE_GO_INSTALL_FAIL"))
			os.Exit(1)
		}
		name := "vloop"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		dir := os.Getenv("FAKE_GOBIN")
		if dir == "" {
			dir = os.Getenv("GOBIN")
		}
		exe := filepath.Join(dir, name)
		_ = os.MkdirAll(filepath.Dir(exe), 0o755)
		if src := os.Getenv("FAKE_GO_INSTALL_EXE"); src != "" {
			in, err := os.Open(src)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			out, err := os.OpenFile(exe, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
			if err == nil {
				_, err = io.Copy(out, in)
				_ = out.Close()
			}
			_ = in.Close()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		} else if err := os.WriteFile(exe, []byte("installed by fake go\n"), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	fmt.Fprintln(os.Stderr, "unexpected fake go arguments:", strings.Join(args, " "))
	os.Exit(1)
}
