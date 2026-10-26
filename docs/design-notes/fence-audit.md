# Fence audit: who enforces each rule in a Copilot review session

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
  a command word starting with `$`. File tools (`create`, `edit`, `apply_patch`)
  and a shell redirection may write only under `.vloop/tmp/`; reads may not touch
  `~/.claude`, `~/.copilot` or `~/.config/gh`; a `skill` call must name one of
  vloop's skills (K1); every other tool is denied. It fails closed: an error
  answers deny.
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
