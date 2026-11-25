package main

import (
	"os/exec"
	"strings"
	"testing"
)

// TestWorkedExampleLocalConfigFirstBlock walks the first block of the B39
// worked example in its order, in a fresh repository.
func TestWorkedExampleLocalConfigFirstBlock(t *testing.T) {
	t.Parallel()
	s := newScratch(t)
	git := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", s.dir, "-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false"}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("add", "-A")
	git("commit", "-q", "-m", "init")
	expectOK := func(args ...string) { t.Helper(); expect(t, s.run(nil, args...), 0, "", "") }

	if r := s.run(nil, "init"); r.code != 0 {
		t.Fatalf("init: %+v", r)
	}
	cfg := s.read(".vloop/config.toml")
	s.write(".vloop/config.toml", cfg+"\n[[check]]\nname = \"all\"\npaths = [\"**\"]\nrun = \"true\"\n")

	expectOK("config", "set", "agent.work", "copilot")
	expectOK("config", "set", "model.work", "gpt-6.1-terra")
	expectOK("config", "set", "--local", "agent.work", "claude")
	expectOK("config", "set", "--local", "model.work", "sonnet")
	expect(t, s.run(nil, "config", "path", "--local"), 0, ".vloop/config.local.toml\n", "")

	r := s.run(nil, "config", "list")
	if r.code != 0 {
		t.Fatalf("config list: %+v", r)
	}
	var got []string
	for _, l := range strings.Split(r.out, "\n") {
		if strings.HasPrefix(l, "agent.work=") || strings.HasPrefix(l, "model.work=") {
			got = append(got, l)
		}
	}
	if strings.Join(got, "\n") != "model.work=sonnet (local)\nagent.work=claude (local)" {
		t.Fatalf("config list rows %q", got)
	}

	expect(t, s.run([]string{"VLOOP_AGENT_WORK=copilot"}, "config", "get", "--json", "agent.work"), 0,
		`{"key":"agent.work","value":"copilot","source":"env"}`+"\n", "")
	expect(t, s.run(nil, "config", "set", "model.work", "gpt-6-luna"), 0, "",
		"! model.work stays sonnet on this machine: .vloop/config.local.toml overrides it\n")

	found := false
	for _, l := range strings.Split(s.read(".gitignore"), "\n") {
		found = found || l == ".vloop/config.local.toml"
	}
	if !found {
		t.Fatalf(".gitignore lacks the local line:\n%s", s.read(".gitignore"))
	}

	doctorRows := func() []string {
		t.Helper()
		r := s.run(nil, "doctor")
		var rows []string
		for _, l := range strings.Split(r.out, "\n") {
			if strings.Contains(l, "config.local") {
				rows = append(rows, l)
			}
		}
		return rows
	}
	if rows := doctorRows(); len(rows) != 1 || rows[0] != "! config.local overrides model.work, agent.work — on this machine only" {
		t.Fatalf("doctor rows %q", rows)
	}
	expectOK("config", "set", "--local", "agent.work", "")
	expectOK("config", "set", "--local", "model.work", "")
	if rows := doctorRows(); len(rows) != 1 || rows[0] != "✓ config.local" {
		t.Fatalf("doctor rows after removal %q", rows)
	}
	expect(t, s.run(nil, "config", "get", "copilot.credit-usd"), 0, "0.01\n", "")
}
