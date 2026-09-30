package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	vloop "github.com/mvelosop/vloop"
)

const pluginManifest = ".claude-plugin/plugin.json"

func newPlugin(b Build, g *Globals) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Work with the plugin embedded in the binary",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Extract the embedded plugin under .vloop/tmp/plugin/ and print its path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := g.root()
			if err != nil {
				return Problem(err)
			}
			rel, err := extractPlugin(root, b.Version)
			if err != nil {
				return Problem(err)
			}
			out := cmd.OutOrStdout()
			if g.JSON {
				return json.NewEncoder(out).Encode(struct {
					Path string `json:"path"`
				}{rel})
			}
			_, err = fmt.Fprintln(out, rel)
			return err
		},
	})
	return cmd
}

// ensureRealDir makes dir exist as a real directory, refusing a symlink or a
// file in its place, so nothing is ever written through a link.
func ensureRealDir(dir string) error {
	fi, err := os.Lstat(dir)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return os.Mkdir(dir, 0o755)
	case err != nil:
		return err
	case !fi.IsDir():
		return fmt.Errorf("%s is not a directory (a symlink is not followed)", dir)
	}
	return nil
}

// extractPlugin makes .vloop/tmp/plugin/<version>/ under root hold exactly the
// embedded plugin and returns that directory relative to root, with "/".
func extractPlugin(root, version string) (string, error) {
	if version == "" || version == "." || version == ".." || strings.ContainsAny(version, `/\\`) {
		return "", fmt.Errorf("cannot extract the plugin for version %q", version)
	}
	rel := path.Join(".vloop", "tmp", "plugin", version)
	dir := root
	for _, part := range strings.Split(rel, "/") {
		dir = filepath.Join(dir, part)
		if err := ensureRealDir(dir); err != nil {
			return "", err
		}
	}

	want := map[string]bool{} // relative slash paths the tree must hold; true = directory
	err := fs.WalkDir(vloop.Plugin, "plugin", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := strings.TrimPrefix(strings.TrimPrefix(p, "plugin"), "/")
		if name == "" {
			return nil
		}
		want[name] = d.IsDir()
		target := filepath.Join(dir, filepath.FromSlash(name))
		fi, lerr := os.Lstat(target)
		if d.IsDir() {
			if lerr == nil && !fi.IsDir() {
				if err := os.RemoveAll(target); err != nil {
					return err
				}
			}
			return ensureRealDir(target)
		}
		content, err := vloop.Plugin.ReadFile(p)
		if err != nil {
			return err
		}
		if name == pluginManifest {
			if content, err = withVersion(content, version); err != nil {
				return err
			}
		}
		if lerr == nil {
			if fi.Mode().IsRegular() {
				if have, rerr := os.ReadFile(target); rerr == nil && bytes.Equal(have, content) {
					return nil
				}
			} else if err := os.RemoveAll(target); err != nil {
				return err
			}
		}
		return os.WriteFile(target, content, 0o644)
	})
	if err != nil {
		return "", err
	}

	// Remove what the embedded plugin does not contain. WalkDir does not
	// follow symlinks, and RemoveAll removes a link, not its target.
	var stray []string
	err = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		r, err := filepath.Rel(dir, p)
		if err != nil || r == "." {
			return err
		}
		if _, ok := want[filepath.ToSlash(r)]; !ok {
			stray = append(stray, p)
			if d.IsDir() {
				return fs.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	for _, p := range stray {
		if err := os.RemoveAll(p); err != nil {
			return "", err
		}
	}
	return rel, nil
}

// withVersion returns the manifest with its version set to v.
func withVersion(manifest []byte, v string) ([]byte, error) {
	var m map[string]any
	if err := json.Unmarshal(manifest, &m); err != nil {
		return nil, fmt.Errorf("embedded plugin.json: %w", err)
	}
	m["version"] = v
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}
