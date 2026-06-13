---
name: playwright
description: >
  Use when writing end-to-end tests, automating browser interactions, capturing screenshots,
  or generating test code using Playwright or the Playwright MCP server.
license: Apache-2.0
---

## When to Use

Activate this skill when the task involves:

- Writing E2E tests for web applications using Playwright
- Automating browser navigation, form filling, button clicks, or file uploads
- Capturing screenshots or PDFs of web pages
- Debugging flaky tests or improving test reliability
- Generating Playwright test code from user stories or feature specs
- Running tests in CI pipelines (headless, across browsers)
- Using the Playwright MCP server for agent-driven browser automation

## Playwright MCP Server

The Playwright MCP server enables AI agents to control a browser programmatically.
It is a separate install — this skill teaches usage patterns only.

```bash
# Install via agentkit (one-time setup)
agentkit install playwright-mcp

# Or install manually
npm install -g @playwright/mcp
```

After installation, the MCP server exposes these key tools to the agent:

| Tool | Action |
|------|--------|
| `browser_navigate` | Go to a URL |
| `browser_click` | Click an element by label/role/selector |
| `browser_type` | Type text into an input |
| `browser_screenshot` | Capture page screenshot |
| `browser_evaluate` | Run JavaScript in the page |
| `browser_wait_for` | Wait for element/navigation |
| `browser_select_option` | Select dropdown option |
| `browser_hover` | Hover over an element |

## Quick Reference

### Project Setup

```bash
# Init Playwright in an existing project
npm init playwright@latest

# Install browsers
npx playwright install

# Install specific browser
npx playwright install chromium

# Run all tests
npx playwright test

# Run specific test file
npx playwright test tests/login.spec.ts

# Run in headed mode (watch browser)
npx playwright test --headed

# Run with UI mode (interactive debugger)
npx playwright test --ui
```

### Writing Tests

```typescript
import { test, expect } from '@playwright/test';

test('user can log in', async ({ page }) => {
  await page.goto('https://example.com/login');
  await page.getByLabel('Email').fill('user@example.com');
  await page.getByLabel('Password').fill('password');
  await page.getByRole('button', { name: 'Sign in' }).click();
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
});
```

### Locator Strategies

```typescript
// Prefer semantic locators (robust to DOM changes)
page.getByRole('button', { name: 'Submit' })
page.getByLabel('Email address')
page.getByPlaceholder('Search...')
page.getByTestId('submit-btn')      // data-testid attribute
page.getByText('Welcome back')

// CSS/XPath as fallback (fragile — avoid for primary selectors)
page.locator('.submit-button')
page.locator('xpath=//button[@type="submit"]')
```

### Assertions

```typescript
await expect(locator).toBeVisible()
await expect(locator).toBeHidden()
await expect(locator).toHaveText('Expected text')
await expect(locator).toContainText('partial text')
await expect(locator).toHaveValue('input value')
await expect(locator).toBeChecked()
await expect(locator).toBeDisabled()
await expect(page).toHaveURL('https://example.com/dashboard')
await expect(page).toHaveTitle('Dashboard')
```

### Debugging

```bash
# Debug mode (pauses at each step)
npx playwright test --debug

# Generate test code by recording actions
npx playwright codegen https://example.com

# Show trace viewer for failed test
npx playwright show-trace test-results/trace.zip
```

## Reference Files

Load the appropriate reference file for deep-dive tasks:

| Task | Reference file |
|------|---------------|
| Page Object Model, locator strategies, auth fixtures, retries | `references/e2e-patterns.md` |

## Common Gotchas

- **`waitForLoadState` vs element waits** — prefer waiting for specific elements (`expect(locator).toBeVisible()`) over `waitForLoadState('networkidle')`; network-idle can be unreliable on SPAs with background requests
- **Strict mode** — Playwright throws if a locator matches multiple elements; use `.first()`, `.nth(n)`, or a more specific locator
- **Parallelism** — Playwright runs test files in parallel by default; tests within a file run serially; use `test.describe.configure({ mode: 'parallel' })` for in-file parallelism
- **Browser context isolation** — each test gets a fresh browser context; shared auth state requires `storageState` fixture
- **Flaky test diagnosis** — use `--retries=2` in CI; check `test-results/` for traces on failures; the trace viewer shows every action, screenshot, and network request
