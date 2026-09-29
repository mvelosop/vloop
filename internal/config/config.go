// Package config resolves the repo root and the repo-local configuration:
// .vloop/config.toml under the root, overridden by VLOOP_* variables.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
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
}

// Keys lists every key, in display order.
var Keys = []Key{
	{"language", "en", []string{"en", "es"}},
	{"model.plan", "opus", nil},
	{"model.work", "sonnet", nil},
	{"model.review", "sonnet", nil},
	{"effort.plan", "", efforts},
	{"effort.work", "", efforts},
	{"effort.review", "", efforts},
}

// Value is a resolved key. Set is false for an unset effort.
type Value struct {
	Key    string
	Value  string
	Set    bool
	Source string
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
	if e.Valid != nil {
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
	return "VLOOP_" + strings.ToUpper(strings.ReplaceAll(name, ".", "_"))
}

// Validate checks value (non-empty) against the key's valid values.
func (k Key) Validate(value string) error {
	if value == "" {
		return &InvalidValueError{k.Name, value, k.Valid}
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

func resolve(k Key, file map[string]any) (Value, error) {
	v := Value{Key: k.Name, Value: k.Default, Set: k.Default != "", Source: SourceDefault}
	if fv, ok := fileValue(file, k.Name); ok {
		s, isStr := fv.(string)
		if !isStr {
			return v, &SourceError{FilePath, fmt.Errorf("%s must be a string", k.Name)}
		}
		if err := k.Validate(s); err != nil {
			return v, &SourceError{FilePath, err}
		}
		v = Value{k.Name, s, true, SourceFile}
	}
	if ev := os.Getenv(EnvVar(k.Name)); ev != "" {
		if err := k.Validate(ev); err != nil {
			return v, &SourceError{EnvVar(k.Name), err}
		}
		v = Value{k.Name, ev, true, SourceEnv}
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
	}
	file, raw, err := readFile(root)
	if err != nil {
		return err
	}
	cur, present := fileValue(file, name)
	if value == "" && !present || value != "" && present && cur == value {
		return nil
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
		dst[leaf] = value
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
