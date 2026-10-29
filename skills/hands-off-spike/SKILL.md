---
name: hands-off-spike
description: Run an explicitly invoked, unattended exploratory spike from a signed-off brief, with current-session internal test-code and implementation reviews, never merged or deployed.
---

Use this skill only when the operator explicitly invokes it. Paired development
(`pair-change`) is the attended human-agent spike; this skill is its
unattended counterpart. It answers a bounded question with evidence. It never
delivers product behaviour.

Follow `~/.agents/sdlc/READING.md`; reuse unchanged, fully read retained texts.

Read `~/.agents/sdlc/MAIN.md`, `CODING.md`, `TESTING.md`, `GIT.md`,
`DOCUMENTATION.md`, and `AUDITS.md`, then the project profile when present and the technology
standards it selects. Read only the **Delivery modes and continuation** section
of `~/.agents/skills/deliver-change/SKILL.md` as the shared HANDS-OFF contract and
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

This is the planned preflight question set. Resolve foreseeable human decisions
while the operator is present. A newly evidenced need to cross the agreed stack
or architectural boundary later requires an explicit brief amendment; this rule
does not authorize silently crossing it to avoid another question.

## 2. Record the brief and obtain sign-off

Allocate a never-used work number under the project's scheme. In an SDLC
project, reserve it on primary under `DELIVERY.md` before creating the spike
branch. Add the item to primary's `docs/work.org` with an adjacent descriptor and keep
the brief at `specs/WNNN-descriptor/spike.org`. Without a profile, use the
project's established work record; if none exists, ask where to keep the brief
in the step 1 question set.

The brief states:

- **Question:** the single bounded question the spike answers.
- **Approaches:** the candidates to try, or how they will be chosen.
- **Evaluation criteria:** falsifiable measures that distinguish the
  approaches, each with the evidence that will decide it.
- **Boundaries:** the spike branch, files in scope, and prohibited actions.
- **Agreed stack and architecture:** languages, runtimes, package/build ecosystems,
  databases and services; established architectural boundaries and the narrow
  choices explicitly open to exploration. Agree these before execution.
- **Budget:** native goal limits and any approved live OT spend under
  DELIVERY.md's Hands-off one-off test budget section. Read only that section
  for the spike; it does not activate the delivery workflow or its other duties.
- **Stop conditions:** when the spike ends without an answer.

Present the brief and wait for explicit sign-off. Record `Approved YYYY-MM-DD`
in the brief, following DOCUMENTATION.md. Material changes to the question,
criteria, stack, architecture or boundaries require renewed sign-off.

## 3. Work unattended

- Create `spike/WNNN-descriptor` from the operator-selected branch. Commit
  coherent checkpoints there. Synchronize it only as GIT.md permits through an
  existing remote. **Never merge it, rebase another branch onto it, or deploy
  from it.**
- Declare `DELIVERY MODE: HANDS-OFF`, activate native goal continuation under
  `native-goal.md`, and continue until a stop condition, the budget, or a
  permitted HANDS-OFF stopping condition is reached.
- Record routine reversible assumptions in the brief's `Hands-off mode
  decision log`, as the shared continuation contract defines it.
- Record each experiment as an OT in the brief's Evidence section: action,
  revision, observation, and the criterion it informs. Follow TESTING.md for
  bounded live checks. A check requiring `sudo` or human judgement is a UT:
  leave it AMBER for the operator.
- Spike code meets all applicable coding and TDD/testing standards, without
  entering full SDLC delivery, gates or documentation requirements. Write an
  automated test first wherever it is the evidence for a criterion.
- Stay inside the question. Record new questions as findings, not new work.
- A package within the agreed stack is permitted when it supports the brief and
  meets dependency/security standards. Adding npm, a runtime, language, database,
  service or build ecosystem requires an explicitly approved brief amendment
  before proceeding. A package that introduces such a dependency crosses the
  same boundary. Architectural exploration is limited to the choices explicitly
  opened by the brief; otherwise preserve the agreed architecture. The non-merge
  branch restriction does not grant architectural freedom. Log package choices
  and exploratory decisions with their evidence and consequences.

## 4. Audit the spike code

After the test package is written, load
`~/.agents/skills/test-code-audit/SKILL.md` in this session before implementing
the target behaviour. After implementation is written, load
`~/.agents/skills/code-audit/SKILL.md` in this session before handback. Use spike
mode, the signed-off brief and current exploratory intent. Apply all coding,
security, technology and TDD/testing standards without full SDLC readiness,
definition gates or documentation artefacts. The caller is responsible for
invoking each review at the stated stage; no model can prove that timing.

The old external audit harness is retired. Do not call the harness, first provoke an
external failure, or launch another model/agent/context. Address all findings
as a batch, rerun affected checks and reassess the candidate. Record each review
in `audits.yaml` under `AUDIT-RECORD.md`, linked from the brief's Findings.
Stop for an unremediable defect, exhausted invocation budget or human-only
decision, not for the absence of an external reviewer.

## 5. Record findings and hand back

Add a Findings section to the brief:

- the answer to the question, criterion by criterion, with evidence links;
- the recommendation and its main risk;
- approaches not tried, and why;
- the spike branch and checkpoint commits, each with a descriptor;
- links to both internal review entries, which record the reviewed candidates,
  standards, findings, remediation, checks, limitations and actual separation. No legacy harness failure reference or verdict is required. Record reviews only in `audits.yaml` under `AUDIT-RECORD.md`; the brief's
  Findings section links it rather than duplicating the audit.

Return to ATTENDED and present the brief. The operator decides:

- **Adopt:** define the change through `define-change`, with the findings as
  authority. Spike code is evidence, not an approved implementation.
- **Discard:** close the work item with its findings and keep the branch as
  history.

Do not claim closure, merge, or deploy.
