# DESIGN.md

> Last updated: 2026-07-01
> Project: Nursery Management System (UK multi-tenant)
> Stack: Angular 21 + Tailwind CSS v4 + ng-icons (Heroicons)
> Template Base: TailAdmin

---

# 1. Product Design Overview

**Product purpose:** Multi-tenant SaaS for UK nursery management — child enrollment, attendance tracking, billing/invoicing, funding claims, and parent communication. Each nursery (tenant) has manager, practitioner, parent, and owner roles.

**Target users:**
- Nursery managers (daily operations, billing, compliance)
- Practitioners (daily attendance, child check-in/out)
- Parents (invoice viewing, payment)
- Owners (cross-site oversight, manager access control)

**Design philosophy:** Modern SaaS dashboard focused on operational efficiency, clarity, and reducing cognitive load. Prioritises task completion speed over visual flourish. Every screen serves a specific operational need with minimal friction.

**UI style direction:**
- Clean, professional, neutral-toned interface with brand blue accent
- Card-based layouts with generous whitespace
- Subtle shadows and borders for depth without excessive layering
- Dark mode support (class-based toggle, `.dark` on `<html>`)
- Consistent rounded corners (`rounded-lg`, `rounded-2xl`, `rounded-full`)
- Icon-supported navigation and actions (Heroicons outline set)

**UX principles:**
- **Role-based experiences** — each user role sees only relevant navigation and actions
- **Progressive disclosure** — complex workflows split into steps (stepper wizard for child registration)
- **Responsive by default** — mobile-first responsive design with desktop tables and mobile card layouts
- **Inline feedback** — field-level validation errors, toast notifications, contextual alerts
- **Minimal page weight** — tables with pagination, search, and filters to handle large datasets

---

# 2. Design Principles

## 2.1 Simplicity over complexity

Every page focuses on a single primary task. The manager dashboard shows exactly four attendance metrics, payment follow-up items, and quick-action links — no charts, no widgets, no noise. The children list page combines search, status filter, pagination, and a sortable table — nothing more.

## 2.2 Consistent spacing

A strict 4px/8px grid system enforced through Tailwind classes. Card padding is uniformly `p-5` (20px) or `p-6` (24px). Stack spacing between sections is `space-y-5` or `space-y-6`. Forms use `space-y-5` between fields. This consistency creates a rhythm that users internalise.

## 2.3 Clear hierarchy

Typography weight and size differentiate headings (`font-semibold text-xl`) from body text (`text-sm text-gray-500`). Page headers combine a bold title with a lighter subtitle. The sidebar uses uppercase section labels for nav grouping. Tables use `font-medium` for data and lighter `text-gray-500` for secondary info.

## 2.4 Progressive disclosure

The child registration wizard breaks a complex form across 6 steps (Child Info → Address → Health → Funding → Documents → Review). Only one step is shown at a time; a step indicator shows progress. The attendance corrections page reveals the form only after a child and date are selected.

## 2.5 Fast task completion

High-frequency actions are surfaced: check-in/check-out buttons on the attendance page are always visible with status-coloured badges. Search with 200ms debounce. Inline loading spinners on buttons during submission. Keyboard-navigable forms with autofocus on the first field.

## 2.6 Error prevention and recovery

Form validation occurs on submit (not on keystroke) with field-level error messages. API errors are mapped through a central `ApiErrorMapper` and presented with human-readable messages. Destructive actions require confirmation dialogs. The attendance corrections page shows issued invoice warnings before allowing edits.

## 2.7 Accessibility foundation

- Semantic HTML (`<nav>`, `<aside>`, `<main>`, `<table>`, `<th>`, `<tr>`)
- `aria-label`, `aria-current="page"`, `aria-modal`, `aria-expanded`, `aria-pressed` used where relevant
- `sr-only` for screen-reader-only text
- `role="alert"` on error banners, `role="status"` on success alerts
- Focus-visible outlines on interactive elements (`outline-2 outline-offset-2 outline-brand-500`)
- All interactive elements keyboard-accessible (buttons, links, form controls)

---

# 3. Design System

## 3.1 Colors

### Primary (Brand) — Blue range

| Token | Hex | Usage |
|---|---|---|
| `brand-25` | `#F2F7FF` | Hover/active table rows, light badge backgrounds |
| `brand-50` | `#ECF3FF` | Active nav item background, light badge backgrounds |
| `brand-100` | `#DDE9FF` | Badge backgrounds (solid primary) |
| `brand-200` | `#C2D6FF` | — |
| `brand-300` | `#9CB9FF` | — |
| `brand-400` | `#7592FF` | Dark mode active nav icon/text |
| `brand-500` | `#465FFF` | **Primary buttons**, links, active nav indicator, focus ring, selected dates |
| `brand-600` | `#3641F5` | Button hover |
| `brand-700` | `#2A31D8` | — |
| `brand-800` | `#252DAE` | — |
| `brand-900` | `#262E89` | — |
| `brand-950` | `#161950` | — |

### Gray (Neutral) scale

