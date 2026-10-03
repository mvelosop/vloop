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
		{"effort.review", "", false}, {"shell", defaultShell(), true}, {"areas", "", false},
		{"metrics.stacks", "", false}, {"metrics.code", "", false}, {"metrics.test", "", false},
		{"metrics.docs", "", false}, {"metrics.excluded", "", false},
		{"run.max-iterations", "30", true}, {"run.cost-ceiling", "40", true}, {"run.max-attempts", "3", true},
		{"run.stall-limit", "2", true}, {"run.convergence-max", "3.0", true}, {"run.convergence-min", "6", true},
		{"run.gate-timeout", "15", true}, {"run.session-timeout", "60", true}, {"run.keep-awake", "on", true},
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

func TestShellKey(t *testing.T) {
	root := scratch(t)
	if v, _ := Get(root, "shell"); v.Value != defaultShell() || v.Source != SourceDefault {
		t.Errorf("default: %+v", v)
	}
	for _, s := range []string{"sh", "bash", "pwsh", "powershell", "cmd"} {
		if err := Set(root, "shell", s); err != nil {
			t.Errorf("shell %s: %v", s, err)
		}
	}
	err := Set(root, "shell", "zsh")
	if err == nil || err.Error() != `invalid value "zsh" for shell: want one of sh, bash, pwsh, powershell, cmd` {
		t.Errorf("zsh: %v", err)
	}
	t.Setenv("VLOOP_SHELL", "cmd")
	if v, _ := Get(root, "shell"); v.Value != "cmd" || v.Source != SourceEnv {
		t.Errorf("env: %+v", v)
	}
}

func TestAreasKey(t *testing.T) {
	root := scratch(t)
	v, err := Get(root, "areas")
	if err != nil || v.Value != "" || v.Source != SourceDefault || v.List == nil || len(v.List) != 0 {
		t.Fatalf("default: %+v %v", v, err)
	}
	if err := Set(root, "areas", "cli,brief,config"); err != nil {
		t.Fatal(err)
	}
	if got := readCfg(t, root); got != "areas = [\"cli\", \"brief\", \"config\"]\n" {
		t.Errorf("file: %q", got)
	}
	v, _ = Get(root, "areas")
	if v.Value != "cli,brief,config" || v.Source != SourceFile || len(v.List) != 3 || v.List[1] != "brief" {
		t.Errorf("file value: %+v", v)
	}
	before := readCfg(t, root)
	for _, bad := range []string{"cli,Bad_Area", "cli,", ",cli", "a b", "Cli"} {
		var iv *InvalidValueError
		if err := Set(root, "areas", bad); !errors.As(err, &iv) {
			t.Errorf("%q accepted: %v", bad, err)
		}
	}
	if readCfg(t, root) != before {
		t.Error("refused value changed the file")
	}
	if err := Set(root, "areas", "cli,brief,config"); err != nil || readCfg(t, root) != before {
		t.Errorf("idempotent set: %v", err)
	}
	t.Setenv("VLOOP_AREAS", "x,y")
	if v, _ := Get(root, "areas"); v.Source != SourceEnv || len(v.List) != 2 {
		t.Errorf("env: %+v", v)
	}
	t.Setenv("VLOOP_AREAS", "")
	if err := Set(root, "areas", ""); err != nil {
		t.Fatal(err)
	}
	if v, _ := Get(root, "areas"); v.Source != SourceDefault || len(v.List) != 0 {
		t.Errorf("removed: %+v", v)
	}
	writeCfg(t, root, "areas = \"cli\"\n")
	var se *SourceError
	if _, err := Get(root, "areas"); !errors.As(err, &se) {
		t.Errorf("non-array: %v", err)
	}
}

func TestMetricsKeys(t *testing.T) {
	root := scratch(t)
	at := -1
	for i, k := range Keys {
		if k.Name == "areas" {
			at = i
		}
	}
	if at < 0 || at+6 > len(Keys) {
		t.Fatalf("no five keys after areas in %d keys", len(Keys))
	}
	var names []string
	for _, k := range Keys[at+1 : at+6] {
		names = append(names, k.Name)
		if !k.List {
			t.Errorf("%s is not a list", k.Name)
		}
	}
	if got := strings.Join(names, " "); got != "metrics.stacks metrics.code metrics.test metrics.docs metrics.excluded" {
		t.Fatalf("keys after areas: %s", got)
	}
	if err := Set(root, "metrics.code", "internal/brief/templates/**,tools/*.go"); err != nil {
		t.Fatal(err)
	}
	if got := readCfg(t, root); got != "[metrics]\n  code = [\"internal/brief/templates/**\", \"tools/*.go\"]\n" {
		t.Errorf("file: %q", got)
	}
	v, _ := Get(root, "metrics.code")
	if v.Source != SourceFile || len(v.List) != 2 {
		t.Errorf("value: %+v", v)
	}
	t.Setenv("VLOOP_METRICS_TEST", "e2e/**")
	if v, _ := Get(root, "metrics.test"); v.Source != SourceEnv || v.List[0] != "e2e/**" {
		t.Errorf("env: %+v", v)
	}
	var iv *InvalidValueError
	if err := Set(root, "metrics.docs", "a,,b"); !errors.As(err, &iv) {
		t.Errorf("empty pattern accepted: %v", err)
	}
	if err := Set(root, "areas", "**"); !errors.As(err, &iv) {
		t.Errorf("areas accepted a glob: %v", err)
	}
}

