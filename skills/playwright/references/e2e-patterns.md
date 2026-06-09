## Overview

End-to-end test patterns for Playwright in TypeScript. Covers the Page Object Model for maintainability, locator strategies for robustness, auth state sharing for speed, and retry/timeout configuration for CI reliability.

## Common Commands

### Page Object Model (POM)

```typescript
// pages/LoginPage.ts
import { Page, Locator } from '@playwright/test';

export class LoginPage {
  readonly page: Page;
  readonly emailInput: Locator;
  readonly passwordInput: Locator;
  readonly submitButton: Locator;
  readonly errorMessage: Locator;

  constructor(page: Page) {
    this.page = page;
    this.emailInput = page.getByLabel('Email');
    this.passwordInput = page.getByLabel('Password');
    this.submitButton = page.getByRole('button', { name: 'Sign in' });
    this.errorMessage = page.getByRole('alert');
  }

  async goto() {
    await this.page.goto('/login');
  }

  async login(email: string, password: string) {
    await this.emailInput.fill(email);
    await this.passwordInput.fill(password);
    await this.submitButton.click();
  }

  async getErrorText(): Promise<string> {
    return this.errorMessage.textContent() ?? '';
  }
}

// tests/login.spec.ts
import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';

test('successful login', async ({ page }) => {
  const loginPage = new LoginPage(page);
  await loginPage.goto();
  await loginPage.login('user@example.com', 'password');
  await expect(page).toHaveURL('/dashboard');
});
```

### Locator Strategies

```typescript
// PREFERRED: Semantic locators (robust to CSS/DOM changes)
page.getByRole('button', { name: 'Submit' })        // ARIA role + name
page.getByRole('textbox', { name: 'Username' })
page.getByRole('link', { name: 'Learn more' })
page.getByLabel('Email address')                     // <label> association
page.getByPlaceholder('Search products')             // placeholder attr
page.getByAltText('Company logo')                    // <img alt>
page.getByTestId('submit-btn')                       // data-testid attr

// FALLBACK: CSS selectors (use sparingly, prefer semantic)
page.locator('[data-qa="submit"]')                   // custom test attr
page.locator('input[type="email"]')                  // attribute selector

// Chaining locators
page.getByRole('region', { name: 'Checkout' })
  .getByRole('button', { name: 'Place order' })

// Filtering
page.getByRole('listitem').filter({ hasText: 'Alice' })
page.getByRole('row').filter({ has: page.getByRole('checkbox', { checked: true }) })
```

### Auth State Sharing (Fixtures)

Reusing authenticated session across tests avoids repeated login in every test.

```typescript
// playwright.config.ts
import { defineConfig } from '@playwright/test';
export default defineConfig({
  globalSetup: './global-setup.ts',
  use: {
    storageState: 'playwright/.auth/user.json',
  },
});

// global-setup.ts
import { chromium, FullConfig } from '@playwright/test';

async function globalSetup(config: FullConfig) {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  await page.goto('http://localhost:3000/login');
  await page.getByLabel('Email').fill('user@example.com');
  await page.getByLabel('Password').fill('password');
  await page.getByRole('button', { name: 'Sign in' }).click();
  await page.waitForURL('/dashboard');
  await page.context().storageState({ path: 'playwright/.auth/user.json' });
  await browser.close();
}

export default globalSetup;

// For tests that need a fresh (unauthenticated) context
test('login page redirects', async ({ browser }) => {
  const context = await browser.newContext();  // no storageState
  const page = await context.newPage();
  // ...
});
```

### Async/Await Patterns

```typescript
// Navigation
await page.goto('https://example.com');
await page.waitForURL('**/dashboard');                     // URL pattern

// Waiting for state
await page.getByRole('button', { name: 'Load' }).click();
await page.waitForLoadState('domcontentloaded');            // DOM ready
// Avoid: await page.waitForLoadState('networkidle');       // unreliable on SPAs

// Wait for element
await expect(page.getByRole('status')).toHaveText('Saved', { timeout: 5000 });

// Wait for network request
const [response] = await Promise.all([
  page.waitForResponse('**/api/save'),
  page.getByRole('button', { name: 'Save' }).click(),
]);
expect(response.status()).toBe(200);

// Intercept and mock network
await page.route('**/api/users', route => {
  route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify([{ id: 1, name: 'Alice' }]),
  });
});
```

### Test Fixtures for Complex Setup

```typescript
// fixtures.ts
import { test as base } from '@playwright/test';
import { LoginPage } from './pages/LoginPage';

type Fixtures = { loginPage: LoginPage };

export const test = base.extend<Fixtures>({
  loginPage: async ({ page }, use) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await use(loginPage);
  },
});

// usage
import { test } from '../fixtures';
import { expect } from '@playwright/test';

test('shows error on wrong password', async ({ loginPage }) => {
  await loginPage.login('user@example.com', 'wrong');
  expect(await loginPage.getErrorText()).toContain('Invalid credentials');
});
```

### Retries and Timeouts

```typescript
// playwright.config.ts
import { defineConfig } from '@playwright/test';
export default defineConfig({
  timeout: 30_000,           // per-test timeout (default 30s)
  expect: {
    timeout: 5_000,          // per-assertion timeout
  },
  retries: process.env.CI ? 2 : 0,   // retry flaky tests in CI
  reporter: [['html'], ['list']],
  use: {
    actionTimeout: 10_000,   // per-action timeout (click, fill, etc.)
    trace: 'on-first-retry', // record trace on first retry for debugging
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
});

// Per-test timeout override
test('slow test', async ({ page }) => {
  test.setTimeout(60_000);
  // ...
});
```

## Patterns

### Data-Driven Tests

```typescript
const credentials = [
  { email: 'admin@example.com', role: 'Admin' },
  { email: 'user@example.com', role: 'User' },
];

for (const { email, role } of credentials) {
  test(`${role} sees correct dashboard`, async ({ page }) => {
    await page.goto('/login');
    await page.getByLabel('Email').fill(email);
    await page.getByLabel('Password').fill('password');
    await page.getByRole('button', { name: 'Sign in' }).click();
    await expect(page.getByRole('heading')).toHaveText(`${role} Dashboard`);
  });
}
```

### Visual Regression

```typescript
// Capture and compare screenshot (requires baseline)
await expect(page).toHaveScreenshot('dashboard.png');

// Update baselines
// npx playwright test --update-snapshots
```

## Gotchas

- **`waitForLoadState('networkidle')` unreliable** — SPAs with polling or WebSocket connections never reach networkidle; use element-based waits instead
- **Strict mode violations** — if `page.getByRole('button')` matches two buttons, Playwright throws; always use `{ name: '...' }` or chain from a parent locator to be specific
- **`page.evaluate()` vs `locator.evaluate()`** — `page.evaluate()` runs in the browser context (no access to Node.js); use locator.evaluate for element-specific operations
- **Shadow DOM** — Playwright auto-pierces open shadow DOM; for closed shadow DOM, query the host and access shadowRoot via `locator.evaluate(el => el.shadowRoot.querySelector(...))`
- **iframes** — use `page.frameLocator('iframe[name="checkout"]')` to interact with content inside iframes; regular locators don't cross frame boundaries
- **Slow CI machines** — increase `actionTimeout` and `expect.timeout` in CI; what passes locally may time out on slow CI runners
