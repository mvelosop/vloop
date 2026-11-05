package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

var reportKinds = map[string]bool{"run": true, "second": true, "work": true,
	"verify": true, "gate-review": true, "plan": true}

type env struct{ repo, vloop string }

// validate is `<vloop> schema validate <schema> <file>` run in the repository.
func (e *env) validate(schema, file string) bool {
	cmd := exec.Command(e.vloop, "schema", "validate", schema, file)
	cmd.Dir = e.repo
	return cmd.Run() == nil
}

func okWord(ok bool) string {
	if ok {
		return "ok"
	}
	return "INVALID"
}

// report prints what report.py <kind> printed at the starting point, with its
// session table replaced by the Sessions blocks.
func report(kind, repo, vloop, rc string) (string, error) {
	abs, err := filepath.Abs(repo)
	if err != nil {
		return "", err
	}
	e := &env{repo: abs, vloop: vloop}
	var heading string
	var tail func() []string
	switch kind {
	case "run":
		heading, tail = "## C5.1 — a real run with Copilot reviews", e.runTail
	case "second":
		heading, tail = "## B22.3 — a feedback-mode run with a copilot second reviewer", e.secondTail
	case "work":
		heading, tail = "## A1 — a real run with Copilot work sessions", e.workTail
	case "verify":
		heading, tail = "## A1 — a real run with a Copilot verify session", e.verifyTail
	case "gate-review":
		heading, tail = "## A1 — a real run with Copilot gate-review sessions", e.gateReviewTail
	default:
		heading, tail = "## A1 — a real run with Copilot plan sessions", e.planTail
	}
	lines := []string{heading, "", fmt.Sprintf("`vloop run` exit %s.", rc), ""}
	lines = append(lines, e.sessionBlocks()...)
	lines = append(lines, tail()...)
	// The Python printed the joined text plus a newline, so the report ends
	// with an empty line.
	return strings.Join(lines, "\n") + "\n\n", nil
}

func (e *env) runTail() []string {
	var out []string
	for _, f := range e.glob("reports/*-verdict.json") {
		out = append(out, fmt.Sprintf("- %s: %s, verdict/v1 %s", filepath.Base(f),
			pyStr(verdictOf(f)), okWord(e.validate("verdict/v1", f))))
	}
	return out
}

func (e *env) workTail() []string {
	out := []string{"Task outcomes (iterations.jsonl):", "",
		"| iteration | task | attempt | outcome |", "|---|---|---|---|"}
	for _, f := range e.glob("iterations.jsonl") {
		for _, it := range jsonLines(f) {
			d, ok := it.(*pyDict)
			if !ok {
				continue
			}
			out = append(out, fmt.Sprintf("| %s | %s | %s | %s |", pyStr(d.field("iteration")),
				pyStr(d.field("task")), pyStr(d.field("attempt")), pyStr(d.field("outcome"))))
		}
	}
	out = append(out, "")
	for _, f := range e.glob("reports/*-verdict.json") {
		out = append(out, fmt.Sprintf("- %s: %s", filepath.Base(f), pyStr(verdictOf(f))))
	}
	return out
}

func (e *env) verifyTail() []string {
	out := []string{"Kept report:", ""}
	reports := e.glob("reports/verify.json")
	if len(reports) == 0 {
		out = append(out, "- verify.json: not present")
	}
	for _, f := range reports {
		n, sch := "unreadable", ""
		if v, ok := readJSON(f); ok {
			if d, ok := v.(*pyDict); ok {
				n = fmt.Sprint(lenOf(d.field("findings")))
				sch, _ = d.field("schema").(string)
			}
		}
		if sch != "verify/v1" && sch != "verify/v2" {
			sch = "verify/v2"
		}
		out = append(out, fmt.Sprintf("- verify.json: present, %s %s, %s finding(s)", sch,
			okWord(e.validate(sch, f)), n))
	}
	return out
}

