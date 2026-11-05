package install

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// SettingsPath is the Claude Code project settings file vloop declares its
// plugin in.
const SettingsPath = ".claude/settings.json"

// jnode is one JSON value read in order: an object keeps its keys' order, a
// scalar keeps its text (numbers their digits).
type jnode struct {
	kind byte // 'o' object, 'a' array, 's' scalar
	keys []string
	vals []*jnode
	elem []*jnode
	raw  []byte // the encoded scalar
}

func (n *jnode) get(k string) *jnode {
	for i, key := range n.keys {
		if key == k {
			return n.vals[i]
		}
	}
	return nil
}

func (n *jnode) set(k string, v *jnode) {
	for i, key := range n.keys {
		if key == k {
			n.vals[i] = v
			return
		}
	}
	n.keys = append(n.keys, k)
	n.vals = append(n.vals, v)
}

func encodeScalar(v any) []byte {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
	return bytes.TrimSuffix(b.Bytes(), []byte("\n"))
}

func readNode(dec *json.Decoder) (*jnode, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	d, isDelim := tok.(json.Delim)
	if !isDelim {
		return &jnode{kind: 's', raw: encodeScalar(tok)}, nil
	}
	if d == '{' {
		n := &jnode{kind: 'o'}
		for dec.More() {
			kt, err := dec.Token()
			if err != nil {
				return nil, err
			}
			v, err := readNode(dec)
			if err != nil {
				return nil, err
			}
			n.set(kt.(string), v)
		}
		if _, err := dec.Token(); err != nil {
			return nil, err
		}
		return n, nil
	}
	n := &jnode{kind: 'a'}
	for dec.More() {
		v, err := readNode(dec)
		if err != nil {
			return nil, err
		}
		n.elem = append(n.elem, v)
	}
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	return n, nil
}

// write lays n out as json.MarshalIndent does with "" and two spaces.
func (n *jnode) write(b *bytes.Buffer, depth int) {
	ind := strings.Repeat("  ", depth+1)
	switch n.kind {
	case 's':
		b.Write(n.raw)
	case 'a':
		if len(n.elem) == 0 {
			b.WriteString("[]")
			return
		}
		b.WriteString("[\n")
		for i, e := range n.elem {
			b.WriteString(ind)
			e.write(b, depth+1)
			if i < len(n.elem)-1 {
				b.WriteByte(',')
			}
			b.WriteByte('\n')
		}
		b.WriteString(strings.Repeat("  ", depth) + "]")
	default:
		if len(n.keys) == 0 {
			b.WriteString("{}")
			return
		}
		b.WriteString("{\n")
		for i, k := range n.keys {
			b.WriteString(ind)
			b.Write(encodeScalar(k))
			b.WriteString(": ")
			n.vals[i].write(b, depth+1)
			if i < len(n.keys)-1 {
				b.WriteByte(',')
			}
			b.WriteByte('\n')
		}
		b.WriteString(strings.Repeat("  ", depth) + "}")
	}
}

func objectOf() *jnode { return &jnode{kind: 'o'} }

// MergeClaudeSettings returns .claude/settings.json holding vloop's plugin
// declaration: extraKnownMarketplaces.vloop and enabledPlugins["vloop@vloop"],
// each added only when its key is absent. Everything else keeps its value and
// order. changed is false when both entries were present; the file is then to
// be left alone. A file that is not a JSON object, or whose two keys are
// present but not objects, is an error naming what to fix.
func MergeClaudeSettings(existing []byte, exists bool) (out []byte, changed bool, err error) {
	root := objectOf()
	if exists {
		dec := json.NewDecoder(bytes.NewReader(existing))
		dec.UseNumber()
		n, rerr := readNode(dec)
		if rerr == nil {
			if _, e := dec.Token(); !errors.Is(e, io.EOF) {
				rerr = errors.New("trailing data")
			}
		}
		if rerr != nil || n.kind != 'o' {
			return nil, false, fmt.Errorf("%s is not a JSON object — fix it, then re-run", SettingsPath)
		}
		root = n
	}
	sub := func(key string) (*jnode, error) {
		v := root.get(key)
		if v == nil {
			v = objectOf()
			root.set(key, v)
			changed = true
			return v, nil
		}
		if v.kind != 'o' {
			return nil, fmt.Errorf("%s's %s is not a JSON object — fix it, then re-run", SettingsPath, key)
		}
		return v, nil
	}
	mk, err := sub("extraKnownMarketplaces")
	if err != nil {
		return nil, false, err
	}
	pl, err := sub("enabledPlugins")
	if err != nil {
		return nil, false, err
	}
	if mk.get("vloop") == nil {
		src := &jnode{kind: 'o'}
		src.set("source", &jnode{kind: 's', raw: encodeScalar("github")})
		src.set("repo", &jnode{kind: 's', raw: encodeScalar("mvelosop/vloop")})
		entry := &jnode{kind: 'o'}
		entry.set("source", src)
		mk.set("vloop", entry)
		changed = true
	}
	if pl.get("vloop@vloop") == nil {
		pl.set("vloop@vloop", &jnode{kind: 's', raw: []byte("true")})
		changed = true
	}
	if !changed {
		return existing, false, nil
	}
	var b bytes.Buffer
	root.write(&b, 0)
	b.WriteByte('\n')
	return b.Bytes(), true, nil
}
