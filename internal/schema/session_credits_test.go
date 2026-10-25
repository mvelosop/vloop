package schema

import (
	"strings"
	"testing"
)

const sessionV3 = `{"schema":"session/v3","run_id":"r","iteration":1,"phase":"review","agent":"AGENT","task":"T1","model":"m","effort":null,
"models_used":{},"started":"2026-01-01T00:00:00Z","ended":"2026-01-01T00:00:01Z","duration_ms":1,"wall_ms":1,
"cost_usd":0.1,"cost_source":"SOURCE","turns":1,"is_error":false,"permission_denials":[]EXTRA}`

func session3(agent, source, extra string) []byte {
	return []byte(strings.NewReplacer("AGENT", agent, "SOURCE", source, "EXTRA", extra).Replace(sessionV3))
}

// session/v3 gains credits and the estimated cost source for other agents than
// claude; a base-written claude record still validates.
func TestSessionV3Credits(t *testing.T) {
	for _, c := range []struct {
		name, agent, source, extra string
		valid                      bool
	}{
		{"copilot estimated with credits", "copilot", "estimated", `,"credits":14.2`, true},
		{"claude reported, no credits", "claude", "reported", "", true},
		{"claude none", "claude", "none", "", true},
		{"claude estimated", "claude", "estimated", "", false},
		{"credits as a string", "copilot", "estimated", `,"credits":"14.2"`, false},
		{"unknown source", "copilot", "bogus", "", false},
	} {
		v, err := Validate("session/v3", session3(c.agent, c.source, c.extra))
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if (len(v) == 0) != c.valid {
			t.Errorf("%s: valid %v, violations %v", c.name, len(v) == 0, v)
		}
	}
}
