# Vercel Optimize (via vercel-labs/agent-skills)

---
name: vercel-optimize
description: >
  Use when optimizing Next.js applications deployed to Vercel — covers Core Web
  Vitals improvement, Image and Font optimization, Edge Runtime, ISR/SSR caching
  strategies, bundle analysis, and Vercel-specific performance features.
license: MIT
source: https://github.com/vercel-labs/agent-skills
---

## When to Use

Activate this skill when the task involves:

- Improving Core Web Vitals (LCP, CLS, INP/FID) scores on a Vercel-deployed app
- Configuring Next.js Image optimization (`next/image`) for performance
- Setting up caching strategies (ISR, SSR with `Cache-Control`, Edge caching)
- Analyzing bundle size and eliminating unnecessary JavaScript
- Using Vercel Edge Runtime or Edge Middleware for low-latency responses
- Configuring `vercel.json` for headers, rewrites, and redirects

## Core Web Vitals Targets

| Metric | Good | Needs Work | Poor |
|--------|------|-----------|------|
| LCP (Largest Contentful Paint) | < 2.5s | 2.5–4s | > 4s |
| CLS (Cumulative Layout Shift) | < 0.1 | 0.1–0.25 | > 0.25 |
| INP (Interaction to Next Paint) | < 200ms | 200–500ms | > 500ms |

Measure with Vercel Speed Insights or `npx lighthouse <url>`.

## Image Optimization

### next/image — Always Use It

```tsx
import Image from "next/image";

// Replace all <img> tags with next/image
<Image
  src="/hero.jpg"
  alt="Hero banner"
  width={1200}
  height={630}
  priority           // preload above-the-fold images (fixes LCP)
  quality={85}       // default 75; 85 is a good balance
  sizes="(max-width: 768px) 100vw, 50vw"
/>
```

### Prevent Layout Shift (CLS)

Always specify `width` and `height`, or use `fill` with a positioned parent:

```tsx
<div style={{ position: "relative", aspectRatio: "16/9" }}>
  <Image
    src="/banner.jpg"
    alt="Banner"
    fill
    sizes="100vw"
    style={{ objectFit: "cover" }}
  />
</div>
```

### Remote Images — Allowlist in next.config

```js
// next.config.js
module.exports = {
  images: {
    remotePatterns: [
      { protocol: "https", hostname: "images.example.com" },
      { protocol: "https", hostname: "**.cdn.com" },
    ],
  },
};
```

## Font Optimization

### next/font (Zero Layout Shift)

```tsx
import { Inter, JetBrains_Mono } from "next/font/google";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
  display: "swap",
});

const mono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
  display: "swap",
});

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`${inter.variable} ${mono.variable}`}>
      <body>{children}</body>
    </html>
  );
}
```

CSS:
```css
body { font-family: var(--font-inter), system-ui, sans-serif; }
code { font-family: var(--font-mono), monospace; }
```

## Caching Strategies

### ISR (Incremental Static Regeneration)

```tsx
// App Router — revalidate at route level
export const revalidate = 60; // seconds

// Granular revalidation
export async function generateStaticParams() {
  const posts = await getPosts();
  return posts.map(p => ({ slug: p.slug }));
}
```

### On-Demand Revalidation

```typescript
import { revalidatePath, revalidateTag } from "next/cache";

// Revalidate a path
export async function POST() {
  revalidatePath("/blog");
  return Response.json({ revalidated: true });
}

// Revalidate by tag (granular)
const data = await fetch("https://api.example.com/posts", {
  next: { tags: ["posts"] },
});
// Later: revalidateTag("posts")
```

### Fetch Cache Control

```typescript
// Static (cached forever until revalidated)
fetch(url, { cache: "force-cache" });

// Dynamic (never cached — fresh each request)
fetch(url, { cache: "no-store" });

// Time-based (ISR style)
fetch(url, { next: { revalidate: 3600 } });
```

## Edge Runtime

Deploy routes to the edge for sub-50ms response times globally:

```typescript
// app/api/fast/route.ts
export const runtime = "edge";

export async function GET() {
  return new Response("Hello from the edge!");
}
```

Edge limitations: no Node.js APIs (`fs`, `crypto` module, child_process), no native addons.
Use Web APIs (`fetch`, `Response`, `crypto.subtle`) only.

## Bundle Analysis

```bash
npm install --save-dev @next/bundle-analyzer
```

```js
// next.config.js
const withBundleAnalyzer = require("@next/bundle-analyzer")({
  enabled: process.env.ANALYZE === "true",
});
module.exports = withBundleAnalyzer({ /* config */ });
```

```bash
ANALYZE=true npm run build
```

Common bundle issues:
- Large date libraries → use `date-fns` with tree-shaking, not `moment`
- Full lodash import → use `lodash-es` or individual function imports
- Icons → import individually (`import { Star } from "lucide-react"`)

## vercel.json Configuration

```json
{
  "headers": [
    {
      "source": "/(.*)",
      "headers": [
        { "key": "X-Content-Type-Options", "value": "nosniff" },
        { "key": "X-Frame-Options", "value": "DENY" },
        { "key": "X-XSS-Protection", "value": "1; mode=block" }
      ]
    },
    {
      "source": "/static/(.*)",
      "headers": [
        { "key": "Cache-Control", "value": "public, max-age=31536000, immutable" }
      ]
    }
  ],
  "rewrites": [
    { "source": "/api/:path*", "destination": "https://internal-api.example.com/:path*" }
  ]
}
```

## Reference Files

| Task | Reference File |
|------|---------------|
| Middleware, A/B testing, feature flags at the edge | `references/edge-middleware.md` |

## Common Gotchas

- **Missing `sizes` on `<Image>`** — without `sizes`, Next.js serves the full-size image on mobile; always set `sizes` for responsive images
- **`priority` overuse** — mark only above-the-fold images as `priority`; every image has priority means no image does
- **ISR stale data window** — ISR shows cached content during revalidation; for financial or real-time data, use `no-store`
- **Edge vs Node runtime** — middleware always runs on Edge; `runtime = "edge"` on API routes opts them in; default is Node
- **`revalidatePath` scope** — it revalidates all routes that include the path; `/blog` revalidates `/blog`, `/blog/my-post`, etc.