func (e *env) gateReviewTail() []string {
	out := []string{"Kept verdicts (reports/gate-review-<round>.json):", "",
		"| round | verdict | gate-verdict/v2 | findings by kind |", "|---|---|---|---|"}
	found := false
	for _, f := range e.glob("reports/gate-review-*.json") {
		found = true
		b := filepath.Base(f)
		rnd := b[len("gate-review-") : len(b)-len(".json")]
		ok := e.validate("gate-verdict/v2", f)
		kinds, verdict := "\"unreadable\"", "unreadable"
		if v, rd := readJSON(f); rd {
			d, _ := v.(*pyDict)
			kinds, verdict = kindCounts(d), pyStr(d.field("verdict"))
		}
		out = append(out, fmt.Sprintf("| %s | present: %s | %s | %s |", rnd, verdict, okWord(ok), kinds))
	}
	if !found {
		out = append(out, "| - | not present | - | - |")
	}
	return out
}

// kindCounts is json.dumps of the findings counted by kind, sort_keys=True.
func kindCounts(d *pyDict) string {
	counts := &pyDict{vals: map[string]any{}}
	if d != nil {
		for _, t := range iterate(d.field("tasks")) {
			td, ok := t.(*pyDict)
			if !ok {
				continue
			}
			for _, x := range iterate(td.field("findings")) {
				xd, ok := x.(*pyDict)
				if !ok {
					continue
				}
				key := "null" // a missing or null kind is a None key, which json.dumps writes as null
				switch k := xd.field("kind").(type) {
				case string:
					key = k
				case bool:
					key = fmt.Sprint(k)
				case pyInt:
					key = string(k)
				case pyFloat:
					key = floatRepr(float64(k))
				}
				n := 0
				if c, ok := counts.vals[key].(pyInt); ok {
					fmt.Sscan(string(c), &n)
				}
				counts.set(key, pyInt(fmt.Sprint(n+1)))
			}
		}
	}
	return dumps(counts, true)
}

func (e *env) planTail() []string {
	out := []string{"The plan (.vloop/state/state.json):", ""}
	sf := filepath.Join(e.repo, ".vloop/state/state.json")
	if _, err := os.Stat(sf); err != nil {
		out = append(out, "- state.json: not present")
	} else {
		ok := e.validate("state/v2", sf)
		count, tasks := "unreadable", pyList(nil)
		if v, rd := readJSON(sf); rd {
			if d, isDict := v.(*pyDict); isDict {
				tasks, _ = d.field("tasks").(pyList)
			}
			count = fmt.Sprint(len(tasks))
		}
		out = append(out, fmt.Sprintf("- state.json: present, state/v2 %s, %s task(s)", okWord(ok), count))
		for _, t := range tasks {
			td, ok := t.(*pyDict)
			if !ok {
				continue
			}
			out = append(out, fmt.Sprintf("  - %s (%s, %s): %s", pyStr(td.field("id")),
				pyStr(td.field("kind")), pyStr(td.field("status")), pyStr(td.field("title"))))
		}
	}
	out = append(out, "", "Gate-review verdict per round (reports/gate-review-<round>.json):", "",
		"| round | verdict | gate-verdict/v2 |", "|---|---|---|")
	found := false
	for _, f := range e.glob("reports/gate-review-*.json") {
		found = true
		b := filepath.Base(f)
		rnd := b[len("gate-review-") : len(b)-len(".json")]
		out = append(out, fmt.Sprintf("| %s | %s | %s |", rnd, pyStr(verdictOf(f)),
			okWord(e.validate("gate-verdict/v2", f))))
	}
	if !found {
		out = append(out, "| - | not present | - |")
	}
	return out
}

func (e *env) secondTail() []string {
	out := []string{"Second-review lines of the run log:", "", "```"}
	for _, f := range append(e.glob("run.log"), e.globShallow("run.log")...) {
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		for _, l := range pySplitlines(string(b)) {
			if strings.Contains(l, "second review") {
				out = append(out, l)
			}
		}
	}
	out = append(out, "```", "", "Dissents (iterations.jsonl records with a second_review that dissents):", "", "```")
	for _, f := range append(e.glob("iterations.jsonl"), e.globShallow("iterations.jsonl")...) {
		for _, it := range jsonLines(f) {
			d, ok := it.(*pyDict)
			if !ok {
				continue
			}
			sr, ok := d.field("second_review").(*pyDict)
			if ok && truthy(sr.field("dissent")) {
				out = append(out, dumps(sr, false))
			}
		}
	}
	return append(out, "```")
}

