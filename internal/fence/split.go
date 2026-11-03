// Package fence judges one tool call of a review session against the embedded
// review fence. It is pure: it reads the fence, the adapter's deny-list stems
// and the plugin's skill names, and never runs the command it judges.
package fence

import (
	"errors"
	"strings"
)

// command is one simple command: its words, and the files it redirects to and
// from.
type command struct {
	words  []string
	plain  []bool // per word: written with no quote or escape
	exp    []bool // per word: the shell expands it (C2)
	writes []string
	reads  []string
}

type pending int

const (
	pendNone pending = iota
	pendWrite
	pendRead
	pendDup  // >&N or <&N: a descriptor, unless the target is not one
	pendSkip // a here-document delimiter or here-string
)

type lexer struct {
	cmds   []command
	cur    command
	word   strings.Builder
	st     []byte // per byte of word: stUnquoted, stDouble or stLiteral
	in     bool   // a word is being built (possibly empty, if quoted)
	quoted bool
	pend   pending
	depth  int  // subshells open
	closed bool // a subshell was just closed: only a redirection or an operator may follow
}

const subshellRefusal = "subshells and process substitution cannot be judged"

// opening are the keywords that start a command whose own words follow.
var opening = map[string]bool{
	"if": true, "then": true, "elif": true, "else": true, "while": true, "until": true, "do": true, "!": true, "{": true,
}

// How a byte of a word was written: bare, inside double quotes, or made
// literal by single quotes or a backslash.
const (
	stUnquoted byte = iota
	stDouble
	stLiteral
)

func (l *lexer) put(b, st byte) {
	l.word.WriteByte(b)
	l.st = append(l.st, st)
	l.in = true
}

func (l *lexer) add(s string) {
	for i := 0; i < len(s); i++ {
		l.put(s[i], stLiteral)
	}
	l.in = true
}

func (l *lexer) resetWord() {
	l.word.Reset()
	l.st = l.st[:0]
}

