package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type graderResult struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
	cost   float64 // the judge calls' cost, kept out of report.json
}

// scoreGrader scores one grader of a run. workspace is the run's repo/; events
// and text are its trace. An error means the grader could not be asked at all
// (the run is then an error, never a score).
func scoreGrader(g grader, workspace string, events []event, text, judgeModel string) (graderResult, error) {
	r := graderResult{Name: strings.TrimSuffix(g.file, ".md"), Weight: 1}
	if w, ok := g.fm.values["weight"]; ok {
		f, err := strconv.ParseFloat(w, 64)
		if err != nil || f < 0 {
			r.Weight = 1
			return fail(r, "weight "+w+" is not a number"), nil
		}
		r.Weight = f
	}
	typ := g.fm.values["type"]
	switch typ {
	case "regex":
		return scoreRegex(g, r, workspace, text), nil
	case "tool_used":
		return scoreToolUsed(g, r, events), nil
	case "file_exists":
		path := g.fm.values["path"]
		if _, err := os.Stat(filepath.Join(workspace, filepath.FromSlash(path))); err == nil {
			r.Score = 1
		} else {
			r.Reason = "no " + path
		}
		return r, nil
	case "llm":
		return scoreLLM(g, r, judgeModel, workspace, events, text)
	}
	return fail(r, "unsupported grader "+typ), nil
}

func fail(r graderResult, reason string) graderResult {
	r.Score, r.Reason = 0, reason
	return r
}

func scoreRegex(g grader, r graderResult, workspace, trace string) graderResult {
	pat, err := regexp.Compile(g.fm.values["pattern"])
	if err != nil {
		return fail(r, "pattern: "+oneLine(err.Error()))
	}
	target := trace
	if g.fm.values["target"] != "trace" {
		t := g.fm.maps["target"]
		if t["source"] != "file" {
			return fail(r, "unsupported target")
		}
		data, err := os.ReadFile(filepath.Join(workspace, filepath.FromSlash(t["path"])))
		if err != nil {
			return fail(r, "no "+t["path"])
		}
		target = string(data)
	}
	found := pat.MatchString(target)
	switch g.fm.values["match"] {
	case "contains":
	case "not_contains":
		found = !found
	default:
		return fail(r, "unsupported match "+g.fm.values["match"])
	}
	if found {
		r.Score = 1
	} else {
		r.Reason = "pattern " + g.fm.values["match"] + " failed"
	}
	return r
}

func scoreToolUsed(g grader, r graderResult, events []event) graderResult {
	var in *regexp.Regexp
	if m, ok := g.fm.values["input_match"]; ok {
		var err error
		if in, err = regexp.Compile(m); err != nil {
			return fail(r, "input_match: "+oneLine(err.Error()))
		}
	}
	lo, hi := 1, -1
	for _, b := range []struct {
		key string
		p   *int
	}{{"min", &lo}, {"max", &hi}} {
		if v, ok := g.fm.values[b.key]; ok {
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				return fail(r, b.key+" "+v+" is not a count")
			}
			*b.p = n
		}
	}
	n := 0
	for _, e := range events {
		if e.Kind == "tool" && e.Tool == g.fm.values["tool"] && (in == nil || in.MatchString(e.Input)) {
			n++
		}
	}
	if n >= lo && (hi < 0 || n <= hi) {
		r.Score = 1
	} else {
		r.Reason = fmt.Sprintf("%d call(s) of %s", n, g.fm.values["tool"])
	}
	return r
}

// weightedMean is the weighted mean of the graders' scores.
func weightedMean(gs []graderResult) (float64, bool) {
	var sum, w float64
	for _, g := range gs {
		sum += g.Weight * g.Score
		w += g.Weight
	}
	if w == 0 {
		return 0, false
	}
	return sum / w, true
}
