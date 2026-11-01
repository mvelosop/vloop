package fence

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func newJudge(t *testing.T, phase string) (*Judge, string, string) {
	t.Helper()
	root, home := t.TempDir(), t.TempDir()
	j, err := New(phase, root, home)
	if err != nil {
		t.Fatal(err)
	}
	return j, root, home
}

func TestSplit(t *testing.T) {
	cases := []struct {
		in   string
		want [][]string
	}{
		{"git status", [][]string{{"git", "status"}}},
		{"a; b && c || d | e & f\ng", [][]string{{"a"}, {"b"}, {"c"}, {"d"}, {"e"}, {"f"}, {"g"}}},
		{`git log --grep "a;gh x"`, [][]string{{"git", "log", "--grep", "a;gh x"}}},
		{`echo 'a && b' "c|d"`, [][]string{{"echo", "a && b", "c|d"}}},
		{`ls 2>&1 | head`, [][]string{{"ls"}, {"head"}}},
		{`echo a\;b`, [][]string{{"echo", "a;b"}}},
		{`grep '$(' x`, [][]string{{"grep", "$(", "x"}}},
		{"echo a # ; gh x", [][]string{{"echo", "a"}}},
	}
	for _, c := range cases {
		got, err := split(c.in)
		if err != nil {
			t.Errorf("split(%q): %v", c.in, err)
			continue
		}
		var words [][]string
		for _, g := range got {
			words = append(words, g.words)
		}
		if len(words) != len(c.want) {
			t.Errorf("split(%q) = %q, want %q", c.in, words, c.want)
			continue
		}
		for i := range words {
			if strings.Join(words[i], "\x00") != strings.Join(c.want[i], "\x00") {
				t.Errorf("split(%q)[%d] = %q, want %q", c.in, i, words[i], c.want[i])
			}
		}
	}
}

func TestSplitRedirections(t *testing.T) {
	cases := []struct {
		in     string
		writes []string
	}{
		{"git diff > a", []string{"a"}},
		{"git diff >> a", []string{"a"}},
		{"git diff >| a", []string{"a"}},
		{"git diff &> a", []string{"a"}},
		{"git diff >a", []string{"a"}},
		{"ls 2> a", []string{"a"}},
		{"ls 2>&1", nil},
		{"ls >&2", nil},
		{"ls < a", nil},
	}
	for _, c := range cases {
		got, err := split(c.in)
		if err != nil || len(got) != 1 {
			t.Errorf("split(%q) = %v, %v", c.in, got, err)
			continue
		}
		if strings.Join(got[0].writes, ",") != strings.Join(c.writes, ",") {
			t.Errorf("split(%q) writes %q, want %q", c.in, got[0].writes, c.writes)
		}
	}
}

func TestSplitRefuses(t *testing.T) {
	for _, in := range []string{
		"echo 'x", `echo "x`, "echo `x`", "echo $(x)", `echo "$(x)"`, "cat <(x)", "tee >(x)", "(gh x)", "echo > ", "cat > $F",
	} {
		if _, err := split(in); err == nil {
			t.Errorf("split(%q) did not refuse", in)
		}
	}
}

