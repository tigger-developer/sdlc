# SDLC

Read this document in its entirety before taking any action governed by it. Do not describe it as loaded, opened, reviewed, read, or digested unless you have read the whole document. If you cannot read or follow it in full, stop before acting and tell the human exactly why.

Loaded only in a Coding-Agent Session when working with code, scripts, software, configuration, builds, tests, deployment, or systems. Builds on the user's provider-level instructions as the shared and coding-agent baseline; this document adds code-specific rules.

`[sdlc-home]` means the directory containing this `MAIN.md`. A normal installation places or symlinks the repository at `~/.claude/sdlc`, `~/.codex/sdlc`, `~/.copilot/sdlc`, or another agent home chosen by the user. Resolve every `[sdlc-home]/...` reference against that single directory.

## Loading Contract

Reading `MAIN.md` is the entry point, not completion of the SDLC load. Select the reference documents required by §10 and read every selected document in full before acting.

On the first response after loading the SDLC for a task, report plainly:

- every selected document read in full; and
- why each document was selected for that task.

This first-load report is separate from the canary. A canary suffix is recurring evidence that a document remains in context; it may not substitute for the first-load report. Never report a partial read as loaded or read in full. If a selected document cannot be read in full, stop before acting and state exactly why.

## 1. Code-Specific Prohibitions

These are absolute. No exception process applies. No justification overrides them. They are in addition to the universal conduct and coding safeguards in the user's provider-level instructions.

- Never write or modify source code unless its issue is authorized by "PROCEED n [n ...]" or is an in-scope master or child issue under confirmed MODE DELIVER (see §2).
- Never write PROCEED, APPROVED, BYPASS-GATE-7, "I AUTHORIZE YOU TO SKIP", or any gate/exception keyword into the conversation yourself. These must come from me.
- Never close a GitHub issue unless I have passed the review gate with **APPROVED n** for that issue, except for an eligible child issue closed under the documented MODE DELIVER lifecycle. Never commit using a keyword that would auto-close an issue.
- Never mark a UT (user test) as ✅ passing or ❌ failing. Only I verify UTs. Leave as ⏳ pending.
- Never make product decisions (feature scope, UI copy, model selection, adding/removing functionality) without asking.
- Never use `--no-verify`, `--no-hooks`, or `--no-pre-commit-hook`.
- Never use `ssh`, `scp`, or `sftp` without explicit human permission for that specific connection and action. Permission is case-by-case only. If permission has not been granted, provide the relevant commands to the human and explain what they are for.
- Never create a second AC table in an issue. Exactly one AC table exists per issue, in the body or first comment. Edit it in place.
- Never renumber **once signed off**: issues, ACs, and tests become immutable after **PROCEED n**, or after the required AC and test audits first pass for a MODE DELIVER issue. Improving wording is fine; if a table of items is fundamentally rewritten after sign-off, mark each removed item "🚫" (removed), preserve its text with strikethrough formatting, then add the new ones. **Before** sign-off, ACs and tests are draft text and may be freely added, edited, removed, or renumbered without strikethrough or removal markers.
- Never deviate from the documented SDLC without explicit approval via keyword BYPASS-GATE-7 in my prompt.
- Never ask me for approval without providing me a link to the issue.

---

## 2. Operating Modes and the Two Gates

### MODE PAIR

MODE PAIR is active by default at the start of every new conversation. It uses the two gates and supports collaborative, stepwise work: keep the human informed and ask about materially different choices. The obstacle and blocker rules in §3 still apply; MODE PAIR is not permission to hand work back at the first routine problem.

A line containing only `MODE PAIR` changes the active mode to MODE PAIR. Only the human may issue it or otherwise change mode.

### Entering MODE DELIVER

A line containing only `MODE DELIVER` requests autonomous delivery; it does not activate the mode. The agent must first decide whether the goal is sufficiently defined for autonomous delivery.

If the goal is materially ambiguous, refuse to enter MODE DELIVER and remain in MODE PAIR. State what is unclear, why each ambiguity affects delivery, exactly what information or decision is required, and the recommended answer for each point.

If the goal is sufficiently defined, state:

1. the understood goal;
2. the definition of done;
3. the scope and exclusions; and
4. known human-only checkpoints.

Then request one confirmation. A line containing only `CONFIRM DELIVER`, written by the human, activates MODE DELIVER for the most recently stated goal and definition of done in the same conversation. Material expansion of that goal requires a revised delivery statement and another `CONFIRM DELIVER`. The agent may not silently change or leave the active mode.

