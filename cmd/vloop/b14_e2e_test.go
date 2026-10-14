package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The learning-loop brief's worked example (B14), played line for line against
// the built binary and the suite's stub claude: one temporary repository per
// scenario with a temporary HOME, addressed with git -C. Values the binary or
// git computes (the root, the end commit, word counts, ids) are read back, not
// hardcoded. The real-data checks read this repository's own history and skip
// where a clone does not carry it.

const (
	b14File = ".vloop/repo-guidelines.md"
	b14Seed = `# Repository guidelines — vloop

Lessons that belong to this repository. Every loop session receives this file.

- A brief never asks a session or a gate to write under ` + "`.vloop/`" + ` outside ` + "`.vloop/tmp/`" + `; records, config and state are the driver's and the operator's (S-4). (D20261004-2304)
- An external tool's format is taken from the tool or its docs, never from memory. (D20261001-1338)

## Pinning tests

- when ` + "`schemas/`" + `: ` + "`TestWorkedExampleB2Commands`, `TestSchemasAreEmbedded`, `TestWorkedExampleB6RealData`" + ` (D20261004-1142, D20261004-2304)
- when ` + "`config key`" + `: ` + "`TestConfigListDefaults`, `TestConfigListJSONOrderAndNull`, `TestDefaultsEveryRow`, `TestWorkedExampleEnglishSession`" + ` (D20261011-1455)
- when ` + "`closing lines`" + `: ` + "`TestInitWritesEverything`, `TestWorkedExampleB5Session`, `TestInitPointsAtBothWaysToLoadThePlugin`" + ` (D20261011-2239)

## External formats

- ` + "`eval case`" + `: ` + "`docs/guide/evals.md`" + ` (D20261001-1338)
`
)

// b14Commit commits everything in the harness repository.
func (r *runRepo) b14Commit(msg string) {
	r.t.Helper()
	r.git("add", "-A")
	r.git("commit", "-q", "-m", msg)
}

// b14Brief is a ready brief that passes every check but the two this brief
// adds: what is "What it is", contract is appended to the behaviour contract,
// constraints are extra Constraints lines, refs extra Binding references.
func b14Brief(name, what, contract, constraints, refs string) string {
	return strings.Join([]string{
		"---", "name: " + name, "description: A demo", "kind: brief", "status: ready",
		"created: 2026-01-01", "seeds: A demo", "depends-on: []", "---",
		"# Brief", "", "## What it is", "", what, "",
		"## Why this shape, and what was rejected", "", "Because.", "",
		"## Binding references", "", "- `docs/x.md` — the error contract", refs, "",
		"## Behaviour contract", "", "An unknown id exits 1.", "", contract, "",
		"## Worked example", "", "```", "$ demo", "ok", "```", "",
		"## Out of scope", "", "- one", "- two", "",
		"## Constraints", "", "- Go.", constraints, "",
		"## Shape", "", "2 to 3 tasks.", "",
	}, "\n")
}

// b14Check runs vloop brief check on a brief written into the repository and
// returns its exit code and the report's lines with the indentation removed.
func (r *runRepo) b14Check(n int, brief string) (int, []string) {
	r.t.Helper()
	rel := fmt.Sprintf("docs/briefs/B20260101-%04d-case.loop-brief.md", 1000+n)
	r.write(rel, brief)
	res := r.vloop("brief", "check", rel)
	var lines []string
	for _, l := range strings.Split(res.out+res.err, "\n") {
		lines = append(lines, strings.TrimSpace(l))
	}
	return res.code, lines
}

func b14HasLine(t *testing.T, lines []string, want string) {
	t.Helper()
	for _, l := range lines {
		if l == want {
			return
		}
	}
	t.Fatalf("no line %q in:\n%s", want, strings.Join(lines, "\n"))
}

func b14NoWarning(t *testing.T, lines []string, prefix string) {
	t.Helper()
	for _, l := range lines {
		if strings.HasPrefix(l, "! "+prefix) {
			t.Fatalf("unexpected warning %q", l)
		}
	}
}

