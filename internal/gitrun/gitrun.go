// Package gitrun is the one place vloop runs git: every other package builds
// its git commands here and keeps its own trimming and error text.
package gitrun

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Plain is git with the arguments as given: no -C and no overrides.
func Plain(args ...string) *exec.Cmd {
	return exec.Command("git", args...)
}

// At is plain git run in root (-C root).
func At(root string, args ...string) *exec.Cmd {
	return Plain(append([]string{"-C", root}, args...)...)
}

// Hardened is git run in root with hooks pointed at hooksDir, fsmonitor off and
// paths unquoted — the form for the driver's own commands. It recreates
// hooksDir empty each time, because it lives where a session can write.
func Hardened(root, hooksDir string, args ...string) *exec.Cmd {
	_ = os.RemoveAll(hooksDir)
	_ = os.MkdirAll(hooksDir, 0o755)
	return Plain(append([]string{"-C", root,
		"-c", "core.hooksPath=" + hooksDir,
		"-c", "core.fsmonitor=false",
		"-c", "core.quotepath=false"}, args...)...)
}

// Capture runs plain git in root and returns its stdout untrimmed; a failure
// reads "git <args>: <err>: <trimmed stderr>".
func Capture(root string, args ...string) ([]byte, error) {
	cmd := At(root, args...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(errb.String()))
	}
	return out.Bytes(), nil
}
