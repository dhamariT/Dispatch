# Dispatch Dashboard — Implementation Plan

This document captures every design decision made so far for the Dispatch
frontend. A future agent should be able to read this and build the
dashboard without re-deriving any of the thinking below.

## Design Philosophy

Dispatch's dashboard is inspired by [Coder's](https://coder.com/) approach
to product UI: clean, spacious, professional. Not flashy, not minimal —
just clear hierarchy and generous whitespace. The dashboard IS the product.
There is no separate marketing site right now.

### Core Principles

1. **Clean and simple** — generous whitespace, clear hierarchy. The
   operator knows exactly what the product does in 5 seconds.
2. **Product screenshot is front and center** — the actual diff view is
   the first thing anyone sees. No login walls, no empty states.
3. **The dashboard is the marketing** — when someone lands on `/`, they
   see Dispatch working. For self-hosted instances, that's the real fleet.
   For the demo, that's simulated data.

### What We're NOT Building

- A separate marketing site (that comes later, wrapping a demo instance)
- A fleet management tool (Balena already does that)
- An alerting system (Dispatch shows diffs, the operator decides)

## Architecture

### Two Deployment Modes, Same Codebase

1. **Self-hosted** — operator runs `dispatchd`, connects their Balena
   fleet, sees their real data.
2. **Demo** — a hosted `dispatchd` instance with simulated fleet data.
   Same code, just seed data.

### Tech Stack

- **Next.js 16** with App Router and Turbopack
- **TypeScript**
- **Tailwind CSS 4** + **shadcn/ui** for components
- **Biome** for linting/formatting
- **Storybook** for component development and visual testing
- **pnpm** as package manager

### Project Structure

```
site/src/
├── app/
│   ├── layout.tsx              # Root layout, nav bar
│   ├── page.tsx                # Main dashboard (diff view)
│   └── (dashboard)/            # Future route group for sub-pages
├── components/
│   ├── ui/                     # shadcn primitives (Button, etc.)
│   ├── admin/ui/               # Admin-specific shadcn components
│   ├── diff-table.tsx          # The core diff table component
│   ├── deploy-tabs.tsx         # Horizontal deploy selector
│   ├── device-group.tsx        # Per-device metric group in the table
│   ├── metric-row.tsx          # Single metric row (before/after/delta)
│   ├── status-badge.tsx        # Deploy status pill (Canary, Soaking, etc.)
│   ├── soak-timer.tsx          # Soak progress indicator
│   ├── demo-banner.tsx         # "This is a demo fleet" banner
│   └── nav.tsx                 # Top navigation bar
├── api/                        # API client for dispatchd
│   └── client.ts               # fetch wrapper for /api/* endpoints
├── hooks/
│   └── use-diff.ts             # Hook to fetch and poll diff data
├── lib/
│   ├── utils.ts                # shadcn utility (already exists)
│   ├── demo-data.ts            # Simulated fleet data for first-run
│   └── format.ts               # Metric formatting (%, MB, ms, etc.)
├── theme/
│   ├── constants.ts            # Layout constants (already exists)
│   ├── roles.ts                # Semantic color roles (already exists)
│   └── index.ts                # Re-exports (already exists)
└── .storybook/                 # Storybook config
```

## Theme System

### Already Built

Theme files exist at `site/src/theme/`. CSS variables are wired into
`globals.css` for both light and dark mode, and registered as Tailwind
color utilities.

### Semantic Roles (Deploy State Colors)

These are the core visual language of the dashboard. Every metric delta,
every device status, every deploy tab uses these:

| Role | Meaning | Usage |
|------|---------|-------|
| `critical` | Metric regression, something broke | LiDAR accuracy dropped, sensor failure |
| `warning` | Elevated but functional | CPU spike, memory pressure |
| `stable` | Within normal range, no regression | Metrics unchanged or improved |
| `neutral` | No data, informational, baseline | Devices not yet deployed to |
| `active` | In progress, currently relevant | Deploy running, soak timer active |
| `offline` | Device not reporting | Device disconnected, agent down |

Each role has `background`, `border`, `text`, and `fill` variants for
both light and dark mode.

**Tailwind usage:** `bg-critical-bg`, `text-critical`, `border-warning`,
`bg-stable-bg`, etc.

### Layout Constants

| Constant | Value | Usage |
|----------|-------|-------|
| `NAV_HEIGHT` | 56px | Top navigation bar |
| `CONTAINER_WIDTH` | 1280px | Main content max-width |
| `CONTAINER_WIDTH_MD` | 960px | Narrower content areas |
| `SIDE_PADDING` | 24px | Page horizontal padding |
| `BORDER_RADIUS` | 8px | Default border radius |

### Fonts

Geist (sans) and Geist Mono — loaded via `next/font` in `layout.tsx`.
Already configured by `create-next-app`.

### Dark Mode

Start with dark mode as default. Fleet dashboards are often monitored in
dim environments. Light mode should work but dark is the primary design
target.

## Before Writing Components

### Set Up Storybook First

Before building any components, Storybook must be configured:

1. Install `@storybook/react-vite` (or the Next.js adapter if available
   for Storybook 10+).
2. Configure `.storybook/main.ts` — stories glob: `"../src/**/*.stories.tsx"`.
3. Configure `.storybook/preview.tsx` with decorators:
   - **Theme decorator** — wraps stories in light/dark mode, applies
     the `dark` class to `<html>` for Tailwind.
   - **Container decorator** — centers stories in a max-width container
     with proper padding.
4. Add viewport presets for desktop (1440px), tablet (768px), and a
   compact view (400px) for monitoring displays.

### Pull Coder Component References

Before building each component, fetch Coder's equivalent from their
GitHub (`coder/coder/site/src/`) to study their patterns:

- **Table component** — how they handle rows, zebra striping, hover states
- **Badge/pill component** — how they show status labels
- **Navigation** — their top bar layout and responsive behavior
- **Color usage** — how they apply semantic colors to data-heavy UIs

Don't copy their code. Study the patterns, adapt for Dispatch.

## Main Screen Layout

### Wireframe

```
┌─────────────────────────────────────────────────────┐
│ Dispatch          Demo Fleet              ⚙         │  ← nav.tsx
├─────────────────────────────────────────────────────┤
│ ⚠ This is a demo fleet — connect Balena to start   │  ← demo-banner.tsx
├─────────────────────────────────────────────────────┤
│ v1.4.2→v1.4.3    v1.4.1→v1.4.2    v1.4.0→v1.4.1  │  ← deploy-tabs.tsx
│ (red tab)         (green tab)       (green tab)     │
├─────────────────────────────────────────────────────┤
│ ● Canary · Car 2 · Soaking 14m / 30m               │  ← status-badge.tsx
│                                                     │     + soak-timer.tsx
├─────────────────────────────────────────────────────┤
│                                                     │
│ Car 2  ● Deployed                                   │  ← device-group.tsx
│ ┌─────────────┬─────────┬─────────┬────────┐       │
│ │ Metric      │ Before  │ After   │ Delta  │       │  ← diff-table.tsx
│ │ LiDAR acc.  │ 98.1%   │ 91.3%   │ -6.8%  │ 🔴   │  ← metric-row.tsx
│ │ CPU         │ 34%     │ 48%     │ +14%   │ 🟡   │     (color from role)
│ │ Memory      │ 412MB   │ 418MB   │ +6MB   │ ⚪   │
│ └─────────────┴─────────┴─────────┴────────┘       │
│                                                     │
│ Car 1  ○ Waiting                                    │  ← device-group.tsx
│ Car 3  ○ Waiting                                    │     (collapsed, no table)
│                                                     │
├─────────────────────────────────────────────────────┤
│              [ Promote to Wave 2 ]  [ Rollback ]    │  ← action buttons
└─────────────────────────────────────────────────────┘
```

### Component Breakdown

Each component maps to a file. Build them bottom-up (smallest first):

#### 1. `metric-row.tsx`
The atomic unit. One row in the diff table.

**Props:**
- `name: string` — metric name ("LiDAR accuracy")
- `before: number` — value before deploy
- `after: number` — value after deploy
- `format?: "percent" | "bytes" | "ms" | "number"` — display format

**Behavior:**
- Computes delta (`after - before`)
- Assigns a role based on delta severity (configurable thresholds later,
  hardcoded for now)
- Colors the row using the role's `background` and the delta using the
  role's `text`

**Stories:**
- `MetricRow.stories.tsx` — critical, warning, stable, neutral states.
  One story per role.

#### 2. `status-badge.tsx`
A colored pill showing deploy phase.

**Props:**
- `status: "canary" | "wave2" | "wave3" | "complete" | "rolledback" | "soaking"`

**Behavior:**
- Maps status to a role color and label
- `canary` / `soaking` → `active`
- `complete` → `stable`
- `rolledback` → `critical`

**Stories:**
- One story per status value.

#### 3. `soak-timer.tsx`
Shows soak progress.

**Props:**
- `elapsed: number` — minutes elapsed
- `total: number` — total soak time configured

**Behavior:**
- Shows "14m / 30m" text
- Optional progress bar (just a styled div, percentage width)
- Uses `active` role color

**Stories:**
- Early soak (10%), mid soak (50%), almost done (90%), complete (100%).

#### 4. `device-group.tsx`
A device section in the diff view.

**Props:**
- `deviceId: string`
- `status: "deployed" | "waiting" | "offline"`
- `metrics?: Array<{ name, before, after, format }>` — only present if
  deployed

**Behavior:**
- If `deployed`: shows device name, status dot (filled, green/red), and
  a table of `metric-row` components
- If `waiting`: shows device name, hollow dot, "Waiting" label, no table
- If `offline`: shows device name, `offline` role color, "Offline" label

**Stories:**
- Deployed with mixed severity metrics, waiting, offline.

#### 5. `deploy-tabs.tsx`
Horizontal tabs for switching between recent deploys.

**Props:**
- `deploys: Array<{ id, from, to, severity }>`
- `selected: string` — currently selected deploy ID
- `onSelect: (id: string) => void`

**Behavior:**
- Each tab shows `v1.4.2 → v1.4.3`
- Tab border/accent colored by worst severity in that deploy
- Selected tab is visually distinct (filled background)

**Stories:**
- Three tabs, one critical, two stable. Selected state for each.

#### 6. `demo-banner.tsx`
Banner shown when viewing simulated data.

**Props:**
- `onConnect?: () => void` — callback when "connect" is clicked

**Behavior:**
- Yellow/amber bar at the top
- "This is a demo fleet — connect your Balena fleet to get started"
- Dismissible? TBD.

**Stories:**
- Default state.

#### 7. `nav.tsx`
Top navigation bar.

**Props:**
- `fleetName: string`

**Behavior:**
- Left: Dispatch logo/wordmark
- Center or left-of-center: fleet name
- Right: settings icon (gear), maybe theme toggle later
- Fixed height: `NAV_HEIGHT` (56px)

**Stories:**
- Default with fleet name, demo mode.

#### 8. `diff-table.tsx`
Orchestrates the full diff view. This is the main page component.

**Props:**
- `deploy: { id, from, to, status, soakElapsed, soakTotal }`
- `devices: Array<DeviceGroup props>`

**Behavior:**
- Renders `status-badge` + `soak-timer` at top
- Renders list of `device-group` components
- Groups by device, sorted: deployed devices first (worst severity
  first), then waiting, then offline
- Action buttons at bottom: "Promote" and "Rollback"

**Stories:**
- Full diff with mixed device states. Canary in progress. Complete deploy.

## Simulation Data

File: `site/src/lib/demo-data.ts`

Hardcoded data representing a simulated fleet of 3 PiRacer cars mid-deploy.
This is what first-time users see.

```typescript
// Shape of the demo data — not final API types, just enough to render.
export const demoFleet = {
  name: "PiRacer Fleet (Demo)",
  deploys: [
    {
      id: "demo-1",
      from: "v1.4.2",
      to: "v1.4.3",
      status: "soaking",
      soakElapsed: 14,
      soakTotal: 30,
      devices: [
        {
          deviceId: "car-2",
          status: "deployed",
          metrics: [
            { name: "LiDAR accuracy", before: 98.1, after: 91.3, format: "percent" },
            { name: "CPU usage", before: 34, after: 48, format: "percent" },
            { name: "Memory", before: 412, after: 418, format: "bytes" },
            { name: "Fusion latency", before: 12, after: 14, format: "ms" },
          ],
        },
        { deviceId: "car-1", status: "waiting" },
        { deviceId: "car-3", status: "waiting" },
      ],
    },
    {
      id: "demo-2",
      from: "v1.4.1",
      to: "v1.4.2",
      status: "complete",
      devices: [
        {
          deviceId: "car-1",
          status: "deployed",
          metrics: [
            { name: "LiDAR accuracy", before: 97.8, after: 98.0, format: "percent" },
            { name: "CPU usage", before: 31, after: 33, format: "percent" },
          ],
        },
        {
          deviceId: "car-2",
          status: "deployed",
          metrics: [
            { name: "LiDAR accuracy", before: 97.9, after: 98.1, format: "percent" },
            { name: "CPU usage", before: 32, after: 34, format: "percent" },
          ],
        },
        {
          deviceId: "car-3",
          status: "deployed",
          metrics: [
            { name: "LiDAR accuracy", before: 97.7, after: 97.9, format: "percent" },
            { name: "CPU usage", before: 30, after: 31, format: "percent" },
          ],
        },
      ],
    },
  ],
};
```

## API Client

File: `site/src/api/client.ts`

Thin fetch wrapper that talks to `dispatchd`. For now, the three
endpoints that exist:

- `POST /api/metrics` — push device metrics
- `POST /api/snapshots` — take a before/after snapshot
- `GET /api/snapshots/:deployID/diff` — get the diff

The client should be configured with a base URL (defaults to
`window.location.origin` for self-hosted, configurable for demo).

For the initial build, the dashboard reads from `demo-data.ts` and
doesn't call the API. API integration comes after the UI is proven.

## Build Order

This is the recommended sequence. Each step produces visible, testable
output in Storybook:

1. **Set up Storybook** — config, theme decorator, verify it runs
2. **`metric-row`** — smallest component, most stories, validates the
   role color system works visually
3. **`status-badge`** — simple, validates the role → status mapping
4. **`soak-timer`** — simple, one visual state
5. **`device-group`** — composes `metric-row`, validates grouping
6. **`deploy-tabs`** — standalone, validates deploy switching
7. **`demo-banner`** — standalone, simple
8. **`nav`** — standalone, simple
9. **`diff-table`** — the full composition, uses everything above
10. **Wire up `page.tsx`** — render `diff-table` with demo data
11. **API client** — connect to real `dispatchd` endpoints
12. **Replace demo data with live API calls**

## Open Questions (Decide Later)

- **Severity thresholds** — how much delta is "critical" vs "warning"
  vs "stable"? Hardcode for now, make configurable later.
- **Metric formatting** — the `format` prop on metric-row. Keep it
  simple: percent appends `%`, bytes appends `MB`, ms appends `ms`,
  number is raw.
- **Responsive layout** — design for desktop first. Tablet/mobile later.
- **Auth** — no auth for now. Self-hosted instances are behind the
  operator's own network. Add API key auth when needed.
- **Real-time updates** — polling vs WebSocket for live metric updates
  during a soak. Polling is simpler. Start there.
- **Theme toggle** — dark mode is default. Add a toggle in nav later.
