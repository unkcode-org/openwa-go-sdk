# Session provisioning SDK extension

The user explicitly corrected the Dexafiate integration: its adapter must create
and recover the WhatsApp session instead of requiring manual provisioning.
The existing SDK publication authorization covers the necessary additive release.

Harness context: Dexafiate compiler executed successfully using
`.harness/runtime/openwa-provisioning-route.yaml`; go, high/audited,
architecture-change and security-sensitive, with architecture/security reviews.
Contract remains the pinned OpenWA 0.23.4 snapshot.

Scope: CreateSession and paginated ListSessions, validation, bounded typed array
responses, contract checks, tests and v0.2.0 publication. No destructive operations,
API key management, built-in retries, application persistence or real gateway calls.

Architecture/security review: state ownership stays in the consumer; exact names
are validated; API errors remain redacted; pagination is bounded by the caller's
context; response/request caps, disabled redirects and cancellation are preserved.
Tests verify malformed responses, wrong creation names, 409s, query parameters and
invalid inputs. The consumer must persist an intent name before POST and recover
by that name after uncertain creation; no stronger idempotency claim is made.
