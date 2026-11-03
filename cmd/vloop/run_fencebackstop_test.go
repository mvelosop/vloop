package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// B34 F1 in every phase: a Copilot session that ran a bash `git push` halts
// with exit 9 where REFS MOVED halts for that phase, and nothing from it is used.

const (
	callBashPush = `{"type":"tool.execution_start","data":{"toolCallId":"c1","toolName":"bash","arguments":{"command":"git push"}}}`
	pushLine     = "vloop: bash git push — "
	bypassTail   = " — the session ran 1 call(s) its fence denies; nothing from it was used"
)

// phaseOnCopilot is a one-task run whose session of phase (and only that one)
// is Copilot's and runs `git push`; env adds more configuration.
func phaseOnCopilot(t *testing.T, phase string, env ...string) (*runRepo, result) {
	t.Helper()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript)
	r.copilotScript(defaultScript + "STUB_CALLS='" + callBashPush + "\n" + doneC1 + "'\n")
	key := strings.ToUpper(strings.ReplaceAll(phase, "-", "_"))
	env = append(env, "GH_TOKEN=gate-gh-token", "VLOOP_AGENT_"+key+"=copilot", "VLOOP_MODEL_"+key+"=gpt-5.6-terra")
	return r, r.runWith(env, runBrief)
}

func wantBypassed(t *testing.T, res result, phase string) {
	t.Helper()
	wantExit(t, res, 9)
	wantIn(t, "stderr", res.err, pushLine, "FENCE BYPASSED "+phase+bypassTail)
}

func TestRunFenceBackstopEveryPhase(t *testing.T) {
	t.Parallel()
	t.Run("plan", func(t *testing.T) {
		t.Parallel()
		r, res := phaseOnCopilot(t, "plan")
		wantBypassed(t, res, "plan")
		if n := r.planCommits(); n != 0 {
			t.Errorf("%d plan commits", n)
		}
		wantNotIn(t, "argv", strings.Join(r.argv(), "\n"), "/vloop:work")
		wantIn(t, "run log", r.runLog(), "FENCE BYPASSED plan"+bypassTail)
	})
	t.Run("gate-review", func(t *testing.T) {
		t.Parallel()
		r, res := phaseOnCopilot(t, "gate-review")
		wantBypassed(t, res, "gate-review")
		if n := r.planCommits(); n != 0 {
			t.Errorf("%d plan commits", n)
		}
		wantNotIn(t, "argv", strings.Join(r.argv(), "\n"), "/vloop:work")
	})
	t.Run("review", func(t *testing.T) {
		t.Parallel()
		r, res := phaseOnCopilot(t, "review")
		wantBypassed(t, res, "review")
		for _, s := range r.subjects("--all") {
			if strings.HasPrefix(s, "[vloop] T1") {
				t.Errorf("committed %q", s)
			}
		}
		wantIn(t, "run log", r.runLog(), pushLine)
	})
	t.Run("second-review", func(t *testing.T) {
		t.Parallel()
		r, res := phaseOnCopilot(t, "second-review", "VLOOP_RUN_SECOND_REVIEW=record")
		wantBypassed(t, res, "second-review")
		for _, s := range r.subjects("--all") {
			if strings.HasPrefix(s, "[vloop] T1") {
				t.Errorf("committed %q", s)
			}
		}
	})
	t.Run("verify", func(t *testing.T) {
		t.Parallel()
		r, res := phaseOnCopilot(t, "verify")
		wantBypassed(t, res, "verify")
		if _, err := os.Stat(filepath.Join(r.runFolder(), "reports", "verify.json")); err == nil {
			t.Error("reports/verify.json was kept")
		}
	})
	t.Run("review-brief", func(t *testing.T) {
		t.Parallel()
		r := newRunRepo(t)
		r.copilotScript(`echo tampered >> ` + runBrief + "\nSTUB_CALLS='" + callBashPush + "\n" + doneC1 + "'\n")
		before := r.read(runBrief)
		res := r.exec(binPath, r.env("GH_TOKEN=gate-gh-token", "VLOOP_AGENT_REVIEW_BRIEF=copilot", "VLOOP_MODEL_REVIEW_BRIEF=gpt-5.6-terra"), "brief", "review", runBrief)
		wantBypassed(t, res, "review-brief")
		wantNotIn(t, "stdout", res.out, "verdict: ")
		if got := r.read(runBrief); got != before {
			t.Errorf("the brief was not restored")
		}
	})
}
