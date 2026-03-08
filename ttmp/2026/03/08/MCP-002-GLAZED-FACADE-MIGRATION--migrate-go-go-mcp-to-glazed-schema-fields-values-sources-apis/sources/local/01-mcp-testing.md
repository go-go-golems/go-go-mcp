---
Title: MCP testing note
Ticket: MCP-002-GLAZED-FACADE-MIGRATION
Status: active
Topics:
    - mcp
    - go
    - glazed
    - refactor
DocType: source
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Imported external note about practical MCP testing tools and harness structure.
LastUpdated: 2026-03-08T19:24:00-04:00
WhatFor: Capture the imported testing guidance verbatim inside the ticket workspace.
WhenToUse: Read when reviewing the external recommendations that informed the testing-plan follow-up.
---

Yes. The main useful options right now are:

**1. Official conformance runner** — best for “does my server actually match MCP?”
The MCP project has an official **Conformance Test Framework** that can run server scenarios against your endpoint, including initialization, tool listing/calls, resources, and prompts. It also has a GitHub Action, but the repo explicitly warns that it is still a **work in progress / unstable**. ([GitHub][1])

Example:

```bash
npx @modelcontextprotocol/conformance server --url http://localhost:3000/mcp
npx @modelcontextprotocol/conformance server --url http://localhost:3000/mcp --scenario server-initialize
npx @modelcontextprotocol/conformance list --server
```

That is the closest thing to an official “compliance test suite.” ([GitHub][1])

**2. Official MCP Inspector CLI** — best for a coding agent doing scripted smoke tests
The official **MCP Inspector** is not just a browser UI; it also has a **CLI mode**. The docs and repo describe it as suitable for scripting, CI, and feedback loops with coding assistants. It can connect to local stdio servers or remote servers, then list tools/resources/prompts and call tools with args. ([Model Context Protocol][2])

Examples:

```bash
# local stdio server
npx @modelcontextprotocol/inspector --cli node build/index.js --method tools/list

# call a tool
npx @modelcontextprotocol/inspector --cli node build/index.js \
  --method tools/call \
  --tool-name mytool \
  --tool-arg 'x=1'

# remote HTTP server
npx @modelcontextprotocol/inspector --cli https://my-mcp-server.example.com \
  --transport http \
  --method tools/list
```

If you want to hand **one tool** to a coding agent, this is probably the most practical starting point. ([GitHub][3])

**3. `mcp-testing-kit`** — good for unit/integration tests in TypeScript
There is a TS library called **mcp-testing-kit** that is specifically for testing MCP servers. It connects directly to a server instance through a dummy transport instead of going over HTTP/SSE, and it works with test runners like Jest/Vitest. That makes it good for deterministic CI tests around your own tools/prompts/resources. ([GitHub][4])

**4. `mcptools`** — good as a generic CLI smoke-test harness
`mcptools` is a third-party CLI that can list tools/resources/prompts, call them, and works over both stdio and HTTP. It also has mock and proxy modes, which can be handy if you want an agent to compose quick diagnostics or fixtures. ([GitHub][5])

There is also an official **reference/test server** called **Everything** in the MCP reference servers repo, which is useful as a known-good fixture when you want to validate your client/harness or compare behavior. ([GitHub][6])

My practical recommendation:

* Use **Conformance** for protocol correctness.
* Use **Inspector CLI** for agent-driven black-box tests.
* Use **mcp-testing-kit** for fast in-process regression tests if your server is TS.
* Optionally use **mcptools** for extra shell-level smoke tests. ([GitHub][1])

A good agent-facing setup is to give the agent a tiny contract like this:

```bash
./scripts/test-mcp.sh
```

where that script runs, in order:

```bash
# 1) protocol-level checks
npx @modelcontextprotocol/conformance server --url "$MCP_URL" --suite active

# 2) enumerate capabilities
npx @modelcontextprotocol/inspector --cli "$MCP_URL" --transport http --method tools/list
npx @modelcontextprotocol/inspector --cli "$MCP_URL" --transport http --method resources/list
npx @modelcontextprotocol/inspector --cli "$MCP_URL" --transport http --method prompts/list

# 3) invoke a few golden-path tools
npx @modelcontextprotocol/inspector --cli "$MCP_URL" --transport http \
  --method tools/call --tool-name health_check
```

That gives the agent a stable shell surface, exit codes, and text/JSON-ish output without needing it to understand your SDK internals. The main caveat is that the **official conformance framework is still marked unstable**, so I would treat it as necessary but not sufficient. ([GitHub][1])

I can sketch a full agent-friendly test harness repo layout next.

[1]: https://github.com/modelcontextprotocol/conformance "GitHub - modelcontextprotocol/conformance: Conformance Tests for MCP · GitHub"
[2]: https://modelcontextprotocol.io/docs/tools/inspector "MCP Inspector - Model Context Protocol"
[3]: https://github.com/modelcontextprotocol/inspector "GitHub - modelcontextprotocol/inspector: Visual testing tool for MCP servers · GitHub"
[4]: https://github.com/thoughtspot/mcp-testing-kit "GitHub - thoughtspot/mcp-testing-kit: The testing library you need to test your MCP servers · GitHub"
[5]: https://github.com/f/mcptools "GitHub - f/mcptools: A command-line interface for interacting with MCP (Model Context Protocol) servers using both stdio and HTTP transport. · GitHub"
[6]: https://github.com/modelcontextprotocol/servers "GitHub - modelcontextprotocol/servers: Model Context Protocol Servers · GitHub"
