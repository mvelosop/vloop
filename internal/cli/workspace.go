package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/mvelosop/vloop/internal/metrics"
)

// wsRepo is one `[[repo]]` table of a workspace file.
type wsRepo struct {
	Path string `toml:"path"`
	Name string `toml:"name"`
}

// loadWorkspace reads a workspace file. Relative paths are resolved later,
// against the file's own directory.
func loadWorkspace(g *Globals, file string) (dir string, repos []wsRepo, err error) {
	if !filepath.IsAbs(file) && g.Dir != "" {
		file = filepath.Join(g.Dir, file)
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		return "", nil, Problem(fmt.Errorf("cannot read workspace file: %s", file))
	}
	var doc struct {
		Repo []wsRepo `toml:"repo"`
	}
	if _, err := toml.Decode(string(raw), &doc); err != nil {
		return "", nil, Problem(fmt.Errorf("workspace file %s: %v", file, err))
	}
	for _, r := range doc.Repo {
		if r.Path == "" {
			return "", nil, Problem(fmt.Errorf("workspace file %s: a [[repo]] has no path", file))
		}
	}
	abs, err := filepath.Abs(filepath.Dir(file))
	return abs, doc.Repo, err
}

// eachWorkspaceRepo calls fn for every listed repository that exists and is a
// git repository, in file order. Each one that is not gets a line on stderr;
// the returned error then makes the command exit 1 without a second message.
func eachWorkspaceRepo(g *Globals, stderr io.Writer, file string, args []string, fn func(root string, r wsRepo) error) error {
	if len(args) > 0 {
		return errors.New("--workspace takes no <brief> arguments")
	}
	dir, repos, err := loadWorkspace(g, file)
	if err != nil {
		return err
	}
	missing := false
	for _, r := range repos {
		root := r.Path
		if !filepath.IsAbs(root) {
			root = filepath.Join(dir, root)
		}
		if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
			fmt.Fprintf(stderr, "vloop: workspace repo not found: %s\n", r.Path)
			missing = true
			continue
		}
		if err := fn(root, r); err != nil {
			return err
		}
	}
	if missing {
		return Problem(errors.New(""))
	}
	return nil
}

func workspaceMetrics(g *Globals, out, stderr io.Writer, file string, args []string) error {
	var names []string
	var reports []*metrics.Report
	err := eachWorkspaceRepo(g, stderr, file, args, func(root string, r wsRepo) error {
		c, err := newClassifier(g, out, root)
		if err != nil {
			return err
		}
		rs, err := allReports(root, c)
		if err != nil {
			return Problem(err)
		}
		name := r.Name
		if name == "" {
			name = repoIdentity(root).Name
		}
		for _, rep := range rs {
			names = append(names, name)
			reports = append(reports, rep)
		}
		return nil
	})
	var pe *ProblemError
	if err != nil && !(errors.As(err, &pe) && err.Error() == "") {
		return err
	}
	if g.JSON {
		if reports == nil {
			reports = []*metrics.Report{}
		}
		if e := json.NewEncoder(out).Encode(reports); e != nil {
			return e
		}
	} else {
		metrics.PrintWorkspaceTable(out, names, reports)
	}
	return err
}

func workspaceExport(g *Globals, out, stderr io.Writer, file string, args []string) error {
	var all string
	err := eachWorkspaceRepo(g, stderr, file, args, func(root string, r wsRepo) error {
		s, err := exportRepoLines(g, out, root, r.Name, nil)
		all += s
		return err
	})
	var pe *ProblemError
	if err != nil && !(errors.As(err, &pe) && err.Error() == "") {
		return err
	}
	fmt.Fprint(out, all)
	return err
}
