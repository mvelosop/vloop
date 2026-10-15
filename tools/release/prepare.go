package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/mvelosop/vloop/internal/config"
)

// refusal is a failed check: its message is printed as "release: <msg>" and
// the run exits with code.
type refusal struct {
	code int
	msg  string
	out  string
}

func (r *refusal) Error() string { return r.msg }

func refuse(code int, format string, a ...any) *refusal {
	return &refusal{code: code, msg: fmt.Sprintf(format, a...)}
}

var versionRE = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)

// semver is a parsed version: vMAJOR.MINOR.PATCH with an optional pre-release.
type semver struct {
	num [3]int
	pre []string
}

func parseVersion(s string) (semver, bool) {
	m := versionRE.FindStringSubmatch(s)
	if m == nil {
		return semver{}, false
	}
	var v semver
	for i := 0; i < 3; i++ {
		n, err := strconv.Atoi(m[i+1])
		if err != nil {
			return semver{}, false
		}
		v.num[i] = n
	}
	if m[4] != "" {
		v.pre = strings.Split(m[4], ".")
	}
	return v, true
}

// cmp orders versions by semver precedence: a pre-release sorts before its release.
func (a semver) cmp(b semver) int {
	for i := 0; i < 3; i++ {
		if a.num[i] != b.num[i] {
			return sign(a.num[i] - b.num[i])
		}
	}
	switch {
	case len(a.pre) == 0 && len(b.pre) == 0:
		return 0
	case len(a.pre) == 0:
		return 1
	case len(b.pre) == 0:
		return -1
	}
	for i := 0; i < len(a.pre) && i < len(b.pre); i++ {
		x, y := a.pre[i], b.pre[i]
		xn, xerr := strconv.Atoi(x)
		yn, yerr := strconv.Atoi(y)
		switch {
		case xerr == nil && yerr == nil:
			if xn != yn {
				return sign(xn - yn)
			}
		case xerr == nil:
			return -1
		case yerr == nil:
			return 1
		case x != y:
			return sign(strings.Compare(x, y))
		}
	}
	return sign(len(a.pre) - len(b.pre))
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	}
	return 0
}

// modulePath and retractions read go.mod under dir.
func readGoMod(dir string) (module string, retracted [][2]string, err error) {
	raw, err := os.ReadFile(dir + "/go.mod")
	if err != nil {
		return "", nil, err
	}
	inBlock := false
	for _, line := range strings.Split(string(raw), "\n") {
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var spec string
		switch {
		case inBlock && line == ")":
			inBlock = false
			continue
		case inBlock:
			spec = line
		case strings.HasPrefix(line, "module "):
			module = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "module ")), `"`)
			continue
		case strings.HasPrefix(line, "retract ("), line == "retract(":
			inBlock = true
			continue
		case strings.HasPrefix(line, "retract "):
			spec = strings.TrimSpace(strings.TrimPrefix(line, "retract "))
		default:
			continue
		}
		spec = strings.TrimSpace(spec)
		if strings.HasPrefix(spec, "[") && strings.HasSuffix(spec, "]") {
			parts := strings.Split(spec[1:len(spec)-1], ",")
			if len(parts) == 2 {
				retracted = append(retracted, [2]string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])})
			}
			continue
		}
		retracted = append(retracted, [2]string{spec, spec})
	}
	return module, retracted, nil
}

func isRetracted(v semver, ranges [][2]string) bool {
	for _, r := range ranges {
		lo, ok1 := parseVersion(r[0])
		hi, ok2 := parseVersion(r[1])
		if ok1 && ok2 && v.cmp(lo) >= 0 && v.cmp(hi) <= 0 {
			return true
		}
	}
	return false
}

// gitOut runs git in dir and returns its trimmed stdout.
func gitOut(dir string, args ...string) (string, error) {
	out, err := git(dir, args...)
	return strings.TrimSpace(out), err
}

// shellCmd builds the command that runs script in the platform shell.
func shellCmd(script string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("pwsh", "-NoProfile", "-Command", script)
	}
	return exec.Command("sh", "-c", script)
}

// runChecked runs cmd in dir; on failure it returns the combined output.
func runChecked(dir string, cmd *exec.Cmd) (string, error) {
	var buf bytes.Buffer
	cmd.Dir = dir
	cmd.Stdout, cmd.Stderr = &buf, &buf
	err := cmd.Run()
	return buf.String(), err
}

