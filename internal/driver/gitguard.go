package driver

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// noHooksDir is the empty directory every driver git command points
// core.hooksPath at, so no repository hook runs on the driver's own commits.
const noHooksDir = tmpDir + "/nohooks"

// gitCmd builds a git command for the driver: hooks and fsmonitor off, paths
// unquoted. The hooks directory is recreated empty each time, because it lives
// where a session can write.
func gitCmd(root string, args ...string) *exec.Cmd {
	dir := filepath.Join(root, filepath.FromSlash(noHooksDir))
	_ = os.RemoveAll(dir)
	_ = os.MkdirAll(dir, 0o755)
	full := append([]string{"-C", root,
		"-c", "core.hooksPath=" + dir,
		"-c", "core.fsmonitor=false",
		"-c", "core.quotepath=false"}, args...)
	return exec.Command("git", full...)
}

// gitGuard is the state of .git/config and of the hooks directory git names,
// as the driver left them before a session or a gate ran.
type gitGuard struct {
	root          string
	config, hooks string
}

func snapshotGit(root string) gitGuard {
	return gitGuard{root: root, config: hashPath(gitPath(root, "config")), hooks: hashPath(gitPath(root, "hooks"))}
}

// changed names what differs now: ".git/config" or "the git hooks"; "" when
// neither moved.
func (g gitGuard) changed() string {
	now := snapshotGit(g.root)
	switch {
	case now.config != g.config:
		return ".git/config"
	case now.hooks != g.hooks:
		return "the git hooks"
	}
	return ""
}

// gitPath is where git keeps the named file or directory; plain git, since the
// driver's own overrides would answer for the hooks.
func gitPath(root, name string) string {
	out, err := exec.Command("git", "-C", root, "rev-parse", "--git-path", name).Output()
	if err != nil {
		return ""
	}
	p := strings.TrimSpace(string(out))
	if p != "" && !filepath.IsAbs(p) {
		p = filepath.Join(root, p)
	}
	return p
}

// hashPath digests a file, or a directory tree by names, modes and contents.
func hashPath(p string) string {
	if p == "" {
		return ""
	}
	h := sha256.New()
	_ = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
		rel, _ := filepath.Rel(p, path)
		h.Write([]byte(filepath.ToSlash(rel) + "\x00"))
		if err != nil {
			h.Write([]byte("missing\x00"))
			return nil
		}
		if info, err := d.Info(); err == nil {
			h.Write([]byte(info.Mode().String() + "\x00"))
		}
		if d.Type().IsRegular() {
			if data, err := os.ReadFile(path); err == nil {
				h.Write(data)
			}
		}
		return nil
	})
	return hex.EncodeToString(h.Sum(nil))
}
