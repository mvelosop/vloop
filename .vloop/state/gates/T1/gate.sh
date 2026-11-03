#!/bin/sh
# Gate T1 — A1: the flags, the cases read and --list. Judged through the command,
# built from ./tools/agent-eval into a temporary directory and run from the
# repository root, with fake claude, copilot and vloop first on PATH (each records
# that it ran) and fixture cases copied from this folder into the temporary
# directory. Nothing in this gate starts a session.
set -u
export LC_ALL=C
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
G=.vloop/state/gates/T1
[ -d "$G/cases" ] || fail "run from the repository root"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") || fail mktemp
t=$(cd "$t" && pwd -P)
trap 'rm -rf "$t"' EXIT
go build -o "$t/agent-eval" ./tools/agent-eval >"$t/build.out" 2>&1 ||
  fail "go build ./tools/agent-eval fails: $(tail -n 5 "$t/build.out")"
AE=$t/agent-eval
fb=$t/bin; st=$t/st; mkdir "$fb" "$st" || fail mkdir
for f in copilot claude vloop; do
  sed "s|@ST@|$st|g" "$G/fake-$f.sh" >"$fb/$f" && chmod +x "$fb/$f" || fail "cannot install the fake $f"
done
cp -R "$G/cases" "$t/cases" && cp -R "$G/bad-frontmatter" "$t/bad1" && cp -R "$G/bad-prompt" "$t/bad2" || fail "cannot copy the fixture cases"
PATH="$fb:$PATH"; export PATH
FX=$t/cases

# ae <name> <args...>: runs the tool; its stdout, stderr and exit code land in $t/<name>.{out,err,rc}.
ae() { n=$1; shift; "$AE" "$@" >"$t/$n.out" 2>"$t/$n.err"; echo $? >"$t/$n.rc"; }
# refuses <name> <stderr line> <args...>: exit 2, that one line on stderr, nothing on stdout.
refuses() {
  n=$1; want=$2; shift 2
  ae "$n" "$@"
  [ "$(cat "$t/$n.rc")" = 2 ] || fail "agent-eval $* exits $(cat "$t/$n.rc"), want 2 (stderr: $(cat "$t/$n.err"))"
  [ "$(cat "$t/$n.err")" = "$want" ] || fail "agent-eval $*: stderr is '$(cat "$t/$n.err")', want '$want'"
  [ -s "$t/$n.out" ] && fail "agent-eval $*: prints on stdout: $(head -n 3 "$t/$n.out")"
  return 0
}

# --- A1's refusals, each in the table's order
refuses r1 "agent-eval: --agent is required" --cases "$FX"
refuses r2 "agent-eval: --model is required" --agent claude --cases "$FX"
refuses r3 "agent-eval: --out is required" --agent claude --model m --cases "$FX"
refuses r4 "agent-eval: --model is required" --agent nope --cases "$FX"
refuses r5 "agent-eval: unknown agent nope" --agent nope --model m --out "$t/o5" --cases "$FX"
refuses r6 "agent-eval: unknown agent nope" --agent nope --model m --out "$t/o5" --runs 0 --cases "$FX"
refuses r7 "agent-eval: --runs must be 1 or more" --agent claude --model m --out "$t/o7" --runs 0 --cases "$FX"
refuses r8 "agent-eval: --runs must be 1 or more" --agent claude --model m --out "$t/o7" --runs x --timeout nope --cases "$FX"
refuses r9 "agent-eval: --timeout 5 is not a duration" --agent claude --model m --out "$t/o9" --timeout 5 --cases "$FX"
mkdir "$t/full" && echo keep >"$t/full/keep.txt" || fail mkdir
refuses r10 "agent-eval: $t/full is not empty" --agent claude --model m --out "$t/full" --case 'nomatch*' --cases "$FX"
[ "$(ls -A "$t/full")" = keep.txt ] || fail "a refused --out gained files: $(ls -A "$t/full" | tr '\n' ' ')"
refuses r11 "agent-eval: no case matches" --agent claude --model m --out "$t/o11" --case 'nomatch*' --cases "$FX"
refuses r12 "agent-eval: no vloop binary for the fence hook" --agent copilot --model m --out "$t/o12" --cases "$FX" --vloop "$t/none/vloop"
for f in copilot.argv claude.calls judge.n vloop.calls; do
  [ -e "$st/$f" ] && fail "a refused invocation started a fake ($f exists)"
done