// prepareChecks makes every check of `prepare <version>`, in the brief's
// order, and writes nothing to the repository. The one thing it touches is
// the remote-tracking ref of origin/main, by the fetch.
func prepareChecks(dir, version string) (latest string, ref *refusal) {
	v, ok := parseVersion(version)
	if !ok {
		return latest, refuse(2, "%s is not a version: vX.Y.Z or vX.Y.Z-<pre>", version)
	}
	module, retracted, err := readGoMod(dir)
	if err != nil {
		return latest, refuse(1, "%s", oneLine(err.Error()))
	}
	if v.num[0] >= 2 && !strings.HasSuffix(module, "/v"+strconv.Itoa(v.num[0])) {
		return latest, refuse(1, "%s needs the module path to end in /v%d, and %s does not", version, v.num[0], module)
	}
	if isRetracted(v, retracted) {
		return latest, refuse(1, "%s is retracted in go.mod", version)
	}
	tags, err := gitOut(dir, "tag", "--list")
	if err != nil {
		return latest, refuse(1, "%s", oneLine(err.Error()))
	}
	var latestV semver
	for _, t := range strings.Fields(tags) {
		if tv, ok := parseVersion(t); ok && (latest == "" || tv.cmp(latestV) > 0) {
			latest, latestV = t, tv
		}
	}
	if latest != "" && v.cmp(latestV) <= 0 {
		return latest, refuse(1, "%s is not after the latest release %s", version, latest)
	}
	if local, _ := gitOut(dir, "tag", "--list", version); local != "" {
		return latest, refuse(1, "tag %s already exists", version)
	}
	remote, err := gitOut(dir, "ls-remote", "--tags", "origin", "refs/tags/"+version)
	if err != nil {
		return latest, refuse(1, "%s", oneLine(err.Error()))
	}
	if remote != "" {
		return latest, refuse(1, "tag %s already exists", version)
	}

	branch, err := gitOut(dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return latest, refuse(1, "%s", oneLine(err.Error()))
	}
	if branch != "main" {
		return latest, refuse(1, "prepare runs on main, not %s", branch)
	}
	status, err := gitOut(dir, "status", "--porcelain")
	if err != nil {
		return latest, refuse(1, "%s", oneLine(err.Error()))
	}
	if status != "" {
		return latest, refuse(1, "the working tree is not clean")
	}
	if _, err := gitOut(dir, "fetch", "--quiet", "origin", "main"); err != nil {
		return latest, refuse(1, "%s", oneLine(err.Error()))
	}
	head, err1 := gitOut(dir, "rev-parse", "HEAD")
	origin, err2 := gitOut(dir, "rev-parse", "refs/remotes/origin/main")
	if err1 != nil || err2 != nil || head != origin {
		return latest, refuse(1, "main is not origin/main — pull or push first")
	}

	checks, err := config.Checks(dir)
	if err != nil {
		return latest, refuse(1, "%s", oneLine(err.Error()))
	}
	for _, c := range checks {
		if out, err := runChecked(dir, shellCmd(c.Run)); err != nil {
			r := refuse(1, "check %s failed", c.Name)
			r.out = out
			return latest, r
		}
	}
	if out, err := runChecked(dir, exec.Command("claude", "plugin", "validate", "plugin")); err != nil {
		r := refuse(1, "claude plugin validate plugin failed")
		r.out = out
		return latest, r
	}
	return latest, nil
}

// prepare checks the release, and with dryRun says what it would do.
func prepare(dir string, args []string, stdout, stderr io.Writer) int {
	var version string
	dryRun := false
	for _, a := range args {
		switch {
		case a == "--dry-run":
			dryRun = true
		case version == "" && !strings.HasPrefix(a, "-"):
			version = a
		default:
			fmt.Fprintln(stderr, "usage: release prepare <version> [--dry-run]")
			return 2
		}
	}
	if version == "" {
		fmt.Fprintln(stderr, "usage: release prepare <version> [--dry-run]")
		return 2
	}
	latest, r := prepareChecks(dir, version)
	if r != nil {
		if r.out != "" {
			fmt.Fprint(stderr, r.out)
			if !strings.HasSuffix(r.out, "\n") {
				fmt.Fprintln(stderr)
			}
		}
		fmt.Fprintln(stderr, "release: "+r.msg)
		return r.code
	}
	if dryRun {
		fmt.Fprintf(stdout, "dry run: every check passed for %s; nothing written\n", version)
		fmt.Fprintf(stdout, "would create release/%s, set plugin.json's version, commit \"Release %s\", push it and open a PR\n", version, version)
		return 0
	}
	url, err := openRelease(dir, version, latest)
	if err != nil {
		fmt.Fprintln(stderr, "release: "+oneLine(err.Error()))
		return 1
	}
	fmt.Fprintf(stdout, "prepared release/%s: %s\n", version, url)
	fmt.Fprintf(stdout, "squash-merge it, then: go run ./tools/release publish %s\n", version)
	return 0
}

var pluginVersionRE = regexp.MustCompile(`("version"\s*:\s*")[^"]*(")`)

// openRelease makes release/<version> from main with the one commit that sets
// plugin.json's version, pushes it and opens the PR; it leaves main checked
// out. It never touches main, tags or merges.
func openRelease(dir, version, latest string) (url string, err error) {
	from := latest
	if from == "" {
		roots, err := gitOut(dir, "rev-list", "--max-parents=0", "main")
		if err != nil {
			return "", err
		}
		from = strings.Fields(roots)[0]
	}
	body, err := notes(dir, from, "main")
	if err != nil {
		return "", err
	}
	branch := "release/" + version
	if _, err := git(dir, "switch", "-q", "-c", branch, "main"); err != nil {
		return "", err
	}
	defer git(dir, "switch", "-q", "main")
	path := dir + "/plugin/.claude-plugin/plugin.json"
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	done := false
	out := pluginVersionRE.ReplaceAllFunc(raw, func(m []byte) []byte {
		if done {
			return m
		}
		done = true
		return pluginVersionRE.ReplaceAll(m, []byte("${1}"+strings.TrimPrefix(version, "v")+"${2}"))
	})
	if !done {
		return "", fmt.Errorf("plugin.json has no version")
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return "", err
	}
	if _, err := git(dir, "commit", "-q", "-m", "Release "+version, "--", "plugin/.claude-plugin/plugin.json"); err != nil {
		return "", err
	}
	if _, err := git(dir, "push", "-q", "-u", "origin", branch); err != nil {
		return "", err
	}
	cmd := exec.Command("gh", "pr", "create", "--base", "main", "--head", branch, "--title", "Release "+version, "--body", body)
	cmd.Dir = dir
	var so, se bytes.Buffer
	cmd.Stdout, cmd.Stderr = &so, &se
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gh pr create: %s %s", se.String(), err)
	}
	return strings.TrimSpace(so.String()), nil
}
