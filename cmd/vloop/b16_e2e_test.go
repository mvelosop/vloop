package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/mvelosop/vloop/internal/budget"
)

// The verification-phase brief's worked example (B16), played line for line
// against the built binary and the suite's stub claude, in temporary
// repositories with a temporary HOME. No verify session runs for real in the
// run that built it, so this is the brief's end-to-end evidence.

const b16Report2 = `case "$PHASE" in
  verify) mkdir -p .vloop/tmp
    printf '{"schema":"verify/v1","findings":[{"kind":"example","claim":"exit 2","observed":"exit 1","where":"Worked example, line 3; vloop greet --loud"},{"kind":"surface","claim":"greet prints in upper case with --loud","observed":"README.md says --loud is not supported","where":"README.md:12"}],"checked":{"example_lines":3,"sentences":4,"surfaces":2},"notes":"stub"}\n' > .vloop/tmp/verify.json;;
esac
`

func b16Squeeze(s string) string { return strings.Join(strings.Fields(s), " ") }

func b16Line(out, word string) string {
	for _, l := range strings.Split(out, "\n") {
		if s := b16Squeeze(l); strings.HasPrefix(s, word+" ") {
			return s
		}
	}
	return ""
}

// b16Run runs the brief in r with the verify script appended.
func b16Run(r *runRepo, verify string) result {
	r.scripted(oneTask(r.t), defaultScript+verify)
	return r.vloop("run", runBrief)
}

func TestWorkedExampleB16ConfigList(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	res := r.vloop("config", "list")
	wantExit(t, res, 0)
	for _, re := range []string{`(?m)^run\.verify=true( |$)`, `(?m)^model\.verify=sonnet( |$)`, `(?m)^effort\.verify=`} {
		if !regexp.MustCompile(re).MatchString(res.out) {
			t.Errorf("config list lacks %s:\n%s", re, res.out)
		}
	}
}

func TestWorkedExampleB16VerifyRunCloseMetrics(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	res := b16Run(r, b16Report2)
	wantExit(t, res, 0)
	r.wantStatus("complete")
	folder := filepath.Base(r.runFolder())
	wantIn(t, "stdout", res.out, "verify: 2 finding(s) — see .vloop/state/runs/"+runID+"/"+folder+"/reports/verify.json")
	files := r.git("show", "--name-only", "--format=", "HEAD")
	prefix := ".vloop/state/runs/" + runID + "/" + folder + "/"
	if !strings.Contains(files, prefix+"reports/verify.json") ||
		!regexp.MustCompile(regexp.QuoteMeta(prefix)+`sessions/\d+-verify\.json`).MatchString(files) {
		t.Errorf("the run commit holds:\n%s", files)
	}

	dry := r.vloop("brief", "close", runBrief, "--dry-run", "--no-findings")
	wantExit(t, dry, 0)
	lines := strings.Split(dry.out, "\n")
	if lines[0] != "verify of "+runID+"/"+folder+": 2 finding(s)" {
		t.Errorf("close --dry-run starts %q", lines[0])
	}
	if len(lines) < 2 || lines[1] != "  1. example — Worked example, line 3; vloop greet --loud" {
		t.Errorf("close --dry-run's first finding: %q", lines[:min(len(lines), 4)])
	}

	m := r.vloop("metrics", runBrief)
	wantExit(t, m, 0)
	if l := b16Line(m.out, "time"); !regexp.MustCompile(` · gate review [0-9.]+ min · verify [0-9.]+ min · gates `).MatchString(l) {
		t.Errorf("time line %q", l)
	}
	if l := b16Line(m.out, "cost"); !regexp.MustCompile(` · gate review [0-9.]+ · verify [0-9.]+ · `).MatchString(l) {
		t.Errorf("cost line %q", l)
	}
}

func TestWorkedExampleB16NoReportEditAndOff(t *testing.T) {
	t.Parallel()
	const warn = "vloop: the verify session left no valid .vloop/tmp/verify.json — verify by hand"
	r := newRunRepo(t)
	res := b16Run(r, "case \"$PHASE\" in verify) rm -f .vloop/tmp/verify.json;; esac\n")
	wantExit(t, res, 0)
	wantIn(t, "output", res.out+res.err, warn)

	r = newRunRepo(t)
	r.write("README.md", "# demo\n")
	r.commitAll("readme")
	res = b16Run(r, verifyEmptyReport+"[ \"$PHASE\" = verify ] && echo edit >> README.md\n")
	wantExit(t, res, 0)
	wantIn(t, "output", res.out+res.err, "vloop: the verify session changed README.md — reverted")
	if got := r.read("README.md"); got != "# demo\n" {
		t.Errorf("README.md = %q, want it reverted", got)
	}

	r = newRunRepo(t)
	wantExit(t, r.vloop("config", "set", "run.verify", "false"), 0)
	r.commitAll("verify off")
	res = b16Run(r, b16Report2)
	wantExit(t, res, 0)
	wantNotIn(t, "stdout", res.out, "verify:")
	if r.verifySessions() != 0 {
		t.Error("a verify session ran with run.verify = false")
	}
	dry := r.vloop("brief", "close", runBrief, "--dry-run", "--no-findings")
	wantExit(t, dry, 0)
	if first := strings.SplitN(dry.out, "\n", 2)[0]; first != "no verify report — verify by hand" {
		t.Errorf("close --dry-run with no report starts %q", first)
	}
}

