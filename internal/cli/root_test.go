package cli

import (
	"errors"

	"bytes"
	"encoding/json"
	"github.com/spf13/cobra"
	"runtime"
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
	want := map[string]string{"version": "1.2.3", "commit": "abc", "plugin": "0.7.0",
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
	want := "vloop 1.2.3 (commit abc, plugin 0.7.0, " + runtime.Version() + " " + runtime.GOOS + "/" + runtime.GOARCH + ")\n"
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
