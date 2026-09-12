# Native goal activation for delivery

Apply only when `deliver-change` enters **HANDS-OFF**. Use the harness hosting
this conversation, not the separately configured external auditor. This skill
requests native goal-backed delivery; obey any additional authorization rule
imposed by that harness before creating a goal.

## Resolve limits

Read `~/.agents/sdlc/bin/sdlc-harness --help` once, resolving the executable to
an absolute path. Run its **read-only** `goal-config --project <project-root>`
operation. Use the returned integers; do not parse budget strings yourself.
An error must be resolved before starting goal-backed delivery. Never substitute
a default after invalid input.

Resolution uses command arguments, process environment, `.sdlc/project.yaml`,
`~/.agents/sdlc.yaml`, then schema defaults. `max_turns` applies only to Hermes
and Copilot; `max_token_budget` only to Codex. Neither changes the audit limit.
Values describe native continuation budgets, not a guaranteed currency ceiling.

## Set the objective

Use the approved work identifiers, specification paths, and the agreed handback
gate. Include TDD, applicable implementation audit and remediation, documentation,
and validation evidence. Name human-only validations that may remain pending
at handback; do not define success as operator closure or manufacture evidence.
Keep the objective bounded to the authorized work, not general improvement of
the repository.

Inspect any existing goal first. Reuse a matching active goal without resetting
its budget. Do not replace another task's goal, replenish an exhausted goal, or
change an active budget without explicit operator authority.

## Codex

When exposed, call `get_goal`, then `create_goal` with the objective and
`token_budget` equal to the resolved `max_token_budget`. Use the actual tool
schema supplied by the current runtime. Do not pass `max_turns`.

Read back the goal and its budget before declaring activation. If the tools are
unavailable, report that native goal activation is unavailable in this session;
the operator can enable goals and use `/goal`. Do not run a child `codex` process
or pretend that printing `/goal` invokes it. If a budget cannot be applied on
that surface, disclose it before proceeding.

At the agreed gate, update the native goal to complete only when its actual
objective is met. Follow native rules for blocked or cancelled goals; budget
exhaustion and missing evidence are not successful completion.
[Native goals](https://learn.chatgpt.com/use-cases/follow-goals).

## Claude Code

Use a native goal control only if the current session exposes it as a callable
tool or supported command. Otherwise ask the operator once to run
`/goal <bounded completion condition>` in this conversation. Do not invent a
`Skill(goal)` capability or launch `claude -p` as a substitute.

Neither configured SDLC goal limit applies. Claude implements `/goal` using a
Stop hook; its documented protection ends a turn after **eight consecutive hook
continuations**. Hooks may be disabled by the current mode or policy. Report
the actual availability and limit; do not claim indefinite continuation.
[Claude goals](https://code.claude.com/docs/en/goal),
[Stop-hook limits](https://code.claude.com/docs/en/hooks#stop-input).

## Hermes

Use an exposed native goal tool if present. Otherwise ask the operator to set
`goals.max_turns` in the Hermes configuration to the resolved `max_turns`, and
activate `/goal <objective>` in the current conversation. Verify `/goal status`
through an exposed control or operator output; SDLC configuration resolution
does not itself configure Hermes.

Do not edit Hermes databases or internal session state to activate a goal.
Do not automatically call `/goal resume` after exhaustion: that command resets
the native turn counter. Ignore `max_token_budget` for this harness.
[Hermes goals and limits](https://hermes-agent.nousresearch.com/docs/user-guide/features/goals).

## GitHub Copilot CLI

Use a same-session autopilot tool supplied by an installed extension if one is
exposed. Require confirmation of both activation and the effective continuation
limit. SDLC does not currently install such an extension.

Otherwise explain the native route: Copilot can be launched by the operator
with `--max-autopilot-continues=<resolved max_turns>`, then `/autopilot <objective>`
activates continuation. An existing session without a supported limit-setting
control needs operator action; do not spawn another delivery session yourself.
Autopilot does not authorize unrestricted permissions. Ignore `max_token_budget`.
[Copilot command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference).

## Report actual state

Declare the native mechanism, **ACTIVE** or **UNAVAILABLE**, and effective native
limit with its unit. Never announce activation from configuration output alone.
Native continuation does not change SDLC stopping conditions, operator authority,
audit budgets, or retained audit identities. On a transition to ATTENDED or an
operator cancellation, pause/clear the delivery goal through supported controls;
do not restart it without renewed authority. Do not reset limits merely because
the context compacted or the model emitted a progress report.
