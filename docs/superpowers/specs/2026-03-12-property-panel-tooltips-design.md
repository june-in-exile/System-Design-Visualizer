# Design Spec: Property Panel Tooltips (Native `title` attribute)

## Overview
Enhance the user experience by providing contextual hints (tooltips) for various system architecture components and their properties. This will help users understand the implications of different configuration choices directly within the UI.

## Goals
- Add descriptive hints to component properties (database types, cache policies, etc.).
- Add descriptive hints to edge properties (protocols, connection types).
- Ensure "unspecified" or "default" options do not show redundant tooltips.
- Avoid adding tooltips to the "Direction" property of edges as requested.
- Use the native HTML `title` attribute for lightweight implementation.

## Implementation Details

### 1. `ComponentPropertyPanel.tsx`
- **Constants Enhancement:** Ensure all option arrays (e.g., `DB_CATEGORIES`, `SQL_PRODUCTS`, `CACHE_TYPES`) have a `description` field.
- **Label Tooltips:** Add `title` to `<label>` elements describing the property's purpose.
- **Select Tooltips:** 
    - Add `title` to `<select>` elements that dynamically reflects the `description` of the currently selected option.
    - If the selected value's label is "unspecified", "default", or equivalent, the `title` should be omitted or set to an empty string.
- **Option Tooltips:** Add `title={opt.description}` to all `<option>` elements.

### 2. `EdgePropertyPanel.tsx`
- **Label Tooltips:** Add `title` to labels for "Protocol", "Connection Type", and "Label".
- **Select Tooltips:**
    - Apply the same dynamic `title` logic to Protocol and Connection Type selects.
    - **Exception:** Do NOT add tooltips to the "Direction" property select or its options.
- **Input Tooltips:** Add a placeholder or title to the "Label" input explaining its optional nature.

### 3. `Sidebar.tsx`
- Add `title` to each draggable component item in the palette to explain what the component represents (e.g., "Service: A generic compute unit or microservice").

### 4. Logic for "Unspecified" Omission
A helper function or inline logic will be used to determine if a tooltip should be shown:
```typescript
const getTooltip = (option: { label: string, description?: string }) => {
  if (!option.description || 
      option.label.toLowerCase().includes('unspecified') || 
      option.label.toLowerCase().includes('default')) {
    return undefined;
  }
  return option.description;
};
```

## User Experience
- Users hover over a label or a dropdown to see a brief explanation.
- Tooltips appear after a short delay (browser default for `title`).
- Visuals remain clean and consistent with the existing theme.

## Verification Plan
- **Manual Test:** Hover over every property label and dropdown in both panels.
- **Corner Case:** Verify that selecting "Default (Auto-detect)" in the consistency level dropdown results in no tooltip on the `select` element.
- **Edge Case:** Verify that the "Direction" dropdown for edges has no tooltips.