| Token | Hex | Usage |
|---|---|---|
| `gray-25` | `#FCFCFD` | — |
| `gray-50` | `#F9FAFB` | Page body background, table header background |
| `gray-100` | `#F2F4F7` | Hover states, border-t separators, table row borders |
| `gray-200` | `#E4E7EC` | Card borders, table borders, scrollbar thumb |
| `gray-300` | `#D0D5DD` | Disabled inputs, placeholder text |
| `gray-400` | `#98A2B3` | Secondary text, dark mode text, muted icons |
| `gray-500` | `#667085` | Body text, description text, secondary labels |
| `gray-600` | `#475467` | Strong body text |
| `gray-700` | `#344054` | Heading text, form labels, nav item text |
| `gray-800` | `#1D2939` | Primary heading text, dark mode heading text |
| `gray-900` | `#101828` | Near-black text, sidebar background dark mode |
| `gray-950` | `#0C111D` | — |
| `gray-dark` | `#1A2231` | Dark mode surface background |

### Status / Semantic colors

| Token | Hex | Usage |
|---|---|---|
| `success-500` | `#12B76A` | Checked-in badges, success toasts/buttons |
| `success-100` | `#D1FADF` | Success badge bg (light variant) |
| `success-50` | `#ECFDF3` | Success alert bg, success table cell bg |
| `error-500` | `#F04438` | Error badges, error buttons, required field indicators |
| `error-100` | `#FEE4E2` | Error badge bg (light variant) |
| `error-50` | `#FEF3F2` | Error alert bg |
| `warning-500` | `#F79009` | Warning badges |
| `warning-100` | `#FEF0C7` | Warning badge bg (light variant) |
| `warning-50` | `#FFFAEB` | Warning alert bg |
| `orange-500` | `#FB6514` | Secondary warning/accent |
| `blue-light-500` | `#0BA5EC` | Info badges |
| `theme-pink-500` | `#EE46BC` | Accent pink |
| `theme-purple-500` | `#7A5AF8` | Accent purple |

### Utility tokens

| Token | Value | Usage |
|---|---|---|
| `white` | `#FFFFFF` | Card/surface backgrounds |
| `black` | `#101828` | Near-black text (same as `gray-900`) |
| `transparent` | `transparent` | — |
| `current` | `currentColor` | Icon color inheritance |

### Border color default

```
border-color: var(--color-gray-200)
```

## 3.2 Typography

### Font family

```
font-family: 'Outfit', sans-serif;
```

Outfit is a modern geometric sans-serif with optical sizing across 9 weights (100–900). Used exclusively across the entire application.

### Type scale

| Token | Size | Line-height | Usage |
|---|---|---|---|
| `text-title-2xl` | 72px | 90px | Hero/404 page heading |
| `text-title-xl` | 60px | 72px | — |
| `text-title-lg` | 48px | 60px | — |
| `text-title-md` | 36px | 44px | Auth page headings, 404 error |
| `text-title-sm` | 30px | 38px | Auth page headings (mobile) |
| `text-theme-xl` | 20px | 30px | Card titles, section headings |
| `text-base` | 16px | 24px | Body text (Tailwind default) |
| `text-theme-sm` | 14px | 20px | **Default body**, table cells, form labels, descriptions, buttons |
| `text-theme-xs` | 12px | 18px | Small meta text, timestamps, status labels |

### Font weights

| Weight | Usage |
|---|---|
| `font-normal` (400) | Body text, table cell data |
| `font-medium` (500) | Form labels, nav items, headings, button text |
| `font-semibold` (600) | Page titles, card titles, section headers |
| `font-bold` (700) | 404 page, auth page titles |

### Line heights

Default line-height is `normal` (1.2–1.4 depending on font size). Explicit line-heights are set on title tokens only.

## 3.3 Spacing

The project uses Tailwind's default spacing scale (multiples of 4px). Dominant spacing values:

| Class | Value | Usage |
|---|---|---|
| `p-3` | 12px | Tight card padding (compact) |
| `p-4` | 16px | Standard component padding |
| `p-5` | 20px | Card padding, page section padding |
| `p-6` | 24px | Wide card padding, page content area (md+) |
| `gap-2` | 8px | Tight flex/grid gaps |
| `gap-3` | 12px | Standard flex/grid gaps |
| `gap-4` | 16px | Section gaps |
| `gap-5` | 20px | Wide section gaps |
| `space-y-5` | 20px | Form field stack spacing |
| `space-y-6` | 24px | Page section stack spacing |
| `mb-6` | 24px | Heading bottom margin |
| `px-3` | 12px | Sidebar padding |
| `px-5`, `px-6` | 20–24px | Page content horizontal padding |

Max-width constraint for page content: `max-w-(--breakpoint-2xl)` (1536px).

## 3.4 Shadows

| Token | Value | Usage |
|---|---|---|
| `shadow-theme-xs` | `0px 1px 2px 0px rgba(16,24,40,0.05)` | Subtle card elevation |
| `shadow-theme-sm` | `0px 1px 3px 0px rgba(16,24,40,0.1), 0px 1px 2px 0px rgba(16,24,40,0.06)` | Cards, small UI elements |
| `shadow-theme-md` | `0px 4px 8px -2px rgba(16,24,40,0.1), 0px 2px 4px -2px rgba(16,24,40,0.06)` | Header dropdown, elevated cards |
| `shadow-theme-lg` | `0px 12px 16px -4px rgba(16,24,40,0.08), 0px 4px 6px -2px rgba(16,24,40,0.03)` | Modals, drawers |
| `shadow-theme-xl` | `0px 20px 24px -4px rgba(16,24,40,0.08), 0px 8px 8px -4px rgba(16,24,40,0.03)` | Datepicker, large overlays |
| `shadow-focus-ring` | `0px 0px 0px 4px rgba(70,95,255,0.12)` | Focus ring on interactive elements |
| `shadow-slider-navigation` | `0px 1px 2px 0px rgba(16,24,40,0.1), 0px 1px 3px 0px rgba(16,24,40,0.1)` | Carousel nav buttons |
| `shadow-tooltip` | `0px 4px 6px -2px rgba(16,24,40,0.05), -8px 0px 20px 8px rgba(16,24,40,0.05)` | Tooltips |

