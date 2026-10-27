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
		{"model.review", "sonnet", true}, {"model.second-review", "sonnet", true}, {"model.gate-review", "sonnet", true}, {"model.verify", "sonnet", true}, {"effort.plan", "", false}, {"effort.work", "", false},
		{"effort.review", "", false}, {"effort.second-review", "", false}, {"effort.gate-review", "", false}, {"effort.verify", "", false}, {"agent.plan", "claude", true}, {"agent.work", "claude", true}, {"agent.review", "claude", true}, {"agent.second-review", "claude", true}, {"agent.gate-review", "claude", true}, {"agent.verify", "claude", true}, {"copilot.credit-usd", "0.0066667", true}, {"shell", defaultShell(), true}, {"areas", "", false},
		{"metrics.stacks", "", false}, {"metrics.code", "", false}, {"metrics.test", "", false},
		{"metrics.docs", "", false}, {"metrics.excluded", "", false},
		{"run.max-iterations", "30", true}, {"run.cost-ceiling", "40", true}, {"run.max-attempts", "3", true},
		{"run.stall-limit", "2", true}, {"run.convergence-max", "3.0", true}, {"run.convergence-min", "6", true},
		{"run.gate-timeout", "15", true}, {"run.gate-scratch", "", false}, {"run.git-config-ignore", "", false}, {"run.session-timeout", "60", true}, {"run.second-review", "off", true}, {"run.verify", "true", true}, {"run.keep-awake", "on", true},
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

func TestGateKeys(t *testing.T) {
	root := scratch(t)
	var iv *InvalidValueError
	if err := Set(root, "effort.gate-review", "huge"); !errors.As(err, &iv) {
		t.Errorf("effort.gate-review=huge: %v", err)
	}
	if err := Set(root, "effort.gate-review", "max"); err != nil {
		t.Fatal(err)
	}
	if err := Set(root, "model.gate-review", "haiku"); err != nil {
		t.Fatal(err)
	}
	if err := Set(root, "run.gate-scratch", "web/.gate/,app/.gate/"); err != nil {
		t.Fatal(err)
	}
	if err := Set(root, "run.gate-scratch", ",x"); !errors.As(err, &iv) {
		t.Errorf("empty scratch entry: %v", err)
	}
	for k, want := range map[string]string{"effort.gate-review": "max", "model.gate-review": "haiku", "run.gate-scratch": "web/.gate/,app/.gate/"} {
		if v, err := Get(root, k); err != nil || v.Value != want || v.Source != SourceFile {
			t.Errorf("get %s = %+v, %v", k, v, err)
		}
	}
	if err := Set(root, "run.gate-scratch", ""); err != nil {
		t.Fatal(err)
	}
	if v, _ := Get(root, "run.gate-scratch"); v.Set || len(v.List) != 0 {
		t.Errorf("removed scratch = %+v", v)
	}
}

const twoChecks = "[[check]]\nname = \"api\"\npaths = [\"api/**\"]\nrun = \"sh api/test.sh\"\n\n" +
	"[[check]]\nname = \"web\"\npaths = [\"web/**\", \"shared/*.ts\"]\nrun = \"sh web/test.sh\"\n"

func TestChecksRead(t *testing.T) {
	root := scratch(t)
	writeCfg(t, root, twoChecks)
	cs, err := Checks(root)
	if err != nil || len(cs) != 2 || cs[0].Name != "api" || cs[1].Name != "web" ||
		strings.Join(cs[1].Paths, "|") != "web/**|shared/*.ts" || cs[1].Run != "sh web/test.sh" {
		t.Fatalf("checks = %+v, %v", cs, err)
	}
	t.Setenv("VLOOP_CHECK", "x")
	if cs2, _ := Checks(root); len(cs2) != 2 {
		t.Error("an environment variable changed the checks")
	}
}

func TestChecksRefusals(t *testing.T) {
	for name, body := range map[string]string{
		"bad name":     "[[check]]\nname = \"API\"\npaths = [\"**\"]\nrun = \"true\"\n",
		"no name":      "[[check]]\npaths = [\"**\"]\nrun = \"true\"\n",
		"duplicate":    "[[check]]\nname = \"a\"\npaths = [\"**\"]\nrun = \"true\"\n[[check]]\nname = \"a\"\npaths = [\"x\"]\nrun = \"true\"\n",
		"empty paths":  "[[check]]\nname = \"a\"\npaths = []\nrun = \"true\"\n",
		"no paths":     "[[check]]\nname = \"a\"\nrun = \"true\"\n",
		"bad glob":     "[[check]]\nname = \"a\"\npaths = [\"[\"]\nrun = \"true\"\n",
		"no run":       "[[check]]\nname = \"a\"\npaths = [\"**\"]\n",
		"blank run":    "[[check]]\nname = \"a\"\npaths = [\"**\"]\nrun = \" \"\n",
		"not an array": "check = \"x\"\n",
	} {
		root := scratch(t)
		writeCfg(t, root, body)
		var se *SourceError
		if _, err := Checks(root); !errors.As(err, &se) {
			t.Errorf("%s: Checks = %v", name, err)
		}
		if _, err := List(root); !errors.As(err, &se) {
			t.Errorf("%s: List = %v", name, err)
		}
		if _, err := Get(root, "language"); !errors.As(err, &se) {
			t.Errorf("%s: Get = %v", name, err)
		}
	}
}

