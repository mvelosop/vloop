package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// evalCase is one directory of the cases directory, read.
type evalCase struct {
	name    string
	dir     string
	prompt  frontmatter
	body    string
	phase   string // "-" when the skill is not a loop phase
	graders []grader
}

type grader struct {
	file string
	fm   frontmatter
}

var phases = map[string]bool{"plan": true, "work": true, "review": true, "gate-review": true, "verify": true, "review-brief": true}

// caseNames lists the directories of dir whose name does not start with `_`, in name order.
func caseNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "_") {
			continue
		}
		if st, err := os.Stat(filepath.Join(dir, e.Name())); err == nil && st.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// readCase reads a case's prompt.md and graders/*.md; an error is already
// worded `<case>/<file>: <reason>`.
func readCase(root, name string) (*evalCase, error) {
	c := &evalCase{name: name, dir: filepath.Join(root, name)}
	text, err := os.ReadFile(filepath.Join(c.dir, "prompt.md"))
	if err != nil {
		return nil, fmt.Errorf("%s/prompt.md: %s", name, reason(err))
	}
	front, body, err := splitFrontmatter(string(text))
	if err != nil {
		body = string(text) // a prompt without frontmatter is all body
		front = nil
	}
	if c.prompt, err = parseFrontmatter(front); err != nil {
		return nil, fmt.Errorf("%s/prompt.md: %v", name, err)
	}
	c.body = body
	c.phase = "-"
	if f := strings.Fields(body); len(f) > 0 && strings.HasPrefix(f[0], "/vloop:") {
		if p := strings.TrimPrefix(f[0], "/vloop:"); phases[p] {
			c.phase = p
		}
	}
	files, _ := filepath.Glob(filepath.Join(c.dir, "graders", "*.md"))
	sort.Strings(files)
	for _, f := range files {
		base := filepath.Base(f)
		gt, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("%s/graders/%s: %s", name, base, reason(err))
		}
		gf, _, err := splitFrontmatter(string(gt))
		if err != nil {
			return nil, fmt.Errorf("%s/graders/%s: %v", name, base, err)
		}
		fm, err := parseFrontmatter(gf)
		if err != nil {
			return nil, fmt.Errorf("%s/graders/%s: %v", name, base, err)
		}
		c.graders = append(c.graders, grader{file: base, fm: fm})
	}
	return c, nil
}

func reason(err error) string {
	if pe, ok := err.(*os.PathError); ok {
		err = pe.Err
	}
	return err.Error()
}
