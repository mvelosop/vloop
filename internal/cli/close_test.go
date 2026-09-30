package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const closeBrief = "B20260101-0900-a"
const closeName = closeBrief + ".loop-brief"

// closeRepo builds a finished run on its work branch: a ready brief, a plan,
// two done tasks and a `complete` run commit. complete=false ends the run
// `stalled` with T2 pending. tail is appended to the brief.
func closeRepo(t *testing.T, complete bool, tail string) string {
	t.Helper()
	dir := t.TempDir()
	at := func(date string, args ...string) { gitIn(t, dir, date, args...) }
	commit := func(date, msg string) {
		at(date, "add", "-A")
		at(date, "commit", "-q", "--allow-empty", "-m", msg)
	}
	at("2026-01-01T08:00:00Z", "init", "-q", "-b", "main")
	at("2026-01-01T08:00:00Z", "config", "user.name", "t")
	at("2026-01-01T08:00:00Z", "config", "user.email", "t@example.com")
	at("2026-01-01T08:00:00Z", "config", "commit.gpgsign", "false")
	write(t, dir, "docs/briefs/"+closeName+".md",
		"---\nname: "+closeName+"\nstatus: ready\n---\n# Brief\n\n- **Status:** ready to plan\n\n## Shape\n\n2 to 3 tasks.\n"+tail)
	write(t, dir, ".vloop/config.toml", "[metrics]\nstacks = [\"go\"]\n")
	commit("2026-01-01T08:00:00Z", "init")
	at("2026-01-01T08:00:00Z", "checkout", "-q", "-b", closeBrief)
	plan := func(a, b, status string) {
		write(t, dir, ".loop/state/state.json", `{"run_id":"`+closeBrief+`","tasks":[{"id":"T1","status":"`+a+`"},{"id":"T2","status":"`+b+`"}]}`)
	}
	plan("pending", "pending", "running")
	commit("2026-01-01T09:00:00Z", "[loop] plan "+closeBrief)
	write(t, dir, "main.go", "package main\n\nfunc main() {}\n")
	plan("done", "pending", "running")
	commit("2026-01-01T09:02:00Z", "[loop] T1: done")
	folder := ".loop/state/runs/" + closeBrief + "/20260101-090000/"
	write(t, dir, folder+"loop.log", "[loop] planning from docs/briefs/"+closeName+".md using opus\n")
	write(t, dir, folder+"iterations.jsonl", `{"iteration":1,"task":"T1","outcome":"done"}`+"\n")
	if complete {
		write(t, dir, "util.go", "package main\n\nfunc u() {}\n")
		plan("done", "done", "complete")
		commit("2026-01-01T09:06:00Z", "[loop] T2: done")
		commit("2026-01-01T09:10:00Z", "[loop] run "+closeBrief+"/20260101-090000: complete")
	} else {
		commit("2026-01-01T09:10:00Z", "[loop] run "+closeBrief+"/20260101-090000: stalled")
	}
	return dir
}

func headOf(t *testing.T, dir string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return string(out)
}

