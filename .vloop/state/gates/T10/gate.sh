#!/bin/sh
# Gate T10 — D1: the domain, the guides and both operate playbooks describe
# the verify session, its report, the clearing of .vloop/tmp/, the three
# config keys and V3's metrics; the retired brief-check warnings and
# guidelines sections are gone from the brief page and the concepts guide; the
# operate skill stays within 1,750 words; docs/guide/commands.md is what
# cmd/gendocs makes now.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
has() { grep -Fq -- "$2" "$1" || fail "$1 does not mention '$2'"; }
hasnt() { grep -Fq -- "$2" "$1" && fail "$1 still has '$2'"; return 0; }
# section <file> <start regex> <end regex>: from the start line up to, not
# including, the next line matching the end.
section() { awk -v s="$2" -v e="$3" 'p && $0 ~ e { exit } $0 ~ s { p = 1 } p' "$1"; }

F=docs/domain/execution/session.md
grep -Eq '^\| \*\*verify\*\* \|' "$F" || fail "$F's kinds table has no verify row"
grep -qi 'five kinds' "$F" || fail "$F does not say there are five kinds"
grep -qi 'the four kinds' "$F" && fail "$F still says four kinds"
has "$F" verify/v1
has "$F" .vloop/tmp/verify.json
section "$F" '^## What a session never does' '^## The record' | grep -q 'verify' || fail "$F: the fence paragraph does not cover the verify phase"

F=docs/domain/execution/run.md
has "$F" reports/verify.json
grep -q 'sessions/NNN-<phase>.json' "$F" || fail "$F's run folder lost sessions/NNN-<phase>.json"
grep -qi 'verify session' "$F" || fail "$F does not describe the verify session"
grep -qi 'final pass' "$F" || fail "$F does not place the verify session after the final pass"
grep -i 'clear' "$F" | grep -q '\.vloop/tmp/' || fail "$F does not say when .vloop/tmp/ is cleared"
grep -Eqi '\.running|lock' "$F" || fail "$F does not say the run's lock survives the clearing"

F=docs/domain/domain-model.md
section "$F" '^## On disk' '^## Invariants' >"$t/od"
grep -q '\.vloop/tmp/verify\.json' "$t/od" || fail "$F's On disk list has no .vloop/tmp/verify.json"
grep -qi 'clear' "$t/od" || fail "$F's On disk does not say .vloop/tmp/ is cleared between sessions"

for F in docs/domain/measurement/metrics.md docs/guide/metrics.md; do
  has "$F" verify_ms
  has "$F" verify_usd
done
grep -q ' · verify <' docs/guide/metrics.md || fail "docs/guide/metrics.md's summary lines show no ' · verify <…>' part"

for F in docs/domain/briefing/brief.md docs/guide/concepts.md; do
  hasnt "$F" 'Pinning tests'
  hasnt "$F" 'External formats'
  hasnt "$F" 'must-change list'
  hasnt "$F" 'external format:'
done

F=docs/guide/configuration.md
for k in run.verify model.verify effort.verify; do has "$F" "$k"; done

O=plugin/skills/operate/SKILL.md
grep -qi 'verify' "$O" || fail "$O does not mention the verify findings"
grep -Eq 'reports/verify\.json|verify findings|verify of ' "$O" || fail "$O's close step does not read the verify findings"
grep -Eq -- '--finding "<[^"]*>: <[^"]*>"' "$O" || fail "$O does not record an accepted finding as --finding \"<origin>: <summary>\""
[ "$(wc -w <"$O" | tr -d ' ')" -le 1750 ] || fail "$O is over its ceiling of 1,750 words"

P=.claude/skills/vloop-operator/SKILL.md
section "$P" '^## 5[.] Closing and merging' '^## [^5]' >"$t/p5"
[ -s "$t/p5" ] || fail "$P has no '## 5. Closing and merging'"
grep -qi 'verify' "$t/p5" || fail "$P step 5 does not read the verify findings"
grep -q -- '--finding' "$t/p5" || fail "$P step 5 does not record an accepted finding with --finding"

C=docs/guide/commands.md
GOFLAGS= go run ./cmd/gendocs "$t/commands.md" >"$t/gen.out" 2>&1 || fail "go run ./cmd/gendocs: $(cat "$t/gen.out")"
cmp -s "$t/commands.md" "$C" || fail "$C is not what cmd/gendocs makes now: regenerate it"
echo "gate T10: ok"
