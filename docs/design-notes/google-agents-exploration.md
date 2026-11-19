# Google agents (Gemini CLI, Antigravity CLI) — exploration

Status: exploration, not a brief. Written 2026-11-19 from an interactive
session, from the vendors' documentation read on 2026-11-17. Nothing here was
run: neither CLI was installed. Every claim below is what the docs say, and the
probes at the end are what has to be measured before a brief relies on it. The
architecture this fits into is `docs/design-notes/multi-agent-backends.md`; the
adapter it would sit beside is `internal/driver/agent_copilot.go`, whose
findings are `docs/design-notes/copilot-cli-experiments.md`.

## Short answer

- **Gemini models already have a route with no code:** the Copilot adapter.
  Copilot's model list of 2026-11-02 shows `gemini-3.8-flash` enabled and
  `gemini-4-argon` disabled by the account's policy
  (`docs/design-notes/copilot-cli-experiments.md`, "Which model ids Copilot
  takes"). `agent.<phase> = "copilot"` with `model.<phase> =
  "gemini-3.8-flash"` reuses the fence hook, the credit estimate and the plugin
  loading. Not yet tried.
- **A native adapter should target Antigravity CLI (`agy`), not Gemini CLI.**
  Google is moving its terminal agent from Gemini CLI to Antigravity CLI.
  `agy` also fits vloop better: plugin skills are namespaced and slash commands
  expand in print mode, `--effort` takes vloop's values, and its permission
  rules look like the Claude fence's.
- **The blocker for `agy` is isolation.** No flag or environment variable gives
  a session its own config directory; everything lives under `~/.gemini/`.
  Until a probe finds a way, a native adapter would have to write the plugin,
  hooks and permissions into the operator's global config. That breaks this
  repository's first rule, and the operator's own global rules could weaken
  the fence.

## What an adapter has to provide

The `Adapter` interface (`internal/driver/agent.go`) asks for: the phases
supported, a preflight, a per-session `Prepare` (plugin, fence, home,
guidelines), the command (argv, working directory, environment), and `Parse`
(stdout to a session record: cost, turns, duration, error, denials, models
used). The fence is enforced three ways for Copilot: a deny list on the command
line, a `preToolUse` hook running `vloop fence hook`
(`internal/cli/fence.go`), and the driver re-judging the session's calls
afterwards (`internal/driver/fencecheck.go`). The sections below measure each
CLI against those needs.

## Gemini CLI

Docs read: geminicli.com/docs and github.com/google-gemini/gemini-cli
(release v0.68.0, 2026-11-10).

### Status of the product

The docs carry a banner: Gemini CLI "was replaced by Antigravity CLI on June
18th, 2026" for unpaid-tier and Google One users, and it "has not been
optimized for Gemini 4 Argon". Google's announcement (see Antigravity below)
narrows it: it still serves paid Gemini API keys and Gemini Code Assist
Standard and Enterprise. No end-of-life date for those was found.

### Mapping

| Adapter need | Gemini CLI | Assessment |
|---|---|---|
| Headless | `-p` / `--prompt`; a non-TTY also runs headless | fine |
| Model | `-m` / `--model`, with aliases `auto`, `pro`, `flash`, `flash-lite` and full ids | fine |
| Effort | no flag; `thinkingConfig` (`thinkingBudget`, `thinkingLevel`) under `modelConfigs` in settings | written into a generated settings file |
| Plugin | Agent Skills (`SKILL.md`) load from extensions (`gemini-extension.json` + `skills/`), user (`~/.gemini/skills`, `~/.agents/skills`) and workspace (`.gemini/skills`, `.agents/skills`). It does not read `.claude-plugin/plugin.json`. | `Prepare` would generate an extension from the embedded plugin |
| Prompt | Slash commands from `-p` are not documented. Skills run when the model calls `activate_skill`, which asks for consent. Extension commands are `/cmd` or `/dir:cmd`, and become `/<ext>.<cmd>` on a clash. | **the riskiest point**: a plain-text prompt naming the skill, its reliability unknown |
| Skill clashes | workspace skills outrank extension skills | a repository's own `review` skill would win over vloop's |
| Isolation | `GEMINI_CLI_HOME` moves the user layer; `GEMINI_CLI_SYSTEM_SETTINGS_PATH`; trust via `GEMINI_CLI_TRUST_WORKSPACE=true` or `--skip-trust`. No way to ignore the project's `.gemini/settings.json` was found. | per-session home works; the project layer still applies |
| Permissions | `--approval-mode default\|auto_edit\|yolo\|plan`. Policy engine: TOML `[[rule]]` with `toolName`, `commandPrefix`, `commandRegex`, `decision = "deny"`, `priority`. In headless mode `ask_user` counts as deny. User policies live in `~/.gemini/policies/`; admin ones come from `--admin-policy` and rank highest. Workspace policies are disabled (issue #18186). `--allowed-tools` and `tools.exclude` are deprecated. | a deny list like Copilot's, from `--admin-policy` |
| Hook | `BeforeTool` in `settings.json` under `hooks`. Stdin carries `tool_name` and `tool_input`. It denies by exit 2 (stderr is the reason) or by stdout `{"decision":"deny","reason":…}`. Timeout in ms, default 60000. | `parseHookPayload` needs a second payload shape |
| Hook failure | **fails open**: another non-zero exit warns and the tool runs; stdout that is not JSON counts as allow. Behaviour on timeout not documented. | **the hook alone cannot be the fence**; the policy deny rules and the driver's re-judge have to back it |
| Tool names | `run_shell_command`, `read_file`, `read_many_files`, `write_file`, `replace`, `glob`, `grep_search` (alias `search_file_content`), `list_directory`, `activate_skill`, `web_fetch`, `google_web_search`, `ask_user`, `write_todos`, … | the judge's switch (`internal/fence/judge.go`) gains these names |
| Web, MCP | deny `web_fetch` and `google_web_search` by policy; `--allowed-mcp-server-names`, or a policy `mcpName = "*"` deny | fine |
| Guidelines | `GEMINI.md` (global and workspace); `context.fileName` can add `AGENTS.md`. `GEMINI_SYSTEM_MD` replaces the whole system prompt. No append flag. | a `GEMINI.md` in the session's home |
| Output | `--output-format json` prints one document (`response`, `stats`, `error`); `stream-json` prints JSON Lines (`init`, `message`, `tool_use`, `tool_result`, `error`, `result` with per-model tokens). Exact stats field names are not documented. Exit codes: 0, 1, 42 (input), 53 (turn limit). | stream-json, for the re-judge; field names by probe |
| Cost | tokens only | a price config, or no cost budget |
| Auth | `GEMINI_API_KEY`, or Vertex (`GOOGLE_CLOUD_PROJECT`, `GOOGLE_CLOUD_LOCATION`, ADC or a service account) | fine |
| Quota | requests per day: free API key 250 (Flash only); Google login 1,000; AI Pro 1,500; Ultra 2,000; pay-as-you-go varies | a run of dozens of sessions can exceed the free tiers |

### Verdict

Feasible, with two unknowns that could each sink it: whether a skill runs
reliably from a plain-text headless prompt, and a hook that fails open. Its
future is limited to API-key and enterprise users, so the effort is better
spent on Antigravity.

## Antigravity CLI (`agy`)

Docs read: antigravity.google/docs, the announcement on
developers.googleblog.com ("An important update: transitioning Gemini CLI to
Antigravity CLI", 2026-05-19), and github.com/google-antigravity/antigravity-cli
(CHANGELOG top entry 1.5.10; the docs' changelog stops at 1.5.9).

### Status of the product

- A new terminal agent written in Go, on the same agent harness as the
  Antigravity app and IDE. It keeps Agent Skills, hooks, subagents and
  extensions (now called plugins), with "no 1:1 feature parity right out of
  the gate".
- Installed by a script (`curl -fsSL https://antigravity.google/cli/install.sh
  | bash`; PowerShell and CMD variants for Windows) into `~/.local/bin/agy`. No
  npm or Homebrew package.
- Migration from Gemini CLI: `agy plugin import gemini` converts extensions to
  plugins. `GEMINI.md`, `AGENTS.md` and `~/.gemini/skills/` still work.
  Workspace skills move from `.gemini/skills/` to `.agents/skills/`.
- Young and moving: ~1,000 open issues, releases every few days, docs a
  release behind.

### Mapping

| Adapter need | Antigravity CLI | Assessment |
|---|---|---|
| Headless | `-p` / `--print` / `--prompt`; `--print-timeout` (default `0s`, no limit) | fine |
| Model | `--model <slug>` from `agy models`: `gemini-4-hydrogen-{xhigh,high,medium,low}`, `gemini-4-carbon-*`, `gemini-4-argon-*`, `gemini-3.8-flash-*`, `gemini-pro-orchestrator`, … The picker also offers Claude Sonnet 5.5 and Opus 5.5 (exact slugs not confirmed). | fine |
| Effort | `--effort low\|medium\|high\|xhigh\|max`, per model; not for the Orchestrator | **maps one to one** onto `effort.<phase>` |
| Plugin | Plugins have `plugin.json` at the plugin root, in `.antigravity-plugin/` or in `.gemini-plugin/`, installed with `agy plugin install`. Changelog 1.5.6: `agy plugin install` honours a Claude Code plugin's `skills` field. Skills come from `.agents/skills/` (workspace and parents), `~/.gemini/config/skills/`, `~/.gemini/config/plugins/<plugin>/skills/`, and, for the CLI, `~/.gemini/antigravity-cli/skills/` and `~/.gemini/skills/`. | vloop's plugin may install unchanged; by probe |
| Prompt | Slash commands expand in print mode (unless `--disable-slash-commands`). Plugin skills are `<plugin>:<skill>`, and each skill is also a slash command. | **`/vloop:review T3` may work as today's prompt** |
| Isolation | **No documented flag or environment variable for another home or settings directory.** Paths are fixed under `~/.gemini/`. Leads: a `--settings` flag that only the 1.5.8 changelog mentions; `--add-dir`, whose folder's customizations are loaded; `HOME` pointed at a temporary directory for the process. Project `.agents/`, `AGENTS.md` and `GEMINI.md` always load. | **the blocker** |
| Permissions | `--mode accept-edits\|auto\|plan` (default `default`; `auto` needs the sandbox). `--dangerously-skip-permissions`. Settings `toolPermission` and `permissions.allow\|ask\|deny`, rules `action(target)` with actions `read_file`, `write_file`, `read_url`, `execute_url`, `command`, `mcp`. Command rules match by word prefix (`command(git push)`) or `regex:`. Deny beats ask beats allow. A line using `$(…)` loses prefix matching. | **close to the Claude fence**: the deny lists in `fence/*.json` translate mechanically |
| Unapproved tool, headless | soft-denied: the run goes on, exits 0, and prints a notice on stderr | denials have to be read from stderr or the stream |
| Hook | `PreToolUse` in `~/.gemini/config/hooks.json`, `<workspace>/.agents/hooks.json` or a plugin's `hooks.json` (not in `settings.json`). Stdin is camelCase: `toolCall.name`, `toolCall.args`, `toolCall.id`, `conversationId`, `workspacePaths`, `cwd`, `hookEventName`. It denies by stdout `{"decision":"deny","reason":…}`; other decisions include `allow`, `ask`, `force_ask`. Timeout in seconds, default 30. | a payload close to Copilot's |
| Hook failure | **not documented.** Changelog 1.5.8: a hook that failed to launch on Windows was "silently skipped", which means it failed open there. Exit-code semantics not documented. | needs the B22.2-style probe |
| Tool names | `run_command`, `view_file`, `write_to_file`, `replace_file_content`, `multi_replace_file_content`, `list_dir`, `find_by_name`, `grep_search`, `search_web`, `read_url_content`, `invoke_subagent`, `define_subagent`, `ask_question`, `run_workflow`, … | the judge gains these names |
| Web, MCP | deny `read_url(*)` and `mcp(*)`; `search_web` has no permission action, so a hook deny by tool name | the hook carries `search_web` |
| Guidelines | `AGENTS.md` and `GEMINI.md` (workspace and parents; global in `~/.gemini/`); `rules/*.md` with a `trigger`. No system-prompt flag found. | a global `AGENTS.md` in the session's home, if isolation is solved |
| Output | `--output-format json` gives one result: `conversation_id`, `status` (`SUCCESS`, `ERROR`, `CANCELED`, …), `response`, `error`, `duration_seconds`, `num_turns`, `usage` (`input_tokens`, `output_tokens`, `thinking_tokens`, `cache_read_tokens`, `total_tokens`). `stream-json` prints NDJSON `init`, `step_update` (with `step_type` `tool` and `tool_info`), `result`. Exit 0 on success, 1 on an unknown model. | `Parse` is close to Claude's; stream-json for the re-judge |
| Cost | tokens only in headless output; `/usage` and `/credits` in the interactive UI | a price config, or no cost budget |
| Auth | a cached login (run `agy` once); or `GEMINI_API_KEY` with `"modelProvider": "gemini"` in `~/.gemini/antigravity-cli/settings.json`; or ADC with `AGY_ADC_AUTH=true` | the API-key route needs the settings file, so it depends on isolation too |
| Quota | Subscriptions have 5-hour and weekly limits. Per-minute waits of up to 4 minutes are retried; spend-cap, prepaid-balance and weekly-quota errors stop the turn. `useG1Credits` falls back to AI credits. | a weekly limit halts a run mid-brief; a paid API key avoids it |
| Platforms | macOS, Linux, Windows (x64, ARM64). The sandbox, and so `auto` mode, is missing on Windows 10, older Windows 11 and some Linux containers. | fine with explicit rules, not `auto` |

### Verdict

`agy` is the better target. If its plugin installs and `/vloop:<phase>` runs in
print mode, the prompt, plugin and effort carry over as they did for Copilot,
and its permission rules translate the Claude fence more faithfully than
Copilot's deny list. It rests on one thing the docs do not give, per-session
isolation, and on one thing they do not say, what a failing hook does.

## Probes before any brief

As for Copilot, an experiment repository outside this one, replaying sessions
of a reference run (`docs/design-notes/copilot-cli-experiments.md`, "Method").
For `agy`, in this order, each one able to stop the work:

1. **Isolation.** Does `HOME=<tmp>`, `--settings <file>` or `--add-dir <dir>`
   load a plugin, `hooks.json`, permission rules and auth settings from a
   directory under `.vloop/tmp/`, with nothing written to or read from the
   real `~/.gemini/`? Do the operator's global allow rules and hooks still
   apply?
2. **Plugin and prompt.** Does `agy plugin install` take vloop's plugin as it
   is (`.claude-plugin/plugin.json`), or does it need a root `plugin.json`?
   Does `agy -p "/vloop:review T3" --output-format json` load the skill and
   leave a `verdict.json` that validates? Does a project skill with the same
   name win?
3. **Fence.** Do `permissions.deny` rules (`command(git push)`,
   `write_file(.vloop/state/**)`, `read_url(*)`) still hold under
   `--dangerously-skip-permissions`, or under explicit `allow` rules? The hook
   failure cases of B22.2: missing binary, exit 1, exit 2, stdout that is not
   JSON, timeout. Does a deny show in `stream-json`?
4. **Record.** The `json` and `stream-json` fields against the session
   record; how a soft-denied call is reported; exit codes on a model error and
   on a quota stop.
5. **Quota.** What a weekly-limit or spend-cap stop looks like to the driver,
   so it halts as a quota halt and not as a failed task.

For the Copilot route, the existing eval runner is enough: `tools/agent-eval`
with `agent = copilot`, `model = gemini-3.8-flash`, on the gate-review and
review cases, as B33 did for Terra.

## Suggested sequence

1. Qualify Gemini through Copilot as a second reviewer. No code; a design-note
   section like B33's records the result.
2. Run the `agy` probes above; record them as `agy-cli-experiments.md`, the
   counterpart of the Copilot findings.
3. If they pass: an `agy` adapter for review sessions (as B19), then the fence
   (as B20/B22), then work, verify, gate-review, plan and review-brief (as
   B24–B29).

Gemini CLI is not on the path unless the operator only has an API key or
enterprise access and `agy` fails its probes.

## Sources

Gemini CLI: geminicli.com/docs (cli/headless, cli/cli-reference,
cli/custom-commands, cli/skills, cli/gemini-md, cli/system-prompt,
reference/policy-engine, reference/configuration, reference/tools, hooks,
hooks/reference, extensions/reference, get-started/authentication,
resources/quota-and-pricing); github.com/google-gemini/gemini-cli/releases.

Antigravity CLI: antigravity.google/docs (getting-started, install, cli/reference,
cli/headless, cli/gcli-migration, cli/credits, hooks, settings, permissions,
models, skills, customization-discovery, changelog);
antigravity.google/product/antigravity-cli;
developers.googleblog.com/an-important-update-transitioning-gemini-cli-to-antigravity-cli;
github.com/google-antigravity/antigravity-cli and its CHANGELOG.md. Not read:
the plans page (prices, quotas), sandbox, plugins, MCP, rules and `agy acp`.
