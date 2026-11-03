# Journal — B20261103-0315-b35-i83-fence-compound-commands

Append-only narrative of this plan. Rendered state lives in .vloop/state/plan.md.

## Plan — B20261103-0315-b35-i83-fence-compound-commands

- **Brief:** `docs/briefs/B20261103-0315-b35-i83-fence-compound-commands.loop-brief.md`
- **Tasks:** 4

## T1 — Judge compound commands through, and let command -v and -V run nothing (C1, C3)

- **Outcome:** done (review: PASS)
- **Summary:** The splitter opens and closes subshells and the judge drops opening keywords, allows lone fi/done/}, reads for NAME [in …], and treats literal command -v/-V as a lookup; every inner command is judged as alone.
- **Files:** internal/fence/split.go, internal/fence/judge.go, internal/fence/judge_test.go
- **Notes for next iteration:** command now carries plain []bool (word written with no quote/escape) parallel to words, passed to Judge.words; C2 can use it. Existing tests changed per Constraints: TestSplitRefuses (gh x) -> (gh x and a=(1 2); work table for f in a b; do echo $f; done now allowed. New TestCompoundCommands holds the C1/C3 worked-example lines. go test ./... run inside this session fails 5 TestTaskGate* tests in internal/cli with 'the plan was changed during this session' (the live driver's plan guard, not the fence); I did not touch that code and did not confirm they pass outside a session.
