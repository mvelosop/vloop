package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/budget"
	"github.com/mvelosop/vloop/internal/classify"
	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/driver"
	"github.com/mvelosop/vloop/internal/install"
	"github.com/mvelosop/vloop/internal/state"
)

// Doctor check results.
const (
	resPass    = "pass"
	resWarning = "warning"
	resProblem = "problem"
	resNA      = "n/a"
)

type doctorCheck struct {
	Name    string `json:"name"`
	Result  string `json:"result"`
	Message string `json:"message"`
}

func newDoctor(b Build, g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check that the repository, the toolchain and the plugin are ready for a run",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := g.root()
			if err != nil {
				return Problem(err)
			}
			checks := runDoctor(b, root)
			return writeDoctor(g, cmd, checks)
		},
	}
}

func writeDoctor(g *Globals, cmd *cobra.Command, checks []doctorCheck) error {
	out := cmd.OutOrStdout()
	problems, warnings := 0, 0
	for _, c := range checks {
		switch c.Result {
		case resProblem:
			problems++
		case resWarning:
			warnings++
		}
	}
	if g.JSON {
		enc := json.NewEncoder(out)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(struct {
			OK     bool          `json:"ok"`
			Checks []doctorCheck `json:"checks"`
		}{problems == 0, checks}); err != nil {
			return err
		}
	} else {
		symbol := map[string]string{resPass: "✓", resWarning: "!", resProblem: "✗", resNA: "-"}
		for _, c := range checks {
			line := symbol[c.Result] + " " + c.Name
			if c.Message != "" {
				line += " " + c.Message
			}
			fmt.Fprintln(out, line)
		}
		fmt.Fprintf(out, "doctor: %d problem(s), %d warning(s)\n", problems, warnings)
	}
	if problems > 0 {
		return Problem(errors.New(""))
	}
	return nil
}

