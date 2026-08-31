# Paired Development

Paired development supports work whose required outcome is refined through
live human review. It is especially suitable for visual, editorial, ergonomic,
and other experience-led changes where a detailed specification would obstruct
the feedback loop.

This is a code-only delivery track, not a conversational default or an
agent-wide operating mode. The operator must select it explicitly. Spec Kit
remains the default for autonomous or pre-specified delivery.

## Specification boundary

The bounded session objective and each explicit operator instruction define the
specification for that iteration. A question, suggestion, or request for an
opinion is not implementation authority.

The project constitution and applicable engineering standards remain in force.
The agent may resolve routine, reversible details that do not change the stated
outcome. A decision affecting product scope, architecture, security, privacy,
access, persisted data, an external contract, or an irreversible outcome must
be recorded in a durable specification or explicitly decided by the operator
before implementation.

Paired work may use artefact-native authority. For example, a Hugo site's
Markdown and YAML may embody accepted content and configuration, while its CSS,
templates, and rendered output may embody the design being reviewed. Do not
create a parallel prose description merely to restate those artefacts.

## Iteration loop

For each bounded iteration:

1. Identify the explicit requested outcome and applicable existing constraints.
2. Implement one reviewable slice without expanding its scope.
3. Run proportionate objective checks that add useful evidence.
4. Present the rendered or otherwise user-visible result.
5. Retain any explicit operator validation in a provisional validation ledger.
6. Continue from the operator's next instruction.

The agent must not manufacture automated tests merely to reproduce visual or
subjective validation the operator has already performed. Follow
`~/.agents/sdlc/TESTING.md` for evidence selection and the final user-test
record.

## Closure

At closure, present one consolidated handback containing:

- the changes made;
- every current operator validation, with superseded validations identified;
- objective checks run and their results;
- unvalidated behaviour and justified automation gaps;
- applicable change-scoped audit results; and
- durable specifications or project documents that require updating.

Ask the operator once whether the current validation ledger may be recorded as
the user tests for the change. Record it only after that confirmation. A later
change invalidates only the validations whose observed behaviour it materially
affects.

Do not merge, deploy, or describe the paired result as delivered unless the
operator has authorized that action and every applicable project requirement is
satisfied.
