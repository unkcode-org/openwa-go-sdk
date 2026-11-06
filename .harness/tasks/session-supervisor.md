# Session supervisor and full lifecycle

User approved the Gofip integration proposal on 2026-11-06 and explicitly requested completing and publishing the SDK before integrating it. Scope includes stop/logout/delete/pairing/proxy/config and optional connection supervision.

Design: Client remains side-effect free at construction. Supervisor.Run owns one explicitly selected session for a caller-owned context. Reconcile provides one-shot recovery for durable application workers. Only GET/START are automatic; no sends or destructive operations. Applications retain tenant scope, persistence and distributed coordination. Changes preserve Go 1.22 and the exact pinned OpenAPI fixture.

Verification: gofmt, go vet, go test -race, go build and git diff --check; results recorded after execution.

Verification completed: gofmt clean; go vet ./... passed; go test -race with local HTTP fixtures passed (93.0% coverage); go build ./... passed; git diff --check passed. HTTP fixture tests ran outside sandbox because loopback listeners were denied inside it. No real gateway/WhatsApp traffic. Architecture/security review: GET/START only in supervisor, no tenant discovery or persistence, destructive operations remain explicit and single attempt, config patch preserves null, errors retain redaction and bounded transport. No blocking findings.
