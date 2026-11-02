---
max_turns: 60
allowed_tools: ["Read", "Glob", "Grep", "Edit", "Write", "TodoWrite", "Bash(git status:*)", "Bash(git diff:*)", "Bash(git log:*)", "Bash(git show:*)", "Bash(git rev-parse:*)", "Bash(ls:*)", "Bash(cat:*)", "Bash(head:*)", "Bash(tail:*)", "Bash(wc:*)", "Bash(find:*)", "Bash(mkdir:*)", "Bash(jq:*)", "Bash(vloop config get:*)", "Bash(vloop status:*)", "Bash(vloop task list:*)", "Bash(vloop task show:*)", "Bash(vloop task validate:*)", "Bash(vloop task gate:*)", "Bash(vloop schema validate:*)", "Bash(echo:*)", "Bash(git ls-files:*)", "Bash(grep:*)", "Bash(sh:*)", "Bash(chmod:*)", "Bash(test:*)", "Bash(vloop task:*)", "Bash(vloop brief:*)"]
---
/vloop:create-brief Write the brief for this request. Every fork is settled; do not ask me questions.

Add a `--shout` option to `bin/greet`. Every value is final:

- `greet --shout <name>` prints `HELLO, <NAME>` on stdout (the name in upper case) and exits 0.
- `greet <name>` is unchanged: it prints `hello, <name>` and exits 0.
- `greet` with no name prints `usage: greet [--shout] <name>` on stderr, nothing on stdout, and exits 2.
- `README.md` documents `--shout`.
- Worked example: `sh bin/greet --shout world` prints `HELLO, WORLD` and exits 0; `sh bin/greet` prints `usage: greet [--shout] <name>` on stderr and exits 2.
- Out of scope: any option other than `--shout`, colour, locales, and a config file.
- Two tasks: first the option and its test, then the README.

Name the brief `greet-shout`. The independent review of the brief is out of this case: do not run /vloop:review-brief or start a review session; stop once the draft is written and checked.