// S1: stale scratch is gone at run start with .running held; a plan stub's
// unformatted scratch is gone before T1's check; the review sees proposal.json
// and checks.json.
func TestWorkedExampleB16Clearing(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.write(".vloop/config.toml", runCheckConfig+"\n[[check]]\nname = \"fmt\"\npaths = [\"**\"]\nrun = \"test -z \\\"$(gofmt -l .)\\\"\"\n")
	r.write("go.mod", "module demo\n\ngo 1.21\n")
	r.commitAll("fmt check")
	r.write(".vloop/tmp/old.txt", "")
	r.scripted(oneTask(t), defaultScript+`seen() { [ -e ".vloop/tmp/$1" ]; }
seen .running || echo "$PHASE: .running missing" >> "$dir/s1"
case "$PHASE" in
  plan) seen old.txt && echo "plan saw old.txt" >> "$dir/s1"
    mkdir -p .vloop/tmp/scratch; printf 'package x\nfunc  F( ){}\n' > .vloop/tmp/scratch/x.go;;
  review) for e in proposal.json checks.json; do seen $e || echo "review found no $e" >> "$dir/s1"; done;;
esac
`)
	wantExit(t, r.vloop("run", runBrief), 0)
	r.wantTask("T1", "done", 0)
	if b, err := os.ReadFile(filepath.Join(r.stub, "s1")); err == nil {
		t.Errorf("clearing broke S1:\n%s", b)
	}
	if r.has(".vloop/tmp/old.txt") {
		t.Error("old.txt outlived the run start")
	}
}

// K1/W: brief check of a ready brief with a "## Pinning tests" guidelines
// section prints no must-change list: line; init's guidelines carry neither
// retired section.
func TestWorkedExampleB16RetiredWarnings(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	g := r.read(".vloop/repo-guidelines.md")
	for _, s := range []string{"## Pinning tests", "## External formats"} {
		if strings.Contains(g, s) {
			t.Errorf("init's guidelines carry %q", s)
		}
	}
	r.write(".vloop/repo-guidelines.md", g+"\n## Pinning tests\n\n- A test that pins a message is listed.\n")
	res := r.vloop("brief", "check", runBrief)
	wantNotIn(t, "brief check", res.out+res.err, "must-change list:", "external format:")
}

// The brief's planted failures are each shown to be caught.
func TestWorkedExampleB16PlantedFailures(t *testing.T) {
	t.Parallel()
	// a verify skill padded past 1,500 words fails the budget test
	dst := t.TempDir()
	src := filepath.Join("..", "..", "plugin", "skills", "verify", "SKILL.md")
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dst, "skills", "verify"), 0o755); err != nil {
		t.Fatal(err)
	}
	padded := string(b) + "\n" + strings.Repeat("pad ", 1501)
	if err := os.WriteFile(filepath.Join(dst, "skills", "verify", "SKILL.md"), []byte(padded), 0o644); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range budget.Check(dst) {
		if strings.HasPrefix(m, "verify:") {
			found = true
		}
	}
	if !found {
		t.Error("a verify skill padded past its ceiling is not named by the budget check")
	}

	// a clearing that removes .running is caught by the sessions' check
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript+`[ "$PHASE" = plan ] && rm -f .vloop/tmp/.running
[ -e .vloop/tmp/.running ] || echo "$PHASE: .running missing" >> "$dir/s1"
`)
	r.vloop("run", runBrief)
	if _, err := os.Stat(filepath.Join(r.stub, "s1")); err != nil {
		t.Error("a missing .running is not caught by the sessions' check")
	}
}

// A driver that ends the run blocked on a verify finding fails V1's
// assertions: exit 0 and status complete.
func TestWorkedExampleB16PlantedBlockedDriver(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript+b16Report2)
	wrap := filepath.Join(t.TempDir(), "plant-vloop")
	body := "#!/bin/sh\n'" + binPath + "' \"$@\"; c=$?\nif [ \"$1\" = run ] && [ \"$c\" = 0 ] && ls .vloop/state/runs/*/*/reports/verify.json >/dev/null 2>&1; then exit 2; fi\nexit $c\n"
	if err := os.WriteFile(wrap, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	res := r.exec(wrap, r.env(), "run", runBrief)
	if res.code == 0 {
		t.Error("a driver that ends blocked on a verify finding exits 0")
	}
	if !strings.Contains(res.out, "verify: 2 finding(s)") {
		t.Errorf("the planted run did not reach the verify line:\n%s", res.out)
	}
}
