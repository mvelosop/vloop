#!/bin/sh
# Gate T4 — L3, second half: sessions cannot change the local file. Judged on
# the fences (fence/*.json, parsed), the audit table's rows, and two whole runs
# of the vloop built here with a stub claude whose work session changes or
# creates the local file: after the run it is as it was.
G=.vloop/state/gates/T4
. "$G/lib.sh"
L=.vloop/config.local.toml

# 1. Every fence that denies Edit and Write of .vloop/config.toml denies them
#    of .vloop/config.local.toml too.
for f in fence/*.json; do
  python3 -I -c 'import json,sys
d=json.load(open(sys.argv[1])).get("permissions",{}).get("deny",[])
for verb in ("Edit","Write"):
    if verb+"(.vloop/config.toml)" in d and verb+"(.vloop/config.local.toml)" not in d:
        print(sys.argv[1]+": denies "+verb+"(.vloop/config.toml) but not "+verb+"(.vloop/config.local.toml)"); sys.exit(1)' "$f" >"$t/fence.out" ||
    fail "$(cat "$t/fence.out")"
done
grep -l 'Edit(.vloop/config.toml)' fence/*.json >/dev/null || fail "no fence denies Edit(.vloop/config.toml) any more"

# 2. The audit table has a row for each new rule, as many as for the shared file.
A=docs/design-notes/fence-audit.md
for verb in Edit Write; do
  n=$(grep -c "^| \`$verb(.vloop/config.toml)\` |" "$A")
  m=$(grep -c "^| \`$verb(.vloop/config.local.toml)\` |" "$A")
  [ "$n" -gt 0 ] && [ "$m" = "$n" ] ||
    fail "$A has $m row(s) for $verb(.vloop/config.local.toml), want $n (one per row for $verb(.vloop/config.toml))"
done

# 3. Whole runs: plan, gate review, one work session, its review, verify.
cat >"$t/stub/plan.json" <<'EOF'
{"schema":"state/v2","run_id":"B20260101-0900-demo","brief":"docs/briefs/B20260101-0900-demo.loop-brief.md",
 "base":"","branch":"","status":"planning","iteration":0,
 "created":"2026-01-01T09:00:00Z","updated":"2026-01-01T09:00:00Z","shell":"sh",
 "checks":[{"name":"all","paths":["**"],"run":"true"}],"gate_scratch":[],"gate_review":{"rounds":0,"verdict":""},
 "tasks":[{"id":"T1","title":"Task T1","goal":"Do T1.","kind":"feature","fixtures":"","references":[],"depends_on":[],
   "acceptance":["T1.out exists"],"verify":"test -f T1.out","status":"pending","attempts":0,"notes":""}]}
EOF
# script.sh is sourced by the stub claude, with PHASE set; TAMPER is the shell
# the work session runs on the repository.
script() {
  cat >"$t/stub/script.sh" <<EOF
case "\$PHASE" in
  plan) mkdir -p .vloop/state; cp "\$dir/plan.json" .vloop/state/state.json;;
  gate-review) mkdir -p .vloop/tmp
    echo '{"schema":"gate-verdict/v2","verdict":"PASS","tasks":[],"notes":"none"}' > .vloop/tmp/gate-verdict.json;;
  work) touch T1.out; mkdir -p .vloop/tmp
    printf '{"schema":"proposal/v1","task":"T1","outcome":"done","summary":"made T1","files":[],"verified":"ok","notes":"none"}\n' > .vloop/tmp/proposal.json
    $1;;
  review) mkdir -p .vloop/tmp
    printf '{"schema":"verdict/v1","task":"T1","verdict":"PASS","criteria":[],"findings":[],"notes":"none"}\n' > .vloop/tmp/verdict.json;;
  verify) mkdir -p .vloop/tmp
    printf '{"schema":"verify/v2","complete":true,"not_run":[],"findings":[],"checked":{"example_lines":0,"sentences":0,"surfaces":0},"notes":"none"}\n' > .vloop/tmp/verify.json;;
esac
EOF
}
IGN='.vloop/tmp/
.vloop/config.local.toml
'

# 3a. A local file the work session changes is put back; so is the shared file.
initrepo a
printf '%s' "$IGN" >"$R/.gitignore" && commitall "$R" ignore
trust "$R"
printf '# this machine\nrun.max-iterations = 9\n' >"$R/$L"
cp "$R/$L" "$t/local.before"
cp "$R/.vloop/config.toml" "$t/shared.before"
script "printf '\\n# edited by a session\\n' >> .vloop/config.toml; printf 'run.cost-ceiling = 1\\n' >> $L"
runin "$R" run "$BRIEF"
wantrc "a run whose work session edits both config files" 0
cmp -s "$t/shared.before" "$R/.vloop/config.toml" || fail "the run did not restore .vloop/config.toml (harness check)"
cmp -s "$t/local.before" "$R/$L" || fail "the work session's edit of $L was not put back:
$(cat "$R/$L")"
[ -z "$(git -C "$R" log --format= --name-only main.. -- "$L")" ] && [ -z "$(git -C "$R" ls-files -- "$L")" ] ||
  fail "a driver commit took $L"

# 3b. A local file the work session creates is removed.
initrepo b
printf '%s' "$IGN" >"$R/.gitignore" && commitall "$R" ignore
trust "$R"
script "printf 'run.max-iterations = 1\\n' > $L"
runin "$R" run "$BRIEF"
wantrc "a run whose work session creates the local file" 0
[ ! -e "$R/$L" ] || fail "the work session created $L and it was left in place:
$(cat "$R/$L")"

echo "gate T4: ok"
