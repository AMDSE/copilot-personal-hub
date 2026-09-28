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

## Local checks on 2026-09-28

- React TypeScript check and Vite production build: passed. npm audit: zero reported vulnerabilities at install time.
- Focused console/auth/account/consumer tests: 61 passed.
- Wider Python run, excluding two uncollectable test modules: 2074 passed, 3 skipped, 2 failed. Failures: absent upstream Docker CI workflow (not copied to this fork), optional Camoufox package not installed in the lightweight environment.
- Full collection blockers: upstream packaging test assumes its original docker-compose.yml; studio diagnostic test refers to an untracked .probe script not present in upstream checkout. The fork uses compose.yaml. These have not been presented as successful tests.
- New scripts/smoke.py verifies deployed administration and removes only its own temporary account/key. It does not call Microsoft or establish provider availability.
