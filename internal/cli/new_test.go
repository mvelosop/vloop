package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var newPathRE = regexp.MustCompile(`^docs/briefs/B[0-9]{8}-[0-9]{4}-pagos\.loop-brief\.md\n$`)

func fileCount(t *testing.T, d string) int {
	t.Helper()
	n := 0
	_ = filepath.Walk(d, func(p string, i os.FileInfo, _ error) error {
		if i != nil && !i.IsDir() {
			n++
		}
		return nil
	})
	return n
}

func TestBriefNewWritesDraftThatChecksSkipped(t *testing.T) {
	for _, lang := range []string{"en", "es"} {
		d := scratchRepo(t)
		if lang == "es" {
			if code, _, e := run(t, "-C", d, "config", "set", "language", "es"); code != 0 {
				t.Fatal(e)
			}
		}
		before := fileCount(t, d)
		code, out, e := run(t, "-C", d, "brief", "new", "pagos")
		if code != 0 || e != "" || !newPathRE.MatchString(out) {
			t.Fatalf("%s: %d %q %q", lang, code, out, e)
		}
		rel := strings.TrimSpace(out)
		if fileCount(t, d) != before+1 {
			t.Fatalf("%s: wrote more than the brief", lang)
		}
		code, out, e = run(t, "-C", d, "brief", "check", rel)
		if code != 0 || !strings.Contains(out, "  - skipped: status is draft\n") || !strings.HasSuffix(out, "briefs ok\n") {
			t.Fatalf("%s: %d %q %q", lang, code, out, e)
		}
	}
}

func TestBriefNewExistsAndDryRun(t *testing.T) {
	d := scratchRepo(t)
	code, out, _ := run(t, "-C", d, "brief", "new", "pagos", "--dry-run")
	if code != 0 || !newPathRE.MatchString(out) || fileCount(t, d) != 0 {
		t.Fatalf("dry-run: %d %q", code, out)
	}
	// Create, then repeat: a same-minute rerun collides (retry once on a minute rollover).
	for i := 0; i < 2; i++ {
		_, first, _ := run(t, "-C", d, "brief", "new", "pagos")
		rel := strings.TrimSpace(first)
		snap, _ := os.ReadFile(filepath.Join(d, rel))
		code, out, e := run(t, "-C", d, "--json", "brief", "new", "pagos")
		if code == 0 {
			continue
		}
		var doc map[string]any
		if code != 1 || e != "vloop: brief exists: "+rel+"\n" || json.Unmarshal([]byte(out), &doc) != nil {
			t.Fatalf("%d %q %q", code, out, e)
		}
		if got, _ := os.ReadFile(filepath.Join(d, rel)); string(got) != string(snap) {
			t.Fatal("existing brief was modified")
		}
		return
	}
	t.Fatal("never saw an exists collision")
}

func TestBriefNewSlugRule(t *testing.T) {
	d := scratchRepo(t)
	for _, s := range []string{"Pagos", "pagos--base", "pagos-", "pagos_base", "pagos base", "ñandu", "-pagos"} {
		args := []string{"-C", d, "brief", "new", "--", s}
		code, out, e := run(t, args...)
		if code != 2 || out != "" || !strings.HasPrefix(e, "vloop: ") {
			t.Fatalf("%q: %d %q %q", s, code, out, e)
		}
	}
	if code, _, _ := run(t, "-C", d, "brief", "new"); code != 2 {
		t.Fatalf("no slug: %d", code)
	}
	if fileCount(t, d) != 0 {
		t.Fatal("a rejected slug wrote something")
	}
}
