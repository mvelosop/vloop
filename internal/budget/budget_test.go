package budget

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPluginSkillsWithinBudget fails with each skill that grew past its
// ceiling; raising one is an edit of Ceilings and its comment.
func TestPluginSkillsWithinBudget(t *testing.T) {
	for _, m := range Check(filepath.Join("..", "..", "plugin")) {
		t.Error(m)
	}
}

func TestCheck(t *testing.T) {
	dir := t.TempDir()
	write := func(skill string, words int) {
		d := filepath.Join(dir, "skills", skill)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		body := "---\nname: x\n---\n" + strings.Repeat("w \t\n\n", words-1)
		if err := os.WriteFile(filepath.Join(d, "SKILL.md"), []byte(body+"w"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, s := range Skills {
		write(s, Ceilings[s]-4) // the header is 4 words: ---, name:, x, ---
	}
	if got := Check(dir); len(got) != 0 {
		t.Fatalf("at ceiling: %v", got)
	}
	write("plan", 4213-4)
	want := "plan: 4,213 words, over its ceiling of 4,100"
	if got := Check(dir); len(got) != 1 || got[0] != want {
		t.Fatalf("got %v, want [%s]", got, want)
	}
}

func TestSep(t *testing.T) {
	for n, want := range map[int]string{0: "0", 999: "999", 1000: "1,000", 4213: "4,213", 1234567: "1,234,567"} {
		if got := Sep(n); got != want {
			t.Errorf("Sep(%d) = %s, want %s", n, got, want)
		}
	}
}