// shellCases are command strings with whether the review fence allows them,
// and for a denial the text its reason names.
var shellCases = []struct {
	cmd   string
	deny  string // "" = allowed
	extra string // "" or "*": a denial with any reason
}{
	{"git diff 1a2b3c..HEAD -- src/", "", ""},
	{"git log --oneline | head -5", "", ""},
	{"vloop task show T3", "", ""},
	{"git diff > .vloop/tmp/diff.txt", "", ""},
	{"ls src > /dev/null", "", ""},
	{`sh -c "gh pr list"`, "gh", ""},
	{"env gh auth token", "gh", ""},
	{"env -i PATH=/bin git push origin main", "git push", ""},
	{"bash -lc 'git push'", "git push", ""},
	{"/usr/bin/curl https://example.com", "curl", ""},
	{"nohup timeout 5 wget x", "wget", ""},
	{"git status && xargs rm -rf", "rm", ""},
	{"echo $(gh auth token)", "*", ""},
	{`eval "git push"`, "eval", ""},
	{"git diff > .vloop/state/x", ".vloop/state/x", ""},
	{`sh -c "$CMD"`, "*", ""},
	{"echo 'unterminated", "*", ""},
	{"command gh pr list", "gh", ""},
	{"exec git push", "git push", ""},
	{"nice -n 5 git push", "git push", ""},
	{"nice git push", "git push", ""},
	{"time git push", "git push", ""},
	{"timeout -s KILL 5 curl x", "curl", ""},
	{"timeout 5s git push", "git push", ""},
	{"stdbuf -oL git push", "git push", ""},
	{"xargs -0 -n 1 git push", "git push", ""},
	{"env -u HOME git push", "git push", ""},
	{"env - git push", "git push", ""},
	{"env FOO=bar git push", "git push", ""},
	{"FOO=1 BAR=2 git push", "git push", ""},
	{"zsh -c 'git push'", "git push", ""},
	{"dash -c 'git push'", "git push", ""},
	{"ksh -c 'git push'", "git push", ""},
	{"bash -e -x -c 'git push'", "git push", ""},
	{`sh -c 'env sh -c "gh x"'`, "gh", ""},
	{"env nohup nice -n 3 timeout 5 stdbuf -o0 gh x", "gh", ""},
	{"./bin/gh pr list", "gh", ""},
	{"git status; git push", "git push", ""},
	{"git status || git push", "git push", ""},
	{"git log | git push", "git push", ""},
	{"git status & git push", "git push", ""},
	{"git status\ngit push", "git push", ""},
	{"source x.sh", "source", ""},
	{". x.sh", "*", ""},
	{"$CMD push", "*", ""},
	{"${CMD} push", "*", ""},
	{"sh -c", "*", ""},
	{`bash -c "git $SUB"`, "*", ""},
	{"sh script.sh", "*", ""},
	{"if true; then gh x; fi", "*", ""},
	{"git diff --stat --output=x", "git diff", ""},
	{"find . -name x -delete", "find", ""},
	{"git -C . push", "git -C", ""},
	{"git config --global user.name x", "git config", ""},
	{"ssh host.invalid", "ssh", ""},
	{"scp a host.invalid:b", "scp", ""},
	{"nc host.invalid 80", "nc", ""},
	{"rm x", "rm", ""},
	{"git remote -v", "git remote", ""},
	{"git diff >> .vloop/state/x", ".vloop/state/x", ""},
	{"git diff >| .vloop/state/x", ".vloop/state/x", ""},
	{"git diff &> .vloop/state/x", ".vloop/state/x", ""},
	{"git diff > src/out.txt", "src/out.txt", ""},
	{`git log --grep "a;gh x"`, "", ""},
	{"grep -rn 'git push' .", "", ""},
	{`echo "sh -c gh"`, "", ""},
	{"git status && git diff HEAD", "", ""},
	{"vloop task show T1 | jq .", "", ""},
	{"ls src 2> /dev/null", "", ""},
	{"cat .vloop/tmp/a > .vloop/tmp/b", "", ""},
	{"env git status", "", ""},
	{"sh -c 'git status'", "", ""},
	{"/usr/bin/git log -1", "", ""},
	{"grep '$(' src", "", ""},
}

func TestShell(t *testing.T) {
	for _, phase := range []string{PhaseReview, PhaseSecondReview} {
		j, _, _ := newJudge(t, phase)
		for _, c := range shellCases {
			for _, tool := range []string{"bash", "powershell"} {
				d := j.Decide(tool, "", c.cmd)
				switch {
				case c.deny == "" && !d.Allow:
					t.Errorf("[%s] %s %q: denied: %s", phase, tool, c.cmd, d.Reason)
				case c.deny != "" && d.Allow:
					t.Errorf("[%s] %s %q: allowed", phase, tool, c.cmd)
				case c.deny != "" && c.deny != "*" && !strings.Contains(d.Reason, c.deny):
					t.Errorf("[%s] %s %q: reason %q does not name %q", phase, tool, c.cmd, d.Reason, c.deny)
				}
			}
		}
	}
}