func TestSetKeepsCheckTables(t *testing.T) {
	root := scratch(t)
	writeCfg(t, root, twoChecks)
	if err := Set(root, "model.gate-review", "opus"); err != nil {
		t.Fatal(err)
	}
	cs, err := Checks(root)
	if err != nil || len(cs) != 2 || cs[0].Name != "api" || cs[1].Run != "sh web/test.sh" {
		t.Fatalf("checks after set = %+v, %v", cs, err)
	}
}

func TestGitConfigIgnoreRefusesProtectedPatterns(t *testing.T) {
	root := scratch(t)
	for value, want := range map[string]string{
		"core.*":           "run.git-config-ignore: core.* would ignore core.* — the guard must see it",
		"*":                "run.git-config-ignore: * would ignore core.* — the guard must see it",
		"filter.lfs.*":     "run.git-config-ignore: filter.lfs.* would ignore filter.* — the guard must see it",
		"gitlens.*,core.*": "run.git-config-ignore: core.* would ignore core.* — the guard must see it",
	} {
		err := Set(root, "run.git-config-ignore", value)
		var cu *UsageError
		if !errors.As(err, &cu) || err.Error() != want {
			t.Errorf("Set %q = %v, want usage error %q", value, err, want)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".vloop", "config.toml")); err == nil {
		t.Error("a refused value wrote the file")
	}
	if err := Set(root, "run.git-config-ignore", "gitlens.*,branch.*.vscode-merge-base"); err != nil {
		t.Errorf("a harmless list was refused: %v", err)
	}
}

func TestGitConfigIgnoreProtectedFromSourceIsAProblem(t *testing.T) {
	root := scratch(t)
	t.Setenv("VLOOP_RUN_GIT_CONFIG_IGNORE", "core.*")
	_, err := Get(root, "run.git-config-ignore")
	var pe *ProtectedPatternError
	var cu *UsageError
	if !errors.As(err, &pe) || errors.As(err, &cu) {
		t.Errorf("Get = %v, want a bare protected pattern error", err)
	}
}

func TestGitConfigIgnoreMixedCaseProtectedRefused(t *testing.T) {
	root := scratch(t)
	err := Set(root, "run.git-config-ignore", "Core.*")
	want := "run.git-config-ignore: Core.* would ignore core.* — the guard must see it"
	if err == nil || err.Error() != want {
		t.Errorf("got %v, want %q", err, want)
	}
	if err := Set(root, "run.git-config-ignore", "Branch.Main.Foo-Bar"); err != nil {
		t.Errorf("mixed-case subsection pattern refused: %v", err)
	}
}

func TestVerifyKeys(t *testing.T) {
	root := scratch(t)
	var iv *InvalidValueError
	if err := Set(root, "run.verify", "maybe"); !errors.As(err, &iv) {
		t.Errorf("run.verify=maybe: %v", err)
	}
	if err := Set(root, "effort.verify", "huge"); !errors.As(err, &iv) {
		t.Errorf("effort.verify=huge: %v", err)
	}
	for k, v := range map[string]string{"run.verify": "false", "effort.verify": "high", "model.verify": "opus"} {
		if err := Set(root, k, v); err != nil {
			t.Fatalf("%s: %v", k, err)
		}
		if got, err := Get(root, k); err != nil || got.Value != v {
			t.Errorf("%s = %q, %v", k, got.Value, err)
		}
	}
}

