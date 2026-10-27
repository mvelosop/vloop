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
	in     bool // a word is being built (possibly empty, if quoted)
	quoted bool
	pend   pending
}

func (l *lexer) add(s string) {
	l.word.WriteString(s)
	l.in = true
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
	l.word.Reset()
	l.in, l.quoted = false, false
	p := l.pend
	l.pend = pendNone
	switch p {
	case pendNone:
		l.cur.words = append(l.cur.words, w)
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
	return nil
}

// split cuts a command string into simple commands at ; && || | & and
// newlines outside quotes. It refuses what it cannot judge: an unterminated
// quote, command substitution, process substitution and grouping.
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
						l.word.WriteByte(n)
						i++
						continue
					}
				}
				l.word.WriteByte(d)
			}
			if !closed {
				return nil, errors.New("unterminated quote")
			}
		case '\\':
			if i+1 < len(s) {
				i++
				if s[i] != '\n' {
					l.word.WriteByte(s[i])
					l.in, l.quoted = true, true
				}
			}
		case '`':
			return nil, errors.New("command substitution cannot be judged")
		case '$':
			if next(i) == '(' {
				return nil, errors.New("command substitution cannot be judged")
			}
			l.word.WriteByte(c)
			l.in = true
		case '(', ')':
			return nil, errors.New("subshells and process substitution cannot be judged")
		case '#':
			if l.in {
				l.word.WriteByte(c)
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
				l.word.Reset()
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
			l.word.WriteByte(c)
			l.in = true
		}
	}
	if err := l.end(); err != nil {
		return nil, err
	}
	return l.cmds, nil
}
