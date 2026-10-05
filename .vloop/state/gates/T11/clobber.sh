# A later task breaks an earlier one once: T2's first work session deletes T1.out.
if [ "$PHASE" = work ] && [ "$TASK" = T2 ] && [ ! -f "$dir/broke" ]; then : > "$dir/broke"; rm -f T1.out; fi
