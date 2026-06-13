# React View Transitions (via vercel-labs/agent-skills)

---
name: react-view-transitions
description: >
  Use when implementing animated page or component transitions in React using the
  View Transitions API — covers shared element transitions, route animations in
  Next.js and React Router, progressive enhancement patterns, and browser support
  fallbacks.
license: MIT
source: https://github.com/vercel-labs/agent-skills
---

## When to Use

Activate this skill when the task involves:

- Adding animated transitions between pages or routes in React apps
- Implementing shared element transitions (e.g. a card expanding to a detail view)
- Using the browser-native View Transitions API with React 18+
- Integrating view transitions with Next.js App Router navigation
- Providing graceful fallbacks for browsers without View Transitions support

## View Transitions API Overview

The View Transitions API enables smooth animated DOM transitions without JavaScript
animation libraries. The browser captures a screenshot of the old state and the new
DOM, then animates between them using CSS.

```javascript
// Bare API — works in any framework
document.startViewTransition(() => {
  // Mutate DOM here — browser animates old → new
  setNewContent();
});
```

**Browser support (June 2025):** Chrome 111+, Edge 111+, Safari 18+, Firefox 130+

## React 18 Integration

React 18's `startTransition` and the View Transitions API compose naturally:

```tsx
import { startTransition, useState } from "react";

function App() {
  const [page, setPage] = useState("home");

  function navigate(to: string) {
    if (!document.startViewTransition) {
      startTransition(() => setPage(to));
      return;
    }
    document.startViewTransition(() => {
      startTransition(() => setPage(to));
    });
  }

  return (
    <>
      <nav>
        <button onClick={() => navigate("home")}>Home</button>
        <button onClick={() => navigate("about")}>About</button>
      </nav>
      {page === "home" && <HomePage />}
      {page === "about" && <AboutPage />}
    </>
  );
}
```

## Next.js App Router

Next.js 14+ exposes a `useRouter` hook whose `push`/`replace` calls can be wrapped
in `startViewTransition` for animated navigation:

```tsx
"use client";
import { useRouter } from "next/navigation";
import { useCallback } from "react";

export function useViewTransitionRouter() {
  const router = useRouter();

  const navigate = useCallback(
    (href: string, options?: { replace?: boolean }) => {
      if (!document.startViewTransition) {
        options?.replace ? router.replace(href) : router.push(href);
        return;
      }
      document.startViewTransition(() => {
        options?.replace ? router.replace(href) : router.push(href);
      });
    },
    [router]
  );

  return navigate;
}

// Usage in any client component
function NavLink({ href, children }: { href: string; children: React.ReactNode }) {
  const navigate = useViewTransitionRouter();
  return <a onClick={() => navigate(href)}>{children}</a>;
}
```

## Shared Element Transitions

Use `view-transition-name` to animate a specific element from one route to another.
The name must be unique per frame.

```css
/* On the list page */
.card[data-id="42"] {
  view-transition-name: card-42;
}

/* On the detail page */
.detail-hero[data-id="42"] {
  view-transition-name: card-42;
}
```

```tsx
// Set view-transition-name via inline style for dynamic IDs
function ProductCard({ product }: { product: Product }) {
  return (
    <div
      className="card"
      style={{ viewTransitionName: `product-${product.id}` } as React.CSSProperties}
    >
      <img src={product.image} alt={product.name} />
      <h2>{product.name}</h2>
    </div>
  );
}
```

## Customizing Animations

```css
/* Default cross-fade — override the pseudo-elements */
::view-transition-old(root) {
  animation: 200ms ease-out fade-out;
}

::view-transition-new(root) {
  animation: 200ms ease-in fade-in;
}

/* Slide transition for named elements */
::view-transition-old(card-42) {
  animation: 300ms ease-in-out slide-out;
}

::view-transition-new(card-42) {
  animation: 300ms ease-in-out slide-in;
}

@keyframes fade-out { from { opacity: 1; } to { opacity: 0; } }
@keyframes fade-in  { from { opacity: 0; } to { opacity: 1; } }
@keyframes slide-out { to { transform: translateY(-20px); opacity: 0; } }
@keyframes slide-in  { from { transform: translateY(20px); opacity: 0; } }
```

### Respect Reduced Motion

```css
@media (prefers-reduced-motion: reduce) {
  ::view-transition-group(*),
  ::view-transition-old(*),
  ::view-transition-new(*) {
    animation-duration: 0.01ms !important;
  }
}
```

## React Router v6

```tsx
import { useNavigate } from "react-router-dom";

function Link({ to, children }: { to: string; children: React.ReactNode }) {
  const navigate = useNavigate();

  function handleClick(e: React.MouseEvent) {
    e.preventDefault();
    if (!document.startViewTransition) {
      navigate(to);
      return;
    }
    document.startViewTransition(() => navigate(to));
  }

  return <a href={to} onClick={handleClick}>{children}</a>;
}
```

## Progressive Enhancement

Always check for API support before using:

```typescript
const supportsViewTransitions = typeof document !== "undefined"
  && "startViewTransition" in document;

function withViewTransition(fn: () => void) {
  if (supportsViewTransitions) {
    (document as any).startViewTransition(fn);
  } else {
    fn(); // immediate, no animation
  }
}
```

## Common Gotchas

- **Duplicate `view-transition-name`** — two elements with the same name in the same frame causes the transition to abort; names must be unique per snapshot
- **SSR environments** — `document.startViewTransition` is undefined on the server; always guard with `typeof document !== "undefined"`
- **Long transitions block interaction** — the page is non-interactive during a view transition; keep animations under 300ms
- **`view-transition-name` cannot be `none`** — that value disables the transition for the element; use `unset` to remove a previously set name
- **Flash of white between routes** — caused by the new page rendering before styles load; ensure critical CSS is inlined or preloaded
