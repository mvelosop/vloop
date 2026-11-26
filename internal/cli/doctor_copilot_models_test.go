package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestDoctorCopilotModelsHTTPMatrix(t *testing.T) {
	const token = "matrix-secret-token"
	cases := []struct {
		name         string
		models       string
		aliases      string
		loggedIn     bool
		modelsStatus int
		planStatus   int
		modelsCalls  int32
		planCalls    int32
		result       string
		message      string
	}{
		{
			name: "pass and distinct count", loggedIn: true,
			models:     `{"data":[{"id":"mai-code-2","model_picker_enabled":true,"policy":{"state":"enabled"}}]}`,
			planStatus: http.StatusOK, modelsStatus: http.StatusOK, modelsCalls: 1, planCalls: 1,
			result: resPass, message: "1 id(s) checked · plan individual_pro",
		},
		{
			name: "missing", loggedIn: true,
			models:     `{"data":[{"id":"other","model_picker_enabled":true,"policy":{"state":"enabled"}}]}`,
			planStatus: http.StatusOK, modelsStatus: http.StatusOK, modelsCalls: 1, planCalls: 1,
			result: resProblem, message: "model.work mai-code-2 is not listed; model.review mai-code-2 is not listed · plan individual_pro — vloop models copilot lists them",
		},
		{
			name: "disabled", loggedIn: true,
			models:     `{"data":[{"id":"mai-code-2","model_picker_enabled":true,"policy":{"state":"disabled"}}]}`,
			planStatus: http.StatusOK, modelsStatus: http.StatusOK, modelsCalls: 1, planCalls: 1,
			result: resProblem, message: "model.work mai-code-2 is disabled; model.review mai-code-2 is disabled · plan individual_pro — vloop models copilot lists them",
		},
		{
			name: "alias without network", loggedIn: true, aliases: "sonnet",
			result: resProblem, message: "model.work sonnet is a Claude alias; model.review sonnet is a Claude alias; ids not checked against the list",
		},
		{
			name: "unreachable models", loggedIn: true,
			modelsStatus: http.StatusServiceUnavailable, planStatus: http.StatusOK, modelsCalls: 1, planCalls: 1,
			result: resWarning, message: "cannot reach Copilot's model list (503 Service Unavailable); ids not checked",
		},
		{
			name: "unknown plan", loggedIn: true,
			models:     `{"data":[{"id":"mai-code-2","model_picker_enabled":true,"policy":{"state":"enabled"}}]}`,
			planStatus: http.StatusServiceUnavailable, modelsStatus: http.StatusOK, modelsCalls: 1, planCalls: 1,
			result: resPass, message: "1 id(s) checked · plan unknown",
		},
		{
			name: "no login", result: "", message: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			model := "mai-code-2"
			if tc.aliases != "" {
				model = tc.aliases
			}
			configPath := filepath.Join(root, ".vloop", "config.toml")
			if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
				t.Fatal(err)
			}
			config := "[agent]\nwork = \"copilot\"\nreview = \"copilot\"\n[model]\nwork = \"" + model + "\"\nreview = \"" + model + "\"\n"
			if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
				t.Fatal(err)
			}
			var modelsCalls, planCalls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("Authorization"); got != "Bearer "+token {
					t.Errorf("Authorization = %q", got)
				}
				switch r.URL.Path {
				case "/models":
					modelsCalls.Add(1)
					w.WriteHeader(tc.modelsStatus)
					fmt.Fprint(w, tc.models)
				case "/copilot_internal/user":
					planCalls.Add(1)
					w.WriteHeader(tc.planStatus)
					fmt.Fprint(w, `{"copilot_plan":"individual_pro"}`)
				default:
					http.NotFound(w, r)
				}
			}))
			t.Cleanup(server.Close)
			t.Setenv("GH_TOKEN", token)
			t.Setenv("VLOOP_COPILOT_API_URL", server.URL)
			t.Setenv("VLOOP_GITHUB_API_URL", server.URL)

			var got []doctorCheck
			doctorCopilotModels(root, []string{"agent.work", "agent.review"}, tc.loggedIn, func(name, result, message string) {
				got = append(got, doctorCheck{Name: name, Result: result, Message: message})
			})
			if len(got) == 0 {
				if tc.result != "" {
					t.Fatalf("no row, want %s %q", tc.result, tc.message)
				}
			} else if len(got) != 1 || got[0].Result != tc.result || got[0].Message != tc.message {
				t.Errorf("rows = %+v, want %s %q", got, tc.result, tc.message)
			}
			if modelsCalls.Load() != tc.modelsCalls || planCalls.Load() != tc.planCalls {
				t.Errorf("calls: models %d, plan %d; want models %d, plan %d", modelsCalls.Load(), planCalls.Load(), tc.modelsCalls, tc.planCalls)
			}
		})
	}
}

