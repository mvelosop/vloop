package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// The code-health and escapes brief's worked example (B15), played line for
// line against the built binary and the suite's stub claude: temporary
// repositories with a temporary HOME, addressed with git -C. "Merged" means the
// brief's consumed commit is on main (M-5).

// b15Script is the default phase script with a work session that writes code
// and test lines for T1 and costs 0.25 a session; failOnce makes T1's first
// work attempt leave its gate failing.
func b15Script(failOnce bool) string {
	touch := `touch "$TASK.out"`
	if failOnce {
		touch = `if [ "$TASK" != T1 ] || [ "$ATTEMPT" != 1 ]; then touch "$TASK.out"; fi`
	}
	return `STUB_COST=0.25
case "$PHASE" in
  plan) mkdir -p .vloop/state; cp "$dir/plan.json" .vloop/state/state.json;;
  work) ` + touch + `; mkdir -p .vloop/tmp
    if [ "$TASK" = T1 ] && [ -z "$B15_NOCODE" ]; then
      mkdir -p src; i=0; : > src/a.go; while [ $i -lt 200 ]; do echo "// $i" >> src/a.go; i=$((i+1)); done
      i=0; : > src/a_test.go; while [ $i -lt 300 ]; do echo "// $i" >> src/a_test.go; i=$((i+1)); done
    fi
    printf '{"schema":"proposal/v1","task":"%s","outcome":"done","summary":"made %s","files":[],"verified":"ok","notes":"none"}\n' "$TASK" "$TASK" > .vloop/tmp/proposal.json;;
  review) mkdir -p .vloop/tmp
    printf '{"schema":"verdict/v1","task":"%s","verdict":"PASS","criteria":[],"findings":[],"notes":"none"}\n' "$TASK" > .vloop/tmp/verdict.json;;
esac
`
}

func b15Want(t *testing.T, res result, code int, stderr string) {
	t.Helper()
	wantExit(t, res, code)
	if stderr != "" && strings.TrimSpace(res.err) != stderr {
		t.Fatalf("stderr:\n%s\nwant:\n%s", strings.TrimSpace(res.err), stderr)
	}
}

func b15Defect(id, brief, foundBy, created string) string {
	return fmt.Sprintf("---\nid: %s\nbrief: %s\norigin: work\nfound-by: %s\nkind: bug\nseverity: low\nstatus: open\nfixed-by: \"\"\ncase: \"\"\ncreated: %s\n---\nfixture defect %s\n",
		id, brief, foundBy, created, id)
}

func b15Squeeze(s string) string { return strings.Join(strings.Fields(s), " ") }

func b15Line(out, prefix string) string {
	for _, l := range strings.Split(out, "\n") {
		if s := b15Squeeze(l); strings.HasPrefix(s, prefix) {
			return s
		}
	}
	return ""
}

func TestWorkedExampleB15GitConfigIgnore(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	for _, p := range []string{"Core.*", "core.*"} {
		b15Want(t, r.vloop("config", "set", "run.git-config-ignore", p), 2,
			"vloop: run.git-config-ignore: "+p+" would ignore core.* — the guard must see it")
	}
	wantExit(t, r.vloop("config", "set", "run.git-config-ignore", "Branch.Main.Foo-Bar"), 0)
	if got := strings.TrimSpace(r.vloop("config", "get", "run.git-config-ignore").out); !strings.Contains(got, "Branch.Main.Foo-Bar") {
		t.Fatalf("config get: %q", got)
	}
}