func TestStacksScopeValidation(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "services", "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "file"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Set(root, "metrics.stacks", "go,csharp@services/api"); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"csharp@missing", "csharp@file", "csharp@/services", "csharp@services/", "csharp@./services", "csharp@services/../services", "cobol@services", "@services"} {
		err := Set(root, "metrics.stacks", "go,"+bad)
		var iv *InvalidValueError
		if !errors.As(err, &iv) || iv.Value != bad {
			t.Errorf("%s: err %v", bad, err)
			continue
		}
		want := "invalid value \"" + bad + "\" for metrics.stacks: want <stack> or <stack>@<existing directory>"
		if err.Error() != want {
			t.Errorf("got %q", err.Error())
		}
	}
	v, err := Get(root, "metrics.stacks")
	if err != nil || v.Value != "go,csharp@services/api" {
		t.Errorf("file changed: %v %v", v, err)
	}
}

func TestStacksScopeReadIgnoresMissingDirectory(t *testing.T) {
	root := t.TempDir()
	write := func(s string) {
		if err := os.MkdirAll(filepath.Join(root, ".vloop"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, ".vloop", "config.toml"), []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("[metrics]\nstacks = [\"go\", \"csharp@gone\"]\n")
	if v, err := Get(root, "metrics.stacks"); err != nil || v.Value != "go,csharp@gone" {
		t.Errorf("%v %v", v, err)
	}
	write("[metrics]\nstacks = [\"go\", \"csharp@/abs\"]\n")
	var se *SourceError
	if _, err := Get(root, "metrics.stacks"); !errors.As(err, &se) {
		t.Errorf("want SourceError, got %v", err)
	}
}

func TestRunKeysValidation(t *testing.T) {
	root := scratch(t)
	cases := []struct {
		key, good string
		bad       []string
	}{
		{"run.max-iterations", "0", []string{"-1", "1.5", "abc"}},
		{"run.convergence-min", "0", []string{"-1", "x"}},
		{"run.max-attempts", "1", []string{"0", "2.5"}},
		{"run.stall-limit", "1", []string{"0", ""}},
		{"run.cost-ceiling", "12.5", []string{"0", "-3", "abc", "NaN", "Inf"}},
		{"run.convergence-max", "2.5", []string{"0", "-0.5", "x"}},
		{"run.gate-timeout", "1", []string{"0", "1.5", "-2", "x"}},
		{"run.session-timeout", "90", []string{"0", "2.5", "-1", "x"}},
	}
	for _, c := range cases {
		if err := Set(root, c.key, c.good); err != nil {
			t.Errorf("set %s=%s: %v", c.key, c.good, err)
		}
		if v, err := Get(root, c.key); err != nil || v.Value != c.good || v.Source != SourceFile {
			t.Errorf("get %s = %+v, %v", c.key, v, err)
		}
		for _, b := range c.bad {
			var iv *InvalidValueError
			err := Check(root, c.key, b)
			if b == "" {
				continue
			}
			if !errors.As(err, &iv) || !strings.HasPrefix(err.Error(), `invalid value "`+b+`" for `+c.key+`: want `) {
				t.Errorf("%s=%q: %v", c.key, b, err)
			}
		}
	}
	if err := Set(root, "run.max-attempts", "0"); err == nil || err.Error() != `invalid value "0" for run.max-attempts: want an integer of at least 1` {
		t.Errorf("message: %v", err)
	}
}

func TestRunKeysEnvFileDefault(t *testing.T) {
	root := scratch(t)
	writeCfg(t, root, "[run]\nmax-attempts = 4\nconvergence-max = 2.5\ncost-ceiling = 10\n")
	for key, want := range map[string]string{"run.max-attempts": "4", "run.convergence-max": "2.5", "run.cost-ceiling": "10"} {
		if v, err := Get(root, key); err != nil || v.Value != want || v.Source != SourceFile {
			t.Errorf("%s = %+v, %v", key, v, err)
		}
	}
	if EnvVar("run.convergence-max") != "VLOOP_RUN_CONVERGENCE_MAX" {
		t.Error(EnvVar("run.convergence-max"))
	}
	t.Setenv("VLOOP_RUN_MAX_ATTEMPTS", "9")
	if v, _ := Get(root, "run.max-attempts"); v.Value != "9" || v.Source != SourceEnv {
		t.Errorf("env: %+v", v)
	}
	t.Setenv("VLOOP_RUN_MAX_ATTEMPTS", "0")
	var se *SourceError
	if _, err := Get(root, "run.max-attempts"); !errors.As(err, &se) || se.Source != "VLOOP_RUN_MAX_ATTEMPTS" {
		t.Errorf("bad env: %v", err)
	}
	t.Setenv("VLOOP_RUN_MAX_ATTEMPTS", "")
	writeCfg(t, root, "[run]\nmax-attempts = \"x\"\n")
	if _, err := Get(root, "run.max-attempts"); !errors.As(err, &se) {
		t.Errorf("string in file: %v", err)
	}
}

func TestRunKeysSetWritesRunTable(t *testing.T) {
	root := scratch(t)
	if err := Set(root, "run.max-attempts", "4"); err != nil {
		t.Fatal(err)
	}
	if got := readCfg(t, root); got != "[run]\n  max-attempts = 4\n" {
		t.Errorf("file = %q", got)
	}
	if err := Set(root, "run.max-attempts", ""); err != nil {
		t.Fatal(err)
	}
}
