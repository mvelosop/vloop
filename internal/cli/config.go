package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/copilot"
	"github.com/mvelosop/vloop/internal/gitrun"
)

// root resolves the repo root from -C or the working directory.
func (g *Globals) root() (string, error) {
	dir := g.Dir
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		dir = wd
	}
	return config.Root(dir), nil
}

// gitRoot is root for a command that needs git: it refuses a directory that is
// not inside a git repository in one line, instead of letting git's own error out.
func (g *Globals) gitRoot() (string, error) {
	root, err := g.root()
	if err != nil {
		return "", err
	}
	if err := gitrun.At(root, "rev-parse", "--git-dir").Run(); err != nil {
		return "", Problem(fmt.Errorf("%s is not a git repository", root))
	}
	return root, nil
}

// configErr maps config errors: bad source values are problems (exit 1), and
// under --json stdout still carries one JSON document.
func configErr(g *Globals, out io.Writer, err error) error {
	var se *config.SourceError
	if errors.As(err, &se) {
		if g.JSON {
			_ = json.NewEncoder(out).Encode(struct {
				Error string `json:"error"`
			}{err.Error()})
		}
		return Problem(err)
	}
	return err
}

func jsonValue(v config.Value) any {
	if k, err := config.Lookup(v.Key); err == nil && k.List {
		return v.List
	}
	if !v.Set {
		return nil
	}
	return v.Value
}

func newConfig(g *Globals) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Read and write repo-local configuration",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List every key with its value and source",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			root, err := g.root()
			if err != nil {
				return err
			}
			vals, err := config.List(root)
			if err != nil {
				return configErr(g, out, err)
			}
			if g.JSON {
				return writeOrdered(out, vals, func(v config.Value) (string, any) {
					return v.Key, struct {
						Value  any    `json:"value"`
						Source string `json:"source"`
					}{jsonValue(v), v.Source}
				})
			}
			for _, v := range vals {
				fmt.Fprintf(out, "%s=%s (%s)\n", v.Key, v.Value, v.Source)
			}
			checks, err := config.Checks(root)
			if err != nil {
				return configErr(g, out, err)
			}
			for _, c := range checks {
				fmt.Fprintf(out, "check.%s=%s (paths: %s)\n", c.Name, c.Run, strings.Join(c.Paths, ", "))
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "get <key>",
		Short: "Show the resolved value of a key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			root, err := g.root()
			if err != nil {
				return err
			}
			v, err := config.Get(root, args[0])
			if err != nil {
				return configErr(g, out, err)
			}
			if g.JSON {
				return json.NewEncoder(out).Encode(struct {
					Key    string `json:"key"`
					Value  any    `json:"value"`
					Source string `json:"source"`
				}{v.Key, jsonValue(v), v.Source})
			}
			fmt.Fprintln(out, v.Value)
			return nil
		},
	})

	var setLocal bool
	var setAgent, setModel, setEffort, setKinds string
	set := &cobra.Command{
		Use:   "set [<key> <value>]",
		Short: "Set any key config list shows, in .vloop/config.toml or with --local in .vloop/config.local.toml (not platforms, nor shell on a mixed team; '' removes it)",
		Args: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("kinds") {
				if len(args) != 0 {
					return Usage(errors.New("--kinds takes no key or value: use --agent, --model and --effort"))
				}
				return nil
			}
			if cmd.Flags().Changed("agent") || cmd.Flags().Changed("model") || cmd.Flags().Changed("effort") {
				return Usage(errors.New("--agent, --model and --effort require --kinds"))
			}
			return cobra.ExactArgs(2)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := g.root()
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("kinds") {
				assignments, err := configSetAssignments(setKinds, cmd, setAgent, setModel, setEffort)
				if err != nil {
					return err
				}
				names := make([]string, 0, len(assignments)*2)
				for _, assignment := range assignments {
					_, kind, _ := strings.Cut(assignment.Key, ".")
					names = append(names, "agent."+kind, "model."+kind)
				}
				values, err := config.PreviewSet(root, assignments, setLocal, names...)
				if err != nil {
					return configErr(g, cmd.OutOrStdout(), err)
				}
				pairs := make([]copilotConfigPair, 0, len(values)/2)
				seen := map[string]bool{}
				for _, assignment := range assignments {
					_, kind, _ := strings.Cut(assignment.Key, ".")
					if seen[kind] {
						continue
					}
					seen[kind] = true
					pairs = append(pairs, copilotConfigPair{
						kind:  kind,
						agent: values["agent."+kind],
						model: values["model."+kind],
					})
				}
				warnings, err := checkConfigSetCopilotPairs(pairs, cmd.Flags().Changed("agent") &&
					setAgent == "" && !cmd.Flags().Changed("model") && !cmd.Flags().Changed("effort"))
				if err != nil {
					return err
				}
				if setLocal {
					err = config.SetManyLocal(root, assignments)
				} else {
					err = config.SetMany(root, assignments)
				}
				if err != nil {
					return configErr(g, cmd.OutOrStdout(), err)
				}
				for _, warning := range warnings {
					fmt.Fprintln(cmd.ErrOrStderr(), warning)
				}
				for _, assignment := range assignments {
					printConfigSetOverride(cmd.ErrOrStderr(), root, assignment.Key, setLocal)
				}
				return nil
			}
			warning, err := checkConfigSetCopilotModel(root, args[0], args[1], setLocal)
			if err != nil {
				return configErr(g, cmd.OutOrStdout(), err)
			}
			write := config.Set
			if setLocal {
				write = config.SetLocal
			}
			if err := write(root, args[0], args[1]); err != nil {
				return configErr(g, cmd.OutOrStdout(), err)
			}
			if warning != "" {
				fmt.Fprintln(cmd.ErrOrStderr(), warning)
			}
			printConfigSetOverride(cmd.ErrOrStderr(), root, args[0], setLocal)
			return nil
		},
	}
	set.Flags().BoolVar(&setLocal, "local", false, "write "+config.LocalFilePath+", this machine's overrides, instead of the shared file")
	set.Flags().StringVar(&setKinds, "kinds", "", "set agent, model and effort for a comma-separated list of session kinds")
	set.Flags().StringVar(&setAgent, "agent", "", "agent for every kind named by --kinds")
	set.Flags().StringVar(&setModel, "model", "", "model for every kind named by --kinds")
	set.Flags().StringVar(&setEffort, "effort", "", "effort for every kind named by --kinds")
	cmd.AddCommand(set)

	var pathLocal bool
	path := &cobra.Command{
		Use:   "path",
		Short: "Show the config file path, relative to the repo root",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			p := config.FilePath
			if pathLocal {
				p = config.LocalFilePath
			}
			fmt.Fprintln(cmd.OutOrStdout(), p)
			return nil
		},
	}
	path.Flags().BoolVar(&pathLocal, "local", false, "show the local file's path instead")
	cmd.AddCommand(path)
	return cmd
}

