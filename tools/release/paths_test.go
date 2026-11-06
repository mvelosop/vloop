package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPaths(t *testing.T) {
	const bad = `: malformed file path "%s": invalid char '\''`
	line := func(p string) string { return p + strings.Replace(bad, "%s", p, 1) }

	t.Run("committed it's repo/.keep", func(t *testing.T) {
		r := newRepo(t)
		r.write("go.mod", "module m\n")
		r.write("it's repo/.keep", "")
		r.commit("keep", "")
		got, err := rejectedPaths(r.dir)
		if err != nil || len(got) != 1 || got[0] != line("it's repo/.keep") {
			t.Fatalf("got %q, %v", got, err)
		}
	})
	t.Run("untracked a'b.txt", func(t *testing.T) {
		r := newRepo(t)
		r.write("go.mod", "module m\n")
		r.commit("base", "")
		r.write("a'b.txt", "")
		got, err := rejectedPaths(r.dir)
		if err != nil || len(got) != 1 || got[0] != line("a'b.txt") {
			t.Fatalf("got %q, %v", got, err)
		}
	})
	t.Run("ignored a'b.txt", func(t *testing.T) {
		r := newRepo(t)
		r.write("go.mod", "module m\n")
		r.write(".gitignore", "a'b.txt\n")
		r.commit("base", "")
		r.write("a'b.txt", "")
		got, err := rejectedPaths(r.dir)
		if err != nil || len(got) != 0 {
			t.Fatalf("got %q, %v", got, err)
		}
	})
	t.Run("only go.mod and x.go", func(t *testing.T) {
		r := newRepo(t)
		r.write("go.mod", "module m\n")
		r.write("x.go", "package m\n")
		r.commit("base", "")
		got, err := rejectedPaths(r.dir)
		if err != nil || len(got) != 0 {
			t.Fatalf("got %q, %v", got, err)
		}
	})
}

func TestRepositoryPaths(t *testing.T) {
	top, err := git(".", "rev-parse", "--show-toplevel")
	if err != nil {
		t.Skip("not in a git repository: " + err.Error())
	}
	top = strings.TrimSpace(top)
	mod, err := os.ReadFile(filepath.Join(top, "go.mod"))
	if err != nil || !bytes.Contains(mod, []byte("module github.com/mvelosop/vloop\n")) {
		t.Skip("the top directory's go.mod does not declare module github.com/mvelosop/vloop")
	}
	lines, err := rejectedPaths(top)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) > 0 {
		t.Fatalf("the module zip rejects %d path(s):\n%s", len(lines), strings.Join(lines, "\n"))
	}
}
