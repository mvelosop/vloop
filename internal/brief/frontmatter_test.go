package brief

import (
	"strings"
	"testing"
)

func TestFrontmatterRules(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(string) string
		want   string
	}{
		{"missing", func(s string) string { return s[strings.Index(s, "# Brief"):] }, "missing frontmatter"},
		{"unclosed list", func(s string) string { return strings.Replace(s, "depends-on: []", "depends-on: [unclosed", 1) }, "unparseable frontmatter"},
		{"no closing", func(s string) string { return strings.Replace(s, "---\n# Brief", "# Brief", 1) }, "unparseable frontmatter"},
		{"no name", func(s string) string { return strings.Replace(s, "name: B20260101-0900-a.loop-brief\n", "", 1) }, "frontmatter has no name"},
		{"name mismatch", func(s string) string { return strings.Replace(s, "name: pass", "name: other", 1) }, "name other.loop-brief does not match filename B20260101-0900-a.loop-brief"},
		{"bad status", func(s string) string { return strings.Replace(s, "status: ready", "status: bogus", 1) }, `status "bogus" is not one of draft|ready|consumed|abandoned`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root, path, text := scratch(t)
			text = c.mutate(strings.Replace(text, "name: pass.loop-brief", "name: B20260101-0900-a.loop-brief", 1))
			if c.name == "name mismatch" {
				text = strings.Replace(text, "name: B20260101-0900-a.loop-brief", "name: pass.loop-brief", 1)
				text = c.mutate(text)
			}
			res := Check(root, Parse(path, text), SetFor("en"))
			if !has(res.Problems(), c.want) || !res.Failed() {
				t.Fatalf("want problem %q, got %v", c.want, res.Lines)
			}
		})
	}
}

func TestFrontmatterAppliesToDraftsAndToleratesOptionalKeys(t *testing.T) {
	root, path, text := scratch(t)
	text = strings.Replace(text, "name: pass.loop-brief", "name: B20260101-0900-a.loop-brief", 1)
	draft := strings.Replace(text, "status: ready", "status: draft", 1)
	bad := strings.Replace(draft, "name: B20260101-0900-a", "name: B20260101-0901-a", 1)
	if res := Check(root, Parse(path, bad), SetFor("en")); res.Skipped || !res.Failed() {
		t.Fatalf("a draft with a wrong name must fail: %+v", res)
	}
	if res := Check(root, Parse(path, draft), SetFor("en")); !res.Skipped {
		t.Fatalf("a valid draft is skipped: %+v", res)
	}
	lean := text
	for _, k := range []string{"description: Build a\n", "created: 2026-01-01\n", "seeds: A run that builds a\n"} {
		lean = strings.Replace(lean, k, "", 1)
	}
	lean = strings.Replace(lean, "kind: brief", "owner: someone", 1)
	if res := Check(root, Parse(path, lean), SetFor("en")); res.Failed() || len(res.Warnings()) != 0 {
		t.Fatalf("optional and unknown keys: %v", res.Lines)
	}
}
