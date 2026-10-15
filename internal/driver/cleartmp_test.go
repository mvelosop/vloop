package driver

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestClearTmp(t *testing.T) {
	root := t.TempDir()
	tmp := filepath.Join(root, ".vloop", "tmp")
	outside := t.TempDir()
	keep := filepath.Join(outside, "keep.txt")
	for _, d := range []string{filepath.Join(tmp, "olddir"), filepath.Join(tmp, "plugin", "0.1.0")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{keep, filepath.Join(tmp, ".running"), filepath.Join(tmp, "proposal.json"), filepath.Join(tmp, "olddir", "x")} {
		if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(outside, filepath.Join(tmp, "linkdir")); err != nil {
		t.Skipf("no symlinks: %v", err)
	}

	var warned []string
	ClearTmp(root, func(entry string, err error) { warned = append(warned, entry) })

	if len(warned) != 0 {
		t.Fatalf("warned %v", warned)
	}
	left, _ := os.ReadDir(tmp)
	if len(left) != 1 || left[0].Name() != ".running" {
		t.Fatalf("left %v, want only .running", left)
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatalf("a symbolic link was followed: %v", err)
	}
}

func TestClearTmpWithoutFolder(t *testing.T) {
	ClearTmp(t.TempDir(), func(entry string, err error) { t.Fatalf("warned %q: %v", entry, err) })
}

func TestClearTmpWarnsAndGoesOn(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("cannot make an entry unremovable here")
	}
	root := t.TempDir()
	tmp := filepath.Join(root, ".vloop", "tmp")
	locked := filepath.Join(tmp, "locked")
	if err := os.MkdirAll(locked, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(locked, "f"), nil, 0o644)
	os.WriteFile(filepath.Join(tmp, "other"), nil, 0o644)
	os.Chmod(locked, 0o555)
	t.Cleanup(func() { os.Chmod(locked, 0o755) })

	var warned []string
	ClearTmp(root, func(entry string, err error) { warned = append(warned, entry+": "+err.Error()) })

	if len(warned) != 1 || !strings.HasPrefix(warned[0], "locked: ") {
		t.Fatalf("warnings %v, want one for locked", warned)
	}
	if _, err := os.Stat(filepath.Join(tmp, "other")); err == nil {
		t.Fatal("the clearing stopped at the entry it could not remove")
	}
}
