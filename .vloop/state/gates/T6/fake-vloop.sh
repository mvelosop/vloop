#!/bin/sh
# Stands for --vloop: the binary a Copilot session's fence hook would run.
echo "fake vloop $*" >> '@ST@/vloop.calls'
exit 0
