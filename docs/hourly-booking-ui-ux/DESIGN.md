---
name: NurseryPro Modern Admin
colors:
  surface: '#f9f9ff'
  surface-dim: '#d0daf2'
  surface-bright: '#f9f9ff'
  surface-container-lowest: '#ffffff'
  surface-container-low: '#F0F3FF'
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
  tertiary-container: '#b9531b'
  on-tertiary-container: '#fff5f2'
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
  tertiary-fixed-dim: '#ffb694'
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
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  xs: 4px
  sm: 8px
  md: 16px
  base: 4px
  lg: 24px
  xl: 32px
  sidebar-width: 280px
  container-max: 1440px
---

## Brand & Style
The brand personality is professional, organized, and reassuring—tailored for education management. The style is **Corporate / Modern** with a focus on high information density delivered through a clean, systematic interface. It utilizes a structured sidebar for navigation and a bento-grid inspired dashboard to categorize complex data into digestible "metric cards." The aesthetic prioritizes legibility and trust, using a restrained color palette with strategic pops of color for status indicators and calls to action.

## Colors
The palette uses a deep "Navy Night" (`#101828`) for the primary navigation sidebar, creating a strong anchor for the application. The primary action color is a vibrant "Vivid Indigo" (`#465FFF`), used for active states and primary buttons. Semantic colors (Success, Error, Warning, Info) follow standard web conventions but are used with low-opacity backgrounds (10-20%) for UI "pills" and iconography to maintain a soft, approachable feel. Neutral tones emphasize a light-mode experience with off-white surfaces and subtle slate borders.

## Typography
The system relies exclusively on **Outfit**, a geometric sans-serif that balances modern efficiency with a slight friendliness. Display sizes use heavy weights (700) and negative letter spacing to create a commanding presence for key metrics. Body text is kept at a comfortable 14px for density, while labels utilize uppercase tracking and semi-bold weights for metadata and category headers. On mobile, `headline-lg` should scale down to 24px to ensure readability without excessive wrapping.

## Layout & Spacing
The application uses a **fixed-fluid hybrid grid**. A permanent 280px sidebar persists on the left (collapsible on mobile), while the main content area occupies the remaining width up to a 1440px max-width container. 

The spacing rhythm is based on a 4px base unit. Dashboards use a 24px (`lg`) gutter for the primary grid and 16px (`md`) for internal card padding. On mobile devices, side margins should reduce to 16px, and multi-column grids (like the 4-column metrics) should reflow into a single-column stack.

## Elevation & Depth
Depth is achieved primarily through **Tonal Layers** and **Low-contrast Outlines**. Surfaces are distinct:
- The background uses a very light tint (`#F9F9FF`).
- Cards and containers are pure white (`#FFFFFF`) with a 1px border (`#E2E8F0`).
- Subtle shadows (`0 1px 2px 0 rgba(0, 0, 0, 0.05)`) are applied to cards to lift them slightly off the background.
- Interactive hover states utilize a transition to a primary color border (`#465FFF`) rather than traditional heavy drop shadows, keeping the UI feeling crisp and responsive.

## Shapes
The UI uses a **Rounded** shape language to soften the corporate data-heavy environment. 
- Primary containers and dashboard cards use 0.75rem (`xl`) to 1rem (`2xl`) corner radii.
- Input fields and small buttons use a 0.5rem (`lg`) radius.
- System icons and status "dots" are fully circular (pill-shaped) to distinguish them from structural elements.

## Components
- **Buttons**: Primary buttons are solid indigo with white text. Ghost/Secondary buttons use a subtle background tint (`bg-white/10`) or simple borders. All buttons should have a 110% brightness hover effect.
- **Metric Cards**: Should always include a label, a large display value, and a footer section for trends (using semantic color text for up/down movements).
- **Navigation Items**: The active state in the sidebar is indicated by a 4px left-border and a 10% opacity white background overlay.
- **Input Fields**: Search bars use a "surface-container-low" background with internal icons and transition to a primary color border on focus.
- **Activity Feed**: Uses a vertical line (timeline) connector with circular icons for each entry, distinguishing types of activity through icon color (e.g., green for new registrations, red for alerts).
- **Progress Bars**: Used within cards to show targets; they should be 6px high with fully rounded caps.