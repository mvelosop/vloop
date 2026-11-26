package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

const selfUpdateModule = "github.com/mvelosop/vloop"

var selfUpdateExecutable = os.Executable
var selfUpdateGOOS = runtime.GOOS
var selfUpdateRemove = os.Remove
var selfUpdateRename = os.Rename

type selfUpdatePlugin struct {
	Agent   string `json:"agent"`
	Scope   string `json:"scope,omitempty"`
	Result  string `json:"result"`
	Message string `json:"message"`
}

type selfUpdateDocument struct {
	Schema    string             `json:"schema"`
	Current   string             `json:"current"`
	Target    string             `json:"target"`
	Newer     bool               `json:"newer"`
	Installed bool               `json:"installed"`
	Dir       string             `json:"dir"`
	Plugins   []selfUpdatePlugin `json:"plugins"`
}

func newSelfUpdate(b Build, g *Globals) *cobra.Command {
	var check bool
	var version string
	cmd := &cobra.Command{
		Use:   "self-update [--check] [--version <v>]",
		Short: "Update the vloop binary and installed plugins",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSelfUpdate(b, g, check, cmd.Flags().Changed("version"), version, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "report whether an update is available without installing")
	cmd.Flags().StringVar(&version, "version", "", "install the exact `version` instead of the latest release")
	return cmd
}

func runSelfUpdate(b Build, g *Globals, check, explicit bool, requested string, out interface {
	Write([]byte) (int, error)
}) error {
	goPath, err := exec.LookPath("go")
	if err != nil {
		return Problem(errors.New("self-update needs go on PATH — install the release by hand: go install github.com/mvelosop/vloop/cmd/vloop@latest"))
	}
	if b.Module == "" || b.Module == "(devel)" {
		return Problem(errors.New("this vloop was built from a checkout, not installed from a release — build it again from the checkout, or go install github.com/mvelosop/vloop/cmd/vloop@latest"))
	}

	target := requested
	if explicit {
		if !strings.HasPrefix(target, "v") || !semver.IsValid(target) {
			return Usage(fmt.Errorf("--version %s is not a version like v0.15.0", target))
		}
	} else {
		target, err = latestVloopVersion(goPath)
		if err != nil {
			return Problem(fmt.Errorf("cannot find the latest release: %s", err))
		}
	}

	newer := semver.Compare(target, b.Module) > 0
	doc := selfUpdateDocument{
		Schema:  "self-update/v1",
		Current: b.Module,
		Target:  target,
		Newer:   newer,
		Plugins: []selfUpdatePlugin{},
	}
	if check {
		if g.JSON {
			return writeSelfUpdateJSON(out, doc)
		}
		_, err = fmt.Fprintf(out, "current  %s\n", b.Module)
		if err != nil {
			return err
		}
		label, action := "latest", "up to date"
		if explicit {
			label = "target"
			action = "already running"
			if target != b.Module {
				action = "vloop self-update --version " + target + " installs it"
			}
		} else if newer {
			action = "vloop self-update installs it"
		}
		_, err = fmt.Fprintf(out, "%s   %s — %s\n", label, target, action)
		return err
	}

	installNeeded := (explicit && target != b.Module) || (!explicit && newer)
	if !installNeeded {
		if g.JSON {
			return writeSelfUpdateJSON(out, doc)
		}
		if !explicit {
			_, err = fmt.Fprintf(out, "vloop %s is up to date (latest release %s)\n", b.Module, target)
		}
		return err
	}

	dir, err := selfUpdateInstallDir(goPath)
	if err != nil {
		return Problem(err)
	}
	exe, err := selfUpdateExecutable()
	if err != nil {
		return Problem(fmt.Errorf("cannot find running vloop: %s", err))
	}
	expected := filepath.Join(dir, "vloop")
	if selfUpdateGOOS == "windows" {
		expected += ".exe"
	}
	exeResolved, err := resolveSelfUpdatePath(exe)
	if err != nil {
		return Problem(fmt.Errorf("cannot resolve running vloop %s: %s", displaySelfUpdatePath(exe), err))
	}
	expectedResolved, err := resolveSelfUpdatePathAllowMissing(expected)
	if err != nil {
		return Problem(fmt.Errorf("cannot resolve go install path %s: %s", displaySelfUpdatePath(expected), err))
	}
	if !sameSelfUpdatePath(exeResolved, expectedResolved) {
		return Problem(fmt.Errorf(
			"this vloop runs from %s, but go install writes to %s — update it the way you installed it, or put %s first on PATH",
			displaySelfUpdatePath(exe), displaySelfUpdatePath(dir), displaySelfUpdatePath(dir),
		))
	}

	doc.Dir = displaySelfUpdatePath(dir)
	if !g.JSON {
		if _, err := fmt.Fprintf(out, "installing vloop %s (from %s)…\n", target, b.Module); err != nil {
			return err
		}
	}
	if err := installSelfUpdate(goPath, target, exe); err != nil {
		return Problem(err)
	}
	doc.Installed = true
	if g.JSON {
		return writeSelfUpdateJSON(out, doc)
	}
	_, err = fmt.Fprintf(out, "installed vloop %s at %s\n", target, displaySelfUpdatePath(dir))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, "next: vloop upgrade in each repository that uses vloop, and vloop self-update there for a plugin installed for that repository")
	return err
}

