// Package budget holds the word budgets of the plugin's skills (brief L10) and
// the one word count the metrics instructions line shares with them.
package budget

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Skills lists the plugin skills in the order Check reports them.
var Skills = []string{"plan", "gate-review", "work", "review", "operate", "verify"}

// Ceilings is each skill's word ceiling, so that growth by drift becomes a
// recorded decision: a failing test, and an edit of the number and of its
// comment, which says when and why the ceiling was last set.
var Ceilings = map[string]int{
	"plan":        4100, // 2026-10-12, B14 (L10): its size after B14's task 3, plus headroom
	"gate-review": 1550, // 2026-10-12, B14 (L10): its size after B14's task 3, plus headroom
	"work":        1800, // 2026-10-12, B14 (L10): its size after B14's task 3, plus headroom
	"review":      1820, // 2026-10-26, B20 (J4): 1650 (2026-10-12, B14, L10) plus the 166 words of "The product" section and the reworded path question
	"operate":     1750, // 2026-10-12, B14 (L10): its size after B14's task 3, plus headroom
	"verify":      1500, // 2026-10-15, B16 (V2, L10): the ceiling set with the skill
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

var guidelineID = regexp.MustCompile(`\b[DI][0-9]{8}-[0-9]{4}(?:[0-9]{2})?`)

// GuidelineRules returns the 1-based line numbers of the rules in a guidelines
// file — lines starting `- ` outside HTML comments — and those among them
// that cite no defect or intervention ID.
func GuidelineRules(text string) (rules, noID []int) {
	inComment := false
	for i, l := range strings.Split(text, "\n") {
		l = strings.TrimRight(l, "\r")
		isRule := !inComment && strings.HasPrefix(l, "- ")
		// Drop comment spans from the rest of the line, tracking an open one.
		var kept strings.Builder
		rest := l
		for rest != "" {
			if inComment {
				_, after, found := strings.Cut(rest, "-->")
				if !found {
					rest = ""
					break
				}
				inComment = false
				rest = after
				continue
			}
			before, after, found := strings.Cut(rest, "<!--")
			kept.WriteString(before)
			if !found {
				break
			}
			inComment = true
			rest = after
		}
		if !isRule {
			continue
		}
		rules = append(rules, i+1)
		if !guidelineID.MatchString(kept.String()) {
			noID = append(noID, i+1)
		}
	}
	return rules, noID
}