# --- --list over a fixture directory: _-prefixed entries are not cases
lines() { awk 'NF' "$1"; }
ae l1 --list --cases "$FX"
[ "$(cat "$t/l1.rc")" = 0 ] || fail "agent-eval --list --cases <fixtures> exits $(cat "$t/l1.rc"): $(cat "$t/l1.err")"
[ -s "$t/l1.err" ] && fail "agent-eval --list --cases <fixtures> prints on stderr: $(cat "$t/l1.err")"
[ "$(lines "$t/l1.out" | wc -l | tr -d ' ')" = 3 ] || fail "--list over the fixtures prints $(lines "$t/l1.out" | wc -l | tr -d ' ') lines, want a header and 2 cases:
$(cat "$t/l1.out")"
[ "$(lines "$t/l1.out" | awk 'NR==1 { print $1, $2, $3, NF }')" = "CASE PHASE GRADERS 3" ] || fail "--list's header is not CASE PHASE GRADERS: $(head -n 1 "$t/l1.out")"
[ "$(lines "$t/l1.out" | awk 'NR==2 { print $1, $2, $3, NF }')" = "fixture-operate - 0 3" ] || fail "--list's first case line is not 'fixture-operate - 0': $(lines "$t/l1.out" | sed -n 2p)"
[ "$(lines "$t/l1.out" | awk 'NR==3 { print $1, $2, $3, NF }')" = "fixture-review review 4 3" ] || fail "--list's second case line is not 'fixture-review review 4': $(lines "$t/l1.out" | sed -n 3p)"
# aligned: every line's second and third columns start where the header's do, at least two spaces after the previous column
lines "$t/l1.out" | awk '
  { match($0, /^[^ ]+ +/); c2 = RLENGTH; gap2 = RLENGTH - length($1)
    rest = substr($0, c2 + 1); match(rest, /^[^ ]+ +/); c3 = c2 + RLENGTH; gap3 = RLENGTH - length($2) }
  NR == 1 { h2 = c2; h3 = c3 }
  c2 != h2 || c3 != h3 || gap2 < 2 || gap3 < 2 { bad = 1 }
  END { exit bad }' || fail "--list's columns are not aligned with two spaces or more between them:
$(cat "$t/l1.out")"

# --case is a path.Match glob, repeatable, a case selected when any matches
ae l2 --list --cases "$FX" --case 'fixture-o*' --case 'fixture-r?view'
[ "$(cat "$t/l2.rc")" = 0 ] || fail "--list with two --case flags exits $(cat "$t/l2.rc"): $(cat "$t/l2.err")"
[ "$(lines "$t/l2.out" | awk 'NR > 1 { print $1 }' | tr '\n' ' ')" = "fixture-operate fixture-review " ] ||
  fail "--case 'fixture-o*' --case 'fixture-r?view' does not select both cases: $(cat "$t/l2.out")"
ae l3 --list --cases "$FX" --case 'fixture-r*'
[ "$(lines "$t/l3.out" | awk 'NR > 1 { print $1 }' | tr '\n' ' ')" = "fixture-review " ] ||
  fail "--case 'fixture-r*' does not select fixture-review alone: $(cat "$t/l3.out")"

# --- --list over plugin/evals, the default --cases: one line per case, as the cases say
ae l4 --list
[ "$(cat "$t/l4.rc")" = 0 ] || fail "agent-eval --list exits $(cat "$t/l4.rc"): $(cat "$t/l4.err")"
[ -s "$t/l4.err" ] && fail "agent-eval --list prints on stderr: $(cat "$t/l4.err")"
: >"$t/want"
for d in plugin/evals/*/; do
  c=$(basename "$d")
  case $c in _*) continue ;; esac
  skill=$(awk 'n >= 2 && NF { print $1; exit } $0 == "---" { n++ }' "$d/prompt.md")
  ph=${skill#/vloop:}
  case $ph in plan | work | review | gate-review | verify | review-brief) ;; *) ph=- ;; esac
  ng=$(ls "$d"graders/*.md 2>/dev/null | wc -l | tr -d ' ')
  echo "$c $ph $ng" >>"$t/want"
done
[ -s "$t/want" ] || fail "plugin/evals holds no case"
lines "$t/l4.out" | awk 'NR > 1 { print $1, $2, $3 }' >"$t/got"
cmp -s "$t/want" "$t/got" || fail "agent-eval --list over plugin/evals differs from its cases (want, then got):
$(diff "$t/want" "$t/got")"

# --- a case that cannot be read: exit 1 before anything runs, with or without --list
for mode in list run; do
  if [ $mode = list ]; then set -- --list; else set -- --agent claude --model m --out "$t/b-$mode"; fi
  ae b1$mode "$@" --cases "$t/bad1"
  [ "$(cat "$t/b1$mode.rc")" = 1 ] || fail "an unreadable grader frontmatter ($mode) exits $(cat "$t/b1$mode.rc"), want 1"
  case $(cat "$t/b1$mode.err") in
    "agent-eval: fixture-bad/"*"x.md: "?*) ;;
    *) fail "an unreadable grader frontmatter ($mode): stderr is not 'agent-eval: fixture-bad/<file>: <reason>': $(cat "$t/b1$mode.err")" ;;
  esac
  [ "$(wc -l <"$t/b1$mode.err" | tr -d ' ')" = 1 ] || fail "an unreadable grader ($mode) prints more than one line on stderr"
  ae b2$mode "$@" --cases "$t/bad2"
  [ "$(cat "$t/b2$mode.rc")" = 1 ] || fail "a case without prompt.md ($mode) exits $(cat "$t/b2$mode.rc"), want 1"
  case $(cat "$t/b2$mode.err") in
    "agent-eval: fixture-noprompt/prompt.md: "?*) ;;
    *) fail "a case without prompt.md ($mode): stderr is not 'agent-eval: fixture-noprompt/prompt.md: <reason>': $(cat "$t/b2$mode.err")" ;;
  esac
done
for f in claude.calls judge.n copilot.argv; do
  [ -e "$st/$f" ] && fail "an unreadable case still started a fake ($f exists)"
done
[ -e "$t/b-run/report.json" ] && fail "an unreadable case still wrote a report"
echo "gate T1: ok"
