#!/usr/bin/env bash
#
# Render .loop/state/state.json as markdown for human reading.
#
#   .loop/render-plan.sh [state.json] [out.md]
#
# The design note this loop is built from is explicit that structured state
# needs a rendered view — "markdown rendered from state for human and mobile
# reading, never edited as the source of truth"
# (docs/references/executable-loop-harness.md). JSON is the right thing for the
# driver to compute over and the wrong thing to read on a phone.
#
# One direction only: state.json -> plan.md. Nothing ever parses plan.md back.
# The driver calls this after every state change, so the file is always current
# and hand edits to it are always lost.

set -uo pipefail

STATE="${1:-$(dirname "${BASH_SOURCE[0]}")/state/state.json}"
OUT="${2:-$(dirname "$STATE")/plan.md}"

[[ -f "$STATE" ]] || { echo "render-plan: no state at $STATE" >&2; exit 2; }
jq -e . "$STATE" >/dev/null 2>&1 || { echo "render-plan: $STATE is not valid JSON" >&2; exit 2; }

TMP="$(mktemp "${TMPDIR:-/tmp}/plan.XXXXXX")"

# `files` is rendered even though the driver never reads it back, because the
# operator does. --plan-only exists so a human can read the plan before the run
# spends anything on it, and one of the few things worth catching there -- a task
# that ships behaviour but names nothing that outlives the run -- was only
# visible in state.json. A rule the driver does not enforce has to be at least
# legible in the artifact people are told to read.

jq -r '
  def mark: if . == "done" then "x" else " " end;
  def note(t):
    if t.status == "blocked" then " · **blocked**"
    elif (t.attempts // 0) > 0 then " · \(t.attempts) attempt(s)"
    else "" end;

  "# Plan — \(.run_id)",
  "",
  "<!-- Rendered from .loop/state/state.json by .loop/render-plan.sh.",
  "     Do NOT edit: regenerated on every state change, your edits will be lost.",
  "     The source of truth is .loop/state/state.json. -->",
  "",
  "**Status:** \(.status) · **\([.tasks[] | select(.status == "done")] | length)/\(.tasks | length) done** · iteration \(.iteration)",
  "",
  "**Brief:** `\(.brief)` · **Updated:** \(.updated)",
  "",
  "## Progress",
  "",
  (.tasks[] | "- [\(.status | mark)] **\(.id)** — \(.title)\(note(.))"),
  "",
  "## Tasks",
  "",
  (.tasks[] |
    "### \(.id) — \(.title)",
    "",
    "`\(.status)`\(note(.)) · depends on: \(if (.depends_on | length) == 0 then "none" else (.depends_on | join(", ")) end)",
    "",
    (if ((.files // []) | length) > 0
     then ("**Files:** " + ([.files[] | "`" + . + "`"] | join(", ")), "")
     else ("**Files:** _none named_", "") end),
    (.goal // "_no goal recorded_"),
    "",
    "**Acceptance**",
    "",
    (.acceptance[]? | "- \(.)"),
    "",
    (if (.notes // "") != "" then ("**From the last attempt:** \(.notes)", "") else empty end),
    "<details><summary>verify command</summary>",
    "",
    "```sh",
    (.verify // ""),
    "```",
    "",
    "</details>",
    ""
  )
' "$STATE" >"$TMP" || { rm -f "$TMP"; echo "render-plan: jq failed" >&2; exit 2; }

mv "$TMP" "$OUT"