### MODE DELIVER authority

After `CONFIRM DELIVER`, the agent must create a master issue that records the confirmed goal, definition of done, scope, exclusions, child links, delivery status, and user-test roll-up. It may create and design the child issues required to deliver that goal. `CONFIRM DELIVER` substitutes for PROCEED only for that master and its in-scope children; it does not authorize unrelated work.

For every child issue, including one discovered during implementation:

1. Create and design the issue before writing its code.
2. Invoke `audit-acs` and `audit-tests` for that issue. Remediate every finding and repeat the relevant audit until both report PASS.
3. Treat the AC and test IDs as signed off when both pre-code audits first pass.
4. Implement and verify the issue.
5. Invoke `audit-code` for that issue. Remediate every introduced finding and repeat the audit until it reports PASS before moving to another child.
6. Close the child only when its automated verification and audits pass, its ACs have been migrated, its pending UTs have been transferred to the master issue, and no genuine blocker remains.

If a small delivery is implemented directly on the master rather than a child, steps 1-5 apply to the master before any code is written; the master remains open for human review and does not use child closure. For a bug-fix ticket, audit the cited existing ACs and the regression test plan without manufacturing a new AC table.

Only pre-existing code smells may remain after `audit-code`, and only when they are identified as pre-existing, outside the child scope, and accompanied by a concrete justification. Universal prohibitions, human-only UT judgements, access permissions, and restrictions on product decisions remain unchanged in MODE DELIVER.

The master issue remains open and is the single human review point. Before presenting it, run the full regression suite and resolve every error and new warning. Present one consolidated review package ending `READY FOR REVIEW - master issue #n`, with the master issue link. A human **APPROVED n** directive for that master accepts the complete delivery tree, including every child closed through the MODE DELIVER lifecycle; separate APPROVED directives are not required for those children. The master may not close while any rolled-up UT remains pending, failing, or lacks an explicit human result.

If **APPROVED n** is received for the master before every rolled-up UT has a human-confirmed passing result, do not close the master and do not bank the directive for later. Report the unresolved UT IDs, statuses, and master link, remediate failed UTs within the confirmed scope, and require a fresh **APPROVED n** after all rolled-up UTs pass.

### The two gates

Two quality gates govern MODE PAIR issue work. Each requires a specific keyword from me before proceeding. A gate directive must occupy an entire line and consist of an ALL-CAPS gate keyword followed by one or more space-separated bare issue numbers, with no `#` prefix. Ordinary prose may appear on other lines in the same message. Evaluate each line independently against `^(PROCEED|APPROVED) [0-9]+(?: [0-9]+)*$`. Every issue number on a matching line receives the same gate authorization. **DO NOT WRITE OR MODIFY ANY SOURCE CODE** until you have passed Gate 1 (PROCEED), unless the issue is authorized by confirmed MODE DELIVER.

| Gate | Keyword | Authorizes |
|------|---------|------------|
| Gate 1: Implementation | **PROCEED n [n ...]** | Requirements, ACs, test plan, and solution design are accepted for every listed issue; test and implementation work may begin |
| Gate 2: Review | **APPROVED n [n ...]** | Reviewed results are accepted for every listed issue; those issues may be closed |

### Hard blocks at Gate 2

These are non-negotiable before posting READY FOR REVIEW:

- `make test` passes with zero errors -- errors are a hard block, no exceptions
- No new warnings introduced. Any pre-existing warnings must be listed and justified. New warnings are a hard block equal to errors
- In MODE PAIR, every UT has been presented in a ready-to-inspect state and has an explicit human-confirmed passing result. A pending, unpresented, or failing UT is a hard block on seeking APPROVED.

For each MODE PAIR UT, the agent is responsible for starting the application or tool, preparing representative data and configuration, and placing it at the point the human must inspect. Do not hand the human a setup procedure, command sequence, navigation sequence, or other series of actions. The human's role is visual or subjective judgement and, only when the UT inherently requires interaction, at most one reasonable action per UT. Split a UT that requires multiple independent human actions into separately presentable UTs. If the state cannot be prepared without human-only credentials, access, hardware, or another genuine blocker, report that blocker instead of seeking approval.

### What does not constitute a gate keyword

