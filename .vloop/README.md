# vloop

Written by vloop 0.15.0-rc.4. vloop owns this file: `vloop upgrade` rewrites it, so edits here are lost.

## What vloop is

vloop is a command-line tool and a plugin for Claude Code and GitHub Copilot CLI. It turns a brief into tasks and, for each task, has a work session do it, a gate check it and an independent review judge it, with one commit per task. It runs on macOS, Linux and Windows.

## How to use it

Install it with `go install github.com/mvelosop/vloop/cmd/vloop@latest`. Run `vloop doctor` to check the setup. Write a brief with `/vloop:create-brief` in the agent and run it with `vloop run <brief>`. Keep vloop current with `vloop self-update`, then `vloop upgrade` in each repository.

## What is in this folder

- `config.toml`: shared configuration.
- `config.local.toml`: this machine's configuration, ignored by Git.
- `install.json`: vloop's installation data.
- `repo-guidelines.md`: repository lessons.
- `README.md`: this generated description.
- `defects/`: recorded defects.
- `interventions/`: recorded interventions.
- `state/`: the plan, runs, metrics and brief reviews.
- `tmp/`: scratch files, ignored by Git.

## More

https://github.com/mvelosop/vloop
