# ArchitectMind — Warning 光暈效果重構（方案 A+C）

## 問題

現有的 `warningPulse` 使用 CSS `box-shadow` 在 `<div>` 上打光，但節點邊框已改為 SVG rough.js 渲染，導致光暈是方形的、與手繪不規則邊框不協調。

## 解法

- **方案 A**：SVG filter glow — 光暈沿著手繪邊框形狀擴散
- **方案 C**：SVG stroke 呼吸動畫 — 邊框本身做透明度脈衝
- **移除**：原本的 CSS `box-shadow` warningPulse

---

## 需要改動的檔案

| 檔案 | 改動 |
|---|---|
| `frontend/src/nodes/ArchitectureNode.tsx` | useEffect 裡加 SVG filter + stroke 動畫 |
| `frontend/src/index.css` | 新增 `strokePulse` keyframes，移除 `warningPulse` |

---

## Step 1：修改 `index.css`

### 1.1 移除舊的 warningPulse

刪掉：

```css
@keyframes warningPulse {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(245, 158, 11, 0);
  }
  50% {
    box-shadow: 0 0 12px 4px rgba(245, 158, 11, 0.5);
  }
}
```

### 1.2 新增 strokePulse

加入：

```css
@keyframes strokePulse {
  0%, 100% { stroke-opacity: 1; }
  50% { stroke-opacity: 0.4; }
}
```

---

## Step 2：修改 `ArchitectureNode.tsx`

### 2.1 移除 div 上的 warningPulse animation

找到 return 裡的外層 `<div>` style，**刪掉** 這段：

```tsx
// ❌ 刪掉這個
...(hasWarnings && {
  animation: 'warningPulse 2s ease-in-out infinite',
}),
```

改成（不需要任何 animation）：

```tsx
style={{
  position: 'relative',
  width: width + extraOffset,
  height: height + extraOffset,
  fontFamily: 'var(--font-hand)',
  // ← 不再有 animation
}}
```

### 2.2 修改 useEffect — 加入 SVG filter glow + stroke 動畫

將整個 `useEffect` 替換為：

```typescript
useEffect(() => {
  if (!svgRef.current) return
  const svg = svgRef.current
  while (svg.firstChild) svg.removeChild(svg.firstChild)

  const rc = rough.svg(svg)

  // --------------------------------------------------------
  // A) 如果有 warning，建立 SVG glow filter
  // --------------------------------------------------------
  if (hasWarnings) {
    const defs = document.createElementNS('http://www.w3.org/2000/svg', 'defs')
    const filter = document.createElementNS('http://www.w3.org/2000/svg', 'filter')
    filter.setAttribute('id', `glow-${id}`)
    // 擴大 filter 範圍避免裁切
    filter.setAttribute('x', '-50%')
    filter.setAttribute('y', '-50%')
    filter.setAttribute('width', '200%')
    filter.setAttribute('height', '200%')
    filter.innerHTML = `
      <feGaussianBlur stdDeviation="3" result="blur" />
      <feFlood flood-color="#f59e0b" flood-opacity="0.35" />
      <feComposite in2="blur" operator="in" result="glow" />
      <feMerge>
        <feMergeNode in="glow" />
        <feMergeNode in="SourceGraphic" />
      </feMerge>
    `
    defs.appendChild(filter)
    svg.appendChild(defs)
  }

  // --------------------------------------------------------
  // B) 主矩形
  // --------------------------------------------------------
  const rect = rc.rectangle(4, 4, width - 8, height - 8, {
    stroke: hasWarnings ? '#f59e0b' : primaryColor,
    strokeWidth: hasWarnings ? 2.5 : 2,
    fill: hasWarnings ? '#f59e0b15' : `${primaryColor}18`,
    fillStyle: 'solid',
    roughness: 1.5,
    seed,
  })

  // 套用 glow filter
  if (hasWarnings) {
    rect.setAttribute('filter', `url(#glow-${id})`)
  }

  svg.appendChild(rect)

  // --------------------------------------------------------
  // C) 有 warning 時，對邊框 path 加上呼吸動畫
  // --------------------------------------------------------
  if (hasWarnings) {
    const paths = rect.querySelectorAll('path')
    paths.forEach(path => {
      path.style.animation = 'strokePulse 2s ease-in-out infinite'
    })
  }

  // --------------------------------------------------------
  // D) Replica 額外層（邏輯不變）
  // --------------------------------------------------------
  if (replicas > 1) {
    for (let i = extraLayers; i >= 1; i--) {
      const offset = i * 6
      const bgRect = rc.rectangle(
        4 + offset, 4 + offset,
        width - 8, height - 8,
        {
          stroke: hasWarnings ? '#f59e0b' : primaryColor,
          strokeWidth: 1.5,
          fill: hasWarnings ? '#f59e0b08' : `${primaryColor}10`,
          fillStyle: 'solid',
          roughness: 1.5,
          seed: seed + i,
        }
      )
      svg.insertBefore(bgRect, svg.firstChild)
    }
  }
}, [id, primaryColor, seed, width, height, replicas, extraLayers, hasWarnings])
```

> **注意**：`useEffect` 的依賴陣列新增了 `id` 和 `hasWarnings`。

---

## 完整改動對照

### ArchitectureNode.tsx — useEffect 依賴陣列

```diff
- }, [primaryColor, seed, width, height, replicas, extraLayers])
+ }, [id, primaryColor, seed, width, height, replicas, extraLayers, hasWarnings])
```

### ArchitectureNode.tsx — 外層 div style

```diff
  style={{
    position: 'relative',
    width: width + extraOffset,
    height: height + extraOffset,
    fontFamily: 'var(--font-hand)',
-   ...(hasWarnings && {
-     animation: 'warningPulse 2s ease-in-out infinite',
-   }),
  }}
