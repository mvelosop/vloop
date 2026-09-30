package classify

import "github.com/bmatcuk/doublestar/v4"

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
	name string
	p    Preset
}

// New builds a classifier from the repo's globs and the preset names of
// metrics.stacks. Unknown names are ignored (config validates them).
func New(repo Preset, stacks []string) *Classifier {
	c := &Classifier{repo: repo}
	for _, s := range stacks {
		if p, ok := Lookup(s); ok {
			c.presets = append(c.presets, namedPreset{s, p})
		}
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
	for _, s := range steps {
		for _, np := range c.presets {
			if g, ok := first(s.globs(np.p), path); ok {
				return Result{s.cat, np.name, g}
			}
		}
	}
	return Result{Category: Other}
}