func b14Count(s string) int { return len(strings.Fields(s)) }

// b14Thousands formats n with a comma between thousands, as the report does.
func b14Thousands(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

// b14AfterLine is the report line after the first line whose text (mark
// removed, spaces collapsed) starts with name.
func b14AfterLine(t *testing.T, out, name string) string {
	t.Helper()
	lines := b5Lines(out)
	for i, l := range lines {
		rest := l
		if mark, r, ok := strings.Cut(l, " "); ok && len([]rune(mark)) == 1 {
			rest = r
		}
		if strings.HasPrefix(rest, name) && i+1 < len(lines) {
			return lines[i+1]
		}
	}
	t.Fatalf("no %q line in %q", name, out)
	return ""
}

// b14Mask replaces the paths of this run's repository and stub so session
// arguments of two repositories compare.
func (r *runRepo) b14Mask(s string) string {
	real, _ := filepath.EvalSymlinks(r.dir)
	for _, p := range []string{real, r.dir} {
		s = strings.ReplaceAll(s, p, "<repo>")
	}
	return strings.ReplaceAll(s, r.stub, "<stub>")
}

// b14Sessions runs vloop run with stub sessions and returns the logged
// arguments of the -p sessions, paths masked.
func (r *runRepo) b14Sessions() []string {
	r.t.Helper()
	r.scripted(planJSON(r.t, planTask("T1", nil)), defaultScript)
	res := r.vloop("run", runBrief)
	wantExit(r.t, res, 0)
	var out []string
	for _, l := range r.argv() {
		if strings.Contains(l, " -p ") || strings.HasPrefix(l, "-p ") {
			out = append(out, r.b14Mask(l))
		}
	}
	if len(out) == 0 {
		r.t.Fatalf("no session was started: %v", r.argv())
	}
	return out
}

func TestWorkedExampleB14Init(t *testing.T) {
	t.Parallel()
	const version = "0.9.0"
	r := b5New(t, false, false)

	first := r.run(version, "init")
	b5Code(t, first, 0)
	lines := b5Lines(first.out)
	at := -1
	for i, l := range lines {
		if l == "wrote .vloop/install.json" {
			at = i
		}
	}
	if at < 0 || at+1 >= len(lines) || lines[at+1] != "wrote "+b14File {
		t.Fatalf("init does not print 'wrote %s' right after install.json: %q", b14File, lines)
	}
	for _, h := range []string{"## Pinning tests", "## External formats"} {
		if !strings.Contains(r.read(b14File), "\n"+h+"\n") {
			t.Fatalf("the template has no %q:\n%s", h, r.read(b14File))
		}
	}

	r.write(b14File, "# Mine\n\n- keep me (D20261011-1455)\n")
	mine := r.read(b14File)
	again := r.run(version, "init") // refuses: already initialized
	if r.read(b14File) != mine {
		t.Fatalf("init changed an existing file:\n%s", r.read(b14File))
	}
	for _, l := range b5Lines(again.out) {
		if l == "wrote "+b14File {
			t.Fatalf("init printed %q for a kept file", l)
		}
	}

	// An existing file in a repository not yet initialised is kept as well.
	k := b5New(t, false, false)
	k.write(b14File, mine)
	kept := k.run(version, "init")
	b5Code(t, kept, 0)
	if k.read(b14File) != mine {
		t.Fatalf("init changed an existing file:\n%s", k.read(b14File))
	}
	for _, l := range b5Lines(kept.out) {
		if l == "wrote "+b14File {
			t.Fatalf("init printed %q for a kept file", l)
		}
	}
	b5Has(t, kept, "wrote .vloop/install.json")
}

func TestWorkedExampleB14BriefCheck(t *testing.T) {
	t.Parallel()
	const (
		all  = "TestConfigListDefaults, TestConfigListJSONOrderAndNull, TestDefaultsEveryRow, TestWorkedExampleEnglishSession"
		tail = " (" + b14File + ")"
		mc   = `! must-change list: the contract mentions "config key" but Constraints do not name `
		ef   = `! external format: the contract mentions "eval case" but Binding references do not cite docs/guide/evals.md` + tail
	)
	r := newRunRepo(t)
	r.write(b14File, b14Seed)
	n := 0
	check := func(what, contract, constraints, refs string) (int, []string) {
		n++
		return r.b14Check(n, b14Brief(fmt.Sprintf("B20260101-%04d-case.loop-brief", 1000+n), what, contract, constraints, refs))
	}

	code, lines := check("A thing.", "It adds a config key.", "- Keep TestConfigListDefaults green.", "")
	if code != 0 {
		t.Fatalf("a warning exits %d, not 0: %v", code, lines)
	}
	b14HasLine(t, lines, mc+"TestConfigListJSONOrderAndNull, TestDefaultsEveryRow, TestWorkedExampleEnglishSession"+tail)

	_, lines = check("A thing.", "It adds a config key.", "- Keep "+all+" green.", "")
	b14NoWarning(t, lines, "must-change list:")

	_, lines = check("It adds a config key.", "Nothing else.", "- Keep TestConfigListDefaults green.", "")
	b14NoWarning(t, lines, "must-change list:")

	_, lines = check("A thing.", "A new eval case.", "", "")
	b14HasLine(t, lines, ef)

	_, lines = check("A thing.", "A new eval case.", "", "- `docs/guide/evals.md` — the eval case format")
	b14NoWarning(t, lines, "external format:")

	// No guidelines file: neither warning, and nothing else changes.
	if err := os.Remove(filepath.Join(r.dir, filepath.FromSlash(b14File))); err != nil {
		t.Fatal(err)
	}
	code, lines = check("A thing.", "It adds a config key and a new eval case.", "", "")
	if code != 0 {
		t.Fatalf("no guidelines file: exit %d: %v", code, lines)
	}
	b14NoWarning(t, lines, "must-change list:")
	b14NoWarning(t, lines, "external format:")
}

func TestWorkedExampleB14RunArguments(t *testing.T) {
	t.Parallel()
	const flag = " --append-system-prompt-file " + b14File + " "

	none := newRunRepo(t)
	if err := os.Remove(filepath.Join(none.dir, filepath.FromSlash(b14File))); err != nil {
		t.Fatal(err)
	}
	none.b14Commit("no guidelines")
	base := none.b14Sessions()
	for _, l := range base {
		if strings.Contains(l, "--append-system-prompt-file") {
			t.Fatalf("no guidelines file, yet a session got the flag: %s", l)
		}
	}

	empty := newRunRepo(t)
	empty.write(b14File, "")
	empty.b14Commit("empty guidelines")
	if got := empty.b14Sessions(); strings.Join(got, "\n") != strings.Join(base, "\n") {
		t.Fatalf("an empty guidelines file changed the arguments:\n%s\nagainst\n%s", strings.Join(got, "\n"), strings.Join(base, "\n"))
	}

	full := newRunRepo(t)
	full.write(b14File, "# Repository guidelines\n\n- A lesson. (D20261011-1455)\n")
	full.b14Commit("guidelines")
	got := full.b14Sessions()
	if len(got) != len(base) {
		t.Fatalf("%d sessions, want %d", len(got), len(base))
	}
	for i, l := range got {
		l += " "
		if !strings.Contains(l, flag) {
			t.Fatalf("session %d lacks the flag: %s", i+1, l)
		}
		if stripped := strings.TrimSuffix(strings.Replace(l, flag, " ", 1), " "); stripped != base[i] {
			t.Fatalf("session %d, flag removed:\n%s\nis not today's:\n%s", i+1, stripped, base[i])
		}
	}
}

func TestWorkedExampleB14Doctor(t *testing.T) {
	t.Parallel()
	const (
		over = "! guidelines 21 rules, over the budget of 20 — retire or merge some before adding more"
		noID = "! guidelines 2 rule(s) cite no defect or intervention: line 5, line 6"
	)
	r := newRunRepo(t)
	guidelines := func(want string) {
		t.Helper()
		res := r.vloop("doctor")
		b5Code(t, res, 0)
		if got := b14AfterLine(t, res.out, "gate edits"); got != want {
			t.Fatalf("after the gate edits line %q, want %q in:\n%s", got, want, res.out)
		}
	}
	extra := ""
	for i := 1; i <= 14; i++ {
		extra += fmt.Sprintf("- Rule %d. (D20261012-%04d)\n", i, 100+i)
	}
	extra += "- Rule 15. (I20261012-0115-a-decision)\n"

	guidelines("✓ guidelines 0 rule(s)") // the template init wrote
	r.write(b14File, b14Seed)
	guidelines("✓ guidelines 6 rule(s)")
	r.write(b14File, b14Seed+extra)
	guidelines(over)

	ids := regexp.MustCompile(` \(D[0-9-]*\)$`)
	src := strings.Split(b14Seed, "\n")
	for _, i := range []int{4, 5} { // lines 5 and 6
		src[i] = ids.ReplaceAllString(src[i], "")
	}
	stripped := strings.Join(src, "\n")
	r.write(b14File, stripped)
	guidelines(noID)

	if err := os.Remove(filepath.Join(r.dir, filepath.FromSlash(b14File))); err != nil {
		t.Fatal(err)
	}
	guidelines("- guidelines")
}

func TestWorkedExampleB14Instructions(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	words := func(n int) string { return strings.TrimSpace(strings.Repeat("word ", n)) + "\n" }
	r.write("plugin/skills/plan/SKILL.md", words(100))
	r.write("plugin/skills/work/SKILL.md", words(50))
	r.write(b14File, b14Seed)
	r.b14Commit("skills and guidelines")
	r.scripted(planJSON(t, planTask("T1", nil)), defaultScript)
	wantExit(t, r.vloop("run", runBrief), 0)

	// The end commit's own files, counted by git, not assumed.
	cw := b14Count(r.git("show", "HEAD:CLAUDE.md"))
	gw := b14Count(r.git("show", "HEAD:"+b14File))
	if cw == 0 || gw == 0 {
		t.Fatalf("fixture: CLAUDE.md has %d words and the guidelines %d at the end commit", cw, gw)
	}
	plan, work := 100+cw+gw, 50+cw+gw
	want := fmt.Sprintf("instructions plan %s · gate review n/a · work %s · review n/a words · guidelines 6 rules",
		b14Thousands(plan), b14Thousands(work))
	res := r.vloop("metrics", runID)
	wantExit(t, res, 0)
	if got := b14AfterLine(t, res.out, "interventions"); got != want {
		t.Fatalf("after the interventions line:\n%q\nwant\n%q", got, want)
	}

	js := r.vloop("metrics", runID, "--json")
	wantExit(t, js, 0)
	var doc struct {
		Instructions map[string]any `json:"instructions"`
	}
	if err := json.Unmarshal([]byte(js.out), &doc); err != nil {
		t.Fatalf("metrics --json: %v\n%s", err, js.out)
	}
	wantObj := map[string]any{"plan": float64(plan), "gate_review": nil, "work": float64(work), "review": nil, "guidelines_rules": float64(6)}
	gotB, _ := json.Marshal(doc.Instructions)
	wantB, _ := json.Marshal(wantObj)
	if string(gotB) != string(wantB) {
		t.Fatalf("instructions = %s, want %s", gotB, wantB)
	}
	file := filepath.Join(t.TempDir(), "metrics.json")
	if err := os.WriteFile(file, []byte(js.out), 0o644); err != nil {
		t.Fatal(err)
	}
	if v := r.vloop("schema", "validate", "metrics/v2", file); v.code != 0 {
		t.Fatalf("metrics --json is not metrics/v2: %+v", v)
	}
}

func TestWorkedExampleB14DefectReport(t *testing.T) {
	t.Parallel()
	const (
		rule = "brief-check: must-change list"
		hint = "vloop defect set <id> rule '<where>: <which>' and rule-outcome fired|missed|absent|none"
	)
	r := newRunRepo(t)
	r.write("docs/briefs/B20250101-0900-old.loop-brief.md",
		"---\nname: B20250101-0900-old.loop-brief\ndescription: An old brief\nkind: brief\nstatus: draft\n---\n# Old\n")
	var olds []string
	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("D2025010%d-0900-old-%d", i, i)
		olds = append(olds, id)
		r.write(".vloop/defects/"+id+".md", fmt.Sprintf("---\nid: %s\nbrief: B20250101-0900-old.loop-brief\norigin: brief\n"+
			"found-by: operator\nkind: spec-gap\nseverity: low\nstatus: open\nfixed-by: \"\"\ncase: \"\"\n"+
			"created: 2025-01-0%dT09:00:00Z\n---\nold defect %d\n", id, i, i))
	}
	r.b14Commit("old defects")
	r.scripted(planJSON(t, planTask("T1", nil)), defaultScript)
	wantExit(t, r.vloop("run", runBrief), 0)

	// A brief with no defects: none, and a dry run writes nothing.
	dry := r.vloop("brief", "close", runBrief, "--dry-run", "--no-findings")
	wantExit(t, dry, 0)
	b14HasLine(t, b14Trim(dry.out), "defects of "+runID+": none")
	if s := r.git("status", "--porcelain"); s != "" {
		t.Fatalf("close --dry-run wrote: %s", s)
	}

	b15MergeBrief(r.t, r.dir, runBrief)
	add := r.vloop("defect", "add", "the must-change list missed a test", "--brief", runBriefName,
		"--found-by", "review", "--kind", "spec-gap", "--origin", "brief")
	wantExit(t, add, 0)
	file := strings.TrimSpace(strings.SplitN(add.out, "\n", 2)[0])
	id1 := strings.TrimSuffix(filepath.Base(file), ".md")
	record := func() string { return r.read(".vloop/defects/" + id1 + ".md") }
	set := func(code int, args ...string) {
		t.Helper()
		before := record()
		res := r.vloop(append([]string{"defect", "set"}, args...)...)
		if res.code != code {
			t.Fatalf("defect set %v: exit %d, want %d: %+v", args, res.code, code, res)
		}
		if code != 0 && record() != before {
			t.Fatalf("a refused defect set %v changed the record", args)
		}
	}
	set(0, id1, "rule", rule)
	set(0, id1, "rule-outcome", "missed")
	set(2, id1, "rule-outcome", "forgot")
	set(2, id1, "rule", "linter: x")
	for _, w := range []string{"guideline: D20261011-1455", "brief-check: x", "gate: x", "gate-review: unpassable", "review: x", "skill: x"} {
		set(0, id1, "rule", w)
	}
	for _, o := range []string{"fired", "missed", "absent", "none"} {
		set(0, id1, "rule-outcome", o)
	}
	set(0, id1, "rule", rule)
	set(0, id1, "rule-outcome", "missed")
	for _, id := range olds {
		set(0, id, "rule", rule)
	}
	r.b14Commit("triage")

	res := r.vloop("brief", "close", runBrief, "--finding", "a", "--finding", "b")
	wantExit(t, res, 0)
	lines := strings.Split(strings.TrimSuffix(res.out, "\n"), "\n")
	if last := lines[len(lines)-1]; !strings.HasPrefix(last, "squash-merge with the trailer: ") {
		t.Fatalf("the last line is %q", last)
	}
	head := -1
	for i, l := range lines {
		if l == "defects of "+runID {
			head = i
		}
	}
	if head < 0 {
		t.Fatalf("close prints no 'defects of %s':\n%s", runID, res.out)
	}
	var rows []string
	for i := head + 1; i < len(lines) && strings.HasPrefix(lines[i], " "); i++ {
		rows = append(rows, lines[i])
	}
	hintAt := -1
	for i := head + 1 + len(rows); i < len(lines)-1; i++ {
		if strings.TrimSpace(lines[i]) == "2 untriaged: "+hint {
			hintAt = i
		}
	}
	if len(rows) != 3 {
		t.Fatalf("the report has %d rows, want one per defect of the brief (3):\n%s", len(rows), res.out)
	}
	if hintAt < 0 {
		t.Fatalf("no '2 untriaged: %s' line before the squash-merge line:\n%s", hint, res.out)
	}
	cols := regexp.MustCompile(` {2,}`)
	short := strings.Join(strings.SplitN(id1, "-", 3)[:2], "-")
	seen := false
	for _, row := range rows {
		f := cols.Split(strings.TrimSpace(row), -1)
		if f[0] == id1 || f[0] == short {
			seen = true
			want := []string{"spec-gap", "review", rule, "missed",
				"recurs ×3 (D20250101-0900, D20250102-0900, D20250103-0900)"}
			if strings.Join(f[1:], "|") != strings.Join(want, "|") {
				t.Fatalf("the row for %s is %q, want kind, found-by, rule, outcome, recurs: %q", id1, f, want)
			}
			continue
		}
		if strings.Join(f[1:], "|") != "bug|operator|—|untriaged" {
			t.Fatalf("a finding's row is %q, want bug, operator, —, untriaged", f)
		}
	}
	if !seen {
		t.Fatalf("no row for %s in %q", id1, rows)
	}
	if h := r.vloop("defect", "set", "--help"); !strings.Contains(h.out+h.err, "rule-outcome") {
		t.Fatalf("defect set --help does not list rule-outcome: %+v", h)
	}
}