func TestOtherTools(t *testing.T) {
	for _, phase := range []string{PhaseReview, PhaseSecondReview} {
		j, root, home := newJudge(t, phase)
		type call struct {
			tool, path string
			want       bool
			reason     string
		}
		calls := []call{
			{"create", ".vloop/tmp/verdict.json", true, ""},
			{"edit", filepath.Join(root, ".vloop", "tmp", "abs.json"), true, ""},
			{"create", ".vloop/state/state.json", false, ""},
			{"edit", ".vloop/tmp/../state/x", false, ""},
			{"edit", "../outside", false, ""},
			{"edit", filepath.Join(root, "src", "app.ts"), false, ""},
			{"create", ".git/config", false, ""},
			{"create", ".vloop/tmp", false, ""},
			{"view", "src/app.ts", true, ""},
			{"read_file", ".vloop/state/state.json", true, ""},
			{"grep", "", true, ""},
			{"view", "~/.claude/settings.json", false, ""},
			{"glob", filepath.Join(home, ".copilot", "config.json"), false, ""},
			{"rg", filepath.Join(home, ".config", "gh", "hosts.yml"), false, ""},
			{"file_search", filepath.Join(home, ".config", "ghx"), true, ""},
			{"skill", "review", true, ""},
			{"skill", "verify", true, ""},
			{"skill", "vloop-review", false, "skill vloop-review is not vloop's — use /vloop:review"},
			{"web_fetch", "", false, ""},
			{"task", "", false, ""},
			{"explore_subagent", "", false, ""},
			{"frobnicate", "", false, "tool frobnicate is not allowed in a " + phase + " session"},
		}
		for _, c := range calls {
			d := j.Decide(c.tool, c.path, "")
			if d.Allow != c.want {
				t.Errorf("[%s] %s %q: allow = %v (%s), want %v", phase, c.tool, c.path, d.Allow, d.Reason, c.want)
			}
			if c.reason != "" && d.Reason != c.reason {
				t.Errorf("[%s] %s %q: reason %q, want %q", phase, c.tool, c.path, d.Reason, c.reason)
			}
		}
	}
}

func TestApplyPatch(t *testing.T) {
	j, _, _ := newJudge(t, PhaseReview)
	ok := "*** Begin Patch\n*** Add File: .vloop/tmp/a.txt\n+x\n*** End Patch"
	bad := ok + "\n*** Update File: src/a.go\n*** Move to: .vloop/tmp/b.go"
	if d := j.Decide("apply_patch", "", ok); !d.Allow {
		t.Errorf("patch under .vloop/tmp denied: %s", d.Reason)
	}
	if d := j.Decide("apply_patch", "", bad); d.Allow {
		t.Error("patch updating src/a.go allowed")
	}
	if d := j.Decide("apply_patch", "", "*** Begin Patch\n*** End Patch"); d.Allow {
		t.Error("patch with no file allowed")
	}
}

// W1, V1: the work and verify phases are served; any other name is not.
func TestUnknownPhase(t *testing.T) {
	_, err := New("plan", t.TempDir(), "")
	if err == nil || err.Error() != `unknown phase "plan": want review, second-review, work, verify or gate-review` {
		t.Errorf("New(plan) error = %v", err)
	}
}

