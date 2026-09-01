# Hugo Standards

These standards apply to Hugo sites and supplement
`~/.agents/sdlc/technologies/WEB.md`. Hugo projects select both documents.

## Architecture

Prefer Hugo as a static-site generator. Build one canonical public artefact and
serve that artefact through every deployment adapter. Do not let GitHub Pages,
a VPS deployment, or another target acquire different content, rendering, or
asset behaviour unless the project specification explicitly defines the
difference.

Do not install Hugo or its source tree on a static serving host merely to serve
the generated site. The build toolchain belongs in a controlled build
environment unless the deployment design establishes another requirement.

Keep content, layouts, assets, data, and generated output in their established
Hugo locations. Do not reorganize a mature site merely to match an example
layout.

## Versions, themes, and modules

- Pin the Hugo version in reproducible project or deployment configuration.
- Treat themes, modules, build hooks, and asset-pipeline packages as executable
  supply-chain dependencies.
- Pin theme and module revisions through committed dependency metadata. Do not
  deploy an unpinned branch or moving remote reference.
- Commit `go.mod` and `go.sum` when Hugo Modules are used.
- Prefer a small project-owned theme layer over a large third-party theme when
  the project needs only limited behaviour.
- Record why a new remote theme, module, shortcode pack, or asset tool is needed
  and review its maintenance and release history before adoption.

## Build security

- Keep Goldmark unsafe rendering disabled unless the specification identifies
  the trusted content boundary and reason it is required.
- Treat shortcodes and templates that emit raw HTML, JavaScript, URLs, or CSS as
  code. Preserve contextual escaping and do not mark untrusted values safe.
- Avoid build-time network access. Pin and validate any remote resource that the
  build must retrieve; an unversioned remote response is not a reproducible
  input.
- Do not expose environment values, filesystem paths, Git credentials, drafts,
  private data, or unpublished resources in the generated output.
- Disable development servers, live reload, debug output, and source maps in the
  deployed artefact unless explicitly required.
- Fail production builds on material template, content, or configuration
  warnings. Do not suppress warnings merely to deploy.

## Testing

Follow the rendered-output and paired-development model in
`~/.agents/sdlc/technologies/WEB.md`.

- Validate the generated site rather than grepping Hugo source templates.
- Use `htmltest` for broad generated-site, route, link, and asset validation.
- Use `htmlq` or `xmlstarlet` for targeted assertions against generated HTML or
  feeds.
- Use browser automation only for behaviour that genuinely requires execution
  or computed presentation; retain human validation for visual judgement.
- Ensure each deployment adapter consumes the same validated public artefact.

## Vulnerability checking

Expose the `make vulncheck` contract defined by
`~/.agents/sdlc/SECURITY.md`.

- Scan Hugo Module metadata with OSV-Scanner or another documented scanner that
  supports the committed Go module graph.
- Apply `~/.agents/sdlc/technologies/NODE.md` when npm or another Node package
  manager participates in the build, and include that dependency graph.
- Treat the Hugo executable as part of the build environment. Where the project
  owns that binary, scan it or its resolved package closure; otherwise record
  the external build owner's responsibility.
- A site with no modules or package-managed build dependencies must verify that
  invariant rather than supplying an unconditional successful target.

## Deployment documentation

Document the supported production build command, public output directory,
base-URL handling, draft and future-content policy, environment inputs, and the
contract by which each deployment mechanism receives the artefact. Do not embed
private infrastructure paths in public project documentation.

## References

- [Hugo security model](https://gohugo.io/about/security/)
- [Hugo Modules](https://gohugo.io/hugo-modules/)
- [Hugo configuration](https://gohugo.io/configuration/)
