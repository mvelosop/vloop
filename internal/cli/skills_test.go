package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	vloop "github.com/mvelosop/vloop"
	"github.com/mvelosop/vloop/internal/schema"
)

// The structural checks over the plugin's skills and evals. They run over the
// directory named by SKILLCHECK_PLUGIN_DIR, or ../../plugin when it is unset or
// empty, so a gate can plant a defect in a copy and see it rejected. They never
// run claude and read nothing of this repository's .loop/ or .vloop/state/.

const (
	skillCheckPluginEnv = "SKILLCHECK_PLUGIN_DIR"
	skillCheckDefault   = "../../plugin"
	skillCheckDomainDoc = "../../docs/domain/domain-model.md"
	skillCheckFixtures  = "testdata/skillcheck"
)

// fencedSkills are the skills whose sessions run under the fence; the operate
// skill is the operator's, in an interactive session, and is exempt.
var fencedSkills = map[string]bool{"plan": true, "work": true, "review": true}

func TestPluginSkills(t *testing.T) {
	dir := os.Getenv(skillCheckPluginEnv)
	if dir == "" {
		dir = skillCheckDefault
	}
	for _, v := range checkSkillPlugin(dir) {
		t.Errorf("%s", v)
	}
}

// fixtureWants names, per committed fixture plugin, what its rejection must say.
// A fixture with no entries is well-formed and must be accepted.
var fixtureWants = map[string][]string{
	"good":                         nil,
	"bad-frontmatter-name":         {"skill review", `"reviewer"`},
	"bad-frontmatter-field":        {"skill plan", "description"},
	"bad-command":                  {"skill work", "frobnicate"},
	"bad-flag":                     {"skill work", "--frobnicate"},
	"bad-schema-name":              {"skill work", "frobnicate/v1"},
	"bad-schema-example":           {"skill work", "verified"},
	"bad-path":                     {"skill work", ".vloop/scratch/notes.md"},
	"bad-fence-denied":             {"skill work", "vloop task verify", "denied"},
	"bad-fence-not-allowed":        {"skill work", "vloop config list", "not allowed"},
	"bad-eval-no-graders":          {"eval case-b", "graders"},
	"bad-eval-max-turns":           {"eval case-c", "max_turns"},
	"bad-eval-weight":              {"eval case-d", "weight"},
	"bad-eval-unknown-key":         {"eval case-a", "setup", "not one claude plugin eval knows"},
	"bad-eval-scaffold-undeclared": {"eval case-a", "scaffold_script"},
	"bad-eval-grader-type":         {"eval case-a", "command", "grader type"},
	"bad-eval-tools-string":        {"eval case-a", "allowed_tools", "yaml array"},
	"bad-eval-regex-keys":          {"eval case-a", "regex grader", "match", "target"},
	"operate-exempt":               nil,
	"bad-operate-command":          {"skill operate", "frobnicate"},
}

func TestSkillChecksRejectFixtures(t *testing.T) {
	ents, err := os.ReadDir(skillCheckFixtures)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		seen[name] = true
		wants, ok := fixtureWants[name]
		if !ok {
			t.Errorf("fixture %s has no expectation in fixtureWants", name)
			continue
		}
		t.Run(name, func(t *testing.T) {
			got := checkSkillPlugin(filepath.Join(skillCheckFixtures, name))
			if wants == nil {
				if len(got) > 0 {
					t.Fatalf("well-formed fixture rejected:\n%s", strings.Join(got, "\n"))
				}
				return
			}
			if len(got) == 0 {
				t.Fatalf("fixture accepted, want a violation naming %q", wants)
			}
			all := strings.ToLower(strings.Join(got, "\n"))
			for _, w := range wants {
				if !strings.Contains(all, strings.ToLower(w)) {
					t.Errorf("violations do not name %q:\n%s", w, strings.Join(got, "\n"))
				}
			}
		})
	}
	for name := range fixtureWants {
		if !seen[name] {
			t.Errorf("fixtureWants names %s, which is not under %s", name, skillCheckFixtures)
		}
	}
}

