package main

import (
	"encoding/json"
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
	root, err := os.MkdirTemp("", "vloop-copilot-models-")
	must(err)
	defer os.RemoveAll(root)

	binary := filepath.Join(root, "vloop")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-o", binary, "./cmd/vloop")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	must(build.Run())

	repo := filepath.Join(root, "repo")
	must(os.MkdirAll(filepath.Join(repo, ".git"), 0o755))

	failModels, failPlan := false, false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer stub-token" {
			http.Error(w, "bad authorization", http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/models":
			if r.Header.Get("Copilot-Integration-Id") != "copilot-developer-cli" {
				http.Error(w, "bad integration id", http.StatusBadRequest)
				return
			}
			if failModels {
				http.Error(w, "unavailable", http.StatusServiceUnavailable)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"data":[{"id":"mai-code-2","model_picker_enabled":true,"policy":{"state":"enabled"}},{"id":"claude-fable-5.1","model_picker_enabled":true,"policy":{"state":"disabled"}},{"id":"text-embedding-9","model_picker_enabled":false},{"id":"gpt-6.1-terra","model_picker_enabled":true}]}`)
		case "/copilot_internal/user":
			if failPlan {
				http.Error(w, "unavailable", http.StatusServiceUnavailable)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"copilot_plan":"individual_pro"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	env := map[string]string{
		"COPILOT_GITHUB_TOKEN":  "",
		"GH_TOKEN":              "stub-token",
		"GITHUB_TOKEN":          "",
		"VLOOP_COPILOT_API_URL": server.URL,
		"VLOOP_GITHUB_API_URL":  server.URL,
	}
	wantText := "plan  individual_pro\nclaude-fable-5.1  disabled\ngpt-6.1-terra     -\nmai-code-2        enabled\n"
	r := invoke(binary, repo, env, "models", "copilot")
	equal(r.code, 0, "models exit")
	equal(r.stdout, wantText, "models text")
	equal(r.stderr, "", "models stderr")

	r = invoke(binary, repo, env, "models", "copilot", "--all")
	equal(r.code, 0, "--all exit")
	equal(r.stdout, wantText+"text-embedding-9  -\n", "--all text")

	r = invoke(binary, repo, env, "models", "copilot", "--json")
	equal(r.code, 0, "json exit")
	equal(r.stderr, "", "json stderr")
	var doc struct {
		Schema string  `json:"schema"`
		Plan   *string `json:"plan"`
		Models []struct {
			ID     string `json:"id"`
			Policy string `json:"policy"`
			Picker bool   `json:"picker"`
		} `json:"models"`
	}
	must(json.Unmarshal([]byte(r.stdout), &doc))
	equal(doc.Schema, "copilot-models/v1", "json schema")
	equal(doc.Plan == nil, false, "json plan presence")
	equal(*doc.Plan, "individual_pro", "json plan")
	equal(len(doc.Models), 3, "json picker count")
	equal(doc.Models[0].ID, "claude-fable-5.1", "json sort")
	equal(doc.Models[1].Policy, "", "json null policy")

	for _, args := range [][]string{{"models"}, {"models", "claude"}} {
		r = invoke(binary, repo, env, args...)
		equal(r.code, 2, "models usage exit")
		equal(r.stdout, "", "models usage stdout")
		equal(r.stderr, "vloop: models lists Copilot models only: vloop models copilot\n", "models usage stderr")
	}

	failPlan = true
	r = invoke(binary, repo, env, "models", "copilot", "--json")
	equal(r.code, 0, "unknown plan exit")
	must(json.Unmarshal([]byte(r.stdout), &doc))
	equal(doc.Plan == nil, true, "unknown plan json")
	equal(doc.Models[0].ID, "claude-fable-5.1", "unknown plan models")
	failPlan = false

	failModels = true
	r = invoke(binary, repo, env, "models", "copilot")
	equal(r.code, 1, "models failure exit")
	equal(r.stdout, "", "models failure stdout")
	equal(r.stderr, "vloop: cannot list Copilot models: 503 Service Unavailable\n", "models failure stderr")

	failModels = false
	shim := filepath.Join(root, "bin")
	must(os.MkdirAll(shim, 0o755))
	gh := filepath.Join(shim, "gh")
	if runtime.GOOS == "windows" {
		gh += ".cmd"
		must(os.WriteFile(gh, []byte("@echo off\r\nexit /b 1\r\n"), 0o755))
		env["PATH"] = shim + string(os.PathListSeparator) + os.Getenv("PATH")
	} else {
		must(os.WriteFile(gh, []byte("#!/bin/sh\nexit 1\n"), 0o755))
		env["PATH"] = shim + string(os.PathListSeparator) + os.Getenv("PATH")
	}
	env["GH_TOKEN"] = ""
	r = invoke(binary, repo, env, "models", "copilot")
	equal(r.code, 1, "no login exit")
	equal(r.stdout, "", "no login stdout")
	equal(r.stderr, "vloop: no GitHub login reaches Copilot: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login\n", "no login stderr")
}

func invoke(binary, dir string, overrides map[string]string, args ...string) result {
	argv := append([]string{"-C", dir}, args...)
	cmd := exec.Command(binary, argv...)
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
		prefix := key + "="
		kept := env[:0]
		for _, item := range env {
			if !strings.HasPrefix(strings.ToUpper(item), strings.ToUpper(prefix)) {
				kept = append(kept, item)
			}
		}
		env = append(kept, prefix+value)
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
