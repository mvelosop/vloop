---
name: B20261103-0315-b35-i83-fence-compound-commands.loop-brief
description: Let the fence judge read if, while, until, for, !, braces and subshells by judging every command inside them, match a word holding $ as any text, and treat command -v as a lookup
kind: brief
status: consumed
created: 2026-11-03
issue: 83
seeds: Issue 83 under 69; the hook denials counted in docs/design-notes/copilot-cli-experiments.md (B22.0, B24, B26, B27, the gpt-6.1-terra planning test) and its count of the kept sessions' shell calls; B32's kept Copilot sessions; decided by the assistant under the operator's delegation of 2026-11-02 for the #81–#88 chain
depends-on: [B20261103-0058-b34-i82-fence-backstop.loop-brief]
---
# Brief — vloop B35: the fence judges the compound commands it can read

- **Issue:** #83 (under #69). The PR closes it.
- **Starting point:** extends `main` at `ee4adf63`, plus this brief's ready
  merge. The planner pins that SHA as the **base** every gate compares
  against — never `HEAD~n`.
- **Produced by:** the operator's delegation of 2026-11-02 for a chain of
  briefs on #84, #81, #82, #83, #96–#97 and #88. Written with
  `plugin/skills/create-brief/`; every decision below was taken by the
  assistant under that delegation.
- **Review:** `vloop brief review` (Claude Opus), 13 rounds. Round 1: ASK,
  8 findings; the one `ASK` (braces and globs bypass the fence as `$`
  does) decided by widening C2 to every word the shell expands; the rest
  fixed (the `command -v` option pinned literally, the argument template,
  `for ((`, `env !`, the counts re-attributed and the kept sessions'
  figures recorded in the design note, the message's words, C4's
  sample). Rounds 2–13: FIX, 1–4 findings each, all fixed: a wrapper's
  expanded option before the name, a `$` inside a command name, which `}`
  closes a brace, `((…))` refused as arithmetic, every doc message as the
  judge gives it, the seeds, S-4, a glob in a redirection target,
  `xargs` (its input, its replace string, BSD's `-J`, `-R`, `-S`, no
  command), a shell's expanded `-c` string, and C2's governing rule.
  Round 7's `ASK` (`let`, `${a[x]}`, `declare -i`) decided out of scope.
  Round 13's one finding (`xargs` with no command) fixed and not
  re-reviewed: it pins one case and adds no reach. No `ASK` left.

## Executive summary

