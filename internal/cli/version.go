package cli

import (
	"encoding/json"
	"fmt"
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

func newVersion(b Build, g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the vloop and embedded plugin versions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
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
}
