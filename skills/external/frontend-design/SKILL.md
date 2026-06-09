# Frontend Design (via anthropics/skills)

---
name: frontend-design
description: >
  Use when designing or building UI components — covers layout systems, color theory,
  typography, spacing, accessibility, responsive design, and visual hierarchy for web
  and mobile interfaces.
license: MIT
source: https://github.com/anthropics/skills
---

## When to Use

Activate this skill when the task involves:

- Designing or critiquing visual layouts (grid, flexbox, spacing, hierarchy)
- Choosing or applying color palettes, contrast ratios, and accessible color pairings
- Setting typographic scales, font pairing, and line-height conventions
- Building or reviewing responsive layouts for mobile, tablet, and desktop breakpoints
- Writing or auditing CSS/Tailwind for consistency and maintainability
- Creating component design systems or design tokens
- Reviewing UI for WCAG accessibility compliance

## Core Principles

### Visual Hierarchy

Apply size, weight, contrast, and whitespace to guide the user's eye:

- **Scale**: Headlines 2×–3× larger than body; captions 0.75× body
- **Weight**: Bold sparingly — one level of emphasis per section
- **Contrast**: Minimum 4.5:1 for normal text (WCAG AA); 3:1 for large text
- **Whitespace**: Generous padding signals premium; cramped padding signals urgency

### Layout Systems

```css
/* 8-point spacing grid */
:root {
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 16px;
  --space-4: 24px;
  --space-5: 32px;
  --space-6: 48px;
  --space-7: 64px;
}

/* Responsive container */
.container {
  width: min(100% - 2rem, 72ch);
  margin-inline: auto;
}
```

### Color System

```css
/* Design token approach — semantic, not literal */
:root {
  --color-primary: hsl(220 90% 56%);
  --color-primary-hover: hsl(220 90% 46%);
  --color-surface: hsl(0 0% 100%);
  --color-surface-raised: hsl(220 13% 97%);
  --color-text: hsl(220 13% 13%);
  --color-text-muted: hsl(220 9% 46%);
  --color-border: hsl(220 13% 87%);
}
```

Always test colors for contrast. Never use color alone to convey meaning (add icon or text).

### Typography Scale

```css
:root {
  --font-size-xs: 0.75rem;   /* 12px — captions, labels */
  --font-size-sm: 0.875rem;  /* 14px — secondary body */
  --font-size-base: 1rem;    /* 16px — body */
  --font-size-lg: 1.125rem;  /* 18px — large body */
  --font-size-xl: 1.25rem;   /* 20px — small heading */
  --font-size-2xl: 1.5rem;   /* 24px — h3 */
  --font-size-3xl: 1.875rem; /* 30px — h2 */
  --font-size-4xl: 2.25rem;  /* 36px — h1 */
  --line-height-tight: 1.25;
  --line-height-normal: 1.5;
  --line-height-relaxed: 1.75;
}
```

## Component Patterns

### Accessible Button

```html
<button
  type="button"
  class="btn btn-primary"
  aria-label="Save document"
>
  Save
</button>
```

```css
.btn {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  border-radius: 0.375rem;
  font-size: var(--font-size-sm);
  font-weight: 500;
  cursor: pointer;
  transition: background-color 150ms ease, box-shadow 150ms ease;
}

.btn:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.btn-primary {
  background: var(--color-primary);
  color: white;
  border: none;
}

.btn-primary:hover {
  background: var(--color-primary-hover);
}
```

### Card Component

```css
.card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 0.5rem;
  padding: var(--space-4);
  box-shadow: 0 1px 3px hsl(0 0% 0% / 0.1);
}
```

## Accessibility Checklist

Before shipping any UI:

- [ ] All interactive elements reachable and operable by keyboard
- [ ] Focus indicators visible (never `outline: none` without a replacement)
- [ ] Images have descriptive `alt` text (or `alt=""` for decorative images)
- [ ] Form inputs have associated `<label>` elements
- [ ] ARIA roles and attributes used correctly (prefer native HTML semantics first)
- [ ] Color contrast meets WCAG AA (4.5:1 normal text, 3:1 large text)
- [ ] Motion respects `prefers-reduced-motion`

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
```

## Responsive Design

```css
/* Mobile-first breakpoints */
/* sm: 640px, md: 768px, lg: 1024px, xl: 1280px */

.grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--space-4);
}

@media (min-width: 640px) {
  .grid { grid-template-columns: repeat(2, 1fr); }
}

@media (min-width: 1024px) {
  .grid { grid-template-columns: repeat(3, 1fr); }
}
```

## Reference Files

| Task | Reference File |
|------|---------------|
| Design system tokens and Tailwind config | `references/design-system.md` |
| Accessibility deep-dive (ARIA, screen readers) | `references/accessibility.md` |
| Animation and motion patterns | `references/motion.md` |

## Common Gotchas

- **Font loading CLS** — preload fonts and use `font-display: swap` to avoid layout shift
- **Dark mode flicker** — set `color-scheme` on `<html>` and read `prefers-color-scheme` in CSS, not JS
- **`z-index` wars** — define a stacking context map in tokens; never use arbitrary large values
- **Flexbox alignment confusion** — `align-items` aligns on cross axis; `justify-content` aligns on main axis
- **Touch target size** — minimum 44×44px for all interactive elements (WCAG 2.5.5)
