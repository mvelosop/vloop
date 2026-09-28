#!/usr/bin/env bash
# A documented path under .loop/tmp/ is checked against its PRODUCER, not the disk.
#
# That directory is gitignored and empty in a clean checkout, so "does it exist"
# can only ever answer by accident — and it did. `.loop/tmp/proposal.json`
# passed in six documents because a calibration fixture happens to be named
# proposal.json; `.loop/tmp/verdict.json`, with no namesake anywhere, failed in
# four. Same rule, opposite answers, and neither had anything to do with whether
# run.sh still produced the file.
#
# Both directions are asserted here because fixing only the noisy one leaves the
# accident intact: the first attempt at this exempted transients that run.sh
# names and let everything else fall through to the filesystem, which meant a
# renamed proposal.json still passed on the strength of its decoy.
. "$(dirname "$0")/../lib.sh"

CHECK="$(cd "$(dirname "$0")/.." && pwd)/check-docs.sh"
TREE=""
cleanup_tree() { [[ -n "$TREE" && "$TREE" == */loopdocs.* ]] && rm -rf "$TREE"; }
trap 'cleanup_tree; fixture_cleanup' EXIT

# A tree small enough to reason about: one run.sh that produces one transient,
# one doc that cites it by the name run.sh USED to produce, and a decoy file
# elsewhere carrying that basename. The decoy is the whole point — it is the
# unrelated namesake that used to decide the answer.
plant() {
  cleanup_tree
  TREE="$(mktemp -d "${TMPDIR:-/tmp}/loopdocs.XXXXXX")"
  mkdir -p "$TREE/.loop" "$TREE/elsewhere"
  cat >"$TREE/.loop/run.sh" <<RUNSH
TMP_DIR="\$LOOP_DIR/tmp"
PROPOSAL="\$TMP_DIR/$1"
RUNSH
  printf 'The work session writes `.loop/tmp/proposal.json`.\n' >"$TREE/d.md"
  : >"$TREE/elsewhere/proposal.json"
}

note "── a transient run.sh still produces is live ──"
plant proposal.json
if bash "$CHECK" "$TREE" >"$TREE/out.log" 2>&1; then ok "live transient passes"
else bad "a transient run.sh produces was called dead"; sed 's/^/      /' "$TREE/out.log"; fi

note "── rename it in run.sh and the reference is stale, decoy or no decoy ──"
plant handoff.json
if bash "$CHECK" "$TREE" >"$TREE/out.log" 2>&1; then
  bad "a transient run.sh no longer produces still passed — the namesake decided it"
  sed 's/^/      /' "$TREE/out.log"
else
  ok "stale transient fails"
  grep -q 'dead path' "$TREE/out.log" && ok "log: dead path" || bad "no 'dead path' in the log"
  grep -q '.loop/tmp/proposal.json' "$TREE/out.log" && ok "log: names the stale path" \
    || bad "the log does not name the stale path"
fi

finish
