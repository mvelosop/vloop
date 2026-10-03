package driver

import (
	"slices"
	"testing"

	"github.com/mvelosop/vloop/internal/state"
)

func TestMatchGlob(t *testing.T) {
	for _, c := range []struct {
		glob, path string
		want       bool
	}{
		{"**", "a.go", true},
		{"**", "a/b/c.go", true},
		{"api/**", "api/x.go", true},
		{"api/**", "api/a/b/x.go", true},
		{"api/**", "web/x.go", false},
		{"api/**", "apix/x.go", false},
		{"**/*.go", "a/b/c.go", true},
		{"**/*.go", "c.go", true},
		{"**/*.go", "a/c.txt", false},
		{"*.go", "a/c.go", false},
		{"go.mod", "go.mod", true},
	} {
		if got := matchGlob(c.glob, c.path); got != c.want {
			t.Errorf("matchGlob(%q, %q) = %v, want %v", c.glob, c.path, got, c.want)
		}
	}
}

func TestChecksForKeepsConfigOrder(t *testing.T) {
	checks := []state.PlanCheck{
		{Name: "api", Paths: []string{"api/**"}},
		{Name: "web", Paths: []string{"web/**"}},
		{Name: "docs", Paths: []string{"**/*.md"}},
	}
	var names []string
	for _, c := range checksFor(checks, []string{"web/a.ts", "README.md"}) {
		names = append(names, c.Name)
	}
	if !slices.Equal(names, []string{"web", "docs"}) {
		t.Errorf("checksFor = %v, want [web docs]", names)
	}
	if got := checksFor(checks, nil); len(got) != 0 {
		t.Errorf("no changed path selected %v", got)
	}
}
