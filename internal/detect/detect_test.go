package detect

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func fixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, body := range files {
		p := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestStacks(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{"nothing", map[string]string{"NOTES.md": "hi"}, nil},
		{"go root", map[string]string{"go.mod": "x"}, []string{"go"}},
		{"go sub", map[string]string{"svc/go.mod": "x"}, []string{"go@svc"}},
		{"js", map[string]string{"package.json": `{"name":"x"}`}, []string{"javascript"}},
		{"ts tsconfig beside", map[string]string{"package.json": "{}", "tsconfig.json": "{}"}, []string{"javascript", "typescript"}},
		{"ts dev dependency", map[string]string{"web/package.json": `{"devDependencies":{"typescript":"5"}}`}, []string{"javascript@web", "typescript@web"}},
		{"tsconfig elsewhere", map[string]string{"web/package.json": "{}", "web/src/tsconfig.json": "{}", "tsconfig.json": "{}"}, []string{"javascript@web"}},
		{"tsconfig alone", map[string]string{"tsconfig.json": "{}"}, nil},
		{"react dependency", map[string]string{"package.json": `{"dependencies":{"react":"18"}}`}, []string{"javascript", "react"}},
		{"react only mentioned", map[string]string{"package.json": `{"name":"react-tools","description":"react","dependencies":{"preact":"1"}}`}, []string{"javascript"}},
		{"malformed package.json", map[string]string{"package.json": "{"}, []string{"javascript"}},
		{"csharp solution", map[string]string{"s/Api.sln": "", "s/Api/Api.csproj": "", "s/Api.Tests/T.csproj": ""}, []string{"csharp@s"}},
		{"csharp root solution", map[string]string{"App.sln": "", "src/App/App.csproj": ""}, []string{"csharp"}},
		{"csharp projects", map[string]string{"a/A.csproj": "", "b/c/C.csproj": ""}, []string{"csharp@a", "csharp@b/c"}},
		{"csharp sln beside csproj", map[string]string{"lib/Lib.sln": "", "lib/Lib.csproj": ""}, []string{"csharp@lib"}},
		{"python pyproject", map[string]string{"pyproject.toml": ""}, []string{"python"}},
		{"python requirements", map[string]string{"tools/scripts/requirements.txt": ""}, []string{"python@tools/scripts"}},
		{"python setup", map[string]string{"pkg/setup.py": ""}, []string{"python@pkg"}},
		{"java pom", map[string]string{"pom.xml": ""}, []string{"java"}},
		{"java gradle", map[string]string{"app/build.gradle": ""}, []string{"java@app"}},
		{"kotlin not java", map[string]string{"app/build.gradle.kts": ""}, []string{"kotlin@app"}},
		{"rust", map[string]string{"crates/core/Cargo.toml": ""}, []string{"rust@crates/core"}},
		{"depth three", map[string]string{"a/b/c/go.mod": "", "a/b/c/d/Cargo.toml": "", "x/y/z/w/package.json": "{}"}, []string{"go@a/b/c"}},
		{"ordering", map[string]string{"go.mod": "", "Cargo.toml": "", "b/package.json": "{}", "a/pyproject.toml": "", "a/package.json": "{}"},
			[]string{"go", "rust", "javascript@a", "javascript@b", "python@a"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Stacks(fixture(t, tc.files))
			if err != nil {
				t.Fatal(err)
			}
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestStacksSkippedDirectories(t *testing.T) {
	files := map[string]string{
		"web/node_modules/react/package.json": "{}",
		"svc/vendor/x/go.mod":                 "x",
		"svc/build/pom.xml":                   "",
		"svc/obj/X.csproj":                    "",
	}
	for _, s := range []string{"node_modules", "vendor", "bin", "obj", "dist", "target", "build", ".loop", ".vloop", ".git"} {
		files[s+"/go.mod"] = "x"
		files[s+"/package.json"] = "{}"
	}
	got, err := Stacks(fixture(t, files))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("got %v, want none", got)
	}
}

func TestStacksSymlinkNotFollowed(t *testing.T) {
	outside := fixture(t, map[string]string{"go.mod": "x"})
	root := fixture(t, map[string]string{"NOTES.md": "x"})
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skip("symlinks unavailable:", err)
	}
	got, err := Stacks(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("got %v, want none", got)
	}
}
