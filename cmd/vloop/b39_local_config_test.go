package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const localPath = ".vloop/config.local.toml"

const (
	localNotIgnored = "✗ config.local " + localPath + " is not git-ignored — the next git add -A would commit it: add it to .gitignore"
	localTracked    = "✗ config.local " + localPath + " is tracked by git — it is shared with everyone who clones: run git rm --cached " + localPath + " and add it to .gitignore"
)

// doctorLocalRows is the config.local rows of vloop doctor.
func doctorLocalRows(t *testing.T, r *runRepo) []string {
	t.Helper()
	res := r.vloop("doctor")
	var rows []string
	for _, l := range strings.Split(res.out, "\n") {
		if strings.Contains(l, " config.local") {
			rows = append(rows, l)
		}
	}
	return rows
}

func TestInitKeepsLocalConfigIgnored(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	gi := r.read(".gitignore")
	for _, l := range []string{".vloop/tmp/", localPath} {
		if strings.Count("\n"+gi, "\n"+l+"\n") != 1 {
			t.Errorf(".gitignore does not hold %s once:\n%s", l, gi)
		}
	}
}

func TestDoctorConfigLocalRow(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	if rows := doctorLocalRows(t, r); len(rows) != 0 {
		t.Errorf("no local file, rows %q", rows)
	}
	r.write(localPath, "")
	if rows := doctorLocalRows(t, r); len(rows) != 1 || rows[0] != "✓ config.local" {
		t.Errorf("empty ignored local file, rows %q", rows)
	}
	r.write(localPath, "[agent]\nwork = \"claude\"\n\n[model]\nwork = \"sonnet\"\n")
	rows := doctorLocalRows(t, r)
	if len(rows) != 1 || rows[0] != "! config.local overrides model.work, agent.work — on this machine only" {
		t.Errorf("ignored local file, rows %q", rows)
	}
	// a pattern that covers the file satisfies the test
	r.write(".gitignore", ".vloop/tmp/\n.vloop/*.local.toml\n")
	if rows := doctorLocalRows(t, r); len(rows) != 1 || !strings.HasPrefix(rows[0], "! config.local overrides") {
		t.Errorf("pattern-ignored local file, rows %q", rows)
	}
	// a local file that cannot be read: no row, the config row is the problem
	r.write(localPath, "agent.work = \"nope\"\n")
	if rows := doctorLocalRows(t, r); len(rows) != 0 {
		t.Errorf("unreadable local file, rows %q", rows)
	}
}

func TestDoctorAndRunRefuseLocalConfigNotIgnored(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.write(".gitignore", ".vloop/tmp/\n")
	r.write(localPath, "[agent]\nwork = \"claude\"\n")
	r.commitAll("setup")
	// commitAll took the file in; plant the failure as untracked
	r.git("rm", "--cached", "-q", localPath)
	r.git("commit", "-q", "-m", "untrack")
	if rows := doctorLocalRows(t, r); len(rows) != 1 || rows[0] != localNotIgnored {
		t.Fatalf("doctor rows %q", rows)
	}
	head := r.head()
	res := r.vloop("run", runBriefPath)
	refusedBeforePlanning(t, r, head, res, "vloop: preflight failed — 1 problem(s), nothing has run")
	wantIn(t, "stderr", res.err, "  "+localNotIgnored+"\n")
}

func TestDoctorAndRunRefuseLocalConfigTracked(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.write(localPath, "[agent]\nwork = \"claude\"\n")
	r.git("add", "-f", localPath)
	r.commitAll("share the local file")
	if rows := doctorLocalRows(t, r); len(rows) != 1 || rows[0] != localTracked {
		t.Fatalf("doctor rows %q", rows)
	}
	head := r.head()
	res := r.vloop("run", runBriefPath)
	refusedBeforePlanning(t, r, head, res, "vloop: preflight failed — 1 problem(s), nothing has run")
	wantIn(t, "stderr", res.err, "  "+localTracked+"\n")
}

