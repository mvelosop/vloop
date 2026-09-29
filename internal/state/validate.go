package state

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Report is what Check found: problems fail the plan, warnings do not.
type Report struct {
	Problems []string
	Warnings []string
}

// Check validates the plan file under root: the state/v1 schema, then the
// structural rules the schema cannot express. areas is the repo's configured
// area list; when empty, a task's area is optional and unchecked. It never
// writes. A missing plan is ErrNoPlan.
func Check(root string, areas []string) (*Report, error) {
	violations, err := Validate(root)
	if err != nil {
		return nil, err
	}
	r := &Report{Problems: []string{}, Warnings: []string{}}
	for _, v := range violations {
		r.Problems = append(r.Problems, fmt.Sprintf("schema: %s: %s", v.Pointer, v.Message))
	}
	p, err := Load(root)
	if err != nil {
		if len(violations) > 0 {
			return r, nil // unparseable; the schema violations say so
		}
		return nil, err
	}
	r.checkIDs(p)
	r.checkCycles(p)
	r.checkReferences(root, p)
	r.checkAreas(p, areas)
	r.checkHistory(p)
	return r, nil
}

func (r *Report) problem(format string, a ...any) {
	r.Problems = append(r.Problems, fmt.Sprintf(format, a...))
}

func (r *Report) checkIDs(p *Plan) {
	seen := map[string]int{}
	for _, t := range p.Tasks {
		seen[t.ID]++
		if seen[t.ID] == 2 {
			r.problem("duplicate task id: %s", t.ID)
		}
	}
	for _, t := range p.Tasks {
		for _, d := range t.DependsOn {
			if seen[d] == 0 {
				r.problem("%s depends on %s, which does not exist", t.ID, d)
			}
		}
	}
}

// idNum is the number after the T of a task id; ids that do not parse sort last.
func idNum(id string) int {
	n, err := strconv.Atoi(strings.TrimPrefix(id, "T"))
	if err != nil {
		return int(^uint(0) >> 1)
	}
	return n
}

func idLess(a, b string) bool {
	if na, nb := idNum(a), idNum(b); na != nb {
		return na < nb
	}
	return a < b
}

// checkCycles reports each depends_on cycle as the loop itself, starting from
// its lowest id.
func (r *Report) checkCycles(p *Plan) {
	deps := map[string][]string{}
	var ids []string
	for _, t := range p.Tasks {
		if _, dup := deps[t.ID]; dup {
			continue // a duplicate id is already reported; its first task speaks for it
		}
		ids = append(ids, t.ID)
		deps[t.ID] = t.DependsOn
	}
	sort.Slice(ids, func(i, j int) bool { return idLess(ids[i], ids[j]) })

	const (
		gray  = 1
		black = 2
	)
	color := map[string]int{}
	var stack []string
	reported := map[string]bool{}
	var visit func(id string)
	visit = func(id string) {
		color[id] = gray
		stack = append(stack, id)
		for _, d := range deps[id] {
			if _, ok := deps[d]; !ok {
				continue
			}
			switch color[d] {
			case 0:
				visit(d)
			case gray:
				i := len(stack) - 1
				for stack[i] != d {
					i--
				}
				cyc := append([]string(nil), stack[i:]...)
				lo := 0
				for j, c := range cyc {
					if idLess(c, cyc[lo]) {
						lo = j
					}
				}
				cyc = append(cyc[lo:], cyc[:lo]...)
				cyc = append(cyc, cyc[0])
				line := strings.Join(cyc, " -> ")
				if !reported[line] {
					reported[line] = true
					r.problem("depends_on cycle: %s", line)
				}
			}
		}
		stack = stack[:len(stack)-1]
		color[id] = black
	}
	for _, id := range ids {
		if color[id] == 0 {
			visit(id)
		}
	}
}

func (r *Report) checkReferences(root string, p *Plan) {
	for _, t := range p.Tasks {
		for _, ref := range t.References {
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(ref.Path))); err != nil {
				r.problem("%s: reference does not resolve: %s", t.ID, ref.Path)
			}
			if strings.TrimSpace(ref.Why) == "" {
				r.problem("%s: reference has no reason: %s", t.ID, ref.Path)
			}
		}
	}
}

func (r *Report) checkAreas(p *Plan, areas []string) {
	if len(areas) == 0 {
		return
	}
	for _, t := range p.Tasks {
		if t.Area == "" {
			r.problem("%s: no area — areas is set", t.ID)
			continue
		}
		found := false
		for _, a := range areas {
			found = found || a == t.Area
		}
		if !found {
			r.problem("%s: area %q is not in areas (%s)", t.ID, t.Area, strings.Join(areas, ", "))
		}
	}
}

func (r *Report) checkHistory(p *Plan) {
	for _, t := range p.Tasks {
		for _, h := range t.GateHistory {
			if strings.TrimSpace(h.Reason) == "" {
				r.Warnings = append(r.Warnings, fmt.Sprintf("%s: no gate_history reason", t.ID))
			}
		}
	}
}
