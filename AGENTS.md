# OpenWA Go SDK

This repository is an independent Go client library. Keep it free of application
domain models, database drivers, process configuration, and framework dependencies.
The package name is `openwa`; its module is `github.com/unkcode-org/openwa-go-sdk`.

## Contract and architecture

- Use the pinned contract in `testdata/openapi.json`; provenance is in README.md.
- One public operation per implementation file; tests stay next to that operation.
- Preserve the application's ownership of session/tenant scope and notification persistence.
- Use SDK sentinels, `APIError` and `SendError`. Application adapters translate
  them into their own domain error catalogs; do not import a backend domain here.
- New or destructive gateway operations require explicit scope and tests.

## Transport invariants

- Preserve context cancellation and bounded requests/responses.
- Never automatically retry sends or follow redirects with X-API-Key.
- Never log keys, QR images, documents, message text or upstream bodies.
- Acceptance is not delivery. Unknown send outcomes must remain distinguishable
  with errors.Is(err, ErrOutcomeUnknown), including through wrapping.
- Keep existing caller-owned HTTP clients unchanged.

## Verification

Run gofmt, `go vet ./...`, `go test -race ./...`, `go build ./...` and
`git diff --check` for changes. Preserve Go 1.22 compatibility. Use local HTTP
fixtures; do not send real messages or require production credentials in tests.

The initial task's organization harness resolution, approval and reviews are
recorded in `.harness/tasks/initial-sdk.md`.