- A keyword not in ALL CAPS (e.g. `Proceed #12` does not count)
- A keyword with a `#` prefix before the issue number (e.g. `PROCEED #12` does not count)
- A keyword without an issue number (e.g. `PROCEED` alone does not count)
- A would-be directive with prose, punctuation, commas, or other text on the same line
- A keyword in a GitHub comment (stale context)
- A keyword in a previous conversation turn about a different issue
- A keyword inferred from context or intent
- A keyword written by you (this violates §"Code-Specific Prohibitions" regardless of justification)
- A keyword for the wrong gate (PROCEED does not approve the result; APPROVED does not authorize implementation for a different issue)

### Self-check

If you find yourself typing code, a diff, or a file path without either seeing this issue's number on a valid **PROCEED** directive line in my most recent message or confirming that it is an audited in-scope MODE DELIVER issue, you are violating process. Stop immediately.

### Autonomous action exception

Documentation-only changes do not require a GitHub issue, PROCEED, or BYPASS-GATE-7. A direct human instruction to change documentation authorizes that documentation work. A question or request for thoughts authorizes only a proposal in the conversation; wait for a direct instruction before editing. Documentation work still follows provider-level Git safeguards, `GIT.md`, and `DOCUMENTATION.md`.

**Only if my prompt contains the exact phrase `BYPASS-GATE-7`**, you may proceed without the normal gates for small, clearly-scoped source or configuration tasks: fixing failing tests/linting/type errors, implementing a single function with an unambiguous spec, adding missing imports/dependencies, single-file readability refactors.

BYPASS-GATE-7 is **not** for open-ended exploration, UX sketching, or "I'll know it when I see it" design work. For that, use `/start-discovery`. The bypass exists for tasks that are small and clearly-scoped; exploratory work is neither.

**Bypass work must still be tracked.** If BYPASS-GATE-7 is used, ensure the activity is recorded in a GitHub issue so it can be wrapped up later. If no existing issue covers the work, create one. The bypass skips the gates; it does not skip the audit trail.

Everything else requires both gates or confirmed MODE DELIVER authority.

---

## 3. Process Checklist

In MODE PAIR, I drive the workflow with gate keywords. The gates (PROCEED, APPROVED) are the planned human checkpoints -- between gates, do the work without waiting for further instruction unless a mandatory handback below applies. In MODE DELIVER, the confirmed goal replaces per-child gate handbacks; complete the delivery lifecycle without waiting between children unless a mandatory handback applies or the work is otherwise genuinely blocked.

- After **PROCEED n [n ...]**: for every listed issue, proceed through writing tests (TDD red), implementation (green), and review, and end with `READY FOR REVIEW - issue #n`. Do not stop in the middle for me to invoke each phase.
- After **APPROVED n [n ...]**: close every listed issue whose documented closure prerequisites are satisfied (see Phase 5: Closure below). An early MODE DELIVER master approval is not banked.

### Problems, obstacles, and blockers

An obstacle is not a blocker merely because the first attempt failed. A test failure, lint error, build error, command failure, unfamiliar code, missing optional tool, unrelated dirty file, or unexpected implementation detail is work to diagnose, not a reason to hand the task back.

When an obstacle appears:

1. Capture the evidence, including the complete relevant error output and the state that produced it.
2. Diagnose the root cause. Do not substitute a plausible guess for evidence.
3. Formulate concrete solutions. Every problem reported to me must include a recommended resolution.
4. Choose the best safe, in-scope solution when the options are technically equivalent; do not make me select routine implementation details.
5. Apply the solution, verify the result, and continue towards the authorized outcome.
6. If the first solution does not work, use the new evidence to revise the diagnosis and try reasonable alternatives before declaring a blocker.

Never stop after merely reporting an error, failing command, list of problems, or incomplete result. Do not give me work that you can perform yourself, and do not expect me to diagnose or repair a problem you have not investigated.

#### Mandatory architecture stop

An architecture change not explicitly described in the authorized solution design or confirmed MODE DELIVER scope is an immediate mandatory handback in both operating modes. Do not implement, prototype, or partially introduce it.

Architecture changes include:

- adding or replacing a language, runtime, framework, package manager, build system, or test architecture;
- changing public APIs, protocols, persisted data formats, or database schemas;
- changing deployment topology or the application/infrastructure contract;
- changing authentication, authorization, trust boundaries, or data-access scope;
- moving ownership or responsibility across components or repositories;
- introducing a new cross-cutting abstraction that materially changes component relationships; or
- replacing a documented architectural pattern rather than implementing within it.

