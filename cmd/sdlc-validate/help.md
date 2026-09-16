sdlc-validate: check Org schema and audit readiness without invoking an agent.
Read-only. Never runs tests, modifies documents, or consumes audit rounds.

Usage:
  sdlc-validate --readiness-check=test-code|delivery-code [options]

Options:
  --spec FILE              Specification (default: ./spec.org).
  --validation FILE        Results (default: validation.org beside the spec).
  --readiness-check CHECK  Required: test-code or delivery-code.
  -h, --help               Show this help.
  --version                Show release version.

Checks:
  Both: validate headings, tags, unique IDs and declared execution states.
  test-code: every specified test has RED, AMBER or GREEN and state evidence.
  delivery-code: additionally require every RT and OT GREEN. UTs may be AMBER.
  A missing state is schema-valid but never audit-ready. Schema errors prevent
  readiness evaluation. Unsupported checks (including test-spec) are errors.

stdout: YAML under audit-readiness, with schema, selected check and inventory.
Includes AC/test counts, IDs, titles, states and diagnostics with file/line.
Exit: 0 ready; 1 ineligible; 2 invalid schema; 3 invocation/read/write error.

Examples:
  sdlc-validate --spec specs/007-example/spec.org --readiness-check=test-code
  sdlc-validate --spec specs/007-example/spec.org --readiness-check=delivery-code

The validator checks recorded evidence, not its truth or behavioural adequacy.
Org record contract: ~/.agents/sdlc/ORG-SCHEMA.md
