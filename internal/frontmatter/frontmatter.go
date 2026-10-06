// Package frontmatter reads and rewrites the Markdown records that open with a
// "---" fenced block of "key: value" lines. A leading UTF-8 BOM is ignored on
// read and kept on write; CRLF is read as LF and written back as CRLF, every
// line, so a rewrite never mixes endings; and an edit touches only the lines
// it names.
package frontmatter

import (
	"errors"
	"slices"
	"strconv"
	"strings"
)

const (
	bom   = "\xef\xbb\xbf"
	fence = "---"
)

// Errors Parse reports.
var (
	ErrMissing      = errors.New("missing frontmatter")
	ErrUnterminated = errors.New("unterminated frontmatter")
)

// Field is one "key: value" line of the frontmatter, the value trimmed and
// unquoted.
type Field struct{ Key, Value string }

// Doc is a parsed record: the frontmatter lines and the body, with the file's
// BOM and line endings remembered for String.
type Doc struct {
	bom   bool
	crlf  bool
	lines []string // every line, without endings; a final "" stands for a trailing newline
	end   int      // index of the closing fence in lines
}

// Parse splits text into its frontmatter and body. The first line must be the
// opening fence (after an optional BOM) and a closing fence must follow. A file
// whose first line ends CRLF is a CRLF file, whatever the rest do.
func Parse(text string) (*Doc, error) {
	d := &Doc{}
	if rest, ok := strings.CutPrefix(text, bom); ok {
		d.bom, text = true, rest
	}
	first, _, _ := strings.Cut(text, "\n")
	d.crlf = strings.HasSuffix(first, "\r")
	d.lines = strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if d.lines[0] != fence {
		return nil, ErrMissing
	}
	d.end = slices.Index(d.lines[1:], fence) + 1
	if d.end == 0 {
		return nil, ErrUnterminated
	}
	return d, nil
}

// Fields are the frontmatter's "key: value" lines in order; lines with no colon
// are skipped.
func (d *Doc) Fields() []Field {
	var out []Field
	for _, l := range d.lines[1:d.end] {
		if k, v, ok := strings.Cut(l, ":"); ok {
			out = append(out, Field{k, Unquote(v)})
		}
	}
	return out
}

// Body is the text after the closing fence, with LF endings.
func (d *Doc) Body() string { return strings.Join(d.lines[d.end+1:], "\n") }

// SetBody replaces the text after the closing fence; body uses LF endings.
func (d *Doc) SetBody(body string) {
	d.lines = append(d.lines[:d.end+1:d.end+1], strings.Split(body, "\n")...)
}

// Lines are the frontmatter lines, between the fences, in order.
func (d *Doc) Lines() []string { return slices.Clone(d.lines[1:d.end]) }

// Len is the number of frontmatter lines.
func (d *Doc) Len() int { return d.end - 1 }

// Find is the index of the first frontmatter line that begins "key:", or -1.
func (d *Doc) Find(key string) int {
	for i, l := range d.lines[1:d.end] {
		if strings.HasPrefix(l, key+":") {
			return i
		}
	}
	return -1
}

// Replace sets the first "key:" line to line, or deletes it when line is empty.
// It reports whether the key was there.
func (d *Doc) Replace(key, line string) bool {
	i := d.Find(key)
	if i < 0 {
		return false
	}
	if line == "" {
		d.lines = slices.Delete(d.lines, i+1, i+2)
		d.end--
	} else {
		d.lines[i+1] = line
	}
	return true
}

// Insert puts line at frontmatter index i; Len() appends it.
func (d *Doc) Insert(i int, line string) {
	d.lines = slices.Insert(d.lines, i+1, line)
	d.end++
}

// String is the record as it is written: the BOM if it had one, every line
// ending in the file's line ending.
func (d *Doc) String() string {
	eol := "\n"
	if d.crlf {
		eol = "\r\n"
	}
	s := strings.Join(d.lines, eol)
	if d.bom {
		s = bom + s
	}
	return s
}

// Quote is s as a frontmatter value: quoted when it is empty, holds ':', '#' or
// a quote, or has surrounding space.
func Quote(s string) string {
	if s == "" || strings.ContainsAny(s, ":#\"'") || s != strings.TrimSpace(s) {
		return strconv.Quote(s)
	}
	return s
}

// Unquote is a frontmatter value trimmed, and unquoted when it is a Go-quoted
// string.
func Unquote(v string) string {
	v = strings.TrimSpace(v)
	if strings.HasPrefix(v, `"`) {
		if u, err := strconv.Unquote(v); err == nil {
			return u
		}
	}
	return v
}