func TestRunWarnsOfLocalConfigOverrides(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.write(localPath, "[agent]\nwork = \"claude\"\n")
	res := r.vloop("run", runBriefPath)
	if strings.Contains(res.err, "preflight failed") {
		t.Errorf("preflight failed: %q", res.err)
	}
	wantIn(t, "stderr", res.err, "  ! config.local overrides agent.work — on this machine only\n")
}

// noClaudeRepo is a harness repository on a PATH holding the stub copilot and
// the few tools the stub and git need, and no claude; every agent.* key is
// copilot with a Copilot model id and run.verify is false.
func noClaudeRepo(t *testing.T, extra string) (*runRepo, []string) {
	t.Helper()
	r := newRunRepo(t)
	if err := os.Remove(filepath.Join(r.stub, "claude")); err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	for _, tool := range []string{"git", "sh", "dirname", "env", "sort", "sed", "cat", "true"} {
		p, err := exec.LookPath(tool)
		if err != nil {
			t.Skipf("%s not found: %v", tool, err)
		}
		if err := os.Symlink(p, filepath.Join(bin, tool)); err != nil {
			t.Fatal(err)
		}
	}
	var cfg strings.Builder
	over := map[string]bool{}
	for _, l := range strings.Split(extra, "\n") {
		if k, _, ok := strings.Cut(l, " = "); ok {
			over[k] = true
			cfg.WriteString(l + "\n")
		}
	}
	for _, p := range []string{"plan", "work", "review", "second-review", "gate-review", "verify", "review-brief"} {
		if !over["agent."+p] {
			cfg.WriteString("agent." + p + " = \"copilot\"\n")
		}
		cfg.WriteString("model." + p + " = \"gpt-6.1-terra\"\n")
	}
	if !over["run.verify"] {
		cfg.WriteString("run.verify = \"false\"\n")
	}
	r.write(".vloop/config.toml", cfg.String()+"\n"+runCheckConfig)
	r.commitAll("config")
	env := append(cleanEnv(r.home), "PATH="+r.stub+string(os.PathListSeparator)+bin,
		"GH_TOKEN=test-token", "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0")
	return r, env
}

func TestClaudeRowsNotInUseWithoutClaude(t *testing.T) {
	t.Parallel()
	r, env := noClaudeRepo(t, "")
	res := r.exec(binPath, env, "doctor")
	lines := strings.Split(res.out, "\n")
	for _, n := range []string{"claude", "trust", "plugin", "gate edits"} {
		want := "- " + n + " no session in use runs on claude"
		found := false
		for _, l := range lines {
			found = found || l == want
		}
		if !found {
			t.Errorf("doctor has no line %q:\n%s", want, res.out)
		}
	}
	if strings.Contains(res.out, "✗ claude") {
		t.Errorf("doctor refuses on claude:\n%s", res.out)
	}
	run := r.exec(binPath, env, "run", runBriefPath)
	if strings.Contains(run.err, "preflight failed") {
		t.Errorf("the run failed its preflight:\n%s", run.err)
	}
	if b, _ := os.ReadFile(filepath.Join(r.stub, "copilot.log")); !strings.Contains(string(b), "arg /vloop:plan") {
		t.Errorf("no copilot plan session ran:\n%s", run.err)
	}
}

func TestClaudeRowsInUseWhenReviewIsClaude(t *testing.T) {
	t.Parallel()
	r, env := noClaudeRepo(t, `agent.review = "claude"`)
	res := r.exec(binPath, env, "doctor")
	if !strings.Contains(res.out, "✗ claude claude is not on PATH\n") {
		t.Errorf("doctor lacks the missing-claude row:\n%s", res.out)
	}
	if strings.Contains(res.out, "no session in use runs on claude") {
		t.Errorf("a Claude row is n/a though review is on claude:\n%s", res.out)
	}
	head := r.head()
	run := r.exec(binPath, env, "run", runBriefPath)
	if !strings.Contains(run.err, "  ✗ claude claude is not on PATH\n") || !strings.Contains(run.err, "vloop: preflight failed") {
		t.Errorf("the run's preflight did not refuse on claude:\n%s", run.err)
	}
	if r.sessions() != 0 || strings.TrimSpace(r.git("rev-parse", "HEAD")) != head {
		t.Error("the refused run ran a session or committed")
	}
}
