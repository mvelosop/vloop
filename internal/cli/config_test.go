package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func scratchRepo(t *testing.T) string {
	t.Helper()
	for _, v := range []string{"LANGUAGE", "MODEL_PLAN", "MODEL_WORK", "MODEL_REVIEW", "EFFORT_PLAN", "EFFORT_WORK", "EFFORT_REVIEW"} {
		t.Setenv("VLOOP_"+v, "")
	}
	d := t.TempDir()
	if err := os.Mkdir(filepath.Join(d, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	return d
}

func TestConfigListDefaults(t *testing.T) {
	d := scratchRepo(t)
	code, out, _ := run(t, "-C", d, "config", "list")
	want := "language=en (default)\nmodel.plan=opus (default)\nmodel.work=sonnet (default)\nmodel.review=sonnet (default)\n" +
		"effort.plan= (default)\neffort.work= (default)\neffort.review= (default)\n"
	if code != 0 || out != want {
		t.Fatalf("code %d out %q", code, out)
	}
}

func TestConfigListJSONOrderAndNull(t *testing.T) {
	d := scratchRepo(t)
	_, out, _ := run(t, "-C", d, "config", "list", "--json")
	want := `{"language":{"value":"en","source":"default"},"model.plan":{"value":"opus","source":"default"},` +
		`"model.work":{"value":"sonnet","source":"default"},"model.review":{"value":"sonnet","source":"default"},` +
		`"effort.plan":{"value":null,"source":"default"},"effort.work":{"value":null,"source":"default"},` +
		`"effort.review":{"value":null,"source":"default"}}` + "\n"
	if out != want {
		t.Fatalf("got %s", out)
	}
}

func TestConfigGetSetPath(t *testing.T) {
	d := scratchRepo(t)
	if code, out, _ := run(t, "-C", d, "config", "get", "effort.plan"); code != 0 || out != "\n" {
		t.Errorf("unset effort: %d %q", code, out)
	}
	if code, out, e := run(t, "-C", d, "config", "set", "language", "es"); code != 0 || out != "" || e != "" {
		t.Errorf("set: %d %q %q", code, out, e)
	}
	_, out, _ := run(t, "-C", d, "config", "get", "language", "--json")
	if out != `{"key":"language","value":"es","source":"file"}`+"\n" {
		t.Errorf("get json: %q", out)
	}
	if _, out, _ := run(t, "-C", d, "config", "path"); out != ".vloop/config.toml\n" {
		t.Errorf("path: %q", out)
	}
	t.Setenv("VLOOP_LANGUAGE", "en")
	if _, out, _ := run(t, "-C", d, "config", "get", "language", "--json"); !strings.Contains(out, `"source":"env"`) {
		t.Errorf("env: %q", out)
	}
}

func TestConfigPathCreatesNothing(t *testing.T) {
	d := scratchRepo(t)
	run(t, "-C", d, "config", "path")
	if _, err := os.Stat(filepath.Join(d, ".vloop")); err == nil {
		t.Error("config path created .vloop")
	}
}

func TestConfigUsageErrorsExit2(t *testing.T) {
	d := scratchRepo(t)
	cases := map[string]string{
		"config get colour":          `vloop: unknown config key "colour"` + "\n",
		"config set colour red":      `vloop: unknown config key "colour"` + "\n",
		"config set language fr":     `vloop: invalid value "fr" for language: want one of en, es` + "\n",
		"config set effort.work bad": `vloop: invalid value "bad" for effort.work: want one of low, medium, high, xhigh, max` + "\n",
	}
	for cmd, wantErr := range cases {
		code, out, e := run(t, append([]string{"-C", d}, strings.Fields(cmd)...)...)
		if code != 2 || out != "" || e != wantErr {
			t.Errorf("%s: %d %q %q", cmd, code, out, e)
		}
	}
	for _, cmd := range []string{"config get", "config set language", "config nosuch"} {
		if code, out, _ := run(t, append([]string{"-C", d}, strings.Fields(cmd)...)...); code != 2 || out != "" {
			t.Errorf("%s: %d %q", cmd, code, out)
		}
	}
}

func TestConfigBadSourceExit1(t *testing.T) {
	d := scratchRepo(t)
	t.Setenv("VLOOP_LANGUAGE", "fr")
	code, _, e := run(t, "-C", d, "config", "get", "language")
	if code != 1 || !strings.Contains(e, "VLOOP_LANGUAGE") || strings.Count(e, "\n") != 1 {
		t.Errorf("env: %d %q", code, e)
	}
	t.Setenv("VLOOP_LANGUAGE", "")
	if err := os.MkdirAll(filepath.Join(d, ".vloop"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, ".vloop", "config.toml"), []byte("language = \n[[[\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, e := run(t, "-C", d, "config", "list", "--json")
	if code != 1 || !strings.Contains(e, "config.toml") || strings.Count(out, "\n") != 1 || !strings.HasPrefix(out, `{"error":`) {
		t.Errorf("file: %d out %q err %q", code, out, e)
	}
}
