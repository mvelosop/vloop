package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunGateLogRedacted: a gate that dumps the environment leaves the value
// of a secret-named variable out of the committed gate log.
func TestRunGateLogRedacted(t *testing.T) {
	r := newRunRepo(t)
	r.scripted(planJSON(t, planTask("T1", map[string]any{"verify": "env; test -f T1.out"})), defaultScript)
	wantExit(t, r.runWith([]string{"FOO_TOKEN=s3cr3t-value", "SHORT_KEY=abc123"}, runBrief), 0)

	log, err := os.ReadFile(filepath.Join(r.runFolder(), "gates", "T1.log"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(log), "<redacted:FOO_TOKEN>") || strings.Contains(string(log), "s3cr3t-value") {
		t.Errorf("gate log:\n%s", log)
	}
	if !strings.Contains(string(log), "SHORT_KEY=abc123") {
		t.Error("a value under 8 characters should stay")
	}
	if out := r.git("log", "--all", "-p", "-S", "s3cr3t-value", "--oneline"); out != "" {
		t.Errorf("committed: %s", out)
	}
}
