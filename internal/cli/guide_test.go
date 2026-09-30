package cli

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/mvelosop/vloop/internal/classify"
	"github.com/mvelosop/vloop/internal/schema"
)

// schemaWords collects every property name at any depth and, when withEnums
// is set, every string enum value.
func schemaWords(v any, withEnums bool, out map[string]bool) {
	switch x := v.(type) {
	case map[string]any:
		if props, ok := x["properties"].(map[string]any); ok {
			for k := range props {
				out[k] = true
			}
		}
		if withEnums {
			if enum, ok := x["enum"].([]any); ok {
				for _, e := range enum {
					if s, ok := e.(string); ok {
						out[s] = true
					}
				}
			}
		}
		for _, c := range x {
			schemaWords(c, withEnums, out)
		}
	case []any:
		for _, c := range x {
			schemaWords(c, withEnums, out)
		}
	}
}

func guideWords(t *testing.T, name string, withEnums bool) []string {
	t.Helper()
	doc, err := schema.Document(name)
	if err != nil {
		t.Fatal(err)
	}
	var v any
	if err := json.Unmarshal(doc, &v); err != nil {
		t.Fatal(err)
	}
	set := map[string]bool{}
	schemaWords(v, withEnums, set)
	words := make([]string, 0, len(set))
	for w := range set {
		words = append(words, w)
	}
	sort.Strings(words)
	return words
}

func readGuide(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("../../docs/guide/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestGuideMetricsCoversSchemaAndPresets(t *testing.T) {
	text := readGuide(t, "metrics.md")
	words := guideWords(t, "metrics/v1", false)
	if len(words) < 20 {
		t.Fatalf("metrics/v1 declares only %d property names", len(words))
	}
	for _, w := range words {
		if !strings.Contains(text, w) {
			t.Errorf("metrics.md does not mention the metrics/v1 key %q", w)
		}
	}
	for _, name := range classify.Names() {
		if !strings.Contains(text, name) {
			t.Errorf("metrics.md does not mention the preset %q", name)
		}
		p, _ := classify.Lookup(name)
		for _, list := range [][]string{p.Code, p.Test, p.Docs, p.Excluded} {
			for _, g := range list {
				if !strings.Contains(text, g) {
					t.Errorf("metrics.md does not show the glob %q of preset %s", g, name)
				}
			}
		}
	}
}

func TestGuideDefectsCoversSchema(t *testing.T) {
	text := readGuide(t, "defects.md")
	for _, w := range guideWords(t, "defect/v1", true) {
		if !strings.Contains(text, w) {
			t.Errorf("defects.md does not mention the defect field or value %q", w)
		}
	}
}

func TestGuideConfigurationCoversStampAndScopes(t *testing.T) {
	text := readGuide(t, "configuration.md")
	for _, w := range guideWords(t, "install/v1", false) {
		if !strings.Contains(text, w) {
			t.Errorf("configuration.md does not mention the install/v1 key %q", w)
		}
	}
	for _, s := range []string{".vloop/install.json", "install/v1", "csharp@services/api", "longest", ".vloop/tmp/"} {
		if !strings.Contains(text, s) {
			t.Errorf("configuration.md does not mention %q", s)
		}
	}
}
