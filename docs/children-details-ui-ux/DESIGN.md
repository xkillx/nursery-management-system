---
name: Nursery Core
colors:
  surface: '#f9f9ff'
  surface-dim: '#d0daf2'
  surface-bright: '#f9f9ff'
  surface-container-lowest: '#ffffff'
  surface-container-low: '#f0f3ff'
  surface-container: '#e8eeff'
  surface-container-high: '#dfe8ff'
  surface-container-highest: '#d9e3fb'
  on-surface: '#111c2d'
  on-surface-variant: '#444656'
  inverse-surface: '#273143'
  inverse-on-surface: '#ecf0ff'
  outline: '#757687'
  outline-variant: '#c5c5d8'
  surface-tint: '#2e48eb'
  primary: '#2642e6'
  on-primary: '#ffffff'
  primary-container: '#465fff'
  on-primary-container: '#f9f7ff'
  inverse-primary: '#bbc3ff'
  secondary: '#565e71'
  on-secondary: '#ffffff'
  secondary-container: '#dbe2f9'
  on-secondary-container: '#5c6477'
  tertiary: '#993c00'
  on-tertiary: '#ffffff'
  tertiary-container: '#c14d00'
  on-tertiary-container: '#fff6f3'
  error: '#EF4444'
  on-error: '#ffffff'
  error-container: '#ffdad6'
  on-error-container: '#93000a'
  primary-fixed: '#dfe0ff'
  primary-fixed-dim: '#bbc3ff'
  on-primary-fixed: '#000d5f'
  on-primary-fixed-variant: '#0029d2'
  secondary-fixed: '#dbe2f9'
  secondary-fixed-dim: '#bfc6dc'
  on-secondary-fixed: '#141b2c'
  on-secondary-fixed-variant: '#3f4759'
  tertiary-fixed: '#ffdbcc'
  tertiary-fixed-dim: '#ffb695'
  on-tertiary-fixed: '#351000'
  on-tertiary-fixed-variant: '#7b2f00'
  background: '#f9f9ff'
  on-background: '#111c2d'
  surface-variant: '#d9e3fb'
  success: '#10B981'
  warning: '#F59E0B'
  info: '#3B82F6'
  border-subtle: '#E2E8F0'
typography:
  display:
    fontFamily: Outfit
    fontSize: 36px
    fontWeight: '700'
    lineHeight: 44px
    letterSpacing: -0.02em
  headline-lg:
    fontFamily: Outfit
    fontSize: 28px
    fontWeight: '600'
    lineHeight: 36px
  headline-md:
    fontFamily: Outfit
    fontSize: 20px
    fontWeight: '600'
    lineHeight: 28px
  body-lg:
    fontFamily: Outfit
    fontSize: 16px
    fontWeight: '400'
    lineHeight: 24px
  body-md:
    fontFamily: Outfit
    fontSize: 14px
    fontWeight: '400'
    lineHeight: 20px
  label-md:
    fontFamily: Outfit
    fontSize: 12px
    fontWeight: '500'
    lineHeight: 16px
    letterSpacing: 0.01em
  label-sm:
    fontFamily: Outfit
    fontSize: 11px
    fontWeight: '600'
    lineHeight: 14px
rounded:
  sm: 0.125rem
  DEFAULT: 0.25rem
  md: 0.375rem
  lg: 0.5rem
  xl: 0.75rem
  full: 9999px
spacing:
  base: 4px
  xs: 4px
  sm: 8px
  md: 16px
  lg: 24px
  xl: 32px
  container-max: 1440px
  sidebar-width: 280px
---

## Brand & Style

The design system is engineered for the high-stakes, high-empathy environment of UK nursery management. The brand personality is **Professional, Dependable, and Orderly**, balancing the technical requirements of administrative data with the warmth required for childcare settings.

The chosen style is **Corporate / Modern**, emphasizing clarity and structural integrity. It utilizes a sidebar-driven architecture with clean lines, ample white space, and a refined "SaaS" aesthetic. The visual language avoids decorative clutter to ensure nursery managers can process complex information—attendance, staffing ratios, and billing—with minimal cognitive load.

