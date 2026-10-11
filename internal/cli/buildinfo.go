package cli

import (
	"regexp"
	"runtime/debug"
)

// readBuildInfo is the seam for Go's build info; tests inject a module
// version, a revision and a modified flag.
var readBuildInfo = debug.ReadBuildInfo

// Resolve completes a Build from what the toolchain recorded when no stamp
// named the commit: the VCS revision (with -dirty when the tree was modified),
// else unknown. Module is the build's module version, empty for (devel).
func Resolve(version, commit string) Build {
	b := Build{Version: version, Commit: commit}
	info, ok := readBuildInfo()
	if !ok || info == nil {
		return b
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		b.Module = v
	}
	if commit != "" && commit != "unknown" {
		return b
	}
	var rev string
	var modified bool
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}
	if rev != "" {
		b.Commit = rev
		if modified {
			b.Commit += "-dirty"
		}
	}
	return b
}

var releaseVersion = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

// isRelease reports a module version that is a release: vX.Y.Z with no
// pre-release or pseudo-version suffix.
func isRelease(module string) bool { return releaseVersion.MatchString(module) }
