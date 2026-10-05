package runs

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/mvelosop/vloop/internal/frontmatter"
)

// DefaultBranch resolves the branch a release lands on: origin/HEAD's target,
// else main, else master. It returns a ref name usable with git log and blame.
func DefaultBranch(root string) (string, error) {
	if b, err := git(root, "symbolic-ref", "--quiet", "refs/remotes/origin/HEAD"); err == nil {
		ref := strings.TrimSpace(string(b))
		if ref != "" && refExists(root, ref) {
			return ref, nil
		}
	}
	for _, name := range []string{"main", "master"} {
		if refExists(root, "refs/heads/"+name) {
			return name, nil
		}
	}
	return "", fmt.Errorf("no default branch: origin/HEAD, main and master are all missing")
}

func refExists(root, ref string) bool {
	return exec.Command("git", "-C", root, "rev-parse", "--verify", "--quiet", ref+"^{commit}").Run() == nil
}

// StatusAt reports the `status` in the frontmatter of the file at path as
// committed at sha; "" when the file, the frontmatter or the key is missing.
func StatusAt(root, sha, path string) string {
	b, err := git(root, "show", sha+":"+path)
	if err != nil {
		return ""
	}
	return frontmatterStatus(string(b))
}

func frontmatterStatus(text string) string {
	doc, err := frontmatter.Parse(text)
	if err != nil {
		return ""
	}
	for _, f := range doc.Fields() {
		if f.Key == "status" {
			v, _, _ := strings.Cut(f.Value, " #")
			return strings.Trim(strings.TrimSpace(v), `"'`)
		}
	}
	return ""
}

// Release returns the full SHA of the first first-parent commit on the default
// branch where the brief's frontmatter says `status: consumed`, or "" when it
// has not been released. briefPath is repo-relative. It errors only when there
// is no default branch.
func Release(root, briefPath string) (string, error) {
	branch, err := DefaultBranch(root)
	if err != nil {
		return "", err
	}
	b, err := git(root, "log", "--first-parent", "--reverse", "--format=%H", branch, "--", briefPath)
	if err != nil {
		return "", err
	}
	for _, sha := range strings.Fields(string(b)) {
		if StatusAt(root, sha, briefPath) == "consumed" {
			return sha, nil
		}
	}
	return "", nil
}

// FrontmatterStatus reads `status:` from a document's frontmatter; "" if none.
func FrontmatterStatus(text string) string { return frontmatterStatus(text) }
