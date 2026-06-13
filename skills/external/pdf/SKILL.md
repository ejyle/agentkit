# PDF (via anthropics/skills)

---
name: pdf
description: >
  Use when generating, parsing, or manipulating PDF files — covers server-side PDF
  generation with Puppeteer and React-PDF, client-side rendering with PDF.js, text
  extraction, form handling, and image embedding.
license: MIT
source: https://github.com/anthropics/skills
---

## When to Use

Activate this skill when the task involves:

- Generating PDFs from HTML/CSS using Puppeteer or Playwright
- Building typed PDF documents with React-PDF or PDFKit
- Rendering PDFs in the browser with PDF.js
- Extracting text from PDF files for processing
- Filling or reading PDF form fields
- Embedding images, tables, or charts in PDFs
- Sending PDF files as API responses (content-disposition: attachment)

## PDF Generation

### Approach 1: HTML-to-PDF via Puppeteer (Best for styled documents)

```typescript
import puppeteer from "puppeteer";

async function generatePDF(htmlContent: string): Promise<Buffer> {
  const browser = await puppeteer.launch({ headless: "new" });
  const page = await browser.newPage();

  await page.setContent(htmlContent, { waitUntil: "networkidle0" });

  const pdf = await page.pdf({
    format: "A4",
    printBackground: true,
    margin: { top: "20mm", right: "15mm", bottom: "20mm", left: "15mm" },
  });

  await browser.close();
  return Buffer.from(pdf);
}

// Usage
const html = `
<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: Arial, sans-serif; font-size: 12pt; }
    h1 { color: #1a56db; }
    table { width: 100%; border-collapse: collapse; }
    td, th { border: 1px solid #ddd; padding: 8px; }
  </style>
</head>
<body>
  <h1>Invoice #1001</h1>
  <p>Date: 2024-01-15</p>
</body>
</html>
`;
const pdfBuffer = await generatePDF(html);
```

### Approach 2: React-PDF (Typed, programmatic documents)

```bash
npm install @react-pdf/renderer
```

```typescript
import { Document, Page, Text, View, StyleSheet } from "@react-pdf/renderer";
import { renderToBuffer } from "@react-pdf/renderer";

const styles = StyleSheet.create({
  page: { padding: 40, fontFamily: "Helvetica" },
  title: { fontSize: 24, marginBottom: 16, color: "#1a56db" },
  text: { fontSize: 12, lineHeight: 1.5 },
  table: { display: "flex", flexDirection: "column", marginTop: 16 },
  row: { flexDirection: "row", borderBottom: "1px solid #ddd", paddingVertical: 6 },
  cell: { flex: 1, fontSize: 11 },
  header: { fontFamily: "Helvetica-Bold" },
});

const InvoicePDF = ({ invoice }) => (
  <Document>
    <Page size="A4" style={styles.page}>
      <Text style={styles.title}>Invoice #{invoice.id}</Text>
      <Text style={styles.text}>Date: {invoice.date}</Text>
      <View style={styles.table}>
        <View style={[styles.row, styles.header]}>
          <Text style={styles.cell}>Item</Text>
          <Text style={styles.cell}>Qty</Text>
          <Text style={styles.cell}>Price</Text>
        </View>
        {invoice.items.map((item, i) => (
          <View style={styles.row} key={i}>
            <Text style={styles.cell}>{item.name}</Text>
            <Text style={styles.cell}>{item.qty}</Text>
            <Text style={styles.cell}>${item.price}</Text>
          </View>
        ))}
      </View>
    </Page>
  </Document>
);

const buffer = await renderToBuffer(<InvoicePDF invoice={invoiceData} />);
```

## Serving PDF as HTTP Response

```typescript
// Express / Node.js
app.get("/invoice/:id", async (req, res) => {
  const invoice = await getInvoice(req.params.id);
  const pdf = await generateInvoicePDF(invoice);

  res.setHeader("Content-Type", "application/pdf");
  res.setHeader("Content-Disposition", `attachment; filename="invoice-${invoice.id}.pdf"`);
  res.setHeader("Content-Length", pdf.length);
  res.end(pdf);
});

// Next.js App Router
export async function GET(request: Request, { params }: { params: { id: string } }) {
  const pdf = await generatePDF(params.id);
  return new Response(pdf, {
    headers: {
      "Content-Type": "application/pdf",
      "Content-Disposition": `attachment; filename="invoice-${params.id}.pdf"`,
    },
  });
}
```

## PDF Rendering in Browser (PDF.js)

```bash
npm install pdfjs-dist
```

```typescript
import * as pdfjsLib from "pdfjs-dist";
pdfjsLib.GlobalWorkerOptions.workerSrc = "/pdf.worker.min.js";

async function renderPDF(url: string, canvasEl: HTMLCanvasElement) {
  const loadingTask = pdfjsLib.getDocument(url);
  const pdf = await loadingTask.promise;
  const page = await pdf.getPage(1);

  const viewport = page.getViewport({ scale: 1.5 });
  canvasEl.width = viewport.width;
  canvasEl.height = viewport.height;

  await page.render({
    canvasContext: canvasEl.getContext("2d")!,
    viewport,
  }).promise;
}
```

## Text Extraction

```typescript
import * as pdfjsLib from "pdfjs-dist";

async function extractText(pdfBuffer: Buffer): Promise<string> {
  const pdf = await pdfjsLib.getDocument({ data: new Uint8Array(pdfBuffer) }).promise;
  let fullText = "";

  for (let pageNum = 1; pageNum <= pdf.numPages; pageNum++) {
    const page = await pdf.getPage(pageNum);
    const content = await page.getTextContent();
    const pageText = content.items
      .filter((item): item is pdfjsLib.TextItem => "str" in item)
      .map((item) => item.str)
      .join(" ");
    fullText += pageText + "\n";
  }

  return fullText;
}
```

## Reference Files

| Task | Reference File |
|------|---------------|
| PDF forms, digital signatures, encryption | `references/advanced-pdf.md` |

## Common Gotchas

- **Font embedding** — Puppeteer renders with Chrome fonts; React-PDF needs fonts registered explicitly
- **Puppeteer sandbox in CI** — add `--no-sandbox --disable-setuid-sandbox` args for Docker/CI environments
- **PDF.js worker path** — the worker file must be served as a separate static file; bundling it inline breaks large PDFs
- **Page breaks** — use `page-break-before: always` or `<View break />` in React-PDF; Puppeteer respects CSS `@page` rules
- **Images must be base64 or URL** — file paths don't work in React-PDF browser context; always use data URIs or HTTPS URLs
