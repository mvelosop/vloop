package cli

// The T1 gate's oracle (B28, worked example): a project skill named like one of
// the two new plugin skills is a problem for a copilot review agent. Copied into
// a copy of the tree by the gate; never committed.

import (
	"strings"
	"testing"
)

func TestGateT1DoctorClash(t *testing.T) {
	for _, name := range []string{"review-brief", "create-brief"} {
		e := newDoctorEnv(t)
		e.write(".vloop/config.toml", doctorCheckConfig+"[agent]\nreview = \"copilot\"\n[model]\nreview = \"gpt-x\"\n")
		e.copilot(true)
		e.skill(name)
		out, code := e.run("doctor")
		want := "✗ .claude/skills/" + name + " has the name of vloop's " + name + " skill; a copilot session may load it instead — rename it"
		if code != 1 || !strings.Contains(out, want+"\n") {
			t.Errorf("doctor with .claude/skills/%s: exit %d, want 1 and the line %q:\n%s", name, code, want, out)
		}
	}
}
