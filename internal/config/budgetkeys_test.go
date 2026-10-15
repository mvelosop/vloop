package config

import (
	"reflect"
	"testing"
)

func TestBudgetKeysAreConfigKeys(t *testing.T) {
	known := map[string]bool{}
	for _, k := range Keys {
		known[k.Name] = true
	}
	v := reflect.ValueOf(BudgetKeys)
	if v.NumField() != 9 {
		t.Fatalf("BudgetKeys names %d keys, want 9", v.NumField())
	}
	for i := 0; i < v.NumField(); i++ {
		if k := v.Field(i).String(); !known[k] {
			t.Errorf("BudgetKeys.%s = %q is not in Keys", v.Type().Field(i).Name, k)
		}
	}
}
