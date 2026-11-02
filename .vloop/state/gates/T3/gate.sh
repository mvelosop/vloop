# T3 (K3): the CLAUDE.md section vloop writes sends an interactive session to
# /vloop:create-brief and /vloop:review-brief — `vloop init` in a fresh
# repository writes it, and an older section is replaced by it between the
# markers (the gate's oracle, in a copy of the tree).
G=.vloop/state/gates/T3
. "$G/lib.sh"
cd "$top" || fail cd

srccopy
oracle zz_gate_t3_section_test.go internal/install
(cd "$t/src" && GATE_SECTION_WANT="$top/$G/section.want" go test -count=1 -run '^TestGateT3' ./internal/install >"$t/oracle.out" 2>&1) ||
  fail "the section oracle failed:
$(tail -40 "$t/oracle.out")"

newbin
freshrepo
[ -f "$t/repo/CLAUDE.md" ] || fail "vloop init wrote no CLAUDE.md"
n=$(grep -c '/vloop:create-brief' "$t/repo/CLAUDE.md")
[ "$n" = 1 ] || fail "grep -c /vloop:create-brief CLAUDE.md: $n, want 1"
sed 's/^Written by vloop .*\.$/Written by vloop @VERSION@./' "$t/repo/CLAUDE.md" >"$t/claude.got" || fail sed
cmp -s "$t/claude.got" "$top/$G/section.want" || fail "vloop init's CLAUDE.md is not the section with the new interactive-session paragraph:
$(diff "$top/$G/section.want" "$t/claude.got" | head -20)"
echo "T3 gate: ok"