func b14Trim(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		out = append(out, strings.TrimSpace(l))
	}
	return out
}

// b14Root is this repository's top level, or skips the test.
func b14Root(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Skip("not inside a git work tree")
	}
	root := strings.TrimSpace(string(out))
	if _, err := os.Stat(filepath.Join(root, "docs", "briefs")); err != nil {
		t.Skip("no docs/briefs in the work tree")
	}
	return root
}

// b14Git runs git -C dir with no global configuration.
func b14Git(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(cleanEnv(t.TempDir()), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.Output()
	return string(out), err
}

// b14EndOf is the latest run commit of a run id in either layout, or "".
func b14EndOf(t *testing.T, root, run string) string {
	t.Helper()
	log, err := b14Git(t, root, "log", "--all", "--format=%ct %H %s")
	if err != nil {
		t.Skip("cannot read the history")
	}
	best, hash := "", ""
	for _, l := range strings.Split(log, "\n") {
		f := strings.SplitN(l, " ", 3)
		if len(f) < 3 {
			continue
		}
		if !strings.HasPrefix(f[2], "[vloop] run "+run+"/") && !strings.HasPrefix(f[2], "[loop] run "+run+"/") {
			continue
		}
		if len(f[0]) > len(best) || (len(f[0]) == len(best) && f[0] >= best) {
			best, hash = f[0], f[1]
		}
	}
	return hash
}

// b14Words is wc -w of a file at a commit; 0 when absent.
func b14Words(t *testing.T, root, commit, path string) int {
	t.Helper()
	out, err := b14Git(t, root, "show", commit+":"+path)
	if err != nil {
		return 0
	}
	return b14Count(out)
}

// b14Consumed lists the briefs of docs/briefs whose front matter says consumed.
func b14Consumed(t *testing.T, root string) []string {
	t.Helper()
	files, _ := filepath.Glob(filepath.Join(root, "docs", "briefs", "*.loop-brief.md"))
	var names []string
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(string(raw), "\n")
		if lines[0] != "---" {
			continue
		}
		for _, l := range lines[1:] {
			if l == "---" {
				break
			}
			if l == "status: consumed" {
				names = append(names, strings.TrimSuffix(filepath.Base(f), ".loop-brief.md"))
				break
			}
		}
	}
	return names
}

