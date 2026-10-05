// Package intervention records what the operator does around a run besides
// testing, one file each under .vloop/interventions/.
package intervention

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/defect"
)

// Dir is the repo-relative folder the intervention files live in.
const Dir = ".vloop/interventions"

// Enum values, in the order errors list them (and, for phases, the lifecycle).
var (
	Phases       = []string{"setup", "design", "run", "halt", "verify", "close", "next"}
	Kinds        = []string{"direction", "decision", "context-supply", "halt", "verification-finding", "repair", "carry-forward", "ceremony"}
	Automatables = []string{"yes", "partly", "no"}
	Bys          = []string{"operator", "assistant", "both"}
)

// SetFields are the fields `intervention set` may change.
var SetFields = []string{"brief", "phase", "kind", "automatable", "by", "occurred", "recommended", "decided", "adjusted"}

// Agreements are the values of the derived agreement field.
var Agreements = []string{"recommended", "other-option", "adjusted", "different", "no-options"}

// MaxOptions is how many options a record may carry.
const MaxOptions = 3

// Recommended is the option the operator's assistant recommended, and why.
type Recommended struct {
	Option int    `json:"option"`
	Why    string `json:"why"`
}

// Decided is what was decided: option 0 is another choice, or none.
type Decided struct {
	Option   int    `json:"option"`
	Adjusted bool   `json:"adjusted"`
	Text     string `json:"text"`
}

// Intervention is the intervention/v2 frontmatter as JSON, plus the body. A
// v1 record reads into it with Schema intervention/v1 and agreement no-options.
type Intervention struct {
	Schema      string `json:"schema"`
	ID          string `json:"id"`
	Brief       string `json:"brief"`
	Phase       string `json:"phase"`
	Kind        string `json:"kind"`
	Automatable string `json:"automatable"`
	By          string `json:"by"`
	Occurred    string `json:"occurred"`
	Recorded    string `json:"recorded"`
	Backfilled  bool   `json:"backfilled,omitempty"`
	Summary     string `json:"summary"`
	Trigger     string `json:"trigger"`
	Done        string `json:"done"`
	Automation  string `json:"automation"`

	Context     string      `json:"context,omitempty"`
	Suggested   string      `json:"suggested,omitempty"`
	Options     []string    `json:"options"`
	Recommended Recommended `json:"recommended"`
	Decided     Decided     `json:"decided"`
	Agreement   string      `json:"agreement"`

	// decision is the stored decided: value, 1-3, other or "".
	decision string
}

// DeriveAgreement is the one place the agreement comes from: the option
// count, the recommended option, the decision (1-3, "other" or "") and
// whether the decided option was adjusted.
func DeriveAgreement(options, recommended int, decided string, adjusted bool) (string, error) {
	if options == 0 {
		return "no-options", nil
	}
	switch decided {
	case "":
		return "", errors.New("options need a decision: decided is empty")
	case "other":
		return "different", nil
	}
	n, err := strconv.Atoi(decided)
	if err != nil || n < 1 || n > options {
		return "", fmt.Errorf("decided %q is not one of 1-%d or other", decided, options)
	}
	switch {
	case adjusted:
		return "adjusted", nil
	case n == recommended:
		return "recommended", nil
	}
	return "other-option", nil
}

// NewInput is what `intervention add` takes.
type NewInput struct {
	Summary, Brief, Phase, Kind, Automatable, By, Trigger, Done, Automation string

	Context string
	Options []string
	// Why is the reason for the recommended option.
	Why string
	// Recommended and DecidedOption are 1-based; the *Set flags say whether
	// the flag was given, so that 0 is refused rather than read as unset.
	Recommended, DecidedOption       int
	RecommendedSet, DecidedOptionSet bool
	DecidedOther, Adjusted           bool
	Decided                          string
}

