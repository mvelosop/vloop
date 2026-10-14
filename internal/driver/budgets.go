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
	GateTimeout    int  // minutes a gate may run
	SessionTimeout int  // minutes a session may run
	KeepAwake      bool // hold the machine awake for the run
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
	integer(config.BudgetKeys.MaxIterations, &b.MaxIterations)
	num(config.BudgetKeys.CostCeiling, &b.CostCeiling)
	integer(config.BudgetKeys.MaxAttempts, &b.MaxAttempts)
	integer(config.BudgetKeys.StallLimit, &b.StallLimit)
	num(config.BudgetKeys.ConvergenceMax, &b.ConvergenceMax)
	integer(config.BudgetKeys.ConvergenceMin, &b.ConvergenceMin)
	integer(config.BudgetKeys.GateTimeout, &b.GateTimeout)
	integer(config.BudgetKeys.SessionTimeout, &b.SessionTimeout)
	if err == nil {
		var s string
		if s, err = val(config.BudgetKeys.KeepAwake); err == nil {
			b.KeepAwake = s == "on"
		}
	}
	return b, err
}