func (r *runRepo) b14MetricsAt(root string, args ...string) result {
	r.t.Helper()
	cmd := exec.Command(binPath, append([]string{"metrics"}, args...)...)
	cmd.Dir = root
	cmd.Env = r.env()
	var o, e bytes.Buffer
	cmd.Stdout, cmd.Stderr = &o, &e
	code := 0
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			r.t.Fatal(err)
		}
		code = ee.ExitCode()
	}
	return result{o.String(), e.String(), code}
}

func TestWorkedExampleB14RealDataBriefCheck(t *testing.T) {
	t.Parallel()
	root := b14Root(t)
	const (
		p    = "! must-change list: the contract mentions "
		tail = " (" + b14File + ")"
	)
	r := newRunRepo(t)
	r.write(b14File, b14Seed)
	for _, c := range []struct{ brief, want string }{
		{"B20261003-2049-gate-model", `"schemas/" but Constraints do not name TestWorkedExampleB2Commands, TestSchemasAreEmbedded, TestWorkedExampleB6RealData`},
		{"B20261004-1434-interventions-options", `"schemas/" but Constraints do not name TestSchemasAreEmbedded, TestWorkedExampleB6RealData`},
		{"B20261008-2144-b12-field-reliability", `"config key" but Constraints do not name TestConfigListDefaults, TestConfigListJSONOrderAndNull, TestDefaultsEveryRow`},
		{"B20261011-2100-b13-docs-onboarding", `"closing lines" but Constraints do not name TestInitPointsAtBothWaysToLoadThePlugin`},
	} {
		raw, err := os.ReadFile(filepath.Join(root, "docs", "briefs", c.brief+".loop-brief.md"))
		if err != nil {
			t.Fatal(err)
		}
		text := strings.Replace(string(raw), "\nstatus: consumed\n", "\nstatus: ready\n", 1)
		if text == string(raw) {
			t.Fatalf("%s: no 'status: consumed' to set to ready", c.brief)
		}
		rel := "docs/briefs/" + c.brief + ".loop-brief.md"
		r.write(rel, text)
		res := r.vloop("brief", "check", rel)
		b14HasLine(t, b14Trim(res.out+res.err), p+c.want+tail)
	}
}

