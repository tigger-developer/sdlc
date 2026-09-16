sdlc-audit: run or resume an independent SDLC audit.
Alias of sdlc-harness; both use the same interface.

Usage:
  sdlc-audit start|resume [options] < context.txt
  sdlc-audit resume --reset-session --gate GATE --audit-record FILE --work-item ID
  sdlc-audit goal-config [options]    Print resolved goal limits as JSON.

Audit options:
  --gate definition|implementation   Required for audits.
  --readiness-check CHECK            Implementation: test-code or delivery-code
                                    (default: delivery-code).
  --audit-record FILE                Harness-owned audits.yaml record.
  --work-item ID                     Work item owning that record.
  --input FILE                       Evidence file; repeat for each input.
  --project DIR                      Project root (default: current directory).
  --phase audit|definition|build     Configuration phase (default: audit).
  --global-config FILE               Override ~/.agents/sdlc.yaml.
  --harness NAME                     Override configured harness.
  --provider NAME                    Override provider where supported.
  --model NAME                       Override configured model.
  --timeout DURATION                 Override timeout, e.g. 8m.
  --audit-prompts FILE               Override installed audit-prompt YAML.
  --session ID                       Resume native ID; normally read from record.
  --reset-session                    Reset selected budget only, then exit.

Goal-config options:
  --project DIR  --global-config FILE Same meanings as above.
  --sdlc-root DIR                    Root containing the config schema.
  --goal-max-turns N                 Override continuation limit.
  --goal-max-token-budget N          Override token budget (e.g. "100,000").

  -h, --help                        Show this reference.
  --version                         Show installed version (top-level only).

Start creates a session; resume loads it from the audit record. Do not pass
--session to start. Reset requires operator authorization: omit --input and
--session, then rerun the audit WITHOUT --reset-session. History is preserved.
stdout: final result. stderr: progress and diagnostics.

Implementation requires --audit-record. Its sibling spec.org and validation.org
are schema/readiness-checked before any provider call, then included as evidence.
Refusal: YAML on stdout; exit 1 ineligible or 2 invalid schema; no round consumed.
Run sdlc-validate --help for local preflight. Test-code and delivery-code reviews
share the implementation session and budget. A test-code PASS is not delivery PASS.

  sdlc-audit start --gate implementation --readiness-check test-code \
    --audit-record specs/007-example/audits.yaml --work-item W007 --input tests.go
  sdlc-audit resume --gate implementation --readiness-check delivery-code \
    --audit-record specs/007-example/audits.yaml --work-item W007 \
    --input tests.go --input application.go

Include all relevant source, test and authority files; examples abbreviate inputs.

Examples, caching, fallback and recovery: ~/.agents/sdlc/HARNESS.md
