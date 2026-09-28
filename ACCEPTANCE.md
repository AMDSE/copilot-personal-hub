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

## Deployed checks

- 2026-09-28: gateway and private consumer transport both running/healthy; combined idle memory sample about 54MiB. Only 127.0.0.1:4143 is mapped; no public consumer transport port.
- Original administrator password verified without reset. Secure + HttpOnly session cookie checked.
- Original panels, accounts, usage, keys, proxy settings, personal model catalog and local conversation routes returned expected responses. Anonymous admin access and import return 401.
- Synthetic personal account import, schedule disable, API key creation/revocation/deletion passed against the actual deployed Go container. Only these temporary records were removed. No Microsoft chat request was made.
- Private transport: anonymous request 401; authenticated malformed request 400; Go container can reach its health endpoint on the private Docker network.
- Public HTTPS root returns the original M365 Copilot2API panel; obsolete /console/ redirects to root. Source-side verification confirms personal import card exists.
- Original legacy CSS block compares byte-for-byte equal to upstream; React theme tokens file hash equals upstream. Browser login/navigation/import interaction checked locally with synthetic credentials.
- Previous Python main container stopped with restart disabled; its data/configuration retained for rollback. Other server services were not modified.
