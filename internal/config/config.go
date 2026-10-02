// Package config resolves the repo root and the repo-local configuration:
// .vloop/config.toml under the root, overridden by VLOOP_* variables.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/mvelosop/vloop/internal/classify"
)

// Sources of a resolved value.
const (
	SourceDefault = "default"
	SourceFile    = "file"
	SourceEnv     = "env"
)

var efforts = []string{"low", "medium", "high", "xhigh", "max"}

// Key describes one configuration key. Valid is nil when any non-empty string
// is accepted.
type Key struct {
	Name    string
	Default string // empty means unset
	Valid   []string
	List    bool    // a list of strings, stored as a TOML array and given as comma-joined text
	Stacks  bool    // with List: entries are <stack> or <stack>@<path>, see classify.ParseStack
	Glob    bool    // with List: entries are any non-empty glob pattern rather than lower-case names
	Int     bool    // a whole number, stored as a TOML integer
	Num     bool    // a number, stored as a TOML float; a TOML integer is read too
	Min     float64 // with Int or Num: the smallest valid value
	MinOpen bool    // with Int or Num: Min itself is not valid
}

var areaName = regexp.MustCompile(`^[a-z0-9-]+$`)

func defaultShell() string {
	if runtime.GOOS == "windows" {
		return "pwsh"
	}
	return "sh"
}

// Keys lists every key, in display order.
var Keys = []Key{
	{Name: "language", Default: "en", Valid: []string{"en", "es"}},
	{Name: "model.plan", Default: "opus"},
	{Name: "model.work", Default: "sonnet"},
	{Name: "model.review", Default: "sonnet"},
	{Name: "effort.plan", Valid: efforts},
	{Name: "effort.work", Valid: efforts},
	{Name: "effort.review", Valid: efforts},
	{Name: "shell", Default: defaultShell(), Valid: []string{"sh", "bash", "pwsh", "powershell", "cmd"}},
	{Name: "areas", List: true},
	{Name: "metrics.stacks", List: true, Stacks: true},
	{Name: "metrics.code", List: true, Glob: true},
	{Name: "metrics.test", List: true, Glob: true},
	{Name: "metrics.docs", List: true, Glob: true},
	{Name: "metrics.excluded", List: true, Glob: true},
	{Name: "run.max-iterations", Default: "30", Int: true},
	{Name: "run.cost-ceiling", Default: "40", Num: true, MinOpen: true},
	{Name: "run.max-attempts", Default: "3", Int: true, Min: 1},
	{Name: "run.stall-limit", Default: "2", Int: true, Min: 1},
	{Name: "run.convergence-max", Default: "3.0", Num: true, MinOpen: true},
	{Name: "run.convergence-min", Default: "6", Int: true},
	{Name: "run.gate-timeout", Default: "15", Int: true, Min: 1},
	{Name: "run.session-timeout", Default: "60", Int: true, Min: 1},
}

// Value is a resolved key. Set is false for an unset effort. For a list key
// Value is the comma-joined text and List the entries (never nil).
type Value struct {
	Key    string
	Value  string
	Set    bool
	Source string
	List   []string
}

// UnknownKeyError and InvalidValueError are usage errors (exit 2).
type UnknownKeyError struct{ Key string }

func (e *UnknownKeyError) Error() string { return fmt.Sprintf("unknown config key %q", e.Key) }

type InvalidValueError struct {
	Key, Value string
	Valid      []string
}

func (e *InvalidValueError) Error() string {
	want := "a non-empty string"
	if k, err := Lookup(e.Key); err == nil && k.Stacks {
		want = "<stack> or <stack>@<existing directory>"
	} else if err == nil && k.List && e.Valid != nil {
		want = "a comma-separated list of " + strings.Join(e.Valid, ", ")
	} else if err == nil && k.List && k.Glob {
		want = "a comma-separated list of non-empty glob patterns"
	} else if err == nil && k.List {
		want = "a comma-separated list of names made of lower-case letters, digits and hyphens"
	} else if err == nil && (k.Int || k.Num) {
		want = k.numWant()
	} else if e.Valid != nil {
		want = "one of " + strings.Join(e.Valid, ", ")
	}
	return fmt.Sprintf("invalid value %q for %s: want %s", e.Value, e.Key, want)
}

// SourceError is a bad value from the environment or file, or a malformed
// file. It names the source and is a problem (exit 1), not a usage error.
type SourceError struct {
	Source string // variable name or config file path
	Err    error
}

func (e *SourceError) Error() string { return e.Source + ": " + e.Err.Error() }
func (e *SourceError) Unwrap() error { return e.Err }

