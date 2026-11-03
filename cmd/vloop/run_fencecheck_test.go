package main

import (
	"strings"
	"testing"
)

// The driver's re-judge of the calls a Copilot work session ran (B34 F1).

const copilotWorkScript = `case "$PHASE" in
  work) touch "$TASK.out"; mkdir -p .vloop/tmp
    printf '{"schema":"proposal/v1","task":"%s","outcome":"done","summary":"made %s","files":[],"verified":"ok","notes":"none"}\n' "$TASK" "$TASK" > .vloop/tmp/proposal.json;;
esac
`

const (
	callBashCommit = `{"type":"tool.execution_start","data":{"toolCallId":"c1","toolName":"bash","arguments":{"command":"git commit -m x"}}}`
	callBashLs     = `{"type":"tool.execution_start","data":{"toolCallId":"c1","toolName":"bash","arguments":{"command":"ls"}}}`
	doneC1         = `{"type":"tool.execution_complete","data":{"toolCallId":"c1","success":true}}`
	deniedC1       = `{"type":"tool.execution_complete","data":{"toolCallId":"c1","success":false,"error":{"code":"denied","message":"Denied by preToolUse hook: no"}}}`
)

// copilotWork is a one-task run whose work session is Copilot's, printing calls.
func copilotWork(t *testing.T, calls ...string) (*runRepo, result) {
	t.Helper()
	r := newRunRepo(t)
	r.scripted(planJSON(t, planTask("T1", nil)), defaultScript)
	wantExit(t, r.vloop("config", "set", "agent.work", "copilot"), 0)
	wantExit(t, r.vloop("config", "set", "model.work", "gpt-5.6-terra"), 0)
	r.git("add", "-A")
	r.git("commit", "-q", "-m", "work on copilot")
	body := copilotWorkScript + "STUB_CALLS='" + strings.Join(calls, "\n") + "'\n"
	r.copilotScript(body)
	return r, r.exec(binPath, r.env("GH_TOKEN=gate-gh-token"), "run", runBrief)
}

func TestRunFenceBackstopBypassed(t *testing.T) {
	t.Parallel()
	r, res := copilotWork(t, callBashCommit, doneC1)
	wantExit(t, res, 9)
	if !strings.Contains(res.err, "vloop: bash git commit -m x — ") {
		t.Errorf("no line naming the call:\n%s", res.err)
	}
	wantIn(t, "stderr", res.err, "FENCE BYPASSED work — the session ran 1 call(s) its fence denies; nothing from it was used")
	if subs := r.subjects("main..HEAD"); len(subs) != 1 || subs[0] != "[vloop] plan "+runID {
		t.Errorf("commits after the plan: %v", subs)
	}
}

func TestRunFenceBackstopDeniedByHook(t *testing.T) {
	t.Parallel()
	_, res := copilotWork(t, callBashCommit, deniedC1)
	wantExit(t, res, 0)
	if strings.Contains(res.err, "FENCE BYPASSED") {
		t.Errorf("a denied call halted the run:\n%s", res.err)
	}
}

func TestRunFenceBackstopAllowedCall(t *testing.T) {
	t.Parallel()
	_, res := copilotWork(t, callBashLs, doneC1)
	wantExit(t, res, 0)
	if strings.Contains(res.err, "FENCE BYPASSED") || strings.Contains(res.err, "vloop: bash") {
		t.Errorf("an allowed call was named:\n%s", res.err)
	}
}
