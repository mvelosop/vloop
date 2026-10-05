package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRefusesFlagLikeArgument(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "gendocs")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	for _, a := range []string{"--help", "-h", "-o"} {
		dir := t.TempDir()
		cmd := exec.Command(bin, a)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Errorf("gendocs %s exited 0", a)
		}
		if len(out) == 0 {
			t.Errorf("gendocs %s printed no usage", a)
		}
		if entries, _ := os.ReadDir(dir); len(entries) != 0 {
			t.Errorf("gendocs %s wrote %d file(s)", a, len(entries))
		}
	}
}
