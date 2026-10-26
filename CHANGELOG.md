# Changelog

## v0.1.0 — 2026-10-26

- Initial independent Go module, package `openwa`, requiring Go 1.22 or later.
- Session inspection/start, QR retrieval and number lookup.
- Synchronous text/document sends using the pinned OpenWA 0.23.4 contract.
- Typed API errors and explicit unknown send outcomes, without automatic retries.
- Configurable request/response/document limits, context cancellation and disabled redirects.
- Local HTTP tests, contract fixtures, usage documentation and GitHub Actions CI.

No live WhatsApp deployment, session administration or application notification
persistence is included in this release.
