package state

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func scratch(t *testing.T, fixture string) string {
	t.Helper()
	root := t.TempDir()
	data, err := os.ReadFile(filepath.Join("testdata", fixture))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(Path(root)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(root), data, 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestRoundTripByteForByte(t *testing.T) {
	root := scratch(t, "plan.json")
	want, _ := os.ReadFile(Path(root))
	p, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("marshal differs:\n%s", got)
	}
	if !strings.Contains(string(got), "<a> & b") {
		t.Fatal("<, > and & must be written literally")
	}
	if err := Save(root, p); err != nil {
		t.Fatal(err)
	}
	saved, _ := os.ReadFile(Path(root))
	if !bytes.Equal(bytes.Replace(saved, []byte(p.Updated), []byte("2026-01-01T09:00:00Z"), 1), want) {
		t.Fatalf("saved plan differs beyond updated:\n%s", saved)
	}
}

func TestSaveSetsUpdatedAndLeavesNoTempFile(t *testing.T) {
	root := scratch(t, "plan.json")
	p, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(root, p); err != nil {
		t.Fatal(err)
	}
	ts, err := time.Parse(time.RFC3339, p.Updated)
	if err != nil || !strings.HasSuffix(p.Updated, "Z") || time.Since(ts) > time.Minute {
		t.Fatalf("updated = %q, %v", p.Updated, err)
	}
	ents, err := os.ReadDir(filepath.Dir(Path(root)))
	if err != nil {
		t.Fatal(err)
	}
	if len(ents) != 1 || ents[0].Name() != "state.json" {
		t.Fatalf("directory holds %v, want only state.json", ents)
	}
	if vs, err := Validate(root); err != nil || len(vs) != 0 {
		t.Fatalf("saved plan invalid: %v %v", vs, err)
	}
}

func TestLoadMissing(t *testing.T) {
	if _, err := Load(t.TempDir()); !errors.Is(err, ErrNoPlan) {
		t.Fatalf("err = %v", err)
	}
	if _, err := Validate(t.TempDir()); !errors.Is(err, ErrNoPlan) {
		t.Fatalf("err = %v", err)
	}
}

func TestLoadSchemaInvalidStillReads(t *testing.T) {
	root := scratch(t, "plan.json")
	data, _ := os.ReadFile(Path(root))
	data = bytes.Replace(data, []byte(`"work": "opus"`), []byte(`"work": ""`), 1)
	data = bytes.Replace(data, []byte(`"review": "high"`), []byte(`"review": "turbo"`), 1)
	if err := os.WriteFile(Path(root), data, 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := Load(root)
	if err != nil || len(p.Tasks) != 2 {
		t.Fatalf("load: %v", err)
	}
	if vs, err := Validate(root); err != nil || len(vs) == 0 {
		t.Fatalf("want violations, got %v %v", vs, err)
	}
}

func TestMarkdown(t *testing.T) {
	p, err := Load(scratch(t, "plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	md := p.Markdown()
	for _, want := range []string{
		"# Plan — B20260101-0900-a\n",
		"- [x] **T1** — Skeleton <a> & b\n",
		"- [ ] **T2** — Config · 1 attempt(s)\n",
		"`pending` · 1 attempt(s) · depends on: T1\n",
		"`done` · depends on: none\n",
		"**From the last attempt:** watch the path\n",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown lacks %q", want)
		}
	}
}
