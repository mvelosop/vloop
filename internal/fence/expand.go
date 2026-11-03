package fence

import (
	"fmt"
	"strings"
)

// xargsInput stands for the words xargs appends to its command.
const xargsInput = "<xargs input>"

// xargsCommand is the command xargs runs, as the judge sees it: with a
// replace string every word holding it is expanded and nothing is appended;
// without, one more expanded word (the input) ends the command. No command
// is echo.
func xargsCommand(ws []string, plain, exp []bool, repl string) ([]string, []bool, []bool) {
	ws, plain, exp = append([]string(nil), ws...), append([]bool(nil), plain...), append([]bool(nil), exp...)
	if len(ws) == 0 {
		ws, plain, exp = []string{"echo"}, []bool{true}, []bool{false}
	}
	if repl != "" {
		for i, w := range ws {
			if strings.Contains(w, repl) {
				exp[i] = true
			}
		}
		return ws, plain, exp
	}
	return append(ws, xargsInput), append(plain, false), append(exp, true)
}

// match denies a simple command that a deny rule of the review fence or one of
// the adapter's stems covers: as written first, then with each expanded word
// standing for any text.
func (j *Judge) match(ws []string, exp []bool) string {
	var shown []string
	anyExp := false
	for i, w := range ws {
		if w != xargsInput || !exp[i] {
			shown = append(shown, w)
		}
		anyExp = anyExp || exp[i]
	}
	text := strings.Join(shown, " ")
	for _, p := range j.fence {
		if patternMatch(p, text) {
			return fmt.Sprintf("%s is denied by the %s fence (Bash(%s))", text, j.fenceLabel(), p)
		}
	}
	for _, p := range j.stems {
		if patternMatch(p, text) {
			return fmt.Sprintf("%s is denied by the Copilot deny list (%s)", text, p)
		}
	}
	if !anyExp {
		return ""
	}
	const suffix = ": a word the shell expands may be any text"
	for _, p := range j.fence {
		if patternMeets(p, ws, exp) {
			return fmt.Sprintf("%s is denied by the %s fence (Bash(%s))%s", text, j.fenceLabel(), p, suffix)
		}
	}
	for _, p := range j.stems {
		if patternMeets(p, ws, exp) {
			return fmt.Sprintf("%s is denied by the Copilot deny list (%s)%s", text, p, suffix)
		}
	}
	return ""
}

// skipXargs skips xargs' options and reports its replace string ("" when it
// has none).
func skipXargs(rest []string) ([]string, string) {
	repl := ""
	for len(rest) > 0 && isOption(rest[0]) {
		o := rest[0]
		rest = rest[1:]
		if o == "--" {
			break
		}
		if strings.HasPrefix(o, "--") {
			name, val, _ := strings.Cut(o, "=")
			switch {
			case name == "--replace":
				repl = val
				if repl == "" {
					repl = "{}"
				}
			case xargsLongArgs[o] && len(rest) > 0:
				rest = rest[1:]
			}
			continue
		}
		for k := 1; k < len(o); k++ {
			if strings.IndexByte("adEILnPsJRS", o[k]) >= 0 {
				replace := o[k] == 'I' || o[k] == 'J'
				if k == len(o)-1 {
					if len(rest) > 0 {
						if replace {
							repl = rest[0]
						}
						rest = rest[1:]
					}
				} else if replace {
					repl = o[k+1:]
				}
				break
			}
			if strings.IndexByte("eil", o[k]) >= 0 {
				if o[k] == 'i' {
					repl = o[k+1:]
					if repl == "" {
						repl = "{}"
					}
				}
				break
			}
		}
	}
	return rest, repl
}

