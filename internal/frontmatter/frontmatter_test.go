package frontmatter

import (
	"errors"
	"reflect"
	"testing"
)

const lf = "---\nid: D1\nstatus: open\nnote: \"a: b\"\n---\nbody one\nbody two\n"

func variants(text string) map[string]string {
	crlf := ""
	for _, r := range text {
		if r == '\n' {
			crlf += "\r"
		}
		crlf += string(r)
	}
	return map[string]string{"LF": text, "CRLF": crlf, "BOM": "\xef\xbb\xbf" + text, "BOM+CRLF": "\xef\xbb\xbf" + crlf}
}

func TestParseReadsEveryVariantAlike(t *testing.T) {
	want := []Field{{"id", "D1"}, {"status", "open"}, {"note", "a: b"}}
	for name, text := range variants(lf) {
		d, err := Parse(text)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got := d.Fields(); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: fields %v, want %v", name, got, want)
		}
		if got := d.Body(); got != "body one\nbody two\n" {
			t.Errorf("%s: body %q", name, got)
		}
		if got := d.String(); got != text {
			t.Errorf("%s: round trip %q, want %q", name, got, text)
		}
	}
}

func TestReplaceChangesOnlyItsLine(t *testing.T) {
	for name, text := range variants(lf) {
		d, _ := Parse(text)
		if !d.Replace("status", "status: fixed") {
			t.Fatalf("%s: status not found", name)
		}
		d.Insert(d.Find("id")+1, "task: T1")
		d.Insert(d.Len(), "last: x")
		want, _ := Parse(variants("---\nid: D1\ntask: T1\nstatus: fixed\nnote: \"a: b\"\nlast: x\n---\nbody one\nbody two\n")[name])
		if d.String() != want.String() {
			t.Errorf("%s: got %q, want %q", name, d.String(), want.String())
		}
		if !d.Replace("task", "") || d.Replace("task", "") {
			t.Errorf("%s: delete should succeed once", name)
		}
	}
}

func TestMixedInputWritesOneEnding(t *testing.T) {
	d, err := Parse("---\r\nid: D1\nstatus: open\r\n---\nbody\r\nmore\n")
	if err != nil {
		t.Fatal(err)
	}
	d.Replace("status", "status: fixed")
	if got, want := d.String(), "---\r\nid: D1\r\nstatus: fixed\r\n---\r\nbody\r\nmore\r\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	d, _ = Parse("---\nid: D1\r\n---\r\nbody\n")
	if got, want := d.String(), "---\nid: D1\n---\nbody\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseErrors(t *testing.T) {
	for _, tc := range []struct {
		text string
		want error
	}{
		{"", ErrMissing},
		{"id: D1\n", ErrMissing},
		{"\n---\nid: D1\n---\n", ErrMissing},
		{"---\nid: D1\nbody\n", ErrUnterminated},
		{"---\n", ErrUnterminated},
	} {
		if _, err := Parse(tc.text); !errors.Is(err, tc.want) {
			t.Errorf("Parse(%q) = %v, want %v", tc.text, err, tc.want)
		}
	}
}

func TestQuoteUnquote(t *testing.T) {
	for _, s := range []string{"plain", "", "a: b", "# x", " pad", `say "hi"`, "it's"} {
		if got := Unquote(Quote(s)); got != s {
			t.Errorf("Unquote(Quote(%q)) = %q", s, got)
		}
	}
	if Quote("plain") != "plain" || Quote("") != `""` {
		t.Error("Quote changed a plain or empty value unexpectedly")
	}
	if got := Unquote("  bare  "); got != "bare" {
		t.Errorf("Unquote trims: %q", got)
	}
	if got := Unquote(`"bad`); got != `"bad` {
		t.Errorf("Unquote of a broken quote: %q", got)
	}
}
