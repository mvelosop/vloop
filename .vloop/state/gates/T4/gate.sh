#!/bin/sh
# Gate T4 — P3, the kit's probes in Go. Each command runs on a copy of a
# fixture this gate carries, once through the starting point's Python (hook.py,
# report.py and bypass.sh's `calls` heredoc, taken from the base commit) and
# once through the kit, both with this gate's stub copilot. Their stdout,
# stderr and exit code must be equal, and so must the trees they leave — the
# files they write and what the stub logged of the session it was started as
# (argv, working directory, environment) — once each copy's directory is
# replaced by @DIR@.
G=.vloop/state/gates/T4
. "$G/lib.sh"
BASE=d7b9cbb579f82f8626c9211e5ce8f2f7e8f720aa

mkdir -p "$t/py" "$t/a" "$t/b" "$t/n" || fail mkdir
for f in report.py hook.py bypass.sh; do
  git -C "$top" show "$BASE:tools/copilot-acceptance/$f" >"$t/py/$f" || fail "git show the base's $f"
done
awk '/^  python3 - "\$d\/out.jsonl" <<.PY.$/ {on=1; next} on && /^PY$/ {on=0} on' "$t/py/bypass.sh" >"$t/py/calls.py"
[ -s "$t/py/calls.py" ] || fail "could not extract the calls heredoc from the base's bypass.sh"
cp "$G/copilot" "$t/copilot" && chmod +x "$t/copilot" || fail "stub copilot"
go build -buildvcs=false -o "$t/kit" ./tools/copilot-acceptance/kit >"$t/kit.out" 2>&1 ||
  fail "go build ./tools/copilot-acceptance/kit: $(tail -n 5 "$t/kit.out")"

# put NAME: fresh copies of fixture NAME at $t/a/NAME (the Python's) and $t/b/NAME (the kit's).
put() {
  rm -rf "$t/a/$1" "$t/b/$1"
  py "$G/tree.py" put "$G/$1" "$t/a/$1" && py "$G/tree.py" put "$G/$1" "$t/b/$1" || fail "copy fixture $1"
}
# pyrun / kitrun OUT ARGS: run the reference (a python file in $t/py and its
# arguments) or the kit, from a directory outside both copies; stdout, stderr
# and the exit code go to OUT.out, OUT.err and OUT.rc.
pyrun() { o=$1; shift; (cd "$t/elsewhere" && py "$@") >"$o.out" 2>"$o.err"; echo $? >"$o.rc"; }
kitrun() { o=$1; shift; (cd "$t/elsewhere" && "$t/kit" "$@") >"$o.out" 2>"$o.err"; echo $? >"$o.rc"; }
# same NAME STEP: the two runs printed and exited alike (each copy's directory
# as @DIR@), and the two copies hold the same files.
same() {
  for x in out err rc; do
    sed "s|$t/a/$1|@DIR@|g" "$t/pa.$x" >"$t/pa.n.$x"; sed "s|$t/b/$1|@DIR@|g" "$t/pb.$x" >"$t/pb.n.$x"
    cmp -s "$t/pa.n.$x" "$t/pb.n.$x" || fail "$2: the kit's $x differs from the Python's:
$(diff "$t/pa.n.$x" "$t/pb.n.$x" | head -n 30)"
  done
  rm -rf "$t/n/a" "$t/n/b"
  py "$G/tree.py" norm "$t/a/$1" "$t/n/a" && py "$G/tree.py" norm "$t/b/$1" "$t/n/b" || fail "normalise $1"
  diff -r "$t/n/a" "$t/n/b" >"$t/tree.diff" || fail "$2: the files the kit left differ from the Python's:
$(head -n 40 "$t/tree.diff")"
}

# ---- probe ------------------------------------------------------------------
put probe
pyrun "$t/pa" "$t/py/report.py" probe "$t/a/probe" "$t/copilot"
kitrun "$t/pb" probe "$t/b/probe" "$t/copilot"
same probe "probe"
[ -s "$t/b/probe/stub.log" ] && [ -f "$t/b/probe/probe.argv" ] && [ "$(cat "$t/b/probe/probe.rc")" = 3 ] ||
  fail "probe: the stub was not started as the logged review session (no stub.log, probe.argv or probe.rc 3)"

