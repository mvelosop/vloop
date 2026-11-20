package main

import (
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
