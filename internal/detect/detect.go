// Package detect finds the stacks a repository uses, from marker files, so
// vloop init can write metrics.stacks. It only reads.
package detect

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/mvelosop/vloop/internal/classify"
	"github.com/mvelosop/vloop/internal/config"
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

// starterRuns is each stack's usual test command.
var starterRuns = map[string]string{
	"csharp":     "dotnet test",
	"go":         "go test ./... && go vet ./...",
	"java":       "mvn test",
	"javascript": "npm test",
	"kotlin":     "gradle test",
	"python":     "pytest",
	"react":      "npm test",
	"rust":       "cargo test",
	"typescript": "npm test",
}

var nonName = regexp.MustCompile(`[^a-z0-9]+`)

// StarterChecks returns one check per entry of Stacks (<stack> or
// <stack>@<dir>): an unscoped stack applies to "**", a scoped one to
// "<dir>/**" and runs in <dir>. Stacks that share a directory and a command
// (javascript, typescript and react all run npm test) yield one check. The
// checks are starting points for the repository to edit.
func StarterChecks(entries []string) []config.CheckDef {
	var out []config.CheckDef
	seenName, seenRun := map[string]bool{}, map[string]bool{}
	for _, e := range entries {
		stack, scope, err := classify.ParseStack(e)
		if err != nil {
			continue
		}
		run, ok := starterRuns[stack]
		if !ok {
			continue
		}
		paths := []string{"**"}
		if scope != "" {
			run = "cd " + scope + " && " + run
			paths = []string{scope + "/**"}
		}
		if seenRun[run] {
			continue
		}
		seenRun[run] = true
		name := stack
		if scope != "" {
			name += "-" + strings.Trim(nonName.ReplaceAllString(strings.ToLower(scope), "-"), "-")
		}
		for base, n := name, 2; seenName[name]; n++ {
			name = base + "-" + strconv.Itoa(n)
		}
		seenName[name] = true
		out = append(out, config.CheckDef{Name: name, Paths: paths, Run: run})
	}
	return out
}

// FormatCheck is how init and upgrade show a check: name, command, paths.
func FormatCheck(c config.CheckDef) string {
	return c.Name + ": " + c.Run + " (paths: " + strings.Join(c.Paths, ", ") + ")"
}
