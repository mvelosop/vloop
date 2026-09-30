# Command reference

This page is generated from vloop's command tree. Do not edit it: run
`go generate ./...` to regenerate it. A test fails when it is out of date.

## Global flags

These work on every command.

- `-C, --dir path`: act as if started in path
- `--json`: machine-readable output on stdout
- `--no-color`: disable colour (also: non-empty NO_COLOR)
- `-q, --quiet`: print less
- `-v, --verbose`: print more

## vloop brief

Check loop briefs

```
vloop brief
```

## vloop brief check

Check that briefs are fit to plan

```
vloop brief check <path>...
```

## vloop brief close

Record findings, snapshot the metrics, write the run record, mark the brief consumed and commit

```
vloop brief close <brief> [flags]
```

- `--abandon string`: close a brief whose plan did not complete, as abandoned, with the reason
- `--dry-run`: print what a close would record, write and commit, and write nothing
- `--finding stringArray`: a defect you found in the run, recorded as an operator defect (repeatable)
- `--no-findings`: state that you found nothing

## vloop brief list

List briefs in dependency order

```
vloop brief list
```

## vloop brief new

Write a draft brief from the template

```
vloop brief new <slug> [flags]
```

- `--dry-run`: print the path and write nothing

## vloop config

Read and write repo-local configuration

```
vloop config
```

## vloop config get

Print the resolved value of a key

```
vloop config get <key>
```

## vloop config list

Print every key with its value and source

```
vloop config list
```

## vloop config path

Print the config file path, relative to the repo root

```
vloop config path
```

## vloop config set

Write a key to .vloop/config.toml ('' removes it)

```
vloop config set <key> <value>
```

## vloop defect

Record, list and update defects the loop cannot see

```
vloop defect
```

## vloop defect add

Record a defect as .vloop/defects/D<stamp>-<slug>.md and print its path

```
vloop defect add "<summary>" [flags]
```

- `--blame file:line`: attribute to the brief that wrote file:line when --brief is absent
- `--brief string`: the loop brief that introduced it
- `--case string`: the failing test written first
- `--found-by string`: who caught it: gate, review, operator, user
- `--kind string`: kind: bug, spec-gap, gate, gate-gap, regression (default bug)
- `--origin string`: where it came from: brief, plan, work, env (default work)
- `--severity string`: severity: low, medium, high, critical (default medium)
- `--task string`: the task in that brief's plan

## vloop defect list

Print the recorded defects, or with --matrix the origin × catcher counts

```
vloop defect list [flags]
```

- `--brief string`: only defects of this loop brief
- `--matrix`: print origin × catcher counts, derived defects included

## vloop defect set

Set status, fixed-by, case, severity, origin, kind, task of a defect

```
vloop defect set <id> <field> <value>
```

## vloop metrics

Summarise what a brief cost and delivered, from its runs and commits

```
vloop metrics [<brief>…] [flags]
```

- `--by task`: break the summary down by task
- `--workspace file`: show every repository the workspace file lists, with a repo column

## vloop metrics classify

Show how paths are classified and by which glob

```
vloop metrics classify <path>…
```

## vloop metrics export

Print briefs, tasks and defects as JSON Lines (export/v1), with the repository's identity

```
vloop metrics export [<brief>…] [flags]
```

- `--workspace file`: export every repository the workspace file lists

## vloop metrics stacks

Print the built-in stack presets

```
vloop metrics stacks [name]
```

## vloop schema

List, print and validate against the embedded JSON Schemas

```
vloop schema
```

## vloop schema list

Print the schema names

```
vloop schema list
```

## vloop schema show

Print a schema document

```
vloop schema show <name>
```

## vloop schema validate

Validate a JSON file against a schema

```
vloop schema validate <name> <file>
```

## vloop status

Show the plan's progress

```
vloop status [flags]
```

- `--markdown`: print the plan as markdown

## vloop task

Inspect the plan's tasks

```
vloop task
```

## vloop task drop

Remove a task nothing depends on

```
vloop task drop <id>
```

## vloop task gate

Run a task's verify command in the plan's shell

```
vloop task gate <id>
```

## vloop task list

Print one line per task

```
vloop task list
```

## vloop task note

Replace a task's notes

```
vloop task note <id> <text>
```

## vloop task reset

Set a task back to pending with no attempts

```
vloop task reset <id>
```

## vloop task set

Set (or with '' clear) area, kind, model.work, model.review, effort.work, effort.review

```
vloop task set <id> <field> <value>
```

## vloop task show

Print a task with the model and effort its sessions resolve to

```
vloop task show <id>
```

## vloop task validate

Check the plan's structure

```
vloop task validate
```

## vloop task verify

Replace a task's verify command, recording why

```
vloop task verify <id> <command> [flags]
```

- `--reason string`: why the gate is being replaced (required)

## vloop version

Print the vloop and embedded plugin versions

```
vloop version
```

