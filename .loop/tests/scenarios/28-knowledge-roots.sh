#!/usr/bin/env bash
# A declared knowledge root the planner cannot scan is the expensive kind of
# silence. Planning still succeeds; the tasks simply carry no references; and
# nothing downstream can tell "nothing bound this task" apart from "the planner
# could not see what did". The run completes, looking fine, having ignored the
# repo's conventions.
#
# So preflight says it, before the run spends anything — and says nothing at
# all when the repo has declared no knowledge, which is the right behaviour for
# a greenfield target and must stay free.
. "$(dirname "$0")/../lib.sh"
fixture_new

note "── a root with neither an index nor descriptions is called out ──"
mkdir -p "$FX/repo/docs/domain"
printf '# entities\n' >"$FX/repo/docs/domain/transaction.md"
cat >"$FX/repo/.claude/loop-knowledge.md" <<'KNOW'
# Knowledge roots
| Surface | Path | How to find what is in it |
|---|---|---|
| Domain model | `docs/domain/` | index at README.md |
KNOW
fixture_stub <<'STUB'
case "$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
{"run_id":"know","brief":"docs/briefs/0003-runstat-cli.md","status":"running","iteration":0,
 "created":"2026-08-15T00:00:00Z","updated":"2026-08-15T00:00:00Z","tasks":[
 {"id":"T1","title":"t","goal":"g","files":[],"depends_on":[],"references":[],
  "acceptance":["a"],"verify":"true","status":"pending","attempts":0,"notes":""}]}
PLANJSON
    ;;
  work)   jq -nc --arg t "$TASK" '{task:$t,outcome:"done",summary:"s",files:[],verified:"ok",notes:"none"}' > .loop/tmp/proposal.json ;;
  review) jq -nc --arg t "$TASK" '{task:$t,verdict:"PASS",criteria:[],findings:[],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB
fixture_run docs/briefs/0003-runstat-cli.md
assert_log "no index and no descriptions"
assert_log "docs/domain/"
# A warning, not a refusal: the repo may genuinely have nothing to say about
# that root yet, and refusing would make declaring a root a liability.
assert_exit 0

note "── an indexed root is scannable, and says so ──"
fixture_cleanup; fixture_new
mkdir -p "$FX/repo/docs/domain"
printf '# entities\n' >"$FX/repo/docs/domain/README.md"
cat >"$FX/repo/.claude/loop-knowledge.md" <<'KNOW'
# Knowledge roots
| Surface | Path | How to find what is in it |
|---|---|---|
| Domain model | `docs/domain/` | index at README.md |
KNOW
fixture_stub <<'STUB'
case "$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
{"run_id":"know2","brief":"docs/briefs/0003-runstat-cli.md","status":"running","iteration":0,
 "created":"2026-08-15T00:00:00Z","updated":"2026-08-15T00:00:00Z","tasks":[
 {"id":"T1","title":"t","goal":"g","files":[],"depends_on":[],
  "references":[{"path":"docs/domain/README.md","why":"the entities this touches"}],
  "acceptance":["a"],"verify":"true","status":"pending","attempts":0,"notes":""}]}
PLANJSON
    ;;
  work)   jq -nc --arg t "$TASK" '{task:$t,outcome:"done",summary:"s",files:[],verified:"ok",notes:"none"}' > .loop/tmp/proposal.json ;;
  review) jq -nc --arg t "$TASK" '{task:$t,verdict:"PASS",criteria:[],findings:[],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB
fixture_run docs/briefs/0003-runstat-cli.md
assert_log "knowledge root(s) scannable"
assert_exit 0

note "── a repo that declares nothing pays nothing ──"
fixture_cleanup; fixture_new
fixture_stub <<'STUB'
case "$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
{"run_id":"know3","brief":"docs/briefs/0003-runstat-cli.md","status":"running","iteration":0,
 "created":"2026-08-15T00:00:00Z","updated":"2026-08-15T00:00:00Z","tasks":[
 {"id":"T1","title":"t","goal":"g","files":[],"depends_on":[],
  "acceptance":["a"],"verify":"true","status":"pending","attempts":0,"notes":""}]}