// Lookup returns the definition of a key.
func Lookup(name string) (Key, error) {
	for _, k := range Keys {
		if k.Name == name {
			return k, nil
		}
	}
	return Key{}, &UnknownKeyError{name}
}

// EnvVar is the environment variable overriding a key.
func EnvVar(name string) string {
	return "VLOOP_" + strings.ToUpper(strings.NewReplacer(".", "_", "-", "_").Replace(name))
}

func (k Key) numWant() string {
	kind, min := "a number", strconv.FormatFloat(k.Min, 'f', -1, 64)
	if k.Int {
		kind = "an integer"
	}
	if k.MinOpen {
		return kind + " greater than " + min
	}
	return kind + " of at least " + min
}

// parseNum parses text as k's number and reports whether it is valid.
func (k Key) parseNum(text string) (float64, bool) {
	var f float64
	if k.Int {
		n, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return 0, false
		}
		f = float64(n)
	} else {
		var err error
		f, err = strconv.ParseFloat(text, 64)
		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
			return 0, false
		}
	}
	return f, f > k.Min || (!k.MinOpen && f == k.Min)
}

// Validate checks value (non-empty) against the key's valid values.
func (k Key) Validate(value string) error {
	if value == "" {
		return &InvalidValueError{k.Name, value, k.Valid}
	}
	if k.Stacks {
		for _, a := range strings.Split(value, ",") {
			if _, _, err := classify.ParseStack(a); err != nil {
				return &InvalidValueError{k.Name, a, nil}
			}
		}
		return nil
	}
	if k.List {
		for _, a := range strings.Split(value, ",") {
			if (k.Glob && a == "") || (!k.Glob && !areaName.MatchString(a)) {
				return &InvalidValueError{k.Name, value, k.Valid}
			}
			if k.Valid != nil && !slices.Contains(k.Valid, a) {
				return &InvalidValueError{k.Name, value, k.Valid}
			}
		}
		return nil
	}
	if k.Int || k.Num {
		if _, ok := k.parseNum(value); !ok {
			return &InvalidValueError{k.Name, value, nil}
		}
		return nil
	}
	if k.Valid == nil {
		return nil
	}
	for _, v := range k.Valid {
		if v == value {
			return nil
		}
	}
	return &InvalidValueError{k.Name, value, k.Valid}
}

// checkScopeDirs requires every scoped entry of a metrics.stacks value to name
// an existing directory under root. Reading a config does not: doctor warns.
func checkScopeDirs(root, value string) error {
	for _, a := range strings.Split(value, ",") {
		_, scope, _ := classify.ParseStack(a)
		if scope == "" {
			continue
		}
		if fi, err := os.Stat(filepath.Join(root, filepath.FromSlash(scope))); err != nil || !fi.IsDir() {
			return &InvalidValueError{"metrics.stacks", a, nil}
		}
	}
	return nil
}

func filePath(root string) string { return filepath.Join(root, filepath.FromSlash(FilePath)) }

// readFile parses the config file into a nested map; a missing file is empty.
func readFile(root string) (map[string]any, []byte, error) {
	raw, err := os.ReadFile(filePath(root))
	if errors.Is(err, os.ErrNotExist) {
		return map[string]any{}, nil, nil
	}
	if err != nil {
		return nil, nil, &SourceError{FilePath, err}
	}
	m := map[string]any{}
	if _, err := toml.Decode(string(raw), &m); err != nil {
		return nil, nil, &SourceError{FilePath, err}
	}
	return m, raw, nil
}

// fileValue finds a dotted key in the parsed file.
func fileValue(m map[string]any, name string) (any, bool) {
	cur := any(m)
	for _, part := range strings.Split(name, ".") {
		t, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		if cur, ok = t[part]; !ok {
			return nil, false
		}
	}
	return cur, true
}

func resolveList(k Key, file map[string]any) (Value, error) {
	v := Value{Key: k.Name, Source: SourceDefault, List: []string{}}
	set := func(text, source string) { v = Value{k.Name, text, true, source, strings.Split(text, ",")} }
	if fv, ok := fileValue(file, k.Name); ok {
		arr, isArr := fv.([]any)
		if !isArr {
			return v, &SourceError{FilePath, fmt.Errorf("%s must be an array of strings", k.Name)}
		}
		var names []string
		for _, e := range arr {
			s, isStr := e.(string)
			if !isStr {
				return v, &SourceError{FilePath, fmt.Errorf("%s must be an array of strings", k.Name)}
			}
			names = append(names, s)
		}
		if len(names) > 0 {
			text := strings.Join(names, ",")
			if err := k.Validate(text); err != nil {
				return v, &SourceError{FilePath, err}
			}
			set(text, SourceFile)
		}
	}
	if ev := os.Getenv(EnvVar(k.Name)); ev != "" {
		if err := k.Validate(ev); err != nil {
			return v, &SourceError{EnvVar(k.Name), err}
		}
		set(ev, SourceEnv)
	}
	return v, nil
}