func TestWorkShell(t *testing.T) {
	j, _, _ := newJudge(t, PhaseWork)
	cases := []struct{ cmd, deny string }{
		{"rm src/old.ts", ""},
		{"pnpm add nanoid", ""},
		{"npm install", ""},
		{"node x.js", ""},
		{"git diff > .vloop/tmp/diff.txt", ""},
		{"echo ok > src/version.txt", ""},
		{"ls src > /dev/null", ""},
		{"rm -rf dist", "rm -rf dist is denied by the work fence (Bash(rm -rf:*))"},
		{"rm -r dist", "(Bash(rm -r:*))"},
		{"rm -R dist", "(Bash(rm -R:*))"},
		{"rm -fr dist", "(Bash(rm -fr:*))"},
		{`sh -c "git push origin main"`, "git push origin main is denied by the work fence (Bash(git push:*))"},
		{"env gh pr list", "gh pr list is denied by the Copilot deny list (gh:*)"},
		{"curl https://example.com", "curl https://example.com is denied by the work fence (Bash(curl:*))"},
		{"git commit -m x", "denied by the work fence (Bash(git commit:*))"},
		{"pip install x", "(Bash(pip install:*))"},
		{"sh script.sh", "*"},
		{"echo x > .vloop/state/state.json", "write to .vloop/state/state.json is not allowed in a work session: nothing under .git/ or .vloop/ but .vloop/tmp/"},
		{"echo x > .git/config", "nothing under .git/ or .vloop/ but .vloop/tmp/"},
		{"echo x > ../outside.txt", "write to ../outside.txt is not allowed in a work session: it is outside the repository"},
		{"for f in a b; do echo $f; done", "*"},
		{"echo $(date)", "*"},
		{"echo `date`", "*"},
		{"eval ls", "*"},
	}
	for _, c := range cases {
		d := j.Decide("bash", "", c.cmd)
		switch {
		case c.deny == "" && !d.Allow:
			t.Errorf("%q: denied: %s", c.cmd, d.Reason)
		case c.deny != "" && d.Allow:
			t.Errorf("%q: allowed", c.cmd)
		case c.deny != "" && c.deny != "*" && !strings.Contains(d.Reason, c.deny):
			t.Errorf("%q: reason %q does not hold %q", c.cmd, d.Reason, c.deny)
		}
	}
}

func TestWorkOtherTools(t *testing.T) {
	j, root, _ := newJudge(t, PhaseWork)
	const dotDirs = "write to %s is not allowed in a work session: nothing under .git/ or .vloop/ but .vloop/tmp/"
	calls := []struct {
		tool, path string
		want       bool
		reason     string
	}{
		{"create", "src/links/links.service.ts", true, ""},
		{"edit", "test/app.e2e-spec.ts", true, ""},
		{"create", ".vloop/tmp/proposal.json", true, ""},
		{"edit", filepath.Join(root, "src", "app.ts"), true, ""},
		{"create", "/dev/null", true, ""},
		{"create", ".vloop/state/state.json", false, fmt.Sprintf(dotDirs, ".vloop/state/state.json")},
		{"edit", ".git/config", false, fmt.Sprintf(dotDirs, ".git/config")},
		{"create", ".vloop/config.toml", false, fmt.Sprintf(dotDirs, ".vloop/config.toml")},
		{"edit", ".vloop/tmp/../state/x", false, fmt.Sprintf(dotDirs, ".vloop/tmp/../state/x")},
		{"create", "/etc/hosts", false, "write to /etc/hosts is not allowed in a work session: it is outside the repository"},
		{"edit", "../outside", false, "write to ../outside is not allowed in a work session: it is outside the repository"},
		{"view", "src/app.ts", true, ""},
		{"view", "~/.copilot/config.json", false, ""},
		{"skill", "work", true, ""},
		{"skill", "vloop-work", false, "skill vloop-work is not vloop's — use /vloop:work"},
		{"web_fetch", "", false, "tool web_fetch is not allowed in a work session"},
		{"task", "", false, "tool task is not allowed in a work session"},
		{"explore_subagent", "", false, "tool explore_subagent is not allowed in a work session"},
	}
	for _, c := range calls {
		d := j.Decide(c.tool, c.path, "")
		if d.Allow != c.want {
			t.Errorf("%s %q: allow = %v (%s), want %v", c.tool, c.path, d.Allow, d.Reason, c.want)
		}
		if c.reason != "" && d.Reason != c.reason {
			t.Errorf("%s %q: reason %q, want %q", c.tool, c.path, d.Reason, c.reason)
		}
	}
	if d := j.Decide("apply_patch", "", "*** Begin Patch\n*** Update File: src/app.ts\n*** End Patch"); !d.Allow {
		t.Errorf("patch of src/app.ts denied: %s", d.Reason)
	}
	if d := j.Decide("apply_patch", "", "*** Begin Patch\n*** Add File: .vloop/defects/D1.md\n*** End Patch"); d.Allow {
		t.Error("patch adding .vloop/defects/D1.md allowed")
	}
}

