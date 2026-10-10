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
		{"Branch.*", "branch.x", false},
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
