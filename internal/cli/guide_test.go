package cli

import (
	"encoding/json"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/mvelosop/vloop/internal/classify"
	"github.com/mvelosop/vloop/internal/config"
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

func TestGuideDefectsCoversInterventions(t *testing.T) {
	text := readGuide(t, "defects.md")
	for _, w := range guideWords(t, "intervention/v1", true) {
		if !strings.Contains(text, w) {
			t.Errorf("defects.md does not mention the intervention field or value %q", w)
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

// TestGuideConfigurationNamesEveryKey checks the configuration guide names
// every config key, in backticks.
func TestGuideConfigurationNamesEveryKey(t *testing.T) {
	text := readGuide(t, "configuration.md")
	for _, k := range config.Keys {
		if !strings.Contains(text, "`"+k.Name+"`") {
			t.Errorf("configuration.md does not mention config key %s", k.Name)
		}
	}
}

// TestGuideCommandsNamesEveryCommandAndFlag checks the command reference names
// every runnable command and every flag.
func TestGuideCommandsNamesEveryCommandAndFlag(t *testing.T) {
	text := readGuide(t, "commands.md")
	root, _ := NewRoot(Build{})
	cmds, flags := map[string]bool{}, map[string]bool{}
	walk(root, "", cmds, flags)
	delete(flags, "help")
	for c := range cmds {
		if !strings.Contains(text, "vloop "+c) {
			t.Errorf("commands.md does not document vloop %s", c)
		}
	}
	for f := range flags {
		if !strings.Contains(text, "--"+f) {
			t.Errorf("commands.md does not document --%s", f)
		}
	}
}

// TestGuideShellDefaultMatchesCode checks the configuration guide's stated default for the
// `shell` key on each OS against config's (F1).
func TestGuideShellDefaultMatchesCode(t *testing.T) {
	b, err := os.ReadFile("../../docs/guide/configuration.md")
	if err != nil {
		t.Fatal(err)
	}
	var row string
	for _, l := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(l, "| `shell` |") {
			row = l
		}
	}
	if row == "" {
		t.Fatal("configuration guide has no shell row")
	}
	def := strings.TrimSpace(strings.Split(row, "|")[2])
	m := regexp.MustCompile("^`([a-z]+)` \\(`([a-z]+)` on Windows\\)$").FindStringSubmatch(def)
	if m == nil {
		t.Fatalf("the guide's shell default %q is not of the form `x` (`y` on Windows)", def)
	}
	// The code's default on the other OS is read from defaultShell's source,
	// since config.Keys holds only the running OS's.
	src, err := os.ReadFile("../../internal/config/config.go")
	if err != nil {
		t.Fatal(err)
	}
	fn := regexp.MustCompile(`(?s)func defaultShell\(\) string \{.*?return "([a-z]+)"\n\t\}\n\treturn "([a-z]+)"`).FindStringSubmatch(string(src))
	if fn == nil {
		t.Fatal("cannot read defaultShell in internal/config/config.go")
	}
	if m[2] != fn[1] || m[1] != fn[2] {
		t.Errorf("the guide says shell defaults to %s (%s on Windows); the code says %s (%s on Windows)", m[1], m[2], fn[2], fn[1])
	}
	for _, k := range config.Keys {
		if k.Name == "shell" {
			want := m[1]
			if runtime.GOOS == "windows" {
				want = m[2]
			}
			if k.Default != want {
				t.Errorf("the guide says the shell default here is %s; config says %s", want, k.Default)
			}
		}
	}
}
