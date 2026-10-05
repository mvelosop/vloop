package state

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func writeFile(t *testing.T, root, rel, data string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestFixturesDigest pins the digest: <path> NUL <contents> NUL per regular
// file, sorted by "/"-separated relative path, the same on every OS.
func TestFixturesDigest(t *testing.T) {
	dir := t.TempDir()
	if d, err := FixturesDigest(filepath.Join(dir, "none")); err != nil || d != "" {
		t.Fatalf("missing folder: %q %v", d, err)
	}
	writeFile(t, dir, "seed/rows.txt", "row1\n")
	writeFile(t, dir, "oracle.sh", "test -f x\n")
	sum := sha256.Sum256([]byte("oracle.sh\x00test -f x\n\x00seed/rows.txt\x00row1\n\x00"))
	want := hex.EncodeToString(sum[:])
	if got, err := FixturesDigest(dir); err != nil || got != want {
		t.Fatalf("got %q %v, want %q", got, err, want)
	}
	writeFile(t, dir, "seed/rows.txt", "row2\n")
	if got, _ := FixturesDigest(dir); got == want {
		t.Fatal("a changed file kept the digest")
	}
	empty := t.TempDir()
	if err := os.MkdirAll(filepath.Join(empty, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if d, _ := FixturesDigest(empty); d != "" {
		t.Fatalf("a folder with no file: %q", d)
	}
}

func TestStampAndStray(t *testing.T) {
	root := t.TempDir()
	p := &Plan{Tasks: []Task{{ID: "T1"}, {ID: "T2"}}}
	writeFile(t, root, GateFolder("T2")+"/oracle.sh", "true\n")
	if err := StampFixtures(root, p); err != nil {
		t.Fatal(err)
	}
	if p.Tasks[0].Fixtures != "" || len(p.Tasks[1].Fixtures) != 64 {
		t.Fatalf("stamp: %+v", p.Tasks)
	}
	if id, _ := FixturesMismatch(root, p); id != "" {
		t.Fatalf("fresh stamp mismatches %s", id)
	}
	writeFile(t, root, GateFolder("T9")+"/x.sh", "true\n")
	if stray, _ := StrayGateFolders(root, p); !reflect.DeepEqual(stray, []string{"T9"}) {
		t.Fatalf("stray: %v", stray)
	}
	writeFile(t, root, GateFolder("T2")+"/oracle.sh", "false\n")
	if id, _ := FixturesMismatch(root, p); id != "T2" {
		t.Fatalf("mismatch: %q", id)
	}
}

func TestPlanDigestAndRestore(t *testing.T) {
	plan := []byte("{}\n")
	sum := sha256.Sum256(plan)
	if PlanDigest(plan, nil) != hex.EncodeToString(sum[:]) {
		t.Fatal("with no gate folder the digest is the plan's SHA-256")
	}
	root := t.TempDir()
	writeFile(t, root, GateFolder("T2")+"/a.sh", "a")
	before, _ := GateFiles(root)
	d0 := PlanDigest(plan, before)
	if d0 == PlanDigest(plan, nil) {
		t.Fatal("gate files are not covered")
	}
	writeFile(t, root, GateFolder("T2")+"/a.sh", "b")
	writeFile(t, root, GateFolder("T3")+"/n/x.sh", "x")
	changed := RestoreGateFiles(root, before)
	if !reflect.DeepEqual(changed, []string{"T2/a.sh", "T3/n/x.sh"}) {
		t.Fatalf("changed: %v", changed)
	}
	after, _ := GateFiles(root)
	if PlanDigest(plan, after) != d0 {
		t.Fatal("not restored")
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(GateFolder("T3")))); err == nil {
		t.Fatal("the stray folder was left")
	}
}

func TestRecordGate(t *testing.T) {
	root := t.TempDir()
	p := &Plan{Tasks: []Task{{ID: "T1", Verify: "old"}}}
	at := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := RecordGate(root, p, "T1", "", "r", at); err == nil || err.Error() != "nothing to record — T1's gate and fixtures are unchanged" {
		t.Fatalf("nothing changed: %v", err)
	}
	if err := RecordGate(root, p, "T1", "old", "r", at); err == nil {
		t.Fatal("the same command counted as new")
	}
	if err := RecordGate(root, p, "T9", "x", "r", at); err == nil || err.Error() != "no task T9 — vloop task list shows the plan's tasks" {
		t.Fatalf("unknown task: %v", err)
	}
	if err := RecordGate(root, p, "T1", "new", "why", at); err != nil {
		t.Fatal(err)
	}
	want := []GateReplace{{Verify: "old", ReplacedAt: "2026-01-02T03:04:05Z", Reason: "why", By: "operator"}}
	if p.Tasks[0].Verify != "new" || !reflect.DeepEqual(p.Tasks[0].GateHistory, want) {
		t.Fatalf("got %+v", p.Tasks[0])
	}
	writeFile(t, root, GateFolder("T1")+"/s.txt", "s")
	if err := RecordGate(root, p, "T1", "", "seed", at); err != nil {
		t.Fatal(err)
	}
	h := p.Tasks[0].GateHistory
	if len(h) != 2 || h[1].Verify != "new" || h[1].Fixtures != "" || p.Tasks[0].Fixtures == "" || p.Tasks[0].Verify != "new" {
		t.Fatalf("got %+v", p.Tasks[0])
	}
}
