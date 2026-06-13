# Advanced PDF Reference

## PDF Form Handling

### Reading Form Fields (pdf-lib)

```bash
npm install pdf-lib
```

```typescript
import { PDFDocument } from "pdf-lib";
import { readFileSync } from "fs";

async function readFormFields(pdfPath: string): Promise<Record<string, string>> {
  const pdfBytes = readFileSync(pdfPath);
  const pdfDoc = await PDFDocument.load(pdfBytes);
  const form = pdfDoc.getForm();

  const fields: Record<string, string> = {};
  for (const field of form.getFields()) {
    const name = field.getName();
    if (field.constructor.name === "PDFTextField") {
      fields[name] = (field as any).getText() ?? "";
    } else if (field.constructor.name === "PDFCheckBox") {
      fields[name] = (field as any).isChecked() ? "true" : "false";
    }
  }
  return fields;
}
```

### Filling Form Fields

```typescript
async function fillPDFForm(
  templatePath: string,
  data: Record<string, string>
): Promise<Uint8Array> {
  const templateBytes = readFileSync(templatePath);
  const pdfDoc = await PDFDocument.load(templateBytes);
  const form = pdfDoc.getForm();

  for (const [fieldName, value] of Object.entries(data)) {
    try {
      const field = form.getTextField(fieldName);
      field.setText(value);
    } catch {
      // Field might be a checkbox or not exist — skip gracefully
    }
  }

  form.flatten(); // Flatten to prevent further editing
  return pdfDoc.save();
}
```

## Merging PDF Files

```typescript
import { PDFDocument } from "pdf-lib";

async function mergePDFs(pdfBuffers: Buffer[]): Promise<Uint8Array> {
  const merged = await PDFDocument.create();

  for (const buffer of pdfBuffers) {
    const doc = await PDFDocument.load(buffer);
    const pageIndices = doc.getPageIndices();
    const copiedPages = await merged.copyPages(doc, pageIndices);
    copiedPages.forEach(page => merged.addPage(page));
  }

  return merged.save();
}
```

## PDF Password Protection

```typescript
async function protectPDF(pdfBytes: Buffer, userPassword: string): Promise<Uint8Array> {
  const pdfDoc = await PDFDocument.load(pdfBytes);
  return pdfDoc.save({
    userPassword,
    ownerPassword: crypto.randomUUID(), // Strong owner password
    permissions: {
      printing: "highResolution",
      modifying: false,
      copying: false,
      annotating: false,
    },
  });
}
```

## Watermarking

```typescript
import { PDFDocument, rgb, degrees } from "pdf-lib";

async function addWatermark(pdfBytes: Buffer, text: string): Promise<Uint8Array> {
  const pdfDoc = await PDFDocument.load(pdfBytes);
  const pages = pdfDoc.getPages();

  for (const page of pages) {
    const { width, height } = page.getSize();
    page.drawText(text, {
      x: width / 2 - 100,
      y: height / 2,
      size: 48,
      color: rgb(0.75, 0.75, 0.75),
      opacity: 0.3,
      rotate: degrees(45),
    });
  }

  return pdfDoc.save();
}
```

## Streaming Large PDFs

For large PDFs, stream to the client instead of buffering in memory:

```typescript
// Express streaming response
app.get("/large-report", async (req, res) => {
  res.setHeader("Content-Type", "application/pdf");
  res.setHeader("Content-Disposition", 'attachment; filename="report.pdf"');

  const browser = await puppeteer.launch();
  const page = await browser.newPage();
  await page.goto("https://internal-report-service/report");

  // Stream directly to response
  const stream = await page.createPDFStream({ format: "A4" });
  stream.pipe(res);
  stream.on("end", () => browser.close());
});
```

## Digital Signatures

Digital signatures require a certificate (PKCS#12 / .p12 format). Use `node-signpdf`
for self-signed certificates in development:

```bash
npm install node-signpdf
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365 -nodes
openssl pkcs12 -export -out cert.p12 -inkey key.pem -in cert.pem -passout pass:password
```

For production, obtain a certificate from a trusted CA (DocuSign, GlobalSign, etc.)
and implement signature workflows according to the PDF/A-3 or PAdES standard.