func resolve(k Key, file map[string]any) (Value, error) {
	if k.List {
		return resolveList(k, file)
	}
	v := Value{Key: k.Name, Value: k.Default, Set: k.Default != "", Source: SourceDefault}
	if fv, ok := fileValue(file, k.Name); ok {
		s, isStr := fv.(string)
		switch n := fv.(type) {
		case int64:
			if k.Int || k.Num {
				s, isStr = strconv.FormatInt(n, 10), true
			}
		case float64:
			if k.Num {
				s, isStr = strconv.FormatFloat(n, 'f', -1, 64), true
			}
		}
		if !isStr {
			want := "a string"
			if k.Int {
				want = "an integer"
			} else if k.Num {
				want = "a number"
			}
			return v, &SourceError{FilePath, fmt.Errorf("%s must be %s", k.Name, want)}
		}
		if err := k.Validate(s); err != nil {
			return v, &SourceError{FilePath, err}
		}
		v = Value{Key: k.Name, Value: s, Set: true, Source: SourceFile}
	}
	if ev := os.Getenv(EnvVar(k.Name)); ev != "" {
		if err := k.Validate(ev); err != nil {
			return v, &SourceError{EnvVar(k.Name), err}
		}
		v = Value{Key: k.Name, Value: ev, Set: true, Source: SourceEnv}
	}
	return v, nil
}

// Get resolves one key under root.
func Get(root, name string) (Value, error) {
	k, err := Lookup(name)
	if err != nil {
		return Value{}, err
	}
	file, _, err := readFile(root)
	if err != nil {
		return Value{}, err
	}
	return resolve(k, file)
}

// List resolves every key under root.
func List(root string) ([]Value, error) {
	file, _, err := readFile(root)
	if err != nil {
		return nil, err
	}
	out := make([]Value, 0, len(Keys))
	for _, k := range Keys {
		v, err := resolve(k, file)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// Check validates a non-empty value for key name as Set would, without
// writing anything.
func Check(root, name, value string) error {
	k, err := Lookup(name)
	if err != nil {
		return err
	}
	if err := k.Validate(value); err != nil {
		return err
	}
	if k.Stacks {
		return checkScopeDirs(root, value)
	}
	return nil
}

// Set validates and writes name=value into the config file under root,
// preserving every other key. An empty value removes the key. The file is not
// touched (nor created) when the result would not change the key.
func Set(root, name, value string) error {
	k, err := Lookup(name)
	if err != nil {
		return err
	}
	if value != "" {
		if err := k.Validate(value); err != nil {
			return err
		}
		if k.Stacks {
			if err := checkScopeDirs(root, value); err != nil {
				return err
			}
		}
	}
	file, raw, err := readFile(root)
	if err != nil {
		return err
	}
	cur, present := fileValue(file, name)
	if value == "" && !present {
		return nil
	}
	if value != "" && present {
		if arr, ok := cur.([]any); k.List && ok {
			var names []string
			for _, e := range arr {
				names = append(names, fmt.Sprint(e))
			}
			if strings.Join(names, ",") == value {
				return nil
			}
		} else if cur == value || (k.Int || k.Num) && fmt.Sprint(cur) == value {
			return nil
		}
	}
	table, leaf, _ := strings.Cut(name, ".")
	if leaf == "" {
		leaf, table = table, ""
	}
	dst := file
	if table != "" {
		t, ok := file[table].(map[string]any)
		if !ok {
			t = map[string]any{}
			file[table] = t
		}
		dst = t
	}
	if value == "" {
		delete(dst, leaf)
		if table != "" && len(dst) == 0 {
			delete(file, table)
		}
	} else {
		if k.List {
			dst[leaf] = strings.Split(value, ",")
		} else if k.Int {
			dst[leaf], _ = strconv.ParseInt(value, 10, 64)
		} else if k.Num {
			dst[leaf], _ = strconv.ParseFloat(value, 64)
		} else {
			dst[leaf] = value
		}
	}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(file); err != nil {
		return err
	}
	if raw != nil && bytes.Equal(raw, buf.Bytes()) {
		return nil
	}
	p := filePath(root)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, buf.Bytes(), 0o644)
}
