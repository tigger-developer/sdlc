sdlc-audit: request an independent SDLC or standalone standards audit.
Alias of sdlc-harness; both use the same interface.

Usage:
  sdlc-audit --input FILE [--input FILE ...] [options] < context.txt
  sdlc-audit --gate GATE --audit-record FILE --work-item ID [options] < context.txt
  sdlc-audit --reset --gate GATE --audit-record FILE --work-item ID
  sdlc-audit goal-config [options]

Audit options:
  --gate GATE          SDLC projects: definition, test-code, delivery-code.
  --audit-record FILE  Harness-owned audits.yaml beside the specification.
  --work-item ID       Work item owning that record.
  --input FILE         Project evidence file; repeat for every relevant input.
  --project DIR        Audit root (default: invocation Git root or current directory).
  --reset              Operator-authorized reset of gate counters and context; then exit.
  --agent-spike        Hands-off spike: standalone review even in an SDLC project.
  -h, --help           Show this reference.
  --version            Show installed version.

When the selected project has no .sdlc/project.yaml, omit --gate,
--audit-record and --work-item. The harness warns STANDALONE MODE and reviews
the supplied original project files against CODING.md, DOCUMENTATION.md and
STANDALONE-AUDIT.md, with other applicable installed standards available.
This review has no SDLC gate or persistent audit record.
In an SDLC project, --agent-spike selects the same standalone review for a
hands-off spike and warns SPIKE MODE. It accepts the same options and rejects
--gate, --audit-record, --work-item, --reset and any phase other than audit.
Audits are anchored to the invocation project. --project cannot select another
root; --input accepts only files inside that project tree. Do not pass
~/.agents/sdlc/MAIN.md, ARCHITECTURE.md, DOCUMENTATION.md or other installed
SDLC standards: the harness supplies its full installed Markdown inventory,
including mandatory gate rules and selected technology standards, automatically.
External --input paths are rejected.
Project evidence and audit records must stay inside that tree. Provider
runtime directories and tmp/temp/temporary directory variants are excluded;
ordinary hidden project directories are allowed. System temporary directories
and symlink aliases are refused before a provider starts. A bounded project
file inventory exposes omitted paths but does not prove their contents were read.

Use the same normal command for the first audit and every remediation.
The harness owns provider selection, context, fallback and bounded recovery.
On FAIL, address all findings before submitting the next audit.
On 'Human intervention required', stop auditing and report the diagnostic reference.
Do not troubleshoot the provider, inspect private diagnostics, or reset without authority.
After authorization, run --reset once without --input, then audit without --reset.
History and logs are preserved; the next audit uses a fresh context.
Reset accepts running state; no process check is required. Old invocations lose
write authority and cannot overwrite the reset or a new audit with late results.
Other projects are unaffected; the selected gate retains its existing context scope.

Each work item and gate has its own verdict allowance (default five).
Unusable responses and cached results do not consume that allowance.
Provider failures have a separate internal bound; no verdict is fabricated.

Test-code and delivery-code locate spec.org and validation.org beside --audit-record.
Schema/readiness rejection returns YAML without invoking a provider.
Test-code requires RTs RED/GREEN; delivery-code requires RTs/OTs GREEN.
UTs may remain AMBER. Test-code PASS is not delivery approval.
Definition has no execution-readiness check. With no approved RTs, skip test-code.
Final delivery review includes affected product documentation.

Do not start delivery-code review until all code, tests and documentation are
written, all required automated and one-off checks (including applicable lint
and vulnerability checks) have finished and passed, and validation.org records
settled GREEN RT/OT evidence. Do not run those checks in parallel with the audit
or record GREEN while checks are still running. Outstanding UTs may remain AMBER.

Audit inputs are frozen when review starts. Any change to reviewed code, tests,
documentation or validation records during or after review immediately invalidates
that audit for the changed candidate. Rerun affected checks, settle the evidence,
then submit the changed candidate for review again. The harness detects changed
inputs when the provider returns; it does not immediately cancel a running provider.
A previous verdict remains historical evidence only for its unchanged candidate.
Every provider start emits a reminder on stderr; heartbeat notices stay brief.

Example:
  sdlc-audit --gate delivery-code \
    --audit-record specs/WNNN-descriptor/audits.yaml --work-item WNNN \
    --input tests.go --input application.go
Include all relevant project source, test and authority files; this example
abbreviates inputs. Do not add paths under ~/.agents/sdlc.

Goal-config options:
  --project DIR  --global-config FILE  Project and global configuration.
  --sdlc-root DIR                      Configuration schema root.
  --goal-max-turns N                   Continuation limit override.
  --goal-max-token-budget N            Token budget override.

Operator configuration and diagnostic details: ~/.agents/sdlc/HARNESS.md
