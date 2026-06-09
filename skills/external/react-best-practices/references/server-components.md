# React Server Components Reference

## RSC vs Client Components

| Feature | Server Component | Client Component |
|---------|-----------------|-----------------|
| Default in Next.js App Router | Yes | No (add `"use client"`) |
| Runs on | Server only | Browser (+ server for SSR) |
| Can use hooks | No | Yes |
| Can use browser APIs | No | Yes |
| Can access filesystem, DB | Yes | No |
| Adds JS to bundle | No | Yes |
| Can be async | Yes | No (use `use()` hook) |

## Server Component Patterns

### Data Fetching in RSC

```tsx
// app/posts/page.tsx — Server Component (no "use client")
import { db } from "@/lib/db";

export default async function PostsPage() {
  // Direct DB access — no API route needed
  const posts = await db.post.findMany({ orderBy: { createdAt: "desc" } });

  return (
    <main>
      <h1>Posts</h1>
      {posts.map(post => (
        <article key={post.id}>
          <h2>{post.title}</h2>
          <p>{post.excerpt}</p>
        </article>
      ))}
    </main>
  );
}
```

### Parallel Data Fetching

```tsx
export default async function Dashboard() {
  // Parallel — both start simultaneously
  const [user, stats] = await Promise.all([
    getUser(),
    getStats(),
  ]);

  return (
    <>
      <UserCard user={user} />
      <StatsPanel stats={stats} />
    </>
  );
}
```

### Streaming with Suspense

```tsx
import { Suspense } from "react";

export default function Page() {
  return (
    <main>
      <h1>Dashboard</h1>
      {/* Slow data renders after fast data */}
      <Suspense fallback={<StatsLoading />}>
        <SlowStats />
      </Suspense>
    </main>
  );
}

async function SlowStats() {
  const stats = await getExpensiveStats(); // slow query
  return <StatsPanel stats={stats} />;
}
```

## Client Component Boundaries

Add `"use client"` only when the component needs:
- Event handlers (`onClick`, `onChange`)
- Browser APIs (`window`, `document`, `navigator`)
- React hooks (`useState`, `useEffect`, `useContext`)
- Real-time subscriptions (WebSocket, SSE)

```tsx
"use client";
import { useState } from "react";

// This entire subtree is client-side
export function Counter() {
  const [count, setCount] = useState(0);
  return <button onClick={() => setCount(c => c + 1)}>{count}</button>;
}
```

### Passing Server Data to Client Components

```tsx
// Server Component
import { Counter } from "./Counter";
import { db } from "@/lib/db";

export default async function Page() {
  const initialCount = await db.counter.findFirst();
  // Pass serializable props — no functions, no class instances, no circular refs
  return <Counter initialCount={initialCount.value} />;
}
```

## use() Hook (React 19 / Next.js 15)

Stream promises into client components:

```tsx
"use client";
import { use } from "react";

function UserCard({ userPromise }: { userPromise: Promise<User> }) {
  const user = use(userPromise); // Suspends until resolved
  return <div>{user.name}</div>;
}
```

## Caching in Next.js

```typescript
// Deduplicated within a single request (default)
async function getUser(id: string) {
  return fetch(`/api/users/${id}`).then(r => r.json());
}

// Cache for 1 hour across requests (ISR-style)
async function getProducts() {
  return fetch("/api/products", { next: { revalidate: 3600 } }).then(r => r.json());
}

// Never cache (always fresh)
async function getLivePrice() {
  return fetch("/api/price", { cache: "no-store" }).then(r => r.json());
}
```
