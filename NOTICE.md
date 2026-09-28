# Sources and changes

This is a combined derivative, not an independently invented Copilot transport.

- Ciallo-Ms-365-OpenAI-Proxy-Docker by MurasameCyan and contributors, commit 72cd93aa82d6a4dcffa76085f1623d8e1998c0c6: backend, protocols, account stores, original admin/user pages, userscript, tests and Docker foundation. Apache-2.0; full notice in licenses/Ciallo-LICENSE. Original README retained as README.upstream.md.
- M365-Copilot2API by HEXUXIU and contributors, commit a182c0f2c5d3c86096ed1631aab6b08961823130: console design tokens, React build foundation and navigation/workflow reference. Its complete AGPL-3.0 text and additional non-commercial relay restriction are preserved in licenses/M365-Copilot2API-LICENSE and LICENSE.

Integration changes dated 2026-09-28: new Chinese React console bound to Ciallo account/key APIs, personal-account onboarding, real chat test, static asset mounts, resource-limited local-only deployment. No upstream OAuth flow is presented as personal-account authorization. No Microsoft credentials are bundled.

The combined distribution retains the upstream AGPL terms and non-commercial relay restriction; Apache notices remain attached to the Ciallo components. Do not offer paid API relay access under this distribution without resolving upstream permission requirements. The console links to the corresponding public source.
