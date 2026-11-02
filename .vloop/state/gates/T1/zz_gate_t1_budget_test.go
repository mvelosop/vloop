package budget

// The T1 gate's oracle (B28, K1): the skill list, the ceilings, and a skill
// padded one word past its ceiling reported in the brief's words. Copied into a
// copy of the tree by the gate; never committed.

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGateT1BudgetLists(t *testing.T) {
	wantSkills := []string{"plan", "gate-review", "work", "review", "operate", "verify", "create-brief", "review-brief"}
	if !reflect.DeepEqual(Skills, wantSkills) {
		t.Errorf("Skills = %q, want %q", Skills, wantSkills)
	}
	wantCeilings := map[string]int{
		"plan": 4100, "gate-review": 1550, "work": 1800, "review": 1820,
		"operate": 1750, "verify": 1500, "create-brief": 1200, "review-brief": 1000,
	}
	if !reflect.DeepEqual(Ceilings, wantCeilings) {
		t.Errorf("Ceilings = %v, want %v", Ceilings, wantCeilings)
	}
	if got := Check(filepath.Join("..", "..", "plugin")); len(got) != 0 {
		t.Errorf("the plugin's skills are over budget: %q", got)
	}
}

func TestGateT1BudgetPadded(t *testing.T) {
	src := filepath.Join("..", "..", "plugin", "skills")
	for _, c := range []struct {
		skill string
		words int
		want  string
	}{
		{"create-brief", 1201, "create-brief: 1,201 words, over its ceiling of 1,200"},
		{"review-brief", 1001, "review-brief: 1,001 words, over its ceiling of 1,000"},
	} {
		dir := t.TempDir()
		ents, err := os.ReadDir(src)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range ents {
			if !e.IsDir() {
				continue
			}
			data, err := os.ReadFile(filepath.Join(src, e.Name(), "SKILL.md"))
			if err != nil {
				t.Fatalf("%s: %v", e.Name(), err)
			}
			if e.Name() == c.skill {
				n := len(strings.Fields(string(data)))
				if n > c.words {
					t.Fatalf("%s has %d words, more than the %d the gate pads it to", c.skill, n, c.words)
				}
				data = append(data, []byte("\n"+strings.Repeat("pad ", c.words-n)+"\n")...)
			}
			if err := os.MkdirAll(filepath.Join(dir, "skills", e.Name()), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "skills", e.Name(), "SKILL.md"), data, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		got := Check(dir)
		if len(got) != 1 || got[0] != c.want {
			t.Errorf("%s padded to %d words: Check = %q, want [%q]", c.skill, c.words, got, c.want)
		}
	}
}
