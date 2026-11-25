package config

import (
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
