package driver

import (
	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/state"
)

// Source of a resolved model or effort that comes from the task itself.
const SourceTask = "task"

// ResolveSession resolves the model or effort (field "model" or "effort") for
// a session kind ("plan", "work", "review") by P-6: the task's own value, else
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
