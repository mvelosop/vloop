package main

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// The reports print JSON values the way Python does: json.dumps for some,
// str() for others. This file holds the value model that keeps what Python
// keeps (key order, int against float, None against missing) and the two
// renderers.

type (
	pyNull  struct{}
	pyInt   string // the integer's digits, normalized
	pyFloat float64
	pyList  []any
	pyDict  struct {
		keys []string
		vals map[string]any
	}
)

func (d *pyDict) get(k string) (any, bool) {
	v, ok := d.vals[k]
	return v, ok
}

// field is dict.get(k): nil when absent, pyNull{} when null.
func (d *pyDict) field(k string) any {
	if d == nil {
		return nil
	}
	return d.vals[k]
}

func (d *pyDict) set(k string, v any) {
	if _, ok := d.vals[k]; !ok {
		d.keys = append(d.keys, k)
	}
	d.vals[k] = v
}

type parser struct {
	s   string
	pos int
}

var errJSON = errors.New("invalid JSON")

// parseJSON reads one JSON document as Python's json.loads does, and fails
// where it fails (NaN, Infinity and lone surrogates aside).
func parseJSON(s string) (any, error) {
	if !utf8.ValidString(s) {
		return nil, errJSON
	}
	p := &parser{s: s}
	p.ws()
	v, err := p.value()
	if err != nil {
		return nil, err
	}
	p.ws()
	if p.pos != len(p.s) {
		return nil, errJSON
	}
	return v, nil
}

func (p *parser) ws() {
	for p.pos < len(p.s) && strings.IndexByte(" \t\n\r", p.s[p.pos]) >= 0 {
		p.pos++
	}
}

func (p *parser) lit(w string, v any) (any, error) {
	if strings.HasPrefix(p.s[p.pos:], w) {
		p.pos += len(w)
		return v, nil
	}
	return nil, errJSON
}

func (p *parser) value() (any, error) {
	if p.pos >= len(p.s) {
		return nil, errJSON
	}
	switch c := p.s[p.pos]; {
	case c == '{':
		p.pos++
		d := &pyDict{vals: map[string]any{}}
		p.ws()
		if p.pos < len(p.s) && p.s[p.pos] == '}' {
			p.pos++
			return d, nil
		}
		for {
			p.ws()
			if p.pos >= len(p.s) || p.s[p.pos] != '"' {
				return nil, errJSON
			}
			k, err := p.str()
			if err != nil {
				return nil, err
			}
			p.ws()
			if p.pos >= len(p.s) || p.s[p.pos] != ':' {
				return nil, errJSON
			}
			p.pos++
			p.ws()
			v, err := p.value()
			if err != nil {
				return nil, err
			}
			d.set(k, v)
			p.ws()
			if p.pos >= len(p.s) {
				return nil, errJSON
			}
			if p.s[p.pos] == ',' {
				p.pos++
				continue
			}
			if p.s[p.pos] == '}' {
				p.pos++
				return d, nil
			}
			return nil, errJSON
		}
	case c == '[':
		p.pos++
		l := pyList{}
		p.ws()
		if p.pos < len(p.s) && p.s[p.pos] == ']' {
			p.pos++
			return l, nil
		}
		for {
			p.ws()
			v, err := p.value()
			if err != nil {
				return nil, err
			}
			l = append(l, v)
			p.ws()
			if p.pos >= len(p.s) {
				return nil, errJSON
			}
			if p.s[p.pos] == ',' {
				p.pos++
				continue
			}
			if p.s[p.pos] == ']' {
				p.pos++
				return l, nil
			}
			return nil, errJSON
		}
	case c == '"':
		return p.str()
	case c == 't':
		return p.lit("true", true)
	case c == 'f':
		return p.lit("false", false)
	case c == 'n':
		return p.lit("null", pyNull{})
	case c == '-' || (c >= '0' && c <= '9'):
		return p.num()
	}
	return nil, errJSON
}

