# Edge Middleware Reference

## What is Middleware in Next.js

Middleware runs at the edge (before the cache, before the route handler) on every
matching request. It has access to the request object and can rewrite, redirect, set
headers, or return a response directly.

Middleware uses the Edge Runtime — no Node.js APIs, web APIs only.

## Setup

```typescript
// middleware.ts (project root, next to app/)
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
  // Runs on every request that matches the config below
  return NextResponse.next();
}

export const config = {
  // Only run on these paths (avoids static assets)
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
```

## Common Patterns

### Authentication Guard

```typescript
export function middleware(request: NextRequest) {
  const token = request.cookies.get("session")?.value;
  const isProtected = request.nextUrl.pathname.startsWith("/dashboard");

  if (isProtected && !token) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  return NextResponse.next();
}
```

### A/B Testing

```typescript
const EXPERIMENT_COOKIE = "ab-variant";
const VARIANTS = ["control", "treatment-a", "treatment-b"];

export function middleware(request: NextRequest) {
  const response = NextResponse.next();
  let variant = request.cookies.get(EXPERIMENT_COOKIE)?.value;

  if (!variant || !VARIANTS.includes(variant)) {
    variant = VARIANTS[Math.floor(Math.random() * VARIANTS.length)];
    response.cookies.set(EXPERIMENT_COOKIE, variant, { maxAge: 60 * 60 * 24 * 30 });
  }

  // Rewrite to variant-specific route
  if (request.nextUrl.pathname === "/pricing") {
    return NextResponse.rewrite(new URL(`/pricing/${variant}`, request.url));
  }

  return response;
}
```

### Internationalization (i18n)

```typescript
import { match } from "@formatjs/intl-localematcher";
import Negotiator from "negotiator";

const LOCALES = ["en", "fr", "de"];
const DEFAULT_LOCALE = "en";

function getLocale(request: NextRequest): string {
  const acceptLanguage = request.headers.get("accept-language") ?? "";
  const languages = new Negotiator({ headers: { "accept-language": acceptLanguage } })
    .languages();
  return match(languages, LOCALES, DEFAULT_LOCALE);
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const hasLocale = LOCALES.some(l => pathname.startsWith(`/${l}/`) || pathname === `/${l}`);

  if (!hasLocale) {
    const locale = getLocale(request);
    return NextResponse.redirect(new URL(`/${locale}${pathname}`, request.url));
  }
}
```

### Security Headers

```typescript
export function middleware(request: NextRequest) {
  const response = NextResponse.next();

  response.headers.set("X-Frame-Options", "DENY");
  response.headers.set("X-Content-Type-Options", "nosniff");
  response.headers.set("Referrer-Policy", "strict-origin-when-cross-origin");
  response.headers.set(
    "Content-Security-Policy",
    "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'"
  );

  return response;
}
```

## Feature Flags at the Edge

```typescript
// Evaluate flags from a KV store or edge config
import { get } from "@vercel/edge-config";

export async function middleware(request: NextRequest) {
  const newCheckout = await get<boolean>("enable-new-checkout");

  if (newCheckout && request.nextUrl.pathname === "/checkout") {
    return NextResponse.rewrite(new URL("/checkout-v2", request.url));
  }

  return NextResponse.next();
}
```

Vercel Edge Config (`@vercel/edge-config`) provides sub-1ms global config reads.

## Middleware Limitations

- No Node.js APIs — use Web APIs only (`fetch`, `crypto.subtle`, `Request`/`Response`)
- 1MB size limit for the middleware bundle
- 25ms CPU time limit on Vercel (soft limit, extendable)
- Cannot read response body (only request body)
- Cannot write to filesystem