PLANJSON
    ;;
  work)   jq -nc --arg t "$TASK" '{task:$t,outcome:"done",summary:"s",files:[],verified:"ok",notes:"none"}' > .loop/tmp/proposal.json ;;
  review) jq -nc --arg t "$TASK" '{task:$t,verdict:"PASS",criteria:[],findings:[],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB
fixture_run docs/briefs/0003-runstat-cli.md
# No declaration, no references field at all: still a clean run. The whole
# feature has to be free for a repo with nothing to declare.
assert_exit 0
# Grep for text the driver actually emits. "knowledge root" appears only in the
# success line, so asserting its absence passed while the fixture was inheriting
# this repo's declaration and warning about roots it did not have.
for phrase in "knowledge root" "declared root(s)" "index does not name" "loop-knowledge"; do
  grep -q -- "$phrase" "$FX/out.log" \
    && bad "preflight said '$phrase' with nothing declared" \
    || ok "silent about: $phrase"
done
note "── a root whose descriptions are one folder down is not blind ──"
fixture_cleanup; fixture_new
mkdir -p "$FX/repo/docs/handoffs/spa87"
printf -- '---\ndescription: the split-editor handoff\n---\n' >"$FX/repo/docs/handoffs/spa87/brief.md"
cat >"$FX/repo/.claude/loop-knowledge.md" <<'KNOW'
# Knowledge roots
| Surface | Path | How to find what is in it |
|---|---|---|
| Design handoffs | `docs/handoffs/` | every file carries a description |
KNOW
fixture_stub <<'STUB'
case "$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
{"run_id":"nested","brief":"docs/briefs/0003-runstat-cli.md","status":"running","iteration":0,
 "created":"2026-08-15T00:00:00Z","updated":"2026-08-15T00:00:00Z","tasks":[
 {"id":"T1","title":"t","goal":"g","files":[],"depends_on":[],
  "references":[{"path":"docs/handoffs/spa87/","why":"the bundle this renders"}],
  "acceptance":["a"],"verify":"true","status":"pending","attempts":0,"notes":""}]}
PLANJSON
    ;;
  work)   jq -nc --arg t "$TASK" '{task:$t,outcome:"done",summary:"s",files:[],verified:"ok",notes:"none"}' > .loop/tmp/proposal.json ;;
  review) jq -nc --arg t "$TASK" '{task:$t,verdict:"PASS",criteria:[],findings:[],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB
fixture_run docs/briefs/0003-runstat-cli.md
# The descriptions live a folder down, which is normal — one bundle per slice.
# A top-level glob would call this root blind while sitting on the answer.
assert_log "knowledge root(s) scannable"
# ...but the folder that got CITED has no way in, and that is worth saying.
assert_log "folder reference(s) with no index.md"
assert_log "docs/handoffs/spa87/"
# A warning, not a refusal: the plan is sound, the docs are not, and the
# planner cannot fix them.
assert_exit 0

note "── a folder with any of the three entry points is quiet ──"
for entry in index.md README.md README-handoff.md; do
  fixture_cleanup; fixture_new
  mkdir -p "$FX/repo/docs/handoffs/spa87"
  printf '# what is in here\n' >"$FX/repo/docs/handoffs/spa87/$entry"
  cat >"$FX/repo/.claude/loop-knowledge.md" <<'KNOW'
# Knowledge roots
| Surface | Path | How to find what is in it |
|---|---|---|
| Design handoffs | `docs/handoffs/` | index per bundle |
KNOW
  fixture_stub <<'STUB'
case "$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
{"run_id":"entry","brief":"docs/briefs/0003-runstat-cli.md","status":"running","iteration":0,
 "created":"2026-08-15T00:00:00Z","updated":"2026-08-15T00:00:00Z","tasks":[
 {"id":"T1","title":"t","goal":"g","files":[],"depends_on":[],
  "references":[{"path":"docs/handoffs/spa87/","why":"the bundle this renders"}],
  "acceptance":["a"],"verify":"true","status":"pending","attempts":0,"notes":""}]}
PLANJSON
    ;;
  work)   jq -nc --arg t "$TASK" '{task:$t,outcome:"done",summary:"s",files:[],verified:"ok",notes:"none"}' > .loop/tmp/proposal.json ;;
  review) jq -nc --arg t "$TASK" '{task:$t,verdict:"PASS",criteria:[],findings:[],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB
  fixture_run docs/briefs/0003-runstat-cli.md
  grep -q "no index.md, README.md or README-\*.md" "$FX/out.log" \
    && bad "$entry was not accepted as an entry point" \
    || ok "$entry accepted as an entry point"
