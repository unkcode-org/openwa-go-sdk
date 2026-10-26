# Changelog

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
