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
	"sync/atomic"
)

type result struct {
	code   int
	stdout string
	stderr string
}

func main() {
	root, err := os.MkdirTemp("", "vloop-doctor-copilot-models-")
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
	must(os.MkdirAll(filepath.Join(repo, ".vloop"), 0o755))
	must(os.MkdirAll(filepath.Join(repo, ".git"), 0o755))

	failModels := false
	var modelCalls, planCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/models":
			modelCalls.Add(1)
			if failModels {
				http.Error(w, "unavailable", http.StatusServiceUnavailable)
				return
			}
			fmt.Fprint(w, `{"data":[{"id":"mai-code-2","model_picker_enabled":true,"policy":{"state":"enabled"}},{"id":"claude-fable-5.1","model_picker_enabled":true,"policy":{"state":"disabled"}}]}`)
		case "/copilot_internal/user":
			planCalls.Add(1)
			fmt.Fprint(w, `{"copilot_plan":"individual_pro"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	bin := filepath.Join(root, "bin")
	must(os.MkdirAll(bin, 0o755))
	writeShims(bin)
	env := map[string]string{
		"COPILOT_GITHUB_TOKEN":  "",
		"GH_TOKEN":              "stub-token",
		"GITHUB_TOKEN":          "",
		"PATH":                  bin + string(os.PathListSeparator) + os.Getenv("PATH"),
		"VLOOP_COPILOT_API_URL": server.URL,
		"VLOOP_GITHUB_API_URL":  server.URL,
		"NO_COLOR":              "1",
	}
	setModels := func(work, review string) {
		doc := "[agent]\nwork = \"copilot\"\nreview = \"copilot\"\n[model]\nwork = " + quoted(work) + "\nreview = " + quoted(review) + "\n"
		must(os.WriteFile(filepath.Join(repo, ".vloop", "config.toml"), []byte(doc), 0o644))
	}

	setModels("mai-code-2", "mai-code-2")
	r := invoke(binary, repo, env, "doctor")
	contains(r.stdout, "✓ copilot models 1 id(s) checked · plan individual_pro\n", "doctor pass row")
	equal(modelCalls.Load(), int32(1), "one models request")
	equal(planCalls.Load(), int32(1), "one plan request")
	r = invoke(binary, repo, env, "doctor", "--json")
	equal(r.code, 1, "doctor unrelated problems exit")
	var doc struct {
		Checks []struct {
			Name    string `json:"name"`
			Result  string `json:"result"`
			Message string `json:"message"`
		} `json:"checks"`
	}
	must(json.Unmarshal([]byte(r.stdout), &doc))
	found := false
	for _, check := range doc.Checks {
		if check.Name == "copilot models" {
			found = true
			equal(check.Result, "pass", "doctor json result")
			equal(check.Message, "1 id(s) checked · plan individual_pro", "doctor json message")
		}
	}
	equal(found, true, "doctor json row")

	setModels("mai-code-3", "claude-fable-5.1")
	r = invoke(binary, repo, env, "doctor")
	contains(r.stdout, "✗ copilot models model.work mai-code-3 is not listed; model.review claude-fable-5.1 is disabled · plan individual_pro — vloop models copilot lists them\n", "doctor problem row")

	failModels = true
	setModels("mai-code-2", "mai-code-2")
	r = invoke(binary, repo, env, "doctor")
	contains(r.stdout, "! copilot models cannot reach Copilot's model list (503 Service Unavailable); ids not checked\n", "doctor warning row")

	setModels("sonnet", "mai-code-2")
	r = invoke(binary, repo, env, "doctor")
	contains(r.stdout, "✗ copilot models model.work sonnet is a Claude alias; ids not checked against the list\n", "doctor alias row")

	env["GH_TOKEN"] = ""
	noAuthBin := filepath.Join(root, "no-login")
	must(os.MkdirAll(noAuthBin, 0o755))
	writeShims(noAuthBin)
	env["PATH"] = noAuthBin + string(os.PathListSeparator) + os.Getenv("PATH")
	setModels("mai-code-2", "mai-code-2")
	beforeModels, beforePlan := modelCalls.Load(), planCalls.Load()
	r = invoke(binary, repo, env, "doctor")
	if strings.Contains(r.stdout, "copilot models") {
		fail("doctor added models row without login: %s", r.stdout)
	}
	equal(modelCalls.Load(), beforeModels, "no-login models request")
	equal(planCalls.Load(), beforePlan, "no-login plan request")
}

func quoted(s string) string { return `"` + s + `"` }

func writeShims(dir string) {
	name := "copilot"
	content := "#!/bin/sh\nexit 0\n"
	if runtime.GOOS == "windows" {
		name += ".cmd"
		content = "@echo off\r\nexit /b 0\r\n"
	}
	must(os.WriteFile(filepath.Join(dir, name), []byte(content), 0o755))
	name = "gh"
	content = "#!/bin/sh\nexit 1\n"
	if runtime.GOOS == "windows" {
		name += ".cmd"
		content = "@echo off\r\nexit /b 1\r\n"
	}
	must(os.WriteFile(filepath.Join(dir, name), []byte(content), 0o755))
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

func contains(got, part, label string) {
	if !strings.Contains(got, part) {
		fail("%s: missing %#v in %#v", label, part, got)
	}
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
