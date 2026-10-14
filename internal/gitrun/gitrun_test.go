package gitrun

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestHardenedCarriesTheOverrides(t *testing.T) {
	hooks := filepath.Join(t.TempDir(), "nohooks")
	_ = os.MkdirAll(hooks, 0o755)
	_ = os.WriteFile(filepath.Join(hooks, "stale"), nil, 0o644)
	cmd := Hardened("r", hooks, "status", "-z")
	want := []string{"git", "-C", "r", "-c", "core.hooksPath=" + hooks,
		"-c", "core.fsmonitor=false", "-c", "core.quotepath=false", "status", "-z"}
	if !slices.Equal(cmd.Args, want) {
		t.Errorf("args = %q, want %q", cmd.Args, want)
	}
	if ents, _ := os.ReadDir(hooks); len(ents) != 0 {
		t.Errorf("hooks dir not recreated empty: %v", ents)
	}
}

func TestPlainFormsCarryNoOverrides(t *testing.T) {
	for _, cmd := range [][]string{At("r", "log").Args, Plain("config", "--list").Args} {
		for _, a := range cmd {
			if a == "-c" || strings.HasPrefix(a, "core.") {
				t.Errorf("plain form carries %q: %q", a, cmd)
			}
		}
	}
	if got, want := At("r", "log").Args, []string{"git", "-C", "r", "log"}; !slices.Equal(got, want) {
		t.Errorf("At = %q, want %q", got, want)
	}
	if got, want := Plain("config").Args, []string{"git", "config"}; !slices.Equal(got, want) {
		t.Errorf("Plain = %q, want %q", got, want)
	}
}

func TestCaptureErrorText(t *testing.T) {
	_, err := Capture(t.TempDir(), "rev-parse", "--verify", "nosuchref")
	if err == nil {
		t.Fatal("want an error")
	}
	if !strings.HasPrefix(err.Error(), "git rev-parse --verify nosuchref: exit status ") {
		t.Errorf("error text = %q", err)
	}
}

func TestCaptureReturnsStdoutUntrimmed(t *testing.T) {
	out, err := Capture(t.TempDir(), "--version")
	if err != nil || !strings.HasPrefix(string(out), "git version") || !strings.HasSuffix(string(out), "\n") {
		t.Errorf("out = %q, err = %v", out, err)
	}
}
