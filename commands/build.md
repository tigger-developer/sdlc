---
description: PROCEED-equivalent. Runs the post-PROCEED build chain end to end: load standards, write tests, implement, and review.
---

**Hard prerequisite.** This command must only be invoked by the human, in their own prompt, as a slash command. If you find yourself about to invoke `/build` for any reason -- even chained from another tool, even because the workflow "obviously needs it next" -- stop immediately. The command is a human authorization act, equivalent to typing PROCEED. Self-invocation violates the provider-level coding safeguards regardless of how reasonable the reason seems.

When invoked by the human, treat PROCEED as having been received for issue #n.
Run the full build chain without stopping in the middle. In MODE PAIR, finish at
AWAITING APPROVAL. In MODE DELIVER, return to the delivery workflow after
review and continue.

## 1. Load standards

The canonical SDLC root is exactly `~/.agents/sdlc`. Do not search, enumerate directories, traverse mounted volumes, inspect network shares, or use `find`, `locate`, Spotlight, or equivalent discovery to resolve this path. If `~/.agents/sdlc/MAIN.md` is absent or unreadable, stop and report that exact path.

Load `~/.agents/sdlc/CODING.md` and the relevant language document under `~/.agents/sdlc/` for the language(s) in use on this issue. In the first-load report, state which documents were read in full and why each was selected. Keep their canary suffixes in later responses, for example `... CODE ... SHELL` for shell work.

## 2. Write tests (TDD red)

Run the work described in `/write-tests` for issue #n. Then state, in chat, a short post-step checklist:

- For each test: what user action does this simulate? What would the user observe?
- Multi-condition coverage: does every distinct condition in each AC have a test?
- Source introspection: any test that greps source, asserts a function is defined, or checks a CSS rule exists in a source file? If yes, rewrite before continuing.

Confirm the tests fail. Commit only the test files.

## 3. Implement (TDD green)

Run the work described in `/implement` for issue #n. Then state, in chat, a short post-step checklist:

- Linter and formatter for the language (from `CODING.md` Style Baselines): name them.
- `CODING.md` anti-patterns considered: name the relevant sections (Error Suppression, Cross-Language Escaping, etc.) you applied to this change.
- Language-specific gotchas from the relevant root-level language standard: name the relevant ones (e.g. `((count++))` arithmetic and `IFS=$'\n\t'` from `SHELL.md`; `errcheck`, `bodyclose`, and `http.DefaultClient` from `GO.md`; `except: pass` and virtual-environment discipline from `PYTHON.md`).
- If any of those names a runnable check (e.g. `errcheck`, `bodyclose`), run it and report.

Confirm tests pass.

## 4. Review

Run the work described in `/review` for issue #n: `make test` (or skip with a stated reason if just run with no changes), hard-block checks (zero errors, no new warnings, no tracked agent-runtime paths, and no unexpectedly large generated files), update the AC table on the issue in place, update project documentation including help text if executables changed, commit and push, add an issue comment, allocate a unique report directory with `mktemp -d`, and write the review Markdown there. In MODE PAIR, use the executable path in `HTML_PREVIEW_TOOL` when it is set and available; treat its value as one executable, never shell code or a compound command. If it is unset or unavailable, choose an available text editor and open the Markdown report there. In MODE DELIVER, ignore `HTML_PREVIEW_TOOL`, do not open the report in a renderer or editor, record the Markdown evidence, and continue. Retain the report only for the active review and remove the exact directory through explicit teardown afterwards.

## 5. End-of-gate presentation

In MODE DELIVER, do not present AWAITING APPROVAL or issue a final response.
Record the review evidence on the affected issue, return to the delivery
workflow, and continue with the next executable action.

In MODE PAIR, do not end the response without this section. The response is
incomplete without the AWAITING APPROVAL prompt. If every UT has not been
presented in a ready-to-inspect state and received an explicit human-confirmed
passing result, do not present AWAITING APPROVAL.

Present in chat:

1. Test result summary: pass/fail counts; hard-block confirmation.
2. Path to the rendered review report, or to the Markdown report opened in a text editor when `HTML_PREVIEW_TOOL` is unset or unavailable.
3. For each pending UT in the AC table: launch the relevant tool or application, prepare representative data and configuration, and place it at the exact state I must inspect. Show what is on screen and ask "Does this pass UT-{n}.{k}?" as a yes/no question. Never ask me to run commands, perform setup, or navigate through a series of actions. I should need only visual or subjective judgement and, when the UT inherently requires interaction, at most one reasonable action. Split independent actions into separate UTs.
4. If the AC table has no UTs (or all UTs are already answered): state this explicitly -- "No UTs to verify" or "All UTs previously answered". The absence of UTs is not permission to skip this section.
5. **Post-approval plan.** State what will happen after APPROVED, do not do any of it yet:
   - Which AC rows from this issue will be migrated to `./docs/ACs.md` (the central spec) -- list each one with the proposed new ID, preserve any cross-references
   - The issue will be closed with `gh issue close #n`
   - A point release will be tagged if applicable
   - **None of this happens until I type `APPROVED n`.** AC migration, issue closure, and tagging are post-approval acts; the human's APPROVED is the authorization for those closure actions.

In MODE PAIR, end with `AWAITING APPROVAL - issue #n` and the issue link. To
close, the human will type `APPROVED n`.