// checkSkillPlugin runs every structural check over the plugin at dir.
func checkSkillPlugin(dir string) []string {
	var out []string
	root, _ := NewRoot(Build{})
	layout, err := onDiskLayout(skillCheckDomainDoc)
	if err != nil {
		out = append(out, "On disk layout: "+err.Error())
	}
	fence, err := loadFence()
	if err != nil {
		out = append(out, "embedded fence: "+err.Error())
	}
	skills, _ := filepath.Glob(filepath.Join(dir, "skills", "*", "SKILL.md"))
	sort.Strings(skills)
	// A skill folder without a SKILL.md is a frontmatter failure of its own.
	dirs, _ := filepath.Glob(filepath.Join(dir, "skills", "*"))
	for _, d := range dirs {
		if st, err := os.Stat(d); err == nil && st.IsDir() {
			if _, err := os.Stat(filepath.Join(d, "SKILL.md")); err != nil {
				out = append(out, fmt.Sprintf("skill %s: no SKILL.md", filepath.Base(d)))
			}
		}
	}
	for _, f := range skills {
		folder := filepath.Base(filepath.Dir(f))
		raw, err := os.ReadFile(f)
		if err != nil {
			out = append(out, fmt.Sprintf("skill %s: %v", folder, err))
			continue
		}
		text := string(raw)
		out = append(out, checkFrontmatter(folder, text)...)
		cmds := skillCommands(text)
		for _, c := range cmds {
			problems, resolved := checkCommand(root, c)
			for _, p := range problems {
				out = append(out, fmt.Sprintf("skill %s: `%s`: %s", folder, c, p))
			}
			if resolved && fencedSkills[folder] && fence != nil {
				if p := fence.judge(c); p != "" {
					out = append(out, fmt.Sprintf("skill %s: `%s`: %s", folder, c, p))
				}
			}
		}
		out = append(out, checkSchemas(folder, text)...)
		if layout != nil {
			out = append(out, checkPaths(folder, text, layout)...)
		}
	}
	out = append(out, checkEvals(dir)...)
	return out
}

// frontmatter returns the key: value pairs between the leading --- lines.
func frontmatter(text string) map[string]string {
	fm := map[string]string{}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return fm
	}
	for _, l := range lines[1:] {
		if strings.TrimSpace(l) == "---" {
			break
		}
		k, v, ok := strings.Cut(l, ":")
		if !ok || strings.HasPrefix(l, " ") || strings.HasPrefix(l, "\t") {
			continue
		}
		fm[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return fm
}

func checkFrontmatter(folder, text string) []string {
	var out []string
	fm := frontmatter(text)
	for _, k := range []string{"name", "description"} {
		if fm[k] == "" {
			out = append(out, fmt.Sprintf("skill %s: frontmatter is missing %s", folder, k))
		}
	}
	if n := fm["name"]; n != "" && n != folder {
		out = append(out, fmt.Sprintf("skill %s: frontmatter name %q does not equal its folder %q", folder, n, folder))
	}
	return out
}

// block is one fenced code block of a markdown document.
type block struct {
	lang string
	body string
}

// splitMarkdown separates fenced blocks from the prose around them.
func splitMarkdown(text string) (blocks []block, prose string) {
	var p strings.Builder
	var cur *block
	var body []string
	for _, l := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			if cur == nil {
				cur = &block{lang: strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(l), "```"))}
				body = nil
			} else {
				cur.body = strings.Join(body, "\n")
				blocks = append(blocks, *cur)
				cur = nil
			}
			continue
		}
		if cur != nil {
			body = append(body, l)
		} else {
			p.WriteString(l + "\n")
		}
	}
	if cur != nil {
		cur.body = strings.Join(body, "\n")
		blocks = append(blocks, *cur)
	}
	return blocks, p.String()
}

var codeSpan = regexp.MustCompile("`([^`\n]+)`")

// codeSnippets are the code spans of the prose and the lines of every fenced
// block that is not JSON: the places a skill tells a session to run something.
func codeSnippets(text string) []string {
	blocks, prose := splitMarkdown(text)
	var out []string
	for _, m := range codeSpan.FindAllStringSubmatch(prose, -1) {
		out = append(out, m[1])
	}
	for _, b := range blocks {
		if b.lang == "json" {
			continue
		}
		out = append(out, strings.Split(b.body, "\n")...)
	}
	return out
}

