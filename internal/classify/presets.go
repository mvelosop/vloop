// Package classify sorts repo-relative paths into code, test, docs, other or
// excluded, using the repo's own globs and the built-in stack presets.
package classify

import "sort"

// Preset is one stack's globs. Every preset's docs is Docs.
type Preset struct {
	Code     []string `json:"code"`
	Test     []string `json:"test"`
	Docs     []string `json:"docs"`
	Excluded []string `json:"excluded"`
}

var docs = []string{"**/*.md"}

var nodeExcluded = []string{"package-lock.json", "yarn.lock", "pnpm-lock.yaml", "**/dist/**", "**/node_modules/**"}

var presets = map[string]Preset{
	"csharp": {
		Code:     []string{"**/*.cs", "**/*.razor", "**/*.cshtml"},
		Test:     []string{"**/*.Tests/**", "**/*Tests.cs"},
		Excluded: []string{"**/bin/**", "**/obj/**", "**/*.Designer.cs", "**/packages.lock.json"},
	},
	"go": {
		Code:     []string{"**/*.go"},
		Test:     []string{"**/*_test.go", "**/testdata/**"},
		Excluded: []string{"go.sum", "vendor/**"},
	},
	"java": {
		Code:     []string{"**/*.java"},
		Test:     []string{"**/src/test/**"},
		Excluded: []string{"**/build/**", "**/target/**"},
	},
	"javascript": {
		Code:     []string{"**/*.js", "**/*.mjs", "**/*.cjs"},
		Test:     []string{"**/*.test.js", "**/*.spec.js", "**/__tests__/**", "**/*.e2e-spec.js", "**/test/**"},
		Excluded: nodeExcluded,
	},
	"kotlin": {
		Code:     []string{"**/*.kt", "**/*.kts"},
		Test:     []string{"**/src/test/**"},
		Excluded: []string{"**/build/**", "**/target/**"},
	},
	"python": {
		Code:     []string{"**/*.py"},
		Test:     []string{"**/test_*.py", "**/*_test.py", "**/tests/**"},
		Excluded: []string{"**/__pycache__/**", "poetry.lock", "uv.lock", "Pipfile.lock"},
	},
	"react": {
		Code:     []string{"**/*.tsx", "**/*.jsx", "**/*.css", "**/*.scss"},
		Test:     []string{"**/*.test.tsx", "**/*.spec.tsx", "**/*.test.jsx", "**/*.spec.jsx", "**/*.stories.*"},
		Excluded: nodeExcluded,
	},
	"rust": {
		Code:     []string{"**/*.rs"},
		Test:     []string{"**/tests/**", "**/benches/**"},
		Excluded: []string{"**/target/**", "Cargo.lock"},
	},
	"typescript": {
		Code:     []string{"**/*.ts"},
		Test:     []string{"**/*.test.ts", "**/*.spec.ts", "**/__tests__/**", "**/*.e2e-spec.ts", "**/test/**"},
		Excluded: nodeExcluded,
	},
}

// Names returns the preset names, sorted.
func Names() []string {
	names := make([]string, 0, len(presets))
	for n := range presets {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Lookup returns the named preset (with docs filled in) and whether it exists.
// The returned slices are copies.
func Lookup(name string) (Preset, bool) {
	p, ok := presets[name]
	if !ok {
		return Preset{}, false
	}
	return Preset{
		Code:     append([]string{}, p.Code...),
		Test:     append([]string{}, p.Test...),
		Docs:     append([]string{}, docs...),
		Excluded: append([]string{}, p.Excluded...),
	}, true
}
