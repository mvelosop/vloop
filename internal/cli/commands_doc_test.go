package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/mvelosop/vloop/internal/classify"
	"github.com/mvelosop/vloop/internal/config"
)

// TestCommandsReferenceIsCurrent fails, without rewriting anything, when the
// committed page differs from a fresh generation.
func TestCommandsReferenceIsCurrent(t *testing.T) {
	got, err := os.ReadFile("../../docs/guide/commands.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != CommandsReference() {
		t.Errorf("docs/guide/commands.md is out of date: run `go generate ./...` and commit the result")
	}
}

func TestCommandsReferenceListsEveryCommand(t *testing.T) {
	root, _ := NewRoot(Build{})
	cmds, flags := map[string]bool{}, map[string]bool{}
	walk(root, "", cmds, flags)
	page := CommandsReference()
	for c := range cmds {
		if !strings.Contains(page, "## vloop "+c+"\n") {
			t.Errorf("commands reference lacks vloop %s", c)
		}
	}
	for f := range flags {
		if f != "help" && !strings.Contains(page, "--"+f) {
			t.Errorf("commands reference lacks --%s", f)
		}
	}
}

func TestGuideConfigurationCoversKeysAndPresets(t *testing.T) {
	text := readGuide(t, "configuration.md")
	for _, k := range config.Keys {
		if !strings.Contains(text, "`"+k.Name+"`") {
			t.Errorf("configuration.md does not document the key %q", k.Name)
		}
		v := "VLOOP_" + strings.ToUpper(strings.ReplaceAll(k.Name, ".", "_"))
		if !strings.Contains(text, v) {
			t.Errorf("configuration.md does not name %s", v)
		}
	}
	for _, name := range classify.Names() {
		if !strings.Contains(text, "`"+name+"`") {
			t.Errorf("configuration.md does not list the preset %q", name)
		}
		p, _ := classify.Lookup(name)
		for _, list := range [][]string{p.Code, p.Test, p.Excluded} {
			for _, g := range list {
				if !strings.Contains(text, g) {
					t.Errorf("configuration.md does not show the glob %q of preset %s", g, name)
				}
			}
		}
	}
}
