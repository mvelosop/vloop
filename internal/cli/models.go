package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/copilot"
)

const modelsUsage = "models lists Copilot models only: vloop models copilot"

type copilotModelsDocument struct {
	Schema string          `json:"schema"`
	Plan   *string         `json:"plan"`
	Models []copilot.Model `json:"models"`
}

func newModels(g *Globals) *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "models copilot",
		Short: "List the Copilot model ids available to this account",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 || args[0] != "copilot" {
				return Usage(errors.New(modelsUsage))
			}
			client, err := copilot.New(20 * time.Second)
			if err != nil {
				return Problem(errors.New("no GitHub login reaches Copilot: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login"))
			}
			models, err := client.Models(context.Background())
			if err != nil {
				return Problem(fmt.Errorf("cannot list Copilot models: %w", err))
			}
			plan, _ := client.Plan(context.Background())
			models = visibleModels(models, all)
			sort.Slice(models, func(i, j int) bool { return models[i].ID < models[j].ID })
			if g.JSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(copilotModelsDocument{
					Schema: "copilot-models/v1",
					Plan:   plan,
					Models: models,
				})
			}
			return writeCopilotModels(cmd, plan, models)
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "include models not shown in Copilot's model picker")
	return cmd
}

func visibleModels(models []copilot.Model, all bool) []copilot.Model {
	if all {
		return models
	}
	visible := make([]copilot.Model, 0, len(models))
	for _, model := range models {
		if model.Picker {
			visible = append(visible, model)
		}
	}
	return visible
}

func writeCopilotModels(cmd *cobra.Command, plan *string, models []copilot.Model) error {
	planText := "unknown"
	if plan != nil && *plan != "" {
		planText = *plan
	}
	out := cmd.OutOrStdout()
	if _, err := fmt.Fprintf(out, "plan  %s\n", planText); err != nil {
		return err
	}
	width := 0
	for _, model := range models {
		if len(model.ID) > width {
			width = len(model.ID)
		}
	}
	for _, model := range models {
		policy := "-"
		if model.Policy != nil {
			policy = *model.Policy
		}
		if _, err := fmt.Fprintf(out, "%s%s%s\n", model.ID, strings.Repeat(" ", width-len(model.ID)+2), policy); err != nil {
			return err
		}
	}
	return nil
}
