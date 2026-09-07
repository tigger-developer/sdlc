# Paired Development

Paired development supports work whose required outcome is refined through
live human review. It is especially suitable for visual, editorial, ergonomic,
and other experience-led changes where a detailed specification would obstruct
the feedback loop.

This is a code-only delivery track, not a conversational default or an
agent-wide operating mode. The operator must select it explicitly. The normal
two-phase workflow remains the default for autonomous or pre-specified delivery.

## Delivery flow

```text
Bounded objective
  -> explicit iteration instruction
  -> reviewable implementation slice
  -> objective checks and operator validation
  -> repeat as directed
  -> consolidate the durable specification and validation record
  -> run applicable change-scoped audits
  -> operator closure decision
```

## Specification boundary

The bounded session objective and each explicit operator instruction define the
specification for that iteration. A question, suggestion, or request for an
opinion is not implementation authority.

The working specification therefore precedes every implementation slice. The
durable specification may be consolidated after the paired implementation from
the final accepted behaviour, evidence, and design. That consolidation records
what the operator authorized and validated; it does not grant permission
retrospectively or introduce unapproved behaviour.

The project profile, named architecture authorities, and applicable engineering
standards remain in force.
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
2. Select proportionate evidence without manufacturing automation.
3. Implement one reviewable slice without expanding its scope.
4. Run proportionate objective checks that add useful evidence.
5. Present the rendered or otherwise user-visible result.
6. Retain any explicit operator validation in a provisional validation ledger.
7. Continue from the operator's next instruction.

The agent must not manufacture automated tests merely to reproduce visual or
subjective validation the operator has already performed. Follow
`~/.agents/sdlc/TESTING.md` for evidence selection and the final user-test
record.

## Closure

When paired work establishes or changes durable product behaviour, consolidate
it into one durable specification before closure. Use the normal unified
specification structure: an accurate summary, described requirements, traced
test definitions, and a concise solution design. Retain executed results in the
validation record. Describe the accepted final state rather than reconstructing
an imagined advance plan. Reference artefact-native authority instead of
duplicating CSS, templates, Markdown, YAML, or other reviewed artefacts in
prose.

Do not route the completed paired change through the normal pre-implementation
definition checkpoint retrospectively. Run the focused audits required by
`~/.agents/sdlc/AUDITS.md`: a change-scoped `audit-code` for material code or
equivalent artefact changes, and the other specialist audits only when the work
creates their corresponding durable specification, design, or test material.

At closure, present one consolidated handback containing:

- the changes made;
- every current operator validation, with superseded validations identified;
- objective checks run and their results;
- unvalidated behaviour and justified automation gaps;
- applicable change-scoped audit results; and
- durable specifications or project documents that require updating.

Ask the operator once whether the durable specification accurately captures the
accepted outcome and whether the current validation ledger may be recorded as
the user tests for the change. Record the validation only after that
confirmation. A later change invalidates only the validations whose observed
behaviour it materially affects.

Do not merge, deploy, or describe the paired result as delivered unless the
operator has authorized that action and every applicable project requirement is
satisfied.
# Canary

Suffix the canary string with "PAIR "
