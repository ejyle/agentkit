# Canvas Design (via anthropics/skills)

---
name: canvas-design
description: >
  Use when working with the HTML Canvas API or WebGL for 2D/3D graphics — covers
  drawing primitives, animation loops, image manipulation, hit detection, and
  performance optimization for browser-based visual applications.
license: MIT
source: https://github.com/anthropics/skills
---

## When to Use

Activate this skill when the task involves:

- Drawing shapes, paths, text, or images using the Canvas 2D API
- Building animation loops with `requestAnimationFrame`
- Implementing hit detection for interactive canvas elements
- Processing or manipulating images pixel-by-pixel
- Integrating canvas rendering in React or other frameworks
- Optimizing canvas performance (offscreen canvas, dirty regions, batching)
- Building data visualizations directly on canvas

## Canvas 2D Fundamentals

### Setup

```html
<canvas id="canvas" width="800" height="600"></canvas>
```

```typescript
const canvas = document.getElementById("canvas") as HTMLCanvasElement;
const ctx = canvas.getContext("2d")!;

// Handle device pixel ratio for sharp rendering on HiDPI displays
const dpr = window.devicePixelRatio || 1;
canvas.width = 800 * dpr;
canvas.height = 600 * dpr;
canvas.style.width = "800px";
canvas.style.height = "600px";
ctx.scale(dpr, dpr);
```

### Drawing Shapes

```typescript
// Rectangle
ctx.fillStyle = "hsl(220 90% 56%)";
ctx.fillRect(10, 10, 100, 60);
ctx.strokeStyle = "hsl(220 90% 40%)";
ctx.lineWidth = 2;
ctx.strokeRect(10, 10, 100, 60);

// Circle
ctx.beginPath();
ctx.arc(200, 50, 40, 0, Math.PI * 2);
ctx.fillStyle = "hsl(142 71% 45%)";
ctx.fill();

// Custom path
ctx.beginPath();
ctx.moveTo(300, 10);
ctx.lineTo(350, 90);
ctx.lineTo(250, 90);
ctx.closePath();
ctx.fillStyle = "hsl(38 92% 50%)";
ctx.fill();
```

### Text Rendering

```typescript
ctx.font = "bold 24px Inter, system-ui, sans-serif";
ctx.fillStyle = "hsl(220 13% 13%)";
ctx.textBaseline = "top";
ctx.fillText("Hello Canvas", 10, 120);

// Measure text for alignment
const metrics = ctx.measureText("Centered");
const x = (canvas.width / dpr - metrics.width) / 2;
ctx.fillText("Centered", x, 160);
```

### Images

```typescript
// Draw an image
const img = new Image();
img.onload = () => {
  ctx.drawImage(img, 0, 0);                   // natural size
  ctx.drawImage(img, 0, 0, 200, 150);         // scaled
  ctx.drawImage(img, 50, 50, 100, 100,        // source crop
                10, 10, 200, 200);             // destination rect
};
img.src = "/assets/photo.jpg";

// Pixel manipulation
const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
const data = imageData.data; // Uint8ClampedArray — RGBA per pixel
for (let i = 0; i < data.length; i += 4) {
  // Invert colors
  data[i]     = 255 - data[i];     // R
  data[i + 1] = 255 - data[i + 1]; // G
  data[i + 2] = 255 - data[i + 2]; // B
  // data[i + 3] = alpha (unchanged)
}
ctx.putImageData(imageData, 0, 0);
```

## Animation Loop

```typescript
interface State {
  x: number;
  y: number;
  vx: number;
  vy: number;
}

const state: State = { x: 400, y: 300, vx: 2, vy: 1.5 };
let animationId: number;

function update(dt: number) {
  state.x += state.vx * dt;
  state.y += state.vy * dt;
  // Bounce off edges
  if (state.x < 20 || state.x > 780) state.vx *= -1;
  if (state.y < 20 || state.y > 580) state.vy *= -1;
}

function render() {
  ctx.clearRect(0, 0, 800, 600);
  ctx.beginPath();
  ctx.arc(state.x, state.y, 20, 0, Math.PI * 2);
  ctx.fillStyle = "hsl(220 90% 56%)";
  ctx.fill();
}

let lastTime = 0;
function loop(timestamp: number) {
  const dt = Math.min((timestamp - lastTime) / 16.67, 3); // cap at 3 frames
  lastTime = timestamp;
  update(dt);
  render();
  animationId = requestAnimationFrame(loop);
}

// Start
animationId = requestAnimationFrame(loop);

// Stop
function stop() { cancelAnimationFrame(animationId); }
```

## Hit Detection

```typescript
interface Shape {
  x: number; y: number; w: number; h: number;
  label: string;
}

const shapes: Shape[] = [
  { x: 10, y: 10, w: 100, h: 60, label: "Box A" },
  { x: 150, y: 10, w: 100, h: 60, label: "Box B" },
];

function getCanvasPoint(e: MouseEvent): { x: number; y: number } {
  const rect = canvas.getBoundingClientRect();
  return { x: e.clientX - rect.left, y: e.clientY - rect.top };
}

canvas.addEventListener("click", (e) => {
  const { x, y } = getCanvasPoint(e);
  const hit = shapes.find(s => x >= s.x && x <= s.x + s.w && y >= s.y && y <= s.y + s.h);
  if (hit) console.log("Clicked:", hit.label);
});
```

## React Integration

```tsx
import { useEffect, useRef } from "react";

export function CanvasComponent({ width = 800, height = 600 }) {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current!;
    const ctx = canvas.getContext("2d")!;
    const dpr = window.devicePixelRatio || 1;
    canvas.width = width * dpr;
    canvas.height = height * dpr;
    ctx.scale(dpr, dpr);

    let animId: number;
    function render() {
      ctx.clearRect(0, 0, width, height);
      // ... draw here
      animId = requestAnimationFrame(render);
    }
    animId = requestAnimationFrame(render);
    return () => cancelAnimationFrame(animId);
  }, [width, height]);

  return <canvas ref={canvasRef} style={{ width, height }} />;
}
```

## Reference Files

| Task | Reference File |
|------|---------------|
| Offscreen canvas, performance, WebWorker rendering | `references/performance.md` |

## Common Gotchas

- **Blurry canvas on HiDPI** — always multiply canvas dimensions by `devicePixelRatio` and scale the context
- **Forgetting `beginPath()`** — accumulated paths from prior draws will be redrawn; always call `beginPath()` before a new shape
- **clearRect vs fillRect** — use `clearRect` to erase to transparent; `fillRect` with white fills to solid white (breaks transparency)
- **Canvas tainted by cross-origin image** — use `img.crossOrigin = "anonymous"` before setting `src` for images from other origins
- **`requestAnimationFrame` not paused on tab hide** — check `document.visibilityState` to pause animation loops
