package config

import (
	"os"
	"path/filepath"
	"testing"
)

func mkdirs(t *testing.T, paths ...string) {
	t.Helper()
	for _, p := range paths {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRootNearestVloopWinsOverGit(t *testing.T) {
	base := t.TempDir()
	mkdirs(t, filepath.Join(base, ".git"), filepath.Join(base, "sub", ".vloop"), filepath.Join(base, "sub", "deep"))
	if got, want := Root(filepath.Join(base, "sub", "deep")), filepath.Join(base, "sub"); got != want {
		t.Fatalf("Root = %q, want %q", got, want)
	}
}

func TestRootNearestGit(t *testing.T) {
	base := t.TempDir()
	mkdirs(t, filepath.Join(base, ".git"), filepath.Join(base, "a", "b"))
	if got := Root(filepath.Join(base, "a", "b")); got != base {
		t.Fatalf("Root = %q, want %q", got, base)
	}
}

func TestRootStartDirWithoutMarkers(t *testing.T) {
	base := t.TempDir()
	start := filepath.Join(base, "x")
	mkdirs(t, start)
	if got := Root(start); got != start {
		t.Fatalf("Root = %q, want %q", got, start)
	}
}
