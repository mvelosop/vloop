package cli

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func scratchRepo(t *testing.T) string {
	t.Helper()
	for _, v := range []string{"LANGUAGE", "MODEL_PLAN", "MODEL_WORK", "MODEL_REVIEW", "EFFORT_PLAN", "EFFORT_WORK", "EFFORT_REVIEW", "AGENT_PLAN", "AGENT_WORK", "AGENT_REVIEW", "AGENT_SECOND_REVIEW", "AGENT_GATE_REVIEW", "AGENT_VERIFY", "AGENT_REVIEW_BRIEF", "PLATFORMS", "SHELL", "AREAS", "RUN_MAX_ITERATIONS", "RUN_COST_CEILING", "RUN_MAX_ATTEMPTS", "RUN_STALL_LIMIT", "RUN_CONVERGENCE_MAX", "RUN_CONVERGENCE_MIN"} {
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
	want := "language=en (default)\nmodel.plan=opus (default)\nmodel.work=sonnet (default)\nmodel.review=sonnet (default)\nmodel.second-review=sonnet (default)\nmodel.gate-review=sonnet (default)\nmodel.verify=sonnet (default)\nmodel.review-brief=opus (default)\n" +
		"effort.plan= (default)\neffort.work= (default)\neffort.review= (default)\neffort.second-review= (default)\neffort.gate-review= (default)\neffort.verify= (default)\neffort.review-brief= (default)\n" +
		"agent.plan=claude (default)\nagent.work=claude (default)\nagent.review=claude (default)\nagent.second-review=claude (default)\nagent.gate-review=claude (default)\nagent.verify=claude (default)\nagent.review-brief=claude (default)\ncopilot.credit-usd=0.01 (default)\n" +
		"platforms= (default)\nshell=" + defaultShellForTest() + " (default)\nareas= (default)\n" +
		"metrics.stacks= (default)\nmetrics.code= (default)\nmetrics.test= (default)\nmetrics.docs= (default)\nmetrics.excluded= (default)\n" +
		"run.max-iterations=30 (default)\nrun.cost-ceiling=40 (default)\nrun.max-attempts=3 (default)\n" +
		"run.stall-limit=2 (default)\nrun.convergence-max=3.0 (default)\nrun.convergence-min=6 (default)\nrun.gate-timeout=15 (default)\nrun.gate-scratch= (default)\nrun.git-config-ignore= (default)\nrun.session-timeout=60 (default)\nrun.second-review=off (default)\nrun.verify=true (default)\nrun.keep-awake=on (default)\n"
	if code != 0 || out != want {
		t.Fatalf("code %d out %q", code, out)
	}
}

func TestConfigListJSONOrderAndNull(t *testing.T) {
	d := scratchRepo(t)
	_, out, _ := run(t, "-C", d, "config", "list", "--json")
	want := `{"language":{"value":"en","source":"default"},"model.plan":{"value":"opus","source":"default"},` +
		`"model.work":{"value":"sonnet","source":"default"},"model.review":{"value":"sonnet","source":"default"},` +
		`"model.second-review":{"value":"sonnet","source":"default"},"model.gate-review":{"value":"sonnet","source":"default"},"model.verify":{"value":"sonnet","source":"default"},"model.review-brief":{"value":"opus","source":"default"},` +
		`"effort.plan":{"value":null,"source":"default"},"effort.work":{"value":null,"source":"default"},` +
		`"effort.review":{"value":null,"source":"default"},"effort.second-review":{"value":null,"source":"default"},"effort.gate-review":{"value":null,"source":"default"},"effort.verify":{"value":null,"source":"default"},"effort.review-brief":{"value":null,"source":"default"},` +
		`"agent.plan":{"value":"claude","source":"default"},"agent.work":{"value":"claude","source":"default"},"agent.review":{"value":"claude","source":"default"},"agent.second-review":{"value":"claude","source":"default"},"agent.gate-review":{"value":"claude","source":"default"},"agent.verify":{"value":"claude","source":"default"},"agent.review-brief":{"value":"claude","source":"default"},"copilot.credit-usd":{"value":"0.01","source":"default"},` +
		`"platforms":{"value":[],"source":"default"},"shell":{"value":"` + defaultShellForTest() + `","source":"default"},"areas":{"value":[],"source":"default"},` +
		`"metrics.stacks":{"value":[],"source":"default"},"metrics.code":{"value":[],"source":"default"},` +
		`"metrics.test":{"value":[],"source":"default"},"metrics.docs":{"value":[],"source":"default"},` +
		`"metrics.excluded":{"value":[],"source":"default"},` +
		`"run.max-iterations":{"value":"30","source":"default"},"run.cost-ceiling":{"value":"40","source":"default"},` +
		`"run.max-attempts":{"value":"3","source":"default"},"run.stall-limit":{"value":"2","source":"default"},` +
		`"run.convergence-max":{"value":"3.0","source":"default"},"run.convergence-min":{"value":"6","source":"default"},` +
		`"run.gate-timeout":{"value":"15","source":"default"},"run.gate-scratch":{"value":[],"source":"default"},"run.git-config-ignore":{"value":[],"source":"default"},"run.session-timeout":{"value":"60","source":"default"},"run.second-review":{"value":"off","source":"default"},"run.verify":{"value":"true","source":"default"},` +
		`"run.keep-awake":{"value":"on","source":"default"}}` + "\n"
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
	_ = os.MkdirAll(filepath.Join(d, ".vloop"), 0o755)
	if err := os.WriteFile(filepath.Join(d, ".vloop", "config.toml"), []byte("language = \n[[[\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, e := run(t, "-C", d, "config", "list", "--json")
	if code != 1 || !strings.Contains(e, "config.toml") || strings.Count(out, "\n") != 1 || !strings.HasPrefix(out, `{"error":`) {
		t.Errorf("file: %d out %q err %q", code, out, e)
	}
}

func defaultShellForTest() string {
	if runtime.GOOS == "windows" {
		return "pwsh"
	}
	return "sh"
}

func TestConfigAreasAndShell(t *testing.T) {
	d := scratchRepo(t)
	if code, _, e := run(t, "-C", d, "config", "set", "areas", "cli,brief,config"); code != 0 || e != "" {
		t.Fatalf("set areas: %d %q", code, e)
	}
	if code, out, _ := run(t, "-C", d, "config", "get", "areas"); code != 0 || out != "cli,brief,config\n" {
		t.Errorf("get areas: %d %q", code, out)
	}
	if _, out, _ := run(t, "-C", d, "config", "get", "areas", "--json"); out != `{"key":"areas","value":["cli","brief","config"],"source":"file"}`+"\n" {
		t.Errorf("get areas --json: %q", out)
	}
	if _, out, _ := run(t, "-C", d, "config", "list"); !strings.Contains(out, "areas=cli,brief,config (file)\n") {
		t.Errorf("list: %q", out)
	}
	before, _ := os.ReadFile(filepath.Join(d, ".vloop", "config.toml"))
	code, out, e := run(t, "-C", d, "config", "set", "areas", "cli,Bad_Area")
	if code != 2 || out != "" || !strings.HasPrefix(e, `vloop: invalid value "cli,Bad_Area" for areas: want `) {
		t.Errorf("bad area: %d %q %q", code, out, e)
	}
	if after, _ := os.ReadFile(filepath.Join(d, ".vloop", "config.toml")); string(after) != string(before) {
		t.Error("refused set changed the file")
	}
	run(t, "-C", d, "config", "set", "areas", "")
	if _, out, _ := run(t, "-C", d, "config", "get", "areas", "--json"); out != `{"key":"areas","value":[],"source":"default"}`+"\n" {
		t.Errorf("cleared: %q", out)
	}
	code, _, e = run(t, "-C", d, "config", "set", "shell", "zsh")
	if code != 2 || e != "vloop: invalid value \"zsh\" for shell: want one of sh, bash, pwsh, powershell, cmd\n" {
		t.Errorf("shell zsh: %d %q", code, e)
	}
	t.Setenv("VLOOP_SHELL", "cmd")
	if _, out, _ := run(t, "-C", d, "config", "get", "shell", "--json"); out != `{"key":"shell","value":"cmd","source":"env"}`+"\n" {
		t.Errorf("env shell: %q", out)
	}
}

func TestConfigPlatformsSharedOnlyRefusals(t *testing.T) {
	d := scratchRepo(t)
	if code, _, e := run(t, "-C", d, "config", "set", "platforms", "windows,macos"); code != 0 || e != "" {
		t.Fatalf("set platforms: %d %q", code, e)
	}
	if code, out, e := run(t, "-C", d, "config", "get", "platforms"); code != 0 || out != "windows,macos\n" || e != "" {
		t.Fatalf("get platforms: %d %q %q", code, out, e)
	}
	shared, err := os.ReadFile(filepath.Join(d, ".vloop", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	local := filepath.Join(d, ".vloop", "config.local.toml")
	if code, out, e := run(t, "-C", d, "config", "set", "platforms", "macos,beos"); code != 2 || out != "" ||
		e != "vloop: invalid value \"macos,beos\" for platforms: want a comma-separated list of linux, macos, windows\n" {
		t.Fatalf("invalid platforms: %d %q %q", code, out, e)
	}
	t.Setenv("VLOOP_PLATFORMS", "linux")
	if code, out, e := run(t, "-C", d, "config", "get", "language"); code != 1 || out != "" ||
		e != "vloop: VLOOP_PLATFORMS: platforms belongs in .vloop/config.toml — the platforms are the repository's, not one machine's\n" {
		t.Fatalf("environment platforms: %d %q %q", code, out, e)
	}
	t.Setenv("VLOOP_PLATFORMS", "")
	if err := os.WriteFile(local, []byte("platforms = [\"linux\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, out, e := run(t, "-C", d, "config", "get", "language"); code != 1 || out != "" ||
		e != "vloop: .vloop/config.local.toml: platforms belongs in .vloop/config.toml — the platforms are the repository's, not one machine's\n" {
		t.Fatalf("local file platforms: %d %q %q", code, out, e)
	}
	if err := os.WriteFile(local, []byte("shell = \"pwsh\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, out, e := run(t, "-C", d, "config", "get", "shell"); code != 1 || out != "" ||
		e != "vloop: .vloop/config.local.toml: shell belongs in .vloop/config.toml while platforms lists macos, windows — the checks and the gates are written for one shell\n" {
		t.Fatalf("local file shell: %d %q %q", code, out, e)
	}
	if err := os.Remove(local); err != nil {
		t.Fatal(err)
	}
	if code, out, e := run(t, "-C", d, "config", "set", "--local", "platforms", "linux"); code != 2 || out != "" ||
		e != "vloop: .vloop/config.local.toml: platforms belongs in .vloop/config.toml — the platforms are the repository's, not one machine's\n" {
		t.Fatalf("local platforms: %d %q %q", code, out, e)
	}
	if code, out, e := run(t, "-C", d, "config", "set", "--local", "shell", "pwsh"); code != 2 || out != "" ||
		e != "vloop: .vloop/config.local.toml: shell belongs in .vloop/config.toml while platforms lists macos, windows — the checks and the gates are written for one shell\n" {
		t.Fatalf("local shell: %d %q %q", code, out, e)
	}
	t.Setenv("VLOOP_SHELL", "sh")
	if code, out, e := run(t, "-C", d, "config", "get", "shell"); code != 1 || out != "" ||
		e != "vloop: VLOOP_SHELL: shell belongs in .vloop/config.toml while platforms lists macos, windows — the checks and the gates are written for one shell\n" {
		t.Fatalf("environment shell: %d %q %q", code, out, e)
	}
	after, err := os.ReadFile(filepath.Join(d, ".vloop", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(shared) {
		t.Errorf("refusals changed shared file: %q", after)
	}
	if _, err := os.Stat(local); !os.IsNotExist(err) {
		t.Errorf("refusals wrote local file: %v", err)
	}
}

func TestConfigRunKeys(t *testing.T) {
	d := scratchRepo(t)
	if code, _, e := run(t, "-C", d, "config", "set", "run.max-attempts", "0"); code != 2 || !strings.HasPrefix(e, `vloop: invalid value "0" for run.max-attempts`) {
		t.Errorf("invalid set: %d %q", code, e)
	}
	if code, _, _ := run(t, "-C", d, "config", "set", "run.cost-ceiling", "12.5"); code != 0 {
		t.Errorf("set cost-ceiling: %d", code)
	}
	if _, out, _ := run(t, "-C", d, "config", "get", "run.cost-ceiling"); out != "12.5\n" {
		t.Errorf("get: %q", out)
	}
	for _, k := range []string{"run.gate-timeout", "run.session-timeout"} {
		for _, bad := range []string{"0", "1.5", "-2", "x"} {
			if code, _, _ := run(t, "-C", d, "config", "set", k, bad); code != 2 {
				t.Errorf("set %s %s: %d, want 2", k, bad, code)
			}
		}
	}
	t.Setenv("VLOOP_RUN_GATE_TIMEOUT", "3")
	if _, out, _ := run(t, "-C", d, "config", "get", "run.gate-timeout"); out != "3\n" {
		t.Errorf("gate-timeout env: %q", out)
	}
	t.Setenv("VLOOP_RUN_GATE_TIMEOUT", "")
	t.Setenv("VLOOP_RUN_COST_CEILING", "7")
	if _, out, _ := run(t, "-C", d, "config", "get", "run.cost-ceiling", "--json"); out != `{"key":"run.cost-ceiling","value":"7","source":"env"}`+"\n" {
		t.Errorf("env: %q", out)
	}
	t.Setenv("VLOOP_RUN_COST_CEILING", "abc")
	if code, _, e := run(t, "-C", d, "config", "list"); code != 1 || !strings.Contains(e, "VLOOP_RUN_COST_CEILING") {
		t.Errorf("bad env: %d %q", code, e)
	}
}

func TestConfigChecksListAndSurviveSet(t *testing.T) {
	d := scratchRepo(t)
	body := "[[check]]\nname = \"api\"\npaths = [\"api/**\"]\nrun = \"sh api/test.sh\"\n\n" +
		"[[check]]\nname = \"web\"\npaths = [\"web/**\", \"shared/*.ts\"]\nrun = \"sh web/test.sh\"\n"
	_ = os.MkdirAll(filepath.Join(d, ".vloop"), 0o755)
	if err := os.WriteFile(filepath.Join(d, ".vloop", "config.toml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, errs := run(t, "-C", d, "config", "set", "effort.gate-review", "high"); code != 0 {
		t.Fatalf("set: %d %s", code, errs)
	}
	code, out, _ := run(t, "-C", d, "config", "list")
	tail := "check.api=sh api/test.sh (paths: api/**)\ncheck.web=sh web/test.sh (paths: web/**, shared/*.ts)\n"
	if code != 0 || !strings.HasSuffix(out, tail) {
		t.Fatalf("code %d out %q", code, out)
	}
}

func TestConfigBadCheckExit1(t *testing.T) {
	d := scratchRepo(t)
	_ = os.MkdirAll(filepath.Join(d, ".vloop"), 0o755)
	if err := os.WriteFile(filepath.Join(d, ".vloop", "config.toml"), []byte("[[check]]\nname = \"A\"\npaths = [\"**\"]\nrun = \"true\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, errs := run(t, "-C", d, "config", "list")
	if code != 1 || out != "" || !strings.HasPrefix(errs, "vloop: ") {
		t.Fatalf("code %d out %q err %q", code, out, errs)
	}
}

func TestConfigSetLocal(t *testing.T) {
	d := scratchRepo(t)
	local := filepath.Join(d, ".vloop", "config.local.toml")
	if code, _, _ := run(t, "-C", d, "config", "set", "--local", "agent.work", ""); code != 0 {
		t.Fatalf("absent remove: %d", code)
	}
	if _, err := os.Stat(filepath.Join(d, ".vloop")); err == nil {
		t.Fatal("absent remove created .vloop")
	}
	if code, _, errOut := run(t, "-C", d, "config", "set", "--local", "agent.work", "nope"); code != 2 || !strings.Contains(errOut, `invalid value "nope" for agent.work`) {
		t.Fatalf("invalid: %d %q", code, errOut)
	}
	if code, _, _ := run(t, "-C", d, "config", "set", "--local", "colour", "red"); code != 2 {
		t.Fatalf("unknown: %d", code)
	}
	if _, err := os.Stat(local); err == nil {
		t.Fatal("refused set wrote the file")
	}
	if err := os.MkdirAll(filepath.Dir(local), 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(local, []byte("# mine\nrun.max-iterations = 9\n"), 0o644)
	if code, out, errOut := run(t, "-C", d, "config", "set", "--local", "agent.work", "copilot"); code != 0 || out != "" || errOut != "" {
		t.Fatalf("set: %d %q %q", code, out, errOut)
	}
	b, _ := os.ReadFile(local)
	if !strings.Contains(string(b), "# mine") || !strings.Contains(string(b), "run.max-iterations = 9") || !strings.Contains(string(b), "copilot") {
		t.Fatalf("layout lost: %q", b)
	}
	if _, err := os.Stat(filepath.Join(d, ".vloop", "config.toml")); err == nil {
		t.Fatal("created shared file")
	}
	run(t, "-C", d, "config", "set", "--local", "agent.work", "")
	b, _ = os.ReadFile(local)
	if strings.Contains(string(b), "agent") {
		t.Fatalf("not removed: %q", b)
	}
}

func TestConfigSetOverrideNotice(t *testing.T) {
	d := scratchRepo(t)
	run(t, "-C", d, "config", "set", "--local", "agent.work", "claude")
	code, out, errOut := run(t, "-C", d, "config", "set", "agent.work", "copilot")
	if code != 0 || out != "" || errOut != "! agent.work stays claude on this machine: .vloop/config.local.toml overrides it\n" {
		t.Fatalf("local notice: %d %q %q", code, out, errOut)
	}
	t.Setenv("VLOOP_AGENT_WORK", "copilot")
	_, out, errOut = run(t, "-C", d, "config", "set", "agent.work", "claude")
	if out != "" || errOut != "! agent.work stays copilot on this machine: VLOOP_AGENT_WORK overrides it\n" {
		t.Fatalf("env notice: %q %q", out, errOut)
	}
	_, _, errOut = run(t, "-C", d, "config", "set", "--local", "agent.work", "claude")
	if !strings.Contains(errOut, "VLOOP_AGENT_WORK overrides it") {
		t.Fatalf("local env notice: %q", errOut)
	}
	t.Setenv("VLOOP_AGENT_WORK", "")
	if _, _, errOut = run(t, "-C", d, "config", "set", "language", "es"); errOut != "" {
		t.Fatalf("unexpected notice %q", errOut)
	}
}

func TestConfigPathLocal(t *testing.T) {
	d := scratchRepo(t)
	if _, out, _ := run(t, "-C", d, "config", "path", "--local"); out != ".vloop/config.local.toml\n" {
		t.Fatalf("%q", out)
	}
	if _, out, _ := run(t, "-C", d, "config", "path"); out != ".vloop/config.toml\n" {
		t.Fatalf("%q", out)
	}
}

func TestConfigSetCopilotModelRefusalsAndWarnings(t *testing.T) {
	mode := "ok"
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Header.Get("Authorization") != "Bearer stub-token" || r.Header.Get("Copilot-Integration-Id") != "copilot-developer-cli" {
			http.Error(w, "bad request", http.StatusUnauthorized)
			return
		}
		switch mode {
		case "failed":
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
		case "malformed":
			fmt.Fprint(w, `{`)
		case "missing-data":
			fmt.Fprint(w, `{}`)
		default:
			fmt.Fprint(w, `{"data":[
				{"id":"mai-code-2","model_picker_enabled":true,"policy":{"state":"enabled"}},
				{"id":"gpt-6.1-terra","model_picker_enabled":true},
				{"id":"claude-fable-5.1","model_picker_enabled":true,"policy":{"state":"disabled"}}]}`)
		}
	}))
	defer server.Close()
	t.Setenv("COPILOT_GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "stub-token")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("VLOOP_COPILOT_API_URL", server.URL)
	t.Setenv("VLOOP_GITHUB_API_URL", server.URL)

	d := scratchRepo(t)
	mustWriteConfig(t, d, "[agent]\nwork = \"copilot\"\n[model]\nwork = \"mai-code-2\"\n")
	assertConfigRefusal := func(value, want string) {
		t.Helper()
		path := filepath.Join(d, ".vloop", "config.toml")
		before, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		code, out, errOut := run(t, "-C", d, "config", "set", "model.work", value)
		if code != 2 || out != "" || errOut != "vloop: "+want+"\n" {
			t.Fatalf("%s: code=%d stdout=%q stderr=%q", value, code, out, errOut)
		}
		after, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(after) != string(before) {
			t.Fatalf("%s changed config from %q to %q", value, before, after)
		}
	}

	assertConfigRefusal("mai-code-3", "mai-code-3 is not a Copilot model this account lists: vloop models copilot lists them")
	assertConfigRefusal("claude-fable-5.1", "claude-fable-5.1 is disabled by this account's Copilot policy: vloop models copilot lists the enabled ones")
	aliasRequests := requests
	assertConfigRefusal("sonnet", "sonnet is a Claude alias, and agent.work is copilot: vloop models copilot lists the Copilot ids")
	if requests != aliasRequests {
		t.Fatal("Claude alias made a models request")
	}

	mode = "failed"
	code, out, errOut := run(t, "-C", d, "config", "set", "model.work", "gpt-6.1-terra")
	wantWarning := "! cannot check gpt-6.1-terra against Copilot's model list (503 Service Unavailable); set anyway\n"
	if code != 0 || out != "" || errOut != wantWarning {
		t.Fatalf("failed call: code=%d stdout=%q stderr=%q", code, out, errOut)
	}
	if _, out, _ := run(t, "-C", d, "config", "get", "model.work"); out != "gpt-6.1-terra\n" {
		t.Fatalf("failed call did not write: %q", out)
	}

	mode = "malformed"
	code, _, errOut = run(t, "-C", d, "config", "set", "model.work", "mai-code-2")
	if code != 0 || !strings.HasPrefix(errOut, "! cannot check mai-code-2 against Copilot's model list (") || !strings.HasSuffix(errOut, "); set anyway\n") {
		t.Fatalf("malformed call: code=%d stderr=%q", code, errOut)
	}
	mode = "missing-data"
	code, _, errOut = run(t, "-C", d, "config", "set", "model.work", "gpt-6.1-terra")
	if code != 0 || errOut != "! cannot check gpt-6.1-terra against Copilot's model list (cannot read JSON response: missing data); set anyway\n" {
		t.Fatalf("missing data: code=%d stderr=%q", code, errOut)
	}

	before := requests
	mode = "failed"
	code, _, errOut = run(t, "-C", d, "config", "set", "model.work", "")
	if code != 0 || errOut != "" || requests != before {
		t.Fatalf("removal: code=%d stderr=%q requests=%d", code, errOut, requests-before)
	}

	mode = "ok"
	code, _, errOut = run(t, "-C", d, "config", "set", "--local", "model.work", "mai-code-2")
	if code != 0 || errOut != "" {
		t.Fatalf("local: code=%d stderr=%q", code, errOut)
	}
	_, out, _ = run(t, "-C", d, "config", "get", "model.work", "--json")
	if !strings.Contains(out, `"value":"mai-code-2","source":"local"`) {
		t.Fatalf("local source: %q", out)
	}
}

func TestConfigSetCopilotModelNoLoginAndResolvedPrecedence(t *testing.T) {
	d := scratchRepo(t)
	mustWriteConfig(t, d, "[agent]\nwork = \"copilot\"\n[model]\nwork = \"mai-code-2\"\n")
	t.Setenv("COPILOT_GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("PATH", t.TempDir())
	code, _, errOut := run(t, "-C", d, "config", "set", "model.work", "gpt-6.1-terra")
	if code != 0 || errOut != "! cannot check gpt-6.1-terra: no GitHub login reaches Copilot\n" {
		t.Fatalf("no login: code=%d stderr=%q", code, errOut)
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		fmt.Fprint(w, `{"data":[{"id":"mai-code-2","policy":{"state":"enabled"}},{"id":"disabled","policy":{"state":"disabled"}}]}`)
	}))
	defer server.Close()
	t.Setenv("GH_TOKEN", "stub-token")
	t.Setenv("VLOOP_COPILOT_API_URL", server.URL)

	t.Setenv("VLOOP_AGENT_WORK", "claude")
	code, _, errOut = run(t, "-C", d, "config", "set", "model.work", "not-listed")
	if code != 0 || errOut != "" || requests != 0 {
		t.Fatalf("non-Copilot agent: code=%d stderr=%q requests=%d", code, errOut, requests)
	}

	t.Setenv("VLOOP_AGENT_WORK", "copilot")
	t.Setenv("VLOOP_MODEL_WORK", "disabled")
	code, _, errOut = run(t, "-C", d, "config", "set", "model.work", "mai-code-2")
	want := "vloop: disabled is disabled by this account's Copilot policy: vloop models copilot lists the enabled ones\n"
	if code != 2 || errOut != want {
		t.Fatalf("environment model precedence: code=%d stderr=%q", code, errOut)
	}

	t.Setenv("VLOOP_MODEL_WORK", "")
	code, _, errOut = run(t, "-C", d, "config", "set", "--local", "model.work", "disabled")
	if code != 2 || errOut != want {
		t.Fatalf("local disabled: code=%d stderr=%q", code, errOut)
	}
	code, _, errOut = run(t, "-C", d, "config", "set", "--local", "model.work", "mai-code-2")
	if code != 0 || errOut != "" {
		t.Fatalf("local enabled: code=%d stderr=%q", code, errOut)
	}
	code, _, errOut = run(t, "-C", d, "config", "set", "model.work", "not-listed")
	if code != 0 {
		t.Fatalf("local override should validate local model: code=%d stderr=%q", code, errOut)
	}
}

func mustWriteConfig(t *testing.T, root, body string) {
	t.Helper()
	path := filepath.Join(root, ".vloop", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