var vloopWord = regexp.MustCompile(`(^|[\s(|;&$"'])vloop(\s|$)`)

// skillCommands lists the vloop command lines a skill names: the words after
// each `vloop` in a code snippet, up to a shell operator or comment.
func skillCommands(text string) []string {
	var out []string
	for _, s := range codeSnippets(text) {
		for _, loc := range vloopWord.FindAllStringIndex(s, -1) {
			after := s[loc[0]:]
			toks := cmdTokens(after[strings.Index(after, "vloop")+len("vloop"):])
			if len(toks) == 0 {
				continue
			}
			out = append(out, "vloop "+strings.Join(toks, " "))
		}
	}
	return out
}

var stopTokens = map[string]bool{"|": true, "||": true, "&&": true, ";": true, ">": true, ">>": true, "2>": true, "<<": true}

// cmdTokens splits a command line into words, keeping quoted strings and
// <placeholders> whole, and stops at a shell operator or a comment.
func cmdTokens(s string) []string {
	var toks []string
	i := 0
	for i < len(s) {
		for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
			i++
		}
		if i >= len(s) {
			break
		}
		start := i
		switch s[i] {
		case '"', '\'':
			q := s[i]
			i++
			for i < len(s) && s[i] != q {
				i++
			}
			if i < len(s) {
				i++
			}
			toks = append(toks, s[start:i])
			continue
		case '<':
			if j := strings.IndexByte(s[i:], '>'); j > 0 && !strings.ContainsAny(s[i:i+j], "|;&") {
				i += j + 1
				for i < len(s) && s[i] != ' ' && s[i] != '\t' {
					i++
				}
				toks = append(toks, s[start:i])
				continue
			}
		}
		for i < len(s) && s[i] != ' ' && s[i] != '\t' {
			i++
		}
		tok := s[start:i]
		if stopTokens[tok] || strings.HasPrefix(tok, "#") || strings.HasPrefix(tok, ">") {
			break
		}
		end := strings.TrimRight(tok, ";)`,.:")
		if end != tok && strings.HasSuffix(tok, ";") {
			if end != "" {
				toks = append(toks, end)
			}
			break
		}
		if end == "" {
			break
		}
		toks = append(toks, end)
	}
	return toks
}

func isPlaceholder(tok string) bool {
	return strings.HasPrefix(tok, "<") || strings.HasPrefix(tok, "[") || strings.HasPrefix(tok, "...") ||
		strings.HasSuffix(tok, ">") || strings.HasPrefix(tok, "$")
}

var flagName = regexp.MustCompile(`^--[A-Za-z][A-Za-z0-9-]*`)

// checkCommand resolves a "vloop <words> [--flags]" line against the command
// tree. resolved reports whether the words named a real command.
func checkCommand(root *cobra.Command, line string) (problems []string, resolved bool) {
	toks := strings.Fields(line)[1:]
	cmd := root
	i := 0
	for i < len(toks) && !strings.HasPrefix(toks[i], "-") {
		tok := toks[i]
		if isPlaceholder(tok) {
			break
		}
		var next *cobra.Command
		for _, c := range cmd.Commands() {
			if c.Name() == tok || c.HasAlias(tok) {
				next = c
				break
			}
		}
		if next == nil {
			if cmd == root && (tok == "help" || tok == "completion") {
				return nil, true
			}
			if cmd.HasSubCommands() && (cmd.Args == nil || cmd.Args(cmd, []string{tok}) != nil) {
				problems = append(problems, fmt.Sprintf("unknown command %q in %q", tok, cmd.CommandPath()))
				return problems, false
			}
			break
		}
		cmd = next
		i++
	}
	if cmd == root && len(toks) > 0 && !strings.HasPrefix(toks[0], "-") && !isPlaceholder(toks[0]) {
		problems = append(problems, fmt.Sprintf("unknown command %q", toks[0]))
		return problems, false
	}
	for _, tok := range toks[i:] {
		if strings.HasPrefix(tok, "\"") || strings.HasPrefix(tok, "'") {
			continue
		}
		tok = strings.Trim(tok, "[]")
		switch {
		case strings.HasPrefix(tok, "--"):
			m := flagName.FindString(tok)
			if m == "" {
				continue
			}
			if m == "--help" || cmdHasFlag(cmd, strings.TrimPrefix(m, "--"), "") {
				continue
			}
			problems = append(problems, fmt.Sprintf("unknown flag %s for %q", m, cmd.CommandPath()))
		case len(tok) == 2 && tok[0] == '-' && tok[1] != '-' && tok[1] != 'h':
			if !cmdHasFlag(cmd, "", tok[1:]) {
				problems = append(problems, fmt.Sprintf("unknown flag %s for %q", tok, cmd.CommandPath()))
			}
		}
	}
	return problems, true
}

