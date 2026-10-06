package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The findings of B11's verification, each pinned before its fix.

func TestRunHelpExitCodesFollowUsageOnly(t *testing.T) {
	_, out, _ := runCLI(t, t.TempDir(), "run", "--help")
	if !strings.Contains(strings.Join(strings.Fields(out), " "), "1 preflight or failure; 2 usage or blocked;") {
		t.Errorf("run --help does not say 1 preflight or failure; 2 usage or blocked:\n%s", out)
	}
}

func TestShortsFollowListAndShow(t *testing.T) {
	short := func(help, name string) string {
		for _, l := range strings.Split(help, "\n") {
			if f := strings.Fields(l); len(f) > 1 && f[0] == name {
				return strings.Join(f[1:], " ")
			}
		}
		return ""
	}
	_, root, _ := runCLI(t, t.TempDir(), "--help")
	_, metrics, _ := runCLI(t, t.TempDir(), "metrics", "--help")
	for name, got := range map[string]string{"metrics": short(root, "metrics"), "schema": short(root, "schema"), "metrics export": short(metrics, "export")} {
		if !strings.HasPrefix(got, "List") && !strings.HasPrefix(got, "Show") {
			t.Errorf("%s: Short %q starts with neither List nor Show", name, got)
		}
	}
	_, out, _ := runCLI(t, t.TempDir(), "config", "set", "--help")
	if first := strings.SplitN(out, "\n", 2)[0]; !strings.Contains(first, "config list") {
		t.Errorf("config set's Short %q does not say which keys it sets", first)
	}
}

func TestInitDoesNotPointAtAMarketplace(t *testing.T) {
	d := initRepo(t, map[string]string{"go.mod": "module a\n"})
	out, _, _ := runPluginCLI(t, initBuild, d, "init")
	if strings.Contains(out, "marketplace") || !strings.Contains(out, `claude --plugin-dir "$(vloop plugin path)"`) {
		t.Errorf("init's next steps:\n%s", out)
	}
}

func TestInterventionsMetricsRefusesArguments(t *testing.T) {
	code, _, e := runCLI(t, t.TempDir(), "metrics", "--interventions", "bogus")
	if code != 2 {
		t.Errorf("metrics --interventions bogus: exit %d (%q), want 2", code, e)
	}
	_, out, _ := runCLI(t, t.TempDir(), "metrics", "--help")
	if strings.Contains(out, "--interventions kind") {
		t.Errorf("metrics --help shows --interventions as taking a value:\n%s", out)
	}
}

func TestTaskValidateSchemaProblemHasNoEmptyField(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".vloop", "state"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".vloop", "state", "state.json"), []byte(`{"version":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, out, e := runCLI(t, dir, "task", "validate")
	if strings.Contains(out+e, ": :") {
		t.Errorf("task validate prints an empty field:\n%s%s", out, e)
	}
}

func TestBriefCheckMissingFileOneWording(t *testing.T) {
	dir := t.TempDir()
	_, _, e := runCLI(t, dir, "brief", "check", "docs/briefs/nope.loop-brief.md")
	if !strings.Contains(e, "vloop: no such file: docs/briefs/nope.loop-brief.md") {
		t.Errorf("brief check on a missing brief: %q", e)
	}
}
