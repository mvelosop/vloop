package install

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCompare(t *testing.T) {
	cases := []struct {
		from, to string
		want     Relation
	}{
		{"0.1.0", "0.1.0", Same},
		{"0.2.0", "0.1.0", FromNewer},
		{"1.0.0", "0.9.9", FromNewer},
		{"0.1.9", "0.1.10", Newer}, // numeric, not textual
		{"0.1.10", "0.1.9", FromNewer},
		{"0.1.0", "0.2.0", Breaking}, // 0.x minor
		{"0.1.0", "1.0.0", Breaking}, // majors differ
		{"1.2.0", "2.0.0", Breaking},
		{"0.1.0", "0.1.1", Newer}, // 0.x patch
		{"1.0.0", "1.1.0", Newer}, // same-major minor
		{"1.0.0", "1.0.1", Newer},
		{"0.0.0-dev", "0.1.0", NeedsYes},
		{"0.1.0", "0.0.0-dev", NeedsYes},
		{"0.0.0-dev", "0.0.0-dev", Same},
		{"0.1.0-rc1", "0.1.0", NeedsYes},
	}
	for _, c := range cases {
		got, err := Compare(c.from, c.to)
		if err != nil || got != c.want {
			t.Errorf("Compare(%q, %q) = %v, %v; want %v", c.from, c.to, got, err, c.want)
		}
	}
}

func TestCompareRejectsGarbage(t *testing.T) {
	for _, v := range []string{"", "1.2", "a.b.c", "1.2.3.4", "01.2.3"} {
		if _, err := Compare(v, "0.1.0"); err == nil {
			t.Errorf("Compare(%q, 0.1.0): want an error", v)
		}
	}
}

func TestParsePreRelease(t *testing.T) {
	if _, err := Parse("0.0.0-dev"); !errors.Is(err, ErrPreRelease) {
		t.Fatalf("got %v", err)
	}
}

func TestReadWrite(t *testing.T) {
	root := t.TempDir()
	if _, err := Read(root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing stamp: %v", err)
	}
	up := "2026-10-02T09:00:00Z"
	for _, want := range []Stamp{
		{Version: "0.1.0", Commit: "abc1234", Initialized: "2026-10-01T09:00:00Z"},
		{Version: "0.1.1", Commit: "def5678", Initialized: "2026-10-01T09:00:00Z", Upgraded: &up},
	} {
		if err := Write(root, want); err != nil {
			t.Fatal(err)
		}
		got, err := Read(root)
		if err != nil {
			t.Fatal(err)
		}
		want.Schema = SchemaName
		if got.Schema != want.Schema || got.Version != want.Version || got.Commit != want.Commit ||
			got.Initialized != want.Initialized || (got.Upgraded == nil) != (want.Upgraded == nil) ||
			(got.Upgraded != nil && *got.Upgraded != *want.Upgraded) {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	}
}

func TestWriteEmitsNullUpgraded(t *testing.T) {
	root := t.TempDir()
	if err := Write(root, Stamp{Version: "0.1.0", Commit: "c", Initialized: "i"}); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(root, ".vloop", "install.json"))
	want := `{"schema":"install/v1","version":"0.1.0","commit":"c","initialized":"i","upgraded":null}` + "\n"
	if string(b) != want {
		t.Fatalf("got %s", b)
	}
}

func TestReadRejectsWrongSchema(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".vloop"), 0o755)
	os.WriteFile(filepath.Join(root, ".vloop", "install.json"), []byte(`{"schema":"state/v1"}`), 0o644)
	if _, err := Read(root); err == nil {
		t.Fatal("want an error")
	}
}
