package driver

import (
	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/state"
)

// Source of a resolved model or effort that comes from the task itself.
const SourceTask = "task"

// ResolveSession resolves the model or effort (field "model" or "effort") for
// a session kind ("plan", "work", "review", "gate-review") by P-6: the task's own value, else
// the environment, else the config file, else the default. An unset effort
// resolves to "".
func ResolveSession(root, field, kind, taskValue string) (value, source string, err error) {
	if taskValue != "" {
		return taskValue, SourceTask, nil
	}
	v, err := config.Get(root, field+"."+kind)
	if err != nil {
		return "", "", err
	}
	if !v.Set {
		return "", v.Source, nil
	}
	return v.Value, v.Source, nil
}

// ModelEffort is the model and effort a session of the phase runs with. task
// is nil for a plan; a task's override applies to its own phase only, so a
// review never takes the work model.
func ModelEffort(root string, task *state.Task, phase string) (model, effort string, err error) {
	var tm, te string
	if task != nil {
		tm, te = sessionOverride(task.Model, phase), sessionOverride(task.Effort, phase)
	}
	if model, _, err = ResolveSession(root, "model", phase, tm); err != nil {
		return "", "", err
	}
	effort, _, err = ResolveSession(root, "effort", phase, te)
	return model, effort, err
}

func sessionOverride(s *state.Sessions, phase string) string {
	switch {
	case s == nil:
		return ""
	case phase == PhaseWork:
		return s.Work
	case phase == PhaseReview:
		return s.Review
	}
	return ""
}

// Resolved is the model and effort of each session kind, read from the config
// once at the start of a run: a session that rewrites the config afterwards
// changes nothing of the run in progress.
type Resolved struct {
	model, effort map[string]string
}

// ResolveRun resolves the work and review defaults.
func ResolveRun(root string) (*Resolved, error) {
	r := &Resolved{model: map[string]string{}, effort: map[string]string{}}
	for _, phase := range []string{PhaseWork, PhaseReview} {
		var err error
		if r.model[phase], _, err = ResolveSession(root, "model", phase, ""); err != nil {
			return nil, err
		}
		if r.effort[phase], _, err = ResolveSession(root, "effort", phase, ""); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// For is ModelEffort over the values held: the task's override for its own
// phase, else the default resolved at the start.
func (r *Resolved) For(task *state.Task, phase string) (model, effort string) {
	model, effort = r.model[phase], r.effort[phase]
	if task != nil {
		if v := sessionOverride(task.Model, phase); v != "" {
			model = v
		}
		if v := sessionOverride(task.Effort, phase); v != "" {
			effort = v
		}
	}
	return model, effort
}
