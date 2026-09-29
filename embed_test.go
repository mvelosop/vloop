package vloop

import "testing"

// Fails without the all: prefix on the go:embed directive.
func TestPluginManifestIsEmbedded(t *testing.T) {
	b, err := Plugin.ReadFile("plugin/.claude-plugin/plugin.json")
	if err != nil {
		t.Fatalf("embedded FS lacks plugin/.claude-plugin/plugin.json: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("embedded plugin.json is empty")
	}
}
