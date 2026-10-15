package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// proxyInterval is how long publish pauses between asks of the module proxy.
var proxyInterval = 30 * time.Second

const pluginJSON = "plugin/.claude-plugin/plugin.json"

// findReleaseCommit returns the newest commit on origin/main whose subject
// starts "Release <version>" and whose plugin.json holds the version.
func findReleaseCommit(dir, version string) (string, error) {
	raw, err := git(dir, "log", "--format=%H"+fieldSep+"%s"+recordSep, "refs/remotes/origin/main")
	if err != nil {
		return "", err
	}
	prefix := "Release " + version
	for _, rec := range strings.Split(raw, recordSep) {
		f := strings.SplitN(strings.Trim(rec, "\n"), fieldSep, 2)
		if len(f) != 2 || !strings.HasPrefix(f[1], prefix) {
			continue
		}
		if rest := f[1][len(prefix):]; rest != "" && !strings.ContainsAny(rest[:1], " (:") {
			continue
		}
		src, err := git(dir, "show", f[0]+":"+pluginJSON)
		if err != nil {
			continue
		}
		var p struct {
			Version string `json:"version"`
		}
		if json.Unmarshal([]byte(src), &p) == nil && p.Version == strings.TrimPrefix(version, "v") {
			return f[0], nil
		}
	}
	return "", nil
}

// remoteTags lists the tags on origin with the commit each points at.
func remoteTags(dir string) (map[string]string, error) {
	out, err := gitOut(dir, "ls-remote", "--tags", "origin")
	if err != nil {
		return nil, err
	}
	tags := map[string]string{}
	for _, l := range strings.Split(out, "\n") {
		f := strings.Fields(l)
		if len(f) != 2 || !strings.HasPrefix(f[1], "refs/tags/") {
			continue
		}
		name := strings.TrimPrefix(f[1], "refs/tags/")
		if strings.HasSuffix(name, "^{}") {
			tags[strings.TrimSuffix(name, "^{}")] = f[0]
		} else if _, ok := tags[name]; !ok {
			tags[name] = f[0]
		}
	}
	return tags, nil
}

// briefTitles lists the titles of the briefs consumed in from..to, in order.
func briefTitles(dir, from, to string) ([]string, error) {
	raw, err := git(dir, "log", "--reverse", "--format=%(trailers:key=Vloop-Brief,valueonly,unfold)"+recordSep, from+".."+to)
	if err != nil {
		return nil, err
	}
	var titles []string
	seen := map[string]bool{}
	for _, rec := range strings.Split(raw, recordSep) {
		name := strings.TrimSpace(rec)
		if i := strings.Index(name, "\n"); i >= 0 {
			name = strings.TrimSpace(name[:i])
		}
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		title := name
		if src, err := git(dir, "show", to+":docs/briefs/"+name+".md"); err == nil {
			if h, _ := parseBrief(src); h != "" {
				title = h
			}
		}
		titles = append(titles, title)
	}
	return titles, nil
}

// waitProxy asks `go list -m <module>@<version>` until it succeeds or wait has
// passed, outside the repository so its go.mod plays no part.
func waitProxy(module, version string, wait time.Duration) bool {
	deadline := time.Now().Add(wait)
	for {
		cmd := exec.Command("go", "list", "-m", module+"@"+version)
		cmd.Dir = os.TempDir()
		if err := cmd.Run(); err == nil {
			return true
		}
		left := time.Until(deadline)
		if left <= 0 {
			return false
		}
		time.Sleep(min(proxyInterval, left))
	}
}

// installCheck installs module/cmd/vloop@version into a temporary GOBIN and
// returns what the binary reports as version and module.
func installCheck(module, version string) (string, string, error) {
	tmp, err := os.MkdirTemp("", "release-install")
	if err != nil {
		return "", "", err
	}
	defer os.RemoveAll(tmp)
	bin := filepath.Join(tmp, "bin")
	cmd := exec.Command("go", "install", module+"/cmd/vloop@"+version)
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(), "GOBIN="+bin)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", "", fmt.Errorf("go install %s/cmd/vloop@%s: %s %w", module, version, out, err)
	}
	exe := filepath.Join(bin, "vloop")
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	vc := exec.Command(exe, "version", "--json")
	vc.Dir = tmp
	out, err := vc.Output()
	if err != nil {
		return "", "", fmt.Errorf("vloop version --json: %w", err)
	}
	var r struct {
		Version string `json:"version"`
		Module  string `json:"module"`
	}
	if err := json.Unmarshal(out, &r); err != nil {
		return "", "", fmt.Errorf("vloop version --json: %w", err)
	}
	return r.Version, r.Module, nil
}

