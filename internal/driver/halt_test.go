package driver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSpendSumsThisRunsSessions(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "sessions"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"001-plan.json": `{"cost_usd": 1.25}`, "002-work.json": `{"cost_usd": 0.5}`, "003-bad.json": `not json`,
	} {
		if err := os.WriteFile(filepath.Join(dir, "sessions", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if got := (&Iterator{RunDir: dir}).spend(); got != 1.75 {
		t.Errorf("spend = %v, want 1.75", got)
	}
	if got := (&Iterator{RunDir: t.TempDir()}).spend(); got != 0 {
		t.Errorf("spend with no sessions = %v, want 0", got)
	}
}