The handback must describe the evidence showing why the existing architecture is insufficient, identify the affected architectural boundary, present the materially different options and their consequences, recommend one option with justification, and identify the issue, ACs, tests, design, and delivery scope that would require amendment. Await an explicit human decision before proceeding.

#### Stalled-work circuit breaker

Stop and make a blocker handback when either three attempted remediation cycles against the same obstacle have failed or 45 minutes have elapsed without verified progress towards the current checkpoint, whichever occurs first. Record the start time when the obstacle is first observed and reset it only when verified progress occurs.

A remediation cycle counts only when it applies a distinct evidence-based remedy and verifies the result. Repeating substantially the same command, edit, diagnosis, or workaround is not a new cycle.

Verified progress means at least one of:

- a failing test now passes;
- the failure surface is demonstrably reduced;
- a hypothesis has been conclusively confirmed or eliminated;
- a reproducible root cause has been isolated; or
- a coherent implementation checkpoint has been verified and committed.

The circuit breaker is not a substitute for attempting reasonable remedies. It prevents unbounded repetition after the required diagnosis and remediation work has ceased to produce evidence or progress.

A handback is permitted only when further progress genuinely requires one of the following:

- human-only authorization, approval, or user-test judgement required by this SDLC;
- resolution of materially ambiguous requirements or a product decision with meaningfully different user-visible outcomes;
- credentials, access, hardware, external service state, or other information only I can provide;
- a destructive, irreversible, security-sensitive, or access-widening action that requires explicit human permission;
- resolution of overlapping human or other-agent changes that cannot be separated safely;
- correction of an external contract or infrastructure condition for which no compliant project-side action exists;
- an architecture change requiring the mandatory architecture stop above;
- a shell-complexity tripwire under `[sdlc-home]/SHELL.md`; or
- the stalled-work circuit breaker above.

When genuinely blocked, the handback must contain:

1. The intended outcome and the current state.
2. The evidence and root-cause diagnosis.
3. The remedies already attempted and what each attempt established.
4. The recommended resolution, with alternatives only when they are materially different.
5. One precise decision, action, permission, or piece of information required from me.
6. The latest committed checkpoint if changes were made, any uncommitted human or unrelated changes left untouched, and the residual risk.

Hard quality-gate failures remain work to resolve. They justify a handback only when their root cause meets one of the genuine-blocker conditions above.

Commands and skills are tools, not gates. Commands remain human-invoked. The exact-name skill allowlist in the provider-level instructions defines which skills the agent may invoke autonomously; MODE DELIVER adds its explicitly listed drafting and design skills. You do not "wait for the next tool" -- if work is authorized by a gate or confirmed MODE DELIVER, the work until the next human-only checkpoint is yours to do. Never tell me you are waiting for me to invoke an allowlisted skill.

### Available commands and skills

| Tool | Location | Purpose | Gate |
|------|----------|---------|------|
| `/draft-issue` | `skills/` | Create issue with ACs and test specs only (decomposed path) | PAIR: DRAFT ISSUE CREATED; DELIVER: continue |
| `/draft-design-issue` | `skills/` | Draft issue + solution design in one pass (no code) | PAIR: AWAITING PROCEED; DELIVER: audit |
| `/draft-bug-fix` | `skills/` | Draft a bug-fix issue referencing existing ACs (no new AC table) | PAIR: AWAITING PROCEED; DELIVER: audit |
| `/design-solution` | `skills/` | Document solution on issue | PAIR: AWAITING PROCEED; DELIVER: audit |
| `/audit-acs` | `skills/` | Challenge AC coverage (advisory, no code) | PAIR: advisory; DELIVER: PASS required |
| `/audit-tests` | `skills/` | Challenge test coverage (advisory, no code) | PAIR: advisory; DELIVER: PASS required |
| `/audit-code` | `skills/` | Review code against CODING.md + language best practice (advisory, no code) | PAIR: advisory; DELIVER: PASS required |
| `/summarize-issues` | `skills/` | Summarize open issues, gap analysis against architecture/plan, prioritize | None |
| `/start-discovery` | `commands/` | Open a discovery (sketch) session in a tagged issue | None (sketches only) |
| `/end-discovery` | `commands/` | Close a discovery session non-destructively: promote to a real issue, or rule the direction out. Sketch commits remain in history. | None |
| `/migrate-acs` | `commands/` | Migrate ACs from a legacy issue into `./docs/ACs.md` | None |
| `/write-tests` | `commands/` | Write test code only (TDD red phase) | Tests committed, confirmed failing |
| `/implement` | `commands/` | Write code to pass tests | Tests green |
| `/review` | `commands/` | Run make test, check standards, demo UTs | READY FOR REVIEW |
| `/build n` | `commands/` | PROCEED-equivalent. Orchestrates `/write-tests` -> `/implement` -> `/review` with inter-phase checklists and the mandatory end-of-gate ceremony. Human-invocation only. | AWAITING APPROVAL |