func TestCloseRefusals(t *testing.T) {
	const msg = "vloop: say what you found before merge: --finding \"<summary>\" (repeatable) or --no-findings\n"
	cases := []struct {
		name  string
		setup func(t *testing.T) string
		args  []string
		code  int
		want  string
	}{
		{"neither flag", func(t *testing.T) string { return closeRepo(t, true, "") }, nil, 2, msg},
		{"both flags", func(t *testing.T) string { return closeRepo(t, true, "") }, []string{"--finding", "x", "--no-findings"}, 2, msg},
		{"no runs", func(t *testing.T) string {
			d := closeRepo(t, true, "")
			write(t, d, "docs/briefs/B20260102-0900-c.loop-brief.md", "---\nname: B20260102-0900-c.loop-brief\nstatus: ready\n---\n")
			gitIn(t, d, "2026-01-02T00:00:00Z", "add", "-A")
			gitIn(t, d, "2026-01-02T00:00:00Z", "commit", "-q", "-m", "c")
			return d
		}, []string{"B20260102-0900-c.loop-brief", "--no-findings"}, 1, "vloop: no runs for B20260102-0900-c.loop-brief\n"},
		{"already consumed", func(t *testing.T) string {
			d := closeRepo(t, true, "")
			write(t, d, "docs/briefs/"+closeName+".md", "---\nname: "+closeName+"\nstatus: consumed\n---\n")
			gitIn(t, d, "2026-01-02T00:00:00Z", "add", "-A")
			gitIn(t, d, "2026-01-02T00:00:00Z", "commit", "-q", "-m", "c")
			return d
		}, nil, 1, "vloop: " + closeName + " is already consumed\n"},
		{"already abandoned", func(t *testing.T) string {
			d := closeRepo(t, true, "")
			write(t, d, "docs/briefs/"+closeName+".md", "---\nname: "+closeName+"\nstatus: abandoned\n---\n")
			gitIn(t, d, "2026-01-02T00:00:00Z", "add", "-A")
			gitIn(t, d, "2026-01-02T00:00:00Z", "commit", "-q", "-m", "c")
			return d
		}, nil, 1, "vloop: " + closeName + " is already abandoned\n"},
		{"on main", func(t *testing.T) string {
			d := closeRepo(t, true, "")
			gitIn(t, d, "2026-01-02T00:00:00Z", "checkout", "-q", "main")
			gitIn(t, d, "2026-01-02T00:00:00Z", "merge", "-q", "--ff-only", closeBrief)
			return d
		}, nil, 1, "vloop: close on the brief's work branch, not main\n"},
		{"untracked file", func(t *testing.T) string {
			d := closeRepo(t, true, "")
			write(t, d, "stray/x.txt", "x")
			return d
		}, nil, 1, "vloop: commit or stash your changes first — close makes one commit of its own\n"},
		{"modified file", func(t *testing.T) string {
			d := closeRepo(t, true, "")
			write(t, d, "main.go", "changed\n")
			return d
		}, nil, 1, "vloop: commit or stash your changes first — close makes one commit of its own\n"},
		{"stalled run", func(t *testing.T) string { return closeRepo(t, false, "") }, nil, 1,
			"vloop: the plan is not complete (1/2 done) — finish it, or pass --abandon \"<reason>\"\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := c.setup(t)
			head := headOf(t, dir)
			status := gitOutput(t, dir, "status", "--porcelain", "--untracked-files=all")
			args := c.args
			if len(args) == 0 || strings.HasPrefix(args[0], "--") {
				args = append([]string{closeName}, args...)
			}
			if len(c.args) == 0 {
				args = append(args, "--no-findings")
			}
			if c.name == "neither flag" {
				args = []string{closeName}
			}
			code, out, errs := runCLI(t, dir, append([]string{"brief", "close"}, args...)...)
			if code != c.code || out != "" || errs != c.want {
				t.Fatalf("exit %d, want %d\nstdout %q\nstderr %q\nwant   %q", code, c.code, out, errs, c.want)
			}
			if headOf(t, dir) != head || gitOutput(t, dir, "status", "--porcelain", "--untracked-files=all") != status {
				t.Error("a refusal wrote, staged or committed something")
			}
			if _, err := os.Stat(filepath.Join(dir, ".vloop/defects")); err == nil {
				t.Error("a refusal recorded a defect")
			}
		})
	}
}