// CheckOptions refuses a malformed combination of the options flags, with the
// message the CLI prints; nothing is written before it passes.
func CheckOptions(in NewInput) error {
	n := len(in.Options)
	if n == 0 {
		if in.RecommendedSet || in.DecidedOptionSet || in.DecidedOther || in.Why != "" {
			return errors.New("--recommended, --decided-option and --decided-other need options")
		}
		if in.Adjusted {
			return errors.New("--adjusted needs --decided-option")
		}
		return nil
	}
	if n > MaxOptions {
		return errors.New("at most three options")
	}
	if !in.RecommendedSet || in.Why == "" {
		return errors.New("options need --recommended and --why")
	}
	if in.Recommended < 1 || in.Recommended > n {
		return fmt.Errorf("--recommended must name an option, 1 to %d", n)
	}
	if in.DecidedOptionSet && in.DecidedOther {
		return errors.New("--decided-option and --decided-other exclude each other")
	}
	if !in.DecidedOptionSet && !in.DecidedOther {
		return errors.New("options need --decided-option or --decided-other")
	}
	if in.DecidedOptionSet && (in.DecidedOption < 1 || in.DecidedOption > n) {
		return fmt.Errorf("--decided-option must name an option, 1 to %d", n)
	}
	if in.Adjusted && !in.DecidedOptionSet {
		return errors.New("--adjusted needs --decided-option")
	}
	return nil
}

// Validate checks one enum value, returning the usage error the CLI reports.
func Validate(field, value string) error {
	var valid []string
	switch field {
	case "phase":
		valid = Phases
	case "kind":
		valid = Kinds
	case "automatable":
		valid = Automatables
	case "by":
		valid = Bys
	case "occurred":
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return &config.InvalidValueError{Key: field, Value: value}
		}
		return nil
	default:
		return nil
	}
	if !slices.Contains(valid, value) {
		return &config.InvalidValueError{Key: field, Value: value, Valid: valid}
	}
	return nil
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