// shellString finds the -c string of a shell invocation's arguments. exp says
// which words the shell expands; repls are the replace strings of the xargs
// looked through.
func shellString(rest []string, exp []bool, repls []string) (string, string) {
	const before = " expands before the command string and cannot be judged"
	for i := 0; i < len(rest); i++ {
		o := rest[i]
		isOpt := isOption(o) || (len(o) > 1 && o[0] == '+')
		if isOpt && exp[i] {
			return "", o + before
		}
		switch {
		case o == "-o" || o == "+o":
			i++
			if i < len(rest) && exp[i] {
				return "", rest[i] + before
			}
		case o != "" && o[0] == '-' && !strings.HasPrefix(o, "--") && strings.Contains(o, "c"):
			if i+1 >= len(rest) {
				return "", "-c has no command string"
			}
			if strings.Contains(rest[i+1], "$") {
				return "", "the -c string holds $ and cannot be judged"
			}
			for _, r := range repls {
				if strings.Contains(rest[i+1], r) {
					return "", "the -c string holds xargs' replace string " + r + " and cannot be judged"
				}
			}
			if exp[i+1] {
				return "", "the -c string expands and cannot be judged"
			}
			return rest[i+1], ""
		case isOpt:
		default:
			return "", "runs a script or stdin, not a literal -c string"
		}
	}
	return "", "runs a script or stdin, not a literal -c string"
}

// patternMeets reports whether some text in place of the expanded words makes
// the rule p match the command: an expanded word is none, one or several
// words. Both are literal text and *, so it asks whether the two languages
// share a string.
func patternMeets(p string, ws []string, exp []bool) bool {
	var a, b nfa
	a.start, b.start = a.node(), b.node()
	end := b.start
	if pre, ok := strings.CutSuffix(p, ":*"); ok {
		end = b.lit(end, pre)
		end = b.opt(end, func(n int) int { return b.star(b.lit(n, " ")) })
	} else {
		for i, q := range strings.Split(p, "*") {
			if i > 0 {
				end = b.star(end)
			}
			end = b.lit(end, q)
		}
	}
	b.accept = end
	end = a.start
	for i, w := range ws {
		switch {
		case i == 0:
			end = a.lit(end, w)
		case exp[i]:
			end = a.opt(end, func(n int) int { return a.star(a.lit(n, " ")) })
		default:
			end = a.lit(a.lit(end, " "), w)
		}
	}
	a.accept = end
	return a.meets(&b)
}

type edge struct {
	to   int
	kind byte // 'e' empty, 'c' the byte c, 'a' any byte
	c    byte
}

type nfa struct {
	edges         [][]edge
	start, accept int
	alpha         map[byte]bool
}

func (n *nfa) node() int {
	n.edges = append(n.edges, nil)
	return len(n.edges) - 1
}

func (n *nfa) lit(from int, s string) int {
	if n.alpha == nil {
		n.alpha = map[byte]bool{}
	}
	for i := 0; i < len(s); i++ {
		to := n.node()
		n.edges[from] = append(n.edges[from], edge{to, 'c', s[i]})
		n.alpha[s[i]] = true
		from = to
	}
	return from
}

func (n *nfa) star(from int) int {
	s := n.node()
	n.edges[from] = append(n.edges[from], edge{s, 'e', 0})
	n.edges[s] = append(n.edges[s], edge{s, 'a', 0})
	return s
}

func (n *nfa) opt(from int, f func(int) int) int {
	out := n.node()
	n.edges[from] = append(n.edges[from], edge{out, 'e', 0})
	end := f(from)
	n.edges[end] = append(n.edges[end], edge{out, 'e', 0})
	return out
}

// meets reports whether n and m accept a common string.
func (n *nfa) meets(m *nfa) bool {
	alpha := map[byte]bool{0: true}
	for c := range n.alpha {
		alpha[c] = true
	}
	for c := range m.alpha {
		alpha[c] = true
	}
	type pair struct{ a, b int }
	seen := map[pair]bool{{n.start, m.start}: true}
	queue := []pair{{n.start, m.start}}
	push := func(p pair) {
		if !seen[p] {
			seen[p] = true
			queue = append(queue, p)
		}
	}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		if p.a == n.accept && p.b == m.accept {
			return true
		}
		for _, e := range n.edges[p.a] {
			if e.kind == 'e' {
				push(pair{e.to, p.b})
			}
		}
		for _, e := range m.edges[p.b] {
			if e.kind == 'e' {
				push(pair{p.a, e.to})
			}
		}
		for c := range alpha {
			for _, e := range n.edges[p.a] {
				if e.kind == 'e' || (e.kind == 'c' && e.c != c) {
					continue
				}
				for _, f := range m.edges[p.b] {
					if f.kind == 'e' || (f.kind == 'c' && f.c != c) {
						continue
					}
					push(pair{e.to, f.to})
				}
			}
		}
	}
	return false
}
