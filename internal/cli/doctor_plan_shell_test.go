package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/state"
)

func TestDoctorPlanShellUsesOneReplacedPlanDecision(t *testing.T) {
	const (
		briefA      = "docs/briefs/a.loop-brief.md"
		briefB      = "docs/briefs/b.loop-brief.md"
		planShell   = "vloop-plan-shell"
		configShell = "vloop-config-shell"
		planMessage = "the plan's shell \"vloop-plan-shell\" is not on PATH"
		configMsg   = "the configured shell \"vloop-config-shell\" is not on PATH"
	)
	none := ""
	same, other := briefA, briefB

	bin := t.TempDir()
	name := configShell
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if err := os.WriteFile(filepath.Join(bin, name), []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}
	empty := t.TempDir()

	type testCase struct {
		name        string
		plan        doctorPlanInfo
		hasPlan     bool
		runBrief    *string
		mixed       bool
		configured  bool
		result      string
		message     string
		planResult  string
		planMessage string
	}
	v2 := func(status string) doctorPlanInfo {
		return doctorPlanInfo{Schema: state.SchemaName, Shell: planShell, Status: status, Brief: briefA}
	}
	v1 := func(status string) doctorPlanInfo {
		return doctorPlanInfo{Schema: state.SchemaV1, Status: status}
	}
	tests := []testCase{
		{"no plan, configured present", doctorPlanInfo{}, false, &other, false, true, resPass, "", resNA, ""},
		{"no plan, configured absent", doctorPlanInfo{}, false, &other, false, false, resWarning, configMsg, "", ""},
		{"no plan, configured absent, mixed", doctorPlanInfo{}, false, &other, true, false, resProblem, configMsg, "", ""},
		{"complete v2, bare doctor", v2("complete"), true, nil, false, true, resPass, "", resPass, "a complete plan of " + briefA + " — the next brief's plan replaces it"},
		{"complete v2, another brief", v2("complete"), true, &other, false, true, resPass, "", "", ""},
		{"complete v2, bare, configured absent", v2("complete"), true, nil, false, false, resWarning, configMsg, "", ""},
		{"complete v2, another brief, configured absent, mixed", v2("complete"), true, &other, true, false, resProblem, configMsg, "", ""},
		{"complete v2, own brief", v2("complete"), true, &same, false, true, resProblem, planMessage, "", ""},
		{"complete v2, unnamed run", v2("complete"), true, &none, false, true, resProblem, planMessage, "", ""},
		{"complete v1, bare doctor", v1("complete"), true, nil, false, true, resPass, "", resPass, "a complete state/v1 plan — the next brief's plan replaces it"},
		{"complete v1, own brief", v1("complete"), true, &same, false, true, resPass, "", resPass, "a complete state/v1 plan — the next brief's plan replaces it"},
		{"complete v1, configured absent", v1("complete"), true, &other, false, false, resWarning, configMsg, "", ""},
		{"complete v1, configured absent, mixed", v1("complete"), true, &other, true, false, resProblem, configMsg, "", ""},
	}
	for _, status := range []string{"planning", "running", "blocked", "stalled", "halted"} {
		tests = append(tests, testCase{
			name: "v2 " + status, plan: v2(status), hasPlan: true, runBrief: &other,
			configured: true, result: resProblem, message: planMessage,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := empty
			if tt.configured {
				path = bin
			}
			t.Setenv("PATH", path)
			vals := []config.Value{{Key: "shell", Value: configShell}}
			if tt.mixed {
				vals = append(vals, config.Value{Key: "platforms", List: []string{"macos", "windows"}})
			}
			var got doctorCheck
			var gotPlan doctorCheck
			doctorPlanShell(t.TempDir(), tt.plan, tt.hasPlan, tt.runBrief, vals, false,
				func(name, result, message string) {
					if name == "gate shell" {
						got = doctorCheck{Name: name, Result: result, Message: message}
					}
					if name == "plan" {
						gotPlan = doctorCheck{Name: name, Result: result, Message: message}
					}
				})
			if got.Result != tt.result || got.Message != tt.message {
				t.Errorf("gate shell = %s %q, want %s %q", got.Result, got.Message, tt.result, tt.message)
			}
			if tt.planResult != "" && (gotPlan.Result != tt.planResult || gotPlan.Message != tt.planMessage) {
				t.Errorf("plan = %s %q, want %s %q", gotPlan.Result, gotPlan.Message, tt.planResult, tt.planMessage)
			}
		})
	}
}

func TestDoctorPlanShellPreflightCarriesConfiguredWarning(t *testing.T) {
	const message = `the configured shell "pwsh" is not on PATH`
	var out bytes.Buffer
	bad := writeRunPreflightChecks(&out, []doctorCheck{{
		Name: "gate shell", Result: resWarning, Message: message,
	}}, false)
	if bad != 0 || !strings.Contains(out.String(), "  ! gate shell "+message+"\n") {
		t.Fatalf("preflight bad=%d output=%q", bad, out.String())
	}
}
