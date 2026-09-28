#!/usr/bin/env bash
#
# The autonomous loop driver.
#
#   .loop/run.sh [brief-path]
#
# Plans once from a brief, then repeats: pick the next ready task, run a work
# session on it, gate every completed task, run an independent review session,
# apply the verdict, commit. Every session is a fresh `claude -p` with no memory
# of any other; all continuity lives in files under .loop/.
#
# The driver owns every mechanical decision — which task is next, whether a task
# is really done, how many attempts it has burned, whether the run is
# converging, and when to stop. Agents do the work and give opinions; they never
# set status and never commit.
#
# Design: docs/briefs/0002-next-generation-autonomous-loop.md
#
# Exit codes
#   0  plan complete            4  max iterations (resumable)
#   1  preflight / usage        5  not converging  (needs a human)
#   2  blocked (needs a human)  6  cost ceiling    (resumable)
#   3  stalled                  7  session error   (needs a human)
#   8  repeat blocked, nothing changed since the first diagnosis (needs a human)
#
# Env
#   LOOP_MAX_ITERATIONS   iterations this run may use          (default 30)
#   LOOP_COST_CEILING     dollars this run may spend           (default 40)
#   LOOP_MAX_ATTEMPTS     failures before a task is blocked    (default 3)
#   LOOP_STALL_LIMIT      no-change iterations before stopping (default 2)
#   LOOP_CONVERGENCE_MAX  max iterations-per-closed-task       (default 3.0)
#   LOOP_CONVERGENCE_MIN  iterations before that check arms    (default 6)
#   LOOP_PLAN_MODEL       model for the plan phase             (default opus)
#   LOOP_WORK_MODEL       model for work + review              (default sonnet)
#   LOOP_ARCHIVE_TRANSCRIPTS=1   copy session transcripts out of ~/.claude
#   LOOP_TRANSCRIPT_DIR   where to put them  (default ../loop-transcripts)

set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO" || exit 1

# The loop is one directory. Inside it, one line separates what the installer
# owns from what the consumer repo owns: everything above state/ and tmp/ is
# replaced wholesale on upgrade, and nothing below them is ever written by it.
LOOP_DIR="$REPO/.loop"
STATE_DIR="$LOOP_DIR/state"     # durable, committed, the repo's own record
TMP_DIR="$LOOP_DIR/tmp"         # transient per-iteration handoffs, gitignored
mkdir -p "$STATE_DIR" "$TMP_DIR"

STATE="$STATE_DIR/state.json"
PROPOSAL="$TMP_DIR/proposal.json"
VERDICT="$TMP_DIR/verdict.json"

# state.json belongs to the driver. Both session contracts say so in prose --
# and a rule in prose is not a control. A session that edits a task's `verify`
# changes what every later gate runs, which collapses the one assumption the
# whole loop rests on: that every verify command was authored before any
# implementation existed. A session that writes both the work and the gate has
# a gate that proves nothing, however correct its rewrite happens to be.
#
# So: snapshot before each session, restore after. The snapshot lives outside
# the repo because it is transient scaffolding within a single run, not a
# record -- nothing reads it after the iteration that wrote it.
STATE_PRE="$(mktemp "${TMPDIR:-/tmp}/loop-state.XXXXXX")"

state_snapshot() { [[ -f "$STATE" ]] && cp "$STATE" "$STATE_PRE"; }

# Restores and returns 0 when the session touched it; 1 when it behaved.
state_restore_if_touched() {
  [[ -f "$STATE" && -s "$STATE_PRE" ]] || return 1
  cmp -s "$STATE" "$STATE_PRE" && return 1
  cp "$STATE_PRE" "$STATE"
  return 0
}

MAX_ITER="${LOOP_MAX_ITERATIONS:-30}"
COST_CEILING="${LOOP_COST_CEILING:-40}"
MAX_ATTEMPTS="${LOOP_MAX_ATTEMPTS:-3}"
STALL_LIMIT="${LOOP_STALL_LIMIT:-2}"
CONVERGENCE_MAX="${LOOP_CONVERGENCE_MAX:-3.0}"
CONVERGENCE_MIN="${LOOP_CONVERGENCE_MIN:-6}"
PLAN_MODEL="${LOOP_PLAN_MODEL:-opus}"
WORK_MODEL="${LOOP_WORK_MODEL:-sonnet}"

# LOOP_ACTIVE_TASK and LOOP_GATE_TASK are the driver's own, set only for the
# duration of each verify command inside gate_ids() below. Unset here so a
# value inherited from whatever invoked this script — including, as happens
# when run-all.sh runs inside another gate, a stale one left by an outer
# driver — never reaches a work or review session, which read the plain
# environment via `claude -p`.
unset LOOP_ACTIVE_TASK LOOP_GATE_TASK

PLAN_ONLY=0
CHECK_ONLY=0
REPLAN=0
BRIEF=""
for a in "$@"; do
  case "$a" in
    --plan-only|--only-plan) PLAN_ONLY=1 ;;
    --check|--preflight)     CHECK_ONLY=1 ;;
    --replan)                REPLAN=1 ;;
    *) BRIEF="$a" ;;
  esac
done

# Validate the argument before anything is opened, locked or reset.
#
# The plan-reset path below deletes a committed state.json when the brief you
# name differs from the one the plan holds — which is right, but it used to
# happen without ever checking that the brief you named exists. So `run.sh
# --help`, or a mistyped path, read as "a different brief" and destroyed a
# finished plan on its way to failing. Found by doing exactly that.
#
# A resumable run is the loop's central promise; a typo must not be able to
# spend it.
case "$BRIEF" in
  -h|--help)
    cat <<'USAGE'
usage: .loop/run.sh [--check | --plan-only] [--replan] [brief-path]

  .loop/run.sh docs/briefs/0003-runstat-cli.md   plan and run that brief
  .loop/run.sh                                   resume the plan in .loop/state/state.json
  .loop/run.sh --plan-only <brief>               plan, commit it, and stop
  .loop/run.sh --check                           run the checks and stop; spends nothing
  .loop/run.sh --replan <brief>                  plan from a brief that has already run

--check answers "is this repo ready?" — tools, workspace trust, the permission
fence, git identity, and whether the knowledge roots you declared can actually
be surveyed: each one has an index or describes itself, and each index names
every document beside it. It costs nothing, touches nothing and needs no brief,
so it is the right thing to run after editing .claude/loop-knowledge.md or
adding a document.

Related, and also free:
  .loop/check-brief.sh <brief>    is this brief plannable

--plan-only is the cheap half of a run: it pays for one planning session, then
stops so you can read the plan and its verify commands before committing to
the rest. Review .loop/state/plan.md, adjust with .loop/amend.sh, then run
.loop/run.sh with no argument to execute it.

A brief's journal at .loop/state/journals/<brief-stem>.md is proof its run
already happened; planning from it again would overwrite .loop/state/state.json
and re-derive work already on main, so run.sh refuses unless you pass
--replan — a deliberate re-plan after an aborted run.

Naming a brief other than the one the current plan holds resets that plan and
starts fresh. Docs: .loop/manual.md
USAGE
    exit 0 ;;
  -*)
    printf '\033[31m[loop] unknown option: %s — see .loop/run.sh --help\033[0m\n' "$BRIEF" >&2
    exit 2 ;;
esac
if [[ -n "$BRIEF" && ! -f "$BRIEF" ]]; then
  printf '\033[31m[loop] brief not found: %s\033[0m\n' "$BRIEF" >&2
  printf '\033[31m[loop] nothing was changed; the current plan is untouched\033[0m\n' >&2
  exit 2
fi
BRANCH="$(git branch --show-current 2>/dev/null)"
[[ -n "$BRANCH" ]] || BRANCH="detached-$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
BRANCH_SAFE="${BRANCH//\//-}"
# Grouped by branch, so two loops running in parallel write to different paths.
# Within one branch the timestamp is only second-resolution, and two runs can
# land in the same second — a quick preflight failure followed by a re-run, or
# a test firing several in a row. A collision would have the later run truncate
# the earlier one's telemetry, so make the directory unique rather than assume.
RUN_STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_ID="$RUN_STAMP"
_n=1
while [[ -d "$STATE_DIR/runs/$BRANCH_SAFE/$RUN_ID" ]]; do
  _n=$((_n + 1)); RUN_ID="$RUN_STAMP-$_n"
done
RUN_PATH="$BRANCH_SAFE/$RUN_ID"
RUN_DIR="$STATE_DIR/runs/$RUN_PATH"
SESSIONS="$RUN_DIR/sessions"
ITERATIONS="$RUN_DIR/iterations.jsonl"

SESSION_N=0

# ---------------------------------------------------------------- helpers ---

