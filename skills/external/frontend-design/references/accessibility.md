# Accessibility Deep Dive

## WCAG 2.1 Conformance Levels

- **Level A**: Minimum — missing this makes content inaccessible to some users
- **Level AA**: Standard target — required by most regulations and contracts
- **Level AAA**: Enhanced — aim for where possible, not always achievable for all content

## ARIA Roles and Landmarks

Use native HTML semantics first. ARIA is a patch, not a replacement.

```html
<!-- Landmark roles (screen readers build a page map from these) -->
<header role="banner">...</header>
<nav role="navigation" aria-label="Main">...</nav>
<main role="main">...</main>
<aside role="complementary">...</aside>
<footer role="contentinfo">...</footer>

<!-- Widget roles -->
<button role="button" aria-pressed="false">Toggle</button>
<div role="dialog" aria-modal="true" aria-labelledby="dialog-title">
  <h2 id="dialog-title">Confirm Delete</h2>
</div>
<ul role="listbox" aria-label="Options">
  <li role="option" aria-selected="true">Option 1</li>
</ul>
```

## Keyboard Navigation Requirements

Every interactive element must be:
1. Reachable by Tab key
2. Operable by Enter/Space
3. Have a visible focus indicator

```css
/* Never remove focus without replacement */
:focus-visible {
  outline: 2px solid hsl(220 90% 56%);
  outline-offset: 2px;
  border-radius: 2px;
}

/* Hide outline for mouse users, show for keyboard */
:focus:not(:focus-visible) { outline: none; }
```

### Focus Trap (for Modals)

```typescript
function trapFocus(container: HTMLElement) {
  const focusable = container.querySelectorAll<HTMLElement>(
    'a, button, input, select, textarea, [tabindex]:not([tabindex="-1"])'
  );
  const first = focusable[0];
  const last = focusable[focusable.length - 1];

  container.addEventListener("keydown", (e) => {
    if (e.key !== "Tab") return;
    if (e.shiftKey) {
      if (document.activeElement === first) { e.preventDefault(); last.focus(); }
    } else {
      if (document.activeElement === last) { e.preventDefault(); first.focus(); }
    }
  });

  first.focus();
}
```

## Screen Reader Announcements

```tsx
// Live regions for dynamic content
<div aria-live="polite" aria-atomic="true">
  {statusMessage}
</div>

// For urgent announcements (error, alert)
<div role="alert">
  Error: Please fill in all required fields.
</div>

// Visually hidden but readable by screen reader
<span className="sr-only">Loading...</span>
```

```css
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
```

## Testing Tools

- **axe DevTools** browser extension — automated WCAG audit
- **Lighthouse** — `npx lighthouse <url> --view` — accessibility score
- **NVDA** (Windows) / **VoiceOver** (macOS/iOS) — manual screen reader testing
- **@axe-core/react** — integrate axe into development mode

```bash
npm install --save-dev @axe-core/react
```

```typescript
// Run in development only
if (process.env.NODE_ENV !== "production") {
  const { default: axe } = await import("@axe-core/react");
  const { default: React } = await import("react");
  const { default: ReactDOM } = await import("react-dom");
  axe(React, ReactDOM, 1000);
}
```
