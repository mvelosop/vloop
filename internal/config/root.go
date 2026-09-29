package config

import (
	"os"
	"path/filepath"
)

// FilePath is the config file's location relative to the repo root, with "/"
// separators on every OS.
const FilePath = ".vloop/config.toml"

// Root resolves the repo root from start: the nearest ancestor (start
// included) containing .vloop, else the nearest containing .git, else start.
func Root(start string) string {
	if abs, err := filepath.Abs(start); err == nil {
		start = abs
	}
	for _, marker := range []string{".vloop", ".git"} {
		for dir := start; ; {
			if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
				return dir
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return start
}