var configSessionKinds = []string{"plan", "gate-review", "work", "review", "second-review", "verify", "review-brief"}

func configSetAssignments(kinds string, cmd *cobra.Command, agent, model, effort string) ([]config.Assignment, error) {
	changedAgent := cmd.Flags().Changed("agent")
	changedModel := cmd.Flags().Changed("model")
	changedEffort := cmd.Flags().Changed("effort")
	if !changedAgent && !changedModel && !changedEffort {
		return nil, Usage(errors.New("--kinds needs --agent, --model or --effort"))
	}
	parts := strings.Split(kinds, ",")
	names := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		kind := strings.TrimSpace(part)
		found := false
		for _, allowed := range configSessionKinds {
			if kind == allowed {
				found = true
				break
			}
		}
		if !found {
			return nil, Usage(fmt.Errorf("unknown session kind %s: want %s or %s",
				kind, strings.Join(configSessionKinds[:len(configSessionKinds)-1], ", "), configSessionKinds[len(configSessionKinds)-1]))
		}
		if !seen[kind] {
			seen[kind] = true
			names = append(names, kind)
		}
	}
	assignments := make([]config.Assignment, 0, len(names)*3)
	for _, kind := range names {
		if changedAgent {
			assignments = append(assignments, config.Assignment{Key: "agent." + kind, Value: agent})
		}
		if changedModel {
			assignments = append(assignments, config.Assignment{Key: "model." + kind, Value: model})
		}
		if changedEffort {
			assignments = append(assignments, config.Assignment{Key: "effort." + kind, Value: effort})
		}
	}
	return assignments, nil
}

func printConfigSetOverride(w io.Writer, root, key string, local bool) {
	if v, err := config.Get(root, key); err == nil {
		switch {
		case v.Source == config.SourceEnv:
			fmt.Fprintf(w, "! %s stays %s on this machine: %s overrides it\n", v.Key, v.Value, config.EnvVar(v.Key))
		case v.Source == config.SourceLocal && !local:
			fmt.Fprintf(w, "! %s stays %s on this machine: %s overrides it\n", v.Key, v.Value, config.LocalFilePath)
		}
	}
}

type copilotConfigPair struct {
	kind         string
	agent, model config.Value
}