// publish finds the release commit before any write, then tags it, pushes the
// tag, creates the GitHub release and proves the public install. Each step
// already done is skipped, so a re-run finishes an interrupted release.
func publish(dir string, args []string, stdout, stderr io.Writer) int {
	usage := func() int {
		fmt.Fprintln(stderr, "usage: release publish <version> [--wait <duration>] [--dry-run]")
		return 2
	}
	version, dryRun, wait := "", false, 10*time.Minute
	for i := 0; i < len(args); i++ {
		switch a := args[i]; {
		case a == "--dry-run":
			dryRun = true
		case a == "--wait" && i+1 < len(args):
			d, err := time.ParseDuration(args[i+1])
			if err != nil {
				return usage()
			}
			wait = d
			i++
		case strings.HasPrefix(a, "-") || version != "":
			return usage()
		default:
			version = a
		}
	}
	if version == "" {
		return usage()
	}
	v, ok := parseVersion(version)
	if !ok {
		fmt.Fprintf(stderr, "release: %s is not a version: vX.Y.Z or vX.Y.Z-<pre>\n", version)
		return 2
	}
	fail := func(format string, a ...any) int {
		fmt.Fprintln(stderr, "release: "+oneLine(fmt.Sprintf(format, a...)))
		return 1
	}
	if _, err := gitOut(dir, "fetch", "--quiet", "origin", "main"); err != nil {
		return fail("%s", err)
	}
	rc, err := findReleaseCommit(dir, version)
	if err != nil {
		return fail("%s", err)
	}
	if rc == "" {
		return fail("no release commit for %s on origin/main — merge the release PR first", version)
	}
	module, _, err := readGoMod(dir)
	if err != nil {
		return fail("%s", err)
	}
	tags, err := remoteTags(dir)
	if err != nil {
		return fail("%s", err)
	}
	var prev string
	var prevV semver
	for t := range tags {
		if tv, ok := parseVersion(t); ok && tv.cmp(v) < 0 && (prev == "" || tv.cmp(prevV) > 0) {
			prev, prevV = t, tv
		}
	}
	if sha, ok := tags[version]; ok && sha != rc {
		return fail("tag %s points at %s, not the release commit %s", version, sha, rc)
	}
	if _, err := git(dir, "fetch", "--quiet", "origin", "+refs/tags/*:refs/tags/*"); err != nil {
		return fail("%s", err)
	}
	from := prev
	if from == "" {
		roots, err := gitOut(dir, "rev-list", "--max-parents=0", rc)
		if err != nil {
			return fail("%s", err)
		}
		from = strings.Fields(roots)[0]
	}

	if _, ok := tags[version]; ok {
		fmt.Fprintln(stdout, "tag: already done")
	} else {
		titles, err := briefTitles(dir, from, rc)
		if err != nil {
			return fail("%s", err)
		}
		msg := version
		if len(titles) > 0 {
			msg += ": " + strings.Join(titles, "; ")
		}
		local, _ := gitOut(dir, "rev-parse", "-q", "--verify", "refs/tags/"+version+"^{commit}")
		if dryRun {
			fmt.Fprintf(stdout, "tag: would tag %s on %s and push it\n", version, rc)
		} else if local != "" && local != rc {
			return fail("tag %s points at %s, not the release commit %s", version, local, rc)
		} else {
			if local == "" {
				if _, err := git(dir, "tag", "-a", version, "-m", msg, rc); err != nil {
					return fail("%s", err)
				}
			}
			if _, err := git(dir, "push", "-q", "origin", "refs/tags/"+version); err != nil {
				return fail("%s", err)
			}
			fmt.Fprintf(stdout, "tag: %s on %s\n", version, rc)
		}
	}

	url, done := "", false
	if out, err := runChecked(dir, exec.Command("gh", "release", "view", version, "--json", "url", "--jq", ".url")); err == nil {
		url, done = strings.TrimSpace(out), true
		fmt.Fprintln(stdout, "release: already done")
	}
	if !done && dryRun {
		fmt.Fprintf(stdout, "release: would create the GitHub release %s\n", version)
	} else if !done {
		body, err := notes(dir, from, rc)
		if err != nil {
			return fail("%s", err)
		}
		cargs := []string{"release", "create", version, "--verify-tag", "--title", version, "--notes", body}
		if len(v.pre) > 0 {
			cargs = append(cargs, "--prerelease")
		}
		cmd := exec.Command("gh", cargs...)
		cmd.Dir = dir
		var so, se bytes.Buffer
		cmd.Stdout, cmd.Stderr = &so, &se
		if err := cmd.Run(); err != nil {
			return fail("gh release create: %s %s", se.String(), err)
		}
		url = strings.TrimSpace(so.String())
	}

	if dryRun {
		fmt.Fprintf(stdout, "proxy: would wait up to %s for go list -m %s@%s\n", wait, module, version)
		fmt.Fprintf(stdout, "install: would go install %s/cmd/vloop@%s into a temporary GOBIN and check what it reports\n", module, version)
		fmt.Fprintln(stdout, "dry run: nothing written")
		return 0
	}
	if !waitProxy(module, version, wait) {
		return fail("the module proxy does not serve %s after %s", version, wait)
	}
	gotV, gotM, err := installCheck(module, version)
	if err != nil {
		return fail("%s", err)
	}
	if gotV != strings.TrimPrefix(version, "v") || gotM != version {
		return fail("go install %s/cmd/vloop@%s reports version %s, module %s", module, version, gotV, gotM)
	}
	fmt.Fprintf(stdout, "published %s: %s\n", version, url)
	fmt.Fprintf(stdout, "install: go install %s/cmd/vloop@%s\n", module, version)
	fmt.Fprintln(stdout, "then, in each repository: vloop upgrade")
	fmt.Fprintln(stdout, "and for the plugin: claude plugin marketplace update vloop && claude plugin update vloop@vloop")
	return 0
}
