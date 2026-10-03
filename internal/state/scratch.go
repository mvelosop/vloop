package state

import (
	"os"
	"path/filepath"
)

// EmptyScratch removes everything inside each gate scratch folder of the
// plan, keeping the folders themselves. A folder that does not exist is
// already empty. It returns the first error, after trying every folder.
func EmptyScratch(root string, dirs []string) error {
	var first error
	for _, d := range dirs {
		abs := filepath.Join(root, filepath.FromSlash(d))
		entries, err := os.ReadDir(abs)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if err := os.RemoveAll(filepath.Join(abs, e.Name())); err != nil && first == nil {
				first = err
			}
		}
	}
	return first
}
