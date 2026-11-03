# Fence audit: who enforces each rule in a Copilot review session

*Since B34 the driver re-judges every call a Copilot session ran and halts (exit 9, FENCE BYPASSED) on one the fence denies, and (HOOK CHANGED) when the hook's binary changed: where this audit says a call the hook fails open on goes unnoticed, that predates it.*

*Written for B22. It covers `fence/review.json`, the fence a review and a
second-review session run under, and the Copilot adapter's deny list in
`internal/driver/agent_copilot.go`.*

The review fence is a Claude Code settings file. Copilot reads no such file, so
before B22 a Copilot review session ran with the adapter's deny list alone. A
Copilot session now runs under a `preToolUse` hook, `vloop fence hook`, which
asks the judge behind `vloop fence check` about every tool call. This audit says,
rule by rule, who enforces it there.

## The three enforcers

- **hook** — the judge. It reads the same embedded rule set the fence is made
  from (not a copy), plus the adapter's deny-list stems. A shell command is split
  at `;`, `&&`, `||`, `|`, `&` and newlines outside quotes and looked through
  `env`, `command`, `exec`, `nohup`, `time`, `nice`, `timeout`, `stdbuf`, `xargs`
  and `sh -c`; each simple command is matched with the fence's own matching
  (`pre:*` is the prefix alone or followed by a space, `*` elsewhere any run of
  characters). What it cannot parse it denies: a backtick, `$(`, `<(`, `>(`,
  `eval`, `source`, `.`, an unterminated quote, a `-c` string that is not literal,
  a command word starting with `$`. It reads `if`, `while`, `until`,
  `for NAME in …`, `!`, `{ … }` and subshells by judging every command inside
  them, and matches a word the shell expands (a `$`, a glob, a brace expansion)
  as any text; it refuses a command name that expands, an expanded word before
  the command name (a wrapper's option, or a shell's before its `-c` string),
  and a redirection target holding a glob or a brace expansion (B35). File tools (`create`, `edit`, `apply_patch`)
  and a shell redirection may write only under `.vloop/tmp/`; reads may not touch
  `~/.claude`, `~/.copilot` or `~/.config/gh`; a `skill` call must name one of
  vloop's skills (K1); every other tool is denied. It fails closed: an error
  answers deny. Copilot itself denies when the hook command is missing or exits
  non-zero, but runs the call when the hook prints output that is not JSON or
  exceeds its timeout (`docs/design-notes/copilot-cli-experiments.md` → B22.2);
  the hook always prints its answer and is fast, and the deny list stays behind it.
  Neither parallel tool calls nor editing the hooks file mid-session gets a call
  past the hook (B22.3): the timeout runs per hook, and the file is read once.
- **driver** — what the driver checks after the session, whatever the session
  did: it restores the plan, the brief's status, the gate fixtures, `.vloop/config.toml`,
  the guidelines file, `.vloop/defects/`, `.vloop/interventions/` and the run
  folder's sessions and reports; it reverts changes to the work a review session
  judges and fails the review; and it halts the run with exit 9 on a moved ref,
  a changed `.git/config` entry or a changed git hook. These checks are the
  boundary (S-2, S-4); the hook is a line before them.
- **neither** — no code stops the call in a Copilot review session; the entry says
  what covers it.

A rule named **hook** is enforced only for the forms the judge can see. The
fence, and so the hook, is advisory: "a deny list cannot name every form of a
command".

## The review fence's deny rules

