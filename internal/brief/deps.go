package brief

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Dir is where vloop finds briefs, relative to the repo root.
const Dir = "docs/briefs"

// Entry is one loop brief in the briefs directory.
type Entry struct {
	Name      string // filename minus .md
	Path      string // repo-root-relative, slash-separated
	Status    string
	DependsOn []string
	Ready     string // "ready", "blocked", or "-" once consumed
}

// Load reads every *.loop-brief.md directly under docs/briefs, sorted by name.
// Other files are invisible.
func Load(root string) ([]Entry, error) {
	files, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(Dir)))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Entry
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), Suffix) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(Dir), f.Name()))
		if err != nil {
			return nil, err
		}
		fm := parseFrontmatter(string(data))
		out = append(out, Entry{
			Name:      strings.TrimSuffix(f.Name(), ".md"),
			Path:      Dir + "/" + f.Name(),
			Status:    fm.Status,
			DependsOn: append([]string{}, fm.DependsOn...),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// graphProblems reports the dangling names and the first cycle reachable from
// start, whose own dependencies are deps. Names resolve against entries. A
// cycle is reported from start unless cycleOnly trims it to the loop.
func graphProblems(entries []Entry, start string, deps []string, cycleOnly bool) []string {
	byName := map[string][]string{}
	for _, e := range entries {
		byName[e.Name] = e.DependsOn
	}
	byName[start] = deps
	var out []string
	seen := map[string]bool{}
	var walk func(n string, stack []string) bool
	walk = func(n string, stack []string) bool {
		for _, d := range byName[n] {
			if _, ok := byName[d]; !ok {
				if !seen[d] {
					seen[d] = true
					out = append(out, "depends-on does not resolve: "+d)
				}
				continue
			}
			for i, s := range stack {
				if s == d {
					path := stack
					if cycleOnly {
						path = stack[i:] // the loop itself, not the way into it
					}
					out = append(out, "depends-on cycle: "+strings.Join(append(path[:len(path):len(path)], d), " -> "))
					return true
				}
			}
			if walk(d, append(stack, d)) {
				return true
			}
		}
		return false
	}
	walk(start, []string{start})
	return out
}

// dependencyLines is the depends-on rule of `brief check`.
func dependencyLines(root string, b *Brief) []Line {
	if len(b.fm.DependsOn) == 0 {
		return nil
	}
	entries, err := Load(root)
	if err != nil {
		return []Line{{Problem, fmt.Sprintf("cannot read %s: %v", Dir, err)}}
	}
	probs := graphProblems(entries, strings.TrimSuffix(pathBase(b.Path), ".md"), b.fm.DependsOn, false)
	if len(probs) == 0 {
		return []Line{{Pass, "depends-on resolves"}}
	}
	out := make([]Line, len(probs))
	for i, p := range probs {
		out[i] = Line{Problem, p}
	}
	return out
}

// Order validates the whole briefs graph and returns the entries in dependency
// order (a brief follows its dependencies, ties by lowest name) with Ready
// filled in. The error is the first problem found, worded as in `brief check`.
func Order(entries []Entry) ([]Entry, error) {
	for _, e := range entries {
		if p := graphProblems(entries, e.Name, e.DependsOn, true); len(p) > 0 {
			return nil, fmt.Errorf("%s", p[0])
		}
	}
	status := map[string]string{}
	for _, e := range entries {
		status[e.Name] = e.Status
	}
	done := map[string]bool{}
	var out []Entry
	for len(out) < len(entries) {
		next := -1
		for i, e := range entries { // entries are sorted by name
			if done[e.Name] {
				continue
			}
			ok := true
			for _, d := range e.DependsOn {
				ok = ok && done[d]
			}
			if ok {
				next = i
				break
			}
		}
		e := entries[next]
		done[e.Name] = true
		e.Ready = "ready"
		for _, d := range e.DependsOn {
			if status[d] != "consumed" {
				e.Ready = "blocked"
			}
		}
		if e.Status == "consumed" {
			e.Ready = "-"
		}
		out = append(out, e)
	}
	return out, nil
}
