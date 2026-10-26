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

func TestRootNearerGitWinsOverOuterVloop(t *testing.T) {
	base := t.TempDir()
	mkdirs(t, filepath.Join(base, ".vloop"), filepath.Join(base, "sub", ".git"), filepath.Join(base, "sub", "deep"))
	if got, want := Root(filepath.Join(base, "sub", "deep")), filepath.Join(base, "sub"); got != want {
		t.Fatalf("Root = %q, want %q", got, want)
	}
}

func TestRootGitFileCounts(t *testing.T) {
	base := t.TempDir()
	mkdirs(t, filepath.Join(base, ".vloop"), filepath.Join(base, "sub", "deep"))
	if err := os.WriteFile(filepath.Join(base, "sub", ".git"), []byte("gitdir: ../x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, want := Root(filepath.Join(base, "sub", "deep")), filepath.Join(base, "sub"); got != want {
		t.Fatalf("Root = %q, want %q", got, want)
	}
}

func TestRootStartHoldingVloopUnderOuterGit(t *testing.T) {
	base := t.TempDir()
	mkdirs(t, filepath.Join(base, ".git"), filepath.Join(base, "a", ".vloop"))
	if got, want := Root(filepath.Join(base, "a")), filepath.Join(base, "a"); got != want {
		t.Fatalf("Root = %q, want %q", got, want)
	}
}

func TestRootBothMarkersInOneDirectory(t *testing.T) {
	base := t.TempDir()
	mkdirs(t, filepath.Join(base, ".git"), filepath.Join(base, ".vloop"), filepath.Join(base, "x"))
	if got := Root(filepath.Join(base, "x")); got != base {
		t.Fatalf("Root = %q, want %q", got, base)
	}
}

func TestRootNeitherMarkerUnderVloopRepo(t *testing.T) {
	base := t.TempDir()
	mkdirs(t, filepath.Join(base, ".git"), filepath.Join(base, ".vloop"), filepath.Join(base, ".vloop", "tmp", "plain"))
	if got := Root(filepath.Join(base, ".vloop", "tmp", "plain")); got != base {
		t.Fatalf("Root = %q, want %q", got, base)
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
