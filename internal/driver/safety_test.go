package driver

import (
	"slices"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mvelosop/vloop/internal/state"
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

func TestLockExactlyOneOfConcurrent(t *testing.T) {
	root := t.TempDir()
	var wg sync.WaitGroup
	errs := make([]error, 2)
	start := make(chan struct{})
	for i := range errs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, errs[i] = AcquireLock(root, "main", "r", time.Now(), nil)
		}()
	}
	close(start)
	wg.Wait()
	won := 0
	for _, err := range errs {
		if err == nil {
			won++
			continue
		}
		var h *Halt
		if !errors.As(err, &h) || h.Code != ExitPreflight || !strings.Contains(err.Error(), "pid "+itoa(os.Getpid())) {
			t.Errorf("the loser: %v", err)
		}
	}
	if won != 1 {
		t.Fatalf("%d acquisitions succeeded, want exactly 1", won)
	}
}

func TestLockSymlinkNotFollowed(t *testing.T) {
	root := t.TempDir()
	lock := filepath.Join(root, filepath.FromSlash(lockFile))
	os.MkdirAll(filepath.Dir(lock), 0o755)
	target := filepath.Join(root, "target")
	if err := os.Symlink(target, lock); err != nil {
		t.Skip("no symlinks:", err)
	}
	release, err := AcquireLock(root, "main", "r", time.Now(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if _, err := os.Stat(target); err == nil {
		t.Error("the lock was written through the symlink")
	}
}

func TestHandoffSymlinkIsMissing(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real.json")
	os.WriteFile(real, []byte("{}"), 0o644)
	if _, err := readHandoff(real); err != nil {
		t.Fatalf("a regular file: %v", err)
	}
	link := filepath.Join(dir, "link.json")
	if err := os.Symlink(real, link); err != nil {
		t.Skip("no symlinks:", err)
	}
	if _, err := readHandoff(link); err == nil {
		t.Error("a symlinked handoff was read")
	}
}

// TestStateGuardGateFolders: a session's edit, addition or removal under the
// gate folders is undone, and the hash handed to the session covers them.
func TestStateGuardGateFolders(t *testing.T) {
	root := t.TempDir()
	write := func(rel, data string) {
		t.Helper()
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(".vloop/state/state.json", "{}\n")
	write(".vloop/state/gates/T2/oracle.sh", "test -f x\n")
	g := snapshotState(root)
	if g.restoreIfTouched() {
		t.Fatal("an untouched snapshot reported a change")
	}
	write(".vloop/state/gates/T2/oracle.sh", "exit 0\n")
	write(".vloop/state/gates/T3/new.sh", "true\n")
	// The warning names what the session changed, not the plan it left alone.
	if got := g.restoreTouched(); !slices.Equal(got, []string{".vloop/state/gates/T2/oracle.sh", ".vloop/state/gates/T3/new.sh"}) {
		t.Fatalf("restoreTouched = %v, want the two gate files and not state.json", got)
	}
	write(".vloop/state/gates/T2/oracle.sh", "exit 0\n")
	if !g.restoreIfTouched() {
		t.Fatal("a rewritten gate folder was not noticed")
	}
	if got, _ := os.ReadFile(filepath.Join(root, ".vloop", "state", "gates", "T2", "oracle.sh")); string(got) != "test -f x\n" {
		t.Fatalf("oracle: %q", got)
	}
	if _, err := os.Stat(filepath.Join(root, ".vloop", "state", "gates", "T3")); err == nil {
		t.Fatal("the added folder survived")
	}
	if g.hash() == planHashOfPlanOnly(g) {
		t.Fatal("the hash does not cover the gate folders")
	}
}

func planHashOfPlanOnly(g stateGuard) string { return state.PlanDigest(g.pre, nil) }