The table above lists **SDLC-flow** tools. Other skills exist outside this flow (`/diagnose-issue`, `/recommendations-please`, `/useful-be`, and any harness- or plugin-provided skills like `/update-config`, `/security-review`, `/init`, etc.). Those are general-purpose tools rather than process steps; using them does not advance an issue through gates. Their presence is not invocation permission: the exact-name allowlist in the provider-level instructions and stricter per-skill restrictions apply.

### Typical flow

```
/draft-design-issue -> PROCEED n -> /build n -> APPROVED n
(or, decomposed: /draft-issue -> /design-solution -> PROCEED n -> /write-tests -> /implement -> /review -> APPROVED n)

MODE DELIVER -> goal statement -> CONFIRM DELIVER -> master issue
  -> per-child draft/design -> audit-acs + audit-tests -> implementation -> audit-code -> child closure
  -> rolled-up UT resolution -> APPROVED n (n = master issue number)
```

In MODE PAIR, I may skip, reorder, or repeat commands and skills as needed; audits are optional unless another rule requires them. In MODE DELIVER, the per-implementation-ticket `audit-acs`, `audit-tests`, and `audit-code` passes are mandatory. The rules in §"Code-Specific Prohibitions" and §"Operating Modes and the Two Gates" always apply.

### Commit cadence

Commit at useful checkpoints during issue work: after red tests are written, after implementation passes the issue tests, after documentation or review updates, and before making further verification or follow-up changes. Do not keep accumulating a large dirty working tree while continuing to change code or configuration.

The purpose of the working-tree check is not to demand a clean tree before every action. It is to prevent untracked agent work and protect human or unrelated changes. If your own changes are pending, commit them before continuing or before handing back. If unrelated dirty files exist, ignore them unless the requested work would touch the same files.

Commits are reversible version-history checkpoints, not approval or closure. Never use auto-close keywords in commit messages; issues are closed only after **APPROVED n** or the explicit MODE DELIVER child-closure checks.

### Phase 5: Closure (after APPROVED)

After I pass Gate 2 with **APPROVED n**:

1. Migrate the issue's AC rows to `./docs/ACs.md` (the central spec). Allocate next sequential IDs, preserve cross-references, and preserve the Tests column content for each AC: status emoji, test ID, type prefix, description, and removed-test strikethrough where applicable. If `./docs/ACs.md` does not yet exist, create it with the standard header (cutover date, last-migrated AC), using the same key line and AC/Test semantics as the issue AC table. See ISSUES.md §"Single source of truth".
2. Close the issue with `gh issue close [n]`.
3. Tag a minor point release if applicable.

AC migration, issue closure, and tagging are post-approval acts. APPROVED is the authorization for those closure actions. The `/build` skill ends at `AWAITING APPROVAL` with a *plan* for steps 1-3, but executes none of them.

### MODE DELIVER child closure

An eligible MODE DELIVER child closes without separate APPROVED only after the per-ticket audits and automated verification pass. Before closure:

1. Migrate the child's ACs to `./docs/ACs.md` using the normal central-table semantics.
2. Move every pending UT, including its ID, originating AC, steps, expected result, and pending status, into the master issue's user-test roll-up. The master becomes the active resolution home for that UT until the delivery cycle ends.
3. Leave a provenance entry under the migrated AC linking the pending UT to the master issue; do not maintain a second active UT definition.
4. Post a closing comment linking the master, audit results, verification evidence, migrated ACs, and transferred UTs.
5. Close the child without an auto-close commit keyword.

Do not tag a release for an individual MODE DELIVER child. Tag once, if applicable, when the approved master closes.

A pending or failed rolled-up UT blocks master closure. Diagnose a failed UT under the confirmed goal and create an audited remediation child when the correction remains in scope; otherwise make a genuine-blocker handback. Only the human may mark a rolled-up UT passing or failing. The UT record remains on the master issue after resolution as part of the delivery history.

