package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeLocal(t *testing.T, root, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(LocalFilePath))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLocalPrecedence(t *testing.T) {
	root := scratch(t)
	writeCfg(t, root, "[agent]\nwork = \"copilot\"\nreview = \"copilot\"\n")
	writeLocal(t, root, "[agent]\nwork = \"claude\"\n")
	cases := []struct{ key, val, src string }{
		{"agent.work", "claude", SourceLocal},
		{"agent.review", "copilot", SourceFile},
		{"agent.plan", "claude", SourceDefault},
	}
	for _, c := range cases {
		v, err := Get(root, c.key)
		if err != nil || v.Value != c.val || v.Source != c.src {
			t.Errorf("%s = %+v, %v; want %s (%s)", c.key, v, err, c.val, c.src)
		}
	}
	t.Setenv(EnvVar("agent.work"), "copilot")
	if v, _ := Get(root, "agent.work"); v.Source != SourceEnv || v.Value != "copilot" {
		t.Errorf("env must win over both: %+v", v)
	}
	vals, err := List(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range vals {
		if v.Key == "agent.review" && v.Source != SourceFile {
			t.Errorf("List agent.review source %s", v.Source)
		}
	}
}

func TestLocalListWholeAndEmpty(t *testing.T) {
	root := scratch(t)
	writeCfg(t, root, "[metrics]\ncode = [\"a/**\"]\ntest = [\"t/**\"]\n")
	writeLocal(t, root, "[metrics]\ncode = [\"b/**\", \"c/**\"]\ntest = []\n")
	v, err := Get(root, "metrics.code")
	if err != nil || v.Source != SourceLocal || !reflect.DeepEqual(v.List, []string{"b/**", "c/**"}) {
		t.Errorf("metrics.code = %+v, %v; want the local list whole", v, err)
	}
	v, err = Get(root, "metrics.test")
	if err != nil || v.Source != SourceFile || !reflect.DeepEqual(v.List, []string{"t/**"}) {
		t.Errorf("metrics.test = %+v, %v; an empty local list sets nothing", v, err)
	}
}

func TestLocalEmptyOrUnknownChangesNothing(t *testing.T) {
	root := scratch(t)
	writeCfg(t, root, "model.work = \"x\"\n")
	before, err := List(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{"", "[unknown]\nthing = 1\n"} {
		writeLocal(t, root, body)
		after, err := List(root)
		if err != nil || !reflect.DeepEqual(before, after) {
			t.Errorf("local %q changed the result or failed: %v", body, err)
		}
	}
}

func TestLocalBadValueAndMalformed(t *testing.T) {
	root := scratch(t)
	writeLocal(t, root, "agent.work = \"nope\"\n")
	_, err := Get(root, "agent.work")
	want := ".vloop/config.local.toml: invalid value \"nope\" for agent.work: want one of claude, copilot"
	if err == nil || err.Error() != want {
		t.Errorf("bad value: %v; want %s", err, want)
	}
	if _, err := List(root); err == nil {
		t.Error("List accepted a bad local value")
	}
	writeLocal(t, root, "agent.work = \n")
	_, err = Get(root, "language")
	if err == nil || !strings.HasPrefix(err.Error(), ".vloop/config.local.toml: ") {
		t.Errorf("malformed: %v", err)
	}
}

func TestLocalCheckRefused(t *testing.T) {
	root := scratch(t)
	writeLocal(t, root, "[[check]]\nname = \"mine\"\npaths = [\"**\"]\nrun = \"true\"\n")
	want := ".vloop/config.local.toml: [[check]] belongs in .vloop/config.toml — the checks are the repository's, not one machine's"
	if defs, err := Checks(root); err == nil || err.Error() != want || defs != nil {
		t.Errorf("Checks: %v, %v", defs, err)
	}
	if _, err := Get(root, "language"); err == nil || err.Error() != want {
		t.Errorf("Get: %v", err)
	}
	if _, err := List(root); err == nil || err.Error() != want {
		t.Errorf("List: %v", err)
	}
}

func TestPlatformsAndMixedShellAreSharedOnly(t *testing.T) {
	root := scratch(t)
	platformErr := "platforms belongs in .vloop/config.toml — the platforms are the repository's, not one machine's"
	writeLocal(t, root, "platforms = [\"linux\"]\n")
	for _, get := range []func() error{
		func() error { _, err := Get(root, "language"); return err },
		func() error { _, err := List(root); return err },
	} {
		if err := get(); err == nil || err.Error() != LocalFilePath+": "+platformErr {
			t.Errorf("local platforms: %v", err)
		}
	}
	writeLocal(t, root, "")
	t.Setenv(EnvVar("platforms"), "linux")
	if _, err := Get(root, "language"); err == nil || err.Error() != EnvVar("platforms")+": "+platformErr {
		t.Errorf("environment platforms: %v", err)
	}
	t.Setenv(EnvVar("platforms"), "")

	writeCfg(t, root, "platforms = [\"windows\", \"macos\", \"macos\"]\n")
	shellErr := "shell belongs in .vloop/config.toml while platforms lists macos, windows — the checks and the gates are written for one shell"
	writeLocal(t, root, "shell = \"pwsh\"\n")
	if _, err := Get(root, "language"); err == nil || err.Error() != LocalFilePath+": "+shellErr {
		t.Errorf("local shell: %v", err)
	}
	writeLocal(t, root, "")
	t.Setenv(EnvVar("shell"), "pwsh")
	if _, err := Get(root, "language"); err == nil || err.Error() != EnvVar("shell")+": "+shellErr {
		t.Errorf("environment shell: %v", err)
	}
	t.Setenv(EnvVar("shell"), "")

	writeCfg(t, root, "platforms = [\"windows\"]\n")
	if err := SetLocal(root, "shell", "sh"); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvVar("shell"), "bash")
	if v, err := Get(root, "shell"); err != nil || v.Value != "bash" || v.Source != SourceEnv {
		t.Errorf("single platform shell overrides: %+v %v", v, err)
	}
}

func TestSetLocalRefusesSharedOnlyKeysWithoutWriting(t *testing.T) {
	root := scratch(t)
	for _, tc := range []struct {
		key, value, config, want string
	}{
		{"platforms", "linux", "", LocalFilePath + ": platforms belongs in " + FilePath},
		{"shell", "pwsh", "platforms = [\"macos\", \"windows\"]\n", LocalFilePath + ": shell belongs in " + FilePath + " while platforms lists macos, windows"},
	} {
		if tc.config != "" {
			writeCfg(t, root, tc.config)
		}
		before, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(LocalFilePath)))
		err := SetLocal(root, tc.key, tc.value)
		var usage *UsageError
		if err == nil || !errors.As(err, &usage) || !strings.HasPrefix(err.Error(), tc.want) {
			t.Errorf("SetLocal(%s): %v", tc.key, err)
		}
		after, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(LocalFilePath)))
		if string(after) != string(before) {
			t.Errorf("SetLocal(%s) wrote %q", tc.key, after)
		}
	}
}

func TestSetLocalShellHonoursPlatformsOwnership(t *testing.T) {
	root := scratch(t)
	t.Setenv(EnvVar("platforms"), "linux")
	err := SetLocal(root, "shell", "pwsh")
	want := EnvVar("platforms") + ": platforms belongs in " + FilePath
	var source *SourceError
	if err == nil || !errors.As(err, &source) || !strings.HasPrefix(err.Error(), want) {
		t.Errorf("environment platforms: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(LocalFilePath))); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("refusal wrote local file: %v", err)
	}
}
