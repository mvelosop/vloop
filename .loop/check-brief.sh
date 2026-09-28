#!/usr/bin/env bash
#
# Check a brief before you spend anything on it.
#
#   .loop/check-brief.sh docs/briefs/0003-runstat-cli.md
#   .loop/check-brief.sh docs/briefs/*.md
#
# The brief is the highest-leverage artefact in the loop: every gate the planner
# writes is derived from it, and a vague brief produces weak gates that quietly
# lower the bar for the whole run. It was also, until this script, the only
# major artefact with no validation at all — the plan, the docs, the driver and
# the reviewer all have one.
#
# This checks the structure a brief needs. It cannot check the thing that
# matters most, which is whether the brief pins DECISIONS and leaves MECHANICS
# open. That stays a judgement, and the manual's "Writing a brief" section is
# where it lives. A brief can pass every check here and still be bad.
#
# Exit 0 clean, 1 with problems. Warnings do not fail.

set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.." || exit 1

fails=0

check_one() {
  local f="$1" problems=0 warnings=0
  local body; body="$(cat "$f")"

  # Only briefs that declare themselves plannable are checked. A discussion
  # document or a finished design record is not a worse brief, it is a
  # different kind of document, and flagging it forever teaches people to
  # ignore the checker. Positive marker, not an inference from absence.
  # tolerate the metadata being written as a list; it is the same statement
  if ! grep -qiE '^[-*]? *\*\*Status:\*\* *(ready to plan|ready to execute)' <<<"$body"; then
    printf '\n\033[1m%s\033[0m\n  \033[36m-\033[0m skipped: not marked "ready to plan"\n' "$f"
    return
  fi

  printf '\n\033[1m%s\033[0m\n' "$f"

  bad()  { printf '  \033[31m✗\033[0m %s\n' "$*"; problems=$((problems+1)); }
  warn() { printf '  \033[33m!\033[0m %s\n' "$*"; warnings=$((warnings+1)); }
  ok()   { printf '  \033[32m✓\033[0m %s\n' "$*"; }

  # Nothing retires a brief, so a spent one reads plannable forever. A run's
  # journal is named for the brief's own stem -- the number-and-slug a
  # planner writes as run_id -- so its existence is the check: no stamp, no
  # new field. A `.loop-brief.md` brief's stem drops that suffix too, since
  # that is the run id its planner actually wrote.
  local stem; stem="$(basename "$f" .md)"; stem="${stem%.loop-brief}"
  local journal=".loop/state/journals/$stem.md"
  if [[ -f "$journal" ]]; then
    bad "already run — $journal exists; this brief declares itself plannable and is not"
  fi

  # A worked example is what arbitrates when two implementations disagree, and
  # what becomes the end-to-end acceptance test.
  if grep -qiE '^#+ .*(worked example|acceptance)' <<<"$body"; then ok "has a worked example"
  else bad "no worked example section — nothing arbitrates a disagreement"; fi

  # ...and it has to contain concrete values, not a description of some.
  if grep -qE '^```' <<<"$body"; then ok "has a fenced block with concrete values"
  else bad "no fenced code block — the worked example needs exact values, not prose"; fi

  # The out-of-scope list is how scope creep becomes a finding rather than a
  # matter of taste. The reviewer catches that shape; it needs something to cite.
  if grep -qiE '^#+ .*out of scope' <<<"$body"; then
    local n; n="$(awk '
      /^#+ .*[Oo]ut of scope/{f=1; c=0; next}
      f && /^#+ /{ if (c>m) m=c; f=0 }
      f && /^[-*] /{ c++ }
      END{ if (c>m) m=c; print m+0 }' <<<"$body")"
    if [[ "$n" -ge 2 ]]; then ok "out of scope: $n item(s)"
    else bad "out-of-scope section has $n item(s) — name what you are NOT asking for"; fi
  else bad "no out-of-scope section — scope creep has nothing to be measured against"; fi

  grep -qiE '^#+ .*constraint' <<<"$body" && ok "has constraints" \
    || warn "no constraints section — toolchain and limits are left to guesswork"

  # A task count calibrates decomposition.
  if grep -qiE '[0-9]+ (to|–|-) ?[0-9]* ?tasks|[a-z]+ to [a-z]+ tasks|[0-9]+ tasks' <<<"$body"; then
    ok "states an expected task count"
  else warn "no expected task count — decomposition has nothing to calibrate against"; fi

  # A precise behaviour contract almost always pins codes of some kind.
  if grep -qiE 'exit (code|[0-9])|\b(200|201|302|400|404|409|500)\b|exits? [0-9]' <<<"$body"; then
    ok "pins exit or status codes"
  else warn "no exit/status codes — is the behaviour contract precise enough to gate?"; fi

  # Hard rule 2, and briefs get published.
  if grep -qE '/Users/|/home/[a-z]' <<<"$body"; then
    bad "contains an absolute home path"
  else ok "no absolute paths"; fi

  # Pinning internal structure is the planner's job where the brief is silent;
  # a brief that does it has chosen mechanics, which is worth a second look.
  local pinned; pinned="$(grep -coE '`[a-z_]+\.(py|ts|js)::|`[a-z_]+\.[a-z_]+\(\)' <<<"$body" || true)"
  [[ "${pinned:-0}" -gt 2 ]] && warn "names $pinned internal symbols — pinning mechanics, not decisions?"

  # Every referenced path should resolve, or be an obvious template.
  #
  # Not just .md. A brief derived from an existing design process cites what
  # that process produced -- diagrams, token files, schemas, a handoff
  # directory -- and a reference that silently is not checked is worse than no
  # reference: it reads as authority and points at nothing.
  local dead=0 r
  while read -r r; do
    [[ -n "$r" ]] || continue
    [[ "$r" == ~* || "$r" == http* || "$r" == *NNN* || "$r" == *'<'* ]] && continue
    # repo path, doc-relative, or a bare filename that exists somewhere: a
    # worked example legitimately names fixture files rather than repo paths
    [[ -e "$r" || -e "$(dirname "$f")/$r" ]] && continue
    find . -name "$(basename "${r%/}")" -not -path './.git/*' -print -quit 2>/dev/null | grep -q . && continue
    warn "path does not resolve: $r"; dead=$((dead+1))
  done < <(grep -oE '`[A-Za-z0-9_./-]+(\.(md|puml|json|jsonl|ya?ml|sh|py|ts|tsx|js|sql|toml|csv|svg|png)|/)`' \
             <<<"$body" | tr -d '`' | sort -u)
  [[ "$dead" -eq 0 ]] && ok "referenced docs resolve"

  # References the reader provably cannot follow.
  #
  # Every session downstream of this brief -- planner, work AND review -- reads
  # it as a fresh `claude -p` with --strict-mcp-config, no memory and
  # WebFetch/WebSearch denied. A tracker key or an issue URL names something
  # none of them can open. The author's session usually CAN open it, which is
  # exactly why this goes unnoticed: the brief reads complete to the one person
  # who will never be its reader.
  #
  # Same class as the absolute-path check above -- a reference that does not
  # resolve where it is consumed -- and it is the one rule that is checkable
  # rather than a matter of judgement, so it is checked.
  # A brief numbered by creation time -- BYYYYMMDD-HHMM, the convention a repo
  # adopts once two people drafting on different branches would otherwise
  # collide on the next counter value -- has exactly the shape of a tracker
  # key. So a brief that names itself, in its own frontmatter or its title,
  # warned about itself, on every brief, forever. Match the date-time SHAPE
  # rather than adding a letter to the list above: the prefix is the repo's
  # choice and B is only the one seen first.
  #
  # And the prefix is not one character. This matched the shape with `[A-Z]`
  # and was then wrong for the first id that used two -- `.loop/todo/`'s
  # `TDYYYYMMDD-HHMM`, a convention that landed after this check -- so every
  # brief citing a parked TODO warned about a file sitting in this repo. Same
  # bug as the one the paragraph above describes, one layer in: the SHAPE is a
  # date-time, the prefix in front of it is free.
  local keys
  keys="$(grep -oE '\b[A-Z][A-Z0-9]{1,9}-[0-9]+\b' <<<"$body" \
    | grep -vE '^(UTF|SHA|ISO|RFC|HTTP|MD|AES|RSA|SPDX|ASCII|CVE|ES|EC|PEP|ADR|UC|SPA)-' \
    | grep -vE '^[A-Z]{1,3}[0-9]{8}-[0-9]{4}$' \
    | sort -u | tr '\n' ' ')"
  if [[ -n "${keys// /}" ]]; then
    warn "names issue(s) no session can open: ${keys}-- carry what they say, not the key"
  elif grep -qiE 'https?://[^ )]*(linear\.app|atlassian\.net|/browse/|/issues/)' <<<"$body"; then
    warn "links an issue tracker -- no session downstream has web access; inline what it says"
  else
    ok "no references a memoryless offline session cannot follow"
  fi

  if [[ $problems -gt 0 ]]; then
    printf '  \033[31m%d problem(s)\033[0m, %d warning(s)\n' "$problems" "$warnings"; fails=$((fails+1))
  else
    printf '  \033[32mok\033[0m, %d warning(s)\n' "$warnings"
  fi
}

[[ $# -gt 0 ]] || { echo "usage: .loop/check-brief.sh <brief.md> [...]" >&2; exit 1; }
for f in "$@"; do
  [[ -f "$f" ]] || { echo "no such brief: $f" >&2; fails=$((fails+1)); continue; }
  check_one "$f"
done
printf '\n'
[[ $fails -eq 0 ]] && echo "briefs ok" || echo "$fails brief(s) need work"
exit $(( fails > 0 ? 1 : 0 ))