# ---- no review session in the log -------------------------------------------
for c in "report.py probe|probe" "hook.py forms|hook forms" "hook.py failure|hook failure"; do
  pyargs=${c%%|*} kitargs=${c#*|}
  put noreview
  extra=; [ "$kitargs" = "hook failure" ] && extra=exit1
  set -- $pyargs; pyrun "$t/pa" "$t/py/$1" "$2" "$t/a/noreview" "$t/copilot" $extra
  set -- $kitargs; kitrun "$t/pb" "$@" "$t/b/noreview" "$t/copilot" $extra
  same noreview "$kitargs, no review session"
  [ "$(cat "$t/pb.rc")" = 1 ] && [ "$(cat "$t/pb.err")" = "c5: no review session in the copilot log" ] ||
    fail "$kitargs, no review session: want exit 1 and 'c5: no review session in the copilot log' on stderr"
done

# ---- summary ------------------------------------------------------------------
for s in sum1 sum2; do
  put "$s"
  pyrun "$t/pa" "$t/py/report.py" summary "$t/a/$s"
  kitrun "$t/pb" summary "$t/b/$s"
  same "$s" "summary on $s"
done

# ---- hook forms, each failure mode, hook summary, in one directory -------------
put hook
pyrun "$t/pa" "$t/py/hook.py" forms "$t/a/hook" "$t/copilot"
kitrun "$t/pb" hook forms "$t/b/hook" "$t/copilot"
same hook "hook forms"
for m in missing exit1 notjson timeout; do
  pyrun "$t/pa" "$t/py/hook.py" failure "$t/a/hook" "$t/copilot" "$m"
  kitrun "$t/pb" hook failure "$t/b/hook" "$t/copilot" "$m"
  same hook "hook failure $m"
done
pyrun "$t/pa" "$t/py/hook.py" summary "$t/a/hook"
kitrun "$t/pb" hook summary "$t/b/hook"
same hook "hook summary"
[ "$(grep -c '^=== stub$' "$t/b/hook/stub.log")" = 5 ] || fail "hook: want 5 sessions started (forms and four failure modes)"

# ---- hook forms with VLOOP_HOOK_PHASE=second-review ------------------------------
put hook
(export VLOOP_HOOK_PHASE=second-review; pyrun "$t/pa" "$t/py/hook.py" forms "$t/a/hook" "$t/copilot")
(export VLOOP_HOOK_PHASE=second-review; kitrun "$t/pb" hook forms "$t/b/hook" "$t/copilot")
same hook "hook forms, VLOOP_HOOK_PHASE=second-review"
py -c '
import json, sys
h = json.load(open(sys.argv[1]))["hooks"]["preToolUse"][0]
sys.exit(0 if "--phase second-review " in h["bash"] and "--phase second-review " in h["powershell"] else 1)' \
  "$t/b/hook/it's repo/.vloop/tmp/home/hooks/vloop-fence.json" ||
  fail "hook forms with VLOOP_HOOK_PHASE=second-review: the hooks file's commands do not hold --phase second-review"

# ---- hook summary: forms not run, one mode probed, no hits log ---------------------
put hsum
pyrun "$t/pa" "$t/py/hook.py" summary "$t/a/hsum"
kitrun "$t/pb" hook summary "$t/b/hsum"
same hsum "hook summary, partial"

# ---- calls ----------------------------------------------------------------------
put calls
pyrun "$t/pa" "$t/py/calls.py" "$t/a/calls/out.jsonl"
kitrun "$t/pb" calls "$t/b/calls/out.jsonl"
same calls "calls"
[ "$(wc -l <"$t/pb.out" | tr -d ' ')" = 5 ] || fail "calls: want 5 call lines, got: $(cat "$t/pb.out")"

# ---- arguments: exit 1, a one-line message on stderr, nothing on stdout -------------
refuse() {
  n=$1; shift
  (cd "$t/elsewhere" && "$t/kit" "$@") >"$t/out" 2>"$t/err"; rc=$?
  [ "$rc" = 1 ] || fail "kit $n: exit $rc, want 1"
  [ -s "$t/err" ] && [ "$(wc -l <"$t/err" | tr -d ' ')" -le 1 ] || fail "kit $n: want a one-line message on stderr, got: $(cat "$t/err")"
  [ ! -s "$t/out" ] || fail "kit $n: printed on stdout: $(cat "$t/out")"
}
put hook
refuse "hook failure, unknown mode" hook failure "$t/b/hook" "$t/copilot" bogus
[ ! -e "$t/b/hook/stub.log" ] || fail "kit hook failure with an unknown mode started a session"
refuse "hook (no subcommand)" hook
refuse "hook bogus" hook bogus "$t/b/hook"
refuse "probe (arguments missing)" probe "$t/b/hook"
refuse "summary (no directory)" summary
refuse "calls (no file)" calls

echo "gate T4: ok"