func TestCloseCommitsFindingsSnapshotAndBrief(t *testing.T) {
	dir := closeRepo(t, true, "")
	before := headOf(t, dir)
	code, out, errs := runCLI(t, dir, "brief", "close", closeName, "--finding", "README omits the flag", "--finding", "second one")
	if code != 0 || errs != "" {
		t.Fatalf("exit %d: %s", code, errs)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	var rec []string
	for _, l := range lines {
		if strings.HasPrefix(l, "recorded ") {
			rec = append(rec, strings.TrimPrefix(l, "recorded "))
		}
	}
	if len(rec) != 2 || !strings.HasSuffix(rec[0], "-readme-omits-the-flag.md") || !strings.HasSuffix(rec[1], "-second-one.md") {
		t.Fatalf("recorded lines: %q", out)
	}
	n := len(lines)
	sha7 := gitOutput(t, dir, "rev-parse", "--short=7", "HEAD")
	sha7 = strings.TrimSpace(sha7)
	for i, w := range []string{
		"wrote .vloop/state/metrics/" + closeBrief + ".json",
		"updated docs/briefs/" + closeName + ".md",
		"committed " + sha7 + " [vloop] close " + closeBrief,
		"squash-merge with the trailer: Vloop-Brief: " + closeName,
	} {
		if lines[n-4+i] != w {
			t.Errorf("line %d = %q, want %q", n-4+i, lines[n-4+i], w)
		}
	}
	if !strings.HasPrefix(collapse(out), closeBrief+" consumed · not merged\n") {
		t.Errorf("summary: %q", out)
	}

	if p := strings.TrimSpace(gitOutput(t, dir, "show", "-s", "--format=%P", "HEAD")); p != before {
		t.Errorf("parents %q, want %q", p, before)
	}
	if m := gitOutput(t, dir, "show", "-s", "--format=%B", "HEAD"); m != "[vloop] close "+closeBrief+"\n\nVloop-Brief: "+closeName+"\n\n" {
		t.Errorf("message %q", m)
	}
	got := strings.Fields(gitOutput(t, dir, "diff-tree", "--no-commit-id", "--name-only", "-r", "HEAD"))
	want := append(append([]string{}, rec...), ".vloop/state/metrics/"+closeBrief+".json", "docs/briefs/"+closeName+".md")
	if strings.Join(sortedCopy(got), " ") != strings.Join(sortedCopy(want), " ") {
		t.Errorf("commit files %v, want %v", got, want)
	}
	if s := gitOutput(t, dir, "status", "--porcelain", "--untracked-files=all"); s != "" {
		t.Errorf("tree dirty: %s", s)
	}
	if b := strings.TrimSpace(gitOutput(t, dir, "branch", "--format=%(refname:short)")); b != closeBrief+"\nmain" {
		t.Errorf("branches %q", b)
	}

	// The snapshot equals metrics --json run right after the close.
	snap, err := os.ReadFile(filepath.Join(dir, ".vloop/state/metrics/"+closeBrief+".json"))
	if err != nil {
		t.Fatal(err)
	}
	_, live, _ := runCLI(t, dir, "metrics", closeName, "--json")
	if strings.TrimSpace(string(snap)) != "" && compactJSON(t, snap) != compactJSON(t, []byte(live)) {
		t.Errorf("snapshot differs from metrics --json:\n%s\n%s", snap, live)
	}
	brief, _ := os.ReadFile(filepath.Join(dir, "docs/briefs/"+closeName+".md"))
	for _, w := range []string{
		"status: consumed\n",
		"- **Status:** consumed — closed ",
		" as run " + closeBrief + ". **Do not re-plan from this brief.**\n",
		"## Run record\n",
		"- " + strings.TrimSuffix(filepath.Base(rec[0]), ".md") + " — bug, work, found by operator: README omits the flag\n",
	} {
		if !strings.Contains(string(brief), w) {
			t.Errorf("brief lacks %q:\n%s", w, brief)
		}
	}
}

func TestCloseJSON(t *testing.T) {
	dir := closeRepo(t, true, "")
	code, out, errs := runCLI(t, dir, "brief", "close", closeName, "--no-findings", "--json")
	if code != 0 || errs != "" {
		t.Fatalf("exit %d: %s", code, errs)
	}
	for _, w := range []string{
		`"brief":"` + closeName + `"`, `"status":"consumed"`, `"commit":"` + headOf(t, dir) + `"`,
		`"files":[".vloop/state/metrics/` + closeBrief + `.json","docs/briefs/` + closeName + `.md"]`,
		`"trailer":"Vloop-Brief: ` + closeName + `"`, `"schema":"metrics/v1"`,
	} {
		if !strings.Contains(out, w) {
			t.Errorf("json lacks %s:\n%s", w, out)
		}
	}
}

func TestCloseRerenderKeepsTextAfterMarkers(t *testing.T) {
	after := "<!-- vloop:run-record:end -->\n\nHand-written note  \nkept byte for byte, trailing spaces and all."
	dir := closeRepo(t, true, "\n<!-- vloop:run-record:begin -->\nstale record\n"+after)
	if code, _, errs := runCLI(t, dir, "brief", "close", closeName, "--no-findings"); code != 0 {
		t.Fatalf("exit %d: %s", code, errs)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "docs/briefs/"+closeName+".md"))
	if !strings.HasSuffix(string(b), after) {
		t.Errorf("text after the end marker changed:\n%q", b)
	}
	if strings.Contains(string(b), "stale record") || strings.Count(string(b), "vloop:run-record:begin") != 1 {
		t.Errorf("run record not replaced in place:\n%s", b)
	}
	if !strings.Contains(string(b), "# Brief\n\n- **Status:** consumed") {
		t.Errorf("text before the markers changed unexpectedly:\n%s", b)
	}
}

func sortedCopy(s []string) []string {
	c := append([]string(nil), s...)
	for i := range c {
		for j := i + 1; j < len(c); j++ {
			if c[j] < c[i] {
				c[i], c[j] = c[j], c[i]
			}
		}
	}
	return c
}

func compactJSON(t *testing.T, b []byte) string {
	t.Helper()
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	out, _ := json.Marshal(v)
	return string(out)
}