// The config file is the operator's: set and remove edit its text, so comments,
// order and layout survive (B16: config set dropped this repository's comments).
func TestSetKeepsCommentsAndLayout(t *testing.T) {
	root := scratch(t)
	const body = `# The check below is slow.
run.gate-timeout = 30

[metrics]
  code = ["schemas/**"]
  # eval cases are tests
  test = [
    "plugin/evals/**",
  ]
  stacks = ["go"]

# The repository's checks.
[[check]]
  name = "go"
  paths = ["**"]
  run = "go test ./..."
`
	writeCfg(t, root, body)
	steps := []struct{ key, value, want string }{
		// a new key of a table written as dotted keys at the root
		{"run.git-config-ignore", "branch.*.github-pr-*", `run.gate-timeout = 30
run.git-config-ignore = ["branch.*.github-pr-*"]
`},
		// an existing key of a dotted table
		{"run.gate-timeout", "45", "run.gate-timeout = 45\n"},
		// an existing multi-line key of a [table]
		{"metrics.test", "x/**", "  # eval cases are tests\n  test = [\"x/**\"]\n  stacks = [\"go\"]\n"},
		// a new key of a [table]
		{"metrics.docs", "docs/**", "  stacks = [\"go\"]\n  docs = [\"docs/**\"]\n"},
		// a new root key, a new table
		{"language", "es", "run.gate-timeout = 45\nrun.git-config-ignore = [\"branch.*.github-pr-*\"]\nlanguage = \"es\"\n"},
		{"model.plan", "opus", "  run = \"go test ./...\"\n\n[model]\n  plan = \"opus\"\n"},
	}
	for _, s := range steps {
		if err := Set(root, s.key, s.value); err != nil {
			t.Fatalf("set %s: %v", s.key, err)
		}
		got := readCfg(t, root)
		if !strings.Contains(got, s.want) {
			t.Errorf("after set %s=%s, want %q in:\n%s", s.key, s.value, s.want, got)
		}
		for _, c := range []string{"# The check below is slow.", "# eval cases are tests", "# The repository's checks."} {
			if !strings.Contains(got, c) {
				t.Errorf("after set %s, comment %q lost:\n%s", s.key, c, got)
			}
		}
		if v, err := Get(root, s.key); err != nil || (v.Value != s.value && strings.Join(v.List, ",") != s.value) {
			t.Errorf("get %s = %+v, %v", s.key, v, err)
		}
	}
	if err := Set(root, "model.plan", ""); err != nil {
		t.Fatal(err)
	}
	if got := readCfg(t, root); strings.Contains(got, "[model]") || !strings.HasSuffix(got, "  run = \"go test ./...\"\n") {
		t.Errorf("removing the last key of [model] left:\n%s", got)
	}
	if err := Set(root, "metrics.test", ""); err != nil {
		t.Fatal(err)
	}
	if got := readCfg(t, root); strings.Contains(got, "x/**") || !strings.Contains(got, "# eval cases are tests") {
		t.Errorf("removing metrics.test left:\n%s", got)
	}
	cs, err := Checks(root)
	if err != nil || len(cs) != 1 || cs[0].Run != "go test ./..." {
		t.Errorf("checks = %+v, %v", cs, err)
	}
}

func TestAddChecksKeepsComments(t *testing.T) {
	root := scratch(t)
	writeCfg(t, root, "# mine\nlanguage = \"en\" # inline\n")
	if err := AddChecks(root, []CheckDef{{Name: "go", Paths: []string{"**"}, Run: "go test ./..."}}); err != nil {
		t.Fatal(err)
	}
	got := readCfg(t, root)
	if !strings.HasPrefix(got, "# mine\nlanguage = \"en\" # inline\n\n[[check]]\n") {
		t.Errorf("file = %q", got)
	}
	if cs, err := Checks(root); err != nil || len(cs) != 1 {
		t.Errorf("checks = %+v, %v", cs, err)
	}
}

