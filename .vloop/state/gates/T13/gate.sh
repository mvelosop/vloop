#!/bin/sh
# Gate T13 — help (brief Q4): the three Shorts the brief quotes; `list`
# subcommands' Shorts begin "List " and `show` subcommands' begin "Show "; Long
# help for the root, run, brief, task, metrics, defect and intervention (the
# help text above "Usage:" is more than the Short); the root's says what a
# brief is and to start with vloop init and vloop brief new; run's says how to
# resume and where the exit codes are documented; and docs/guide/commands.md is
# what cmd/gendocs writes from the command tree.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR CLAUDE_PLUGIN_ROOT
export LC_ALL=C
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
go build -o "$t/gendocs" ./cmd/gendocs || fail "go build ./cmd/gendocs"
B="$t/vloop"
# shorts <command...>: "<name>\t<short>" for each entry of its Available Commands.
shorts() {
  "$B" "$@" --help 2>&1 | awk '/^Available Commands:/ { on = 1; next } on && /^[^ ]/ { on = 0 } on && NF { n = $1; sub(/^ +[^ ]+ +/, ""); print n "\t" $0 }'
}
# short <parent...> <name>: the Short of one subcommand, as its parent lists it.
short() { n=$1; shift; shorts "$@" | awk -F '\t' -v n="$n" '$1 == n { print $2 }'; }
# above <command...>: the help text above "Usage:", blank lines dropped.
above() { "$B" "$@" --help 2>&1 | awk '/^Usage:/ { exit } NF { print }'; }

[ "$(short run)" = "Plan a brief and work it, task by task, on a work branch" ] || fail "run's Short is '$(short run)'"
[ "$(short brief)" = "Write, check, list and close loop briefs" ] || fail "brief's Short is '$(short brief)'"
[ "$(short task)" = "Show, amend, gate and reset the plan's tasks" ] || fail "task's Short is '$(short task)'"

for g in brief config defect intervention schema task; do
  s=$(short list "$g")
  [ -n "$s" ] || fail "vloop $g has no list subcommand in its help"
  case "$s" in "List "*) ;; *) fail "vloop $g list's Short does not begin 'List ': '$s'" ;; esac
done
for g in intervention schema task; do
  s=$(short show "$g")
  [ -n "$s" ] || fail "vloop $g has no show subcommand in its help"
  case "$s" in "Show "*) ;; *) fail "vloop $g show's Short does not begin 'Show ': '$s'" ;; esac
done

# Long help: the text above Usage: is not just the Short.
[ "$(above | wc -l | tr -d ' ')" -ge 3 ] || fail "vloop --help has no Long text: $(above)"
for c in run brief task metrics defect intervention; do
  a=$(above "$c")
  [ "$a" != "$(short "$c")" ] || fail "vloop $c --help has no Long text, only its Short: '$a'"
  [ "$(above "$c" | wc -l | tr -d ' ')" -ge 2 ] || fail "vloop $c --help's Long text is one line: '$a'"
done
above >"$t/root"
grep -qi 'brief' "$t/root" || fail "vloop --help does not say what a brief is"
grep -Fq 'vloop init' "$t/root" || fail "vloop --help does not say to start with vloop init"
grep -Fq 'vloop brief new' "$t/root" || fail "vloop --help does not name vloop brief new"
above run >"$t/run"
grep -qi 'resum' "$t/run" || fail "vloop run --help does not say how to resume a run"
grep -qi 'exit' "$t/run" || fail "vloop run --help does not give the exit codes"
grep -Fq 'concepts.md' "$t/run" || fail "vloop run --help does not say where the exit codes are documented (docs/guide/concepts.md)"

# The generated reference matches the tree.
(cd "$t" && "$t/gendocs" "$t/commands.md") >"$t/gd.out" 2>&1 || fail "gendocs failed: $(cat "$t/gd.out")"
cmp -s "$t/commands.md" docs/guide/commands.md || fail "docs/guide/commands.md is not what cmd/gendocs writes from the command tree: regenerate it"
grep -Fq 'Plan a brief and work it, task by task, on a work branch' docs/guide/commands.md || fail "docs/guide/commands.md does not carry run's new Short"
echo "gate T13: ok"
