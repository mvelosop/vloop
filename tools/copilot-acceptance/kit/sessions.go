package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var sessionFile = regexp.MustCompile(`^([0-9]+)-([a-z-]+)\.json$`)

var sessionPhases = map[string]bool{"plan": true, "work": true, "review": true,
	"second-review": true, "gate-review": true, "verify": true}

// runIDs are the brief ids with a run folder and a brief file, in name order.
func (e *env) runIDs() []string {
	ents, err := os.ReadDir(filepath.Join(e.repo, ".vloop/state/runs"))
	if err != nil {
		return nil
	}
	var ids []string
	for _, d := range ents {
		if !d.IsDir() {
			continue
		}
		st, err := os.Stat(filepath.Join(e.repo, "docs/briefs", d.Name()+".loop-brief.md"))
		if err == nil && st.Mode().IsRegular() {
			ids = append(ids, d.Name())
		}
	}
	sort.Strings(ids)
	return ids
}

// sessionBlocks replaces a report's session table: one Sessions block per
// qualifying run id, or the one line saying there is none.
func (e *env) sessionBlocks() []string {
	ids := e.runIDs()
	if len(ids) == 0 {
		return []string{"Sessions: no brief with runs.", ""}
	}
	var out []string
	for _, id := range ids {
		out = append(out, e.sessionBlock(id)...)
	}
	return out
}

type recFile struct {
	seq   string // digits without leading zeros ("0" for zero)
	name  string
	phase string
}

func trimZeros(s string) string {
	s = strings.TrimLeft(s, "0")
	if s == "" {
		return "0"
	}
	return s
}

func lessRec(a, b recFile) bool {
	if len(a.seq) != len(b.seq) {
		return len(a.seq) < len(b.seq)
	}
	if a.seq != b.seq {
		return a.seq < b.seq
	}
	if a.name != b.name {
		return a.name < b.name
	}
	return a.phase < b.phase
}

func (e *env) sessionBlock(id string) []string {
	var cmdOut, cmdErr bytes.Buffer
	cmd := exec.Command(e.vloop, "metrics", id, "--by", "session")
	cmd.Dir = e.repo
	cmd.Stdout, cmd.Stderr = &cmdOut, &cmdErr
	code := 0
	if err := cmd.Run(); err != nil {
		code = 1
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			cmdErr.WriteString(err.Error())
		}
	}
	var fenced string
	if code == 0 {
		fenced = cmdOut.String()
		if !strings.HasSuffix(fenced, "\n") {
			fenced += "\n"
		}
	} else {
		fenced = fmt.Sprintf("vloop metrics exited %d: %s\n", code,
			strings.ReplaceAll(strings.TrimSpace(cmdErr.String()), "\n", " "))
	}
	lines := []string{fmt.Sprintf("Sessions (`vloop metrics %s --by session`):", id), "", "```"}
	lines = append(lines, strings.Split(strings.TrimSuffix(fenced, "\n"), "\n")...)
	lines = append(lines, "```", "")

	n, k, noschema := 0, 0, 0
	counts := map[string]int{}
	var recs, dens []string
	rdir := filepath.Join(e.repo, ".vloop/state/runs", id)
	for _, folder := range subdirs(rdir) {
		sdir := filepath.Join(rdir, folder, "sessions")
		ents, err := os.ReadDir(sdir)
		if err != nil {
			continue
		}
		var files []recFile
		for _, f := range ents {
			if m := sessionFile.FindStringSubmatch(f.Name()); m != nil {
				files = append(files, recFile{trimZeros(m[1]), f.Name(), m[2]})
			}
		}
		sort.Slice(files, func(i, j int) bool { return lessRec(files[i], files[j]) })
		for _, f := range files {
			n++
			rel := ".vloop/state/runs/" + id + "/" + folder + "/sessions/" + f.name
			v, ok := readJSON(filepath.Join(sdir, f.name))
			if !ok {
				recs = append(recs, "- "+rel+": unreadable")
				continue
			}
			r, _ := v.(*pyDict)
			sch := ""
			if r != nil {
				sch, _ = r.field("schema").(string)
			}
			if sch == "" {
				noschema++
				recs = append(recs, "- "+rel+": no schema")
			} else {
				counts[sch]++
				if e.validate(sch, rel) {
					k++
				} else {
					recs = append(recs, "- "+rel+": "+sch+" INVALID")
				}
			}
			if r == nil {
				continue
			}
			phase := f.phase
			if p := r.field("phase"); truthy(p) {
				phase, _ = p.(string)
			}
			if !sessionPhases[phase] {
				continue
			}
			for _, d := range iterate(r.field("permission_denials")) {
				tn := ""
				if dd, ok := d.(*pyDict); ok {
					tn, _ = dd.field("tool_name").(string)
				}
				seq := f.seq
				for len(seq) < 3 {
					seq = "0" + seq
				}
				dens = append(dens, fmt.Sprintf("- %s/%s %s %s: %s", folder, seq, phase, tn,
					strings.ReplaceAll(denialText(d), "\n", " ")))
			}
		}
	}
	if n == 0 {
		lines = append(lines, "Session records: 0.")
	} else {
		var schemas []string
		for s := range counts {
			schemas = append(schemas, s)
		}
		sort.Strings(schemas)
		var parts []string
		for _, s := range schemas {
			parts = append(parts, s+" "+strconv.Itoa(counts[s]))
		}
		if noschema > 0 {
			parts = append(parts, "no schema "+strconv.Itoa(noschema))
		}
		paren := ""
		if len(parts) > 0 {
			paren = " (" + strings.Join(parts, ", ") + ")"
		}
		lines = append(lines, fmt.Sprintf("Session records: %d, schema-valid %d%s.", n, k, paren))
	}
	lines = append(lines, recs...)
	if len(dens) > 0 {
		lines = append(lines, "", "Permission denials:", "")
		lines = append(lines, dens...)
	}
	return append(lines, "")
}

// denialText is the entry's first non-empty string of message, command,
// file_path, tool_input.command and tool_input.file_path, else "".
func denialText(d any) string {
	dd, ok := d.(*pyDict)
	if !ok {
		return ""
	}
	cands := []any{dd.field("message"), dd.field("command"), dd.field("file_path")}
	if ti, ok := dd.field("tool_input").(*pyDict); ok {
		cands = append(cands, ti.field("command"), ti.field("file_path"))
	}
	for _, c := range cands {
		if s, ok := c.(string); ok && s != "" {
			return s
		}
	}
	return ""
}

// iterate is `for x in v or []`: a list's items, a dict's keys.
func iterate(v any) []any {
	switch x := v.(type) {
	case pyList:
		return x
	case *pyDict:
		out := make([]any, len(x.keys))
		for i, k := range x.keys {
			out[i] = k
		}
		return out
	}
	return nil
}