## 3.5 Breakpoints

| Breakpoint | Min-width | Usage |
|---|---|---|
| `2xsm` | 375px | Small phones |
| `xsm` | 425px | Large phones |
| `sm` | 640px | Tablets portrait |
| `md` | 768px | Tablets landscape |
| `lg` | 1024px | Desktop narrow |
| `xl` | 1280px | **Desktop standard** — sidebar collapse toggle activates |
| `2xl` | 1536px | Desktop wide (content max-width) |
| `3xl` | 2000px | Ultra-wide |

## 3.6 Border radius

| Radius | Value | Usage |
|---|---|---|
| `rounded-lg` | 8px | Buttons, inputs, cards, nav items |
| `rounded-xl` | 12px | Alerts, dropdown menus |
| `rounded-2xl` | 16px | **Primary card radius** — page headers, table shells, component cards, modals |
| `rounded-full` | 9999px | Badges, avatars, status indicators |

## 3.7 Z-index scale

| Token | Value | Used by |
|---|---|---|
| `z-1` | 1 | Body, page content |
| `z-9` | 9 | — |
| `z-40` | 40 | Dropdown menus |
| `z-50` | 50 | Sidebar |
| `z-99` | 99 | — |
| `z-999` | 999 | Header |
| `z-9999` | 9,999 | — |
| `z-99999` | 99,999 | Modals |
| `z-999999` | 999,999 | Toast container |

## 3.8 Icons

**Library:** `@ng-icons/core` with `@ng-icons/heroicons/outline`

**Icon usage conventions:**
- Sidebar navigation: 24px Heroicons outline (`size="24"`)
- Form input icons (email, password): 18px (`size="18"`)
- Button icons: 16–20px via SVG string (passed as `startIcon`/`endIcon`)
- Status/category icons: 20px inline SVGs
- Toast/alert icons: 20px inline SVGs
- Dropdown chevrons: 20px Heroicons

**Frequently used icons:** `heroSquares2x2` (dashboard), `heroUserGroup` (children), `heroEnvelope` (invites), `heroClipboardDocumentCheck` (attendance), `heroClipboardDocumentList` (corrections), `heroDocumentText` (invoices), `heroCog6Tooth` (settings), `heroHome` (breadcrumb home), `heroEnvelope`/`heroLockClosed`/`heroEye`/`heroEyeSlash` (auth form).

---

# 4. Layout

## 4.1 App shell (authenticated)

```
┌─────────────────────────────────────────────┐
│  Header (sticky, z-99999)                    │
│  [☰ toggle] [logo]    [🌙] [user▼]          │
├──────────┬──────────────────────────────────┤
│ Sidebar  │ Breadcrumb                        │
│ (fixed)  │───────────────────────────────────│
│ z-50     │ page-content (max-w-2xl)          │
│ 290px    │ router-outlet                      │
│ or 90px  │                                    │
│          │                                    │
│          │                                    │
│          │                                    │
├──────────┴──────────────────────────────────┤
│              Toast container (fixed top-right)│
└─────────────────────────────────────────────┘
```

**Layout structure:** Flex column on mobile, flex row on desktop. Sidebar is fixed `left-0` with `z-50`. Content area uses `xl:ml-[290px]` when sidebar is expanded, `xl:ml-[90px]` when collapsed.

**Transition:** `transition-all duration-300 ease-in-out` on both sidebar width and content margin.

**Page content wrapper:** `p-4 mx-auto max-w-(--breakpoint-2xl) md:p-6`

## 4.2 Sidebar

- **States:** Expanded (290px), Collapsed (90px), Hover (expands when collapsed if hovered), Mobile (full overlay)
- **Mobile:** Slides in from left (`translate-x-0` / `-translate-x-full`), with gray backdrop overlay (`z-40 bg-gray-900/50`)
- **Logo:** Full logo (150×40) when expanded, icon (32×32) when collapsed. Light/dark variants
- **Nav groups:** Role-based filtering; each group has an uppercase section header (hidden when collapsed, replaced by a thin divider line)
- **Active indicator:** 3px brand-colored left border (`left-0 w-[3px] h-5 rounded-full bg-brand-500`) + `bg-brand-50` background
- **Hover:** `bg-gray-100` on inactive items
- **Icons:** Always visible; 24px Heroicons
- **Labels:** Only visible when expanded/hovered/mobile
- **Scroll:** `overflow-y-auto no-scrollbar`

## 4.3 Header

- **Position:** `sticky top-0 z-99999`
- **Left section:** Hamburger/cross toggle button (44×44 rounded-lg border), mobile logo, mobile application menu dots button
- **Right section:** Theme toggle (sun/moon), User dropdown (avatar initials + email + role + sign out)
- **Responsive:** On mobile (`xl:hidden`), right section is hidden behind the application menu toggle

## 4.4 Breadcrumb

- **Location:** Below header, above `router-outlet`
- **Data source:** Route `data.breadcrumb` object with `label`, optional `link`, and optional `resolve` function
- **Structure:** Home → Section → Subsection → Current page
- **Home link:** Dynamic — points to the authenticated user's default route based on role
- **Mobile collapse:** When breadcrumbs exceed `mobileCollapseAfter` (default 2), intermediate crumbs are hidden behind a "..." toggle button
- **Separator:** Chevron icon between crumbs
- **Last crumb:** Current page (plain text, no link)

