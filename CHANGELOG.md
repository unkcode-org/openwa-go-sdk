# Changelog

## v0.5.1 — 2026-11-14

- Update the pinned OpenAPI snapshot to OpenWA 0.30.1; the contract changes only its version.

## v0.5.0 — 2026-11-14

- Add image messages from base64 or a URL with captions and typed send outcomes.
- Add session-scoped webhook listing, creation and updates with partial configuration.
- Keep the OpenWA 0.30.0 contract and Go 1.22 support.

## v0.4.0 — 2026-11-14

- Update the pinned OpenAPI snapshot to OpenWA 0.30.0.
- Support custom link preview thumbnails and the passkey restriction kind.
- Validate pairing phone length and document gateway compatibility changes.

## v0.3.0 — 2026-11-06

- Explicit stop, logout, delete and pairing code operations; creation proxy fields.
- Read/patch session configuration with distinct omitted/null/zero semantics.
- Optional cancelable session supervisor and one-shot reconciliation, bounded concurrency, per-session serialization and capped recovery backoff with jitter.
- Loaded engines are left to native reconnection; missing sessions are never recreated and human linking is surfaced. No automatic sends or destructive recovery.
- Additional pinned-contract, lifecycle, concurrency and recovery tests.


## v0.2.0 — 2026-10-26

- Explicit session creation with unique-name validation and typed configuration.
- Paginated session listing for recovery after an uncertain creation response.
- Object/array transport decoding with the existing response limits and redaction.
- Contract and regression coverage for creation, pagination and conflicts.
- Session persistence, ownership and automatic provisioning remain responsibilities
  of the consuming application; no key management or destructive operations added.

## v0.1.0 — 2026-10-26

- Initial independent Go module, package `openwa`, requiring Go 1.22 or later.
- Session inspection/start, QR retrieval and number lookup.
- Synchronous text/document sends using the pinned OpenWA 0.23.4 contract.
- Typed API errors and explicit unknown send outcomes, without automatic retries.
- Configurable request/response/document limits, context cancellation and disabled redirects.
- Local HTTP tests, contract fixtures, usage documentation and GitHub Actions CI.

No live WhatsApp deployment, session administration or application notification
persistence is included in this release.
