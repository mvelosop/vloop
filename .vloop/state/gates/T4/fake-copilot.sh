#!/bin/sh
# A fake Copilot CLI. It records how it was started, writes the verdict file in
# its working directory (the workspace) and prints copilot-session.jsonl.
ST='@ST@'
[ "${1-}" = "--version" ] && { echo "1.0.105"; exit 0; }
printf '%s\n' "$@" > "$ST/copilot.argv"
pwd -P > "$ST/copilot.pwd"
p=''; pd=''; prev=''
for a in "$@"; do
  [ "$prev" = "-p" ] && p=$a
  [ "$prev" = "--plugin-dir" ] && pd=$a
  prev=$a
done
( cd "$pd" 2>/dev/null && pwd -P ) > "$ST/copilot.plugin"
if [ -n "${COPILOT_HOME-}" ] && [ -f "$COPILOT_HOME/hooks/vloop-fence.json" ]; then
  cp "$COPILOT_HOME/hooks/vloop-fence.json" "$ST/copilot.hooks"
fi
printf '%s\t%s\n' "$p" "${VLOOP_PLAN_SHA256-unset}" >> "$ST/copilot.log"
mode=$(cat "$ST/mode" 2>/dev/null || true)
case $mode in
  sleep) sleep 20; exit 0 ;;
  silent) echo "no events at all" >&2; exit 1 ;;
esac
verdict=FAIL
[ "$mode" = pass ] && verdict=PASS
mkdir -p .vloop/tmp
printf '{"verdict": "%s"}\n' "$verdict" > .vloop/tmp/verdict.json
cat "$ST/copilot-session.jsonl"
