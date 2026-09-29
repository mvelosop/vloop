package brief

import (
	"embed"
	"fmt"
	"regexp"
	"strings"
	"time"
)

//go:embed templates/en.md templates/es.md
var templates embed.FS

var slugRE = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// ValidSlug reports whether s is lower-case letters, digits and single hyphens,
// with no leading or trailing hyphen.
func ValidSlug(s string) bool { return slugRE.MatchString(s) }

// NewName is the brief's name, its filename minus .md, for a local time.
func NewName(slug string, t time.Time) string {
	return "B" + t.Format("20060102-1504") + "-" + slug + ".loop-brief"
}

// NewPath is the root-relative path of a new brief.
func NewPath(slug string, t time.Time) string {
	return Dir + "/" + NewName(slug, t) + ".md"
}

// Render returns the embedded template for language with name and today filled
// in. An unknown language gets English.
func Render(language, slug string, t time.Time) (string, error) {
	file := "templates/en.md"
	if language == "es" {
		file = "templates/es.md"
	}
	data, err := templates.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("template %s: %w", file, err)
	}
	r := strings.NewReplacer("{{NAME}}", NewName(slug, t), "{{CREATED}}", t.Format("2006-01-02"))
	return r.Replace(string(data)), nil
}