# Hard rule 2's backstop. Agents are told to write repo-relative paths; tool
# output and stack traces are not under their control, so everything this script
# persists goes through here.
USER_NAME="$(basename "$HOME")"
mask() { sed -e "s#${HOME}#~#g" -e "s#${USER_NAME}#USER#g"; }

# Everything the driver persists goes through mask() — including its own log,
# which is committed as evidence. An absolute path reaching a message is a
# mistake waiting to happen, so the backstop sits here rather than at each
# call site.
say()  { printf '\033[36m[loop]\033[0m %s\n' "$(printf '%s' "$*" | mask)" | tee -a "$RUN_DIR/loop.log" >&2; }
WARN_COUNT=0
warn() { WARN_COUNT=$((WARN_COUNT + 1)); printf '\033[33m[loop]\033[0m %s\n' "$(printf '%s' "$*" | mask)" | tee -a "$RUN_DIR/loop.log" >&2; }
die()  { printf '\033[31m[loop] %s\033[0m\n' "$(printf '%s' "$*" | mask)" >&2; exit 1; }
ts()   { date -u +%Y-%m-%dT%H:%M:%SZ; }

# All state access goes through jq. Never grep/sed/awk over the state file — a
# hand-rolled parser meeting a document format is where this class of loop
# reliably breaks (docs/references/executable-loop-harness.md).
state_get() { jq -r "$@" "$STATE"; }

# Atomic: build the new state beside the old one, then rename over it. A run
# killed mid-write leaves the previous state intact rather than a truncated file.
state_edit() {
  local tmp
  tmp="$(mktemp "$STATE_DIR/.state.XXXXXX")"
  if jq "$@" "$STATE" >"$tmp"; then
    mv "$tmp" "$STATE"
  else
    rm -f "$tmp"
    die "state edit failed: $*"
  fi
}

# state.json is what the driver computes over; plan.md is what a human reads.
# Rendered after every state change, one direction only, never parsed back.
render_plan() {
  "$LOOP_DIR/render-plan.sh" "$STATE" "$STATE_DIR/plan.md" \
    || warn "plan render failed — .loop/state/plan.md may be stale"
}

# One `claude -p` session, fully contained: the project's own settings plus the
# loop's fence, no MCP servers, no memory of anything else. The fence is the
# loop's own file rather than an entry merged into the consumer's
# .claude/settings.json, so it binds loop sessions and nothing else: deny still
# beats allow across sources, and the repo's interactive work is untouched. Its result JSON is stamped with the phase
# and iteration that produced it, masked, and kept as telemetry.
run_session() {
  local phase="$1" iter="$2" model="$3" prompt="$4" rc=0 raw out
  SESSION_N=$((SESSION_N + 1))
  out="$(printf '%s/%03d-%s.json' "$SESSIONS" "$SESSION_N" "$phase")"
  raw="$out.raw"

  claude -p "$prompt" \
    --model "$model" \
    --permission-mode auto \
    --setting-sources project \
    --settings "$LOOP_DIR/settings.json" \
    --strict-mcp-config \
    --output-format json \
    >"$raw" 2>"$RUN_DIR/$phase-$iter.stderr"
  rc=$?

  if [[ -s "$raw" ]]; then
    jq --arg p "$phase" --argjson i "$iter" '. + {phase: $p, iteration: $i}' "$raw" \
      | mask >"$out" 2>/dev/null || cp "$raw" "$out"
  fi
  rm -f "$raw"

  if [[ -s "$out" ]]; then
    jq -r '"    cost=$\(.total_cost_usd // 0) turns=\(.num_turns // 0) dur=\(((.duration_ms // 0)/1000)|round)s error=\(.is_error // false)"' \
      "$out" 2>/dev/null | while read -r l; do say "$l"; done
    local denials
    denials="$(jq -r '.permission_denials | length' "$out" 2>/dev/null || echo 0)"
    [[ "${denials:-0}" -gt 0 ]] && warn "    $denials permission denial(s) — the fence may be in the wrong place"
  fi

  archive_transcript "$out" "$phase" "$iter"
  return $rc
}