## 4.5 Auth page layout

```
┌──────────────────────────────┬──────────────┐
│                              │  Decorative   │
│   Auth form content          │  grid pattern │
│   (centered, max-w-md)       │  (hidden on   │
│                              │   mobile)     │
│                              │   Logo        │
│                              │   Tagline     │
├──────────────────────────────┴──────────────┤
│          🌙 Theme toggle (bottom-right)      │
└─────────────────────────────────────────────┘
```

Two-column flex layout: left (form) takes full width on mobile, `lg:w-1/2` on desktop. Right column (hidden on mobile) has a decorative grid SVG background, the app logo, and a product description.

## 4.6 Page header

Each page starts with a `app-page-header` component:
- White card (`rounded-2xl border border-gray-200 bg-white`)
- Flex column on mobile, row on md+ (`flex-col md:flex-row md:items-center md:justify-between`)
- Title (`text-xl font-semibold text-gray-800`)
- Optional description (`text-sm text-gray-500`)
- Optional actions area (projected via `[actions]` selector)
- Padding: `p-5`

## 4.7 Table shell

Wraps all data tables in a `app-table-shell`:
- White card (`rounded-2xl border border-gray-200 bg-white`)
- Optional header with title + description + actions
- `overflow-x-auto` for horizontal scroll on narrow screens
- Optional footer slot

---

# 5. Navigation

## 5.1 Role-based routing

Managed through Angular route guards:
- `authGuard` — redirects to `/signin` if unauthenticated
- `roleGuard` — restricts routes by role (`data.roles` array), redirects to 403/not-found
- `roleDefaultRedirectGuard` — redirects root `/` to the authenticated user's home page

### Route structure

```
/ (protected, redirects to role home)
├── manager/
│   ├── dashboard
│   ├── children/ (index, :childId, new, :childId/edit, :childId/booking-pattern, :childId/:tab)
│   ├── invites
│   ├── attendance
│   ├── attendance-corrections
│   ├── rooms/ (index, new, :roomId/edit)
│   ├── session-types/ (index, new, :sessionTypeId/edit)
│   ├── session-templates
│   ├── funding
│   ├── invoices/ (index, new, :invoiceId, :invoiceId/edit)
│   ├── billing-setup
│   ├── site-settings
│   └── site-profile
├── practitioner/
│   └── attendance
├── parent/
│   └── invoices/ (index, :invoiceId)
├── owner/
│   ├── (overview)
│   ├── rooms/ (index, new, :roomId/edit)
│   ├── session-types/ (index, new, :sessionTypeId/edit)
│   └── manager-access
├── signin (public)
├── signup (public)
├── forgot-password (public)
├── reset-password (public)
├── invite-accept (public)
└── 404 (catch-all)
```

## 5.2 Sidebar navigation groups

| Role | Groups | Items |
|---|---|---|
| Manager | Overview | Dashboard |
| | People | Children |
| | Attendance | Attendance |
| | Billing | Invoices |
| | Setup | Site settings |
| Practitioner | Workday | Attendance |
| Owner | Overview | Overview |
| | Access | Manager access |
| Parent | Billing | Invoices |

---

# 6. Components

## 6.1 Button (`app-button`)

**Variants:** `primary`, `outline`, `secondary`, `success`, `warning`, `danger`, `ghost`, `link`

**Sizes:** `xs`, `sm`, `md`, `lg`

**States:** Normal, hover, disabled, loading (shows spinner, disables interaction)

**Features:**
- Optional `startIcon` / `endIcon` SVG strings (sanitized via `safeHtml` pipe)
- `type` attribute configurable (button/submit)
- Custom `className` for additional styling
- Emits `btnClick` on click

**Default classes:** `inline-flex items-center justify-center gap-2 rounded-lg transition`

**Loading spinner:** 16px animated SVG (`animate-spin`)

**Primary variant:** `bg-brand-500 text-white hover:bg-brand-600`
**Outline variant:** `border border-gray-300 text-gray-700 hover:bg-gray-50`
**Disabled state:** `opacity-50 cursor-not-allowed`

## 6.2 Badge (`app-badge`)

**Variants:** `light` (colored bg + colored text), `solid` (filled bg + white text)

**Sizes:** `sm`, `md`

**Colors:** `primary`, `success`, `error`, `warning`, `info`, `light`, `dark`

**Shape:** `rounded-full` (pill)

**Features:** Optional `startIcon`/`endIcon` SVGs

**Status badge (`app-status-badge`):** Domain-aware wrapper that maps known status keys to the correct BadgeComponent color/label. Handles: `active`, `paid`, `overdue`, `checked_in`, `absent`, `pending`, `expired`, `revoked`, `enrolled`, `draft`, and more. Falls back to title-case for unknown statuses.

## 6.3 Alert (`app-alert`)

**Variants:** `success`, `error`, `warning`, `info`

**Features:**
- Title + message
- Optional router link (href + text)
- Compact mode (`p-2` instead of `p-4`)
- Dismissible mode (emits `dismissed`)
- Unique SVG icon per variant
- Correct ARIA role (`alert` for error/warning, `status` otherwise)

**Shape:** `rounded-xl border p-4`

## 6.4 Toast (`app-toast-container`)

