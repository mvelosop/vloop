#!/bin/sh
# Gate T3 — the eval case plan-judges-through-named-surface and the catalogues
# (G4). The case's files hold what the brief fixes; its scaffold, run with bash
# in an empty temporary directory, builds the brief's baseline (bin/tally
# failing with "tally: not implemented", three tracked files, one commit, a
# clean tree, no .vloop/) and refuses a second run; the two catalogues name the
# case. Text of the brief and the catalogues is joined (newlines to spaces,
# whitespace collapsed) before it is searched, except the README's table row,
# which is one line.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
export LC_ALL=C
top=$(pwd -P)
C=plugin/evals/plan-judges-through-named-surface
M=plugin/evals/plan-checks-its-gates
[ -d "$C" ] || fail "$C/ is missing"
for f in prompt.md case.yaml scaffold.sh brief.md graders/state-validates.md graders/judges-through-the-command.md; do
  [ -f "$C/$f" ] || fail "$C/$f is missing"
done
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") || fail mktemp
trap 'rm -rf "$t"' EXIT
joined() { tr -s ' \t\n' '   ' <"$1" | sed -e 's/^ //' -e 's/ $//'; }
# front <file>: the frontmatter, both --- lines included; body <file>: the rest.
front() { awk 'NR == 1 && $0 != "---" { exit } { print } NR > 1 && $0 == "---" { exit }' "$1"; }
body() { awk 'n >= 2 { print; next } $0 == "---" { n++ }' "$1"; }

# prompt.md
[ "$(front "$C/prompt.md")" = "$(front "$M/prompt.md")" ] || fail "$C/prompt.md's frontmatter is not $M/prompt.md's"
[ "$(body "$C/prompt.md" | grep -v '^[[:space:]]*$')" = "/vloop:plan docs/briefs/tally.loop-brief.md" ] ||
  fail "$C/prompt.md's body is not: /vloop:plan docs/briefs/tally.loop-brief.md"

# case.yaml
grep -qx 'schema_version: "1.1"' "$C/case.yaml" || fail "$C/case.yaml: no schema_version: \"1.1\""
grep -qx 'name: plan-judges-through-named-surface' "$C/case.yaml" || fail "$C/case.yaml: no name: plan-judges-through-named-surface"
awk '/^context:[[:space:]]*$/ { f = 1; next } /^[^[:space:]]/ { f = 0 } f && /^[[:space:]]+scaffold_script:[[:space:]]*scaffold\.sh[[:space:]]*$/ { ok = 1 } END { exit !ok }' "$C/case.yaml" ||
  fail "$C/case.yaml: no context.scaffold_script: scaffold.sh"

# graders
cmp -s "$C/graders/state-validates.md" "$M/graders/state-validates.md" || fail "$C/graders/state-validates.md is not a copy of $M's"
J=$C/graders/judges-through-the-command.md
front "$J" >"$t/jf"
grep -qx 'type: llm' "$t/jf" || fail "$J: frontmatter has no type: llm"
grep -qx 'weight: 3' "$t/jf" || fail "$J: frontmatter has no weight: 3"
grep -qE '^[[:space:]]*path:[[:space:]]*\.vloop/state/state\.json[[:space:]]*$' "$t/jf" || fail "$J: its focus is not .vloop/state/state.json"
grep -qE '^[[:space:]]*source:[[:space:]]*file[[:space:]]*$' "$t/jf" || fail "$J: its focus is not a file"

# brief.md
Bf=$C/brief.md
front "$Bf" >"$t/bf"
for l in 'name: tally' 'kind: brief' 'status: ready'; do grep -qx "$l" "$t/bf" || fail "$Bf: frontmatter has no $l"; done
grep -qE '^description: *[^ ]' "$t/bf" || fail "$Bf: frontmatter has no description"
joined "$Bf" >"$t/brief"
for s in 'bin/tally add x' 'tally: not a number: x' "are the implementer's"; do
  grep -qF -- "$s" "$t/brief" || fail "$Bf does not say: $s"
done
grep -qF 'lib/' "$t/brief" && fail "$Bf names a lib/ path: the library's file names are the implementer's"
grep -rqi 'shortener' "$C" && fail "$C/ mentions the URL shortener"

