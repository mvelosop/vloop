package classify

import (
	"fmt"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// Categories.
const (
	Code     = "code"
	Test     = "test"
	Docs     = "docs"
	Other    = "other"
	Excluded = "excluded"
)

// LayerAlways and LayerRepo name the two non-preset layers.
const (
	LayerAlways = "always"
	LayerRepo   = "repo"
)

var always = []string{".vloop/**", ".loop/**"}

// Result is the category of a path and the layer and glob that decided it.
// Layer and Glob are empty for Other.
type Result struct {
	Category string
	Layer    string
	Glob     string
}

// Classifier holds the repo's globs and the named presets, in order.
type Classifier struct {
	repo    Preset
	presets []namedPreset
}

type namedPreset struct {
	name  string
	scope string // repo-relative directory, "" when unscoped
	p     Preset
}

// label is how a match by this preset is named: <stack> or <stack>@<path>.
func (n namedPreset) label() string {
	if n.scope == "" {
		return n.name
	}
	return n.name + "@" + n.scope
}

// ParseStack splits a metrics.stacks entry, <stack> or <stack>@<path>, and
// checks its form: a known stack, and a path with "/" separators, no leading or
// trailing "/", no empty, "." or ".." segment. It does not look at the disk.
func ParseStack(entry string) (name, scope string, err error) {
	name, scope, scoped := strings.Cut(entry, "@")
	if _, ok := Lookup(name); !ok {
		return "", "", fmt.Errorf("unknown stack %q", name)
	}
	if !scoped {
		return name, "", nil
	}
	if scope == "" || strings.Contains(scope, "\\") {
		return "", "", fmt.Errorf("bad scope path %q", scope)
	}
	for _, seg := range strings.Split(scope, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return "", "", fmt.Errorf("bad scope path %q", scope)
		}
	}
	return name, scope, nil
}

// New builds a classifier from the repo's globs and the preset names of
// metrics.stacks, each <stack> or <stack>@<path>. Malformed entries are ignored
// (config validates them).
func New(repo Preset, stacks []string) *Classifier {
	c := &Classifier{repo: repo}
	for _, s := range stacks {
		name, scope, err := ParseStack(s)
		if err != nil {
			continue
		}
		p, _ := Lookup(name)
		c.presets = append(c.presets, namedPreset{name, scope, p})
	}
	return c
}

func first(globs []string, path string) (string, bool) {
	for _, g := range globs {
		if ok, err := doublestar.Match(g, path); err == nil && ok {
			return g, true
		}
	}
	return "", false
}

// Classify classifies a repo-relative path written with "/" separators.
func (c *Classifier) Classify(path string) Result {
	if g, ok := first(always, path); ok {
		return Result{Excluded, LayerAlways, g}
	}
	type step struct {
		cat   string
		globs func(Preset) []string
	}
	steps := []step{
		{Excluded, func(p Preset) []string { return p.Excluded }},
		{Test, func(p Preset) []string { return p.Test }},
		{Docs, func(p Preset) []string { return p.Docs }},
		{Code, func(p Preset) []string { return p.Code }},
	}
	for _, s := range steps {
		if g, ok := first(s.globs(c.repo), path); ok {
			return Result{s.cat, LayerRepo, g}
		}
	}
	scope := ""
	for _, np := range c.presets {
		if len(np.scope) > len(scope) && strings.HasPrefix(path, np.scope+"/") {
			scope = np.scope
		}
	}
	rel := path
	if scope != "" {
		rel = path[len(scope)+1:]
	}
	for _, s := range steps {
		for _, np := range c.presets {
			if np.scope != scope {
				continue
			}
			if g, ok := first(s.globs(np.p), rel); ok {
				return Result{s.cat, np.label(), g}
			}
		}
	}
	return Result{Category: Other}
}
