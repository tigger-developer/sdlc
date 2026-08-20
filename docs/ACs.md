# Acceptance Criteria

This is the canonical spec. ACs introduced from 2026-08-20 onward live here.
Pre-cutover ACs remain in their originating issues until cited or migrated.

Last migrated: AC4.6 from #4 on 2026-08-20

---

## Shared SDLC deployment

### AC4.1 - Given drift between staging and the shared live SDLC tree, interactive installation presents one shared-deployment confirmation; declining leaves the live tree unchanged and accepting synchronizes only `~/.agents/sdlc`.

- Introduced: #4 (closed 2026-08-20)
- Migrated: 2026-08-20
- Tests:
  - ✅ RT-4.1: Shared drift produces one confirmation
  - ✅ RT-4.2: Declining shared deployment leaves the live tree byte-for-byte unchanged
  - ✅ RT-4.3: Accepting shared deployment synchronizes the owned subtree and excludes root Git metadata

### AC4.2 - Given any subset of known provider homes, interactive installation recognizes exactly those providers and offers confirmation only for adapters that differ from the expected live targets.

- Introduced: #4 (closed 2026-08-20)
- Migrated: 2026-08-20
- Tests:
  - ✅ RT-4.4: Provider detection follows the available provider-home subset
  - ✅ RT-4.5: Current provider adapters are reported without a confirmation
  - ✅ RT-4.6: Drifted or absent provider adapters receive individual confirmations

### AC4.3 - Given mixed provider responses, only accepted provider adapters change; declined adapters and provider configuration files remain unchanged.

- Introduced: #4 (closed 2026-08-20)
- Migrated: 2026-08-20
- Tests:
  - ✅ RT-4.7: Selective acceptance changes only accepted provider links
  - ✅ RT-4.8: Declining every provider leaves every provider home unchanged
  - ✅ RT-4.9: Interactive adapter installation does not alter provider configuration

### AC4.4 - Given the top-level SDLC skill directories, every skill has its exact common live link and every accepted supported provider has its expected skill links, while unrelated common and provider entries remain unchanged.

- Introduced: #4 (closed 2026-08-20)
- Migrated: 2026-08-20
- Tests:
  - ✅ RT-4.10: Source-driven iteration covers every common SDLC skill link
  - ✅ RT-4.11: Provider-table iteration covers every expected provider skill link
  - ✅ RT-4.12: Unrelated common and provider entries survive installation

### AC4.5 - Given a live deployment matching staging and current provider adapters, repeated interactive installation has no filesystem effects and offers no redundant confirmations.

- Introduced: #4 (closed 2026-08-20)
- Migrated: 2026-08-20
- Tests:
  - ✅ RT-4.13: A second installation leaves filesystem state unchanged
  - ✅ RT-4.14: A second analysis reports the shared tree and provider adapters current

### AC4.6 - Given the staging repository, `make install` starts the interactive installer and propagates its success, refusal, or failure status.

- Introduced: #4 (closed 2026-08-20)
- Migrated: 2026-08-20
- Tests:
  - ✅ OT-4.1: The Makefile entry point starts the built interactive installer
  - ✅ OT-4.2: The Makefile entry point preserves installer exit status

**Key:** ✅ passing · ⏳ pending · ❌ failing · ~~🚫 removed~~