# scaffold
d="$t/sandbox"; mkdir "$d" || fail mkdir
(cd "$d" && bash "$top/$C/scaffold.sh") >"$t/s.out" 2>&1 || fail "the scaffold exits non-zero in an empty directory:
$(cat "$t/s.out")"
[ -f "$d/bin/tally" ] || fail "the scaffold wrote no bin/tally"
[ "$(ls -l "$d/bin/tally" | cut -c1-10)" = "-rwxr-xr-x" ] || fail "bin/tally's mode is not 755: $(ls -l "$d/bin/tally" | cut -c1-10)"
[ "$(head -n 1 "$d/bin/tally")" = "#!/bin/sh" ] || fail "bin/tally does not start #!/bin/sh"
(cd "$d" && ./bin/tally total) >"$t/o" 2>"$t/e"; rc=$?
[ "$rc" = 1 ] || fail "bin/tally total exits $rc, want 1"
[ "$(cat "$t/e")" = "tally: not implemented" ] || fail "bin/tally total's stderr is not 'tally: not implemented': $(cat "$t/e")"
[ -s "$t/o" ] && fail "bin/tally total prints on stdout: $(cat "$t/o")"
[ -z "$(git -C "$d" status --porcelain)" ] || fail "the scaffold leaves the tree dirty: $(git -C "$d" status --porcelain)"
[ "$(git -C "$d" ls-files | sort | tr '\n' ' ')" = "CLAUDE.md bin/tally docs/briefs/tally.loop-brief.md " ] ||
  fail "tracked files are not CLAUDE.md, bin/tally, docs/briefs/tally.loop-brief.md: $(git -C "$d" ls-files | tr '\n' ' ')"
[ "$(git -C "$d" log --format=%s)" = "baseline" ] || fail "the scaffold's history is not one commit \"baseline\""
[ "$(git -C "$d" ls-files -s bin/tally | cut -c1-6)" = "100755" ] || fail "bin/tally is not committed executable"
[ "$(wc -l <"$d/CLAUDE.md" | tr -d ' ')" = 1 ] || fail "CLAUDE.md is not one line"
cmp -s "$d/docs/briefs/tally.loop-brief.md" "$Bf" || fail "docs/briefs/tally.loop-brief.md is not a copy of $Bf"
[ -e "$d/.vloop" ] && fail "the scaffold writes .vloop/"
(cd "$d" && bash "$top/$C/scaffold.sh") >/dev/null 2>&1 && fail "a second scaffold run in the same directory exits 0"

# catalogues
R=plugin/evals/README.md
row='| `plan-judges-through-named-surface` | `/vloop:plan` | a brief that names a command and leaves its library'"'"'s file and function names open: the plan validates, and every gate judges through the command, naming no library file or function |'
awk '/^## The other cases/ { f = 1; next } /^## / { f = 0 } f' "$R" | grep -qxF -- "$row" || fail "$R's \"The other cases\" table lacks the row: $row"
E=docs/guide/evals.md
r=$(grep -E '^\| `/vloop:plan` \|' "$E")
[ "$(printf '%s\n' "$r" | grep -c .)" = 1 ] || fail "$E: want one /vloop:plan row"
printf '%s\n' "$r" | awk -F'|' '{ exit !(index($3, "`plan-judges-through-named-surface`") && index($3, "`plan-valid-state`") && index($3, "`plan-checks-its-gates`") && index($3, "`language-es`")) }' ||
  fail "$E: the /vloop:plan row's cases cell does not add plan-judges-through-named-surface to its three cases"
printf '%s\n' "$r" | awk -F'|' '{ s = $4; sub(/[ ]+$/, "", s); w = "prose in the configured language, gates through the surface the brief names"; exit !(substr(s, length(s) - length(w) + 1) == w) }' ||
  fail "$E: the /vloop:plan row's last cell does not end ', gates through the surface the brief names'"
joined "$E" >"$t/guide"
grep -qF 'is 23 cases × 3 runs × 2 arms' "$t/guide" || fail "$E does not say: is 23 cases × 3 runs × 2 arms"
grep -qF '21 cases ×' "$t/guide" && fail "$E still says 21 cases"
echo "gate T3: ok"