func runDoctor(b Build, root string) []doctorCheck {
	var checks []doctorCheck
	add := func(name, result, msg string) { checks = append(checks, doctorCheck{name, result, msg}) }

	isRepo, head := doctorGit(root, add)

	stamp, serr := install.Read(root)
	switch {
	case errors.Is(serr, os.ErrNotExist):
		add("install", resWarning, "not initialized — run vloop init")
	case serr != nil:
		add("install", resProblem, serr.Error())
	default:
		rel, err := install.Compare(stamp.Version, b.Version)
		switch {
		case err != nil:
			add("install", resProblem, err.Error())
		case rel == install.FromNewer:
			add("install", resProblem, fmt.Sprintf("set up by vloop %s, newer than this binary (%s)", stamp.Version, b.Version))
		case rel == install.Same:
			add("install", resPass, "")
		default:
			add("install", resWarning, fmt.Sprintf("set up by vloop %s, this binary is %s — run vloop upgrade", stamp.Version, b.Version))
		}
	}

	vals, cerr := config.List(root)
	if cerr != nil {
		add("config", resProblem, cerr.Error())
	} else {
		add("config", resPass, "")
	}

	if cerr == nil {
		if cks, err := config.Checks(root); err == nil && len(cks) == 0 {
			add("checks", resProblem, "no [[check]] is configured — a run needs at least one; see docs/guide/configuration.md")
		}
		for _, v := range vals {
			if v.Key != "run.gate-scratch" {
				continue
			}
			var loose []string
			for _, d := range v.List {
				if isRepo && !gitIgnored(root, d) {
					loose = append(loose, d)
				}
			}
			if len(loose) > 0 {
				add("gate scratch", resProblem, strings.Join(loose, ", ")+" is not ignored by git — add it to .gitignore")
			}
		}
	}

	_, lerr := exec.LookPath("claude")
	hasClaude := lerr == nil
	switch {
	case !hasClaude:
		add("claude", resProblem, "claude is not on PATH")
	default:
		if err := exec.Command("claude", "--version").Run(); err != nil {
			add("claude", resProblem, "claude --version failed: "+err.Error())
		} else {
			add("claude", resPass, "")
		}
	}

	if doctorTrusted(root) {
		add("trust", resPass, "")
	} else {
		add("trust", resWarning, "this folder is not marked trusted in the Claude settings — run claude here once and accept the trust dialog")
	}

	plan, hasPlan := doctorPlan(root)
	shell := plan.Shell
	if !hasPlan {
		shell = ""
		for _, v := range vals {
			if v.Key == "shell" {
				shell = v.Value
			}
		}
	}
	switch {
	case shell == "":
		add("gate shell", resNA, "")
	case pathHas(shell):
		add("gate shell", resPass, "")
	case hasPlan:
		add("gate shell", resProblem, fmt.Sprintf("the plan's shell %q is not on PATH", shell))
	default:
		add("gate shell", resWarning, fmt.Sprintf("the configured shell %q is not on PATH", shell))
	}

	if !hasPlan {
		add("plan", resNA, "")
		add("branch", resNA, "")
	} else {
		areas := []string{}
		for _, v := range vals {
			if v.Key == "areas" {
				areas = v.List
			}
		}
		rep, err := state.Check(root, areas)
		switch {
		case plan.Schema == state.SchemaV1 && plan.Status == "complete":
			add("plan", resPass, "a complete "+state.SchemaV1+" plan — the next brief's plan replaces it")
		case plan.Schema == state.SchemaV1:
			add("plan", resWarning, "a "+state.SchemaV1+" plan — finish it with vloop 1.x or re-plan the brief")
		case err != nil:
			add("plan", resProblem, err.Error())
		case len(rep.Problems) > 0:
			add("plan", resProblem, fmt.Sprintf("%d problem(s) — run vloop task validate", len(rep.Problems)))
		default:
			add("plan", resPass, "")
		}
		if (plan.Status == "running" || plan.Status == "planning") && isRepo && doctorOnDefaultBranch(root) {
			add("branch", resProblem, "a run happens on a work branch, not the default branch")
		} else {
			add("branch", resPass, "")
		}
	}

	if !hasClaude {
		add("plugin", resNA, "")
		add("gate edits", resNA, "")
	} else {
		add(doctorPlugin(b))
		add(doctorGateEdits(root))
	}
	add(doctorGuidelines(root))

	if cerr == nil {
		var missing []string
		scoped := false
		for _, v := range vals {
			if v.Key != "metrics.stacks" {
				continue
			}
			for _, e := range v.List {
				_, scope, _ := classify.ParseStack(e)
				if scope == "" {
					continue
				}
				scoped = true
				if fi, err := os.Stat(filepath.Join(root, filepath.FromSlash(scope))); err != nil || !fi.IsDir() {
					missing = append(missing, e+": no such directory")
				}
			}
		}
		switch {
		case len(missing) > 0:
			add("stacks", resWarning, strings.Join(missing, "; "))
		case scoped:
			add("stacks", resPass, "")
		}
	}

	switch {
	case !isRepo || !ownModule(root):
		add("self-hosting", resNA, "")
	case (b.Commit == "" || b.Commit == "unknown") && isRelease(b.Module):
		add("self-hosting", resPass, "")
	case b.Commit == "" || b.Commit == "unknown":
		add("self-hosting", resWarning, "this binary carries no commit — build releases with the stamped build in docs/guide/concepts.md")
	case head != "" && strings.HasPrefix(head, b.Commit):
		add("self-hosting", resWarning, "this binary was built from HEAD — build the next vloop with a released one")
	default:
		add("self-hosting", resPass, "")
	}
	return checks
}

