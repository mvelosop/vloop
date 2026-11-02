#!/bin/sh
# Gate T5 — the docs (A6). Judged on the documents the brief names: the tool's
# README, the section "## Other agents" in docs/guide/evals.md between
# "## Running" and "## vloop's suites", and the sentence in the design note.
# Prose is joined (newlines to spaces, whitespace collapsed, backticks and
# asterisks dropped) before it is searched.
set -u
export LC_ALL=C
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") || fail mktemp
trap 'rm -rf "$t"' EXIT
joined() { tr -d '`*' <"$1" | tr -s ' \t\n' '   '; }

# --- tools/agent-eval/README.md
RM=tools/agent-eval/README.md
[ -f "$RM" ] || fail "$RM is missing"
joined "$RM" >"$t/readme"
for s in --agent --model --out --cases --case --runs --judge-model --timeout --vloop --list \
  report.json trace.jsonl transcript.jsonl 'claude plugin eval' allowed_tools max_turns; do
  grep -qF -- "$s" "$t/readme" || fail "$RM does not mention $s"
done
grep -qi 'skip' "$t/readme" || fail "$RM does not say which cases are skipped"
grep -qi 'judge' "$t/readme" || fail "$RM does not say who judges the llm graders"
grep -qi 'fence' "$t/readme" || fail "$RM does not say the phase's fence replaces allowed_tools and max_turns"

# --- docs/guide/evals.md: ## Other agents, after ## Running and before ## vloop's suites
E=docs/guide/evals.md
hr=$(grep -n '^## Running$' "$E" | head -n 1 | cut -d: -f1)
ho=$(grep -n '^## Other agents$' "$E" | head -n 1 | cut -d: -f1)
hs=$(grep -n "^## vloop's suites\$" "$E" | head -n 1 | cut -d: -f1)
[ -n "$ho" ] || fail "$E has no section '## Other agents'"
[ -n "$hr" ] && [ -n "$hs" ] || fail "$E lost '## Running' or '## vloop's suites'"
[ "$hr" -lt "$ho" ] && [ "$ho" -lt "$hs" ] || fail "'## Other agents' is not after '## Running' and before '## vloop's suites'"
awk -v a="$ho" -v b="$hs" 'NR > a && NR < b' "$E" >"$t/sec.md"
grep -q '^## ' "$t/sec.md" && fail "'## Other agents' is not immediately followed by '## vloop's suites': another section sits between them"
joined "$t/sec.md" >"$t/sec"
grep -qF 'go run ./tools/agent-eval' "$t/sec" || fail "'## Other agents' does not give the command go run ./tools/agent-eval"
grep -qF 'tools/agent-eval/README.md' "$t/sec" || fail "'## Other agents' does not point to tools/agent-eval/README.md"
tr 'A-Z' 'a-z' <"$t/sec" >"$t/secl"
for s in 'a model qualifies for a phase on an agent' 'at 3 runs' 'every case of that phase' 'no lower than' 'default claude model' 'same tool' 'same runs'; do
  grep -qF -- "$s" "$t/secl" || fail "'## Other agents' does not state the qualifying rule: missing '$s'"
done

# --- the design note's sentence, after "another agent needs it first." and before "The fence audit is worth doing"
N=docs/design-notes/multi-agent-backends.md
joined "$N" >"$t/note"
grep -qF 'another agent needs it first. It is tools/agent-eval since B33. The fence audit is worth doing' "$t/note" ||
  fail "$N does not read 'another agent needs it first. It is tools/agent-eval since B33. The fence audit is worth doing'"
echo "gate T5: ok"