func TestWorkedExampleB15DefectsInterventionsMetrics(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	unmerged := "B20260101-0700-unmerged.loop-brief"
	r.write("docs/briefs/"+unmerged+".md", strings.ReplaceAll(runBriefText(), runBriefName, unmerged))
	wantExit(t, r.vloop("config", "set", "metrics.stacks", "go"), 0)
	r.commitAll("unmerged brief and stacks")
	r.scripted(planJSON(t, planTask("T1", nil), planTask("T2", map[string]any{"depends_on": []string{"T1"}})), b15Script(false))
	wantExit(t, r.vloop("run", runBrief), 0)
	wantExit(t, r.vloop("brief", "close", runBrief, "--no-findings"), 0)
	branch := strings.TrimSpace(r.git("symbolic-ref", "--short", "HEAD"))
	if branch == "main" {
		t.Fatal("the run left main checked out")
	}
	r.git("checkout", "-q", "main")
	r.git("merge", "-q", "--no-ff", "-m", "merge "+branch, branch)

	// defect add on an unmerged brief names the intervention to use.
	res := r.vloop("defect", "add", "x", "--brief", unmerged, "--found-by", "user")
	b15Want(t, res, 1, "vloop: "+unmerged+" is not merged: record what you found before merge as an intervention — "+
		"vloop intervention add \"<summary>\" --brief "+unmerged+" --kind verification-finding --origin <brief|plan|work|env> --phase verify --automatable partly --by operator")
	if m, _ := filepath.Glob(filepath.Join(r.dir, ".vloop", "defects", "*")); len(m) != 0 {
		t.Fatalf("defect add on an unmerged brief wrote %v", m)
	}

	// Eleven records of another brief hold the next seconds: the same second.
	// The pause puts the defects after the merge, which is what escaped means.
	time.Sleep(2 * time.Second)
	now := time.Now().Unix()
	stamp := func(k int64) string { return time.Unix(now+k, 0).Format("20060102-150405") }
	for k := int64(0); k <= 10; k++ {
		id := "D" + stamp(k) + "-other"
		r.write(".vloop/defects/"+id+".md", b15Defect(id, unmerged, "user", "2026-01-01T00:00:00Z"))
	}
	a1 := r.vloop("defect", "add", "x", "--brief", runBriefName, "--found-by", "operator")
	wantExit(t, a1, 0)
	if got, want := strings.TrimSpace(a1.out), ".vloop/defects/D"+stamp(11)+"-x.md"; got != want {
		t.Fatalf("defect add printed %q, want %q", got, want)
	}
	a2 := r.vloop("defect", "add", "x", "--brief", runBriefName, "--found-by", "operator")
	wantExit(t, a2, 0)
	if got, want := strings.TrimSpace(a2.out), ".vloop/defects/D"+stamp(12)+"-x.md"; got != want {
		t.Fatalf("the same again printed %q, want %q", got, want)
	}
	m := r.vloop("metrics", runBrief)
	wantExit(t, m, 0)
	if got, want := b15Line(m.out, "defects "), "defects in-loop 0 · operator 0 · escaped 2 · removal efficiency 0%"; got != want {
		t.Fatalf("metrics defects line %q, want %q", got, want)
	}

	// intervention add --origin.
	iv := func(extra ...string) result {
		return r.vloop(append([]string{"intervention", "add", "y", "--brief", runBriefName, "--phase", "verify", "--automatable", "no", "--by", "operator"}, extra...)...)
	}
	b15Want(t, iv("--kind", "verification-finding"), 2, "vloop: --kind verification-finding needs --origin brief, plan, work or env")
	b15Want(t, iv("--kind", "repair", "--origin", "work"), 2, "vloop: --origin is only for --kind verification-finding")
	i3 := iv("--kind", "verification-finding", "--origin", "plan")
	wantExit(t, i3, 0)
	p := strings.TrimSpace(i3.out)
	if !regexp.MustCompile(`^\.vloop/interventions/I\d{8}-\d{6}-y\.md$`).MatchString(p) {
		t.Fatalf("intervention add printed %q", p)
	}
	if !strings.Contains("\n"+r.read(p), "\norigin: plan\n") {
		t.Fatalf("%s has no origin: plan", p)
	}
	b15Want(t, r.vloop("defect", "set", strings.TrimSuffix(filepath.Base(strings.TrimSpace(a1.out)), ".md"), "rule", "gate: x"), 2,
		`vloop: cannot set "rule": want one of status, fixed-by, case, severity, origin, kind, task`)

	// Per task: T1 code +200, test +300, $0.50; T2 no code.
	bt := r.vloop("metrics", runBrief, "--by", "task")
	wantExit(t, bt, 0)
	rows := strings.Split(strings.TrimSpace(bt.out), "\n")
	head := strings.Fields(rows[0])
	col := func(name string) int {
		for i, h := range head {
			if h == name {
				return i
			}
		}
		t.Fatalf("no column %s in %q", name, rows[0])
		return -1
	}
	o := col("other+")
	if col("test:code") != o+1 || col("$/kloc") != o+2 || col("agent") != o+3 {
		t.Fatalf("per-task header: %s", rows[0])
	}
	got := map[string]string{}
	for _, row := range rows[1:] {
		f := strings.Fields(row)
		if len(f) > o+2 {
			got[f[0]] = f[o+1] + " " + f[o+2]
		}
	}
	if got["T1"] != "1.50 2.50" || got["T2"] != "- -" {
		t.Fatalf("per-task test:code and $/kloc: %v\n%s", got, bt.out)
	}
	sum := r.vloop("metrics", runBrief)
	if !regexp.MustCompile(`(?m)^ *cost .* · \$[0-9,]+\.[0-9]{2}/kloc$`).MatchString(sum.out) {
		t.Fatalf("the cost line does not end ' · $<c>/kloc':\n%s", sum.out)
	}
	j := r.vloop("metrics", runBrief, "--by", "task", "--json")
	wantExit(t, j, 0)
	var doc struct {
		ByTask []map[string]any `json:"by_task"`
	}
	if err := json.Unmarshal([]byte(j.out), &doc); err != nil {
		t.Fatal(err)
	}
	for _, tk := range doc.ByTask {
		switch tk["id"] {
		case "T1":
			if tk["test_code_ratio"] != 1.5 || tk["per_1000_code_lines_usd"] != 2.5 {
				t.Fatalf("--json T1: %v", tk)
			}
		case "T2":
			v, has := tk["test_code_ratio"]
			if !has || v != nil || tk["per_1000_code_lines_usd"] != nil {
				t.Fatalf("--json T2: %v", tk)
			}
		}
	}
}