**Position:** Fixed top-right (`fixed top-4 right-4 z-100000`)

**Managed by:** `ToastService` — provides `success()`, `error()`, `warning()`, `info()` methods

**Behaviour:**
- Auto-dismisses after 5000ms
- Max 5 visible toasts
- Each toast has unique ID, variant, title, message
- Dismiss button on each toast

## 6.5 Modal (`app-modal`)

**Features:**
- Overlay with backdrop blur (`bg-gray-400/50 backdrop-blur-[32px]`)
- Close button (top-right, `right-3 top-3 z-999`)
- Escape key closes
- Backdrop click closes
- Focus trapping (restores previous focus on close)
- Scroll locking (`overflow-y-auto` on container)
- Fullscreen mode
- ARIA: `role="dialog"`, `aria-modal="true"`, `aria-label` / `aria-labelledby`

**Shape:** `rounded-3xl bg-white dark:bg-gray-900` (standard), full viewport (fullscreen)

**Z-index:** `z-99999`

## 6.6 Confirmation dialog (`app-confirmation-dialog`)

Wraps `app-modal`. Provides:
- Title, message, content projection
- Cancel + Confirm buttons
- Loading spinner on confirm
- `primary` or `danger` variant for confirm button
- Emits `confirmed`, `cancelled`

## 6.7 Drawer (`app-drawer`)

**Positions:** Left, Right

**Sizes:** `sm` (w-80), `md` (w-96), `lg` (w-[480px]), `xl` (w-full max-w-md)

**Features:**
- Backdrop with blur
- Header with title + close button
- Scrollable content area
- Escape key closes

## 6.8 Dropdown (`app-dropdown`)

**Features:**
- Absolutely positioned below trigger
- Click-outside-to-close (document mousedown listener)
- Ignores clicks on `.dropdown-toggle` elements
- Rounded-xl border card
- Z-index: `z-40`

**Dropdown item (`app-dropdown-item`):** Button-styled item, emits `itemClick`
**Dropdown item two (`app-dropdown-item-two`):** Router-link variant (`<a>` with `[routerLink]`)

## 6.9 Avatar (`app-avatar`)

**Sizes:** `xsmall` (24px), `small` (32px), `medium` (40px default), `large` (48px), `xlarge` (56px), `xxlarge` (80px)

**Status indicator:** Optional online (green), offline (gray), busy (red) dot — positioned bottom-right, sized proportionally

**Text avatar (`app-avatar-text`):** Extracts up to 2 initials from name. Deterministic background color from palette of 8 colors

## 6.10 Table system

### Table shell (`app-table-shell`)
Card wrapper: `rounded-2xl border border-gray-200 bg-white`. Optional header (title + description + actions slot), `overflow-x-auto` content area, optional footer.

### Table (`app-table`)
Semantic `<table>` with `min-w-full text-left text-sm`.

### Table header (`app-table-header`)
`<thead>` with `border-b border-gray-200 text-gray-500`.

### Table body (`app-table-body`)
`<tbody>` wrapper.

### Table row (`app-table-row`)
`<tr>` with `border-b border-gray-100 dark:border-gray-800/60`.

### Table cell (`app-table-cell`)
Dual-purpose: renders `<th>` if `isHeader`, else `<td>`. Padding: `px-5 py-3`.

### Table pagination (`app-table-pagination`)
Previous/Next navigation with "Showing X - Y of Z" text. Buttons disabled at boundaries or when loading. Emits `previous`, `next`. Inputs: `offset`, `limit`, `total`.

**Responsive table pattern:** Many list pages use a dual-rendering approach:
- Desktop (`hidden lg:block`): Standard `<table>` from the table component system
- Mobile (`lg:hidden`): Card-based layout — each row is a card with key-value pairs

## 6.11 Form field (`app-form-field`)

Wrapper with:
- Optional `<label>` with `*` required indicator
- Content projection for input control
- Optional hint text (shown when no error)
- Optional error message `<p>` in `text-error-500`

## 6.12 Input field (`app-input-field`)

**Features:**
- Implements `ControlValueAccessor` (usable with `[(ngModel)]` or reactive forms)
- Supports: `text`, `email`, `password`, `number`, `tel`, `url`, `search`, `date` types
- Error state: red border + optional hint text
- Success state: green border
- Disabled state: `opacity-50 cursor-not-allowed bg-gray-100`
- Hint text: below input, `text-xs`
- `autocomplete`, `inputMode`, `min`/`max`/`step` for number type
- Emits `valueChange` and `blurred`

## 6.13 Select (`app-select`)

Native `<select>` with custom arrow SVG. Implements `ControlValueAccessor`. Supports placeholder (first disabled option), error state, custom className.

## 6.14 Multi-select (`app-multi-select`)

Custom multi-select dropdown. Selected options shown as removable chips inside input. Dropdown opens below with selectable options (click to toggle). Emits `selectionChange` with `string[]`.

## 6.15 Date picker (`app-date-picker`)

Uses Flatpickr. Modes: `single`, `multiple`, `range`, `time`. Configurable min/max dates. Emits `dateChange` with selected dates.

## 6.16 Time picker (`app-time-picker`)

Flatpickr time-only mode (`enableTime: true, noCalendar: true`). 12-hour format, 1-minute increment. Emits `timeChange` with `"HH:mm"`.

## 6.17 Checkbox (`app-checkbox`)

