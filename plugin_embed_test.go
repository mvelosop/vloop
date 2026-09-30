package vloop

import (
	"encoding/json"
	"testing"
)

func TestPluginHookIsEmbedded(t *testing.T) {
	b, err := Plugin.ReadFile("plugin/hooks/hooks.json")
	if err != nil {
		t.Fatalf("embedded FS lacks plugin/hooks/hooks.json: %v", err)
	}
	var doc struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Type    string `json:"type"`
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Hooks) != 1 || len(doc.Hooks["SessionStart"]) != 1 || len(doc.Hooks["SessionStart"][0].Hooks) != 1 {
		t.Fatalf("want exactly one SessionStart hook, got %s", b)
	}
	h := doc.Hooks["SessionStart"][0].Hooks[0]
	if h.Type != "command" || h.Command != `vloop version --check-plugin "${CLAUDE_PLUGIN_ROOT}"` {
		t.Fatalf("unexpected hook: %+v", h)
	}
}