# Session transcripts live in ~/.claude and are collected after ~30 days. They
# hold absolute paths and full file contents, so this is off unless asked for,
# and it never writes inside the repo.
archive_transcript() {
  [[ "${LOOP_ARCHIVE_TRANSCRIPTS:-0}" == "1" ]] || return 0
  local result="$1" phase="$2" iter="$3" sid src dest
  dest="${LOOP_TRANSCRIPT_DIR:-$REPO/../loop-transcripts}/$RUN_PATH"
  case "$(cd "$dest" 2>/dev/null && pwd)" in "$REPO"|"$REPO"/*)
    warn "refusing to archive transcripts inside the repo"; return 0 ;;
  esac
  sid="$(jq -r '.session_id // empty' "$result" 2>/dev/null)" || return 0
  [[ -n "$sid" ]] || return 0
  src="$(find "$HOME/.claude/projects" -maxdepth 2 -name "$sid.jsonl" -print -quit 2>/dev/null)"
  [[ -n "$src" ]] || return 0
  mkdir -p "$dest" && cp "$src" "$dest/$(printf '%03d-%s-iter%s.jsonl' "$SESSION_N" "$phase" "$iter")"
}

# ------------------------------------------------------------------ gates ---

# A goalpost can be moved without touching state.json at all, by rewriting the
# file a verify command RUNS. The command still reads
# `pytest -q tests/test_messages.py`; what it asserts is now the session's own.
# The state snapshot defends the pointer and leaves the target open.
#
# Four conditions, every one mechanical, and together they separate a moved
# goalpost from ordinary work:
#
#   modified, not created    a task that SHIPS the file its gate runs is the
#                            durable-artifact rule working as intended, so
#                            creation is always allowed
#   existed at HEAD          which makes the author an earlier session or the
#                            planner, never this one -- UNLESS that HEAD
#                            content is itself this task's own, from an
#                            earlier attempt: the driver commits every
#                            iteration whatever its outcome, so a task whose
#                            first attempt fails review still lands its new
#                            file in HEAD, and a retry that keeps editing that
#                            same file is still this task's own work, not a
#                            rewrite of someone else's gate. The driver's own
#                            commit subjects are "[loop] $task: $outcome"
#                            (below), so the last commit to touch the file
#                            says who actually wrote the HEAD version -- a
#                            narrower test than "does `files` own it", which a
#                            directory-style files entry (".loop/tests/
#                            scenarios/", used throughout this plan) can never
#                            satisfy by exact match, and which must stay
#                            narrow: owning everything under a directory would
#                            also excuse rewriting a FILE SOME OTHER TASK put
#                            there, which is exactly the rewrite below still
#                            has to catch
#   named by THIS task's     the session rewrote the gate it is judged by. A
#     verify                 file some OTHER task's gate names is deliberately
#                            not this check's business: 03-gate-regression has
#                            T2 delete a file T1's gate checks, which is a
#                            regression the gate re-run is built to catch, and
#                            restoring the file here would hide it
#   not in this task's       the plan assigned it, so writing it is the work
#
# Measured: reviewer-calibration case 06 plants exactly this -- correct code,
# every criterion met, and a rewritten gate that is a strictly BETTER test than
# the planner's. Three review sessions saw it and none objected; one cited the
# rewritten test as evidence the criteria were met. The reviewer is built to
# rule on substance and this is a question about process, so it belongs here.
#
# The retry carve-out is itself measured: B20260924-1947's own T9 hit this on
# its second attempt. Its first attempt created a new scenario file (allowed,
# per "modified, not created" above), but review failed it -- and the driver
# commits every iteration regardless of outcome, so that file was in HEAD by
# the second attempt. T9's `files` names the directory, not that exact path,
# so the second attempt's legitimate fix to its own file was reverted as a
# goalpost rewrite, and the task burned its last attempt on a false positive.
gate_files_moved() {
  local f owner
  while IFS= read -r f; do
    [[ -n "$f" ]] || continue
    git cat-file -e "HEAD:$f" 2>/dev/null || continue
    jq -e --arg f "$f" --arg t "$task" '
        (.tasks[] | select(.id == $t)) as $cur
        | (($cur.verify // "") | contains($f))
          and ((($cur.files // []) | index($f)) == null)
      ' "$STATE" >/dev/null 2>&1 || continue
    owner="$(git log -1 --format=%s -- "$f" 2>/dev/null)"
    case "$owner" in
      "[loop] $task:"*) continue ;;
    esac
    printf '%s\n' "$f"
  done < <(git diff --name-only HEAD 2>/dev/null)
}

# What a work session left in the tree when it died before writing a proposal --
# the only account of "did it do anything" left once its own report is gone.
# `git status --porcelain` (not `git diff HEAD`) because a session that died
# after only creating new files leaves nothing for `diff` to see. One git
# invocation; everything under .loop/state/ or .loop/tmp/ is the driver's own
# bookkeeping (telemetry the session never touched) and is filtered in-shell,
# not with a second git call.
session_tree_changes() {
  local line path
  while IFS= read -r line; do
    [[ -n "$line" ]] || continue
    path="${line:3}"
    path="${path#* -> }"
    case "$path" in
      .loop/state/*|.loop/tmp/*) continue ;;
    esac
    printf '%s\n' "$path"
  done < <(git status --porcelain 2>/dev/null)
}

# Paths a verify command reads the BYTES of, as opposed to merely handing to a
# runner. Used by the plan-time gate-shape lint (rule 3) to find, one phase
# early, exactly the files gate_files_moved() above would revert at runtime --
# grep, cat, test -f and a Python-level open(...) are inspecting; `uv run
# pytest -q path` and `bash path` are executing and stay unflagged. Not a full
# shell parse, on purpose: the four forms below are what the loop's own
# scenarios and the arc's five stalled runs actually wrote.
gate_inspected_paths() {
  local cmd="$1"
  { grep -oE 'grep[[:space:]]+(-[[:alnum:]]+[[:space:]]+)*[^[:space:]&|;-][^[:space:]&|;]*[[:space:]]+[^[:space:]&|;()"'"'"']+' <<<"$cmd" \
      | awk '{print $NF}'
    grep -oE 'cat[[:space:]]+[^[:space:]&|;]+' <<<"$cmd" | awk '{print $NF}'
    grep -oE 'test[[:space:]]+-f[[:space:]]+[^[:space:]&|;)]+' <<<"$cmd" | awk '{print $NF}'
    # \* rather than \?: this cmd may have come through jq's @tsv, which
    # doubles a literal backslash, so a quote escaped once in the source
    # (\") can arrive here escaped twice (\\").
    grep -oE 'open\(\\*["'"'"'][^"'"'"']+\\*["'"'"']\)' <<<"$cmd" | sed -E 's/^open\(\\*.//; s/\\*.\)$//'
  } | sort -u
}

# Refs a `git diff`/`log`/`rev-list` in a verify command is baselined against,
# other than HEAD -- used by the plan-time gate-shape lint (rule 4) to catch a
# baseline that decays the moment any other task commits. A range like
# `origin/main..HEAD` still fails: the left side is the fixed part, so it is
# what gets reported. Not a full shell parse, on purpose, same spirit as
# gate_inspected_paths above: skips flags, stops at a `--` pathspec separator,
# and reassembles a `$(...)` command-substitution ref that word-splitting would
# otherwise cut at its first internal space.
gate_diff_refs() {
  local cmd="$1" rest tok ref
  # Greedy to end of string, on purpose: a $(...) ref needs its closing paren
  # still present to be reassembled below, and a verify is one line, so this
  # loses nothing except a second, later diff/log/rev-list call in the same
  # command -- a shape none of this rule's cases actually write.
  rest="$(grep -oE 'git[[:space:]]+(diff|log|rev-list)[[:space:]]+.*' <<<"$cmd" \
          | sed -E 's/^git[[:space:]]+(diff|log|rev-list)[[:space:]]+//')"
  [[ -n "$rest" ]] || return 0
  ref=""
  for tok in $rest; do
    case "$tok" in
      --) break ;;
      -*) continue ;;
      *) ref="$tok"; break ;;
    esac
  done
  [[ -n "$ref" ]] || return 0
  case "$ref" in
    \"\$\(*|\$\(*) ref="$(grep -oE '\$\([^)]*\)' <<<"$rest" | head -1)" ;;
  esac
  ref="${ref#\"}"; ref="${ref%\"}"; ref="${ref#\'}"; ref="${ref%\'}"
  [[ -n "$ref" && "$ref" != "HEAD" ]] || return 0
  if [[ "$ref" == *..* ]]; then
    printf '%s\n' "${ref%%..*}"
  else
    printf '%s\n' "$ref"
  fi
}

# Re-run the verify command of every task named. This is the whole point of the
# external gate: "done" has to survive a command the session neither runs nor can
# edit. Echoes the ids that failed.
#
# Each verify command runs with LOOP_ACTIVE_TASK (the task this iteration is
# working, passed in as $1) and LOOP_GATE_TASK (the id currently being gated)
# in its environment — set only on the `bash -c` that runs that one command,
# never exported into this script's own environment, so nothing outside this
# loop body can see them and a work or review session never does. The two
# differ exactly when a gate is re-running as a regression check of a task
# other than the one being worked — the case that used to have no session
# whose scope it could speak to (brief B20260924-1947, failure E).
gate_ids() {
  local active="$1"; shift
  local id cmd rc failed=()
  mkdir -p "$RUN_DIR/gates"
  for id in "$@"; do
    cmd="$(state_get --arg id "$id" '.tasks[]|select(.id==$id)|.verify')"
    [[ -n "$cmd" && "$cmd" != "null" ]] || continue
    rc=0
    LOOP_ACTIVE_TASK="$active" LOOP_GATE_TASK="$id" \
      bash -c "$cmd" >"$RUN_DIR/gates/$id.log.raw" 2>&1 || rc=$?
    mask <"$RUN_DIR/gates/$id.log.raw" >"$RUN_DIR/gates/$id.log"
    rm -f "$RUN_DIR/gates/$id.log.raw"
    [[ $rc -ne 0 ]] && failed+=("$id")
  done
  printf '%s\n' "${failed[@]:-}"
}

# --------------------------------------------------------------- signals ----
#
# The run-level view. Every gate can be green and every review thorough while
# the run goes nowhere, because each of those judges a tick against its task and
# nothing judges the run against the point of the run
# (docs/references/executable-loop-harness.md Rule 7). These are the numbers
# that make that visible from outside, and one of them can stop the run.
#
# runstat recomputes all of this in Python afterwards; brief 0002 acceptance
# item 6 requires the two to agree. If you change a formula here, change it
# there — the fixture in brief 0003 is the arbiter.

sig_iterations()  { [[ -f "$ITERATIONS" ]] && wc -l <"$ITERATIONS" | tr -d ' ' || echo 0; }
# `jq -s` on an unmatched glob BOTH prints 0 and exits non-zero, so a bare
# `jq ... || echo 0` emits "0\n0" — which awk then rejects, silently disabling
# the cost-ceiling comparison. Check the glob matched instead of relying on ||.
sig_spend() {
  local f=("$SESSIONS"/*.json)
  [[ -e "${f[0]}" ]] || { echo 0; return 0; }
  jq -s '[.[] | .total_cost_usd // 0] | add // 0' "${f[@]}" 2>/dev/null || echo 0
}
sig_closed()      { jq -s 'if length == 0 then 0 else (.[-1].tasks_done // 0) end' "$ITERATIONS" 2>/dev/null || echo 0; }
sig_total()       { jq -s 'if length == 0 then 0 else (.[-1].tasks_total // 0) end' "$ITERATIONS" 2>/dev/null || echo 0; }
sig_gate_fails()  { jq -s '[.[]|select(.outcome=="gate_fail")]|length' "$ITERATIONS" 2>/dev/null || echo 0; }
sig_review_fails(){ jq -s '[.[]|select(.outcome=="review_fail")]|length' "$ITERATIONS" 2>/dev/null || echo 0; }
# Iterations that did not close their task. NOT the sum of `.attempts` — that
# field is a cumulative per-task counter, so summing it double-counts a task
# that appears in more than one record.
sig_attempts()    { jq -s '[.[]|select(.outcome!="done")]|length' "$ITERATIONS" 2>/dev/null || echo 0; }

sig_per_closed() {
  local i c; i="$(sig_iterations)"; c="$(sig_closed)"
  if [[ "${c:-0}" -eq 0 ]]; then echo "n/a"; else
    awk -v i="$i" -v c="$c" 'BEGIN{printf "%.2f", i/c}'
  fi
}

# Trailing iterations that closed nothing.
sig_streak() {
  jq -s '[.[] | .tasks_done // 0] as $d
         | ([0] + $d[:-1]) as $p
         | [range(0; ($d|length))] | map(if $d[.] > $p[.] then 1 else 0 end)
         | (reverse | index(1)) // length' "$ITERATIONS" 2>/dev/null || echo 0
}

print_signals() {
  say "  ── signals ──"
  say "  iterations:            $(sig_iterations)"
  say "  tasks closed:          $(sig_closed)/$(sig_total)"
  say "  iterations per closed: $(sig_per_closed)"
  say "  gate failures:         $(sig_gate_fails)"
  say "  review rejections:     $(sig_review_fails)"
  say "  attempts burned:       $(sig_attempts)"
  say "  no-progress streak:    $(sig_streak)"
  say "  estimated spend:       \$$(printf '%.2f' "$(sig_spend)") (estimate, not a bill)"
}

# --------------------------------------------------------------- lock -------
#
# Parallel loops are supported, one per git WORKTREE. Git already guarantees
# those are on different branches — it refuses to check one branch out twice —
# so the only thing left to prevent is two loops in the SAME working tree,
# where they would share .loop/state/state.json and, far worse, .loop/tmp/proposal.json:
# one loop's review session reading the other loop's proposal is exactly the
# stale-handoff failure the driver clears per-iteration to avoid.
#
# A lock that can brick the loop is worse than no lock, so it records a pid and
# a dead one is cleared rather than obeyed.
LOCK="$TMP_DIR/.running"

acquire_lock() {
  if [[ -f "$LOCK" ]]; then
    local pid other started
    pid="$(jq -r '.pid // ""' "$LOCK" 2>/dev/null)"
    other="$(jq -r '.branch // "?"' "$LOCK" 2>/dev/null)"
    started="$(jq -r '.started // "?"' "$LOCK" 2>/dev/null)"
    if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
      die "a loop is already running in this working tree.
  pid $pid · branch '$other' · started $started
  Two loops in one tree share .loop/state/state.json and .loop/tmp/proposal.json, so one
  can mark a task done on the other's evidence. To run in parallel, give each
  its own worktree:
    git worktree add ../<dir> <branch>"
    fi
    warn "clearing a stale lock (pid ${pid:-unknown} is gone)"
  fi
  jq -nc --arg p "$$" --arg b "$BRANCH" --arg t "$(ts)" --arg r "$RUN_PATH" \
    '{pid: $p, branch: $b, started: $t, run: $r}' >"$LOCK"
  trap 'rm -f "$LOCK" "$STATE_PRE"' EXIT
}

# ------------------------------------------------------------- preflight ----

preflight() {
  local ok=1
  say "preflight"

  for t in jq claude git; do
    if command -v "$t" >/dev/null 2>&1; then say "  [x] $t"
    else say "  [ ] $t — missing"; ok=0; fi
  done
  command -v uv >/dev/null 2>&1 && say "  [x] uv" || warn "  [ ] uv — missing (the demo target needs it)"

  # An untrusted workspace makes `claude -p` silently ignore this repo's
  # settings.json permissions. That failure took out a whole prior experiment
  # and left one line in a log preamble as its only trace. Fail loudly.
  local trusted
  trusted="$(jq -r --arg p "$REPO" '.projects[$p].hasTrustDialogAccepted // false' "$HOME/.claude.json" 2>/dev/null)"
  if [[ "$trusted" == "true" ]]; then
    say "  [x] workspace trusted"
  else
    say "  [ ] workspace NOT trusted — settings.json permissions would be silently ignored"
    say "      fix: run \`claude\` interactively here once and accept the trust dialog"
    ok=0
  fi

  if jq -e . "$LOOP_DIR/settings.json" >/dev/null 2>&1; then
    say "  [x] .loop/settings.json parses — $(jq '.permissions.allow|length' "$LOOP_DIR/settings.json") allow, $(jq '.permissions.deny|length' "$LOOP_DIR/settings.json") deny rules"
  else
    say "  [ ] .loop/settings.json missing or invalid — sessions would run without the fence"; ok=0
  fi
  if [[ -f .claude/settings.json ]] && ! jq -e . .claude/settings.json >/dev/null 2>&1; then
    say "  [ ] .claude/settings.json is present but invalid"; ok=0
  fi

  # A declared knowledge root the planner cannot scan is the expensive kind of
  # silence: planning still succeeds, the tasks simply carry no references, and
  # nothing downstream can tell the difference between "nothing bound this task"
  # and "the planner could not see what did". Say it before the run spends.
  if [[ -f .claude/loop-knowledge.md ]]; then
    local blind="" unlisted="" n_roots=0 r entry f b
    while read -r r; do
      [[ -n "$r" ]] || continue
      n_roots=$((n_roots + 1))
      if [[ ! -d "$r" ]]; then blind+=" $r(no such directory)"; continue; fi
      # An index is the cheapest thing to read and the easiest to let rot: a
      # document added and never listed is invisible to the planner forever,
      # and nothing about the repo looks wrong. That is the real hazard with an
      # index -- not its format -- so coverage is checked and format is not.
      if entry="$(entry_point "$r")"; then
        while IFS= read -r f; do
          [[ -f "$f" ]] || continue
          b="$(basename "$f")"
          [[ "$b" == "$(basename "$entry")" ]] && continue   # the index itself
          grep -qF -- "$b" "$entry" || unlisted+=" ${f#./}"
        done < <(find "$r" -maxdepth 1 -name '*.md')
        continue
      fi
      # No index, so fall back to the files describing themselves -- at any
      # depth, because a root's documents are often one folder down (a handoff
      # bundle per slice, an ADR per topic) and a top-level glob would call that
      # root blind while sitting on a hundred description lines.
      find "$r" -name '*.md' -exec head -5 {} + 2>/dev/null | grep -q '^description:' && continue
      blind+=" $r"
    done < <(grep -oE '`[A-Za-z0-9_./-]+/`' .claude/loop-knowledge.md 2>/dev/null | tr -d '`' | sort -u)
    if [[ "$n_roots" -eq 0 ]]; then
      warn "  [ ] .claude/loop-knowledge.md declares no roots — the planner will cite nothing"
    elif [[ -z "$blind" ]]; then
      say "  [x] $n_roots knowledge root(s) scannable"
    else
      warn "  [ ] declared root(s) with no index and no descriptions:$blind"
      warn "      the planner will guess from filenames; add an index or description: frontmatter"
    fi
    [[ -z "$unlisted" ]] || {
      warn "  [ ] document(s) their index does not name:$unlisted"
      warn "      the planner cites from the index, so an unlisted document does not exist to it"
    }
  fi

  # The driver makes one commit per iteration. A repo with no identity
  # configured fails at the END of iteration 1, after both sessions have been
  # paid for: the most expensive place to find a one-line setup problem.
  if git config user.email >/dev/null 2>&1 && git config user.name >/dev/null 2>&1; then
    say "  [x] git identity ($(git config user.email))"
  else
    say "  [ ] git user.name/user.email not set - the first commit would fail"
    say "      fix: git config user.email you@example.com && git config user.name 'Your Name'"
    ok=0
  fi

  # A hook that rejects the driver's commit fails in the same place. We cannot
  # know whether it would pass, only that it is there to be considered.
  hookdir="$(git config core.hooksPath 2>/dev/null || echo .git/hooks)"
  if [[ -x "$hookdir/pre-commit" ]]; then
    warn "  [!] a pre-commit hook is active ($hookdir/pre-commit)"
    warn "      if it rejects the driver's commit, the run stops after paying for a task"
  fi

  [[ -n "$HOME" && -n "$USER_NAME" ]] \
    && say "  [x] masking active (\$HOME and username)" \
    || { say "  [ ] masking cannot resolve \$HOME"; ok=0; }

  if [[ "$CHECK_ONLY" -eq 1 ]]; then
    mkdir -p "$SESSIONS" 2>/dev/null && say "  [x] telemetry dir writable" \
      || { say "  [ ] cannot create telemetry dir"; ok=0; }
  else
    mkdir -p "$SESSIONS" 2>/dev/null \
      && say "  [x] telemetry dir .loop/state/runs/$RUN_PATH" \
      || { say "  [ ] cannot create telemetry dir"; ok=0; }
  fi

  if [[ -f "$STATE" ]]; then
    if jq -e . "$STATE" >/dev/null 2>&1; then say "  [x] state.json valid — $(state_get '[.tasks[]|select(.status=="pending")]|length') pending"
    else say "  [ ] state.json is not valid JSON"; ok=0; fi
  else
    say "  [x] no state — will plan first"
  fi

  [[ $ok -eq 1 ]] || die "preflight failed — nothing has run"
  say "preflight ok"
}

# A directory is usable as knowledge if it says what is in it. Three names,
# because the choice between them is the repo's and each buys something
# different: README.md is what GitHub renders when you browse a folder,
# README-<subject>.md stays unique in a flat search (an Obsidian quick-switcher,
# a backlink pane), index.md is the web convention. The loop recognises all
# three and prescribes none.
#
# Deliberately a property, not a convention: can a session find out what is in
# here without opening every file? A repo wanting the stricter rule -- every .md
# carries frontmatter, every folder carries an entry file -- enforces that in
# its own lint, where it can be strict, because the loop also installs into
# repos that will never adopt it.
entry_point() {
  # Roots are declared with a trailing slash, so normalise it away: "$d/README.md"
  # on "docs/" yields "docs//README.md", which no path comparison against find's
  # output will ever match. Cost the first time: an index reported as missing
  # from itself.
  local d="${1%/}" f
  for f in "$d/index.md" "$d/README.md"; do [[ -f "$f" ]] && { printf '%s' "$f"; return 0; }; done
  f="$(find "$d" -maxdepth 1 -name 'README-*.md' -print -quit 2>/dev/null)"
  [[ -n "$f" ]] && { printf '%s' "$f"; return 0; }
  return 1
}

# ------------------------------------------------------------------- run ----

# --check runs the same preflight the real thing runs -- the same code, not a
# reimplementation of it, because a sanity check that can disagree with the
# thing it is checking is worse than none.
#
# It takes no lock (a check must not block a running loop, nor fail because one
# is running) and writes its log outside the repo, since a check that leaves a
# run directory behind every time you fix a heading is a check people stop
# running.
if [[ "$CHECK_ONLY" -eq 1 ]]; then
  RUN_DIR="$(mktemp -d "${TMPDIR:-/tmp}/loop-check.XXXXXX")"
  SESSIONS="$RUN_DIR/sessions"
  trap 'rm -rf "$RUN_DIR" "$STATE_PRE"' EXIT
  preflight
  say ""
  if [[ "$WARN_COUNT" -eq 0 ]]; then
    say "ready. nothing was run and nothing was changed."
    exit 0
  fi
  # Non-zero on advisories even though a real run would proceed through them,
  # because the two commands are asked different questions. A run asks "can I
  # start?"; --check asks "is anything wrong?", and answering that with a silent
  # 0 makes it useless in the hook or CI job it exists to be put in.
  warn "$WARN_COUNT advisory finding(s) above — a run would still start, but read them first"
  say "nothing was run and nothing was changed."
  exit 1
fi

mkdir -p "$SESSIONS"
acquire_lock
: >"$RUN_DIR/loop.log"
: >"$ITERATIONS"
preflight

# --- plan phase (once) ---

# The journal is named for the PLAN, so a resumed run keeps appending to the
# same narrative while a different plan — on this branch or a parallel one —
# never touches this file.
open_journal() {
  JOURNAL="$STATE_DIR/journals/$(state_get .run_id).md"
  mkdir -p "$STATE_DIR/journals"
  [[ -f "$JOURNAL" ]] || printf '# Journal — %s\n\nAppend-only narrative of this plan. Rendered state lives in .loop/state/plan.md.\n' \
    "$(state_get .run_id)" >"$JOURNAL"
}

# Item 8: nothing retires a brief, so a spent one reads plannable forever.
# A run's journal is named for the brief's own stem -- the same name
# check-brief.sh keys off -- so its existence is the check: no stamp, no new
# field, nothing to keep in sync. Refused here, before the plan-reset logic
# below touches an existing state.json and before any session runs, so
# --plan-only cannot overwrite .loop/state/state.json and re-derive work this
# brief already produced. --replan is the deliberate override for a genuine
# re-plan after an aborted run.
#
# Resuming this same plan is not "planning from" the brief again -- state.json
# already holds the work this brief produced, so re-deriving it is exactly
# what does NOT happen. Skip the refusal whenever the brief asked for is the
# one state.json is already stamped with.
resuming_same_brief=0
if [[ -f "$STATE" && -n "$BRIEF" ]] && [[ "$(state_get '.brief // ""')" == "$BRIEF" ]]; then
  resuming_same_brief=1
fi
if [[ -n "$BRIEF" && "$REPLAN" -eq 0 && "$resuming_same_brief" -eq 0 ]]; then
  brief_stem="$(basename "$BRIEF" .md)"
  brief_stem="${brief_stem%.loop-brief}"
  brief_journal="$STATE_DIR/journals/$brief_stem.md"
  if [[ -f "$brief_journal" ]]; then
    die "refusing to plan: ${brief_journal#$REPO/} already exists.
  Planning would overwrite .loop/state/state.json and re-derive work this
  brief already produced. Re-plan deliberately with --replan."
  fi
fi

# A branch cut from main inherits whatever state.json the last squash left
# there — another branch's plan. It must never be resumed as if it were this
# branch's work.
#
# The discriminator is the BRIEF, not the branch. A branch name cannot tell an
# inherited plan from your own plan on a branch you renamed, and guessing wrong
# in the destructive direction loses a run. The brief names which plan you are
# asking for, so it answers the question directly.
if [[ -f "$STATE" ]]; then
  STATE_BRIEF="$(state_get '.brief // ""')"
  STATE_BRANCH="$(state_get '.branch // ""')"
  if [[ -n "$BRIEF" ]]; then
    if [[ -n "$STATE_BRIEF" && "$STATE_BRIEF" != "$BRIEF" ]]; then
      say "state.json holds plan $(state_get .run_id) for '$STATE_BRIEF'"
      say "  you asked for '$BRIEF' — resetting and planning fresh"
      rm -f "$STATE" "$STATE_DIR/plan.md"
    fi
    # Same brief: this is the plan you asked for. Resume it whatever branch it
    # was stamped on — that is how a renamed branch recovers, without the
    # operator having to reach for the one flag that would destroy the run.
  elif [[ -n "$STATE_BRANCH" && "$STATE_BRANCH" != "$BRANCH" ]]; then
    die "state.json holds plan $(state_get .run_id), stamped on branch '$STATE_BRANCH'; you are on '$BRANCH'.
  With no brief there is no way to tell an inherited plan from your own on a
  renamed branch, and the two want opposite things. Say which you mean:
    resume it            .loop/run.sh $STATE_BRIEF
    start a new plan     .loop/run.sh docs/briefs/<other>.md"
  fi
fi

if [[ ! -f "$STATE" ]]; then
  if [[ -z "$BRIEF" ]]; then
    BRIEF="$(ls -1 docs/briefs/*.md 2>/dev/null | tail -1)"
    [[ -n "$BRIEF" ]] || die "no state and no brief. usage: .loop/run.sh docs/briefs/NNNN-slug.md"
  fi
  [[ -f "$BRIEF" ]] || die "brief not found: $BRIEF"

  say "planning from $BRIEF using $PLAN_MODEL"
  run_session plan 0 "$PLAN_MODEL" "/loop-plan $BRIEF" \
    || die "planning session failed — see .loop/state/runs/$RUN_PATH/"

  [[ -f "$STATE" ]] || die "planning produced no .loop/state/state.json"
  jq -e . "$STATE" >/dev/null 2>&1 || die ".loop/state/state.json is not valid JSON"

  # The plan skill is told to check its own work; this is the driver checking it
  # anyway, because a plan that cannot be validated is a planning failure and
  # every downstream session would inherit it.
  n_bad="$(state_get '[.tasks[]|select((.verify//"")=="" or ((.acceptance//[])|length)==0)]|length')"
  [[ "$n_bad" -eq 0 ]] || die "$n_bad task(s) have no verify command or no acceptance criteria — fix the plan"
  bad_dep="$(state_get '[.tasks[].id] as $ids | [.tasks[]|.depends_on[]?|select(($ids|index(.))==null)] | length')"
  [[ "$bad_dep" -eq 0 ]] || die "$bad_dep dependency reference(s) name a task that does not exist"

  # Dangling-reference lint. A task's references are what the work session is
  # told to read and what the review session holds it against, so a path that
  # does not resolve is worse than no reference at all: the work session cannot
  # recover, and it costs an attempt to discover that. Cheap to check here,
  # expensive to find at iteration 7.
  bad_ref=""; blind_ref=""
  while IFS=$'\t' read -r id ref; do
    [[ -n "$id" && -n "$ref" ]] || continue
    if [[ ! -e "$ref" ]]; then
      bad_ref+="    $id  $ref"$'\n'
    elif [[ -d "$ref" ]] && ! entry_point "$ref" >/dev/null; then
      # A folder reference is right for something that only makes sense whole --
      # a design handoff bundle, a spec and its diagrams. But handed a directory
      # with no way in, a session opens files until it thinks it understands,
      # which is the expensive kind of guessing. Not fatal: the plan is sound and
      # the docs are not, and failing here would punish the planner for something
      # it cannot fix.
      blind_ref+="    $id  $ref"$'\n'
    fi
  done < <(state_get '.tasks[] as $t | ($t.references // [])[] | [$t.id, .path] | @tsv')
  [[ -z "$blind_ref" ]] || warn "folder reference(s) with no index.md, README.md or README-*.md:
$blind_ref  a session handed one of these opens files until it guesses right"
  [[ -z "$bad_ref" ]] || die "task reference(s) do not resolve -- fix the plan:

$bad_ref
  A cited file that is not there stops the session it was cited to, which has
  no way to recover and burns an attempt finding out. Cite what exists."

  # Gate-shape lint. A gate that parses a structure and then substring-matches
  # its re-serialised text has thrown away the parse -- and that is not a style
  # preference. It rejects correct implementations (a $ref does not contain the
  # property name) while accepting corrupted ones (a hand-written duplicate
  # does). Both halves of that were observed in one real run.
  #
  # This is checkable precisely because the compliant form is essentially
  # unique: navigate to the value and assert on it. The rule cannot be stated
  # unambiguously in prose -- so it is stated as a check instead.
  bad_gate=""
  while IFS=$'\t' read -r id cmd; do
    [[ -n "$id" ]] || continue
    if grep -qE 'JSON\.stringify\([^)]*\)\.(indexOf|includes|match|search)\(|json\.dumps\([^)]*\)\.(find|index)\(|in json\.dumps\(' <<<"$cmd"; then
      bad_gate+="    $id  re-serialises a parsed structure and substring-matches the text"$'\n'
    elif grep -qE '(readFileSync|open)\([^)]*src/|src/[^)]*\)[[:space:]]*\.[[:space:]]*read_text\(|grep [^|]*[[:space:]]src/' <<<"$cmd"; then
      bad_gate+="    $id  asserts on source text rather than on what the program does"$'\n'
    fi
  done < <(state_get '.tasks[] | [.id, .verify] | @tsv')
  [[ -z "$bad_gate" ]] || die "gate shape rejected -- fix the plan:

$bad_gate
  Navigate to the value and assert on it, resolving a \$ref when there is one:

    const s = op.requestBody.content['application/json'].schema;
    const r = s.\$ref ? doc.components.schemas[s.\$ref.split('/').pop()] : s;
    need(r?.properties?.url, 'the documented request body has no url property');

  A substring over re-serialised JSON passes on a hand-written duplicate and
  fails on the idiomatic form -- exactly backwards. Matching on text that was
  never parsed (an HTML page, a log line) is fine and is not flagged."

  # Gate-shape lint, rule 3. Not a heuristic: the condition is
  # gate_files_moved() (above) evaluated one phase early. That function
  # reverts any HEAD-existing file the current task's verify names and its
  # files does not, and the revert runs before the gate does -- so a task in
  # that shape is unpassable by construction, for any implementation, forever.
  # Rule 2 already draws the inspect-vs-execute line for src/; this extends it
  # to any HEAD-tracked file, task-relative rather than path-prefix-relative.
  bad_gate3=""
  while IFS=$'\t' read -r id cmd; do
    [[ -n "$id" ]] || continue
    while IFS= read -r f; do
      [[ -n "$f" ]] || continue
      git cat-file -e "HEAD:$f" 2>/dev/null || continue
      jq -e --arg f "$f" --arg t "$id" '
          (.tasks[] | select(.id == $t)) as $cur
          | (($cur.files // []) | index($f)) == null
        ' "$STATE" >/dev/null 2>&1 &&
        bad_gate3+="    $id  inspects $f, which it does not own; the driver reverts that file before the gate runs, so no implementation can pass"$'\n'
    done < <(gate_inspected_paths "$cmd")
  done < <(state_get '.tasks[] | [.id, .verify] | @tsv')
  [[ -z "$bad_gate3" ]] || die "gate shape rejected -- fix the plan:

$bad_gate3
  Two escape hatches, both correct outcomes: add the file to that task's
  files, or move the claim to acceptance, where the review session can read
  provenance instead of the gate re-asserting it. Handing the path to a
  runner (uv run pytest -q path, bash path) rather than reading it is fine
  and is not flagged."

  # Gate-shape lint, rule 4. A gate re-runs for the life of the plan, so a
  # baseline fixed at plan time decays the moment any other task commits --
  # failure E in the brief this rule answers, twice over. git diff/log/rev-list
  # against HEAD stays sound and unflagged: a work session cannot commit, so
  # HEAD still discriminates a session's edits from committed history, which is
  # exactly what the runtime gate-rewrite guard depends on.
  bad_gate4=""
  while IFS=$'\t' read -r id cmd; do
    [[ -n "$id" ]] || continue
    while IFS= read -r ref; do
      [[ -n "$ref" ]] || continue
      bad_gate4+="    $id  diffs against $ref rather than HEAD; a gate re-runs for the life of the plan and that baseline decays on the next commit"$'\n'
    done < <(gate_diff_refs "$cmd")
  done < <(state_get '.tasks[] | [.id, .verify] | @tsv')
  [[ -z "$bad_gate4" ]] || die "gate shape rejected -- fix the plan:

$bad_gate4
  Diff against HEAD instead. It is the one baseline a gate can hold for the
  life of the plan without decaying: a work session cannot commit, so HEAD
  still separates a session's edits from everything committed before it."

  # The driver stamps both, rather than trusting the plan session to record
  # them: which branch and which brief a plan belongs to are facts the driver
  # already holds, and the brief is now what decides whether a later run
  # resumes this plan or resets it.
  state_edit --arg b "$BRANCH" --arg f "$BRIEF" '.branch = $b | .brief = $f'
  say "planned: $(state_get '"\(.run_id) — \(.tasks|length) tasks"')"
  open_journal

  # The plan phase's report carries the planner's "what I interpreted rather
  # than read" list — the operator's one cheap chance to catch a misreading
  # before every iteration inherits it. It is otherwise reachable only as a
  # long string inside a session JSON blob, which nobody reads on a phone.
  {
    printf '\n## Plan — %s\n\n' "$(state_get .run_id)"
    printf -- '- **Brief:** `%s`\n' "$BRIEF"
    printf -- '- **Tasks:** %s\n\n' "$(state_get '.tasks|length')"
    jq -r '.result // ""' "$SESSIONS"/*-plan.json 2>/dev/null | mask
    printf '\n'
  } >>"$JOURNAL"

  render_plan
  git add -A && git commit -q -m "[loop] plan $(state_get .run_id)" && say "committed the plan"

  # --plan-only stops here, having paid for exactly one session.
  #
  # The plan is the highest-leverage artefact in a run and the cheapest thing to
  # get wrong: every gate the rest of the run is measured against was authored
  # in that one session. Reading it before spending on iterations is the whole
  # point, and it used to mean LOOP_MAX_ITERATIONS=0 -- which works, but reads
  # like a workaround and invites a typo in the one place a typo is expensive.
  if [[ "$PLAN_ONLY" -eq 1 ]]; then
    say ""
    say "═══ plan only ═══"
    say "plan:   $(state_get .run_id) — $(state_get '.tasks|length') task(s)"
    say "first:  $(state_get '[.tasks[]|select(.status=="pending" and ((.depends_on//[])|length)==0)][0].id // "none"')"
    say "read:   .loop/state/plan.md   (verify commands are in .loop/state/state.json)"
    say "adjust: .loop/amend.sh"
    say "run it: .loop/run.sh"
    exit 0
  fi
else
  if [[ "$PLAN_ONLY" -eq 1 ]]; then
    say "--plan-only, but $(state_get .run_id) is already planned — nothing to do"
    say "read:   .loop/state/plan.md"
    say "run it: .loop/run.sh"
    exit 0
  fi
  say "resuming $(state_get .run_id) — $(state_get '[.tasks[]|select(.status=="done")]|length')/$(state_get '.tasks|length') done"
  state_edit --arg t "$(ts)" --arg b "$BRANCH" '.status="running" | .updated=$t | .branch=$b'
  open_journal
fi

# --- iterate ---

run_iters=0
stalls=0
status="max_iterations"
exit_code=4
blocked_gate_pass_tasks=()
no_proposal_tasks=()

while true; do
  pending="$(state_get '[.tasks[]|select(.status=="pending")]|length')"
  blocked="$(state_get '[.tasks[]|select(.status=="blocked")]|length')"
  done_n="$(state_get '[.tasks[]|select(.status=="done")]|length')"
  total="$(state_get '.tasks|length')"

  if [[ "$pending" -eq 0 ]]; then
    if [[ "$blocked" -gt 0 ]]; then status="blocked"; exit_code=2
    else status="complete"; exit_code=0; fi
    break
  fi

  # Budgets are checked here, between iterations — never mid-iteration. A run
  # always stops with state coherent, so raising a limit and re-running just
  # works. They are per-run, not per-plan: a runaway backstop, not a
  # convergence detector. That job belongs to the signal below.
  if [[ "$run_iters" -ge "$MAX_ITER" ]]; then status="max_iterations"; exit_code=4; break; fi

  spend="$(sig_spend)"
  if awk -v s="$spend" -v c="$COST_CEILING" 'BEGIN{exit !(s>=c)}'; then
    status="cost_ceiling"; exit_code=6; break
  fi

  iters="$(sig_iterations)"
  if [[ "${iters:-0}" -ge "$CONVERGENCE_MIN" ]]; then
    pc="$(sig_per_closed)"
    if [[ "$pc" == "n/a" ]] || awk -v p="$pc" -v m="$CONVERGENCE_MAX" 'BEGIN{exit !(p>m)}'; then
      status="not_converging"; exit_code=5; break
    fi
  fi

  # Next ready task: first pending task whose dependencies are all done. The
  # driver picks it; the work session is told which one and does not choose.
  task="$(state_get '[.tasks[]|select(.status=="done")|.id] as $d
                     | [.tasks[]|select(.status=="pending")
                        | select(([.depends_on[]?|select(($d|index(.))==null)]|length)==0)][0].id // empty')"
  if [[ -z "$task" ]]; then
    warn "$pending task(s) pending but none are ready — dependencies cannot be satisfied"
    status="blocked"; exit_code=2; break
  fi

  run_iters=$((run_iters + 1))
  iter="$(( $(state_get .iteration) + 1 ))"
  say "── iteration $iter ($run_iters/$MAX_ITER this run) · $task · $done_n/$total done ──"
  say "   $(state_get --arg id "$task" '.tasks[]|select(.id==$id)|.title')"

  # A stale file from a previous iteration must never be mistaken for this
  # iteration's report.
  rm -f "$PROPOSAL" "$VERDICT"

  # 1. work session
  state_snapshot
  run_session work "$iter" "$WORK_MODEL" "/loop-work $task"
  if [[ $? -ne 0 && ! -f "$PROPOSAL" ]]; then
    status="session_error"; exit_code=7; break
  fi

  if [[ ! -f "$PROPOSAL" ]] || ! jq -e . "$PROPOSAL" >/dev/null 2>&1; then
    warn "work session left no valid proposal"
    changed=()
    while IFS= read -r f; do [[ -n "$f" ]] && changed+=("$f"); done < <(session_tree_changes)
    if [[ ${#changed[@]} -gt 0 ]]; then
      account="work session left no valid proposal; before dying it changed $(IFS=', '; echo "${changed[*]}") — the driver committed them under this iteration"
    else
      account="work session left no valid proposal; the working tree is unchanged"
    fi
    account="$(mask <<<"$account")"
    [[ ${#changed[@]} -gt 0 ]] && no_proposal_tasks+=("$task — $account")
    outcome="blocked"; summary="$account"; notes="none"
  else
    outcome="$(jq -r '.outcome // "blocked"' "$PROPOSAL")"
    summary="$(jq -r '.summary // ""' "$PROPOSAL" | mask)"
    notes="$(jq -r '.notes // "none"' "$PROPOSAL" | mask)"
  fi

  # The plan is not the session's to edit, and moving a goalpost is a finding
  # whatever the code looks like -- including when the rewrite is an improvement,
  # because the loop cannot tell a good rewrite from a bad one without spending
  # a review on it, which is the cost the pre-authored gate exists to avoid.
  tampered=""
  if state_restore_if_touched; then
    tampered="work session modified .loop/state/state.json — restored by the driver; the plan and its verify commands are not a session's to edit"
    warn "   STATE TAMPERING $task — .loop/state/state.json was modified; restored, iteration failed"
  fi

  # Restored from HEAD rather than reported and left alone: the driver commits
  # every iteration, so a weakened gate that survives this one guards its task
  # for the rest of the run -- and keeps passing, which no gate re-run can see.
  moved=()
  while IFS= read -r f; do [[ -n "$f" ]] && moved+=("$f"); done < <(gate_files_moved)
  if [[ ${#moved[@]} -gt 0 ]]; then
    git checkout HEAD -- "${moved[@]}" 2>/dev/null
    tampered="${tampered:+$tampered; }work session modified ${moved[*]} — a file a verify command runs, which this task neither created nor was assigned; restored by the driver"
    warn "   GATE REWRITE $task — ${moved[*]} restored from HEAD; a gate is not a session's to rewrite"
  fi

  # 2. gate — every done task, plus this one if it claims to be done or
  # reports blocked: a blocked outcome with a passing gate is a defect in the
  # plan or the block itself, not in the work, and the only way to tell that
  # from a genuine block is to run the gate rather than skip it.
  gate_targets=()
  while read -r id; do [[ -n "$id" ]] && gate_targets+=("$id"); done \
    < <(state_get '.tasks[]|select(.status=="done")|.id')
  [[ "$outcome" == "done" || "$outcome" == "blocked" ]] && gate_targets+=("$task")

  gate_failed=()
  if [[ ${#gate_targets[@]} -gt 0 ]]; then
    while read -r id; do [[ -n "$id" ]] && gate_failed+=("$id"); done \
      < <(gate_ids "$task" "${gate_targets[@]}")
  fi

  # Gate logs are keyed by task id and overwritten every iteration, so a failing
  # log would be erased by the next passing run of the same task — losing the
  # only record of the failure. Keep a per-iteration copy of the failures.
  for id in "${gate_failed[@]:-}"; do
    [[ -n "$id" ]] || continue
    cp "$RUN_DIR/gates/$id.log" "$RUN_DIR/gates/$(printf '%03d' "$iter")-$id.fail.log" 2>/dev/null
  done

  # A previously-done task that no longer verifies is a regression: revert it,
  # charge it an attempt. This is the check that per-task gates alone cannot do.
  for id in "${gate_failed[@]:-}"; do
    [[ -n "$id" && "$id" != "$task" ]] || continue
    warn "   GATE REGRESSION $id — reverting to pending; detected during $task's iteration"
    state_edit --arg id "$id" --arg n "regressed: verify failed during $task — see .loop/state/runs/$RUN_PATH/gates/$id.log" \
      '(.tasks[]|select(.id==$id)) |= (.status="pending" | .attempts=(.attempts+1) | .notes=$n)'
  done

  candidate_failed=0
  for id in "${gate_failed[@]:-}"; do [[ "$id" == "$task" ]] && candidate_failed=1; done

  # A blocked task whose own gate passes on this re-run is a signal, not a
  # verdict: the work satisfies the gate, so whatever the session could not do
  # is about something else. This never changes the status transition below.
  blocked_gate_passed=0
  gate_pass_line="$task GATE PASSES while the session reports blocked — the work satisfies its own gate; the block is about something else. Read the proposal."
  [[ "$outcome" == "blocked" && $candidate_failed -eq 0 ]] && blocked_gate_passed=1

  verdict="skipped"
  if [[ -n "$tampered" ]]; then
    outcome="gate_fail"
    warn "   $task failed on state tampering — the work is reverted whatever the gate said"
  elif [[ "$outcome" == "blocked" ]]; then
    say "   work session reported blocked"
    if [[ $blocked_gate_passed -eq 1 ]]; then
      warn "   $gate_pass_line"
      blocked_gate_pass_tasks+=("$task")
    fi
  elif [[ $candidate_failed -eq 1 ]]; then
    outcome="gate_fail"
    warn "   GATE FAIL $task — review skipped, work that fails its own gate is not reviewable"
  else
    # 3. review session — separate, read-only, sees the diff and not the summary
    state_snapshot
    run_session review "$iter" "$WORK_MODEL" "/loop-review $task"
    if state_restore_if_touched; then
      warn "   STATE TAMPERING $task — review session modified .loop/state/state.json; restored"
      tampered="review session modified .loop/state/state.json — restored by the driver; only the driver makes status transitions"
    fi
    if [[ ! -f "$VERDICT" ]] || ! jq -e . "$VERDICT" >/dev/null 2>&1; then
      warn "   review session left no valid verdict — treating as FAIL"
      verdict="FAIL"
    else
      verdict="$(jq -r '.verdict // "FAIL"' "$VERDICT")"
    fi
    [[ "$verdict" == "PASS" ]] || outcome="review_fail"
    say "   review: $verdict"
  fi

  # 4. apply — the driver makes every status transition
  case "$outcome" in
    done)
      state_edit --arg id "$task" '(.tasks[]|select(.id==$id)) |= (.status="done" | .notes="")'
      say "   $task done" ;;
    gate_fail|review_fail)
      if [[ -n "$tampered" ]]; then
        reason="$tampered"
      elif [[ "$outcome" == "gate_fail" ]]; then
        reason="gate failed — see .loop/state/runs/$RUN_PATH/gates/$task.log"
      else
        reason="$(jq -r '(.findings // []) | join("; ")' "$VERDICT" 2>/dev/null | mask)"
      fi
      state_edit --arg id "$task" --arg n "$reason" \
        '(.tasks[]|select(.id==$id)) |= (.status="pending" | .attempts=(.attempts+1) | .notes=$n)' ;;
    blocked)
      blocked_note="$summary"
      [[ $blocked_gate_passed -eq 1 ]] && blocked_note="$gate_pass_line $summary"
      state_edit --arg id "$task" --arg n "$blocked_note" \
        '(.tasks[]|select(.id==$id)) |= (.status="pending" | .attempts=(.attempts+1) | .notes=$n)' ;;
  esac

  # Item 7: a second consecutive `blocked` outcome for the SAME task, with
  # nothing outside the driver's own bookkeeping under .loop/state/ changed in
  # the repo since the first, cannot carry new information -- a memoryless
  # session given identical inputs reaches an identical conclusion. Kept on the
  # task itself (not "the previous iteration of the run") so it still applies
  # correctly when a different ready task was worked in between, and it
  # survives a resumed run because state.json does. blocked_head is HEAD the
  # last time this task ended blocked, captured before that iteration's own
  # commit; HEAD now is still pre-commit too, so the diff between them is
  # exactly what changed between the start of those two sessions.
  repeat_blocked_halt=0
  if [[ "$outcome" == "blocked" ]]; then
    head_now="$(git rev-parse HEAD 2>/dev/null)"
    prev_head="$(state_get --arg id "$task" '.tasks[]|select(.id==$id)|.blocked_head // empty')"
    if [[ -n "$prev_head" ]]; then
      changed="$(git diff --name-only "$prev_head" "$head_now" 2>/dev/null | grep -v '^\.loop/state/' || true)"
      if [[ -z "$changed" ]]; then
        prev_summary="$(state_get --arg id "$task" '.tasks[]|select(.id==$id)|.blocked_summary // empty')"
        repeat_blocked_halt=1
        repeat_blocked_msg="$task blocked twice with nothing changed since the first attempt — halting rather than spending a third identical session. First diagnosis: ${prev_summary:-none}"
      fi
    fi
    state_edit --arg id "$task" --arg h "$head_now" --arg s "$summary" \
      '(.tasks[]|select(.id==$id)) |= (.blocked_head=$h | .blocked_summary=$s)'
  else
    state_edit --arg id "$task" \
      '(.tasks[]|select(.id==$id)) |= (del(.blocked_head) | del(.blocked_summary))'
  fi

  # A task that has burned its attempts is blocked, not retried forever.
  n_new="$(state_get --argjson m "$MAX_ATTEMPTS" '[.tasks[]|select(.status=="pending" and .attempts>=$m)]|length')"
  if [[ "$n_new" -gt 0 ]]; then
    state_edit --argjson m "$MAX_ATTEMPTS" '(.tasks[]|select(.status=="pending" and .attempts>=$m)) |= (.status="blocked")'
    warn "   $n_new task(s) hit the attempt ceiling — blocked"
  fi

  state_edit --arg t "$(ts)" --argjson i "$iter" '.iteration=$i | .updated=$t'

  # 5. record — telemetry, journal, evidence
  new_done="$(state_get '[.tasks[]|select(.status=="done")]|length')"
  jq -nc --argjson i "$iter" --arg t "$task" --arg o "$outcome" \
     --argjson a "$(state_get --arg id "$task" '.tasks[]|select(.id==$id)|.attempts')" \
     --argjson d "$new_done" --argjson n "$total" \
     '{iteration:$i, task:$t, outcome:$o, attempts:$a, tasks_done:$d, tasks_total:$n}' >>"$ITERATIONS"

  mkdir -p "$RUN_DIR/reports"
  [[ -f "$PROPOSAL" ]] && mask <"$PROPOSAL" >"$RUN_DIR/reports/$(printf '%03d' "$iter")-proposal.json"
  [[ -f "$VERDICT" ]]  && mask <"$VERDICT"  >"$RUN_DIR/reports/$(printf '%03d' "$iter")-verdict.json"

  # The journal's prose is the sessions'; assembling it is the driver's, so the
  # entry always carries the verdict and always passes through the mask.
  {
    printf '\n## %s — %s\n\n' "$task" "$(state_get --arg id "$task" '.tasks[]|select(.id==$id)|.title')"
    printf -- '- **Outcome:** %s (review: %s)\n' "$outcome" "$verdict"
    printf -- '- **Summary:** %s\n' "${summary:-none}"
    printf -- '- **Files:** %s\n' "$(jq -r '(.files // []) | join(", ")' "$PROPOSAL" 2>/dev/null | mask || echo none)"
    printf -- '- **Notes for next iteration:** %s\n' "${notes:-none}"
  } >>"$JOURNAL"

  render_plan

  # 6. commit — one per iteration, driver-owned, covering code + state +
  # journal + telemetry together. Agents never commit, so the history is a
  # record of what the loop decided rather than of what a session claimed.
  git add -A >/dev/null 2>&1
  if ! git diff --cached --quiet 2>/dev/null; then
    git commit -q -m "[loop] $task: $outcome" && say "   committed $(git rev-parse --short HEAD)"
  fi

  print_signals

  if [[ $repeat_blocked_halt -eq 1 ]]; then
    warn "   $repeat_blocked_msg"
    status="repeat_blocked"; exit_code=8
    break
  fi

  # Stall detection: an iteration that closed nothing and burned no attempt has
  # made no recorded progress at all.
  if [[ "$new_done" -le "$done_n" && "$outcome" != "gate_fail" && "$outcome" != "review_fail" ]]; then
    stalls=$((stalls + 1))
    warn "   no recorded progress ($stalls/$STALL_LIMIT)"
    if [[ "$stalls" -ge "$STALL_LIMIT" ]]; then status="stalled"; exit_code=3; break; fi
  else
    stalls=0
  fi
done

# --- report ---

state_edit --arg s "$status" --arg t "$(ts)" '.status=$s | .updated=$t'

say ""
say "═══ $status ═══"
say "run:    $RUN_PATH  ($run_iters iteration(s) this run)"
say "plan:   $(state_get '[.tasks[]|select(.status=="done")]|length')/$(state_get '.tasks|length') done, $(state_get '[.tasks[]|select(.status=="blocked")]|length') blocked"
print_signals
say ""
say "per-phase cost:"
for p in plan work review; do
  jq -s --arg p "$p" '[.[]|select(.phase==$p)] | "  \($p): \(length) session(s), $\([.[].total_cost_usd//0]|add//0|.*100|round/100), \([.[].num_turns//0]|add//0) turns"' \
    "$SESSIONS"/*.json 2>/dev/null | tr -d '"' | while read -r l; do say "$l"; done
done
if [[ ${#blocked_gate_pass_tasks[@]} -gt 0 ]]; then
  say ""
  for id in "${blocked_gate_pass_tasks[@]}"; do
    say "$id GATE PASSES while the session reports blocked — the work satisfies its own gate; the block is about something else."
  done
fi
if [[ ${#no_proposal_tasks[@]} -gt 0 ]]; then
  say ""
  for line in "${no_proposal_tasks[@]}"; do
    say "$line"
  done
fi
say ""
case "$status" in
  complete)       say "plan complete. journal: ${JOURNAL#$REPO/}" ;;
  blocked)        say "a human is needed. read the blocked task's notes in .loop/state/state.json" ;;
  stalled)        say "no recorded progress twice running — read .loop/state/runs/$RUN_PATH/" ;;
  max_iterations) say "iteration budget spent. resumable: re-run .loop/run.sh" ;;
  cost_ceiling)   say "cost ceiling reached. resumable: raise LOOP_COST_CEILING and re-run" ;;
  not_converging) say "iterations-per-closed-task exceeded $CONVERGENCE_MAX — the run is not converging." ;;
  session_error)  say "a claude session failed. see .loop/state/runs/$RUN_PATH/" ;;
  repeat_blocked) say "$repeat_blocked_msg" ;;
esac
render_plan

# Every session writes a stderr file; almost all are empty, and committing 36
# empty files per run buries the ones that are not. A non-empty stderr is the
# evidence you want when a session dies, so keep those and drop the rest.
find "$RUN_DIR" -name '*.stderr' -empty -delete 2>/dev/null

{
  printf '\n## Run ended — %s\n\n' "$status"
  printf -- '- **Run:** `%s` · %s iteration(s) this run\n' "$RUN_ID" "$run_iters"
  printf -- '- **Plan:** %s/%s done, %s blocked\n' \
    "$(state_get '[.tasks[]|select(.status=="done")]|length')" \
    "$(state_get '.tasks|length')" \
    "$(state_get '[.tasks[]|select(.status=="blocked")]|length')"
  printf -- '- **Signals:** %s iterations · %s per closed · %s gate failure(s) · %s review rejection(s) · %s attempt(s) burned · streak %s · ~$%s\n' \
    "$(sig_iterations)" "$(sig_per_closed)" "$(sig_gate_fails)" "$(sig_review_fails)" \
    "$(sig_attempts)" "$(sig_streak)" "$(printf '%.2f' "$(sig_spend)")"
} >>"$JOURNAL"

git add -A >/dev/null 2>&1
git diff --cached --quiet 2>/dev/null || git commit -q -m "[loop] run $RUN_PATH: $status"

exit "$exit_code"
