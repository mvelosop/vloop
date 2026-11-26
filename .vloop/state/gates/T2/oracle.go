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
)

type result struct {
	code   int
	stdout string
	stderr string
}

func main() {
	root, err := os.MkdirTemp("", "vloop-config-copilot-")
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
	config := "[agent]\nwork = \"copilot\"\n[model]\nwork = \"mai-code-2\"\n"
	write := func(s string) { must(os.WriteFile(filepath.Join(repo, ".vloop", "config.toml"), []byte(s), 0o644)) }
	write(config)

	failModels := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer stub-token" {
			http.Error(w, "bad token", http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/models" {
			http.NotFound(w, r)
			return
		}
		if failModels {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		fmt.Fprint(w, `{"data":[{"id":"mai-code-2","model_picker_enabled":true,"policy":{"state":"enabled"}},{"id":"gpt-6.1-terra","model_picker_enabled":true},{"id":"claude-fable-5.1","model_picker_enabled":true,"policy":{"state":"disabled"}}]}`)
	}))
	defer server.Close()
	env := apiEnv(server.URL)

	assertRefusal := func(id, want string) {
		before, err := os.ReadFile(filepath.Join(repo, ".vloop", "config.toml"))
		must(err)
		r := invoke(binary, repo, env, "config", "set", "model.work", id)
		equal(r.code, 2, id+" exit")
		equal(r.stdout, "", id+" stdout")
		equal(r.stderr, want, id+" stderr")
		after, err := os.ReadFile(filepath.Join(repo, ".vloop", "config.toml"))
		must(err)
		equal(string(after), string(before), id+" changed config")
	}
	assertRefusal("mai-code-3", "vloop: mai-code-3 is not a Copilot model this account lists: vloop models copilot lists them\n")
	assertRefusal("claude-fable-5.1", "vloop: claude-fable-5.1 is disabled by this account's Copilot policy: vloop models copilot lists the enabled ones\n")
	assertRefusal("sonnet", "vloop: sonnet is a Claude alias, and agent.work is copilot: vloop models copilot lists the Copilot ids\n")

	failModels = true
	r := invoke(binary, repo, env, "config", "set", "model.work", "gpt-6.1-terra")
	equal(r.code, 0, "unreachable exit")
	equal(r.stdout, "", "unreachable stdout")
	equal(r.stderr, "! cannot check gpt-6.1-terra against Copilot's model list (503 Service Unavailable); set anyway\n", "unreachable stderr")
	r = invoke(binary, repo, env, "config", "get", "model.work")
	equal(r.stdout, "gpt-6.1-terra\n", "unreachable value")

	noLogin := apiEnv(server.URL)
	noLogin["GH_TOKEN"] = ""
	shim := filepath.Join(root, "bin")
	must(os.MkdirAll(shim, 0o755))
	gh := filepath.Join(shim, "gh")
	if runtime.GOOS == "windows" {
		gh += ".cmd"
		must(os.WriteFile(gh, []byte("@echo off\r\nexit /b 1\r\n"), 0o755))
	} else {
		must(os.WriteFile(gh, []byte("#!/bin/sh\nexit 1\n"), 0o755))
	}
	noLogin["PATH"] = shim + string(os.PathListSeparator) + os.Getenv("PATH")
	r = invoke(binary, repo, noLogin, "config", "set", "model.work", "mai-code-2")
	equal(r.code, 0, "no login exit")
	equal(r.stderr, "! cannot check mai-code-2: no GitHub login reaches Copilot\n", "no login stderr")
	r = invoke(binary, repo, noLogin, "config", "get", "model.work")
	equal(r.stdout, "mai-code-2\n", "no login value")

	failModels = true
	r = invoke(binary, repo, env, "config", "set", "model.work", "")
	equal(r.code, 0, "remove exit")
	equal(r.stderr, "", "remove checked model")
	r = invoke(binary, repo, env, "config", "get", "model.work")
	equal(r.stdout, "sonnet\n", "remove value")

	failModels = false
	r = invoke(binary, repo, env, "config", "set", "--local", "model.work", "mai-code-2")
	equal(r.code, 0, "local exit")
	equal(r.stderr, "", "local stderr")
	r = invoke(binary, repo, env, "config", "get", "model.work", "--json")
	if !strings.Contains(r.stdout, `"source":"local"`) {
		fail("local value did not resolve from local source: %s", r.stdout)
	}
}

func apiEnv(url string) map[string]string {
	return map[string]string{
		"COPILOT_GITHUB_TOKEN":  "",
		"GH_TOKEN":              "stub-token",
		"GITHUB_TOKEN":          "",
		"VLOOP_COPILOT_API_URL": url,
		"VLOOP_GITHUB_API_URL":  url,
	}
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
