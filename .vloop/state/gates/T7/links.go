// Command links checks that every relative Markdown link in the named files
// resolves: the file exists and, when the link has a #fragment, the target
// has a heading whose GitHub anchor is that fragment.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var link = regexp.MustCompile(`\]\(([^)\s]+)\)`)

func anchors(path string) (map[string]bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	seen := map[string]int{}
	out := map[string]bool{}
	fence := false
	for _, l := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			fence = !fence
			continue
		}
		if fence || !strings.HasPrefix(l, "#") {
			continue
		}
		h := strings.TrimSpace(strings.TrimLeft(l, "#"))
		var b strings.Builder
		for _, r := range strings.ToLower(h) {
			switch {
			case r == ' ':
				b.WriteRune('-')
			case r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r):
				b.WriteRune(r)
			}
		}
		a := b.String()
		if n := seen[a]; n > 0 {
			out[fmt.Sprintf("%s-%d", a, n)] = true
		} else {
			out[a] = true
		}
		seen[a]++
	}
	return out, nil
}

func main() {
	bad := 0
	for _, f := range os.Args[1:] {
		raw, err := os.ReadFile(f)
		if err != nil {
			fmt.Println(err)
			bad++
			continue
		}
		fence := false
		for _, l := range strings.Split(string(raw), "\n") {
			if strings.HasPrefix(strings.TrimSpace(l), "```") {
				fence = !fence
				continue
			}
			if fence {
				continue
			}
			for _, m := range link.FindAllStringSubmatch(l, -1) {
				u := m[1]
				if strings.Contains(u, "://") || strings.HasPrefix(u, "mailto:") {
					continue
				}
				p, frag, _ := strings.Cut(u, "#")
				target := f
				if p != "" {
					target = filepath.Join(filepath.Dir(f), filepath.FromSlash(p))
				}
				if _, err := os.Stat(target); err != nil {
					fmt.Printf("%s: %s: no such file\n", f, u)
					bad++
					continue
				}
				if frag == "" {
					continue
				}
				if !strings.HasSuffix(target, ".md") {
					continue
				}
				as, err := anchors(target)
				if err != nil || !as[frag] {
					fmt.Printf("%s: %s: no heading with that anchor\n", f, u)
					bad++
				}
			}
		}
	}
	if bad > 0 {
		os.Exit(1)
	}
}
