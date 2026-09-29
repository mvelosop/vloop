package cli

import (
	"bytes"
	"strings"
	"testing"
)

func runCLI(t *testing.T, dir string, args ...string) (int, string, string) {
	t.Helper()
	var o, e bytes.Buffer
	code := Execute(Build{}, append([]string{"-C", dir}, args...), &o, &e)
	return code, o.String(), e.String()
}

func TestMetricsStacks(t *testing.T) {
	dir := t.TempDir()
	code, out, _ := runCLI(t, dir, "metrics", "stacks")
	if code != 0 || !strings.HasPrefix(out, "csharp\ngo\njava\n") {
		t.Errorf("stacks: %d %q", code, out)
	}
	_, out, _ = runCLI(t, dir, "metrics", "stacks", "go")
	if out != "code:\n  **/*.go\ntest:\n  **/*_test.go\n  **/testdata/**\ndocs:\n  **/*.md\nexcluded:\n  go.sum\n  vendor/**\n" {
		t.Errorf("stacks go: %q", out)
	}
	_, out, _ = runCLI(t, dir, "--json", "metrics", "stacks", "go")
	if !strings.HasPrefix(out, `{"go":{"code":["**/*.go"]`) {
		t.Errorf("stacks go --json: %q", out)
	}
	code, out, errs := runCLI(t, dir, "metrics", "stacks", "cobol")
	if code != 2 || out != "" || errs != "vloop: unknown stack \"cobol\"\n" {
		t.Errorf("unknown: %d %q %q", code, out, errs)
	}
}

func TestMetricsClassify(t *testing.T) {
	dir := t.TempDir()
	if code, _, _ := runCLI(t, dir, "config", "set", "metrics.stacks", "go"); code != 0 {
		t.Fatal("set stacks")
	}
	code, out, _ := runCLI(t, dir, "metrics", "classify", "main.go", "go.sum", ".loop/x", "notes.txt")
	want := "main.go  code  (go: **/*.go)\ngo.sum  excluded  (go: go.sum)\n.loop/x  excluded  (always: .loop/**)\nnotes.txt  other  (none)\n"
	if code != 0 || out != want {
		t.Errorf("classify: %d %q", code, out)
	}
}

func TestConfigSetUnknownStack(t *testing.T) {
	dir := t.TempDir()
	code, _, errs := runCLI(t, dir, "config", "set", "metrics.stacks", "cobol")
	if code != 2 || !strings.HasPrefix(errs, "vloop: invalid value \"cobol\" for metrics.stacks: want ") {
		t.Errorf("%d %q", code, errs)
	}
}
