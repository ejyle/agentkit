# MCP Tools and Resources Reference

## Tool Schema Design

Every tool parameter needs a clear `description` — Claude uses it to decide what to pass:

```typescript
server.tool(
  "search_database",
  "Search records in the database by field and value",
  {
    table: z.enum(["users", "orders", "products"])
      .describe("Database table to search"),
    field: z.string()
      .describe("Column name to filter on (e.g. 'email', 'status')"),
    value: z.string()
      .describe("Value to match — exact match, case-insensitive"),
    limit: z.number().int().min(1).max(100).default(10)
      .describe("Maximum number of results to return"),
  },
  async ({ table, field, value, limit }) => {
    const results = await db.query(table, field, value, limit);
    return {
      content: [{
        type: "text",
        text: JSON.stringify(results, null, 2),
      }],
    };
  }
);
```

## Tool Return Types

```typescript
// Text response
return { content: [{ type: "text", text: "Result string" }] };

// Image response
return {
  content: [{
    type: "image",
    data: base64String,
    mimeType: "image/png",
  }],
};

// Multiple content blocks
return {
  content: [
    { type: "text", text: "Here is the chart:" },
    { type: "image", data: chartBase64, mimeType: "image/png" },
  ],
};

// Error (isError: true tells Claude the tool failed)
return {
  content: [{ type: "text", text: "Error: record not found" }],
  isError: true,
};
```

## Resource Patterns

Resources are read-only data sources that Claude can access by URI:

```typescript
// Static resource
server.resource(
  "config://settings",
  "Application settings",
  async () => ({
    contents: [{
      uri: "config://settings",
      mimeType: "application/json",
      text: JSON.stringify(await loadSettings()),
    }],
  })
);

// Dynamic resource with URI template
server.resource(
  "db://users/{id}",
  "User record by ID",
  async (uri) => {
    const id = uri.pathname.split("/").pop();
    const user = await db.users.findById(id);
    return {
      contents: [{
        uri: uri.href,
        mimeType: "application/json",
        text: JSON.stringify(user),
      }],
    };
  }
);
```

## Prompts

Named, reusable message templates:

```typescript
server.prompt(
  "code_review",
  "Review code for quality, security, and correctness",
  {
    code: z.string().describe("Code to review"),
    language: z.string().describe("Programming language"),
  },
  ({ code, language }) => ({
    messages: [
      {
        role: "user",
        content: {
          type: "text",
          text: `Review this ${language} code:\n\n\`\`\`${language}\n${code}\n\`\`\`\n\nFocus on: correctness, security, readability, and performance.`,
        },
      },
    ],
  })
);
```

## Error Handling Best Practices

```typescript
server.tool("risky_operation", "Performs a risky operation", {
  id: z.string(),
}, async ({ id }) => {
  try {
    const result = await performOperation(id);
    return { content: [{ type: "text", text: JSON.stringify(result) }] };
  } catch (err) {
    // Log server-side for debugging
    console.error("risky_operation failed:", err);

    // Return structured error to Claude
    return {
      content: [{
        type: "text",
        text: `Operation failed: ${err instanceof Error ? err.message : "Unknown error"}`,
      }],
      isError: true,
    };
  }
});
```

## Testing MCP Servers

Use the MCP Inspector to test without a full Claude Code setup:

```bash
npx @modelcontextprotocol/inspector node dist/index.js
```

The inspector provides a web UI to call tools, read resources, and inspect the protocol.

For automated testing, use the in-process client:

```typescript
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { InMemoryTransport } from "@modelcontextprotocol/sdk/inMemory.js";

const [clientTransport, serverTransport] = InMemoryTransport.createLinkedPair();
const client = new Client({ name: "test-client", version: "1.0.0" }, { capabilities: {} });
await client.connect(clientTransport);
await server.connect(serverTransport);

const result = await client.callTool("get_weather", { city: "Tokyo" });
```
