#!/bin/sh
# Gate T3 — F3 and F4. F3 through the product: a `vloop run` with every phase
# on the stub copilot (run.second-review = "record") and a `vloop brief
# review`, each session running only calls its fence allows; every hooks file
# the stub copilot reads has timeoutSec 120, both commands end as at the base
# (exit 0), and nothing is named. F4: every edited passage, each file's lines
# joined and whitespace collapsed before the match (so any wrapping passes),
# and `vloop run --help` and `vloop brief review --help` read the same way.
G=.vloop/state/gates/T3
. "$G/lib.sh"
newbin

# F3.
CFG="run.second-review record"; fixture
for p in plan gate-review work review second-review verify review-brief; do
  { start c1 bash "$(bashargs 'git status')"; done_ c1; start c2 view "$(fileargs docs/x.md)"; done_ c2; } >"$t/stub/calls.$p"
done
runfix ok
[ "$rc" = 0 ] || fail "a run whose sessions ran only allowed calls: want exit 0, got $(show ok)"
brev okb
[ "$rc" = 0 ] || fail "a brief review whose session ran only allowed calls: want exit 0, got $(show okb)"
for n in ok okb; do
  grep -qE 'FENCE BYPASSED|HOOK CHANGED' "$t/$n.err" && fail "sessions that ran only allowed calls: a halt was printed: $(grep -E 'FENCE BYPASSED|HOOK CHANGED' "$t/$n.err")"
  grep -qE '^vloop: (bash|view) ' "$t/$n.err" && fail "sessions that ran only allowed calls: a call was named: $(grep -E '^vloop: (bash|view) ' "$t/$n.err")"
done
cat "$t/ok.ev" "$t/okb.ev" >"$t/all.ev"
for p in plan gate-review work review second-review verify review-brief; do
  got=$(sed -n "s/^timeout $p //p" "$t/all.ev" | sort -u | tr '\n' ' ')
  [ "$got" = "120 " ] || fail "the $p session's hooks file has timeoutSec '${got:-no session}', want 120"
done

# F4. flat <file>: the file on one line, whitespace collapsed.
flat() { tr '\n\t' '  ' <"$1" | tr -s ' '; }
# has <file> <text>: the flattened file contains the text.
has() { flat "$1" | grep -qF -- "$2" || fail "$1 lacks: $2"; }
# row <file> <first cells> <text>: the table row beginning with the cells contains the text.
row() { grep -F -- "$2" "$1" | head -n 1 | tr -s ' \t' '  ' | grep -qF -- "$3" || fail "$1: the row '$2' lacks: $3"; }

D=docs/domain/domain-model.md
has $D "and restores what a session changed among the driver's inputs. For a Copilot session it also halts (exit 9) when a call that ran is one the fence denies, or when the binary the fence hook runs changed."
has $D "8 repeat blocked, 9 refs or repository configuration moved, or a Copilot session got past its fence. Exit 2 also covers"
has docs/domain/execution/session.md "halts the run with exit 9 and nothing is committed. For a Copilot session the driver also halts (exit 9) when a call that ran is one the fence denies, or when the binary the fence hook runs changed. A session writes nothing"
C=docs/guide/concepts.md
row $C '| `9` |' '| `9` | refs or repository configuration moved, or a Copilot session got past its fence: a session or a gate changed a ref,'
row $C '| `9` |' 'or the git hooks, or a Copilot session ran a call its fence denies (FENCE BYPASSED, the calls named) or the binary its fence hook runs changed (HOOK CHANGED); nothing was committed.'
row $C '| `9` |' '; after FENCE BYPASSED or HOOK CHANGED, inspect the work tree and the named calls first |'
has docs/guide/configuration.md "B22.2 and B22.3). The driver backs it: after every Copilot session it judges again every call that ran, and halts with exit 9 (FENCE BYPASSED) when the fence denies one, or (HOOK CHANGED) when the hook's binary changed."
row docs/domain/execution/run.md '| 9 |' '| 9 | refs or repository configuration moved, or a Copilot session got past its fence; nothing committed |'
has docs/design-notes/copilot-cli-experiments.md "it is not probed. Since B34 the driver re-judges every call a Copilot session ran and halts on one the fence denies, checks the hook's binary before and after each session, and gives the hook a timeoutSec of 120."
O=plugin/skills/operate/SKILL.md
row $O '| 9 |' '| 9 | a session moved git refs, or got past its fence |'
row $O '| 9 |' 'For FENCE BYPASSED or HOOK CHANGED, inspect the work tree and the named calls before anything else; do not resume until the operator decides. |'

# The fence audit: six headings, each section with the note on a line of its own.
A=docs/design-notes/fence-audit.md
NOTE="*Since B34 the driver re-judges every call a Copilot session ran and halts (exit 9, FENCE BYPASSED) on one the fence denies, and (HOOK CHANGED) when the hook's binary changed: where this audit says a call the hook fails open on goes unnoticed, that predates it.*"
[ "$(grep -c '^# Fence audit: ' $A)" = 6 ] || fail "$A: want its six '# Fence audit: …' headings"
awk -v d="$t/audit" '/^# Fence audit: /{ n++ } n { print > (d "." n) }' $A || fail "split $A"
i=1
while [ "$i" -le 6 ]; do
  grep -q '^\*Since B34 the driver re-judges' "$t/audit.$i" && flat "$t/audit.$i" | grep -qF -- "$NOTE" ||
    fail "$A: the section under '$(head -n 1 "$t/audit.$i")' lacks the B34 note on a line of its own"
  i=$((i + 1))
done

# The commands' help.
(cd "$t/repo" && venv "$S" run --help >"$t/run.help" 2>&1) || fail "vloop run --help failed"
has "$t/run.help" "9 refs or repository configuration moved, or a Copilot session got past its fence."
(cd "$t/repo" && venv "$S" brief review --help >"$t/br.help" 2>&1) || fail "vloop brief review --help failed"
has "$t/br.help" "9 the session changed the brief, git refs or the repository configuration, or a Copilot session got past its fence."
echo "gate T3: ok"
