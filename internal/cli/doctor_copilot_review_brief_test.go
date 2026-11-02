package cli

import (
	"strings"
	"testing"
)

func TestCopilotReviewSkillNamedReviewBriefClashes(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", doctorCheckConfig+"[agent]\nreview = \"copilot\"\n[model]\nreview = \"gpt-x\"\n")
	e.copilot(true)
	e.skill("review-brief")
	out, code := e.run("doctor")
	want := "✗ .claude/skills/review-brief has the name of vloop's review-brief skill; a copilot session may load it instead — rename it"
	if code != 1 || !strings.Contains(out, want) {
		t.Errorf("doctor exit %d, want line %q:\n%s", code, want, out)
	}
}