// Add writes a new intervention file and returns its repo-relative path.
// Values are validated by the caller (Validate).
func Add(root string, in NewInput, now time.Time) (string, error) {
	if err := CheckOptions(in); err != nil {
		return "", err
	}
	summary := oneLine(in.Summary)
	v := Intervention{
		Schema: "intervention/v2", Agreement: "no-options", Options: []string{}, Brief: defect.BriefName(in.Brief), Phase: in.Phase, Kind: in.Kind,
		Automatable: in.Automatable, By: in.By, Occurred: now.Format("2006-01-02"),
		Recorded: now.UTC().Format("2006-01-02T15:04:05Z"), Summary: summary,
		Trigger: oneLine(in.Trigger), Done: oneLine(in.Done), Automation: oneLine(in.Automation),
	}
	v.Context, v.Decided.Text = oneLine(in.Context), oneLine(in.Decided)
	for _, o := range in.Options {
		v.Options = append(v.Options, oneLine(o))
	}
	if len(v.Options) > 0 {
		v.Recommended = Recommended{Option: in.Recommended, Why: oneLine(in.Why)}
		v.decision = "other"
		if in.DecidedOptionSet {
			v.Decided.Option, v.decision = in.DecidedOption, strconv.Itoa(in.DecidedOption)
			v.Decided.Adjusted = in.Adjusted
		}
		a, err := DeriveAgreement(len(v.Options), in.Recommended, v.decision, in.Adjusted)
		if err != nil {
			return "", err
		}
		v.Agreement = a
	}
	dir := filepath.Join(root, filepath.FromSlash(Dir))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	slug := defect.Slug(summary)
	if slug == "defect" && summary == "" {
		slug = "intervention"
	}
	base := "I" + now.Format("20060102-1504") + "-" + slug
	for n := 1; ; n++ {
		id := base
		if n > 1 {
			id = base + "-" + strconv.Itoa(n)
		}
		v.ID = id
		f, err := os.OpenFile(filepath.Join(dir, id+".md"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		_, werr := f.WriteString(render(v))
		if cerr := f.Close(); werr == nil {
			werr = cerr
		}
		return Dir + "/" + id + ".md", werr
	}
}

func quote(s string) string {
	if s == "" || strings.ContainsAny(s, ":#\"'") || s != strings.TrimSpace(s) {
		return strconv.Quote(s)
	}
	return s
}

func render(v Intervention) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %s\nbrief: %s\nphase: %s\nkind: %s\nautomatable: %s\nby: %s\n", v.ID, quote(v.Brief), v.Phase, v.Kind, v.Automatable, v.By)
	fmt.Fprintf(&b, "schema: intervention/v2\noptions: %d\nrecommended: %d\ndecided: %s\n", len(v.Options), v.Recommended.Option, quote(v.decision))
	if v.Decided.Adjusted {
		b.WriteString("adjusted: true\n")
	}
	fmt.Fprintf(&b, "agreement: %s\n", v.Agreement)
	fmt.Fprintf(&b, "occurred: %s\nrecorded: %s\n", v.Occurred, v.Recorded)
	if v.Backfilled {
		b.WriteString("backfilled: true\n")
	}
	fmt.Fprintf(&b, "---\n%s\n", v.Summary)
	section := func(name, text string) {
		if text != "" {
			fmt.Fprintf(&b, "\n**%s.** %s\n", name, text)
		}
	}
	section("Trigger", v.Trigger)
	section("Done", v.Done)
	section("Context", v.Context)
	if len(v.Options) > 0 {
		b.WriteString("\n**Options.**\n")
		for i, o := range v.Options {
			fmt.Fprintf(&b, "%d. %s\n", i+1, o)
		}
	}
	section("Recommended", v.Recommended.Why)
	section("Suggested", v.Suggested)
	section("Decided", v.Decided.Text)
	section("What would automate it", v.Automation)
	return b.String()
}

func unquote(v string) string {
	v = strings.TrimSpace(v)
	if strings.HasPrefix(v, `"`) {
		if u, err := strconv.Unquote(v); err == nil {
			return u
		}
	}
	return v
}

var (
	sectionRe = regexp.MustCompile(`^\*\*(Trigger|Done|Context|Options|Recommended|Suggested|Decided|What would automate it)\.\*\*\s*`)
	optionRe  = regexp.MustCompile(`^(\d+)\.\s+(.*)$`)
)

func parse(text string) (Intervention, error) {
	v := Intervention{Schema: "intervention/v1", Options: []string{}, Agreement: "no-options"}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return v, errors.New("missing frontmatter")
	}
	end := -1
	var nOptions, nRecommended int
	var stored string
	var adjusted bool
	num := func(key, val string) (int, error) {
		n, err := strconv.Atoi(val)
		if err != nil || n < 0 || n > MaxOptions {
			return 0, fmt.Errorf("%s: %q is not 0-%d", key, val, MaxOptions)
		}
		return n, nil
	}
	for i, l := range lines[1:] {
		if l == "---" {
			end = i + 1
			break
		}
		k, val, ok := strings.Cut(l, ":")
		if !ok {
			continue
		}
		val = unquote(val)
		var err error
		switch k {
		case "id":
			v.ID = val
		case "brief":
			v.Brief = val
		case "phase":
			v.Phase = val
		case "kind":
			v.Kind = val
		case "automatable":
			v.Automatable = val
		case "by":
			v.By = val
		case "occurred":
			v.Occurred = val
		case "recorded":
			v.Recorded = val
		case "backfilled":
			v.Backfilled = val == "true"
		case "schema":
			v.Schema = val
		case "options":
			nOptions, err = num(k, val)
		case "recommended":
			nRecommended, err = num(k, val)
		case "decided":
			v.decision = val
		case "adjusted":
			adjusted = val == "true"
		case "agreement":
			stored = val
		}
		if err != nil {
			return v, err
		}
	}
	if end < 0 {
		return v, errors.New("unterminated frontmatter")
	}
	v2 := v.Schema == "intervention/v2"
	// The summary is the first body line; each section runs to the next marker.
	var cur *string
	var list []string
	inOptions := false
	var summary []string
	for _, l := range lines[end+1:] {
		if m := sectionRe.FindStringSubmatch(l); m != nil && (v2 || m[1] != "Options" && m[1] != "Recommended") {
			inOptions = false
			switch m[1] {
			case "Trigger":
				cur = &v.Trigger
			case "Done":
				cur = &v.Done
			case "Context":
				cur = &v.Context
			case "Options":
				cur, inOptions = new(string), true
			case "Recommended":
				cur = &v.Recommended.Why
			case "Suggested":
				cur = &v.Suggested
			case "Decided":
				cur = &v.Decided.Text
			default:
				cur = &v.Automation
			}
			*cur = strings.TrimSpace(l[len(m[0]):])
			continue
		}
		switch {
		case inOptions:
			if t := strings.TrimSpace(l); t != "" {
				list = append(list, t)
			}
		case cur != nil:
			*cur = strings.TrimSpace(*cur + " " + strings.TrimSpace(l))
		default:
			summary = append(summary, l)
		}
	}
	v.Summary = strings.TrimSpace(strings.Join(summary, "\n"))
	if !v2 {
		return v, nil
	}
	for i, l := range list {
		m := optionRe.FindStringSubmatch(l)
		if m == nil || m[1] != strconv.Itoa(i+1) {
			return v, fmt.Errorf("options: line %q is not the numbered item %d", l, i+1)
		}
		v.Options = append(v.Options, m[2])
	}
	if len(v.Options) != nOptions {
		return v, fmt.Errorf("options: %d declared, %d listed", nOptions, len(v.Options))
	}
	if nRecommended > nOptions {
		return v, fmt.Errorf("recommended: %d is beyond the %d options", nRecommended, nOptions)
	}
	v.Recommended.Option = nRecommended
	v.Decided.Adjusted = adjusted
	if n, err := strconv.Atoi(v.decision); err == nil {
		v.Decided.Option = n
	}
	agreement, err := DeriveAgreement(nOptions, nRecommended, v.decision, adjusted)
	if err != nil {
		return v, err
	}
	if stored != agreement {
		return v, fmt.Errorf("agreement: stored %q, derived %q", stored, agreement)
	}
	v.Agreement = agreement
	return v, nil
}

