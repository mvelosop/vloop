package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fakeSession = `{"type":"assistant.message","data":{"content":"hello","toolRequests":[]}}
{"type":"tool.execution_start","data":{"toolCallId":"c1","toolName":"bash","arguments":{"command":"echo a<b&c"}}}
{"type":"tool.execution_complete","data":{"toolCallId":"c1","result":{"content":"out"}}}
{"type":"tool.execution_start","data":{"toolCallId":"c2","toolName":"report_intent","arguments":{"intent":"x"}}}
{"type":"tool.execution_complete","data":{"toolCallId":"c2","result":{}}}
{"type":"assistant.message","data":{"content":""}}
{"type":"result","exitCode":0,"usage":{"sessionDurationMs":5}}
`

// pathWith installs fake copilot and vloop scripts first on PATH and returns the vloop path.
func pathWith(t *testing.T, copilot string) string {
	t.Helper()
	bin := t.TempDir()
	write(t, filepath.Join(bin, "copilot"), "#!/bin/sh\n"+copilot)
	write(t, filepath.Join(bin, "vloop"), "#!/bin/sh\nexit 0\n")
	for _, n := range []string{"copilot", "vloop"} {
		if err := os.Chmod(filepath.Join(bin, n), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return filepath.Join(bin, "vloop")
}

func graderFile(t *testing.T, cases, kase, name, front string) {
	write(t, filepath.Join(cases, kase, "graders", name+".md"), "---\n"+front+"---\nbody\n")
}

func readReport(t *testing.T, out string) report {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(out, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var r report
	if err := json.Unmarshal(data, &r); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestObservedCopilotTrace(t *testing.T) {
	raw, err := os.ReadFile("testdata/observed/copilot-events.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	events := copilotTrace(raw)
	if len(events) == 0 {
		t.Fatal("no events read from the observed sample")
	}
	var tools int
	for _, e := range events {
		if e.Kind == "text" && e.Text == "" {
			t.Errorf("an empty text event")
		}
		if e.Kind == "tool" {
			tools++
			if e.Tool == "bash" || e.Tool == "view" || e.Tool == "create" {
				t.Errorf("tool %s was not mapped", e.Tool)
			}
		}
	}
	if tools == 0 {
		t.Error("no tool event in the observed sample")
	}
}

func TestCopilotRunEndToEnd(t *testing.T) {
	cop := `case "$1" in --version) echo 1; exit 0;; esac
mkdir -p .vloop/tmp
echo '{"verdict": "FAIL"}' > .vloop/tmp/verdict.json
cat <<'EOF2'
` + fakeSession + `EOF2
`
	vl := pathWith(t, cop)
	cases := t.TempDir()
	write(t, filepath.Join(cases, "_lib.sh"), "")
	write(t, filepath.Join(cases, "rev", "prompt.md"), "---\nmax_turns: 1\n---\n\n/vloop:review T1\n")
	write(t, filepath.Join(cases, "rev", "scaffold.sh"), "mkdir -p .vloop/tmp\n")
	write(t, filepath.Join(cases, "op", "prompt.md"), "/vloop:operate\n")
	write(t, filepath.Join(cases, "bad", "prompt.md"), "/vloop:review T1\n")
	write(t, filepath.Join(cases, "bad", "scaffold.sh"), "echo one >&2\necho broke >&2\nexit 3\n")
	graderFile(t, cases, "rev", "a", "type: regex\nweight: 1\ntarget: trace\nmatch: contains\npattern: 'Bash .*a<b&c'\n")
	graderFile(t, cases, "rev", "b", "type: regex\nweight: 2\ntarget:\n  source: file\n  path: .vloop/tmp/verdict.json\nmatch: contains\npattern: PASS\n")
	graderFile(t, cases, "rev", "c", "type: tool_used\ntool: report_intent\n")
	graderFile(t, cases, "rev", "d", "type: tool_used\ntool: Bash\ninput_match: 'echo\\s+a'\nmax: 0\n")
	graderFile(t, cases, "rev", "e", "type: file_exists\npath: .vloop/tmp/verdict.json\n")
	graderFile(t, cases, "rev", "f", "type: regex\ntarget:\n  source: file\n  path: gone\nmatch: contains\npattern: x\n")
	graderFile(t, cases, "rev", "g", "type: tool_order\n")
	out := filepath.Join(t.TempDir(), "o")
	code, stdout, stderr := invoke("--agent", "copilot", "--model", "m", "--cases", cases, "--out", out, "--vloop", vl, "--timeout", "1m")
	if code != 1 {
		t.Fatalf("exit %d, want 1 (the bad case errors): %s", code, stderr)
	}
	rep := readReport(t, out)
	byName := map[string]caseResult{}
	for _, c := range rep.Cases {
		byName[c.Name] = c
	}
	op := byName["op"]
	if op.Status != "skipped" || op.Reason != "no phase" || op.Phase != "" || len(op.Runs) != 0 {
		t.Errorf("op = %+v", op)
	}
	bad := byName["bad"]
	if bad.Status != "error" || bad.Runs[0].Reason != "scaffold exited 3: broke" || bad.Mean != nil || bad.Runs[0].Score != nil {
		t.Errorf("bad = %+v", bad)
	}
	rev := byName["rev"]
	got := map[string]float64{}
	for _, g := range rev.Runs[0].Graders {
		got[g.Name] = g.Score
	}
	want := map[string]float64{"a": 1, "b": 0, "c": 1, "d": 0, "e": 1, "f": 0, "g": 0}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("grader %s = %v, want %v", k, got[k], v)
		}
	}
	if rev.Status != "scored" || rev.Phase != "review" || rev.Runs[0].Credits == nil {
		t.Errorf("rev = %+v", rev)
	}
	for _, r := range rev.Runs[0].Graders {
		if r.Name == "f" && r.Reason != "no gone" {
			t.Errorf("f's reason = %q", r.Reason)
		}
		if r.Name == "g" && r.Reason != "unsupported grader tool_order" {
			t.Errorf("g's reason = %q", r.Reason)
		}
	}
	dir := filepath.Join(out, "rev", "run-1")
	for _, p := range []string{"repo/.vloop/tmp/verdict.json", "transcript.jsonl", "trace.jsonl"} {
		if _, err := os.Stat(filepath.Join(dir, p)); err != nil {
			t.Errorf("missing %s", p)
		}
	}
	if m, _ := filepath.Glob(filepath.Join(dir, "sessions", "*.json")); len(m) != 1 {
		t.Errorf("session records: %v", m)
	}
	trace, _ := os.ReadFile(filepath.Join(dir, "trace.jsonl"))
	if !strings.Contains(string(trace), `"input":"{\"command\":\"echo a<b&c\"}"`) || strings.Contains(string(trace), "\\u003c") {
		t.Errorf("trace = %s", trace)
	}
	hook, _ := filepath.Glob(filepath.Join(dir, "repo", ".vloop", "tmp", "copilot-home-*", "hooks", "vloop-fence.json"))
	if len(hook) != 1 {
		t.Fatalf("hooks files: %v", hook)
	}
	if data, _ := os.ReadFile(hook[0]); !strings.Contains(string(data), "'"+vl+"' fence hook --phase review --root ") {
		t.Errorf("the hook does not run --vloop's binary: %s", data)
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 5 || !strings.HasPrefix(lines[0], "CASE") || !strings.HasSuffix(lines[4], "3 case(s) · 1 skipped · 1 error(s) · $0.00") {
		t.Fatalf("stdout:\n%s", stdout)
	}
	if !strings.Contains(lines[1], "scaffold exited 3: broke") || !strings.HasSuffix(lines[2], "no phase") || !strings.HasSuffix(strings.TrimSpace(lines[3]), "d, f, g") {
		t.Errorf("stdout:\n%s", stdout)
	}
}

func TestNoResultAndTimeout(t *testing.T) {
	cases := t.TempDir()
	write(t, filepath.Join(cases, "rev", "prompt.md"), "/vloop:review T1\n")
	write(t, filepath.Join(cases, "rev", "scaffold.sh"), "true\n")
	graderFile(t, cases, "rev", "e", "type: file_exists\npath: x\n")
	for _, c := range []struct{ script, timeout, reason string }{
		{"echo nothing >&2; exit 1\n", "1m", "no session result"},
		{"sleep 20\n", "1s", "timed out after 1s"},
	} {
		vl := pathWith(t, "case \"$1\" in --version) echo 1; exit 0;; esac\n"+c.script)
		out := filepath.Join(t.TempDir(), "o")
		code, _, _ := invoke("--agent", "copilot", "--model", "m", "--cases", cases, "--out", out, "--vloop", vl, "--timeout", c.timeout)
		r := readReport(t, out).Cases[0].Runs[0]
		if code != 1 || r.Status != "error" || r.Reason != c.reason {
			t.Errorf("%s: exit %d, run %+v", c.reason, code, r)
		}
	}
}
