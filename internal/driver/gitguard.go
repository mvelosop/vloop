package driver

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/gitcfg"
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
// as the driver left them before a session or a gate ran. The config is held as
// its own entries, minus the keys matching the ignore patterns.
type gitGuard struct {
	root   string
	ignore []string
	config map[string][]string // key -> its values, sorted
	hooks  string
}

// snapshotGit takes the guard, resolving the ignore patterns once: the built-in
// one and run.git-config-ignore. A pattern source that cannot be read leaves
// the built-in one alone; the run refuses such a source before it starts.
func snapshotGit(root string) gitGuard {
	ignore := []string{gitcfg.BuiltinIgnore}
	if v, err := config.Get(root, "run.git-config-ignore"); err == nil {
		ignore = append(ignore, v.List...)
	}
	return snapshotGitWith(root, ignore)
}

func snapshotGitWith(root string, ignore []string) gitGuard {
	return gitGuard{root: root, ignore: ignore,
		config: configEntries(gitPath(root, "config"), ignore),
		hooks:  hashPath(gitPath(root, "hooks"))}
}

// maxNamedKeys is how many changed keys a halt names before "and n more".
const maxNamedKeys = 3

// changed names what differs now: ".git/config (<keys>)" or "the git hooks"; ""
// when neither moved.
func (g gitGuard) changed() string {
	now := snapshotGitWith(g.root, g.ignore)
	if keys := diffEntries(g.config, now.config); len(keys) > 0 {
		named := keys
		more := ""
		if len(keys) > maxNamedKeys {
			named, more = keys[:maxNamedKeys], fmt.Sprintf(" and %d more", len(keys)-maxNamedKeys)
		}
		return ".git/config (" + strings.Join(named, ", ") + more + ")"
	}
	if now.hooks != g.hooks {
		return "the git hooks"
	}
	return ""
}

// unparsedKey stands for a config git cannot read, held by a digest of its bytes.
const unparsedKey = "(unreadable)"

// configEntries reads the file's own entries — keys and values, include
// directives as entries, the included files never — through git, leaving out
// keys an ignore pattern matches. A missing file has none.
func configEntries(path string, ignore []string) map[string][]string {
	out := map[string][]string{}
	if path == "" {
		return out
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return out
	}
	list, err := exec.Command("git", "config", "--file", path, "--no-includes", "--list", "-z").Output()
	if err != nil {
		sum := sha256.Sum256(raw)
		out[unparsedKey] = []string{hex.EncodeToString(sum[:])}
		return out
	}
	for _, e := range strings.Split(string(list), "\x00") {
		if e == "" {
			continue
		}
		key, val, _ := strings.Cut(e, "\n")
		if gitcfg.Ignored(ignore, key) {
			continue
		}
		out[key] = append(out[key], val)
	}
	for _, vals := range out {
		slices.Sort(vals)
	}
	return out
}

// diffEntries returns the sorted keys whose values differ between two configs.
func diffEntries(before, after map[string][]string) []string {
	var keys []string
	for k, v := range before {
		if !slices.Equal(v, after[k]) {
			keys = append(keys, k)
		}
	}
	for k := range after {
		if _, ok := before[k]; !ok {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)
	return keys
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
