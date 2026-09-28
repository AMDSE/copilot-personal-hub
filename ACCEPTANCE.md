# Acceptance status

The upstream AGENTS.md requires live acceptance. Local tests alone are insufficient.

| Live path | Status before account onboarding |
| --- | --- |
| M365 direct/native | BLOCKED: no authorized organizational account supplied |
| Router planning/tool | BLOCKED: no authorized upstream account supplied |
| Studio planning/tool | BLOCKED: no Studio-capable account supplied |
| Anthropic Messages tool loop | BLOCKED: no authorized upstream account supplied |
| OpenAI Responses tool loop | BLOCKED: no authorized upstream account supplied |
| Consumer personal chat | BLOCKED: owner must sign in and push own credentials |

Deployment health and administration tests do not establish Microsoft chat availability. Do not describe these six paths as passed until real requests and results are recorded without secrets.
