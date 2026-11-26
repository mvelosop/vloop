package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
)

type result struct {
	code   int
	stdout string
	stderr string
}

func main() {
	root, err := os.MkdirTemp("", "vloop-config-kinds-")
	must(err)
	defer os.RemoveAll(root)
	binary := filepath.Join(root, "vloop")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-o", binary, "./cmd/vloop")
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	must(build.Run())
	repo := filepath.Join(root, "repo")
	must(os.MkdirAll(filepath.Join(repo, ".git"), 0o755))
	must(os.MkdirAll(filepath.Join(repo, ".vloop"), 0o755))
	configPath := filepath.Join(repo, ".vloop", "config.toml")
	initial := "[agent]\nwork = \"claude\"\nreview = \"claude\"\n[model]\nwork = \"sonnet\"\nreview = \"sonnet\"\n"
	must(os.WriteFile(configPath, []byte(initial), 0o644))

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			http.NotFound(w, r)
			return
		}
		calls.Add(1)
		fmt.Fprint(w, `{"data":[{"id":"mai-code-2","model_picker_enabled":true,"policy":{"state":"enabled"}},{"id":"gpt-6.1-terra","model_picker_enabled":true},{"id":"claude-fable-5.1","model_picker_enabled":true,"policy":{"state":"disabled"}}]}`)
	}))
	defer server.Close()
	env := map[string]string{
		"COPILOT_GITHUB_TOKEN":  "",
		"GH_TOKEN":              "stub-token",
		"GITHUB_TOKEN":          "",
		"VLOOP_COPILOT_API_URL": server.URL,
		"VLOOP_GITHUB_API_URL":  server.URL,
	}

	r := invoke(binary, repo, env, "config", "set", "--kinds", "work,review", "--agent", "copilot", "--model", "mai-code-2", "--effort", "high")
	equal(r.code, 0, "valid kinds exit")
	equal(r.stdout, "", "valid kinds stdout")
	equal(r.stderr, "", "valid kinds stderr")
	equal(calls.Load(), int32(1), "one models call")
	for _, key := range []string{"agent.work", "agent.review"} {
		r = invoke(binary, repo, env, "config", "get", key)
		equal(r.stdout, "copilot\n", key)
	}
	for _, key := range []string{"model.work", "model.review"} {
		r = invoke(binary, repo, env, "config", "get", key)
		equal(r.stdout, "mai-code-2\n", key)
	}
	for _, key := range []string{"effort.work", "effort.review"} {
		r = invoke(binary, repo, env, "config", "get", key)
		equal(r.stdout, "high\n", key)
	}

	assertRefusal := func(args []string, want string) {
		before, err := os.ReadFile(configPath)
		must(err)
		r := invoke(binary, repo, env, append([]string{"config", "set"}, args...)...)
		equal(r.code, 2, strings.Join(args, " ")+" exit")
		equal(r.stdout, "", strings.Join(args, " ")+" stdout")
		equal(r.stderr, want, strings.Join(args, " ")+" stderr")
		after, err := os.ReadFile(configPath)
		must(err)
		equal(string(after), string(before), strings.Join(args, " ")+" wrote config")
	}
	assertRefusal([]string{"--kinds", "work,beta", "--model", "mai-code-2"}, "vloop: unknown session kind beta: want plan, gate-review, work, review, second-review, verify or review-brief\n")
	assertRefusal([]string{"--kinds", "work"}, "vloop: --kinds needs --agent, --model or --effort\n")
	assertRefusal([]string{"--kinds", "work", "model.review", "mai-code-2"}, "vloop: --kinds takes no key or value: use --agent, --model and --effort\n")
	assertRefusal([]string{"--kinds", "work,review", "--model", "claude-fable-5.1"}, "vloop: claude-fable-5.1 is disabled by this account's Copilot policy: vloop models copilot lists the enabled ones\n")

	calls.Store(0)
	envAlias := copyMap(env)
	envAlias["VLOOP_MODEL_WORK"] = "sonnet"
	before, err := os.ReadFile(configPath)
	must(err)
	r = invoke(binary, repo, envAlias, "config", "set", "--kinds", "work", "--model", "")
	equal(r.code, 2, "resolved alias exit")
	equal(r.stderr, "vloop: sonnet is a Claude alias, and agent.work is copilot: vloop models copilot lists the Copilot ids\n", "resolved alias stderr")
	after, err := os.ReadFile(configPath)
	must(err)
	equal(string(after), string(before), "resolved alias changed config")
	equal(calls.Load(), int32(0), "resolved alias did not call models")
	delete(envAlias, "VLOOP_MODEL_WORK")

	r = invoke(binary, repo, envAlias, "config", "set", "--kinds", "work", "--agent", "")
	equal(r.code, 0, "remove agent exit")
	equal(r.stderr, "", "remove agent refused")
	r = invoke(binary, repo, envAlias, "config", "get", "agent.work")
	equal(r.stdout, "claude\n", "remove agent value")

	r = invoke(binary, repo, envAlias, "config", "set", "--local", "--kinds", "review", "--effort", "max")
	equal(r.code, 0, "local kinds exit")
	equal(r.stderr, "", "local kinds stderr")
	r = invoke(binary, repo, envAlias, "config", "get", "effort.review", "--json")
	if !strings.Contains(r.stdout, `"source":"local"`) {
		fail("local --kinds value did not resolve locally: %s", r.stdout)
	}
}

func copyMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func invoke(binary, dir string, overrides map[string]string, args ...string) result {
	cmd := exec.Command(binary, append([]string{"-C", dir}, args...)...)
	cmd.Env = mergedEnv(overrides)
	var out, errOut strings.Builder
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	code := 0
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			code = exit.ExitCode()
		} else {
			fail("%v", err)
		}
	}
	return result{code, out.String(), errOut.String()}
}

func mergedEnv(overrides map[string]string) []string {
	env := os.Environ()
	for key, value := range overrides {
		kept := env[:0]
		prefix := strings.ToUpper(key + "=")
		for _, item := range env {
			if !strings.HasPrefix(strings.ToUpper(item), prefix) {
				kept = append(kept, item)
			}
		}
		env = append(kept, key+"="+value)
	}
	return env
}

func must(err error) {
	if err != nil {
		fail("%v", err)
	}
}

func equal(got, want any, label string) {
	if got != want {
		fail("%s: got %#v, want %#v", label, got, want)
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
