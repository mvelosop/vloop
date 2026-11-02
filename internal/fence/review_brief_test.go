package fence

import "testing"

func TestReviewBriefPhase(t *testing.T) {
	const B = "docs/briefs/B20990101-0000-x.loop-brief.md"
	const ow = " is not allowed in a review-brief session: only .vloop/tmp/ may be written"
	j, _, _ := newJudge(t, PhaseReviewBrief)
	cases := []struct {
		tool, path, cmd string
		want            string // "" allows, else the exact reason
	}{
		{"bash", "", "vloop brief check .vloop/tmp/rb/x.loop-brief.md", ""},
		{"bash", "", "cat " + B, ""},
		{"create", ".vloop/tmp/brief-review.json", "", ""},
		{"create", B, "", "write to " + B + ow},
		{"edit", ".vloop/state/state.json", "", "write to .vloop/state/state.json" + ow},
		{"bash", "", "git commit -m x", "git commit -m x is denied by the review-brief fence (Bash(git commit:*))"},
		{"bash", "", "rm " + B, "rm " + B + " is denied by the Copilot deny list (rm:*)"},
		{"bash", "", "echo x >> " + B, "write to " + B + ow},
		{"skill", "review-brief", "", ""},
		{"skill", "vloop-review-brief", "", "skill vloop-review-brief is not vloop's — use /vloop:review-brief"},
		{"web_fetch", "", "", "tool web_fetch is not allowed in a review-brief session"},
		{"bash", "", "vloop brief review " + B, "vloop brief review " + B + " is denied by the review-brief fence (Bash(vloop brief review:*))"},
	}
	for _, c := range cases {
		d := j.Decide(c.tool, c.path, c.cmd)
		if c.want == "" && !d.Allow || c.want != "" && (d.Allow || d.Reason != c.want) {
			t.Errorf("%s %q %q: got allow=%v %q, want %q", c.tool, c.path, c.cmd, d.Allow, d.Reason, c.want)
		}
	}
}

func TestNestedBriefReviewDeniedInEveryPhase(t *testing.T) {
	fences := map[string]string{"plan": "plan", "review": "review", "second-review": "review", "verify": "verify",
		"gate-review": "gate-review", "work": "work", "review-brief": "review-brief"}
	for ph, name := range fences {
		j, _, _ := newJudge(t, ph)
		want := "vloop brief review x.loop-brief.md is denied by the " + name + " fence (Bash(vloop brief review:*))"
		for _, cmd := range []string{"vloop brief review x.loop-brief.md", "env vloop brief review x.loop-brief.md"} {
			if d := j.Decide("bash", "", cmd); d.Allow || d.Reason != want {
				t.Errorf("%s %q: got allow=%v %q, want %q", ph, cmd, d.Allow, d.Reason, want)
			}
		}
	}
}

func TestGateReviewSkillHintUnchanged(t *testing.T) {
	j, _, _ := newJudge(t, PhaseGateReview)
	d := j.Decide("skill", "vloop-review-brief", "")
	if d.Allow || d.Reason != "skill vloop-review-brief is not vloop's — use /vloop:gate-review" {
		t.Errorf("got %v %q", d.Allow, d.Reason)
	}
}