func TestWorkedExampleB15CloseWithFinding(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(planJSON(t, planTask("T1", nil)), b15Script(true))
	wantExit(t, r.vloop("run", runBrief), 0)
	b15Want(t, r.vloop("brief", "close", runBrief, "--finding", "spec gap: z"), 2,
		`vloop: a finding starts with whose defect it is: --finding "<brief|plan|work|env>: <summary>"`)
	res := r.vloop("brief", "close", runBrief, "--finding", "brief: z")
	wantExit(t, res, 0)
	m := regexp.MustCompile(`(?m)^recorded (\.vloop/interventions/(I\d{8}-\d{6}-z)\.md)$`).FindStringSubmatch(res.out)
	if m == nil {
		t.Fatalf("close does not print recorded .vloop/interventions/I<stamp>-z.md:\n%s", res.out)
	}
	if strings.Contains(res.out, "defects of") {
		t.Fatalf("close still prints 'defects of':\n%s", res.out)
	}
	lines := strings.Split(strings.TrimSpace(res.out), "\n")
	if last := lines[len(lines)-1]; !strings.HasPrefix(last, "squash-merge with the trailer: ") {
		t.Fatalf("close's last line: %q", last)
	}
	if got, want := b15Line(res.out, "defects "), "defects in-loop 1 · operator 1 · escaped 0 · removal efficiency 100%"; got != want {
		t.Fatalf("close's defects line %q, want %q", got, want)
	}
	if want := "- " + m[2] + " — verification-finding, brief, found by operator: z"; !strings.Contains("\n"+r.read(runBrief), "\n"+want+"\n") {
		t.Fatalf("the run record has no %q", want)
	}
}

// Every existing record of this repository is still listed and shown.
func TestWorkedExampleB15RealDataRecords(t *testing.T) {
	t.Parallel()
	root := b14Root(t)
	count := func(dir, prefix string) int {
		m, _ := filepath.Glob(filepath.Join(root, ".vloop", dir, prefix+"*.md"))
		return len(m)
	}
	run := func(args ...string) []byte {
		c := exec.Command(binPath, args...)
		c.Dir = root
		c.Env = cleanEnv(t.TempDir())
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("vloop %v: %v\n%s", args, err, out)
		}
		return out
	}
	for _, k := range []struct{ list, dir, prefix string }{{"defect", "defects", "D"}, {"intervention", "interventions", "I"}} {
		var recs []json.RawMessage
		if err := json.Unmarshal(run(k.list, "list", "--json"), &recs); err != nil {
			t.Fatal(err)
		}
		if want := count(k.dir, k.prefix); len(recs) != want {
			t.Errorf("%s list lists %d of %d records", k.list, len(recs), want)
		}
	}
	files, _ := filepath.Glob(filepath.Join(root, ".vloop", "interventions", "I*.md"))
	for _, f := range files {
		if _, err := os.Stat(f); err == nil {
			run("intervention", "show", strings.TrimSuffix(filepath.Base(f), ".md"))
		}
	}
}