// List reads every recorded intervention, sorted by phase in lifecycle order,
// then kind, then id, keeping only briefs matching brief when it is not empty.
func List(root, brief string) ([]Intervention, error) {
	dir := filepath.Join(root, filepath.FromSlash(Dir))
	ents, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return []Intervention{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := []Intervention{}
	for _, e := range ents {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "I") || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		v, err := parse(string(b))
		if err != nil {
			return nil, fmt.Errorf("%s/%s: %w", Dir, e.Name(), err)
		}
		if brief != "" && v.Brief != defect.BriefName(brief) {
			continue
		}
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if pa, pb := slices.Index(Phases, a.Phase), slices.Index(Phases, b.Phase); pa != pb {
			return pa < pb
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.ID < b.ID
	})
	return out, nil
}

// Set changes one frontmatter field of a recorded intervention, keeping the
// rest of the file. Validation happens before the file is touched. The choice
// fields (recommended, decided, adjusted) rewrite the derived agreement in the
// same write; agreement and options are never set.
func Set(root, id, field, value string) error {
	switch field {
	case "agreement":
		return errors.New("agreement is derived — set decided, adjusted or recommended instead")
	case "options":
		return errors.New("options are recorded with the intervention, not set")
	}
	if !slices.Contains(SetFields, field) {
		return fmt.Errorf("cannot set %q: want one of %s", field, strings.Join(SetFields, ", "))
	}
	if err := Validate(field, value); err != nil {
		return err
	}
	if strings.ContainsAny(value, "\r\n") {
		return &config.InvalidValueError{Key: field, Value: value}
	}
	path := filepath.Join(root, filepath.FromSlash(Dir), filepath.Base(id)+".md")
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("no intervention %s", id)
	}
	v, err := parse(string(b))
	if err != nil {
		return err
	}
	// updates maps a frontmatter key to its new line; "" removes the line.
	updates := map[string]string{field: field + ": " + value}
	switch field {
	case "brief":
		updates[field] = field + ": " + quote(defect.BriefName(value))
	case "recommended", "decided", "adjusted":
		n := len(v.Options)
		if n == 0 {
			return fmt.Errorf("cannot set %s: the record has no options", field)
		}
		rec, decided, adjusted := v.Recommended.Option, v.decision, v.Decided.Adjusted
		switch field {
		case "recommended":
			r, err := strconv.Atoi(value)
			if err != nil || r < 1 || r > n {
				return fmt.Errorf("recommended must name an option, 1 to %d", n)
			}
			rec = r
		case "decided":
			if d, err := strconv.Atoi(value); value != "other" && (err != nil || d < 1 || d > n) {
				return fmt.Errorf("decided must name an option, 1 to %d, or other", n)
			}
			decided = value
		case "adjusted":
			if value != "true" && value != "false" {
				return fmt.Errorf("adjusted must be true or false")
			}
			adjusted = value == "true"
		}
		if decided == "other" {
			adjusted = false
		}
		a, err := DeriveAgreement(n, rec, decided, adjusted)
		if err != nil {
			return err
		}
		updates["recommended"] = "recommended: " + strconv.Itoa(rec)
		updates["decided"] = "decided: " + quote(decided)
		updates["adjusted"] = ""
		if adjusted {
			updates["adjusted"] = "adjusted: true"
		}
		updates["agreement"] = "agreement: " + a
	}
	lines := strings.Split(string(b), "\n")
	end := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return errors.New("unterminated frontmatter")
	}
	keys := make([]string, 0, len(updates))
	for k := range updates {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		line, found := updates[k], false
		for i := 1; i < end; i++ {
			if strings.HasPrefix(lines[i], k+":") {
				found = true
				if line == "" {
					lines = slices.Delete(lines, i, i+1)
					end--
				} else {
					lines[i] = line
				}
				break
			}
		}
		if !found && line != "" { // field absent: insert where I1 orders it
			at := end
			if k == "adjusted" { // between decided and agreement
				for i := 1; i < end; i++ {
					if strings.HasPrefix(lines[i], "agreement:") {
						at = i
						break
					}
				}
			}
			lines = slices.Insert(lines, at, line)
			end++
		}
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
}

