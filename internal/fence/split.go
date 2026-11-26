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
	words      []string
	plain      []bool // per word: written with no quote or escape
	exp        []bool // per word: the shell expands it (C2)
	writes     []string
	reads      []string
	powerShell string
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
	syntax splitSyntax
	in     bool // a word is being built (possibly empty, if quoted)
	quoted bool
	pend   pending
	depth  int  // subshells open
	closed bool // a subshell was just closed: only a redirection or an operator may follow
	pipe   bool // the current command is after a pipeline operator
}

type splitSyntax int

const (
	syntaxPOSIX splitSyntax = iota
	syntaxPowerShell
	syntaxCMD
)

const subshellRefusal = "subshells and process substitution cannot be judged"

const scriptBlockRefusal = "script blocks cannot be judged — use Where-Object <Property> -like <value>, ForEach-Object -MemberName <Property> or Select-Object -ExpandProperty <Property>"

const powerShellParenthesesRefusal = "parentheses cannot be judged here — (<command>).<Property> is read; or pipe into Measure-Object or Select-Object -ExpandProperty"

const powerShellSubstitutionRefusal = "command substitution cannot be judged — run the inner command on its own"

const powerShellBacktickRefusal = "a backtick cannot be judged — write the command on one line, without escapes"

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
	if l.syntax == syntaxCMD {
		ex = false
	}
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
	if l.syntax != syntaxCMD && strings.ContainsAny(w, "$`") {
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

func (l *lexer) startsPowerShellRead() bool {
	return !l.in && !l.closed && !l.pipe && l.pend == pendNone &&
		len(l.cur.words) == 0 && len(l.cur.writes) == 0 && len(l.cur.reads) == 0
}

// split cuts a command string into simple commands at ; && || | & and
// newlines outside quotes. It refuses what it cannot judge: an unterminated
// quote, command substitution, process substitution and what a ( opens
// anywhere but a subshell. A subshell's ( and ) cut commands like ; does.
func split(s string) ([]command, error) {
	return splitMode(s, syntaxPOSIX)
}

func splitPowerShell(s string) ([]command, error) {
	return splitMode(s, syntaxPowerShell)
}

func splitCMDPart(s string) ([]command, error) {
	return splitMode(s, syntaxCMD)
}

func splitMode(s string, syntax splitSyntax) (cmds []command, err error) {
	l := &lexer{syntax: syntax}
	defer func() {
		if err != nil && syntax == syntaxPowerShell && cmds == nil {
			cmds = l.cmds
		}
	}()
	next := func(i int) byte {
		if i+1 < len(s) {
			return s[i+1]
		}
		return 0
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if syntax == syntaxPowerShell && c == '$' && !l.pipe && l.startsPowerShellRead() {
			end, first, ok, assignment := powerShellEnvAssignment(s, i)
			if assignment {
				if !ok {
					l.cmds = append(l.cmds, command{
						words: []string{first},
						plain: []bool{true},
						exp:   []bool{false},
					})
				}
				i = end - 1
				continue
			}
		}
		switch c {
		case '\'':
			if syntax == syntaxCMD {
				l.put(c, stUnquoted)
				break
			}
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
				if d == '`' && syntax != syntaxCMD {
					reason := "command substitution cannot be judged"
					if syntax == syntaxPowerShell {
						reason = powerShellBacktickRefusal
					}
					return nil, errors.New(reason)
				}
				if syntax != syntaxCMD && d == '$' && next(i) == '(' {
					reason := "command substitution cannot be judged"
					if syntax == syntaxPowerShell {
						reason = powerShellSubstitutionRefusal
					}
					return nil, errors.New(reason)
				}
				if syntax != syntaxCMD && d == '\\' && i+1 < len(s) {
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
			word := l.word.String()
			windowsPath := len(word) >= 2 && word[1] == ':' && ((word[0] >= 'A' && word[0] <= 'Z') || (word[0] >= 'a' && word[0] <= 'z'))
			if syntax != syntaxPOSIX || windowsPath {
				l.put(c, stUnquoted)
			} else if i+1 < len(s) {
				i++
				if s[i] != '\n' {
					l.put(s[i], stLiteral)
					l.in, l.quoted = true, true
				}
			}
		case '{':
			if syntax == syntaxPowerShell {
				return nil, errors.New(scriptBlockRefusal)
			}
			l.put(c, stUnquoted)
		case '`':
			if syntax == syntaxCMD {
				l.put(c, stUnquoted)
			} else if syntax == syntaxPowerShell {
				return nil, errors.New(powerShellBacktickRefusal)
			} else {
				return nil, errors.New("command substitution cannot be judged")
			}
		case '$':
			if syntax != syntaxCMD && next(i) == '(' {
				if syntax == syntaxPowerShell {
					return nil, errors.New(powerShellSubstitutionRefusal)
				}
				return nil, errors.New("command substitution cannot be judged")
			}
			l.put(c, stUnquoted)
		case '(':
			if syntax == syntaxCMD {
				if err := l.end(); err != nil {
					return nil, err
				}
				l.depth++
				break
			}
			if syntax == syntaxPowerShell {
				if !l.startsPowerShellRead() {
					return nil, errors.New(powerShellParenthesesRefusal)
				}
				inner, end, ok := powerShellProperty(s, i)
				if ok {
					if _, nestedErr := splitPowerShell(inner); nestedErr != nil && nestedErr.Error() == scriptBlockRefusal {
						return nil, nestedErr
					}
					l.cmds = append(l.cmds, command{powerShell: inner})
					i = end
					break
				}
				return nil, errors.New(powerShellParenthesesRefusal)
			}
			if (i > 0 && s[i-1] == '(') || !l.opens() {
				return nil, errors.New(subshellRefusal)
			}
			if err := l.end(); err != nil {
				return nil, err
			}
			l.depth++
		case ')':
			if syntax == syntaxCMD {
				if l.depth == 0 {
					return nil, errors.New("cmd command group cannot be judged")
				}
				if err := l.end(); err != nil {
					return nil, err
				}
				l.depth--
				l.closed = true
				break
			}
			if l.depth == 0 {
				if syntax == syntaxPowerShell {
					return nil, errors.New(powerShellParenthesesRefusal)
				}
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
		case '\n':
			if syntax == syntaxCMD {
				if err := l.flush(); err != nil {
					return nil, err
				}
				break
			}
			if err := l.end(); err != nil {
				return nil, err
			}
			l.pipe = false
		case ';':
			if syntax == syntaxCMD {
				l.put(c, stUnquoted)
				break
			}
			if err := l.end(); err != nil {
				return nil, err
			}
			l.pipe = false
		case '|':
			if n := next(i); n == '|' || n == '&' {
				i++
			}
			if err := l.end(); err != nil {
				return nil, err
			}
			l.pipe = true
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
					l.pipe = false
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
		if syntax == syntaxCMD {
			return nil, errors.New("cmd command group cannot be judged")
		}
		return nil, errors.New(subshellRefusal)
	}
	return l.cmds, nil
}

func powerShellProperty(s string, start int) (inner string, end int, ok bool) {
	depth := 0
	quote := byte(0)
	for i := start; i < len(s); i++ {
		c := s[i]
		if quote != 0 {
			if c == quote {
				if quote == '\'' && i+1 < len(s) && s[i+1] == '\'' {
					i++
					continue
				}
				quote = 0
			}
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			continue
		}
		switch c {
		case '(':
			depth++
		case ')':
			depth--
			if depth != 0 {
				continue
			}
			if strings.TrimSpace(s[start+1:i]) == "" || i+1 >= len(s) || s[i+1] != '.' {
				return "", 0, false
			}
			j := i + 2
			components := 0
			for {
				componentStart := j
				for j < len(s) && ((s[j] >= 'A' && s[j] <= 'Z') ||
					(s[j] >= 'a' && s[j] <= 'z') ||
					(s[j] >= '0' && s[j] <= '9') || s[j] == '_') {
					j++
				}
				if j == componentStart {
					return "", 0, false
				}
				components++
				if j >= len(s) || s[j] != '.' {
					break
				}
				j++
			}
			if components == 0 {
				return "", 0, false
			}
			for j < len(s) && (s[j] == ' ' || s[j] == '\t' || s[j] == '\r') {
				j++
			}
			if j < len(s) {
				switch s[j] {
				case '|', ';', '\n', '&':
				default:
					return "", 0, false
				}
			}
			return s[start+1 : i], j - 1, true
		}
	}
	return "", 0, false
}

func powerShellEnvAssignment(s string, start int) (end int, first string, ok, assignment bool) {
	const prefix = "$env:"
	if len(s)-start < len(prefix) || !strings.EqualFold(s[start:start+len(prefix)], prefix) {
		return start + 1, "", false, false
	}
	assignment = true
	end = start
	quote := byte(0)
	for end < len(s) {
		c := s[end]
		if quote != 0 {
			if c == quote {
				if quote == '\'' && end+1 < len(s) && s[end+1] == '\'' {
					end += 2
					continue
				}
				quote = 0
			}
			end++
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			end++
			continue
		}
		if c == ';' || c == '\n' || c == '|' {
			break
		}
		if c == '&' {
			break
		}
		end++
	}
	statement := strings.TrimSpace(s[start:end])
	first = powerShellFirstWord(statement)
	if len(prefix) >= len(statement) {
		return end, first, false, true
	}
	i := len(prefix)
	nameStart := i
	for i < len(statement) && powerShellNameChar(statement[i]) {
		i++
	}
	if i == nameStart || statement[nameStart] >= '0' && statement[nameStart] <= '9' {
		return end, first, false, true
	}
	for i < len(statement) && (statement[i] == ' ' || statement[i] == '\t' || statement[i] == '\r') {
		i++
	}
	if i >= len(statement) || statement[i] != '=' {
		return end, first, false, true
	}
	i++
	for i < len(statement) && (statement[i] == ' ' || statement[i] == '\t' || statement[i] == '\r') {
		i++
	}
	if end < len(s) && s[end] == '|' {
		return end, first, false, true
	}
	return end, first, powerShellLiteral(statement[i:]), true
}

func powerShellFirstWord(s string) string {
	quote := byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if quote != 0 {
			if c == quote {
				if quote == '\'' && i+1 < len(s) && s[i+1] == '\'' {
					i++
					continue
				}
				quote = 0
			}
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			continue
		}
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == ';' || c == '|' || c == '&' {
			return s[:i]
		}
	}
	return s
}

func powerShellNameChar(c byte) bool {
	return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' ||
		c >= '0' && c <= '9' || c == '_'
}

func powerShellLiteral(s string) bool {
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		for i := 1; i < len(s)-1; i++ {
			if s[i] == '\'' {
				if i+1 >= len(s)-1 || s[i+1] != '\'' {
					return false
				}
				i++
			}
		}
		return true
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return !strings.ContainsAny(s[1:len(s)-1], "$`")
	}
	if s == "" || strings.IndexAny(s, " \t\r\n$(){}[]@`") >= 0 {
		return false
	}
	return true
}
