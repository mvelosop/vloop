#!/bin/sh
# A fake Claude Code CLI. A session (its argv has --plugin-dir) records how it
# was started, writes the verdict file and prints claude-stream.jsonl; a judge
# call (no --plugin-dir) records its argv, directory and stdin and answers.
ST='@ST@'
[ "${1-}" = "--version" ] && { echo "2.1.0 (Claude Code)"; exit 0; }
session=no
for a in "$@"; do [ "$a" = "--plugin-dir" ] && session=yes; done
if [ $session = yes ]; then
  echo x >> "$ST/claude.calls"
  printf '%s\n' "$@" > "$ST/claude.argv"
  pd=''; fe=''; prev=''
  for a in "$@"; do
    [ "$prev" = "--plugin-dir" ] && pd=$a
    [ "$prev" = "--settings" ] && fe=$a
    prev=$a
  done
  ( cd "$pd" 2>/dev/null && pwd -P ) > "$ST/claude.plugin"
  if [ -f "$fe" ]; then
    ( cd "$(dirname "$fe")" && printf '%s/%s\n' "$(pwd -P)" "$(basename "$fe")" ) > "$ST/claude.fence"
  fi
  mode=$(cat "$ST/mode" 2>/dev/null || true)
  mkdir -p .vloop/tmp
  printf '{"verdict": "FAIL"}\n' > .vloop/tmp/verdict.json
  if [ "$mode" = noresult ]; then
    sed '$d' "$ST/claude-stream.jsonl"
  else
    cat "$ST/claude-stream.jsonl"
  fi
  exit 0
fi
n=$(( $(cat "$ST/judge.n" 2>/dev/null || echo 0) + 1 ))
echo "$n" > "$ST/judge.n"
ls -A | wc -l | tr -d ' ' > "$ST/judge.files.$n"
pwd -P > "$ST/judge.pwd.$n"
printf '%s\n' "$@" > "$ST/judge.argv.$n"
cat > "$ST/judge.stdin.$n"
jm=$(cat "$ST/judge.mode" 2>/dev/null || true)
case $jm in
  fail) echo "judge is down" >&2; exit 1 ;;
  seq) ans=$(sed -n "${n}p" "$ST/judge.seq") ;;
  *) ans=PASS ;;
esac
printf '{"type":"result","is_error":false,"result":"I read the focus.\\n%s","total_cost_usd":0.01}\n' "$ans"
