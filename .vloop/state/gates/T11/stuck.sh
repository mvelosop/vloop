# T3's work never makes T3.out, so its gate always fails.
if [ "$PHASE" = work ] && [ "$TASK" = T3 ]; then rm -f T3.out; fi
