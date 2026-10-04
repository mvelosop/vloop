package cli

import (
	"strings"
	"testing"
)

var fencePhases = []string{"plan", "work", "review", "gate-review"}

// f10Forms are the bypasses the v1.0 review listed (F10): each must be denied
// by every phase's fence.
var f10Forms = []string{
	"git -C . commit -m x",
	"git -c core.hooksPath=x push",
	"git -c k=v push origin main",
	"git merge main",
	"git cherry-pick abc123",
	"git revert HEAD",
	"git am x.patch",
	"git pull",
	"git fetch origin",
	"git push https://example.com/x.git main",
	"git config user.name x",
	"git notes add -m x",
	"git replace a b",
	"git filter-branch --all",
	"git worktree add ../x",
	"git reflog expire --all",
	"git gc --prune=now",
	"find . -delete",
	"find . -name x -delete",
	"find . -exec rm {} ;",
	"find . -fprint out.txt",
	"git diff --output=x.txt",
	"git diff HEAD --output=x.txt",
	"git log --output=x.txt",
	"git show HEAD --output=x.txt",
	"rm -fr x",
	"rm -r -f x",
	"vloop -C . task set T1 status done",
	"vloop brief new demo",
}

// skillAllowed are commands the skills tell sessions to run; every phase's
// fence still allows them.
var skillAllowed = []string{
	"vloop task show T1 --json",
	"vloop status",
	"vloop task gate T1",
	"vloop task list",
	"vloop task validate",
	"vloop config get language",
	"vloop schema validate proposal/v1 .vloop/tmp/proposal.json",
	"git status",
	"git diff",
	"git diff HEAD~1 -- internal/x.go",
	"git log -n 3",
	"git show HEAD",
	"git rev-parse HEAD",
	"ls .vloop/state/journals",
	"cat CLAUDE.md",
	"head -n 5 CLAUDE.md",
	"tail -n 5 CLAUDE.md",
	"wc -l CLAUDE.md",
	"find . -name '*.go'",
	"mkdir -p .vloop/tmp",
	"jq . .vloop/tmp/proposal.json",
}

func TestFenceDeniesBypasses(t *testing.T) {
	for _, phase := range fencePhases {
		f, err := loadFence(phase)
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range f10Forms {
			if _, ok := prefixMatch(f.deny, c); !ok {
				t.Errorf("%s fence does not deny `%s`", phase, c)
			}
		}
		for _, c := range skillAllowed {
			if p := f.judge(c); p != "" {
				t.Errorf("%s fence: `%s` is %s", phase, c, p)
			}
		}
	}
}

// pathMatch matches a repo-relative path against a Read/Edit/Write rule body:
// "dir/**" covers everything under dir, a body with no "*" one path.
func pathMatch(p, path string) bool {
	if pre, ok := strings.CutSuffix(p, "/**"); ok {
		return strings.HasPrefix(path, pre+"/")
	}
	return p == path
}

func TestPhaseFenceWrites(t *testing.T) {
	type w struct {
		path               string
		plan, work, review bool // allowed in the plan, work and review fences
	}
	writes := []w{
		{".vloop/state/state.json", true, false, false},
		{".vloop/tmp/proposal.json", true, true, true},
		{".vloop/tmp/verdict.json", true, true, true},
		{".vloop/config.toml", false, false, false},
		{".vloop/defects/D1.md", false, false, false},
		{".vloop/interventions/I1.md", false, false, false},
		{".vloop/state/runs/r/sessions/001-plan.json", false, false, false},
		{".git/config", false, false, false},
		{".git/hooks/pre-commit", false, false, false},
		{"internal/x/x.go", false, true, false},
		{"docs/x.md", false, true, false},
	}
	for _, phase := range fencePhases {
		f, err := readFence(phase)
		if err != nil {
			t.Fatal(err)
		}
		for _, tool := range []string{"Edit", "Write"} {
			for _, x := range writes {
				want := map[string]bool{"plan": x.plan, "work": x.work, "review": x.review, "gate-review": x.review}[phase]
				if got := writeAllowed(f.Permissions.Allow, f.Permissions.Deny, tool, x.path); got != want {
					t.Errorf("%s fence: %s %s allowed = %v, want %v", phase, tool, x.path, got, want)
				}
			}
			if writeAllowed(f.Permissions.Allow, f.Permissions.Deny, tool, "~/.claude/settings.json") {
				t.Errorf("%s fence allows %s on ~/.claude", phase, tool)
			}
		}
	}
}

func writeAllowed(allow, deny []string, tool, path string) bool {
	match := func(rules []string) bool {
		for _, r := range rules {
			if r == tool {
				return true
			}
			if b, ok := strings.CutPrefix(r, tool+"("); ok && pathMatch(strings.TrimSuffix(b, ")"), path) {
				return true
			}
		}
		return false
	}
	return !match(deny) && match(allow)
}