Custom styled checkbox implementing `ControlValueAccessor`. Hidden native input, branded checkmark SVG on checked state. Disabled variant with gray checkmark. Supports label text.

## 6.18 Radio (`app-radio`, `app-radio-sm`)

Custom styled radio. Hidden native input, brand-500 fill when selected, inner white dot. `app-radio-sm` is smaller (16px vs 20px).

## 6.19 Switch toggle (`app-switch`)

Toggle switch implementing `ControlValueAccessor`. Two color variants: `blue` (brand-500) and `gray`. Circular knob slides horizontally on track. Supports disabled state, label, aria-label.

## 6.20 Text area (`app-text-area`)

Multiline textarea, implements `ControlValueAccessor`. Configurable rows (default 3), min-height 120px. Error/disabled states, hint text.

## 6.21 Phone input (`app-phone-input`)

Phone number with country code selector. Country code dropdown at `start` or `end`. Pre-populates with country code prefix.

## 6.22 File input (`app-file-input`)

Custom styled "Browse" button using Tailwind's `file:` modifier. Emits `change` event.

## 6.23 Dropzone (`app-dropzone`)

Drag-and-drop file upload area. Accepts PNG/JPG/WebP/SVG. Visual feedback on drag over. Click to browse.

## 6.24 Label (`app-label`)

`<label>` element: `mb-2 block text-sm font-medium text-gray-700 dark:text-gray-400`.

## 6.25 Empty state (`app-empty-state`)

Centered layout:
- Optional icon (HTML, rendered via `[innerHTML]`)
- Title (`text-sm font-medium text-gray-800`)
- Optional message (`text-sm text-gray-500`)
- Optional action buttons (content projection)
- Compact mode (`py-4` vs `py-8`)

## 6.26 Loading state (`app-loading-state`)

Animated SVG spinner with label. Three variants:
- `block` — centered `<div>` for full-page loading
- `table-row` — `<tr><td colspan={colspan}>` for table loading
- `inline` — `inline-flex` span for inline loading

## 6.27 Component card (`app-component-card`)

Generic card container: `rounded-2xl border border-gray-200 bg-white`. Optional title + description in header (separated by `border-t`), content projection in `p-4 sm:p-6`.

## 6.28 Page breadcrumb (`app-page-breadcrumb`)

Dynamic breadcrumbs from route data. Pre-pends "Home" with role-based default route. Supports resolvers (observable/string). Mobile collapse with "..." toggle. Uses `@ng-icons/core` for home icon. `ChangeDetectionStrategy.OnPush`.

## 6.29 Theme toggle (`app-theme-toggle-button`)

Circular 44×44 button (`rounded-full border`). Toggles between moon (dark mode) and sun (light mode) SVGs. Uses `ThemeService` which persists to localStorage and adds/removes `dark` class on `<html>`.

## 6.30 Table dropdown (`app-table-dropdown`)

Row action dropdown using Popper.js. Two content projection slots: `[dropdown-button]` (trigger) and `[dropdown-content]` (popped menu). Opens on click, closes on outside click.

## 6.31 User dropdown (`app-user-dropdown`)

Header dropdown showing avatar-text initials, email, session role label, and "Sign out" button. Calls `authService.logout()`.

## 6.32 Over-capacity banner (`app-over-capacity-banner`)

Error-styled banner (`role="alert"`) showing rooms exceeding capacity. Takes `OverCapacityRoom[]`. Shows exclamation icon, heading, anchor-linked room names.

## 6.33 Countdown timer (`app-countdown-timer`)

Live countdown to target date. Shows `dd : hh : mm : ss` in brand-500 color. Updates every 1 second via `setInterval`. Cleans up on `ngOnDestroy`.

## 6.34 Chart tab (`app-chart-tab`)

Three-option segmented control (Monthly/Quarterly/Annually). Pill-style toggle inside `rounded-lg` container. Active option has shadow elevation.

---

# 7. Form patterns

## 7.1 Form layout

- Single-column forms with `max-w-md` or full-width
- `space-y-5` between fields
- Field groups use `<div>` per field with `label + input/select + optional hint`
- Submit button full-width (`w-full`) for auth forms, or inline for data-entry forms

## 7.2 Validation

- Client-side validation on submit (not keystroke)
- Field-level error messages shown via `[hint]` binding on `app-input-field`
- General form errors shown in red alert banner above the form
- API errors mapped through `ApiErrorMapper` → `presentApiError` → user-readable messages

## 7.3 Multi-step form (child registration)

6-step stepper wizard:
```
1. Child Info → 2. Address → 3. Health → 4. Funding → 5. Documents → 6. Review
```

- Steps shown as numbered circles with connector lines
- Active step highlighted with brand-500
- Completed steps shown with checkmark
- Each step validates before allowing "Continue"
- Data persisted in `registration-draft.storage.ts` (localStorage)
- "Back" button returns to previous step without validation
- Final step shows review summary before submitting

## 7.4 Membership challenge

When a user account has multiple nursery memberships, the sign-in form shows a selection step after email/password validation. Available memberships are displayed as selectable cards showing tenant name, branch name, and role badge.

---

# 8. Data display patterns

## 8.1 Dashboard metrics

The manager dashboard uses a 4-column metric tile grid with icon + label + value + meta + description. Tiles have semantic tone classes (`success`/`warning`/`neutral`). Each tile links to the relevant detail page.

## 8.2 Lists with table/card dual rendering

