package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestWorkedExampleB11 plays the quality-pass brief's worked example line for
// line against the built binary, in temporary directories.

const b11Plan = `{
  "schema": "state/v2",
  "run_id": "B20260101-0900-demo",
  "brief": "docs/briefs/B20260101-0900-demo.loop-brief.md",
  "base": "",
  "branch": "",
  "status": "planning",
  "iteration": 0,
  "created": "2026-01-01T09:00:00Z",
  "updated": "2026-01-01T09:00:00Z",
  "shell": "sh",
  "checks": [{"name": "all", "paths": ["**"], "run": "true"}],
  "gate_scratch": [],
  "gate_review": {"rounds": 0, "verdict": ""},
  "tasks": [{
    "id": "T1", "title": "Task T1", "goal": "Do T1.", "kind": "feature",
    "fixtures": "", "references": [], "depends_on": [],
    "acceptance": ["T1.out exists"], "verify": "test -f T1.out",
    "status": "pending", "attempts": 0, "notes": ""
  }]
}
`

func crlf(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\n", "\r\n")
}

func allCRLF(s string) bool {
	return strings.Count(s, "\n") == strings.Count(s, "\r\n")
}

func TestWorkedExampleB11(t *testing.T) {
	t.Parallel()
	s := newScratch(t)
	git := func(args ...string) {
		t.Helper()
		full := append([]string{"-c", "user.name=e2e", "-c", "user.email=e2e@example.invalid", "-c", "commit.gpgsign=false"}, args...)
		cmd := exec.Command("git", full...)
		cmd.Dir = s.dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("add", "-A")
	git("commit", "-q", "-m", "base")
	if r := s.run(nil, "init"); r.code != 0 {
		t.Fatalf("vloop init: %d %s%s", r.code, r.out, r.err)
	}
	fail := func(r result, code int, msg string, args ...string) {
		t.Helper()
		if r.code != code || strings.TrimSpace(r.err) != msg {
			t.Errorf("vloop %v: exit %d, stderr %q; want exit %d, %q", args, r.code, r.err, code, msg)
		}
	}
	errOf := func(code int, msg string, args ...string) {
		t.Helper()
		fail(s.run(nil, args...), code, msg, args...)
	}

	errOf(1, "vloop: -C /nonexistent: no such directory", "-C", "/nonexistent", "status")
	errOf(1, "vloop: brief not found: docs/briefs/nope.md", "run", "docs/briefs/nope.md")
	errOf(1, "vloop: no such file: missing.md", "brief", "check", "missing.md")
	if r := s.run(nil, "--bogus"); r.code != 2 || !strings.Contains(r.err, "unknown flag") {
		t.Errorf("vloop --bogus: exit %d, stderr %q; want exit 2, unknown flag", r.code, r.err)
	}
	errOf(1, "vloop: no plan — vloop run <brief> makes one", "task", "show", "T1")

	s.write(".vloop/state/state.json", b11Plan)
	errOf(1, "vloop: no task T99 — vloop task list shows the plan's tasks", "task", "show", "T99")
	errOf(1, "vloop: no brief nope — vloop brief list shows the briefs", "metrics", "nope")
	s.write(".vloop/state/state.json", "{bad")
	errOf(1, "vloop: .vloop/state/state.json is not valid JSON (line 1, column 2)", "status")
	s.write(".vloop/state/state.json", `{"version":"x"}`)
	r := s.run(nil, "status", "--json")
	var e struct{ Error string }
	if err := json.Unmarshal([]byte(r.out), &e); err != nil || r.code != 1 ||
		!strings.Contains(e.Error, "is not a valid plan — vloop task validate lists the problems") {
		t.Errorf("status --json on an invalid plan: exit %d, out %q", r.code, r.out)
	}
	if err := os.Remove(filepath.Join(s.dir, ".vloop", "state", "state.json")); err != nil {
		t.Fatal(err)
	}

	// Only draft briefs.
	d := newScratch(t)
	if r := d.run(nil, "init"); r.code != 0 {
		t.Fatalf("vloop init: %d %s%s", r.code, r.out, r.err)
	}
	if r := d.run(nil, "brief", "new", "second"); r.code != 0 {
		t.Fatalf("brief new: %d %s", r.code, r.err)
	}
	drafts, err := filepath.Glob(filepath.Join(d.dir, "docs", "briefs", "*.loop-brief.md"))
	if err != nil || len(drafts) != 2 {
		t.Fatalf("drafts: %v %v", drafts, err)
	}
	args := []string{"brief", "check"}
	for _, p := range drafts {
		rel, _ := filepath.Rel(d.dir, p)
		args = append(args, filepath.ToSlash(rel))
	}
	if r := d.run(nil, args...); r.code != 0 || !strings.Contains("\n"+r.out, "\nnothing checked: 2 draft brief(s) skipped\n") {
		t.Errorf("brief check on drafts: exit %d, out %q, err %q", r.code, r.out, r.err)
	}

	// Outside any repository, and in a directory that is not a git repository.
	outside := t.TempDir()
	run := func(dir string, args ...string) result {
		c := &scratch{t: t, dir: dir, home: s.home}
		return c.run(nil, args...)
	}
	r = run(outside, "plugin", "path")
	fail(r, 1, "vloop: not in a vloop repository — run vloop init first", "plugin", "path")
	if _, err := os.Stat(filepath.Join(outside, ".vloop")); err == nil {
		t.Error("plugin path outside any repository created .vloop/")
	}
	r = run(outside, "metrics")
	if r.code != 1 || !strings.HasPrefix(r.err, "vloop: ") || !strings.HasSuffix(strings.TrimSpace(r.err), " is not a git repository") {
		t.Errorf("metrics outside a git repository: exit %d, %q", r.code, r.err)
	}

	// Records with CRLF and a BOM.
	briefs, _ := filepath.Glob(filepath.Join(s.dir, "docs", "briefs", "*.loop-brief.md"))
	if len(briefs) == 0 {
		t.Fatal("no brief after init")
	}
	brief := strings.TrimSuffix(filepath.Base(briefs[0]), ".md")
	r = s.run(nil, "defect", "add", "a CRLF defect", "--found-by", "operator", "--brief", brief)
	if r.code != 0 {
		t.Fatalf("defect add: %d %s", r.code, r.err)
	}
	dRel := strings.TrimSpace(r.out)
	s.write(dRel, crlf(s.read(dRel)))
	if r := s.run(nil, "defect", "set", strings.TrimSuffix(filepath.Base(dRel), ".md"), "status", "fixed"); r.code != 0 {
		t.Fatalf("defect set on a CRLF defect: %d %s", r.code, r.err)
	}
	if got := s.read(dRel); !strings.Contains(got, "status: fixed\r\n") || !allCRLF(got) {
		t.Errorf("CRLF defect after set: %q", got)
	}
	r = s.run(nil, "intervention", "add", "a BOM and CRLF intervention", "--phase", "run", "--kind", "repair",
		"--automatable", "no", "--by", "operator", "--option", "one", "--option", "two", "--recommended", "1",
		"--why", "because", "--decided-option", "1")
	if r.code != 0 {
		t.Fatalf("intervention add: %d %s", r.code, r.err)
	}
	iRel := strings.TrimSpace(r.out)
	s.write(iRel, "\xef\xbb\xbf"+crlf(s.read(iRel)))
	iID := strings.TrimSuffix(filepath.Base(iRel), ".md")
	if r := s.run(nil, "intervention", "set", iID, "phase", "verify"); r.code != 0 {
		t.Fatalf("intervention set on a BOM + CRLF record: %d %s", r.code, r.err)
	}
	if got := s.read(iRel); !strings.HasPrefix(got, "\xef\xbb\xbf") || !allCRLF(got) || !strings.Contains(got, "phase: verify\r\n") {
		t.Errorf("BOM + CRLF intervention after set: %q", got)
	}
	if r := s.run(nil, "intervention", "set", iID, "decided", "5"); r.code != 2 {
		t.Errorf("intervention set decided 5: exit %d, want 2", r.code)
	}

	// cmd/gendocs refuses a flag-like argument.
	gd := t.TempDir()
	gdBin := filepath.Join(t.TempDir(), "gendocs")
	if runtime.GOOS == "windows" {
		gdBin += ".exe"
	}
	if out, err := exec.Command("go", "build", "-o", gdBin, "../gendocs").CombinedOutput(); err != nil {
		t.Fatalf("build gendocs: %v\n%s", err, out)
	}
	cmd := exec.Command(gdBin, "--help")
	cmd.Dir = gd
	if err := cmd.Run(); err == nil {
		t.Error("cmd/gendocs --help was not refused")
	}
	if _, err := os.Stat(filepath.Join(gd, "--help")); err == nil {
		t.Error("cmd/gendocs --help wrote a file named --help")
	}

	// doctor with no marketplace plugin: a fake claude that lists none.
	if runtime.GOOS != "windows" {
		bin := t.TempDir()
		fake := "#!/bin/sh\ncase \"$1\" in\n  --version) echo \"2.0.0 (Claude Code)\"; exit 0;;\n  plugin) echo '[]'; exit 0;;\nesac\nexit 1\n"
		if err := os.WriteFile(filepath.Join(bin, "claude"), []byte(fake), 0o755); err != nil {
			t.Fatal(err)
		}
		r := s.run([]string{"PATH=" + bin + string(os.PathListSeparator) + os.Getenv("PATH")}, "doctor")
		if !strings.Contains(r.out+r.err, "✓ plugin the plugin is supplied by vloop run (--plugin-dir); ") {
			t.Errorf("doctor with no marketplace plugin: %s%s", r.out, r.err)
		}
	}

	// --help says what a brief is and to start with vloop init.
	r = s.run(nil, "--help")
	if h := strings.ToLower(r.out + r.err); !strings.Contains(h, "brief") || !strings.Contains(h, "vloop init") {
		t.Errorf("vloop --help: %s", r.out+r.err)
	}
}
