package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// B7's close: the extracted plugin holds exactly the four skills and no evals,
// and an intervention is added, listed and exported with repo.stacks.

func TestWorkedExampleB7PluginPath(t *testing.T) {
	s := newScratch(t)
	res := s.run(nil, "plugin", "path")
	if res.code != 0 {
		t.Fatalf("plugin path: %+v", res)
	}
	root := filepath.Join(s.dir, filepath.FromSlash(strings.TrimSpace(res.out)))
	entries, err := os.ReadDir(filepath.Join(root, "skills"))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, e := range entries {
		got = append(got, e.Name())
		if _, err := os.Stat(filepath.Join(root, "skills", e.Name(), "SKILL.md")); err != nil {
			t.Errorf("skill %s has no SKILL.md: %v", e.Name(), err)
		}
	}
	if strings.Join(got, " ") != "gate-review operate plan review work" {
		t.Errorf("extracted skills = %q, want gate-review operate plan review work", got)
	}
	if _, err := os.Stat(filepath.Join(root, "evals")); !os.IsNotExist(err) {
		t.Errorf("evals/ was extracted (stat err = %v)", err)
	}
}

func TestWorkedExampleB7Interventions(t *testing.T) {
	s := newScratch(t)
	s.write(".vloop/config.toml", "[metrics]\nstacks = [\"go\"]\n")

	add := s.run(nil, "intervention", "add", "Supplied the missing brief context",
		"--kind", "context-supply", "--phase", "design", "--by", "operator", "--automatable", "partly",
		"--trigger", "the brief was silent", "--done", "added the paragraph", "--automation", "none")
	if add.code != 0 {
		t.Fatalf("intervention add: %+v", add)
	}
	path := strings.TrimSpace(add.out)
	if !strings.HasPrefix(path, ".vloop/interventions/I") || !strings.HasSuffix(path, "-supplied-the-missing-brief-context.md") {
		t.Fatalf("intervention add printed %q", add.out)
	}
	if _, err := os.Stat(filepath.Join(s.dir, filepath.FromSlash(path))); err != nil {
		t.Fatalf("record not written: %v", err)
	}

	list := s.run(nil, "intervention", "list")
	if list.code != 0 || !strings.Contains(list.out, "context-supply") || !strings.Contains(list.out, "supplied-the-missing-brief-context") {
		t.Errorf("intervention list = %+v", list)
	}

	exp := s.run(nil, "metrics", "export")
	if exp.code != 0 {
		t.Fatalf("metrics export: %+v", exp)
	}
	found := 0
	for _, line := range strings.Split(strings.TrimSpace(exp.out), "\n") {
		var rec struct {
			Type string `json:"type"`
			Repo struct {
				Stacks []string `json:"stacks"`
			} `json:"repo"`
		}
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("export line %q: %v", line, err)
		}
		if strings.Join(rec.Repo.Stacks, ",") != "go" {
			t.Errorf("record %q repo.stacks = %v, want [go]", rec.Type, rec.Repo.Stacks)
		}
		if rec.Type == "intervention" {
			found++
		}
	}
	if found != 1 {
		t.Errorf("export carries %d intervention records, want 1:\n%s", found, exp.out)
	}
}
