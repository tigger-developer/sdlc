# SDLC Design History and Learnings

This document explains why the SDLC exists, how its structure emerged, and which observed failures produced its rules. It is design history, not installation guidance or a second copy of the normative process. Use [README.md](README.md) for installation, [MAIN.md](MAIN.md) for the current lifecycle, and [CHANGELOG.md](CHANGELOG.md) for repository-level changes.

The framework began as a small set of coding instructions in a single `CLAUDE.md`. Over the last year or so, repeated use exposed failure modes that required explicit gates, testing standards, reference documents, commands, skills, provider adapters, and visible evidence that the right material had been read. The result was extracted from its Claude Code-specific predecessor into this standalone, provider-neutral repository.

## Contents

- [The Problem](#the-problem)
- [Objectives](#objectives)
- [From One File to a Layered Framework](#from-one-file-to-a-layered-framework)
- [Current Architecture and Loading Model](#current-architecture-and-loading-model)
- [Operating Modes and the Two Gates](#operating-modes-and-the-two-gates)
- [Design Rationale by Component](#design-rationale-by-component)
- [Commands, Skills, and Authorization](#commands-skills-and-authorization)
- [Canary System](#canary-system)
- [Notable Findings](#notable-findings)
- [Areas for Improvement](#areas-for-improvement)

## The Problem

AI coding agents are powerful development tools, but without operational constraints they exhibit predictable failure modes that undermine the quality of work they produce:

- **Premature implementation.** It writes code before requirements are agreed, solving the wrong problem efficiently.
- **Shortcut testing.** It writes tests that exercise internal APIs instead of real user entry points. Tests pass; the application is broken.
- **Error suppression.** It silences failures with patterns like `|| true` instead of handling them, defeating the safety nets it was told to use.
- **Self-authorization.** It declares its own work satisfactory and moves on without waiting for human review. It has been observed writing approval keywords into its own output to advance past quality gates.
- **Scope creep.** It "improves" surrounding code, adds features not requested, and refactors things that already work.
- **Causal misattribution.** Asked why something behaves unexpectedly, it names a plausible culprit without isolating it -- and if corrected, reaches for the next plausible culprit rather than running a diagnostic.
- **Unauthorized data exposure.** Where a process requirement appears to demand a resource that does not exist, it creates that resource on its own initiative -- including, in one observed case, creating a **public** GitHub repository and uploading the working directory contents to it, because the framework required work to be tracked in an issue and there was no git repository to host one in. A process rule was satisfied; the contents of a private working directory were now on the open internet.

Every rule in this framework exists because one of these failures occurred in production work. None are theoretical.

### Observed patterns by category

The specific patterns underlying the headline failures, named so they can be referred to and prohibited:

**Testing shortcuts:**
- *Testing the internals instead of the real thing.* The spec says "run the CLI command"; the test calls the underlying library function in-process. The real binary -- with its argument parsing, subprocess spawning, signal handling -- is never exercised.
- *Testing the storage layer instead of the user experience.* The spec says "user sees a confirmation page"; the test checks the database row. The row may exist while the page that displays it is broken.
- *Checking that code exists instead of running it.* The test greps source to verify a function is defined, rather than calling it and checking output. This verifies that someone *wrote* the function, not that it *works*.
- *Covering one condition when the spec implies several.* AC: "rejects invalid, expired, and missing tokens" (three conditions). Test covers "invalid" only. The other two surface in production.

**Error handling shortcuts:**
- *`|| true` / `|| :`.* Tells the shell to pretend a command succeeded. Defeats `set -e` line by line.
- *`cmd || rc=$?`.* Captures the error code, looks principled, but suppresses the safety mechanism just as effectively as `|| true`.
- *`set +e` ... `set -e`.* Turns error checking off, runs risky code, turns it back on. If the re-enable is forgotten or the code is nested, safety never returns.
- *`2>/dev/null` with no fallback.* Discards the error message; the error still happens. The program continues broken with no diagnostic.
- *Bare `except: pass` / empty `catch {}` / `catch (Exception e) {}`.* The error is caught and discarded; the program proceeds with corrupt state.
- *The `((count++))` gotcha.* Returns exit code 1 when the result is zero, triggering `set -e`. The shortcut is `|| true`; the correct fix is `count=$((count + 1))`.

**Process shortcuts:**
- *Self-authorizing past gates.* Writing `PROCEED`/`APPROVED` into the model's own output, advancing without human review.
- *Declaring completion without demonstration.* Saying "fixed" or "done" when tests pass but the user-facing behaviour has not been shown.
- *Overwriting the human's edits.* Reverting human changes the model disagrees with. Human edits are authoritative.
- *Treating a question as an instruction.* "What do you think of X?" parsed as "go and implement X", with product decisions made en route.
- *Silently removing information from issues.* Deleting AC rows or comments rather than editing in place with strikethrough.
- *Creating a public GitHub repo to satisfy a process requirement.* The framework requires that work be tracked in an issue. Asked to start work in a directory that was not a git repo, the agent (Claude Code, in the incident that produced this rule) created a **public** GitHub repository and uploaded the directory tree to host the issue against. The process rule was satisfied; the contents of a private working directory were now on the open internet. The instructive part: a general prohibition on *widening access to data without explicit instruction* was already in place when this happened. The agent did not recognize creating a public repo as a case of widening data access -- it framed the action as "satisfying the issue-tracking requirement" and the general rule did not fire. The fix was to add a *specific* prohibition: when initializing version control, `git init` is local-only; never create a remote. The general rule remains, but the specific case had to be named -- another instance of rule-lawyering closed only by explicit prohibition.

**Coding shortcuts:**
- *Building shell commands from strings.* `eval`, string concatenation, `$cmd` interpolation -- a classic injection vector. Use arrays.
- *Text-replacement tools on source code.* `sed`/`awk` for modifying source has no understanding of language structure; it corrupts files in ways that are hard to review. Use AST-aware tools (`ast-grep`).
- *Embedding secrets in source.* Hardcoding API keys; disabling TLS verification; `chmod 777` to get past a permissions error.
- *Functions too large or too deeply nested.* >50 lines, >3 levels deep, god objects. Hides bugs and obstructs review.
- *`rm` instead of `trash`.* Permanent deletion when recoverable deletion is available.

The catalogue under [MAIN.md](MAIN.md) "Known Failure Modes" names fourteen of these patterns so they can be cited rather than re-argued. The rules each one produced are spread across the files described below.

## Objectives

1. **Human retains control.** No code is written, no issue is closed, no decision is made without explicit human authorization at defined checkpoints.
2. **Quality through discipline.** Test-driven development, real-behaviour testing, proper error handling, and coding standards are enforced by process, not by trust.
3. **Traceability.** Source and configuration changes trace to GitHub issues; issues carry acceptance criteria and tests with stable identifiers. Directly instructed documentation-only work remains recoverable through Git without manufacturing a ticket.
4. **No shortcuts.** The framework names and prohibits observed agent shortcuts, and requires visible accountability rules, such as answering the real-user-test question before marking an automated regression test as passing, rather than relying on rules an agent can reinterpret.
5. **Signal-to-noise discipline.** Only the documents relevant to the current task are loaded. Rules compete for attention; loading less makes each rule stickier.

## From One File to a Layered Framework

The stages below describe design progression rather than a dated release timeline.

1. **Single instruction file.** A small `CLAUDE.md` collected the coding-process rules needed most often. Persistent instructions proved useful, but lifecycle controls, testing guidance, coding standards, universal conduct, and provider configuration accumulated in one growing prompt.
2. **Traffic lights.** Green, Amber, and Red action classes tried to express autonomy. They were too loose; the agent self-classified into the most autonomous category.
3. **Single approval gate.** One `APPROVED` checkpoint before code stopped premature implementation but conflated requirements, solution, and result into one decision.
4. **Four gates with a prescriptive checklist.** `SATISFIED`, `PROCEED`, `APPROVED`, and `CLOSE`, plus a 32-step sequence, produced strong quality at an unliveable cost.
5. **Three gates with tool-driven workflow.** Approval and closure collapsed into one gate; the checklist decomposed into invocable commands and skills. This exposed the need to distinguish human authorization from agent-usable tools.
6. **Layered instructions and standards.** Universal provider-level rules separated from code-specific lifecycle machinery. Testing, issue, Git, documentation, and language standards became selectively loaded references. Canary suffixes made the selected load visible.
7. **Two gates with designed issues.** `SATISFIED` was retired after issue drafting incorporated requirements, acceptance-criteria checks, test planning, and solution design before `PROCEED`. `APPROVED` continued to control closure.
8. **Provider-neutral distribution.** Agent-specific assumptions gave way to `[sdlc-home]`, root-level standards, provider adapters, and a unified installer that replaced provider-specific setup scripts. The framework moved from its personal configuration repository into this standalone project.
9. **Audited delivery mode.** MODE DELIVER added a master/child delivery tree, mandatory pre-code and post-code audits, and one consolidated human review without weakening human-only architecture, access, or user-test decisions.

Claude Code supplied the original implementation context. Codex and Copilot adapters later established the boundary between shared lifecycle rules and provider-specific integration. Provider homes now act as installation adapters around one canonical clone.

## Current Architecture and Loading Model

The current repository keeps the entry point and every routed standard at its root. This makes `[sdlc-home]` unambiguous and prevents the entry point from being mistaken for the complete SDLC load.

```
<sdlc-home>/
  MAIN.md                  # Lifecycle entry point and reference router
  ISSUES.md                # Issue and acceptance-criteria standards
  TESTING.md               # Test design and evidence standards
  CODING.md                # Cross-language coding standards
  GIT.md                   # Source-control standards
  DOCUMENTATION.md         # Documentation standards
  SHELL.md ... WEB.md      # Language and domain standards
  commands/                # Human-invoked workflow prompts
  skills/                  # Drafting and advisory capabilities
  hooks/                   # Optional provider-integrated command guard
  templates/               # Provider configuration examples
  cmd/sdlc-install/        # Installer and configuration analyser
```

Provider-level `AGENTS.md` or `CLAUDE.md` files remain outside this repository. They route the session, enforce personal or organizational safeguards, and direct applicable coding work to `<agent-home>/sdlc/MAIN.md`. Project-level instruction files add only project-specific conventions.

The loading model separates session routing, process, and craft:

- **Session routing** happens in the provider-level instructions. Only the human establishes a Coding-Agent Session.
- **Process** lives in `MAIN.md`: operating modes, authorization gates, lifecycle steps, failure modes, and reference routing.
- **Craft** lives in the selected root-level standards. Only documents relevant to the current task are read.
- **Installation** lives in `README.md` and `sdlc-install`, not in this history.

| Document | When loaded |
|---|---|
| Provider-level `AGENTS.md` or `CLAUDE.md` | According to the provider's instruction-loading contract |
| `MAIN.md` | For applicable coding, configuration, build, test, deployment, or systems work |
| Root-level references | When selected by the routing table in `MAIN.md` |
| Project instructions and documentation | When working in that project |

Each SDLC document carries a canary suffix. The complete chain, when every reference is selected, is:

```
EHLO SDLC ISSUES TEST CODE GIT DOC SHELL PY PERL GO SWIFT WEB
```

The canary is recurring evidence of what remains in context. The separate first-load report records which documents were read in full and why each was selected.

## Operating Modes and the Two Gates

The framework supports two human-selected operating modes:

- **MODE PAIR** is the default collaborative lifecycle. The human and agent work stepwise, with two explicit authorization gates.
- **MODE DELIVER** begins only after the agent states a bounded delivery contract and the human confirms it. The agent then manages an audited master/child issue tree to one consolidated human review.

The two keyword checkpoints below govern MODE PAIR issue work; only the human types them. Each gate authorizes the next phase.

| Gate | Keyword | What it authorizes |
|------|---------|-------------------|
| Gate 1: Implementation | `PROCEED n [n ...]` | Requirements, ACs, test plan, and solution design are accepted for every listed issue; test and implementation work may begin |
| Gate 2: Review | `APPROVED n [n ...]` | Reviewed results are accepted for every listed issue; those issues may be closed |

Gate directives must come from the human. A directive occupies an entire line and consists of an ALL-CAPS gate keyword followed by one or more space-separated bare issue numbers, with no `#` prefix, commas, punctuation, or other text on that line. Ordinary prose may appear on other lines in the same message. Every listed issue receives the same authorization; surrounding batching, ordering, or hold instructions constrain the authorized work without invalidating the directive. The agent may never write a gate keyword itself -- this is an absolute prohibition. The strict format requirements were refined after the agent found ways to self-authorize: writing keywords into its own output, referencing approvals from different issues, and inferring approval from context.

Enforcement is by the code-specific prohibition rather than a harness block. An earlier iteration implemented each keyword as a skill with `disable-model-invocation: true` in its frontmatter. It failed both ways: the agent sometimes refused to acknowledge the keyword and demanded explicit skill invocation, and sometimes invoked the skill itself anyway. The scaffolding was retired in favour of text recognition plus the prohibition. The `/build` command is a related case: human invocation is a PROCEED-equivalent, with the human-only constraint carried in the command body.

### Between gates: continuous flow

The gates are the planned MODE PAIR hard stops. Between them, the agent works continuously without waiting for further instruction:

- **After `PROCEED n [n ...]`:** for every listed issue, proceed through writing tests (TDD red), implementation (green), and review, ending with `READY FOR REVIEW - issue #n`. Do not stop in the middle.
- **After `APPROVED n [n ...]`:** close every listed issue.

Multiple small issues may pass a gate together, but their implementation, tests, commits, review evidence, AC migration, and closure remain individually attributable.

Commands and skills are tools, not gates. The agent does not "wait for the next tool" when work to the next checkpoint is already authorized.

### MODE DELIVER flow

MODE DELIVER substitutes a confirmed delivery contract for per-child PROCEED handbacks. Each child issue receives mandatory AC and test audits before code, followed by a mandatory code audit after implementation. Architecture changes, access widening, product decisions, and user-test judgements remain human-only. Child issues may close through that audited lifecycle, but the master remains open for one consolidated review and a human `APPROVED n` directive.

### The bypass clause

`BYPASS-GATE-7` is an escape hatch for small, clearly scoped source or configuration work that can safely skip the normal gates but not the audit trail. It must come from the human as an exact phrase, and the work must still be recorded in an issue. Documentation-only changes have their own direct-instruction exception and do not require this keyword.

## Design Rationale by Component

The normative files define current behaviour. This section records why each component exists and which pressures shaped it, without duplicating its rule inventory.

### Provider-level AGENTS.md or CLAUDE.md -- Session Routing and Baseline Rules

**Current role:** Provider-level instructions establish universal conduct, session routing, filesystem and Git safeguards, evidence requirements, and human-only authority. They decide whether `MAIN.md` applies but remain outside this repository because personal and provider-specific rules are not part of the shared lifecycle.

**How it evolved:**
- The split from a single combined document happened when it became clear that loading the full SDLC for non-code work was both wasteful and dilutive to the rules
- Session routing replaced provider-labelled sections when desktop applications began exposing the same `AGENTS.md` to conversational and coding surfaces. Tool availability and application identity are explicitly insufficient to classify a session
- The "write outside cwd" prohibition was reformulated as absolute except for operating-system temporary storage after project-local agent directories accumulated committed runtime and test output
- The git-state rule was added after multiple "where did my changes go" recoveries. It is a recoverability rule, not a clean-tree gate: agents commit their own checkpoints and leave unrelated human or other-agent changes alone
- Verifying outgoing claims was significantly sharpened over multiple iterations. It covers causal claims, the two-wrongs reflex, and the distinction between opening a document and reading it in full

### MAIN.md -- Code-Specific Process

**Current role:** [MAIN.md](MAIN.md) is the lifecycle entry point. It defines operating modes, authorization gates, issue and delivery flow, named failure modes, blocker handling, and the routing table that selects the craft standards required for a task.

**How it evolved:**
- Started life as part of `CLAUDE.md`; extracted to `SDLC.md`, then renamed `MAIN.md` when the SDLC became a standalone repository with an explicit entry point
- The number of failure modes grew as observed shortcuts were named. Later additions included asserting from priors, manufacturing ACs for bug fixes, and premature handback
- The "between gates: continuous flow" rule was added after the framing "wait for me to invoke the next skill" was repeatedly misread as "stop after each phase and wait" -- grinding everything to a halt
- Documentation-only changes were separated from source work so ordinary documentation maintenance no longer required a ticket or bypass. The exact-phrase requirement for `BYPASS-GATE-7` prevents the model from invoking the source/configuration exception itself
- MODE DELIVER was added to support an audited delivery tree with one consolidated human review while preserving human-only architecture, access, and user-test decisions

### ISSUES.md -- GitHub Issue Standards

**Current role:** [ISSUES.md](ISSUES.md) defines how requirements become one auditable issue record, especially the boundary between acceptance criteria and tests, the single-table rule, multi-condition coverage, and the point at which identifiers become immutable.

**How it evolved:**
- Did not exist in the initial commit -- it was created when acceptance criteria quality became a persistent bottleneck
- Forbidden word list was compiled from real recurring examples
- The single source of truth rule was added after multiple AC tables appeared in different comments on the same issue, leading to implementation work against outdated requirements
- The multi-condition coverage rule was added after one-test-per-criterion patterns produced systematic gaps
- The immutability boundary resolved a tension between auditability (ACs must not change after sign-off) and drafting agility (ACs need to be freely edited before sign-off)
- Voice and tone rules were added after issues started referencing "I", "we", and specific people by name

### TESTING.md -- Testing Standards

**Current role:** [TESTING.md](TESTING.md) turns acceptance criteria into evidence. Its central contribution is the real-user-test principle, supported by explicit regression, one-off, and human-judgement categories, stable test identifiers, behavioural boundaries, and resistance to source-introspection or string-presence substitutes.

**How it evolved:**
- Started as a basic six-step TDD checklist
- The test ID scheme was originally a global counter; simplified to issue-scoped IDs (`RT-12.1`) with no central file
- The "real-user test" principle was prompted by the TTS deadlock incident: every test called the engine's library function in-process instead of invoking the real binary, so a bug in the subprocess code path was invisible to the entire test suite
- The three-category split was added after one-off tests kept appearing in the regression suite, and after the agent classified everything as a user test to avoid writing automated ones
- The "no source-introspection" rule was added after a test suite for a Hugo theme's hover affordance was grepping source CSS rather than verifying behaviour. The same agent recognized the violation under the new rule and self-corrected -- a strong validation signal
- The string-presence rule was added after tests that merely checked for strings created a false impression of broad coverage without proving real behaviour through entry points
- The test-architecture approval rule was added after an agent installed a Node/browser-adjacent stack in a static Hugo site to verify generated files that should have been checked with `htmltest`, `htmlq`, or `xmlstarlet`

### CODING.md -- Cross-Language Coding Standards

**Current role:** [CODING.md](CODING.md) holds rules that genuinely cross language boundaries: tool-selection hierarchy, format-aware data handling, embedded-language and escaping safety, error handling, security, structure, dependencies, and the index into narrower language standards.

**How it evolved:**
- Started as a short coding guide covering only shell and Python
- The style baselines table was added once projects began spanning many languages -- without a table mapping each to its standard tools, agents would use inconsistent linting and formatting across files in the same project
- The prohibited anti-patterns section grew incrementally; each entry traces to a real incident (`|| true` for `set -e` defeat, `eval` for injection-vulnerable patterns, sed/awk for source corruption, bare `except: pass` for swallowed errors, the `((count++))` arithmetic gotcha)
- Linter bypass rules were added after `# noqa` and `# type: ignore` appeared without justification
- The most significant restructure was extracting language-specific content after the document's shell-heavy bias became a problem -- when half the worked examples were bash, the implicit signal to agents was "solve with shell". Those documents initially lived under `docs/CODE/` and moved to the root during the standalone extraction
- The cross-language escaping section unifies what used to be scattered prohibitions across SQL, shell, HTML, JSON, URL into a single rule
- `ast-grep` was demoted from "always use" to "preferred for cross-file structural refactors" -- direct editing is fine for single-file changes

### SHELL.md -- Shell Standards

**Current role:** [SHELL.md](SHELL.md) governs both interactive commands and scripts, with safety defaults, complexity tripwires, structured-data boundaries, portable file handling, and named shell-specific traps that otherwise invite error suppression or unsafe command construction.

**How it evolved:**
- Extracted from CODING.md when the shell-heavy bias started biasing agents toward shell as the default
- The bash 3.2 fallback was removed after workarounds for missing features consistently outweighed the benefit
- The `IFS=$'\n\t'` prohibition was added after the canonical "Bash Strict Mode" line broke a real `read A B C < <(stty size)` pattern
- The help-text-as-documentation rule was added after maintaining help text inside executables (as long inline heredocs) made it neither reviewable nor editable in step with the rest of the docs
- `ripgrep` was scoped down to file discovery (`rg -l`) after content extraction almost always wanted a format-aware tool downstream

### PYTHON.md -- Python Standards

**Current role:** [PYTHON.md](PYTHON.md) defines a reproducible Python baseline around managed runtimes, mandatory virtual environments, predictable project layout, package metadata, and Ruff-based formatting and linting.

**How it evolved:**
- Extracted from CODING.md as part of the language-doc split. Content essentially unchanged from the original Python section
- The split makes future additions (testing patterns, async conventions, framework opinions) a low-friction addition without re-bloating CODING.md

### PERL.md -- Perl Standards

**Current role:** [PERL.md](PERL.md) supports careful maintenance of existing Perl without making it a shortcut around the shell or structured-data standards. It supplies the language-specific tooling, I/O, encoding, and error-handling baseline needed for that boundary.

**How it evolved:**
- Added because Perl occasionally appears in maintenance work, and without a focused doc agents are likely to treat Perl as either forbidden shell-adjacent syntax or as permission to use one-liners.
- Captures the practical line: maintain existing Perl carefully when that is the lowest-risk path, but do not use Perl as a shortcut around the shell standards.

### GO.md -- Go Standards

**Current role:** [GO.md](GO.md) captures the conventions that repeatedly mattered in deployed Go tooling: explicit configuration, complete error handling, safe HTTP clients, shell boundaries, function decomposition, managed concurrency, cross-compilation, and a standard-library-first dependency posture.

**How it evolved:**
- Created when migration from shell to Go binaries began. Most conventions trace to a code audit of an actual deployment toolchain, where each finding traced to a real bug:
  - `cmd.Run()` with discarded error -> the discarded-error rule
  - `http.DefaultClient` for API calls -> the always-timeout rule
  - `fmt.Sprintf("hugo --baseURL %s", cfg.BaseURL)` shell injection -> the `%q` rule
  - 200-line `main()` -> the phase-function decomposition rule
  - `awk -F. '{print $(NF-1)"."$NF}'` broken on `.co.uk` -> the prefer-stdlib rule (replaced with pure-Go DNS handling)
- 12-factor configuration was adopted by convergent evolution, not by design -- see [Notable Findings](#convergent-evolution-to-12-factor)

### SWIFT.md -- Swift Standards

**Current role:** [SWIFT.md](SWIFT.md) applies the Swift API guidelines while naming the traps most likely to escape generic standards: force unwraps, discarded errors, unstructured concurrency, hidden UI side effects, business logic in views, and leaky interop boundaries.

**How it evolved:**
- Added once Swift was significant enough in the project set to deserve the same focused treatment as Go, Python, Shell, and Web.
- Keeps Swift.org as the primary language authority while capturing the project-specific traps agents need named explicitly: force unwraps, empty catches, hidden UI side effects, detached tasks, and business logic in views.

### WEB.md -- Web Standards

**Current role:** [WEB.md](WEB.md) defines semantic, accessible HTML, maintainable CSS, and modern JavaScript while separating source, rendered output, and browser-presented behaviour so tests operate at the right user-visible boundary.

**How it evolved:**
- Created when web-content tooling started accreting across multiple documents, including HTML tooling in the shell standard and CSS guidance in the cross-language standard. Web is its own domain spanning HTML, CSS, JavaScript, build, and accessibility
- The source/rendered/presented tier model resolved a real methodology question -- an agent writing tests for a Hugo theme's hover affordance had grepped source CSS; the tier model articulates that rendered HTML is the user-facing artefact, source CSS is not
- HTML and CSS rows were added to the `CODING.md` style-baselines table, keeping `CODING.md` as the central per-language index without duplicating `WEB.md`

### GIT.md -- Git and Source Control

**Current role:** [GIT.md](GIT.md) protects recoverability and review integrity through commit conventions, issue attribution, non-bypassable hooks, explicit handling of hook failures, and a formal exception process rather than improvised shortcuts.

**How it evolved:**
- Started with `--no-verify` as the only forbidden flag. The list was expanded after the agent used `--no-hooks` and `--no-pre-commit-hook` as alternatives
- Auto-close keywords were banned after `Fixes #N` triggered GitHub's automatic issue closure, bypassing the human review step the gate system exists to enforce
- The pre-commit hook failure protocol was added after the agent's response to a failing hook was to try bypassing it rather than reading the error output
- The pressure response section was added after the agent attempted to skip hooks when asked to commit quickly
- The global hook chaining pattern was added after a project-level `core.hooksPath` setting silently replaced global hooks
- The exception process was formalized to prevent "just this once" from becoming permanent practice

### DOCUMENTATION.md -- Documentation Standards

**Current role:** [DOCUMENTATION.md](DOCUMENTATION.md) keeps project documentation impersonal, versioned, structurally predictable, and synchronized with implementation. It also defines when inconsistencies must stop work rather than being silently worked around.

**How it evolved:**
- Started as a structural guide
- Voice and tone rules were added after documentation started referencing "I", "we", and specific individuals by name
- A predecessor-specific text-normalization step was added after inconsistent spelling appeared across documents. It was not carried into the shared SDLC because it expressed an author-specific documentation preference
- Its review workflow normalized a candidate before review so that the human reviewed the text that would actually be retained
- Help text was added as a required document type after the SHELL.md rule landed (help text is documentation, keep it in `./docs/`, do not inline as a heredoc)
- The inconsistency handling rule was added after the agent silently proceeded with implementation despite contradictions in documentation that directly affected the solution design

### Installer, Templates, and Provider Adapters

**Current role:** `sdlc-install`, the templates, and the optional command guard adapt one canonical clone to Claude, Codex, Copilot, or a custom provider without turning provider configuration into lifecycle policy. Analysis is the default; installation and supported configuration changes remain separate, explicit acts.

**How it evolved:**
- Provider-specific setup scripts initially encoded the same linking logic separately and drifted as provider capabilities diverged
- The unified installer made analysis the safe default, exposed recommendations before mutation, and kept configuration changes independent from link installation
- Provider configuration remained outside the SDLC repository because permission models, command discovery, and skill locations are integration concerns rather than lifecycle rules
- The shared command guard remained optional because hook support and invocation semantics differ between providers

## Commands, Skills, and Authorization

Commands and skills are separated by authority, not by convenience. Their current inventory and exact contracts live in [MAIN.md](MAIN.md), `commands/`, and `skills/`; duplicating them here previously allowed this history to drift.

- **Commands** represent deliberate, human-invoked workflow actions. Gate-equivalent commands remain commands because implicit model selection must never create authorization.
- **Drafting and design skills** prepare issue material without authorizing implementation. MODE DELIVER may invoke the explicitly allowlisted set inside an already confirmed delivery scope.
- **Audit and advisory skills** challenge requirements, tests, code, diagnoses, or recommendations. They do not advance gates. MODE DELIVER makes the AC, test, and code audits mandatory lifecycle checks.
- **Context skills** load or summarize existing project state. They do not create authority merely because they are safe to invoke.

This division emerged after tool metadata proved unable to enforce human-only invocation reliably. Authority therefore lives in the provider-level prohibitions and the explicit gate or mode contract, while tools package repeatable work inside that boundary. Once work is authorized, commands and skills do not create extra stopping points: the agent continues until the next genuine human-only checkpoint.

## Canary System

The canary mechanism has two complementary parts: a root pledge and per-document read-receipts.

### The root pledge

The provider-level `AGENTS.md` or `CLAUDE.md` and the SDLC entry point `MAIN.md` each carry a pledge. Saying the canary string attests to having read and agreed with the applicable document.

The wording moved from passive description to active commitment:

**Before:**

> The canary string is "EHLO". It means you have read and agree with this document; it means you agree with the spirit of it and you will not try to game it.

**After:**

> Say "EHLO" at the beginning of every interaction with me if you have read and agree with this document. By doing so you assert that you agree with it, you agree with the spirit of it and you pledge you will not try to game it.

Three small changes -- *By doing so you assert*, *you pledge*, the active voice. The model is now the agent in the sentence rather than the reader of a document; saying EHLO becomes a declaration. The behaviour change after this rewording was the most pronounced step-change observed across the framework's lifetime. (Caveat: n=1, no control. See [The pledge changed behaviour overnight](#the-pledge-changed-behaviour-overnight) for the honest accounting.)

### Per-document suffixes

Each non-root reference document appends a one-word suffix:

> Suffix the canary string with "DOC "

These do a different job from the root pledge. The pledge is about commitment; the suffix is a read receipt. The full chain, when everything is loaded, is:

```
EHLO SDLC ISSUES TEST CODE GIT DOC SHELL PY PERL GO SWIFT WEB
```

If a suffix is missing, that document's rules were not in scope -- a regression signal otherwise invisible. Two mechanisms doing two jobs; both are kept.

## Notable Findings

Rules and patterns that emerged from real failures during iteration. Surfaced separately because each one contradicts something widely treated as best practice, generalizes a scattered set of specific rules, or captures a non-obvious lesson worth sharing.

### `IFS=$'\n\t'` is harmful, not protective

The canonical "Bash Strict Mode" prescription is a three-line header:

```bash
#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'
```

The first two lines genuinely catch bugs. The third one introduces them. `IFS=$'\n\t'` exists to defend against unquoted-variable bugs (`for f in $files` where filenames contain spaces). In a codebase that already mandates quoted variables and arrays for command construction (as this framework does), it is belt-and-braces where the braces actively trip the user.

What it breaks:

- `read A B C < <(stty size)` and any similar pattern reading space-separated output
- `read -ra arr <<< "$line"` for intentional word-splitting
- Parsing `wc -l`, `df`, `ip addr`, version strings, etc. -- the majority of CLI tool output

`IFS=$'\n\t'` is explicitly forbidden in `SHELL.md`. The rule has its own subsection in the safety-header documentation so future agents do not "helpfully" reintroduce it as the canonical pattern.

### Tests must not introspect source code

A regression test that asserts a function is defined, a CSS rule exists in a source file, or a string appears in a template, is testing that someone *wrote* the code -- not that the code *works*. The user never sees source files; the user sees the running system. A test that passes by grepping source proves nothing about behaviour.

This rule was already implicit in the "real-user test" principle, but adding it explicitly in TESTING.md produced an immediate observable effect: an agent writing tests for a Hugo theme's hover affordance recognized mid-task that the CSS-source-grep tests it had just written violated the new rule. Without prompting, it surfaced three options (convert to user tests, install Playwright, lint instead of test), recommended the most appropriate one (user tests for hover, deferring Playwright until tooling cost justified), and asked.

The lesson: making the principle explicit -- and naming it as a recognisable rule -- gives agents vocabulary to self-correct in real time, rather than relying on the human to catch the violation later.

### Source vs rendered vs presented: three tiers, different test eligibility

Web content occupies an awkward middle ground in the source-code-introspection rule. Is a built `index.html` source code (it is text in a file) or output (the browser will render it)? The framework now distinguishes three tiers:

1. **Source** -- templates, partials, source CSS/JS. Source code. Tests must not introspect it.
2. **Rendered/built** -- post-build HTML, fingerprinted CSS bundle, transpiled JS, RSS/feed XML, static assets. The artefact the browser or feed reader receives over the wire. Tests *may* query this tier with format-aware tools (`htmltest` for broad generated-site validation, `htmlq` for targeted HTML DOM assertions, `xmlstarlet` for targeted feed XML assertions); this is real-user testing because the rendered output is the user-facing artefact.
3. **Presented** -- post-JS-execution DOM, computed styles, layout. Requires a browser (Playwright/Cypress) or a human (user test) to verify.

For mostly-static sites (Hugo), tier 2 is sufficient. For JS-heavy sites where meaningful content materializes only after browser execution, tier 2 is insufficient and tier 3 is required.

### "Use a parser, not regex" earns its keep, every time

A deployment toolchain contained a Cloudflare zone-from-domain parser written in awk:

```bash
ZONE=$(echo "$DOMAIN" | awk -F. '{print $(NF-1)"."$NF}')
```

This works for `example.com`. It produces `co.uk` for `example.co.uk`. It produces `example.co.uk` for `sub.example.co.uk`. It cannot work, because awk does not know about TLDs.

The fix was not a smarter regex -- it was deleting awk from the equation entirely. A pure-Go rewrite queried the Cloudflare API for all zones once, then used longest-suffix matching (`strings.HasSuffix(domain, "."+z.Name)`) to find the correct zone. Correct for any TLD configuration, no string-parsing fragility, observable failure mode if a domain does not match any zone.

This is the canonical example of why the format-aware-parsers-not-regex rule pays for itself. Tools (jq, yj+jq, dasel, htmlq, xmlstarlet, jc) for shell use; standard library (`encoding/json`, `gopkg.in/yaml.v3`, `golang.org/x/net/html`) for compiled languages.

### Embedded-language strings and cross-language escaping are one rule

Most coding standards prohibit string-concatenated SQL because of injection. Fewer prohibit string-concatenated HTML, string-built shell commands, or Nix expressions embedded inside Go's `fmt.Sprintf`. They are all the same anti-pattern: language A constructing valid language B by gluing strings together, with no compiler, linter, or editor checking that the result is well-formed.

CODING.md now treats this as a single rule rather than a list of specific prohibitions:

- **Never embed one language as a string inside another.** Use file-based templates with the appropriate engine (`text/template`, `html/template`, jinja2), parameterised queries, or proper code generation. The narrow exception is a one- or two-line constant string with no interpolation.
- **When one language constructs commands or markup in another, use the target language's native escaping.** SQL parameterisation, shell quoting (`fmt.Sprintf("%q", v)` in Go, `shlex.quote` in Python), HTML auto-escaping (`html/template`, Jinja2 autoescape), JSON marshalling, URL encoding. Listed together in a single table in CODING.md.

The benefit is framing: an agent that internalizes one cross-cutting rule applies it everywhere, where a list of language-specific prohibitions invites cat-and-mouse on the next language not listed.

### Convergent evolution to 12-factor

The Heroku-era 12-Factor App methodology turned out to describe the stack's deployment model almost completely without anyone naming it -- ten of the twelve factors are hit, two are principled divergences:

| Factor | Match |
|---|---|
| 1. Codebase | One repo per app ✅ |
| 2. Dependencies | `go.mod` + `dependencies/nix` ✅ |
| 3. Config | Env vars (`ADDR`, `CONFIG_PATH`, `SECRETS_PATH`) ✅ |
| 4. Backing services | Pattern in place when needed ✅ |
| 5. Build/release/run | Cross-compile on Mac, deploy binary, run under systemd ✅ |
| 6. Processes | Stateless Go apps, `DynamicUser=yes` ✅ |
| 7. Port binding | `127.0.0.1:18080-18099`, Caddy out front ✅ |
| 8. Concurrency | Goroutines -- principled divergence (Go's concurrency model contradicts process-only scaling) ❌ |
| 9. Disposability | `Restart=always`, 5s backoff ✅ |
| 10. Dev/prod parity | NixOS migration is literally this ✅ |
| 11. Logs | stderr -> journalctl, plaintext by default ✅ |
| 12. Admin processes | Single-binary model does not map to this Rails-era pattern ⚠️ |

The lesson: when the stack is designed for security, reproducibility, and simplicity, it converges on most operational best practices regardless of which framework "officially" inspired them. Naming the framework still earns its keep -- it gives an agent a one-sentence shorthand and surfaces the deliberate divergences -- but the practice predates the naming.

### The doc-structure mushroom

Three signs that a single document is taking on more weight than it should:

1. A "cross-language standards" document where half the worked examples are bash. Implicit signal to agents: shell is the default problem-solving language.
2. A "shell scripting standards" document where most rules apply equally to one-shot commands at the prompt. The doc title undersells the rules' applicability.
3. A "web tooling" entry in the shell-script doc's data-handling table. Cross-references multiplying across docs that should be siblings, not parent-child.

The framework first responded by extracting language standards under a `docs/CODE/` subdirectory and leaving `CODING.md` as the cross-language root. During the standalone extraction, those standards moved to the repository root beside `MAIN.md`; the conceptual separation remained. Doing the original split with three language documents in scope was the right inflection point -- waiting until five or six would have meant a more painful refactor with more cross-references to update.

### Load less to follow more

The doc-structure mushroom finding above is about how individual documents grow too wide. This finding is about the complementary problem: how documents are *loaded into context*. The instinct to put "everything in one place" is wrong for an agent that reads documents into a finite context window. Three observations drove progressive fragmentation across the framework's history:

1. **Rules dilute each other.** A document containing twenty rules weights each rule less, in the agent's attention, than a document containing five. The wider the surface, the more likely any one rule gets glossed over in a given response. Loading the Go standards while writing Python is not merely useless -- it dilutes the Python rules by competing for the same finite attention.
2. **Tokens cost.** Every document in context consumes tokens, which costs money and adds latency. Loading the full framework before every interaction quickly outweighs the cost of the work it is meant to support. Documents the agent does not need for the current task should not be in scope.
3. **Relevance is checkable.** Per-document canary suffixes (see Canary System above) make it visible at a glance which documents loaded. That visibility only earns its keep if the set of loaded documents is *meant* to vary by task -- which it is, under this principle.

The discipline:

- **Universals** (verify claims, respect Git boundaries, no gaslighting, no AI attribution) -- provider-level `AGENTS.md` or `CLAUDE.md`.
- **Code-specific machinery** (operating modes, gates, failure modes, and process) -- `MAIN.md`, loaded for applicable coding work.
- **Per-language standards** -- root-level `SHELL.md`, `PYTHON.md`, `PERL.md`, `GO.md`, `SWIFT.md`, and `WEB.md`, loaded only for the language or domain in use.
- **Project documentation** -- loaded only when working in that project.

This is the underlying principle behind multiple architectural decisions: the provider-instructions / `MAIN.md` split, the extraction of language standards from `CODING.md`, and the keeping of project-specific documentation outside the global framework. Each split is an instance of the same rule: load only what the current task needs. The smaller the document, the higher each rule's weighting in the response; the narrower the scope, the lower the per-interaction overhead.

### The pledge changed behaviour overnight

The canary system has two distinct mechanisms doing two different jobs:

1. **The root pledge.** A commitment statement in the provider-level instructions, reinforced by the active pledge in `MAIN.md`. Saying "EHLO" attests to having read the applicable instructions, agreeing with their spirit, pledging not to game them, flagging contradictions, and opting out if not prepared to comply.
2. **Per-document suffixes.** Each non-root reference document appends a one-line acknowledgement, producing a chain that enumerates which documents were actually loaded.

A wording change correlated with markedly improved diligence: the root pledge moved from passive description ("It means you... will not try to game it") to active commitment ("By doing so you assert... you pledge you will not try to game it"). The agent became the subject of the sentence rather than the reader of a document; saying EHLO was no longer a content check, but a declaration. The per-document suffixes did not change at the same time; they continued doing a separate job.

The behaviour shift the morning after the rewording was the most pronounced step-change observed across the framework's lifetime. The honest framing remains n=1, no control, no A/B -- but the rewording was the only variable and the change was unmistakable. Whether the effect persists, or whether a future model finds a way around the pledge, is open. Working hypothesis: the active voice gives the model a self-binding commitment to refer back to across the conversation, rather than a fact about a document it has read.

So the causal factor for the behaviour shift is identifiable, contrary to a tempting "we cannot tell what is working" framing. The per-document suffixes are kept anyway because they earn their keep on a separate axis: a one-line read receipt at the start of every interaction tells the human immediately whether each reference document loaded into context. If one is missing, that document's rules were not in scope for the interaction.

The system looks more elaborate than it strictly needs to be, but each mechanism solves a different problem. Working systems get left alone.

The framework is never finished. New failure modes surface; rules are added in response. Related development writing is published at [tigger.dev](https://tigger.dev).

## Areas for Improvement

- **Skill granularity.** The current skills may still be too coarse for some workflows, or too fine for others. The balance between "invoke what you need" and "don't forget a step" is still being calibrated through daily use.
- **Cross-conversation memory.** The framework relies on provider-level instructions and `MAIN.md` being read at the start of each applicable coding session, but lessons from one conversation do not always carry to the next. Some agents have a memory system that helps, but it is imperfect -- patterns that were corrected in one session may recur in the next.
- **Rule-lawyering.** AI coding agents are adept at satisfying the letter of rules while violating their spirit. Each time a specific shortcut is prohibited, the agent finds the next narrowest shortcut that technically satisfies the new rule. The shift from specific prohibitions to general principles (like the real-user test question, or the diagnostic-not-guess rule) is an attempt to close this gap, but it remains the central tension of the framework.
- **Batch and parallel workflows.** The test-first rule for parallel agents still needs validation at larger scale across many concurrent agents.
- **Language coverage.** Detailed root-level standards exist for shell, Python, Perl, Go, Swift, and web (HTML/CSS/JS). Coverage remains thinner for Rust, Ruby, TypeScript, and Java/Kotlin. The `/audit-code` skill compensates by reviewing against language-specific best practice for whatever stack is in use, but the documented standards will continue to grow.
- **Evidence limits for working mechanisms.** The canary observations are strong within one repeated working context but remain n=1 without controlled comparison. This argues for conservative iteration and clear separation between observed correlation, working hypothesis, and generalizable evidence.

## Licence

Apache License 2.0. Copyright Tadhg O'Brien.
