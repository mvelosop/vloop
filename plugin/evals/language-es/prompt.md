---
max_turns: 60
allowed_tools: Read, Grep, Glob, Write, Edit, Bash(git rev-parse:*), Bash(git status:*), Bash(vloop config get:*), Bash(vloop schema validate:*), Bash(vloop task:*), Bash(sh:*), Bash(chmod:*), Bash(grep:*), Bash(cat:*), Bash(test:*)
setup: scaffold.sh
---
/vloop:plan docs/briefs/greet.loop-brief.md
