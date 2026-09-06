# SDLC v1 Definitions Retained in v3

SDLC v3 deliberately retains a small set of definitions from the latest v1
release, `v1.0.3`. It does not restore the v1 ticket or gate ceremony.

| V3 concept | V1 source | Retained meaning |
|---|---|---|
| Acceptance criterion | `v1.0.3:ISSUES.md` | One falsifiable system state, not a test action or implementation mechanism |
| Regression test (RT) | `v1.0.3:TESTING.md` | Persistent automated evidence in the supported regression suite |
| User test (UT) | `v1.0.3:TESTING.md` | Human validation owned by the operator or named reviewer |
| One-off test (OT) | `v1.0.3:TESTING.md` | Bounded evidence that is recorded but not retained in regression automation |
| Automated TDD | `v1.0.3:TESTING.md` | Write the justified automated test first, observe RED, implement, and observe GREEN |

The v3 `ISSUES.md` and `TESTING.md` are the current normative texts. The v1 tag
is provenance for the resurrected terminology only.
