# Agent Browser (via vercel-labs/agent-browser)

---
name: agent-browser
description: >
  Use when automating web browsing tasks — covers Playwright-based browser automation,
  web scraping, form submission, screenshot capture, and multi-step web workflows
  orchestrated by an AI agent with tool calls.
license: Apache-2.0
source: https://github.com/vercel-labs/agent-browser
---

## When to Use

Activate this skill when the task involves:

- Automating multi-step web workflows (login, form fill, navigate, extract data)
- Scraping web pages that require JavaScript rendering
- Capturing screenshots of pages or elements for verification
- Filling and submitting forms programmatically
- Testing web application flows via browser automation
- Using browser automation as a tool within an AI agent loop

## Core Concepts

The agent-browser pattern wraps Playwright in a set of AI-callable tools so an agent
can control a browser in a tool-use loop. The agent issues tool calls to navigate,
click, type, and read page content without writing Playwright scripts directly.

### Tool-Use Architecture

```
Agent (Claude)
  │
  ├── navigate(url)              → open URL in browser
  ├── click(selector)            → click an element
  ├── type(selector, text)       → fill an input
  ├── read_page()                → return visible text content
  ├── screenshot()               → capture and return image
  ├── wait_for(selector)         → wait for element to appear
  └── evaluate(script)           → run JS in page context
```

## Playwright Setup

```bash
npm install playwright @playwright/test
npx playwright install chromium
```

## Browser Session Management

```typescript
import { chromium, Browser, BrowserContext, Page } from "playwright";

class AgentBrowser {
  private browser: Browser | null = null;
  private context: BrowserContext | null = null;
  private page: Page | null = null;

  async init(options: { headless?: boolean } = {}) {
    this.browser = await chromium.launch({ headless: options.headless ?? true });
    this.context = await this.browser.newContext({
      viewport: { width: 1280, height: 720 },
      userAgent: "Mozilla/5.0 (compatible; AgentBrowser/1.0)",
    });
    this.page = await this.context.newPage();
  }

  async navigate(url: string): Promise<string> {
    await this.page!.goto(url, { waitUntil: "domcontentloaded" });
    return `Navigated to ${url}`;
  }

  async click(selector: string): Promise<string> {
    await this.page!.click(selector, { timeout: 5000 });
    return `Clicked: ${selector}`;
  }

  async type(selector: string, text: string): Promise<string> {
    await this.page!.fill(selector, text);
    return `Filled ${selector} with text`;
  }

  async readPage(): Promise<string> {
    const content = await this.page!.evaluate(() => document.body.innerText);
    return content.slice(0, 4000); // truncate for context window
  }

  async screenshot(): Promise<Buffer> {
    return this.page!.screenshot({ type: "png" });
  }

  async waitFor(selector: string, timeout = 5000): Promise<string> {
    await this.page!.waitForSelector(selector, { timeout });
    return `Element visible: ${selector}`;
  }

  async evaluate(script: string): Promise<unknown> {
    return this.page!.evaluate(script);
  }

  async close() {
    await this.browser?.close();
  }
}
```

## MCP Tool Definitions

Expose browser actions as MCP tools for agent consumption:

```typescript
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { z } from "zod";

const browser = new AgentBrowser();
const server = new McpServer({ name: "agent-browser", version: "1.0.0" });

server.tool("browser_navigate", "Navigate to a URL", {
  url: z.string().url().describe("URL to navigate to"),
}, async ({ url }) => {
  const result = await browser.navigate(url);
  return { content: [{ type: "text", text: result }] };
});

server.tool("browser_click", "Click an element on the page", {
  selector: z.string().describe("CSS selector or text content of element to click"),
}, async ({ selector }) => {
  const result = await browser.click(selector);
  return { content: [{ type: "text", text: result }] };
});

server.tool("browser_type", "Type text into an input field", {
  selector: z.string().describe("CSS selector of the input element"),
  text: z.string().describe("Text to type"),
}, async ({ selector, text }) => {
  const result = await browser.type(selector, text);
  return { content: [{ type: "text", text: result }] };
});

server.tool("browser_read", "Read visible text content from current page", {},
async () => {
  const content = await browser.readPage();
  return { content: [{ type: "text", text: content }] };
});

server.tool("browser_screenshot", "Capture a screenshot of the current page", {},
async () => {
  const png = await browser.screenshot();
  return {
    content: [{
      type: "image",
      data: png.toString("base64"),
      mimeType: "image/png",
    }],
  };
});
```

## Common Automation Patterns

### Login Flow

```typescript
async function loginToSite(browser: AgentBrowser, url: string, email: string, password: string) {
  await browser.navigate(url + "/login");
  await browser.waitFor('[name="email"]');
  await browser.type('[name="email"]', email);
  await browser.type('[name="password"]', password);
  await browser.click('[type="submit"]');
  await browser.waitFor('[data-testid="dashboard"]');
  return "Logged in successfully";
}
```

### Data Extraction

```typescript
async function extractTableData(page: Page): Promise<Record<string, string>[]> {
  return page.evaluate(() => {
    const rows = Array.from(document.querySelectorAll("table tbody tr"));
    const headers = Array.from(document.querySelectorAll("table thead th"))
      .map(th => th.textContent?.trim() ?? "");
    return rows.map(row => {
      const cells = Array.from(row.querySelectorAll("td")).map(td => td.textContent?.trim() ?? "");
      return Object.fromEntries(headers.map((h, i) => [h, cells[i] ?? ""]));
    });
  });
}
```

### Wait for Network Idle

```typescript
await page.waitForLoadState("networkidle"); // after all requests settle
await page.waitForResponse(resp => resp.url().includes("/api/data")); // specific API
```

## Security Considerations

- **Never store credentials in code** — pass via environment variables only
- **Sandbox browser context** — use `context.addInitScript` to inject CSP headers in tests
- **Headless detection bypass** — some sites block headless browsers; set a realistic `userAgent`
- **Rate limiting** — add delays between actions to avoid triggering bot detection
- **Data isolation** — use separate browser contexts for separate user sessions; never share cookies across sessions

## Common Gotchas

- **Dynamic selectors** — prefer `data-testid` attributes over CSS classes that change with UI rebuilds
- **Race conditions** — always `waitFor` after navigation or before reading data; never assume instant rendering
- **Iframes** — switch frame context with `page.frame({ name: "frame-name" })` before interacting with iframe content
- **File downloads** — set `downloadPath` in browser context options before triggering downloads
- **Playwright vs Puppeteer** — Playwright supports Chromium, Firefox, and WebKit; Puppeteer is Chromium-only