done

note "── a document its index does not name is reported ──"
fixture_cleanup; fixture_new
mkdir -p "$FX/repo/docs/decisions"
cat >"$FX/repo/docs/decisions/README-decisions.md" <<'IDX'
# Decisions
| File | What it decides |
| --- | --- |
| `0001-storage.md` | where uploads live |
IDX
printf '# storage\n'  >"$FX/repo/docs/decisions/0001-storage.md"
printf '# currency\n' >"$FX/repo/docs/decisions/0002-currency.md"
cat >"$FX/repo/.claude/loop-knowledge.md" <<'KNOW'
# Knowledge roots
| Surface | Path | How to find what is in it |
|---|---|---|
| Decisions | `docs/decisions/` | index |
KNOW
fixture_stub <<'STUB'
case "$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
{"run_id":"cover","brief":"docs/briefs/0003-runstat-cli.md","status":"running","iteration":0,
 "created":"2026-08-15T00:00:00Z","updated":"2026-08-15T00:00:00Z","tasks":[
 {"id":"T1","title":"t","goal":"g","files":[],"depends_on":[],"references":[],
  "acceptance":["a"],"verify":"true","status":"pending","attempts":0,"notes":""}]}
PLANJSON
    ;;
  work)   jq -nc --arg t "$TASK" '{task:$t,outcome:"done",summary:"s",files:[],verified:"ok",notes:"none"}' > .loop/tmp/proposal.json ;;
  review) jq -nc --arg t "$TASK" '{task:$t,verdict:"PASS",criteria:[],findings:[],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB
fixture_run docs/briefs/0003-runstat-cli.md
# The unlisted one: committed, invisible to the planner, and nothing about the
# repo looks wrong. This is the failure an index has that frontmatter does not.
assert_log "index does not name"
assert_log "0002-currency.md"
# The listed one must NOT be reported, or the warning is noise within one run.
grep -q "0001-storage.md" "$FX/out.log" && bad "a listed document was reported as missing" \
  || ok "the listed document was not reported"
assert_exit 0

note "── an index is never reported as missing from itself ──"
# Once per name, because the bug this pins is in how the path is BUILT, and the
# three names take different routes: index.md and README.md are concatenated
# onto the root ("docs/decisions/" + "/README.md" = a doubled slash that no
# comparison against find's normalised output can match), while README-*.md
# comes back from find already normalised. Testing only the last one passes
# while the first two are broken, which is exactly what happened here.
for idx in index.md README.md README-decisions.md; do
  fixture_cleanup; fixture_new
  mkdir -p "$FX/repo/docs/decisions"
  printf '# decisions\n\n- `0001-storage.md` where uploads live\n' >"$FX/repo/docs/decisions/$idx"
  printf '# storage\n' >"$FX/repo/docs/decisions/0001-storage.md"
  cat >"$FX/repo/.claude/loop-knowledge.md" <<'KNOW'
# Knowledge roots
| Surface | Path | How to find what is in it |
|---|---|---|
| Decisions | `docs/decisions/` | index |
KNOW
  fixture_stub <<'STUB'
case "$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
{"run_id":"selfidx","brief":"docs/briefs/0003-runstat-cli.md","status":"running","iteration":0,
 "created":"2026-08-15T00:00:00Z","updated":"2026-08-15T00:00:00Z","tasks":[
 {"id":"T1","title":"t","goal":"g","files":[],"depends_on":[],"references":[],
  "acceptance":["a"],"verify":"true","status":"pending","attempts":0,"notes":""}]}
PLANJSON
    ;;
  work)   jq -nc --arg t "$TASK" '{task:$t,outcome:"done",summary:"s",files:[],verified:"ok",notes:"none"}' > .loop/tmp/proposal.json ;;
  review) jq -nc --arg t "$TASK" '{task:$t,verdict:"PASS",criteria:[],findings:[],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB
  fixture_run docs/briefs/0003-runstat-cli.md
  grep -q "index does not name" "$FX/out.log" \
    && bad "$idx was reported as missing from itself" \
    || ok "$idx is not reported as missing from itself"
done

assert_no_tool_errors
finish
