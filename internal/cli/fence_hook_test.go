package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func runHook(t *testing.T, phase, stdin string) (int, string) {
	t.Helper()
	root := t.TempDir()
	var o, e bytes.Buffer
	root2, _ := NewRoot(Build{})
	root2.SetArgs([]string{"fence", "hook", "--phase", phase, "--root", root})
	root2.SetIn(strings.NewReader(stdin))
	root2.SetOut(&o)
	root2.SetErr(&e)
	err := root2.Execute()
	code := 0
	if err != nil {
		code = ExitProblems
		if xe, ok := err.(*ExitError); ok {
			code = xe.Code
		}
	}
	return code, o.String()
}

func decision(t *testing.T, out string) (string, string) {
	t.Helper()
	if strings.Count(out, "\n") != 1 || !strings.HasSuffix(out, "\n") {
		t.Fatalf("want one line, got %q", out)
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("not JSON: %q", out)
	}
	return m["permissionDecision"], m["permissionDecisionReason"]
}

func TestFenceHookAnswers(t *testing.T) {
	patch := "*** Begin Patch\n*** Update File: .git/config\n@@\n-a\n+b\n*** End Patch\n"
	enc := func(v any) string { b, _ := json.Marshal(v); return string(b) }
	cases := []struct {
		name, payload, want string
	}{
		{"shell allow", enc(map[string]any{"toolName": "bash", "cwd": "x", "toolArgs": map[string]any{"command": "git diff HEAD"}}), "allow"},
		{"shell deny", enc(map[string]any{"toolName": "bash", "toolArgs": map[string]any{"command": "env gh auth token"}}), "deny"},
		{"toolArgs as string", enc(map[string]any{"toolName": "bash", "toolArgs": `{"command":"sh -c \"gh pr list\""}`}), "deny"},
		{"toolArgs string allow", enc(map[string]any{"toolName": "bash", "toolArgs": `{"command":"git status"}`}), "allow"},
		{"file allow", enc(map[string]any{"toolName": "create", "toolArgs": map[string]any{"path": ".vloop/tmp/v.json"}}), "allow"},
		{"file deny", enc(map[string]any{"toolName": "edit", "toolArgs": map[string]any{"path": "src/app.ts"}}), "deny"},
		{"file without path", enc(map[string]any{"toolName": "create", "toolArgs": map[string]any{}}), "deny"},
		{"shell without toolArgs", `{"toolName":"bash","cwd":"x"}`, "deny"},
		{"shell with an unread toolArgs", enc(map[string]any{"toolName": "bash", "toolArgs": "git status"}), "deny"},
		{"unknown tool", enc(map[string]any{"toolName": "frobnicate", "toolArgs": map[string]any{}}), "deny"},
		{"patch raw string", enc(map[string]any{"toolName": "apply_patch", "toolArgs": patch}), "deny"},
		{"patch input field", enc(map[string]any{"toolName": "apply_patch", "toolArgs": map[string]any{"input": patch}}), "deny"},
		{"patch in a string object", enc(map[string]any{"toolName": "apply_patch", "toolArgs": enc(map[string]any{"patch": patch})}), "deny"},
		{"patch in an unread shape", enc(map[string]any{"toolName": "apply_patch", "toolArgs": map[string]any{"diff": patch}}), "deny"},
		{"patch without paths", enc(map[string]any{"toolName": "apply_patch", "toolArgs": "nothing"}), "deny"},
		{"patch allowed paths", enc(map[string]any{"toolName": "apply_patch", "toolArgs": "*** Add File: .vloop/tmp/a.txt\n+a\n"}), "allow"},
	}
	for _, ph := range []string{"review", "second-review"} {
		for _, c := range cases {
			code, out := runHook(t, ph, c.payload)
			d, why := decision(t, out)
			if code != 0 || d != c.want {
				t.Errorf("[%s] %s: exit %d, %q", ph, c.name, code, out)
			}
			if d == "deny" && !strings.HasPrefix(why, "vloop fence: ") {
				t.Errorf("[%s] %s: reason %q", ph, c.name, why)
			}
		}
	}
}

func TestFenceHookFailsClosed(t *testing.T) {
	ok := `{"toolName":"bash","toolArgs":{"command":"git status"}}`
	for _, c := range []struct{ name, phase, payload string }{
		{"not JSON", "review", "not json"},
		{"empty", "review", ""},
		{"no toolName", "review", `{"toolArgs":{"command":"git status"}}`},
		{"empty object", "second-review", `{}`},
		{"plan phase", "plan", ok},
		{"unknown phase", "bogus", ok},
		{"no phase", "", ok},
	} {
		code, out := runHook(t, c.phase, c.payload)
		d, why := decision(t, out)
		if code != ExitUsage || d != "deny" || !strings.HasPrefix(why, "vloop fence: ") {
			t.Errorf("%s: exit %d, %q", c.name, code, out)
		}
	}
}