`vloop fence hook` judges every shell call of a Copilot session. It denies
whatever shell syntax it cannot read, and today that includes every
compound command: `if`, `for`, `while`, `until`, `{ … }` and a subshell
answer `shell syntax … cannot be judged` (a subshell: `subshells and
process substitution cannot be judged`). Each is a harmless form a session
then retries one command at a time. The Copilot acceptances recorded in
`docs/design-notes/copilot-cli-experiments.md` count 9 `if`, 5 `for` and 2
subshells denied this way: B22.0 1 `if` and 1 subshell, B24 5 `if` and 1
subshell, B26 4 `for`, B27 1 `if` and 1 `for`, the `gpt-6.1-terra`
planning test 2 `if`. B32's brief review session was also denied `command -v
curl`, which runs nothing, because the judge looks through `command` to
`curl` (the note's "shell calls of the kept sessions").

Reading these forms means each command inside them is judged on its own,
and that exposes a hole the judge has today: a word the shell expands is
matched as written. `X=push; git $X origin` is allowed, and so are `git
pu{sh,} origin` (brace expansion) and `git pu?h origin` (a glob, with a
file named `push` in the directory); a `for` loop would make the first
easy (`for x in push; do git "$x" origin; done`).

**The goal:** the judge reads `if`, `while`, `until`, `for … in`, `!`,
`{ … }` and subshells, judging every command inside them as it judges any
command; a word the shell expands (a `$`, a glob, a brace expansion)
counts as any text when the deny rules are matched; and `command -v` is a
lookup, not a run.

- **C1 — compound commands are judged through.** A leading `if`, `then`,
  `elif`, `else`, `while`, `until`, `do`, `!` or `{` is dropped and the
  rest of the command judged; `fi`, `done` and `}` alone are judged only
  for their redirections; `for NAME` and `for NAME in WORDS` run nothing;
  a subshell `( … )` is split like any other list.
- **C2 — a word the shell expands stands for any text.** A command no
  deny rule matches as written is denied when a rule matches it with some
  text in place of each word that holds an unquoted or double-quoted `$`,
  or an unquoted glob (`*`, `?`, `[…]`) or brace expansion (`{a,b}`,
  `{1..3}`). A command name that expands anywhere in it is refused; today
  only one that starts with `$` is.
- **C3 — `command -v` and `command -V` run nothing.** Allowed whatever the
  names after them.
- **C4 — the docs:** the configuration guide's fence section, the fence
  audit's hook description and its work-session paragraph, and `vloop fence
  check`'s help.
- Unchanged: the shell syntax still refused (`case`, `select`, `function`,
  `coproc`, `[[`, a function definition, a backtick, `$(`, `<(`, `>(`,
  `eval`, `source`, `.`, a command word starting with `$`, a redirection
  target holding `$`, a `-c` string holding `$`); every fence's rules; the
  driver's backstop (B34); Claude sessions, whose fence is Claude Code's.

## What it is

A wider reading in the judge (`internal/fence`), which both the hook and
B34's re-judge use. It allows forms that today cost a retry, and denies
one form that today gets through.

## Why this shape, and what was rejected

Decided by the assistant on 2026-11-03 under the operator's delegation.
**Do not re-litigate these** — the planner inherits them.

- **Drop the keyword, judge the rest; do not parse the grammar.** The
  splitter already cuts at `;`, `&&`, `||`, `|`, `&` and newlines, so
  `if a; then b; else c; fi` reaches the judge as `if a`, `then b`, `else
  c`, `fi`. Dropping the leading keyword judges `a`, `b` and `c`, which are
  every command the compound can run. Nothing is judged less than if the
  session had sent the commands one per line. A real shell grammar
  (`mvdan.cc/sh`) was rejected: a new dependency, and its tree would be a
  second reading of the command next to the splitter's.
- **`case`, `select`, `function`, `coproc` and `[[` stay refused.** `case`
  patterns end in `)`, which the splitter cannot tell from a subshell;
  `select` and `coproc` are rare; a function body runs later, under a name;
  `[[` holds `&&`, `||`, `<` and `>` that are not operators, which the
  splitter would cut or read as redirections. `[ … ]` and `test` are
  allowed today and cover what `[[` is used for. Not one of them appears
  among the denials counted above.
- **An expanded word is any text, rather than resolving it.** The judge
  never runs the command and cannot know a variable's value; an assignment
  earlier in the string, a `for` list, the environment or `read` can each
  set it. A glob's words depend on the files present when it runs. Brace
  expansion could be computed, but it can also hold a `$` or a glob, and
  one rule for all three is simpler to judge and to check. Reading the
  word as any text is the only answer that holds for all of them. It can deny a harmless command (`git "$sub" --help`, since `$sub`
  may be `push`), which costs a retry, as the refusals did, and so can a
  rule with an inner `*`: `git diff "$BASE"
  --stat` is denied, since `$BASE` may be `--output=…` (`Bash(git
  diff*--output*)`), and so is `find "$d" -name x` (`Bash(find*-delete*)`).
  In a work session, `rm "$f"` and `rm *.tmp` are denied (`$f` may be
  `-rf x`, `Bash(rm -rf:*)`), though `rm` of a file is allowed there on
  purpose, and so are `git show "$SHA"` and `git log -- *.go` (`Bash(git
  show*--output*)`, `Bash(git log*--output*)`). A session writes the
  path or the SHA instead; `git diff -- src/*.go` is denied too
  (a file can be named `--output=x`), and `git diff -- 'src/*.go'` is not.
  Of the 52 shell calls in the 7 kept Copilot sessions of B32 and B33 (the
  design note's "shell calls of the kept sessions"), 5 hold a `$`; none
  names a command a fence denies.
- **Quoted is literal.** `'$x'`, `'*.go'`, `"*.go"` and `\*` are never
  expanded, so such a word is matched as written, as today; `grep '$(' x`
  stays allowed. A `{` with no `,` or `..` before its `}` (`@{u}`, `{}`)
  is not a brace expansion, and a `[` with no `]` after it (`[ -f x ]`) is
  not a glob.
- **`command -v` and `-V` only.** They print how a name would resolve and
  run nothing (POSIX `command`). `type`, `which` and `hash` are not
  looked through today and need nothing.
- **Subshells, not only braces.** The issue names them, and B22.0's and
  B24's runs counted one each. `$(`, `<(` and `>(` stay refused: they splice output into
  another command.

## Binding references

- `internal/fence/split.go` — `split` and the lexer: where `(` and `)`
  are refused today and where words are built (the quoting each `$`, glob
  and brace comes in).
- `internal/fence/judge.go` — `keywords`, `words` (where `shell syntax …
  cannot be judged` comes from and where `command` is looked through),
  `match` and `patternMatch` (the fence's matching and its two messages).
- `internal/fence/judge_test.go` — `TestSplitRefuses`, `shellCases` and
  the work fence's table: the cases that pin today's refusals.
- `docs/domain/domain-model.md` — the glossary's **hook** row and S-2:
  the hook is a line before the driver, which stays the boundary; S-4
  (where a session writes): what the redirections of `fi`, `done`, `}` and
  `)` are judged against.
- `docs/guide/configuration.md` — the fence section's bullet on the work
  fence, whose sentence lists the syntax denied.
- `docs/design-notes/fence-audit.md` — "The three enforcers" (the hook's
  description) and the work-session paragraph after the work fence's
  list.
- `internal/cli/fence.go` — `vloop fence check`'s help text.
- `docs/design-notes/copilot-cli-experiments.md` — the denial counts cited
  above: the hook-denial rows of B22.0, B24, B26, B27 and the
  `gpt-6.1-terra` planning test, and "The shell calls of the kept
  sessions" (added with this brief's ready merge, so on the base).

## Behaviour contract

### C1 — compound commands are judged through

The splitter cuts the string into simple commands as today. Then, for each
simple command, before assignments are skipped:

- **Opening keywords.** While the command's first word is `if`, `then`,
  `elif`, `else`, `while`, `until`, `do`, `!` or `{`, that word is dropped.
  What remains is judged as any command (assignments, wrappers, the deny
  rules); nothing remaining is allowed. A keyword counts only as the first
  word: `X=1 if true` stays denied (`shell syntax if cannot be judged`),
  and `echo if` is an `echo`.
- **Closing keywords.** A command whose words are exactly `fi`, `done` or
  `}` is allowed once its redirections are (they are judged as any
  redirection: `done > .vloop/state/x` is denied in a review session). `fi`,
  `done` or `}` followed by another word is denied: `shell syntax fi cannot
  be judged` (the word at hand).
- **`for`.** A command `for NAME` or `for NAME in WORDS…`, `NAME` matching
  `[A-Za-z_][A-Za-z0-9_]*`, runs nothing and is allowed, its redirections
  judged as above. Any other `for` (a missing or bad name, a second word
  not `in`) is denied: `shell syntax for cannot be judged`. `for ((…))`
  never reaches this rule: the splitter refuses its `(` first, `subshells
  and process substitution cannot be judged`.
- **Still refused**, with today's message `shell syntax <word> cannot be
  judged`: `case`, `esac`, `select`, `function`, `coproc`, `[[`, as the
  command's first word or after an opening keyword.
- **Subshells.** A `(` opens a subshell when no word is being built and the
  current command holds no word but opening keywords and no redirection,
  complete or still waiting for its target (so `<(ls) cat` stays refused):
  the current command ends there (judged as above, so `if` alone is
  allowed), and the commands inside are split and judged as any other. A
  `)` while a subshell is open ends the current command and closes it; a
  redirection may follow (`(cd src && ls) > .vloop/tmp/x` judges the write
  to `.vloop/tmp/x`), a word may not (`subshells and process substitution
  cannot be judged`). Subshells nest, written apart (`( (ls) )`). A `(` written directly
  after another `(`, `((`, is arithmetic, not two subshells: bash evaluates
  a variable's value there as an expression, and a subscript in it can run
  a command substitution (`x='a[$(git push)]'; ((x))` runs `git push`).
  It is refused: `subshells and process substitution cannot be judged`. A `)` with none open, a `(` anywhere
  else (`a=(1 2)`, `f() { …; }`, `<(…)`, `>(…)`), and a string that ends
  with a subshell still open, are refused with that same message. `$(`
  stays `command substitution cannot be judged`.

### C2 — a word the shell expands stands for any text

- A word is **expanded** when it holds:
  - a `$` that is unquoted or inside double quotes;
  - an unquoted `*` or `?`, or an unquoted `[` followed later in the word
    by an unquoted `]` (a glob);
  - a brace expansion: an unquoted `{` and the unquoted `}` that closes
    it (the first unquoted `}` after it with no unquoted `{` between), with
    an unquoted `,` or `..` between the two. `HEAD@{1}..HEAD@{0}` holds two
    such pairs, `{1}` and `{0}`, neither with a `,` or `..` inside: not a
    brace expansion.

  Single quotes, double quotes (for the glob and the brace) and a
  backslash make the character literal: `'$x'`, `"*.go"`, `\*`, `'{a,b}'`
  are not expanded.
- **The governing rule.** An expanded word is matched as any text only
  where it is an argument the deny rules look at. Wherever the judge reads
  a word as structure (a command name, a wrapper's option or its value, a
  shell's option or its `-c` string, a redirection target, xargs' input
  standing as a command name), an expanded word there is refused. The
  bullets below are this rule's cases; a form they miss follows the rule.
- `match` first matches the command as written, exactly as today, with
  today's messages. When no rule matches, it matches again with each
  expanded word standing for any text (none, one word, several words): a
  command is denied when some text in place of those words makes a deny
  rule, of the phase's fence or of the Copilot adapter's deny list, match.
  The rules are tried in today's order (the fence's, then the deny list's),
  and the first that can match is named. Both the rule and the command are
  then patterns of literal text and `*`; how to decide that two such
  patterns share a string is the plan's.
- The message is the rule's, with the command's words joined by spaces as
  today (quotes removed), and this suffix: `git $X origin is denied by the
  review fence (Bash(git push:*)): a word the shell expands may be any
  text`; for the deny list, `… is denied by the Copilot deny list
  (<stem>): a word the shell expands may be any text`.
- The command's name (after assignments and wrappers) that is expanded,
  by a `$` anywhere in it, a glob or a brace expansion, is refused:
  `command word <word> expands and cannot be judged`. A name starting with
  `$` keeps today's message (`command word $CMD starts with $ and cannot be
  judged`).
- An expanded word can become several words, so it must not stand before
  the command's name: a word a wrapper skips that is expanded (an option or
  its value, `timeout`'s duration, `exec -a`'s name, `env`'s `NAME=value`,
  `nice -n`'s number, `xargs`' and `stdbuf`'s options) refuses the
  command:
  `<word> expands before the command name and cannot be judged`. With
  `N='1 git'`, `nice -n $N push origin` runs `git push origin`. Assignments
  before the first word (`X=$Y git status`) are not split by the shell and
  are not affected.
- The same holds for `sh`, `bash`, `zsh`, `dash` and `ksh`: an option word
  before the `-c` string (or `-o`'s value) that is expanded is refused,
  with the shell's prefix as today: `bash: <word> expands before the
  command string and cannot be judged`. A `-c` string that is itself
  expanded by a glob or a brace expansion (`sh -c git\ \'pu{sh,}\'`) is
  refused: `sh: the -c string expands and cannot be judged`; one holding
  `$` keeps today's message.
- `xargs` appends the words it reads from its input to the command it
  runs, so that command is judged as if it ended with one more expanded
  word, added before the judge looks through the command's wrappers: a
  wrapper with nothing after it (`xargs env`, `xargs nohup`, `xargs
  timeout 5`, `xargs xargs`) has that word as its command name and is
  refused, `command word <xargs input> expands and cannot be judged`.
  `xargs` with no command runs `echo`, and is judged as `echo` with the
  word appended, so `ls | xargs` and `ls | xargs -n 1` stay allowed; `echo push origin | xargs git` is denied (`git is denied by the
  review fence (Bash(git push:*)): a word the shell expands may be any
  text`); `xargs wc -l` and `xargs grep -n x` stay allowed. With a
  replace option (`-I <s>`, `-i[<s>]`, `--replace[=<s>]`, BSD `-J <s>`;
  `-i` and `--replace` default to `{}`), the input goes where the replace
  string is instead: every word of the command holding it counts as
  expanded, nothing is appended, and a command name holding it is refused
  (`command word <word> expands and cannot be judged`). A shell's `-c`
  string holding the replace string is refused too, since the judge
  splits that string again and would read the replace string as literal:
  `sh: the -c string holds xargs' replace string % and cannot be judged`
  (the shell and the string at hand).
- `xargs` options that take an argument include BSD's (macOS's) `-J`, `-R`
  and `-S`, besides today's `-a`, `-d`, `-E`, `-I`, `-L`, `-n`, `-P` and
  `-s`: today the judge reads `-J`'s, `-R`'s or `-S`'s argument as the
  command name, so `xargs -J % git push` is allowed.
- Unchanged: a redirection target holding `$` and a `-c` string holding
  `$` stay refused with today's messages. Writes and reads judged by path
  are not affected by the "any text" reading. A redirection target that
  holds a glob or a brace expansion is refused, as one holding `$` is
  today: bash writes to the file a glob matches (`echo x > .vloop/*`), so
  the target cannot be judged as written. The message is `a redirection
  target that expands cannot be resolved: <word>`; one holding `$` keeps
  today's.

### C3 — `command -v` and `command -V` run nothing

When the judge looks through `command` and an option before the name is a
`-` followed only by the letters `p`, `v` and `V`, at least one of them `v`
or `V` (`-v`, `-V`, `-pv`, `-vp`), written literally (no `$`, no quote),
the command runs nothing and is allowed, whatever follows (`command -v
curl`, `command -v "$tool"`). Only `command`: `time`, which the judge
looks through alongside it, runs its command whatever its options (`time -v
git push` stays denied). Without such an option `command` is looked
through as today (`command gh pr list` stays denied); an option holding
`$` is a word `command` skips that expands, so `command -$v git push` is
refused by C2's rule on words before the name.

### C4 — the docs

- `docs/guide/configuration.md`, the work-fence bullet: the sentence
  "Shell syntax the judge cannot read is denied, not guessed at: `if`,
  `for`, `while`, `until`, `case`, `function`, `{`, `[[`, `$(…)`, a
  subshell, answering `shell syntax if cannot be judged` (or the keyword at
  hand). Write the commands one per line instead." becomes: "The judge
  reads `if`, `while`, `until`, `for NAME in …`, `!`, `{ … }` and
  subshells, and judges every command inside them. Shell syntax it cannot
  read is denied, not guessed at: `select`, `function`, `coproc` and `[[`
  answer `shell syntax [[ cannot be judged` (the word at hand); `$(…)` and
  a backtick, `command substitution cannot be judged`; `case`, `<(…)`,
  `>(…)` and `((…))`, `subshells and process substitution cannot be
  judged`; a command name that expands, an expanded word before the
  command name (a wrapper's option, or a shell's before its `-c` string)
  and a redirection target holding a glob or a brace expansion, `… expands
  … and cannot be judged` or `… cannot be resolved`. A word the shell expands (a `$`, a glob, a brace
  expansion) may be any text: `git $X origin` and `git pu{sh,} origin` are
  denied wherever `git push` is."
- `docs/design-notes/fence-audit.md`, "The three enforcers", the **hook**
  bullet: after "a command word starting with `$`." add "It reads `if`,
  `while`, `until`, `for NAME in …`, `!`, `{ … }` and subshells by judging
  every command inside them, and matches a word the shell expands (a `$`,
  a glob, a brace expansion) as any text; it refuses a command name that
  expands, an expanded word before the command name (a wrapper's option,
  or a shell's before its `-c` string), and a redirection target holding a
  glob or a brace expansion (B35)."
- `docs/design-notes/fence-audit.md`, the work-session paragraph that
  begins "Shell syntax the judge cannot read is denied in a work session as
  in a review one": its first sentence becomes "Shell syntax the
  judge cannot read is denied in a work session as in a review one: a
  backtick or `$(` answers `command substitution cannot be judged`; `<(`,
  `>(`, `((` and `case`, `subshells and process substitution cannot be
  judged`; `eval`, `source` and `.`, `… is not allowed`; `select`,
  `function`, `coproc` and `[[`, `shell syntax … cannot be judged`. `if`,
  `while`, `until`, `for`, `!`, `{ … }` and subshells are judged through
  (B35)." The rest of the paragraph (from "The hook fails open") stays.
- `internal/cli/fence.go`, `vloop fence check`'s long help: after "looked
  through env, command, exec, nohup, time, nice, timeout, stdbuf, xargs and
  sh -c," add "read through if, while, until, for, !, { } and subshells,"
  and, after the sentence on the fence's rules, "A word the shell expands
  ($, a glob, a brace expansion) matches as any text."

### Decided here, because the cited documents leave it underdetermined

- The order inside a command: opening keywords dropped, then (C1's other
  rules) closing keywords, `for` and the still-refused words, then
  assignments and wrappers as today.
- A keyword after a wrapper is not dropped: `env ! x` stays denied with
  today's answer, `shell syntax ! cannot be judged`.
- C2 applies in every phase and to the words after every wrapper, the
  command's own name excluded (a name that expands is refused, above).

### Violations the review must rule on

- Any form under "still refused" allowed.
- A command inside a compound or a subshell judged less than it would be
  alone: `if true; then git push; fi`, `(git push)` and `while true; do
  git push; done` must each be denied exactly as `git push` is.
- A quoted or escaped `$`, glob or brace read as any text.
- A dependency added to `go.mod`.

## Worked example

Every line is `vloop fence check --root <tmp> --phase <phase> --tool bash
<command>`, built from the branch, in a temporary directory, with the
command passed as one argument exactly as shown: a Go string in the
judge's tests, where each line is a case; at the close, single-quoted, each
`'` in it written `'\''`. `allow` exits 0; `deny: …` exits 1.

```
# C1, the real denials (plan session)
plan    if test -f .vloop/tmp/plan-problems.md; then cat .vloop/tmp/plan-problems.md; else printf 'No revision feedback\n'; fi   → allow
plan    for tool in node pnpm npm curl vloop; do command -v "$tool" || true; done   → allow
review  while read -r f; do wc -l "$f"; done < .vloop/tmp/list   → allow
review  { git status; git diff; } > .vloop/tmp/out   → allow
review  (cd src && ls) > .vloop/tmp/x   → allow
review  if ! git diff --quiet; then echo changed; fi   → allow
# C1, what is inside is judged
review  if true; then git push; fi   → deny: git push is denied by the review fence (Bash(git push:*))
review  (git push)   → deny: git push is denied by the review fence (Bash(git push:*))
review  while true; do git push; done   → deny: git push is denied by the review fence (Bash(git push:*))
review  { git status; } > .vloop/state/x   → deny: write to .vloop/state/x is not allowed in a review session: only .vloop/tmp/ may be written
review  fi now   → deny: shell syntax fi cannot be judged
review  for ((i=0; i<3; i++)); do ls; done   → deny: subshells and process substitution cannot be judged
review  case x in a) ls;; esac   → deny: subshells and process substitution cannot be judged
review  [[ -f x ]]   → deny: shell syntax [[ cannot be judged
review  env ! git status   → deny: shell syntax ! cannot be judged
review  X=1 if true; then ls; fi   → deny: shell syntax if cannot be judged
review  ((x))   → deny: subshells and process substitution cannot be judged
review  x='a[$(git push)]'; ((x))   → deny: subshells and process substitution cannot be judged
review  ( (ls) )   → allow
review  a=(1 2)   → deny: subshells and process substitution cannot be judged
review  (ls   → deny: subshells and process substitution cannot be judged
review  cat <(ls)   → deny: subshells and process substitution cannot be judged
# C2
review  X=push; git $X origin   → deny: git $X origin is denied by the review fence (Bash(git push:*)): a word the shell expands may be any text
work    for x in push; do git "$x" origin; done   → deny: git $x origin is denied by the work fence (Bash(git push:*)): a word the shell expands may be any text
review  git pu{sh,} origin   → deny: git pu{sh,} origin is denied by the review fence (Bash(git push:*)): a word the shell expands may be any text
review  git pu?h origin   → deny: git pu?h origin is denied by the review fence (Bash(git push:*)): a word the shell expands may be any text
review  git 'pu{sh,}' origin   → allow
review  git log @{u}..HEAD   → allow
review  [ -f x ] && echo yes   → allow
review  g?t push   → deny: command word g?t expands and cannot be judged
review  g$X push origin   → deny: command word g$X expands and cannot be judged
review  /usr/bin/$X push origin   → deny: command word /usr/bin/$X expands and cannot be judged
review  git log HEAD@{1}..HEAD@{0}   → allow
review  bash -$O -c 'git status'   → deny: bash: -$O expands before the command string and cannot be judged
review  echo push | xargs -I{} git {} origin   → deny: git {} origin is denied by the review fence (Bash(git push:*)): a word the shell expands may be any text
review  xargs -J % git push   → deny: git push is denied by the review fence (Bash(git push:*))
review  xargs -I {} -R 1 git push   → deny: git push is denied by the review fence (Bash(git push:*))
review  xargs -S 255 -I {} git push   → deny: git push is denied by the review fence (Bash(git push:*))
review  ls | xargs -n 1   → allow
review  echo git push | xargs env   → deny: command word <xargs input> expands and cannot be judged
review  sh -c git\ \'pu{sh,}\'   → deny: sh: the -c string expands and cannot be judged
review  echo push | xargs -I % sh -c 'git %'   → deny: sh: the -c string holds xargs' replace string % and cannot be judged
review  echo git | xargs -I{} {} push   → deny: command word {} expands and cannot be judged
review  echo push origin | xargs git   → deny: git is denied by the review fence (Bash(git push:*)): a word the shell expands may be any text
review  find src -name '*.go' | xargs wc -l   → allow
review  nice -n $N push origin   → deny: $N expands before the command name and cannot be judged
review  X=$Y git status   → allow
review  git '$X' origin   → allow
work    echo x > .vloop/*   → deny: a redirection target that expands cannot be resolved: .vloop/*
work    echo x > 'src/*.txt'   → allow
review  git diff "$BASE" --stat   → deny: git diff $BASE --stat is denied by the review fence (Bash(git diff*--output*)): a word the shell expands may be any text
review  git diff -- 'src/*.go'   → allow
review  printf 'Gate exit: %s\n' "$gate_status"   → allow
# C3
review  command -v curl   → allow
review  command -V git   → allow
review  time -v git push   → deny: git push is denied by the review fence (Bash(git push:*))
review  command gh pr list   → deny: gh pr list is denied by the Copilot deny list (gh:*)
review  command -$v git push   → deny: -$v expands before the command name and cannot be judged
```

The failures it must catch: a keyword dropped from the middle of a command
(`echo if` is an `echo`, not an `if`); the inside of a compound or a
subshell judged less than alone; a `$`, glob or brace word matched as
written; a quoted one read as any text; `command` without a literal `-v`
let through.

## Out of scope

- `case`, `select`, `function`, `coproc`, `[[` and function definitions:
  refused as today.
- Resolving variables, globs or brace expansions, or judging `$(…)` by its
  inside.
- The other arithmetic contexts that evaluate a variable's value (`let`,
  `${a[x]}`, `declare -i`): allowed today and unchanged here; a later
  brief's. B35 refuses only `((…))`, the one it would otherwise open.
- Relative paths after a `cd` (judged against the root, as today).
- Claude sessions' fence; the driver's backstop (B34).
- #96 and #97 (the plugin declared per repository, B36) and #88
  (`vloop metrics` per session, B37).

## Constraints

- Gates run on macOS's BSD tools: no empty alternatives in `grep -E`, no
  `sed -i` without a suffix, no `\+` or `\|` in basic regexes.
- Sessions never move refs.
- Tests build their fixtures in temporary directories and address them
  with `git -C`; no test runs a real `copilot`.
- The existing tests pass unchanged, except the cases that pin a form C1
  now reads, each of which changes to the C1 answer: in `TestSplitRefuses`,
  `(gh x)` (it now splits; `(gh x` and `a=(1 2)` take its place), and in
  the work fence's table `for f in a b; do echo $f; done` (now allowed).
  `if true; then gh x; fi` stays denied (by the deny list now, not the
  keyword); its case keeps `*`.
- No new module dependency.
- A docs gate joins lines and collapses whitespace before it greps, so an
  edit wrapped in the file's style passes.
- Repo-relative paths everywhere.

## Shape

3–4 tasks; every configured check green after each.

1. **C1 and C3** in the splitter and the judge, gated on the worked
   example's C1 and C3 lines as named tests, but `command -$v git push`,
   which is C2's.
2. **C2**, the `$`, the glob and the brace, gated on its lines.
3. **C4.**
4. **Close:** every gate at once, and the worked example line for line
   through the built binary.
<!-- vloop:run-record:begin -->
## Run record

Generated by vloop brief close on 2026-11-03. Recompute with: vloop metrics B20261103-0315-b35-i83-fence-compound-commands.loop-brief

```
B20261103-0315-b35-i83-fence-compound-commands  consumed · not merged
 tasks     4 planned · 4 done · 0 blocked · first-pass 3/4
 size      delivered  code 472 · test 195 · docs 12 · test:code 0.41
           churn      code 477 · test 195 · docs 12 · rework 1.01
 time      agent 21.5 min (work 19.5 · review 2.0) · plan 6.5 min · gate review 0.7 min · verify 1.3 min · gates 0.1 · wall 49.9 min
 rate      22.0 code lines/min · 31.0 incl. tests
 cost      $4.60 · plan 2.03 · work 1.61 · review 0.52 · gate review 0.19 · verify 0.25 · $9.74/kloc
 models    plan claude-opus-5-5 · work claude-sonnet-5-5 · review claude-sonnet-5-5 · gate review claude-sonnet-5-5 · verify claude-sonnet-5-5
 defects   in-loop 0 · operator 1 · escaped 0 · removal efficiency 100%
 interventions  2 · recommended 0 · other-option 0 · adjusted 0 · different 0 · no-options 2
 instructions  plan 4,693 · gate review 2,338 · work 2,435 · review 2,456 · verify 2,109 words · guidelines 2 rules
```

```
id  area  kind     att  code+  test+  docs+  other+  test:code  $/kloc  agent  cost   model
T1  -     feature  1    97     67     0      0       0.69       5.86    7m06s  $0.57  claude-sonnet-5-5
T2  -     feature  2    378    128    0      0       0.34       2.83    5m33s  $1.07  claude-sonnet-5-5
T3  -     docs     1    2      0      12     0       0.00       115.02  0m43s  $0.23  claude-sonnet-5-5
T4  -     test     1    0      0      0      0       -          -       8m07s  $0.26  claude-sonnet-5-5
```

- I20261103-045334-t2-s-gate-wanted-the-any-text-message-wh — verification-finding, plan, found by assistant: T2's gate wanted the any-text message where the as-written match comes first
<!-- vloop:run-record:end -->
