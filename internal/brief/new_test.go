package brief

import (
	"strings"
	"testing"
	"time"
)

func TestValidSlug(t *testing.T) {
	for s, want := range map[string]bool{
		"a": true, "pagos-base": true, "v2": true, "a-b-c": true,
		"": false, "Pagos": false, "a--b": false, "-a": false, "a-": false, "a_b": false, "a b": false, "ñandu": false,
	} {
		if ValidSlug(s) != want {
			t.Errorf("ValidSlug(%q) != %v", s, want)
		}
	}
}

func TestNamePathLocalTime(t *testing.T) {
	ts := time.Date(2026, 9, 29, 7, 5, 0, 0, time.FixedZone("x", 14*3600))
	if got := NewPath("pagos", ts); got != "docs/briefs/B20260929-0705-pagos.loop-brief.md" {
		t.Fatal(got)
	}
}

func TestRenderTemplatesFrontmatterAndHeadings(t *testing.T) {
	ts := time.Date(2026, 9, 29, 7, 5, 0, 0, time.UTC)
	for lang, set := range sets {
		text, err := Render(lang, "pagos", ts)
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{
			"name: B20260929-0705-pagos.loop-brief\n", "kind: brief\n", "status: draft\n",
			"created: 2026-09-29\n", "depends-on: []\n",
		} {
			if !strings.Contains(text, want) {
				t.Errorf("%s: missing %q", lang, want)
			}
		}
		if strings.Contains(text, "**Status:**") || strings.Contains(text, "{{") {
			t.Errorf("%s: status line or unfilled placeholder", lang)
		}
		for _, h := range set.Sections {
			if !strings.Contains(text, "\n## "+h+"\n") {
				t.Errorf("%s: missing heading %q", lang, h)
			}
		}
		b := Parse("docs/briefs/B20260929-0705-pagos.loop-brief.md", text)
		if res := Check(t.TempDir(), b, set); !res.Skipped {
			t.Errorf("%s: fresh brief must check as skipped: %v", lang, res.Lines)
		}
	}
}

func TestSpanishTemplateIsSpanish(t *testing.T) {
	es, _ := Render("es", "x", time.Now())
	en, _ := Render("en", "x", time.Now())
	if es == en || strings.Contains(es, "## What it is") || !strings.Contains(es, "Sustituye todo") {
		t.Fatal("es template is not Spanish")
	}
}