// glob is the sorted .vloop/state/runs/*/*/<rest> of the repository, with
// Python's glob rules (a * does not match a name that starts with a dot).
func (e *env) glob(rest string) []string {
	return e.globFrom(append([]string{".vloop", "state", "runs", "*", "*"}, strings.Split(rest, "/")...))
}

// globShallow is the sorted .vloop/state/runs/*/<rest>.
func (e *env) globShallow(rest string) []string {
	return e.globFrom(append([]string{".vloop", "state", "runs", "*"}, strings.Split(rest, "/")...))
}

func (e *env) globFrom(segs []string) []string {
	paths := []string{e.repo}
	for _, s := range segs {
		var next []string
		for _, p := range paths {
			if !strings.ContainsAny(s, "*?[") {
				if _, err := os.Lstat(filepath.Join(p, s)); err == nil {
					next = append(next, filepath.Join(p, s))
				}
				continue
			}
			ents, err := os.ReadDir(p)
			if err != nil {
				continue
			}
			for _, d := range ents {
				if strings.HasPrefix(d.Name(), ".") && !strings.HasPrefix(s, ".") {
					continue
				}
				if ok, _ := path.Match(s, d.Name()); ok {
					next = append(next, filepath.Join(p, d.Name()))
				}
			}
		}
		paths = next
	}
	sort.Strings(paths)
	return paths
}

func subdirs(dir string) []string {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, d := range ents {
		if st, err := os.Stat(filepath.Join(dir, d.Name())); err == nil && st.IsDir() {
			out = append(out, d.Name())
		}
	}
	sort.Strings(out)
	return out
}

func readJSON(f string) (any, bool) {
	b, err := os.ReadFile(f)
	if err != nil {
		return nil, false
	}
	v, err := parseJSON(string(b))
	return v, err == nil
}

// verdictOf is json.load(f).get("verdict"), "unreadable" when f is not JSON.
func verdictOf(f string) any {
	v, ok := readJSON(f)
	if !ok {
		return "unreadable"
	}
	d, _ := v.(*pyDict)
	if d == nil {
		return nil
	}
	return d.field("verdict")
}

// jsonLines are the JSON values of a file's lines, the lines that are not JSON skipped.
func jsonLines(f string) []any {
	b, err := os.ReadFile(f)
	if err != nil {
		return nil
	}
	s := strings.ReplaceAll(strings.ReplaceAll(string(b), "\r\n", "\n"), "\r", "\n")
	var out []any
	for _, l := range strings.Split(s, "\n") {
		if v, err := parseJSON(l); err == nil {
			out = append(out, v)
		}
	}
	return out
}

func truthy(v any) bool {
	switch x := v.(type) {
	case nil, pyNull:
		return false
	case bool:
		return x
	case pyInt:
		return x != "0"
	case pyFloat:
		return x != 0
	case string:
		return x != ""
	case pyList:
		return len(x) > 0
	case *pyDict:
		return len(x.keys) > 0
	}
	return false
}

// lenOf is len(v or []).
func lenOf(v any) int {
	switch x := v.(type) {
	case pyList:
		return len(x)
	case *pyDict:
		return len(x.keys)
	case string:
		return len([]rune(x))
	}
	return 0
}

// pySplitlines is str.splitlines().
func pySplitlines(s string) []string {
	var out []string
	start := 0
	rs := []rune(s)
	for i := 0; i < len(rs); i++ {
		switch rs[i] {
		case '\n', '\v', '\f', '\x1c', '\x1d', '\x1e', '\x85', ' ', ' ', '\r':
			out = append(out, string(rs[start:i]))
			if rs[i] == '\r' && i+1 < len(rs) && rs[i+1] == '\n' {
				i++
			}
			start = i + 1
		}
	}
	if start < len(rs) {
		out = append(out, string(rs[start:]))
	}
	return out
}
