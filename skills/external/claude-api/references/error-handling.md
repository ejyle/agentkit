# Claude API Error Handling Reference

## Error Types

The Anthropic SDK throws typed errors that map to HTTP status codes:

```typescript
import Anthropic, {
  APIError,
  AuthenticationError,
  PermissionDeniedError,
  NotFoundError,
  RateLimitError,
  InternalServerError,
  APIConnectionError,
} from "@anthropic-ai/sdk";

try {
  const response = await client.messages.create({ ... });
} catch (err) {
  if (err instanceof AuthenticationError) {
    // 401 — invalid API key
  } else if (err instanceof PermissionDeniedError) {
    // 403 — key lacks access to this model/feature
  } else if (err instanceof NotFoundError) {
    // 404 — model not found
  } else if (err instanceof RateLimitError) {
    // 429 — too many requests; retry with backoff
    const retryAfter = err.headers?.["retry-after"];
  } else if (err instanceof InternalServerError) {
    // 5xx — Anthropic server error; retry with backoff
  } else if (err instanceof APIConnectionError) {
    // Network error; retry
  } else if (err instanceof APIError) {
    // Other API error
    console.error(err.status, err.message);
  }
}
```

## Retry with Exponential Backoff

```typescript
async function withRetry<T>(
  fn: () => Promise<T>,
  options: { maxRetries?: number; baseDelay?: number } = {}
): Promise<T> {
  const { maxRetries = 3, baseDelay = 1000 } = options;

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      return await fn();
    } catch (err) {
      if (attempt === maxRetries) throw err;
      if (err instanceof RateLimitError || err instanceof InternalServerError) {
        const delay = baseDelay * 2 ** attempt + Math.random() * 1000;
        await new Promise(resolve => setTimeout(resolve, delay));
        continue;
      }
      throw err; // Don't retry auth/not-found errors
    }
  }
  throw new Error("Unreachable");
}

// Usage
const response = await withRetry(() => client.messages.create({ ... }));
```

## Rate Limits

| Tier | RPM | TPM |
|------|-----|-----|
| Free | 5 | 25,000 |
| Build | 50 | 100,000 |
| Scale | 2,000 | 4,000,000 |

Rate limit headers returned with every response:
- `anthropic-ratelimit-requests-limit`
- `anthropic-ratelimit-requests-remaining`
- `anthropic-ratelimit-tokens-limit`
- `anthropic-ratelimit-tokens-remaining`
- `anthropic-ratelimit-tokens-reset`

## Batch API (Large-Scale Workloads)

For more than 100 messages, use the Batch API to process asynchronously at 50% cost:

```typescript
const batch = await client.beta.messages.batches.create({
  requests: items.map((item, i) => ({
    custom_id: `item-${i}`,
    params: {
      model: "claude-haiku-3-5",
      max_tokens: 256,
      messages: [{ role: "user", content: item.prompt }],
    },
  })),
});

// Poll until complete
let status = batch;
while (status.processing_status !== "ended") {
  await new Promise(resolve => setTimeout(resolve, 10_000));
  status = await client.beta.messages.batches.retrieve(batch.id);
}

// Stream results
for await (const result of await client.beta.messages.batches.results(batch.id)) {
  if (result.result.type === "succeeded") {
    console.log(result.custom_id, result.result.message.content[0].text);
  }
}
```
