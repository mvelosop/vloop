package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func scratch(t *testing.T) string {
	t.Helper()
	for _, k := range Keys {
		t.Setenv(EnvVar(k.Name), "")
	}
	return t.TempDir()
}

func writeCfg(t *testing.T, root, body string) {
	t.Helper()
	p := filepath.Join(root, ".vloop", "config.toml")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readCfg(t *testing.T, root string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, ".vloop", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestDefaultsEveryRow(t *testing.T) {
	root := scratch(t)
	want := []struct {
		key, val string
		set      bool
	}{
		{"language", "en", true}, {"model.plan", "opus", true}, {"model.work", "sonnet", true},
		{"model.review", "sonnet", true}, {"effort.plan", "", false}, {"effort.work", "", false},
		{"effort.review", "", false},
	}
	vals, err := List(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(vals) != len(want) {
		t.Fatalf("got %d keys", len(vals))
	}
	for i, w := range want {
		v := vals[i]
		if v.Key != w.key || v.Value != w.val || v.Set != w.set || v.Source != SourceDefault {
			t.Errorf("row %d = %+v, want %+v", i, v, w)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".vloop")); err == nil {
		t.Error("List created .vloop")
	}
}

func TestValidValuesEveryRow(t *testing.T) {
	root := scratch(t)
	good := map[string]string{"language": "es", "model.plan": "x", "model.work": "y", "model.review": "z",
		"effort.plan": "low", "effort.work": "xhigh", "effort.review": "max"}
	for k, v := range good {
		if err := Set(root, k, v); err != nil {
			t.Fatalf("set %s: %v", k, err)
		}
		got, err := Get(root, k)
		if err != nil || got.Value != v || got.Source != SourceFile {
			t.Errorf("get %s = %+v, %v", k, got, err)
		}
	}
	bad := map[string]string{"language": "fr", "effort.plan": "extreme", "effort.work": "Low", "effort.review": "none"}
	for k, v := range bad {
		var iv *InvalidValueError
		if err := Set(root, k, v); !errors.As(err, &iv) {
			t.Errorf("set %s=%s: %v", k, v, err)
		}
	}
	err := Set(root, "language", "fr")
	if err.Error() != `invalid value "fr" for language: want one of en, es` {
		t.Errorf("message: %v", err)
	}
	k, _ := Lookup("model.plan")
	if err := k.Validate("  "); err != nil {
		t.Error("whitespace-only is a non-empty string")
	}
	if err := (&InvalidValueError{"model.plan", "", nil}); !strings.HasSuffix(err.Error(), "want a non-empty string") {
		t.Error(err)
	}
}

func TestUnknownKey(t *testing.T) {
	root := scratch(t)
	var uk *UnknownKeyError
	if _, err := Get(root, "colour"); !errors.As(err, &uk) || err.Error() != `unknown config key "colour"` {
		t.Errorf("get: %v", err)
	}
	if err := Set(root, "colour", "red"); !errors.As(err, &uk) {
		t.Errorf("set: %v", err)
	}
}

func TestEnvOverFileOverDefault(t *testing.T) {
	root := scratch(t)
	writeCfg(t, root, "[model]\nwork = \"from-file\"\n")
	if v, _ := Get(root, "model.work"); v.Value != "from-file" || v.Source != SourceFile {
		t.Errorf("file: %+v", v)
	}
	t.Setenv("VLOOP_MODEL_WORK", "from-env")
	if v, _ := Get(root, "model.work"); v.Value != "from-env" || v.Source != SourceEnv {
		t.Errorf("env: %+v", v)
	}
	if v, _ := Get(root, "model.plan"); v.Value != "opus" || v.Source != SourceDefault {
		t.Errorf("default: %+v", v)
	}
}

func TestBadSourcesAreSourceErrors(t *testing.T) {
	root := scratch(t)
	t.Setenv("VLOOP_LANGUAGE", "fr")
	var se *SourceError
	if _, err := Get(root, "language"); !errors.As(err, &se) || se.Source != "VLOOP_LANGUAGE" {
		t.Errorf("env: %v", err)
	}
	t.Setenv("VLOOP_LANGUAGE", "")
	writeCfg(t, root, "language = \"fr\"\n")
	if _, err := Get(root, "language"); !errors.As(err, &se) || se.Source != ".vloop/config.toml" {
		t.Errorf("file: %v", err)
	}
	writeCfg(t, root, "language = \n[[[\n")
	if _, err := List(root); !errors.As(err, &se) {
		t.Errorf("malformed: %v", err)
	}
	writeCfg(t, root, "language = 3\n")
	if _, err := Get(root, "language"); !errors.As(err, &se) {
		t.Errorf("non-string: %v", err)
	}
}

func TestSetIdempotentByteForByte(t *testing.T) {
	root := scratch(t)
	if err := Set(root, "language", "es"); err != nil {
		t.Fatal(err)
	}
	// A hand-formatted file whose value already matches must not be rewritten.
	body := "# mine\nlanguage   =   \"es\"\n"
	writeCfg(t, root, body)
	if err := Set(root, "language", "es"); err != nil {
		t.Fatal(err)
	}
	if got := readCfg(t, root); got != body {
		t.Errorf("file rewritten: %q", got)
	}
}

func TestSetPreservesOtherKeysAndRemoves(t *testing.T) {
	root := scratch(t)
	writeCfg(t, root, "language = \"es\"\nextra = 7\n[model]\nplan = \"p\"\n")
	if err := Set(root, "model.work", "w"); err != nil {
		t.Fatal(err)
	}
	if err := Set(root, "model.plan", ""); err != nil {
		t.Fatal(err)
	}
	got := readCfg(t, root)
	for _, s := range []string{`language = "es"`, "extra = 7", `work = "w"`} {
		if !strings.Contains(got, s) {
			t.Errorf("missing %q in %q", s, got)
		}
	}
	if strings.Contains(got, "plan") {
		t.Errorf("plan not removed: %q", got)
	}
	if err := Set(root, "model.work", ""); err != nil {
		t.Fatal(err)
	}
	if got := readCfg(t, root); strings.Contains(got, "model") {
		t.Errorf("empty table kept: %q", got)
	}
}

func TestRemoveMissingCreatesNothing(t *testing.T) {
	root := scratch(t)
	if err := Set(root, "language", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".vloop")); err == nil {
		t.Error("removal of an absent key created .vloop")
	}
}

func TestRejectedSetLeavesFileUntouched(t *testing.T) {
	root := scratch(t)
	writeCfg(t, root, "language = \"es\"\n")
	_ = Set(root, "language", "fr")
	_ = Set(root, "colour", "red")
	if got := readCfg(t, root); got != "language = \"es\"\n" {
		t.Errorf("file changed: %q", got)
	}
}
