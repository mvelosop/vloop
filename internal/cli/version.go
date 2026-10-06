package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	vloop "github.com/mvelosop/vloop"
)

// versionInfo's field order is the JSON key order.
type versionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Plugin  string `json:"plugin"`
	Go      string `json:"go"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

// pluginVersion reads the version from the embedded plugin.json, never from
// the working tree.
func pluginVersion() (string, error) {
	b, err := vloop.Plugin.ReadFile("plugin/.claude-plugin/plugin.json")
	if err != nil {
		return "", err
	}
	var m struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return "", fmt.Errorf("embedded plugin.json: %w", err)
	}
	return m.Version, nil
}

// checkPlugin is the SessionStart hook's handshake. It reads one JSON file and
// nothing else, prints to stdout only, and never fails: a failing hook would
// interrupt every session.
func checkPlugin(out io.Writer, b Build, dir string) {
	var m struct {
		Version string `json:"version"`
	}
	raw, err := os.ReadFile(filepath.Join(dir, ".claude-plugin", "plugin.json"))
	if err == nil {
		err = json.Unmarshal(raw, &m)
	}
	switch {
	case err != nil || m.Version == "":
		fmt.Fprintf(out, "vloop: cannot read the vloop plugin's version in %s\n", dir)
	case m.Version != b.Version:
		fmt.Fprintf(out, "vloop: the vloop plugin is %s but the vloop binary is %s \u2014 update the one that is behind\n", m.Version, b.Version)
	}
}

func newVersion(b Build, g *Globals) *cobra.Command {
	var check string
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the vloop and embedded plugin versions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if cmd.Flags().Changed("check-plugin") {
				checkPlugin(cmd.OutOrStdout(), b, check)
				return nil
			}
			plugin, err := pluginVersion()
			if err != nil {
				return Problem(err)
			}
			v := versionInfo{b.Version, b.Commit, plugin, runtime.Version(), runtime.GOOS, runtime.GOARCH}
			out := cmd.OutOrStdout()
			if g.JSON {
				return json.NewEncoder(out).Encode(v)
			}
			_, err = fmt.Fprintf(out, "vloop %s (commit %s, plugin %s, %s %s/%s)\n",
				v.Version, v.Commit, v.Plugin, v.Go, v.OS, v.Arch)
			return err
		},
	}
	cmd.Flags().StringVar(&check, "check-plugin", "", "compare the plugin in `dir` with this binary's version; always exits 0")
	return cmd
}
