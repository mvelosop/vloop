package install

import (
	"bytes"
	"strings"
)

// Markers bracket vloop's section of CLAUDE.md; upgrade replaces what is
// between them.
const (
	Begin = "<!-- vloop:begin -->"
	End   = "<!-- vloop:end -->"
)

// IgnoreLine is the .gitignore line vloop's scratch space needs.
const IgnoreLine = ".vloop/tmp/"

// Section is vloop's CLAUDE.md section for version, markers included, ending
// with a newline.
func Section(version string) string {
	return Begin + `
## vloop

Sessions started by the vloop loop (plan, gate review, work, review) follow these rules:

1. Use repo-relative paths only, in files, logs and commit messages.
2. A work session does one task and stops; it does not start the next one.
3. Sessions never commit, set a task status or move git refs; the driver does.
4. A task is done only when its gate passes and the review passes it.
5. Halting cleanly with an account of what blocked you is a success; faking progress is the only failure.

In an interactive session you are the operator's hands: follow the operator's playbook (/vloop:operate), not these session rules.

Written by vloop ` + version + `.
` + End + "\n"
}

// MergeClaudeMD returns CLAUDE.md with vloop's section: the text between the
// markers replaced when both are present, the section appended after the
// existing text when they are not, the section alone when there is no file.
// Everything outside the markers is kept byte for byte.
func MergeClaudeMD(existing []byte, exists bool, version string) []byte {
	sec := Section(version)
	if !exists {
		return []byte(sec)
	}
	s := string(existing)
	if b := strings.Index(s, Begin); b >= 0 {
		if e := strings.Index(s[b:], End); e >= 0 {
			e += b
			return []byte(s[:b] + strings.TrimSuffix(sec, "\n") + s[e+len(End):])
		}
	}
	if s != "" && !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	if s != "" {
		s += "\n"
	}
	return []byte(s + sec)
}

// MergeGitignore returns .gitignore holding IgnoreLine once, on a line of its
// own, changing nothing else.
func MergeGitignore(existing []byte) (out []byte, changed bool) {
	for _, l := range strings.Split(string(existing), "\n") {
		if strings.TrimSpace(l) == IgnoreLine {
			return existing, false
		}
	}
	out = bytes.Clone(existing)
	if len(out) > 0 && out[len(out)-1] != '\n' {
		out = append(out, '\n')
	}
	return append(out, IgnoreLine+"\n"...), true
}
