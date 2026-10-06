package cli

import (
	"runtime"
	"strings"
)

// trustKeyMatch reports whether a ~/.claude.json project key names the
// repository path: both are compared with "/" separators, and
// case-insensitively on Windows and macOS (goos is runtime.GOOS's value).
func trustKeyMatch(goos, key, path string) bool {
	key = strings.ReplaceAll(key, `\`, "/")
	path = strings.ReplaceAll(path, `\`, "/")
	if goos == "windows" || goos == "darwin" {
		return strings.EqualFold(key, path)
	}
	return key == path
}

// hookCounts reports whether a pre-commit hook file is active: on Windows by
// its presence, elsewhere by its executable bit.
func hookCounts(goos string, mode uint32) bool {
	if goos == "windows" {
		return true
	}
	return mode&0o111 != 0
}

var hostOS = runtime.GOOS
