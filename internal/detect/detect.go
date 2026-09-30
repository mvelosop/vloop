// Package detect finds the stacks a repository uses, from marker files, so
// vloop init can write metrics.stacks. It only reads.
package detect

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// maxDepth is how many directory levels below the root are examined.
const maxDepth = 3

var skipped = map[string]bool{
	".git": true, ".vloop": true, ".loop": true, "node_modules": true,
	"vendor": true, "bin": true, "obj": true, "dist": true, "target": true,
	"build": true,
}

// Stacks returns the entries init writes to metrics.stacks: stacks found at
// root unscoped, stacks found below it as <stack>@<dir> with "/". Unscoped
// entries come first, each group sorted. The result is empty when nothing is
// found.
func Stacks(root string) ([]string, error) {
	found := map[string]bool{}
	if err := walk(root, "", 0, false, found); err != nil {
		return nil, err
	}
	var unscoped, scoped []string
	for e := range found {
		if strings.Contains(e, "@") {
			scoped = append(scoped, e)
		} else {
			unscoped = append(unscoped, e)
		}
	}
	sort.Strings(unscoped)
	sort.Strings(scoped)
	return append(unscoped, scoped...), nil
}

func entry(stack, rel string) string {
	if rel == "" {
		return stack
	}
	return stack + "@" + rel
}

// walk examines the directory rel (slash-separated, "" for the root).
// underSln reports whether a *.sln sits in this directory's ancestry.
func walk(root, rel string, depth int, underSln bool, found map[string]bool) error {
	dir := filepath.Join(root, filepath.FromSlash(rel))
	ents, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var subdirs []string
	var hasSln, hasCsproj bool
	for _, e := range ents {
		name := e.Name()
		switch {
		case e.IsDir():
			// A symlink reports as ModeSymlink, not a directory: never followed.
			if !skipped[name] {
				subdirs = append(subdirs, name)
			}
			continue
		case !e.Type().IsRegular():
			continue
		}
		switch lower := strings.ToLower(name); {
		case lower == "go.mod":
			found[entry("go", rel)] = true
		case lower == "package.json":
			for _, s := range jsStacks(filepath.Join(dir, name), ents) {
				found[entry(s, rel)] = true
			}
		case strings.HasSuffix(lower, ".sln"):
			hasSln = true
		case strings.HasSuffix(lower, ".csproj"):
			hasCsproj = true
		case lower == "pyproject.toml", lower == "requirements.txt", lower == "setup.py":
			found[entry("python", rel)] = true
		case lower == "pom.xml", lower == "build.gradle":
			found[entry("java", rel)] = true
		case lower == "build.gradle.kts":
			found[entry("kotlin", rel)] = true
		case lower == "cargo.toml":
			found[entry("rust", rel)] = true
		}
	}
	if hasSln || (hasCsproj && !underSln) {
		found[entry("csharp", rel)] = true
	}
	if depth == maxDepth {
		return nil
	}
	for _, s := range subdirs {
		if err := walk(root, path.Join(rel, s), depth+1, underSln || hasSln, found); err != nil {
			return err
		}
	}
	return nil
}

// jsStacks reads the package.json at file; siblings are its directory listing.
// An unreadable or malformed package.json still marks javascript.
func jsStacks(file string, siblings []os.DirEntry) []string {
	out := []string{"javascript"}
	ts := false
	for _, e := range siblings {
		if strings.EqualFold(e.Name(), "tsconfig.json") && e.Type().IsRegular() {
			ts = true
		}
	}
	var pkg struct {
		Dependencies    map[string]json.RawMessage `json:"dependencies"`
		DevDependencies map[string]json.RawMessage `json:"devDependencies"`
	}
	react := false
	if b, err := os.ReadFile(file); err == nil && json.Unmarshal(b, &pkg) == nil {
		for _, deps := range []map[string]json.RawMessage{pkg.Dependencies, pkg.DevDependencies} {
			if _, ok := deps["typescript"]; ok {
				ts = true
			}
			if _, ok := deps["react"]; ok {
				react = true
			}
		}
	}
	if ts {
		out = append(out, "typescript")
	}
	if react {
		out = append(out, "react")
	}
	return out
}