```

### index.css

```diff
- @keyframes warningPulse {
-   0%, 100% {
-     box-shadow: 0 0 0 0 rgba(245, 158, 11, 0);
-   }
-   50% {
-     box-shadow: 0 0 12px 4px rgba(245, 158, 11, 0.5);
-   }
- }

+ @keyframes strokePulse {
+   0%, 100% { stroke-opacity: 1; }
+   50% { stroke-opacity: 0.4; }
+ }
```

---

## 視覺效果說明

| 狀態 | 效果 |
|---|---|
| 正常節點 | 各 componentType 顏色邊框 + 淡色填充 |
| Warning 節點（靜態） | 邊框變 `#f59e0b` 琥珀色 + 微弱柔光（SVG filter glow） |
| Warning 節點（動態） | 邊框 stroke 透明度在 1.0 ↔ 0.4 間呼吸，2 秒一個循環 |
| Replica + Warning | 額外層也變琥珀色，但不加呼吸動畫（只有主矩形呼吸） |

---

## 測試驗證

1. 載入任一 Preset
2. 點擊 **Analyze**
3. 觸發一個 warning（例如：只放 1 個 Load Balancer → SPOF 警告）
4. 確認：
   - ✅ 邊框變琥珀色
   - ✅ 柔光沿著手繪邊框形狀擴散（不是方形 box-shadow）
   - ✅ 邊框在「呼吸」（透明度脈衝）
   - ✅ hover 時 tooltip 仍正常顯示
   - ✅ 沒有 warning 的節點完全不受影響

---

## 可調參數

| 參數 | 位置 | 預設值 | 說明 |
|---|---|---|---|
| `stdDeviation` | SVG filter | `3` | 光暈擴散程度（越大越模糊） |
| `flood-opacity` | SVG filter | `0.35` | 光暈亮度（0~1） |
| `strokeWidth` | rc.rectangle | `2.5` | Warning 時的邊框粗度 |
| `stroke-opacity` range | CSS keyframes | `1 ↔ 0.4` | 呼吸的幅度 |
| animation duration | CSS keyframes | `2s` | 呼吸速度 |