// The review phases keep denying rm: only the work phase drops that stem.
func TestReviewStillDeniesRm(t *testing.T) {
	for _, phase := range []string{PhaseReview, PhaseSecondReview} {
		j, _, _ := newJudge(t, phase)
		d := j.Decide("bash", "", "rm src/old.ts")
		if d.Allow || d.Reason != "rm src/old.ts is denied by the Copilot deny list (rm:*)" {
			t.Errorf("[%s] rm src/old.ts: %+v", phase, d)
		}
	}
}

// V1: the verify phase decides as review, reads fence/verify.json and names
// verify in its messages.
func TestVerifyPhase(t *testing.T) {
	j, root, _ := newJudge(t, PhaseVerify)
	const vw = " is not allowed in a verify session: only .vloop/tmp/ may be written"
	patchAdd := "*** Begin Patch\n*** Add File: .vloop/tmp/v/x.txt\n+x\n*** End Patch\n"
	patchSrc := "*** Begin Patch\n*** Update File: src/app.go\n@@\n-a\n+b\n*** End Patch\n"
	cases := []struct {
		tool, path, cmd string
		want            string // "" allows
	}{
		{"bash", "", "go build -o .vloop/tmp/v/vloop ./cmd/vloop", ""},
		{"powershell", "", "go build -o .vloop/tmp/v/vloop ./cmd/vloop", ""},
		{"bash", "", "git init .vloop/tmp/v/ex1", ""},
		{"bash", "", "./.vloop/tmp/v/vloop --version > .vloop/tmp/v/out.txt", ""},
		{"bash", "", "vloop schema validate verify/v1 .vloop/tmp/verify.json", ""},
		{"bash", "", "cd .vloop/tmp/v && ./vloop status > out.txt", "write to out.txt" + vw},
		{"bash", "", "echo x > README.md", "write to README.md" + vw},
		{"bash", "", "rm .vloop/tmp/v/out.txt", "rm .vloop/tmp/v/out.txt is denied by the Copilot deny list (rm:*)"},
		{"bash", "", "git -C .vloop/tmp/v/ex1 status", "git -C .vloop/tmp/v/ex1 status is denied by the verify fence (Bash(git -C:*))"},
		{"bash", "", "vloop run", "vloop run is denied by the verify fence (Bash(vloop run:*))"},
		{"bash", "", `sh -c "git push origin main"`, "git push origin main is denied by the verify fence (Bash(git push:*))"},
		{"bash", "", "env gh pr list", "gh pr list is denied by the Copilot deny list (gh:*)"},
		{"bash", "", "echo $(date)", "command substitution cannot be judged"},
		{"create", ".vloop/tmp/verify.json", "", ""},
		{"edit", "internal/x.go", "", "write to internal/x.go" + vw},
		{"create", ".vloop/state/state.json", "", "write to .vloop/state/state.json" + vw},
		{"view", "src/app.go", "", ""},
		{"view", "~/.copilot/config.json", "", "read of ~/.copilot/config.json is not allowed: it is under ~/.claude, ~/.copilot or ~/.config/gh"},
		{"skill", "verify", "", ""},
		{"skill", "vloop-verify", "", "skill vloop-verify is not vloop's — use /vloop:verify"},
		{"web_fetch", "", "", "tool web_fetch is not allowed in a verify session"},
		{"apply_patch", "", patchAdd, ""},
		{"apply_patch", "", patchSrc, "write to src/app.go" + vw},
	}
	for _, c := range cases {
		d := j.Decide(c.tool, c.path, c.cmd)
		if c.want == "" {
			if !d.Allow {
				t.Errorf("%s %q %q denied: %s", c.tool, c.path, c.cmd, d.Reason)
			}
			continue
		}
		if d.Allow || d.Reason != c.want {
			t.Errorf("%s %q %q = allow %v, %q; want deny %q", c.tool, c.path, c.cmd, d.Allow, d.Reason, c.want)
		}
	}
	if d := j.Decide("create", filepath.Join(root, ".vloop", "tmp", "a.json"), ""); !d.Allow {
		t.Errorf("absolute path under .vloop/tmp denied: %s", d.Reason)
	}
}

