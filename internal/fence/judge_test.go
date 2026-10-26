package fence

import (
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

func TestUnknownPhase(t *testing.T) {
	if _, err := New("work", t.TempDir(), ""); err == nil {
		t.Error("New accepted the work phase")
	}
}
