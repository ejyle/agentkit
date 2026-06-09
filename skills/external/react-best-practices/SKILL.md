# React Best Practices (via vercel-labs/agent-skills)

---
name: react-best-practices
description: >
  Use when building or reviewing React applications — covers component design, hooks
  patterns, state management, performance optimization, error boundaries, and modern
  React 18+ patterns including concurrent features and Server Components.
license: MIT
source: https://github.com/vercel-labs/agent-skills
---

## When to Use

Activate this skill when the task involves:

- Designing or refactoring React component architecture
- Choosing between useState, useReducer, useContext, or external state
- Optimizing renders with memo, useMemo, useCallback
- Implementing error boundaries and suspense boundaries
- Working with React 18 concurrent features (transitions, deferred values)
- Using React Server Components in Next.js or similar frameworks
- Writing hooks that are correct, composable, and testable

## Component Design

### Single Responsibility

Each component should do one thing. Split when a component:
- Has more than one reason to change
- Renders more than one distinct visual section
- Manages both UI state and business logic

```tsx
// BAD: God component
function UserProfile({ userId }: { userId: string }) {
  const [user, setUser] = useState(null);
  const [posts, setPosts] = useState([]);
  // ... 200 lines of mixed concerns
}

// GOOD: Separated by responsibility
function UserProfilePage({ userId }: { userId: string }) {
  return (
    <>
      <UserCard userId={userId} />
      <UserPostFeed userId={userId} />
    </>
  );
}
```

### Props Design

```tsx
// Use explicit prop interfaces, not spreading unknown props
interface ButtonProps {
  label: string;
  onClick: () => void;
  variant?: "primary" | "secondary" | "ghost";
  disabled?: boolean;
  loading?: boolean;
}

// Composition over prop explosion
// BAD: 20 props for all icon positions
function BadButton({ iconLeft, iconRight, iconTop, ... }) {}

// GOOD: Slot pattern
function Button({ children, className, ...props }: ButtonProps & React.ButtonHTMLAttributes<HTMLButtonElement>) {
  return <button className={cn("btn", className)} {...props}>{children}</button>;
}
```

## Hooks Patterns

### useState vs useReducer

Use `useState` for independent, simple values. Use `useReducer` when:
- Multiple state values are updated together
- Next state depends on prior state in a complex way
- State transitions need to be testable in isolation

```tsx
// Prefer useReducer for form state
type FormState = { name: string; email: string; errors: Record<string, string> };
type FormAction =
  | { type: "set_field"; field: keyof FormState; value: string }
  | { type: "set_errors"; errors: Record<string, string> }
  | { type: "reset" };

function formReducer(state: FormState, action: FormAction): FormState {
  switch (action.type) {
    case "set_field": return { ...state, [action.field]: action.value };
    case "set_errors": return { ...state, errors: action.errors };
    case "reset": return { name: "", email: "", errors: {} };
  }
}
```

### Custom Hooks

Extract stateful logic into custom hooks when it's reused or complex:

```tsx
function useLocalStorage<T>(key: string, initial: T) {
  const [value, setValue] = useState<T>(() => {
    try {
      const item = window.localStorage.getItem(key);
      return item ? JSON.parse(item) : initial;
    } catch {
      return initial;
    }
  });

  const setStored = useCallback((newValue: T | ((v: T) => T)) => {
    setValue((prev) => {
      const next = typeof newValue === "function" ? (newValue as (v: T) => T)(prev) : newValue;
      window.localStorage.setItem(key, JSON.stringify(next));
      return next;
    });
  }, [key]);

  return [value, setStored] as const;
}
```

## Performance

### Memoization — Only When Measured

```tsx
// memo only prevents re-render when props are stable references
const ExpensiveList = React.memo(function ExpensiveList({ items }: { items: Item[] }) {
  return <ul>{items.map(item => <li key={item.id}>{item.name}</li>)}</ul>;
});

// useMemo for expensive derivations
const sortedItems = useMemo(
  () => [...items].sort((a, b) => a.name.localeCompare(b.name)),
  [items]
);

// useCallback for stable function references passed to memo'd children
const handleDelete = useCallback((id: string) => {
  setItems(prev => prev.filter(item => item.id !== id));
}, []); // empty dep — no captured state
```

### Transitions (React 18)

```tsx
import { startTransition, useTransition } from "react";

function SearchPage() {
  const [isPending, startTransition] = useTransition();
  const [query, setQuery] = useState("");
  const [results, setResults] = useState([]);

  function handleSearch(value: string) {
    setQuery(value); // urgent — update input immediately
    startTransition(() => {
      setResults(computeResults(value)); // non-urgent — can be deferred
    });
  }

  return (
    <>
      <input value={query} onChange={e => handleSearch(e.target.value)} />
      {isPending ? <Spinner /> : <ResultList items={results} />}
    </>
  );
}
```

## Error Boundaries

```tsx
import { Component, ReactNode } from "react";

interface Props { children: ReactNode; fallback?: ReactNode; }
interface State { hasError: boolean; error?: Error; }

class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false };

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: { componentStack: string }) {
    console.error("Uncaught error:", error, info.componentStack);
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback ?? <p>Something went wrong.</p>;
    }
    return this.props.children;
  }
}

// Usage
<ErrorBoundary fallback={<ErrorScreen />}>
  <App />
</ErrorBoundary>
```

## Reference Files

| Task | Reference File |
|------|---------------|
| React Server Components, Suspense, data fetching | `references/server-components.md` |
| Testing hooks and components with React Testing Library | `references/testing.md` |

## Common Gotchas

- **Stale closure in useEffect** — always include all values read inside the effect in the dependency array; use the exhaustive-deps ESLint rule
- **Object/array as deps** — new object/array on every render triggers the effect every render; memoize or extract the value
- **Keys in lists** — use stable, unique IDs as keys; never use array index as key in dynamic lists
- **useEffect for data fetching** — prefer React Query, SWR, or RSC; raw `useEffect` data fetching has race conditions
- **Context re-render storm** — split context into value-context and dispatch-context to prevent all consumers re-rendering on any state change
