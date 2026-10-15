// Command release is this repository's release tool, run as `go run ./tools/release`.
// It is a repository-owned program, not a vloop subcommand.
//
//	release notes <from> <to>   the release notes for the range from..to, on stdout
//	release prepare <version>   check a release; --dry-run writes nothing
//	release publish <version>   tag the merged release commit and create the GitHub release
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	os.Exit(run(os.Args[1:], ".", os.Stdout, os.Stderr))
}

// run executes one invocation against the repository at dir and returns the
// exit code: 0 done, 1 refused or failed, 2 usage.
func run(args []string, dir string, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "prepare" {
		return prepare(dir, args[1:], stdout, stderr)
	}
	if len(args) > 0 && args[0] == "publish" {
		return publish(dir, args[1:], stdout, stderr)
	}
	if len(args) == 0 || args[0] != "notes" {
		fmt.Fprintln(stderr, "usage: release notes <from> <to> | prepare <version> [--dry-run] | publish <version>")
		return 2
	}
	if len(args) != 3 {
		fmt.Fprintln(stderr, "usage: release notes <from> <to>")
		return 2
	}
	out, err := notes(dir, args[1], args[2])
	if err != nil {
		fmt.Fprintln(stderr, "release: "+oneLine(err.Error()))
		return 1
	}
	fmt.Fprint(stdout, out)
	return 0
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var out, errb strings.Builder
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", args[0], msg)
	}
	return out.String(), nil
}

// Control characters that cannot occur in a subject or a trailer value.
const (
	fieldSep  = "\x1f"
	recordSep = "\x1e"
)

type commit struct {
	hash, subject, brief string
}

// notes renders the release notes of the commits after from, up to and
// including to.
func notes(dir, from, to string) (string, error) {
	for _, rev := range []string{from, to} {
		if _, err := git(dir, "rev-parse", "--verify", "--quiet", rev+"^{commit}"); err != nil {
			return "", fmt.Errorf("unknown revision %q", rev)
		}
	}
	format := "--format=%H" + fieldSep + "%s" + fieldSep +
		"%(trailers:key=Vloop-Brief,valueonly,unfold)" + recordSep
	raw, err := git(dir, "log", "--reverse", format, from+".."+to)
	if err != nil {
		return "", err
	}
	var commits []commit
	for _, rec := range strings.Split(raw, recordSep) {
		rec = strings.Trim(rec, "\n")
		if rec == "" {
			continue
		}
		f := strings.SplitN(rec, fieldSep, 3)
		if len(f) != 3 {
			return "", fmt.Errorf("unreadable git log record")
		}
		c := commit{hash: f[0], subject: f[1]}
		for _, line := range strings.Split(f[2], "\n") {
			if line = strings.TrimSpace(line); line != "" {
				c.brief = line
				break
			}
		}
		commits = append(commits, c)
	}

	var briefs, others []string
	seen := map[string]bool{}
	for _, c := range commits {
		if c.brief != "" {
			if seen[c.brief] {
				continue
			}
			seen[c.brief] = true
			briefs = append(briefs, briefLine(dir, to, c.brief))
			continue
		}
		if strings.HasPrefix(c.subject, "Release v") {
			continue
		}
		files, err := git(dir, "diff-tree", "--no-commit-id", "--name-only", "-r", "--root", "-m", c.hash)
		if err != nil {
			return "", err
		}
		if touchesCode(files) {
			others = append(others, "- "+c.subject)
		}
	}

	var sections []string
	if len(briefs) > 0 {
		sections = append(sections, "## Briefs\n\n"+strings.Join(briefs, "\n")+"\n")
	}
	if len(others) > 0 {
		sections = append(sections, "## Other changes\n\n"+strings.Join(others, "\n")+"\n")
	}
	if len(sections) == 0 {
		return "No changes.\n", nil
	}
	return strings.Join(sections, "\n"), nil
}

// touchesCode reports whether any listed file is outside docs/, .vloop/ and CLAUDE.md.
func touchesCode(files string) bool {
	for _, f := range strings.Split(files, "\n") {
		if f == "" {
			continue
		}
		if strings.HasPrefix(f, "docs/") || strings.HasPrefix(f, ".vloop/") || f == "CLAUDE.md" {
			continue
		}
		return true
	}
	return false
}

// briefLine is the brief's line, its title and description read at rev.
func briefLine(dir, rev, name string) string {
	title := name
	desc := ""
	src, err := git(dir, "show", rev+":docs/briefs/"+name+".md")
	if err == nil {
		var heading string
		heading, desc = parseBrief(src)
		if heading != "" {
			title = heading
		}
	}
	if desc == "" {
		return "- **" + title + "**"
	}
	return "- **" + title + "** — " + desc
}

// parseBrief returns the first heading, cleaned of a leading "Brief — " and
// "vloop ", and the frontmatter description.
func parseBrief(src string) (heading, desc string) {
	lines := strings.Split(src, "\n")
	i := 0
	if len(lines) > 0 && strings.TrimRight(lines[0], "\r") == "---" {
		for i = 1; i < len(lines); i++ {
			l := strings.TrimRight(lines[i], "\r")
			if l == "---" {
				i++
				break
			}
			if desc == "" && strings.HasPrefix(l, "description: ") {
				desc = strings.TrimSpace(strings.TrimPrefix(l, "description: "))
			}
		}
	}
	for ; i < len(lines); i++ {
		l := strings.TrimRight(lines[i], "\r")
		if strings.HasPrefix(l, "# ") {
			heading = strings.TrimSpace(strings.TrimPrefix(l, "# "))
			heading = strings.TrimPrefix(heading, "Brief — ")
			heading = strings.TrimPrefix(heading, "vloop ")
			break
		}
	}
	return heading, desc
}