func checkConfigSetCopilotPairs(pairs []copilotConfigPair, agentRemovalOnly bool) ([]string, error) {
	if agentRemovalOnly {
		return nil, nil
	}
	candidates := make([]copilotConfigPair, 0, len(pairs))
	seen := map[string]bool{}
	for _, pair := range pairs {
		if pair.agent.Value != "copilot" {
			continue
		}
		if isClaudeAlias(pair.model.Value) {
			return nil, Usage(fmt.Errorf("%s is a Claude alias, and agent.%s is copilot: vloop models copilot lists the Copilot ids", pair.model.Value, pair.kind))
		}
		key := pair.model.Value
		if !seen[key] {
			seen[key] = true
			candidates = append(candidates, pair)
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	client, err := copilot.New(5 * time.Second)
	if errors.Is(err, copilot.ErrNoLogin) {
		warnings := make([]string, 0, len(candidates))
		for _, pair := range candidates {
			warnings = append(warnings, fmt.Sprintf("! cannot check %s: no GitHub login reaches Copilot", pair.model.Value))
		}
		return warnings, nil
	}
	if err != nil {
		warnings := make([]string, 0, len(candidates))
		for _, pair := range candidates {
			warnings = append(warnings, fmt.Sprintf("! cannot check %s against Copilot's model list (%s); set anyway", pair.model.Value, err))
		}
		return warnings, nil
	}
	models, err := client.Models(context.Background())
	if err != nil {
		warnings := make([]string, 0, len(candidates))
		for _, pair := range candidates {
			warnings = append(warnings, fmt.Sprintf("! cannot check %s against Copilot's model list (%s); set anyway", pair.model.Value, err))
		}
		return warnings, nil
	}
	for _, pair := range candidates {
		found := false
		for _, listed := range models {
			if listed.ID != pair.model.Value {
				continue
			}
			found = true
			if listed.Policy != nil && *listed.Policy == "disabled" {
				return nil, Usage(fmt.Errorf("%s is disabled by this account's Copilot policy: vloop models copilot lists the enabled ones", pair.model.Value))
			}
			break
		}
		if !found {
			return nil, Usage(fmt.Errorf("%s is not a Copilot model this account lists: vloop models copilot lists them", pair.model.Value))
		}
	}
	return nil, nil
}

func checkConfigSetCopilotModel(root, key, value string, local bool) (string, error) {
	kind, ok := strings.CutPrefix(key, "model.")
	if !ok || value == "" {
		return "", nil
	}
	agentKey := "agent." + kind
	if _, err := config.Lookup(agentKey); err != nil {
		return "", nil
	}
	if err := config.Check(root, key, value); err != nil {
		return "", err
	}
	agent, err := config.Get(root, agentKey)
	if err != nil {
		return "", err
	}
	if agent.Value != "copilot" {
		return "", nil
	}
	model, err := configValueAfterSet(root, key, value, local)
	if err != nil {
		return "", err
	}
	if isClaudeAlias(model.Value) {
		return "", Usage(fmt.Errorf("%s is a Claude alias, and %s is copilot: vloop models copilot lists the Copilot ids", model.Value, agentKey))
	}
	client, err := copilot.New(5 * time.Second)
	if errors.Is(err, copilot.ErrNoLogin) {
		return fmt.Sprintf("! cannot check %s: no GitHub login reaches Copilot", model.Value), nil
	}
	if err != nil {
		return fmt.Sprintf("! cannot check %s against Copilot's model list (%s); set anyway", model.Value, err), nil
	}
	models, err := client.Models(context.Background())
	if err != nil {
		return fmt.Sprintf("! cannot check %s against Copilot's model list (%s); set anyway", model.Value, err), nil
	}
	for _, listed := range models {
		if listed.ID != model.Value {
			continue
		}
		if listed.Policy != nil && *listed.Policy == "disabled" {
			return "", Usage(fmt.Errorf("%s is disabled by this account's Copilot policy: vloop models copilot lists the enabled ones", model.Value))
		}
		return "", nil
	}
	return "", Usage(fmt.Errorf("%s is not a Copilot model this account lists: vloop models copilot lists them", model.Value))
}

func configValueAfterSet(root, key, value string, local bool) (config.Value, error) {
	current, err := config.Get(root, key)
	if err != nil {
		return config.Value{}, err
	}
	if current.Source == config.SourceEnv || current.Source == config.SourceLocal && !local {
		return current, nil
	}
	return config.Value{Key: key, Value: value, Set: true}, nil
}

func isClaudeAlias(model string) bool {
	switch model {
	case "opus", "sonnet", "haiku", "fable":
		return true
	default:
		return false
	}
}

// writeOrdered emits a JSON object whose keys keep the order of vals.
func writeOrdered[T any](w io.Writer, vals []T, kv func(T) (string, any)) error {
	buf := []byte{'{'}
	for i, v := range vals {
		k, x := kv(v)
		kb, _ := json.Marshal(k)
		xb, err := json.Marshal(x)
		if err != nil {
			return err
		}
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(append(append(buf, kb...), ':'), xb...)
	}
	buf = append(buf, '}', '\n')
	_, err := w.Write(buf)
	return err
}
