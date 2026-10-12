// Package budget holds the word budgets of the plugin's skills (brief L10) and
// the one word count the metrics instructions line shares with them.
package budget

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Skills lists the plugin skills in the order Check reports them.
var Skills = []string{"plan", "gate-review", "work", "review", "operate"}

// Ceilings is each skill's word ceiling. Each was set on 2026-10-12 by brief
// B20261011-2251-b14-learning-loop (L10), at the size the skills had after its
// task 3 plus a little headroom, so that growth by drift becomes a recorded
// decision: a failing test, and an edit of the number and of this comment.
var Ceilings = map[string]int{
	"plan":        4100,
	"gate-review": 1550,
	"work":        1800,
	"review":      1650,
	"operate":     1750,
}

// CountWords counts the whitespace-separated words of data, as wc -w does.
func CountWords(data []byte) int {
	return len(strings.Fields(string(data)))
}

// FileWords counts the words of the whole file at path, frontmatter included.
func FileWords(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return CountWords(data), nil
}

// Check counts <pluginDir>/skills/<skill>/SKILL.md for each skill and returns
// one message per skill over its ceiling; a skill exactly at it is not
// reported. A skill that cannot be read is reported too.
func Check(pluginDir string) []string {
	var out []string
	for _, skill := range Skills {
		n, err := FileWords(filepath.Join(pluginDir, "skills", skill, "SKILL.md"))
		if err != nil {
			out = append(out, fmt.Sprintf("%s: %v", skill, err))
			continue
		}
		if c := Ceilings[skill]; n > c {
			out = append(out, fmt.Sprintf("%s: %s words, over its ceiling of %s", skill, Sep(n), Sep(c)))
		}
	}
	return out
}

// Sep formats n with thousands separators.
func Sep(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0 && s[i-1] != '-'; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
