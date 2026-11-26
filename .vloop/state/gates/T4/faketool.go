package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	tool := strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]))
	args := os.Args[1:]
	joined := strings.Join(args, " ")
	switch tool {
	case "git":
		switch {
		case strings.Contains(joined, "log") && strings.Contains(joined, "--format=%H") && strings.Contains(joined, "%(trailers:"):
			fmt.Print("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\x1fRelease v0.15.0\x1f\x1e")
		case strings.Contains(joined, "log") && strings.Contains(joined, "--format=%H"):
			fmt.Print("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\x1fRelease v0.15.0\x1e")
		case strings.Contains(joined, "log") && strings.Contains(joined, "--reverse"):
			// No brief trailers are needed for this release proof.
		case len(args) >= 3 && contains(args, "show"):
			fmt.Print("{\"version\":\"0.15.0\"}")
		case contains(args, "ls-remote"):
			// No earlier tags; the release uses its synthetic root.
		case contains(args, "rev-list"):
			fmt.Println("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
		case contains(args, "rev-parse"):
			// The tag is not already local.
		default:
			// Fetch, tag and push are successful no-output operations.
		}
	case "gh":
		if contains(args, "view") {
			os.Exit(1)
		}
		fmt.Println("https://example.invalid/releases/v0.15.0")
	default:
		fmt.Fprintln(os.Stderr, "unexpected fake tool")
		os.Exit(1)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
