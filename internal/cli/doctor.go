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
	} else {
		add(doctorPlugin(b))
	}

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
	case b.Commit != "" && head != "" && strings.HasPrefix(head, b.Commit):
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
	for _, k := range keys {
		if doc.Projects[k].Trusted {
			return true
		}
	}
	return false
}

// doctorPlan reads just the plan's shell and status, leniently: validity is
// the plan check's business.
func doctorPlan(root string) (p struct{ Shell, Status string }, ok bool) {
	raw, err := os.ReadFile(state.Path(root))
	if err != nil {
		return p, false
	}
	_ = json.Unmarshal(raw, &p)
	return p, true
}

func doctorPlugin(b Build) (string, string, string) {
	out, err := exec.Command("claude", "plugin", "list", "--json").Output()
	var list []struct {
		ID      string `json:"id"`
		Version string `json:"version"`
		Enabled bool   `json:"enabled"`
	}
	if err == nil {
		err = json.Unmarshal(out, &list)
	}
	if err != nil {
		return "plugin", resWarning, "cannot read claude plugin list --json"
	}
	found := ""
	for _, p := range list {
		if !p.Enabled || !strings.HasPrefix(p.ID, "vloop@") {
			continue
		}
		if p.Version == b.Version {
			return "plugin", resPass, ""
		}
		found = p.Version
	}
	if found == "" {
		return "plugin", resWarning, "no enabled vloop plugin — run claude plugin install vloop@vloop"
	}
	return "plugin", resWarning, fmt.Sprintf("the vloop plugin is %s but this binary is %s — update the one that is behind", found, b.Version)
}

var goModule = regexp.MustCompile(`(?m)^module\s+github\.com/mvelosop/vloop\s*$`)

func ownModule(root string) bool {
	raw, err := os.ReadFile(filepath.Join(root, "go.mod"))
	return err == nil && goModule.Match(raw)
}