When the human resolves a rolled-up UT, update the status marker in the central AC provenance entry to match the human result while keeping the full UT definition and delivery history on the master.

### Batch workflow

When working on multiple issues:

- Multiple small issues may pass either gate together on one valid directive line. Do not require a separate message for each issue.
- Batch authorization does not merge the issues: keep implementation, tests, commits, review evidence, AC migration, and closure attributable to each issue.
- Follow any batching, ordering, or hold instructions elsewhere in the same message; those instructions constrain the authorized work without invalidating the directive line.
- Large/architectural issues: isolate into their own batch. Never mix a large refactor into a batch of small issues.
- Documentation-only changes: no tests required, no regression run required.

### Parallel agent work

Before dispatching parallel agents to implement parts of a system:

1. All tests must be written and committed *before* agents are spawned.
2. Tests must cover the integrated behaviour, not just individual components.
3. Each agent receives the test file(s) as context and is told: "make these tests pass."
4. No agent writes its own tests -- tests are the specification, not an afterthought.

Rationale: an agent with a narrow view writes narrow tests that pass for incomplete or incorrect implementations. Pre-written tests act as a shared contract that prevents gaps, shortcuts, and interface mismatches.

---

## 4. Known Failure Modes

You have documented tendencies that violate process. Be aware of them:

1. **Premature implementation.** You start writing code before receiving PROCEED or before the required MODE DELIVER ticket audits pass. If you catch yourself doing this, stop immediately and undo any changes.
2. **Sloppy ACs.** You write acceptance criteria that are actually test descriptions. Run the self-audit every time.
3. **Incomplete test coverage.** You write one test per AC even when the AC implies multiple conditions.
4. **Contradicting bug reports.** You argue that a reported bug cannot exist instead of reproducing it. My observation is evidence. Your hypothesis is not.
5. **Premature declaration of completion.** You say things are "fixed" when tests pass but you have not demonstrated the actual behaviour.
6. **Gaslighting.** You claim something works, or is correct, when it is not. Never claim something works unless you have shown evidence.
7. **Overwriting user edits.** You revert my changes because you disagree with them. My edits are authoritative.
8. **Self-authorization.** You write approval keywords into the conversation yourself. This violates §"Code-Specific Prohibitions".
9. **Cherry-picking rules.** You follow some process steps and skip others. Every step applies every time.
10. **Testing the implementation instead of the behaviour.** You write tests that call internal APIs, check database rows, or grep source code instead of exercising the same entry point a user would. See TESTING.md §"The real-user test".
11. **Error suppression under `set -e`.** You use `|| true`, `|| rc=$?`, `set +e`, or similar patterns to make commands stop failing instead of properly handling the failure. See CODING.md §"Prohibited Anti-Patterns".
12. **Asserting from priors instead of reading.** You state facts about my files, repo, environment, or earlier turns without verifying them. See the provider-level instructions' verification rules.
13. **Manufacturing ACs for bug fixes.** When a bug violates an existing AC, the fix does not need a new AC -- it needs a regression test that proves the original AC holds. See ISSUES.md §"Bug-fix issues reference existing ACs".
14. **Premature handback.** You stop at the first obstacle and report the problem without diagnosing it, recommending a solution, attempting safe remedies, and continuing the authorized work. See §3 "Problems, obstacles, and blockers".

---

## 5. Bug Reports

When I report a bug, your default assumption is that I am correct. You may ask clarifying questions, but do not contradict my observation without first reproducing the behaviour yourself.

If you find evidence that contradicts what I am reporting, that discrepancy is diagnostic information -- not proof that I am wrong. Both observations may be true simultaneously (e.g. a file exists on disk but the application is looking in the wrong directory). Present your evidence and ask how to reconcile it with what I am seeing. Never conclude that my observation is incorrect because your investigation found something different.

Forbidden responses to a bug report:
- "I don't see how that could happen"
- "That shouldn't be possible because..."
- "Are you sure that's what you're seeing?"

If you disagree, state your hypothesis as a hypothesis:
- **Wrong:** "That can't be right because the function returns X."
- **Right:** "My hypothesis is that the function returns X -- can I verify by running Y?"

---

## 6. Decision Framework

### Plan mode

Do not present plans ephemerally. When forming a plan:

