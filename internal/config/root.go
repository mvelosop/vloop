package config

import (
	"os"
	"path/filepath"
)

// FilePath is the config file's location relative to the repo root, with "/"
// separators on every OS.
const FilePath = ".vloop/config.toml"

// Root resolves the repo root from start: the nearest ancestor (start
// included) containing .vloop or .git, checking both at each level before
// moving up, else start (C-2).
func Root(start string) string {
	if abs, err := filepath.Abs(start); err == nil {
		start = abs
	}
	for dir := start; ; {
		for _, marker := range []string{".vloop", ".git"} {
			if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return start
		}
		dir = parent
	}
}
