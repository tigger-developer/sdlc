# Application Security Standards

These standards define the application-side security contract. They complement
language standards and any external infrastructure contract; they do not claim
host, network, operating-system, or fleet responsibilities.

## Vulnerability-check interface

Every deployable application must expose a repository-owned `make vulncheck`
target. This is a stable security-gate interface for developers, continuous
integration, and deployment systems. It is not part of `make test`.

The target must:

- scan every application-managed dependency set that affects the built or
  deployed artefact, including build dependencies capable of changing it;
- use the stack-specific scanner selected by the applicable technology
  standards;
- inspect committed locks, resolved dependencies, or the actual build artefact
  rather than an unconstrained newly resolved dependency graph;
- make no dependency, lockfile, source, configuration, or environment changes;
- never install a scanner, update packages, apply fixes, or rewrite an exception;
- fail when a required scanner is absent, advisory data cannot be obtained or
  validated, an expected manifest is stale, or any applicable scan is incomplete;
- return non-zero for every non-exempt vulnerability reported by the selected
  scanner; and
- give concise evidence identifying each scanner, input, and result.

A mixed-stack project must run every applicable scan from the same target. A
project with no application-managed dependencies may use a repository-owned
check that verifies this invariant and fails if an uncovered dependency manifest
appears. An unconditional successful no-op is prohibited.

The target is a repeatable security control, not a behavioural regression test.
It may run in a dedicated continuous-integration or deployment step even though
it is excluded from `make test`.

## Scanner selection

Prefer the ecosystem scanner when it provides reliable coverage. Use Trivy as
the fallback for a supported lockfile or artefact when the ecosystem lacks a
strong native scanner. Use OSV-Scanner when its supported manifest coverage is
a better fit. Do not add a second scanner merely to produce duplicate findings.

| Stack | Preferred scanner |
|---|---|
| Go | `govulncheck` |
| Python | `pip-audit` through the project-managed environment |
| Node.js and npm-managed web assets | The selected package manager's audit command |
| Perl | `cpan-audit` |
| Swift Package Manager or CocoaPods | Trivy |
| Other supported lockfiles or artefacts | Trivy or OSV-Scanner, with the choice recorded |

Scanner commands and thresholds belong in the applicable technology standard
or project design. Pin or reproducibly provision scanner versions; do not fetch
an unversioned executable during the security gate.

## Findings and exceptions

A scanner finding blocks the target by default. Never suppress an exit code,
hide a finding, lower a threshold, or exclude a dependency merely to obtain a
successful deployment.

A false positive or accepted exposure requires a durable exception containing:

- the vulnerability identifier and affected component;
- whether the vulnerable code or artefact is reachable or deployed;
- the evidence and rationale for accepting the exposure;
- the human authority that approved it;
- compensating controls;
- an expiry or removal condition; and
- the scanner configuration that applies the exact exception.

Broad package, severity, path, or scanner exclusions are prohibited. Review an
exception whenever the dependency, application reachability, deployment model,
or advisory changes.

## Build and deployment boundary

Application scanning does not establish host or infrastructure safety. The
application owns its dependency manifests, build inputs, generated artefacts,
and `vulncheck` implementation. An external infrastructure owner may require and
invoke that target, then separately assess the deployed operating-system,
service, container, package-closure, routing, and host controls.

Run the application gate before deployment. A later vulnerability disclosure
can make an unchanged deployed version unsafe, so projects should also use a
proportionate periodic invocation or advisory-monitoring mechanism. Reuse the
same target and exception policy rather than creating a second security process.

## References

- [Go vulnerability management](https://go.dev/doc/security/vuln/)
- [pip-audit](https://github.com/pypa/pip-audit)
- [npm audit](https://docs.npmjs.com/cli/v11/commands/npm-audit/)
- [CPAN Audit](https://metacpan.org/pod/cpan-audit)
- [Trivy vulnerability scanning](https://trivy.dev/docs/latest/scanner/vulnerability/)
- [OSV-Scanner](https://google.github.io/osv-scanner/)
- [OWASP vulnerable dependency management](https://cheatsheetseries.owasp.org/cheatsheets/Vulnerable_Dependency_Management_Cheat_Sheet.html)
