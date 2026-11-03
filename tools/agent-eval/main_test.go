package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func invoke(args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	code := run(args, ".", &out, &errb)
	return code, out.String(), errb.String()
}

func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func fixtureCases(t *testing.T) string {
	root := filepath.Join(t.TempDir(), "cases")
	write(t, filepath.Join(root, "_lib.sh"), "x")
	write(t, filepath.Join(root, "b-review", "prompt.md"), "---\nmax_turns: 3\n---\n\n/vloop:review T1\n")
	write(t, filepath.Join(root, "b-review", "graders", "a.md"), "---\ntype: regex\ntarget:\n  source: file\n  path: x\n---\nbody\n")
	write(t, filepath.Join(root, "b-review", "graders", "b.md"), "---\ntype: llm\n---\nbody\n")
	write(t, filepath.Join(root, "a-op", "prompt.md"), "---\nmax_turns: 3\n---\n/vloop:operate\n")
	return root
}

func TestRefusalsInOrder(t *testing.T) {
	cases := fixtureCases(t)
	full := t.TempDir()
	write(t, filepath.Join(full, "keep"), "x")
	tmp := t.TempDir()
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"--cases", cases}, "--agent is required"},
		{[]string{"--agent", "claude", "--cases", cases}, "--model is required"},
		{[]string{"--agent", "claude", "--model", "m", "--cases", cases}, "--out is required"},
		{[]string{"--agent", "nope", "--cases", cases}, "--model is required"},
		{[]string{"--agent", "nope", "--model", "m", "--out", tmp + "/o", "--runs", "0", "--cases", cases}, "unknown agent nope"},
		{[]string{"--agent", "claude", "--model", "m", "--out", tmp + "/o", "--runs", "x", "--timeout", "q", "--cases", cases}, "--runs must be 1 or more"},
		{[]string{"--agent", "claude", "--model", "m", "--out", tmp + "/o", "--timeout", "5", "--cases", cases}, "--timeout 5 is not a duration"},
		{[]string{"--agent", "claude", "--model", "m", "--out", full, "--case", "zzz", "--cases", cases}, full + " is not empty"},
		{[]string{"--agent", "claude", "--model", "m", "--out", tmp + "/o", "--case", "zzz", "--cases", cases}, "no case matches"},
		{[]string{"--agent", "copilot", "--model", "m", "--out", tmp + "/o", "--cases", cases, "--vloop", tmp + "/none"}, "no vloop binary for the fence hook"},
	}
	for _, tc := range tests {
		code, out, errs := invoke(tc.args...)
		if code != 2 || out != "" || errs != "agent-eval: "+tc.want+"\n" {
			t.Errorf("%v: code %d, stdout %q, stderr %q; want 2, empty, %q", tc.args, code, out, errs, tc.want)
		}
	}
	if _, err := os.Stat(filepath.Join(tmp, "o")); err == nil {
		t.Error("a refusal created --out")
	}
}

func TestList(t *testing.T) {
	cases := fixtureCases(t)
	code, out, errs := invoke("--list", "--cases", cases)
	want := "CASE      PHASE   GRADERS\na-op      -       0\nb-review  review  2\n"
	if code != 0 || errs != "" || out != want {
		t.Fatalf("code %d, stderr %q, stdout %q; want %q", code, errs, out, want)
	}
	_, out, _ = invoke("--list", "--cases", cases, "--case", "b-*", "--case", "zz")
	if strings.Count(out, "\n") != 2 || !strings.Contains(out, "b-review") {
		t.Errorf("--case glob: %q", out)
	}
}

func TestUnreadableCases(t *testing.T) {
	cases := fixtureCases(t)
	write(t, filepath.Join(cases, "c-bad", "prompt.md"), "---\nk: v\n---\n/vloop:work T1\n")
	write(t, filepath.Join(cases, "c-bad", "graders", "x.md"), "---\ntype: regex\n")
	for _, args := range [][]string{{"--list"}, {"--agent", "claude", "--model", "m", "--out", filepath.Join(t.TempDir(), "o")}} {
		code, out, errs := invoke(append(args, "--cases", cases)...)
		if code != 1 || out != "" || !strings.HasPrefix(errs, "agent-eval: c-bad/graders/x.md: ") || strings.Count(errs, "\n") != 1 {
			t.Errorf("%v: code %d, stdout %q, stderr %q", args, code, out, errs)
		}
	}
	os.Remove(filepath.Join(cases, "c-bad", "prompt.md"))
	code, _, errs := invoke("--list", "--cases", cases)
	if code != 1 || !strings.HasPrefix(errs, "agent-eval: c-bad/prompt.md: ") {
		t.Errorf("no prompt.md: code %d, stderr %q", code, errs)
	}
}

func TestFrontmatter(t *testing.T) {
	text := "---\ntype: regex\nname: 'it''s'\nq: \"a\\sb\"\nplain: x\\d+ y: z\nlist: [a, b]\ntarget:\n  source: file\n  path: 'p'\n---\nbody"
	lines, body, err := splitFrontmatter(text)
	if err != nil || body != "body" {
		t.Fatal(err, body)
	}
	fm, err := parseFrontmatter(lines)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range map[string]string{"type": "regex", "name": "it's", "q": `a\sb`, "plain": `x\d+ y: z`, "list": "[a, b]"} {
		if fm.values[k] != v {
			t.Errorf("%s = %q, want %q", k, fm.values[k], v)
		}
	}
	if fm.maps["target"]["source"] != "file" || fm.maps["target"]["path"] != "p" {
		t.Errorf("map: %v", fm.maps)
	}
}

func TestFrontmatterErrors(t *testing.T) {
	for name, text := range map[string]string{
		"never closed":    "---\ntype: x\n",
		"none":            "just text\n",
		"not a key":       "---\njust text\n---\n",
		"orphan indent":   "---\n  a: b\n---\n",
		"deep indent":     "---\nm:\n    a: b\n---\n",
		"unclosed single": "---\na: 'x\n---\n",
		"stray single":    "---\na: 'x'y'\n---\n",
		"unclosed double": "---\na: \"x\n---\n",
		"map entry bare":  "---\nm:\n  a:\n---\n",
		"repeated key":    "---\na: 1\na: 2\n---\n",
		"no space":        "---\na:b\n---\n",
		"tab indent":      "---\n\ta: b\n---\n",
	} {
		lines, _, err := splitFrontmatter(text)
		if err == nil {
			_, err = parseFrontmatter(lines)
		}
		if err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}