// expands reports whether the shell expands a word (C2): a $ unquoted or in
// double quotes, an unquoted * or ?, an unquoted [ with an unquoted ] later,
// or a brace expansion: an unquoted { and the first unquoted } after it, with
// no unquoted { between and an unquoted , or .. between the two.
func expands(w string, st []byte) bool {
	bare := func(i int) bool { return st[i] == stUnquoted }
	for i := 0; i < len(w); i++ {
		switch {
		case w[i] == '$' && st[i] != stLiteral:
			return true
		case (w[i] == '*' || w[i] == '?') && bare(i):
			return true
		case w[i] == '[' && bare(i):
			for j := i + 1; j < len(w); j++ {
				if w[j] == ']' && bare(j) {
					return true
				}
			}
		case w[i] == '{' && bare(i):
			for j := i + 1; j < len(w); j++ {
				if !bare(j) {
					continue
				}
				if w[j] == '{' {
					break
				}
				if w[j] == '}' {
					for k := i + 1; k < j; k++ {
						if bare(k) && (w[k] == ',' || (w[k] == '.' && k+1 < j && w[k+1] == '.' && bare(k+1))) {
							return true
						}
					}
					break
				}
			}
		}
	}
	return false
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func (l *lexer) flush() error {
	if !l.in {
		return nil
	}
	w := l.word.String()
	ex := expands(w, l.st)
	l.resetWord()
	plain := !l.quoted
	l.in, l.quoted = false, false
	p := l.pend
	l.pend = pendNone
	switch p {
	case pendNone:
		if l.closed {
			return errors.New(subshellRefusal)
		}
		l.cur.words = append(l.cur.words, w)
		l.cur.plain = append(l.cur.plain, plain)
		l.cur.exp = append(l.cur.exp, ex)
		return nil
	case pendSkip:
		return nil
	case pendDup:
		if allDigits(w) || w == "-" {
			return nil
		}
		p = pendWrite
	}
	if strings.ContainsAny(w, "$`") {
		return errors.New("a redirection target holding $ cannot be resolved: " + w)
	}
	if ex {
		return errors.New("a redirection target that expands cannot be resolved: " + w)
	}
	if p == pendWrite {
		l.cur.writes = append(l.cur.writes, w)
	} else {
		l.cur.reads = append(l.cur.reads, w)
	}
	return nil
}

func (l *lexer) end() error {
	if err := l.flush(); err != nil {
		return err
	}
	if l.pend != pendNone {
		return errors.New("a redirection has no target")
	}
	if len(l.cur.words)+len(l.cur.writes)+len(l.cur.reads) > 0 {
		l.cmds = append(l.cmds, l.cur)
	}
	l.cur = command{}
	l.closed = false
	return nil
}

// opens reports whether a ( here opens a subshell: no word is being built,
// no redirection is complete or waiting, and the command holds nothing but
// opening keywords.
func (l *lexer) opens() bool {
	if l.in || l.closed || l.pend != pendNone || len(l.cur.writes)+len(l.cur.reads) > 0 {
		return false
	}
	for _, w := range l.cur.words {
		if !opening[w] {
			return false
		}
	}
	return true
}

// split cuts a command string into simple commands at ; && || | & and
// newlines outside quotes. It refuses what it cannot judge: an unterminated
// quote, command substitution, process substitution and what a ( opens
// anywhere but a subshell. A subshell's ( and ) cut commands like ; does.
func split(s string) ([]command, error) {
	l := &lexer{}
	next := func(i int) byte {
		if i+1 < len(s) {
			return s[i+1]
		}
		return 0
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\'':
			j := strings.IndexByte(s[i+1:], '\'')
			if j < 0 {
				return nil, errors.New("unterminated quote")
			}
			l.add(s[i+1 : i+1+j])
			l.quoted = true
			i += j + 1
		case '"':
			l.in, l.quoted = true, true
			closed := false
			for i++; i < len(s); i++ {
				d := s[i]
				if d == '"' {
					closed = true
					break
				}
				if d == '`' || (d == '$' && next(i) == '(') {
					return nil, errors.New("command substitution cannot be judged")
				}
				if d == '\\' && i+1 < len(s) {
					n := s[i+1]
					if n == '\n' {
						i++
						continue
					}
					if n == '"' || n == '\\' || n == '$' || n == '`' {
						l.put(n, stLiteral)
						i++
						continue
					}
				}
				l.put(d, stDouble)
			}
			if !closed {
				return nil, errors.New("unterminated quote")
			}
		case '\\':
			if i+1 < len(s) {
				i++
				if s[i] != '\n' {
					l.put(s[i], stLiteral)
					l.in, l.quoted = true, true
				}
			}
		case '`':
			return nil, errors.New("command substitution cannot be judged")
		case '$':
			if next(i) == '(' {
				return nil, errors.New("command substitution cannot be judged")
			}
			l.put(c, stUnquoted)
		case '(':
			if (i > 0 && s[i-1] == '(') || !l.opens() {
				return nil, errors.New(subshellRefusal)
			}
			if err := l.end(); err != nil {
				return nil, err
			}
			l.depth++
		case ')':
			if l.depth == 0 {
				return nil, errors.New(subshellRefusal)
			}
			if err := l.end(); err != nil {
				return nil, err
			}
			l.depth--
			l.closed = true
		case '#':
			if l.in {
				l.put(c, stUnquoted)
				break
			}
			for i < len(s) && s[i] != '\n' {
				i++
			}
			i--
		case ' ', '\t', '\r':
			if err := l.flush(); err != nil {
				return nil, err
			}
		case '\n', ';':
			if err := l.end(); err != nil {
				return nil, err
			}
		case '|':
			if n := next(i); n == '|' || n == '&' {
				i++
			}
			if err := l.end(); err != nil {
				return nil, err
			}
		case '&', '<', '>':
			if c == '&' {
				n := next(i)
				if n != '>' {
					if n == '&' {
						i++
					}
					if err := l.end(); err != nil {
						return nil, err
					}
					break
				}
			}
			if l.in && !l.quoted && allDigits(l.word.String()) {
				l.resetWord()
				l.in = false
			} else if err := l.flush(); err != nil {
				return nil, err
			}
			if l.pend != pendNone {
				return nil, errors.New("a redirection has no target")
			}
			kind := pendWrite
			switch c {
			case '&': // &> and &>>
				i++
				if next(i) == '>' {
					i++
				}
			case '>':
				switch next(i) {
				case '>', '|':
					i++
				case '&':
					i++
					kind = pendDup
				}
			case '<':
				switch next(i) {
				case '<':
					i++
					kind = pendSkip
					if n := next(i); n == '<' || n == '-' {
						i++
					}
				case '&':
					i++
					kind = pendDup
				case '>':
					i++
				default:
					kind = pendRead
				}
			}
			l.pend = kind
		default:
			l.put(c, stUnquoted)
		}
	}
	if err := l.end(); err != nil {
		return nil, err
	}
	if l.depth > 0 {
		return nil, errors.New(subshellRefusal)
	}
	return l.cmds, nil
}
