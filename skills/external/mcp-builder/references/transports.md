# MCP Transports Reference

## stdio Transport (Default)

stdio is the recommended transport for local MCP servers. The client spawns the server
as a child process and communicates over stdin/stdout.

**Advantages:** No network port, no auth needed, zero latency, process lifecycle managed by client.
**Disadvantages:** Single-client only, not suitable for remote or shared servers.

```typescript
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";

const transport = new StdioServerTransport();
await server.connect(transport);
```

Claude Code `settings.json` entry:
```json
{
  "mcpServers": {
    "my-server": {
      "command": "node",
      "args": ["dist/index.js"]
    }
  }
}
```

## SSE Transport (Remote)

Server-Sent Events transport enables remote MCP servers over HTTP. Requires an HTTP
server that handles GET `/sse` (event stream) and POST `/message` (tool calls).

```typescript
import express from "express";
import { SSEServerTransport } from "@modelcontextprotocol/sdk/server/sse.js";

const app = express();
const transports: Record<string, SSEServerTransport> = {};

app.get("/sse", async (req, res) => {
  const transport = new SSEServerTransport("/message", res);
  transports[transport.sessionId] = transport;
  await server.connect(transport);
});

app.post("/message", async (req, res) => {
  const sessionId = req.query.sessionId as string;
  await transports[sessionId].handlePostMessage(req, res);
});

app.listen(3000);
```

Claude Code `settings.json` entry for SSE:
```json
{
  "mcpServers": {
    "my-remote-server": {
      "url": "http://localhost:3000/sse"
    }
  }
}
```

## Authentication Patterns

### Bearer Token (SSE servers)

```typescript
app.use("/sse", (req, res, next) => {
  const token = req.headers.authorization?.replace("Bearer ", "");
  if (token !== process.env.MCP_AUTH_TOKEN) {
    res.status(401).json({ error: "Unauthorized" });
    return;
  }
  next();
});
```

Claude Code `settings.json` with auth:
```json
{
  "mcpServers": {
    "my-server": {
      "url": "http://my-server.example.com/sse",
      "headers": {
        "Authorization": "Bearer ${MCP_AUTH_TOKEN}"
      }
    }
  }
}
```

### OAuth 2.0 (Advanced)

For production remote servers, implement OAuth 2.0 with PKCE. The MCP spec includes
an `Authorization` capability that clients like Claude Code can negotiate. Refer to the
MCP specification at modelcontextprotocol.io for the full OAuth flow.
