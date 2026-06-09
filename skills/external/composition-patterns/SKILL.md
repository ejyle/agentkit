# Composition Patterns (via vercel-labs/agent-skills)

---
name: composition-patterns
description: >
  Use when designing flexible, reusable component APIs in React — covers compound
  components, render props, slot patterns, headless components, and provider patterns
  for building UI libraries and design systems.
license: MIT
source: https://github.com/vercel-labs/agent-skills
---

## When to Use

Activate this skill when the task involves:

- Building a component API that needs to be flexible without a prop explosion
- Designing reusable UI library components (modals, dropdowns, tabs, accordions)
- Refactoring tightly coupled components for better composability
- Implementing headless (behavior-only) components consumed by styled wrappers
- Choosing between compound components, render props, and slot/children patterns

## Core Patterns

### 1. Compound Components

A parent component shares implicit context with specific child components.
The parent owns state; children subscribe to it through context.

```tsx
import { createContext, useContext, useState } from "react";

// --- Tabs implementation ---
interface TabsCtx {
  active: string;
  setActive: (id: string) => void;
}

const TabsContext = createContext<TabsCtx | null>(null);

function useTabs() {
  const ctx = useContext(TabsContext);
  if (!ctx) throw new Error("useTabs must be used within <Tabs>");
  return ctx;
}

function Tabs({ defaultValue, children }: { defaultValue: string; children: React.ReactNode }) {
  const [active, setActive] = useState(defaultValue);
  return (
    <TabsContext.Provider value={{ active, setActive }}>
      <div className="tabs">{children}</div>
    </TabsContext.Provider>
  );
}

function TabsList({ children }: { children: React.ReactNode }) {
  return <div role="tablist" className="tabs-list">{children}</div>;
}

function TabsTrigger({ value, children }: { value: string; children: React.ReactNode }) {
  const { active, setActive } = useTabs();
  return (
    <button
      role="tab"
      aria-selected={active === value}
      onClick={() => setActive(value)}
    >
      {children}
    </button>
  );
}

function TabsContent({ value, children }: { value: string; children: React.ReactNode }) {
  const { active } = useTabs();
  if (active !== value) return null;
  return <div role="tabpanel">{children}</div>;
}

// Attach sub-components for ergonomic API
Tabs.List = TabsList;
Tabs.Trigger = TabsTrigger;
Tabs.Content = TabsContent;

// Consumer API
<Tabs defaultValue="account">
  <Tabs.List>
    <Tabs.Trigger value="account">Account</Tabs.Trigger>
    <Tabs.Trigger value="billing">Billing</Tabs.Trigger>
  </Tabs.List>
  <Tabs.Content value="account"><AccountForm /></Tabs.Content>
  <Tabs.Content value="billing"><BillingForm /></Tabs.Content>
</Tabs>
```

### 2. Render Props

Pass a function as a child or prop to give the consumer control over rendering.
Use when the consumer needs access to internal state for layout, not just behavior.

```tsx
interface DisclosureProps {
  children: (state: { isOpen: boolean; toggle: () => void }) => React.ReactNode;
  defaultOpen?: boolean;
}

function Disclosure({ children, defaultOpen = false }: DisclosureProps) {
  const [isOpen, setIsOpen] = useState(defaultOpen);
  const toggle = () => setIsOpen(v => !v);
  return <>{children({ isOpen, toggle })}</>;
}

// Usage — consumer decides the layout
<Disclosure>
  {({ isOpen, toggle }) => (
    <div>
      <button onClick={toggle}>{isOpen ? "Collapse" : "Expand"}</button>
      {isOpen && <details>Hidden content</details>}
    </div>
  )}
</Disclosure>
```

### 3. Slot Pattern (asChild)

Allow consumers to pass their own element as the rendered root, merging behavior.
Used extensively in Radix UI.

```tsx
import { Slot } from "@radix-ui/react-slot";

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  asChild?: boolean;
  variant?: "primary" | "ghost";
}

function Button({ asChild = false, variant = "primary", className, ...props }: ButtonProps) {
  const Comp = asChild ? Slot : "button";
  return <Comp className={cn("btn", `btn-${variant}`, className)} {...props} />;
}

// Usage — renders as <a> with button behavior
<Button asChild>
  <a href="/dashboard">Go to Dashboard</a>
</Button>
```

### 4. Headless Components

Separate behavior (state, a11y, keyboard) from appearance. The headless layer
provides hooks and primitives; the styled layer owns every CSS class.

```tsx
// Headless hook — no styling
function useToggle(initial = false) {
  const [on, setOn] = useState(initial);
  return {
    on,
    toggle: () => setOn(v => !v),
    setOn,
    setOff: () => setOn(false),
    // Spread onto the trigger element for a11y
    triggerProps: {
      onClick: () => setOn(v => !v),
      "aria-expanded": on,
      "aria-pressed": on,
    },
  };
}

// Styled consumer
function DarkModeToggle() {
  const { on, triggerProps } = useToggle(false);
  return (
    <button
      className={cn("toggle", on && "toggle-active")}
      {...triggerProps}
    >
      {on ? "Dark" : "Light"}
    </button>
  );
}
```

### 5. Provider Pattern

Share global or feature-scoped state without prop drilling.

```tsx
interface ThemeContextValue {
  theme: "light" | "dark";
  toggleTheme: () => void;
}

const ThemeContext = createContext<ThemeContextValue | null>(null);

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setTheme] = useState<"light" | "dark">("light");
  const toggleTheme = useCallback(() => {
    setTheme(t => (t === "light" ? "dark" : "light"));
  }, []);
  return (
    <ThemeContext.Provider value={{ theme, toggleTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new Error("useTheme must be used within ThemeProvider");
  return ctx;
}
```

## Pattern Selection Guide

| Need | Pattern |
|------|---------|
| Multiple related components sharing state | Compound Components |
| Consumer controls the rendered output structure | Render Props |
| Consumer supplies their own element type | Slot / asChild |
| Reuse behavior without enforcing style | Headless Hook |
| App-wide or feature-wide shared state | Provider |

## Common Gotchas

- **Context value reference stability** — wrap context value in `useMemo` to prevent consumers re-rendering when parent re-renders with the same logical state
- **Compound component displayName** — set `ComponentName.displayName = "Tabs.Trigger"` for clear React DevTools tree
- **Render props and `memo`** — render prop functions are recreated each render; wrap the consumer in `useCallback` or `useMemo` if it's inside a memo boundary
- **Slot and ref forwarding** — when using `asChild`, forward refs through the slot so parent code can access the DOM node
