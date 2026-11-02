package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The plugin's brief templates are the binary's templates with an
// executive-summary section added. checkPluginTemplates holds each to the
// binary's; it is a function of the plugin directory and the binary's
// template directory, so fixtures can run it on copies.

const (
	pluginTemplatesRel = "plugin/skills/create-brief/templates"
	binTemplatesRel    = "internal/brief/templates"
)

var pluginTemplateLangs = []struct{ plugin, bin, heading string }{
	{"brief-en.md", "en.md", "## Executive summary"},
	{"brief-es.md", "es.md", "## Resumen ejecutivo"},
}

// withoutSummary removes the lines from heading up to, not including, the next
// line starting "## ". It reports whether the heading was found.
func withoutSummary(b []byte, heading string) ([]byte, bool) {
	var out []byte
	found, skipping := false, false
	for _, line := range bytes.SplitAfter(b, []byte("\n")) {
		s := strings.TrimRight(string(line), "\r\n")
		switch {
		case !found && s == heading:
			found, skipping = true, true
			continue
		case skipping && strings.HasPrefix(s, "## "):
			skipping = false
		}
		if !skipping {
			out = append(out, line...)
		}
	}
	return out, found
}

func checkPluginTemplates(pluginDir, binDir string) []string {
	var vs []string
	for _, l := range pluginTemplateLangs {
		pRel := pluginTemplatesRel + "/" + l.plugin
		bRel := binTemplatesRel + "/" + l.bin
		p, err := os.ReadFile(filepath.Join(pluginDir, "skills", "create-brief", "templates", l.plugin))
		if err != nil {
			vs = append(vs, pRel+": "+err.Error())
			continue
		}
		b, err := os.ReadFile(filepath.Join(binDir, l.bin))
		if err != nil {
			vs = append(vs, bRel+": "+err.Error())
			continue
		}
		rest, found := withoutSummary(p, l.heading)
		if !found {
			vs = append(vs, pRel+" has no executive summary")
			continue
		}
		if !bytes.Equal(rest, b) {
			vs = append(vs, pRel+" differs from "+bRel+" outside its executive summary")
		}
	}
	return vs
}

func TestPluginBriefTemplatesMatchBinary(t *testing.T) {
	for _, v := range checkPluginTemplates("../../plugin", "../brief/templates") {
		t.Errorf("%s", v)
	}
}

// templateCopy copies the repository's four templates into a temporary tree and
// returns its plugin and binary template directories.
func templateCopy(t *testing.T) (pluginDir, binDir string) {
	t.Helper()
	root := t.TempDir()
	pluginDir = filepath.Join(root, "plugin")
	binDir = filepath.Join(root, "bin")
	for _, d := range []string{filepath.Join(pluginDir, "skills", "create-brief", "templates"), binDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, l := range pluginTemplateLangs {
		for src, dst := range map[string]string{
			filepath.Join("../../plugin/skills/create-brief/templates", l.plugin): filepath.Join(pluginDir, "skills", "create-brief", "templates", l.plugin),
			filepath.Join("../brief/templates", l.bin):                            filepath.Join(binDir, l.bin),
		} {
			b, err := os.ReadFile(src)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(dst, b, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return pluginDir, binDir
}

func editFile(t *testing.T, path string, f func(string) string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	n := f(string(b))
	if n == string(b) {
		t.Fatalf("edit of %s changed nothing", path)
	}
	if err := os.WriteFile(path, []byte(n), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPluginBriefTemplateCheckFixtures(t *testing.T) {
	enP := func(d string) string { return filepath.Join(d, "skills", "create-brief", "templates", "brief-en.md") }
	esP := func(d string) string { return filepath.Join(d, "skills", "create-brief", "templates", "brief-es.md") }
	cases := []struct {
		name string
		plan func(t *testing.T, pluginDir, binDir string)
		want string // empty: accepted
	}{
		{"copy", func(*testing.T, string, string) {}, ""},
		{"heading changed", func(t *testing.T, p, b string) {
			editFile(t, enP(p), func(s string) string {
				return strings.Replace(s, "\n## Out of scope\n", "\n## Out of scope, planted\n", 1)
			})
		}, "plugin/skills/create-brief/templates/brief-en.md differs from internal/brief/templates/en.md outside its executive summary"},
		{"binary es changed", func(t *testing.T, p, b string) {
			editFile(t, filepath.Join(b, "es.md"), func(s string) string {
				return strings.Replace(s, "\n## Fuera de alcance\n", "\n## Fuera de alcance, plantado\n", 1)
			})
		}, "plugin/skills/create-brief/templates/brief-es.md differs from internal/brief/templates/es.md outside its executive summary"},
		{"no es summary", func(t *testing.T, p, b string) {
			editFile(t, esP(p), func(s string) string { return strings.Replace(s, "\n## Resumen ejecutivo\n", "\n", 1) })
		}, "plugin/skills/create-brief/templates/brief-es.md has no executive summary"},
		{"no en summary", func(t *testing.T, p, b string) {
			editFile(t, enP(p), func(s string) string { return strings.Replace(s, "\n## Executive summary\n", "\n", 1) })
		}, "plugin/skills/create-brief/templates/brief-en.md has no executive summary"},
		{"line inside summary", func(t *testing.T, p, b string) {
			editFile(t, enP(p), func(s string) string {
				return strings.Replace(s, "\n## Executive summary\n", "\n## Executive summary\n\nA line planted inside.\n", 1)
			})
		}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, b := templateCopy(t)
			c.plan(t, p, b)
			got := strings.Join(checkPluginTemplates(p, b), "\n")
			if c.want == "" && got != "" {
				t.Fatalf("want accepted, got %s", got)
			}
			if c.want != "" && !strings.Contains(got, c.want) {
				t.Fatalf("want %q, got %q", c.want, got)
			}
		})
	}
}
