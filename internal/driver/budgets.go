package driver

import (
	"strconv"

	"github.com/mvelosop/vloop/internal/config"
)

// Budgets bound one run. They are checked between iterations: a run always
// stops with the plan coherent, so raising one and re-running resumes.
type Budgets struct {
	MaxIterations  int
	CostCeiling    float64 // dollars
	MaxAttempts    int
	StallLimit     int
	ConvergenceMax float64
	ConvergenceMin int
}

// ResolveBudgets resolves each budget by flag, then VLOOP_RUN_*, then the
// config file, then the default. over holds the flags that were given, keyed
// by config key, as typed; each must already pass config.Check. A bad
// environment or file value is a *config.SourceError.
func ResolveBudgets(root string, over map[string]string) (Budgets, error) {
	val := func(key string) (string, error) {
		if v, ok := over[key]; ok {
			return v, nil
		}
		got, err := config.Get(root, key)
		return got.Value, err
	}
	var b Budgets
	var err error
	num := func(key string, dst *float64) {
		if err != nil {
			return
		}
		var s string
		if s, err = val(key); err == nil {
			*dst, err = strconv.ParseFloat(s, 64)
		}
	}
	integer := func(key string, dst *int) {
		var f float64
		num(key, &f)
		*dst = int(f)
	}
	integer("run.max-iterations", &b.MaxIterations)
	num("run.cost-ceiling", &b.CostCeiling)
	integer("run.max-attempts", &b.MaxAttempts)
	integer("run.stall-limit", &b.StallLimit)
	num("run.convergence-max", &b.ConvergenceMax)
	integer("run.convergence-min", &b.ConvergenceMin)
	return b, err
}
