# Sources and integration changes

Main application: HEXUXIU/M365-Copilot2API, commit a182c0f2c5d3c86096ed1631aab6b08961823130. Go gateway, account/key stores, settings, usage, conversations, protocol adapters, original HTML and React consoles are retained. LICENSE preserves its complete AGPL text and additional non-commercial relay restriction.

Personal transport: MurasameCyan/Ciallo-Ms-365-OpenAI-Proxy-Docker, commit 72cd93aa82d6a4dcffa76085f1623d8e1998c0c6. consumer/consumer_client.py is its consumer protocol implementation. The credential capture script is adapted from get_token.user.js to export locally rather than push to the former Python admin API. The Go MSA refresh exchange follows its consumer_refresh_via_rt identity-binding logic. Apache notice retained in licenses/Ciallo-LICENSE.

2026-09-28 correction: replace the previous Python-based main app with the actual Go enterprise application. Add encrypted personal credential fields, identity-checked refresh, private transport bridge, stream/error tests, personal model catalog, minimal same-style import cards and deployment hardening. Existing panel styling is not redesigned. Public source is linked from both consoles.

Changes are distributed subject to the retained upstream notices and restrictions. No account secrets are included. No Microsoft endorsement or official API status is claimed.
