package driver

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestLockLiveStaleAndRelease(t *testing.T) {
	root := t.TempDir()
	lock := filepath.Join(root, filepath.FromSlash(lockFile))
	if err := os.MkdirAll(filepath.Dir(lock), 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(s string) {
		t.Helper()
		if err := os.WriteFile(lock, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write(`{"pid":` + itoa(os.Getpid()) + `,"branch":"b"}`)
	_, err := AcquireLock(root, "main", "r", time.Now(), nil)
	var h *Halt
	if !errors.As(err, &h) || h.Code != ExitPreflight || !strings.Contains(err.Error(), "git worktree add") {
		t.Fatalf("a live lock: %v", err)
	}
	if _, err := os.Stat(lock); err != nil {
		t.Error("a live lock was removed")
	}

	write(`{"pid":"2147483646"}`)
	var warned string
	release, err := AcquireLock(root, "main", "r", time.Now(), func(f string, a ...any) { warned = f })
	if err != nil || !strings.Contains(warned, "stale lock") {
		t.Fatalf("a stale lock: %v, warned %q", err, warned)
	}
	release()
	release()
	if _, err := os.Stat(lock); !errors.Is(err, os.ErrNotExist) {
		t.Error("the lock survived its release")
	}
}

func itoa(n int) string { return strconv.Itoa(n) }

func TestStateGuard(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, ".vloop", "state", "state.json")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(p, []byte("{\"a\": 1}\n"), 0o644)
	g := snapshotState(root)
	if g.restoreIfTouched() {
		t.Error("an untouched plan reported as touched")
	}
	os.WriteFile(p, []byte("{\"a\": 1} \n"), 0o644)
	if !g.restoreIfTouched() {
		t.Fatal("a one-byte edit went unnoticed")
	}
	if b, _ := os.ReadFile(p); string(b) != "{\"a\": 1}\n" {
		t.Errorf("not restored: %q", b)
	}
}
