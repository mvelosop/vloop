package cli

import (
	"errors"

	"bytes"
	"encoding/json"
	"github.com/spf13/cobra"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
)

func run(t *testing.T, args ...string) (code int, out, errOut string) {
	t.Helper()
	var o, e bytes.Buffer
	code = Execute(Build{Version: "1.2.3", Commit: "abc"}, args, &o, &e)
	return code, o.String(), e.String()
}

func TestVersionJSON(t *testing.T) {
	code, out, _ := run(t, "version", "--json")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"version": "1.2.3", "commit": "abc", "plugin": "0.8.0",
		"go": runtime.Version(), "os": runtime.GOOS, "arch": runtime.GOARCH}
	for k, v := range want {
		if m[k] != v {
			t.Errorf("%s = %q, want %q", k, m[k], v)
		}
	}
	if len(m) != len(want) {
		t.Errorf("keys = %v", m)
	}
	order := []string{"version", "commit", "plugin", "go", "os", "arch"}
	last := -1
	for _, k := range order {
		i := strings.Index(out, `"`+k+`"`)
		if i <= last {
			t.Errorf("key %s out of order in %s", k, out)
		}
		last = i
	}
}

func TestVersionText(t *testing.T) {
	_, out, _ := run(t, "version")
	want := "vloop 1.2.3 (commit abc, plugin 0.8.0, " + runtime.Version() + " " + runtime.GOOS + "/" + runtime.GOARCH + ")\n"
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestUsageErrors(t *testing.T) {
	for _, args := range [][]string{{"nosuch"}, {"version", "--nosuchflag"}, {"--json", "nosuch"}, {"version", "extra"}} {
		code, out, e := run(t, args...)
		if code != ExitUsage || out != "" || !strings.HasPrefix(e, "vloop: ") || strings.Count(e, "\n") != 1 {
			t.Errorf("%v: code=%d out=%q err=%q", args, code, out, e)
		}
	}
}

func TestProblemErrorExitsOne(t *testing.T) {
	var e bytes.Buffer
	root, _ := NewRoot(Build{})
	root.AddCommand(newProblemCmd())
	root.SetArgs([]string{"boom"})
	root.SetErr(&e)
	err := root.Execute()
	if _, ok := err.(*ProblemError); !ok {
		t.Fatalf("got %T", err)
	}
}

func newProblemCmd() *cobra.Command {
	return &cobra.Command{Use: "boom", RunE: func(*cobra.Command, []string) error {
		return Problem(errors.New("found problems"))
	}}
}

func TestResolveBuild(t *testing.T) {
	defer func(f func() (*debug.BuildInfo, bool)) { readBuildInfo = f }(readBuildInfo)
	info := func(mod string, kv ...string) func() (*debug.BuildInfo, bool) {
		return func() (*debug.BuildInfo, bool) {
			bi := &debug.BuildInfo{Main: debug.Module{Version: mod}}
			for i := 0; i+1 < len(kv); i += 2 {
				bi.Settings = append(bi.Settings, debug.BuildSetting{Key: kv[i], Value: kv[i+1]})
			}
			return bi, true
		}
	}
	cases := []struct {
		name, stamp, mod string
		kv               []string
		commit, module   string
	}{
		{"stamped", "abc", "", []string{"vcs.revision", "r1"}, "abc", ""},
		{"revision", "unknown", "(devel)", []string{"vcs.revision", "r1", "vcs.modified", "false"}, "r1", ""},
		{"dirty", "unknown", "", []string{"vcs.revision", "r1", "vcs.modified", "true"}, "r1-dirty", ""},
		{"installed", "unknown", "v0.8.0", nil, "unknown", "v0.8.0"},
		{"nothing", "unknown", "(devel)", nil, "unknown", ""},
	}
	for _, c := range cases {
		readBuildInfo = info(c.mod, c.kv...)
		b := Resolve("1.0.0", c.stamp)
		if b.Commit != c.commit || b.Module != c.module {
			t.Errorf("%s: got commit %q module %q, want %q %q", c.name, b.Commit, b.Module, c.commit, c.module)
		}
	}
	for v, want := range map[string]bool{"v0.8.0": true, "v0.8.0-rc1": false, "v0.0.0-20260101-abcdef": false, "": false} {
		if isRelease(v) != want {
			t.Errorf("isRelease(%q) = %v", v, !want)
		}
	}
}