Pages like children list, rooms list, invoices list use:
- **Desktop:** Standard HTML table with sortable column headers, checkboxes for selection, action buttons per row
- **Mobile:** Card-per-item layout with key info displayed as label-value pairs

## 8.3 Tabbed detail view

The child detail page uses tab navigation (`overview | attendance | funding | health | contacts`) that updates the URL (`/manager/children/:childId/:tab`). Active tab has brand-500 underline indicator. Content changes without page reload via paramMap subscription.

## 8.4 Pagination

Offset-based pagination with Previous/Next buttons. Shows "Showing X - Y of Z". Used on children list, invoices list, and similar data tables.

## 8.5 Search and filter

- Search with 200ms debounce to avoid excessive API calls
- Status filter dropdowns (e.g., `active/inactive/all` for children)
- Attendance status filter pills (all/checked-in/not checked-in/absent) on practitioner page
- Sortable columns with ascending/descending toggle via sort icons

## 8.6 Attendance status display

Individual child attendance records show:
- Status badge: `checked_in` (green), `absent` (red), `not_checked_in` (gray)
- Check-in/out times with 24h format
- Incomplete session warning indicator
- Action buttons: check-in, check-out, mark absent, clear absence
- Real-time clock in header, auto-polling every 30 seconds

## 8.7 Invoice status display

Invoice items show status badges: `paid`, `overdue`, `pending`, `draft`, `cancelled`. Totals displayed in GBP format (minor units → formatted string). Payment follow-up items show outstanding amount and status.

---

# 9. Response handling

## 9.1 Loading

- **Initial load:** Centered `app-loading-state` (block variant) with spinner + "Loading..." label
- **Background refresh:** Subtle spinner/indicator without disrupting the current view
- **Button loading:** 16px animated spinner replaces button text; button is disabled during submission
- **Table loading:** `app-loading-state` with `table-row` variant showing colspan spinner
- **Card loading:** `isLoadingCards` boolean hides mobile card view until loaded

## 9.2 Empty states

When a list has no items:
- Centered layout with optional icon
- Title ("No children found")
- Descriptive message
- Optional action button (e.g., "Add your first child")

## 9.3 Error states

- **Form errors:** Red border on input + hint text below field
- **General errors:** Red alert banner at top of form/page (rounded-lg border border-red-200 bg-red-50)
- **API errors:** Mapped through `ApiErrorMapper` → `presentApiError` → human-readable message
- **Network/down errors:** Descriptive message with retry option
- **Correction form errors:** Field-level error display + "issued invoice warning" banner for billing-impacting corrections
- **Token errors:** Specific screen states for expired/invalid/accepted/revoked invite/reset tokens

## 9.4 Success feedback

- **Toast notifications:** Brief success message auto-dismissed after 5s
- **Screen transition:** Auth flows transition to success state (check email, password reset, invitation accepted)
- **Inline message:** Correction submission shows success message above form

---

# 10. Interaction patterns

## 10.1 Navigation

- Sidebar: 300ms ease-in-out transition
- Breadcrumb: Click any crumb (except current) to navigate
- Tab navigation: URL-driven, preserves query params
- Mobile sidebar: slides in from left, backdrop click to close

## 10.2 Data entry

- Tab key navigates between fields in natural DOM order
- Submit on Enter within forms
- Password visibility toggle
- Date/time via Flatpickr calendar popover
- Multi-select via chip-based dropdown

## 10.3 Table interactions

- Row hover: subtle background change
- Sort: click column header, toggle between asc/desc/none
- Search: debounced 200ms, real-time client-side filtering after initial load
- Pagination: offset-based, Previous/Next controls

## 10.4 Attendance (practitioner)

- Auto-polling every 30s with background refresh indicator
- Manual refresh button
- Real-time clock display (24h UK format)
- Status filter pills with counts
- Check-in/out via single click with loading state
- Mark/clear absence with confirmation

## 10.5 Corrections (manager)

- Step 1: Select child
- Step 2: Select date → shows sessions for that day
- Step 3: Select session or choose "missed session" mode
- Step 4: Enter corrected check-in/out times + reason
- Invoice warning displayed if corrections affect issued invoices
- Reason selection: dropdown (late arrival, early pick-up, missed swipe, etc.) + optional note

## 10.6 Dark mode

- Toggled by `ThemeToggleButtonComponent` → `ThemeService`
- Persisted to `localStorage`
- `.dark` class on `<html>` element
- All components have `dark:` variants for every color, shadow, border, and background
- Smooth class-based switching (no flash)

---

# 11. Responsive behavior

## 11.1 Mobile (< 640px)
- Sidebar hidden behind hamburger menu overlay
- Tables converted to card-per-item layouts
- Header shows application menu (3-dot) for secondary controls
- Single-column layouts throughout
- Breadcrumb collapse to last 2 items with "..." toggle

## 11.2 Tablet (640px – 1023px)
- Sidebar hidden behind hamburger menu overlay
- 2-column grids possible
- Tables still scroll horizontally (`overflow-x-auto`)
- Forms remain single-column

## 11.3 Desktop (1024px – 1279px)
- Sidebar visible, always expanded (290px)
- Multi-column layouts active
- Standard tables visible
- Action buttons shown inline

## 11.4 Wide desktop (1280px+)
- Sidebar collapsible (290px ↔ 90px)
- Hover expand when collapsed
- Page content constrained to `max-w-(--breakpoint-2xl)` (1536px)
- Multi-column metric grids

---

# 12. Implementation rules

## 12.1 Component authoring

