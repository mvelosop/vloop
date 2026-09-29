package defect

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mvelosop/vloop/internal/runs"
)

// ErrUnattributable is returned by Blame when the line cannot be tied to a
// loop brief.
type ErrUnattributable struct {
	File string
	Line int
}

func (e *ErrUnattributable) Error() string {
	return fmt.Sprintf("cannot attribute %s:%d to a loop brief — pass --brief", e.File, e.Line)
}

// Attribution says which brief a line belongs to and how that was found.
type Attribution struct {
	Brief string // the brief name
	How   string // "trailer" or "consumed"
	SHA7  string
}

// String is the stderr line.
func (a Attribution) String() string {
	how := "trailer on"
	if a.How == "consumed" {
		how = "consumed in"
	}
	return fmt.Sprintf("attributed to %s (%s %s)", a.Brief, how, a.SHA7)
}

func gitOut(root string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(errb.String()))
	}
	return out.String(), nil
}

// Blame attributes file:line, as it is on the default branch (not HEAD), to a
// loop brief: the last commit's Vloop-Brief trailer, else the brief whose
// frontmatter became `status: consumed` in that commit.
func Blame(root, file string, line int) (*Attribution, error) {
	un := &ErrUnattributable{File: file, Line: line}
	branch, err := runs.DefaultBranch(root)
	if err != nil {
		return nil, err
	}
	out, err := gitOut(root, "blame", "--porcelain", "-L", fmt.Sprintf("%d,%d", line, line), branch, "--", file)
	if err != nil {
		return nil, un
	}
	fields := strings.Fields(out)
	if len(fields) == 0 {
		return nil, un
	}
	sha := fields[0]
	sha7 := sha
	if len(sha7) > 7 {
		sha7 = sha7[:7]
	}
	tr, err := gitOut(root, "show", "-s", "--format=%(trailers:key=Vloop-Brief,valueonly,separator=%x0a)", sha)
	if err == nil {
		for _, l := range strings.Split(tr, "\n") {
			if name := BriefName(strings.TrimSpace(l)); name != "" {
				return &Attribution{Brief: name, How: "trailer", SHA7: sha7}, nil
			}
		}
	}
	files, err := gitOut(root, "diff-tree", "--root", "-r", "--no-commit-id", "--name-only", "-m", "--first-parent", sha)
	if err != nil {
		return nil, un
	}
	parent := ""
	if p, err := gitOut(root, "rev-parse", "--verify", "--quiet", sha+"^"); err == nil {
		parent = strings.TrimSpace(p)
	}
	for _, f := range strings.Split(files, "\n") {
		if !strings.HasPrefix(f, BriefsDir+"/") || !strings.HasSuffix(f, ".loop-brief.md") || strings.Count(f, "/") != 2 {
			continue
		}
		if runs.StatusAt(root, sha, f) != "consumed" {
			continue
		}
		if parent != "" && runs.StatusAt(root, parent, f) == "consumed" {
			continue
		}
		return &Attribution{Brief: BriefName(f), How: "consumed", SHA7: sha7}, nil
	}
	return nil, un
}
