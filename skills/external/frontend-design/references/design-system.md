# Design System Reference

## Design Tokens

Design tokens are named variables that encode every visual decision — colors, spacing, type,
shadows, radii — in a single source of truth consumed by CSS, Tailwind, and component libraries.

### Full Token Set

```css
:root {
  /* --- Colors --- */
  /* Brand */
  --color-brand-50:  hsl(220 100% 97%);
  --color-brand-100: hsl(220 95%  93%);
  --color-brand-200: hsl(220 92%  86%);
  --color-brand-300: hsl(220 90%  74%);
  --color-brand-400: hsl(220 90%  63%);
  --color-brand-500: hsl(220 90%  56%); /* primary */
  --color-brand-600: hsl(220 85%  46%);
  --color-brand-700: hsl(220 80%  38%);
  --color-brand-800: hsl(220 76%  30%);
  --color-brand-900: hsl(220 72%  22%);

  /* Neutrals */
  --color-neutral-0:   hsl(0 0% 100%);
  --color-neutral-50:  hsl(220 13% 97%);
  --color-neutral-100: hsl(220 13% 94%);
  --color-neutral-200: hsl(220 11% 87%);
  --color-neutral-300: hsl(220 10% 76%);
  --color-neutral-400: hsl(220 9%  60%);
  --color-neutral-500: hsl(220 9%  46%);
  --color-neutral-600: hsl(220 10% 36%);
  --color-neutral-700: hsl(220 11% 27%);
  --color-neutral-800: hsl(220 12% 20%);
  --color-neutral-900: hsl(220 13% 13%);

  /* Semantic */
  --color-success: hsl(142 71% 45%);
  --color-warning: hsl(38  92% 50%);
  --color-error:   hsl(0   84% 60%);
  --color-info:    hsl(199 89% 48%);

  /* --- Spacing (8pt grid) --- */
  --space-px:  1px;
  --space-0:   0;
  --space-1:   0.25rem; /* 4px  */
  --space-2:   0.5rem;  /* 8px  */
  --space-3:   0.75rem; /* 12px */
  --space-4:   1rem;    /* 16px */
  --space-5:   1.25rem; /* 20px */
  --space-6:   1.5rem;  /* 24px */
  --space-8:   2rem;    /* 32px */
  --space-10:  2.5rem;  /* 40px */
  --space-12:  3rem;    /* 48px */
  --space-16:  4rem;    /* 64px */
  --space-20:  5rem;    /* 80px */
  --space-24:  6rem;    /* 96px */

  /* --- Typography --- */
  --font-family-sans: "Inter", system-ui, -apple-system, sans-serif;
  --font-family-mono: "JetBrains Mono", "Fira Code", monospace;

  /* --- Radii --- */
  --radius-sm:   0.125rem; /* 2px  */
  --radius-md:   0.375rem; /* 6px  */
  --radius-lg:   0.5rem;   /* 8px  */
  --radius-xl:   0.75rem;  /* 12px */
  --radius-2xl:  1rem;     /* 16px */
  --radius-full: 9999px;

  /* --- Shadows --- */
  --shadow-sm:  0 1px 2px hsl(0 0% 0% / 0.05);
  --shadow-md:  0 4px 6px hsl(0 0% 0% / 0.07), 0 2px 4px hsl(0 0% 0% / 0.05);
  --shadow-lg:  0 10px 15px hsl(0 0% 0% / 0.1), 0 4px 6px hsl(0 0% 0% / 0.05);
  --shadow-xl:  0 20px 25px hsl(0 0% 0% / 0.1), 0 8px 10px hsl(0 0% 0% / 0.05);
}
```

## Tailwind Config Mapping

```js
// tailwind.config.js
module.exports = {
  theme: {
    extend: {
      colors: {
        brand: {
          50: 'hsl(220 100% 97%)',
          // ... map all brand tokens
          500: 'hsl(220 90% 56%)',
          900: 'hsl(220 72% 22%)',
        },
      },
      spacing: {
        '18': '4.5rem',
        '88': '22rem',
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
    },
  },
}
```

## Component Library Integration

When integrating with shadcn/ui, Radix UI, or Headless UI:

1. Map your design tokens to CSS variables the library expects
2. Override only semantic variables, not primitive palette values
3. Keep component overrides in a single `globals.css`, not scattered across component files