func cmdHasFlag(cmd *cobra.Command, long, short string) bool {
	local, inherited := cmd.LocalFlags(), cmd.InheritedFlags()
	if long != "" {
		return local.Lookup(long) != nil || inherited.Lookup(long) != nil
	}
	return local.ShorthandLookup(short) != nil || inherited.ShorthandLookup(short) != nil
}

var schemaName = regexp.MustCompile(`(^|[^A-Za-z0-9_/.-])([a-z][a-z0-9-]*/v[0-9]+)\b`)

func checkSchemas(folder, text string) []string {
	var out []string
	known := map[string]bool{}
	for _, n := range schema.Names() {
		known[n] = true
	}
	for _, m := range schemaName.FindAllStringSubmatch(text, -1) {
		if !known[m[2]] {
			out = append(out, fmt.Sprintf("skill %s: names schema %s, which is not in vloop schema list", folder, m[2]))
		}
	}
	blocks, _ := splitMarkdown(text)
	for _, b := range blocks {
		if b.lang != "json" || !strings.Contains(b.body, `"schema"`) {
			continue
		}
		var doc map[string]any
		if err := json.Unmarshal([]byte(b.body), &doc); err != nil {
			out = append(out, fmt.Sprintf("skill %s: a json block carrying a schema field is not valid JSON: %v", folder, err))
			continue
		}
		name, ok := doc["schema"].(string)
		if !ok {
			continue
		}
		vs, err := schema.Validate(name, []byte(b.body))
		if err != nil {
			out = append(out, fmt.Sprintf("skill %s: example of %s: %v", folder, name, err))
			continue
		}
		for _, v := range vs {
			out = append(out, fmt.Sprintf("skill %s: example of %s fails at %s: %s", folder, name, v.Pointer, v.Message))
		}
	}
	return out
}

var (
	placeholderRe = regexp.MustCompile(`<[^>\n]*>`)
	vloopPathRe   = regexp.MustCompile("\\.vloop/(?:<[^>`\\n]*>|[A-Za-z0-9_.*-]|/)*")
)