func TestAgentKeys(t *testing.T) {
	for _, p := range []string{"plan", "work", "review", "gate-review", "verify"} {
		root, k := scratch(t), "agent."+p
		if v, err := Get(root, k); err != nil || v.Value != "claude" || v.Source != SourceDefault {
			t.Errorf("%s default: %+v %v", k, v, err)
		}
		if err := Set(root, k, "claude"); err != nil {
			t.Fatal(err)
		}
		if v, _ := Get(root, k); v.Source != SourceFile || !strings.Contains(readCfg(t, root), "[agent]") {
			t.Errorf("%s file: %+v", k, v)
		}
		before := readCfg(t, root)
		var iv *InvalidValueError
		if p == "review" {
			if err := Set(root, k, "copilot"); err != nil {
				t.Errorf("%s set copilot: %v", k, err)
			}
			if v, _ := Get(root, k); v.Value != "copilot" || v.Source != SourceFile {
				t.Errorf("%s copilot: %+v", k, v)
			}
			if err := Set(root, k, "grok"); !errors.As(err, &iv) || err.Error() != "invalid value \"grok\" for agent.review: want one of claude, copilot" {
				t.Errorf("%s set grok: %v", k, err)
			}
			if err := Set(root, k, "claude"); err != nil {
				t.Fatal(err)
			}
			before = readCfg(t, root)
			t.Setenv(EnvVar(k), "copilot")
			if v, err := Get(root, k); err != nil || v.Value != "copilot" || v.Source != SourceEnv {
				t.Errorf("%s env copilot: %+v %v", k, v, err)
			}
			t.Setenv(EnvVar(k), "")
			writeCfg(t, root, "[agent]\nreview = \"copilot\"\n")
			if v, err := Get(root, k); err != nil || v.Value != "copilot" {
				t.Errorf("%s file copilot: %+v %v", k, v, err)
			}
			writeCfg(t, root, before)
			continue
		}
		if err := Set(root, k, "copilot"); !errors.As(err, &iv) || err.Error() != "invalid value \"copilot\" for "+k+": want one of claude" {
			t.Errorf("%s set copilot: %v", k, err)
		}
		if readCfg(t, root) != before {
			t.Errorf("%s: refused set changed the file", k)
		}
		ev := EnvVar(k)
		t.Setenv(ev, "claude")
		if v, _ := Get(root, k); v.Source != SourceEnv {
			t.Errorf("%s env: %+v", k, v)
		}
		t.Setenv(ev, "copilot")
		var se *SourceError
		if _, err := Get(root, k); !errors.As(err, &se) || se.Source != ev {
			t.Errorf("%s bad env: %v", k, err)
		}
		if _, err := List(root); !errors.As(err, &se) {
			t.Errorf("%s bad env list: %v", k, err)
		}
		t.Setenv(ev, "")
		writeCfg(t, root, "[agent]\n"+p+" = \"copilot\"\n")
		if _, err := Get(root, k); !errors.As(err, &se) {
			t.Errorf("%s bad file: %v", k, err)
		}
	}
	if EnvVar("agent.gate-review") != "VLOOP_AGENT_GATE_REVIEW" {
		t.Error("env var name")
	}
}

func TestCopilotCreditKey(t *testing.T) {
	root := scratch(t)
	k := "copilot.credit-usd"
	if v, err := Get(root, k); err != nil || v.Value != "0.0066667" || v.Source != SourceDefault {
		t.Fatalf("default: %+v %v", v, err)
	}
	if EnvVar(k) != "VLOOP_COPILOT_CREDIT_USD" {
		t.Errorf("env var = %s", EnvVar(k))
	}
	writeCfg(t, root, "language = \"en\"\n")
	before := readCfg(t, root)
	for _, bad := range []string{"0", "0.0", "-1", "abc"} {
		var iv *InvalidValueError
		if err := Set(root, k, bad); !errors.As(err, &iv) {
			t.Errorf("set %q: %v", bad, err)
		}
		if readCfg(t, root) != before {
			t.Errorf("refused set %q changed the file", bad)
		}
		t.Setenv(EnvVar(k), bad)
		var se *SourceError
		if _, err := Get(root, k); !errors.As(err, &se) || se.Source != EnvVar(k) {
			t.Errorf("env %q: %v", bad, err)
		}
		t.Setenv(EnvVar(k), "")
	}
	t.Setenv(EnvVar(k), "0.02")
	if v, err := Get(root, k); err != nil || v.Value != "0.02" || v.Source != SourceEnv {
		t.Errorf("env: %+v %v", v, err)
	}
	t.Setenv(EnvVar(k), "")
	if err := Set(root, k, "0.01"); err != nil {
		t.Fatal(err)
	}
	if v, err := Get(root, k); err != nil || v.Value != "0.01" || v.Source != SourceFile {
		t.Errorf("file: %+v %v", v, err)
	}
	writeCfg(t, root, "[copilot]\n  credit-usd = -2\n")
	var se *SourceError
	if _, err := Get(root, k); !errors.As(err, &se) {
		t.Errorf("bad file: %v", err)
	}
}

func TestSecondReviewKeys(t *testing.T) {
	root := t.TempDir()
	for k, want := range map[string]string{"run.second-review": "off", "agent.second-review": "claude", "model.second-review": "sonnet", "effort.second-review": ""} {
		if v, err := Get(root, k); err != nil || v.Value != want {
			t.Errorf("%s = %q, %v; want %q", k, v.Value, err, want)
		}
	}
	if err := Set(root, "run.second-review", "always"); err == nil {
		t.Error("run.second-review always accepted")
	}
	if err := Set(root, "agent.second-review", "codex"); err == nil {
		t.Error("agent.second-review codex accepted")
	}
}
