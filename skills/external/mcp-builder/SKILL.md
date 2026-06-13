# MCP Builder (via anthropics/skills)

---
name: mcp-builder
description: >
  Use when building Model Context Protocol (MCP) servers — covers server scaffolding,
  tool and resource definition, transport configuration (stdio/SSE), error handling,
  and publishing an MCP server for use with Claude Code and other MCP clients.
license: MIT
source: https://github.com/anthropics/skills
---

## When to Use

Activate this skill when the task involves:

- Scaffolding a new MCP server in TypeScript or Python
- Defining MCP tools (callable functions) and resources (readable data)
- Wiring up stdio or SSE transports for Claude Code integration
- Implementing input schema validation for tool parameters
- Publishing or distributing an MCP server via npm or PyPI
- Debugging MCP server connectivity and tool registration issues

## MCP Architecture Overview

```
Claude Code / MCP Client
        |
   MCP Protocol (JSON-RPC over stdio or SSE)
        |
   MCP Server  ─── tools: [ search(), create(), delete() ]
                ─── resources: [ file://, db://, api:// ]
                ─── prompts: [ named prompt templates ]
```

An MCP server exposes capabilities to AI clients through a standardized protocol. Tools
are functions the AI can call. Resources are data sources the AI can read. Prompts are
reusable message templates.

## TypeScript Quick Start

```bash
npm create mcp-server@latest my-server
cd my-server
npm install
```

### Minimal MCP Server (TypeScript)

```typescript
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";

const server = new McpServer({
  name: "my-server",
  version: "1.0.0",
});

// Define a tool
server.tool(
  "get_weather",
  "Returns current weather for a city",
  {
    city: z.string().describe("City name"),
    units: z.enum(["metric", "imperial"]).default("metric"),
  },
  async ({ city, units }) => {
    // Implementation
    const data = await fetchWeather(city, units);
    return {
      content: [{ type: "text", text: JSON.stringify(data) }],
    };
  }
);

// Define a resource
server.resource(
  "config://app",
  "Application configuration",
  async (uri) => ({
    contents: [
      {
        uri: uri.href,
        mimeType: "application/json",
        text: JSON.stringify({ version: "1.0.0" }),
      },
    ],
  })
);

// Start the server
const transport = new StdioServerTransport();
await server.connect(transport);
```

## Python Quick Start

```bash
pip install mcp
```

### Minimal MCP Server (Python)

```python
from mcp.server import Server
from mcp.server.stdio import stdio_server
from mcp.types import Tool, TextContent

server = Server("my-server")

@server.list_tools()
async def list_tools() -> list[Tool]:
    return [
        Tool(
            name="search_docs",
            description="Search project documentation",
            inputSchema={
                "type": "object",
                "properties": {
                    "query": {"type": "string", "description": "Search query"},
                },
                "required": ["query"],
            },
        )
    ]

@server.call_tool()
async def call_tool(name: str, arguments: dict) -> list[TextContent]:
    if name == "search_docs":
        results = await search(arguments["query"])
        return [TextContent(type="text", text=str(results))]
    raise ValueError(f"Unknown tool: {name}")

async def main():
    async with stdio_server() as (read, write):
        await server.run(read, write, server.create_initialization_options())

import asyncio
asyncio.run(main())
```

## Claude Code Integration

Register the server in `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "my-server": {
      "command": "node",
      "args": ["/path/to/my-server/dist/index.js"],
      "env": {}
    }
  }
}
```

For npm-published servers:

```json
{
  "mcpServers": {
    "my-server": {
      "command": "npx",
      "args": ["-y", "my-mcp-server@latest"],
      "env": { "API_KEY": "${MY_API_KEY}" }
    }
  }
}
```

## Tool Design Principles

- **Single responsibility** — one tool does one thing clearly
- **Descriptive schema** — every parameter needs a `description`; Claude uses it to fill arguments
- **Return structured data** — prefer JSON for machine-readable results; prose for user-facing
- **Error surfaces** — throw typed errors with helpful messages; never swallow exceptions
- **Idempotent reads** — tools that only read data should be safe to call multiple times

## Reference Files

| Task | Reference File |
|------|---------------|
| Transport types, SSE setup, auth patterns | `references/transports.md` |
| Tool schema patterns, validation, error codes | `references/tools-and-resources.md` |

## Common Gotchas

- **stdio vs SSE** — stdio is the default and requires no network; SSE is for remote servers
- **Schema validation** — always validate inputs with zod/pydantic; malformed args crash the server
- **Blocking the event loop** — use async/await for all I/O; synchronous calls block all clients
- **Missing initialization** — the MCP handshake must complete before tools are callable; never call tools before `connect()`
- **Environment variable leakage** — never log env vars; use a `.env` file and add it to `.gitignore`