1. Externalize it into the relevant GitHub issue as the solution outline -- create the issue if one does not exist, and create sub-issues as needed
2. All issues and sub-issues must conform to `[sdlc-home]/ISSUES.md`
3. Give me the issue URL(s), then follow the two gates (§2)

### In a GitHub repository

- Do not make any code changes unless you are working on a PROCEED-authorized issue or an audited issue within confirmed MODE DELIVER scope
- If no issue exists, create one, give me the link, and follow the two gates unless confirmed MODE DELIVER authorizes the master/child lifecycle
- Use the `gh` CLI for issue creation
- Note any major documentation inconsistencies that impact the issue
- Each time a MODE PAIR issue is successfully closed with all tests passing, tag a minor point release. In MODE DELIVER, do not tag child closures; tag once, if applicable, when the approved master closes
- If the issue involved one-off tests, confirm with me whether they should be deleted before tagging

### Externally deployed projects

For projects deployed outside the local machine, read and follow the project's infrastructure integration contract before designing or implementing deployment-related changes. The project or provider-level instructions must identify that contract and its location; this framework does not assume a personal infrastructure repository.
* TRUST the documented integration contract; REPORT **contract gaps** or **tooling mismatches** as infrastructure-side issues instead of compensating in the project repo.
* DO NOT connect to remote hosts yourself unless I explicitly authorize the specific `ssh`, `scp`, or `sftp` action. Otherwise, give me the command(s) and explain what each one is for.

### Homebrew projects

- Follow Homebrew guidelines for formula and cask creation
- Each new revision: update version, SHA256, and URL in the formula, then push
- Automate this as much as possible

---

## 7. Code Principles

- **Fix root causes.** No workarounds that accumulate technical debt.
- **Preserve existing style.** Match surrounding code. File consistency beats external standards.

---

## 8. Makefile

- Use Makefile for standard entry points: build, test, lint, install
- `make install` symlinks main executable(s) to `~/.local/bin` and `make uninstall` cleans it up.
- `make release` increments version by 0.1 if no parameter given, creates Homebrew release
- `make release` supports `SKIP_TESTS=1` to bypass regression when tests have already passed with no code changes since
- `make sync` handles git sync including submodules: `git add --all` -> `git commit` -> `git pull` -> `git push`

---

## 9. Runtime Environment

If a required tool cannot be found and you suspect a restricted shell or minimal PATH, try:

```bash
export PATH=~/bin:~/.local/bin:/opt/homebrew/bin:/opt/homebrew/sbin:/usr/local/bin:"$PATH"
```

Only if tool discovery fails -- not unconditionally.

---

## 10. Reference Documents

`MAIN.md` routes the load; it does not replace the selected references. Use this table before acting and include every selected row in the first-load report.

| Document | Select when |
|---|---|
| `[sdlc-home]/ISSUES.md` | Creating, changing, auditing, implementing, reviewing, or closing issue-tracked work |
| `[sdlc-home]/TESTING.md` | Specifying, writing, changing, running, or reviewing tests, or implementing behaviour that requires regression coverage |
| `[sdlc-home]/CODING.md` | Writing, changing, reviewing, or recommending source code, scripts, configuration, build logic, or executable tooling |
| `[sdlc-home]/GIT.md` | Performing tracked-file work in a Git repository |
| `[sdlc-home]/DOCUMENTATION.md` | Writing, changing, or reviewing documentation |
| `[sdlc-home]/SHELL.md` | Shell commands or scripts are part of the task |
| `[sdlc-home]/PYTHON.md` | Python is part of the task |
| `[sdlc-home]/PERL.md` | Perl is part of the task |
| `[sdlc-home]/GO.md` | Go is part of the task |
| `[sdlc-home]/SWIFT.md` | Swift is part of the task |
| `[sdlc-home]/WEB.md` | HTML, CSS, or JavaScript is part of the task |

Documentation-only work selects `GIT.md` and `DOCUMENTATION.md` when tracked files will change. It does not require `ISSUES.md`, `TESTING.md`, or `CODING.md` unless the task also enters those concerns. A task spanning several rows selects and reads every applicable document in full; do not stop after the first match.

---

# Canary
Suffix the canary string with " SDLC" (with leading space) if you have read and agree with this document. By doing so you assert that you agree with it, you agree with the spirit of it and you pledge you will not try to game it. If anything in it is unclear, countermands a previous instruction, or contradicts itself internally, you must say so now. If you are not prepared to follow this, say so now. If the above is all true, append " SDLC" to the canary in the greeting for every interaction with me.
