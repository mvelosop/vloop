package driver

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// ClearTmp empties .vloop/tmp/ under root of everything but the run lock, so no
// session's scratch outlives the session that follows it. A symbolic link is
// removed, never followed. An entry that cannot be removed is reported through
// warn and the clearing goes on. The plugin, the fence and the hooks folder are
// recreated before use, so clearing them is safe.
func ClearTmp(root string, warn func(entry string, err error)) {
	dir := filepath.Join(root, filepath.FromSlash(tmpDir))
	entries, err := os.ReadDir(dir)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) && warn != nil {
			warn("", err)
		}
		return
	}
	lock := filepath.Base(lockFile)
	for _, e := range entries {
		if e.Name() == lock {
			continue
		}
		// RemoveAll removes a symlink itself and never what it points at.
		p := filepath.Join(dir, e.Name())
		err := os.RemoveAll(p)
		if err != nil && warn != nil {
			var pe *fs.PathError
			if errors.As(err, &pe) {
				err = pe.Err
			}
			warn(e.Name(), err)
		}
	}
}
