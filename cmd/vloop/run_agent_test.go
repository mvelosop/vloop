package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestRunClaudeInvocations pins what a default run hands claude for each
// phase: the argv, the working directory, the variables that pass through and
// the ones the driver drops or sets itself.
func TestRunClaudeInvocations(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript)
	res := r.runWith([]string{
		"CLAUDE_CODE_USE_FOUNDRY=1", "ANTHROPIC_FOUNDRY_RESOURCE=res", "ANTHROPIC_DEFAULT_SONNET_MODEL=snt",
		"VLOOP_ACTIVE_TASK=leaked", "VLOOP_GATE_TASK=leaked", "VLOOP_PLAN_SHA256=leaked",
	}, runBrief)
	wantExit(t, res, 0)

	wantCwd, err := filepath.EvalSymlinks(r.dir)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, b := range r.invocations() {
		lines := strings.Split(strings.TrimSpace(b), "\n")
		if !strings.HasPrefix(lines[0], "arg -p") {
			continue // doctor's and preflight's probes
		}
		prompt := strings.TrimPrefix(lines[1], "arg ")
		phase := strings.TrimPrefix(strings.Fields(prompt)[0], "/vloop:")
		seen[phase] = true
		want := []string{"arg -p", "arg " + prompt, "arg --model"}
		for i, w := range want {
			if lines[i] != w {
				t.Errorf("%s: argument line %d = %q, want %q", phase, i, lines[i], w)
			}
		}
		tail := []string{"arg --permission-mode", "arg auto", "arg --setting-sources", "arg project", "arg --settings"}
		joined := "\n" + strings.Join(lines, "\n") + "\n"
		for _, w := range tail {
			if !strings.Contains(joined, "\n"+w+"\n") {
				t.Errorf("%s: no line %q", phase, w)
			}
		}
		for _, w := range []string{"arg --plugin-dir", "arg --strict-mcp-config", "arg --output-format", "arg json"} {
			if !strings.Contains(joined, "\n"+w+"\n") {
				t.Errorf("%s: no line %q", phase, w)
			}
		}
		for _, w := range []string{"env CLAUDE_CODE_USE_FOUNDRY=1", "env ANTHROPIC_FOUNDRY_RESOURCE=res", "env ANTHROPIC_DEFAULT_SONNET_MODEL=snt"} {
			if !strings.Contains(joined, "\n"+w+"\n") {
				t.Errorf("%s: %q did not reach claude", phase, w)
			}
		}
		for _, w := range []string{"env VLOOP_ACTIVE_TASK=", "env VLOOP_GATE_TASK="} {
			if strings.Contains(joined, "\n"+w) {
				t.Errorf("%s: a dropped variable reached claude: %s", phase, w)
			}
		}
		if strings.Contains(joined, "\nenv VLOOP_PLAN_SHA256=leaked\n") {
			t.Errorf("%s: the stale plan hash reached claude", phase)
		}
		if !strings.Contains(joined, "\ncwd "+wantCwd+"\n") {
			t.Errorf("%s: claude did not start in the repository root %s", phase, wantCwd)
		}
		hash := strings.Contains(joined, "\nenv VLOOP_PLAN_SHA256=")
		if wantHash := phase != "plan"; hash != wantHash && phase != "gate-review" {
			t.Errorf("%s: plan hash present = %v", phase, hash)
		}
	}
	for _, p := range []string{"plan", "gate-review", "work", "review", "verify"} {
		if !seen[p] {
			t.Errorf("no %s session was started", p)
		}
	}
}

// TestRunRecordsAreSessionV3 pins agent and cost_source in the records, with
// the cost reported and with none.
func TestRunRecordsAreSessionV3(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript)
	wantExit(t, r.vloop("run", runBrief), 0)
	for _, n := range r.sessionFiles() {
		var rec map[string]any
		sessionRecord(t, r.runFolder()+"/sessions/"+n, &rec)
		if rec["schema"] != "session/v3" || rec["agent"] != "claude" || rec["cost_source"] != "reported" {
			t.Errorf("%s: schema, agent, cost_source = %v, %v, %v", n, rec["schema"], rec["agent"], rec["cost_source"])
		}
		if _, ok := rec["ended"]; !ok {
			t.Errorf("%s: no ended", n)
		}
	}
}

// TestRunRecordsWithoutCost: a result with no total_cost_usd is a record with
// cost_source none and cost_usd 0.
func TestRunRecordsWithoutCost(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), "STUB_NOCOST=1\n"+defaultScript)
	wantExit(t, r.vloop("run", runBrief), 0)
	for _, n := range r.sessionFiles() {
		var rec map[string]any
		sessionRecord(t, r.runFolder()+"/sessions/"+n, &rec)
		if rec["cost_source"] != "none" || rec["cost_usd"] != 0.0 {
			t.Errorf("%s: cost_source, cost_usd = %v, %v", n, rec["cost_source"], rec["cost_usd"])
		}
	}
	if v := r.vloop("schema", "validate", "session/v3", r.runFolder()+"/sessions/"+r.sessionFiles()[0]); v.code != 0 {
		t.Errorf("a record with cost_source none does not validate: %s%s", v.out, v.err)
	}
}