- **Standalone components** (Angular 14+ style) — all components use `standalone: true`
- **Inline templates** preferred for simple components (<30 lines), separate `.html` for complex ones
- **OnPush change detection** only for the breadcrumb component (performance-critical)
- **Content projection** via `<ng-content>` with selectors for multiple slots
- **ControlValueAccessor** for all form controls (compatible with template-driven and reactive forms)

## 12.2 Styling

- **Tailwind CSS v4** with `@theme` custom properties
- **No component-level CSS files** — all styling via Tailwind utility classes
- **Dark mode** via `dark:` variant prefix on every component
- **Custom utilities** defined with `@utility` for frequently used patterns (menu items, scrollbars)
- **Third-party overrides** in `styles.css` for Flatpickr, ApexCharts, FullCalendar, Swiper, Simplebar, Prism.js
- **Layer base** for global defaults (border-color, cursor on buttons, body background)

## 12.3 Icons

- Heroicons outline via `@ng-icons/core`
- Provided locally per component via `provideIcons()`
- Inline SVG strings used for custom/action icons
- `safeHtml` pipe for rendering raw SVG in templates

## 12.4 Services

- `ToastService` — singleton, manages toast state via signal
- `ModalService` — singleton, manages modal open/close via BehaviorSubject
- `SidebarService` — singleton, manages expand/collapse/hover via BehaviorSubjects
- `ThemeService` — singleton, manages dark/light toggle with localStorage persistence
- All services `providedIn: 'root'`

## 12.5 Error handling pattern

```
API error → ApiErrorMapper.map() → MappedApiError
         → presentApiError(mappedError, context) → PresentedApiError
         → display fieldErrors on inputs + formError alert
```

- Field errors: `mapped.fieldErrors['fieldName']` → `input.hint`
- General errors: `formatPresentedApiError(presented)` → form-level alert banner
- Token/invite errors: mapped to specific screen states (not generic error messages)

## 12.6 Guard pattern

```
Route → authGuard (check session) → roleGuard (check role in data.roles) → component
```

- `authGuard`: redirects to `/signin` if no valid session
- `roleGuard`: shows 404 if user's role not in route's `data.roles` array
- `roleDefaultRedirectGuard`: redirects `/` to `defaultRouteForRole(currentRole())`

## 12.7 Naming conventions

- **Selectors:** `app-{feature}-{component}` (e.g., `app-page-header`, `app-table-pagination`, `app-input-field`)
- **Files:** `{feature}.{type}.ts` (e.g., `button.component.ts`, `toast.service.ts`, `parent-invoice-formatters.ts`)
- **Routes:** lowercase kebab-case (`/manager/attendance-corrections`, `/owner/manager-access`)
- **TS classes:** PascalCase (e.g., `ManagerChildrenComponent`, `ToastService`)
- **Testing:** Co-located `.spec.ts` files next to their source files
- **Models:** Separate `*.models.ts` files per feature module

## 12.8 File structure per feature

```
features/{role}/
├── data/
│   └── {feature}-api.service.ts
├── models/
│   └── {feature}.models.ts
├── pages/
│   └── {feature}/
│       ├── {feature}.component.ts
│       └── {feature}.component.html
└── utils/
    └── {feature}-formatters.ts
```

Shared components go under `shared/components/{category}/`. Layout under `shared/layout/`. Auth under `pages/auth-pages/`. Core services/middleware under `core/{services,guards,interceptors,models}/`.

---

# 13. Design tokens (CSS custom properties)

All design tokens are defined in `web/src/styles.css` via Tailwind CSS v4 `@theme` directive and are available as utility classes throughout the app.

| Category | Tokens |
|---|---|
| Font | `--font-outfit` |
| Breakpoints | `--breakpoint-2xsm`, `-xsm`, `-3xl`, `-sm` through `-2xl` |
| Text | `--text-title-2xl` through `-sm`, `--text-theme-xl` through `-xs` |
| Colors | `--color-brand-25` through `-950` (11 steps), `--color-gray-25` through `-950` (12 steps), `--color-success-*`, `--color-error-*`, `--color-warning-*`, `--color-orange-*`, `--color-blue-light-*`, `--color-theme-pink-500`, `--color-theme-purple-500` |
| Shadows | `--shadow-theme-xs` through `-xl`, `--shadow-focus-ring`, `--shadow-slider-navigation`, `--shadow-tooltip`, `--shadow-datepicker` |
| Z-index | `--z-index-1` through `-999999` |

---

# 14. Third-party libraries (UI)

| Library | Purpose | Integration |
|---|---|---|
| Tailwind CSS v4 | Utility-first CSS framework | `@tailwindcss` + `@theme` directive |
| Flatpickr | Date/time pickers | Custom overrides in `styles.css` for consistent theming |
| ApexCharts | Dashboard charts | Themed via CSS overrides for legend, tooltip, grid |
| FullCalendar | Calendar view | Custom button styles, event colors, toolbar overrides |
| Swiper | Carousels | Custom nav button and pagination styles |
| Simplebar | Custom scrollbars | Themed via CSS pseudo-element overrides |
| Prism.js | Code syntax highlighting | Token color overrides for light/dark themes |
| ng-icons | Icon library | Heroicons outline set, provided per-component |

---

# 15. Testing conventions

- Tests co-located with source: `{component}.spec.ts` next to `{component}.ts`
- Test patterns found: component creation tests, service behavior tests, formatter tests
- No end-to-end or visual regression tests detected in the codebase
