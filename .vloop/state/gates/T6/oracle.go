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
	root, err := os.MkdirTemp("", "vloop-copilot-close-")
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
	must(os.MkdirAll(repo, 0o755))
	gitInit := exec.Command("git", "init", "-q", repo)
	must(gitInit.Run())
	r := invoke(binary, repo, nil, "init")
	equal(r.code, 0, "vloop init exit")

	bin := filepath.Join(root, "bin")
	must(os.MkdirAll(bin, 0o755))
	writeCopilotShim(bin)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/models":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"data":[{"id":"mai-code-2","model_picker_enabled":true,"policy":{"state":"enabled"}},{"id":"gpt-6.1-terra","model_picker_enabled":true},{"id":"claude-fable-5.1","model_picker_enabled":true,"policy":{"state":"disabled"}},{"id":"text-embedding-9","model_picker_enabled":false}]}`)
		case "/copilot_internal/user":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"copilot_plan":"individual_pro"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	env := map[string]string{
		"COPILOT_GITHUB_TOKEN":  "",
		"GH_TOKEN":              "stub-token",
		"GITHUB_TOKEN":          "",
		"PATH":                  bin + string(os.PathListSeparator) + os.Getenv("PATH"),
		"VLOOP_COPILOT_API_URL": server.URL,
		"VLOOP_GITHUB_API_URL":  server.URL,
		"NO_COLOR":              "1",
	}
	combined := ""
	record := func(got result) result {
		combined += got.stdout + got.stderr
		return got
	}

	r = record(invoke(binary, repo, env, "models", "copilot"))
	equal(r.code, 0, "models exit")
	equal(r.stdout, "plan  individual_pro\nclaude-fable-5.1  disabled\ngpt-6.1-terra     -\nmai-code-2        enabled\n", "worked models")
	r = record(invoke(binary, repo, env, "models", "copilot", "--all"))
	equal(r.code, 0, "models --all exit")
	equal(r.stdout, "plan  individual_pro\nclaude-fable-5.1  disabled\ngpt-6.1-terra     -\nmai-code-2        enabled\ntext-embedding-9  -\n", "worked --all")

	r = record(invoke(binary, repo, env, "config", "set", "--kinds", "work,review", "--agent", "copilot", "--model", "mai-code-2"))
	equal(r.code, 0, "worked kinds exit")
	equal(r.stdout+r.stderr, "", "worked kinds output")
	r = record(invoke(binary, repo, env, "config", "get", "model.review"))
	equal(r.stdout, "mai-code-2\n", "worked get review")

	r = record(invoke(binary, repo, env, "config", "set", "model.work", "mai-code-3"))
	equal(r.code, 2, "worked missing exit")
	equal(r.stderr, "vloop: mai-code-3 is not a Copilot model this account lists: vloop models copilot lists them\n", "worked missing")

	beforeShared, _ := os.ReadFile(filepath.Join(repo, ".vloop", "config.toml"))
	beforeLocal, _ := os.ReadFile(filepath.Join(repo, ".vloop", "config.local.toml"))
	r = record(invoke(binary, repo, env, "config", "set", "--kinds", "work,review", "--model", "claude-fable-5.1"))
	equal(r.code, 2, "worked disabled exit")
	equal(r.stderr, "vloop: claude-fable-5.1 is disabled by this account's Copilot policy: vloop models copilot lists the enabled ones\n", "worked disabled")
	r = record(invoke(binary, repo, env, "config", "get", "model.work"))
	equal(r.stdout, "mai-code-2\n", "worked atomic")
	afterShared, _ := os.ReadFile(filepath.Join(repo, ".vloop", "config.toml"))
	afterLocal, _ := os.ReadFile(filepath.Join(repo, ".vloop", "config.local.toml"))
	equal(string(afterShared), string(beforeShared), "worked atomic shared file")
	equal(string(afterLocal), string(beforeLocal), "worked atomic local file")

	r = record(invoke(binary, repo, env, "config", "set", "model.work", "sonnet"))
	equal(r.code, 2, "worked alias exit")
	equal(r.stderr, "vloop: sonnet is a Claude alias, and agent.work is copilot: vloop models copilot lists the Copilot ids\n", "worked alias")
	r = record(invoke(binary, repo, env, "doctor"))
	contains(r.stdout, "✓ copilot models 1 id(s) checked · plan individual_pro\n", "worked doctor")
	if strings.Contains(r.stdout, "✗ copilot models") {
		fail("worked doctor: the copilot models row is a problem")
	}

	server.Close()
	r = record(invoke(binary, repo, env, "config", "set", "model.work", "gpt-6.1-terra"))
	equal(r.code, 0, "stopped stub config exit")
	contains(r.stderr, "! cannot check gpt-6.1-terra against Copilot's model list (", "stopped stub config warning")
	contains(r.stderr, "); set anyway\n", "stopped stub config suffix")
	r = record(invoke(binary, repo, env, "doctor"))
	contains(r.stdout, "! copilot models cannot reach Copilot's model list (", "stopped stub doctor warning")
	contains(r.stdout, "); ids not checked\n", "stopped stub doctor suffix")
	if strings.Contains(r.stdout, "✗ copilot models") {
		fail("stopped stub doctor: an unreachable API made the copilot models row a problem")
	}
	if strings.Contains(combined, "stub-token") {
		fail("token leaked into command output")
	}
	// Nothing the commands left behind (config, logs, records, fixtures)
	// may hold the token either.
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if b, err := os.ReadFile(p); err == nil && strings.Contains(string(b), "stub-token") {
			fail("token leaked into " + p)
		}
		return nil
	})
}

func writeCopilotShim(dir string) {
	name, content := "copilot", "#!/bin/sh\nexit 0\n"
	if runtime.GOOS == "windows" {
		name += ".cmd"
		content = "@echo off\r\nexit /b 0\r\n"
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
