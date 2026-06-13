# Claude API (via anthropics/skills)

---
name: claude-api
description: >
  Use when integrating with the Anthropic Claude API — covers authentication, message
  construction, streaming, tool use, vision inputs, prompt caching, and cost optimization
  for production deployments.
license: MIT
source: https://github.com/anthropics/skills
---

## When to Use

Activate this skill when the task involves:

- Calling the Anthropic Messages API from TypeScript, Python, or other languages
- Implementing streaming responses for real-time output
- Building tool use (function calling) with Claude
- Sending images or documents to the API (vision)
- Optimizing costs with prompt caching or batching
- Handling rate limits, retries, and error responses
- Multi-turn conversation management

## Authentication

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

Never hard-code the API key. Use environment variables or a secrets manager.

## Basic Message (TypeScript)

```typescript
import Anthropic from "@anthropic-ai/sdk";

const client = new Anthropic(); // reads ANTHROPIC_API_KEY from env

const message = await client.messages.create({
  model: "claude-opus-4-5",
  max_tokens: 1024,
  messages: [
    { role: "user", content: "Explain the CAP theorem in plain English." }
  ],
});

console.log(message.content[0].text);
```

## Basic Message (Python)

```python
import anthropic

client = anthropic.Anthropic()  # reads ANTHROPIC_API_KEY from env

message = client.messages.create(
    model="claude-opus-4-5",
    max_tokens=1024,
    messages=[
        {"role": "user", "content": "Explain the CAP theorem in plain English."}
    ],
)
print(message.content[0].text)
```

## Streaming

```typescript
const stream = client.messages.stream({
  model: "claude-opus-4-5",
  max_tokens: 1024,
  messages: [{ role: "user", content: "Write a short story." }],
});

for await (const chunk of stream) {
  if (chunk.type === "content_block_delta" && chunk.delta.type === "text_delta") {
    process.stdout.write(chunk.delta.text);
  }
}

const finalMessage = await stream.finalMessage();
console.log("\nTotal tokens:", finalMessage.usage.input_tokens + finalMessage.usage.output_tokens);
```

## Tool Use (Function Calling)

```typescript
const tools = [
  {
    name: "get_weather",
    description: "Get current weather for a location",
    input_schema: {
      type: "object",
      properties: {
        location: { type: "string", description: "City and country" },
        units: { type: "string", enum: ["celsius", "fahrenheit"] },
      },
      required: ["location"],
    },
  },
];

let messages = [{ role: "user", content: "What's the weather in Tokyo?" }];

while (true) {
  const response = await client.messages.create({
    model: "claude-opus-4-5",
    max_tokens: 1024,
    tools,
    messages,
  });

  if (response.stop_reason === "end_turn") {
    console.log(response.content[0].text);
    break;
  }

  if (response.stop_reason === "tool_use") {
    const toolUse = response.content.find((b) => b.type === "tool_use");
    const result = await callTool(toolUse.name, toolUse.input); // your dispatch

    messages = [
      ...messages,
      { role: "assistant", content: response.content },
      { role: "user", content: [{ type: "tool_result", tool_use_id: toolUse.id, content: result }] },
    ];
  }
}
```

## Vision (Image Inputs)

```typescript
const response = await client.messages.create({
  model: "claude-opus-4-5",
  max_tokens: 1024,
  messages: [
    {
      role: "user",
      content: [
        { type: "text", text: "Describe what you see in this image." },
        {
          type: "image",
          source: {
            type: "base64",
            media_type: "image/jpeg",
            data: imageBase64String,
          },
        },
      ],
    },
  ],
});
```

## Prompt Caching

Cache large system prompts or document context to reduce costs on repeated calls:

```typescript
const response = await client.messages.create({
  model: "claude-opus-4-5",
  max_tokens: 1024,
  system: [
    {
      type: "text",
      text: largeDocumentContext, // 10k+ tokens
      cache_control: { type: "ephemeral" }, // cached for ~5 min
    },
  ],
  messages: [{ role: "user", content: "Summarize section 3." }],
});
```

Cache hits reduce input token cost by ~90%. Cache writes cost 25% more than normal.

## Model Reference

| Model | Best For | Context |
|-------|----------|---------|
| claude-opus-4-5 | Complex reasoning, nuanced tasks | 200k tokens |
| claude-sonnet-4-5 | Balanced speed and quality | 200k tokens |
| claude-haiku-3-5 | Fast, cost-efficient tasks | 200k tokens |

## Reference Files

| Task | Reference File |
|------|---------------|
| Error handling, rate limits, retries, batching | `references/error-handling.md` |
| Multi-turn conversations, context management | `references/conversations.md` |

## Common Gotchas

- **max_tokens is required** — the API never sets a default; always pass it explicitly
- **Tool result must reference tool_use_id** — mismatch causes a 400 validation error
- **Stop reason check** — always check `stop_reason` before accessing content; `tool_use` stops before generating text
- **Context window math** — input + output must fit the model's context; count tokens before sending large docs
- **Base64 images** — strip the `data:image/jpeg;base64,` prefix; send only the raw base64 string