// onDiskLayout reads the .vloop/ entries of the domain model's "On disk" block.
func onDiskLayout(doc string) ([]string, error) {
	raw, err := os.ReadFile(doc)
	if err != nil {
		return nil, err
	}
	in, fenced := false, false
	var out []string
	for _, l := range strings.Split(string(raw), "\n") {
		switch {
		case strings.HasPrefix(l, "## "):
			if in {
				return out, nil
			}
			in = strings.TrimSpace(strings.TrimPrefix(l, "## ")) == "On disk"
		case in && strings.HasPrefix(l, "```"):
			if fenced {
				return out, nil
			}
			fenced = true
		case in && fenced:
			l = placeholderRe.ReplaceAllString(l, "<>")
			if f := strings.Fields(l); len(f) > 0 && strings.HasPrefix(f[0], ".vloop/") {
				out = append(out, f[0])
			}
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no .vloop/ entries found in the On disk block of %s", doc)
	}
	return out, nil
}

func checkPaths(folder, text string, layout []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, p := range vloopPathRe.FindAllString(text, -1) {
		p = strings.TrimRight(p, ".,:;")
		if seen[p] {
			continue
		}
		seen[p] = true
		if !pathInLayout(placeholderRe.ReplaceAllString(p, "<>"), layout) {
			out = append(out, fmt.Sprintf("skill %s: names path %s, which is not in the On disk layout", folder, p))
		}
	}
	return out
}

// pathInLayout reports whether p is an entry, a directory on the way to one, or
// something inside an entry that is a directory.
func pathInLayout(p string, layout []string) bool {
	segs := strings.Split(strings.TrimSuffix(p, "/"), "/")
	for _, e := range layout {
		esegs := strings.Split(strings.TrimSuffix(e, "/"), "/")
		dir := strings.HasSuffix(e, "/")
		n := len(segs)
		if n > len(esegs) {
			if !dir {
				continue
			}
			n = len(esegs)
		}
		ok := true
		for i := 0; i < n && ok; i++ {
			ok = segMatch(esegs[i], segs[i])
		}
		if ok {
			return true
		}
	}
	return false
}

// segMatch matches one path segment, a <> placeholder standing for any run of
// characters in the layout and in the skill's path alike.
func segMatch(entry, seg string) bool {
	re := "^" + strings.ReplaceAll(regexp.QuoteMeta(entry), regexp.QuoteMeta("<>"), "[^/]+") + "$"
	return regexp.MustCompile(re).MatchString(seg)
}

// fenceRules are the embedded fence's Bash(vloop ...) rules, as command prefixes.
type fenceRules struct{ allow, deny []string }

func loadFence() (*fenceRules, error) {
	var f struct {
		Permissions struct {
			Allow []string `json:"allow"`
			Deny  []string `json:"deny"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(vloop.Fence, &f); err != nil {
		return nil, err
	}
	return &fenceRules{allow: bashPrefixes(f.Permissions.Allow), deny: bashPrefixes(f.Permissions.Deny)}, nil
}

func bashPrefixes(rules []string) []string {
	var out []string
	for _, r := range rules {
		if !strings.HasPrefix(r, "Bash(") || !strings.HasSuffix(r, ")") {
			continue
		}
		p := strings.TrimSuffix(strings.TrimPrefix(r, "Bash("), ")")
		out = append(out, strings.TrimSuffix(p, ":*"))
	}
	return out
}

func prefixMatch(rules []string, cmd string) (string, bool) {
	for _, p := range rules {
		if cmd == p || strings.HasPrefix(cmd, p+" ") {
			return p, true
		}
	}
	return "", false
}

// judge returns "" when cmd is allowed and not denied, else why not.
func (f *fenceRules) judge(cmd string) string {
	if p, ok := prefixMatch(f.deny, cmd); ok {
		return fmt.Sprintf("denied by the fence (%s)", p)
	}
	if _, ok := prefixMatch(f.allow, cmd); !ok {
		return "not allowed by the fence"
	}
	return ""
}

// The eval case format is claude plugin eval's (code.claude.com/docs/en/plugin-evals),
// not ours: these lists are its documented keys. A key the tool does not know
// makes it refuse the case outright — every case B7's run wrote declared its
// scaffold as `setup:`, and the tool loaded none of them.
var (
	evalPromptKeys = []string{"schema_version", "name", "description", "tags", "plugins", "runs",
		"expected_outcome", "model", "max_turns", "timeout_seconds", "allowed_tools",
		"artifact_publish", "growthbook_overrides", "append_system_prompt", "env"}
	evalCaseKeys = []string{"schema_version", "name", "description", "tags", "plugins", "runs",
		"expected_outcome", "execution", "context", "graders"}
	// Each grader type and the keys it requires; an llm grader's criteria are
	// its body. There is no grader that runs a command.
	evalGraderKeys = map[string][]string{
		"regex": {"pattern", "match", "target"}, "tool_used": {"tool"},
		"tool_order": {"before", "after"}, "file_exists": {"path"},
		"llm": nil, "baseline": {"baseline_file", "criteria"},
	}
)

// frontmatterKeys returns every top-level key between the leading --- lines,
// including keys whose value is a nested block (frontmatter skips those).
func frontmatterKeys(text string) []string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil
	}
	var keys []string
	for _, l := range lines[1:] {
		if strings.TrimSpace(l) == "---" {
			break
		}
		if k, _, ok := strings.Cut(l, ":"); ok && l != "" && !strings.HasPrefix(l, " ") && !strings.HasPrefix(l, "\t") && !strings.HasPrefix(l, "-") && !strings.HasPrefix(l, "#") {
			keys = append(keys, strings.TrimSpace(k))
		}
	}
	return keys
}

// yamlTopKeys returns a YAML document's top-level keys, by indentation.
func yamlTopKeys(text string) []string {
	return frontmatterKeys("---\n" + text + "\n---")
}

func inList(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func checkEvals(dir string) []string {
	var out []string
	cases, _ := filepath.Glob(filepath.Join(dir, "evals", "*"))
	sort.Strings(cases)
	for _, c := range cases {
		if st, err := os.Stat(c); err != nil || !st.IsDir() {
			continue
		}
		name := filepath.Base(c)
		if raw, err := os.ReadFile(filepath.Join(c, "prompt.md")); err != nil {
			out = append(out, fmt.Sprintf("eval %s: no prompt.md", name))
		} else {
			fm := frontmatter(string(raw))
			for _, k := range []string{"max_turns", "allowed_tools"} {
				if fm[k] == "" {
					out = append(out, fmt.Sprintf("eval %s: prompt.md frontmatter is missing %s", name, k))
				}
			}
			if v := fm["allowed_tools"]; v != "" && !strings.HasPrefix(v, "[") {
				out = append(out, fmt.Sprintf("eval %s: prompt.md allowed_tools must be a YAML array, not %q", name, v))
			}
			for _, k := range frontmatterKeys(string(raw)) {
				if !inList(evalPromptKeys, k) {
					out = append(out, fmt.Sprintf("eval %s: prompt.md frontmatter key %q is not one claude plugin eval knows", name, k))
				}
			}
		}
		// A scaffold runs only when case.yaml declares it under context.
		if _, err := os.Stat(filepath.Join(c, "scaffold.sh")); err == nil {
			raw, err := os.ReadFile(filepath.Join(c, "case.yaml"))
			switch {
			case err != nil:
				out = append(out, fmt.Sprintf("eval %s: scaffold.sh is not declared: no case.yaml with context.scaffold_script", name))
			case !strings.Contains(string(raw), "\n  scaffold_script: scaffold.sh"):
				out = append(out, fmt.Sprintf("eval %s: case.yaml does not declare context.scaffold_script: scaffold.sh", name))
			}
		}
		if raw, err := os.ReadFile(filepath.Join(c, "case.yaml")); err == nil {
			keys := yamlTopKeys(string(raw))
			for _, k := range keys {
				if !inList(evalCaseKeys, k) {
					out = append(out, fmt.Sprintf("eval %s: case.yaml key %q is not one claude plugin eval knows", name, k))
				}
			}
			for _, k := range []string{"schema_version", "name"} {
				if !inList(keys, k) {
					out = append(out, fmt.Sprintf("eval %s: case.yaml is missing %s", name, k))
				}
			}
		}
		graders, _ := filepath.Glob(filepath.Join(c, "graders", "*.md"))
		if len(graders) == 0 {
			out = append(out, fmt.Sprintf("eval %s: no graders/*.md", name))
		}
		for _, g := range graders {
			raw, err := os.ReadFile(g)
			if err != nil {
				out = append(out, fmt.Sprintf("eval %s: %v", name, err))
				continue
			}
			fm := frontmatter(string(raw))
			for _, k := range []string{"type", "weight"} {
				if fm[k] == "" {
					out = append(out, fmt.Sprintf("eval %s: grader %s frontmatter is missing %s", name, filepath.Base(g), k))
				}
			}
			need, ok := evalGraderKeys[fm["type"]]
			if !ok && fm["type"] != "" {
				out = append(out, fmt.Sprintf("eval %s: grader %s type %q is not a claude plugin eval grader type (regex, tool_used, tool_order, file_exists, llm, baseline)", name, filepath.Base(g), fm["type"]))
				continue
			}
			keys := frontmatterKeys(string(raw))
			for _, k := range need {
				if !inList(keys, k) {
					out = append(out, fmt.Sprintf("eval %s: %s grader %s is missing %s", name, fm["type"], filepath.Base(g), k))
				}
			}
		}
	}
	return out
}
