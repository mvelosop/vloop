package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModelsCopilotTextJSONAndFilters(t *testing.T) {
	var integrationHeaders int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer copilot-token" {
			t.Errorf("Authorization = %q", got)
		}
		switch r.URL.Path {
		case "/models":
			integrationHeaders++
			if got := r.Header.Get("Copilot-Integration-Id"); got != "copilot-developer-cli" {
				t.Errorf("Copilot-Integration-Id = %q", got)
			}
			fmt.Fprint(w, `{"data":[
				{"id":"mai-code-2","model_picker_enabled":true,"policy":{"state":"enabled"}},
				{"id":"claude-fable-5.1","model_picker_enabled":true,"policy":{"state":"disabled"}},
				{"id":"text-embedding-9","model_picker_enabled":false},
				{"id":"gpt-6.1-terra","model_picker_enabled":true}]}`)
		case "/copilot_internal/user":
			fmt.Fprint(w, `{"copilot_plan":"individual_pro"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	setModelAPIEnv(t, server.URL)

	code, out, errOut := runModels(t, "models", "copilot")
	want := "plan  individual_pro\nclaude-fable-5.1  disabled\ngpt-6.1-terra     -\nmai-code-2        enabled\n"
	if code != 0 || out != want || errOut != "" {
		t.Fatalf("text: code=%d stdout=%q stderr=%q", code, out, errOut)
	}

	code, out, errOut = runModels(t, "models", "copilot", "--all")
	if code != 0 || out != want+"text-embedding-9  -\n" || errOut != "" {
		t.Fatalf("--all: code=%d stdout=%q stderr=%q", code, out, errOut)
	}

	code, out, errOut = runModels(t, "--json", "models", "copilot")
	if code != 0 || errOut != "" {
		t.Fatalf("json: code=%d stderr=%q", code, errOut)
	}
	var doc copilotModelsDocument
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Schema != "copilot-models/v1" || doc.Plan == nil || *doc.Plan != "individual_pro" || len(doc.Models) != 3 {
		t.Fatalf("document = %+v", doc)
	}
	if doc.Models[0].ID != "claude-fable-5.1" || doc.Models[1].Policy != nil {
		t.Fatalf("models = %+v", doc.Models)
	}
	if integrationHeaders != 3 {
		t.Fatalf("models requests = %d, want 3", integrationHeaders)
	}
}

func TestModelsCopilotUnknownPlanAndFailures(t *testing.T) {
	failModels, planBody := false, `{}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/models":
			if failModels {
				http.Error(w, "unavailable", http.StatusServiceUnavailable)
				return
			}
			fmt.Fprint(w, `{"data":[]}`)
		case "/copilot_internal/user":
			fmt.Fprint(w, planBody)
		}
	}))
	defer server.Close()
	setModelAPIEnv(t, server.URL)

	code, out, errOut := runModels(t, "--json", "models", "copilot")
	var doc copilotModelsDocument
	if code != 0 || errOut != "" || json.Unmarshal([]byte(out), &doc) != nil || doc.Plan != nil {
		t.Fatalf("unknown plan: code=%d stdout=%q stderr=%q", code, out, errOut)
	}

	planBody = `{`
	code, out, errOut = runModels(t, "models", "copilot")
	if code != 0 || out != "plan  unknown\n" || errOut != "" {
		t.Fatalf("malformed plan: code=%d stdout=%q stderr=%q", code, out, errOut)
	}

	failModels = true
	code, out, errOut = runModels(t, "models", "copilot")
	if code != 1 || out != "" || errOut != "vloop: cannot list Copilot models: 503 Service Unavailable\n" {
		t.Fatalf("failed models: code=%d stdout=%q stderr=%q", code, out, errOut)
	}
}

func TestModelsCopilotUsageDoesNotNetwork(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests++
	}))
	defer server.Close()
	setModelAPIEnv(t, server.URL)

	for _, args := range [][]string{{"models"}, {"models", "claude"}, {"models", "copilot", "extra"}} {
		code, out, errOut := runModels(t, args...)
		if code != 2 || out != "" || errOut != "vloop: models lists Copilot models only: vloop models copilot\n" {
			t.Fatalf("%v: code=%d stdout=%q stderr=%q", args, code, out, errOut)
		}
	}
	if requests != 0 {
		t.Fatalf("invalid argument made %d requests", requests)
	}
}

func TestModelsCopilotNoLogin(t *testing.T) {
	t.Setenv("COPILOT_GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("PATH", t.TempDir())
	code, out, errOut := runModels(t, "models", "copilot")
	want := "vloop: no GitHub login reaches Copilot: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login\n"
	if code != 1 || out != "" || errOut != want {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, out, errOut)
	}
}

func TestModelsSchema(t *testing.T) {
	raw := `{"schema":"copilot-models/v1","plan":null,"models":[{"id":"mai-code-2","policy":null,"picker":true}]}`
	path := filepath.Join(t.TempDir(), "models.json")
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	root, _ := NewRoot(Build{})
	root.SetArgs([]string{"schema", "validate", "copilot-models/v1", path})
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	if err := root.Execute(); err != nil || stderr.Len() != 0 {
		t.Fatalf("validate: %v, stderr %q", err, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != path+": ok" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func setModelAPIEnv(t *testing.T, url string) {
	t.Helper()
	t.Setenv("COPILOT_GITHUB_TOKEN", "copilot-token")
	t.Setenv("GH_TOKEN", "gh-token")
	t.Setenv("GITHUB_TOKEN", "github-token")
	t.Setenv("VLOOP_COPILOT_API_URL", url)
	t.Setenv("VLOOP_GITHUB_API_URL", url)
}

func runModels(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	dir := t.TempDir()
	args = append([]string{"-C", dir}, args...)
	var stdout, stderr bytes.Buffer
	code := Execute(Build{}, args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}
