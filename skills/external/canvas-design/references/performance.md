# Canvas Performance Reference

## Offscreen Canvas

Offscreen canvas moves rendering work to a Web Worker, keeping the main thread free:

```typescript
// main.ts
const offscreen = document.getElementById("canvas").transferControlToOffscreen();
const worker = new Worker("render-worker.js");
worker.postMessage({ canvas: offscreen }, [offscreen]);

// render-worker.ts
self.onmessage = ({ data }) => {
  const canvas = data.canvas;
  const ctx = canvas.getContext("2d");
  // Render loop here — runs off main thread
  function loop() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    // ... draw
    requestAnimationFrame(loop);
  }
  loop();
};
```

## Dirty Region Rendering

Only redraw the parts of the canvas that changed:

```typescript
interface DirtyRegion { x: number; y: number; w: number; h: number; }

class OptimizedCanvas {
  private dirtyRegions: DirtyRegion[] = [];

  markDirty(region: DirtyRegion) {
    this.dirtyRegions.push(region);
  }

  render(ctx: CanvasRenderingContext2D) {
    for (const region of this.dirtyRegions) {
      ctx.clearRect(region.x, region.y, region.w, region.h);
      // Redraw only this region
      ctx.save();
      ctx.beginPath();
      ctx.rect(region.x, region.y, region.w, region.h);
      ctx.clip();
      this.drawScene(ctx);
      ctx.restore();
    }
    this.dirtyRegions = [];
  }

  private drawScene(ctx: CanvasRenderingContext2D) {
    // Full scene draw — clipping ensures only dirty region pixels are updated
  }
}
```

## Path Batching

Minimize state changes — each `fillStyle` or `font` change flushes the GPU pipeline:

```typescript
// BAD — state change per shape
shapes.forEach(s => {
  ctx.fillStyle = s.color;
  ctx.fillRect(s.x, s.y, s.w, s.h);
});

// GOOD — group by color
const byColor = new Map<string, typeof shapes>();
shapes.forEach(s => {
  const arr = byColor.get(s.color) ?? [];
  arr.push(s);
  byColor.set(s.color, arr);
});

for (const [color, group] of byColor) {
  ctx.fillStyle = color;
  ctx.beginPath();
  group.forEach(s => ctx.rect(s.x, s.y, s.w, s.h));
  ctx.fill();
}
```

## Image Caching

Pre-render complex shapes to an offscreen canvas and use `drawImage` in the loop:

```typescript
function createCachedShape(size: number, color: string): HTMLCanvasElement {
  const cache = document.createElement("canvas");
  cache.width = size;
  cache.height = size;
  const ctx = cache.getContext("2d")!;
  // Expensive draw once
  ctx.beginPath();
  ctx.arc(size / 2, size / 2, size / 2, 0, Math.PI * 2);
  ctx.fillStyle = color;
  ctx.fill();
  return cache;
}

const cachedCircle = createCachedShape(40, "hsl(220 90% 56%)");

// Cheap drawImage in loop
function renderFrame(ctx: CanvasRenderingContext2D, x: number, y: number) {
  ctx.drawImage(cachedCircle, x - 20, y - 20);
}
```

## Performance Checklist

- [ ] Use `requestAnimationFrame`, never `setTimeout` for animation loops
- [ ] Cap delta time in animation loop (`Math.min(dt, 3)`) to prevent spiral of death
- [ ] Pre-compute static elements once; only animate what moves
- [ ] Use integer coordinates — fractional pixels trigger anti-aliasing (slower)
- [ ] Use `willReadFrequently: true` in `getContext` if calling `getImageData` often
- [ ] Profile with Chrome DevTools Performance tab — look for long paint frames
