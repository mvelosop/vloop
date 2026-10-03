package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/state"
)

func newTaskGate(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "gate <id>",
		Short: "Run a task's verify command in the plan's shell",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			p, err := loadPlan(g, out)
			if err != nil {
				return err
			}
			if want := os.Getenv("VLOOP_PLAN_SHA256"); want != "" {
				if err := checkPlanHash(g, want); err != nil {
					return jsonProblem(g, out, err)
				}
			}
			t := p.Find(args[0])
			if t == nil {
				return jsonProblem(g, out, &state.NoTaskError{ID: args[0]})
			}
			root, err := g.root()
			if err != nil {
				return err
			}
			gateOut := out
			if g.JSON {
				gateOut = cmd.ErrOrStderr()
			}
			minutes, err := config.Get(root, "run.gate-timeout")
			if err != nil {
				return configErr(g, out, err)
			}
			n, _ := strconv.Atoi(minutes.Value)
			timeout := time.Duration(n) * time.Minute
			code, d, timedOut, err := state.RunGateWithin(root, p.Shell, t.Verify, gateOut, cmd.ErrOrStderr(), timeout)
			if err != nil {
				return jsonProblem(g, out, err)
			}
			if timedOut {
				if code == 0 {
					code = 1
				}
				fmt.Fprintln(cmd.ErrOrStderr(), state.GateTimedOutLine(t.ID, n))
			}
			if g.JSON {
				b, err := json.Marshal(struct {
					Task       string `json:"task"`
					Passed     bool   `json:"passed"`
					Exit       int    `json:"exit"`
					DurationMS int64  `json:"duration_ms"`
				}{t.ID, code == 0, code, d.Milliseconds()})
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s\n", b)
			} else if code == 0 {
				fmt.Fprintf(out, "gate %s: pass (%s)\n", t.ID, d.Round(time.Millisecond))
			} else {
				fmt.Fprintf(out, "gate %s: fail (exit %d, %s)\n", t.ID, code, d.Round(time.Millisecond))
			}
			if code != 0 {
				return Problem(errors.New(""))
			}
			return nil
		},
	}
}

func newTaskVerify(g *Globals) *cobra.Command {
	var reason string
	cmd := &cobra.Command{
		Use:   "verify <id> [<command>]",
		Short: "Record a change to a task's gate: a new verify command, a changed gate folder, or both",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := g.root()
			if err != nil {
				return err
			}
			command := ""
			if len(args) == 2 {
				if args[1] == "" {
					return Problem(errors.New("verify command is empty"))
				}
				command = args[1]
			}
			return amend(g, cmd, func(p *state.Plan) error {
				return state.RecordGate(root, p, args[0], command, reason, time.Now())
			})
		},
	}
	cmd.Flags().StringVar(&reason, "reason", "", "why the gate is being replaced (required)")
	_ = cmd.MarkFlagRequired("reason")
	return cmd
}

// checkPlanHash refuses to run a gate when the plan on disk is not the one the
// driver handed this session, and leaves the driver a note that it did.
func checkPlanHash(g *Globals, want string) error {
	root, err := g.root()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(state.Path(root))
	if err != nil {
		return Problem(err)
	}
	gates, err := state.GateFiles(root)
	if err != nil {
		return Problem(err)
	}
	if state.PlanDigest(data, gates) == strings.ToLower(strings.TrimSpace(want)) {
		return nil
	}
	note := filepath.Join(root, ".vloop", "tmp", "gate-refused")
	if os.MkdirAll(filepath.Dir(note), 0o755) == nil {
		_ = os.WriteFile(note, []byte("plan changed\n"), 0o644)
	}
	return Problem(errors.New("the plan was changed during this session — gates run only from the plan the driver holds"))
}
