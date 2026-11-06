package main

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/mod/module"
)

// rejectedPaths lists the files of the repository that dir belongs to, tracked
// or untracked and not ignored, and returns one line per path that
// module.CheckFilePath rejects, in the order git lists them.
func rejectedPaths(dir string) ([]string, error) {
	top, err := git(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, err
	}
	top = strings.TrimRight(top, "\r\n")
	raw, err := git(top, "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, p := range strings.Split(raw, "\x00") {
		if p == "" {
			continue
		}
		if err := module.CheckFilePath(p); err != nil {
			lines = append(lines, p+": "+err.Error())
		}
	}
	return lines, nil
}

// paths implements `release paths`.
func paths(dir string, args []string, stdout, stderr io.Writer) int {
	if len(args) != 0 {
		fmt.Fprintln(stderr, "usage: release paths")
		return 2
	}
	lines, err := rejectedPaths(dir)
	if err != nil {
		fmt.Fprintln(stderr, "release: "+oneLine(err.Error()))
		return 1
	}
	if len(lines) == 0 {
		return 0
	}
	for _, l := range lines {
		fmt.Fprintln(stdout, l)
	}
	fmt.Fprintf(stderr, "release: the module zip rejects %d path(s)\n", len(lines))
	return 1
}
