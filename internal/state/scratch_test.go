package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEmptyScratch(t *testing.T) {
	root := t.TempDir()
	d := filepath.Join(root, "web", ".gate")
	if err := os.MkdirAll(filepath.Join(d, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"a.txt", "sub/b.txt"} {
		if err := os.WriteFile(filepath.Join(d, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := EmptyScratch(root, []string{"web/.gate/", "missing/"}); err != nil {
		t.Fatal(err)
	}
	if es, err := os.ReadDir(d); err != nil || len(es) != 0 {
		t.Fatalf("scratch holds %d entries (%v), want the folder kept and empty", len(es), err)
	}
}
