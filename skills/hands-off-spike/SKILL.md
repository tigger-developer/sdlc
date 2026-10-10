---
name: hands-off-spike
description: Run an explicitly invoked, unattended exploratory spike from a signed-off brief, with mandatory independent standalone code review and recorded findings, never merged or deployed.
---

Use this skill only when the operator explicitly invokes it. Paired development
(`pair-change`) is the attended human-agent spike; this skill is its
unattended counterpart. It answers a bounded question with evidence. It never
delivers product behaviour.

If `~/.agents/sdlc/MAIN.md` is absent or unreadable, report that exact path;
never search for another copy.

Read `~/.agents/sdlc/MAIN.md`, `CODING.md`, `TESTING.md`, `GIT.md`,
`DOCUMENTATION.md`, and `AUDITS.md`, then the project profile when present and the technology
standards it selects. Apply MAIN's **Delivery modes** HANDS-OFF rules and
`~/.agents/skills/deliver-change/references/native-goal.md` by reference; do not
restate them here.

## 1. Clarify the objective before anything else

Before creating a brief, branch, or work item:

- **No objective supplied:** ask the operator for it in ordinary prose. Do not
  offer a menu of options.
- **Objective supplied:** restate it as a draft brief and ask one consolidated
  set of questions covering everything unclear or missing: the question to
  answer, approaches, evaluation criteria, boundaries, budget, and stop
  conditions. If nothing is unclear, say so explicitly rather than asking for
  the sake of it.

This is the only question set. Every decision a human must make belongs here,
while the operator is present.

## 2. Record the brief and obtain sign-off

Allocate a never-used work number under the project's scheme. In an SDLC
project, add the item to `docs/work.org` with an adjacent descriptor and keep
the brief at `specs/WNNN-descriptor/spike.org`. Without a profile, use the
project's established work record; if none exists, ask where to keep the brief
in the step 1 question set.

The brief states:

- **Question:** the single bounded question the spike answers.
- **Approaches:** the candidates to try, or how they will be chosen.
- **Evaluation criteria:** falsifiable measures that distinguish the
  approaches, each with the evidence that will decide it.
- **Boundaries:** the spike branch, files in scope, and prohibited actions.
- **Budget:** native goal limits and any approved live OT spend under
  TESTING.md's hands-off OT budget.
- **Stop conditions:** when the spike ends without an answer.

Present the brief and wait for explicit sign-off. Record `Approved YYYY-MM-DD`
in the brief, following DOCUMENTATION.md. Material changes to the question,
criteria, or boundaries require renewed sign-off.

## 3. Work unattended

- Create `spike/WNNN-descriptor` from the operator-selected branch. Commit
  coherent checkpoints there. Synchronize it only as GIT.md permits through an
  existing remote. **Never merge it, rebase another branch onto it, or deploy
  from it.**
- Declare `DELIVERY MODE: HANDS-OFF`, activate native goal continuation under
  `native-goal.md`, and continue until a stop condition, the budget, or a
  permitted HANDS-OFF stopping condition is reached.
- Record routine reversible assumptions in the brief's `Hands-off mode
  decision log`, as MAIN defines it.
- Record each experiment as an OT in the brief's Evidence section: action,
  revision, observation, and the criterion it informs. Follow TESTING.md for
  bounded live checks. A check requiring `sudo` or human judgement is a UT:
  leave it AMBER for the operator.
- Spike code meets all applicable coding and TDD/testing standards, without
  entering full SDLC delivery, gates or documentation requirements. Write an
  automated test first wherever it is the evidence for a criterion.
- Stay inside the question. Record new questions as findings, not new work.

## 4. Audit the spike code

Code produced by a spike is always independently reviewed. The harness's
`SPIKE-AUDIT.md` profile supplies all coding, security, technology and testing
standards; it does not require SDLC workflow or documentation artefacts.
Before handback, invoke the resolved absolute `~/.agents/sdlc/bin/sdlc-harness` path with
`--agent-spike`, every spike code and test file, and the brief as `--input`.
Supply `Workflow: hands-off spike` and the brief's question on stdin. Do not
pass `--gate`, `--audit-record`, `--work-item`, or installed SDLC paths.

Address every finding as a batch: remediate valid findings and challenge
incorrect ones with evidence under AUDITS.md, then resubmit the same command.
A standalone review has no harness verdict budget; stop after five verdicts
without `PASS` or `PROVISIONAL PASS` unless the operator authorized more. On
"Human intervention required", report the diagnostic reference and stop
auditing.

## 5. Record findings and hand back

Add a Findings section to the brief:

- the answer to the question, criterion by criterion, with evidence links;
- the recommendation and its main risk;
- approaches not tried, and why;
- the spike branch and checkpoint commits, each with a descriptor;
- the final standalone audit verdict and its findings, quoted. A standalone
  review writes no `audits.yaml`; the spike is deliberately outside the SDLC
  gate record, so its findings section is the record.

Return to ATTENDED and present the brief. The operator decides:

- **Adopt:** define the change through `define-change`, with the findings as
  authority. Spike code is evidence, not an approved implementation.
- **Discard:** close the work item with its findings and keep the branch as
  history.

Do not claim closure, merge, or deploy.