func TestDoctorCopilotModelsUsesResolvedConfigValues(t *testing.T) {
	const token = "resolved-secret-token"
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".vloop"), 0o755); err != nil {
		t.Fatal(err)
	}
	shared := "[agent]\nwork = \"copilot\"\nreview = \"copilot\"\n[model]\nwork = \"shared-work\"\nreview = \"shared-review\"\n"
	if err := os.WriteFile(filepath.Join(root, ".vloop", "config.toml"), []byte(shared), 0o644); err != nil {
		t.Fatal(err)
	}
	local := "[model]\nreview = \"local-review\"\n"
	if err := os.WriteFile(filepath.Join(root, ".vloop", "config.local.toml"), []byte(local), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("VLOOP_MODEL_WORK", "mai-code-2")
	var seen atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.Add(1)
		switch r.URL.Path {
		case "/models":
			fmt.Fprint(w, `{"data":[{"id":"mai-code-2","policy":{"state":"enabled"}},{"id":"local-review","policy":{"state":"enabled"}}]}`)
		case "/copilot_internal/user":
			fmt.Fprint(w, `{"copilot_plan":"individual_pro"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	t.Setenv("GH_TOKEN", token)
	t.Setenv("VLOOP_COPILOT_API_URL", server.URL)
	t.Setenv("VLOOP_GITHUB_API_URL", server.URL)

	var got []doctorCheck
	doctorCopilotModels(root, []string{"agent.work", "agent.review"}, true, func(name, result, message string) {
		got = append(got, doctorCheck{Name: name, Result: result, Message: message})
	})
	if len(got) != 1 || got[0].Result != resPass || got[0].Message != "2 id(s) checked · plan individual_pro" {
		t.Errorf("resolved model row = %+v", got)
	}
	if seen.Load() != 2 {
		t.Errorf("network calls = %d, want models and plan", seen.Load())
	}
}

func TestDoctorCopilotModelsPassDistinctOrderingAndJSON(t *testing.T) {
	e := newDoctorEnv(t)
	e.copilot(true)
	var modelsCalls, planCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/models":
			modelsCalls.Add(1)
			fmt.Fprint(w, `{"data":[{"id":"mai-code-2","model_picker_enabled":true,"policy":{"state":"enabled"}},{"id":"gpt-6.1-terra","model_picker_enabled":true}]}`)
		case "/copilot_internal/user":
			planCalls.Add(1)
			fmt.Fprint(w, `{"copilot_plan":"individual_pro"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	t.Setenv("VLOOP_COPILOT_API_URL", server.URL)
	t.Setenv("VLOOP_GITHUB_API_URL", server.URL)
	e.write(".vloop/config.toml", doctorCheckConfig+"[agent]\nwork = \"copilot\"\nreview = \"copilot\"\n[model]\nwork = \"mai-code-2\"\nreview = \"gpt-6.1-terra\"\n")

	checks := runDoctor(e.build, e.repo, nil)
	var copilot, models int
	for i, check := range checks {
		switch check.Name {
		case "copilot":
			copilot = i
		case "copilot models":
			models = i
			if check.Result != resPass || check.Message != "2 id(s) checked · plan individual_pro" {
				t.Errorf("copilot models = %s %q", check.Result, check.Message)
			}
		}
	}
	if models != copilot+1 {
		t.Errorf("copilot models row %d is not immediately after copilot row %d", models, copilot)
	}
	if modelsCalls.Load() != 1 || planCalls.Load() != 1 {
		t.Errorf("calls: models %d, plan %d", modelsCalls.Load(), planCalls.Load())
	}

	out, code := e.run("--json", "doctor")
	var doc struct {
		Checks []doctorCheck `json:"checks"`
	}
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, check := range doc.Checks {
		if check.Name == "copilot models" {
			found = true
			if check.Result != resPass || check.Message != "2 id(s) checked · plan individual_pro" {
				t.Errorf("JSON copilot models = %s %q, exit %d", check.Result, check.Message, code)
			}
		}
	}
	if !found {
		t.Fatal("JSON doctor has no copilot models row")
	}
}

func TestDoctorCopilotModelsProblemsAndPreflight(t *testing.T) {
	e := newDoctorEnv(t)
	e.copilot(true)
	var modelsCalls, planCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/models":
			modelsCalls.Add(1)
			fmt.Fprint(w, `{"data":[{"id":"claude-fable-5.1","model_picker_enabled":true,"policy":{"state":"disabled"}}]}`)
		case "/copilot_internal/user":
			planCalls.Add(1)
			fmt.Fprint(w, `{"copilot_plan":"individual_pro"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	t.Setenv("VLOOP_COPILOT_API_URL", server.URL)
	t.Setenv("VLOOP_GITHUB_API_URL", server.URL)
	e.write(".vloop/config.toml", doctorCheckConfig+"[agent]\nwork = \"copilot\"\nreview = \"copilot\"\n[model]\nwork = \"mai-code-3\"\nreview = \"claude-fable-5.1\"\n")

	out, code := e.run("doctor")
	want := "✗ copilot models model.work mai-code-3 is not listed; model.review claude-fable-5.1 is disabled · plan individual_pro — vloop models copilot lists them\n"
	if code != 1 || !strings.Contains(out, want) {
		t.Errorf("doctor exit %d, missing %q:\n%s", code, want, out)
	}
	errOut, err := e.preflight()
	if err == nil || !strings.Contains(errOut, "  ✗ copilot models "+strings.TrimPrefix(want, "✗ copilot models ")) {
		t.Errorf("preflight did not refuse the model row: %v\n%s", err, errOut)
	}
}

func TestDoctorCopilotModelsAliasNeedsNoNetwork(t *testing.T) {
	e := newDoctorEnv(t)
	e.copilot(true)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		http.Error(w, "unreachable", http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)
	t.Setenv("VLOOP_COPILOT_API_URL", server.URL)
	t.Setenv("VLOOP_GITHUB_API_URL", server.URL)
	e.write(".vloop/config.toml", doctorCheckConfig+"[agent]\nwork = \"copilot\"\n[model]\nwork = \"sonnet\"\n")

	check, code := e.check("copilot models")
	if code != 1 || check.Result != resProblem ||
		check.Message != "model.work sonnet is a Claude alias; ids not checked against the list" {
		t.Errorf("alias row = %s %q, exit %d", check.Result, check.Message, code)
	}
	if calls.Load() != 0 {
		t.Errorf("alias made %d network calls", calls.Load())
	}
}

func TestDoctorCopilotModelsUnavailableIsWarningAndPlanUnknown(t *testing.T) {
	for _, tc := range []struct {
		name        string
		modelsFail  bool
		planFail    bool
		wantResult  string
		wantMessage string
	}{
		{"models unavailable", true, false, resWarning, "cannot reach Copilot's model list (503 Service Unavailable); ids not checked"},
		{"plan unavailable", false, true, resPass, "1 id(s) checked · plan unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := newDoctorEnv(t)
			e.copilot(true)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/models":
					if tc.modelsFail {
						http.Error(w, "unavailable", http.StatusServiceUnavailable)
						return
					}
					fmt.Fprint(w, `{"data":[{"id":"mai-code-2","model_picker_enabled":true,"policy":{"state":"enabled"}}]}`)
				case "/copilot_internal/user":
					if tc.planFail {
						http.Error(w, "unavailable", http.StatusServiceUnavailable)
						return
					}
					fmt.Fprint(w, `{"copilot_plan":"individual_pro"}`)
				default:
					http.NotFound(w, r)
				}
			}))
			t.Cleanup(server.Close)
			t.Setenv("VLOOP_COPILOT_API_URL", server.URL)
			t.Setenv("VLOOP_GITHUB_API_URL", server.URL)
			e.write(".vloop/config.toml", doctorCheckConfig+"[agent]\nwork = \"copilot\"\n[model]\nwork = \"mai-code-2\"\n")

			check, code := e.check("copilot models")
			if code != 0 || check.Result != tc.wantResult || check.Message != tc.wantMessage {
				t.Errorf("row = %s %q, exit %d", check.Result, check.Message, code)
			}
			if _, err := e.preflight(); err != nil {
				t.Errorf("warning refused preflight: %v", err)
			}
		})
	}
}

func TestDoctorCopilotModelsSkipsWithoutLoginOrCopilotKind(t *testing.T) {
	for _, copilotKind := range []bool{true, false} {
		t.Run(fmt.Sprintf("copilot-kind-%v", copilotKind), func(t *testing.T) {
			e := newDoctorEnv(t)
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				calls.Add(1)
			}))
			t.Cleanup(server.Close)
			t.Setenv("VLOOP_COPILOT_API_URL", server.URL)
			t.Setenv("VLOOP_GITHUB_API_URL", server.URL)
			if copilotKind {
				e.copilot(true)
				e.gh("exit 1")
				e.write(".vloop/config.toml", doctorCheckConfig+"[agent]\nwork = \"copilot\"\n[model]\nwork = \"mai-code-2\"\n")
			} else {
				e.copilot(true)
			}

			out, _ := e.run("--json", "doctor")
			if strings.Contains(out, `"name":"copilot models"`) {
				t.Errorf("unexpected copilot models row:\n%s", out)
			}
			if calls.Load() != 0 {
				t.Errorf("unexpected %d API calls", calls.Load())
			}
		})
	}
}

func TestDoctorCopilotModelsRedactsToken(t *testing.T) {
	e := newDoctorEnv(t)
	e.copilot(true)
	const token = "doctor-secret-token"
	t.Setenv("GH_TOKEN", token)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		http.Error(w, token, http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)
	t.Setenv("VLOOP_COPILOT_API_URL", server.URL)
	t.Setenv("VLOOP_GITHUB_API_URL", server.URL)
	e.write(".vloop/config.toml", doctorCheckConfig+"[agent]\nwork = \"copilot\"\n[model]\nwork = \"mai-code-2\"\n")

	out, _ := e.run("doctor")
	jsonOut, _ := e.run("--json", "doctor")
	for _, output := range []string{out, jsonOut} {
		if strings.Contains(output, token) {
			t.Errorf("token leaked in output:\n%s", output)
		}
	}
}
