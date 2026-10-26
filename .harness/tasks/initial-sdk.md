# Initial reusable OpenWA SDK

## Authorization and scope

The user approved implementing the proposed standalone SDK and publishing it to
the newly created `unkcode-org/openwa-go-sdk` repository. That request satisfies
the architectural approval boundary for this SDK. It does not authorize deploying
client migrations or sending real WhatsApp messages.

Included: construction/configuration, transport, session read/start, QR, number
lookup, text/document sends, typed errors, contract fixtures, documentation and CI.
Excluded: Dexafiate adapter, notification persistence/workers, administration of
sessions/keys, destructive operations and live infrastructure changes.

## Organization harness context

The task continues the design governed by Dexafiate's organization harness. Its
compiler was executed successfully again before SDK implementation. The resolved
output was:

```yaml
version: 1
profiles: [go]
capabilities:
  direct: []
  resolved: [go-core, testing]
risk:
  requested_level: high
  resolved_level: high
  factors: [architecture-change, security-sensitive]
mode: audited
policies:
  - architecture-decisions.md
  - backend/project-structure.md
  - backend/modularity-srp.md
  - backend/domain-errors.md
  - testing.md
  - security.md
skills: []
verification: [backend_architecture, security]
reviews: [architecture, security]
human:
  approval_required: true
  resolution_required: false
```

The SDK follows the approved architectural boundary: provider DTOs/errors remain
in the library and each consuming application's adapter maps them to domain
models/errors. Backend-specific internal/core, REST envelopes and domain error
codes are owned by the consuming application, not by this transport library.
The existing harness kernel is not copied or reimplemented in this library.

## Implementation plan

1. Pin the user-confirmed OpenAPI and preserve its exact provenance/hash.
2. Implement synchronous, bounded HTTP operations with context cancellation,
   disabled redirects, no automatic retries and caller-owned transport support.
3. Return safe typed HTTP errors and explicit unknown outcomes after ambiguous sends.
4. Validate payloads and required response data; preserve provider session states.
5. Exercise contract shape, actual wire requests, failures, limits, concurrency,
   redirects, cancellation and document encoding with local fixtures.
6. Document consumption, ownership, defaults and failure semantics; add CI.
7. Verify, perform the resolved targeted reviews and publish initial version.

## Validation and review

- Go 1.26.1: gofmt, `go vet ./...`, `go build ./...` and
  `go test -race -coverprofile=coverage.out ./...` passed. Statement coverage: 97.7%.
- Go 1.22.12: `go test -ldflags=-linkmode=external ./...` passed. The default
  internal linker produces a binary rejected by this newer macOS host because
  it lacks LC_UUID; external linking was used to run the minimum-version suite.
  CI uses Linux for the normal Go 1.22 and Go 1.26 commands.
- The cross-version timeout test found that Go 1.22's HTTP timeout wrapper lost
  errors.Is(context.DeadlineExceeded). The implementation now places deadlines
  in the request context and preserves cancellation causes for headers and bodies;
  the regression suite passes on both versions.
- Organization `backend_architecture` script executed: explicitly skipped its
  backend directory check because this SDK has no internal/ application tree.
- Organization `security` script executed from this repository: passed.
- Architecture review: approved SDK-only boundary maintained; standard-library
  dependencies, operations split by file, provider types/errors kept out of any
  consumer domain, no process/tenant/persistence orchestration in the library.
- Security review: no blocking findings after checking path segments, redirect
  refusal, bounded bodies, key handling, error redaction, context cancellation,
  absence of retries and conservative handling of ambiguous send outcomes.
  URL documents remain a caller/gateway SSRF responsibility, documented explicitly.
- Contract hash and supported operation/field checks passed against the exact
  user-confirmed OpenAPI. Tests used local fixtures only; no live instance,
  session permissions, actual delivery or migrations were exercised.
- GitHub Actions is configured; remote run results are separate from local checks.

Publication target: `main` and initial version tag `v0.1.0`. Git whitespace check
and remote commit/tag verification are performed as the final publication steps.
