package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCloseAbandonStalledRun(t *testing.T) {
	dir := closeRepo(t, false, "")
	before := headOf(t, dir)
	code, out, errs := runCLI(t, dir, "brief", "close", closeName, "--abandon", "superseded", "--no-findings")
	if code != 0 || errs != "" {
		t.Fatalf("exit %d: %s", code, errs)
	}
	if !strings.HasPrefix(collapse(out), closeBrief+" abandoned · not merged\n") {
		t.Errorf("summary: %q", out)
	}
	brief, err := os.ReadFile(filepath.Join(dir, "docs/briefs/"+closeName+".md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range []string{"\nstatus: abandoned\n", "\n- **Status:** abandoned — superseded\n", "<!-- vloop:run-record:begin -->"} {
		if !strings.Contains(string(brief), w) {
			t.Errorf("brief lacks %q:\n%s", w, brief)
		}
	}
	if p := strings.TrimSpace(gitOutput(t, dir, "show", "-s", "--format=%P", "HEAD")); p != before {
		t.Errorf("parents %q, want %q", p, before)
	}
	if m := gitOutput(t, dir, "show", "-s", "--format=%B", "HEAD"); m != "[vloop] close "+closeBrief+"\n\nVloop-Brief: "+closeName+"\n\n" {
		t.Errorf("message %q", m)
	}
	snap, err := os.ReadFile(filepath.Join(dir, ".vloop/state/metrics/"+closeBrief+".json"))
	if err != nil || !strings.Contains(string(snap), `"status": "abandoned"`) {
		t.Errorf("snapshot: %v %s", err, snap)
	}
	code, _, errs = runCLI(t, dir, "brief", "close", closeName, "--no-findings")
	if code != 1 || errs != "vloop: "+closeName+" is already abandoned\n" {
		t.Errorf("second close: exit %d %q", code, errs)
	}
}

func TestCloseAbandonAcceptedOnCompleteRunAndKeepsRefusals(t *testing.T) {
	dir := closeRepo(t, true, "")
	if code, _, errs := runCLI(t, dir, "brief", "close", closeName, "--abandon", "x", "--finding", "f"); code != 0 {
		t.Fatalf("exit %d: %s", code, errs)
	}

	dir = closeRepo(t, false, "")
	head := headOf(t, dir)
	code, out, errs := runCLI(t, dir, "brief", "close", closeName, "--abandon", "x")
	if code != 2 || out != "" || errs != "vloop: "+closeFindingsMsg+"\n" {
		t.Errorf("without findings flag: exit %d %q %q", code, out, errs)
	}
	write(t, dir, "stray.txt", "x")
	code, out, errs = runCLI(t, dir, "brief", "close", closeName, "--abandon", "x", "--no-findings")
	if code != 1 || out != "" || !strings.Contains(errs, "commit or stash your changes first") {
		t.Errorf("dirty tree: exit %d %q %q", code, out, errs)
	}
	if headOf(t, dir) != head {
		t.Error("a refusal committed")
	}
}

func TestCloseDryRunWritesNothing(t *testing.T) {
	dir := closeRepo(t, true, "")
	head := headOf(t, dir)
	code, out, errs := runCLI(t, dir, "brief", "close", closeName, "--finding", "README omits the flag", "--dry-run")
	if code != 0 || errs != "" {
		t.Fatalf("exit %d: %s", code, errs)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	n := len(lines)
	if !strings.HasPrefix(collapse(out), closeBrief+" consumed · not merged\n") || !strings.Contains(collapse(out), "operator 1") {
		t.Errorf("summary: %q", out)
	}
	if !strings.HasPrefix(lines[n-4], "would record .vloop/defects/D") || !strings.HasSuffix(lines[n-4], "-readme-omits-the-flag.md") {
		t.Errorf("record line: %q", lines[n-4])
	}
	for i, w := range []string{
		"would write .vloop/state/metrics/" + closeBrief + ".json",
		"would update docs/briefs/" + closeName + ".md",
		"would commit [vloop] close " + closeBrief,
	} {
		if lines[n-3+i] != w {
			t.Errorf("line %d = %q, want %q", n-3+i, lines[n-3+i], w)
		}
	}
	if headOf(t, dir) != head || gitOutput(t, dir, "status", "--porcelain", "--untracked-files=all") != "" {
		t.Error("dry run wrote or committed")
	}
	for _, p := range []string{".vloop/defects", ".vloop/state"} {
		if _, err := os.Stat(filepath.Join(dir, p)); err == nil {
			t.Errorf("dry run created %s", p)
		}
	}
}

func TestCloseDryRunRefusalsStillWin(t *testing.T) {
	dir := closeRepo(t, false, "")
	code, out, errs := runCLI(t, dir, "brief", "close", closeName, "--no-findings", "--dry-run")
	if code != 1 || out != "" || errs != "vloop: the plan is not complete (1/2 done) — finish it, or pass --abandon \"<reason>\"\n" {
		t.Errorf("stalled: exit %d %q %q", code, out, errs)
	}
	code, out, errs = runCLI(t, dir, "brief", "close", closeName, "--dry-run")
	if code != 2 || out != "" || errs != "vloop: "+closeFindingsMsg+"\n" {
		t.Errorf("no findings flag: exit %d %q %q", code, out, errs)
	}
	code, out, _ = runCLI(t, dir, "brief", "close", closeName, "--abandon", "x", "--no-findings", "--dry-run")
	if code != 0 || !strings.HasPrefix(collapse(out), closeBrief+" abandoned") {
		t.Errorf("abandon dry run: exit %d %q", code, out)
	}
}
