// Package schema validates documents against vloop's embedded JSON Schemas.
// The documents in the root schemas/ directory are the authority; nothing
// here defines a rule of its own.
package schema

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	vloop "github.com/mvelosop/vloop"
)

// Violation is one place where a document breaks its schema.
type Violation struct {
	Pointer string `json:"pointer"`
	Message string `json:"message"`
}

// UnknownError is returned for a name that is not one of the schemas.
type UnknownError struct{ Name string }

func (e *UnknownError) Error() string { return fmt.Sprintf("unknown schema %q", e.Name) }

const suffix = ".json"

// file maps "state/v1" to its embedded path "schemas/state.v1.json".
func file(name string) string {
	n, v, ok := strings.Cut(name, "/")
	if !ok || strings.ContainsAny(n+v, "./\\") {
		return ""
	}
	return "schemas/" + n + "." + v + suffix
}

// Names lists the schema names, sorted.
func Names() []string {
	ents, err := vloop.Schemas.ReadDir("schemas")
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range ents {
		base := strings.TrimSuffix(e.Name(), suffix)
		n, v, ok := strings.Cut(base, ".")
		if !ok || base == e.Name() {
			continue
		}
		names = append(names, n+"/"+v)
	}
	sort.Strings(names)
	return names
}

// Document returns the schema document for name.
func Document(name string) ([]byte, error) {
	f := file(name)
	if f == "" {
		return nil, &UnknownError{name}
	}
	b, err := vloop.Schemas.ReadFile(f)
	if err != nil {
		return nil, &UnknownError{name}
	}
	return b, nil
}

func url(name string) string { return "vloop:///" + file(name) }

func compile(name string) (*jsonschema.Schema, error) {
	doc, err := Document(name)
	if err != nil {
		return nil, err
	}
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(doc))
	if err != nil {
		return nil, fmt.Errorf("embedded schema %s: %w", name, err)
	}
	c := jsonschema.NewCompiler()
	c.UseLoader(noLoader{})
	if err := c.AddResource(url(name), inst); err != nil {
		return nil, err
	}
	return c.Compile(url(name))
}

// noLoader refuses every load: schemas are added as resources, so a $ref that
// would need loading is an error rather than a network or disk access.
type noLoader struct{}

func (noLoader) Load(u string) (any, error) { return nil, fmt.Errorf("no loading of %s", u) }

// Validate checks the JSON document data against the named schema. A document
// that is not JSON is one violation at pointer "". A nil result means valid.
func Validate(name string, data []byte) ([]Violation, error) {
	sch, err := compile(name)
	if err != nil {
		return nil, err
	}
	if msg := JSONError(data); msg != "" {
		return []Violation{{Pointer: "", Message: msg}}, nil
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var inst any
	if err := dec.Decode(&inst); err != nil {
		return nil, err
	}
	verr := sch.Validate(inst)
	if verr == nil {
		return nil, nil
	}
	var ve *jsonschema.ValidationError
	if !errors.As(verr, &ve) {
		return nil, verr
	}
	p := message.NewPrinter(language.English)
	var out []Violation
	collect(ve, p, &out)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Pointer < out[j].Pointer })
	return out, nil
}

// collect gathers the leaf causes: the most specific violations.
func collect(ve *jsonschema.ValidationError, p *message.Printer, out *[]Violation) {
	if len(ve.Causes) == 0 {
		*out = append(*out, Violation{pointer(ve.InstanceLocation), ve.ErrorKind.LocalizedString(p)})
		return
	}
	for _, c := range ve.Causes {
		collect(c, p, out)
	}
}

func pointer(loc []string) string {
	var b strings.Builder
	for _, s := range loc {
		b.WriteByte('/')
		s = strings.ReplaceAll(s, "~", "~0")
		b.WriteString(strings.ReplaceAll(s, "/", "~1"))
	}
	return b.String()
}

// JSONError says where data stops being JSON: "not valid JSON (line l, column
// c)", both 1-based, c counting characters. It is empty when data is one JSON
// document.
func JSONError(data []byte) string {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var inst any
	err := dec.Decode(&inst)
	off := len(data)
	switch {
	case err == nil:
		if _, e := dec.Token(); errors.Is(e, io.EOF) {
			return ""
		}
		rest := data[dec.InputOffset():]
		off = len(data) - len(bytes.TrimLeft(rest, " \t\r\n")) + 1
	default:
		var se *json.SyntaxError
		if errors.As(err, &se) {
			off = int(se.Offset)
		}
	}
	off = min(off, len(data))
	line := 1 + bytes.Count(data[:off], []byte("\n"))
	start := bytes.LastIndexByte(data[:off], '\n') + 1
	col := utf8.RuneCount(data[start:off])
	if col == 0 {
		col = 1
	}
	return fmt.Sprintf("not valid JSON (line %d, column %d)", line, col)
}
