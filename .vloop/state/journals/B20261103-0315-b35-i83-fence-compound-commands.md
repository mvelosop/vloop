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

## T2 — Match a word the shell expands as any text (C2)

- **Outcome:** blocked (review: skipped)
- **Summary:** C2 is implemented: the lexer marks each word as expanded or literal, the judge refuses expanded words wherever it reads structure, xargs gets the appended or replace-string word, and match tries expanded words as any text after the as-written pass. Gate T2 passes 108 of 109 cases. The one miss is a gate case that contradicts acceptance 3.
- **Files:** internal/fence/split.go, internal/fence/judge.go, internal/fence/expand.go, internal/fence/judge_test.go
- **Notes for next iteration:** Per-byte quote state (stUnquoted, stDouble, stLiteral) in the lexer feeds expands(); command.exp runs parallel to words and plain. internal/fence/expand.go holds match (as written, then any-text), skipXargs, shellString, xargsCommand and the NFA intersection used to decide whether a rule and a command with wildcard words share a string. The xargs input is the word <xargs input>, left out of the message text. TestExpandedWords in judge_test.go holds every gate case plus `command -$v git push`; the `git remote $X` row there expects the as-written stem message, as acceptance 3 requires. No existing test changed in this task.

## Run ended — blocked

- **Run:** `B20261103-0315-b35-i83-fence-compound-commands` · 2 iteration(s) this run
- **Plan:** 1/4 done, 1 blocked

## T2 — Match a word the shell expands as any text (C2)

- **Outcome:** done (review: PASS)
- **Summary:** C2 is implemented: the lexer marks each word expanded or literal, the judge refuses expanded words wherever it reads structure (command name, wrapper options, -c string, redirection target, xargs input), and match retries with expanded words as any text after the as-written pass.
- **Files:** internal/fence/split.go, internal/fence/judge.go, internal/fence/expand.go, internal/fence/judge_test.go
- **Notes for next iteration:** Work was already on disk from the earlier blocked attempt; the replaced gate (git remote $X expects the Copilot stem message) now passes. Per-byte quote state in the lexer feeds expands(); internal/fence/expand.go holds match, skipXargs, shellString, xargsCommand and an NFA intersection deciding whether a rule and a command with wildcard words share a string. TestExpandedWords holds every gate case plus command -$v git push. No existing test changed.

## T3 — Document the wider reading in the guide, the fence audit and fence check's help (C4)

- **Outcome:** done (review: PASS)
- **Summary:** The configuration guide, the fence audit (hook bullet and work-session paragraph) and `vloop fence check --help` now describe the wider reading, with the brief's C4 text word for word.
- **Files:** docs/guide/configuration.md, docs/design-notes/fence-audit.md, internal/cli/fence.go
- **Notes for next iteration:** none

## T4 — Close the run: every gate at once, and the worked example line for line

- **Outcome:** done (review: PASS)
- **Summary:** Ran gates T1, T2 and T3 together on the final tree (all ok) and checked every worked-example line against the committed tests; no code changed.
- **Files:** none
- **Notes for next iteration:** Line-for-line: every C1, C2 and C3 worked-example command is a case in gate T1's cases.tsv (C1, C3 lines; command -$v git push is in T2's) or T2's cases.tsv (C2 lines), and a Go-string case in internal/fence/judge_test.go: TestCompoundCommands (C1 lines, including the (ls case as 'open-subshell'), TestExpandedWords (C2 lines, the for x in push line, git diff "$BASE" --stat, command -$v git push), and C3 command -v/-V, time -v, command gh pr list in the same tests (grep -F found each at least once). No gap found, no test added. The TestTaskGate* failures in internal/cli are the live driver's plan guard firing because the plan differs from the driver's in-session; not touched by this run.

## Run ended — complete

- **Run:** `B20261103-0315-b35-i83-fence-compound-commands` · 3 iteration(s) this run
- **Plan:** 4/4 done, 0 blocked