func (p *parser) num() (any, error) {
	st := p.pos
	if p.s[p.pos] == '-' {
		p.pos++
	}
	digits := func() int {
		n := 0
		for p.pos < len(p.s) && p.s[p.pos] >= '0' && p.s[p.pos] <= '9' {
			p.pos++
			n++
		}
		return n
	}
	if p.pos >= len(p.s) {
		return nil, errJSON
	}
	if p.s[p.pos] == '0' {
		p.pos++
	} else if digits() == 0 {
		return nil, errJSON
	}
	isFloat := false
	if p.pos+1 < len(p.s) && p.s[p.pos] == '.' && p.s[p.pos+1] >= '0' && p.s[p.pos+1] <= '9' {
		isFloat = true
		p.pos++
		digits()
	}
	if p.pos < len(p.s) && (p.s[p.pos] == 'e' || p.s[p.pos] == 'E') {
		q := p.pos + 1
		if q < len(p.s) && (p.s[q] == '+' || p.s[q] == '-') {
			q++
		}
		if q < len(p.s) && p.s[q] >= '0' && p.s[q] <= '9' {
			isFloat = true
			p.pos = q
			digits()
		}
	}
	lit := p.s[st:p.pos]
	if !isFloat {
		neg := strings.HasPrefix(lit, "-")
		d := strings.TrimPrefix(lit, "-")
		if neg && d == "0" {
			neg = false
		}
		if neg {
			d = "-" + d
		}
		return pyInt(d), nil
	}
	f, err := strconv.ParseFloat(lit, 64)
	if err != nil && !math.IsInf(f, 0) {
		return nil, errJSON
	}
	return pyFloat(f), nil
}

func (p *parser) str() (string, error) {
	p.pos++ // the opening quote
	var b strings.Builder
	for {
		if p.pos >= len(p.s) {
			return "", errJSON
		}
		c := p.s[p.pos]
		switch {
		case c == '"':
			p.pos++
			return b.String(), nil
		case c < 0x20:
			return "", errJSON
		case c == '\\':
			p.pos++
			if p.pos >= len(p.s) {
				return "", errJSON
			}
			e := p.s[p.pos]
			p.pos++
			switch e {
			case '"', '\\', '/':
				b.WriteByte(e)
			case 'b':
				b.WriteByte('\b')
			case 'f':
				b.WriteByte('\f')
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case 'u':
				r, ok := p.hex4()
				if !ok {
					return "", errJSON
				}
				if r >= 0xD800 && r < 0xDC00 && strings.HasPrefix(p.s[p.pos:], `\u`) {
					save := p.pos
					p.pos += 2
					r2, ok := p.hex4()
					if ok && r2 >= 0xDC00 && r2 < 0xE000 {
						r = 0x10000 + (r-0xD800)<<10 + (r2 - 0xDC00)
					} else {
						p.pos = save
					}
				}
				b.WriteRune(r) // a lone surrogate becomes U+FFFD
			default:
				return "", errJSON
			}
		default:
			b.WriteByte(c)
			p.pos++
		}
	}
}

func (p *parser) hex4() (rune, bool) {
	if p.pos+4 > len(p.s) {
		return 0, false
	}
	n, err := strconv.ParseUint(p.s[p.pos:p.pos+4], 16, 32)
	if err != nil || strings.ContainsAny(p.s[p.pos:p.pos+4], "+-_") {
		return 0, false
	}
	p.pos += 4
	return rune(n), true
}

// floatRepr is Python's repr(float).
func floatRepr(f float64) string {
	switch {
	case math.IsInf(f, 1):
		return "inf"
	case math.IsInf(f, -1):
		return "-inf"
	case math.IsNaN(f):
		return "nan"
	}
	e := strconv.FormatFloat(f, 'e', -1, 64) // d.ddde±XX
	neg := strings.HasPrefix(e, "-")
	e = strings.TrimPrefix(e, "-")
	mant, expS, _ := strings.Cut(e, "e")
	exp, _ := strconv.Atoi(expS)
	digits := strings.Replace(mant, ".", "", 1)
	var out string
	if exp >= -4 && exp < 16 {
		switch {
		case exp < 0:
			out = "0." + strings.Repeat("0", -exp-1) + digits
		case len(digits) <= exp+1:
			out = digits + strings.Repeat("0", exp+1-len(digits)) + ".0"
		default:
			out = digits[:exp+1] + "." + digits[exp+1:]
		}
	} else {
		out = digits[:1]
		if len(digits) > 1 {
			out += "." + digits[1:]
		}
		sign := "+"
		if exp < 0 {
			sign, exp = "-", -exp
		}
		out += "e" + sign + pad2(exp)
	}
	if neg {
		out = "-" + out
	}
	return out
}

