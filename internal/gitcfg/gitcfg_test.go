package gitcfg

import "testing"

func TestMatch(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		pattern, key string
		want         bool
	}{
		{BuiltinIgnore, "branch.feat/x.vscode-merge-base", true},
		{BuiltinIgnore, "branch.a.b.vscode-merge-base", true},
		{BuiltinIgnore, "branch.main.remote", false},
		{BuiltinIgnore, "core.vscode-merge-base", false},
		{BuiltinIgnorePR, "branch.B1.github-pr-base-branch", true},
		{BuiltinIgnorePR, "branch.release/v1.github-pr-owner-number", true},
		{BuiltinIgnorePR, "core.github-pr-base-branch", false},
		{"gitlens.*", "gitlens.x.y", true},
		{"gitlens.*", "gitlens.", true},
		{"gitlens.*", "gitlen.x", false},
		{"core.fsmonitor", "core.fsmonitor", true},
		{"core.fsmonitor", "core.fsmonitor2", false},
		{"*", "anything.at.all", true},
		{"a*b*c", "axxbyyc", true},
		{"a*b*c", "axxcyyb", false},
		{"a*b", "ab", true},
		{"a*a", "a", false},
		{"Branch.*", "branch.x", true},
		{"Branch.X.*", "branch.x.y", false},
	} {
		if got := Match(c.pattern, c.key); got != c.want {
			t.Errorf("Match(%q, %q) = %v, want %v", c.pattern, c.key, got, c.want)
		}
	}
}

func TestProtected(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		pattern, section string
	}{
		{"*", "core"},
		{"c*", "core"},
		{"co", "core"},
		{"core.*", "core"},
		{"core.fsmonitor", "core"},
		{"filter.lfs.*", "filter"},
		{"f*", "filter"},
		{"i*", "include"},
		{"include.path", "include"},
		{"includeif.*", "includeif"},
		{"includeif.gitdir:x.path", "includeif"},
		{"gitlens.*", ""},
		{"branch.*.vscode-merge-base", ""},
		{"corex.*", ""},
		{"alias.*", ""},
	} {
		got, ok := Protected(c.pattern)
		if got != c.section || ok != (c.section != "") {
			t.Errorf("Protected(%q) = %q, %v; want %q", c.pattern, got, ok, c.section)
		}
	}
}

func TestIgnored(t *testing.T) {
	t.Parallel()
	if !Ignored([]string{"x.*", BuiltinIgnore}, "branch.m.vscode-merge-base") {
		t.Error("the built-in pattern did not match")
	}
	if Ignored(nil, "x.y") {
		t.Error("no patterns ignored a key")
	}
}

func TestNormalizeMixedCase(t *testing.T) {
	if _, ok := Protected("Core.*"); !ok {
		t.Error("Core.* not protected")
	}
	if _, ok := Protected("IncludeIf.*"); !ok {
		t.Error("IncludeIf.* not protected")
	}
	if !Match("Branch.Main.Foo-Bar", "branch.Main.foo-bar") {
		t.Error("mixed-case pattern does not match its own subsection")
	}
	if Match("Branch.Main.Foo-Bar", "branch.main.foo-bar") {
		t.Error("subsection case must be kept")
	}
	if !Match("branch.*.vscode-merge-base", "branch.x.vscode-merge-base") || Match("core.x", "core.y") {
		t.Error("lower-case patterns changed")
	}
}

func TestBuiltinIgnores(t *testing.T) {
	t.Parallel()
	for _, k := range []string{"branch.b.vscode-merge-base", "branch.b.github-pr-base-branch", "branch.b.github-pr-owner-number"} {
		if !Ignored(BuiltinIgnores, k) {
			t.Errorf("%s is not ignored by the built-in patterns", k)
		}
	}
	for _, k := range []string{"branch.b.remote", "branch.b.merge", "core.hookspath"} {
		if Ignored(BuiltinIgnores, k) {
			t.Errorf("%s is ignored by the built-in patterns", k)
		}
	}
}
