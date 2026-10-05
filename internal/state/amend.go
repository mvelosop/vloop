package state

import (
	"errors"
	"fmt"
)

// ErrInvalidPlan is returned by Amend when the plan on disk already fails its
// schema; nothing is written.
var ErrInvalidPlan = errors.New("plan is not valid — run vloop task validate")

// NoTaskError is returned when a task id is not in the plan.
type NoTaskError struct{ ID string }

func (e *NoTaskError) Error() string {
	return "no task " + e.ID + " — vloop task list shows the plan's tasks"
}

// Find returns the task with the given id, or nil.
func (p *Plan) Find(id string) *Task {
	for i := range p.Tasks {
		if p.Tasks[i].ID == id {
			return &p.Tasks[i]
		}
	}
	return nil
}

// Amend loads the plan under root, applies change, and saves the result. It
// refuses (writing nothing) a plan that already fails its schema, and a result
// that would fail task validate; areas is the configured area list.
func Amend(root string, areas []string, change func(*Plan) error) error {
	violations, err := Validate(root)
	if err != nil {
		return err
	}
	if len(violations) > 0 {
		return ErrInvalidPlan
	}
	p, err := Load(root)
	if err != nil {
		return err
	}
	if err := change(p); err != nil {
		return err
	}
	data, err := Marshal(p)
	if err != nil {
		return err
	}
	rep, err := CheckBytes(root, data, areas)
	if err != nil {
		return err
	}
	if len(rep.Problems) > 0 {
		return fmt.Errorf("refusing to write a plan that fails task validate: %s", rep.Problems[0])
	}
	return Save(root, p)
}

// Reset sets a task pending with no attempts.
func Reset(p *Plan, id string) error {
	t := p.Find(id)
	if t == nil {
		return &NoTaskError{id}
	}
	t.Status, t.Attempts = "pending", 0
	return nil
}

// SetNote replaces a task's notes.
func SetNote(p *Plan, id, text string) error {
	t := p.Find(id)
	if t == nil {
		return &NoTaskError{id}
	}
	t.Notes = text
	return nil
}

// Drop removes a task, refusing while another task depends on it.
func Drop(p *Plan, id string) error {
	if p.Find(id) == nil {
		return &NoTaskError{id}
	}
	kept := make([]Task, 0, len(p.Tasks))
	for _, t := range p.Tasks {
		for _, d := range t.DependsOn {
			if d == id && t.ID != id {
				return fmt.Errorf("%s depends on %s", t.ID, id)
			}
		}
		if t.ID != id {
			kept = append(kept, t)
		}
	}
	p.Tasks = kept
	return nil
}

// SetField sets, or with "" clears, area, kind, model.work, model.review,
// effort.work or effort.review. The value is not validated here.
func SetField(p *Plan, id, field, value string) error {
	t := p.Find(id)
	if t == nil {
		return &NoTaskError{id}
	}
	switch field {
	case "area":
		t.Area = value
	case "kind":
		t.Kind = value
	case "model.work":
		t.Model = setSession(t.Model, "work", value)
	case "model.review":
		t.Model = setSession(t.Model, "review", value)
	case "effort.work":
		t.Effort = setSession(t.Effort, "work", value)
	case "effort.review":
		t.Effort = setSession(t.Effort, "review", value)
	default:
		return fmt.Errorf("unknown field %q", field)
	}
	return nil
}

// setSession sets one key of s, returning nil when nothing is left.
func setSession(s *Sessions, kind, value string) *Sessions {
	if s == nil {
		s = &Sessions{}
	}
	if kind == "work" {
		s.Work = value
	} else {
		s.Review = value
	}
	if *s == (Sessions{}) {
		return nil
	}
	return s
}