func TestWorkedExampleB14RealDataInstructions(t *testing.T) {
	t.Parallel()
	root := b14Root(t)
	r := newRunRepo(t)
	for _, c := range []struct{ brief, skill string }{
		{"B20261011-2100-b13-docs-onboarding", "plugin/skills/plan/SKILL.md"},
		{"B20260929-1804-vloop-skeleton-config-briefs", ".claude/skills/loop-plan/SKILL.md"},
	} {
		end := b14EndOf(t, root, c.brief)
		if end == "" {
			t.Skipf("no run commit for %s in this clone's history", c.brief)
		}
		if _, err := b14Git(t, root, "cat-file", "-e", end+":"+c.skill); err != nil {
			t.Fatalf("no %s at %s", c.skill, end)
		}
		want := b14Thousands(b14Words(t, root, end, c.skill) + b14Words(t, root, end, "CLAUDE.md") + b14Words(t, root, end, b14File))
		res := r.b14MetricsAt(root, c.brief)
		wantExit(t, res, 0)
		got := b14AfterLine(t, res.out, "interventions")
		if !strings.HasPrefix(got, "instructions plan "+want+" · ") {
			t.Fatalf("%s: the instructions line does not give plan %s: %q", c.brief, want, got)
		}
		if strings.HasPrefix(c.brief, "B2026092") && !strings.Contains(got, " · gate review n/a · ") {
			t.Fatalf("%s: gate review is not n/a: %q", c.brief, got)
		}
	}
}

func TestWorkedExampleB14RealDataMetricsEveryBrief(t *testing.T) {
	t.Parallel()
	root := b14Root(t)
	r := newRunRepo(t)
	names := b14Consumed(t, root)
	if len(names) == 0 {
		t.Skip("no consumed briefs")
	}
	for _, b := range names {
		res := r.b14MetricsAt(root, b)
		if res.code != 0 {
			t.Fatalf("vloop metrics %s exits %d: %s%s", b, res.code, res.out, res.err)
		}
		if !strings.Contains(strings.Join(b5Lines(res.out), "\n"), "\ninstructions plan ") {
			t.Fatalf("vloop metrics %s has no instructions line:\n%s", b, res.out)
		}
	}
}