## Colors

This design system utilizes a foundation of professional blues and high-utility neutrals. 

- **Primary (#465FFF):** An energetic yet trustworthy blue used for primary actions, active navigation states, and key data points.
- **Secondary (#101828):** A deep charcoal used for text and the sidebar background to provide a strong visual anchor.
- **Neutral (#667085):** A cool gray reserved for secondary text, icons, and metadata.
- **Background (#F9FAFB):** A soft, off-white surface that reduces glare during long periods of administrative use.

Semantic colors (Success, Warning, Error) follow standard UK web accessibility conventions to ensure critical alerts regarding child safety or billing issues are immediately recognizable.

## Typography

The design system uses **Outfit** exclusively to provide a contemporary, geometric feel that remains highly legible in data-dense tables. 

The hierarchy is built to emphasize scanning. Large display titles are used for dashboard overviews, while `label-sm` is utilized for table headers and metadata tags. Line heights are generous to prevent the "wall of text" effect common in management software. On mobile devices, `display` and `headline-lg` should scale down by 15% to maintain viewport harmony.

## Layout & Spacing

The design system follows a **Fixed-Fluid Hybrid Grid**. The sidebar remains fixed at 280px, while the main content area utilizes a fluid grid that expands to a maximum width of 1440px.

- **Grid:** A 12-column system is used for dashboard widgets. Gutters are fixed at 24px (`lg`) to ensure clear separation between data cards.
- **Rhythm:** A 4px baseline shift is used for all internal component spacing.
- **Breakpoints:**
  - **Desktop (1024px+):** Sidebar is visible; content is fluid.
  - **Tablet (768px - 1023px):** Sidebar collapses to icons or a hamburger menu; margins reduce to 16px.
  - **Mobile (<768px):** Single-column layout; cards stack vertically; horizontal scrolling enabled only for large data tables.

## Elevation & Depth

Hierarchy in this design system is communicated through **Tonal Layers** and **Low-Contrast Outlines**.

- **Surfaces:** The primary background is `#F9FAFB`. White cards (`#FFFFFF`) sit on top of this surface to denote interactive or informative sections.
- **Borders:** Instead of heavy shadows, components use a 1px solid border (`#E2E8F0`). This creates a clean, "flat" look that excels in data-heavy environments.
- **Interactive Depth:** Only the most critical floating elements (modals, dropdowns) use a subtle, diffused shadow: `0px 4px 6px -1px rgba(0, 0, 0, 0.1)`. This keeps the interface feeling grounded and professional.

## Shapes

The shape language is **Soft** and restrained. 

- **Standard Elements:** Buttons, input fields, and cards use a 4px (0.25rem) corner radius. This provides a professional edge that is softer than a sharp corner but more serious than a fully rounded one.
- **Containers:** Large dashboard sections and modals use `rounded-lg` (8px) to soften the overall layout.
- **Indicators:** Status chips (e.g., "Present", "Absent") use a fully pill-shaped radius to distinguish them from interactive buttons.

## Components

### Buttons
- **Primary:** Solid `#465FFF` with white text. 
- **Secondary:** White background with a `#E2E8F0` border and `#101828` text.
- **Ghost:** No background or border, used for less frequent actions.

### Inputs & Form Fields
- Fields use a white background with a 1px `#E2E8F0` border.
- On focus, the border shifts to `#465FFF` with a subtle 2px outer glow of the same color at 10% opacity.
- Labels are positioned above the field using `label-md`.

### Data Tables
- Headers use `label-sm` with a light gray background (`#F9FAFB`).
- Rows have a subtle hover state (`#F1F5F9`) to assist with line-tracking across wide screens.
- Inline status chips are mandatory for attendance and payment tracking.

### Cards
- Dashboard widgets are contained in white cards with 1px borders. 
- Headers within cards should be separated by a 1px horizontal rule to distinguish title from content.

### Sidebar
- The sidebar uses the `secondary` color (`#101828`) for its background.
- Nav items use a semi-transparent white text, shifting to pure white with a left-side `#465FFF` border accent when active.