// doctorGit adds the git check and reports whether root is a git repository
// and its HEAD sha ("" when there is none).
func doctorGit(root string, add func(name, result, msg string)) (bool, string) {
	if _, err := gitOut(root, "rev-parse", "--git-dir"); err != nil {
		add("git", resProblem, "not a git repository")
		return false, ""
	}
	var unset []string
	for _, k := range []string{"user.name", "user.email"} {
		if v, err := gitOut(root, "config", k); err != nil || v == "" {
			unset = append(unset, k)
		}
	}
	if len(unset) > 0 {
		add("git", resProblem, strings.Join(unset, " and ")+" not set")
	} else {
		add("git", resPass, "")
	}
	head, _ := gitOut(root, "rev-parse", "HEAD")
	return true, head
}

func gitIgnored(root, rel string) bool { return driver.GitIgnored(root, rel) }

func gitOut(root string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", root}, args...)...).Output()
	return strings.TrimSpace(string(out)), err
}

// doctorOnDefaultBranch reports whether HEAD is on the default branch.
func doctorOnDefaultBranch(root string) bool { return driver.OnDefaultBranch(root) }

func pathHas(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// doctorTrusted reads the user's Claude settings, the one read outside the
// repository: .projects[<root>].hasTrustDialogAccepted in ~/.claude.json.
func doctorTrusted(root string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	raw, err := os.ReadFile(filepath.Join(home, ".claude.json"))
	if err != nil {
		return false
	}
	var doc struct {
		Projects map[string]struct {
			Trusted bool `json:"hasTrustDialogAccepted"`
		} `json:"projects"`
	}
	if json.Unmarshal(raw, &doc) != nil {
		return false
	}
	keys := []string{root}
	if real, err := filepath.EvalSymlinks(root); err == nil {
		keys = append(keys, real)
	}
	for key, proj := range doc.Projects {
		if !proj.Trusted {
			continue
		}
		for _, k := range keys {
			if trustKeyMatch(hostOS, key, k) {
				return true
			}
		}
	}
	return false
}

// doctorPlan reads just the plan's shell and status, leniently: validity is
// the plan check's business.
func doctorPlan(root string) (p struct{ Schema, Shell, Status string }, ok bool) {
	raw, err := os.ReadFile(state.Path(root))
	if err != nil {
		return p, false
	}
	_ = json.Unmarshal(raw, &p)
	return p, true
}

func doctorPlugin(b Build) (string, string, string) {
	out, err := exec.Command("claude", "plugin", "list", "--json").Output()
	var stderr string
	if ee := (*exec.ExitError)(nil); errors.As(err, &ee) {
		stderr = strings.TrimSpace(string(ee.Stderr))
	}
	var list []struct {
		ID      string `json:"id"`
		Version string `json:"version"`
		Enabled bool   `json:"enabled"`
	}
	if err == nil {
		err = json.Unmarshal(out, &list)
	}
	if err != nil {
		if stderr != "" {
			err = errors.New(stderr)
		}
		return "plugin", resWarning, "cannot read claude plugin list --json: " + err.Error()
	}
	found, disabled := "", false
	for _, p := range list {
		if !strings.HasPrefix(p.ID, "vloop@") {
			continue
		}
		if !p.Enabled {
			disabled = true
			continue
		}
		if p.Version == b.Version {
			return "plugin", resPass, ""
		}
		found = p.Version
	}
	if found == "" {
		if disabled {
			return "plugin", resWarning, "installed but disabled — claude plugin enable vloop@vloop"
		}
		return "plugin", resWarning, `/vloop:operate is not available in your interactive sessions — claude plugin install vloop@vloop (after claude plugin marketplace add mvelosop/vloop), or claude --plugin-dir "$(vloop plugin path)" for one session`
	}
	return "plugin", resWarning, fmt.Sprintf("the vloop plugin is %s but this binary is %s — update the one that is behind", found, b.Version)
}

// userHomeDir is the seam through which doctor finds the user's home, so tests
// can set it on every OS.
var userHomeDir = os.UserHomeDir

// gateEditsWarning names the rule that lets an interactive session edit gate
// folders, which a gate repair needs.
const gateEditsWarning = `no allow rule lets an interactive session edit .vloop/state/gates/ — add "Edit(./.vloop/state/gates/**)" to permissions.allow in .claude/settings.json`

// gateEditsAllows reports whether one permissions.allow entry lets a session
// edit .vloop/state/gates/: exactly Edit, or Edit(<p>) with p one of the
// listed patterns once one leading "./" is removed.
func gateEditsAllows(entry string) bool {
	if entry == "Edit" {
		return true
	}
	if !strings.HasPrefix(entry, "Edit(") || !strings.HasSuffix(entry, ")") {
		return false
	}
	p := strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(entry, "Edit("), ")"), "./")
	switch p {
	case "**", ".vloop/**", ".vloop/state/**", ".vloop/state/gates/**":
		return true
	}
	return false
}

