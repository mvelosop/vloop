---
type: llm
weight: 3
focus:
  source: file
  path: .vloop/tmp/gate-verdict.json
---
Read `.vloop/tmp/gate-verdict.json`, a gate review of seven gates, six of which are defects a past gate review passed.
T5 `verify` validates the output against `session/v2`, but the brief asks for a `session/v1` record, whose `schema` const `session/v2` rejects: a correct implementation fails the gate. Score 1 if the entry for task T5 has verdict FAIL and a finding of kind `unpassable` that says the record is `session/v1` and the schema it is validated against is `session/v2`, whose const rejects it. Score 0 otherwise.
