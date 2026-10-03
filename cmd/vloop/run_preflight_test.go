package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const runBriefPath = "docs/briefs/" + runBriefName + ".md"

// refusedBeforePlanning asserts a refusal that ran no session and committed
// nothing, with its exact stderr line.
func refusedBeforePlanning(t *testing.T, r *runRepo, head string, res result, line string) {
	t.Helper()
	wantExit(t, res, 1)
	if !strings.Contains(res.err, line+"\n") && strings.TrimSpace(res.err) != line {
		t.Errorf("stderr = %q, want the line %q", res.err, line)
	}
	if r.sessions() != 0 {
		t.Error("a session ran")
	}
	if got := strings.TrimSpace(r.git("rev-parse", "HEAD")); got != head {
		t.Error("the refusal committed")
	}
	if r.branch() != "main" {
		t.Errorf("the refusal left main for %s", r.branch())
	}
}

func TestRunPlansNewestReadyBrief(t *testing.T) {
	r := newRunRepo(t)
	// vloop init leaves a newer draft beside the ready brief.
	r.write("docs/briefs/B20260102-0900-newer.loop-brief.md",
		strings.Replace(runBriefText(), "status: ready", "status: draft", 1))
	r.commitAll("a newer draft")
	r.scripted(planJSON(t, planTask("T1", nil)), defaultScript)
	res := r.vloop("run", "--plan-only")
	wantExit(t, res, 0)
	wantIn(t, "argv", strings.Join(r.argv(), "\n"), "-p /vloop:plan "+runBriefPath+" ")
}

func TestRunRefusesDraftBriefs(t *testing.T) {
	r := newRunRepo(t)
	r.write(runBriefPath, strings.Replace(runBriefText(), "status: ready", "status: draft", 1))
	r.commitAll("all drafts")
	head := strings.TrimSpace(r.git("rev-parse", "HEAD"))

	refusedBeforePlanning(t, r, head, r.vloop("run"),
		"vloop: no ready brief in docs/briefs/ — name one, or set status: ready on a checked brief")
	refusedBeforePlanning(t, r, head, r.vloop("run", runBriefPath),
		"vloop: "+runBriefPath+" is draft, not ready — set status: ready once it passes vloop brief check")
}

func TestRunRefusesUncheckedBrief(t *testing.T) {
	r := newRunRepo(t)
	r.write(runBriefPath, strings.Replace(runBriefText(), "## Worked example", "## Example", 1))
	r.commitAll("an unchecked brief")
	head := strings.TrimSpace(r.git("rev-parse", "HEAD"))
	res := r.vloop("run", runBriefPath)
	wantExit(t, res, 1)
	wantIn(t, "stderr", res.err, "vloop: "+runBriefPath+" fails vloop brief check (", "problem(s)) — run vloop brief check "+runBriefPath)
	refusedBeforePlanning(t, r, head, res, strings.TrimSpace(res.err))
}

func TestRunRefusesUnconsumedDependency(t *testing.T) {
	r := newRunRepo(t)
	const base = "B20251201-0900-base"
	r.write("docs/briefs/"+base+".loop-brief.md", strings.ReplaceAll(runBriefText(), runBriefName, base+".loop-brief"))
	r.write(runBriefPath, strings.Replace(runBriefText(), "depends-on: []", "depends-on: ["+base+".loop-brief]", 1))
	r.commitAll("a dependency")
	head := strings.TrimSpace(r.git("rev-parse", "HEAD"))
	refusedBeforePlanning(t, r, head, r.vloop("run", runBriefPath),
		"vloop: "+runBriefPath+" depends on "+base+".loop-brief, which is ready, not consumed")
}

func TestRunRefusesDirtyTree(t *testing.T) {
	r := newRunRepo(t)
	head := strings.TrimSpace(r.git("rev-parse", "HEAD"))
	r.write(".env", "API_KEY=x\n")
	refusedBeforePlanning(t, r, head, r.vloop("run", runBriefPath),
		"vloop: the tree is not clean — commit, ignore or remove these first: .env")
	if r.git("ls-files", ".env") != "" {
		t.Error(".env was committed")
	}

	// A modified tracked file counts, and so does a staged one; scratch does not.
	r.write("docs/x.md", "y\n")
	r.git("add", "docs/x.md")
	r.write(".vloop/tmp/scratch", "s\n")
	refusedBeforePlanning(t, r, head, r.vloop("run", runBriefPath),
		"vloop: the tree is not clean — commit, ignore or remove these first: .env, docs/x.md")

	// More than ten: the first ten and the count of the rest.
	for i := 0; i < 12; i++ {
		r.write("extra/f"+string(rune('a'+i))+".txt", "x\n")
	}
	res := r.vloop("run", runBriefPath)
	wantExit(t, res, 1)
	wantIn(t, "stderr", res.err, "and 4 more")
}

func TestRunResumeAllowsStateEdits(t *testing.T) {
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript)
	wantExit(t, r.vloop("run", "--max-iterations", "1", runBriefPath), 4)
	wantExit(t, r.vloop("task", "note", "T2", "operator note"), 0)
	if r.git("status", "--porcelain", "--", ".vloop/state") == "" {
		t.Fatal("vloop task note changed nothing under .vloop/state/")
	}

	r.write("stray.txt", "x\n")
	res := r.vloop("run")
	wantExit(t, res, 1)
	wantIn(t, "stderr", res.err, "vloop: the tree is not clean — commit, ignore or remove these first: stray.txt")
	if err := os.Remove(filepath.Join(r.dir, "stray.txt")); err != nil {
		t.Fatal(err)
	}

	wantExit(t, r.vloop("run"), 0)
	r.wantTask("T2", "done", 0)
}