// Migrate rewrites every v1 record under Dir to intervention/v2 by inserting
// the five no-options lines right after by: and changing no other byte. With
// dryRun nothing is written. It returns the names of the records migrated (or,
// on a dry run, that would be), in name order.
func Migrate(root string, dryRun bool) ([]string, error) {
	dir := filepath.Join(root, filepath.FromSlash(Dir))
	ents, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var done []string
	for _, e := range ents {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "I") || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		out, ok, err := migrateText(string(b))
		if err != nil {
			return nil, fmt.Errorf("%s/%s: %w", Dir, e.Name(), err)
		}
		if !ok {
			continue
		}
		if !dryRun {
			if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
				return nil, err
			}
		}
		done = append(done, e.Name())
	}
	return done, nil
}

// migrateText inserts the v2 lines after the by: line of a v1 record. ok is
// false when the record already carries a schema line.
func migrateText(text string) (string, bool, error) {
	lines := strings.SplitAfter(text, "\n")
	if strings.TrimRight(lines[0], "\r\n") != "---" {
		return "", false, errors.New("missing frontmatter")
	}
	by := -1
	for i := 1; i < len(lines); i++ {
		l := strings.TrimRight(lines[i], "\r\n")
		if l == "---" {
			break
		}
		if strings.HasPrefix(l, "schema:") {
			return text, false, nil
		}
		if by < 0 && strings.HasPrefix(l, "by:") {
			by = i
		}
	}
	if by < 0 {
		return "", false, errors.New("no by: line in the frontmatter")
	}
	eol := "\n"
	if strings.HasSuffix(lines[by], "\r\n") {
		eol = "\r\n"
	}
	ins := strings.Join([]string{"schema: intervention/v2", "options: 0", "recommended: 0", `decided: ""`, "agreement: no-options"}, eol) + eol
	if !strings.HasSuffix(lines[by], "\n") {
		ins = eol + strings.TrimSuffix(ins, eol)
	}
	lines[by] += ins
	return strings.Join(lines, ""), true, nil
}
