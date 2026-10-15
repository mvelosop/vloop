package metrics

import (
	"github.com/mvelosop/vloop/internal/budget"
	"github.com/mvelosop/vloop/internal/gitrun"
)

// Instructions is the words each session kind starts with, counted at the
// commit a brief's delivered lines are measured to: its skill, the repository's
// guidelines and the root CLAUDE.md. A kind with no skill there is nil.
type Instructions struct {
	Plan            *int `json:"plan"`
	GateReview      *int `json:"gate_review"`
	Work            *int `json:"work"`
	Review          *int `json:"review"`
	GuidelinesRules int  `json:"guidelines_rules"`
}

const guidelinesPath = ".vloop/repo-guidelines.md"

// showAt returns the file at the commit, and whether it exists there.
func showAt(root, sha, path string) ([]byte, bool) {
	out, err := gitrun.At(root, "show", sha+":"+path).Output()
	return out, err == nil
}

// MeasureInstructions computes the instructions of the tree at sha from git.
// A kind's skill is plugin/skills/<kind>/SKILL.md, else for plan, work and
// review the shell loop's .claude/skills/loop-<kind>/SKILL.md, else the kind is
// not counted.
func MeasureInstructions(root, sha string) *Instructions {
	var shared int
	var rules int
	if b, ok := showAt(root, sha, "CLAUDE.md"); ok {
		shared += budget.CountWords(b)
	}
	if b, ok := showAt(root, sha, guidelinesPath); ok {
		shared += budget.CountWords(b)
		r, _ := budget.GuidelineRules(string(b))
		rules = len(r)
	}
	kind := func(dir, shell string) *int {
		b, ok := showAt(root, sha, "plugin/skills/"+dir+"/SKILL.md")
		if !ok && shell != "" {
			b, ok = showAt(root, sha, ".claude/skills/loop-"+shell+"/SKILL.md")
		}
		if !ok {
			return nil
		}
		n := budget.CountWords(b) + shared
		return &n
	}
	return &Instructions{
		Plan:            kind("plan", "plan"),
		GateReview:      kind("gate-review", ""),
		Work:            kind("work", "work"),
		Review:          kind("review", "review"),
		GuidelinesRules: rules,
	}
}