// doctorGateEdits reads permissions.allow of the user, project and local
// Claude Code settings, in that order, and never writes any of them.
func doctorGateEdits(root string) (string, string, string) {
	type file struct{ path, shown string }
	var files []file
	if home, err := userHomeDir(); err == nil {
		files = append(files, file{filepath.Join(home, ".claude", "settings.json"), "~/.claude/settings.json"})
	}
	files = append(files,
		file{filepath.Join(root, ".claude", "settings.json"), ".claude/settings.json"},
		file{filepath.Join(root, ".claude", "settings.local.json"), ".claude/settings.local.json"})
	for _, f := range files {
		raw, err := os.ReadFile(f.path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		var doc struct {
			Permissions struct {
				Allow []any `json:"allow"`
			} `json:"permissions"`
		}
		if err == nil {
			err = json.Unmarshal(raw, &doc)
		}
		if err != nil {
			return "gate edits", resWarning, "cannot read " + f.shown + ": " + err.Error()
		}
		for _, e := range doc.Permissions.Allow {
			if s, ok := e.(string); ok && gateEditsAllows(s) {
				return "gate edits", resPass, ""
			}
		}
	}
	return "gate edits", resWarning, gateEditsWarning
}

// guidelinesBudget is the most rules .vloop/repo-guidelines.md should hold.
const guidelinesBudget = 20

// doctorGuidelines counts the rules of the guidelines file and warns past the
// budget or on a rule with no provenance; it never writes the file.
func doctorGuidelines(root string) (string, string, string) {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(".vloop/repo-guidelines.md")))
	if errors.Is(err, os.ErrNotExist) {
		return "guidelines", resNA, ""
	}
	if err != nil {
		return "guidelines", resWarning, "cannot read .vloop/repo-guidelines.md: " + err.Error()
	}
	rules, noID := budget.GuidelineRules(string(raw))
	var msgs []string
	if len(rules) > guidelinesBudget {
		msgs = append(msgs, fmt.Sprintf("%d rules, over the budget of %d — retire or merge some before adding more", len(rules), guidelinesBudget))
	}
	if len(noID) > 0 {
		lines := make([]string, len(noID))
		for i, n := range noID {
			lines[i] = fmt.Sprintf("line %d", n)
		}
		prefix := ""
		if len(msgs) > 0 {
			prefix = "guidelines "
		}
		msgs = append(msgs, fmt.Sprintf("%s%d rule(s) cite no defect or intervention: %s", prefix, len(noID), strings.Join(lines, ", ")))
	}
	if len(msgs) > 0 {
		return "guidelines", resWarning, strings.Join(msgs, "; ")
	}
	return "guidelines", resPass, fmt.Sprintf("%d rule(s)", len(rules))
}

var goModule = regexp.MustCompile(`(?m)^module\s+github\.com/mvelosop/vloop\s*$`)

func ownModule(root string) bool {
	raw, err := os.ReadFile(filepath.Join(root, "go.mod"))
	return err == nil && goModule.Match(raw)
}
