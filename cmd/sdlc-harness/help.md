sdlc-audit: request an independent SDLC audit.
Alias of sdlc-harness; both use the same interface.

Usage:
  sdlc-audit --gate GATE --audit-record FILE --work-item ID [options] < context.txt
  sdlc-audit --reset --gate GATE --audit-record FILE --work-item ID
  sdlc-audit goal-config [options]

Audit options:
  --gate GATE          Required: definition, test-code, delivery-code.
  --audit-record FILE  Harness-owned audits.yaml beside the specification.
  --work-item ID       Work item owning that record.
  --input FILE         Evidence file; repeat for every relevant input.
  --project DIR        Project root (default: current directory).
  --reset              Operator-authorized reset of gate counters and context; then exit.
  -h, --help           Show this reference.
  --version            Show installed version.

Use the same normal command for the first audit and every remediation.
The harness owns provider selection, context, fallback and bounded recovery.
On FAIL, address all findings before submitting the next audit.
On 'Human intervention required', stop auditing and report the diagnostic reference.
Do not troubleshoot the provider, inspect private diagnostics, or reset without authority.
After authorization, run --reset once without --input, then audit without --reset.
History and logs are preserved; the next audit uses a fresh context.

Each work item and gate has its own verdict allowance (default five).
Unusable responses and cached results do not consume that allowance.
Provider failures have a separate internal bound; no verdict is fabricated.

Test-code and delivery-code locate spec.org and validation.org beside --audit-record.
Schema/readiness rejection returns YAML without invoking a provider.
Test-code requires RTs RED/GREEN; delivery-code requires RTs/OTs GREEN.
UTs may remain AMBER. Test-code PASS is not delivery approval.
Definition has no execution-readiness check. With no approved RTs, skip test-code.
Final delivery review includes affected product documentation.

Example:
  sdlc-audit --gate delivery-code \
    --audit-record specs/WNNN-descriptor/audits.yaml --work-item WNNN \
    --input tests.go --input application.go
Include all relevant source, test and authority files; this example abbreviates inputs.

Goal-config options:
  --project DIR  --global-config FILE  Project and global configuration.
  --sdlc-root DIR                      Configuration schema root.
  --goal-max-turns N                   Continuation limit override.
  --goal-max-token-budget N            Token budget override.

Operator configuration and diagnostic details: ~/.agents/sdlc/HARNESS.md
