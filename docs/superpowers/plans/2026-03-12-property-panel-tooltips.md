# Property Panel Tooltips Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enhance the user experience by adding descriptive hints (tooltips) using the native HTML `title` attribute to property panels and the sidebar.

**Architecture:** Use a helper function `getTooltip` to determine if a hint should be shown based on the option's label and description. Apply `title` attributes to `<label>`, `<select>`, `<option>`, and draggable sidebar items.

**Tech Stack:** React 19, TypeScript.

---

## Chunk 1: Component Property Panel Enhancements

### Task 1: Add `getTooltip` helper and update `ComponentPropertyPanel.tsx`

**Files:**
- Modify: `frontend/src/components/ComponentPropertyPanel.tsx`

- [ ] **Step 1: Implement `getTooltip` helper and update labels/selects**

Add this helper at the top of the component (outside the component function):
```typescript
const getTooltip = (label: string, description?: string) => {
  const lowerLabel = label.toLowerCase()
  if (!description || lowerLabel.includes('unspecified') || lowerLabel.includes('default') || lowerLabel.includes('auto-detect')) {
    return undefined
  }
  return description
}
```

Then, for each property section (Database, Cache, etc.):
1. Add `title` to the `<label>` explaining the property.
2. Add `title` to the `<select>` that dynamically finds the description of the selected option using `getTooltip`.
3. Add `title` to each `<option>` using `opt.description`.

Example for Database Category:
```tsx
<label 
  style={{ ... }}
  title="Choose between relational (SQL) or non-relational (NoSQL) database types."
>
  Database Category
</label>
<select
  value={properties.dbType || 'sql'}
  onChange={(e) => handleDBTypeChange(e.target.value)}
  title={getTooltip(
    DB_CATEGORIES.find(opt => opt.value === (properties.dbType || 'sql'))?.label || '',
    DB_CATEGORIES.find(opt => opt.value === (properties.dbType || 'sql'))?.description
  )}
  style={{ ... }}
>
  {DB_CATEGORIES.map((opt) => (
    <option key={opt.value} value={opt.value} title={opt.description}>{opt.label}</option>
  ))}
</select>
```

- [ ] **Step 2: Update all other sections in `ComponentPropertyPanel.tsx`**
Repeat the pattern for:
- Product (filtered by DB type)
- Consistency Requirement
- Algorithm (Load Balancer)
- Access Level (Storage)
- Storage Class (Storage)
- Category (Message Queue)
- Product (Message Queue)
- Queue Type (Message Queue)
- Delivery Guarantee (Message Queue)
- Cache Type (Cache)
- Product (Cache)
- Eviction Policy (Cache)

- [ ] **Step 3: Commit changes**
```bash
git add frontend/src/components/ComponentPropertyPanel.tsx
git commit -m "feat: add tooltips to ComponentPropertyPanel labels and selects"
```

---

## Chunk 2: Edge Property Panel and Sidebar Enhancements

### Task 2: Update `EdgePropertyPanel.tsx`

**Files:**
- Modify: `frontend/src/components/EdgePropertyPanel.tsx`

- [ ] **Step 1: Implement `getTooltip` and update Protocol/Connection Type**

Add the same `getTooltip` helper.
Update labels and selects for:
- Protocol
- Connection Type

**Constraint:** Do NOT add tooltips to "Direction" select or its options.

- [ ] **Step 2: Commit changes**
```bash
git add frontend/src/components/EdgePropertyPanel.tsx
git commit -m "feat: add tooltips to EdgePropertyPanel (excluding Direction)"
```

### Task 3: Update `Sidebar.tsx`

**Files:**
- Modify: `frontend/src/components/Sidebar.tsx`
- Modify: `frontend/src/nodes/nodeConfig.ts` (to add descriptions)

- [ ] **Step 1: Add descriptions to `NODE_TYPE_CONFIG`**
Update `frontend/src/nodes/nodeConfig.ts` to include a `description` field for each node type.
Example:
```typescript
service: {
  label: 'Service',
  icon: '⚙️',
  color: '#3b82f6',
  description: 'A generic compute unit, microservice, or application server.'
},
```

- [ ] **Step 2: Add `title` to Sidebar items**
In `frontend/src/components/Sidebar.tsx`, add `title={config.description}` to the draggable `div`.

- [ ] **Step 3: Commit changes**
```bash
git add frontend/src/nodes/nodeConfig.ts frontend/src/components/Sidebar.tsx
git commit -m "feat: add component descriptions to Sidebar tooltips"
```

---

## Chunk 3: Final Verification

### Task 4: Verify All Tooltips

- [ ] **Step 1: Manual verification**
1. Run `npm run dev` in `frontend/`.
2. Open the app and hover over:
   - Sidebar components.
   - Property labels in both panels.
   - Select dropdowns (check if title changes with selection).
   - Individual options within dropdowns.
3. Ensure "Direction" has no tooltips.
4. Ensure "unspecified" or "default" selections show no tooltip on the `select`.

- [ ] **Step 2: Final cleanup**
Check for any console errors or linting issues.
```bash
npm run lint
```