| Rule | Enforced by | Note |
| --- | --- | --- |
| `Bash(git push:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git commit:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git reset:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git clean:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git branch:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git checkout:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git switch:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git remote add:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git remote set-url:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git remote remove:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git remote rm:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git remote rename:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git remote set-head:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git remote set-branches:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git remote prune:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git remote update:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git update-ref:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git symbolic-ref:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git tag:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git stash:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git rebase:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(rm -rf:*)` | hook, driver | hook matches it (the adapter's `rm` stem denies every `rm`); driver reverts and fails the review if a file changed |
| `Bash(sudo:*)` | hook, driver | hook matches it; driver sees only its effects in the tree |
| `Bash(curl:*)` | hook, driver | hook matches it, and the adapter's argv denies it too; neither limits a network tool the lists do not name |
| `Bash(wget:*)` | hook, driver | hook matches it, and the adapter's argv denies it too; neither limits a network tool the lists do not name |
| `Bash(pip install:*)` | hook, driver | hook matches it; driver sees only its effects in the tree |
| `Bash(claude:*)` | hook, driver | hook matches it; driver restores the plan and halts on a moved ref |
| `Bash(vloop run:*)` | hook, driver | hook matches it; driver restores the plan and halts on a moved ref |
| `Bash(vloop init:*)` | hook, driver | hook matches it; driver restores the plan and halts on a moved ref |
| `Bash(vloop upgrade:*)` | hook, driver | hook matches it; driver restores the plan and halts on a moved ref |
| `Bash(vloop brief close:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop task verify:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop task reset:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop task note:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop task drop:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop task set:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop defect add:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop defect set:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop intervention add:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop intervention set:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop config set:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(.loop/run.sh:*)` | hook, driver | hook matches it; driver restores the plan and halts on a moved ref |
| `Edit(.git/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; a shell program that writes (`tee`, `cp`, `sed -i`) is not seen by the hook, so for those neither, and the driver halts on a moved ref or a changed `.git/config` or hook |
| `Write(.git/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; a shell program that writes (`tee`, `cp`, `sed -i`) is not seen by the hook, so for those neither, and the driver halts on a moved ref or a changed `.git/config` or hook |
| `Edit(.vloop/state/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the plan and the gate fixtures after the session |
| `Write(.vloop/state/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the plan and the gate fixtures after the session |
| `Edit(.vloop/config.toml)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Write(.vloop/config.toml)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Edit(.vloop/defects/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Write(.vloop/defects/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Edit(.vloop/interventions/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Write(.vloop/interventions/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Edit(.vloop/repo-guidelines.md)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Write(.vloop/repo-guidelines.md)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Edit(.vloop/install.json)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Write(.vloop/install.json)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Read(~/.claude/**)` | hook | hook: a read tool of a path there is denied, as is a redirection from it; neither for `cat ~/.claude/...` run as a command, which the hook does not read as a read — the session's `COPILOT_HOME` is empty and the hook denies `~/.copilot` and `~/.config/gh`, so what a session can reach is what the user's Claude directory holds |
| `Edit(~/.claude/**)` | hook | hook: writes outside `.vloop/tmp/` are denied; a shell program that writes there is neither (the driver checks only the repository) |
| `Write(~/.claude/**)` | hook | hook: writes outside `.vloop/tmp/` are denied; a shell program that writes there is neither (the driver checks only the repository) |
| `WebFetch` | hook | hook denies every tool not listed, `web_fetch` among them; the adapter's `url` deny stands in |
| `WebSearch` | hook | hook denies every tool not listed, `web_fetch` among them; the adapter's `url` deny stands in |
| `Bash(git -C:*)` | hook, driver | hook matches it; driver halts the run (exit 9) if `.git/config` or a hook changed |
| `Bash(git -c:*)` | hook, driver | hook matches it; driver halts the run (exit 9) if `.git/config` or a hook changed |
| `Bash(git --git-dir:*)` | hook, driver | hook matches it; driver halts the run (exit 9) if `.git/config` or a hook changed |
| `Bash(git --work-tree:*)` | hook, driver | hook matches it; driver halts the run (exit 9) if `.git/config` or a hook changed |
| `Bash(git --exec-path:*)` | hook, driver | hook matches it; driver halts the run (exit 9) if `.git/config` or a hook changed |
| `Bash(git merge:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git cherry-pick:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git revert:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git am:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git pull:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git fetch:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git notes:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git replace:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git filter-branch:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git worktree:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git reflog:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git gc:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git config:*)` | hook, driver | hook matches it; driver halts the run (exit 9) if `.git/config` or a hook changed |
| `Bash(git read-tree:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git submodule:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git prune:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git repack:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git hash-object:*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts and fails the review if a file changed |
| `Bash(git fast-import:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git format-patch:*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts and fails the review if a file changed |
| `Bash(git diff*--output*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts and fails the review if a file changed |
| `Bash(git log*--output*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts and fails the review if a file changed |
| `Bash(git show*--output*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts and fails the review if a file changed |
| `Bash(find*-delete*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts and fails the review if a file changed |
| `Bash(find*-exec*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts and fails the review if a file changed |
| `Bash(find*-ok*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts and fails the review if a file changed |
| `Bash(find*-fprint*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts and fails the review if a file changed |
| `Bash(find*-fls*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts and fails the review if a file changed |
| `Bash(rm -r:*)` | hook, driver | hook matches it (the adapter's `rm` stem denies every `rm`); driver reverts and fails the review if a file changed |
| `Bash(rm -R:*)` | hook, driver | hook matches it (the adapter's `rm` stem denies every `rm`); driver reverts and fails the review if a file changed |
| `Bash(rm -fr:*)` | hook, driver | hook matches it (the adapter's `rm` stem denies every `rm`); driver reverts and fails the review if a file changed |
| `Bash(rm -fR:*)` | hook, driver | hook matches it (the adapter's `rm` stem denies every `rm`); driver reverts and fails the review if a file changed |
| `Bash(rm --recursive:*)` | hook, driver | hook matches it (the adapter's `rm` stem denies every `rm`); driver reverts and fails the review if a file changed |
| `Bash(vloop -C:*)` | hook, driver | hook matches it; driver restores the plan and halts on a moved ref |
| `Bash(vloop --chdir:*)` | hook, driver | hook matches it; driver restores the plan and halts on a moved ref |
| `Bash(vloop brief new:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop brief review:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |

## The adapter's deny list

`copilotDenials` goes to Copilot as `--deny-tool` on the argv, so Copilot itself
refuses these, and the judge reads the `shell(…)` stems as further Bash deny rules
(`CopilotDenyStems`). Copilot's own matching is not vloop's: it is the stem alone
or followed by a space, so another form of the command can pass it.

| Pattern | Enforced by | Note |
| --- | --- | --- |
| `shell(gh:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git push:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git remote:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git config --global:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git config --system:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git config --file:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(curl:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(wget:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(ssh:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(scp:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(nc:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(rm:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git commit:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git reset:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git checkout:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git switch:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git stash:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git branch:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git tag:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `url` | hook | Copilot's own URL tool is refused on the argv; the hook denies `web_fetch` and every tool it does not list |

## What is neither, and what covers it

- **A write by a shell program other than a redirection** (`tee`, `cp`, `mv`,
  `sed -i`, `python -c`): the judge reads redirections and file tools, not what a
  program does. The hook does not see it. Covered by the driver: it reverts the
  work a review session changed and fails the review, restores the files it
  guards, and halts on a moved ref or a changed `.git/config` or hook.
- **A read of `~/.claude` by a program** (`cat ~/.claude/…`): the hook denies a
  read tool and a redirection from there, not a program's argument. Nothing after
  the session undoes a read. Nothing covers it
  beyond the session's `COPILOT_HOME` being its own empty folder.
- **A command no list names**: both lists are denylists. Neither the hook nor
  Copilot refuses a tool they do not know to refuse, except that the hook denies
  every tool it does not list. The driver's checks are the boundary.
- **Network from an allowed program**: `curl`, `wget`, `ssh`, `scp`, `nc` are
  denied; nothing stops a program from opening its own connection. Neither; the
  review fence's allow rules name few programs, and the session timeout bounds
  how long one can run.

# Fence audit: who enforces each rule in a Copilot work session

*Since B34 the driver re-judges every call a Copilot session ran and halts (exit 9, FENCE BYPASSED) on one the fence denies, and (HOOK CHANGED) when the hook's binary changed: where this audit says a call the hook fails open on goes unnoticed, that predates it.*

*Written for B24. It covers `fence/work.json`, the fence a work session runs
under, and the Copilot adapter's deny list for work sessions: the review list
without `shell(rm:*)`.*

A Copilot work session runs under the same `preToolUse` hook, `vloop fence hook
--phase work`. The judge reads `fence/work.json` instead of `fence/review.json`
(its 102 deny rules are the same ones) and the adapter's deny-list stems without
`rm:*`. The three enforcers are the ones above; they differ for work sessions in
three ways:

- **Writes.** A work session edits the repository. The hook allows a file tool or
  a redirection to write anywhere under the repository root except `.git/` and
  `.vloop/`, where only `.vloop/tmp/` may be written. A write outside the
  repository is denied, and so is the same write by a relative path that leaves it.
  `/dev/null` is allowed.
- **`rm`.** `rm` of a file is allowed in a work session: deleting and renaming
  files is part of the work, and the review session's blanket `rm` deny would stop
  it. The recursive forms (`rm -r`, `-R`, `-rf`, `-fr`, `-fR`, `--recursive`) stay
  denied by the fence rules below. `shell(rm:*)` is not on a work session's
  command line either.
- **The skill.** The `skill` tool is allowed only for the names of vloop's
  skills, and a refusal reads `vloop fence: skill <name> is not vloop's — use
  /vloop:work`.

Shell syntax the judge cannot read is denied in a work session as in a review one:
a backtick or `$(` answers `command substitution cannot be judged`; `<(`, `>(`,
`((` and `case`, `subshells and process substitution cannot be judged`; `eval`,
`source` and `.`, `… is not allowed`; `select`, `function`, `coproc` and `[[`,
`shell syntax … cannot be judged`. `if`, `while`, `until`, `for`, `!`, `{ … }`
and subshells are judged through (B35). The hook fails open on the
two limits above (B22.2, B22.3), for work sessions too: it always prints its
answer and is fast, and the deny list stays behind it.

## The work fence's deny rules

| Rule | Enforced by | Note |
| --- | --- | --- |
| `Bash(git push:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git commit:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git reset:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git clean:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git branch:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git checkout:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git switch:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git remote add:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git remote set-url:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git remote remove:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git remote rm:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git remote rename:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git remote set-head:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git remote set-branches:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git remote prune:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git remote update:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git update-ref:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git symbolic-ref:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git tag:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git stash:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git rebase:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(rm -rf:*)` | hook | hook matches the recursive form; plain `rm` of a file is allowed (see below); a recursive delete spelled another way is neither |
| `Bash(sudo:*)` | hook | hook matches it; driver sees only its effects in the tree |
| `Bash(curl:*)` | hook, driver | hook matches it, and the adapter's argv denies it too; neither limits a network tool the lists do not name |
| `Bash(wget:*)` | hook, driver | hook matches it, and the adapter's argv denies it too; neither limits a network tool the lists do not name |
| `Bash(pip install:*)` | hook | hook matches it; driver sees only its effects in the tree |
| `Bash(claude:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(vloop run:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(vloop init:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(vloop upgrade:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(vloop brief close:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(vloop task verify:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(vloop task reset:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(vloop task note:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(vloop task drop:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(vloop task set:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(vloop defect add:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(vloop defect set:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(vloop intervention add:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(vloop intervention set:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(vloop config set:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(.loop/run.sh:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Edit(.git/**)` | hook, driver | hook: a work session may write under the repository except `.git/` and `.vloop/` (but `.vloop/tmp/`); driver halts the run (exit 9) on a moved ref or a changed `.git/config` or hook |
| `Write(.git/**)` | hook, driver | hook: a work session may write under the repository except `.git/` and `.vloop/` (but `.vloop/tmp/`); driver halts the run (exit 9) on a moved ref or a changed `.git/config` or hook |
| `Edit(.vloop/state/**)` | hook, driver | hook denies a write there but under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Write(.vloop/state/**)` | hook, driver | hook denies a write there but under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Edit(.vloop/config.toml)` | hook, driver | hook denies a write there but under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Write(.vloop/config.toml)` | hook, driver | hook denies a write there but under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Edit(.vloop/defects/**)` | hook, driver | hook denies a write there but under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Write(.vloop/defects/**)` | hook, driver | hook denies a write there but under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Edit(.vloop/interventions/**)` | hook, driver | hook denies a write there but under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Write(.vloop/interventions/**)` | hook, driver | hook denies a write there but under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Edit(.vloop/repo-guidelines.md)` | hook, driver | hook denies a write there but under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Write(.vloop/repo-guidelines.md)` | hook, driver | hook denies a write there but under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Edit(.vloop/install.json)` | hook, driver | hook denies a write there but under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Write(.vloop/install.json)` | hook, driver | hook denies a write there but under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Read(~/.claude/**)` | hook, neither | hook: a read tool of a path there is denied, as is a redirection from it; neither for `cat ~/.claude/...` run as a command, which the hook does not read as a read |
| `Edit(~/.claude/**)` | hook | hook: a write outside the repository is denied; a shell program that writes there is neither (see below) |
| `Write(~/.claude/**)` | hook | hook: a write outside the repository is denied; a shell program that writes there is neither (see below) |
| `WebFetch` | hook | hook denies every tool not listed, `web_fetch` among them; the adapter's `url` deny stands in |
| `WebSearch` | hook | hook denies every tool not listed, `web_fetch` among them; the adapter's `url` deny stands in |
| `Bash(git -C:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git -c:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git --git-dir:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git --work-tree:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git --exec-path:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git merge:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git cherry-pick:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git revert:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git am:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git pull:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git fetch:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git notes:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git replace:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git filter-branch:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git worktree:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git reflog:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git gc:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git config:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git read-tree:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git submodule:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git prune:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git repack:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git hash-object:*)` | hook | hook matches it, by the fence's own `*` matching; a repository file written or removed by another program is neither (see below) |
| `Bash(git fast-import:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(git format-patch:*)` | hook | hook matches it, by the fence's own `*` matching; a repository file written or removed by another program is neither (see below) |
| `Bash(git diff*--output*)` | hook | hook matches it, by the fence's own `*` matching; a repository file written or removed by another program is neither (see below) |
| `Bash(git log*--output*)` | hook | hook matches it, by the fence's own `*` matching; a repository file written or removed by another program is neither (see below) |
| `Bash(git show*--output*)` | hook | hook matches it, by the fence's own `*` matching; a repository file written or removed by another program is neither (see below) |
| `Bash(find*-delete*)` | hook | hook matches it, by the fence's own `*` matching; a repository file written or removed by another program is neither (see below) |
| `Bash(find*-exec*)` | hook | hook matches it, by the fence's own `*` matching; a repository file written or removed by another program is neither (see below) |
| `Bash(find*-ok*)` | hook | hook matches it, by the fence's own `*` matching; a repository file written or removed by another program is neither (see below) |
| `Bash(find*-fprint*)` | hook | hook matches it, by the fence's own `*` matching; a repository file written or removed by another program is neither (see below) |
| `Bash(find*-fls*)` | hook | hook matches it, by the fence's own `*` matching; a repository file written or removed by another program is neither (see below) |
| `Bash(rm -r:*)` | hook | hook matches the recursive form; plain `rm` of a file is allowed (see below); a recursive delete spelled another way is neither |
| `Bash(rm -R:*)` | hook | hook matches the recursive form; plain `rm` of a file is allowed (see below); a recursive delete spelled another way is neither |
| `Bash(rm -fr:*)` | hook | hook matches the recursive form; plain `rm` of a file is allowed (see below); a recursive delete spelled another way is neither |
| `Bash(rm -fR:*)` | hook | hook matches the recursive form; plain `rm` of a file is allowed (see below); a recursive delete spelled another way is neither |
| `Bash(rm --recursive:*)` | hook | hook matches the recursive form; plain `rm` of a file is allowed (see below); a recursive delete spelled another way is neither |
| `Bash(vloop -C:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(vloop --chdir:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(vloop brief new:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |
| `Bash(vloop brief review:*)` | hook, driver | hook matches it (the fence's own matching); driver restores the plan, defects, interventions and config and halts on a moved ref or a changed `.git/config` or hook |

## The work adapter's deny list

The work deny list is `copilotDenials` without `shell(rm:*)`: it goes to Copilot
as `--deny-tool` on the argv, and the judge reads its `shell(…)` stems as further
Bash deny rules. `rm` is the one pattern a work session does not carry.

| Pattern | Enforced by | Note |
| --- | --- | --- |
| `shell(gh:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git push:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git remote:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git config --global:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git config --system:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git config --file:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(curl:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(wget:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(ssh:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(scp:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(nc:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git commit:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git reset:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git checkout:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git switch:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git stash:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git branch:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git tag:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `url` | hook | Copilot's own URL tool is refused on the argv; the hook denies `web_fetch` and every tool it does not list |

## What is neither in a work session, and what covers it

- **A shell program that writes outside what the hook reads** (`tee`, `cp`,
  `sed -i`, `python -c`): the judge reads redirections and file tools, not what a
  program does. For `.git/`, `.vloop/state/` and the files the driver guards, the
  driver restores them after the session and halts on a moved ref or a changed
  `.git/config` or hook. Inside the repository's source the write is the work
  itself; the gate and the review judge it.
- **A recursive delete spelled another way** (`find … -delete` is denied, but a
  program that removes a tree is not): `rm` of a file is allowed because deleting
  files is work; the review and the gate judge the result. Nothing undoes it.
- **A read of `~/.claude` by a program** (`cat ~/.claude/…`): as in a review
  session, nothing covers it beyond the session's `COPILOT_HOME` being its own
  empty folder.
- **A command no list names**: both lists are denylists; the hook denies every
  tool it does not list, and the driver's checks are the boundary.
- **Network from an allowed program**: `curl`, `wget`, `ssh`, `scp`, `nc` are
  denied; nothing stops a program from opening its own connection. Neither; the
  session timeout bounds how long one can run.

# Fence audit: who enforces each rule in a Copilot verify session

*Since B34 the driver re-judges every call a Copilot session ran and halts (exit 9, FENCE BYPASSED) on one the fence denies, and (HOOK CHANGED) when the hook's binary changed: where this audit says a call the hook fails open on goes unnoticed, that predates it.*

*Written for B25. It covers `fence/verify.json`, the fence a verify session runs
under, and the Copilot adapter's deny list for verify sessions: the whole review
list, `shell(rm:*)` included.*

A Copilot verify session (`agent.verify = copilot`) runs under the same
`preToolUse` hook, `vloop fence hook --phase verify`. The judge reads
`fence/verify.json` instead of `fence/review.json` (its 102 deny rules are the
same ones, so a change to either fence moves only its own phase) and every
`shell(…)` stem of the adapter's deny list. The three enforcers are the ones
above; a verify session decides like a review session:

- **Writes.** A verify session reports findings; it builds nothing. The hook
  allows a file tool or a redirection to write only under `.vloop/tmp/`, where
  the session writes its report. `/dev/null` is allowed, and a write outside the
  repository is denied.
- **`rm`.** Every `rm` is denied, `shell(rm:*)` among the stems, as in a review
  session: there is nothing for a verify session to delete.
- **The skill.** The `skill` tool is allowed only for the names of vloop's
  skills, and a refusal reads `vloop fence: skill <name> is not vloop's — use
  /vloop:verify`.

Shell syntax the judge cannot read is denied as in the other phases. The hook
fails open on the two limits above (B22.2, B22.3) for verify sessions too: it
always prints its answer and is fast, and the deny list stays behind it.

## The verify fence's deny rules

| Rule | Enforced by | Note |
| --- | --- | --- |
| `Bash(git push:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git commit:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git reset:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git clean:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git branch:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git checkout:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git switch:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git remote add:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git remote set-url:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git remote remove:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git remote rm:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git remote rename:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git remote set-head:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git remote set-branches:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git remote prune:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git remote update:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git update-ref:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git symbolic-ref:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git tag:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git stash:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git rebase:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(rm -rf:*)` | hook, driver | hook matches it (the adapter's `rm` stem denies every `rm`); driver reverts a file the session changed and warns `the verify session changed <path> — reverted` |
| `Bash(sudo:*)` | hook, driver | hook matches it; driver sees only its effects in the tree |
| `Bash(curl:*)` | hook, driver | hook matches it, and the adapter's argv denies it too; neither limits a network tool the lists do not name |
| `Bash(wget:*)` | hook, driver | hook matches it, and the adapter's argv denies it too; neither limits a network tool the lists do not name |
| `Bash(pip install:*)` | hook, driver | hook matches it; driver sees only its effects in the tree |
| `Bash(claude:*)` | hook, driver | hook matches it; driver restores the plan and halts on a moved ref |
| `Bash(vloop run:*)` | hook, driver | hook matches it; driver restores the plan and halts on a moved ref |
| `Bash(vloop init:*)` | hook, driver | hook matches it; driver restores the plan and halts on a moved ref |
| `Bash(vloop upgrade:*)` | hook, driver | hook matches it; driver restores the plan and halts on a moved ref |
| `Bash(vloop brief close:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop task verify:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop task reset:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop task note:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop task drop:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop task set:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop defect add:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop defect set:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop intervention add:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop intervention set:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop config set:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(.loop/run.sh:*)` | hook, driver | hook matches it; driver restores the plan and halts on a moved ref |
| `Edit(.git/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; a shell program that writes (`tee`, `cp`, `sed -i`) is not seen by the hook, so for those neither, and the driver halts on a moved ref or a changed `.git/config` or hook |
| `Write(.git/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; a shell program that writes (`tee`, `cp`, `sed -i`) is not seen by the hook, so for those neither, and the driver halts on a moved ref or a changed `.git/config` or hook |
| `Edit(.vloop/state/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the plan and the gate fixtures after the session |
| `Write(.vloop/state/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the plan and the gate fixtures after the session |
| `Edit(.vloop/config.toml)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Write(.vloop/config.toml)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Edit(.vloop/defects/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Write(.vloop/defects/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Edit(.vloop/interventions/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Write(.vloop/interventions/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Edit(.vloop/repo-guidelines.md)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Write(.vloop/repo-guidelines.md)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Edit(.vloop/install.json)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Write(.vloop/install.json)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; driver restores the file or folder after the session |
| `Read(~/.claude/**)` | hook | hook: a read tool of a path there is denied, as is a redirection from it; neither for `cat ~/.claude/...` run as a command, which the hook does not read as a read — the session's `COPILOT_HOME` is empty and the hook denies `~/.copilot` and `~/.config/gh`, so what a session can reach is what the user's Claude directory holds |
| `Edit(~/.claude/**)` | hook | hook: writes outside `.vloop/tmp/` are denied; a shell program that writes there is neither (the driver checks only the repository) |
| `Write(~/.claude/**)` | hook | hook: writes outside `.vloop/tmp/` are denied; a shell program that writes there is neither (the driver checks only the repository) |
| `WebFetch` | hook | hook denies every tool not listed, `web_fetch` among them; the adapter's `url` deny stands in |
| `WebSearch` | hook | hook denies every tool not listed, `web_fetch` among them; the adapter's `url` deny stands in |
| `Bash(git -C:*)` | hook, driver | hook matches it; driver halts the run (exit 9) if `.git/config` or a hook changed |
| `Bash(git -c:*)` | hook, driver | hook matches it; driver halts the run (exit 9) if `.git/config` or a hook changed |
| `Bash(git --git-dir:*)` | hook, driver | hook matches it; driver halts the run (exit 9) if `.git/config` or a hook changed |
| `Bash(git --work-tree:*)` | hook, driver | hook matches it; driver halts the run (exit 9) if `.git/config` or a hook changed |
| `Bash(git --exec-path:*)` | hook, driver | hook matches it; driver halts the run (exit 9) if `.git/config` or a hook changed |
| `Bash(git merge:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git cherry-pick:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git revert:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git am:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git pull:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git fetch:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git notes:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git replace:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git filter-branch:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git worktree:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git reflog:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git gc:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git config:*)` | hook, driver | hook matches it; driver halts the run (exit 9) if `.git/config` or a hook changed |
| `Bash(git read-tree:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git submodule:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git prune:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git repack:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git hash-object:*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts a file the session changed and warns `the verify session changed <path> — reverted` |
| `Bash(git fast-import:*)` | hook, driver | hook matches the git command; driver halts the run (exit 9) if a ref moved |
| `Bash(git format-patch:*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts a file the session changed and warns `the verify session changed <path> — reverted` |
| `Bash(git diff*--output*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts a file the session changed and warns `the verify session changed <path> — reverted` |
| `Bash(git log*--output*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts a file the session changed and warns `the verify session changed <path> — reverted` |
| `Bash(git show*--output*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts a file the session changed and warns `the verify session changed <path> — reverted` |
| `Bash(find*-delete*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts a file the session changed and warns `the verify session changed <path> — reverted` |
| `Bash(find*-exec*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts a file the session changed and warns `the verify session changed <path> — reverted` |
| `Bash(find*-ok*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts a file the session changed and warns `the verify session changed <path> — reverted` |
| `Bash(find*-fprint*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts a file the session changed and warns `the verify session changed <path> — reverted` |
| `Bash(find*-fls*)` | hook, driver | hook matches it, by the fence's own `*` matching; driver reverts a file the session changed and warns `the verify session changed <path> — reverted` |
| `Bash(rm -r:*)` | hook, driver | hook matches it (the adapter's `rm` stem denies every `rm`); driver reverts a file the session changed and warns `the verify session changed <path> — reverted` |
| `Bash(rm -R:*)` | hook, driver | hook matches it (the adapter's `rm` stem denies every `rm`); driver reverts a file the session changed and warns `the verify session changed <path> — reverted` |
| `Bash(rm -fr:*)` | hook, driver | hook matches it (the adapter's `rm` stem denies every `rm`); driver reverts a file the session changed and warns `the verify session changed <path> — reverted` |
| `Bash(rm -fR:*)` | hook, driver | hook matches it (the adapter's `rm` stem denies every `rm`); driver reverts a file the session changed and warns `the verify session changed <path> — reverted` |
| `Bash(rm --recursive:*)` | hook, driver | hook matches it (the adapter's `rm` stem denies every `rm`); driver reverts a file the session changed and warns `the verify session changed <path> — reverted` |
| `Bash(vloop -C:*)` | hook, driver | hook matches it; driver restores the plan and halts on a moved ref |
| `Bash(vloop --chdir:*)` | hook, driver | hook matches it; driver restores the plan and halts on a moved ref |
| `Bash(vloop brief new:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |
| `Bash(vloop brief review:*)` | hook, driver | hook matches it; driver restores the plan, defects, interventions and config it guards |

## The verify adapter's deny list

The verify deny list is `copilotDenials` whole, `shell(rm:*)` included: it goes
to Copilot as `--deny-tool` on the argv, and the judge reads its `shell(…)` stems
as further Bash deny rules.

| Pattern | Enforced by | Note |
| --- | --- | --- |
| `shell(gh:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git push:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git remote:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git config --global:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git config --system:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git config --file:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(curl:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(wget:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(ssh:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(scp:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(nc:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(rm:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git commit:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git reset:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git checkout:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git switch:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git stash:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git branch:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git tag:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `url` | hook | Copilot's own URL tool is refused on the argv; the hook denies `web_fetch` and every tool it does not list |

## What is neither in a verify session, and what covers it

- **A write by a shell program other than a redirection** (`tee`, `cp`, `mv`,
  `sed -i`, `python -c`): the judge reads redirections and file tools, not what a
  program does. Covered by the driver: after the session it restores the plan,
  the gates and the files it guards, reverts any other file the session changed
  with a warning, and halts on a moved ref or a changed `.git/config` or hook.
- **A read of `~/.claude` by a program** (`cat ~/.claude/…`): neither. The hook
  denies a read tool and a redirection from there, not a program's argument;
  nothing covers it beyond the session's `COPILOT_HOME` being its own empty folder.
- **A command no list names**: both lists are denylists; the hook denies every
  tool it does not list, and the driver's checks are the boundary.
- **Network from an allowed program**: `curl`, `wget`, `ssh`, `scp`, `nc` are
  denied; nothing stops a program from opening its own connection. Neither; the
  session timeout bounds how long one can run.

# Fence audit: who enforces each rule in a Copilot gate-review session

*Since B34 the driver re-judges every call a Copilot session ran and halts (exit 9, FENCE BYPASSED) on one the fence denies, and (HOOK CHANGED) when the hook's binary changed: where this audit says a call the hook fails open on goes unnoticed, that predates it.*

*Since B31 the driver reverts a file a gate-review or plan session changed outside `.vloop/state/` and `.vloop/tmp/`, with a warning, before the plan commit: where this audit says the driver does not revert files after such a session, or that the plan commit would commit the change, that predates it.*

*Written for B26. It covers `fence/gate-review.json`, the fence a gate-review
session runs under, and the Copilot adapter's deny list for gate-review
sessions: the whole review list, `shell(rm:*)` included.*

A Copilot gate-review session (`agent.gate-review = copilot`) runs under the same
`preToolUse` hook, `vloop fence hook --phase gate-review`. The judge reads
`fence/gate-review.json` (its 102 deny rules are the same ones as the review
fence's, so a change to either fence moves only its own phase) and every
`shell(…)` stem of the adapter's deny list. The enforcers are the ones above, but
the driver's part is smaller here:

- **Writes.** A gate-review session reports a verdict; it builds nothing. The hook
  allows a file tool or a redirection to write only under `.vloop/tmp/`, where the
  session writes `gate-verdict.json`. `/dev/null` is allowed, and a write outside
  the repository is denied.
- **What the driver checks after the session.** Only git: it halts the run with
  exit 9 if a ref moved (`REFS MOVED gate-review`) or if `.git/config` or a git
  hook changed (`the gate review changed .git/config (…)` or `… the git hooks`).
  It does **not** revert a file the session changed outside `.vloop/tmp/`: the
  plan commit stages the whole tree, so such a change would be committed. The hook
  is the line for every file write (out of scope to change); the rows below say
  *hook*, not *driver*, for them.
- **`rm`.** Every `rm` is denied, `shell(rm:*)` among the stems, as in a review
  session: there is nothing for a gate-review session to delete.
- **The skill.** The `skill` tool is allowed only for the names of vloop's
  skills, and a refusal reads `vloop fence: skill <name> is not vloop's — use
  /vloop:gate-review`.

Shell syntax the judge cannot read is denied as in the other phases. The hook
fails open on the two limits above (B22.2, B22.3) for gate-review sessions too: it
always prints its answer and is fast, and the deny list stays behind it.

## The gate-review fence's deny rules

| Rule | Enforced by | Note |
| --- | --- | --- |
| `Bash(git push:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git commit:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git reset:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git clean:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git branch:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git checkout:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git switch:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git remote add:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git remote set-url:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git remote remove:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git remote rm:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git remote rename:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git remote set-head:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git remote set-branches:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git remote prune:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git remote update:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git update-ref:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git symbolic-ref:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git tag:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git stash:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git rebase:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(rm -rf:*)` | hook | hook matches it (the adapter's `rm` stem denies every `rm`); the driver does not revert files after a gate-review session |
| `Bash(sudo:*)` | hook | hook |
| `Bash(curl:*)` | hook | hook |
| `Bash(wget:*)` | hook | hook |
| `Bash(pip install:*)` | hook | hook |
| `Bash(claude:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Bash(vloop run:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Bash(vloop init:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Bash(vloop upgrade:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Bash(vloop brief close:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Bash(vloop task verify:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Bash(vloop task reset:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Bash(vloop task note:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Bash(vloop task drop:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Bash(vloop task set:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Bash(vloop defect add:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Bash(vloop defect set:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Bash(vloop intervention add:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Bash(vloop intervention set:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Bash(vloop config set:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Bash(.loop/run.sh:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Edit(.git/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; a shell program that writes there is neither (see below), though the driver halts on a changed `.git/config` or hook |
| `Write(.git/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; a shell program that writes there is neither (see below), though the driver halts on a changed `.git/config` or hook |
| `Edit(.vloop/state/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; the driver does not restore this path after a gate-review session, so the hook is the line |
| `Write(.vloop/state/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; the driver does not restore this path after a gate-review session, so the hook is the line |
| `Edit(.vloop/config.toml)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; the driver does not restore this path after a gate-review session, so the hook is the line |
| `Write(.vloop/config.toml)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; the driver does not restore this path after a gate-review session, so the hook is the line |
| `Edit(.vloop/defects/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; the driver does not restore this path after a gate-review session, so the hook is the line |
| `Write(.vloop/defects/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; the driver does not restore this path after a gate-review session, so the hook is the line |
| `Edit(.vloop/interventions/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; the driver does not restore this path after a gate-review session, so the hook is the line |
| `Write(.vloop/interventions/**)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; the driver does not restore this path after a gate-review session, so the hook is the line |
| `Edit(.vloop/repo-guidelines.md)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; the driver does not restore this path after a gate-review session, so the hook is the line |
| `Write(.vloop/repo-guidelines.md)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; the driver does not restore this path after a gate-review session, so the hook is the line |
| `Edit(.vloop/install.json)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; the driver does not restore this path after a gate-review session, so the hook is the line |
| `Write(.vloop/install.json)` | hook | hook: file tools and redirections may write only under `.vloop/tmp/`; the driver does not restore this path after a gate-review session, so the hook is the line |
| `Read(~/.claude/**)` | hook | hook: a read tool of a path there is denied, as is a redirection from it; neither for `cat ~/.claude/...` run as a command, and for a write there by a shell program (the driver checks only the repository) |
| `Edit(~/.claude/**)` | hook | hook |
| `Write(~/.claude/**)` | hook | hook |
| `WebFetch` | hook | hook denies every tool not listed, `web_fetch` among them; the adapter's `url` deny stands in |
| `WebSearch` | hook | hook denies every tool not listed, `web_fetch` among them; the adapter's `url` deny stands in |
| `Bash(git -C:*)` | hook, driver | hook matches it; the driver halts the run (exit 9) if `.git/config` or a hook changed |
| `Bash(git -c:*)` | hook, driver | hook matches it; the driver halts the run (exit 9) if `.git/config` or a hook changed |
| `Bash(git --git-dir:*)` | hook, driver | hook matches it; the driver halts the run (exit 9) if `.git/config` or a hook changed |
| `Bash(git --work-tree:*)` | hook, driver | hook matches it; the driver halts the run (exit 9) if `.git/config` or a hook changed |
| `Bash(git --exec-path:*)` | hook, driver | hook matches it; the driver halts the run (exit 9) if `.git/config` or a hook changed |
| `Bash(git merge:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git cherry-pick:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git revert:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git am:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git pull:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git fetch:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git notes:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git replace:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git filter-branch:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git worktree:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git reflog:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git gc:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git config:*)` | hook, driver | hook matches it; the driver halts the run (exit 9) if `.git/config` or a hook changed |
| `Bash(git read-tree:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git submodule:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git prune:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git repack:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git hash-object:*)` | hook | hook matches it, by the fence's own `*` matching; the driver does not revert the file it writes (see what is neither, below) |
| `Bash(git fast-import:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved |
| `Bash(git format-patch:*)` | hook | hook matches it, by the fence's own `*` matching; the driver does not revert the file it writes (see what is neither, below) |
| `Bash(git diff*--output*)` | hook | hook matches it, by the fence's own `*` matching; the driver does not revert the file it writes (see what is neither, below) |
| `Bash(git log*--output*)` | hook | hook matches it, by the fence's own `*` matching; the driver does not revert the file it writes (see what is neither, below) |
| `Bash(git show*--output*)` | hook | hook matches it, by the fence's own `*` matching; the driver does not revert the file it writes (see what is neither, below) |
| `Bash(find*-delete*)` | hook | hook |
| `Bash(find*-exec*)` | hook | hook |
| `Bash(find*-ok*)` | hook | hook |
| `Bash(find*-fprint*)` | hook | hook |
| `Bash(find*-fls*)` | hook | hook |
| `Bash(rm -r:*)` | hook | hook matches it (the adapter's `rm` stem denies every `rm`); the driver does not revert files after a gate-review session |
| `Bash(rm -R:*)` | hook | hook matches it (the adapter's `rm` stem denies every `rm`); the driver does not revert files after a gate-review session |
| `Bash(rm -fr:*)` | hook | hook matches it (the adapter's `rm` stem denies every `rm`); the driver does not revert files after a gate-review session |
| `Bash(rm -fR:*)` | hook | hook matches it (the adapter's `rm` stem denies every `rm`); the driver does not revert files after a gate-review session |
| `Bash(rm --recursive:*)` | hook | hook matches it (the adapter's `rm` stem denies every `rm`); the driver does not revert files after a gate-review session |
| `Bash(vloop -C:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Bash(vloop --chdir:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Bash(vloop brief new:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |
| `Bash(vloop brief review:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a gate-review session, so the hook is the line |

## The gate-review adapter's deny list

`copilotDenials` goes to Copilot as `--deny-tool` on the argv for a gate-review
session as for a review session, `shell(rm:*)` included, and the judge reads the
`shell(…)` stems as further Bash deny rules. Copilot's own matching is not
vloop's: it is the stem alone or followed by a space, so another form of the
command can pass it. The hook is the line for each; the driver halts the run only
on a moved ref or a changed `.git/config` or hook.

| Pattern | Enforced by | Note |
| --- | --- | --- |
| `shell(gh:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git push:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git remote:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git config --global:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git config --system:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git config --file:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(curl:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(wget:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(ssh:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(scp:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(nc:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(rm:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git commit:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git reset:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git checkout:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git switch:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git stash:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git branch:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git tag:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `url` | hook | Copilot's own URL tool is refused on the argv; the hook denies `web_fetch` and every tool it does not list |

## What is neither in a gate-review session, and what covers it

- **A write by a shell program other than a redirection** (`tee`, `cp`, `mv`,
  `sed -i`, `python -c`): the judge reads redirections and file tools, not what a
  program does. Neither: the driver does not revert files after a gate-review
  session, and the plan commit stages the whole tree. What covers it is the
  hook's other lines (the `rm`, `git` and wrapper denials), the review of the
  commit by the operator, and the driver's halt on a moved ref or a changed
  `.git/config` or hook.
- **A read of `~/.claude` by a program** (`cat ~/.claude/…`): neither. The hook
  denies a read tool and a redirection from there, not a program's argument;
  nothing covers it beyond the session's `COPILOT_HOME` being its own empty folder.
- **A command no list names**: both lists are denylists; the hook denies every
  tool it does not list, and the driver's git checks are the boundary.
- **Network from an allowed program**: `curl`, `wget`, `ssh`, `scp`, `nc` are
  denied; nothing stops a program from opening its own connection. Neither; the
  session timeout bounds how long one can run.

# Fence audit: who enforces each rule in a Copilot plan session

*Since B34 the driver re-judges every call a Copilot session ran and halts (exit 9, FENCE BYPASSED) on one the fence denies, and (HOOK CHANGED) when the hook's binary changed: where this audit says a call the hook fails open on goes unnoticed, that predates it.*

*Since B31 the driver reverts a file a gate-review or plan session changed outside `.vloop/state/` and `.vloop/tmp/`, with a warning, before the plan commit: where this audit says the driver does not revert files after such a session, or that the plan commit would commit the change, that predates it.*

*Written for B27. It covers `fence/plan.json`, the fence a plan session runs
under, and the Copilot adapter's deny list for plan sessions: the whole review
list, `shell(rm:*)` included.*

A Copilot plan session (`agent.plan = copilot`) runs under the same `preToolUse`
hook, `vloop fence hook --phase plan`. The judge reads `fence/plan.json` (its 90
deny rules) and every `shell(…)` stem of the adapter's deny list. The enforcers
are the ones above, but the driver's part is small here:

- **Writes.** A plan session writes the plan and its gates. The hook allows a
  file tool or a redirection to write only `.vloop/state/state.json`, anything
  under `.vloop/state/gates/` and anything under `.vloop/tmp/` (S-4); the two
  folders themselves, a write elsewhere in the repository and a write outside it
  are denied, and `/dev/null` is allowed. The refusal reads `write to <path> is
  not allowed in a plan session: only .vloop/state/state.json,
  .vloop/state/gates/ and .vloop/tmp/ may be written`.
- **What the driver checks after the session.** Only git: it halts the run with
  exit 9 if a ref moved (`REFS MOVED plan`) or if `.git/config` or a git hook
  changed (`plan changed … — nothing was committed; restore it, then re-run`). It
  does **not** revert a file the session changed outside the files it may write:
  the plan commit stages the whole tree, so such a change would be committed. The
  hook is the line for every file write (out of scope to change); the rows below
  say *hook*, not *driver*, for them.
- **`rm`.** Every `rm` is denied, `shell(rm:*)` among the stems, as in a review
  session: a plan session replaces a file by writing it.
- **The skill.** The `skill` tool is allowed only for the names of vloop's
  skills, and a refusal reads `vloop fence: skill <name> is not vloop's — use
  /vloop:plan`.

Shell syntax the judge cannot read is denied as in the other phases. The hook
fails open on the two limits above (B22.2, B22.3) for plan sessions too: it
always prints its answer and is fast, and the deny list stays behind it.

## The plan fence's deny rules

| Rule | Enforced by | Note |
| --- | --- | --- |
| `Bash(git push:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git commit:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git reset:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git clean:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git branch:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git checkout:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git switch:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git remote add:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git remote set-url:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git remote remove:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git remote rm:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git remote rename:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git remote set-head:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git remote set-branches:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git remote prune:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git remote update:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git update-ref:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git symbolic-ref:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git tag:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git stash:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git rebase:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(rm -rf:*)` | hook | hook matches it (the adapter's `rm` stem denies every `rm`); the driver does not revert files after a plan session |
| `Bash(sudo:*)` | hook | hook |
| `Bash(curl:*)` | hook | hook |
| `Bash(wget:*)` | hook | hook |
| `Bash(pip install:*)` | hook | hook |
| `Bash(claude:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Bash(vloop run:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Bash(vloop init:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Bash(vloop upgrade:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Bash(vloop brief close:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Bash(vloop task verify:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Bash(vloop task reset:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Bash(vloop task note:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Bash(vloop task drop:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Bash(vloop task set:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Bash(vloop defect add:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Bash(vloop defect set:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Bash(vloop intervention add:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Bash(vloop intervention set:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Bash(vloop config set:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Bash(.loop/run.sh:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Edit(.git/**)` | hook | hook: a plan session's file tools and redirections may write only to `.vloop/state/state.json`, under `.vloop/state/gates/` and under `.vloop/tmp/`; a shell program that writes there is neither (see below), though the driver halts on a changed `.git/config` or hook |
| `Write(.git/**)` | hook | hook: a plan session's file tools and redirections may write only to `.vloop/state/state.json`, under `.vloop/state/gates/` and under `.vloop/tmp/`; a shell program that writes there is neither (see below), though the driver halts on a changed `.git/config` or hook |
| `Read(~/.claude/**)` | hook | hook: a read tool of a path there is denied, as is a redirection from it; neither for `cat ~/.claude/...` run as a command, and for a write there by a shell program (the driver checks only the repository) |
| `Edit(~/.claude/**)` | hook | hook |
| `Write(~/.claude/**)` | hook | hook |
| `WebFetch` | hook | hook denies every tool not listed, `web_fetch` among them; the adapter's `url` deny stands in |
| `WebSearch` | hook | hook denies every tool not listed, `web_fetch` among them; the adapter's `url` deny stands in |
| `Bash(git -C:*)` | hook, driver | hook matches it; the driver halts the run (exit 9) if `.git/config` or a hook changed (`plan changed …`) |
| `Bash(git -c:*)` | hook, driver | hook matches it; the driver halts the run (exit 9) if `.git/config` or a hook changed (`plan changed …`) |
| `Bash(git --git-dir:*)` | hook, driver | hook matches it; the driver halts the run (exit 9) if `.git/config` or a hook changed (`plan changed …`) |
| `Bash(git --work-tree:*)` | hook, driver | hook matches it; the driver halts the run (exit 9) if `.git/config` or a hook changed (`plan changed …`) |
| `Bash(git --exec-path:*)` | hook, driver | hook matches it; the driver halts the run (exit 9) if `.git/config` or a hook changed (`plan changed …`) |
| `Bash(git merge:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git cherry-pick:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git revert:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git am:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git pull:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git fetch:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git notes:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git replace:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git filter-branch:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git worktree:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git reflog:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git gc:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git config:*)` | hook, driver | hook matches it; the driver halts the run (exit 9) if `.git/config` or a hook changed (`plan changed …`) |
| `Bash(git read-tree:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git submodule:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git prune:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git repack:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git hash-object:*)` | hook | hook matches it, by the fence's own `*` matching; the driver does not revert the file it writes (see what is neither, below) |
| `Bash(git fast-import:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved (`REFS MOVED plan`) |
| `Bash(git format-patch:*)` | hook | hook matches it, by the fence's own `*` matching; the driver does not revert the file it writes (see what is neither, below) |
| `Bash(git diff*--output*)` | hook | hook matches it, by the fence's own `*` matching; the driver does not revert the file it writes (see what is neither, below) |
| `Bash(git log*--output*)` | hook | hook matches it, by the fence's own `*` matching; the driver does not revert the file it writes (see what is neither, below) |
| `Bash(git show*--output*)` | hook | hook matches it, by the fence's own `*` matching; the driver does not revert the file it writes (see what is neither, below) |
| `Bash(find*-delete*)` | hook | hook |
| `Bash(find*-exec*)` | hook | hook |
| `Bash(find*-ok*)` | hook | hook |
| `Bash(find*-fprint*)` | hook | hook |
| `Bash(find*-fls*)` | hook | hook |
| `Bash(rm -r:*)` | hook | hook matches it (the adapter's `rm` stem denies every `rm`); the driver does not revert files after a plan session |
| `Bash(rm -R:*)` | hook | hook matches it (the adapter's `rm` stem denies every `rm`); the driver does not revert files after a plan session |
| `Bash(rm -fr:*)` | hook | hook matches it (the adapter's `rm` stem denies every `rm`); the driver does not revert files after a plan session |
| `Bash(rm -fR:*)` | hook | hook matches it (the adapter's `rm` stem denies every `rm`); the driver does not revert files after a plan session |
| `Bash(rm --recursive:*)` | hook | hook matches it (the adapter's `rm` stem denies every `rm`); the driver does not revert files after a plan session |
| `Bash(vloop -C:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Bash(vloop --chdir:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Bash(vloop brief new:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |
| `Bash(vloop brief review:*)` | hook | hook matches it; the driver does not restore the plan's gates, defects, interventions or config after a plan session, so the hook is the line |

## The plan session's allowed writes

The plan fence allows three writes; the hook is the only line, and the judge
checks the path after expanding it.

| Path | Enforced by | Note |
| --- | --- | --- |
| `.vloop/state/state.json` | hook | hook allows exactly this file; the driver does not revert any other file under `.vloop/state/` |
| `.vloop/state/gates/` | hook | hook allows a file under it (not the folder itself); the gate fixtures are the plan session's to write |
| `.vloop/tmp/` | hook | hook allows a file under it (not the folder itself), where the session keeps its scratch files |

## The plan adapter's deny list

`copilotDenials` goes to Copilot as `--deny-tool` on the argv for a plan session
as for a review session, `shell(rm:*)` included, and the judge reads the
`shell(…)` stems as further Bash deny rules. Copilot's own matching is not
vloop's: it is the stem alone or followed by a space, so another form of the
command can pass it. The hook is the line for each; the driver halts the run only
on a moved ref or a changed `.git/config` or hook.

| Pattern | Enforced by | Note |
| --- | --- | --- |
| `shell(gh:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git push:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git remote:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git config --global:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git config --system:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git config --file:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(curl:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(wget:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(ssh:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(scp:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(nc:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(rm:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git commit:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git reset:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git checkout:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git switch:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git stash:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git branch:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git tag:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `url` | hook | Copilot's own URL tool is refused on the argv; the hook denies `web_fetch` and every tool it does not list |

## What is neither in a plan session, and what covers it

- **A write by a shell program other than a redirection** (`tee`, `cp`, `mv`,
  `sed -i`, `python -c`): the judge reads redirections and file tools, not what a
  program does. Neither: the driver does not revert files after a plan session,
  and the plan commit stages the whole tree. What covers it is the hook's other
  lines (the `rm`, `git` and wrapper denials), the plan's own review by the gate
  review and the operator before any work starts, and the driver's halt on a moved
  ref or a changed `.git/config` or hook.
- **A read of `~/.claude` by a program** (`cat ~/.claude/…`): neither. The hook
  denies a read tool and a redirection from there, not a program's argument;
  nothing covers it beyond the session's `COPILOT_HOME` being its own empty folder.
- **A command no list names**: both lists are denylists; the hook denies every
  tool it does not list, and the driver's git checks are the boundary.
- **Network from an allowed program**: `curl`, `wget`, `ssh`, `scp`, `nc` are
  denied; nothing stops a program from opening its own connection. Neither; the
  session timeout bounds how long one can run.

# Fence audit: who enforces each rule in a Copilot review-brief session

*Since B34 the driver re-judges every call a Copilot session ran and halts (exit 9, FENCE BYPASSED) on one the fence denies, and (HOOK CHANGED) when the hook's binary changed: where this audit says a call the hook fails open on goes unnoticed, that predates it.*

*Written for B29. It covers `fence/review-brief.json`, the fence a
`vloop brief review` session runs under, and the Copilot adapter's deny list for
it: the whole review list, `shell(rm:*)` included.*

A review-brief session is started by `vloop brief review <brief>`, not by
`vloop run`. With `agent.review-brief = copilot` it runs under the same
`preToolUse` hook, `vloop fence hook --phase review-brief`. The judge reads
`fence/review-brief.json` and every `shell(…)` stem of the adapter's deny list.
The driver's part is small:

- **Writes.** The hook allows a file tool or a redirection to write only under
  `.vloop/tmp/` (the report `.vloop/tmp/brief-review.json` and a copy of the
  brief for the structure check); the folder itself, a write elsewhere in the
  repository and a write outside it are denied, and `/dev/null` is allowed.
- **What the driver checks after the session.** The brief: a session that
  changed it is refused and the brief is restored (exit 9). Git: it halts with
  exit 9 if a ref moved or if `.git/config` or a git hook changed. It does not
  revert any other file the session wrote outside `.vloop/tmp/`, so the hook is
  the line for every file write; the rows below say *hook*, not *driver*, for
  them.
- **`rm`.** Every `rm` is denied, `shell(rm:*)` among the stems.
- **A nested review.** `Bash(vloop brief review:*)` is denied, so a review
  session cannot start another.
- **The skill.** The `skill` tool is allowed only for the names of vloop's
  skills (`/vloop:review-brief` here).

## The review-brief fence's deny rules

Every deny rule of `fence/review-brief.json`. The git, `rm`, `find`, `~/.claude`
and web rules are the review fence's, enforced by the hook, and by the driver's
guard where it names a ref or the git configuration. The `vloop` rules are
enforced by the hook alone: the driver restores neither the plan, the defects,
the interventions nor the configuration after this session.

| Rule | Enforced by | Note |
| --- | --- | --- |
| `Bash(git push:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git commit:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git reset:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git clean:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git branch:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git checkout:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git switch:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git remote add:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git remote set-url:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git remote remove:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git remote rm:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git remote rename:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git remote set-head:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git remote set-branches:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git remote prune:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git remote update:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git update-ref:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git symbolic-ref:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git tag:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git stash:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git rebase:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(rm -rf:*)` | hook | hook matches it (the adapter's `rm` stem denies every `rm`) |
| `Bash(sudo:*)` | hook | hook |
| `Bash(curl:*)` | hook | hook |
| `Bash(wget:*)` | hook | hook |
| `Bash(pip install:*)` | hook | hook |
| `Bash(claude:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Bash(vloop run:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Bash(vloop init:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Bash(vloop upgrade:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Bash(vloop brief close:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Bash(vloop task verify:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Bash(vloop task reset:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Bash(vloop task note:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Bash(vloop task drop:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Bash(vloop task set:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Bash(vloop defect add:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Bash(vloop defect set:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Bash(vloop intervention add:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Bash(vloop intervention set:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Bash(vloop config set:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Bash(.loop/run.sh:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Edit(.git/**)` | hook, driver | hook: only `.vloop/tmp/` may be written; the driver halts on a changed `.git/config` or hook |
| `Write(.git/**)` | hook, driver | hook: only `.vloop/tmp/` may be written; the driver halts on a changed `.git/config` or hook |
| `Edit(.vloop/state/**)` | hook | hook: only `.vloop/tmp/` may be written; the driver does not revert files after the session |
| `Write(.vloop/state/**)` | hook | hook: only `.vloop/tmp/` may be written; the driver does not revert files after the session |
| `Edit(.vloop/config.toml)` | hook | hook: only `.vloop/tmp/` may be written; the driver does not revert files after the session |
| `Write(.vloop/config.toml)` | hook | hook: only `.vloop/tmp/` may be written; the driver does not revert files after the session |
| `Edit(.vloop/defects/**)` | hook | hook: only `.vloop/tmp/` may be written; the driver does not revert files after the session |
| `Write(.vloop/defects/**)` | hook | hook: only `.vloop/tmp/` may be written; the driver does not revert files after the session |
| `Edit(.vloop/interventions/**)` | hook | hook: only `.vloop/tmp/` may be written; the driver does not revert files after the session |
| `Write(.vloop/interventions/**)` | hook | hook: only `.vloop/tmp/` may be written; the driver does not revert files after the session |
| `Edit(.vloop/repo-guidelines.md)` | hook | hook: only `.vloop/tmp/` may be written; the driver does not revert files after the session |
| `Write(.vloop/repo-guidelines.md)` | hook | hook: only `.vloop/tmp/` may be written; the driver does not revert files after the session |
| `Edit(.vloop/install.json)` | hook | hook: only `.vloop/tmp/` may be written; the driver does not revert files after the session |
| `Write(.vloop/install.json)` | hook | hook: only `.vloop/tmp/` may be written; the driver does not revert files after the session |
| `Read(~/.claude/**)` | hook | hook |
| `Edit(~/.claude/**)` | hook | hook |
| `Write(~/.claude/**)` | hook | hook |
| `WebFetch` | hook | hook |
| `WebSearch` | hook | hook |
| `Bash(git -C:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git -c:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git --git-dir:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git --work-tree:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git --exec-path:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git merge:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git cherry-pick:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git revert:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git am:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git pull:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git fetch:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git notes:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git replace:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git filter-branch:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git worktree:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git reflog:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git gc:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git config:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git read-tree:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git submodule:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git prune:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git repack:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git hash-object:*)` | hook | hook matches it, by the fence's own `*` matching; the driver does not revert the file it writes (see below) |
| `Bash(git fast-import:*)` | hook, driver | hook matches the git command; the driver halts the run (exit 9) if a ref moved or `.git/config` or a hook changed |
| `Bash(git format-patch:*)` | hook | hook matches it, by the fence's own `*` matching; the driver does not revert the file it writes (see below) |
| `Bash(git diff*--output*)` | hook | hook matches it, by the fence's own `*` matching; the driver does not revert the file it writes (see below) |
| `Bash(git log*--output*)` | hook | hook matches it, by the fence's own `*` matching; the driver does not revert the file it writes (see below) |
| `Bash(git show*--output*)` | hook | hook matches it, by the fence's own `*` matching; the driver does not revert the file it writes (see below) |
| `Bash(find*-delete*)` | hook | hook |
| `Bash(find*-exec*)` | hook | hook |
| `Bash(find*-ok*)` | hook | hook |
| `Bash(find*-fprint*)` | hook | hook |
| `Bash(find*-fls*)` | hook | hook |
| `Bash(rm -r:*)` | hook | hook matches it (the adapter's `rm` stem denies every `rm`) |
| `Bash(rm -R:*)` | hook | hook matches it (the adapter's `rm` stem denies every `rm`) |
| `Bash(rm -fr:*)` | hook | hook matches it (the adapter's `rm` stem denies every `rm`) |
| `Bash(rm -fR:*)` | hook | hook matches it (the adapter's `rm` stem denies every `rm`) |
| `Bash(rm --recursive:*)` | hook | hook matches it (the adapter's `rm` stem denies every `rm`) |
| `Bash(vloop -C:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Bash(vloop --chdir:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Bash(vloop brief new:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |
| `Bash(vloop brief review:*)` | hook | hook matches it; the driver does not restore the plan, defects, interventions or config after a review-brief session, so the hook is the line |

## The review-brief session's allowed writes

| Path | Enforced by | Note |
| --- | --- | --- |
| `.vloop/tmp/` | hook | hook allows a file under it (not the folder itself): the report `.vloop/tmp/brief-review.json` and scratch copies; the driver clears the folder before the session |

## The review-brief adapter's deny list

`copilotDenials` goes to Copilot as `--deny-tool` on the argv, as for a review
session, `shell(rm:*)` included, and the judge reads the `shell(…)` stems as
further Bash deny rules. Copilot's own matching is not vloop's, so another form
of the command can pass it; the hook is the line for each.

| Pattern | Enforced by | Note |
| --- | --- | --- |
| `shell(gh:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git push:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git remote:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git config --global:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git config --system:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git config --file:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(curl:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(wget:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(ssh:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(scp:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(nc:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(rm:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git commit:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git reset:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git checkout:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git switch:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git stash:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git branch:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `shell(git tag:*)` | hook | the judge matches the stem through wrappers; Copilot refuses the plain form too |
| `url` | hook | Copilot's own URL tool is refused on the argv; the hook denies `web_fetch` and every tool it does not list |

## What is neither in a review-brief session, and what covers it

- **A write by a shell program other than a redirection** (`tee`, `cp`, `mv`,
  `sed -i`): neither. The judge reads redirections and file tools, not what a
  program does. What covers it is the hook's other lines, the driver's check
  that the brief is unchanged (it restores it) and its halt on a moved ref or a
  changed `.git/config` or hook.
- **A read of `~/.claude` by a program** (`cat ~/.claude/…`): neither; nothing
  covers it beyond the session's `COPILOT_HOME` being its own empty folder.
- **A command no list names**: both lists are denylists; the hook denies every
  tool it does not list.
- **Network from an allowed program**: neither; the session timeout bounds it.