func pad2(n int) string {
	s := strconv.Itoa(n)
	if len(s) < 2 {
		return "0" + s
	}
	return s
}

// pyStr is str(v).
func pyStr(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return pyRepr(v)
}

// pyRepr is repr(v).
func pyRepr(v any) string {
	switch x := v.(type) {
	case nil, pyNull:
		return "None"
	case bool:
		if x {
			return "True"
		}
		return "False"
	case pyInt:
		return string(x)
	case pyFloat:
		return floatRepr(float64(x))
	case string:
		return strRepr(x)
	case pyList:
		parts := make([]string, len(x))
		for i, e := range x {
			parts[i] = pyRepr(e)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case *pyDict:
		parts := make([]string, len(x.keys))
		for i, k := range x.keys {
			parts[i] = strRepr(k) + ": " + pyRepr(x.vals[k])
		}
		return "{" + strings.Join(parts, ", ") + "}"
	}
	return ""
}

func strRepr(s string) string {
	q := byte('\'')
	if strings.Contains(s, "'") && !strings.Contains(s, `"`) {
		q = '"'
	}
	var b strings.Builder
	b.WriteByte(q)
	for _, r := range s {
		switch {
		case r == rune(q) || r == '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case r < 0x20 || r == 0x7f:
			b.WriteString(`\x` + hex(int(r), 2))
		case r < 0x7f || unicode.IsPrint(r):
			b.WriteRune(r)
		case r <= 0xff:
			b.WriteString(`\x` + hex(int(r), 2))
		case r <= 0xffff:
			b.WriteString(`\u` + hex(int(r), 4))
		default:
			b.WriteString(`\U` + hex(int(r), 8))
		}
	}
	b.WriteByte(q)
	return b.String()
}

func hex(n, width int) string {
	s := strconv.FormatInt(int64(n), 16)
	for len(s) < width {
		s = "0" + s
	}
	return s
}

// dumps is json.dumps(v) (sortKeys: sort_keys=True) with the default
// separators and ensure_ascii.
func dumps(v any, sortKeys bool) string {
	var b strings.Builder
	dumpTo(&b, v, sortKeys)
	return b.String()
}

func dumpTo(b *strings.Builder, v any, sortKeys bool) {
	switch x := v.(type) {
	case nil, pyNull:
		b.WriteString("null")
	case bool:
		if x {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case pyInt:
		b.WriteString(string(x))
	case pyFloat:
		f := float64(x)
		switch {
		case math.IsInf(f, 1):
			b.WriteString("Infinity")
		case math.IsInf(f, -1):
			b.WriteString("-Infinity")
		default:
			b.WriteString(floatRepr(f))
		}
	case string:
		dumpStr(b, x)
	case pyList:
		b.WriteByte('[')
		for i, e := range x {
			if i > 0 {
				b.WriteString(", ")
			}
			dumpTo(b, e, sortKeys)
		}
		b.WriteByte(']')
	case *pyDict:
		keys := x.keys
		if sortKeys {
			keys = append([]string(nil), keys...)
			sortStrings(keys)
		}
		b.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				b.WriteString(", ")
			}
			dumpStr(b, k)
			b.WriteString(": ")
			dumpTo(b, x.vals[k], sortKeys)
		}
		b.WriteByte('}')
	}
}

func dumpStr(b *strings.Builder, s string) {
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"':
			b.WriteString(`\"`)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case r == '\b':
			b.WriteString(`\b`)
		case r == '\f':
			b.WriteString(`\f`)
		case r < 0x20:
			b.WriteString(`\u` + hex(int(r), 4))
		case r < 0x7f:
			b.WriteRune(r)
		case r < 0x10000:
			b.WriteString(`\u` + hex(int(r), 4))
		default:
			r -= 0x10000
			b.WriteString(`\u` + hex(0xD800+int(r>>10), 4) + `\u` + hex(0xDC00+int(r&0x3ff), 4))
		}
	}
	b.WriteByte('"')
}
