// Package install reads and writes the install stamp, .vloop/install.json,
// and compares the semantic versions that init, upgrade and doctor share.
package install

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Path is the stamp's repo-relative path, with `/` separators.
const Path = ".vloop/install.json"

// SchemaName is the schema every stamp names.
const SchemaName = "install/v1"

// Stamp is the install/v1 document. Upgraded is nil until the first upgrade.
type Stamp struct {
	Schema      string  `json:"schema"`
	Version     string  `json:"version"`
	Commit      string  `json:"commit"`
	Initialized string  `json:"initialized"`
	Upgraded    *string `json:"upgraded"`
}

// Read loads the stamp under root. A missing file is an error satisfying
// errors.Is(err, os.ErrNotExist): the repository is not initialised.
func Read(root string) (Stamp, error) {
	var s Stamp
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(Path)))
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return Stamp{}, fmt.Errorf("%s: %w", Path, err)
	}
	if s.Schema != SchemaName {
		return Stamp{}, fmt.Errorf("%s: schema is %q, want %q", Path, s.Schema, SchemaName)
	}
	return s, nil
}

// Write stores the stamp under root, creating .vloop/ if needed.
func Write(root string, s Stamp) error {
	s.Schema = SchemaName
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	p := filepath.Join(root, filepath.FromSlash(Path))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, append(b, '\n'), 0o644)
}

// Version is a parsed MAJOR.MINOR.PATCH version.
type Version struct{ Major, Minor, Patch int }

// ErrPreRelease is returned by Parse for a version with a pre-release or
// build suffix (0.0.0-dev): it cannot be ordered.
var ErrPreRelease = errors.New("pre-release version")

// Parse reads MAJOR.MINOR.PATCH. A suffix yields ErrPreRelease.
func Parse(v string) (Version, error) {
	if strings.ContainsAny(v, "-+") {
		return Version{}, fmt.Errorf("%q: %w", v, ErrPreRelease)
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("%q is not a MAJOR.MINOR.PATCH version", v)
	}
	var n [3]int
	for i, p := range parts {
		x, err := strconv.Atoi(p)
		if err != nil || x < 0 || p != strconv.Itoa(x) {
			return Version{}, fmt.Errorf("%q is not a MAJOR.MINOR.PATCH version", v)
		}
		n[i] = x
	}
	return Version{n[0], n[1], n[2]}, nil
}

// Relation is how an upgrade from one version to another is judged.
type Relation int

const (
	Same      Relation = iota // from = to
	Newer                     // an ordinary upgrade
	Breaking                  // an upgrade that needs --yes
	FromNewer                 // from is newer than to: refuse
	NeedsYes                  // a version cannot be compared: needs --yes
)

// Compare judges an upgrade from `from` to `to`. Breaking means the majors
// differ or both are 0 and the minors differ. A pre-release version on either
// side yields NeedsYes, unless the two strings are identical (Same). An
// unparsable version is an error.
func Compare(from, to string) (Relation, error) {
	if from == to {
		return Same, nil
	}
	f, ferr := Parse(from)
	t, terr := Parse(to)
	for _, err := range []error{ferr, terr} {
		if err != nil && !errors.Is(err, ErrPreRelease) {
			return 0, err
		}
	}
	if ferr != nil || terr != nil {
		return NeedsYes, nil
	}
	switch {
	case f == t:
		return Same, nil
	case less(t, f):
		return FromNewer, nil
	case f.Major != t.Major || (f.Major == 0 && f.Minor != t.Minor):
		return Breaking, nil
	}
	return Newer, nil
}

func less(a, b Version) bool {
	if a.Major != b.Major {
		return a.Major < b.Major
	}
	if a.Minor != b.Minor {
		return a.Minor < b.Minor
	}
	return a.Patch < b.Patch
}