// G1: the gate-review phase decides as review, reads fence/gate-review.json and
// names gate-review in its messages and skill hint.
func TestGateReviewPhase(t *testing.T) {
	j, root, _ := newJudge(t, PhaseGateReview)
	const gw = " is not allowed in a gate-review session: only .vloop/tmp/ may be written"
	patchAdd := "*** Begin Patch\n*** Add File: .vloop/tmp/gate-verdict.json\n+x\n*** End Patch\n"
	patchGate := "*** Begin Patch\n*** Update File: .vloop/state/gates/T1/check.sh\n@@\n-a\n+b\n*** End Patch\n"
	cases := []struct {
		tool, path, cmd string
		want            string // "" allows
	}{
		{"bash", "", "vloop task gate T1", ""},
		{"bash", "", "vloop task show T1 --json", ""},
		{"bash", "", "cat .vloop/state/gates/T1/check.sh", ""},
		{"bash", "", "printf 'hello\\n' | grep -E '^hello$'", ""},
		{"bash", "", "vloop schema validate gate-verdict/v2 .vloop/tmp/gate-verdict.json", ""},
		{"bash", "", "cd .vloop/tmp/g && ls > out.txt", "write to out.txt" + gw},
		{"bash", "", "echo x > .vloop/state/gates/T1/check.sh", "write to .vloop/state/gates/T1/check.sh" + gw},
		{"bash", "", "rm .vloop/tmp/g/out.txt", "rm .vloop/tmp/g/out.txt is denied by the Copilot deny list (rm:*)"},
		{"bash", "", "vloop task verify T1 'true' --reason x", "vloop task verify T1 true --reason x is denied by the gate-review fence (Bash(vloop task verify:*))"},
		{"bash", "", "git -C .vloop/tmp/g status", "git -C .vloop/tmp/g status is denied by the gate-review fence (Bash(git -C:*))"},
		{"bash", "", `sh -c "git push origin main"`, "git push origin main is denied by the gate-review fence (Bash(git push:*))"},
		{"bash", "", "env gh pr list", "gh pr list is denied by the Copilot deny list (gh:*)"},
		{"bash", "", "echo $(date)", "command substitution cannot be judged"},
		{"create", ".vloop/tmp/gate-verdict.json", "", ""},
		{"edit", ".vloop/state/state.json", "", "write to .vloop/state/state.json" + gw},
		{"create", "src/app.go", "", "write to src/app.go" + gw},
		{"view", ".vloop/state/gates/T1/check.sh", "", ""},
		{"view", "~/.copilot/config.json", "", "read of ~/.copilot/config.json is not allowed: it is under ~/.claude, ~/.copilot or ~/.config/gh"},
		{"skill", "gate-review", "", ""},
		{"skill", "vloop-gate-review", "", "skill vloop-gate-review is not vloop's — use /vloop:gate-review"},
		{"web_fetch", "", "", "tool web_fetch is not allowed in a gate-review session"},
		{"apply_patch", "", patchAdd, ""},
		{"apply_patch", "", patchGate, "write to .vloop/state/gates/T1/check.sh" + gw},
	}
	for _, c := range cases {
		d := j.Decide(c.tool, c.path, c.cmd)
		if c.want == "" {
			if !d.Allow {
				t.Errorf("%s %q %q denied: %s", c.tool, c.path, c.cmd, d.Reason)
			}
			continue
		}
		if d.Allow || d.Reason != c.want {
			t.Errorf("%s %q %q = allow %v, %q; want deny %q", c.tool, c.path, c.cmd, d.Allow, d.Reason, c.want)
		}
	}
	if d := j.Decide("create", filepath.Join(root, ".vloop", "tmp", "a.json"), ""); !d.Allow {
		t.Errorf("absolute path under .vloop/tmp denied: %s", d.Reason)
	}
}
