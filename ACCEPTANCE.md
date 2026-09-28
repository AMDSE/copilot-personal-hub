# Go-base reconstruction acceptance — 2026-09-28

## Local automated evidence

- Original React build and TypeScript: passed.
- Go complete package suite: passed in a fresh isolated test data directory. Reusing test data causes the upstream password rotation test to fail; this is not a production password change.
- Go provider integration tests: encrypted credential persistence/reopen, identity mismatch rejection, admin-only import, redacted account listing, original management routes.
- Consumer bridge: exact stream assembly, no duplicated event deltas, truncated stream rejection.
- Go end-to-end protocol adapters against a simulated transport: Chat Completions, Responses, Messages; both streaming and non-streaming. This is NOT a live Microsoft model test.
- Private Python transport: 3 tests passed (private authentication, streaming, secret-safe errors).
- Credential exporter JavaScript syntax: passed.

## Required live matrix

| Path | Status |
| --- | --- |
| M365 native | BLOCKED: no authorized organizational account |
| Router planning / tool loop | BLOCKED: no live account credentials |
| Studio planning / tool loop | BLOCKED: no licensed Studio account; not a consumer feature |
| Anthropic Messages tool loop | BLOCKED: no live account credentials; protocol-level mock only |
| OpenAI Responses tool loop | BLOCKED: no live account credentials; protocol-level mock only |
| Personal Consumer | BLOCKED: owner must import personal credentials |

Do not claim full live acceptance until account prerequisites are supplied. The deployed interface and admin checks are tracked separately from Microsoft availability.