func installSelfUpdate(goPath, target, exe string) error {
	old := exe + ".old"
	moved := false
	if selfUpdateGOOS == "windows" {
		if err := selfUpdateRemove(old); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("cannot remove %s (%s) — delete it and run vloop self-update again", displaySelfUpdatePath(old), err)
		}
		if err := selfUpdateRename(exe, old); err != nil {
			return fmt.Errorf("cannot move %s aside (%s) — nothing was installed", displaySelfUpdatePath(exe), err)
		}
		moved = true
	}

	cmd := exec.Command(goPath, "install", selfUpdateModule+"/cmd/vloop@"+target)
	raw, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	line := lastSelfUpdateLine(raw)
	if line == "" {
		line = err.Error()
	}
	if moved {
		if removeErr := selfUpdateRemove(exe); removeErr != nil && !os.IsNotExist(removeErr) {
			return fmt.Errorf(
				"go install failed: %s; and %s could not be moved back to %s (%s) — rename it back by hand",
				line, displaySelfUpdatePath(old), displaySelfUpdatePath(exe), removeErr,
			)
		}
		if restoreErr := selfUpdateRename(old, exe); restoreErr != nil {
			return fmt.Errorf(
				"go install failed: %s; and %s could not be moved back to %s (%s) — rename it back by hand",
				line, displaySelfUpdatePath(old), displaySelfUpdatePath(exe), restoreErr,
			)
		}
	}
	return fmt.Errorf("go install failed: %s", line)
}

func latestVloopVersion(goPath string) (string, error) {
	tmp, err := os.MkdirTemp("", "vloop-self-update-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)
	cmd := exec.Command(goPath, "list", "-m", "-json", selfUpdateModule+"@latest")
	cmd.Dir = tmp
	raw, err := cmd.CombinedOutput()
	if err != nil {
		line := firstSelfUpdateLine(raw)
		if line == "" {
			line = err.Error()
		}
		return "", errors.New(line)
	}
	var module struct {
		Version string `json:"Version"`
	}
	if err := json.Unmarshal(raw, &module); err != nil {
		return "", fmt.Errorf("invalid go list -m -json output: %s", firstSelfUpdateLine(raw))
	}
	if !semver.IsValid(module.Version) {
		return "", fmt.Errorf("go list -m -json returned no valid version")
	}
	return module.Version, nil
}

func selfUpdateInstallDir(goPath string) (string, error) {
	gobin, err := selfUpdateGoEnv(goPath, "GOBIN")
	if err != nil {
		return "", fmt.Errorf("cannot find where go install writes: %s", err)
	}
	if gobin == "" {
		gobin, err = selfUpdateGoEnv(goPath, "GOPATH")
		if err != nil {
			return "", fmt.Errorf("cannot find where go install writes: %s", err)
		}
		if gobin != "" {
			gobin = filepath.Join(gobin, "bin")
		}
	}
	if gobin == "" {
		return "", errors.New("cannot find where go install writes: GOBIN and GOPATH are empty")
	}
	return gobin, nil
}

func selfUpdateGoEnv(goPath, key string) (string, error) {
	cmd := exec.Command(goPath, "env", key)
	raw, err := cmd.CombinedOutput()
	if err != nil {
		line := lastSelfUpdateLine(raw)
		if line == "" {
			line = err.Error()
		}
		return "", errors.New(line)
	}
	return strings.TrimSpace(string(raw)), nil
}

func firstSelfUpdateLine(raw []byte) string {
	for _, line := range strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n") {
		if line != "" {
			return line
		}
	}
	return ""
}

func lastSelfUpdateLine(raw []byte) string {
	lines := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if lines[i] != "" {
			return lines[i]
		}
	}
	return ""
}

func resolveSelfUpdatePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err == nil {
		return filepath.Clean(resolved), nil
	}
	return "", err
}

func resolveSelfUpdatePathAllowMissing(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return filepath.Clean(resolved), nil
	}
	parent, err := resolveSelfUpdatePathAllowMissing(filepath.Dir(abs))
	if err != nil {
		return "", err
	}
	return filepath.Join(parent, filepath.Base(abs)), nil
}

func sameSelfUpdatePath(a, b string) bool {
	if selfUpdateGOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

func displaySelfUpdatePath(path string) string {
	abs, err := filepath.Abs(path)
	if err == nil {
		path = abs
	}
	home, err := os.UserHomeDir()
	if err == nil {
		absHome, homeErr := filepath.Abs(home)
		home = absHome
		if homeErr == nil {
			rel, relErr := filepath.Rel(home, path)
			inside := relErr == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
			if selfUpdateGOOS == "windows" {
				inside = relErr == nil && rel != ".." && !strings.HasPrefix(strings.ToLower(rel), ".."+string(filepath.Separator))
			}
			if inside {
				if rel == "." {
					return "~"
				}
				return filepath.ToSlash(filepath.Join("~", rel))
			}
		}
	}
	return filepath.ToSlash(path)
}

func writeSelfUpdateJSON(out interface{ Write([]byte) (int, error) }, doc selfUpdateDocument) error {
	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)
	return enc.Encode(doc)
}
