package budget

import "testing"

func TestGuidelineRulesAcceptsBothIDForms(t *testing.T) {
	text := "- old (D20250101-0900-old)\n- new (I20261014-113845-new)\n- none\n"
	rules, noID := GuidelineRules(text)
	if len(rules) != 3 || len(noID) != 1 || noID[0] != 3 {
		t.Fatalf("rules %v noID %v", rules, noID)
	}
}
